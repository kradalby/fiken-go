# `fiken-go` — Go library, CLI, and MCP server for Fiken

Design document. Folds in the 17 gap resolutions from the 2026-05-15
brainstorm session into one canonical plan. Supersedes the earlier
plan draft.

## Context

Fiken is a Norwegian bookkeeping/invoicing platform with a ~70-operation
OpenAPI 3.0.2 API (`https://api.fiken.no/api/v2`). We want a single Go
module that provides:

1. A typed Go **client library** (replacement for the ad-hoc `sfiber/fiken`
   package; concrete migration of `sfiber` itself is out of scope for this
   project).
2. A **CLI** with human-readable + `--json` output for every command, secure
   credential storage, and bilingual (en / nb) help and messages.
3. An **MCP server** exposing the same operations to agents, with strict
   `--read-only` and `--read-write` modes, served over stdio or streamable
   HTTP.

The design point is that both CLI subcommands and MCP tools call the same
underlying operations layer, so adding a new Fiken endpoint surfaces it in
both frontends with one edit. Every public string identifier (op name,
flag, env var, message key) is a shared constant so naming stays in
lockstep.

## Execution policy

### Hard rules (non-negotiable)

1. **Pre-commit hooks installed and green before the first non-trivial
   commit.** The very first non-trivial commit of the project lands
   `.pre-commit-config.yaml`, runs `prek install`, and confirms
   `prek run --all-files` is clean. Every `/commit` after that is gated by
   the real hook set. No "wire it up later", no merging on red hooks,
   no `--no-verify` shortcuts. Design and brainstorm docs predating the
   hooks are exempt from the linter pass since they are documentation, not
   code.
2. **ogen is the only code generator.** Client, server, types, spec
   embedding — all produced by ogen from `api/fiken-openapi.yaml`. No
   kin-openapi at runtime, no oapi-codegen, no swag, no hand-rolled struct
   mirrors of response shapes. Codegen scaffolding tools (i18n template
   seed, `mutating.gen.go`, OAS unit-format lint) are build-time helpers,
   not parallel code generators.

### Stop and ask whenever in doubt

This plan is the destination, not a script — where the implementer hits
any of the following, pause and surface a question to the user before
guessing:

- Ambiguity in a step (multiple valid readings of an instruction).
- A library / tool used here that turns out to be archived, broken, or the
  wrong fit for what the step actually needs.
- An upstream API/spec field whose meaning, units, or nullability isn't
  obvious from the OAS spec or Fiken docs (especially monetary, tax,
  VAT-code, and date/timezone fields).
- A naming or layout choice not pinned by the plan (op-name, flag spelling,
  i18n key shape, JSON tag) that would be costly to rename later.
- Any place where the design's invariants (single `Result[T]` envelope,
  int64 øre, etc.) would force an awkward shape — propose an alternative
  and wait, don't paper over it.
- A test that fails for a reason that isn't immediately understood. Trace
  to the cause; don't paper over with retries, `t.Skip`, or relaxed
  assertions.
- Anything that would touch files explicitly out of scope (e.g. editing
  generated `fiken/` by hand, adding `pkg/` or `internal/` dirs).

Cost of asking: a short message. Cost of guessing wrong: rework that
cascades through dozens of ops × 2 frontends. Always ask.

## Architecture

```mermaid
flowchart TD
    Spec[api/fiken-openapi.yaml<br/>vendored OAS 3.0.2]
    Spec -- "ogen (go generate)" --> Client[fiken/<br/>typed client + Handler iface + models]
    Spec -- "scaffold tool" --> I18nSeed[i18n/locales/en.template.toml<br/>not loaded at runtime]
    Spec -- "mutating-gen tool" --> MutGen[ops/mutating.gen.go]
    Spec -- "unit-lint tool" --> CIGate[CI: unit/format lint]

    SpecTool[cmd/fiken-spec-update<br/>fetch + difftastic diff]
    SpecTool -. updates .-> Spec

    Config[config/<br/>koanf: file + env + flag<br/>named profiles]
    Auth[auth/<br/>token sources:<br/>flag > env > keyring > file]
    Config --> Auth

    I18n[i18n/<br/>en + nb bundle]
    Output[output/<br/>table + json renderers]

    Ops[ops/<br/>domain operations<br/>Result[T] envelope<br/>shared op-name constants<br/>typed In/Out structs<br/>units invariant test]
    Client --> Ops
    Auth --> Ops
    MutGen --> Ops

    CLI[cli/<br/>ff/v4 command tree<br/>1 subcmd : 1 ops call]
    MCP[mcp/<br/>go-sdk server<br/>stdio + streamable HTTP<br/>1 tool : 1 ops call<br/>read-only filter]
    Ops --> CLI
    Ops --> MCP
    I18n --> CLI
    I18n --> MCP
    Output --> CLI

    Main[cmd/fiken/main.go]
    CLI --> Main
    MCP -. "fiken mcp" subcmd .-> CLI

    MockSpec[Spec-driven mock server<br/>ogen Handler + override registry] -. e2e .-> Ops
```

## Repo layout

