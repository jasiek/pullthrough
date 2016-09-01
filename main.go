package main

import (
//	"github.com/pkg/errors"
	"fmt"
)

func main() {
	hs := NewHttpSource("http://releases.ubuntu.com/14.04/ubuntu-14.04.4-server-i386.template")
	fs := NewFileSink("test")
	
	destination := make(chan []byte)
	source := make(chan []byte)

	go hs.Stream(destination)
	go fs.Consume(source)

	for x := range destination {
		fmt.Println(len(x))
		source <- x
	}
}