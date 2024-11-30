package writer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CreateFile initializes an empty file with the specified size in the given directory.
// It ensures the directory exists and creates the file with the specified name.
func CreateFile(directory, fileName string, size int64) (*os.File, error) {
	if size <= 0 {
		return nil, fmt.Errorf("invalid file size: %d", size)
	}

	if err := os.MkdirAll(directory, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", directory, err)
	}

	filePath := filepath.Join(directory, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file %s: %w", filePath, err)
	}

	if err := file.Truncate(size); err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to set file size for %s: %w", filePath, err)
	}

	return file, nil
}

// WriteChunk writes data from an io.Reader to a specific offset in the file.
// It validates that the data matches the expected size and ensures all data is written.
func WriteChunk(file *os.File, data io.Reader, offset int64, size int64) error {
	if offset < 0 {
		return fmt.Errorf("invalid offset: %d", offset)
	}
	if size <= 0 {
		return fmt.Errorf("invalid size: %d", size)
	}

	offsetWriter := io.NewOffsetWriter(file, offset)
	written, err := io.CopyN(offsetWriter, data, size)
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to write chunk at offset %d: %w", offset, err)
	}

	if written != size {
		return fmt.Errorf("data size mismatch: expected %d bytes, written %d bytes", size, written)
	}

	// Ensure no extra data remains in the reader
	remaining := make([]byte, 1)
	if _, err := data.Read(remaining); err != io.EOF {
		return fmt.Errorf("unexpected data remaining after writing chunk: %w", err)
	}

	return nil
}
