package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

type server struct {
	port string
}

func NewServer(port string) *server {
	return &server{
		port: port,
	}
}

func (s *server) Run() {
	l, err := net.Listen("tcp4", fmt.Sprintf("127.0.0.1:%v", s.port))

	if err != nil {
		log.Fatalf("an error was found during listner setup: %v", err)
	}

	defer l.Close()

	for {
		conn, err := l.Accept()

		if err != nil {
			log.Fatalf("an error was found during connection setup: %v", err)
		}

		go handleConnection(conn)
	}

}

func handleConnection(c net.Conn) {
	fmt.Printf("Server connection: %v", c.RemoteAddr().String())

	reader := bufio.NewReader(c)

	requestLine, _ := reader.ReadString('\n')

	fmt.Println(requestLine)

	parts := strings.Split(strings.TrimSpace(requestLine), " ")

	fmt.Println(parts[0], parts[1], parts[2])

}

// func handleHandshakeRequest() {

// }
