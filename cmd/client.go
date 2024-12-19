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

	tlsConfig := &tls.Config{RootCAs: certPool}
	conn, err := tls.Dial(TYPE, host+":"+port, tlsConfig)
	if err != nil {
		log.Fatal("Failed to establish TLS connection:", err)
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
			Message: "Choose an action:",
			Options: []string{"Upload", "Download", "Exit"},
		}
		if err := survey.AskOne(prompt, &action); err != nil {
			log.Println("Error reading input:", err)
			continue
		}

		var filename string
		switch action {
		case "Upload":
			survey.AskOne(&survey.Input{Message: "Enter the filename to upload:"}, &filename)
			c.UploadToServer(filename)
		case "Download":
			survey.AskOne(&survey.Input{Message: "Enter the filename to download:"}, &filename)
			c.DownloadFromServer(filename)
		case "Exit":
			fmt.Println("Exiting the program...")
			return
		default:
			fmt.Println("Invalid option. Please try again.")
		}
	}
}

func (c *Client) UploadToServer(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Println("File not found:", err)
		return
	}
	defer file.Close()

	checksum, err := utils.GenerateChecksum(file)
	if err != nil {
		log.Println("Checksum generation failed:", err)
		return
	}

	metadata := utils.GenerateUploadMetadata(c.Username, "Upload", file, checksum)
	if _, err := c.Conn.Write([]byte(metadata)); err != nil {
		log.Println("Failed to send metadata:", err)
		return
	}

	file.Seek(0, io.SeekStart)
	if _, err := io.Copy(c.Conn, file); err != nil {
		log.Println("Failed to upload file:", err)
		return
	}
	log.Println("File uploaded successfully")
}

func (c *Client) DownloadFromServer(filename string) {
	metadata := utils.GenerateDownloadMetadata(c.Username, "Download", filename)
	if _, err := c.Conn.Write([]byte(metadata)); err != nil {
		log.Println("Failed to send download request:", err)
		return
	}

	metaDataBuffer := make([]byte, 4096)
	n, err := c.Conn.Read(metaDataBuffer)
	if err != nil {
		log.Println("Failed to receive metadata:", err)
		return
	}

	uploadMetadata, err := utils.ParseUploadMetadata(string(metaDataBuffer[:n]))
	if err != nil || !utils.ValidateUploadMetadata(*uploadMetadata) {
		log.Println("Invalid metadata received")
		return
	}

	dirName := "down-" + strings.TrimSuffix(uploadMetadata.Name, path.Ext(uploadMetadata.Name))
	if err := os.MkdirAll(dirName, 0775); err != nil {
		log.Println("Failed to create directory:", err)
		return
	}

	filePath := path.Join(dirName, uploadMetadata.Name)
	file, err := os.Create(filePath)
	if err != nil {
		log.Println("Failed to create file:", err)
		return
	}
	defer file.Close()

	if _, err := io.CopyN(file, c.Conn, int64(uploadMetadata.FileSize)); err != nil {
		log.Println("Failed to download file:", err)
		return
	}
	log.Println("File downloaded successfully:", filePath)
}
