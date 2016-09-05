package main

import (
	"os"
//	"fmt"
)

type FileSink struct {
	Filename string
}

func NewFileSink(filename string) (fs *FileSink) {
	fs = new(FileSink)
	fs.Filename = filename
	return fs
}

func (fs *FileSink) Consume(source <-chan []byte) {
	file, _ := os.Create(fs.Filename)
	for chunk := range source {
		file.Write(chunk)
	}
	file.Close()
}