package types

type UploadMetadata struct {
	Username string `json:"username"`
	Action   string `json:"action"`
	Name     string `json:"name"`
	FileSize int64  `json:"file_size"`
	Checksum string `json:"checksum"`
}

type DownloadMetadata struct {
	Username string `json:"username"`
	Action   string `json:"action"`
	Name     string `json:"name"`
}

type User struct {
	Username  string `json:"username"`
	Password  string `json:"-"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
