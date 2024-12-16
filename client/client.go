package client

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"time"
	"github.com/n4vxn/FileMove/utils"
)

const (
	HOST = "localhost"
	PORT = "8080"
	TYPE = "tcp"
)

type Client struct {
	conn *tls.Conn
}

func NewClientConn(host, port string) *Client {
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
	return &Client{conn: conn}
}

func (c *Client) ReadInput() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter command: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("error reading input:", err)
			continue
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}
		parts := strings.Fields(input)
		if len(parts) != 2 {
			fmt.Println("Invalid input. Example: upload <filename>")
			continue
		}

		action := parts[0]
		filename := parts[1]

		switch action {
		case "UPLOAD":
			go c.UploadToServer(action, filename)
		case "DOWNLOAD":
			go c.DownloadFromServer(action, filename)
		default:
			fmt.Println("Unknown action:", action)
		}
		time.Sleep(3 * time.Second)
	}
}

func (c *Client) UploadToServer(action, filename string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(file.Name())
	defer file.Close()

	checksum, err := utils.GenerateChecksum(file)
	if err != nil {
		log.Fatal("error generating checksum:", err)
	}
	metadata := utils.GenerateUploadMetadata(action, file, checksum)
	_, err = c.conn.Write([]byte(metadata))
	if err != nil {
		log.Fatal("error sending metadata:", err)
	}

	file.Seek(0, io.SeekStart)

	_, err = io.Copy(c.conn, file)
	if err != nil {
		fmt.Println("cannot send data to the server")
	}
	log.Println("Data sent succesfully")
}

func (c *Client) DownloadFromServer(action, filename string) {
	outMetadata := utils.GenerateDownloadMetadata(action, filename)
	_, err := c.conn.Write([]byte(outMetadata))
	if err != nil {
		log.Fatal("Error sending metadata:", err)
	}

	metaDataBuffer := make([]byte, 4096)

	n, err := c.conn.Read(metaDataBuffer)
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

	_, err = io.CopyN(file, c.conn, int64(metadata.FileSize))
	if err != nil {
		fmt.Println("Error downloading data from server:", err)
		return
	}

	log.Println("Data downloaded successfully")
}
