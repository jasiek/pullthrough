package main

import (
	"io"
	"os"
	"net/http"
	"log"
)

type HttpPuller struct {
	Entry *FileEntry
}

func NewHttpPuller(entry *FileEntry) (ht *HttpPuller) {
	ht = new(HttpPuller)
	ht.Entry = entry
	return ht
}

func (ht* HttpPuller) UpdateProgress(progress int) {
	ht.Entry.Downloaded += int64(progress)
	ht.Entry.NotifySinks()
}

func (ht *HttpPuller) PullAndSave() {
	log.Println("Pulling: " + ht.Entry.URL)
	resp, _ := http.Get(ht.Entry.URL)
	if (resp.StatusCode == 200) {
		ht.Entry.Length = resp.ContentLength
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
	}
}
