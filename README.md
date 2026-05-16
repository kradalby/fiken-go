# fiken-go

Go library, CLI, and MCP server for the [Fiken](https://fiken.no) API.

`fiken-go` exposes every Fiken REST endpoint through three frontends that share one operations layer:

1. A typed Go client library (drop-in for ad-hoc Fiken integrations).
2. A CLI: `fiken <tag> <op>`, human-readable + `--json`.
3. An MCP server: `fiken mcp`, served over stdio (default) or HTTP (`--transport=http`), with strict read-only mode by default.

## Install

```bash
# Nix
nix profile install github:kradalby/fiken-go

# Go
go install github.com/kradalby/fiken-go/cmd/fiken@latest
```

## Quick start

```bash
# Create a personal API token at:
#   https://fiken.no/foretak/<co>/api-tokens
fiken auth login

fiken companies list
fiken --json companies list | jq .ok.items[].slug
fiken --lang=nb companies list --help
fiken --profile work invoices list --company acme-as
```

## Localization

`--lang=nb` (or `--lang=no`) localizes **error messages** and **op descriptions exposed to MCP clients**. The `fiken <cmd> --help` output is rendered by the ff/v4 flag parser from Go struct tags and stays in English for now.

## MCP

```bash
# Default: stdio + read-only
fiken mcp

# Read-write + HTTP for a remote agent
fiken mcp --mode=read-write --transport=http --listen=:8765

# Claude Desktop / Code:
claude mcp add fiken -- fiken mcp
```

## Configuration

```toml
# ~/.config/fiken/config.toml
default_profile = "work"

[profiles.work]
company = "acme-as"
lang    = "nb"

[profiles.test]
company = "sandbox-co"
```

Override via env (`FIKEN_PROFILE`, `FIKEN_TOKEN`, `FIKEN_COMPANY`, `FIKEN_LANG`) or flags (`--profile`, `--token`, `--company`, `--lang`).

## Develop

```bash
nix develop
prek run --all-files
go test ./...
```

See `docs/superpowers/specs/2026-05-15-fiken-go-design.md` for the canonical design and `docs/superpowers/plans/` for the implementation plans.
