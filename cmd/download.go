package cmd

import (
	"chunky/internal/downloadhandler"
	"fmt"
	"github.com/spf13/cobra"
)

const (
	url                = "url"
	shortUrl           = "u"
	directory          = "directory"
	shortDirectory     = "d"
	defaultDirectory   = "."
	parallelism        = "parallelism"
	shortParallelism   = "p"
	defaultParallelism = 4
	size               = "size"
	shortSize          = "s"
	defaultSize        = 1024 * 1024

	downloadCmd              = `download`
	downloadShortDescription = `Download a file in chunks`
	downloadLongDescription  = `Download a file in chunks using multiple goroutines to speed up the process.`
	downloadExample          = `chunky download -u https://files.testfile.org/PDF/50MB-TESTFILE.ORG.pdf -d /tmp -p 4 -s 1048576`
)

func newDownloadCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:          downloadCmd,
		Short:        downloadShortDescription,
		Long:         downloadLongDescription,
		Example:      downloadExample,
		RunE:         run,
		SilenceUsage: true,
	}

	cmd.Flags().StringP(url, shortUrl, "", "URL of the file to download (required)")
	cmd.MarkFlagRequired(url)
	cmd.Flags().StringP(directory, shortDirectory, defaultDirectory, "Target directory to save the downloaded file")
	cmd.Flags().IntP(parallelism, shortParallelism, defaultParallelism, "Number of parallel chunks to download")
	cmd.Flags().Int64P(size, shortSize, defaultSize, "Size of each download chunk in bytes (default: 1MB)")

	return cmd
}

func run(cmd *cobra.Command, _ []string) error {

	urlValue, err := cmd.Flags().GetString(url)
	if err != nil {
		return fmt.Errorf("url flag not found: %w", err)
	}

	directoryValue, err := cmd.Flags().GetString(directory)
	if err != nil {
		return fmt.Errorf("directory flag not found: %w", err)
	}

	parallelismValue, err := cmd.Flags().GetInt(parallelism)
	if err != nil {
		return fmt.Errorf("parallelism flag not found: %w", err)
	}

	sizeValue, err := cmd.Flags().GetInt64(size)
	if err != nil {
		return fmt.Errorf("size flag not found: %w", err)
	}

	err = downloadhandler.Handle(cmd.Context(), urlValue, directoryValue, parallelismValue, sizeValue)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	return nil
}
