package cli

import (
	"context"
	"io"

	"github.com/peterbourgon/ff/v4"

	"github.com/kradalby/fiken-go/fiken"
	"github.com/kradalby/fiken-go/ops"
)

// AddProducts wires `fiken products {list,get,create,update,delete,sales-report}`.
//
// All subcommands are wired against the upstream endpoints.
func AddProducts(root *ff.Command, stdout, stderr io.Writer, sf *sessionFactory) error {
	_ = stderr
	set := ff.NewFlagSet("products")
	if parentFS, ok := root.Flags.(*ff.FlagSet); ok {
		set.SetParent(parentFS)
	}

	cmd := &ff.Command{
		Name:      "products",
		Usage:     "fiken products <subcommand>",
		ShortHelp: "List, fetch, or manage products.",
		Flags:     set,
	}

	cmd.Subcommands = append(
		cmd.Subcommands,
		productsListCmd(set, stdout, stderr, sf),
		productsGetCmd(set, stdout, stderr, sf),
		productsCreateCmd(set, stdout, stderr, sf),
		productsUpdateCmd(set, stdout, stderr, sf),
		productsDeleteCmd(set, stdout, stderr, sf),
		productsSalesReportCmd(set, stdout, stderr, sf),
	)

	root.Subcommands = append(root.Subcommands, cmd)
	return nil
}

// productsListCmd wires `fiken products list ...`.
func productsListCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
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
		Usage:     "fiken products list --company <slug>",
		ShortHelp: "List products for a company.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).ProductsList(ctx, ops.ProductsListIn{
				Company:  *sf.flagCompany,
				Page:     page,
				PageSize: pageSize,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// productsGetCmd wires `fiken products get ...`.
func productsGetCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("get")
	fs.SetParent(parent)
	var productID int64
	fs.Int64Var(&productID, 0, "product-id", 0, "Product id (required)")
	return &ff.Command{
		Name:      "get",
		Usage:     "fiken products get --company <slug> --product-id <id>",
		ShortHelp: "Get a single product by id.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).ProductsGet(ctx, ops.ProductsGetIn{
				Company:   *sf.flagCompany,
				ProductID: productID,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// productsCreateCmd wires `fiken products create ...`.
func productsCreateCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("create")
	fs.SetParent(parent)
	var fromFile string
	fs.StringVar(&fromFile, 0, "from-file", "", "Path to JSON body (required)")
	return &ff.Command{
		Name:      "create",
		Usage:     "fiken products create --company <slug> --from-file <path>",
		ShortHelp: "Create a new product.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			body, e := ReadBodyFile[fiken.Product](fromFile)
			if e != nil {
				e.Op = ops.OpProductsCreate
				return Renderer(ctx).Render(ops.Err[ops.ProductOut](e))
			}
			res := Client(ctx).ProductsCreate(ctx, ops.ProductsCreateIn{
				Company: *sf.flagCompany,
				Body:    body,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// productsUpdateCmd wires `fiken products update ...`.
func productsUpdateCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("update")
	fs.SetParent(parent)
	var (
		productID int64
		fromFile  string
	)
	fs.Int64Var(&productID, 0, "product-id", 0, "Product id to update (required)")
	fs.StringVar(&fromFile, 0, "from-file", "", "Path to JSON body (required)")
	return &ff.Command{
		Name:      "update",
		Usage:     "fiken products update --company <slug> --product-id <id> --from-file <path>",
		ShortHelp: "Update an existing product.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			body, e := ReadBodyFile[fiken.Product](fromFile)
			if e != nil {
				e.Op = ops.OpProductsUpdate
				return Renderer(ctx).Render(ops.Err[ops.ProductOut](e))
			}
			res := Client(ctx).ProductsUpdate(ctx, ops.ProductsUpdateIn{
				Company:   *sf.flagCompany,
				ProductID: productID,
				Body:      body,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// productsDeleteCmd wires `fiken products delete ...` (stub).
func productsDeleteCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	fs := ff.NewFlagSet("delete")
	fs.SetParent(parent)
	var productID int64
	fs.Int64Var(&productID, 0, "product-id", 0, "Product id to delete (required)")
	return &ff.Command{
		Name:      "delete",
		Usage:     "fiken products delete --company <slug> --product-id <id>",
		ShortHelp: "Delete a product.",
		Flags:     fs,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).ProductsDelete(ctx, ops.ProductsDeleteIn{
				Company:   *sf.flagCompany,
				ProductID: productID,
			})
			return Renderer(ctx).Render(res)
		},
	}
}

// productsSalesReportCmd wires `fiken products sales-report create ...`.
// The sales-report endpoint is POST (mutating in the OAS) and returns
// per-product aggregates over a date window.
func productsSalesReportCmd(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	set := ff.NewFlagSet("sales-report")
	set.SetParent(parent)

	createFS := ff.NewFlagSet("create")
	createFS.SetParent(set)
	var (
		from string
		to   string
	)
	createFS.StringVar(&from, 0, "from", "", "Start date inclusive, YYYY-MM-DD (required)")
	createFS.StringVar(&to, 0, "to", "", "End date inclusive, YYYY-MM-DD (required)")
	createCmd := &ff.Command{
		Name:      "create",
		Usage:     "fiken products sales-report create --company <slug> --from <date> --to <date>",
		ShortHelp: "Generate a product sales report for a date window.",
		Flags:     createFS,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).ProductsSalesReportCreate(ctx, ops.ProductsSalesReportCreateIn{
				Company: *sf.flagCompany,
				From:    ops.Date(from),
				To:      ops.Date(to),
			})
			return Renderer(ctx).Render(res)
		},
	}

	return &ff.Command{
		Name:        "sales-report",
		Usage:       "fiken products sales-report <subcommand>",
		ShortHelp:   "Build sales reports across the product catalog.",
		Flags:       set,
		Subcommands: []*ff.Command{createCmd},
	}
}
