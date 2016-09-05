package main

import (
	"io"
//	"fmt"
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
	resp, _ := http.Get(hs.Url)
	if (resp.StatusCode == 200) {
		for {
			chunk := make([]byte, 64000, 64000)
			n, err := resp.Body.Read(chunk)
			destination <- chunk[0:n]
			if n == 0 && err == io.EOF { break }
		}
		close(destination)
	}
}