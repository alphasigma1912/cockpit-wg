package main

import (
	"crypto/sha256"
	"fmt"
)

func main() {
	data := []byte("hello")
	hash := sha256.Sum256(data)
	fmt.Printf("%x\n", hash)
}
