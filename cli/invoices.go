package cli

import (
	"context"
	"io"

	"github.com/peterbourgon/ff/v4"

	"github.com/kradalby/fiken-go/fiken"
	"github.com/kradalby/fiken-go/ops"
)

// AddInvoices wires `fiken invoices {list,get,send,counter,drafts,attachments}`.
//
// list + get + send + counter create are wired against the ogen
// Client; send + counter take their request bodies via --from-file or
// scalar flags. drafts + attachments are wired in their own files.
func AddInvoices(root *ff.Command, stdout, stderr io.Writer, sf *sessionFactory) error {
	_ = stderr
	set := ff.NewFlagSet("invoices")
	if parentFS, ok := root.Flags.(*ff.FlagSet); ok {
		set.SetParent(parentFS)
	}

	cmd := &ff.Command{
		Name:      "invoices",
		Usage:     "fiken invoices <subcommand>",
		ShortHelp: "List, fetch, send, or manage invoices.",
		Flags:     set,
	}

	cmd.Subcommands = append(
		cmd.Subcommands,
		invoicesListCmd(set, stdout, stderr, sf),
		invoicesGetCmd(set, stdout, stderr, sf),
		invoicesCreateCmd(set, stdout, stderr, sf),
		invoicesUpdateCmd(set, stdout, stderr, sf),
		invoicesSendCmd(set, stdout, stderr, sf),
		invoicesCounterCmd(set, stdout, stderr, sf),
		invoicesAttachmentsCmd(set, stdout, stderr, sf),
		invoicesDraftsCmd(set, stdout, stderr, sf),
	)

	root.Subcommands = append(root.Subcommands, cmd)
	return nil
}

// invoicesListCmd wires `fiken invoices list ...`.
func invoicesListCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("list")
	fs.SetParent(parent)
	var (
		page             int
		pageSize         int
		issueDate        string
		issueDateLe      string
		issueDateLt      string
		issueDateGe      string
		issueDateGt      string
		lastModified     string
		lastModifiedLe   string
		lastModifiedLt   string
		lastModifiedGe   string
		lastModifiedGt   string
		dueDate          string
		dueDateLe        string
		dueDateLt        string
		dueDateGe        string
		dueDateGt        string
		customerID       int64
		settled          bool
		settledSet       bool
		orderReference   string
		invoiceDraftUUID string
		invoiceNumber    string
	)
	fs.IntVar(&page, 0, "page", 0, "Page number (0-indexed)")
	fs.IntVar(&pageSize, 0, "page-size", 0, "Page size")
	fs.StringVar(&issueDate, 0, "issue-date", "", "Filter by exact issue date (YYYY-MM-DD)")
	fs.StringVar(&issueDateLe, 0, "issue-date-le", "", "Filter issue date <= value")
	fs.StringVar(&issueDateLt, 0, "issue-date-lt", "", "Filter issue date < value")
	fs.StringVar(&issueDateGe, 0, "issue-date-ge", "", "Filter issue date >= value")
	fs.StringVar(&issueDateGt, 0, "issue-date-gt", "", "Filter issue date > value")
	fs.StringVar(&lastModified, 0, "last-modified", "", "Filter by exact last-modified date")
	fs.StringVar(&lastModifiedLe, 0, "last-modified-le", "", "Filter last-modified <= value")
	fs.StringVar(&lastModifiedLt, 0, "last-modified-lt", "", "Filter last-modified < value")
	fs.StringVar(&lastModifiedGe, 0, "last-modified-ge", "", "Filter last-modified >= value")
	fs.StringVar(&lastModifiedGt, 0, "last-modified-gt", "", "Filter last-modified > value")
	fs.StringVar(&dueDate, 0, "due-date", "", "Filter by exact due date")
	fs.StringVar(&dueDateLe, 0, "due-date-le", "", "Filter due-date <= value")
	fs.StringVar(&dueDateLt, 0, "due-date-lt", "", "Filter due-date < value")
	fs.StringVar(&dueDateGe, 0, "due-date-ge", "", "Filter due-date >= value")
	fs.StringVar(&dueDateGt, 0, "due-date-gt", "", "Filter due-date > value")
	fs.Int64Var(&customerID, 0, "customer-id", 0, "Filter by customer contact id")
	fs.BoolVar(&settled, 0, "settled", "Filter to settled invoices (default: include unsettled)")
	fs.StringVar(&orderReference, 0, "order-reference", "", "Filter by order reference")
	fs.StringVar(&invoiceDraftUUID, 0, "invoice-draft-uuid", "", "Filter by source draft UUID")
	fs.StringVar(&invoiceNumber, 0, "invoice-number", "", "Filter by exact invoice number string")

	return &ff.Command{
		Name:      "list",
		Usage:     "fiken invoices list --company <slug>",
		ShortHelp: "List invoices for a company.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			in := ops.InvoicesListIn{
				Company:          *sf.flagCompany,
				Page:             page,
				PageSize:         pageSize,
				IssueDate:        ops.Date(issueDate),
				IssueDateLe:      ops.Date(issueDateLe),
				IssueDateLt:      ops.Date(issueDateLt),
				IssueDateGe:      ops.Date(issueDateGe),
				IssueDateGt:      ops.Date(issueDateGt),
				LastModified:     ops.Date(lastModified),
				LastModifiedLe:   ops.Date(lastModifiedLe),
				LastModifiedLt:   ops.Date(lastModifiedLt),
				LastModifiedGe:   ops.Date(lastModifiedGe),
				LastModifiedGt:   ops.Date(lastModifiedGt),
				DueDate:          ops.Date(dueDate),
				DueDateLe:        ops.Date(dueDateLe),
				DueDateLt:        ops.Date(dueDateLt),
				DueDateGe:        ops.Date(dueDateGe),
				DueDateGt:        ops.Date(dueDateGt),
				CustomerID:       customerID,
				OrderReference:   orderReference,
				InvoiceDraftUUID: invoiceDraftUUID,
				InvoiceNumber:    invoiceNumber,
			}
			if settledSet {
				in.Settled = &settled
			}
			res := Client(ctx).InvoicesList(ctx, in)
			return Renderer(ctx).Render(res)
		},
	}
}

