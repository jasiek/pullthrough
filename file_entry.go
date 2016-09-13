package main

import (
	"sync"
	"fmt"
)

type FileEntry struct {
	Lock sync.Mutex
	URL string
	Filename string
	Length int64
	Downloaded int64
	ProgressSinks []chan bool
}

func NewFileEntry(url string) (fe *FileEntry) {
	fe = new(FileEntry)
	fe.URL = url
	fe.Filename = randomFilename("/tmp/")
	fe.Length = 0
	fe.Downloaded = 0
	fe.ProgressSinks = make([]chan bool, 0)
	return fe
}

func (fe *FileEntry) IsDone() bool {
	return fe.Downloaded == fe.Length
}

func (fe *FileEntry) AddSink(sink chan bool) {
	fe.Lock.Lock()
	fe.ProgressSinks = append(fe.ProgressSinks, sink)
	fe.Lock.Unlock()
}

func (fe *FileEntry) NotifySinks() {
	fe.Lock.Lock()
	for _, c := range fe.ProgressSinks {
		fmt.Println("notifying channel", c)
		c <- true
	}
	fe.Lock.Unlock()
}

func (fe *FileEntry) RemoveSink(sink chan bool) {
	fe.Lock.Lock()
	index := indexOf(fe.ProgressSinks, sink)
	if index != -1 {
		fe.ProgressSinks = append(fe.ProgressSinks[:index], fe.ProgressSinks[index+1:]...)
	}
	fe.Lock.Unlock()
}

func indexOf(array []chan bool, element chan bool) int {
	for i, e := range array {
		if e == element {
			return i
		}
	}
	return -1
}