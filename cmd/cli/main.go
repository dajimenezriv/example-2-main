package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"exercise/internal/cli"

	"github.com/spf13/cobra"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err := run(ctx)
	stop()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	// Loaded before wiring (adapters need it), but its error is only reported when a command
	// runs, so --help works without any env set.
	cfg, cfgErr := cli.LoadConfig()
	app, err := setupApplication(&cfg)
	if err != nil {
		return fmt.Errorf("setup: %w", err)
	}

	root := newRootCmd(cfgErr)
	registerCommands(root, app)
	if err := root.ExecuteContext(ctx); err != nil {
		return fmt.Errorf("run command: %w", err)
	}
	return nil
}

func newRootCmd(cfgErr error) *cobra.Command {
	return &cobra.Command{
		Use:   "exercise",
		Short: "Command-line client for the Participants API",
		// Runtime errors are not usage errors: don't dump help on them.
		SilenceUsage: true,
		// run() prints the error, so setup errors and command errors look the same.
		SilenceErrors: true,
		PersistentPreRunE: func(*cobra.Command, []string) error {
			if cfgErr != nil {
				return fmt.Errorf("config: %w", cfgErr)
			}
			return nil
		},
	}
}

func registerCommands(root *cobra.Command, app *application) {
	app.LoginCmd.Register(root)

	actors := &cobra.Command{Use: "actors", Short: "Manage actors"}
	root.AddCommand(actors)
	app.ListActorsCmd.Register(actors)
	app.CreateActorCmd.Register(actors)
}
