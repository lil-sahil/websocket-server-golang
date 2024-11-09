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

	// Hande the handshake request
	handleHandshakeRequest(c)

}

func handleHandshakeRequest(c net.Conn) {
	reader := bufio.NewReader(c)

	requestLine, _ := reader.ReadString('\n')

	fmt.Println(requestLine)

	parts := strings.Split(strings.TrimSpace(requestLine), " ")

	method := parts[0]
	httpVersion := parts[2]

	headers := make(map[string]string)

	// Get the headers
	for {
		line, err := reader.ReadString('\n')

		if err != nil {
			log.Printf("Failed to read header line: %v", err)
			return
		}

		headerParts := strings.SplitN(line, ":", 2)

		if len(headerParts) == 2 {
			headers[strings.TrimSpace(headerParts[0])] = strings.TrimSpace(headerParts[1])
		}

		if line == "\r\n" {
			break
		}
	}

	// Verify
	err := verifyHandshakeRequest(httpVersion, method, headers["Upgrade"], headers["Connection"], headers["Sec-WebSocket-Key"], headers["Sec-WebSocket-Version"])

	if err != nil {
		fmt.Printf(err.Error())
		c.Write([]byte(err.Error()))
	}

	// res := `HTTP/1.1 101 Switching Protocols Upgrade: websocket Connection: Upgrade Sec-WebSocket-Accept: s3pPLMBiTxaQ9kYGzzhZRbK+xOo=`

	// c.Write([]byte(res))
}

func verifyHandshakeRequest(httpVersion, method, upgrade, connection, websocketKey, websocketVersion string) error {
	// Verify http version
	if httpVersion != "HTTP/1.1" {
		return fmt.Errorf("HTTP version not supported. recieved %v, expected HTTP/1.1", httpVersion)
	}

	// Verify Method
	if method != "GET" {
		return fmt.Errorf("invalid method. require GET recieved %v", method)
	}

	// Verify Upgrade
	if upgrade != "websocket" {
		return fmt.Errorf("invalid upgrade specified. require websocket recieved %v", upgrade)
	}

	// Verify Connection
	if connection != "Upgrade" {
		return fmt.Errorf("invalid connection specified. require upgrade recieved %v", connection)
	}

	// Verify Websocketkey
	if websocketKey == "" {
		return fmt.Errorf("required websocket key to be passed")
	}
	// Verify websocket version
	if websocketVersion == "" {
		return fmt.Errorf("required websocket version to be passed")
	}

	return nil
}

// func sendHandShakeResponse() {

// }
