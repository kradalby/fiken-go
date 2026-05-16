package cli

import (
	"context"
	"io"

	"github.com/peterbourgon/ff/v4"

	"github.com/kradalby/fiken-go/fiken"
	"github.com/kradalby/fiken-go/ops"
)

// purchasesDraftsCmd wires
// `fiken purchases drafts {list,get,create,update,delete,create-from,attachments}`.
//
// Mirrors sales drafts; the upstream path uses the shared
// DraftRequest / DraftResult schema.
func purchasesDraftsCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	set := ff.NewFlagSet("drafts")
	set.SetParent(parent)

	cmd := &ff.Command{
		Name:      "drafts",
		Usage:     "fiken purchases drafts <subcommand>",
		ShortHelp: "List, fetch, or manage purchase drafts.",
		Flags:     set,
	}

	cmd.Subcommands = append(
		cmd.Subcommands,
		purchasesDraftsListCmd(set, stdout, stderr, sf),
		purchasesDraftsGetCmd(set, stdout, stderr, sf),
		purchasesDraftsCreateCmd(set, stdout, stderr, sf),
		purchasesDraftsUpdateCmd(set, stdout, stderr, sf),
		purchasesDraftsDeleteCmd(set, stdout, stderr, sf),
		purchasesDraftsCreateFromCmd(set, stdout, stderr, sf),
		purchasesDraftsAttachmentsCmd(set, stdout, stderr, sf),
	)
	return cmd
}

func purchasesDraftsListCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
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
		Usage:     "fiken purchases drafts list --company <slug>",
		ShortHelp: "List purchase drafts for a company.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).PurchaseDraftsList(ctx, ops.PurchaseDraftsListIn{
				Company:  *sf.flagCompany,
				Page:     page,
				PageSize: pageSize,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

func purchasesDraftsGetCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("get")
	fs.SetParent(parent)
	var draftID int64
	fs.Int64Var(&draftID, 0, "draft-id", 0, "Draft id (required)")
	return &ff.Command{
		Name:      "get",
		Usage:     "fiken purchases drafts get --company <slug> --draft-id <id>",
		ShortHelp: "Get a single purchase draft by id.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).PurchaseDraftsGet(ctx, ops.PurchaseDraftsGetIn{
				Company: *sf.flagCompany,
				DraftID: draftID,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

func purchasesDraftsCreateCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("create")
	fs.SetParent(parent)
	var fromFile string
	fs.StringVar(&fromFile, 0, "from-file", "", "Path to JSON body (required)")
	return &ff.Command{
		Name:      "create",
		Usage:     "fiken purchases drafts create --company <slug> --from-file <path>",
		ShortHelp: "Create a new purchase draft.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			body, e := ReadBodyFile[fiken.DraftRequest](fromFile)
			if e != nil {
				e.Op = ops.OpPurchasesDraftsCreate
				return Renderer(ctx).Render(ops.Err[ops.PurchaseDraftOut](e))
			}
			res := Client(ctx).PurchaseDraftsCreate(ctx, ops.PurchaseDraftsCreateIn{
				Company: *sf.flagCompany,
				Body:    body,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

func purchasesDraftsUpdateCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
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
		Usage:     "fiken purchases drafts update --company <slug> --draft-id <id> --from-file <path>",
		ShortHelp: "Update an existing purchase draft.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			body, e := ReadBodyFile[fiken.DraftRequest](fromFile)
			if e != nil {
				e.Op = ops.OpPurchasesDraftsUpdate
				return Renderer(ctx).Render(ops.Err[ops.PurchaseDraftOut](e))
			}
			res := Client(ctx).PurchaseDraftsUpdate(ctx, ops.PurchaseDraftsUpdateIn{
				Company: *sf.flagCompany,
				DraftID: draftID,
				Body:    body,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

func purchasesDraftsDeleteCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("delete")
	fs.SetParent(parent)
	var draftID int64
	fs.Int64Var(&draftID, 0, "draft-id", 0, "Draft id to delete (required)")
	return &ff.Command{
		Name:      "delete",
		Usage:     "fiken purchases drafts delete --company <slug> --draft-id <id>",
		ShortHelp: "Delete a purchase draft.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).PurchaseDraftsDelete(ctx, ops.PurchaseDraftsDeleteIn{
				Company: *sf.flagCompany, DraftID: draftID,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

func purchasesDraftsCreateFromCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("create-from")
	fs.SetParent(parent)
	var draftID int64
	fs.Int64Var(&draftID, 0, "draft-id", 0, "Source draft id (required)")
	return &ff.Command{
		Name:      "create-from",
		Usage:     "fiken purchases drafts create-from --company <slug> --draft-id <id>",
		ShortHelp: "Promote a draft to a posted purchase.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).PurchaseDraftsCreateFrom(ctx, ops.PurchaseDraftsCreateFromIn{
				Company: *sf.flagCompany, DraftID: draftID,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

func purchasesDraftsAttachmentsCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	set := ff.NewFlagSet("attachments")
	set.SetParent(parent)

	listFS := ff.NewFlagSet("list")
	listFS.SetParent(set)
	var listDraftID int64
	listFS.Int64Var(&listDraftID, 0, "draft-id", 0, "Draft id (required)")
	listCmd := &ff.Command{
		Name:      "list",
		Usage:     "fiken purchases drafts attachments list --company <slug> --draft-id <id>",
		ShortHelp: "List attachments for a purchase draft.",
		Flags:     listFS,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).PurchaseDraftsAttachmentsList(ctx, ops.PurchaseDraftsAttachmentsListIn{
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
		Usage:     "fiken purchases drafts attachments attach --company <slug> --draft-id <id> --file <path>",
		ShortHelp: "Attach a file to a purchase draft.",
		Flags:     attachFS,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).PurchaseDraftsAttachmentsAttach(ctx, ops.PurchaseDraftsAttachmentsAttachIn{
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
		Usage:       "fiken purchases drafts attachments <subcommand>",
		ShortHelp:   "List or attach files to a purchase draft.",
		Flags:       set,
		Subcommands: []*ff.Command{listCmd, attachCmd},
	}
}
