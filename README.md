# FileMove - TCP File Server

FileMove is a Go-based secure file transfer system designed to handle file uploads and downloads with features like TLS encryption, checksum validation, and PostgreSQL integration.

## Features

- Secure communication using **TLS**.
- **File upload** and **download** functionality.
- Metadata handling for file transfers.
- **Checksum validation** to ensure data integrity.
- PostgreSQL integration for saving metadata.
- Docker setup for easy deployment.
- **Database migrations** using `go-migration`.

## Prerequisites

- Go 1.18 or higher
- Docker and Docker Compose
- PostgreSQL database
- TLS certificates (server and client)

## Project Structure

```plaintext
FileMove/
├── main.go               # Main application entry point
├── server.go             # Server implementation
├── client.go             # Client implementation
├── utils/                
│   ├── checksum.go       # Checksum generation and validation
│   └── metadata.go       # Metadata generation and parsing
├── db/                   
│   ├── database.go       # Database operations
│   └── migrations/       # Folder for migration files
├── tls/                  
│   ├── server.crt        # Server certificate
│   └── server.key        # Server private key
├── Dockerfile            # Dockerfile for the application
├── docker-compose.yml    # Docker Compose configuration
├── README.md             # Project documentation
└── go.mod                # Go module file
