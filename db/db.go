package db

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/n4vxn/FileMove/types"
)

var db *sql.DB

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func ConnectDB() error {
	loadEnv()
	dbAddr := os.Getenv("DB_ADDR")

	var err error
	db, err = sql.Open("postgres", dbAddr)
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		return err
	}

	fmt.Println("Connected to DB")
	return nil
}

func SaveUploadMetadata(metadata *types.UploadMetadata) error {
	query := `INSERT INTO files (filename, file_size, checksum, action) 
			  VALUES ($1, $2, $3, $4) RETURNING id`
	var id int
	err := db.QueryRow(query, &metadata.Name, &metadata.FileSize, &metadata.Checksum, &metadata.Action).Scan(&id)
	if err != nil {
		return err
	}
	fmt.Printf("File metadata saved with ID: %d\n", id)
	return err
}

func SaveUsers(user *types.User) error {
	query := `INSERT INTO users (username, password, created_at, updated_at) 
			  VALUES ($1, $2, NOW(), NOW()) RETURNING id`
	var id int
	err := db.QueryRow(query, &user.Username, &user.Password).Scan(&id)
	if err != nil {
		return err
	}
	return err
}

func RetrieveHashedPassword(username string) (string, error) {
	var hashedPassword string
	query := `SELECT password FROM users WHERE username = $1`
	err := db.QueryRow(query, &username).Scan(&hashedPassword)
	if err != nil {
		return "", err
	}

	return hashedPassword, nil
}

func HandleConnection(conn net.Conn) {
	defer conn.Close()
}
