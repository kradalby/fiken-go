# fiken-go — Plan C: Remaining 16 Fiken Tags

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan. Steps use checkbox (`- [ ]`) syntax.

**Goal:** Replicate Plan B's companies pattern across the remaining 16 Fiken tags so every operationId in `api/fiken-openapi.yaml` is reachable through the CLI (`fiken <tag> <op>`) and through MCP (`tags/list` enumerates them). After Plan C, the CLI covers ~150 operations end-to-end, the parity test enforces CLI-vs-MCP byte equality for every registered tool, and `mockfiken` covers all overrides used by tests.

**Architecture:** No new packages or layers — Plan C is bulk replication of the per-tag pattern Plan B established with `ops/companies.go`, `cli/companies.go`, and the matching MCP tool registrations. Each tag becomes:

- One `ops/<tag>.go` file with `In*` / `Out*` types and methods on `*ops.Client`.
- One `cli/<tag>.go` file with subcommands following the auth/companies pattern.
- New entries in `ops/names.go` (`Op*` consts) and `ops/Registry` (CompanyScoped flag).
- New tools registered in `mcp/server.go` (or split into `mcp/<tag>.go` if `server.go` grows beyond ~500 lines — see Task 17 in this plan).
- New i18n entries in `i18n/locales/en.toml` AND `nb.toml` (with bilingual completeness gated by the `i18n-keys` hook).
- New parity-test cases in `cli/parity_test.go` for every registered op.

**Tech Stack:** Same as Plan B. No new dependencies.

**Spec reference:** `docs/superpowers/specs/2026-05-15-fiken-go-design.md`. This plan implements the remaining 16 Fiken tags listed in the spec's repo layout (§"Repo layout", under `ops/`).

**Prerequisites from Plan B:**

- `ops/{result,errors,paging,date,ratelimit,units_test,client,companies}.go` exist and compile.
- `ops/names.go` has Registry; `mutating.gen.go` lists all 152 ops.
- `auth/`, `config/`, `output/`, `i18n/`, `mockfiken/` packages work end-to-end.
- `cli/{root,context,progress,companies,auth,mcp}.go` exist; `cli/parity_test.go` passes for companies.
- `mcp/{server,readonly,transport,progress}.go` exist; companies tools registered.
- `cmd/fiken/main.go` is the entrypoint.
- `.pre-commit-config.yaml` has `i18n-keys` and `oas-units` hooks active.
- All Plan B verify gates green.

---

## The per-tag template (read once, apply 16 times)

This is the mechanical pattern every Plan C task follows. The template is illustrated for the **contacts** tag (Task 1) because it's the smallest non-companies tag and exhibits every concern (list + get + create + update + delete + sub-resource list). Tasks 2–16 reference this template and call out tag-specific deltas.

### Files per tag

```
ops/<tag>.go          types + methods (~300 lines including In/Out and translates)
cli/<tag>.go          ff/v4 subcommands (~150 lines)
mcp/<tag>.go          tool registrations (only if mcp/server.go would exceed ~500 lines)
ops/<tag>_test.go     happy + 1 error per op via mockfiken (~150 lines)
i18n/locales/en.toml  append: ops.<op>.{summary,when_to_use,returns,example,flags.*}
i18n/locales/nb.toml  append: same keys, bokmål translation
mockfiken/handler_impl.go  per-method overrides if tests need shaped data
cli/parity_test.go    append cases for each op
```

### Steps per tag (TDD)

