package main

import (
	//	"github.com/pkg/errors"
	"github.com/kr/pretty"
	//	"fmt"
	"log"
	"net/http"
	"time"
	"sync"
)

import _ 	"net/http/pprof"


type RequestHandler struct {
	Files map[string]*FileEntry
	Mutex sync.Mutex
}

func NewRequestHandler() (rh *RequestHandler) {
	rh = new(RequestHandler)
	rh.Files = make(map[string]*FileEntry)
	return rh
}

func (rh *RequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := r.URL.String()
	log.Println("Incoming request for " + key)
	rh.Mutex.Lock()
	if fe, ok := rh.Files[key]; ok {
		rh.Mutex.Unlock()
		fe.Push(w)
	} else {
		fe := NewFileEntry(key)
		rh.Files[key] = fe
		rh.Mutex.Unlock()
		log.Println("saving as " + fe.Filename)
		go fe.Pull()
		fe.Push(w)
	}
}

func main() {
	mux := http.NewServeMux()
	handler := NewRequestHandler()

	ticker := time.NewTicker(time.Second * 10)
	go func() {
		for range ticker.C {
		 	pretty.Println(handler)
		}
	}()

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	
	mux.Handle("/", handler)
	http.ListenAndServe(":3000", mux)
}