package main

import (
	"io"
	"net/http"
)

type HttpSource struct {
	Url string
}

func NewHttpSource(url string) (hs *HttpSource) {
	hs = new(HttpSource)
	hs.Url = url
	return hs
}

func (hs *HttpSource) Stream(destination chan<- []byte) {
	chunk := make([]byte, 64000, 64000)
	resp, _ := http.Get(hs.Url)
	if (resp.StatusCode == 200) {
		for {
			n, err := resp.Body.Read(chunk)
			destination <- chunk[0:n]
			if n == 0 && err == io.EOF { break }
		}
		close(destination)
	}
}