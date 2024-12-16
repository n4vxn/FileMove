package utils

import (
	"fmt"
	"strconv"
	"strings"
	"github.com/n4vxn/FileMove/types"
)

func ParseUploadMetadata(data string) (*types.UploadMetadata, error) {
	parts := strings.Split(data, "|")

	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid metadata format")
	}

	size, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("error parsing file size: %v", err)
	}

	return &types.UploadMetadata{
		Action:   parts[0],
		Name:     parts[1],
		FileSize: int64(size),
		Checksum: parts[3],
	}, nil
}

func ParseDownloadMetadata(data string) (*types.DownloadMetadata, error) {
	parts := strings.Split(data, "|")

	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid metadata format")
	}

	return &types.DownloadMetadata{
		Action: parts[0],
		Name:   parts[1],
	}, nil
}
