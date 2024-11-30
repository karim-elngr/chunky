package cmd

import (
	"context"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "chunky",
	Short: "Chunky is a very fast parallel file downloader",
	Long: `Chunky is a high-performance tool designed for downloading large files efficiently.
It divides files into chunks and downloads them in parallel, enabling faster downloads.

Examples:
  chunky download --url=https://example.com/file --directory=/tmp --parallelism=4 --size=1048576`,
	SilenceUsage: true,
}

func Execute(ctx context.Context) error {

	return rootCmd.ExecuteContext(ctx)
}
