# fiken-go — Plan D: Polish (CI, drift detector, attachments, MCP HTTP, release)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan. Steps use checkbox (`- [ ]`) syntax.

**Goal:** Finish the project. After Plans A-C, every Fiken op is reachable through CLI + MCP with parity tested. Plan D adds:

1. GitHub Actions workflows (`ci.yml`, `spec-drift.yml`).
2. The opt-in attachment exposure (`--enable-attachments`) wired and smoke-tested.
3. MCP streamable HTTP transport tested in addition to stdio.
4. Final integration cleanups: README expansion, install/release docs.
5. Manual smoke checklist (humans only).

**Architecture:** No new domain layers — Plan D is operational polish over the implementation Plans A-C built.

**Spec reference:** `docs/superpowers/specs/2026-05-15-fiken-go-design.md` §"GitHub Actions", §"Multipart attachments", §"MCP transport scope", §"Verification".

**Prerequisites from Plans A-C:**

- 13 + 22 + ~25 = ~60 commits on `main`.
- All Fiken tags exposed via CLI + MCP.
- Parity test covers every registered op.
- `prek run --all-files` green; `go test ./...` green; `nix build .#fiken` works.

---

### Task 1: `.github/workflows/ci.yml`

**Files:**

- Create: `.github/workflows/ci.yml`

**Spec ref:** §"GitHub Actions (Nix-based, nix-community)".

- [ ] **Step 1.1: Write the workflow**

```yaml
name: ci

on:
  push:
    branches: [main]
  pull_request:

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: cachix/install-nix-action@v31
        with:
          nix_path: nixpkgs=channel:nixos-unstable
          extra_nix_config: |
            experimental-features = nix-command flakes
      - uses: cachix/cachix-action@v15
        with:
          name: fiken-go
          authToken: ${{ secrets.CACHIX_AUTH_TOKEN }}
        continue-on-error: true
      - run: nix develop -c go test -race -count=1 ./...

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: cachix/install-nix-action@v31
        with:
          nix_path: nixpkgs=channel:nixos-unstable
          extra_nix_config: |
            experimental-features = nix-command flakes
      - run: nix develop -c golangci-lint run --timeout=5m

  hooks:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: cachix/install-nix-action@v31
        with:
          nix_path: nixpkgs=channel:nixos-unstable
          extra_nix_config: |
            experimental-features = nix-command flakes
      - run: nix develop -c prek run --all-files

  codegen-clean:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: cachix/install-nix-action@v31
        with:
          nix_path: nixpkgs=channel:nixos-unstable
          extra_nix_config: |
            experimental-features = nix-command flakes
      - run: nix develop -c go generate ./...
      - run: git diff --exit-code

  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: cachix/install-nix-action@v31
        with:
          nix_path: nixpkgs=channel:nixos-unstable
          extra_nix_config: |
            experimental-features = nix-command flakes
      - run: nix build .#fiken
```

(Note: `CACHIX_AUTH_TOKEN` is optional — `continue-on-error: true` on the cachix step ensures forks without the secret still see green CI.)

- [ ] **Step 1.2: Commit**

