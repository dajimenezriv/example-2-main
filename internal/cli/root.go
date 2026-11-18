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
	// Issuer is the Auth0 tenant URL, the same value the API verifies tokens against.
	Issuer string `env:"AUTH_ISSUER,required,notEmpty"`
	// Audience is the API identifier; tokens are issued for it.
	Audience string `env:"AUTH_AUDIENCE,required,notEmpty"`
	// ClientID of the Auth0 Native app. Public client: there is no secret.
	ClientID string `env:"CLI_AUTH0_CLIENT_ID,required,notEmpty"`
}

// NewRootCmd builds the command tree. Config is loaded in PersistentPreRunE (before any
// subcommand runs), so --help and invalid commands work without any env set.
func NewRootCmd() *cobra.Command {
	cfg := &Config{}

	root := &cobra.Command{
		Use:   "exercise",
		Short: "Command-line client for the Participants API",
		// Runtime errors are not usage errors: don't dump help on them.
		SilenceUsage: true,
		PersistentPreRunE: func(*cobra.Command, []string) error {
			c, err := loadConfig()
			if err != nil {
				return err
			}
			*cfg = c
			return nil
		},
	}

	root.AddCommand(newLoginCmd(cfg))

	return root
}

func loadConfig() (Config, error) {
	// Dev convenience: reuse the repo's .env (filled by `task auth:env`). Real env vars win.
	if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return Config{}, fmt.Errorf("load .env: %w", err)
	}
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}
