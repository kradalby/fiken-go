package cli

import (
	"context"
	"io"

	"github.com/peterbourgon/ff/v4"

	"github.com/kradalby/fiken-go/ops"
)

// AddGroups wires `fiken groups list`. The upstream tag exposes a
// single GET that returns a flat list of customer-group names.
func AddGroups(root *ff.Command, stdout, stderr io.Writer, sf *sessionFactory) error {
	_ = stderr
	set := ff.NewFlagSet("groups")
	if parentFS, ok := root.Flags.(*ff.FlagSet); ok {
		set.SetParent(parentFS)
	}

	cmd := &ff.Command{
		Name:      "groups",
		Usage:     "fiken groups <subcommand>",
		ShortHelp: "Inspect customer groups.",
		Flags:     set,
	}

	listSet := ff.NewFlagSet("list")
	listSet.SetParent(set)
	var (
		listPage     int
		listPageSize int
	)
	listSet.IntVar(&listPage, 0, "page", 0, "Page number (0-indexed)")
	listSet.IntVar(&listPageSize, 0, "page-size", 0, "Page size")
	listCmd := &ff.Command{
		Name:      "list",
		Usage:     "fiken groups list --company <slug>",
		ShortHelp: "List customer groups for a company.",
		Flags:     listSet,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).GroupsList(ctx, ops.GroupsListIn{
				Company:  *sf.flagCompany,
				Page:     listPage,
				PageSize: listPageSize,
			})
			return Renderer(ctx).Render(res)
		},
	}

	cmd.Subcommands = []*ff.Command{listCmd}
	root.Subcommands = append(root.Subcommands, cmd)
	return nil
}
