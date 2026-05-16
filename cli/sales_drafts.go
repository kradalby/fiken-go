package cli

import (
	"context"
	"io"

	"github.com/peterbourgon/ff/v4"

	"github.com/kradalby/fiken-go/fiken"
	"github.com/kradalby/fiken-go/ops"
)

// salesDraftsCmd wires
// `fiken sales drafts {list,get,create,update,delete,create-from,attachments}`.
//
// Layout mirrors invoices drafts. Bodies for create/update flow
// through --from-file into fiken.DraftRequest.
func salesDraftsCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	set := ff.NewFlagSet("drafts")
	set.SetParent(parent)

	cmd := &ff.Command{
		Name:      "drafts",
		Usage:     "fiken sales drafts <subcommand>",
		ShortHelp: "List, fetch, or manage sale drafts.",
		Flags:     set,
	}

	cmd.Subcommands = append(
		cmd.Subcommands,
		salesDraftsListCmd(set, stdout, stderr, sf),
		salesDraftsGetCmd(set, stdout, stderr, sf),
		salesDraftsCreateCmd(set, stdout, stderr, sf),
		salesDraftsUpdateCmd(set, stdout, stderr, sf),
		salesDraftsDeleteCmd(set, stdout, stderr, sf),
		salesDraftsCreateFromCmd(set, stdout, stderr, sf),
		salesDraftsAttachmentsCmd(set, stdout, stderr, sf),
	)
	return cmd
}

func salesDraftsListCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
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
		Usage:     "fiken sales drafts list --company <slug>",
		ShortHelp: "List sale drafts for a company.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).SaleDraftsList(ctx, ops.SaleDraftsListIn{
				Company:  *sf.flagCompany,
				Page:     page,
				PageSize: pageSize,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

func salesDraftsGetCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("get")
	fs.SetParent(parent)
	var draftID int64
	fs.Int64Var(&draftID, 0, "draft-id", 0, "Draft id (required)")
	return &ff.Command{
		Name:      "get",
		Usage:     "fiken sales drafts get --company <slug> --draft-id <id>",
		ShortHelp: "Get a single sale draft by id.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).SaleDraftsGet(ctx, ops.SaleDraftsGetIn{
				Company: *sf.flagCompany,
				DraftID: draftID,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

func salesDraftsCreateCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("create")
	fs.SetParent(parent)
	var fromFile string
	fs.StringVar(&fromFile, 0, "from-file", "", "Path to JSON body (required)")
	return &ff.Command{
		Name:      "create",
		Usage:     "fiken sales drafts create --company <slug> --from-file <path>",
		ShortHelp: "Create a new sale draft.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			body, e := ReadBodyFile[fiken.DraftRequest](fromFile)
			if e != nil {
				e.Op = ops.OpSalesDraftsCreate
				return Renderer(ctx).Render(ops.Err[ops.SaleDraftOut](e))
			}
			res := Client(ctx).SaleDraftsCreate(ctx, ops.SaleDraftsCreateIn{
				Company: *sf.flagCompany,
				Body:    body,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

func salesDraftsUpdateCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
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
		Usage:     "fiken sales drafts update --company <slug> --draft-id <id> --from-file <path>",
		ShortHelp: "Update an existing sale draft.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			body, e := ReadBodyFile[fiken.DraftRequest](fromFile)
			if e != nil {
				e.Op = ops.OpSalesDraftsUpdate
				return Renderer(ctx).Render(ops.Err[ops.SaleDraftOut](e))
			}
			res := Client(ctx).SaleDraftsUpdate(ctx, ops.SaleDraftsUpdateIn{
				Company: *sf.flagCompany,
				DraftID: draftID,
				Body:    body,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

func salesDraftsDeleteCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("delete")
	fs.SetParent(parent)
	var draftID int64
	fs.Int64Var(&draftID, 0, "draft-id", 0, "Draft id to delete (required)")
	return &ff.Command{
		Name:      "delete",
		Usage:     "fiken sales drafts delete --company <slug> --draft-id <id>",
		ShortHelp: "Delete a sale draft.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).SaleDraftsDelete(ctx, ops.SaleDraftsDeleteIn{
				Company: *sf.flagCompany, DraftID: draftID,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

func salesDraftsCreateFromCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("create-from")
	fs.SetParent(parent)
	var draftID int64
	fs.Int64Var(&draftID, 0, "draft-id", 0, "Source draft id (required)")
	return &ff.Command{
		Name:      "create-from",
		Usage:     "fiken sales drafts create-from --company <slug> --draft-id <id>",
		ShortHelp: "Promote a draft to a posted sale.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).SaleDraftsCreateFrom(ctx, ops.SaleDraftsCreateFromIn{
				Company: *sf.flagCompany, DraftID: draftID,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

func salesDraftsAttachmentsCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	set := ff.NewFlagSet("attachments")
	set.SetParent(parent)

	listFS := ff.NewFlagSet("list")
	listFS.SetParent(set)
	var listDraftID int64
	listFS.Int64Var(&listDraftID, 0, "draft-id", 0, "Draft id (required)")
	listCmd := &ff.Command{
		Name:      "list",
		Usage:     "fiken sales drafts attachments list --company <slug> --draft-id <id>",
		ShortHelp: "List attachments for a sale draft.",
		Flags:     listFS,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).SaleDraftsAttachmentsList(ctx, ops.SaleDraftsAttachmentsListIn{
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
		Usage:     "fiken sales drafts attachments attach --company <slug> --draft-id <id> --file <path>",
		ShortHelp: "Attach a file to a sale draft.",
		Flags:     attachFS,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).SaleDraftsAttachmentsAttach(ctx, ops.SaleDraftsAttachmentsAttachIn{
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
		Usage:       "fiken sales drafts attachments <subcommand>",
		ShortHelp:   "List or attach files to a sale draft.",
		Flags:       set,
		Subcommands: []*ff.Command{listCmd, attachCmd},
	}
}
