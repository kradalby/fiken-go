package cli

import (
	"context"
	"io"

	"github.com/peterbourgon/ff/v4"

	"github.com/kradalby/fiken-go/fiken"
	"github.com/kradalby/fiken-go/ops"
)

// AddOffers wires
// `fiken offers {list,get,send,counter,drafts}`.
//
// list + get are fully wired against the upstream read paths. Mutating
// ops (send, counter create, drafts CRUD, create-from-draft, attach)
// register as stubs that surface CodeInternal so CLI help and MCP tool
// discovery stay complete. Mutating wiring lands alongside the broader
// Plan D mutation pass.
func AddOffers(root *ff.Command, stdout, stderr io.Writer, sf *sessionFactory) error {
	_ = stderr
	set := ff.NewFlagSet("offers")
	if parentFS, ok := root.Flags.(*ff.FlagSet); ok {
		set.SetParent(parentFS)
	}

	cmd := &ff.Command{
		Name:      "offers",
		Usage:     "fiken offers <subcommand>",
		ShortHelp: "List, fetch, send, or manage offers.",
		Flags:     set,
	}

	cmd.Subcommands = append(
		cmd.Subcommands,
		offersListCmd(set, stdout, stderr, sf),
		offersGetCmd(set, stdout, stderr, sf),
		offersSendCmd(set, stdout, stderr, sf),
		offersCounterCmd(set, stdout, stderr, sf),
		offersDraftsCmd(set, stdout, stderr, sf),
	)

	root.Subcommands = append(root.Subcommands, cmd)
	return nil
}

// offersListCmd wires `fiken offers list ...`.
func offersListCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
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
		Usage:     "fiken offers list --company <slug>",
		ShortHelp: "List offers for a company.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).OffersList(ctx, ops.OffersListIn{
				Company:  *sf.flagCompany,
				Page:     page,
				PageSize: pageSize,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// offersGetCmd wires `fiken offers get ...`. The upstream path-param
// is a string (not int) so the flag stays --offer-id without int
// parsing.
func offersGetCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("get")
	fs.SetParent(parent)
	var offerID string
	fs.StringVar(&offerID, 0, "offer-id", "", "Offer id (required)")
	return &ff.Command{
		Name:      "get",
		Usage:     "fiken offers get --company <slug> --offer-id <id>",
		ShortHelp: "Get a single offer by id.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).OffersGet(ctx, ops.OffersGetIn{
				Company: *sf.flagCompany,
				OfferID: offerID,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// offersSendCmd wires `fiken offers send ...`. The full send payload
// (recipients, method, message, attachments toggle) is supplied via
// --from-file as JSON matching SendOfferRequest.
func offersSendCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("send")
	fs.SetParent(parent)
	var fromFile string
	fs.StringVar(&fromFile, 0, "from-file", "", "Path to JSON body (required)")
	return &ff.Command{
		Name:      "send",
		Usage:     "fiken offers send --company <slug> --from-file <path>",
		ShortHelp: "Send an offer to its customer.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			body, e := ReadBodyFile[fiken.SendOfferRequest](fromFile)
			if e != nil {
				e.Op = ops.OpOffersSend
				return Renderer(ctx).Render(ops.Err[ops.OffersSendOut](e))
			}
			res := Client(ctx).OffersSend(ctx, ops.OffersSendIn{
				Company: *sf.flagCompany,
				Body:    body,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// offersCounterCmd wires `fiken offers counter create ...` (stub).
func offersCounterCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	set := ff.NewFlagSet("counter")
	set.SetParent(parent)

	createFS := ff.NewFlagSet("create")
	createFS.SetParent(set)
	var value int64
	createFS.Int64Var(&value, 0, "value", 0, "Starting counter value (required)")
	createCmd := &ff.Command{
		Name:      "create",
		Usage:     "fiken offers counter create --company <slug> --value <n>",
		ShortHelp: "Initialize the offer counter for the fiscal year.",
		Flags:     createFS,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			// value clamps to int32 bounds; mirror the invoices template.
			var v32 int32
			switch {
			case value > 2147483647:
				v32 = 2147483647
			case value < -2147483648:
				v32 = -2147483648
			default:
				v32 = int32(value)
			}
			res := Client(ctx).OffersCounterCreate(ctx, ops.OffersCounterCreateIn{
				Company: *sf.flagCompany, Value: v32,
			})
			return Renderer(ctx).Render(res)
		},
	}

	getFS := ff.NewFlagSet("get")
	getFS.SetParent(set)
	getCmd := &ff.Command{
		Name:      "get",
		Usage:     "fiken offers counter get --company <slug>",
		ShortHelp: "Get the current offer counter value.",
		Flags:     getFS,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).OffersCounterGet(ctx, ops.CounterGetIn{
				Company: *sf.flagCompany,
			})
			return Renderer(ctx).Render(res)
		},
	}

	return &ff.Command{
		Name:        "counter",
		Usage:       "fiken offers counter <subcommand>",
		ShortHelp:   "Manage the offer counter.",
		Flags:       set,
		Subcommands: []*ff.Command{createCmd, getCmd},
	}
}

// offersDraftsCmd wires
// `fiken offers drafts {list,get,create,update,delete,create-from,attachments}`.
//
// list, get, and attachments list are fully wired against read paths;
// the rest are CLI/MCP stubs.
func offersDraftsCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	set := ff.NewFlagSet("drafts")
	set.SetParent(parent)

	cmd := &ff.Command{
		Name:      "drafts",
		Usage:     "fiken offers drafts <subcommand>",
		ShortHelp: "List, fetch, or manage offer drafts.",
		Flags:     set,
	}

	cmd.Subcommands = append(
		cmd.Subcommands,
		offersDraftsListCmd(set, stdout, stderr, sf),
		offersDraftsGetCmd(set, stdout, stderr, sf),
		offersDraftsCreateCmd(set, stdout, stderr, sf),
		offersDraftsUpdateCmd(set, stdout, stderr, sf),
		offersDraftsDeleteCmd(set, stdout, stderr, sf),
		offersDraftsCreateFromCmd(set, stdout, stderr, sf),
		offersDraftsAttachmentsCmd(set, stdout, stderr, sf),
	)
	return cmd
}

