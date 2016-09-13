package main

import (
	//	"github.com/pkg/errors"
	"github.com/kr/pretty"
//	"fmt"
	"io"
	"os"
	"log"
	"net/http"
	"time"
)

type RequestHandler struct {
	Files map[string]*FileEntry
}

func NewRequestHandler() (rh *RequestHandler) {
	rh = new(RequestHandler)
	rh.Files = make(map[string]*FileEntry)
	return rh
}

func (rh *RequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := r.URL.String()
	log.Println("Incoming request for " + key)
	if fe, ok := rh.Files[key]; ok {
		if fe.IsDone() {
			log.Println("Found in cache, sending")
			f, _ := os.Open(fe.Filename)
			io.Copy(w, f)
			f.Close()
		} else {
			log.Println("Found partial, sending")
			go fe.Push(w)
		}
		return
	}

	// create a new partial file and start streaming
	fe := NewFileEntry(key)
	rh.Files[key] = fe
	log.Println("saving as " + fe.Filename)
	go fe.Pull()
	defer fe.Push(w)
}

func main() {
	mux := http.NewServeMux()
	handler := NewRequestHandler()

	ticker := time.NewTicker(time.Second * 10)
	go func() {
		for {
			<- ticker.C
			pretty.Println(handler)
		}
	}()
	
	mux.Handle("/", handler)
	http.ListenAndServe(":3000", mux)
}