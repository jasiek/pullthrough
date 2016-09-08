package main

import (
	"io"
	"os"
	"net/http"
	"log"
)

type PartialFileEntry struct {
	Filename string
	Length int64
	Downloaded int64
}

type HttpPuller struct {
	URL string
	Entry *PartialFileEntry
}

func NewHttpPuller(url string, entry *PartialFileEntry) (ht *HttpPuller) {
	ht = new(HttpPuller)
	ht.URL = url
	ht.Entry = entry
	return ht
}

func (ht* HttpPuller) UpdateProgress(progress int) {
	ht.Entry.Downloaded += int64(progress)
}

func (ht *HttpPuller) PullAndSave(done chan<- bool) {
	log.Println("Pulling: " + ht.URL)
	resp, _ := http.Get(ht.URL)
	if (resp.StatusCode == 200) {
		chunk := make([]byte, 64000, 64000)
		f, _ := os.Create(ht.Entry.Filename)
		for {
			n, err := resp.Body.Read(chunk)
			f.Write(chunk[0:n])
			ht.UpdateProgress(n)
			log.Print(".")
			if n == 0 && err == io.EOF { break }
		}
		defer f.Close()
		log.Print("Finished")
		done <- true
	}
}
