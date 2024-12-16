package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path"
	"strings"
	"github.com/n4vxn/FileMove/utils"
)

const (
	HOST    = "localhost"
	PORT    = "8080"
	TYPE    = "tcp"
	maxConn = 10
)

type Config struct {
	HOST string
	PORT string
}

type Server struct {
	listener net.Listener
	Config
}

func NewServer(cfg Config) *Server {
	if len(cfg.PORT) == 0 {
		cfg.PORT = PORT
	}
	return &Server{
		Config: cfg,
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
		}
		sem <- struct{}{} // Bloc if max connections reached
		go func(conn net.Conn) {
			defer func() { <-sem }() // Release semaphore slots
			s.handleRequests(conn)
		}(conn)
	}
}

func (s *Server) handleRequests(conn net.Conn) {
	defer conn.Close()
	metaDataBuffer := make([]byte, 2048)

	fmt.Println("Received a request:", conn.RemoteAddr().String())
	// Receive metadata

	n, err := conn.Read(metaDataBuffer)
	if err != nil {
		log.Printf("Error reading metadata: %v", err)
		return
	}

	metadata := string(metaDataBuffer[:n])
	metaDataBuffer = nil

	parts := strings.Split(strings.TrimSpace(metadata), "|")

	if parts[0] == "UPLOAD" {
		if len(parts) != 4 {
			conn.Write([]byte("Invalid file upload metadata format"))
			return
		} else {
			s.handleUpload(metadata, conn)
		}

	} else if parts[0] == "DOWNLOAD" {
		if len(parts) != 2 {
			conn.Write([]byte("Invalid file download metadata format"))
			return
		} else {
			s.handleDownload(metadata, parts[0], conn)
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

	ext := path.Ext(metaData.Name)
	folderName := strings.TrimSuffix(metaData.Name, ext)

	err = os.Mkdir(folderName, 0775)
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

func (s *Server) handleDownload(filename, action string, conn net.Conn) {
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

	metadata := utils.GenerateUploadMetadata(action, file, checksum)
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
