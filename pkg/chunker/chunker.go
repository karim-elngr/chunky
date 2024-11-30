package chunker

import "fmt"

type Chunker interface {
	Split(fileSize int64) ([]Chunk, error)
}

type defaultChunker struct {
	chunkSize int64
}

func NewChunker(chunkSize int64) Chunker {
	return &defaultChunker{
		chunkSize: chunkSize,
	}
}

// Chunk represents a file chunk with an offset and size.
type Chunk struct {
	Index  int
	Offset int64
	Size   int64
}

// Split divides the file into chunks based on the file size and chunk size.
func (c *defaultChunker) Split(fileSize int64) ([]Chunk, error) {
	if fileSize <= 0 {
		return nil, fmt.Errorf("invalid file size: %d", fileSize)
	}
	if c.chunkSize <= 0 {
		return nil, fmt.Errorf("invalid chunk size: %d", c.chunkSize)
	}

	var chunks []Chunk
	var index int
	var offset int64
	for offset < fileSize {
		size := c.chunkSize
		if remaining := fileSize - offset; remaining < c.chunkSize {
			size = remaining
		}
		chunks = append(chunks, Chunk{
			Index:  index,
			Offset: offset,
			Size:   size,
		})
		offset += size
		index++
	}
	return chunks, nil
}
