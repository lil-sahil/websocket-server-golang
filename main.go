package main

import "github.com/lil-sahil/websocket-server-golang/server"

func main() {
	// Initialize new TCP Server
	server := server.NewServer("8080")
	server.Run()

}
