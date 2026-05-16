package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"

	"github.com/kradalby/fiken-go/cli"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cmd, err := cli.Root(os.Stdout, os.Stderr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fiken: %v\n", err)
		os.Exit(1)
	}
	if err := cmd.ParseAndRun(ctx, os.Args[1:], ff.WithEnvVarPrefix("FIKEN")); err != nil {
		if errors.Is(err, ff.ErrHelp) {
			fmt.Fprintln(os.Stderr, ffhelp.Command(cmd))
			return
		}
		fmt.Fprintf(os.Stderr, "fiken: %v\n", err)
		os.Exit(1)
	}
}
