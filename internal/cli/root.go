package cli

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

type Config struct {
	Issuer   string `env:"AUTH_ISSUER,required,notEmpty"`
	Audience string `env:"AUTH_AUDIENCE,required,notEmpty"`
	ClientID string `env:"CLI_AUTH0_CLIENT_ID,required,notEmpty"`
}

func NewRootCmd() *cobra.Command {
	cfg := &Config{}

	root := &cobra.Command{
		Use:   "exercise",
		Short: "Command-line client for the Participants API",
		// Runtime errors are not usage errors: don't dump help on them.
		SilenceUsage: true,
		PersistentPreRunE: func(*cobra.Command, []string) error {
			if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
				return fmt.Errorf("load .env: %w", err)
			}
			c, err := env.ParseAs[Config]()
			if err != nil {
				return fmt.Errorf("parse config: %w", err)
			}
			*cfg = c
			return nil
		},
	}

	root.AddCommand(newLoginCmd(cfg))

	return root
}
