package main

import (
//	"github.com/pkg/errors"
	"fmt"
	"io"
	"os"
	"net/http"
)

type RequestHandler struct {
	CompleteFiles map[string]string
}

func NewRequestHandler() (rh *RequestHandler) {
	rh = new(RequestHandler)
	rh.CompleteFiles = make(map[string]string)
	rh.CompleteFiles["http://localhost:8000/wl.txt"] = "test"
	return rh
}

func (rh *RequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.String())
	if filename, ok := rh.CompleteFiles[r.URL.String()]; ok {
		f, _ := os.Open(filename)
		io.Copy(w, f)
		f.Close()
		return
	}
}


func main() {
	mux := http.NewServeMux()
	handler := NewRequestHandler()
	mux.Handle("/", handler)
	http.ListenAndServe(":3000", mux)
}