// invoicesGetCmd wires `fiken invoices get ...`.
func invoicesGetCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("get")
	fs.SetParent(parent)
	var invoiceID int64
	fs.Int64Var(&invoiceID, 0, "invoice-id", 0, "Invoice id (required)")
	return &ff.Command{
		Name:      "get",
		Usage:     "fiken invoices get --company <slug> --invoice-id <id>",
		ShortHelp: "Get a single invoice by id.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).InvoicesGet(ctx, ops.InvoicesGetIn{
				Company:   *sf.flagCompany,
				InvoiceID: invoiceID,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// invoicesSendCmd wires `fiken invoices send ...`.
func invoicesSendCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("send")
	fs.SetParent(parent)
	var fromFile string
	fs.StringVar(&fromFile, 0, "from-file", "", "Path to JSON body (required)")
	return &ff.Command{
		Name:      "send",
		Usage:     "fiken invoices send --company <slug> --from-file <path>",
		ShortHelp: "Send an invoice to its customer.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			body, e := ReadBodyFile[fiken.SendInvoiceRequest](fromFile)
			if e != nil {
				e.Op = ops.OpInvoicesSend
				return Renderer(ctx).Render(ops.Err[ops.InvoicesSendOut](e))
			}
			res := Client(ctx).InvoicesSend(ctx, ops.InvoicesSendIn{
				Company: *sf.flagCompany,
				Body:    body,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// invoicesCreateCmd wires `fiken invoices create ...`.
func invoicesCreateCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("create")
	fs.SetParent(parent)
	var fromFile string
	fs.StringVar(&fromFile, 0, "from-file", "", "Path to JSON body (required)")
	return &ff.Command{
		Name:      "create",
		Usage:     "fiken invoices create --company <slug> --from-file <path>",
		ShortHelp: "Create a posted invoice directly (no draft step).",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			body, e := ReadBodyFile[fiken.InvoiceRequest](fromFile)
			if e != nil {
				e.Op = ops.OpInvoicesCreate
				return Renderer(ctx).Render(ops.Err[ops.InvoicesCreateOut](e))
			}
			res := Client(ctx).InvoicesCreate(ctx, ops.InvoicesCreateIn{
				Company: *sf.flagCompany,
				Body:    body,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// invoicesUpdateCmd wires `fiken invoices update ...`.
func invoicesUpdateCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("update")
	fs.SetParent(parent)
	var (
		invoiceID int64
		fromFile  string
	)
	fs.Int64Var(&invoiceID, 0, "invoice-id", 0, "Invoice id to update (required)")
	fs.StringVar(&fromFile, 0, "from-file", "", "Path to JSON body (required)")
	return &ff.Command{
		Name:      "update",
		Usage:     "fiken invoices update --company <slug> --invoice-id <id> --from-file <path>",
		ShortHelp: "Update mutable fields on a posted invoice.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			body, e := ReadBodyFile[fiken.UpdateInvoiceRequest](fromFile)
			if e != nil {
				e.Op = ops.OpInvoicesUpdate
				return Renderer(ctx).Render(ops.Err[ops.InvoicesUpdateOut](e))
			}
			res := Client(ctx).InvoicesUpdate(ctx, ops.InvoicesUpdateIn{
				Company:   *sf.flagCompany,
				InvoiceID: invoiceID,
				Body:      body,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// invoicesCounterCmd wires `fiken invoices counter {create,get}`.
func invoicesCounterCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	set := ff.NewFlagSet("counter")
	set.SetParent(parent)

	createFS := ff.NewFlagSet("create")
	createFS.SetParent(set)
	var value int64
	createFS.Int64Var(&value, 0, "value", 0, "Starting counter value (required)")
	createCmd := &ff.Command{
		Name:      "create",
		Usage:     "fiken invoices counter create --company <slug> --value <n>",
		ShortHelp: "Initialize the invoice counter for the fiscal year.",
		Flags:     createFS,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			// value is bounded to int32 by the upstream Counter schema;
			// clamp to int32 max defensively (gosec G115 dislikes a bare
			// int->int32 cast even from int64).
			var v32 int32
			switch {
			case value > 2147483647:
				v32 = 2147483647
			case value < -2147483648:
				v32 = -2147483648
			default:
				v32 = int32(value)
			}
			res := Client(ctx).InvoicesCounterCreate(ctx, ops.InvoicesCounterCreateIn{
				Company: *sf.flagCompany, Value: v32,
			})
			return Renderer(ctx).Render(res)
		},
	}

	getFS := ff.NewFlagSet("get")
	getFS.SetParent(set)
	getCmd := &ff.Command{
		Name:      "get",
		Usage:     "fiken invoices counter get --company <slug>",
		ShortHelp: "Get the current invoice counter value.",
		Flags:     getFS,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).InvoicesCounterGet(ctx, ops.CounterGetIn{
				Company: *sf.flagCompany,
			})
			return Renderer(ctx).Render(res)
		},
	}

	return &ff.Command{
		Name:        "counter",
		Usage:       "fiken invoices counter <subcommand>",
		ShortHelp:   "Manage the invoice counter.",
		Flags:       set,
		Subcommands: []*ff.Command{createCmd, getCmd},
	}
}
