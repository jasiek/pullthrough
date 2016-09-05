package main

import (
//	"github.com/pkg/errors"
//	"fmt"
)

func main() {
	hs := NewHttpSource("http://localhost:8000/wl.txt")
	fs := NewFileSink("test")
	
	destination := make(chan []byte)
	source := make(chan []byte)
	
	go hs.Stream(destination)
	go fs.Consume(source)

	for x := range destination {
		source <- x
	}
}