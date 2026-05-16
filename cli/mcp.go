package cli

import (
	"context"
	"io"

	"github.com/peterbourgon/ff/v4"

	"github.com/kradalby/fiken-go/mcp"
)

// AddMCP wires `fiken mcp` subcommand.
func AddMCP(root *ff.Command, stdout, stderr io.Writer, sf *sessionFactory) error {
	set := ff.NewFlagSet("mcp")
	if parentFS, ok := root.Flags.(*ff.FlagSet); ok {
		set.SetParent(parentFS)
	}
	var (
		mode      string
		transport string
		listen    string
		enableAtt bool
	)
	set.StringVar(&mode, 0, "mode", "read-only", "read-only | read-write")
	set.StringVar(&transport, 0, "transport", "stdio", "stdio | http")
	set.StringVar(&listen, 0, "listen", ":8765", "HTTP listen address")
	set.BoolVar(&enableAtt, 0, "enable-attachments", "Expose 6 multipart attachment ops")

	cmd := &ff.Command{
		Name:      "mcp",
		Usage:     "fiken mcp [--mode=...] [--transport=...]",
		ShortHelp: "Run the MCP server.",
		Flags:     set,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}

			modeVal := mcp.ModeReadOnly
			if mode == "read-write" {
				modeVal = mcp.ModeReadWrite
			}
			srv, err := mcp.New(mcp.Options{
				Client:            Client(ctx),
				Mode:              modeVal,
				Bundle:            Bundle(ctx),
				Lang:              Lang(ctx),
				EnableAttachments: enableAtt,
			})
			if err != nil {
				return err
			}
			if transport == "http" {
				return mcp.RunHTTP(ctx, srv, listen)
			}
			return mcp.RunStdio(ctx, srv)
		},
	}
	root.Subcommands = append(root.Subcommands, cmd)
	return nil
}
