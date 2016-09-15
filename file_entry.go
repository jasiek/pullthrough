
package main

import (
	"io"
	"os"
	"net/http"
	"log"
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
	Filename string
	Length int64
	Downloaded int64
	Created bool
	Completed bool
	CreatedNotifier chan bool
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
		fe.Length = resp.ContentLength
		chunk := make([]byte, 64000, 64000)
		f, _ := os.Create(fe.Filename)

		fe.CreatedNotifier <- true
		fe.Created = true

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

func (fe *FileEntry) Push(w http.ResponseWriter) {
	if fe.Completed {
		fe.PushCompleted(w)
	} else {
		fe.PushIncomplete(w)
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
	for pos := int64(0); pos < fe.Length ; {
		log.Println(pos, fe.Downloaded, fe.Length)
		for {
			if pos < fe.Downloaded {
				n, _ := io.CopyN(w, f, 64000)
				pos += n
			} else {
				break
			}
		}

		select {
		case <- w.(http.CloseNotifier).CloseNotify():
			fe.RemoveConsumerChannel <- consumer
			break
		case <- consumer.Progress:
		default:
		}
	}
	fe.RemoveConsumerChannel <- consumer
}