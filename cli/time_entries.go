package cli

import (
	"context"
	"io"

	"github.com/peterbourgon/ff/v4"

	"github.com/kradalby/fiken-go/fiken"
	"github.com/kradalby/fiken-go/ops"
)

// AddTimeEntries wires `fiken time-entries
// {list,get,create,update,delete,invoice-draft}`. All subcommands are
// wired against the upstream endpoints.
func AddTimeEntries(root *ff.Command, stdout, stderr io.Writer, sf *sessionFactory) error {
	_ = stderr
	set := ff.NewFlagSet("time-entries")
	if parentFS, ok := root.Flags.(*ff.FlagSet); ok {
		set.SetParent(parentFS)
	}

	cmd := &ff.Command{
		Name:      "time-entries",
		Usage:     "fiken time-entries <subcommand>",
		ShortHelp: "Manage time entries.",
		Flags:     set,
	}

	cmd.Subcommands = append(
		cmd.Subcommands,
		timeEntriesListCmd(set, stdout, stderr, sf),
		timeEntriesGetCmd(set, stdout, stderr, sf),
		timeEntriesCreateCmd(set, stdout, stderr, sf),
		timeEntriesUpdateCmd(set, stdout, stderr, sf),
		timeEntriesDeleteCmd(set, stdout, stderr, sf),
		timeEntriesInvoiceDraftCmd(set, stdout, stderr, sf),
	)

	root.Subcommands = append(root.Subcommands, cmd)
	return nil
}

// timeEntriesListCmd wires `fiken time-entries list ...`. --invoiced is
// a tri-state string mirroring projects --completed; "" → no filter.
func timeEntriesListCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("list")
	fs.SetParent(parent)
	var (
		page       int
		pageSize   int
		date       string
		dateGe     string
		dateLe     string
		projectID  int64
		activityID int64
		timeUserID int64
		invoiced   string
	)
	fs.IntVar(&page, 0, "page", 0, "Page number (0-indexed)")
	fs.IntVar(&pageSize, 0, "page-size", 0, "Page size")
	fs.StringVar(&date, 0, "date", "", "Exact date YYYY-MM-DD")
	fs.StringVar(&dateGe, 0, "date-ge", "", "Date >= YYYY-MM-DD")
	fs.StringVar(&dateLe, 0, "date-le", "", "Date <= YYYY-MM-DD")
	fs.Int64Var(&projectID, 0, "project-id", 0, "Filter by project id")
	fs.Int64Var(&activityID, 0, "activity-id", 0, "Filter by activity id")
	fs.Int64Var(&timeUserID, 0, "time-user-id", 0, "Filter by time user id")
	fs.StringVar(&invoiced, 0, "invoiced", "", "Filter by invoiced: 'true', 'false', or omit for all")
	return &ff.Command{
		Name:      "list",
		Usage:     "fiken time-entries list --company <slug>",
		ShortHelp: "List time entries for a company.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			in := ops.TimeEntriesListIn{
				Company:    *sf.flagCompany,
				Page:       page,
				PageSize:   pageSize,
				Date:       ops.Date(date),
				DateGe:     ops.Date(dateGe),
				DateLe:     ops.Date(dateLe),
				ProjectID:  projectID,
				ActivityID: activityID,
				TimeUserID: timeUserID,
			}
			switch invoiced {
			case "true":
				v := true
				in.Invoiced = &v
			case "false":
				v := false
				in.Invoiced = &v
			}
			res := Client(ctx).TimeEntriesList(ctx, in)
			return Renderer(ctx).Render(res)
		},
	}
}

