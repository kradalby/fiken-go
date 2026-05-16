package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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
		mode          string
		transport     string
		listen        string
		enableAtt     bool
		tsnetEnable   bool
		tsnetHostname string
		tsnetAuthKey  string
		tsnetAuthFile string
		tsnetStateDir string
	)
	set.StringVar(&mode, 0, "mode", "read-only", "read-only | read-write (ignored when --tsnet)")
	set.StringVar(&transport, 0, "transport", "stdio", "stdio | http (ignored when --tsnet)")
	set.StringVar(&listen, 0, "listen", ":8765", "HTTP listen address (ignored when --tsnet)")
	set.BoolVar(&enableAtt, 0, "enable-attachments", "Expose 6 multipart attachment ops")
	set.BoolVar(&tsnetEnable, 0, "tsnet", "Serve MCP over Tailscale (tsnet); disables stdio/HTTP")
	set.StringVar(&tsnetHostname, 0, "tsnet-hostname", "fiken-mcp", "Tailscale device name")
	set.StringVar(&tsnetAuthKey, 0, "tsnet-authkey", "", "Tailscale pre-auth key (mutex with --tsnet-authkey-file)")
	set.StringVar(&tsnetAuthFile, 0, "tsnet-authkey-file", "", "File containing a Tailscale pre-auth key")
	set.StringVar(&tsnetStateDir, 0, "tsnet-state-dir", "", "tsnet state directory (default $XDG_STATE_HOME/fiken-mcp/tsnet)")

	cmd := &ff.Command{
		Name:      "mcp",
		Usage:     "fiken mcp [--mode=...] [--transport=...] [--tsnet]",
		ShortHelp: "Run the MCP server.",
		Flags:     set,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}

			if tsnetEnable {
				return runTsnet(ctx, tsnetHostname, tsnetAuthKey, tsnetAuthFile, tsnetStateDir)
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

func runTsnet(ctx context.Context, hostname, authKey, authFile, stateDir string) error {
	if authKey != "" && authFile != "" {
		return fmt.Errorf("--tsnet-authkey and --tsnet-authkey-file are mutually exclusive")
	}
	if authFile != "" {
		b, err := os.ReadFile(authFile) //nolint:gosec // operator-supplied path is the documented mechanism for loading the tsnet auth key
		if err != nil {
			return fmt.Errorf("read --tsnet-authkey-file: %w", err)
		}
		authKey = strings.TrimSpace(string(b))
	}
	if stateDir == "" {
		stateDir = defaultTsnetStateDir()
	}
	if err := os.MkdirAll(stateDir, 0o700); err != nil {
		return fmt.Errorf("create tsnet state dir: %w", err)
	}
	// tsnet exposes every tool; per-request capability gates writes.
	srv, err := mcp.New(mcp.Options{
		Client:            Client(ctx),
		Mode:              mcp.ModeReadWrite,
		Bundle:            Bundle(ctx),
		Lang:              Lang(ctx),
		EnableAttachments: true,
		CapGated:          true,
	})
	if err != nil {
		return err
	}
	return mcp.RunTsnet(ctx, srv, mcp.TsnetOptions{
		Hostname: hostname,
		AuthKey:  authKey,
		StateDir: stateDir,
	})
}

func defaultTsnetStateDir() string {
	if x := os.Getenv("XDG_STATE_HOME"); x != "" {
		return filepath.Join(x, "fiken-mcp", "tsnet")
	}
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".local", "state", "fiken-mcp", "tsnet")
	}
	return ".tsnet-state"
}
