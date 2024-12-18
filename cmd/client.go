package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/n4vxn/FileMove/utils"
)

type Client struct {
	ContextualConn
}

func NewClientConn(ctx context.Context, host, port string) *Client {
	certPool := x509.NewCertPool()
	serverCert, err := os.ReadFile("./tls/server.crt")
	if err != nil {
		log.Fatal("Failed to read server certificate:", err)
	}
	certPool.AppendCertsFromPEM(serverCert)

	// Configure TLS
	tlsConfig := &tls.Config{
		RootCAs: certPool,
	}

	// Establish the TLS connection
	conn, err := tls.Dial(TYPE, host+":"+port, tlsConfig)
	if err != nil {
		log.Fatal(err)
	}

	return &Client{
		ContextualConn: ContextualConn{
			Conn:     conn,
			ctx:      ctx,
			Username: currentUser,
		},
	}
}

func (c *Client) ReadInput() {
	for {
		var action string
		prompt := &survey.Select{
			Message: "What would you like to do?",
			Options: []string{"Upload", "Download", "Exit"},
			Default: "Upload",
		}

		err := survey.AskOne(prompt, &action)
		if err != nil {
			log.Fatalf("Failed to get input: %v", err)
		}

		switch action {
		case "Upload":
			var filename string
			survey.AskOne(&survey.Input{Message: "Enter the filename to upload:"}, &filename)
			go c.UploadToServer(c.Username, action, filename)
		case "Download":
			var filename string
			survey.AskOne(&survey.Input{Message: "Choose the file to download:"}, &filename)
			go c.DownloadFromServer(c.Username, action, filename)
		default:
			return
		}
	}
}

func (c *Client) UploadToServer(username, action, filename string) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		log.Println("File does not exist.")
		return
	}
	
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	checksum, err := utils.GenerateChecksum(file)
	if err != nil {
		log.Fatal("error generating checksum:", err)
	}
	metadata := utils.GenerateUploadMetadata(username, action, file, checksum)
	_, err = c.Conn.Write([]byte(metadata))
	if err != nil {
		log.Fatal("error sending metadata:", err)
	}

	file.Seek(0, io.SeekStart)

	_, err = io.Copy(c.Conn, file)
	if err != nil {
		fmt.Println("cannot send data to the server")
	}
	log.Println("Data sent succesfully")
}

func (c *Client) DownloadFromServer(username, action, filename string) {
	outMetadata := utils.GenerateDownloadMetadata(username, action, filename)
	_, err := c.Conn.Write([]byte(outMetadata))
	if err != nil {
		log.Fatal("Error sending metadata:", err)
	}

	metaDataBuffer := make([]byte, 4096)

	n, err := c.Conn.Read(metaDataBuffer)
	if err != nil {
		log.Printf("Error reading metadata: %v", err)
		return
	}

	incMetadata := string(metaDataBuffer[:n])
	metadata, err := utils.ParseUploadMetadata(incMetadata)
	if err != nil {
		log.Printf("Error parsing metadata: %v", err)
		return
	}

	log.Println(metadata)

	if utils.ValidateUploadMetadata(*metadata) {
	} else {
		log.Println("Invalid metadata")
		return
	}

	ext := path.Ext(metadata.Name)
	folderName := strings.TrimSuffix(metadata.Name, ext)

	if _, err := os.Stat("down-" + folderName); os.IsNotExist(err) {
		err = os.Mkdir("down-"+folderName, 0775)
		if err != nil {
			fmt.Println("Error creating directory:", err)
			return
		}
	}

	file, err := os.Create("./down-" + folderName + "/" + metadata.Name)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	_, err = io.CopyN(file, c.Conn, int64(metadata.FileSize))
	if err != nil {
		fmt.Println("Error downloading data from server:", err)
		return
	}

	log.Println("Data downloaded successfully")
}
