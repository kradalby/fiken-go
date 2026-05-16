package cli

import (
	"context"
	"io"

	"github.com/peterbourgon/ff/v4"

	"github.com/kradalby/fiken-go/fiken"
	"github.com/kradalby/fiken-go/ops"
)

// AddCreditNotes wires
// `fiken credit-notes {list,get,send,counter,drafts,full-create,partial-create}`.
//
// list + get are fully wired against the upstream read paths. Mutating
// ops (send, counter create, full/partial create, drafts CRUD,
// create-from-draft, attach) register as stubs that surface
// CodeInternal so CLI help and MCP tool discovery stay complete.
// Mutating wiring lands alongside the broader Plan D mutation pass.
func AddCreditNotes(root *ff.Command, stdout, stderr io.Writer, sf *sessionFactory) error {
	_ = stderr
	set := ff.NewFlagSet("credit-notes")
	if parentFS, ok := root.Flags.(*ff.FlagSet); ok {
		set.SetParent(parentFS)
	}

	cmd := &ff.Command{
		Name:      "credit-notes",
		Usage:     "fiken credit-notes <subcommand>",
		ShortHelp: "List, fetch, send, or manage credit notes.",
		Flags:     set,
	}

	cmd.Subcommands = append(
		cmd.Subcommands,
		creditNotesListCmd(set, stdout, stderr, sf),
		creditNotesGetCmd(set, stdout, stderr, sf),
		creditNotesSendCmd(set, stdout, stderr, sf),
		creditNotesCounterCmd(set, stdout, stderr, sf),
		creditNotesFullCreateCmd(set, stdout, stderr, sf),
		creditNotesPartialCreateCmd(set, stdout, stderr, sf),
		creditNotesDraftsCmd(set, stdout, stderr, sf),
	)

	root.Subcommands = append(root.Subcommands, cmd)
	return nil
}

// creditNotesListCmd wires `fiken credit-notes list ...`.
func creditNotesListCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("list")
	fs.SetParent(parent)
	var (
		page                int
		pageSize            int
		issueDate           string
		issueDateLe         string
		issueDateLt         string
		issueDateGe         string
		issueDateGt         string
		lastModified        string
		lastModifiedLe      string
		lastModifiedLt      string
		lastModifiedGe      string
		lastModifiedGt      string
		customerID          int64
		settled             bool
		settledSet          bool
		creditNoteDraftUUID string
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
	fs.Int64Var(&customerID, 0, "customer-id", 0, "Filter by customer contact id")
	fs.BoolVar(&settled, 0, "settled", "Filter to settled credit notes (default: include unsettled)")
	fs.StringVar(&creditNoteDraftUUID, 0, "credit-note-draft-uuid", "", "Filter by source draft UUID")

	return &ff.Command{
		Name:      "list",
		Usage:     "fiken credit-notes list --company <slug>",
		ShortHelp: "List credit notes for a company.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			in := ops.CreditNotesListIn{
				Company:             *sf.flagCompany,
				Page:                page,
				PageSize:            pageSize,
				IssueDate:           ops.Date(issueDate),
				IssueDateLe:         ops.Date(issueDateLe),
				IssueDateLt:         ops.Date(issueDateLt),
				IssueDateGe:         ops.Date(issueDateGe),
				IssueDateGt:         ops.Date(issueDateGt),
				LastModified:        ops.Date(lastModified),
				LastModifiedLe:      ops.Date(lastModifiedLe),
				LastModifiedLt:      ops.Date(lastModifiedLt),
				LastModifiedGe:      ops.Date(lastModifiedGe),
				LastModifiedGt:      ops.Date(lastModifiedGt),
				CustomerID:          customerID,
				CreditNoteDraftUUID: creditNoteDraftUUID,
			}
			if settledSet {
				in.Settled = &settled
			}
			res := Client(ctx).CreditNotesList(ctx, in)
			return Renderer(ctx).Render(res)
		},
	}
}

