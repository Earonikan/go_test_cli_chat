package main

import (
	"fmt"
	"log"
	"os"

	"github.com/lmikolajczak/go-cli-chat/pkg/frontend"
)

func main() {

	// AUTH

	// authUI, err := frontend.NewAuthUI()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// // defer authUI.Close()

	// authUI.SetManagerFunc(authUI.Layout)
	// authUI.SetKeyBindings(authUI.Gui)

	// if err = authUI.Serve(); err != nil {
	// 	log.Fatal(err)
	// }

	// authUI.Close()

	// CHAT

	chatUI, err := frontend.NewChatUI()
	if err != nil {
		log.Fatal(err)
	}
	defer chatUI.Close()

	logType := os.Args[1]
	username := os.Args[2]
	password := os.Args[3]
	// email := os.Args[4]
	var email string
	switch logType {
	case "-S":
		email = os.Args[4]
	default:
		fmt.Println("Wrong key! Need to use -S for new user registration, or -L for login")
	}

	// log.Println("")
	chatUI.SetUsername(username)
	chatUI.SetPassword(password)
	chatUI.SetEmail(email)

	if err = chatUI.Connect(logType); err != nil {
		log.Fatal(err)
	}

	chatUI.SetManagerFunc(chatUI.Layout)
	chatUI.SetKeyBindings(chatUI.Gui)

	if err = chatUI.Serve(); err != nil {
		log.Fatal(err)
	}
}