// timeEntriesGetCmd wires `fiken time-entries get ...`.
func timeEntriesGetCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("get")
	fs.SetParent(parent)
	var timeEntryID int64
	fs.Int64Var(&timeEntryID, 0, "time-entry-id", 0, "Time entry id (required)")
	return &ff.Command{
		Name:      "get",
		Usage:     "fiken time-entries get --company <slug> --time-entry-id <id>",
		ShortHelp: "Get a single time entry by id.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).TimeEntriesGet(ctx, ops.TimeEntriesGetIn{
				Company:     *sf.flagCompany,
				TimeEntryID: timeEntryID,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// timeEntriesCreateCmd wires `fiken time-entries create ...`.
func timeEntriesCreateCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("create")
	fs.SetParent(parent)
	var fromFile string
	fs.StringVar(&fromFile, 0, "from-file", "", "Path to JSON body (required)")
	return &ff.Command{
		Name:      "create",
		Usage:     "fiken time-entries create --company <slug> --from-file <path>",
		ShortHelp: "Create a time entry.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			body, e := ReadBodyFile[fiken.TimeEntryRequest](fromFile)
			if e != nil {
				e.Op = ops.OpTimeEntriesCreate
				return Renderer(ctx).Render(ops.Err[ops.TimeEntryOut](e))
			}
			res := Client(ctx).TimeEntriesCreate(ctx, ops.TimeEntriesCreateIn{
				Company: *sf.flagCompany,
				Body:    body,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// timeEntriesUpdateCmd wires `fiken time-entries update ...`.
func timeEntriesUpdateCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("update")
	fs.SetParent(parent)
	var (
		timeEntryID int64
		fromFile    string
	)
	fs.Int64Var(&timeEntryID, 0, "time-entry-id", 0, "Time entry id (required)")
	fs.StringVar(&fromFile, 0, "from-file", "", "Path to JSON body (required)")
	return &ff.Command{
		Name:      "update",
		Usage:     "fiken time-entries update --company <slug> --time-entry-id <id> --from-file <path>",
		ShortHelp: "Update a time entry.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			body, e := ReadBodyFile[fiken.UpdateTimeEntryRequest](fromFile)
			if e != nil {
				e.Op = ops.OpTimeEntriesUpdate
				return Renderer(ctx).Render(ops.Err[ops.TimeEntryOut](e))
			}
			res := Client(ctx).TimeEntriesUpdate(ctx, ops.TimeEntriesUpdateIn{
				Company:     *sf.flagCompany,
				TimeEntryID: timeEntryID,
				Body:        body,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// timeEntriesDeleteCmd wires `fiken time-entries delete ...`.
func timeEntriesDeleteCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("delete")
	fs.SetParent(parent)
	var timeEntryID int64
	fs.Int64Var(&timeEntryID, 0, "time-entry-id", 0, "Time entry id (required)")
	return &ff.Command{
		Name:      "delete",
		Usage:     "fiken time-entries delete --company <slug> --time-entry-id <id>",
		ShortHelp: "Delete a time entry.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).TimeEntriesDelete(ctx, ops.TimeEntriesDeleteIn{
				Company:     *sf.flagCompany,
				TimeEntryID: timeEntryID,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// timeEntriesInvoiceDraftCmd wires `fiken time-entries invoice-draft ...`.
// Bundles a set of time entries into a new invoice draft.
func timeEntriesInvoiceDraftCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("invoice-draft")
	fs.SetParent(parent)
	var fromFile string
	fs.StringVar(&fromFile, 0, "from-file", "", "Path to JSON body (required)")
	return &ff.Command{
		Name:      "invoice-draft",
		Usage:     "fiken time-entries invoice-draft --company <slug> --from-file <path>",
		ShortHelp: "Create an invoice draft from time entries.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			body, e := ReadBodyFile[fiken.TimeEntryInvoiceDraftRequest](fromFile)
			if e != nil {
				e.Op = ops.OpTimeEntriesInvoiceDraftFromTimes
				return Renderer(ctx).Render(ops.Err[ops.TimeEntriesInvoiceDraftFromTimesOut](e))
			}
			res := Client(ctx).TimeEntriesInvoiceDraftFromTimes(ctx, ops.TimeEntriesInvoiceDraftFromTimesIn{
				Company: *sf.flagCompany,
				Body:    body,
			})
			return Renderer(ctx).Render(res)
		},
	}
}
