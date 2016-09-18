package main

import (
	"io"
	"os"
	"net/http"
	"log"
	"time"
	"strconv"
)

const (
	DEFAULT_CHUNK_SIZE = 1024 * 1024
)

type Consumer struct {
	Progress chan int64
}

func NewConsumer() (c *Consumer) {
	c = new(Consumer)
	c.Progress = make(chan int64)
	return c
}

type FileEntry struct {
	URL string
	ETag string
	LastModified time.Time
	
	Filename string
	ContentType string
	Length int64
	Downloaded int64
	Created bool
	CreatedNotifier chan bool
	MetadataAvailable bool
	MetadataAvailableNotifier chan bool
	Completed bool
	Consumers []*Consumer
	AddConsumerChannel chan *Consumer
	RemoveConsumerChannel chan *Consumer
}

func NewFileEntry(url string) (fe *FileEntry) {
	fe = new(FileEntry)
	fe.URL = url
	fe.Filename = randomFilename("/tmp/")
	fe.Length = 0
	fe.Downloaded = 0
	fe.Created = false
	fe.Completed = false
	fe.CreatedNotifier = make(chan bool)
	fe.MetadataAvailableNotifier = make(chan bool)
	fe.Consumers = make([]*Consumer, 0)
	fe.AddConsumerChannel = make(chan *Consumer)
	fe.RemoveConsumerChannel = make(chan *Consumer)

	go func() {
		for c := range fe.RemoveConsumerChannel {
			var index int
			for i, e := range fe.Consumers {
				if e == c {
					index = i
				}
			}
			fe.Consumers = append(fe.Consumers[:index], fe.Consumers[index+1:]...)
		}
	}()

	go func() {
		for c := range fe.AddConsumerChannel {
			fe.Consumers = append(fe.Consumers, c)
		}
	}()

	return fe
}

func (fe *FileEntry) createFile() (f *os.File) {
	f, _ = os.Create(fe.Filename)
	fe.CreatedNotifier <- true
	fe.Created = true
	return f
}

func (fe *FileEntry) UpdateProgress(n int) {
	fe.Downloaded += int64(n)
	for _, c := range fe.Consumers {
		c.Progress <- fe.Downloaded
	}
}

func (fe *FileEntry) Pull() {
	log.Println("Pulling: " + fe.URL)
	resp, _ := http.Get(fe.URL)
	if (resp.StatusCode == 200) {
		// what happens if this is chunked?
		fe.Length = resp.ContentLength
		fe.ContentType = resp.Header.Get("content-type")
		etagString := resp.Header.Get("etag")
		if etagString != "" {
			fe.ETag = etagString
		}

		lastModifiedString := resp.Header.Get("last-modified")
		lastModified, err := time.Parse(time.RFC1123, lastModifiedString)
		if err != nil {
			fe.LastModified = lastModified
		}

		fe.MetadataAvailable = true
		fe.MetadataAvailableNotifier <- true
		
		chunk := make([]byte, DEFAULT_CHUNK_SIZE, DEFAULT_CHUNK_SIZE)

		f := fe.createFile()

		for {
			n, err := resp.Body.Read(chunk)
			f.Write(chunk[0:n])
			fe.UpdateProgress(n)
			if n == 0 && err == io.EOF { break }
		}

		f.Close()
		fe.Completed = true
		close(fe.AddConsumerChannel)
		close(fe.RemoveConsumerChannel)
		log.Print("Finished")
	}
	
}

func expired(r *http.Request, etag string, lastModified time.Time) (bool) {
	reqEtag := r.Header.Get("if-not-match")
	if reqEtag != "" && reqEtag != etag { return true }
	requestDateString := r.Header.Get("if-modified-since")
	requestDate, err := time.Parse(time.RFC1123, requestDateString)
	if err == nil {
		return requestDate.Before(lastModified)
	}
	return reqEtag == "" && requestDateString == ""

}

func (fe *FileEntry) PushNotModified(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotModified)
}

func (fe *FileEntry) Push(w http.ResponseWriter, r *http.Request) {
	// push headers first, regardless of what we're doing

	if !fe.MetadataAvailable {
		<- fe.MetadataAvailableNotifier
	}

	if fe.Completed {
		w.Header().Set("content-length", strconv.FormatInt(fe.Length, 10))
	}
	w.Header().Set("content-type", fe.ContentType)
	if !fe.LastModified.IsZero() {
		w.Header().Set("last-modified", fe.LastModified.Format(time.RFC1123))
	}
	if fe.ETag != "" {
		w.Header().Set("etag", fe.ETag)
	}

	if expired(r, fe.ETag, fe.LastModified) {
		if fe.Completed {
			fe.PushCompleted(w)
		} else {
			fe.PushIncomplete(w)
		}
	} else {
		fe.PushNotModified(w)
	}
}

func (fe *FileEntry) PushCompleted(w http.ResponseWriter) {
	log.Println("Pushing from cache: " + fe.URL)
	f, _ := os.Open(fe.Filename)
	io.Copy(w, f)
	f.Close()
}

func (fe *FileEntry) PushIncomplete(w http.ResponseWriter) {
	log.Println("Pushing: " + fe.URL)
	if !fe.Created {
		log.Println("Waiting until file gets created")
		<- fe.CreatedNotifier
		log.Println("file created")
	}
	f, err := os.Open(fe.Filename)
	log.Println("File opened", f, err)
	consumer := NewConsumer()
	fe.AddConsumerChannel <- consumer

	notifier := w.(http.CloseNotifier).CloseNotify()
	
	for pos := int64(0); pos < fe.Length ; {
		for {
			if pos < fe.Downloaded {
				n, err := io.CopyN(w, f, DEFAULT_CHUNK_SIZE)
				if n == 0 && err == io.EOF { break }
				pos += n
			} else {
				break
			}

			select {
			case <- notifier:
				fe.RemoveConsumerChannel <- consumer
				break
			default:
			}
		}

		<- consumer.Progress
	}
	fe.RemoveConsumerChannel <- consumer
}