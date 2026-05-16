# fiken-go

Go library, CLI, and MCP server for the [Fiken](https://fiken.no) API.

## Install

```bash
# Nix
nix profile install github:kradalby/fiken-go

# Go
go install github.com/kradalby/fiken-go/cmd/fiken@latest
```

## Usage

Get a token in Fiken: _Rediger konto → Sikkerhet → Personlige API-nøkler_.

```bash
fiken auth login            # paste token, stored in OS keyring
fiken companies list
fiken --json invoices list --company acme-as
fiken --lang=nb companies list --help
```

## MCP

```bash
fiken mcp                                                        # stdio, read-only
fiken mcp --mode=read-write --transport=http --listen=:8765      # HTTP
fiken mcp --tsnet --tsnet-hostname fiken-mcp                     # tailnet

# Claude Desktop / Code:
claude mcp add fiken -- fiken mcp
```

Over tsnet, reads are implicit for any tailnet peer. Writes require a Tailscale ACL grant under `kradalby.no/cap/fiken-mcp` with `{"write": true}`.

## NixOS

`nixosModules.fiken-mcp` + `packages.fiken-mcp`. See [`nix/example-configuration.nix`](nix/example-configuration.nix).

## Config

```toml
# ~/.config/fiken/config.toml
default_profile = "work"

[profiles.work]
company = "acme-as"
lang    = "nb"
```

Env: `FIKEN_PROFILE`, `FIKEN_TOKEN`, `FIKEN_COMPANY`, `FIKEN_LANG`. Flags: same names with `--`.

## Develop

```bash
nix develop
prek run --all-files
go test ./...
```

Design notes: `docs/superpowers/specs/2026-05-15-fiken-go-design.md`.
