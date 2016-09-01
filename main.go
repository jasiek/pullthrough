package main

import (
//	"github.com/pkg/errors"
	"fmt"
)

func main() {
	prod := NewProducer("http://releases.ubuntu.com/14.04/ubuntu-14.04.4-server-i386.template", "test")
	progress := make(chan int64)
	done := make(chan bool)

	go prod.Produce(progress, done)

	for {
		select {
		case p := <-progress:
			fmt.Println(p)
		case <-done:
			fmt.Println("finished")
			return
		}
	}
}