package cli

import (
	"context"
	"io"

	"github.com/peterbourgon/ff/v4"

	"github.com/kradalby/fiken-go/ops"
)

// AddUser wires `fiken user me`.
//
// The Fiken /user endpoint returns the caller identified by the bearer
// token, so the command is user-level (no --company flag). The subcommand
// is named `me` to make the intent obvious — `fiken user me` reads as
// "tell me about me".
func AddUser(root *ff.Command, stdout, stderr io.Writer, sf *sessionFactory) error {
	_ = stderr
	set := ff.NewFlagSet("user")
	if parentFS, ok := root.Flags.(*ff.FlagSet); ok {
		set.SetParent(parentFS)
	}

	cmd := &ff.Command{
		Name:      "user",
		Usage:     "fiken user <subcommand>",
		ShortHelp: "Inspect the authenticated user.",
		Flags:     set,
	}

	meSet := ff.NewFlagSet("me")
	meSet.SetParent(set)
	meCmd := &ff.Command{
		Name:      "me",
		Usage:     "fiken user me",
		ShortHelp: "Show the authenticated user's name and email.",
		Flags:     meSet,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).UserGet(ctx, ops.UserGetIn{})
			return Renderer(ctx).Render(res)
		},
	}

	cmd.Subcommands = []*ff.Command{meCmd}
	root.Subcommands = append(root.Subcommands, cmd)
	return nil
}
