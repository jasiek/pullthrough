package main

import (
	"os"
	"io"
	"net/http"
)

const (
	DEFAULT_CHUNK_LENGTH = 1000000
)

type ContentInfo struct {
	Etag string
	Length int64
}

type FileProducer struct {
	Url string
	Filename string
	Info ContentInfo
}

func NewProducer(url string, filename string) (p *FileProducer) {
	p = new(FileProducer)
	p.Url = url
	p.Filename = filename
	return p
}

func (fp *FileProducer) Produce(progress chan int64, done chan bool) {
	var written int64
	
	resp, _ := http.Get(fp.Url)
	if (resp.StatusCode == 200) {
		fp.Info.Length = resp.ContentLength
		f, _ := os.Create(fp.Filename)
		
		for written = 0 ; written < fp.Info.Length ; {
			n, _ := io.CopyN(f, resp.Body, DEFAULT_CHUNK_LENGTH)
			written += n
			progress <- written
		}

		defer f.Close()
		
		done <- true
	}
}
