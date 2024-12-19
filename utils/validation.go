package utils

import (
	"log"

	"github.com/n4vxn/FileMove/types"
)

func ValidateUploadMetadata(metaData types.UploadMetadata) bool {
	if metaData.Username == "" || metaData.Action == "" || metaData.FileSize == 0 || metaData.Name == "" || metaData.Checksum == "" {
		log.Println("validation error: invalid metadata")
		return false
	}
	return true
}

func ValidateDownloadMetadata(metaData types.DownloadMetadata) bool {
	if metaData.Username == "" || metaData.Action == "" || metaData.Name == "" {
		log.Println("validation error: invalid metadata")
		return false
	}
	return true
}
