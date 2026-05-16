package cli

import (
	"context"
	"io"

	"github.com/peterbourgon/ff/v4"

	"github.com/kradalby/fiken-go/ops"
)

// contactsAttachmentsCmd wires `fiken contacts attachments attach ...`.
//
// The upstream tag exposes only the attachment upload op (no
// listAttachmentsForContact path), so we surface a single attach
// subcommand under contacts attachments.
func contactsAttachmentsCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	set := ff.NewFlagSet("attachments")
	set.SetParent(parent)

	attachFS := ff.NewFlagSet("attach")
	attachFS.SetParent(set)
	var (
		contactID int64
		filename  string
		filePath  string
		comment   string
	)
	attachFS.Int64Var(&contactID, 0, "contact-id", 0, "Contact id (required)")
	attachFS.StringVar(&filename, 0, "filename", "", "Override filename (defaults to basename of --file)")
	attachFS.StringVar(&filePath, 0, "file", "", "Local file path to upload (required)")
	attachFS.StringVar(&comment, 0, "comment", "", "Optional free-text annotation")
	attachCmd := &ff.Command{
		Name:      "attach",
		Usage:     "fiken contacts attachments attach --company <slug> --contact-id <id> --file <path>",
		ShortHelp: "Attach a file to a contact.",
		Flags:     attachFS,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).ContactsAttachmentsAttach(ctx, ops.ContactsAttachmentsAttachIn{
				Company:   *sf.flagCompany,
				ContactID: contactID,
				Filename:  filename,
				FilePath:  filePath,
				Comment:   comment,
			})
			return Renderer(ctx).Render(res)
		},
	}

	return &ff.Command{
		Name:        "attachments",
		Usage:       "fiken contacts attachments <subcommand>",
		ShortHelp:   "Attach files to a contact.",
		Flags:       set,
		Subcommands: []*ff.Command{attachCmd},
	}
}
