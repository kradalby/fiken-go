package cli

import (
	"context"
	"io"

	"github.com/peterbourgon/ff/v4"

	"github.com/kradalby/fiken-go/ops"
)

// AddTimeUsers wires `fiken time-users {list,get}`. The upstream tag
// exposes only GETs, so there are no mutating stubs to wire here.
func AddTimeUsers(root *ff.Command, stdout, stderr io.Writer, sf *sessionFactory) error {
	_ = stderr
	set := ff.NewFlagSet("time-users")
	if parentFS, ok := root.Flags.(*ff.FlagSet); ok {
		set.SetParent(parentFS)
	}

	cmd := &ff.Command{
		Name:      "time-users",
		Usage:     "fiken time-users <subcommand>",
		ShortHelp: "Inspect persons who can register time entries.",
		Flags:     set,
	}

	listSet := ff.NewFlagSet("list")
	listSet.SetParent(set)
	var (
		listPage     int
		listPageSize int
		listName     string
		listEmail    string
	)
	listSet.IntVar(&listPage, 0, "page", 0, "Page number (0-indexed)")
	listSet.IntVar(&listPageSize, 0, "page-size", 0, "Page size")
	listSet.StringVar(&listName, 0, "name", "", "Filter by name (partial match)")
	listSet.StringVar(&listEmail, 0, "email", "", "Filter by email")
	listCmd := &ff.Command{
		Name:      "list",
		Usage:     "fiken time-users list --company <slug>",
		ShortHelp: "List time users for a company.",
		Flags:     listSet,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).TimeUsersList(ctx, ops.TimeUsersListIn{
				Company:  *sf.flagCompany,
				Page:     listPage,
				PageSize: listPageSize,
				Name:     listName,
				Email:    listEmail,
			})
			return Renderer(ctx).Render(res)
		},
	}

	getSet := ff.NewFlagSet("get")
	getSet.SetParent(set)
	var getTimeUserID int64
	getSet.Int64Var(&getTimeUserID, 0, "time-user-id", 0, "Time user id (required)")
	getCmd := &ff.Command{
		Name:      "get",
		Usage:     "fiken time-users get --company <slug> --time-user-id <id>",
		ShortHelp: "Get a single time user by id.",
		Flags:     getSet,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).TimeUsersGet(ctx, ops.TimeUsersGetIn{
				Company:    *sf.flagCompany,
				TimeUserID: getTimeUserID,
			})
			return Renderer(ctx).Render(res)
		},
	}

	cmd.Subcommands = []*ff.Command{listCmd, getCmd}
	root.Subcommands = append(root.Subcommands, cmd)
	return nil
}
