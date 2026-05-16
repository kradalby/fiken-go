package cli

import (
	"context"
	"io"

	"github.com/peterbourgon/ff/v4"

	"github.com/kradalby/fiken-go/fiken"
	"github.com/kradalby/fiken-go/ops"
)

// AddPurchases wires
// `fiken purchases {list,get,create,delete,payments,attachments}`.
//
// list + get + attachments-list + payments-list/get are fully wired
// against the upstream read paths. Mutating ops (create, delete,
// payments create, attach) register as stubs that surface CodeInternal
// so CLI help and MCP tool discovery stay complete; mutating wiring
// lands in a follow-up task.
func AddPurchases(root *ff.Command, stdout, stderr io.Writer, sf *sessionFactory) error {
	_ = stderr
	set := ff.NewFlagSet("purchases")
	if parentFS, ok := root.Flags.(*ff.FlagSet); ok {
		set.SetParent(parentFS)
	}

	cmd := &ff.Command{
		Name:      "purchases",
		Usage:     "fiken purchases <subcommand>",
		ShortHelp: "List, fetch, or manage purchases.",
		Flags:     set,
	}

	cmd.Subcommands = append(
		cmd.Subcommands,
		purchasesListCmd(set, stdout, stderr, sf),
		purchasesGetCmd(set, stdout, stderr, sf),
		purchasesCreateCmd(set, stdout, stderr, sf),
		purchasesDeleteCmd(set, stdout, stderr, sf),
		purchasesPaymentsCmd(set, stdout, stderr, sf),
		purchasesAttachmentsCmd(set, stdout, stderr, sf),
		purchasesDraftsCmd(set, stdout, stderr, sf),
	)

	root.Subcommands = append(root.Subcommands, cmd)
	return nil
}

// purchasesListCmd wires `fiken purchases list ...`.
func purchasesListCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("list")
	fs.SetParent(parent)
	var (
		page          int
		pageSize      int
		date          string
		dateLe        string
		dateLt        string
		dateGe        string
		dateGt        string
		settledDate   string
		settledDateLe string
		settledDateLt string
		settledDateGe string
		settledDateGt string
		paid          bool
		paidSet       bool
	)
	fs.IntVar(&page, 0, "page", 0, "Page number (0-indexed)")
	fs.IntVar(&pageSize, 0, "page-size", 0, "Page size")
	fs.StringVar(&date, 0, "date", "", "Filter by exact purchase date (YYYY-MM-DD)")
	fs.StringVar(&dateLe, 0, "date-le", "", "Filter date <= value")
	fs.StringVar(&dateLt, 0, "date-lt", "", "Filter date < value")
	fs.StringVar(&dateGe, 0, "date-ge", "", "Filter date >= value")
	fs.StringVar(&dateGt, 0, "date-gt", "", "Filter date > value")
	fs.StringVar(&settledDate, 0, "settled-date", "", "Filter by exact settled date")
	fs.StringVar(&settledDateLe, 0, "settled-date-le", "", "Filter settled-date <= value")
	fs.StringVar(&settledDateLt, 0, "settled-date-lt", "", "Filter settled-date < value")
	fs.StringVar(&settledDateGe, 0, "settled-date-ge", "", "Filter settled-date >= value")
	fs.StringVar(&settledDateGt, 0, "settled-date-gt", "", "Filter settled-date > value")
	fs.BoolVar(&paid, 0, "paid", "Filter to paid purchases (default: include unpaid)")
	return &ff.Command{
		Name:      "list",
		Usage:     "fiken purchases list --company <slug>",
		ShortHelp: "List purchases for a company.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			in := ops.PurchasesListIn{
				Company:       *sf.flagCompany,
				Page:          page,
				PageSize:      pageSize,
				Date:          ops.Date(date),
				DateLe:        ops.Date(dateLe),
				DateLt:        ops.Date(dateLt),
				DateGe:        ops.Date(dateGe),
				DateGt:        ops.Date(dateGt),
				SettledDate:   ops.Date(settledDate),
				SettledDateLe: ops.Date(settledDateLe),
				SettledDateLt: ops.Date(settledDateLt),
				SettledDateGe: ops.Date(settledDateGe),
				SettledDateGt: ops.Date(settledDateGt),
			}
			if paidSet {
				in.Paid = &paid
			}
			res := Client(ctx).PurchasesList(ctx, in)
			return Renderer(ctx).Render(res)
		},
	}
}

