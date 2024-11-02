package server

import (
	"fmt"
	"log"
	"net"
)

func Server() {
	l, err := net.Listen("tcp4", "127.0.0.1:8080")

	if err != nil {
		log.Fatalf("an error was found during listner setup: %v", err)
	}

	defer l.Close()

	for {
		conn, err := l.Accept()

		if err != nil {
			log.Fatalf("an error was found during connection setup: %v", err)
		}

		fmt.Print(conn)

		go handleConnection(conn)
	}

}

func handleConnection(c net.Conn) {
	fmt.Printf("Server connection: %v", c.RemoteAddr().String())

	packet := make([]byte, 4096)

	n, err := c.Read(packet)

	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(n)

	fmt.Println(string(packet[:n]))

}