```
api/
  fiken-openapi.yaml              vendored OAS 3.0.2 spec (from official source)
  SOURCE.txt                      provenance: source URL + SHA-256

fiken/                            ogen-generated client + Handler + models
  ogen.yml                        codegen config
  doc.go                          //go:generate directive
  (rest generated)

auth/
  auth.go                         Source iface, ChainSource
  credential.go                   Credential storage value (bare token v1)
  keyring.go                      go-keyring wrapper

config/
  config.go                       koanf loader, Profile, Config

output/
  output.go                       Renderer iface
  table.go                        text/tabwriter renderer
  json.go                         encoding/json renderer (same bytes MCP returns)

i18n/
  i18n.go                         bundle + T() helper
  locales/
    en.toml                       hand-authored
    en.template.toml              scaffold-generated, gitignored (or committed
                                  for diff visibility — pick one), NOT loaded
    nb.toml                       hand-authored

ops/                              domain operations layer
  names.go                        op-name consts + Registry (Mutating sourced
                                  from mutating.gen.go, not hand-marked)
  mutating.gen.go                 generated from OAS HTTP method
  result.go                       Result[T any] envelope + Error type
  errors.go                       ogen-error → ops.Error mapping
  date.go                         ops.Date civil type
  paging.go                       ListMeta, ListOut[T any]
  ratelimit.go                    sem=1 + 429-aware backoff RoundTripper
  units_test.go                   reflect-walk invariant test
  companies.go
  accounts.go
  bank_accounts.go
  contacts.go
  journal_entries.go
  transactions.go
  invoices.go
  credit_notes.go
  offers.go
  order_confirmations.go
  products.go
  sales.go
  purchases.go
  inbox.go
  projects.go
  user.go
  (~17 files, one per Fiken tag)

cli/                              ff/v4 command tree
  root.go                         root cmd, global flags, context wiring
  auth.go                         fiken auth login|status|logout|list
  mcp.go                          fiken mcp subcommand
  progress.go                     stderr progress writer for paged ops
  parity_test.go                  CLI --json bytes == MCP tool result bytes
  companies.go
  invoices.go
  ...                             one per ops domain

mcp/                              MCP server
  server.go                       go-sdk server, tool registration
  readonly.go                     Mode enum + filter (consults mutating.gen.go)
  transport.go                    stdio + streamable HTTP wiring
  progress.go                     MCP progress notifications during paging
  attachments.go                  opt-in path-only multipart support

mockfiken/                        ogen-Handler-based mock for tests
  server.go                       httptest.Server wrapping fiken.NewServer(Handler)
                                  + per-op override registry

cmd/
  fiken/
    main.go                       ~30 line entrypoint
  fiken-spec-update/
    main.go                       spec fetcher + difftastic diff
  fiken-i18n-scaffold/
    main.go                       seeds i18n/locales/en.template.toml from OAS
  fiken-mutating-gen/
    main.go                       emits ops/mutating.gen.go from OAS HTTP method
  fiken-spec-lint/
    main.go                       OAS unit-format checks (date/datetime
                                  declarations on *At/*Date fields)

flake.nix                         Nix devShell + package
flake.lock
.envrc                            direnv: use flake
.golangci.yml                     lint config
.pre-commit-config.yaml           prek hooks
.github/
  workflows/
    ci.yml                        test + lint + codegen-clean + hooks
    spec-drift.yml                daily upstream-spec drift detector
```

No `pkg/`, no `internal/`, per repo convention.

## Key design decisions

### Spec and generator

- **OpenAPI generator: `github.com/ogen-go/ogen`** — strong typing, `OptT`
  generics for optional/nullable, no `interface{}`, no reflection,
  supports OAS 3.0.2. Pin a specific ogen version. Generated package is
  `fiken` (matches the import path needed for drop-in into sfiber later).
- **Spec source: official Fiken docs**, canonical URL
  `https://api.fiken.no/api/v2/docs/swagger.yaml` (OpenAPI 3.0.2). Vendored
  at `api/fiken-openapi.yaml`. **No third-party mirror.**
  `cmd/fiken-spec-update` always fetches from this URL. The original Fiken
  response is byte-preserved before any normalisation so we can diff
  cleanly.
- **Spec characteristics observed in current snapshot** (informs the
  decisions below):
  - `summary` field is **empty** for most operations; `description` carries
    the prose.
  - Parameter `description` quality is **mixed** ("From the request header"
    next to good ones).
  - No `securitySchemes` block — auth is plain `Authorization: Bearer …`
    described in prose.
  - 6 endpoints declare `multipart/form-data` (attachments).
  - Zero `callbacks:` / webhook declarations.

### Auth

- **Personal API Tokens only in phase 1.** Fiken's spec only declares
  `authorizationCode` OAuth2 flow (no device grant); a browser-redirect
  login is worth it but parked for phase 2.
- **Resolution order**, highest first: `--token` flag → `FIKEN_TOKEN` env
  → keyring (`zalando/go-keyring`, service `fiken-go`, user = profile name)
  → config file.
- **Storage shape v1**: bare token string in keyring (or chmod-0600 fallback
  file). Phase 2 OAuth migrates the value to a structured `Credential`
  JSON; migration is the phase-2 designer's problem, not paid up front.
- **Auth UX** (`fiken auth login|status|logout|list`):
  - `login`: prints "Create personal API token: <Fiken URL>" (exact URL
    **TBD verify against Fiken UI** — likely
    `https://fiken.no/foretak/<co>/api-tokens` but confirm before
    landing; bilingual link if `--lang=nb`), reads token via `x/term`
    no-echo, trims whitespace, verifies against `/user/me` before
    saving, then writes to keyring (or 0600 file if keyring unavailable
    — prompt confirms fallback).
  - `status`: prints user + companies + credential source.
  - `logout`: removes the credential entry.
  - `list`: enumerates configured profiles.
  - Overwrite of an existing profile credential always prompts confirm.

### Config

- **`koanf v2`** with providers in this load order (later wins): file
  (`~/.config/fiken/config.toml` default; `--config` overrides) → env
  (`FIKEN_*`, e.g. `FIKEN_PROFILE`, `FIKEN_TOKEN`, `FIKEN_COMPANY`) →
  ff/v4 flags via the `posflag` adapter. `--profile` selects a named
  profile block; default profile named in `[default] profile = "..."`.
- `Profile` carries optional `company` slug — this also feeds the MCP
  server-default (see Company scoping below).

### CLI

- **Framework: `github.com/peterbourgon/ff/v4`** — one `*ff.Command` per
  subcommand, child `FlagSet`s use `SetParent()` so global flags
  (`--profile`, `--token`, `--company`, `--lang`, `--json`, `--config`,
  `--log-json`, `-v / -vv`) work everywhere.
- ff/v4 doesn't ship YAML/JSON config parsers, so koanf owns config
  loading; ff only parses args.

### MCP

- **`github.com/modelcontextprotocol/go-sdk/mcp`**. Transport flag:
  `--transport=stdio|http`, default `stdio`. HTTP mode listens on
  `--listen=:8765` (default) using the SDK's streamable HTTP transport.
- Mode flag `--mode=read-only|read-write`, default `read-only`. The server
  only registers tools whose op-name is `false` in
  `ops/mutating.gen.go`'s table.
