package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/peterbourgon/ff/v4"

	"github.com/kradalby/fiken-go/auth"
	"github.com/kradalby/fiken-go/ops"
	"github.com/kradalby/fiken-go/output"
)

// resolveAuthProfile returns the profile name for auth subcommands.
// It reads FIKEN_PROFILE from the environment and falls back to
// "default". The global --profile flag is parsed by the parent command
// but not re-exposed here (re-registering it on the auth FlagSet
// collides with the inherited entry).
func resolveAuthProfile() string {
	if v := os.Getenv("FIKEN_PROFILE"); v != "" {
		return v
	}
	return "default"
}

// AddAuth attaches `fiken auth {login,status,logout,list}`. stderr is
// accepted for symmetry with future subcommand registrars even though
// auth currently writes only to stdout.
func AddAuth(root *ff.Command, stdout, stderr io.Writer, sf *sessionFactory) error {
	_ = stderr
	set := ff.NewFlagSet("auth")
	if parentFS, ok := root.Flags.(*ff.FlagSet); ok {
		set.SetParent(parentFS)
	}

	authCmd := &ff.Command{
		Name:      "auth",
		Usage:     "fiken auth <subcommand>",
		ShortHelp: "Manage credentials.",
		Flags:     set,
	}

	loginSet := ff.NewFlagSet("login")
	loginSet.SetParent(set)
	loginCmd := &ff.Command{
		Name:      "login",
		Usage:     "fiken auth login",
		ShortHelp: "Prompt for a personal API token and store it.",
		Flags:     loginSet,
		Exec: func(ctx context.Context, _ []string) error {
			profile := resolveAuthProfile()
			b := sf.bundle
			lang := "en" // default; for full --lang plumbing see root flags
			_, _ = fmt.Fprintln(stdout, b.T(lang, "auth.login.prompt_url", nil))
			_, _ = fmt.Fprint(stdout, b.T(lang, "auth.login.prompt_paste", nil)+" ")
			tokBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
			_, _ = fmt.Fprintln(stdout)
			if err != nil {
				return fmt.Errorf("read token: %w", err)
			}
			tok := strings.TrimSpace(string(tokBytes))
			if tok == "" {
				return errors.New("empty token")
			}
			_, _ = fmt.Fprintln(stdout, b.T(lang, "auth.login.verifying", nil))
			if err := verifyToken(ctx, tok); err != nil {
				return fmt.Errorf("token verification failed: %w", err)
			}
			ks := auth.KeyringSource{Profile: profile}
			loc, err := ks.Save(auth.NewPersonal(tok))
			if err != nil {
				return fmt.Errorf("save: %w", err)
			}
			if loc == "keyring" {
				_, _ = fmt.Fprintln(stdout, b.T(lang, "auth.login.saved_keyring",
					map[string]any{"profile": profile}))
			} else {
				_, _ = fmt.Fprintln(stdout, b.T(lang, "auth.login.saved_file",
					map[string]any{"path": loc}))
			}
			return nil
		},
	}

	statusSet := ff.NewFlagSet("status")
	statusSet.SetParent(set)
	statusCmd := &ff.Command{
		Name:      "status",
		Usage:     "fiken auth status",
		ShortHelp: "Verify the stored token works.",
		Flags:     statusSet,
		Exec: func(ctx context.Context, _ []string) error {
			// Pre-resolve renderer because sf.Build may fail before
			// the Renderer is installed in ctx (e.g. config load error,
			// missing token). We want even those failures to honor --json
			// and render through the Result[T] envelope.
			var r output.Renderer
			if *sf.flagJSON {
				r = output.JSON(stdout)
			} else {
				bundle := sf.bundle
				lang := "en"
				if v := *sf.flagLang; v != "" {
					lang = v
				}
				r = output.Table(stdout, func(code, msg string) string {
					if v := bundle.T(lang, "error."+code+".short",
						map[string]any{"detail": msg}); v != "" {
						return v
					}
					return msg
				})
			}
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return r.Render(ops.Result[ops.UserOut]{
					Error: &ops.Error{
						Code:    ops.CodeAuthMissing,
						Message: err.Error(),
						Op:      "auth_status",
					},
				})
			}
			res := Client(ctx).UserGet(ctx, ops.UserGetIn{})
			return Renderer(ctx).Render(res)
		},
	}

	logoutSet := ff.NewFlagSet("logout")
	logoutSet.SetParent(set)
	logoutCmd := &ff.Command{
		Name:      "logout",
		Usage:     "fiken auth logout",
		ShortHelp: "Delete the stored token.",
		Flags:     logoutSet,
		Exec: func(_ context.Context, _ []string) error {
			profile := resolveAuthProfile()
			b := sf.bundle
			lang := "en"
			ks := auth.KeyringSource{Profile: profile}
			if err := ks.Delete(); err != nil {
				return err
			}
			_, _ = fmt.Fprintln(stdout, b.T(lang, "auth.logout.removed",
				map[string]any{"profile": profile}))
			return nil
		},
	}

	listSet := ff.NewFlagSet("list")
	listSet.SetParent(set)
	listCmd := &ff.Command{
		Name:      "list",
		Usage:     "fiken auth list",
		ShortHelp: "List configured profiles (placeholder).",
		Flags:     listSet,
		Exec: func(_ context.Context, _ []string) error {
			_, _ = fmt.Fprintln(stdout, "(profile listing not yet implemented in Plan B)")
			return nil
		},
	}

	authCmd.Subcommands = []*ff.Command{loginCmd, statusCmd, logoutCmd, listCmd}
	root.Subcommands = append(root.Subcommands, authCmd)
	return nil
}

// verifyToken hits /user against the canonical URL with the token.
// Production simplicity; tests use mockfiken via the ops layer.
func verifyToken(ctx context.Context, tok string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.fiken.no/api/v2/user", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	return fmt.Errorf("HTTP %d", resp.StatusCode)
}
