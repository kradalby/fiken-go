package cli

import (
	"context"
	"io"

	"github.com/peterbourgon/ff/v4"

	"github.com/kradalby/fiken-go/ops"
)

// AddBankBalances wires `fiken bank-balances list`. The upstream tag
// exposes a single GET so there is no get / mutate subcommand.
func AddBankBalances(root *ff.Command, stdout, stderr io.Writer, sf *sessionFactory) error {
	_ = stderr
	set := ff.NewFlagSet("bank-balances")
	if parentFS, ok := root.Flags.(*ff.FlagSet); ok {
		set.SetParent(parentFS)
	}

	cmd := &ff.Command{
		Name:      "bank-balances",
		Usage:     "fiken bank-balances <subcommand>",
		ShortHelp: "Inspect bank balances.",
		Flags:     set,
	}

	listSet := ff.NewFlagSet("list")
	listSet.SetParent(set)
	var (
		listDate     string
		listPage     int
		listPageSize int
	)
	listSet.StringVar(&listDate, 0, "date", "", "Balance date YYYY-MM-DD (optional)")
	listSet.IntVar(&listPage, 0, "page", 0, "Page number (0-indexed)")
	listSet.IntVar(&listPageSize, 0, "page-size", 0, "Page size")
	listCmd := &ff.Command{
		Name:      "list",
		Usage:     "fiken bank-balances list --company <slug>",
		ShortHelp: "List bank balances for a company.",
		Flags:     listSet,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).BankBalancesList(ctx, ops.BankBalancesListIn{
				Company:  *sf.flagCompany,
				Date:     ops.Date(listDate),
				Page:     listPage,
				PageSize: listPageSize,
			})
			return Renderer(ctx).Render(res)
		},
	}

	cmd.Subcommands = []*ff.Command{listCmd}
	root.Subcommands = append(root.Subcommands, cmd)
	return nil
}