- Mutating bit is **never hand-marked**. `cmd/fiken-mutating-gen` reads
  `api/fiken-openapi.yaml`, classifies each operation by HTTP method
  (`GET`/`HEAD` = non-mutating; everything else = mutating), and emits
  `ops/mutating.gen.go`. CI's `codegen-clean` job fails if anyone edits
  by hand.

### Shared op identifiers and the single `Out` type

- **Shared op identifiers** in `ops/names.go`:
  ```go
  const (
      OpCompaniesList = "companies_list"   // CLI: `fiken companies list`
                                           // MCP tool name: same string
      OpInvoicesGet   = "invoices_get"
      // ... one per operation
  )
  ```
  Both `cli/companies.go` and `mcp/server.go` reference the same const, so
  a rename touches one file.
- **Single `Out` type, two frontends.** Each `ops.Op*Out` struct is the
  _one_ return type for both `--json` CLI output and the MCP tool result.
  Both frontends wrap it in the same `Result[T]` envelope (below).
  - JSON tags are authoritative and stable; `omitempty` only where the
    absence is meaningful (don't blank out empty arrays — agents handle
    them fine).
  - `output.JSON` renderer just calls `json.NewEncoder(w).Encode(result)`
    on the same value the MCP handler returns; no per-frontend reshaping.
  - Money stays as `int64` øre in the `Out` types (no float drift);
    `output.Table` is the only place that formats to NOK.
  - Tax rates are **basis points** `int` (25% = 2500, 11.11% = 1111).
    Exact, no float drift, table renderer formats as `%`.
  - Date-only fields use `ops.Date` (`type Date string` with `"2006-01-02"`
    invariant + custom MarshalJSON).
  - Datetime fields use stdlib `time.Time` (Fiken's spec declares
    `format: date-time`; verified by codegen-time lint).
  - The `Out` types live in `ops/` because both frontends consume them
    and `output/` is presentation-layer.

### Result envelope (universal)

All ops return `Result[T]` rather than a bare `Out`. Both frontends emit
the envelope as their canonical output, including on the error path.

```go
// ops/result.go
type Result[T any] struct {
    Ok    *T     `json:"ok"`     // never omitempty: presence signals discriminator
    Error *Error `json:"error"`  // never omitempty
}

type Error struct {
    Code       string         `json:"code"`        // stable, machine
    Message    string         `json:"message"`     // stable EN (data, not UX)
    HTTPStatus int            `json:"http_status,omitempty"`
    Op         string         `json:"op,omitempty"`
    Details    map[string]any `json:"details,omitempty"` // field errors, Fiken raw
    RequestID  string         `json:"request_id,omitempty"`
}
```

- **Stable `Code` set**: `auth_missing`, `auth_invalid`, `auth_forbidden`,
  `not_found`, `validation`, `conflict`, `rate_limited`, `server_error`,
  `network`, `cancelled`, `internal`, `read_only_violation`.
- **`Message` is data, not UX.** It stays in stable English regardless of
  `--lang` so JSON bytes are identical across locales. Table renderer
  looks up `Code` in the i18n bundle (`error.<code>`) to produce a
  localized human string for terminal display.
- **CLI exit codes**: `0` ok; `1` user error (`validation`/`not_found`/
  `conflict`/`auth_*`); `2` retriable (`rate_limited`/`server_error`/
  `network`); `130` `cancelled`. `read_only_violation` and `internal`
  map to `1`.
- **CLI `--json` on error**: writes the full envelope to **stdout** plus
  exit code as above. (Pipelines doing `fiken … --json | jq` need to
  branch on `.error != null`; this is the trade for strict parity.)
- **CLI table on error**: writes a localized human message to **stderr**;
  same exit code.
- **MCP**: handler returns `Result[Out]`. SDK serializes into
  `CallToolResult.StructuredContent`; we additionally set `IsError=true`
  when `Error != nil`. The wire JSON in `StructuredContent` is byte-equal
  to the CLI `--json` envelope. This is the parity invariant
  (`cli/parity_test.go` enforces).

### Pagination

Different defaults per frontend; envelope is shared.

```go
// ops/paging.go
type ListMeta struct {
    Truncated bool `json:"truncated"`
    NextPage  int  `json:"next_page,omitempty"`
    Returned  int  `json:"returned"`
    Cancelled bool `json:"cancelled,omitempty"`
}
type ListOut[T any] struct {
    Items []T      `json:"items"`
    Meta  ListMeta `json:"meta"`
}
```

| Frontend | Default `max_results` | Hard ceiling         | Beyond ceiling                                                           |
| -------- | --------------------- | -------------------- | ------------------------------------------------------------------------ |
| CLI      | unlimited (humans)    | none                 | iterates all pages                                                       |
| MCP      | 25                    | 100 (one Fiken page) | `Meta.Truncated=true`, `Meta.NextPage=N+1`; agent must explicit-continue |

Parity test still holds: identical inputs → identical envelope. The
defaults differ only at the frontend boundary where the caller doesn't
specify `max_results`.

### Company scoping (Hybrid)

Fiken paths look like `/companies/{companySlug}/...`. Most ops need a
company slug; some (`/companies`, `/user/me`) don't.

- **CLI**: global `--company` flag (also `FIKEN_COMPANY`, profile field).
- **MCP**: hybrid model.
  - Server has an optional default (`fiken mcp --company=acme`, or
    inherited from the profile).
  - Each company-scoped tool exposes `company` as an **optional**
    parameter in its InputSchema.
  - Resolution: tool-arg → server default → error (`validation`,
    "company required").
  - Tools that don't operate on a company (`companies_list`, `user_me`,
    etc.) omit the param entirely.

This keeps single-tenant clean (no `company` arg noise) while letting
multi-company agents override per call.

### Multipart attachments

6 endpoints declare `multipart/form-data`. ogen generates Go structs with
`ht.MultipartFile` fields (`io.Reader` + filename); not yet exercised in
this repo, but documented as supported.

- **CLI**: `fiken invoices attach <invoice-id> --file path/to/x.pdf
[--name "custom.pdf"]`. `--file -` reads stdin.
- **MCP**: attach ops are **not registered by default**. Server flag
  `--enable-attachments` registers them with a `file_path` parameter
  (local path on the MCP server host). Phase 1 has no base64 input
  variant. Document the trade-off in the tool description: agent must
  have host FS access for the attach to work.

