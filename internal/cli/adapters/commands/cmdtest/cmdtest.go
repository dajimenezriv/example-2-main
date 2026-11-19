package cmdtest

import (
	"bytes"
	"context"
	"testing"

	"github.com/spf13/cobra"
)

type Result struct {
	Stdout string
	Stderr string
	Err    error
}

// Run registers one command under a bare root (no config, no real wiring) and executes it
// with args, e.g. Run(ctx, t, cmd.Register, "list", "--limit", "2").
func Run(ctx context.Context, t *testing.T, register func(parent *cobra.Command), args ...string) Result {
	t.Helper()

	root := &cobra.Command{Use: "exercise", SilenceUsage: true, SilenceErrors: true}
	register(root)

	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs(args)

	err := root.ExecuteContext(ctx)
	return Result{Stdout: stdout.String(), Stderr: stderr.String(), Err: err}
}
