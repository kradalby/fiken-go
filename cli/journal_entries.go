package cli

import (
	"context"
	"io"

	"github.com/peterbourgon/ff/v4"

	"github.com/kradalby/fiken-go/fiken"
	"github.com/kradalby/fiken-go/ops"
)

// AddJournalEntries wires `fiken journal-entries {list,get,create,attachments}`.
//
// list + get + attachments list + create are wired against the ogen
// Client; create takes its JSON body via --from-file. The multipart
// attachments attach subcommand is registered for help completeness
// but surfaces the Plan-D-deferred stub error.
func AddJournalEntries(root *ff.Command, stdout, stderr io.Writer, sf *sessionFactory) error {
	_ = stderr
	set := ff.NewFlagSet("journal-entries")
	if parentFS, ok := root.Flags.(*ff.FlagSet); ok {
		set.SetParent(parentFS)
	}

	cmd := &ff.Command{
		Name:      "journal-entries",
		Usage:     "fiken journal-entries <subcommand>",
		ShortHelp: "Inspect or create journal entries.",
		Flags:     set,
	}

	// list
	listSet := ff.NewFlagSet("list")
	listSet.SetParent(set)
	var (
		listPage           int
		listPageSize       int
		listDate           string
		listDateLe         string
		listDateLt         string
		listDateGe         string
		listDateGt         string
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
	listSet.StringVar(&listDate, 0, "date", "", "Filter by exact posting date (YYYY-MM-DD)")
	listSet.StringVar(&listDateLe, 0, "date-le", "", "Filter posting date <= value")
	listSet.StringVar(&listDateLt, 0, "date-lt", "", "Filter posting date < value")
	listSet.StringVar(&listDateGe, 0, "date-ge", "", "Filter posting date >= value")
	listSet.StringVar(&listDateGt, 0, "date-gt", "", "Filter posting date > value")
	listSet.StringVar(&listLastModified, 0, "last-modified", "", "Filter by exact last-modified date")
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
		Usage:     "fiken journal-entries list --company <slug>",
		ShortHelp: "List journal entries for a company.",
		Flags:     listSet,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).JournalEntriesList(ctx, ops.JournalEntriesListIn{
				Company:        *sf.flagCompany,
				Page:           listPage,
				PageSize:       listPageSize,
				Date:           ops.Date(listDate),
				DateLe:         ops.Date(listDateLe),
				DateLt:         ops.Date(listDateLt),
				DateGe:         ops.Date(listDateGe),
				DateGt:         ops.Date(listDateGt),
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
	var getJournalEntryID int64
	getSet.Int64Var(&getJournalEntryID, 0, "journal-entry-id", 0, "Journal entry id (required)")
	getCmd := &ff.Command{
		Name:      "get",
		Usage:     "fiken journal-entries get --company <slug> --journal-entry-id <id>",
		ShortHelp: "Get a single journal entry by id.",
		Flags:     getSet,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).JournalEntriesGet(ctx, ops.JournalEntriesGetIn{
				Company:        *sf.flagCompany,
				JournalEntryID: getJournalEntryID,
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
		Usage:     "fiken journal-entries create --company <slug> --from-file <path>",
		ShortHelp: "Create a general journal entry.",
		Flags:     createSet,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			body, e := ReadBodyFile[fiken.GeneralJournalEntryRequest](createFromFile)
			if e != nil {
				e.Op = ops.OpJournalEntriesCreate
				return Renderer(ctx).Render(ops.Err[ops.JournalEntryOut](e))
			}
			res := Client(ctx).JournalEntriesCreate(ctx, ops.JournalEntriesCreateIn{
				Company: *sf.flagCompany,
				Body:    body,
			})
			return Renderer(ctx).Render(res)
		},
	}

	// attachments (parent — list + attach children)
	attachmentsSet := ff.NewFlagSet("attachments")
	attachmentsSet.SetParent(set)

	attListFS := ff.NewFlagSet("list")
	attListFS.SetParent(attachmentsSet)
	var attListJournalEntryID int64
	attListFS.Int64Var(&attListJournalEntryID, 0, "journal-entry-id", 0, "Journal entry id (required)")
	attListCmd := &ff.Command{
		Name:      "list",
		Usage:     "fiken journal-entries attachments list --company <slug> --journal-entry-id <id>",
		ShortHelp: "List attachments for a journal entry.",
		Flags:     attListFS,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).JournalEntriesAttachmentsList(ctx, ops.JournalEntriesAttachmentsListIn{
				Company:        *sf.flagCompany,
				JournalEntryID: attListJournalEntryID,
			})
			return Renderer(ctx).Render(res)
		},
	}

	attAttachFS := ff.NewFlagSet("attach")
	attAttachFS.SetParent(attachmentsSet)
	var (
		attAttachJournalEntryID int64
		attAttachFilename       string
		attAttachFilePath       string
	)
	attAttachFS.Int64Var(&attAttachJournalEntryID, 0, "journal-entry-id", 0, "Journal entry id (required)")
	attAttachFS.StringVar(&attAttachFilename, 0, "filename", "", "Override filename (defaults to basename of --file)")
	attAttachFS.StringVar(&attAttachFilePath, 0, "file", "", "Local file path to upload (required)")
	attAttachCmd := &ff.Command{
		Name:      "attach",
		Usage:     "fiken journal-entries attachments attach --company <slug> --journal-entry-id <id> --file <path>",
		ShortHelp: "Attach a file to a journal entry.",
		Flags:     attAttachFS,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).JournalEntriesAttachmentsAttach(ctx, ops.JournalEntriesAttachmentsAttachIn{
				Company:        *sf.flagCompany,
				JournalEntryID: attAttachJournalEntryID,
				Filename:       attAttachFilename,
				FilePath:       attAttachFilePath,
			})
			return Renderer(ctx).Render(res)
		},
	}

	attachmentsCmd := &ff.Command{
		Name:        "attachments",
		Usage:       "fiken journal-entries attachments <subcommand>",
		ShortHelp:   "List or attach files to a journal entry.",
		Flags:       attachmentsSet,
		Subcommands: []*ff.Command{attListCmd, attAttachCmd},
	}

	cmd.Subcommands = []*ff.Command{listCmd, getCmd, createCmd, attachmentsCmd}
	root.Subcommands = append(root.Subcommands, cmd)
	return nil
}
