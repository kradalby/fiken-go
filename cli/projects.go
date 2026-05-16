package cli

import (
	"context"
	"io"

	"github.com/peterbourgon/ff/v4"

	"github.com/kradalby/fiken-go/fiken"
	"github.com/kradalby/fiken-go/ops"
)

// AddProjects wires `fiken projects {list,get,create,update,delete}`.
//
// All subcommands are wired against the upstream endpoints.
func AddProjects(root *ff.Command, stdout, stderr io.Writer, sf *sessionFactory) error {
	_ = stderr
	set := ff.NewFlagSet("projects")
	if parentFS, ok := root.Flags.(*ff.FlagSet); ok {
		set.SetParent(parentFS)
	}

	cmd := &ff.Command{
		Name:      "projects",
		Usage:     "fiken projects <subcommand>",
		ShortHelp: "Manage Fiken projects.",
		Flags:     set,
	}

	cmd.Subcommands = append(
		cmd.Subcommands,
		projectsListCmd(set, stdout, stderr, sf),
		projectsGetCmd(set, stdout, stderr, sf),
		projectsCreateCmd(set, stdout, stderr, sf),
		projectsUpdateCmd(set, stdout, stderr, sf),
		projectsDeleteCmd(set, stdout, stderr, sf),
	)

	root.Subcommands = append(root.Subcommands, cmd)
	return nil
}

// projectsListCmd wires `fiken projects list ...`. The --completed flag
// is a tri-state string ("", "true", "false") so omitting it returns
// all projects (no filter), matching the upstream semantic where the
// completed query parameter is itself optional.
func projectsListCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("list")
	fs.SetParent(parent)
	var (
		page      int
		pageSize  int
		completed string
		name      string
		number    string
	)
	fs.IntVar(&page, 0, "page", 0, "Page number (0-indexed)")
	fs.IntVar(&pageSize, 0, "page-size", 0, "Page size")
	fs.StringVar(&completed, 0, "completed", "", "Filter by completed: 'true', 'false', or omit for all")
	fs.StringVar(&name, 0, "name", "", "Filter by project name")
	fs.StringVar(&number, 0, "number", "", "Filter by project number")
	return &ff.Command{
		Name:      "list",
		Usage:     "fiken projects list --company <slug>",
		ShortHelp: "List projects for a company.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			in := ops.ProjectsListIn{
				Company:  *sf.flagCompany,
				Page:     page,
				PageSize: pageSize,
				Name:     name,
				Number:   number,
			}
			switch completed {
			case "true":
				v := true
				in.Completed = &v
			case "false":
				v := false
				in.Completed = &v
			}
			res := Client(ctx).ProjectsList(ctx, in)
			return Renderer(ctx).Render(res)
		},
	}
}

// projectsGetCmd wires `fiken projects get ...`.
func projectsGetCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("get")
	fs.SetParent(parent)
	var projectID int64
	fs.Int64Var(&projectID, 0, "project-id", 0, "Project id (required)")
	return &ff.Command{
		Name:      "get",
		Usage:     "fiken projects get --company <slug> --project-id <id>",
		ShortHelp: "Get a single project by id.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).ProjectsGet(ctx, ops.ProjectsGetIn{
				Company:   *sf.flagCompany,
				ProjectID: projectID,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// projectsCreateCmd wires `fiken projects create ...`.
func projectsCreateCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("create")
	fs.SetParent(parent)
	var fromFile string
	fs.StringVar(&fromFile, 0, "from-file", "", "Path to JSON body (required)")
	return &ff.Command{
		Name:      "create",
		Usage:     "fiken projects create --company <slug> --from-file <path>",
		ShortHelp: "Create a project.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			body, e := ReadBodyFile[fiken.ProjectRequest](fromFile)
			if e != nil {
				e.Op = ops.OpProjectsCreate
				return Renderer(ctx).Render(ops.Err[ops.ProjectOut](e))
			}
			res := Client(ctx).ProjectsCreate(ctx, ops.ProjectsCreateIn{
				Company: *sf.flagCompany,
				Body:    body,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// projectsUpdateCmd wires `fiken projects update ...`.
func projectsUpdateCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("update")
	fs.SetParent(parent)
	var (
		projectID int64
		fromFile  string
	)
	fs.Int64Var(&projectID, 0, "project-id", 0, "Project id (required)")
	fs.StringVar(&fromFile, 0, "from-file", "", "Path to JSON body (required)")
	return &ff.Command{
		Name:      "update",
		Usage:     "fiken projects update --company <slug> --project-id <id> --from-file <path>",
		ShortHelp: "Update a project.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			body, e := ReadBodyFile[fiken.UpdateProjectRequest](fromFile)
			if e != nil {
				e.Op = ops.OpProjectsUpdate
				return Renderer(ctx).Render(ops.Err[ops.ProjectOut](e))
			}
			res := Client(ctx).ProjectsUpdate(ctx, ops.ProjectsUpdateIn{
				Company:   *sf.flagCompany,
				ProjectID: projectID,
				Body:      body,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// projectsDeleteCmd wires `fiken projects delete ...`.
func projectsDeleteCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("delete")
	fs.SetParent(parent)
	var projectID int64
	fs.Int64Var(&projectID, 0, "project-id", 0, "Project id (required)")
	return &ff.Command{
		Name:      "delete",
		Usage:     "fiken projects delete --company <slug> --project-id <id>",
		ShortHelp: "Delete a project.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).ProjectsDelete(ctx, ops.ProjectsDeleteIn{
				Company:   *sf.flagCompany,
				ProjectID: projectID,
			})
			return Renderer(ctx).Render(res)
		},
	}
}
