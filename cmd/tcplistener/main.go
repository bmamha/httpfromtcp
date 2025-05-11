package main

import (
	"fmt"
	"log"
	"net"

	"github.com/bmamha/httpfromtcp/internal/request"
)

const port = ":42069"

func main() {
	l, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer l.Close()

	fmt.Println("Listening for TCP traffic on", port)
	for {
		conn, err := l.Accept()
		fmt.Println("A connection has been accepted from", conn.RemoteAddr())
		if err != nil {
			log.Fatalf("error parsing request: %s\n", err)
		}
		r, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Request has been processed")
		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", r.RequestLine.Method)
		fmt.Printf("- Target: %s\n", r.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", r.RequestLine.HttpVersion)

		fmt.Println("Headers:")
		for key, value := range r.Headers {
			fmt.Printf("- %s: %s\n", key, value)
		}

		fmt.Println("Body:")
		fmt.Println(string(r.Body))

	}
}
