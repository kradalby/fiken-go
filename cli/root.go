package cli

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"

	"github.com/kradalby/fiken-go/auth"
	"github.com/kradalby/fiken-go/config"
	"github.com/kradalby/fiken-go/i18n"
	"github.com/kradalby/fiken-go/ops"
	"github.com/kradalby/fiken-go/output"
)

// Root builds the fiken root command tree. stdout receives Renderer
// output, stderr receives logs + help. Subcommand wiring
// (auth/companies/mcp) lands in Tasks 15-18.
func Root(stdout, stderr io.Writer) (*ff.Command, error) {
	rootSet := ff.NewFlagSet("fiken")

	var (
		flagConfig  string
		flagProfile string
		flagToken   string
		flagCompany string
		flagLang    string
		flagJSON    bool
		flagLogJSON bool
		flagV       int
	)
	rootSet.StringVar(&flagConfig, 0, "config", defaultConfigPath(), "Path to config TOML")
	rootSet.StringVar(&flagProfile, 0, "profile", "", "Profile name (overrides FIKEN_PROFILE)")
	rootSet.StringVar(&flagToken, 0, "token", "", "Personal API token")
	rootSet.StringVar(&flagCompany, 0, "company", "", "Default company slug")
	rootSet.StringVar(&flagLang, 0, "lang", "", "Locale: en, nb, no")
	rootSet.BoolVar(&flagJSON, 0, "json", "Emit ops.Result[T] envelope as JSON")
	rootSet.BoolVar(&flagLogJSON, 0, "log-json", "Log to stderr in JSON")
	rootSet.IntVar(&flagV, 'v', "verbose", 0, "Increase log verbosity (1=info, 2=debug)")

	bundle := i18n.MustLoad()

	root := &ff.Command{
		Name:      "fiken",
		Usage:     "fiken [global flags] <subcommand>",
		ShortHelp: "Go library, CLI, and MCP server for the Fiken API",
		Flags:     rootSet,
	}
	root.Exec = func(_ context.Context, _ []string) error {
		_, _ = fmt.Fprintln(stderr, ffhelp.Command(root).String())
		return nil
	}

	sf := &sessionFactory{
		bundle:      bundle,
		flagJSON:    &flagJSON,
		flagConfig:  &flagConfig,
		flagProfile: &flagProfile,
		flagToken:   &flagToken,
		flagCompany: &flagCompany,
		flagLang:    &flagLang,
		flagV:       &flagV,
		flagLogJSON: &flagLogJSON,
	}

	if err := AddAuth(root, stdout, stderr, sf); err != nil {
		return nil, err
	}
	if err := AddCompanies(root, stdout, stderr, sf); err != nil {
		return nil, err
	}
	if err := AddContacts(root, stdout, stderr, sf); err != nil {
		return nil, err
	}
	if err := AddAccounts(root, stdout, stderr, sf); err != nil {
		return nil, err
	}
	if err := AddBankAccounts(root, stdout, stderr, sf); err != nil {
		return nil, err
	}
	if err := AddJournalEntries(root, stdout, stderr, sf); err != nil {
		return nil, err
	}
	if err := AddTransactions(root, stdout, stderr, sf); err != nil {
		return nil, err
	}
	if err := AddInvoices(root, stdout, stderr, sf); err != nil {
		return nil, err
	}
	if err := AddCreditNotes(root, stdout, stderr, sf); err != nil {
		return nil, err
	}
	if err := AddOffers(root, stdout, stderr, sf); err != nil {
		return nil, err
	}
	if err := AddOrderConfirmations(root, stdout, stderr, sf); err != nil {
		return nil, err
	}
	if err := AddProducts(root, stdout, stderr, sf); err != nil {
		return nil, err
	}
	if err := AddSales(root, stdout, stderr, sf); err != nil {
		return nil, err
	}
	if err := AddPurchases(root, stdout, stderr, sf); err != nil {
		return nil, err
	}
	if err := AddInbox(root, stdout, stderr, sf); err != nil {
		return nil, err
	}
	if err := AddProjects(root, stdout, stderr, sf); err != nil {
		return nil, err
	}
	if err := AddUser(root, stdout, stderr, sf); err != nil {
		return nil, err
	}
	if err := AddAccountBalances(root, stdout, stderr, sf); err != nil {
		return nil, err
	}
	if err := AddBankBalances(root, stdout, stderr, sf); err != nil {
		return nil, err
	}
	if err := AddGroups(root, stdout, stderr, sf); err != nil {
		return nil, err
	}
	if err := AddActivities(root, stdout, stderr, sf); err != nil {
		return nil, err
	}
	if err := AddTimeUsers(root, stdout, stderr, sf); err != nil {
		return nil, err
	}
	if err := AddTimeEntries(root, stdout, stderr, sf); err != nil {
		return nil, err
	}
	if err := AddMCP(root, stdout, stderr, sf); err != nil {
		return nil, err
	}

	return root, nil
}

