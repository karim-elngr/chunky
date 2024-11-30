package cmd

import (
	"chunky/internal/downloadfilehandler"
	"chunky/pkg/downloader"
	"fmt"
	"github.com/spf13/cobra"
	"log"
)

var downloaderFlags = struct {
	url         string
	directory   string
	parallelism int
	size        int
}{}

func init() {
	downloadCmd.Flags().StringVarP(&downloaderFlags.url, "url", "u", "", "URL of the file to download (required)")
	downloadCmd.MarkFlagRequired("url")
	downloadCmd.Flags().StringVarP(&downloaderFlags.directory, "directory", "d", ".", "Target directory to save the downloaded file")
	downloadCmd.Flags().IntVarP(&downloaderFlags.parallelism, "parallelism", "p", 4, "Number of parallel chunks to download")
	downloadCmd.Flags().IntVarP(&downloaderFlags.size, "size", "s", 1024*1024, "Size of each download chunk in bytes (default: 1MB)")

	rootCmd.AddCommand(downloadCmd)
}

var downloadCmd = newDownloadCommand()

func newDownloadCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "download",
		Short: "Download a file in chunks",
		RunE:  downloadHandler,
	}
}

var downloadHandler = func(cmd *cobra.Command, args []string) error {
	log.Printf("Starting download: URL=%s, Directory=%s, Parallelism=%d, ChunkSize=%d",
		downloaderFlags.url, downloaderFlags.directory, downloaderFlags.parallelism, downloaderFlags.size)

	dl := downloader.NewDownloader(nil)

	handler := downloadfilehandler.NewFileDownloadHandler(dl, downloaderFlags.parallelism)

	ctx := cmd.Context()
	log.Printf("Executing download for URL: %s", downloaderFlags.url)

	path, err := handler.Download(ctx, downloaderFlags.url, downloaderFlags.directory, downloaderFlags.size)
	if err != nil {
		return fmt.Errorf("failed to download file from URL '%s': %w", downloaderFlags.url, err)
	}

	log.Printf("Download completed successfully. File saved in: %s", path)
	return nil
}