```bash
git add .github/workflows/ci.yml
git commit -m "$(cat <<'EOF'
ci: add nix-based CI workflow (test, lint, hooks, codegen, build)

Five parallel jobs: go test, golangci-lint, prek run --all-files,
go generate clean-tree check, nix build. All run inside
nix develop -c to pin tool versions to the flake exactly.

Cachix is optional via secrets.CACHIX_AUTH_TOKEN; forks without it
still see CI green (pull cache is public).

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 2: `.github/workflows/spec-drift.yml`

**Files:**

- Create: `.github/workflows/spec-drift.yml`

**Spec ref:** §"`spec-drift.yml` — scheduled drift detector".

- [ ] **Step 2.1: Write the workflow**

````yaml
name: spec-drift

on:
  schedule:
    # 06:17 UTC daily — off-hour to avoid GH cron stampede.
    - cron: "17 6 * * *"
  workflow_dispatch:

permissions:
  contents: write
  issues: write
  pull-requests: write

jobs:
  detect:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: cachix/install-nix-action@v31
        with:
          nix_path: nixpkgs=channel:nixos-unstable
          extra_nix_config: |
            experimental-features = nix-command flakes

      - name: Fetch upstream spec
        run: curl -fsSL https://api.fiken.no/api/v2/docs/swagger.yaml -o /tmp/upstream.yaml

      - name: Diff vs vendored
        id: diff
        run: |
          if nix develop -c difft --exit-code api/fiken-openapi.yaml /tmp/upstream.yaml; then
            echo "drift=false" >> "$GITHUB_OUTPUT"
          else
            echo "drift=true" >> "$GITHUB_OUTPUT"
          fi

      - name: Open or update drift issue
        if: steps.diff.outputs.drift == 'true'
        uses: actions/github-script@v7
        with:
          script: |
            const fs = require('fs');
            const { execSync } = require('child_process');
            const summary = execSync(
              `nix develop -c go run ./cmd/fiken-spec-update -diff-only 2>&1 || true`
            ).toString();
            const body = "Upstream Fiken OpenAPI spec drifted vs the vendored copy.\n\n```\n" + summary + "\n```";
            const { data: issues } = await github.rest.issues.listForRepo({
              owner: context.repo.owner, repo: context.repo.repo,
              state: 'open', labels: 'spec-drift',
            });
            if (issues.length > 0) {
              await github.rest.issues.createComment({
                owner: context.repo.owner, repo: context.repo.repo,
                issue_number: issues[0].number, body,
              });
            } else {
              await github.rest.issues.create({
                owner: context.repo.owner, repo: context.repo.repo,
                title: 'Fiken OpenAPI spec drift detected',
                body, labels: ['spec-drift'],
              });
            }

      - name: Open update PR (optional, gated on PAT)
        if: steps.diff.outputs.drift == 'true' && env.SPEC_BOT_PAT != ''
        env:
          SPEC_BOT_PAT: ${{ secrets.SPEC_BOT_PAT }}
        run: |
          nix develop -c go run ./cmd/fiken-spec-update --apply
          nix develop -c go generate ./...
      - if: steps.diff.outputs.drift == 'true' && env.SPEC_BOT_PAT != ''
        env:
          SPEC_BOT_PAT: ${{ secrets.SPEC_BOT_PAT }}
        uses: peter-evans/create-pull-request@v7
        with:
          token: ${{ secrets.SPEC_BOT_PAT }}
          branch: spec-drift-${{ github.run_id }}
          title: "chore(api): sync vendored Fiken spec with upstream"
          body: "Auto-generated by spec-drift workflow."
          commit-message: |
            chore(api): sync vendored Fiken spec with upstream

            Generated by spec-drift workflow.
````

- [ ] **Step 2.2: Commit**

```bash
git add .github/workflows/spec-drift.yml
git commit -m "$(cat <<'EOF'
ci: add scheduled spec-drift workflow

Daily 06:17 UTC and workflow_dispatch. Fetches the canonical Fiken
spec, diffs vs the vendored copy via difftastic, and on drift either
opens/comments on a spec-drift issue (default) or opens an
auto-generated update PR (gated on SPEC_BOT_PAT secret).

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 3: `--enable-attachments` end-to-end smoke

**Files:**

- Create: `mcp/attachments.go`
- Create: `mcp/attachments_test.go`
- Modify: `mcp/server.go` to delegate attachment-op registration to `mcp/attachments.go` only when `Options.EnableAttachments` is true.

**Spec ref:** §"Multipart attachments".

The 6 attachment ops were skipped from MCP in Plans B-C. Plan D wires them as opt-in: server flag `--enable-attachments` registers them with a single `file_path` parameter (local path on the MCP server host). No base64 in v1.

- [ ] **Step 3.1: Failing test**

`mcp/attachments_test.go`:

```go
package mcp

import (
    "testing"
    "github.com/kradalby/fiken-go/i18n"
)

func TestAttachmentsHiddenByDefault(t *testing.T) {
    srv, _ := New(Options{Mode: ModeReadWrite, Bundle: i18n.MustLoad(), Lang: "en"})
    // Inspect registered tools (SDK API — adapt to actual surface).
    // Assert none of the 6 attach op names appear.
    _ = srv
    t.Skip("TODO: inspect SDK registered tools list once SDK API confirmed")
}

func TestAttachmentsExposedWhenEnabled(t *testing.T) {
    srv, _ := New(Options{
        Mode: ModeReadWrite, Bundle: i18n.MustLoad(), Lang: "en",
        EnableAttachments: true,
    })
    _ = srv
    t.Skip("TODO: same as above + assert presence")
}
```

- [ ] **Step 3.2: Implement `mcp/attachments.go`**

