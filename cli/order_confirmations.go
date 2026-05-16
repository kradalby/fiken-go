package cli

import (
	"context"
	"io"

	"github.com/peterbourgon/ff/v4"

	"github.com/kradalby/fiken-go/fiken"
	"github.com/kradalby/fiken-go/ops"
)

// AddOrderConfirmations wires
// `fiken order-confirmations {list,get,counter,create-invoice-draft,drafts}`.
//
// list + get are fully wired against the upstream read paths. Mutating
// ops (counter create, create-invoice-draft, drafts CRUD, draft
// create-from, attach) register as stubs that surface CodeInternal so
// CLI help and MCP tool discovery stay complete. Mutating wiring lands
// alongside the broader Plan D mutation pass.
func AddOrderConfirmations(root *ff.Command, stdout, stderr io.Writer, sf *sessionFactory) error {
	_ = stderr
	set := ff.NewFlagSet("order-confirmations")
	if parentFS, ok := root.Flags.(*ff.FlagSet); ok {
		set.SetParent(parentFS)
	}

	cmd := &ff.Command{
		Name:      "order-confirmations",
		Usage:     "fiken order-confirmations <subcommand>",
		ShortHelp: "List, fetch, or manage order confirmations.",
		Flags:     set,
	}

	cmd.Subcommands = append(
		cmd.Subcommands,
		orderConfirmationsListCmd(set, stdout, stderr, sf),
		orderConfirmationsGetCmd(set, stdout, stderr, sf),
		orderConfirmationsCounterCmd(set, stdout, stderr, sf),
		orderConfirmationsCreateInvoiceDraftCmd(set, stdout, stderr, sf),
		orderConfirmationsDraftsCmd(set, stdout, stderr, sf),
	)

	root.Subcommands = append(root.Subcommands, cmd)
	return nil
}

// orderConfirmationsListCmd wires `fiken order-confirmations list ...`.
func orderConfirmationsListCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("list")
	fs.SetParent(parent)
	var (
		page     int
		pageSize int
	)
	fs.IntVar(&page, 0, "page", 0, "Page number (0-indexed)")
	fs.IntVar(&pageSize, 0, "page-size", 0, "Page size")
	return &ff.Command{
		Name:      "list",
		Usage:     "fiken order-confirmations list --company <slug>",
		ShortHelp: "List order confirmations for a company.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).OrderConfirmationsList(ctx, ops.OrderConfirmationsListIn{
				Company:  *sf.flagCompany,
				Page:     page,
				PageSize: pageSize,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// orderConfirmationsGetCmd wires `fiken order-confirmations get ...`. The
// upstream path-param is a string (not int) so the flag stays
// --confirmation-id without int parsing.
func orderConfirmationsGetCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("get")
	fs.SetParent(parent)
	var confirmationID string
	fs.StringVar(&confirmationID, 0, "confirmation-id", "", "Order confirmation id (required)")
	return &ff.Command{
		Name:      "get",
		Usage:     "fiken order-confirmations get --company <slug> --confirmation-id <id>",
		ShortHelp: "Get a single order confirmation by id.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).OrderConfirmationsGet(ctx, ops.OrderConfirmationsGetIn{
				Company:        *sf.flagCompany,
				ConfirmationID: confirmationID,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// orderConfirmationsCounterCmd wires
// `fiken order-confirmations counter create ...` (stub).
func orderConfirmationsCounterCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	set := ff.NewFlagSet("counter")
	set.SetParent(parent)

	createFS := ff.NewFlagSet("create")
	createFS.SetParent(set)
	var value int64
	createFS.Int64Var(&value, 0, "value", 0, "Starting counter value (required)")
	createCmd := &ff.Command{
		Name:      "create",
		Usage:     "fiken order-confirmations counter create --company <slug> --value <n>",
		ShortHelp: "Initialize the order confirmation counter for the fiscal year.",
		Flags:     createFS,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			// value clamps to int32 bounds; mirror the offers template.
			var v32 int32
			switch {
			case value > 2147483647:
				v32 = 2147483647
			case value < -2147483648:
				v32 = -2147483648
			default:
				v32 = int32(value)
			}
			res := Client(ctx).OrderConfirmationsCounterCreate(ctx, ops.OrderConfirmationsCounterCreateIn{
				Company: *sf.flagCompany, Value: v32,
			})
			return Renderer(ctx).Render(res)
		},
	}

	getFS := ff.NewFlagSet("get")
	getFS.SetParent(set)
	getCmd := &ff.Command{
		Name:      "get",
		Usage:     "fiken order-confirmations counter get --company <slug>",
		ShortHelp: "Get the current order-confirmation counter value.",
		Flags:     getFS,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).OrderConfirmationsCounterGet(ctx, ops.CounterGetIn{
				Company: *sf.flagCompany,
			})
			return Renderer(ctx).Render(res)
		},
	}

	return &ff.Command{
		Name:        "counter",
		Usage:       "fiken order-confirmations counter <subcommand>",
		ShortHelp:   "Manage the order confirmation counter.",
		Flags:       set,
		Subcommands: []*ff.Command{createCmd, getCmd},
	}
}

