package types

import "time"

type UploadMetadata struct {
	Action    string    `json:"action"`
	Name      string    `json:"name"`
	FileSize  int64     `json:"file_size"`
	Checksum  string    `json:"checksum"`
	Timestamp time.Time `json:"timestamp"`
}

type DownloadMetadata struct {
	Action    string    `json:"action"`
	Name      string    `json:"name"`
	Timestamp time.Time `json:"timestamp"`
}