### Rate limit, concurrency, progress

- Fiken's only stated rate rule: **one concurrent request per user**. We
  enforce with a buffered token of size 1 on the `ops.Client`.
- **No preemptive ticker.** Drop the 250 ms heartbeat from the earlier
  draft — Fiken does not currently throttle by RPS, and the cost of a
  fixed delay is real (every command pays it).
- **Reactive backoff** on 429: honor `Retry-After` if present, else
  exponential 250 ms → 4 s, capped, with full jitter. Lives in
  `ops/ratelimit.go` as an `http.RoundTripper` chained under the
  concurrency token.
- **No client-side prefetch**. Fiken's concurrency rule forbids.
- **Progress signals**:
  - CLI `--verbose` (`-v` or higher): stderr `page 3/12, 250 items so
far` between page fetches.
  - MCP: emit `progress` notifications via the SDK during paged ops so
    agents can stream rather than wait silently.

### Cancellation

- **Single-request ops** (GET/POST one-shot): cancel mid-request →
  `Result.Error` with `Code = cancelled`. CLI exits 130.
- **List/paged ops**: cancel between page fetches → `Result.Ok` with the
  pages already collected, `Meta.Cancelled = true`, `Meta.Truncated =
true`. CLI exits 130 but still emits the partial envelope. Useful data
  is preserved.

### i18n

- **`github.com/nicksnyder/go-i18n/v2/i18n`** with TOML catalogs in
  `i18n/locales/`. Languages: `en` (default), `nb` (bokmål — Fiken's own
  UI language).
- Accepted `--lang` values: `en`, `nb`, plus `no` as an alias for `nb`.
  No nynorsk.
- Selection: `--lang` flag → `FIKEN_LANG` → `LANG`/`LC_MESSAGES` (with
  locale prefix stripping: `nb_NO.UTF-8` → `nb`, `nn_*` → `en`) → `en`.
- Used for help/error/message strings in CLI and for tool descriptions
  in MCP.
- **Data payloads are never translated.** Field names, enum values, and
  the stable English `Error.Message` stay byte-stable across `--lang`.
- **Help-text source of truth**: all `ops.<op>.{summary,when-to-use,
returns,example}` keys and `ops.<op>.flags.<flag>` keys are
  hand-authored in both `en.toml` and `nb.toml`. The scaffold tool
  `cmd/fiken-i18n-scaffold` reads the OAS spec and emits a starter
  `en.template.toml` with op-name keys and `description`-derived prose;
  humans copy / rewrite into `en.toml`. The template file is **not loaded
  at runtime** — it's a one-shot editor convenience.
- **`nb` quality gate**: CI fails if any key present in `en.toml` is
  missing from `nb.toml`. Hand-written by the maintainer; reviewed on PR.
  No machine translation committed without a human pass.

### Help text and flag descriptions are a first-class deliverable

The same audience reads `fiken … --help` and a tool's MCP description —
humans skimming a terminal and agents picking which tool to call.

- **Every op** gets a four-line spec attached via the `Registry`:
  1. **Summary** — one imperative sentence, ≤80 chars.
  2. **When to use** — one sentence on when to pick this over neighbours.
  3. **Returns** — one sentence on response shape and units.
  4. **Example** — a concrete shell or MCP-JSON invocation.
- **Every flag** description includes (a) what it controls, (b) the
  type/unit, (c) the default, (d) required-ness, (e) one tiny example.
  No bare `string`/`int` descriptions.
- **CLI `--help` and MCP tool description share the same i18n keys**, so
  one source feeds both surfaces. A unit test walks `ops.Registry` and
  fails if any op is missing any of the four keys in either locale, or
  if any flag lacks description + type + example.
- **MCP-specific affordances**: where the SDK supports it
  (`mcp.Tool.InputSchema` annotations, parameter examples), populate from
  the same registry.

### Output

- Global `--json` flag flips `output.Renderer` between `output.Table`
  (`text/tabwriter`) and `output.JSON` (`encoding/json.NewEncoder`).
- The renderer always receives a `Result[T]` — never a bare `Out`. JSON
  mode emits the envelope; table mode formats `Ok.*` or, on error, prints
  a localized `error.<code>` message to stderr.
- For paged ops `Ok` is `ListOut[T]`; the table renderer iterates
  `Ok.Items` and calls `TableHeader()` / `TableRow() []string` on each
  `T`. For single ops `Ok` is the `Out` struct itself, which implements
  the same two methods. The `--json` path is agnostic and emits the
  envelope as-is.

### Read-only enforcement

- **Single layer**: `ops/mutating.gen.go` (generated from OAS HTTP method)
  is the source of truth. MCP read-only mode consults it to decide which
  tools to register. CI's `codegen-clean` gate prevents hand-edit drift.
- The earlier draft's RoundTripper belt is dropped — the registry is
  sufficient given it's machine-generated.

### Units invariants (test-enforced)

`ops/units_test.go` reflect-walks every exported `Out*` struct. Field
name patterns map to required Go types:

| Pattern                                                                              | Required type        |
| ------------------------------------------------------------------------------------ | -------------------- |
| `*Amount`, `*Price`, `*Total`, `*Sum`, `*Net`, `*Gross`, `*Balance`, `*Paid`, `*Due` | `int64`              |
| `*Rate`, `*Percent`                                                                  | `int` (basis points) |
| `*Date` (no time component)                                                          | `ops.Date`           |
| `*At`, `*Time`, `*DateTime`                                                          | `time.Time`          |

Failures list offending field paths so regressions are obvious. Adding a
new pattern is one map entry.

A companion **OAS lint** (`cmd/fiken-spec-lint`) at codegen time asserts
every spec field whose name matches `*At|*Date` declares `format:
date-time` or `format: date`. Catches naive-datetime strings that would
silently parse as UTC.

### Logging

- **`log/slog` only**, always written to **stderr** (stdout is data /
  MCP protocol).
- **Levels**: default WARN. `-v` raises to INFO. `-vv` raises to DEBUG
  (request bodies, headers). MCP server inherits the same flags.
- **Format**: text default; `--log-json` flips to JSON for parseable
  pipelines.
