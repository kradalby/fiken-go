package auth

import (
	"context"
	"errors"
	"testing"
)

type fixedSrc struct{ tok string }

func (s fixedSrc) Token(_ context.Context) (string, error) {
	if s.tok == "" {
		return "", ErrNotFound
	}
	return s.tok, nil
}

func TestChainSourceReturnsFirstNonEmpty(t *testing.T) {
	c := ChainSource{fixedSrc{""}, fixedSrc{"abc"}, fixedSrc{"def"}}
	got, err := c.Token(context.Background())
	if err != nil {
		t.Fatalf("Token: %v", err)
	}
	if got != "abc" {
		t.Fatalf("got %q want abc", got)
	}
}

func TestChainSourceErrNotFound(t *testing.T) {
	c := ChainSource{fixedSrc{""}, fixedSrc{""}}
	_, err := c.Token(context.Background())
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("err=%v want ErrNotFound", err)
	}
}

func TestFlagSourceEnvSource(t *testing.T) {
	if _, err := (FlagSource{Value: ""}).Token(context.Background()); !errors.Is(err, ErrNotFound) {
		t.Errorf("empty FlagSource should return ErrNotFound")
	}
	if got, err := (FlagSource{Value: "tok"}).Token(context.Background()); err != nil || got != "tok" {
		t.Errorf("FlagSource: got %q err %v", got, err)
	}

	t.Setenv("FIKEN_TOKEN", "envtok")
	if got, err := (EnvSource{Var: "FIKEN_TOKEN"}).Token(context.Background()); err != nil || got != "envtok" {
		t.Errorf("EnvSource: got %q err %v", got, err)
	}
}
