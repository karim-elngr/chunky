package filewriter

import (
	"fmt"
	"io"
	"os"
)

// FileWriter manages the creation and chunk-based writing of a file.
type FileWriter struct {
	file   *os.File
	path   string
	size   int64
	closed bool
}

// NewFileWriter initializes a new FileWriter and creates an empty file of the specified size.
// If the file already exists, it will be truncated.
func NewFileWriter(path string, size int64) (*FileWriter, error) {
	if size <= 0 {
		return nil, fmt.Errorf("invalid file size: %d", size)
	}

	// Create or truncate the file
	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}

	// Set the file size
	if err := file.Truncate(size); err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to set file size: %w", err)
	}

	return &FileWriter{
		file: file,
		path: path,
		size: size,
	}, nil
}

// WriteChunk writes data from an io.ReadCloser to a specific offset in the file.
func (fw *FileWriter) WriteChunk(data io.ReadCloser, offset int64) error {
	defer data.Close()

	if fw.closed {
		return fmt.Errorf("file is closed")
	}

	if offset < 0 || offset >= fw.size {
		return fmt.Errorf("offset %d out of bounds", offset)
	}

	offsetWriter := io.NewOffsetWriter(fw.file, offset)
	_, err := io.Copy(offsetWriter, data)
	if err != nil {
		return fmt.Errorf("failed to write chunk at offset %d: %w", offset, err)
	}

	return nil
}

// Cleanup removes the file from disk, useful in case of an unrecoverable error.
func (fw *FileWriter) Cleanup() error {
	if fw.closed {
		return fmt.Errorf("file is already closed")
	}

	if err := os.Remove(fw.path); err != nil {
		return fmt.Errorf("failed to clean up file: %w", err)
	}

	fw.closed = true
	return nil
}

// Close closes the underlying file, releasing resources.
func (fw *FileWriter) Close() error {
	if fw.closed {
		return fmt.Errorf("file already closed")
	}

	if err := fw.file.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}

	fw.closed = true
	return nil
}