- **Per-request line at INFO**: `op=companies_list method=GET path=/companies
status=200 dur=234ms req_id=abc123`.
- **DEBUG adds**: request body (truncated >2KB), full headers, response
  body summary.
- **Request ID**: capture Fiken's `X-Request-Id` if present; else
  generate a ULID client-side. Surfaces in `ops.Error.RequestID` on
  errors.
- **Secret redaction**: log middleware always redacts `Authorization`
  headers, even at DEBUG.

## Implementation steps

Order matters — each step's output is consumed by the next. **Steps 1
and 2 come first** so every commit made during the rest of the build is
gated by the same hooks (formatters, linters, `go test -short`) that
will gate CI.

### 1. `flake.nix` (headscale-style) — devShell first

- Inputs: `nixpkgs` (nixpkgs-unstable), `flake-utils` (numtide).
- Outputs:
  - `packages.${system}.fiken` via `pkgs.buildGo126Module` with
    `vendorHash`; falls back to `buildGoModule` overridden to `go_1_26`.
  - `packages.${system}.default = fiken`.
  - `devShells.${system}.default` = `mkShell` with `go_1_26`, `gopls`,
    `gofumpt`, `golangci-lint`, `gotestsum`, `gotests`, `difftastic`,
    `prek`, `ogen` (built from source via `buildGo126Module` if not in
    nixpkgs), `nodePackages.prettier`, `nixfmt-rfc-style`, `git`.
  - `formatter.${system}` = `pkgs.nixfmt-rfc-style`.
  - `overlays.default` exposes `fiken` package.
- Add `.envrc` with `use flake` for direnv users.
- **Verify**: `nix develop -c bash -lc 'go version && prek --version &&
golangci-lint --version'` succeeds.

### 2. Pre-commit (`prek`) hooks — wire before anything else

**Hooks installed and passing before the first non-trivial commit.** The
first step-2 commit writes `.pre-commit-config.yaml`, runs `prek install`,
runs `prek run --all-files`, and only then commits. From that commit
forward every `/commit` is gated by the real hook set. No "we'll wire
it up later", no "it's failing but I'll fix it after this batch".

- Use `prek` (`github.com/j178/prek`, Rust drop-in for pre-commit) as the
  runner. Config file `.pre-commit-config.yaml`.
- Hooks (adapted from headscale's `.pre-commit-config.yaml`):
  - `pre-commit/pre-commit-hooks` v6.0.0: `check-added-large-files`
    (`--maxkb=1024`), `check-case-conflict`, `check-merge-conflict`,
    `check-symlinks`, `check-json`, `check-toml`, `check-xml`,
    `check-yaml`, `detect-private-key`, `fix-byte-order-marker`,
    `end-of-file-fixer`, `trailing-whitespace`, `mixed-line-ending`.
  - Local hooks (all via `nix develop -c …`):
    - `nixfmt` — `nixfmt-rfc-style` on `*.nix`.
    - `prettier` — JSON/YAML/Markdown (excluding generated + `api/`).
    - `gofumpt` — Go files outside `fiken/` and `**/*_gen.go` /
      `**/*.gen.go`.
    - `goimports` — same scope.
    - `golangci-lint` — `--timeout=5m --fix`, repo-wide, single run per
      commit.
    - `vendor-hash` — regenerate flake hash file when `go.{mod,sum}`
      change; fails if stale.
    - `codegen-clean` — runs `go generate ./...` and fails if `git diff`
      is non-empty. Catches:
      - hand-edits to generated `fiken/`,
      - drift in `ops/mutating.gen.go`,
      - drift in any other `.gen.go`.
    - `i18n-keys` — fails if any key in `en.toml` is missing from
      `nb.toml`.
    - `oas-units` — runs `fiken-spec-lint`; fails on missing
      `format: date|date-time` declarations.
    - `go-test` — `go test -race -count=1 -short ./...`,
      `pass_filenames: false`, `stages: [pre-commit]`. `--no-verify`
      escape exists, but default is on.
- `pre-push` stage adds the non-`-short` integration suite plus
  `golangci-lint` without `--fix`.
- Install once: `nix develop -c prek install --hook-type pre-commit
--hook-type pre-push`.
- The same `.pre-commit-config.yaml` is what `ci.yml`'s `hooks` job runs
  via `prek run --all-files`, so CI and local enforcement are
  bit-identical.

### 3. Codegen plumbing — ogen plus three scaffold tools

**ogen is the only generator for code.** The scaffold tools below are
build-time helpers that emit either non-loaded fixtures (i18n template)
or trivial generated Go (`mutating.gen.go`) — they are not parallel code
generators.

- Add `api/fiken-openapi.yaml` via `curl -fsSL
https://api.fiken.no/api/v2/docs/swagger.yaml -o
api/fiken-openapi.yaml`. Record source URL + SHA-256 in
  `api/SOURCE.txt`.
- Add `fiken/ogen.yml` and `fiken/doc.go`:
  ```go
  //go:generate go run github.com/ogen-go/ogen/cmd/ogen --target . --package fiken --clean ../api/fiken-openapi.yaml
  ```
  Default `--generate-types,client,server,spec` so the same package
  exposes both `Client` (used by `ops/`) and `Handler` interface (used by
  `mockfiken/`).
- Add `cmd/fiken-spec-update/main.go`: HTTP-GET the canonical URL to a
  temp file, run `difft <old> <new>` if available (else
  `git diff --no-index --color=always`), prompt confirm, then move into
  place and rewrite `api/SOURCE.txt` with the new SHA-256. Stdlib +
  `os/exec` only.
- Add `cmd/fiken-mutating-gen/main.go`: reads the YAML, classifies each
  operation by HTTP method, emits `ops/mutating.gen.go`. Header comment
  states `DO NOT EDIT`; `codegen-clean` hook enforces.
- Add `cmd/fiken-i18n-scaffold/main.go`: reads the YAML, emits
  `i18n/locales/en.template.toml` with op-name keys and
  `description`-derived starter prose. Run on demand; output is editor
  scaffolding, not loaded at runtime.
- Add `cmd/fiken-spec-lint/main.go`: reads the YAML, walks all
  `properties`, fails if any property whose name matches `*Date|*At|*Time|
*DateTime` lacks a `format: date` or `format: date-time` declaration.
  Wired into the `oas-units` pre-commit hook.
- Run `go generate ./fiken && go generate ./ops` once; commit the
  result.
- CI `codegen-clean` job: `go generate ./...` then
  `git diff --exit-code`.

### 4. Foundations: `auth/`, `config/`, `output/`, `i18n/`

- `config/config.go`: `type Profile struct { Token, Company, Lang string }`
  and `type Config struct { Default string; Profiles map[string]Profile }`.
  `Load(flags)` composes koanf with `file.Provider` (TOML via
  `koanf/parsers/toml/v2`), `env.Provider("FIKEN_", ".")`,
  `posflag.Provider`. `Resolve(profile string)` returns the merged
  effective profile.
- `auth/auth.go`: `type Source interface { Token(ctx) (string, error) }`.
  Concrete: `FlagSource`, `EnvSource`, `FileSource(profile)`,
  `KeyringSource(profile)`, `ChainSource{...}` returning the first
  non-empty.
- `auth/credential.go`: bare-token storage helpers. Phase 2 will widen
  the value shape; for now keyring/file holds a plain string.
- `auth/keyring.go`: thin wrap of `github.com/zalando/go-keyring` with
  service constant `"fiken-go"` and graceful fallback (return
  `ErrNotFound`, not panic) when no keyring available (CI / containers).
- `output/output.go`: `type Renderer interface { Render(ctx, v any) error }`,
  factories `JSON(w io.Writer)` and `Table(w io.Writer)`. Both expect a
  `Result[T]`. Table consults a `TableRow() []string` /
  `TableHeader() []string` method on `Ok.*` when present, falling back to
  reflect.
- `i18n/i18n.go`: package-level `Bundle` initialized at startup;
  `T(lang, id, data)` helper. Embed locale TOMLs with `//go:embed
locales/*.toml`.

Tests at this layer are unit tests with `t.TempDir()`-backed config
files and an in-memory keyring stub.

### 5. `ops/` — domain operations

Single Go package wrapping the generated client.

```go
type Client struct {
    gen   *fiken.Client
    sem   chan struct{}              // size 1 — Fiken concurrency rule
    rt    http.RoundTripper          // 429-aware backoff wraps base
}
func New(ctx context.Context, cfg Profile, src auth.Source) (*Client, error) { ... }
```

For each Fiken tag (~17 files):

- A typed `In*` request struct (JSON-tagged; MCP unmarshals tool args
  into this).
- A typed `Out*` response struct (or `ListOut[T]` for paged ops).
- A method on `Client` returning `(Result[Out], error)`. The `error`
  return is reserved for "the call itself didn't make sense" (programmer
  errors); user-visible failures land in `Result.Error`. CLI / MCP
  always render the `Result`.
- Registration in `ops/names.go`'s `Registry` map (summary key, flags
  key, etc.). The Mutating bit is read from `ops/mutating.gen.go`, not
  set here.

