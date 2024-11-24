package server

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/lil-sahil/websocket-server-golang/types"
	"github.com/lil-sahil/websocket-server-golang/utils"
)

type server struct {
	port      string
	callbacks map[types.CallbackEvent]func(string)
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
		port:      port,
		callbacks: make(map[types.CallbackEvent]func(string)),
	}
}

func (s *server) RegisterCallBack(event types.CallbackEvent, cb func(string)) {
	s.callbacks[event] = cb
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

		go s.handleConnection(conn)
	}

}

func (s *server) handleConnection(c net.Conn) {
	fmt.Printf("Server connection: %v", c.RemoteAddr().String())

	// Hande the handshake request
	handleHandshakeRequest(c)

	// Create a message interceptor
	recievedMessage := utils.NewRecieveMessage(c, s.callbacks)

	// Read message continuously
	for {
		err := recievedMessage.HandleReciveMessage()
		if err != nil {
			fmt.Println(err.Error())
		}

		fmt.Println("recieving message")
	}

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

	// Derive Accept key
	key, err := handShakeRequest.deriveAcceptKey()

	// Formulate response
	res := handShakeRequest.createHandShakeResponse(*key)

	c.Write([]byte(res))

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

func (h *handshakeRequest) deriveAcceptKey() (*string, error) {
	magicString := "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	hasher := sha1.New()

	_, err := hasher.Write([]byte(h.websocketKey + magicString))

	if err != nil {
		return nil, err
	}

	key := base64.StdEncoding.EncodeToString(hasher.Sum(nil))

	return &key, nil
}

func (h *handshakeRequest) createHandShakeResponse(key string) string {
	return fmt.Sprintf(
		"HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: %v\r\n\r\n", key)
}
