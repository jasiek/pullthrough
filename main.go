package main

import (
	"fmt"
	"net/http"
	"os"
	"log"
	"time"
)

func handler(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open("main.go")
	fmt.Printf("%s", r.URL)
	if err != nil {
		log.Fatal(err)
	} else {
		http.ServeContent(w, r, "lol.txt", time.Now(), file)
		return
	}
}

func main() {
	// map of filename to disk file
	
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}