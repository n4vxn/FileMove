package utils

import (
	"fmt"
	"log"
	"os"
)

func GenerateUploadMetadata(action string, file *os.File, checksum string) string {
	fileInfo, err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}

	data := fmt.Sprintf("%s|%s|%d|%s", action, fileInfo.Name(), fileInfo.Size(), checksum)
	return data
}

// GenerateDownloadMetadata prepares the metadata for a download operation.
func GenerateDownloadMetadata(action, filename string) string {
	data := fmt.Sprintf("%s|%s", action, filename)
	return data
}
