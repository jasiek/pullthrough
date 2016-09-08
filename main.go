package main

import (
//	"github.com/pkg/errors"
	"fmt"
	"io"
	"os"
	"log"
	"net/http"
)

type RequestHandler struct {
	CompleteFiles map[string]string
	PartialFiles map[string]*PartialFileEntry
}

func NewRequestHandler() (rh *RequestHandler) {
	rh = new(RequestHandler)
	rh.CompleteFiles = make(map[string]string)
	rh.PartialFiles = make(map[string]*PartialFileEntry)
	return rh
}

func (rh *RequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := r.URL.String()
	log.Println("Incoming request for " + key)
	if filename, ok := rh.CompleteFiles[key]; ok {
		log.Println("Found in cache, sending")
		f, _ := os.Open(filename)
		io.Copy(w, f)
		f.Close()
		return
	}
	// partial file servicing goes here
	// create a new partial file and start streaming
	pfe := new(PartialFileEntry)
	pfe.Filename = randomFilename("/tmp/")
	log.Println("saving as " + pfe.Filename)
	puller := NewHttpPuller(key, pfe)
	rh.PartialFiles[key] = pfe
	done := make(chan bool)
	go puller.PullAndSave(done)
	go func(done <-chan bool) {
		<- done
		delete(rh.PartialFiles, key)
		rh.CompleteFiles[key] = pfe.Filename
		log.Println("Added to complete files")
	}(done)

	fmt.Println(rh)
}

func main() {
	mux := http.NewServeMux()
	handler := NewRequestHandler()
	mux.Handle("/", handler)
	http.ListenAndServe(":3000", mux)
}