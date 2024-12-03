package cmd

import (
	"context"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
)

const (
	rootCmd              = `chunky`
	rootShortDescription = `Chunky is a very fast file downloader`
	rootLongDescription  = `Chunky is a high-performance tool designed for downloading large files efficiently.`
	rootExample          = `chunky download -u https://files.testfile.org/PDF/50MB-TESTFILE.ORG.pdf -d /tmp -p 4 -s 1048576`
)

func newRootCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:          rootCmd,
		Short:        rootShortDescription,
		Long:         rootLongDescription,
		Example:      rootExample,
		SilenceUsage: true,
	}

	cmd.AddCommand(newDownloadCmd())

	return cmd
}

func Execute() {

	ctx, cancel := context.WithCancel(context.Background())

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		_ = <-signalChan
		cancel()
	}()

	root := newRootCmd()

	root.ExecuteContext(ctx)
}