// offersDraftsListCmd wires `fiken offers drafts list ...`.
func offersDraftsListCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
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
		Usage:     "fiken offers drafts list --company <slug>",
		ShortHelp: "List offer drafts for a company.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).OfferDraftsList(ctx, ops.OfferDraftsListIn{
				Company:  *sf.flagCompany,
				Page:     page,
				PageSize: pageSize,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// offersDraftsGetCmd wires `fiken offers drafts get ...`.
func offersDraftsGetCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("get")
	fs.SetParent(parent)
	var draftID int64
	fs.Int64Var(&draftID, 0, "draft-id", 0, "Draft id (required)")
	return &ff.Command{
		Name:      "get",
		Usage:     "fiken offers drafts get --company <slug> --draft-id <id>",
		ShortHelp: "Get a single offer draft by id.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).OfferDraftsGet(ctx, ops.OfferDraftsGetIn{
				Company: *sf.flagCompany,
				DraftID: draftID,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// offersDraftsCreateCmd wires `fiken offers drafts create ...`.
func offersDraftsCreateCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("create")
	fs.SetParent(parent)
	var fromFile string
	fs.StringVar(&fromFile, 0, "from-file", "", "Path to JSON body (required)")
	return &ff.Command{
		Name:      "create",
		Usage:     "fiken offers drafts create --company <slug> --from-file <path>",
		ShortHelp: "Create a new offer draft.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			body, e := ReadBodyFile[fiken.InvoiceishDraftRequest](fromFile)
			if e != nil {
				e.Op = ops.OpOffersDraftsCreate
				return Renderer(ctx).Render(ops.Err[ops.InvoiceDraftOut](e))
			}
			res := Client(ctx).OfferDraftsCreate(ctx, ops.OfferDraftsCreateIn{
				Company: *sf.flagCompany,
				Body:    body,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// offersDraftsUpdateCmd wires `fiken offers drafts update ...`.
func offersDraftsUpdateCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
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
		Usage:     "fiken offers drafts update --company <slug> --draft-id <id> --from-file <path>",
		ShortHelp: "Update an existing offer draft.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			body, e := ReadBodyFile[fiken.InvoiceishDraftRequest](fromFile)
			if e != nil {
				e.Op = ops.OpOffersDraftsUpdate
				return Renderer(ctx).Render(ops.Err[ops.InvoiceDraftOut](e))
			}
			res := Client(ctx).OfferDraftsUpdate(ctx, ops.OfferDraftsUpdateIn{
				Company: *sf.flagCompany,
				DraftID: draftID,
				Body:    body,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// offersDraftsDeleteCmd wires `fiken offers drafts delete ...` (stub).
func offersDraftsDeleteCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("delete")
	fs.SetParent(parent)
	var draftID int64
	fs.Int64Var(&draftID, 0, "draft-id", 0, "Draft id to delete (required)")
	return &ff.Command{
		Name:      "delete",
		Usage:     "fiken offers drafts delete --company <slug> --draft-id <id>",
		ShortHelp: "Delete an offer draft.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).OfferDraftsDelete(ctx, ops.OfferDraftsDeleteIn{
				Company: *sf.flagCompany, DraftID: draftID,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// offersDraftsCreateFromCmd wires `fiken offers drafts create-from ...`
// (stub) — promotes a draft to a posted offer upstream.
func offersDraftsCreateFromCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("create-from")
	fs.SetParent(parent)
	var draftID int64
	fs.Int64Var(&draftID, 0, "draft-id", 0, "Source draft id (required)")
	return &ff.Command{
		Name:      "create-from",
		Usage:     "fiken offers drafts create-from --company <slug> --draft-id <id>",
		ShortHelp: "Promote a draft to a posted offer.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).OfferDraftsCreateFrom(ctx, ops.OfferDraftsCreateFromIn{
				Company: *sf.flagCompany, DraftID: draftID,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// offersDraftsAttachmentsCmd wires
// `fiken offers drafts attachments {list,attach}`. list is fully
// wired; attach surfaces the Plan-D-deferred stub error.
func offersDraftsAttachmentsCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	set := ff.NewFlagSet("attachments")
	set.SetParent(parent)

	listFS := ff.NewFlagSet("list")
	listFS.SetParent(set)
	var listDraftID int64
	listFS.Int64Var(&listDraftID, 0, "draft-id", 0, "Draft id (required)")
	listCmd := &ff.Command{
		Name:      "list",
		Usage:     "fiken offers drafts attachments list --company <slug> --draft-id <id>",
		ShortHelp: "List attachments for an offer draft.",
		Flags:     listFS,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).OfferDraftsAttachmentsList(ctx, ops.OfferDraftsAttachmentsListIn{
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
		Usage:     "fiken offers drafts attachments attach --company <slug> --draft-id <id> --file <path>",
		ShortHelp: "Attach a file to an offer draft.",
		Flags:     attachFS,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).OfferDraftsAttachmentsAttach(ctx, ops.OfferDraftsAttachmentsAttachIn{
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
		Usage:       "fiken offers drafts attachments <subcommand>",
		ShortHelp:   "List or attach files to an offer draft.",
		Flags:       set,
		Subcommands: []*ff.Command{listCmd, attachCmd},
	}
}
