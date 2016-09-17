package main

import (
	"testing"
	"net/http"
	"net/http/httptest"
	"strconv"
)

func testRequest() (req *http.Request) {
	req, _ = http.NewRequest("GET", "/nonexistent", nil)
	return req
}

func TestMetadata(t *testing.T) {
	fe := NewFileEntry(SAMPLE_URL)
	fe.ContentType = "text/plain"
	fe.Length = 666
	fe.Completed = true
	go fe.createFile()
	<- fe.CreatedNotifier

	req := testRequest()
	resp := httptest.NewRecorder()

	fe.Push(resp, req)
	requestLen, _ := strconv.ParseInt(resp.Header().Get("content-length"), 10, 64)
	if (requestLen != fe.Length) {
		t.Fatal("wrong content length, ", requestLen)
	}
	if (fe.ContentType != resp.Header().Get("content-type")) {
		t.Fatal("wrong content type")
	}
}