1. **Identify the tag's ops** from `api/fiken-openapi.yaml` — grep for `tags:\s*\n\s*-\s*<tag>` then list neighboring `operationId:` lines. Record HTTP method + path + summary.
2. **Add the `Op*` constants** to `ops/names.go` and the matching Registry entries (CompanyScoped: most are true; the two exceptions are `user_*` and `companies_*`).
3. **Write the failing test** in `ops/<tag>_test.go` — at minimum one happy-path test per op and one error-path test per op (404 or 422).
4. **Implement `ops/<tag>.go`** — `In*` / `Out*` structs, method on `*ops.Client`, translate function(s) wired to the ogen-generated client method. Watch the units invariant (int64 for money, basis-points int for tax rates, `ops.Date` for civil dates, `time.Time` for datetimes).
5. **Update `ops/units_test.go`'s `outStructs()` slice** to include the new `Out*` types.
6. **Run the tests, confirm pass.**
7. **Implement `cli/<tag>.go`** — `Add<Tag>` wires a parent `<tag>` subcommand under root + one child per op. Each child uses `sf.build(...)` for session, calls `ops.Client.<Op>`, writes to `Renderer`. **Required flags** match the op's `In*` JSON tags (e.g. `--max-results`, `--company` for company-scoped).
8. **Register the new commands** in `cli/root.go`: add `if err := Add<Tag>(root, ...); err != nil { return nil, err }`.
9. **Register MCP tools** in `mcp/server.go` (or `mcp/<tag>.go` if extracting): one `mcpsdk.AddTool` per op, gated by `AllowOp(opts.Mode, opName)`. Description from `i18n.T(lang, "ops.<op>.summary", nil)`.
10. **Add i18n entries** to BOTH `en.toml` and `nb.toml` for every new op: `summary`, `when_to_use`, `returns`, `example`, plus a `flags.<flag>` entry per declared flag.
11. **Add parity test cases** in `cli/parity_test.go` for every new op. Use `mockfiken.Set` to register a deterministic response shape so both frontends compute byte-identical envelopes. Cover both Ok and Error paths.
12. **Update `mockfiken/handler_impl.go`** if your tests need overrides for tag-specific operations (default zero-value works for most happy-path no-data cases).
13. **Run the full test suite, hooks, and codegen-clean.** Commit:

```bash
git add ops/<tag>.go ops/<tag>_test.go cli/<tag>.go mcp/<tag>.go \
        cli/root.go mcp/server.go ops/units_test.go ops/names.go \
        i18n/locales/en.toml i18n/locales/nb.toml \
        mockfiken/handler_impl.go cli/parity_test.go
git commit -m "feat(<tag>): expose <N> ops via CLI + MCP

Adds <op_name>, <op_name>, ... as both fiken <tag> <op> subcommands
and MCP tools. Parity test extended.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

### Verification per tag

After each tag commit:

- `prek run --all-files` green (includes `i18n-keys` parity check).
- `go test ./...` green (includes parity test).
- `go generate ./...` zero diff.
- Spot-check: `fiken <tag> --help` shows English; `fiken --lang=nb <tag> --help` shows bokmål.

---

## Tag inventory (run once, before starting)

Discover the ops per tag with this small script — the output drives Tasks 1–16's specific deltas:

```bash
nix develop -c bash -lc '
  awk "/^\s+tags:/{getline; gsub(/^\s+-\s+/,\"\"); tag=\$0; next} /operationId:/{gsub(/^\s+operationId:\s+/,\"\"); print tag\" \"\$0}" \
    api/fiken-openapi.yaml | sort -u