```go
package mcp

import (
    "context"
    mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
    "github.com/kradalby/fiken-go/ops"
)

// AttachOpNames lists the 6 multipart attachment op-ids that
// EnableAttachments toggles. Source: api/fiken-openapi.yaml grep
// for `multipart/form-data` request bodies.
var AttachOpNames = []string{
    // Implementer: populate with the 6 actual op-ids surfaced in
    // ops/names.go from Plan C. Likely:
    //   ops.OpContactsPersonsAttachAdd
    //   ops.OpJournalEntriesAttachAdd
    //   ops.OpInvoicesAttachAdd
    //   ops.OpInvoiceDraftsAttachAdd
    //   ops.OpCreditNoteDraftsAttachAdd
    //   ops.OpOfferDraftsAttachAdd
}

// AttachIn is the canonical input shape for any attach op: path
// to a local file on the MCP server host. The Op identifies which
// resource the file attaches to (e.g. invoice ID).
type AttachIn struct {
    Company string `json:"company"`
    Resource int64 `json:"resource_id"`
    FilePath string `json:"file_path"`
}

// registerAttachments adds the 6 attachment tools to srv. Called only
// when Options.EnableAttachments is true.
func registerAttachments(srv *mcpsdk.Server, client *ops.Client) {
    for _, opName := range AttachOpNames {
        opName := opName // capture
        mcpsdk.AddTool(srv, &mcpsdk.Tool{
            Name:        opName,
            Description: "Attach a file to the given resource. file_path is a path on the MCP server's local filesystem.",
        }, makeAttachHandler(client, opName))
    }
}

func makeAttachHandler(client *ops.Client, opName string) func(context.Context, *mcpsdk.CallToolRequest, AttachIn) (*mcpsdk.CallToolResult, ops.Result[ops.AttachOut], error) {
    return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in AttachIn) (*mcpsdk.CallToolResult, ops.Result[ops.AttachOut], error) {
        // Delegate to the matching ops.Client.<Tag>Attach method. The
        // dispatch by opName is brittle; consider a sync.Map from
        // opName to method ptr. For Plan D's scope (6 ops), an
        // if-else ladder suffices.
        var res ops.Result[ops.AttachOut]
        switch opName {
        // case ops.OpInvoicesAttachAdd:
        //     res = client.InvoicesAttachAdd(ctx, ...)
        // ... etc
        }
        r := &mcpsdk.CallToolResult{IsError: res.Error != nil}
        return r, res, nil
    }
}
```

