package cli

import (
	"context"
	"io"

	"github.com/peterbourgon/ff/v4"

	"github.com/kradalby/fiken-go/ops"
)

// AddInbox wires `fiken inbox {list,get,send,delete}`.
//
// list + get + send are fully wired against the upstream paths;
// send streams a local file as multipart form data. delete has no
// upstream endpoint — surface a CodeInternal stub so the surface
// stays uniform with the other tags.
func AddInbox(root *ff.Command, stdout, stderr io.Writer, sf *sessionFactory) error {
	_ = stderr
	set := ff.NewFlagSet("inbox")
	if parentFS, ok := root.Flags.(*ff.FlagSet); ok {
		set.SetParent(parentFS)
	}

	cmd := &ff.Command{
		Name:      "inbox",
		Usage:     "fiken inbox <subcommand>",
		ShortHelp: "List, fetch, or upload inbox documents.",
		Flags:     set,
	}

	cmd.Subcommands = append(
		cmd.Subcommands,
		inboxListCmd(set, stdout, stderr, sf),
		inboxGetCmd(set, stdout, stderr, sf),
		inboxSendCmd(set, stdout, stderr, sf),
		inboxDeleteCmd(set, stdout, stderr, sf),
	)

	root.Subcommands = append(root.Subcommands, cmd)
	return nil
}

// inboxListCmd wires `fiken inbox list ...`. Sort/status flags
// accept the upstream literals ("createdDate asc", "unused" ...) so
// the CLI mirrors the Fiken docs verbatim.
func inboxListCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("list")
	fs.SetParent(parent)
	var (
		page     int
		pageSize int
		sortBy   string
		status   string
		name     string
	)
	fs.IntVar(&page, 0, "page", 0, "Page number (0-indexed)")
	fs.IntVar(&pageSize, 0, "page-size", 0, "Page size")
	fs.StringVar(&sortBy, 0, "sort-by", "", "Sort order ('createdDate asc'|'createdDate desc'|'name asc'|'name desc')")
	fs.StringVar(&status, 0, "status", "", "Filter by status (all|unused|used)")
	fs.StringVar(&name, 0, "name", "", "Case-insensitive substring filter on document name")
	return &ff.Command{
		Name:      "list",
		Usage:     "fiken inbox list --company <slug>",
		ShortHelp: "List inbox documents for a company.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).InboxList(ctx, ops.InboxListIn{
				Company:  *sf.flagCompany,
				Page:     page,
				PageSize: pageSize,
				SortBy:   sortBy,
				Status:   status,
				Name:     name,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// inboxGetCmd wires `fiken inbox get ...`.
func inboxGetCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("get")
	fs.SetParent(parent)
	var documentID int64
	fs.Int64Var(&documentID, 0, "document-id", 0, "Inbox document id (required)")
	return &ff.Command{
		Name:      "get",
		Usage:     "fiken inbox get --company <slug> --document-id <id>",
		ShortHelp: "Get a single inbox document by id.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).InboxGet(ctx, ops.InboxGetIn{
				Company:    *sf.flagCompany,
				DocumentID: documentID,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// inboxSendCmd wires `fiken inbox send ...`.
func inboxSendCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("send")
	fs.SetParent(parent)
	var (
		name        string
		filename    string
		description string
		filePath    string
	)
	fs.StringVar(&name, 0, "name", "", "Document name (defaults to filename)")
	fs.StringVar(&filename, 0, "filename", "", "Override filename (defaults to basename of --file)")
	fs.StringVar(&description, 0, "description", "", "Additional description")
	fs.StringVar(&filePath, 0, "file", "", "Local file path to upload (required)")
	return &ff.Command{
		Name:      "send",
		Usage:     "fiken inbox send --company <slug> --file <path>",
		ShortHelp: "Upload a document to the inbox.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).InboxSend(ctx, ops.InboxSendIn{
				Company:     *sf.flagCompany,
				Name:        name,
				Filename:    filename,
				Description: description,
				FilePath:    filePath,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// inboxDeleteCmd wires `fiken inbox delete ...`. Upstream has no
// delete endpoint for inbox documents; this stub keeps the CLI shape
// uniform with the other tags and surfaces a clear "not implemented"
// error to the caller.
func inboxDeleteCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("delete")
	fs.SetParent(parent)
	var documentID int64
	fs.Int64Var(&documentID, 0, "document-id", 0, "Inbox document id (required)")
	return &ff.Command{
		Name:      "delete",
		Usage:     "fiken inbox delete --company <slug> --document-id <id>",
		ShortHelp: "Delete an inbox document (no upstream endpoint — stub).",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			// No upstream endpoint exposes inbox delete. Surface a
			// declarative Result[InboxDocumentOut] error so CLI callers
			// observe a uniform Err envelope. Op carries the stub name
			// for log/error parity with the registered ops; it is
			// intentionally absent from ops.Registry since the upstream
			// has no corresponding operationId. Flag values are bound
			// for help/discovery but not consumed by the stub itself.
			_ = documentID
			res := ops.Err[ops.InboxDocumentOut](&ops.Error{
				Code:    ops.CodeInternal,
				Message: "inbox delete is not exposed by the Fiken API",
				Op:      "inbox_delete",
			})
			return Renderer(ctx).Render(res)
		},
	}
}
