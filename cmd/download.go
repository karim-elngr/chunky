package cmd

import (
	"chunky/internal/downloadhandler"
	"github.com/spf13/cobra"
)

var (
	url         string
	directory   string
	parallelism int
	size        int64
	retries     int
)

func init() {
	downloadCmd.Flags().StringVarP(&url, "url", "u", "", "URL of the file to download (required)")
	downloadCmd.MarkFlagRequired("url")
	downloadCmd.Flags().StringVarP(&directory, "directory", "d", ".", "Target directory to save the downloaded file")
	downloadCmd.Flags().IntVarP(&parallelism, "parallelism", "p", 4, "Number of parallel chunks to download")
	downloadCmd.Flags().Int64VarP(&size, "size", "s", 1024*1024, "Size of each download chunk in bytes (default: 1MB)")
	downloadCmd.Flags().IntVarP(&retries, "retries", "r", 0, "Number of retries to download a chunk (default: 0)")

	rootCmd.AddCommand(downloadCmd)
}

var downloadCmd = newDownloadCommand()

func newDownloadCommand() *cobra.Command {

	return &cobra.Command{
		Use:   "download",
		Short: "Download a file in chunks",
		RunE:  cobraDownloadHandler,
	}
}

func cobraDownloadHandler(cmd *cobra.Command, args []string) error {
	dh := downloadhandler.NewDownloadHandler(url, directory, parallelism, size, retries)
	return dh.CobraDownloadHandler()(cmd, args)
}