// orderConfirmationsCreateInvoiceDraftCmd wires
// `fiken order-confirmations create-invoice-draft ...` — promotes a
// posted order confirmation to an invoice draft upstream.
func orderConfirmationsCreateInvoiceDraftCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("create-invoice-draft")
	fs.SetParent(parent)
	var confirmationID string
	fs.StringVar(&confirmationID, 0, "confirmation-id", "", "Source order confirmation id (required)")
	return &ff.Command{
		Name:      "create-invoice-draft",
		Usage:     "fiken order-confirmations create-invoice-draft --company <slug> --confirmation-id <id>",
		ShortHelp: "Create an invoice draft from an order confirmation.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).OrderConfirmationsCreateInvoiceDraft(ctx, ops.OrderConfirmationsCreateInvoiceDraftIn{
				Company: *sf.flagCompany, ConfirmationID: confirmationID,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// orderConfirmationsDraftsCmd wires
// `fiken order-confirmations drafts {list,get,create,update,delete,create-from,attachments}`.
//
// list, get, and attachments list are fully wired against read paths;
// the rest are CLI/MCP stubs.
func orderConfirmationsDraftsCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	set := ff.NewFlagSet("drafts")
	set.SetParent(parent)

	cmd := &ff.Command{
		Name:      "drafts",
		Usage:     "fiken order-confirmations drafts <subcommand>",
		ShortHelp: "List, fetch, or manage order confirmation drafts.",
		Flags:     set,
	}

	cmd.Subcommands = append(
		cmd.Subcommands,
		orderConfirmationsDraftsListCmd(set, stdout, stderr, sf),
		orderConfirmationsDraftsGetCmd(set, stdout, stderr, sf),
		orderConfirmationsDraftsCreateCmd(set, stdout, stderr, sf),
		orderConfirmationsDraftsUpdateCmd(set, stdout, stderr, sf),
		orderConfirmationsDraftsDeleteCmd(set, stdout, stderr, sf),
		orderConfirmationsDraftsCreateFromCmd(set, stdout, stderr, sf),
		orderConfirmationsDraftsAttachmentsCmd(set, stdout, stderr, sf),
	)
	return cmd
}

