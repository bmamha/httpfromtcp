package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
)

func main() {
	l, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer l.Close()
	for {
		conn, err := l.Accept()
		fmt.Println("A connection has been accepted")
		if err != nil {
			log.Fatal(err)
		}
		cs := getLinesChannel(conn)
		for line := range cs {
			fmt.Printf("%s\n", line)
		}

		fmt.Println("Channel is closed")
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	stringsChannel := make(chan string)
	go func() {
		defer f.Close()
		defer close(stringsChannel)

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			stringsChannel <- line
		}
		if err := scanner.Err(); err != nil {
			fmt.Printf("error: %s\n", err.Error())
		}
	}()

	return stringsChannel
}
