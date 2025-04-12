package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	file, err := os.Open("./messages.txt")
	if err != nil {
		log.Fatal()
	}

	b := make([]byte, 8)
	for {
		n, err := file.Read(b)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Printf("error: %s\n", err.Error())
			break
		}
		str := string(b[:n])
		fmt.Printf("read: %s\n", str)
	}
}
