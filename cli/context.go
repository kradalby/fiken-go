// Package cli builds the ff/v4 command tree. context.go threads
// per-invocation values (Client, Renderer, Bundle, lang, stderr,
// verbosity) through context.Value lookups.
package cli

import (
	"context"
	"io"

	"github.com/kradalby/fiken-go/i18n"
	"github.com/kradalby/fiken-go/ops"
	"github.com/kradalby/fiken-go/output"
)

type ctxKey int

const (
	keyClient ctxKey = iota
	keyRenderer
	keyBundle
	keyLang
	keyStderr
	keyVerbosity
	keyProfile
)

// WithSession stuffs the per-invocation values into ctx.
func WithSession(
	ctx context.Context,
	c *ops.Client,
	r output.Renderer,
	b *i18n.Bundle,
	lang string,
	stderr io.Writer,
	verbosity int,
	profile string,
) context.Context {
	ctx = context.WithValue(ctx, keyClient, c)
	ctx = context.WithValue(ctx, keyRenderer, r)
	ctx = context.WithValue(ctx, keyBundle, b)
	ctx = context.WithValue(ctx, keyLang, lang)
	ctx = context.WithValue(ctx, keyStderr, stderr)
	ctx = context.WithValue(ctx, keyVerbosity, verbosity)
	ctx = context.WithValue(ctx, keyProfile, profile)
	return ctx
}

// Client returns the per-invocation *ops.Client stashed by WithSession.
func Client(ctx context.Context) *ops.Client { return ctx.Value(keyClient).(*ops.Client) }

// Renderer returns the per-invocation output.Renderer.
func Renderer(ctx context.Context) output.Renderer { return ctx.Value(keyRenderer).(output.Renderer) }

// Bundle returns the i18n bundle shared across the invocation.
func Bundle(ctx context.Context) *i18n.Bundle { return ctx.Value(keyBundle).(*i18n.Bundle) }

// Lang returns the resolved locale string ("en", "nb", ...).
func Lang(ctx context.Context) string { return ctx.Value(keyLang).(string) }

// Stderr returns the writer used for diagnostic output.
func Stderr(ctx context.Context) io.Writer { return ctx.Value(keyStderr).(io.Writer) }

// Verbosity returns the parsed -v count (0 = warn, 1 = info, 2+ = debug).
func Verbosity(ctx context.Context) int { return ctx.Value(keyVerbosity).(int) }

// ProfileName returns the active profile name resolved by the session factory.
func ProfileName(ctx context.Context) string { return ctx.Value(keyProfile).(string) }
