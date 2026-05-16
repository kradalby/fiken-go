package cli

import (
	"context"
	"io"

	"github.com/peterbourgon/ff/v4"

	"github.com/kradalby/fiken-go/ops"
)

// AddAccounts wires `fiken accounts {list,get}`. The accounts tag in
// the upstream OpenAPI exposes only GETs, so there are no mutating
// stubs to wire here.
func AddAccounts(root *ff.Command, stdout, stderr io.Writer, sf *sessionFactory) error {
	_ = stderr
	set := ff.NewFlagSet("accounts")
	if parentFS, ok := root.Flags.(*ff.FlagSet); ok {
		set.SetParent(parentFS)
	}

	cmd := &ff.Command{
		Name:      "accounts",
		Usage:     "fiken accounts <subcommand>",
		ShortHelp: "Inspect bookkeeping accounts.",
		Flags:     set,
	}

	// list
	listSet := ff.NewFlagSet("list")
	listSet.SetParent(set)
	var (
		listPage        int
		listPageSize    int
		listFromAccount int64
		listToAccount   int64
		listRange       string
	)
	listSet.IntVar(&listPage, 0, "page", 0, "Page number (0-indexed)")
	listSet.IntVar(&listPageSize, 0, "page-size", 0, "Page size")
	listSet.Int64Var(&listFromAccount, 0, "from-account", 0, "Lower bound account number")
	listSet.Int64Var(&listToAccount, 0, "to-account", 0, "Upper bound account number")
	listSet.StringVar(&listRange, 0, "range", "", "Comma-separated codes/ranges, e.g. \"1000-1500,3020\"")
	listCmd := &ff.Command{
		Name:      "list",
		Usage:     "fiken accounts list --company <slug>",
		ShortHelp: "List bookkeeping accounts for a company.",
		Flags:     listSet,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).AccountsList(ctx, ops.AccountsListIn{
				Company:     *sf.flagCompany,
				Page:        listPage,
				PageSize:    listPageSize,
				FromAccount: listFromAccount,
				ToAccount:   listToAccount,
				Range:       listRange,
			})
			return Renderer(ctx).Render(res)
		},
	}

	// get
	getSet := ff.NewFlagSet("get")
	getSet.SetParent(set)
	var getAccountCode string
	getSet.StringVar(&getAccountCode, 0, "account-code", "", "Account code (required)")
	getCmd := &ff.Command{
		Name:      "get",
		Usage:     "fiken accounts get --company <slug> --account-code <code>",
		ShortHelp: "Get a single bookkeeping account by code.",
		Flags:     getSet,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).AccountsGet(ctx, ops.AccountsGetIn{
				Company:     *sf.flagCompany,
				AccountCode: getAccountCode,
			})
			return Renderer(ctx).Render(res)
		},
	}

	cmd.Subcommands = []*ff.Command{listCmd, getCmd}
	root.Subcommands = append(root.Subcommands, cmd)
	return nil
}
