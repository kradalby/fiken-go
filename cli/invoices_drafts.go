package cli

import (
	"context"
	"io"

	"github.com/peterbourgon/ff/v4"

	"github.com/kradalby/fiken-go/fiken"
	"github.com/kradalby/fiken-go/ops"
)

// invoicesDraftsCmd wires `fiken invoices drafts {list,get,create,update,delete,create-from}`.
//
// All subcommands hit the ogen Client; create + update read JSON
// payloads via --from-file. Attachments live in their own file.
func invoicesDraftsCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	set := ff.NewFlagSet("drafts")
	set.SetParent(parent)

	cmd := &ff.Command{
		Name:      "drafts",
		Usage:     "fiken invoices drafts <subcommand>",
		ShortHelp: "List, fetch, or manage invoice drafts.",
		Flags:     set,
	}

	cmd.Subcommands = append(
		cmd.Subcommands,
		invoicesDraftsListCmd(set, stdout, stderr, sf),
		invoicesDraftsGetCmd(set, stdout, stderr, sf),
		invoicesDraftsCreateCmd(set, stdout, stderr, sf),
		invoicesDraftsUpdateCmd(set, stdout, stderr, sf),
		invoicesDraftsDeleteCmd(set, stdout, stderr, sf),
		invoicesDraftsCreateFromCmd(set, stdout, stderr, sf),
		invoicesDraftsAttachmentsCmd(set, stdout, stderr, sf),
	)
	return cmd
}

// invoicesDraftsListCmd wires `fiken invoices drafts list ...`.
func invoicesDraftsListCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("list")
	fs.SetParent(parent)
	var (
		page           int
		pageSize       int
		orderReference string
		uuidFlag       string
	)
	fs.IntVar(&page, 0, "page", 0, "Page number (0-indexed)")
	fs.IntVar(&pageSize, 0, "page-size", 0, "Page size")
	fs.StringVar(&orderReference, 0, "order-reference", "", "Filter by order reference")
	fs.StringVar(&uuidFlag, 0, "uuid", "", "Filter by draft UUID")
	return &ff.Command{
		Name:      "list",
		Usage:     "fiken invoices drafts list --company <slug>",
		ShortHelp: "List invoice drafts for a company.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).InvoiceDraftsList(ctx, ops.InvoiceDraftsListIn{
				Company:        *sf.flagCompany,
				Page:           page,
				PageSize:       pageSize,
				OrderReference: orderReference,
				UUID:           uuidFlag,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// invoicesDraftsGetCmd wires `fiken invoices drafts get ...`.
func invoicesDraftsGetCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("get")
	fs.SetParent(parent)
	var draftID int64
	fs.Int64Var(&draftID, 0, "draft-id", 0, "Draft id (required)")
	return &ff.Command{
		Name:      "get",
		Usage:     "fiken invoices drafts get --company <slug> --draft-id <id>",
		ShortHelp: "Get a single invoice draft by id.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).InvoiceDraftsGet(ctx, ops.InvoiceDraftsGetIn{
				Company: *sf.flagCompany,
				DraftID: draftID,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// invoicesDraftsCreateCmd wires `fiken invoices drafts create ...`.
func invoicesDraftsCreateCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("create")
	fs.SetParent(parent)
	var fromFile string
	fs.StringVar(&fromFile, 0, "from-file", "", "Path to JSON body (required)")
	return &ff.Command{
		Name:      "create",
		Usage:     "fiken invoices drafts create --company <slug> --from-file <path>",
		ShortHelp: "Create a new invoice draft.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			body, e := ReadBodyFile[fiken.InvoiceishDraftRequest](fromFile)
			if e != nil {
				e.Op = ops.OpInvoicesDraftsCreate
				return Renderer(ctx).Render(ops.Err[ops.InvoiceDraftOut](e))
			}
			res := Client(ctx).InvoiceDraftsCreate(ctx, ops.InvoiceDraftsCreateIn{
				Company: *sf.flagCompany,
				Body:    body,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// invoicesDraftsUpdateCmd wires `fiken invoices drafts update ...`.
func invoicesDraftsUpdateCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("update")
	fs.SetParent(parent)
	var (
		draftID  int64
		fromFile string
	)
	fs.Int64Var(&draftID, 0, "draft-id", 0, "Draft id to update (required)")
	fs.StringVar(&fromFile, 0, "from-file", "", "Path to JSON body (required)")
	return &ff.Command{
		Name:      "update",
		Usage:     "fiken invoices drafts update --company <slug> --draft-id <id> --from-file <path>",
		ShortHelp: "Update an existing invoice draft.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			body, e := ReadBodyFile[fiken.InvoiceishDraftRequest](fromFile)
			if e != nil {
				e.Op = ops.OpInvoicesDraftsUpdate
				return Renderer(ctx).Render(ops.Err[ops.InvoiceDraftOut](e))
			}
			res := Client(ctx).InvoiceDraftsUpdate(ctx, ops.InvoiceDraftsUpdateIn{
				Company: *sf.flagCompany,
				DraftID: draftID,
				Body:    body,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// invoicesDraftsDeleteCmd wires `fiken invoices drafts delete ...`.
func invoicesDraftsDeleteCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("delete")
	fs.SetParent(parent)
	var draftID int64
	fs.Int64Var(&draftID, 0, "draft-id", 0, "Draft id to delete (required)")
	return &ff.Command{
		Name:      "delete",
		Usage:     "fiken invoices drafts delete --company <slug> --draft-id <id>",
		ShortHelp: "Delete an invoice draft.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).InvoiceDraftsDelete(ctx, ops.InvoiceDraftsDeleteIn{
				Company: *sf.flagCompany, DraftID: draftID,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// invoicesDraftsCreateFromCmd wires `fiken invoices drafts create-from ...`
// — turns a draft into a posted invoice upstream.
func invoicesDraftsCreateFromCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("create-from")
	fs.SetParent(parent)
	var draftID int64
	fs.Int64Var(&draftID, 0, "draft-id", 0, "Source draft id (required)")
	return &ff.Command{
		Name:      "create-from",
		Usage:     "fiken invoices drafts create-from --company <slug> --draft-id <id>",
		ShortHelp: "Promote a draft to a posted invoice.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).InvoiceDraftsCreateFrom(ctx, ops.InvoiceDraftsCreateFromIn{
				Company: *sf.flagCompany, DraftID: draftID,
			})
			return Renderer(ctx).Render(res)
		},
	}
}
