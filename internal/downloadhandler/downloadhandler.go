package downloadhandler

import (
	"chunky/pkg/chunker"
	"chunky/pkg/downloader"
	"chunky/pkg/header"
	"chunky/pkg/md5checker"
	"chunky/pkg/workmanager"
	"chunky/pkg/writer"
	"fmt"
	"log"
	"path/filepath"

	"github.com/spf13/cobra"
)

type DownloadHandler struct {
	url         string
	directory   string
	parallelism int
	size        int64
	retries     int
}

func NewDownloadHandler(url string, directory string, parallelism int, size int64, retries int) *DownloadHandler {
	return &DownloadHandler{
		url:         url,
		directory:   directory,
		parallelism: parallelism,
		size:        size,
		retries:     retries,
	}
}

func (dh *DownloadHandler) CobraDownloadHandler() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return dh.handleInternal(cmd, args)
	}
}

func (dh *DownloadHandler) handleInternal(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	// Step 1: Fetch file metadata
	headerClient := header.NewDefaultHeader(nil)
	meta, err := headerClient.Head(ctx, dh.url)
	if err != nil {
		return fmt.Errorf("failed to fetch file metadata: %w", err)
	}
	log.Printf("File metadata: %+v", meta)

	// Step 2: Initialize chunker and calculate chunks
	chunkerClient := chunker.NewChunker(dh.size)
	chunks, err := chunkerClient.Split(meta.ContentSize)
	if err != nil {
		return fmt.Errorf("failed to calculate chunks: %w", err)
	}
	log.Printf("Total chunks to download: %d", len(chunks))

	// Step 3: Create an empty file
	filePath := filepath.Join(dh.directory, meta.FileName)
	file, err := writer.CreateFile(dh.directory, meta.FileName, meta.ContentSize)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Step 4: Initialize downloader and work manager
	downloaderClient := downloader.NewDefaultDownloader(nil)
	workManager := workmanager.NewWorkManager(ctx, dh.parallelism)
	workManager.WithRetries(dh.retries)

	// Step 5: Submit tasks to download and write chunks
	for _, chunk := range chunks {
		chunk := chunk
		err := workManager.Submit(func() error {
			respBody, err := downloaderClient.DownloadChunk(ctx, dh.url, chunk.Offset, chunk.Size)
			if err != nil {
				return fmt.Errorf("failed to download chunk at offset %d: %w", chunk.Offset, err)
			}
			defer respBody.Close()

			if err := writer.WriteChunk(file, respBody, chunk.Offset, chunk.Size); err != nil {
				return fmt.Errorf("failed to write chunk at offset %d: %w", chunk.Offset, err)
			}

			log.Printf("Chunk at offset %d written successfully", chunk.Offset)
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to submit task for chunk at offset %d: %w", chunk.Offset, err)
		}
	}

	// Step 6: Wait for all tasks to complete
	if err := workManager.Wait(); err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	log.Printf("File downloaded successfully to %s", filePath)

	// Step 7: Calculate and match the MD5 checksum of the file
	checker := md5checker.NewMD5Checker(1024 * 1024)
	actualMD5, err := checker.CalculateMD5(filePath)
	if err != nil {
		return fmt.Errorf("failed to calculate MD5 checksum: %w", err)
	}
	if actualMD5 != meta.Signature {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", meta.Signature, actualMD5)
	}

	return nil
}
