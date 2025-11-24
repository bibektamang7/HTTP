package main

import (
	"fmt"
	"log"
	"net"

	"github.com/bibektamang7/httpFromScratch/internal/request"
)

const PORT = ":8000"

func main() {
	listener, err := net.Listen("tcp", PORT)

	if err != nil {
		log.Fatal("LISTENER FAILED TO INITIALIZE")
	}

	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Fatal("CONNECTION FAILED")
		}
		r, err := request.RequestFromReader(conn)
		if err != nil {
			fmt.Printf("Failed to parse request: %v", err)
			continue
		}
		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", r.RequestLine.Method)
		fmt.Printf("- Target: %s\n", r.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", r.RequestLine.HttpVersion)

		fmt.Println("Headers:")
		r.Headers.ForEach(func(name, value string) {
			fmt.Printf("- %s: %s\n", name, value)
		})

	}
}
