package header

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
)

// FileMeta contains metadata about the file.
type FileMeta struct {
	ContentType string
	ContentSize int64
	FileName    string
	Signature   string
}

// Header is the interface for retrieving file metadata.
type Header interface {
	Head(ctx context.Context, fileURL string) (*FileMeta, error)
}

// DefaultHeader is the default implementation of the Header interface.
type DefaultHeader struct {
	client *http.Client
}

// NewDefaultHeader creates a new DefaultHeader with the provided HTTP client.
// If no client is provided, http.DefaultClient is used.
func NewDefaultHeader(client *http.Client) Header {
	if client == nil {
		client = http.DefaultClient
	}
	return &DefaultHeader{client: client}
}

// Head retrieves metadata about the file from the given URL using an HTTP HEAD request.
func (h *DefaultHeader) Head(ctx context.Context, fileURL string) (*FileMeta, error) {
	parsedURL, err := url.ParseRequestURI(fileURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	fileName := path.Base(parsedURL.Path)
	if fileName == "" || fileName == "/" {
		return nil, errors.New("missing file name in URL")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, fileURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HEAD request: %w", err)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status: %s", resp.Status)
	}

	contentLength := resp.Header.Get("Content-Length")
	size, err := strconv.ParseInt(contentLength, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid content length: %w", err)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		return nil, fmt.Errorf("missing Content-Type header in response")
	}

	signature := resp.Header.Get("ETag")
	if signature == "" {
		return nil, fmt.Errorf("missing ETag header in response")
	}

	acceptRanges := resp.Header.Get("Accept-Ranges")
	if acceptRanges != "bytes" {
		return nil, fmt.Errorf("server does not support byte ranges")
	}

	return &FileMeta{
		ContentType: contentType,
		ContentSize: size,
		FileName:    fileName,
		Signature:   signature,
	}, nil
}
