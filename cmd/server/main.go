package main

import (
	"os"

	"github.com/lmikolajczak/go-cli-chat/pkg/transport"

	"github.com/lmikolajczak/go-cli-chat/pkg/chat"
)

func main() {
	sAddress := os.Args[1]

	serverApp := transport.CreateClientApp()
	serverApp.Serve(chat.NewSupervisor(), sAddress)
}
