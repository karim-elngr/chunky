package md5checker

import (
	"crypto/md5"
	"fmt"
	"hash"
	"io"
	"os"
)

// CalculateMD5 calculates the MD5 checksum of the file at the given path.
// It reads the file in chunks to minimize memory usage for large files.
func CalculateMD5(filePath string, chunkSize int64) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	h := md5.New()
	if err := calculateHash(file, h, chunkSize); err != nil {
		return "", fmt.Errorf("failed to calculate MD5: %w", err)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// calculateHash reads the file in chunks and updates the hash incrementally.
func calculateHash(file *os.File, h hash.Hash, chunkSize int64) error {
	buf := make([]byte, chunkSize)
	for {
		n, err := file.Read(buf)
		if n > 0 {
			if _, hashErr := h.Write(buf[:n]); hashErr != nil {
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
