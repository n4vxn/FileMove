package types

import "time"

type UploadMetadata struct {
	Username string `json:"username"`
	Action    string    `json:"action"`
	Name      string    `json:"name"`
	FileSize  int64     `json:"file_size"`
	Checksum  string    `json:"checksum"`
	Timestamp time.Time `json:"timestamp"`
}

type DownloadMetadata struct {
	Username  string    `json:"username"`
	Action    string    `json:"action"`
	Name      string    `json:"name"`
	Timestamp time.Time `json:"timestamp"`
}

type User struct {
	Username  string `json:"username"`
	Password  string `json:"-"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
