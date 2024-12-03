package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// Downloader provides the functionality to download file chunks.
type Downloader interface {
	DownloadChunk(ctx context.Context, fileURL string, offset int64, size int) (io.ReadCloser, error)
}

// DefaultDownloader is the default implementation of the Downloader interface.
type DefaultDownloader struct {
	client *http.Client
}

// NewDefaultDownloader creates a new DefaultDownloader with the provided HTTP client.
// If no client is provided, http.DefaultClient is used.
func NewDefaultDownloader(client *http.Client) *DefaultDownloader {
	if client == nil {
		client = http.DefaultClient
	}
	return &DefaultDownloader{client: client}
}

// DownloadChunk fetches a specific chunk of a file from the given URL using HTTP GET with a Range header.
func (d *DefaultDownloader) DownloadChunk(ctx context.Context, fileURL string, offset int64, size int64) (io.ReadCloser, error) {
	if offset < 0 || size <= 0 {
		return nil, fmt.Errorf("invalid offset (%d) or size (%d)", offset, size)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", fileURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create GET request: %w", err)
	}

	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", offset, offset+size-1))

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download chunk: %w", err)
	}

	if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected HTTP status: %s", resp.Status)
	}

	if resp.ContentLength != size {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected content size: expected %d, got %d", size, resp.ContentLength)
	}

	return resp.Body, nil
}