// creditNotesGetCmd wires `fiken credit-notes get ...`. The upstream
// path-param is a string (not int) so the flag stays --credit-note-id
// rather than --credit-note-id with int parsing.
func creditNotesGetCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("get")
	fs.SetParent(parent)
	var creditNoteID string
	fs.StringVar(&creditNoteID, 0, "credit-note-id", "", "Credit note id (required)")
	return &ff.Command{
		Name:      "get",
		Usage:     "fiken credit-notes get --company <slug> --credit-note-id <id>",
		ShortHelp: "Get a single credit note by id.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).CreditNotesGet(ctx, ops.CreditNotesGetIn{
				Company:      *sf.flagCompany,
				CreditNoteID: creditNoteID,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// creditNotesSendCmd wires `fiken credit-notes send ...`. The full
// send payload (recipients, method, message, attachments toggle) is
// supplied via --from-file as JSON matching SendCreditNoteRequest.
func creditNotesSendCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("send")
	fs.SetParent(parent)
	var fromFile string
	fs.StringVar(&fromFile, 0, "from-file", "", "Path to JSON body (required)")
	return &ff.Command{
		Name:      "send",
		Usage:     "fiken credit-notes send --company <slug> --from-file <path>",
		ShortHelp: "Send a credit note to its customer.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			body, e := ReadBodyFile[fiken.SendCreditNoteRequest](fromFile)
			if e != nil {
				e.Op = ops.OpCreditNotesSend
				return Renderer(ctx).Render(ops.Err[ops.CreditNotesSendOut](e))
			}
			res := Client(ctx).CreditNotesSend(ctx, ops.CreditNotesSendIn{
				Company: *sf.flagCompany,
				Body:    body,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// creditNotesCounterCmd wires `fiken credit-notes counter create ...` (stub).
func creditNotesCounterCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	set := ff.NewFlagSet("counter")
	set.SetParent(parent)

	createFS := ff.NewFlagSet("create")
	createFS.SetParent(set)
	var value int64
	createFS.Int64Var(&value, 0, "value", 0, "Starting counter value (required)")
	createCmd := &ff.Command{
		Name:      "create",
		Usage:     "fiken credit-notes counter create --company <slug> --value <n>",
		ShortHelp: "Initialize the credit-note counter for the fiscal year.",
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
			res := Client(ctx).CreditNotesCounterCreate(ctx, ops.CreditNotesCounterCreateIn{
				Company: *sf.flagCompany, Value: v32,
			})
			return Renderer(ctx).Render(res)
		},
	}

	getFS := ff.NewFlagSet("get")
	getFS.SetParent(set)
	getCmd := &ff.Command{
		Name:      "get",
		Usage:     "fiken credit-notes counter get --company <slug>",
		ShortHelp: "Get the current credit-note counter value.",
		Flags:     getFS,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).CreditNotesCounterGet(ctx, ops.CounterGetIn{
				Company: *sf.flagCompany,
			})
			return Renderer(ctx).Render(res)
		},
	}

	return &ff.Command{
		Name:        "counter",
		Usage:       "fiken credit-notes counter <subcommand>",
		ShortHelp:   "Manage the credit-note counter.",
		Flags:       set,
		Subcommands: []*ff.Command{createCmd, getCmd},
	}
}