// sessionFactory holds pointers to the parsed global flags + the
// shared i18n bundle. Subcommand Exec funcs call Build to materialize
// the per-invocation ops.Client + Renderer and inject them via
// WithSession.
type sessionFactory struct {
	bundle      *i18n.Bundle
	flagJSON    *bool
	flagConfig  *string
	flagProfile *string
	flagToken   *string
	flagCompany *string
	flagLang    *string
	flagV       *int
	flagLogJSON *bool
}

// Build resolves a Profile from config + flags, constructs ops.Client +
// Renderer, configures slog, and returns ctx via WithSession.
func (sf *sessionFactory) Build(ctx context.Context, stdout, stderr io.Writer) (context.Context, error) {
	cfg, err := config.Load(*sf.flagConfig, map[string]string{
		"profile": *sf.flagProfile,
		"token":   *sf.flagToken,
		"company": *sf.flagCompany,
		"lang":    *sf.flagLang,
	})
	if err != nil {
		return ctx, fmt.Errorf("config load: %w", err)
	}
	prof, ok := cfg.Resolve(*sf.flagProfile)
	if !ok {
		return ctx, fmt.Errorf("profile %q not found", *sf.flagProfile)
	}
	lang := prof.Lang
	if lang == "" {
		lang = "en"
	}

	src := auth.ChainSource{
		auth.FlagSource{Value: *sf.flagToken},
		auth.EnvSource{Var: "FIKEN_TOKEN"},
		auth.KeyringSource{Profile: cfg.DefaultProfile},
	}
	if prof.Token != "" {
		src = append(auth.ChainSource{auth.FlagSource{Value: prof.Token}}, src...)
	}

	client, err := ops.New(ctx, ops.Options{
		BaseURL: os.Getenv("FIKEN_API_URL"),
		Auth:    src,
		Company: prof.Company,
	})
	if err != nil {
		return ctx, err
	}

	var renderer output.Renderer
	if *sf.flagJSON {
		renderer = output.JSON(stdout)
	} else {
		renderer = output.Table(stdout, func(code, msg string) string {
			if v := sf.bundle.T(lang, "error."+code+".short", map[string]any{"detail": msg}); v != "" {
				return v
			}
			return msg
		})
	}

	level := slog.LevelWarn
	switch {
	case *sf.flagV >= 2:
		level = slog.LevelDebug
	case *sf.flagV == 1:
		level = slog.LevelInfo
	}
	var handler slog.Handler
	if *sf.flagLogJSON {
		handler = slog.NewJSONHandler(stderr, &slog.HandlerOptions{Level: level})
	} else {
		handler = slog.NewTextHandler(stderr, &slog.HandlerOptions{Level: level})
	}
	slog.SetDefault(slog.New(handler))

	profile := cfg.DefaultProfile
	if *sf.flagProfile != "" {
		profile = *sf.flagProfile
	}
	return WithSession(ctx, client, renderer, sf.bundle, lang, stderr, *sf.flagV, profile), nil
}

func defaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "fiken", "config.toml")
}
