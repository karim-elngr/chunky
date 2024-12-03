package downloadhandler

import (
	"chunky/pkg/chunker"
	"chunky/pkg/downloader"
	"chunky/pkg/header"
	"chunky/pkg/md5checker"
	"chunky/pkg/writer"
	"context"
	"fmt"
	"github.com/schollz/progressbar/v3"
	"os"
	"path/filepath"
	"strings"
)

func Handle(ctx context.Context, url string, directory string, parallelism int, size int64) (err error) {

	// Step 1: Fetch file metadata
	headerClient := header.NewDefaultHeader(nil)
	meta, err := headerClient.Head(ctx, url)
	if err != nil {
		return fmt.Errorf("failed to fetch file metadata: %w", err)
	}
	fmt.Printf("File name: %v, File size: %v, File signature: %v\n", meta.FileName, meta.ContentSize, meta.Signature)

	// Step 2: Initialize chunker and calculate chunks
	chunks, err := chunker.Split(meta.ContentSize, size)
	if err != nil {
		return fmt.Errorf("failed to calculate chunks: %w", err)
	}
	fmt.Printf("Total chunks to download: %d\n", len(chunks))

	// Step 3: Create an empty file
	filePath := filepath.Join(directory, meta.FileName)
	file, err := writer.CreateFile(directory, meta.FileName, meta.ContentSize)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			fmt.Printf("Failed to close file %s: %v\n", filePath, err)
		}
		if err == nil {
			return
		}
		fmt.Printf("Cleaning up file: %s due to error\n", filePath)
		if err := os.Remove(filePath); err != nil {
			fmt.Printf("Failed to clean up file %s: %v\n", filePath, err)
		}
	}(file)

	// Step 4: Initialize downloader and work manager
	downloaderClient := downloader.NewDefaultDownloader(nil)

	bar := progressbar.DefaultBytes(meta.ContentSize, "Downloading file...")
	defer bar.Close()

	// Step 5: Submit tasks to download and write chunks
	for _, chunk := range chunks {
		chunk := chunk
		err = func() error {
			respBody, err := downloaderClient.DownloadChunk(ctx, url, chunk.Offset, chunk.Size)
			if err != nil {
				return fmt.Errorf("failed to download chunk at offset %d: %w", chunk.Offset, err)
			}
			defer respBody.Close()

			if err := writer.WriteChunk(file, respBody, chunk.Offset, chunk.Size); err != nil {
				return fmt.Errorf("failed to write chunk at offset %d: %w", chunk.Offset, err)
			}

			bar.Add64(chunk.Size)

			return nil
		}()
		if err != nil {
			return err
		}
	}

	// Step 6: Calculate and match the MD5 checksum of the file
	actualMD5, err := md5checker.CalculateMD5(filePath, size)
	if err != nil {
		return fmt.Errorf("failed to calculate MD5 checksum: %w", err)
	}
	if strings.EqualFold(meta.Signature, actualMD5) {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", meta.Signature, actualMD5)
	}
	fmt.Printf("File downloaded successfully to: %s\n", filePath)
	fmt.Printf("Checksum matched: %s\n", actualMD5)

	return nil
}
