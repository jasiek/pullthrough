
package main

import (
	"io"
	"os"
	"net/http"
	"log"
)

type FileEntry struct {
	URL string
	Filename string
	Length int64
	Downloaded int64
	Created chan bool
	Watchers []chan int64
}

func NewFileEntry(url string) (fe *FileEntry) {
	fe = new(FileEntry)
	fe.URL = url
	fe.Filename = randomFilename("/tmp/")
	fe.Length = 0
	fe.Downloaded = 0
	fe.Created = make(chan bool)
	fe.Watchers = make([]chan int64, 0)
	return fe
}

func (fe *FileEntry) IsDone() bool {
	return fe.Downloaded == fe.Length
}

func (fe *FileEntry) UpdateProgress(n int) {
	fe.Downloaded += int64(n)
	for _, ch := range fe.Watchers {
		ch <- fe.Downloaded
	}
}

func (fe *FileEntry) Pull() {
	log.Println("Pulling: " + fe.URL)
	resp, _ := http.Get(fe.URL)
	if (resp.StatusCode == 200) {
		fe.Length = resp.ContentLength
		chunk := make([]byte, 64000, 64000)
		f, _ := os.Create(fe.Filename)
		fe.Created <- true
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

func (fe *FileEntry) Push(w io.Writer) {
	log.Println("Pushing: " + fe.URL)
	<- fe.Created
	f, err := os.Open(fe.Filename)
	log.Println("File opened", f, err)
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
	}
}