'
```

Expected output (truncated): pairs of `<tag> <operationId>` covering all 152 ops. Use this list to identify which ops fall under each Plan C task.

If a tag's HTTP method varies across ops (some GET, some POST), each op still maps deterministically via `ops.IsMutating(snake_case(operationId))` — no per-tag mutating bit to maintain.

---

### Task 1: `contacts` tag — full template walkthrough

The **contacts** tag is implemented in full detail here as the canonical Plan C task. Subsequent tag tasks reference this one for pattern; only the tag-specific deltas appear in their own task bodies.

**Files:**

- Create: `ops/contacts.go`
- Create: `ops/contacts_test.go`
- Create: `cli/contacts.go`
- Modify: `ops/names.go`, `ops/units_test.go`, `cli/root.go`, `mcp/server.go`, `i18n/locales/en.toml`, `i18n/locales/nb.toml`, `cli/parity_test.go`.

**Contacts tag ops (verify via `awk` from inventory step):**

- `getContacts` → GET `/companies/{companySlug}/contacts` → `contacts_list`
- `createContact` → POST `/companies/{companySlug}/contacts` → `contacts_create`
- `getContact` → GET `/companies/{companySlug}/contacts/{contactId}` → `contacts_get`
- `updateContact` → PUT `/companies/{companySlug}/contacts/{contactId}` → `contacts_update`
- `deleteContact` → DELETE `/companies/{companySlug}/contacts/{contactId}` → `contacts_delete`
- `getContactPersons` → GET `/companies/{companySlug}/contacts/{contactId}/contactPerson` → `contacts_persons_list`
- `createContactPerson` → POST `.../contactPerson` → `contacts_persons_create`
- `getContactPerson` → GET `.../contactPerson/{contactPersonId}` → `contacts_persons_get`
- `updateContactPerson` → PUT `.../contactPerson/{contactPersonId}` → `contacts_persons_update`
- `deleteContactPerson` → DELETE `.../contactPerson/{contactPersonId}` → `contacts_persons_delete`
- (verify the list against the actual spec; this is from memory of similar Fiken APIs)

That's ~10 ops in one tag. For the first vertical of Plan C, implement **all** of them to confirm the pattern scales.

- [ ] **Step 1.1: Identify the actual op list**

```bash
nix develop -c bash -lc '
  awk "/^\s+tags:/{getline; gsub(/^\s+-\s+/,\"\"); tag=\$0; next}
       /operationId:/{gsub(/^\s+operationId:\s+/,\"\"); if(tag==\"contacts\") print \$0}" \
    api/fiken-openapi.yaml
'
```

Record the actual operationIds; if the above list is wrong/incomplete, adjust the task accordingly.

- [ ] **Step 1.2: Add Op\* constants and Registry entries**

In `ops/names.go`, append to the const block:

```go
const (
    // ... existing ...
    OpContactsList          = "contacts_list"
    OpContactsCreate        = "contacts_create"
    OpContactsGet           = "contacts_get"
    OpContactsUpdate        = "contacts_update"
    OpContactsDelete        = "contacts_delete"
    OpContactsPersonsList   = "contacts_persons_list"
    OpContactsPersonsCreate = "contacts_persons_create"
    OpContactsPersonsGet    = "contacts_persons_get"
    OpContactsPersonsUpdate = "contacts_persons_update"
    OpContactsPersonsDelete = "contacts_persons_delete"
)
```

Append Registry entries (all `CompanyScoped: true`):

```go
OpContactsList:          {Mutating: IsMutating(OpContactsList), CompanyScoped: true},
OpContactsCreate:        {Mutating: IsMutating(OpContactsCreate), CompanyScoped: true},
// ... and the other 8.
```

- [ ] **Step 1.3: Write the failing test**

`ops/contacts_test.go` covers 1 happy + 1 error per op. Use the override registry for happy paths:

```go
package ops

import (
    "context"
    "testing"
    "github.com/kradalby/fiken-go/auth"
    "github.com/kradalby/fiken-go/fiken"
)

func TestContactsListAgainstMock(t *testing.T) {
    mock := startMockForTest(t)
    mock.Set(OpContactsList, []fiken.Contact{ /* one zero-value contact */ })
    c, _ := New(context.Background(), Options{
        BaseURL: mock.URL(), Auth: auth.FlagSource{Value: "t"},
    })
    res := c.ContactsList(context.Background(), ContactsListIn{Company: "acme"})
    if res.Error != nil {
        t.Fatalf("err: %+v", res.Error)
    }
    if res.Ok == nil {
        t.Fatal("nil Ok")
    }
}

func TestContactsListMissingCompany(t *testing.T) {
    c, _ := New(context.Background(), Options{BaseURL: "x", Auth: auth.FlagSource{Value: "t"}})
    res := c.ContactsList(context.Background(), ContactsListIn{})
    if res.Error == nil || res.Error.Code != CodeValidation {
        t.Fatalf("want validation, got %+v", res)
    }
}

// ... 18 more (1 happy + 1 error for each of the other 9 ops)
```

(The implementer writes all 20 tests; the template above shows the shape.)

- [ ] **Step 1.4: Confirm fail.** `go test ./ops/...` → many `undefined: ContactsList*` etc.

- [ ] **Step 1.5: Implement `ops/contacts.go`**

Pattern (one example op shown — implementer replicates across the 10):

```go
package ops

import (
    "context"
    "github.com/kradalby/fiken-go/fiken"
)

// ContactsListIn carries paged-list input for contacts under one company.
type ContactsListIn struct {
    Company    string `json:"company"`
    MaxResults int    `json:"max_results,omitempty"`
    PageSize   int    `json:"page_size,omitempty"`
    Page       int    `json:"page,omitempty"`
}

// ContactOut is the canonical single-contact shape.
type ContactOut struct {
    ContactID          int64       `json:"contact_id,omitempty"`
    Name               string      `json:"name"`
    Email              string      `json:"email,omitempty"`
    OrganizationNumber string      `json:"organization_number,omitempty"`
    PhoneNumber        string      `json:"phone_number,omitempty"`
    Currency           string      `json:"currency,omitempty"`
    Language           string      `json:"language,omitempty"`
    Inactive           bool        `json:"inactive,omitempty"`
    LastModifiedDate   ops.Date    `json:"last_modified_date,omitempty"`
    CreatedAt          time.Time   `json:"created_at,omitempty"`
}

func (c ContactOut) TableHeader() []string {
    return []string{"ID", "NAME", "EMAIL", "ORG.NR"}
}
func (c ContactOut) TableRow() []string {
    return []string{
        fmt.Sprintf("%d", c.ContactID),
        c.Name, c.Email, c.OrganizationNumber,
    }
}

type ContactsListOut = ListOut[ContactOut]

func (c *Client) ContactsList(ctx context.Context, in ContactsListIn) Result[ContactsListOut] {
    if in.Company == "" {
        return Err[ContactsListOut](&Error{
            Code: CodeValidation, Message: "company is required",
            Op: OpContactsList,
        })
    }
    resp, err := c.gen.GetContacts(ctx, fiken.GetContactsParams{
        CompanySlug: in.Company,
        // pageSize / page / etc. follow the params struct's shape.
    })
    if err != nil {
        return Err[ContactsListOut](MapErr(OpContactsList, err))
    }
    return Ok[ContactsListOut](translateContactsList(resp))
}

func translateContactsList(resp any) ContactsListOut {
    // Implementer: cast resp to the ogen-emitted response type
    // (likely *fiken.GetContactsOKApplicationJSON or similar),
    // iterate the slice, map each item to ContactOut. Watch unit
    // rules: int64 money, ops.Date for date-only, time.Time for
    // *At fields.
    return ContactsListOut{}
}

// ... 9 more methods following the same pattern.
```

- [ ] **Step 1.6: Update `ops/units_test.go`**

```go
func outStructs() []reflect.Type {
    return []reflect.Type{
        reflect.TypeOf(CompanyOut{}),
        reflect.TypeOf(CompaniesListOut{}),
        reflect.TypeOf(ContactOut{}),
        reflect.TypeOf(ContactsListOut{}),
        reflect.TypeOf(ContactPersonOut{}),
        reflect.TypeOf(ContactPersonsListOut{}),
    }
}
```

- [ ] **Step 1.7: Pass tests.** `go test ./ops/...` green.

- [ ] **Step 1.8: Implement `cli/contacts.go`**

Mirrors `cli/companies.go`:

```go
package cli

import (
    "context"
    "fmt"
    "io"
    "github.com/peterbourgon/ff/v4"
    "github.com/kradalby/fiken-go/ops"
)

func AddContacts(root *ff.Command, stdout, stderr io.Writer, sf *sessionFactory) error {
    set := ff.NewFlagSet("contacts").SetParent(root.Flags)
    contactsCmd := &ff.Command{
        Name: "contacts", Usage: "fiken contacts <subcommand>",
        ShortHelp: "Manage Fiken contacts.", Flags: set,
    }

    // list
    listSet := ff.NewFlagSet("list").SetParent(set)
    var maxResults int
    listSet.IntVar(&maxResults, 0, "max-results", 0, "Max contacts to return (0 = unlimited)")
    contactsCmd.Subcommands = append(contactsCmd.Subcommands, &ff.Command{
        Name: "list", Usage: "fiken contacts list", Flags: listSet,
        Exec: func(ctx context.Context, _ []string) error {
            ctx, err := sf.build(ctx, stdout, stderr)
            if err != nil { return err }
            res := Client(ctx).ContactsList(ctx, ops.ContactsListIn{
                Company:    sf.profile(ctx).Company,
                MaxResults: maxResults,
            })
            return Renderer(ctx).Render(res)
        },
    })

    // get, create, update, delete — same pattern.

    root.Subcommands = append(root.Subcommands, contactsCmd)
    return nil
}
```

- [ ] **Step 1.9: Register in `cli/root.go`**

Add inside `Root`:

```go
if err := AddContacts(root, stdout, stderr, sf); err != nil {
    return nil, err
}
```

- [ ] **Step 1.10: Register MCP tools in `mcp/server.go`**

Inside `New(opts Options)`:

```go
if AllowOp(opts.Mode, ops.OpContactsList) {
    mcpsdk.AddTool(srv, &mcpsdk.Tool{
        Name:        ops.OpContactsList,
        Description: opts.Bundle.T(opts.Lang, "ops.contacts_list.summary", nil),
    }, makeContactsListHandler(opts.Client))
}
// ... and the other 9.
```

If `mcp/server.go` grows past ~500 lines after a few tag additions, extract per-tag registrars into `mcp/contacts.go` etc. Defer until file size warrants.

- [ ] **Step 1.11: Add i18n entries**

`i18n/locales/en.toml`:

```toml
[ops.contacts_list]
summary     = "List contacts for a company."
when_to_use = "Use to find a contactId before invoicing or attaching files."
returns     = "Object with `items` (each: contact_id, name, email, ...) and pagination meta."
example     = "fiken contacts list --company acme"

[ops.contacts_list.flags]
company     = "Slug of the company to query. Required. Example: --company=acme-as"
max_results = "Max contacts across all pages. Default: unlimited (CLI), 25 (MCP). Example: --max-results=50"

# ... and the other 9 ops.
```

`i18n/locales/nb.toml`: same keys, bokmål prose. Reviewed by maintainer per the `nb` quality policy.

- [ ] **Step 1.12: Extend `cli/parity_test.go`**

Add one parity case per op:

```go
func TestParityContactsList(t *testing.T) {
    mock := mockfiken.New(t)
    mock.Set(ops.OpContactsList, []fiken.Contact{ /* deterministic data */ })

    // CLI --json path.
    var stdout, stderr bytes.Buffer
    cmd, _ := Root(&stdout, &stderr)
    t.Setenv("FIKEN_TOKEN", "test")
    if err := cmd.ParseAndRun(context.Background(), []string{
        "--json", "--config", "/dev/null", "--company", "acme",
        "contacts", "list",
    }); err != nil {
        t.Fatalf("CLI: %v", err)
    }

    // MCP path.
    client, _ := ops.New(context.Background(), ops.Options{
        BaseURL: mock.URL(), Auth: auth.FlagSource{Value: "test"},
    })
    srv, _ := mcppkg.New(mcppkg.Options{
        Client: client, Mode: mcppkg.ModeReadOnly,
        Bundle: i18n.MustLoad(), Lang: "en",
    })
    // ... (same in-memory transport setup as TestParityCompaniesList)

    // Compare JSON-equal.
    if !jsonEqual(stdout.Bytes(), resp.StructuredContent) {
        t.Fatalf("envelope mismatch")
    }
}
```

Extract a `jsonEqual(a, b []byte) bool` helper into `cli/parity_helpers.go` so each case stays short.

- [ ] **Step 1.13: Run full verification**

```bash
nix develop -c prek run --all-files
nix develop -c go test -race -count=1 ./...
```

Both must be green.

- [ ] **Step 1.14: Commit**

Single commit per the file list above. Commit message format:

```
feat(contacts): expose 10 ops via CLI + MCP

Adds contacts_list, contacts_create, contacts_get, contacts_update,
contacts_delete, contacts_persons_{list,create,get,update,delete}
through fiken contacts <op> and MCP. Parity tests cover all 10.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
```

---

### Tasks 2–16: Remaining tags (one task per tag)

Each task below follows the template above. Only tag-specific deltas appear here: op list, special unit/type concerns, sub-resource notes. Use Task 1 as the worked example.

### Task 2: `accounts`

Likely ops (verify via `awk`): `getAccount`, `getAccounts`, `getAccountBalances`, `getAccountBalance`.

Unit concerns:

- `*Balance` fields → `int64` øre.
- Account number format → opaque string, no special type.

CompanyScoped: yes for all.

---

### Task 3: `bankAccounts`

Likely ops: `getBankAccounts`, `getBankAccount`, `createBankAccount`, plus reconciliation sub-resources.

Unit concerns:

- `*Balance` → `int64` øre.
- `bankAccountNumber` (Norwegian KID) → string.

CompanyScoped: yes.

---

### Task 4: `journalEntries`

Likely ops: `getJournalEntries`, `getJournalEntry`, `createGeneralJournalEntry`, `addAttachmentToJournalEntry` (multipart — opt-out of MCP by default per spec §"Multipart attachments"; CLI exposes `fiken journal-entries attach <id> --file path`).

Unit concerns:

- `amount`, `debit`, `credit` → `int64` øre.
- `transactionDate` → `ops.Date`.

CompanyScoped: yes. Note: the dash form `journal-entries` for CLI subcommand path; underscore form `journal_entries` for op ids and i18n keys (consistent across the codebase).

---

### Task 5: `transactions`

Likely ops: `getTransactions`, `getTransaction`.

Unit concerns: same as journalEntries.

CompanyScoped: yes.

---

### Task 6: `invoices` (LARGE)

This is the biggest tag — ~20 ops. Includes drafts, counters, sendings.

Op highlights:

- `getInvoices`, `getInvoice`, `createInvoice`, `sendInvoice`, `createInvoiceCounter`.
- Drafts: `getInvoiceDrafts`, `createInvoiceDraft`, `getInvoiceDraft`, `updateInvoiceDraft`, `deleteInvoiceDraft`, `createInvoiceFromDraft`, `getInvoiceDraftAttachments`, `addAttachmentToInvoiceDraft` (multipart — opt-out), `setInvoiceDraftCounter`.
- Lines: each invoice has line items — `productId`, `description`, `unitPrice` (int64), `quantity` (int64 or basis-points int), `vatType` (string enum from Fiken), `vatRate` (basis-points int).

Unit concerns:

- All monetary fields → `int64` øre.
- `vatRate` → basis-points int (25% = 2500, 11.11% = 1111, 0% = 0).
- `dueDate`, `issueDate` → `ops.Date`.
- `createdAt` → `time.Time`.

CompanyScoped: yes for all.

**Sub-task split**: invoices has so many ops that one commit is unwieldy. Split into 3 commits:

- **6a**: `getInvoices`, `getInvoice`, `createInvoice`, `sendInvoice`, `createInvoiceCounter` (5 core ops).
- **6b**: invoice draft ops (CRUD + create from draft + counter).
- **6c**: attachment ops (multipart) — opt-out in MCP by default, CLI-only path.

Each sub-task follows the template and produces its own commit.

---

### Task 7: `creditNotes`

Ops: `getCreditNotes`, `getCreditNote`, `getCreditNoteDrafts`, `createCreditNoteDraftFromInvoice`, full draft CRUD, `addAttachmentToCreditNoteDraft` (multipart).

Unit concerns: same as invoices.

CompanyScoped: yes.

Sub-task split: one commit if <10 ops total, otherwise split.

---

### Task 8: `offers`

Ops: `getOffers`, `getOffer`, full draft CRUD, `addAttachmentToOfferDraft` (multipart).

Same unit/scope concerns as invoices/creditNotes.

---

### Task 9: `orderConfirmations`

Ops: `getOrderConfirmations`, `getOrderConfirmation`, draft CRUD.

No attachment endpoints in this tag.

---

### Task 10: `products`

Ops: `getProducts`, `createProduct`, `getProduct`, `updateProduct`, `deleteProduct`, `getProductSalesReport`, `createProductSalesReport`.

Unit concerns:

- `incomeAccount`, `vatType` → string enum.
- `price` → `int64` øre.
- `active` → bool.

The sales report ops (`createProductSalesReport`) return paginated reports — verify the response schema for monetary fields.

CompanyScoped: yes.

---

### Task 11: `sales`

Ops: `getSales`, `getSale`, `createSale`, `deleteSale`, `payments` sub-resource (CRUD), `attachments` sub-resource (likely multipart).

Unit concerns: same as invoices.

CompanyScoped: yes.

Sub-task split likely.

---

### Task 12: `purchases`

Ops: `getPurchases`, `getPurchase`, `createPurchase`, `deletePurchase`, plus payments/attachments sub-resources.

Same shape as sales.

---

### Task 13: `inbox`

Ops: `getInboxDocuments`, `getInboxDocument`, `sendInboxDocument` (multipart — file upload), `deleteInboxDocument`.

This is fully attachment-centric — multipart endpoints are the norm here. CLI exposes them as usual; MCP is opt-out unless `--enable-attachments` (per spec §"Multipart attachments").

---

### Task 14: `projects`

Ops: `getProjects`, `getProject`, `createProject`, `updateProject`, `deleteProject`.

Simple CRUD. No money fields per op; relations are by reference.

---

### Task 15: `user`

Ops: `getUser` (only one).

`getUser` returns the authenticated user's profile (name, email).

CompanyScoped: **no** — this is a user-level op. Set `CompanyScoped: false` in the Registry; MCP tool's InputSchema omits the `company` param.

i18n: small task, only one entry pair to add.

---

### Task 16: Any straggler tags

After the 15 above, run the inventory script again to verify no tag was missed:

```bash
nix develop -c bash -lc '
  awk "/^\s+tags:/{getline; gsub(/^\s+-\s+/,\"\"); print \$0}" \
    api/fiken-openapi.yaml | sort -u
'
```

If a tag appears that wasn't planned (e.g. a future addition to the upstream spec), implement it following the template and append a Task 16-bis commit.

---

### Task 17: Extract MCP per-tag files if `mcp/server.go` is too big

By the time all 16 tag-tasks are landed, `mcp/server.go` has 150+ `AddTool` blocks. If it crosses ~500 lines, extract:

- `mcp/server.go`: keeps `New()` shell + transport wiring.
- `mcp/companies.go`, `mcp/contacts.go`, ... : each defines `register<Tag>(srv, opts)` that `New()` calls.

This task is gated on file size — only if needed. Single refactor commit.

---

### Task 18: End-of-Plan-C verification

- [ ] `git log --oneline | head -50` shows roughly 20-30 new commits (Plan B + Plan C).
- [ ] `nix develop -c prek run --all-files` green.
- [ ] `nix develop -c go test -race -count=1 ./...` green (covers ~150 op-level tests + ~150 parity tests).
- [ ] `nix develop -c go generate ./...` zero diff.
- [ ] `nix build .#fiken` works; `./result/bin/fiken --help` lists every tag subcommand.
- [ ] MCP smoke: `claude mcp add fiken -- ./result/bin/fiken mcp` then call `tools/list` — confirm count matches the non-mutating subset of registered ops.
- [ ] Help completeness: every op has en + nb entries; `i18n-keys` hook never failed.
- [ ] Spec-lint: `oas-units` hook passed (with the known wall-clock `--ignore startTime,endTime` from Plan B Task 21).

Success → Plan D's polish phase begins.

---

## Self-review notes

- Per-tag tests live in `ops/<tag>_test.go`. Each test sets up `mockfiken` deterministically; no real Fiken token anywhere.
- `mockfiken/handler_impl.go` will grow as more tags need shaped responses. If it gets unwieldy, extract `mockfiken/<tag>.go` files mirroring the per-tag split in `ops/`.
- The parity test relies on `mockfiken` returning **byte-identical** envelopes regardless of which frontend triggered it. If a future tag introduces non-deterministic ordering (e.g. map-valued response fields), the parity test will flake — sort keys server-side via `json.Marshal` semantics (Go orders map keys alphabetically) and don't introduce slice-of-map shapes.
- Multipart attachment ops: don't register them in MCP unless `Options.EnableAttachments` is true. CLI always exposes them as `fiken <tag> attach <id> --file <path>`. Plan D's Task 4 wires `--enable-attachments` end-to-end with one happy-path test.
- Tax rate edge: `vatRate` of `11.11%` becomes `1111` basis points. The units invariant test enforces this is `int`, not `float64`. Translate functions must do the multiply-by-100 conversion from spec's number representation.
- Date fields named `*Date` (no time) map to `ops.Date`. Datetime fields named `*At` / `*DateTime` map to `time.Time`. The `oas-units` hook catches drift if a future spec change introduces a `*Date` field without `format: date`.
