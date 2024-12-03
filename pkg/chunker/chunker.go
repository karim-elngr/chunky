package chunker

import "fmt"

// Chunk represents a file chunk with an offset and size.
type Chunk struct {
	Offset int64
	Size   int64
}

// Split divides the file into chunks based on the file size and chunk size.
func Split(fileSize int64, chunkSize int64) ([]Chunk, error) {
	if fileSize <= 0 {
		return nil, fmt.Errorf("invalid file size: %d", fileSize)
	}
	if chunkSize <= 0 {
		return nil, fmt.Errorf("invalid chunk size: %d", chunkSize)
	}

	var chunks []Chunk
	var index int
	var offset int64
	for offset < fileSize {
		size := chunkSize
		if remaining := fileSize - offset; remaining < chunkSize {
			size = remaining
		}
		chunks = append(chunks, Chunk{
			Offset: offset,
			Size:   size,
		})
		offset += size
		index++
	}
	return chunks, nil
}