Pagination: helper `paginate(ctx, fn func(page int) ([]T, bool, error))
ListOut[T]` honoring the per-frontend caps via an `In.MaxResults` + a
caller-supplied `Ceiling` parameter. CLI uses ceiling 0 (unlimited); MCP
uses ceiling 100.

`ops/ratelimit.go`:

- `concurrencyRT` — acquires the size-1 semaphore around each request.
- `backoffRT` — on 429 honors `Retry-After`, else exponential
  250 ms → 4 s with jitter, capped at N retries.
- Composed via `transport: backoffRT(concurrencyRT(http.DefaultTransport))`.

### 6. `cli/` — ff/v4 command tree

- `cli/root.go`: build the root `*ff.Command` named `fiken`. Global flags
  on the parent FlagSet: `--config`, `--profile`, `--token`, `--company`,
  `--lang`, `--json`, `--no-color`, `--log-json`, `-v / -vv`. Construct
  `config.Config`, resolve profile, build `ops.Client` once, stash in
  context.
- One file per ops domain: each builds a parent subcommand with
  `list`/`get`/`create`/... children. Each child calls the matching
  `ops.Client.*` method and writes the result to `output.Renderer` from
  context. The struct passed is the same `Result[Out]` MCP will return.
- `cli/auth.go`: `fiken auth login|status|logout|list`.
- `cli/mcp.go`: `fiken mcp [--mode=read-only|read-write]
[--transport=stdio|http] [--listen=:8765] [--enable-attachments]`.
- `cli/progress.go`: stderr progress writer used by paged ops when
  verbose.
- `cli/parity_test.go`: for every op, run CLI with `--json` and call the
  matching MCP tool with equivalent input; `assert.JSONEq` (or byte-equal
  after canonicalisation) the two outputs — for both success and error
  paths.
- All user-facing strings via `i18n.T(lang, key, …)`. Help text via
  `ff.Command.LongHelp` populated from i18n.

### 7. `mcp/` — MCP server

- `mcp/server.go`: `func New(client *ops.Client, opts Options) *mcp.Server`.
  Walks `ops.Registry`, skips ops where `mutating.gen.go` says `true` in
  read-only mode, calls `mcp.AddTool(srv, &mcp.Tool{Name: opName,
Description: i18n.T(...), InputSchema: schema}, handler)`. Schema for
  company-scoped ops includes optional `company`; non-scoped ops omit.
- Handler signature: `func(ctx, *mcp.CallToolRequest, In) (*mcp.CallToolResult, Result[Out], error)`.
  Delegates to the matching `ops.Client.*` method. If `Result.Error !=
nil`, sets `CallToolResult.IsError = true`. Returns the envelope in
  `StructuredContent`.
- `mcp/readonly.go`: `Mode` enum + filter that consults
  `ops.IsMutating(opName)` (sourced from `mutating.gen.go`).
- `mcp/transport.go`: wires stdio (default) and streamable HTTP based on
  `Options.Transport`.
- `mcp/progress.go`: helper to emit progress notifications during paged
  ops.
- `mcp/attachments.go`: when `Options.EnableAttachments` is set,
  registers the 6 multipart ops with a `file_path` parameter; otherwise
  they don't appear in `tools/list`.

### 8. `mockfiken/` — ogen-generated mock server

