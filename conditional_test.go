package main

import (
	"testing"
	"time"
	"net/http"
	"net/http/httptest"
)

const (
	SAMPLE_URL = "http://lol.biz/100MB.zip"
)

func TestUnconditionalGetPush(t *testing.T) {
	fe := NewFileEntry(SAMPLE_URL)
	fe.LastModified = time.Now()
	fe.ETag = "LOL"
	fe.Completed = true
	go fe.createFile()
	<-fe.CreatedNotifier

	req, _ := http.NewRequest("GET", "/nonexistent", nil)
	resp := httptest.NewRecorder()
	
	fe.Push(resp, req)

	if fe.ETag != resp.Header().Get("etag") {
		t.Fatal("doesn't set etag")
	}
	if fe.LastModified.Format(time.RFC1123) != resp.Header().Get("last-modified") {
		t.Fatal("doesn't set last modified")
	}
}

func TestConditionalGetPush(t *testing.T) {
	fe := NewFileEntry(SAMPLE_URL)
	fe.LastModified = time.Now()
	fe.ETag = "LOL"
	fe.Completed = true
	go fe.createFile()
	<-fe.CreatedNotifier

	req, _ := http.NewRequest("GET", "/nonexistent", nil)
	resp := httptest.NewRecorder()
	t.Run("match by etag", func(t *testing.T) {
		req.Header.Set("if-not-match", "LOL")
	})

	t.Run("match by last-modified", func(t *testing.T) {
		yesterday := time.Now().AddDate(1, 0, 0)
		req.Header.Set("if-modified-since", yesterday.Format(time.RFC1123))
	})

	fe.Push(resp, req)
	if resp.Result().StatusCode != 304 {
		t.Fatal("doesn't return 304", resp.Result().StatusCode)
	}
}
