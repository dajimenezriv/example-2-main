package login

import (
	"fmt"

	"exercise/internal/cli/domain/auth"

	"github.com/spf13/cobra"
)

type Input struct{}

type Command struct {
	Auth   auth.Authenticator
	Tokens auth.TokenStore
}

func NewCommand(authenticator auth.Authenticator, tokens auth.TokenStore) *Command {
	return &Command{Auth: authenticator, Tokens: tokens}
}

func (c *Command) Register(parent *cobra.Command) {
	in := &Input{}
	parent.AddCommand(&cobra.Command{
		Use:   "login",
		Short: "Log in through the browser (OAuth2 Device Authorization flow)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return c.Handle(cmd, in)
		},
	})
}

func (c *Command) Handle(cmd *cobra.Command, _ *Input) error {
	// stderr: stdout is reserved for data output.
	tok, err := c.Auth.DeviceLogin(cmd.Context(), func(dc auth.DeviceCode) {
		cmd.PrintErrf("Open this URL in a browser:\n\n  %s\n\n", dc.VerificationURL)
		if dc.CodePrefilled {
			cmd.PrintErrf("Confirm the code shown there is: %s\n\n", dc.UserCode)
		} else {
			cmd.PrintErrf("Enter this code there: %s\n\n", dc.UserCode)
		}
		cmd.PrintErrln("Waiting for approval (Ctrl+C to cancel)...")
	})
	if err != nil {
		return fmt.Errorf("login: %w", err)
	}
	if err := c.Tokens.Save(tok); err != nil {
		return fmt.Errorf("login: save token: %w", err)
	}

	cmd.PrintErrf("Logged in. Tokens saved to %s\n", c.Tokens.Path())
	if tok.RefreshToken == "" {
		cmd.PrintErrln("Warning: no refresh token issued; you will need to log in again when the access token expires.")
	}
	return nil
}
