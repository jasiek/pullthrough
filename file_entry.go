
package main

import (
	"io"
	"os"
	"net/http"
	"log"
	"sync"
)

type Consumer struct {
	Active bool
	Progress chan int64
}

func NewConsumer() (c *Consumer) {
	c = new(Consumer)
	c.Active = false
	c.Progress = make(chan int64)
	return c
}

type FileEntry struct {
	Mutex sync.Mutex
	URL string
	Filename string
	Length int64
	Downloaded int64
	Created bool
	CreatedNotifier chan bool
	Consumers []*Consumer
}

func NewFileEntry(url string) (fe *FileEntry) {
	fe = new(FileEntry)
	fe.URL = url
	fe.Filename = randomFilename("/tmp/")
	fe.Length = 0
	fe.Downloaded = 0
	fe.Created = false
	fe.CreatedNotifier = make(chan bool)
	fe.Consumers = make([]*Consumer, 0)
	return fe
}

func (fe *FileEntry) IsDone() bool {
	return fe.Downloaded == fe.Length
}

func (fe *FileEntry) UpdateProgress(n int) {
	fe.Downloaded += int64(n)
	for _, c := range fe.Consumers {
		if (c.Active) {
			c.Progress <- fe.Downloaded
		}
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
		defer f.Close()
		log.Print("Finished")
	}
	
}

func (fe *FileEntry) AddConsumer(c *Consumer) {
	fe.Mutex.Lock()
	fe.Consumers = append(fe.Consumers, c)
	fe.Mutex.Unlock()
}

func (fe *FileEntry) RemoveConsumer(c *Consumer) (index int) {
	fe.Mutex.Lock()
	for i, e := range fe.Consumers {
		if e == c {
			index = i
		}
	}
	fe.Consumers = append(fe.Consumers[:index], fe.Consumers[index+1:]...)
	fe.Mutex.Unlock()
	return index
}

func (fe *FileEntry) Push(w http.ResponseWriter) {
	log.Println("Pushing: " + fe.URL)
	if !fe.Created {
		<- fe.CreatedNotifier
	}
	f, err := os.Open(fe.Filename)
	log.Println("File opened", f, err)
	consumer := NewConsumer()
	fe.AddConsumer(consumer)
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

		consumer.Active = true
		select {
		case <- w.(http.CloseNotifier).CloseNotify():
			consumer.Active = false
			fe.RemoveConsumer(consumer)
			break
		case <- consumer.Progress:
			// NOOP
		}
	}
	consumer.Active = false
}