// purchasesGetCmd wires `fiken purchases get ...`.
func purchasesGetCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("get")
	fs.SetParent(parent)
	var purchaseID int64
	fs.Int64Var(&purchaseID, 0, "purchase-id", 0, "Purchase id (required)")
	return &ff.Command{
		Name:      "get",
		Usage:     "fiken purchases get --company <slug> --purchase-id <id>",
		ShortHelp: "Get a single purchase by id.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).PurchasesGet(ctx, ops.PurchasesGetIn{
				Company:    *sf.flagCompany,
				PurchaseID: purchaseID,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// purchasesCreateCmd wires `fiken purchases create ...`.
func purchasesCreateCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("create")
	fs.SetParent(parent)
	var fromFile string
	fs.StringVar(&fromFile, 0, "from-file", "", "Path to JSON body (required)")
	return &ff.Command{
		Name:      "create",
		Usage:     "fiken purchases create --company <slug> --from-file <path>",
		ShortHelp: "Create a new purchase.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			body, e := ReadBodyFile[fiken.PurchaseRequest](fromFile)
			if e != nil {
				e.Op = ops.OpPurchasesCreate
				return Renderer(ctx).Render(ops.Err[ops.PurchasesCreateOut](e))
			}
			res := Client(ctx).PurchasesCreate(ctx, ops.PurchasesCreateIn{
				Company: *sf.flagCompany,
				Body:    body,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// purchasesDeleteCmd wires `fiken purchases delete ...` (stub).
func purchasesDeleteCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("delete")
	fs.SetParent(parent)
	var (
		purchaseID  int64
		description string
	)
	fs.Int64Var(&purchaseID, 0, "purchase-id", 0, "Purchase id to delete (required)")
	fs.StringVar(&description, 0, "description", "", "Audit description (required upstream)")
	return &ff.Command{
		Name:      "delete",
		Usage:     "fiken purchases delete --company <slug> --purchase-id <id> --description <text>",
		ShortHelp: "Soft-archive a purchase with an audit description.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).PurchasesDelete(ctx, ops.PurchasesDeleteIn{
				Company:     *sf.flagCompany,
				PurchaseID:  purchaseID,
				Description: description,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// purchasesPaymentsCmd wires `fiken purchases payments {list,get,create}`.
// list + get are fully wired against read paths; create surfaces the
// stub error.
func purchasesPaymentsCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	set := ff.NewFlagSet("payments")
	set.SetParent(parent)

	cmd := &ff.Command{
		Name:      "payments",
		Usage:     "fiken purchases payments <subcommand>",
		ShortHelp: "List, fetch, or register payments for a purchase.",
		Flags:     set,
	}

	cmd.Subcommands = append(
		cmd.Subcommands,
		purchasesPaymentsListCmd(set, stdout, stderr, sf),
		purchasesPaymentsGetCmd(set, stdout, stderr, sf),
		purchasesPaymentsCreateCmd(set, stdout, stderr, sf),
	)
	return cmd
}

// purchasesPaymentsListCmd wires `fiken purchases payments list ...`.
func purchasesPaymentsListCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("list")
	fs.SetParent(parent)
	var purchaseID int64
	fs.Int64Var(&purchaseID, 0, "purchase-id", 0, "Purchase id (required)")
	return &ff.Command{
		Name:      "list",
		Usage:     "fiken purchases payments list --company <slug> --purchase-id <id>",
		ShortHelp: "List payments for a purchase.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).PurchasesPaymentsList(ctx, ops.PurchasesPaymentsListIn{
				Company:    *sf.flagCompany,
				PurchaseID: purchaseID,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// purchasesPaymentsGetCmd wires `fiken purchases payments get ...`.
func purchasesPaymentsGetCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("get")
	fs.SetParent(parent)
	var (
		purchaseID int64
		paymentID  int64
	)
	fs.Int64Var(&purchaseID, 0, "purchase-id", 0, "Purchase id (required)")
	fs.Int64Var(&paymentID, 0, "payment-id", 0, "Payment id (required)")
	return &ff.Command{
		Name:      "get",
		Usage:     "fiken purchases payments get --company <slug> --purchase-id <id> --payment-id <id>",
		ShortHelp: "Get a single payment by id.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).PurchasesPaymentsGet(ctx, ops.PurchasesPaymentsGetIn{
				Company:    *sf.flagCompany,
				PurchaseID: purchaseID,
				PaymentID:  paymentID,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// purchasesPaymentsCreateCmd wires `fiken purchases payments create ...`.
func purchasesPaymentsCreateCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("create")
	fs.SetParent(parent)
	var (
		purchaseID int64
		fromFile   string
	)
	fs.Int64Var(&purchaseID, 0, "purchase-id", 0, "Purchase id (required)")
	fs.StringVar(&fromFile, 0, "from-file", "", "Path to JSON body (required)")
	return &ff.Command{
		Name:      "create",
		Usage:     "fiken purchases payments create --company <slug> --purchase-id <id> --from-file <path>",
		ShortHelp: "Register a payment for a purchase.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			body, e := ReadBodyFile[fiken.Payment](fromFile)
			if e != nil {
				e.Op = ops.OpPurchasesPaymentsCreate
				return Renderer(ctx).Render(ops.Err[ops.PurchasesPaymentsCreateOut](e))
			}
			res := Client(ctx).PurchasesPaymentsCreate(ctx, ops.PurchasesPaymentsCreateIn{
				Company:    *sf.flagCompany,
				PurchaseID: purchaseID,
				Body:       body,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// purchasesAttachmentsCmd wires `fiken purchases attachments {list,attach}`.
// list is fully wired; attach surfaces the Plan-D-deferred stub error.
func purchasesAttachmentsCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	set := ff.NewFlagSet("attachments")
	set.SetParent(parent)

	listFS := ff.NewFlagSet("list")
	listFS.SetParent(set)
	var listPurchaseID int64
	listFS.Int64Var(&listPurchaseID, 0, "purchase-id", 0, "Purchase id (required)")
	listCmd := &ff.Command{
		Name:      "list",
		Usage:     "fiken purchases attachments list --company <slug> --purchase-id <id>",
		ShortHelp: "List attachments for a purchase.",
		Flags:     listFS,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).PurchasesAttachmentsList(ctx, ops.PurchasesAttachmentsListIn{
				Company:    *sf.flagCompany,
				PurchaseID: listPurchaseID,
			})
			return Renderer(ctx).Render(res)
		},
	}

	attachFS := ff.NewFlagSet("attach")
	attachFS.SetParent(set)
	var (
		attachPurchaseID int64
		attachFilename   string
		attachFilePath   string
		attachToSale     bool
		attachToPayment  bool
	)
	attachFS.Int64Var(&attachPurchaseID, 0, "purchase-id", 0, "Purchase id (required)")
	attachFS.StringVar(&attachFilename, 0, "filename", "", "Override filename (defaults to basename of --file)")
	attachFS.StringVar(&attachFilePath, 0, "file", "", "Local file path to upload (required)")
	attachFS.BoolVar(&attachToSale, 0, "attach-to-sale", "Mark this attachment as documenting the purchase invoice")
	attachFS.BoolVar(&attachToPayment, 0, "attach-to-payment", "Mark this attachment as documenting the payment")
	attachCmd := &ff.Command{
		Name:      "attach",
		Usage:     "fiken purchases attachments attach --company <slug> --purchase-id <id> --file <path>",
		ShortHelp: "Attach a file to a purchase.",
		Flags:     attachFS,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).PurchasesAttach(ctx, ops.PurchasesAttachIn{
				Company:         *sf.flagCompany,
				PurchaseID:      attachPurchaseID,
				Filename:        attachFilename,
				FilePath:        attachFilePath,
				AttachToSale:    attachToSale,
				AttachToPayment: attachToPayment,
			})
			return Renderer(ctx).Render(res)
		},
	}

	return &ff.Command{
		Name:        "attachments",
		Usage:       "fiken purchases attachments <subcommand>",
		ShortHelp:   "List or attach files to a purchase.",
		Flags:       set,
		Subcommands: []*ff.Command{listCmd, attachCmd},
	}
}
