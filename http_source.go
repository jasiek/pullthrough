package main

import (
	"io"
	"net/http"
)

const (
	DEFAULT_CHUNK_LENGTH = 1024 * 1024
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
	chunk := make([]byte, DEFAULT_CHUNK_LENGTH, DEFAULT_CHUNK_LENGTH)
	resp, _ := http.Get(hs.Url)
	if (resp.StatusCode == 200) {
		for {
			n, err := resp.Body.Read(chunk)
			destination <- chunk[0:n]
			if err == io.EOF { break }
		}
		close(destination)
	}
}