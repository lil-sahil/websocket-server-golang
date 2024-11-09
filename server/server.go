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

type handshakeRequest struct {
	method           string
	httpVersion      string
	upgrade          string
	connection       string
	websocketKey     string
	websocketVersion string
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

	// Hande the handshake request
	handleHandshakeRequest(c)

}

func handleHandshakeRequest(c net.Conn) {
	reader := bufio.NewReader(c)

	requestLine, _ := reader.ReadString('\n')

	handShakeRequest := handshakeRequest{}

	parts := strings.Split(strings.TrimSpace(requestLine), " ")

	handShakeRequest.method = parts[0]
	handShakeRequest.httpVersion = parts[2]

	// Get the headers
	for {
		line, err := reader.ReadString('\n')

		if err != nil {
			log.Printf("Failed to read header line: %v", err)
			return
		}

		headerParts := strings.SplitN(line, ":", 2)

		if len(headerParts) == 2 {
			key := strings.TrimSpace(headerParts[0])
			value := strings.TrimSpace(headerParts[1])

			switch key {
			case "Upgrade":
				handShakeRequest.upgrade = value
			case "Connection":
				handShakeRequest.connection = value
			case "Sec-WebSocket-Key":
				handShakeRequest.websocketKey = value
			case "Sec-WebSocket-Version":
				handShakeRequest.websocketVersion = value
			}
		}

		if line == "\r\n" {
			break
		}
	}

	// Verify
	err := handShakeRequest.verifyHandshakeRequest()

	if err != nil {
		fmt.Printf(err.Error())
		c.Write([]byte(err.Error()))
	}

	// res := `HTTP/1.1 101 Switching Protocols Upgrade: websocket Connection: Upgrade Sec-WebSocket-Accept: s3pPLMBiTxaQ9kYGzzhZRbK+xOo=`

	// c.Write([]byte(res))
}

func (h *handshakeRequest) verifyHandshakeRequest() error {
	// Verify http version
	if h.httpVersion != "HTTP/1.1" {
		return fmt.Errorf("HTTP version not supported. recieved %v, expected HTTP/1.1", h.httpVersion)
	}

	// Verify Method
	if h.method != "GET" {
		return fmt.Errorf("invalid method. require GET recieved %v", h.method)
	}

	// Verify Upgrade
	if h.upgrade != "websocket" {
		return fmt.Errorf("invalid upgrade specified. require websocket recieved %v", h.upgrade)
	}

	// Verify Connection
	if h.connection != "Upgrade" {
		return fmt.Errorf("invalid connection specified. require upgrade recieved %v", h.connection)
	}

	// Verify Websocketkey
	if h.websocketKey == "" {
		return fmt.Errorf("required websocket key to be passed")
	}
	// Verify websocket version
	if h.websocketVersion != "13" {
		return fmt.Errorf("required websocket version 13, recieved %v", h.websocketVersion)
	}

	return nil
}

// func sendHandShakeResponse() {

// }
