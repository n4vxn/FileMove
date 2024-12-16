package main

import (
	"context"
	"crypto/tls"
	"log"
	"time"
	"github.com/n4vxn/FileMove/client"
	"github.com/n4vxn/FileMove/server"
)

var (
	HOST     = "localhost"
	PORT     = "8080"
	certfile = "./tls/server.crt"
	keyfile  = "./tls/server.key"
)

func main() {
	serv := server.NewServer(server.Config{})

	cert, err := tls.LoadX509KeyPair(certfile, keyfile)
	if err != nil {
		log.Fatal("Error loading certificate:", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go serv.Start(ctx, tlsConfig)
	time.Sleep(2 * time.Second)

	client := client.NewClientConn(HOST, PORT)
	client.ReadInput()
}
