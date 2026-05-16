package cli

import (
	"context"
	"io"

	"github.com/peterbourgon/ff/v4"

	"github.com/kradalby/fiken-go/ops"
)

// AddAccountBalances wires `fiken account-balances {list,get}`. The
// upstream tag exposes only GETs, so there are no mutating stubs.
//
// The subcommand path uses a hyphen because ff matches subcommands
// verbatim and "account balances" is two words upstream; "account-
// balances" keeps the CLI single-token.
func AddAccountBalances(root *ff.Command, stdout, stderr io.Writer, sf *sessionFactory) error {
	_ = stderr
	set := ff.NewFlagSet("account-balances")
	if parentFS, ok := root.Flags.(*ff.FlagSet); ok {
		set.SetParent(parentFS)
	}

	cmd := &ff.Command{
		Name:      "account-balances",
		Usage:     "fiken account-balances <subcommand>",
		ShortHelp: "Inspect bookkeeping account closing balances.",
		Flags:     set,
	}

	// list
	listSet := ff.NewFlagSet("list")
	listSet.SetParent(set)
	var (
		listDate        string
		listPage        int
		listPageSize    int
		listFromAccount int64
		listToAccount   int64
	)
	listSet.StringVar(&listDate, 0, "date", "", "Closing-balance date YYYY-MM-DD (required)")
	listSet.IntVar(&listPage, 0, "page", 0, "Page number (0-indexed)")
	listSet.IntVar(&listPageSize, 0, "page-size", 0, "Page size")
	listSet.Int64Var(&listFromAccount, 0, "from-account", 0, "Lower bound account number")
	listSet.Int64Var(&listToAccount, 0, "to-account", 0, "Upper bound account number")
	listCmd := &ff.Command{
		Name:      "list",
		Usage:     "fiken account-balances list --company <slug> --date <yyyy-mm-dd>",
		ShortHelp: "List bookkeeping account balances as of a date.",
		Flags:     listSet,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).AccountBalancesList(ctx, ops.AccountBalancesListIn{
				Company:     *sf.flagCompany,
				Date:        ops.Date(listDate),
				Page:        listPage,
				PageSize:    listPageSize,
				FromAccount: listFromAccount,
				ToAccount:   listToAccount,
			})
			return Renderer(ctx).Render(res)
		},
	}

	// get
	getSet := ff.NewFlagSet("get")
	getSet.SetParent(set)
	var (
		getAccountCode string
		getDate        string
	)
	getSet.StringVar(&getAccountCode, 0, "account-code", "", "Account code (required)")
	getSet.StringVar(&getDate, 0, "date", "", "Closing-balance date YYYY-MM-DD (required)")
	getCmd := &ff.Command{
		Name:      "get",
		Usage:     "fiken account-balances get --company <slug> --account-code <code> --date <yyyy-mm-dd>",
		ShortHelp: "Get a single bookkeeping account balance as of a date.",
		Flags:     getSet,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).AccountBalancesGet(ctx, ops.AccountBalancesGetIn{
				Company:     *sf.flagCompany,
				AccountCode: getAccountCode,
				Date:        ops.Date(getDate),
			})
			return Renderer(ctx).Render(res)
		},
	}

	cmd.Subcommands = []*ff.Command{listCmd, getCmd}
	root.Subcommands = append(root.Subcommands, cmd)
	return nil
}
