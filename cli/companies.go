package cli

import (
	"context"
	"io"

	"github.com/peterbourgon/ff/v4"

	"github.com/kradalby/fiken-go/ops"
)

// AddCompanies wires `fiken companies {list,get}`.
func AddCompanies(root *ff.Command, stdout, stderr io.Writer, sf *sessionFactory) error {
	_ = stderr
	set := ff.NewFlagSet("companies")
	if parentFS, ok := root.Flags.(*ff.FlagSet); ok {
		set.SetParent(parentFS)
	}

	cmd := &ff.Command{
		Name:      "companies",
		Usage:     "fiken companies <subcommand>",
		ShortHelp: "Manage Fiken companies.",
		Flags:     set,
	}

	// list
	listSet := ff.NewFlagSet("list")
	listSet.SetParent(set)
	listCmd := &ff.Command{
		Name:      "list",
		Usage:     "fiken companies list",
		ShortHelp: "List all companies the user has access to.",
		Flags:     listSet,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).CompaniesList(ctx, ops.CompaniesListIn{})
			return Renderer(ctx).Render(res)
		},
	}

	// get
	getSet := ff.NewFlagSet("get")
	getSet.SetParent(set)
	getCmd := &ff.Command{
		Name:      "get",
		Usage:     "fiken companies get --company <slug>",
		ShortHelp: "Get a single company by slug.",
		Flags:     getSet,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).CompaniesGet(ctx, ops.CompaniesGetIn{Company: *sf.flagCompany})
			return Renderer(ctx).Render(res)
		},
	}

	cmd.Subcommands = []*ff.Command{listCmd, getCmd}
	root.Subcommands = append(root.Subcommands, cmd)
	return nil
}
