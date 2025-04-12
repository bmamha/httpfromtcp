package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func main() {
	file, err := os.Open("./messages.txt")
	if err != nil {
		log.Fatal()
	}
	stringsChannel := getLinesChannel(file)
	for line := range stringsChannel {
		fmt.Println("read: ", line)
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	stringsChannel := make(chan string)
	go func() {
		defer f.Close()
		defer close(stringsChannel)
		content := ""
		b := make([]byte, 8)
		for {
			n, err := f.Read(b)
			if err != nil {
				if err == io.EOF {
					break
				}
				fmt.Printf("error: %s\n", err.Error())
				return
			}
			str := string(b[:n])
			parts := strings.Split(str, "\n")
			for i := 0; i < len(parts)-1; i++ {

				stringsChannel <- fmt.Sprintf("%s%s", content, parts[i])
				content = ""
			}
			content += parts[len(parts)-1]
		}
	}()

	return stringsChannel
}
