package utils

import (
	"fmt"
	"log"
	"os"
)

func GenerateUploadMetadata(username string, action string, file *os.File, checksum string) string {
	fileInfo, err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}

	data := fmt.Sprintf("%s|%s|%s|%d|%s", username, action, fileInfo.Name(), fileInfo.Size(), checksum)
	return data
}

// GenerateDownloadMetadata prepares the metadata for a download operation.
func GenerateDownloadMetadata(username, action, filename string) string {
	data := fmt.Sprintf("%s|%s|%s", username, action, filename)
	return data
}
