package main

import (
	"io"
	"os"
)


type HttpPusher struct {
	Entry *FileEntry
}

func NewHttpPusher(entry *FileEntry) (hp *HttpPusher) {
	hp = new(HttpPusher)
	hp.Entry = entry
	return hp
}

func (hp *HttpPusher) Push(w io.Writer) {
	f, _ := os.Open(hp.Entry.Filename)
	pos := int64(0)
	sink := make(chan bool)
	hp.Entry.AddSink(sink)
	for ; pos <= hp.Entry.Length ;  {
		for {
			if pos < hp.Entry.Downloaded {
				n, _ := io.CopyN(w, f, 64000)
				pos += n
			} else {
				break
			}
		}
		<- sink
	}
	hp.Entry.RemoveSink(sink)
	defer f.Close()
}