package listactors

import (
	"fmt"
	"strconv"

	"exercise/internal/cli/adapters/commands/cmdutil"
	"exercise/internal/cli/domain/actor"

	"github.com/spf13/cobra"
)

const defaultPageSize = 20

type Input struct {
	Limit  int32
	Offset int32
	Output string
}

type listActorsOutput struct {
	Actors      []actorItem `json:"actors"`
	TotalCount  int64       `json:"totalCount"`
	HasNext     bool        `json:"hasNext"`
	HasPrevious bool        `json:"hasPrevious"`
}

type actorItem struct {
	ID       string        `json:"id"`
	TaxID    string        `json:"taxId"`
	Type     string        `json:"type"`
	Version  int64         `json:"version"`
	Accounts []accountItem `json:"accounts"`
}

type accountItem struct {
	ID           string `json:"id"`
	DltAccountID string `json:"dltAccountId"`
	PublicKey    string `json:"publicKey"`
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
		Use:   "list",
		Short: "List actors with their accounts",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return c.Handle(cmd, in)
		},
	}
	cmd.Flags().Int32Var(&in.Limit, "limit", defaultPageSize, "page size")
	cmd.Flags().Int32Var(&in.Offset, "offset", 0, "number of actors to skip")
	cmdutil.AddOutputFlag(cmd, &in.Output)
	parent.AddCommand(cmd)
}

func (c *Command) Handle(cmd *cobra.Command, in *Input) error {
	if err := cmdutil.CheckOutput(in.Output); err != nil {
		return fmt.Errorf("list actors: %w", err)
	}
	page, err := c.Actors.FindAll(cmd.Context(), in.Limit, in.Offset)
	if err != nil {
		return fmt.Errorf("list actors: %w", err)
	}

	if in.Output == cmdutil.OutputJSON {
		if err := cmdutil.PrintJSON(cmd.OutOrStdout(), toOutput(page)); err != nil {
			return fmt.Errorf("list actors: %w", err)
		}
		return nil
	}

	rows := make([][]string, len(page.Actors))
	for i, a := range page.Actors {
		rows[i] = []string{
			a.ID.String(), a.TaxID, string(a.Type),
			strconv.FormatInt(a.Version, 10),
			strconv.Itoa(len(a.Accounts)),
		}
	}
	if err := cmdutil.PrintTable(cmd.OutOrStdout(), []string{"ID", "TAX ID", "TYPE", "VERSION", "ACCOUNTS"}, rows); err != nil {
		return fmt.Errorf("list actors: %w", err)
	}

	// stderr: piping the table (e.g. to grep) shouldn't include the footer.
	if len(page.Actors) > 0 {
		cmd.PrintErrf("\nShowing %d-%d of %d.", int(in.Offset)+1, int(in.Offset)+len(page.Actors), page.TotalCount)
		if page.HasNext {
			cmd.PrintErrf(" Next page: --offset %d", in.Offset+in.Limit)
		}
		cmd.PrintErrln()
	}
	return nil
}

func toOutput(page actor.Page) listActorsOutput {
	out := listActorsOutput{
		Actors:      make([]actorItem, len(page.Actors)),
		TotalCount:  page.TotalCount,
		HasNext:     page.HasNext,
		HasPrevious: page.HasPrevious,
	}
	for i, a := range page.Actors {
		accounts := make([]accountItem, len(a.Accounts))
		for j, acc := range a.Accounts {
			accounts[j] = accountItem{ID: acc.ID.String(), DltAccountID: acc.DltAccountID, PublicKey: acc.PublicKey}
		}
		out.Actors[i] = actorItem{
			ID:       a.ID.String(),
			TaxID:    a.TaxID,
			Type:     string(a.Type),
			Version:  a.Version,
			Accounts: accounts,
		}
	}
	return out
}
