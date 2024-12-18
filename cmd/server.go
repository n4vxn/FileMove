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
		} else {
			log.Printf("Server started with username: %s\n", currentUser)
		}

		contextualConn := ContextualConn{
			Conn:     conn,
			ctx:      s.ctx,
			Username: currentUser,
		}

		sem <- struct{}{} // Bloc if max connections reached
		go func(conn net.Conn) {
			defer func() { <-sem }() // Release semaphore slots
			s.handleRequests(&contextualConn)
		}(contextualConn)
	}
}

func (s *Server) handleRequests(conn *ContextualConn) {
	defer conn.Close()
	metaDataBuffer := make([]byte, 2048)

	log.Printf("%s connected!\n", conn.Username)

	// Receive metadata
	n, err := conn.Read(metaDataBuffer)
	if err != nil {
		log.Printf("Error reading metadata: %v", err)
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

	} else if parts[1] == "DOWNLOAD" {
		if len(parts) != 2 {
			conn.Write([]byte("Invalid file download metadata format"))
			return
		} else {
			s.handleDownload(s.Username, metadata, parts[0], conn)
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
		log.Println("Metadata received and valid")
	} else {
		log.Println("Invalid metadata")
		return
	}

	folderName := metaData.Username

	err = os.MkdirAll(folderName, 0775)
	if err != nil {
		fmt.Println("error creating directory:", err)
	}

	file, err := os.Create("./" + folderName + "/" + metaData.Name)
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

func (s *Server) handleDownload(username, filename, action string, conn net.Conn) {
	metaData, err := utils.ParseDownloadMetadata(filename)
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
		conn.Write([]byte("File not found"))
		return
	}

	folderContent, err := os.ReadDir("navee")
	if err != nil {
		log.Println("No dir")
	}

	for _, v := range folderContent {
		log.Println(v.Info())
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

	metadata := utils.GenerateUploadMetadata(username, action, file, checksum)
	_, err = conn.Write([]byte(metadata))
	if err != nil {
		logErrorAndRespond(conn, err, "Error sending metadata")
		return
	}

	file.Seek(0, io.SeekStart)
	// s.transferData(file,  conn, "send")
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