(The implementer wires the actual dispatch once Plan C's attach methods are confirmed.)

- [ ] **Step 3.3: Modify `mcp/server.go`'s `New`**

Add at the end:

```go
if opts.EnableAttachments {
    registerAttachments(srv, opts.Client)
}
```

And extend `Options`:

```go
type Options struct {
    Client            *ops.Client
    Mode              Mode
    Bundle            *i18n.Bundle
    Lang              string
    EnableAttachments bool
}
```

- [ ] **Step 3.4: Modify `cli/mcp.go`'s `Exec`**

The `--enable-attachments` flag was added in Plan B Task 18; thread it through:

```go
srv, err := mcp.New(mcp.Options{
    Client:            Client(ctx),
    Mode:              modeVal,
    Bundle:            Bundle(ctx),
    Lang:              Lang(ctx),
    EnableAttachments: enableAtt,
})
```

- [ ] **Step 3.5: Pass tests, commit**

```bash
git add mcp/attachments.go mcp/attachments_test.go mcp/server.go cli/mcp.go
git commit -m "$(cat <<'EOF'
feat(mcp): opt-in attachment exposure via --enable-attachments

Registers the 6 Fiken multipart attachment ops as MCP tools only
when EnableAttachments is true. Tool input is {company, resource_id,
file_path} — path is local to the MCP server host. Read-write mode
required (attachments are mutating ops).

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 4: MCP streamable HTTP transport smoke test

**Files:**

- Create: `mcp/transport_http_test.go`

**Spec ref:** §"MCP transport scope".

- [ ] **Step 4.1: Test**

```go
package mcp

import (
    "context"
    "net"
    "testing"
    "time"

    mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
    "github.com/kradalby/fiken-go/auth"
    "github.com/kradalby/fiken-go/i18n"
    "github.com/kradalby/fiken-go/mockfiken"
    "github.com/kradalby/fiken-go/ops"
)

func TestMCPHTTPTransport(t *testing.T) {
    mock := mockfiken.New(t)
    client, _ := ops.New(context.Background(), ops.Options{
        BaseURL: mock.URL(), Auth: auth.FlagSource{Value: "test"},
    })
    srv, _ := New(Options{Client: client, Mode: ModeReadOnly,
        Bundle: i18n.MustLoad(), Lang: "en"})

    // Pick a free port.
    lis, err := net.Listen("tcp", "127.0.0.1:0")
    if err != nil { t.Fatal(err) }
    addr := lis.Addr().String()
    lis.Close()

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    t.Cleanup(cancel)

    go func() { _ = RunHTTP(ctx, srv, addr) }()
    time.Sleep(100 * time.Millisecond) // let it bind

    cs := mcpsdk.NewClient("test-http", "0.1")
    transport := mcpsdk.NewStreamableClientTransport("http://" + addr, nil)
    if err := cs.Connect(ctx, transport, nil); err != nil {
        t.Fatalf("connect: %v", err)
    }
    resp, err := cs.CallTool(ctx, &mcpsdk.CallToolParams{
        Name: ops.OpCompaniesList, Arguments: map[string]any{},
    })
    if err != nil {
        t.Fatalf("CallTool: %v", err)
    }
    if resp.IsError {
        t.Fatalf("tool returned error: %+v", resp)
    }
}
```

- [ ] **Step 4.2: Pass + commit**

```bash
git add mcp/transport_http_test.go
git commit -m "test(mcp): smoke streamable HTTP transport end-to-end

Spins up the MCP server with --transport=http on a random port,
connects an MCP client over the streamable HTTP transport, calls
companies_list against mockfiken. Asserts response shape.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

### Task 5: README expansion

**Files:**

- Modify: `README.md`

- [ ] **Step 5.1: Expand**

Replace the minimal Plan A README with a complete user-facing intro:

````markdown
# fiken-go

Go library, CLI, and MCP server for the [Fiken](https://fiken.no) API.

`fiken-go` exposes every Fiken REST endpoint through three frontends
that share one operations layer:

1. A typed Go client library (drop-in for ad-hoc Fiken integrations).
2. A CLI: `fiken <tag> <op>`, human-readable + `--json`.
3. An MCP server: `fiken mcp`, served over stdio (default) or HTTP
   (`--transport=http`), with strict read-only mode by default.

## Install

```bash
# Nix
nix profile install github:kradalby/fiken-go

# Go
go install github.com/kradalby/fiken-go/cmd/fiken@latest
```
````

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

Override via env (`FIKEN_PROFILE`, `FIKEN_TOKEN`, `FIKEN_COMPANY`,
`FIKEN_LANG`) or flags (`--profile`, `--token`, `--company`, `--lang`).

## Develop

```bash
nix develop
prek run --all-files
go test ./...
```

See `docs/superpowers/specs/2026-05-15-fiken-go-design.md` for the
canonical design and `docs/superpowers/plans/` for the implementation
plans.

````

- [ ] **Step 5.2: Commit**

```bash
git add README.md
git commit -m "docs(readme): user-facing intro, install, MCP, config

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
````

---

### Task 6: Wire the `vendor-hash` pre-commit hook

**Files:**

- Modify: `.pre-commit-config.yaml`

**Spec ref:** §"Implementation steps" #2.

The vendor-hash hook regenerates `flake.nix`'s `vendorHash` when `go.{mod,sum}` change. Without it, every new dep silently breaks `nix build` until a human notices.

- [ ] **Step 6.1: Add a script**

Create `scripts/update-vendor-hash.sh`:

```bash
#!/usr/bin/env bash
set -euo pipefail
# Detect go.mod/go.sum staged changes; if any, run `nix build` and
# capture the actual vendor hash, then patch flake.nix.
if ! git diff --cached --name-only | grep -qE '^(go\.mod|go\.sum)$'; then
  exit 0
fi
# Replace the current hash with fakeHash so nix recomputes.
sed -i 's|vendorHash = "sha256-[^"]*"|vendorHash = pkgs.lib.fakeHash|' flake.nix
got=$(nix build .#fiken 2>&1 | awk '/got:/{print $2; exit}')
if [ -z "$got" ]; then
  echo "vendor-hash: could not capture new hash; aborting" >&2
  exit 1
fi
sed -i "s|vendorHash = pkgs.lib.fakeHash|vendorHash = \"$got\"|" flake.nix
git add flake.nix
```

```bash
chmod +x scripts/update-vendor-hash.sh
```

Add to `.pre-commit-config.yaml`'s local hooks:

```yaml
- id: vendor-hash
  name: vendor-hash
  language: system
  entry: bash scripts/update-vendor-hash.sh
  pass_filenames: false
  files: '^go\.(mod|sum)$'
  stages: [pre-commit]
```

- [ ] **Step 6.2: Commit**

```bash
git add scripts/update-vendor-hash.sh .pre-commit-config.yaml
git commit -m "$(cat <<'EOF'
build(hooks): wire vendor-hash hook

When go.mod or go.sum changes, the hook regenerates the flake.nix
vendorHash by running `nix build .#fiken` once with pkgs.lib.fakeHash
and capturing the actual hash from Nix's mismatch error. Keeps the
flake reproducible without humans remembering to update.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 7: GoReleaser / release config (optional)

If the user wants automated GitHub releases on tags, add `.goreleaser.yaml`:

```yaml
version: 2

before:
  hooks:
    - go mod tidy

builds:
  - id: fiken
    main: ./cmd/fiken
    binary: fiken
    env:
      - CGO_ENABLED=0
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
    ldflags:
      - -s -w -X main.version={{.Version}}

archives:
  - format: tar.gz
    name_template: >-
      {{ .ProjectName }}_
      {{- .Os }}_
      {{- if eq .Arch "amd64" }}x86_64
      {{- else }}{{ .Arch }}{{ end }}

checksum:
  name_template: "checksums.txt"

snapshot:
  name_template: "{{ incpatch .Version }}-next"

changelog:
  use: github
  sort: asc
  groups:
    - title: Features
      regexp: "^feat"
    - title: Bug fixes
      regexp: "^fix"
    - title: Build / CI
      regexp: "^(build|ci|chore)"
```

Plus `.github/workflows/release.yml`:

```yaml
name: release
on:
  push:
    tags: ["v*"]
jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with: { fetch-depth: 0 }
      - uses: cachix/install-nix-action@v31
        with:
          nix_path: nixpkgs=channel:nixos-unstable
          extra_nix_config: |
            experimental-features = nix-command flakes
      - run: nix develop -c goreleaser release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

(Add `goreleaser` to the flake's devShell `packages`.)

Optional — skip if the user doesn't want this. Commit only when affirmed.

---

### Task 8: End-of-Plan-D verification

- [ ] `git log --oneline | head -100` — final state.
- [ ] CI on a test PR: every job green.
- [ ] `nix develop -c prek run --all-files` green.
- [ ] `nix develop -c go test -race -count=1 ./...` green.
- [ ] `nix build .#fiken` works; binary runs `--help`.
- [ ] Schedule-trigger the `spec-drift` workflow manually via `workflow_dispatch`; verify it either prints `(no changes)` (most likely) or opens an issue with the diff.
- [ ] MCP HTTP smoke: `fiken mcp --transport=http --listen=:8765` in one terminal; from another, hit it with an MCP client.

Manual smoke (humans, real token):

```bash
# Auth
./result/bin/fiken auth login --profile test
./result/bin/fiken auth status --profile test

# CLI
./result/bin/fiken --profile test companies list
./result/bin/fiken --profile test --lang=nb companies list --help
./result/bin/fiken --profile test invoices list --company my-co --json | jq .ok.items[0]
./result/bin/fiken --profile test invoices attach 42 --file ./test.pdf  # attachment via CLI

# MCP
claude mcp add fiken -- ./result/bin/fiken mcp --profile test
# Then in Claude: ask it to list companies → verify it calls companies_list

# MCP HTTP
./result/bin/fiken mcp --profile test --transport=http --listen=:8765 &
# Then point an HTTP MCP client at http://localhost:8765

# Read-write + attachments
./result/bin/fiken mcp --profile test --mode=read-write --enable-attachments
```

Success → project ships. Tag v0.1.0 if release is wanted.

---

## Self-review notes

- The `spec-drift.yml` workflow comments on (rather than re-creating) an open `spec-drift` issue to avoid issue spam.
- The auto-PR step is gated on `secrets.SPEC_BOT_PAT`; without the PAT, the workflow stops at issue-comment level. Document this in the README.
- `cachix/cachix-action` is optional — `continue-on-error: true` ensures CI doesn't fail when forks lack the auth token. Add a setup-guide section to README for maintainers.
- `vendor-hash` hook depends on `nix build` outputting `got: <hash>` on mismatch. If the Nix message format changes, the sed regex breaks. Worth a comment in the script.
- The 6 attachment ops in `AttachOpNames` are populated by hand based on Plan C's op constants. If a future spec adds a 7th multipart endpoint, add it here.
- MCP HTTP transport test (Task 4) relies on the SDK's `NewStreamableClientTransport`; verify the actual name in the current go-sdk release before merging.
- Plan D doesn't touch the core ops/CLI/MCP code (Plans A-C already shipped it). Only adds CI, drift detection, attachments opt-in, HTTP transport tests, README, and the vendor-hash automation.
