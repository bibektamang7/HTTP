package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
)

func getLinesChannel(fs io.ReadCloser) <-chan string {
	ch := make(chan string, 1)

	go func() {
		defer fs.Close()
		defer close(ch)

		str := ""

		read := make([]byte, 8)
		for {
			n, err := fs.Read(read)

			if n == 0 {
				break
			}
			if err != nil {
				break
			}
			idx := bytes.IndexByte(read, '\n')
			if idx == -1 {
				str += string(read[:n])
				continue
			}

			str += string(read[:idx])
			ch <- str
			str = string(read[idx+1 : n])
		}
		if len(str) != 0 {
			ch <- str
		}
	}()
	return ch
}

const port = ":42069"

func main() {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal("Failed to create lisener")
	}
	fmt.Println("Making connection")

	for {

		conn, err := listener.Accept()

		if err != nil {
			log.Fatal("Failed to create connection")
		}
		datas := getLinesChannel(conn)
		for data := range datas {
			fmt.Println(data)
		}
		fmt.Println("Connection is closed")
	}

}
