package cli

import (
	"context"
	"io"

	"github.com/peterbourgon/ff/v4"

	"github.com/kradalby/fiken-go/fiken"
	"github.com/kradalby/fiken-go/ops"
)

// AddActivities wires `fiken activities {list,get,create,update,delete}`.
//
// All subcommands are wired against the upstream endpoints.
func AddActivities(root *ff.Command, stdout, stderr io.Writer, sf *sessionFactory) error {
	_ = stderr
	set := ff.NewFlagSet("activities")
	if parentFS, ok := root.Flags.(*ff.FlagSet); ok {
		set.SetParent(parentFS)
	}

	cmd := &ff.Command{
		Name:      "activities",
		Usage:     "fiken activities <subcommand>",
		ShortHelp: "Manage Fiken activities for time tracking.",
		Flags:     set,
	}

	cmd.Subcommands = append(
		cmd.Subcommands,
		activitiesListCmd(set, stdout, stderr, sf),
		activitiesGetCmd(set, stdout, stderr, sf),
		activitiesCreateCmd(set, stdout, stderr, sf),
		activitiesUpdateCmd(set, stdout, stderr, sf),
		activitiesDeleteCmd(set, stdout, stderr, sf),
	)

	root.Subcommands = append(root.Subcommands, cmd)
	return nil
}

// activitiesListCmd wires `fiken activities list ...`. The --archived
// flag is a tri-state string ("", "true", "false") so omitting it
// returns all activities (no filter) — matching upstream semantics.
func activitiesListCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("list")
	fs.SetParent(parent)
	var (
		page     int
		pageSize int
		name     string
		archived string
	)
	fs.IntVar(&page, 0, "page", 0, "Page number (0-indexed)")
	fs.IntVar(&pageSize, 0, "page-size", 0, "Page size")
	fs.StringVar(&name, 0, "name", "", "Filter by activity name (partial match)")
	fs.StringVar(&archived, 0, "archived", "", "Filter by archived: 'true', 'false', or omit for all")
	return &ff.Command{
		Name:      "list",
		Usage:     "fiken activities list --company <slug>",
		ShortHelp: "List activities for a company.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			in := ops.ActivitiesListIn{
				Company:  *sf.flagCompany,
				Page:     page,
				PageSize: pageSize,
				Name:     name,
			}
			switch archived {
			case "true":
				v := true
				in.Archived = &v
			case "false":
				v := false
				in.Archived = &v
			}
			res := Client(ctx).ActivitiesList(ctx, in)
			return Renderer(ctx).Render(res)
		},
	}
}

// activitiesGetCmd wires `fiken activities get ...`.
func activitiesGetCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("get")
	fs.SetParent(parent)
	var activityID int64
	fs.Int64Var(&activityID, 0, "activity-id", 0, "Activity id (required)")
	return &ff.Command{
		Name:      "get",
		Usage:     "fiken activities get --company <slug> --activity-id <id>",
		ShortHelp: "Get a single activity by id.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).ActivitiesGet(ctx, ops.ActivitiesGetIn{
				Company:    *sf.flagCompany,
				ActivityID: activityID,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// activitiesCreateCmd wires `fiken activities create ...`.
func activitiesCreateCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("create")
	fs.SetParent(parent)
	var fromFile string
	fs.StringVar(&fromFile, 0, "from-file", "", "Path to JSON body (required)")
	return &ff.Command{
		Name:      "create",
		Usage:     "fiken activities create --company <slug> --from-file <path>",
		ShortHelp: "Create an activity.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			body, e := ReadBodyFile[fiken.ActivityRequest](fromFile)
			if e != nil {
				e.Op = ops.OpActivitiesCreate
				return Renderer(ctx).Render(ops.Err[ops.ActivityOut](e))
			}
			res := Client(ctx).ActivitiesCreate(ctx, ops.ActivitiesCreateIn{
				Company: *sf.flagCompany,
				Body:    body,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// activitiesUpdateCmd wires `fiken activities update ...`.
func activitiesUpdateCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("update")
	fs.SetParent(parent)
	var (
		activityID int64
		fromFile   string
	)
	fs.Int64Var(&activityID, 0, "activity-id", 0, "Activity id (required)")
	fs.StringVar(&fromFile, 0, "from-file", "", "Path to JSON body (required)")
	return &ff.Command{
		Name:      "update",
		Usage:     "fiken activities update --company <slug> --activity-id <id> --from-file <path>",
		ShortHelp: "Update an activity.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			body, e := ReadBodyFile[fiken.UpdateActivityRequest](fromFile)
			if e != nil {
				e.Op = ops.OpActivitiesUpdate
				return Renderer(ctx).Render(ops.Err[ops.ActivityOut](e))
			}
			res := Client(ctx).ActivitiesUpdate(ctx, ops.ActivitiesUpdateIn{
				Company:    *sf.flagCompany,
				ActivityID: activityID,
				Body:       body,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// activitiesDeleteCmd wires `fiken activities delete ...`.
func activitiesDeleteCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("delete")
	fs.SetParent(parent)
	var activityID int64
	fs.Int64Var(&activityID, 0, "activity-id", 0, "Activity id (required)")
	return &ff.Command{
		Name:      "delete",
		Usage:     "fiken activities delete --company <slug> --activity-id <id>",
		ShortHelp: "Delete an activity.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).ActivitiesDelete(ctx, ops.ActivitiesDeleteIn{
				Company:    *sf.flagCompany,
				ActivityID: activityID,
			})
			return Renderer(ctx).Render(res)
		},
	}
}
