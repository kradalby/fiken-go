package cli

import (
	"context"
	"io"

	"github.com/peterbourgon/ff/v4"

	"github.com/kradalby/fiken-go/ops"
)

// AddTransactions wires `fiken transactions {list,get}`.
//
// Both subcommands are fully wired against the ogen Client. The
// upstream tag exposes only GETs in Plan C scope — the
// `deleteTransaction` op stays unimplemented until a mutating task
// covers it.
func AddTransactions(root *ff.Command, stdout, stderr io.Writer, sf *sessionFactory) error {
	_ = stderr
	set := ff.NewFlagSet("transactions")
	if parentFS, ok := root.Flags.(*ff.FlagSet); ok {
		set.SetParent(parentFS)
	}

	cmd := &ff.Command{
		Name:      "transactions",
		Usage:     "fiken transactions <subcommand>",
		ShortHelp: "Inspect transactions for a company.",
		Flags:     set,
	}

	// list
	listSet := ff.NewFlagSet("list")
	listSet.SetParent(set)
	var (
		listPage           int
		listPageSize       int
		listLastModified   string
		listLastModifiedLe string
		listLastModifiedLt string
		listLastModifiedGe string
		listLastModifiedGt string
		listCreatedDate    string
		listCreatedDateLe  string
		listCreatedDateLt  string
		listCreatedDateGe  string
		listCreatedDateGt  string
	)
	listSet.IntVar(&listPage, 0, "page", 0, "Page number (0-indexed)")
	listSet.IntVar(&listPageSize, 0, "page-size", 0, "Page size")
	listSet.StringVar(&listLastModified, 0, "last-modified", "", "Filter by exact last-modified date (YYYY-MM-DD)")
	listSet.StringVar(&listLastModifiedLe, 0, "last-modified-le", "", "Filter last-modified <= value")
	listSet.StringVar(&listLastModifiedLt, 0, "last-modified-lt", "", "Filter last-modified < value")
	listSet.StringVar(&listLastModifiedGe, 0, "last-modified-ge", "", "Filter last-modified >= value")
	listSet.StringVar(&listLastModifiedGt, 0, "last-modified-gt", "", "Filter last-modified > value")
	listSet.StringVar(&listCreatedDate, 0, "created-date", "", "Filter by exact created date")
	listSet.StringVar(&listCreatedDateLe, 0, "created-date-le", "", "Filter created-date <= value")
	listSet.StringVar(&listCreatedDateLt, 0, "created-date-lt", "", "Filter created-date < value")
	listSet.StringVar(&listCreatedDateGe, 0, "created-date-ge", "", "Filter created-date >= value")
	listSet.StringVar(&listCreatedDateGt, 0, "created-date-gt", "", "Filter created-date > value")
	listCmd := &ff.Command{
		Name:      "list",
		Usage:     "fiken transactions list --company <slug>",
		ShortHelp: "List transactions for a company.",
		Flags:     listSet,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).TransactionsList(ctx, ops.TransactionsListIn{
				Company:        *sf.flagCompany,
				Page:           listPage,
				PageSize:       listPageSize,
				LastModified:   ops.Date(listLastModified),
				LastModifiedLe: ops.Date(listLastModifiedLe),
				LastModifiedLt: ops.Date(listLastModifiedLt),
				LastModifiedGe: ops.Date(listLastModifiedGe),
				LastModifiedGt: ops.Date(listLastModifiedGt),
				CreatedDate:    ops.Date(listCreatedDate),
				CreatedDateLe:  ops.Date(listCreatedDateLe),
				CreatedDateLt:  ops.Date(listCreatedDateLt),
				CreatedDateGe:  ops.Date(listCreatedDateGe),
				CreatedDateGt:  ops.Date(listCreatedDateGt),
			})
			return Renderer(ctx).Render(res)
		},
	}

	// get
	getSet := ff.NewFlagSet("get")
	getSet.SetParent(set)
	var getTransactionID int64
	getSet.Int64Var(&getTransactionID, 0, "transaction-id", 0, "Transaction id (required)")
	getCmd := &ff.Command{
		Name:      "get",
		Usage:     "fiken transactions get --company <slug> --transaction-id <id>",
		ShortHelp: "Get a single transaction by id.",
		Flags:     getSet,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).TransactionsGet(ctx, ops.TransactionsGetIn{
				Company:       *sf.flagCompany,
				TransactionID: getTransactionID,
			})
			return Renderer(ctx).Render(res)
		},
	}

	// delete (soft-archive, PATCH with required description query param)
	deleteSet := ff.NewFlagSet("delete")
	deleteSet.SetParent(set)
	var (
		deleteTransactionID int64
		deleteDescription   string
	)
	deleteSet.Int64Var(&deleteTransactionID, 0, "transaction-id", 0, "Transaction id (required)")
	deleteSet.StringVar(&deleteDescription, 0, "description", "", "Audit-trail description (required)")
	deleteCmd := &ff.Command{
		Name:      "delete",
		Usage:     "fiken transactions delete --company <slug> --transaction-id <id> --description <text>",
		ShortHelp: "Soft-archive a transaction.",
		Flags:     deleteSet,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).TransactionsDelete(ctx, ops.TransactionsDeleteIn{
				Company:       *sf.flagCompany,
				TransactionID: deleteTransactionID,
				Description:   deleteDescription,
			})
			return Renderer(ctx).Render(res)
		},
	}

	cmd.Subcommands = []*ff.Command{listCmd, getCmd, deleteCmd}
	root.Subcommands = append(root.Subcommands, cmd)
	return nil
}
