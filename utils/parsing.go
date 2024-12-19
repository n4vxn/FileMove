package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/n4vxn/FileMove/types"
)

func ParseUploadMetadata(data string) (*types.UploadMetadata, error) {
	parts := strings.Split(data, "|")

	if len(parts) != 5 {
		return nil, fmt.Errorf("invalid metadata format")
	}

	size, err := strconv.Atoi(parts[3])
	if err != nil {
		return nil, fmt.Errorf("parsing error")
	}

	return &types.UploadMetadata{
		Username: parts[0],
		Action:   parts[1],
		Name:     parts[2],
		FileSize: int64(size),
		Checksum: parts[4],
	}, nil
}

func ParseDownloadMetadata(data string) (*types.DownloadMetadata, error) {
	parts := strings.Split(data, "|")

	if len(parts) != 3 {
		return nil, fmt.Errorf("parsing error")
	}

	return &types.DownloadMetadata{
		Username: parts[0],
		Action:   parts[1],
		Name:     parts[2],
	}, nil
}
