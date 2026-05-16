package cli

import (
	"context"
	"io"

	"github.com/peterbourgon/ff/v4"

	"github.com/kradalby/fiken-go/fiken"
	"github.com/kradalby/fiken-go/ops"
)

// AddBankAccounts wires `fiken bank-accounts {list,get,create}`.
//
// All subcommands hit the ogen Client. create reads its JSON body from
// --from-file via the shared ReadBodyFile helper.
func AddBankAccounts(root *ff.Command, stdout, stderr io.Writer, sf *sessionFactory) error {
	_ = stderr
	set := ff.NewFlagSet("bank-accounts")
	if parentFS, ok := root.Flags.(*ff.FlagSet); ok {
		set.SetParent(parentFS)
	}

	cmd := &ff.Command{
		Name:      "bank-accounts",
		Usage:     "fiken bank-accounts <subcommand>",
		ShortHelp: "Inspect company bank accounts.",
		Flags:     set,
	}

	// list
	listSet := ff.NewFlagSet("list")
	listSet.SetParent(set)
	var (
		listPage     int
		listPageSize int
		listInactive bool
	)
	listSet.IntVar(&listPage, 0, "page", 0, "Page number (0-indexed)")
	listSet.IntVar(&listPageSize, 0, "page-size", 0, "Page size")
	listSet.BoolVar(&listInactive, 0, "inactive", "Return inactive bank accounts instead of active")
	listCmd := &ff.Command{
		Name:      "list",
		Usage:     "fiken bank-accounts list --company <slug>",
		ShortHelp: "List bank accounts for a company.",
		Flags:     listSet,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).BankAccountsList(ctx, ops.BankAccountsListIn{
				Company:  *sf.flagCompany,
				Page:     listPage,
				PageSize: listPageSize,
				Inactive: listInactive,
			})
			return Renderer(ctx).Render(res)
		},
	}

	// get
	getSet := ff.NewFlagSet("get")
	getSet.SetParent(set)
	var getBankAccountID int64
	getSet.Int64Var(&getBankAccountID, 0, "bank-account-id", 0, "Bank account id (required)")
	getCmd := &ff.Command{
		Name:      "get",
		Usage:     "fiken bank-accounts get --company <slug> --bank-account-id <id>",
		ShortHelp: "Get a single bank account by id.",
		Flags:     getSet,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).BankAccountsGet(ctx, ops.BankAccountsGetIn{
				Company:       *sf.flagCompany,
				BankAccountID: getBankAccountID,
			})
			return Renderer(ctx).Render(res)
		},
	}

	// create
	createSet := ff.NewFlagSet("create")
	createSet.SetParent(set)
	var createFromFile string
	createSet.StringVar(&createFromFile, 0, "from-file", "", "Path to JSON body (required)")
	createCmd := &ff.Command{
		Name:      "create",
		Usage:     "fiken bank-accounts create --company <slug> --from-file <path>",
		ShortHelp: "Create a bank account.",
		Flags:     createSet,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			body, e := ReadBodyFile[fiken.BankAccountRequest](createFromFile)
			if e != nil {
				e.Op = ops.OpBankAccountsCreate
				return Renderer(ctx).Render(ops.Err[ops.BankAccountOut](e))
			}
			res := Client(ctx).BankAccountsCreate(ctx, ops.BankAccountsCreateIn{
				Company: *sf.flagCompany,
				Body:    body,
			})
			return Renderer(ctx).Render(res)
		},
	}

	cmd.Subcommands = []*ff.Command{listCmd, getCmd, createCmd}
	root.Subcommands = append(root.Subcommands, cmd)
	return nil
}
