package main

import (
	"os"
	"io"
	"net/http"
)

type HttpConsumer struct {
	Filename string
	Progress chan int64
}

func NewConsumer(filename string) (c *HttpConsumer) {
	c = new(HttpConsumer)
	c.Filename = filename
	return c
}

func (c *HttpConsumer) Consume(w http.ResponseWriter, r *http.Request) {
	f, _ := os.Open(c.Filename)
	for progress := range c.Progress {
		io.Copy(w, f, progress)
	}
	defer f.Close()
}
