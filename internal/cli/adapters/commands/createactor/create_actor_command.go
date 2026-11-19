package createactor

import (
	"fmt"
	"strconv"
	"strings"

	"exercise/internal/cli/adapters/commands/cmdutil"
	"exercise/internal/cli/domain/actor"

	"github.com/spf13/cobra"
)

type Input struct {
	TaxID  string
	Type   string
	Output string
}

type createActorOutput struct {
	ID      string `json:"id"`
	TaxID   string `json:"taxId"`
	Type    string `json:"type"`
	Version int64  `json:"version"`
}

type Command struct {
	Actors actor.Repository
}

func NewCommand(actors actor.Repository) *Command {
	return &Command{Actors: actors}
}

func (c *Command) Register(parent *cobra.Command) {
	in := &Input{}
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create an actor",
		Example: "  exercise actors create --tax-id 123456789 --type issuer",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return c.Handle(cmd, in)
		},
	}
	types := actorTypes()
	cmd.Flags().StringVar(&in.TaxID, "tax-id", "", "tax ID, 9 digits")
	cmd.Flags().StringVar(&in.Type, "type", "", "one of: "+strings.Join(types, ", "))
	cmdutil.MarkRequired(cmd, "tax-id", "type")
	cmdutil.RegisterFlagValues(cmd, "type", types)
	cmdutil.AddOutputFlag(cmd, &in.Output)
	parent.AddCommand(cmd)
}

func (c *Command) Handle(cmd *cobra.Command, in *Input) error {
	if err := cmdutil.CheckOutput(in.Output); err != nil {
		return fmt.Errorf("create actor: %w", err)
	}
	// No local validation of the actor: the API is the source of truth and its 422 lists every invalid field.
	a, err := c.Actors.Create(cmd.Context(), in.TaxID, actor.Type(in.Type))
	if err != nil {
		return fmt.Errorf("create actor: %w", err)
	}

	out := createActorOutput{ID: a.ID.String(), TaxID: a.TaxID, Type: string(a.Type), Version: a.Version}
	if in.Output == cmdutil.OutputJSON {
		err = cmdutil.PrintJSON(cmd.OutOrStdout(), out)
	} else {
		err = cmdutil.PrintTable(cmd.OutOrStdout(),
			[]string{"ID", "TAX ID", "TYPE", "VERSION"},
			[][]string{{out.ID, out.TaxID, out.Type, strconv.FormatInt(out.Version, 10)}})
	}
	if err != nil {
		return fmt.Errorf("create actor: %w", err)
	}
	return nil
}

func actorTypes() []string {
	types := actor.Types()
	out := make([]string, len(types))
	for i, t := range types {
		out[i] = string(t)
	}
	return out
}
