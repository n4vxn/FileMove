package utils

import (
	"fmt"

	"github.com/n4vxn/FileMove/types"
)

func ValidateUploadMetadata(metaData types.UploadMetadata) bool {
	if metaData.Username == "" || metaData.Action == "" || metaData.FileSize == 0 || metaData.Name == "" || metaData.Checksum == "" {
		fmt.Println("validation error: invalid metadata")
		return false
	}
	fmt.Printf("%s|%s|%s|%d|%s", metaData.Username, metaData.Action, metaData.Name, metaData.FileSize, metaData.Checksum)
	return true
}

func ValidateDownloadMetadata(metaData types.DownloadMetadata) bool {
	if metaData.Action == "" || metaData.Name == "" {
		fmt.Println("validation error: invalid metadata")
		return false
	}
	fmt.Printf("%s|%s", metaData.Action, metaData.Name)
	return true
}
