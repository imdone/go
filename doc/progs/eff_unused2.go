package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

var _ = fmt.Printf // For debugging; delete when done.
var _ io.Reader    // For debugging; delete when done.

func main() {
	fd, err := os.Open("test.go")
	if err != nil {
		log.Fatal(err)
	}
	// TODO: use fd. id:8 gh:9
	_ = fd
}
