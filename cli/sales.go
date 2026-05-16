package cli

import (
	"context"
	"io"

	"github.com/peterbourgon/ff/v4"

	"github.com/kradalby/fiken-go/fiken"
	"github.com/kradalby/fiken-go/ops"
)

// AddSales wires
// `fiken sales {list,get,create,delete,settle,write-off,payments,attachments}`.
//
// list + get + attachments-list + payments-list/get are fully wired
// against the upstream read paths. Mutating ops (create, delete,
// settle, write-off, payments create, attach) register as stubs that
// surface CodeInternal so CLI help and MCP tool discovery stay
// complete; mutating wiring lands in a follow-up task.
func AddSales(root *ff.Command, stdout, stderr io.Writer, sf *sessionFactory) error {
	_ = stderr
	set := ff.NewFlagSet("sales")
	if parentFS, ok := root.Flags.(*ff.FlagSet); ok {
		set.SetParent(parentFS)
	}

	cmd := &ff.Command{
		Name:      "sales",
		Usage:     "fiken sales <subcommand>",
		ShortHelp: "List, fetch, or manage sales.",
		Flags:     set,
	}

	cmd.Subcommands = append(
		cmd.Subcommands,
		salesListCmd(set, stdout, stderr, sf),
		salesGetCmd(set, stdout, stderr, sf),
		salesCreateCmd(set, stdout, stderr, sf),
		salesDeleteCmd(set, stdout, stderr, sf),
		salesSettleCmd(set, stdout, stderr, sf),
		salesWriteOffCmd(set, stdout, stderr, sf),
		salesPaymentsCmd(set, stdout, stderr, sf),
		salesAttachmentsCmd(set, stdout, stderr, sf),
		salesDraftsCmd(set, stdout, stderr, sf),
	)

	root.Subcommands = append(root.Subcommands, cmd)
	return nil
}

// salesListCmd wires `fiken sales list ...`.
func salesListCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("list")
	fs.SetParent(parent)
	var (
		page           int
		pageSize       int
		date           string
		dateLe         string
		dateLt         string
		dateGe         string
		dateGt         string
		lastModified   string
		lastModifiedLe string
		lastModifiedLt string
		lastModifiedGe string
		lastModifiedGt string
		saleNumber     string
		settled        bool
		settledSet     bool
		contactID      int64
	)
	fs.IntVar(&page, 0, "page", 0, "Page number (0-indexed)")
	fs.IntVar(&pageSize, 0, "page-size", 0, "Page size")
	fs.StringVar(&date, 0, "date", "", "Filter by exact sale date (YYYY-MM-DD)")
	fs.StringVar(&dateLe, 0, "date-le", "", "Filter date <= value")
	fs.StringVar(&dateLt, 0, "date-lt", "", "Filter date < value")
	fs.StringVar(&dateGe, 0, "date-ge", "", "Filter date >= value")
	fs.StringVar(&dateGt, 0, "date-gt", "", "Filter date > value")
	fs.StringVar(&lastModified, 0, "last-modified", "", "Filter by exact last-modified date")
	fs.StringVar(&lastModifiedLe, 0, "last-modified-le", "", "Filter last-modified <= value")
	fs.StringVar(&lastModifiedLt, 0, "last-modified-lt", "", "Filter last-modified < value")
	fs.StringVar(&lastModifiedGe, 0, "last-modified-ge", "", "Filter last-modified >= value")
	fs.StringVar(&lastModifiedGt, 0, "last-modified-gt", "", "Filter last-modified > value")
	fs.StringVar(&saleNumber, 0, "sale-number", "", "Filter by sale number")
	fs.BoolVar(&settled, 0, "settled", "Filter to settled sales (default: include unsettled)")
	fs.Int64Var(&contactID, 0, "contact-id", 0, "Filter by customer contact id")
	return &ff.Command{
		Name:      "list",
		Usage:     "fiken sales list --company <slug>",
		ShortHelp: "List sales for a company.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			in := ops.SalesListIn{
				Company:        *sf.flagCompany,
				Page:           page,
				PageSize:       pageSize,
				Date:           ops.Date(date),
				DateLe:         ops.Date(dateLe),
				DateLt:         ops.Date(dateLt),
				DateGe:         ops.Date(dateGe),
				DateGt:         ops.Date(dateGt),
				LastModified:   ops.Date(lastModified),
				LastModifiedLe: ops.Date(lastModifiedLe),
				LastModifiedLt: ops.Date(lastModifiedLt),
				LastModifiedGe: ops.Date(lastModifiedGe),
				LastModifiedGt: ops.Date(lastModifiedGt),
				SaleNumber:     saleNumber,
				ContactID:      contactID,
			}
			if settledSet {
				in.Settled = &settled
			}
			res := Client(ctx).SalesList(ctx, in)
			return Renderer(ctx).Render(res)
		},
	}
}