// creditNotesFullCreateCmd wires `fiken credit-notes full-create ...`
// — POST /creditNotes/full credits an invoice in full.
func creditNotesFullCreateCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("full-create")
	fs.SetParent(parent)
	var fromFile string
	fs.StringVar(&fromFile, 0, "from-file", "", "Path to JSON body (required)")
	return &ff.Command{
		Name:      "full-create",
		Usage:     "fiken credit-notes full-create --company <slug> --from-file <path>",
		ShortHelp: "Create a full credit note from an invoice.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			body, e := ReadBodyFile[fiken.FullCreditNoteRequest](fromFile)
			if e != nil {
				e.Op = ops.OpCreditNotesFullCreate
				return Renderer(ctx).Render(ops.Err[ops.CreditNotesFullCreateOut](e))
			}
			res := Client(ctx).CreditNotesFullCreate(ctx, ops.CreditNotesFullCreateIn{
				Company: *sf.flagCompany,
				Body:    body,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// creditNotesPartialCreateCmd wires `fiken credit-notes partial-create ...`
// — POST /creditNotes/partial credits an invoice in part.
func creditNotesPartialCreateCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("partial-create")
	fs.SetParent(parent)
	var fromFile string
	fs.StringVar(&fromFile, 0, "from-file", "", "Path to JSON body (required)")
	return &ff.Command{
		Name:      "partial-create",
		Usage:     "fiken credit-notes partial-create --company <slug> --from-file <path>",
		ShortHelp: "Create a partial credit note.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			body, e := ReadBodyFile[fiken.PartialCreditNoteRequest](fromFile)
			if e != nil {
				e.Op = ops.OpCreditNotesPartialCreate
				return Renderer(ctx).Render(ops.Err[ops.CreditNotesPartialCreateOut](e))
			}
			res := Client(ctx).CreditNotesPartialCreate(ctx, ops.CreditNotesPartialCreateIn{
				Company: *sf.flagCompany,
				Body:    body,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// creditNotesDraftsCmd wires
// `fiken credit-notes drafts {list,get,create,update,delete,create-from,attachments}`.
//
// list, get, and attachments list are fully wired against read paths;
// the rest are CLI/MCP stubs.
func creditNotesDraftsCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	set := ff.NewFlagSet("drafts")
	set.SetParent(parent)

	cmd := &ff.Command{
		Name:      "drafts",
		Usage:     "fiken credit-notes drafts <subcommand>",
		ShortHelp: "List, fetch, or manage credit-note drafts.",
		Flags:     set,
	}

	cmd.Subcommands = append(
		cmd.Subcommands,
		creditNotesDraftsListCmd(set, stdout, stderr, sf),
		creditNotesDraftsGetCmd(set, stdout, stderr, sf),
		creditNotesDraftsCreateCmd(set, stdout, stderr, sf),
		creditNotesDraftsUpdateCmd(set, stdout, stderr, sf),
		creditNotesDraftsDeleteCmd(set, stdout, stderr, sf),
		creditNotesDraftsCreateFromCmd(set, stdout, stderr, sf),
		creditNotesDraftsAttachmentsCmd(set, stdout, stderr, sf),
	)
	return cmd
}

// creditNotesDraftsListCmd wires `fiken credit-notes drafts list ...`.
func creditNotesDraftsListCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
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
		Usage:     "fiken credit-notes drafts list --company <slug>",
		ShortHelp: "List credit-note drafts for a company.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).CreditNoteDraftsList(ctx, ops.CreditNoteDraftsListIn{
				Company:  *sf.flagCompany,
				Page:     page,
				PageSize: pageSize,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// creditNotesDraftsGetCmd wires `fiken credit-notes drafts get ...`.
func creditNotesDraftsGetCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("get")
	fs.SetParent(parent)
	var draftID int64
	fs.Int64Var(&draftID, 0, "draft-id", 0, "Draft id (required)")
	return &ff.Command{
		Name:      "get",
		Usage:     "fiken credit-notes drafts get --company <slug> --draft-id <id>",
		ShortHelp: "Get a single credit-note draft by id.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).CreditNoteDraftsGet(ctx, ops.CreditNoteDraftsGetIn{
				Company: *sf.flagCompany,
				DraftID: draftID,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// creditNotesDraftsCreateCmd wires
// `fiken credit-notes drafts create ...`.
func creditNotesDraftsCreateCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("create")
	fs.SetParent(parent)
	var fromFile string
	fs.StringVar(&fromFile, 0, "from-file", "", "Path to JSON body (required)")
	return &ff.Command{
		Name:      "create",
		Usage:     "fiken credit-notes drafts create --company <slug> --from-file <path>",
		ShortHelp: "Create a new credit-note draft.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			body, e := ReadBodyFile[fiken.InvoiceishDraftRequest](fromFile)
			if e != nil {
				e.Op = ops.OpCreditNotesDraftsCreate
				return Renderer(ctx).Render(ops.Err[ops.InvoiceDraftOut](e))
			}
			res := Client(ctx).CreditNoteDraftsCreate(ctx, ops.CreditNoteDraftsCreateIn{
				Company: *sf.flagCompany,
				Body:    body,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// creditNotesDraftsUpdateCmd wires
// `fiken credit-notes drafts update ...`.
func creditNotesDraftsUpdateCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
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
		Usage:     "fiken credit-notes drafts update --company <slug> --draft-id <id> --from-file <path>",
		ShortHelp: "Update an existing credit-note draft.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			body, e := ReadBodyFile[fiken.InvoiceishDraftRequest](fromFile)
			if e != nil {
				e.Op = ops.OpCreditNotesDraftsUpdate
				return Renderer(ctx).Render(ops.Err[ops.InvoiceDraftOut](e))
			}
			res := Client(ctx).CreditNoteDraftsUpdate(ctx, ops.CreditNoteDraftsUpdateIn{
				Company: *sf.flagCompany,
				DraftID: draftID,
				Body:    body,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// creditNotesDraftsDeleteCmd wires
// `fiken credit-notes drafts delete ...` (stub).
func creditNotesDraftsDeleteCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("delete")
	fs.SetParent(parent)
	var draftID int64
	fs.Int64Var(&draftID, 0, "draft-id", 0, "Draft id to delete (required)")
	return &ff.Command{
		Name:      "delete",
		Usage:     "fiken credit-notes drafts delete --company <slug> --draft-id <id>",
		ShortHelp: "Delete a credit-note draft.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).CreditNoteDraftsDelete(ctx, ops.CreditNoteDraftsDeleteIn{
				Company: *sf.flagCompany, DraftID: draftID,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// creditNotesDraftsCreateFromCmd wires
// `fiken credit-notes drafts create-from ...` (stub) — promotes a
// draft to a posted credit note upstream.
func creditNotesDraftsCreateFromCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("create-from")
	fs.SetParent(parent)
	var draftID int64
	fs.Int64Var(&draftID, 0, "draft-id", 0, "Source draft id (required)")
	return &ff.Command{
		Name:      "create-from",
		Usage:     "fiken credit-notes drafts create-from --company <slug> --draft-id <id>",
		ShortHelp: "Promote a draft to a posted credit note.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).CreditNoteDraftsCreateFrom(ctx, ops.CreditNoteDraftsCreateFromIn{
				Company: *sf.flagCompany, DraftID: draftID,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// creditNotesDraftsAttachmentsCmd wires
// `fiken credit-notes drafts attachments {list,attach}`. list is fully
// wired; attach surfaces the Plan-D-deferred stub error.
func creditNotesDraftsAttachmentsCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	set := ff.NewFlagSet("attachments")
	set.SetParent(parent)

	listFS := ff.NewFlagSet("list")
	listFS.SetParent(set)
	var listDraftID int64
	listFS.Int64Var(&listDraftID, 0, "draft-id", 0, "Draft id (required)")
	listCmd := &ff.Command{
		Name:      "list",
		Usage:     "fiken credit-notes drafts attachments list --company <slug> --draft-id <id>",
		ShortHelp: "List attachments for a credit-note draft.",
		Flags:     listFS,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).CreditNoteDraftsAttachmentsList(ctx, ops.CreditNoteDraftsAttachmentsListIn{
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
		Usage:     "fiken credit-notes drafts attachments attach --company <slug> --draft-id <id> --file <path>",
		ShortHelp: "Attach a file to a credit-note draft.",
		Flags:     attachFS,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).CreditNoteDraftsAttachmentsAttach(ctx, ops.CreditNoteDraftsAttachmentsAttachIn{
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
		Usage:       "fiken credit-notes drafts attachments <subcommand>",
		ShortHelp:   "List or attach files to a credit-note draft.",
		Flags:       set,
		Subcommands: []*ff.Command{listCmd, attachCmd},
	}
}
