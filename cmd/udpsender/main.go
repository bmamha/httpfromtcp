package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	udp, err := net.ResolveUDPAddr("udp", "localhost: 42069")
	if err != nil {
		log.Fatal(err.Error())
	}
	conn, err := net.DialUDP("udp", nil, udp)
	if err != nil {
		log.Fatal(err.Error())
	}

	defer conn.Close()
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("<")
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err.Error())
		}
		_, err = conn.Write([]byte(line))
		if err != nil {
			log.Fatal(err.Error())
		}
	}
}
