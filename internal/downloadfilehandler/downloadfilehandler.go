package downloadfilehandler

import (
	"chunky/pkg/downloader"
	"chunky/pkg/filewriter"
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// FileDownloadHandler orchestrates the file download process using the downloader and filewriter packages.
type FileDownloadHandler struct {
	Downloader  *downloader.Downloader
	Parallelism int
}

// NewFileDownloadHandler initializes a FileDownloadHandler with the necessary dependencies.
func NewFileDownloadHandler(downloaderClient *downloader.Downloader, parallelism int) *FileDownloadHandler {
	return &FileDownloadHandler{
		Downloader:  downloaderClient,
		Parallelism: parallelism,
	}
}

// Download orchestrates the entire download process.
func (fdh *FileDownloadHandler) Download(ctx context.Context, rawURL string, filePath string, chunkSize int) (string, error) {
	df, err := fdh.Downloader.PrepareDownload(ctx, rawURL, chunkSize)
	if err != nil {
		return "", fmt.Errorf("failed to prepare download: %w", err)
	}
	log.Printf("File metadata retrieved: %+v", df.HeadInfo)

	fullPath := filepath.Join(filePath, df.HeadInfo.FileName)
	fileWriter, err := filewriter.NewFileWriter(fullPath, df.HeadInfo.ContentSize)
	if err != nil {
		return "", fmt.Errorf("failed to initialize file writer: %w", err)
	}
	defer fileWriter.Close()

	taskChan := make(chan int, df.HeadInfo.NumberOfChunks)
	errChan := make(chan error, 1)
	var wg sync.WaitGroup
	var once sync.Once

	for i := 0; i < fdh.Parallelism; i++ {
		fdh.launchWorker(ctx, &wg, taskChan, df, &once, errChan, fileWriter, i)
	}

	fdh.distributeWork(ctx, df, taskChan)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
		close(errChan)
	}()

	select {
	case <-done:
		md5Hash, err := generateMd5(fullPath)
		if err != nil {
			return "", fmt.Errorf("failed to generate md5 hash: %w", err)
		}
		md5Err := df.ValidateSignature(md5Hash)
		if md5Err != nil {
			return "", md5Err
		}
		return fullPath, nil
	case err := <-errChan:
		fileWriter.Cleanup()
		return "", fmt.Errorf("file download failed: %w", err)
	case <-ctx.Done():
		fileWriter.Cleanup()
		return "", fmt.Errorf("download canceled: %w", ctx.Err())
	}
}

func (fdh *FileDownloadHandler) launchWorker(ctx context.Context, wg *sync.WaitGroup, taskChan chan int, df *downloader.FileDownload, once *sync.Once, errChan chan error, fileWriter *filewriter.FileWriter, i int) {
	wg.Add(1)
	go func(workerID int) {
		defer wg.Done()
		for chunk := range taskChan {
			err := fdh.processChunk(ctx, df, fileWriter, df.Chunks[chunk], workerID)
			if err != nil {
				once.Do(func() {
					errChan <- err
				})
				return
			}
		}
	}(i)
}

func (fdh *FileDownloadHandler) distributeWork(ctx context.Context, df *downloader.FileDownload, taskChan chan int) {
	go func() {
		for idx := 0; idx < df.HeadInfo.NumberOfChunks; idx++ {
			select {
			case taskChan <- idx:
			case <-ctx.Done():
				close(taskChan)
				return
			}
		}
		close(taskChan)
	}()
}

func (fdh *FileDownloadHandler) processChunk(ctx context.Context, dl *downloader.FileDownload, fileWriter *filewriter.FileWriter, chunk downloader.Chunk, workerID int) error {
	log.Printf("Worker %d: Processing chunk %d at offset %d", workerID, chunk.Index, chunk.Offset)
	respBody, err := dl.DownloadChunk(ctx, chunk.Index)
	if err != nil {
		return fmt.Errorf("worker %d: failed to download chunk %d: %w", workerID, chunk.Index, err)
	}
	defer respBody.Close()

	if err := fileWriter.WriteChunk(respBody, chunk.Offset); err != nil {
		return fmt.Errorf("worker %d: failed to write chunk %d: %w", workerID, chunk.Index, err)
	}

	log.Printf("Worker %d: Successfully processed chunk %d", workerID, chunk.Index)
	return nil
}

func generateMd5(fileName string) (string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return "", nil
	}
	defer file.Close()
	var r io.Reader = file

	w := md5.New()

	if _, err := io.Copy(w, r); err != nil {
		return "", err
	}

	sig := w.Sum(nil)
	return fmt.Sprintf("%x", sig), nil
}