// orderConfirmationsDraftsListCmd wires
// `fiken order-confirmations drafts list ...`.
func orderConfirmationsDraftsListCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("list")
	fs.SetParent(parent)
	var (
		page     int
		pageSize int
	)
	fs.IntVar(&page, 0, "page", 0, "Page number (0-indexed)")
	fs.IntVar(&pageSize, 0, "page-size", 0, "Page size")
	return &ff.Command{
		Name:      "list",
		Usage:     "fiken order-confirmations drafts list --company <slug>",
		ShortHelp: "List order confirmation drafts for a company.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).OrderConfirmationDraftsList(ctx, ops.OrderConfirmationDraftsListIn{
				Company:  *sf.flagCompany,
				Page:     page,
				PageSize: pageSize,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// orderConfirmationsDraftsGetCmd wires
// `fiken order-confirmations drafts get ...`.
func orderConfirmationsDraftsGetCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("get")
	fs.SetParent(parent)
	var draftID int64
	fs.Int64Var(&draftID, 0, "draft-id", 0, "Draft id (required)")
	return &ff.Command{
		Name:      "get",
		Usage:     "fiken order-confirmations drafts get --company <slug> --draft-id <id>",
		ShortHelp: "Get a single order confirmation draft by id.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).OrderConfirmationDraftsGet(ctx, ops.OrderConfirmationDraftsGetIn{
				Company: *sf.flagCompany,
				DraftID: draftID,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// orderConfirmationsDraftsCreateCmd wires
// `fiken order-confirmations drafts create ...`.
func orderConfirmationsDraftsCreateCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("create")
	fs.SetParent(parent)
	var fromFile string
	fs.StringVar(&fromFile, 0, "from-file", "", "Path to JSON body (required)")
	return &ff.Command{
		Name:      "create",
		Usage:     "fiken order-confirmations drafts create --company <slug> --from-file <path>",
		ShortHelp: "Create a new order confirmation draft.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			body, e := ReadBodyFile[fiken.InvoiceishDraftRequest](fromFile)
			if e != nil {
				e.Op = ops.OpOrderConfirmationsDraftsCreate
				return Renderer(ctx).Render(ops.Err[ops.InvoiceDraftOut](e))
			}
			res := Client(ctx).OrderConfirmationDraftsCreate(ctx, ops.OrderConfirmationDraftsCreateIn{
				Company: *sf.flagCompany,
				Body:    body,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// orderConfirmationsDraftsUpdateCmd wires
// `fiken order-confirmations drafts update ...`.
func orderConfirmationsDraftsUpdateCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
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
		Usage:     "fiken order-confirmations drafts update --company <slug> --draft-id <id> --from-file <path>",
		ShortHelp: "Update an existing order confirmation draft.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			body, e := ReadBodyFile[fiken.InvoiceishDraftRequest](fromFile)
			if e != nil {
				e.Op = ops.OpOrderConfirmationsDraftsUpdate
				return Renderer(ctx).Render(ops.Err[ops.InvoiceDraftOut](e))
			}
			res := Client(ctx).OrderConfirmationDraftsUpdate(ctx, ops.OrderConfirmationDraftsUpdateIn{
				Company: *sf.flagCompany,
				DraftID: draftID,
				Body:    body,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// orderConfirmationsDraftsDeleteCmd wires
// `fiken order-confirmations drafts delete ...` (stub).
func orderConfirmationsDraftsDeleteCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("delete")
	fs.SetParent(parent)
	var draftID int64
	fs.Int64Var(&draftID, 0, "draft-id", 0, "Draft id to delete (required)")
	return &ff.Command{
		Name:      "delete",
		Usage:     "fiken order-confirmations drafts delete --company <slug> --draft-id <id>",
		ShortHelp: "Delete an order confirmation draft.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).OrderConfirmationDraftsDelete(ctx, ops.OrderConfirmationDraftsDeleteIn{
				Company: *sf.flagCompany, DraftID: draftID,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// orderConfirmationsDraftsCreateFromCmd wires
// `fiken order-confirmations drafts create-from ...` (stub) — promotes
// a draft to a posted order confirmation upstream.
func orderConfirmationsDraftsCreateFromCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("create-from")
	fs.SetParent(parent)
	var draftID int64
	fs.Int64Var(&draftID, 0, "draft-id", 0, "Source draft id (required)")
	return &ff.Command{
		Name:      "create-from",
		Usage:     "fiken order-confirmations drafts create-from --company <slug> --draft-id <id>",
		ShortHelp: "Promote a draft to a posted order confirmation.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).OrderConfirmationDraftsCreateFrom(ctx, ops.OrderConfirmationDraftsCreateFromIn{
				Company: *sf.flagCompany, DraftID: draftID,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// orderConfirmationsDraftsAttachmentsCmd wires
// `fiken order-confirmations drafts attachments {list,attach}`. list is
// fully wired; attach surfaces the Plan-D-deferred stub error.
func orderConfirmationsDraftsAttachmentsCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	set := ff.NewFlagSet("attachments")
	set.SetParent(parent)

	listFS := ff.NewFlagSet("list")
	listFS.SetParent(set)
	var listDraftID int64
	listFS.Int64Var(&listDraftID, 0, "draft-id", 0, "Draft id (required)")
	listCmd := &ff.Command{
		Name:      "list",
		Usage:     "fiken order-confirmations drafts attachments list --company <slug> --draft-id <id>",
		ShortHelp: "List attachments for an order confirmation draft.",
		Flags:     listFS,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).OrderConfirmationDraftsAttachmentsList(ctx, ops.OrderConfirmationDraftsAttachmentsListIn{
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
		Usage:     "fiken order-confirmations drafts attachments attach --company <slug> --draft-id <id> --file <path>",
		ShortHelp: "Attach a file to an order confirmation draft.",
		Flags:     attachFS,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).OrderConfirmationDraftsAttachmentsAttach(ctx, ops.OrderConfirmationDraftsAttachmentsAttachIn{
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
		Usage:       "fiken order-confirmations drafts attachments <subcommand>",
		ShortHelp:   "List or attach files to an order confirmation draft.",
		Flags:       set,
		Subcommands: []*ff.Command{listCmd, attachCmd},
	}
}