- `mockfiken/server.go`: `type Server struct { ... }` implementing the
  generated `fiken.Handler` interface. Construct via `mockfiken.New(t)`:
  builds `Server`, wraps with `fiken.NewServer(srv)`, wraps that in
  `httptest.NewServer`, exposes `.URL()`.
- **Default response strategy**: each `Handler` method returns the
  zero-value of the generated response type for the lowest 2xx status.
  Zero values are deterministic and satisfy the type system.
- **Override registry** for tests that need concrete data:
  ```go
  mock := mockfiken.New(t)
  mock.Set(ops.OpInvoicesList, []fiken.Invoice{{InvoiceNumber: fiken.NewOptInt64(42)}})
  ```
  Keyed by `ops.Op*` name. The generated `Handler` method looks up the
  registry first, falls back to the zero value.
- **Error injection**: `mock.SetError(ops.OpInvoicesGet,
&fiken.ErrorResponse{...}, 404)` exercises typed error paths.
- **Auth wiring**: wrapping `http.Handler` checks for non-empty
  `Authorization: Bearer …` before delegating; returns 401 otherwise.
- **One reusable package** for `ops` integration tests, `cli` golden
  tests, `mcp` tests, `cli/parity_test.go`. No e2e suite hits the real
  API.

### 9. `cmd/fiken/main.go`

About 30 lines: build root command via `cli.NewRoot()`, call
`root.ParseAndRun(ctx, os.Args[1:], ff.WithEnvVarPrefix("FIKEN"))`,
print errors via `i18n` and exit non-zero.

### 10. GitHub Actions (Nix-based, nix-community)

All jobs run on `ubuntu-latest` and use **nix-community actions** —
`cachix/install-nix-action` for the Nix install, `cachix/cachix-action`
for build cache. **No `DeterminateSystems/*` actions.** Every job's
actual command runs inside `nix develop -c …` so versions match the
flake exactly.

Common setup snippet:

```yaml
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
```

Workflows in `.github/workflows/`:

- **`ci.yml`** — runs on `push` and `pull_request`:
  - `test` — `nix develop -c go test -race -count=1 ./...`
  - `lint` — `nix develop -c golangci-lint run --timeout=5m`
  - `hooks` — `nix develop -c prek run --all-files` (includes
    `i18n-keys`, `oas-units`, `codegen-clean`).
  - `codegen-clean` — explicit job for safety:
    `nix develop -c go generate ./...` then `git diff --exit-code`.
  - `build` — `nix build .#fiken`.
- **`spec-drift.yml`** — scheduled drift detector. Runs on `schedule:
cron: "17 6 * * *"` and `workflow_dispatch`.
  1. Common setup.
  2. `curl -fsSL https://api.fiken.no/api/v2/docs/swagger.yaml -o
/tmp/upstream.yaml`.
  3. `nix develop -c difft --exit-code api/fiken-openapi.yaml
/tmp/upstream.yaml`.
  4. On non-zero: a Go helper prints a structured summary of changed
     paths and the new SHA-256.
  5. If a drift issue exists (label `spec-drift`), comment with the diff
     summary; otherwise open one. Uses `actions/github-script@v7` with
     `GITHUB_TOKEN`.
  6. Optionally open a PR by running `fiken-spec-update --apply` and
     `go generate ./...`, then `peter-evans/create-pull-request@v7`.
     Gated behind a `secrets.SPEC_BOT_PAT`.

### 11. Testing

- **Unit** (`*_test.go` per package): table-driven, no network.
- **Integration** (`ops/ops_integration_test.go`): start
  `mockfiken.New(t)`, point the production client at it via
  `fiken.WithServerURL(mock.URL())`. Cover one happy + one error per op.
- **Units invariant** (`ops/units_test.go`): reflect-walk every `Out*`,
  assert field types match name-pattern table. Listed offending fields.
- **CLI tests** (`cli/*_test.go`): build root command, point at
  `mockfiken`, call `Cmd.ParseAndRun`, assert stdout (golden files for
  both `--json` and table modes; one `nb` golden to prove i18n works).
- **MCP tests** (`mcp/server_test.go`): use SDK's in-process transport
  (`mcp.InMemoryTransport`), call each registered tool against the same
  `mockfiken` server, assert filtering in read-only mode plus
  presence/absence of attachment ops based on `--enable-attachments`.
- **Parity test** (`cli/parity_test.go`): for every op, `assert.JSONEq`
  CLI `--json` bytes vs MCP `StructuredContent` for the same input on
  success and error paths.
- **i18n completeness test**: walks `ops.Registry`, fails if any op is
  missing `summary` / `when-to-use` / `returns` / `example` in either
  locale, or if any flag lacks description + type + example.
- **Lint**: `.golangci.yml` enabling `errcheck`, `govet`, `staticcheck`,
  `gofumpt`, `gosec`, `misspell`, `revive`, `unused`. CI runs
  `golangci-lint run` and `go vet ./...`.
- **No real-API tests.** `mockfiken` covers wire-level behaviour without
  needing a token; upstream-spec drift is caught by `cmd/fiken-spec-
update` + the codegen-clean CI gate + the scheduled drift workflow.

## Critical files to create / modify

- `api/fiken-openapi.yaml`, `api/SOURCE.txt` — vendored spec + provenance
- `fiken/ogen.yml`, `fiken/doc.go` — codegen entrypoint (rest of `fiken/`
  is generated)
- `auth/auth.go`, `auth/credential.go`, `auth/keyring.go`
- `config/config.go`
- `output/output.go`, `output/table.go`, `output/json.go`
- `i18n/i18n.go`, `i18n/locales/en.toml`, `i18n/locales/nb.toml`
- `ops/names.go`, `ops/result.go`, `ops/errors.go`, `ops/paging.go`,
  `ops/date.go`, `ops/ratelimit.go`, `ops/units_test.go`,
  `ops/mutating.gen.go` (generated)
- One ops file per Fiken tag (~17 files)
- `cli/root.go` + one file per ops domain + `cli/mcp.go` +
  `cli/auth.go` + `cli/progress.go` + `cli/parity_test.go`
- `mcp/server.go`, `mcp/readonly.go`, `mcp/transport.go`,
  `mcp/progress.go`, `mcp/attachments.go`
