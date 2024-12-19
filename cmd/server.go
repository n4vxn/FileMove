package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"

	"github.com/n4vxn/FileMove/db"
	"github.com/n4vxn/FileMove/utils"
)

const (
	TYPE    = "tcp"
	maxConn = 10
)

type Config struct {
	HOST string
	PORT string
}

type ContextualConn struct {
	net.Conn
	ctx      context.Context
	Username string
}

func (c *ContextualConn) Context() context.Context {
	return c.ctx
}

type Server struct {
	listener net.Listener
	Config
	ContextualConn
	ctx context.Context
}

func NewServer(cfg Config) *Server {
	if len(cfg.PORT) == 0 {
		cfg.PORT = PORT
	}

	ctx := context.Background()
	return &Server{
		Config: cfg,
		ctx:    ctx,
	}
}

func (s *Server) Start(ctx context.Context, tlsConfig *tls.Config) error {
	var err error
	s.listener, err = tls.Listen(TYPE, HOST+":"+PORT, tlsConfig)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		<-ctx.Done()
		log.Println("Shutting down server...")
		s.listener.Close()
	}()
	go s.StartAcceptLoop()
	return nil
}

func (s *Server) StartAcceptLoop() {
	sem := make(chan struct{}, maxConn)
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			fmt.Println("failed to accept connection:", err)
			continue
		}

		if currentUser == "" {
			log.Println("Username not found!")
		}

		username := currentUser
		contextualConn := ContextualConn{
			Conn:     conn,
			ctx:      s.ctx,
			Username: username,
		}

		sem <- struct{}{} // Block if max connections reached
		go func(conn net.Conn) {
			defer func() { <-sem }() // Release semaphore slots
			s.handleRequests(&contextualConn)
		}(contextualConn)
	}
}

func (s *Server) handleRequests(conn *ContextualConn) {
	defer conn.Close()
	metaDataBuffer := make([]byte, 2048)

	username := conn.Username
	log.Printf("%s connected!\n", username)

	// Receive metadata
	n, err := conn.Read(metaDataBuffer)
	if err != nil {
		fmt.Printf("Error reading metadata: %v", err)
		return
	}

	metadata := string(metaDataBuffer[:n])
	metaDataBuffer = nil

	parts := strings.Split(strings.TrimSpace(metadata), "|")
	if parts[1] == "Upload" {
		if len(parts) != 5 {
			conn.Write([]byte("Invalid file upload metadata format"))
			return
		} else {
			s.handleUpload(metadata, conn)
		}

	} else if parts[1] == "Download" {
		if len(parts) != 3 {
			conn.Write([]byte("Invalid file download metadata format"))
			return
		} else {
			s.handleDownload(username, metadata, parts[1], conn)
		}
	}
}

func (s *Server) handleUpload(metadata string, conn net.Conn) {
	metaData, err := utils.ParseUploadMetadata(metadata)
	if err != nil {
		log.Printf("Error parsing metadata: %v", err)
		return
	}

	// Validation
	if utils.ValidateUploadMetadata(*metaData) {
	} else {
		log.Println("Invalid metadata")
		return
	}
	db.SaveUploadMetadata(metaData)

	folderName := metaData.Username
	dirPath := fmt.Sprintf("./%s/%s", ServerStorage, folderName)

	err = os.MkdirAll(dirPath, 0775)
	if err != nil {
		fmt.Println("error creating directory:", err)
	}

	filePath := fmt.Sprintf("%s/%s", dirPath, metaData.Name)
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("error creating file:", err)
	}

	s.transferData(file, metaData.FileSize, conn, "receive")

	// Verify checksum
	calculatedChecksum, err := utils.CalculateChecksum(file)
	if err != nil {
		log.Println("Error calculating checksum:", err)
		return
	}

	if calculatedChecksum == metaData.Checksum {
		conn.Write([]byte("Checksum match"))
	} else {
		conn.Write([]byte("Checksum mismatch"))

		file.Close()
		os.Remove("./" + folderName + "/" + metaData.Name)
		log.Println("Corrupted file deleted")
	}
}

func (s *Server) handleDownload(username, metadata, action string, conn net.Conn) {
	metaData, err := utils.ParseDownloadMetadata(metadata)
	if err != nil {
		logErrorAndRespond(conn, err, "Error parsing metadata")
		return
	}

	if !utils.ValidateDownloadMetadata(*metaData) {
		log.Println("Invalid metadata during validation")
		return
	}
	log.Println("Metadata validated successfully")

	// Check if file exists
	if _, err := os.Stat(metaData.Name); os.IsNotExist(err) {
		log.Printf("File not found: %s", metaData.Name)
		return
	}

	file, err := os.Open(metaData.Name)
	if err != nil {
		logErrorAndRespond(conn, err, "Error loading the requested file")
		return
	}
	defer file.Close()

	checksum, err := utils.GenerateChecksum(file)
	if err != nil {
		logErrorAndRespond(conn, err, "Error generating checksum")
		return
	}

	metadata = utils.GenerateUploadMetadata(username, action, file, checksum)
	_, err = conn.Write([]byte(metadata))
	if err != nil {
		logErrorAndRespond(conn, err, "Error sending metadata")
		return
	}

	db.SaveDownloadMetadata(metaData)

	file.Seek(0, io.SeekStart)
	s.transferData(file, 0, conn, "send")
}

func (s *Server) transferData(file *os.File, size int64, conn net.Conn, operation string) {
	var err error
	if operation == "send" {
		_, err = io.Copy(conn, file)
		if err != nil {
			if err == io.EOF {
				log.Println("File sent successfully")
			} else {
				log.Printf("Error sending data: %v", err)
			}
		}
	} else if operation == "receive" {
		_, err = io.CopyN(file, conn, size)
		if err != nil {
			log.Printf("Error receiving data: %v", err)
		} else {
			log.Println("Data receieved succesfully")
		}
	}
}

func logErrorAndRespond(conn net.Conn, err error, message string) {
	log.Println(message, err)
	conn.Write([]byte(message + ": " + err.Error()))
}
