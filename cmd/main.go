package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/n4vxn/FileMove/db"
)

var (
	HOST     = "localhost"
	PORT     = "8080"
	certfile = "./tls/server.crt"
	keyfile  = "./tls/server.key"
	RootDir = "FileMove-Storage"
)

func main() {
	db.ConnectDB()

	var action string

	for {
		prompt := &survey.Select{
			Message: "What would you like to do?",
			Options: []string{"Sign Up", "Login", "Exit"},
			Default: "Sign Up",
		}

		err := survey.AskOne(prompt, &action)
		if err != nil {
			log.Fatalf("Failed to get input: %v", err)
		}

		switch action {
		case "Sign Up":
			SignUpCmd.Run(nil, nil)
		case "Login":
			err := LoginCmd.RunE(nil, nil)
			if err != nil {
				fmt.Println("Login failed, try again.")
			} else {
				StartServer()
			}
		case "Exit":
			fmt.Println("Exiting the program...")
			return
		default:
			fmt.Println("Invalid option. Please try again.")
		}
	}
}

func StartServer() {
	serv := NewServer(Config{}) // Initialize the server

	cert, err := tls.LoadX509KeyPair(certfile, keyfile)
	if err != nil {
		log.Fatalf("Error loading certificate: %v", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		err := serv.Start(ctx, tlsConfig)
		if err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	time.Sleep(2 * time.Second)

	client := NewClientConn(ctx, HOST, PORT)
	client.ReadInput()
}

func CreateDir() {
	if err := os.MkdirAll(RootDir, 0775); err != nil {
		log.Println("Failed to create directory:", err)
		return
	}
}
