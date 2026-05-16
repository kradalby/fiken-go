package cli

import (
	"context"
	"io"

	"github.com/peterbourgon/ff/v4"

	"github.com/kradalby/fiken-go/ops"
)

// invoicesAttachmentsCmd wires `fiken invoices attachments {list,attach}`.
func invoicesAttachmentsCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	set := ff.NewFlagSet("attachments")
	set.SetParent(parent)

	listFS := ff.NewFlagSet("list")
	listFS.SetParent(set)
	var listInvoiceID int64
	listFS.Int64Var(&listInvoiceID, 0, "invoice-id", 0, "Invoice id (required)")
	listCmd := &ff.Command{
		Name:      "list",
		Usage:     "fiken invoices attachments list --company <slug> --invoice-id <id>",
		ShortHelp: "List attachments for an invoice.",
		Flags:     listFS,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).InvoicesAttachmentsList(ctx, ops.InvoicesAttachmentsListIn{
				Company:   *sf.flagCompany,
				InvoiceID: listInvoiceID,
			})
			return Renderer(ctx).Render(res)
		},
	}

	attachFS := ff.NewFlagSet("attach")
	attachFS.SetParent(set)
	var (
		attachInvoiceID int64
		attachFilename  string
		attachFilePath  string
	)
	attachFS.Int64Var(&attachInvoiceID, 0, "invoice-id", 0, "Invoice id (required)")
	attachFS.StringVar(&attachFilename, 0, "filename", "", "Override filename (defaults to basename of --file)")
	attachFS.StringVar(&attachFilePath, 0, "file", "", "Local file path to upload (required)")
	attachCmd := &ff.Command{
		Name:      "attach",
		Usage:     "fiken invoices attachments attach --company <slug> --invoice-id <id> --file <path>",
		ShortHelp: "Attach a file to an invoice.",
		Flags:     attachFS,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).InvoicesAttachmentsAttach(ctx, ops.InvoicesAttachmentsAttachIn{
				Company:   *sf.flagCompany,
				InvoiceID: attachInvoiceID,
				Filename:  attachFilename,
				FilePath:  attachFilePath,
			})
			return Renderer(ctx).Render(res)
		},
	}

	return &ff.Command{
		Name:        "attachments",
		Usage:       "fiken invoices attachments <subcommand>",
		ShortHelp:   "List or attach files to an invoice.",
		Flags:       set,
		Subcommands: []*ff.Command{listCmd, attachCmd},
	}
}

// invoicesDraftsAttachmentsCmd wires
// `fiken invoices drafts attachments {list,attach}`.
func invoicesDraftsAttachmentsCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	set := ff.NewFlagSet("attachments")
	set.SetParent(parent)

	listFS := ff.NewFlagSet("list")
	listFS.SetParent(set)
	var listDraftID int64
	listFS.Int64Var(&listDraftID, 0, "draft-id", 0, "Draft id (required)")
	listCmd := &ff.Command{
		Name:      "list",
		Usage:     "fiken invoices drafts attachments list --company <slug> --draft-id <id>",
		ShortHelp: "List attachments for an invoice draft.",
		Flags:     listFS,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).InvoiceDraftsAttachmentsList(ctx, ops.InvoiceDraftsAttachmentsListIn{
				Company: *sf.flagCompany,
				DraftID: listDraftID,
			})
			return Renderer(ctx).Render(res)
		},
	}

	attachFS := ff.NewFlagSet("attach")
	attachFS.SetParent(set)
	var (
		attachDraftID  int64
		attachFilename string
		attachFilePath string
	)
	attachFS.Int64Var(&attachDraftID, 0, "draft-id", 0, "Draft id (required)")
	attachFS.StringVar(&attachFilename, 0, "filename", "", "Override filename (defaults to basename of --file)")
	attachFS.StringVar(&attachFilePath, 0, "file", "", "Local file path to upload (required)")
	attachCmd := &ff.Command{
		Name:      "attach",
		Usage:     "fiken invoices drafts attachments attach --company <slug> --draft-id <id> --file <path>",
		ShortHelp: "Attach a file to an invoice draft.",
		Flags:     attachFS,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).InvoiceDraftsAttachmentsAttach(ctx, ops.InvoiceDraftsAttachmentsAttachIn{
				Company:  *sf.flagCompany,
				DraftID:  attachDraftID,
				Filename: attachFilename,
				FilePath: attachFilePath,
			})
			return Renderer(ctx).Render(res)
		},
	}

	return &ff.Command{
		Name:        "attachments",
		Usage:       "fiken invoices drafts attachments <subcommand>",
		ShortHelp:   "List or attach files to an invoice draft.",
		Flags:       set,
		Subcommands: []*ff.Command{listCmd, attachCmd},
	}
}