- `mockfiken/server.go`
- `cmd/fiken/main.go`
- `cmd/fiken-spec-update/main.go`
- `cmd/fiken-mutating-gen/main.go`
- `cmd/fiken-i18n-scaffold/main.go`
- `cmd/fiken-spec-lint/main.go`
- `flake.nix`, `.envrc`, `.golangci.yml`
- `.pre-commit-config.yaml`
- `.github/workflows/ci.yml`, `.github/workflows/spec-drift.yml`

## Reused / external code

- `github.com/ogen-go/ogen` — generates `fiken/` (client **and** server /
  `Handler` interface for `mockfiken/`)
- `github.com/peterbourgon/ff/v4` — CLI parsing
- `github.com/knadh/koanf/v2` + `parsers/toml/v2` +
  `providers/{file,env/v2,posflag}` — config
- `github.com/zalando/go-keyring` — credential storage
- `github.com/nicksnyder/go-i18n/v2` — bilingual messages
- `github.com/modelcontextprotocol/go-sdk/mcp` — MCP server
- `golang.org/x/term` — no-echo token paste
- stdlib `text/tabwriter`, `encoding/json`, `net/http/httptest`, `embed`,
  `log/slog`
- `difft` (system binary, optional) — spec diffs
- Inspiration only (no code copy):
  - `github.com/jakoblind/fiken-cli` for command shape and table layout
  - `github.com/heim/fiken-read-only-mcp` for tool descriptions and
    integration-test inputs

## Verification

End-to-end checks before declaring done:

1. `go generate ./...` produces zero diff — `git status` clean.
2. `go test ./...` green (uses `mockfiken` throughout — no real Fiken
   token needed in the test suite).
3. `golangci-lint run` clean.
4. `prek run --all-files` clean.
5. **Parity check passes**: `cli/parity_test.go` shows CLI `--json` bytes
   equal MCP tool result bytes for every registered op, on both success
   and error paths.
6. **Help completeness check passes**: registry walk asserts every op
   has `summary` / `when-to-use` / `returns` / `example` in both `en`
   and `nb`, and every flag has description + type + example.
7. **Units invariant test passes**: every `Out*` field matches the
   name-pattern → type table.
8. **OAS unit-format lint passes**: every `*At|*Date` spec field declares
   `format: date` or `format: date-time`.
9. Manual smoke (humans only, real token):
   - `fiken auth login --profile test` then `fiken auth status` returns
     the user.
   - `fiken companies list` shows a table; `fiken companies list --json`
     is valid JSON with the `Result[T]` envelope shape.
   - `fiken --lang=nb companies list --help` shows Norwegian (bokmål)
     help; `--lang=no` returns the same output (alias). The JSON payload
     is **identical** to the `en` run.
   - `fiken --profile test invoices list --company my-co` paginates
     through results; `-v` shows progress on stderr.
   - Ctrl-C during a long paged list returns a partial envelope with
     `meta.cancelled=true, meta.truncated=true` and exit 130.
10. MCP smoke:
    - `fiken mcp` (default `--mode=read-only --transport=stdio`) responds
      to `tools/list` over stdio with only non-mutating tools; attachment
      tools absent.
    - `tools/call companies_list` returns `StructuredContent` byte-equal
      to `fiken companies list --json`.
    - `fiken mcp --transport=http --listen=:8765` serves the same surface
      over streamable HTTP.
    - `fiken mcp --mode=read-write --enable-attachments` exposes mutating
      and attachment tools; calling one against a sandbox token actually
      creates the resource.
    - Drop into Claude Desktop / `claude mcp add fiken -- fiken mcp` and
      verify end-to-end tool call.
11. `nix develop` drops into shell with `go`, `ogen`, `golangci-lint`,
    `difft`. `nix build` produces `./result/bin/fiken` that runs
    `--help` successfully.
12. `go run ./cmd/fiken-spec-update` fetches the canonical YAML, prints a
    difftastic diff, updates `api/fiken-openapi.yaml` and
    `api/SOURCE.txt` on confirm.

## Appendix: resolved gaps (2026-05-15 brainstorm)

| #   | Gap              | Resolution                                                                                                                        |
| --- | ---------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| 1   | Error model      | `Result[T]` + `Error` envelope; stable `Code` set; EN-`Message` as data; CLI exit 0/1/2/130; MCP `IsError` + `StructuredContent`. |
| 2   | Help-text source | Hand-author i18n keys + codegen scaffold tool seeds `en.template.toml` from spec (not loaded at runtime).                         |
| 3   | Company scoping  | Hybrid: server default + optional per-tool override; non-co ops skip param.                                                       |
| 4   | Pagination caps  | CLI default unlimited; MCP default 25 / ceil 100; `ListMeta{Truncated,NextPage,Returned,Cancelled}`.                              |
| 5   | Multipart        | Default off in MCP; `--enable-attachments` flag exposes path-only param.                                                          |
| 6   | Read-only lock   | Codegen `mutating.gen.go` from OAS HTTP method; no RoundTripper belt.                                                             |
| 7   | Units invariants | Basis-points int for tax rates; custom `ops.Date` for civil dates; reflect-walk test.                                             |
| 8   | Cancellation     | List: partial `Ok` + `meta.cancelled=true`. Non-list: `Error` `code=cancelled`. Exit 130.                                         |
| 9   | Concurrency UX   | Sem=1, drop preemptive ticker, add 429-aware backoff, progress signals (stderr / MCP notifications).                              |
| 10  | OAuth seam       | Bare token string in storage v1; migrate at phase 2.                                                                              |
| 11  | MCP transport    | stdio + streamable HTTP (`--transport` flag).                                                                                     |
| 12  | sfiber migration | Out of scope; deferred.                                                                                                           |
| 13  | Logging          | slog stderr text+JSON; `-v` / `-vv` levels; `X-Request-Id` capture + `Authorization` redaction.                                   |
| 14  | Time zones       | OAS lint (`fiken-spec-lint`) + fixture parse test.                                                                                |
| 15  | Webhooks         | None in spec — confirmed out of scope.                                                                                            |
| 16  | nb maintenance   | CI gate on missing keys + self-review on PR.                                                                                      |
| 17  | Auth UX          | `login` / `status` / `logout` / `list`; verify-before-save; keyring + chmod-0600 file fallback.                                   |
