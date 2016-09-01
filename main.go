package main

import (
	"github.com/pkg/errors"
	"fmt"
	"net/http"
	"os"
	"log"
	"time"
	"io"
	"net/url"
	"math/rand"
)

var url_to_cached_filename = make(map[string]string)

func RandomString(strlen int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

func save(body io.ReadCloser, filename string) {
	f, err := os.Create("/tmp/" + filename)
	if err != nil {
		log.Fatal(err)
	}
	n, err := io.Copy(f, body)
	if err != nil {
		log.Fatal(errors.Wrap(err, "couldn't write to file"))
	} else {
		fmt.Println("written: %d\n", n)
	}
	defer body.Close()
	defer f.Close()
}

func fetchAndSave(url *url.URL, filename string) {
	uri := url.String()
	fmt.Println(uri)
	resp, err := http.Get(uri)
	if err != nil {
		log.Fatal(err)
		return
	}
	save(resp.Body, filename)
	url_to_cached_filename[uri] = filename	
}

func handler(w http.ResponseWriter, r *http.Request) {
	uri := r.URL.String()
	log.Println("GET " + uri)
	filename, ok := url_to_cached_filename[uri]
	if ok {
		log.Println("content is in cache")
		// fall through
	} else {
		filename := RandomString(32)
		log.Println("content not in cache, creating new entry: " + filename)
		fetchAndSave(r.URL, filename)
		log.Println("content saved")
	}
	
	file, err := os.Open("/tmp/" + filename)
	if err != nil {
		log.Fatal(errors.Wrap(err, "couldn't open file"))
	} else {
		log.Println("serving content")
		http.ServeContent(w, r, "", time.Now(), file)
		return
	}
}

func main() {
	// map of filename to disk file
	
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}