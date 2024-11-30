package downloader

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
)

// Downloader manages the HTTP client and provides file download capabilities.
type Downloader struct {
	Client *http.Client
}

// NewDownloader initializes a Downloader with the provided HTTP client.
// If the client is nil, a default client is created.
func NewDownloader(client *http.Client) *Downloader {
	if client == nil {
		client = &http.Client{}
	}
	return &Downloader{Client: client}
}

// FileDownload represents a single file download.
type FileDownload struct {
	RawURL    string
	parsedURL *url.URL
	client    *http.Client
	HeadInfo  *HeadInfo
	Chunks    []Chunk
}

// HeadInfo contains metadata about the file being downloaded.
type HeadInfo struct {
	ContentType    string
	ContentSize    int64
	FileName       string
	Signature      string
	NumberOfChunks int
	ChunkSize      int
}

// Chunk represents a portion of the file.
type Chunk struct {
	Index  int
	Offset int64
	Size   int
}

// PrepareDownload initializes a download for the given URL, retrieves metadata,
// and splits the file into chunks of the specified size.
func (d *Downloader) PrepareDownload(ctx context.Context, rawURL string, chunkSize int) (*FileDownload, error) {
	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	fd := &FileDownload{
		RawURL:    rawURL,
		parsedURL: parsedURL,
		client:    d.Client,
	}

	if err := fd.fetchHead(ctx); err != nil {
		return nil, fmt.Errorf("failed to fetch file metadata: %w", err)
	}

	chunks, err := fd.split(chunkSize)
	if err != nil {
		return nil, fmt.Errorf("failed to split file into chunks: %w", err)
	}

	fd.Chunks = chunks
	fd.HeadInfo.NumberOfChunks = len(chunks)
	fd.HeadInfo.ChunkSize = chunkSize

	return fd, nil
}

// DownloadChunk retrieves a specific chunk by its index and returns an io.ReadCloser for the caller to consume.
func (fd *FileDownload) DownloadChunk(ctx context.Context, idx int) (io.ReadCloser, error) {
	if idx < 0 || idx >= len(fd.Chunks) {
		return nil, fmt.Errorf("chunk index %d out of bounds", idx)
	}

	chunk := fd.Chunks[idx]
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fd.RawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create GET request: %w", err)
	}

	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", chunk.Offset, chunk.Offset+int64(chunk.Size)-1))
	resp, err := fd.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download chunk: %w", err)
	}

	if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected HTTP status for chunk: %s", resp.Status)
	}

	return resp.Body, nil
}

// ValidateSignature compares the file's signature with an expected value.
func (fd *FileDownload) ValidateSignature(expectedSignature string) error {
	if fd.HeadInfo == nil {
		return errors.New("file metadata not available")
	}

	if fd.HeadInfo.Signature == "" {
		return errors.New("no signature available in file metadata")
	}

	if strings.EqualFold(fd.HeadInfo.Signature, expectedSignature) {
		return fmt.Errorf("signature mismatch: expected %s, got %s", expectedSignature, fd.HeadInfo.Signature)
	}

	return nil
}

// fetchHead retrieves metadata for the file using an HTTP HEAD request.
func (fd *FileDownload) fetchHead(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, fd.RawURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create HEAD request: %w", err)
	}

	resp, err := fd.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected HTTP status: %s", resp.Status)
	}

	contentLength := resp.Header.Get("Content-Length")
	size, err := strconv.ParseInt(contentLength, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid content length: %w", err)
	}

	fileName := path.Base(fd.parsedURL.Path)
	if fileName == "" || fileName == "/" {
		fileName = "unknown"
	}

	fd.HeadInfo = &HeadInfo{
		ContentType: resp.Header.Get("Content-Type"),
		ContentSize: size,
		FileName:    fileName,
		Signature:   resp.Header.Get("ETag"),
	}

	return nil
}

// split divides the file into chunks of the specified size.
func (fd *FileDownload) split(chunkSize int) ([]Chunk, error) {
	if fd.HeadInfo == nil {
		return nil, errors.New("file metadata not initialized")
	}

	var chunks []Chunk
	var index int
	for offset := int64(0); offset < fd.HeadInfo.ContentSize; offset += int64(chunkSize) {
		size := chunkSize
		if remaining := fd.HeadInfo.ContentSize - offset; remaining < int64(chunkSize) {
			size = int(remaining)
		}
		chunks = append(chunks, Chunk{
			Index:  index,
			Offset: offset,
			Size:   size,
		})
		index++
	}

	return chunks, nil
}
