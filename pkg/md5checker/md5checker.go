package md5checker

import (
	"crypto/md5"
	"fmt"
	"hash"
	"io"
	"os"
)

// MD5Checker calculates the MD5 checksum of a file.
type MD5Checker struct {
	chunkSize int // Size of the chunks to read from the file
}

// NewMD5Checker creates a new MD5Checker with the specified chunk size.
// If chunkSize is <= 0, a default chunk size of 1MB is used.
func NewMD5Checker(chunkSize int) *MD5Checker {
	if chunkSize <= 0 {
		chunkSize = 1024 * 1024 // Default to 1MB
	}
	return &MD5Checker{chunkSize: chunkSize}
}

// CalculateMD5 calculates the MD5 checksum of the file at the given path.
// It reads the file in chunks to minimize memory usage for large files.
func (mc *MD5Checker) CalculateMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	hash := md5.New()
	if err := mc.calculateHash(file, hash); err != nil {
		return "", fmt.Errorf("failed to calculate MD5: %w", err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// calculateHash reads the file in chunks and updates the hash incrementally.
func (mc *MD5Checker) calculateHash(file *os.File, hash hash.Hash) error {
	buf := make([]byte, mc.chunkSize)
	for {
		n, err := file.Read(buf)
		if n > 0 {
			if _, hashErr := hash.Write(buf[:n]); hashErr != nil {
				return fmt.Errorf("failed to update hash: %w", hashErr)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
	}
	return nil
}
