package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"exercise/internal/cli"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err := cli.NewRootCmd().ExecuteContext(ctx)
	stop()
	if err != nil {
		os.Exit(1)
	}
}
