// Package auth resolves the Fiken personal API token for the current
// session. Token sources are arranged into a ChainSource that returns
// the first non-empty value. The resolution order — CLI flag, env,
// keyring, file — is configured by the caller in cli/root.go.
package auth

import (
	"context"
	"errors"
	"os"
)

// ErrNotFound is returned by a Source when it has no token to offer.
var ErrNotFound = errors.New("auth: no token found")

// Source produces a personal API token for an outgoing Fiken request.
type Source interface {
	Token(ctx context.Context) (string, error)
}

// ChainSource calls each Source in order and returns the first
// non-empty token. Any non-ErrNotFound error short-circuits.
type ChainSource []Source

// Token implements Source.
func (c ChainSource) Token(ctx context.Context) (string, error) {
	for _, s := range c {
		tok, err := s.Token(ctx)
		if err == nil {
			return tok, nil
		}
		if !errors.Is(err, ErrNotFound) {
			return "", err
		}
	}
	return "", ErrNotFound
}

// FlagSource carries a token passed via --token flag.
type FlagSource struct{ Value string }

// Token implements Source.
func (f FlagSource) Token(_ context.Context) (string, error) {
	if f.Value == "" {
		return "", ErrNotFound
	}
	return f.Value, nil
}

// EnvSource reads a token from the named env var.
type EnvSource struct{ Var string }

// Token implements Source.
func (e EnvSource) Token(_ context.Context) (string, error) {
	v := os.Getenv(e.Var)
	if v == "" {
		return "", ErrNotFound
	}
	return v, nil
}
