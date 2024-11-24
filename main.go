package main

import (
	"fmt"

	"github.com/lil-sahil/websocket-server-golang/server"
	"github.com/lil-sahil/websocket-server-golang/types"
)

func main() {
	// Initialize new TCP Server
	server := server.NewServer("8080")

	server.RegisterCallBack(types.MessageEvent, func(e string) {
		if e == "hello" {
			fmt.Println("hi")
		}
	})

	server.Run()

}