// salesGetCmd wires `fiken sales get ...`.
func salesGetCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("get")
	fs.SetParent(parent)
	var saleID int64
	fs.Int64Var(&saleID, 0, "sale-id", 0, "Sale id (required)")
	return &ff.Command{
		Name:      "get",
		Usage:     "fiken sales get --company <slug> --sale-id <id>",
		ShortHelp: "Get a single sale by id.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).SalesGet(ctx, ops.SalesGetIn{
				Company: *sf.flagCompany,
				SaleID:  saleID,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// salesCreateCmd wires `fiken sales create ...`.
func salesCreateCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("create")
	fs.SetParent(parent)
	var fromFile string
	fs.StringVar(&fromFile, 0, "from-file", "", "Path to JSON body (required)")
	return &ff.Command{
		Name:      "create",
		Usage:     "fiken sales create --company <slug> --from-file <path>",
		ShortHelp: "Create a new sale.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			body, e := ReadBodyFile[fiken.SaleRequest](fromFile)
			if e != nil {
				e.Op = ops.OpSalesCreate
				return Renderer(ctx).Render(ops.Err[ops.SalesCreateOut](e))
			}
			res := Client(ctx).SalesCreate(ctx, ops.SalesCreateIn{
				Company: *sf.flagCompany,
				Body:    body,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// salesDeleteCmd wires `fiken sales delete ...` (stub).
func salesDeleteCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("delete")
	fs.SetParent(parent)
	var (
		saleID      int64
		description string
	)
	fs.Int64Var(&saleID, 0, "sale-id", 0, "Sale id to delete (required)")
	fs.StringVar(&description, 0, "description", "", "Audit description (required upstream)")
	return &ff.Command{
		Name:      "delete",
		Usage:     "fiken sales delete --company <slug> --sale-id <id> --description <text>",
		ShortHelp: "Soft-archive a sale with an audit description.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).SalesDelete(ctx, ops.SalesDeleteIn{
				Company:     *sf.flagCompany,
				SaleID:      saleID,
				Description: description,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// salesSettleCmd wires `fiken sales settle ...` (stub).
func salesSettleCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("settle")
	fs.SetParent(parent)
	var (
		saleID      int64
		settledDate string
	)
	fs.Int64Var(&saleID, 0, "sale-id", 0, "Sale id to settle (required)")
	fs.StringVar(&settledDate, 0, "settled-date", "", "Settlement date (YYYY-MM-DD, required)")
	return &ff.Command{
		Name:      "settle",
		Usage:     "fiken sales settle --company <slug> --sale-id <id> --settled-date <YYYY-MM-DD>",
		ShortHelp: "Mark a sale settled without payment.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).SalesSettle(ctx, ops.SalesSettleIn{
				Company:     *sf.flagCompany,
				SaleID:      saleID,
				SettledDate: ops.Date(settledDate),
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// salesWriteOffCmd wires `fiken sales write-off ...`.
func salesWriteOffCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("write-off")
	fs.SetParent(parent)
	var (
		saleID   int64
		fromFile string
	)
	fs.Int64Var(&saleID, 0, "sale-id", 0, "Sale id to write off (required)")
	fs.StringVar(&fromFile, 0, "from-file", "", "Path to JSON body (required)")
	return &ff.Command{
		Name:      "write-off",
		Usage:     "fiken sales write-off --company <slug> --sale-id <id> --from-file <path>",
		ShortHelp: "Register a write-off (tapsføring) for a sale.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			body, e := ReadBodyFile[fiken.WriteOffRequest](fromFile)
			if e != nil {
				e.Op = ops.OpSalesWriteOff
				return Renderer(ctx).Render(ops.Err[ops.SaleOut](e))
			}
			res := Client(ctx).SalesWriteOff(ctx, ops.SalesWriteOffIn{
				Company: *sf.flagCompany,
				SaleID:  saleID,
				Body:    body,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// salesPaymentsCmd wires `fiken sales payments {list,get,create}`.
// list + get are fully wired against read paths; create surfaces the
// stub error.
func salesPaymentsCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	set := ff.NewFlagSet("payments")
	set.SetParent(parent)

	cmd := &ff.Command{
		Name:      "payments",
		Usage:     "fiken sales payments <subcommand>",
		ShortHelp: "List, fetch, or register payments for a sale.",
		Flags:     set,
	}

	cmd.Subcommands = append(
		cmd.Subcommands,
		salesPaymentsListCmd(set, stdout, stderr, sf),
		salesPaymentsGetCmd(set, stdout, stderr, sf),
		salesPaymentsCreateCmd(set, stdout, stderr, sf),
	)
	return cmd
}

// salesPaymentsListCmd wires `fiken sales payments list ...`.
func salesPaymentsListCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("list")
	fs.SetParent(parent)
	var saleID int64
	fs.Int64Var(&saleID, 0, "sale-id", 0, "Sale id (required)")
	return &ff.Command{
		Name:      "list",
		Usage:     "fiken sales payments list --company <slug> --sale-id <id>",
		ShortHelp: "List payments for a sale.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).SalesPaymentsList(ctx, ops.SalesPaymentsListIn{
				Company: *sf.flagCompany,
				SaleID:  saleID,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// salesPaymentsGetCmd wires `fiken sales payments get ...`.
func salesPaymentsGetCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("get")
	fs.SetParent(parent)
	var (
		saleID    int64
		paymentID int64
	)
	fs.Int64Var(&saleID, 0, "sale-id", 0, "Sale id (required)")
	fs.Int64Var(&paymentID, 0, "payment-id", 0, "Payment id (required)")
	return &ff.Command{
		Name:      "get",
		Usage:     "fiken sales payments get --company <slug> --sale-id <id> --payment-id <id>",
		ShortHelp: "Get a single payment by id.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).SalesPaymentsGet(ctx, ops.SalesPaymentsGetIn{
				Company:   *sf.flagCompany,
				SaleID:    saleID,
				PaymentID: paymentID,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// salesPaymentsCreateCmd wires `fiken sales payments create ...`.
func salesPaymentsCreateCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("create")
	fs.SetParent(parent)
	var (
		saleID   int64
		fromFile string
	)
	fs.Int64Var(&saleID, 0, "sale-id", 0, "Sale id (required)")
	fs.StringVar(&fromFile, 0, "from-file", "", "Path to JSON body (required)")
	return &ff.Command{
		Name:      "create",
		Usage:     "fiken sales payments create --company <slug> --sale-id <id> --from-file <path>",
		ShortHelp: "Register a payment for a sale.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			body, e := ReadBodyFile[fiken.Payment](fromFile)
			if e != nil {
				e.Op = ops.OpSalesPaymentsCreate
				return Renderer(ctx).Render(ops.Err[ops.SalesPaymentsCreateOut](e))
			}
			res := Client(ctx).SalesPaymentsCreate(ctx, ops.SalesPaymentsCreateIn{
				Company: *sf.flagCompany,
				SaleID:  saleID,
				Body:    body,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// salesAttachmentsCmd wires `fiken sales attachments {list,attach}`.
// list is fully wired; attach surfaces the Plan-D-deferred stub error.
func salesAttachmentsCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	set := ff.NewFlagSet("attachments")
	set.SetParent(parent)

	listFS := ff.NewFlagSet("list")
	listFS.SetParent(set)
	var listSaleID int64
	listFS.Int64Var(&listSaleID, 0, "sale-id", 0, "Sale id (required)")
	listCmd := &ff.Command{
		Name:      "list",
		Usage:     "fiken sales attachments list --company <slug> --sale-id <id>",
		ShortHelp: "List attachments for a sale.",
		Flags:     listFS,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).SalesAttachmentsList(ctx, ops.SalesAttachmentsListIn{
				Company: *sf.flagCompany,
				SaleID:  listSaleID,
			})
			return Renderer(ctx).Render(res)
		},
	}

	attachFS := ff.NewFlagSet("attach")
	attachFS.SetParent(set)
	var (
		attachSaleID    int64
		attachFilename  string
		attachFilePath  string
		attachToSale    bool
		attachToPayment bool
	)
	attachFS.Int64Var(&attachSaleID, 0, "sale-id", 0, "Sale id (required)")
	attachFS.StringVar(&attachFilename, 0, "filename", "", "Override filename (defaults to basename of --file)")
	attachFS.StringVar(&attachFilePath, 0, "file", "", "Local file path to upload (required)")
	attachFS.BoolVar(&attachToSale, 0, "attach-to-sale", "Mark this attachment as documenting the sale")
	attachFS.BoolVar(&attachToPayment, 0, "attach-to-payment", "Mark this attachment as documenting the payment")
	attachCmd := &ff.Command{
		Name:      "attach",
		Usage:     "fiken sales attachments attach --company <slug> --sale-id <id> --file <path>",
		ShortHelp: "Attach a file to a sale.",
		Flags:     attachFS,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).SalesAttach(ctx, ops.SalesAttachIn{
				Company:         *sf.flagCompany,
				SaleID:          attachSaleID,
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
		Usage:       "fiken sales attachments <subcommand>",
		ShortHelp:   "List or attach files to a sale.",
		Flags:       set,
		Subcommands: []*ff.Command{listCmd, attachCmd},
	}
}
