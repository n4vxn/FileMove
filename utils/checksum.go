package utils

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

func GenerateChecksum(file *os.File) (string, error) {
	hasher := sha256.New()

	_, err := file.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}

	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

func CalculateChecksum(file *os.File) (string, error) {
	hasher := sha256.New()

	_, err := file.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}

	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}
