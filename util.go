package main

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
)

func randomFilename(prefix string) string {
	buffer := make([]byte, 32)
	rand.Read(buffer)
	sha1 := sha1.New()
	return prefix + hex.EncodeToString(sha1.Sum(buffer))
}