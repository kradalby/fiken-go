package mcp

import (
	"context"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/kradalby/fiken-go/i18n"
	"github.com/kradalby/fiken-go/ops"
)

// serverName + serverVersion identify the implementation in
// initialize handshakes. Version stays "0.1.0" until we cut a release.
const (
	serverName    = "fiken-go"
	serverVersion = "0.1.0"
)

// Options configures a new server.
type Options struct {
	Client *ops.Client
	Mode   Mode
	Bundle *i18n.Bundle
	Lang   string
	// EnableAttachments is reserved for Plan D's attachment tools; Plan
	// B ignores it. Keeping the field surfaces the future toggle in
	// callers' build code without a follow-up signature change.
	EnableAttachments bool
	// CapGated installs a receiving-middleware that consults the
	// per-request Capability (placed in ctx by the tsnet HTTP layer)
	// before letting a tools/call through. Used only by the tsnet
	// transport — stdio and plain HTTP leave it false.
	CapGated bool
}

// New returns a configured MCP server with companies_{list,get}
// registered (subject to Mode filter).
func New(opts Options) (*mcpsdk.Server, error) {
	srv := mcpsdk.NewServer(&mcpsdk.Implementation{
		Name:    serverName,
		Version: serverVersion,
	}, nil)

	if AllowOp(opts.Mode, ops.OpCompaniesList) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpCompaniesList,
			Description: opts.Bundle.T(opts.Lang, "ops.companies_list.summary", nil),
		}, makeCompaniesListHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpCompaniesGet) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpCompaniesGet,
			Description: opts.Bundle.T(opts.Lang, "ops.companies_get.summary", nil),
		}, makeCompaniesGetHandler(opts.Client))
	}

	if AllowOp(opts.Mode, ops.OpContactsList) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpContactsList,
			Description: opts.Bundle.T(opts.Lang, "ops.contacts_list.summary", nil),
		}, makeContactsListHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpContactsGet) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpContactsGet,
			Description: opts.Bundle.T(opts.Lang, "ops.contacts_get.summary", nil),
		}, makeContactsGetHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpContactsCreate) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpContactsCreate,
			Description: opts.Bundle.T(opts.Lang, "ops.contacts_create.summary", nil),
		}, makeContactsCreateHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpContactsUpdate) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpContactsUpdate,
			Description: opts.Bundle.T(opts.Lang, "ops.contacts_update.summary", nil),
		}, makeContactsUpdateHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpContactsDelete) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpContactsDelete,
			Description: opts.Bundle.T(opts.Lang, "ops.contacts_delete.summary", nil),
		}, makeContactsDeleteHandler(opts.Client))
	}
	// Multipart attach gated behind EnableAttachments — mirrors the
	// journal-entries / invoices pattern.
	if opts.EnableAttachments && AllowOp(opts.Mode, ops.OpContactsAttachmentsAttach) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpContactsAttachmentsAttach,
			Description: opts.Bundle.T(opts.Lang, "ops.contacts_attachments_attach.summary", nil),
		}, makeContactsAttachmentsAttachHandler(opts.Client))
	}

	if AllowOp(opts.Mode, ops.OpContactsPersonsList) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpContactsPersonsList,
			Description: opts.Bundle.T(opts.Lang, "ops.contacts_persons_list.summary", nil),
		}, makeContactsPersonsListHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpContactsPersonsGet) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpContactsPersonsGet,
			Description: opts.Bundle.T(opts.Lang, "ops.contacts_persons_get.summary", nil),
		}, makeContactsPersonsGetHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpContactsPersonsCreate) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpContactsPersonsCreate,
			Description: opts.Bundle.T(opts.Lang, "ops.contacts_persons_create.summary", nil),
		}, makeContactsPersonsCreateHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpContactsPersonsUpdate) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpContactsPersonsUpdate,
			Description: opts.Bundle.T(opts.Lang, "ops.contacts_persons_update.summary", nil),
		}, makeContactsPersonsUpdateHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpContactsPersonsDelete) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpContactsPersonsDelete,
			Description: opts.Bundle.T(opts.Lang, "ops.contacts_persons_delete.summary", nil),
		}, makeContactsPersonsDeleteHandler(opts.Client))
	}

	if AllowOp(opts.Mode, ops.OpAccountsList) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpAccountsList,
			Description: opts.Bundle.T(opts.Lang, "ops.accounts_list.summary", nil),
		}, makeAccountsListHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpAccountsGet) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpAccountsGet,
			Description: opts.Bundle.T(opts.Lang, "ops.accounts_get.summary", nil),
		}, makeAccountsGetHandler(opts.Client))
	}

	if AllowOp(opts.Mode, ops.OpBankAccountsList) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpBankAccountsList,
			Description: opts.Bundle.T(opts.Lang, "ops.bank_accounts_list.summary", nil),
		}, makeBankAccountsListHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpBankAccountsGet) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpBankAccountsGet,
			Description: opts.Bundle.T(opts.Lang, "ops.bank_accounts_get.summary", nil),
		}, makeBankAccountsGetHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpBankAccountsCreate) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpBankAccountsCreate,
			Description: opts.Bundle.T(opts.Lang, "ops.bank_accounts_create.summary", nil),
		}, makeBankAccountsCreateHandler(opts.Client))
	}

	if AllowOp(opts.Mode, ops.OpJournalEntriesList) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpJournalEntriesList,
			Description: opts.Bundle.T(opts.Lang, "ops.journal_entries_list.summary", nil),
		}, makeJournalEntriesListHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpJournalEntriesGet) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpJournalEntriesGet,
			Description: opts.Bundle.T(opts.Lang, "ops.journal_entries_get.summary", nil),
		}, makeJournalEntriesGetHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpJournalEntriesCreate) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpJournalEntriesCreate,
			Description: opts.Bundle.T(opts.Lang, "ops.journal_entries_create.summary", nil),
		}, makeJournalEntriesCreateHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpJournalEntriesAttachmentsList) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpJournalEntriesAttachmentsList,
			Description: opts.Bundle.T(opts.Lang, "ops.journal_entries_attachments_list.summary", nil),
		}, makeJournalEntriesAttachmentsListHandler(opts.Client))
	}
	// Multipart attach is gated behind EnableAttachments. Plan D wires
	// the toggle into the binary-upload path; for now it stays opt-in
	// and surfaces the stub error when explicitly enabled.
	if opts.EnableAttachments && AllowOp(opts.Mode, ops.OpJournalEntriesAttachmentsAttach) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpJournalEntriesAttachmentsAttach,
			Description: opts.Bundle.T(opts.Lang, "ops.journal_entries_attachments_attach.summary", nil),
		}, makeJournalEntriesAttachmentsAttachHandler(opts.Client))
	}

	if AllowOp(opts.Mode, ops.OpTransactionsList) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpTransactionsList,
			Description: opts.Bundle.T(opts.Lang, "ops.transactions_list.summary", nil),
		}, makeTransactionsListHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpTransactionsGet) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpTransactionsGet,
			Description: opts.Bundle.T(opts.Lang, "ops.transactions_get.summary", nil),
		}, makeTransactionsGetHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpTransactionsDelete) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpTransactionsDelete,
			Description: opts.Bundle.T(opts.Lang, "ops.transactions_delete.summary", nil),
		}, makeTransactionsDeleteHandler(opts.Client))
	}

	if AllowOp(opts.Mode, ops.OpInvoicesList) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpInvoicesList,
			Description: opts.Bundle.T(opts.Lang, "ops.invoices_list.summary", nil),
		}, makeInvoicesListHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpInvoicesGet) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpInvoicesGet,
			Description: opts.Bundle.T(opts.Lang, "ops.invoices_get.summary", nil),
		}, makeInvoicesGetHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpInvoicesCreate) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpInvoicesCreate,
			Description: opts.Bundle.T(opts.Lang, "ops.invoices_create.summary", nil),
		}, makeInvoicesCreateHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpInvoicesUpdate) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpInvoicesUpdate,
			Description: opts.Bundle.T(opts.Lang, "ops.invoices_update.summary", nil),
		}, makeInvoicesUpdateHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpInvoicesSend) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpInvoicesSend,
			Description: opts.Bundle.T(opts.Lang, "ops.invoices_send.summary", nil),
		}, makeInvoicesSendHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpInvoicesCounterCreate) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpInvoicesCounterCreate,
			Description: opts.Bundle.T(opts.Lang, "ops.invoices_counter_create.summary", nil),
		}, makeInvoicesCounterCreateHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpInvoicesCounterGet) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpInvoicesCounterGet,
			Description: opts.Bundle.T(opts.Lang, "ops.invoices_counter_get.summary", nil),
		}, makeInvoicesCounterGetHandler(opts.Client))
	}

	if AllowOp(opts.Mode, ops.OpInvoicesDraftsList) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpInvoicesDraftsList,
			Description: opts.Bundle.T(opts.Lang, "ops.invoices_drafts_list.summary", nil),
		}, makeInvoiceDraftsListHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpInvoicesDraftsGet) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpInvoicesDraftsGet,
			Description: opts.Bundle.T(opts.Lang, "ops.invoices_drafts_get.summary", nil),
		}, makeInvoiceDraftsGetHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpInvoicesDraftsCreate) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpInvoicesDraftsCreate,
			Description: opts.Bundle.T(opts.Lang, "ops.invoices_drafts_create.summary", nil),
		}, makeInvoiceDraftsCreateHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpInvoicesDraftsUpdate) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpInvoicesDraftsUpdate,
			Description: opts.Bundle.T(opts.Lang, "ops.invoices_drafts_update.summary", nil),
		}, makeInvoiceDraftsUpdateHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpInvoicesDraftsDelete) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpInvoicesDraftsDelete,
			Description: opts.Bundle.T(opts.Lang, "ops.invoices_drafts_delete.summary", nil),
		}, makeInvoiceDraftsDeleteHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpInvoicesDraftsCreateFrom) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpInvoicesDraftsCreateFrom,
			Description: opts.Bundle.T(opts.Lang, "ops.invoices_drafts_create_from.summary", nil),
		}, makeInvoiceDraftsCreateFromHandler(opts.Client))
	}

	if AllowOp(opts.Mode, ops.OpInvoicesAttachmentsList) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpInvoicesAttachmentsList,
			Description: opts.Bundle.T(opts.Lang, "ops.invoices_attachments_list.summary", nil),
		}, makeInvoicesAttachmentsListHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpInvoicesDraftsAttachmentsList) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpInvoicesDraftsAttachmentsList,
			Description: opts.Bundle.T(opts.Lang, "ops.invoices_drafts_attachments_list.summary", nil),
		}, makeInvoiceDraftsAttachmentsListHandler(opts.Client))
	}
	// Multipart attaches stay gated behind EnableAttachments — matches the
	// journal-entries pattern. Plan D wires the binary-upload path; for
	// now the tools surface a stub error when explicitly enabled.
	if opts.EnableAttachments && AllowOp(opts.Mode, ops.OpInvoicesAttachmentsAttach) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpInvoicesAttachmentsAttach,
			Description: opts.Bundle.T(opts.Lang, "ops.invoices_attachments_attach.summary", nil),
		}, makeInvoicesAttachmentsAttachHandler(opts.Client))
	}
	if opts.EnableAttachments && AllowOp(opts.Mode, ops.OpInvoicesDraftsAttachmentsAttach) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpInvoicesDraftsAttachmentsAttach,
			Description: opts.Bundle.T(opts.Lang, "ops.invoices_drafts_attachments_attach.summary", nil),
		}, makeInvoiceDraftsAttachmentsAttachHandler(opts.Client))
	}

	if AllowOp(opts.Mode, ops.OpCreditNotesList) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpCreditNotesList,
			Description: opts.Bundle.T(opts.Lang, "ops.credit_notes_list.summary", nil),
		}, makeCreditNotesListHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpCreditNotesGet) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpCreditNotesGet,
			Description: opts.Bundle.T(opts.Lang, "ops.credit_notes_get.summary", nil),
		}, makeCreditNotesGetHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpCreditNotesSend) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpCreditNotesSend,
			Description: opts.Bundle.T(opts.Lang, "ops.credit_notes_send.summary", nil),
		}, makeCreditNotesSendHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpCreditNotesCounterCreate) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpCreditNotesCounterCreate,
			Description: opts.Bundle.T(opts.Lang, "ops.credit_notes_counter_create.summary", nil),
		}, makeCreditNotesCounterCreateHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpCreditNotesCounterGet) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpCreditNotesCounterGet,
			Description: opts.Bundle.T(opts.Lang, "ops.credit_notes_counter_get.summary", nil),
		}, makeCreditNotesCounterGetHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpCreditNotesFullCreate) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpCreditNotesFullCreate,
			Description: opts.Bundle.T(opts.Lang, "ops.credit_notes_full_create.summary", nil),
		}, makeCreditNotesFullCreateHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpCreditNotesPartialCreate) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpCreditNotesPartialCreate,
			Description: opts.Bundle.T(opts.Lang, "ops.credit_notes_partial_create.summary", nil),
		}, makeCreditNotesPartialCreateHandler(opts.Client))
	}

	if AllowOp(opts.Mode, ops.OpCreditNotesDraftsList) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpCreditNotesDraftsList,
			Description: opts.Bundle.T(opts.Lang, "ops.credit_notes_drafts_list.summary", nil),
		}, makeCreditNoteDraftsListHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpCreditNotesDraftsGet) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpCreditNotesDraftsGet,
			Description: opts.Bundle.T(opts.Lang, "ops.credit_notes_drafts_get.summary", nil),
		}, makeCreditNoteDraftsGetHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpCreditNotesDraftsCreate) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpCreditNotesDraftsCreate,
			Description: opts.Bundle.T(opts.Lang, "ops.credit_notes_drafts_create.summary", nil),
		}, makeCreditNoteDraftsCreateHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpCreditNotesDraftsUpdate) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpCreditNotesDraftsUpdate,
			Description: opts.Bundle.T(opts.Lang, "ops.credit_notes_drafts_update.summary", nil),
		}, makeCreditNoteDraftsUpdateHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpCreditNotesDraftsDelete) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpCreditNotesDraftsDelete,
			Description: opts.Bundle.T(opts.Lang, "ops.credit_notes_drafts_delete.summary", nil),
		}, makeCreditNoteDraftsDeleteHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpCreditNotesDraftsCreateFrom) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpCreditNotesDraftsCreateFrom,
			Description: opts.Bundle.T(opts.Lang, "ops.credit_notes_drafts_create_from.summary", nil),
		}, makeCreditNoteDraftsCreateFromHandler(opts.Client))
	}

	if AllowOp(opts.Mode, ops.OpCreditNotesDraftsAttachmentsList) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpCreditNotesDraftsAttachmentsList,
			Description: opts.Bundle.T(opts.Lang, "ops.credit_notes_drafts_attachments_list.summary", nil),
		}, makeCreditNoteDraftsAttachmentsListHandler(opts.Client))
	}
	// Multipart attach is gated behind EnableAttachments — matches the
	// invoices pattern. Plan D wires the binary-upload path; for now
	// the tool surfaces a stub error when explicitly enabled.
	if opts.EnableAttachments && AllowOp(opts.Mode, ops.OpCreditNotesDraftsAttachmentsAttach) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpCreditNotesDraftsAttachmentsAttach,
			Description: opts.Bundle.T(opts.Lang, "ops.credit_notes_drafts_attachments_attach.summary", nil),
		}, makeCreditNoteDraftsAttachmentsAttachHandler(opts.Client))
	}

	if AllowOp(opts.Mode, ops.OpOffersList) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpOffersList,
			Description: opts.Bundle.T(opts.Lang, "ops.offers_list.summary", nil),
		}, makeOffersListHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpOffersGet) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpOffersGet,
			Description: opts.Bundle.T(opts.Lang, "ops.offers_get.summary", nil),
		}, makeOffersGetHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpOffersSend) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpOffersSend,
			Description: opts.Bundle.T(opts.Lang, "ops.offers_send.summary", nil),
		}, makeOffersSendHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpOffersCounterCreate) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpOffersCounterCreate,
			Description: opts.Bundle.T(opts.Lang, "ops.offers_counter_create.summary", nil),
		}, makeOffersCounterCreateHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpOffersCounterGet) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpOffersCounterGet,
			Description: opts.Bundle.T(opts.Lang, "ops.offers_counter_get.summary", nil),
		}, makeOffersCounterGetHandler(opts.Client))
	}

	if AllowOp(opts.Mode, ops.OpOffersDraftsList) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpOffersDraftsList,
			Description: opts.Bundle.T(opts.Lang, "ops.offers_drafts_list.summary", nil),
		}, makeOfferDraftsListHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpOffersDraftsGet) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpOffersDraftsGet,
			Description: opts.Bundle.T(opts.Lang, "ops.offers_drafts_get.summary", nil),
		}, makeOfferDraftsGetHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpOffersDraftsCreate) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpOffersDraftsCreate,
			Description: opts.Bundle.T(opts.Lang, "ops.offers_drafts_create.summary", nil),
		}, makeOfferDraftsCreateHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpOffersDraftsUpdate) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpOffersDraftsUpdate,
			Description: opts.Bundle.T(opts.Lang, "ops.offers_drafts_update.summary", nil),
		}, makeOfferDraftsUpdateHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpOffersDraftsDelete) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpOffersDraftsDelete,
			Description: opts.Bundle.T(opts.Lang, "ops.offers_drafts_delete.summary", nil),
		}, makeOfferDraftsDeleteHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpOffersDraftsCreateFrom) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpOffersDraftsCreateFrom,
			Description: opts.Bundle.T(opts.Lang, "ops.offers_drafts_create_from.summary", nil),
		}, makeOfferDraftsCreateFromHandler(opts.Client))
	}

	if AllowOp(opts.Mode, ops.OpOffersDraftsAttachmentsList) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpOffersDraftsAttachmentsList,
			Description: opts.Bundle.T(opts.Lang, "ops.offers_drafts_attachments_list.summary", nil),
		}, makeOfferDraftsAttachmentsListHandler(opts.Client))
	}
	// Multipart attach is gated behind EnableAttachments — matches the
	// credit-notes/invoices pattern. Plan D wires the binary-upload path.
	if opts.EnableAttachments && AllowOp(opts.Mode, ops.OpOffersDraftsAttachmentsAttach) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpOffersDraftsAttachmentsAttach,
			Description: opts.Bundle.T(opts.Lang, "ops.offers_drafts_attachments_attach.summary", nil),
		}, makeOfferDraftsAttachmentsAttachHandler(opts.Client))
	}

	if AllowOp(opts.Mode, ops.OpOrderConfirmationsList) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpOrderConfirmationsList,
			Description: opts.Bundle.T(opts.Lang, "ops.order_confirmations_list.summary", nil),
		}, makeOrderConfirmationsListHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpOrderConfirmationsGet) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpOrderConfirmationsGet,
			Description: opts.Bundle.T(opts.Lang, "ops.order_confirmations_get.summary", nil),
		}, makeOrderConfirmationsGetHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpOrderConfirmationsCounterCreate) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpOrderConfirmationsCounterCreate,
			Description: opts.Bundle.T(opts.Lang, "ops.order_confirmations_counter_create.summary", nil),
		}, makeOrderConfirmationsCounterCreateHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpOrderConfirmationsCounterGet) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpOrderConfirmationsCounterGet,
			Description: opts.Bundle.T(opts.Lang, "ops.order_confirmations_counter_get.summary", nil),
		}, makeOrderConfirmationsCounterGetHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpOrderConfirmationsCreateInvoiceDraft) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpOrderConfirmationsCreateInvoiceDraft,
			Description: opts.Bundle.T(opts.Lang, "ops.order_confirmations_create_invoice_draft.summary", nil),
		}, makeOrderConfirmationsCreateInvoiceDraftHandler(opts.Client))
	}

	if AllowOp(opts.Mode, ops.OpOrderConfirmationsDraftsList) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpOrderConfirmationsDraftsList,
			Description: opts.Bundle.T(opts.Lang, "ops.order_confirmations_drafts_list.summary", nil),
		}, makeOrderConfirmationDraftsListHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpOrderConfirmationsDraftsGet) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpOrderConfirmationsDraftsGet,
			Description: opts.Bundle.T(opts.Lang, "ops.order_confirmations_drafts_get.summary", nil),
		}, makeOrderConfirmationDraftsGetHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpOrderConfirmationsDraftsCreate) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpOrderConfirmationsDraftsCreate,
			Description: opts.Bundle.T(opts.Lang, "ops.order_confirmations_drafts_create.summary", nil),
		}, makeOrderConfirmationDraftsCreateHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpOrderConfirmationsDraftsUpdate) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpOrderConfirmationsDraftsUpdate,
			Description: opts.Bundle.T(opts.Lang, "ops.order_confirmations_drafts_update.summary", nil),
		}, makeOrderConfirmationDraftsUpdateHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpOrderConfirmationsDraftsDelete) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpOrderConfirmationsDraftsDelete,
			Description: opts.Bundle.T(opts.Lang, "ops.order_confirmations_drafts_delete.summary", nil),
		}, makeOrderConfirmationDraftsDeleteHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpOrderConfirmationsDraftsCreateFrom) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpOrderConfirmationsDraftsCreateFrom,
			Description: opts.Bundle.T(opts.Lang, "ops.order_confirmations_drafts_create_from.summary", nil),
		}, makeOrderConfirmationDraftsCreateFromHandler(opts.Client))
	}

	if AllowOp(opts.Mode, ops.OpOrderConfirmationsDraftsAttachmentsList) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpOrderConfirmationsDraftsAttachmentsList,
			Description: opts.Bundle.T(opts.Lang, "ops.order_confirmations_drafts_attachments_list.summary", nil),
		}, makeOrderConfirmationDraftsAttachmentsListHandler(opts.Client))
	}
	// Multipart attach gated behind EnableAttachments — matches offers
	// + credit-notes pattern. Plan D wires the binary-upload path.
	if opts.EnableAttachments && AllowOp(opts.Mode, ops.OpOrderConfirmationsDraftsAttachmentsAttach) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpOrderConfirmationsDraftsAttachmentsAttach,
			Description: opts.Bundle.T(opts.Lang, "ops.order_confirmations_drafts_attachments_attach.summary", nil),
		}, makeOrderConfirmationDraftsAttachmentsAttachHandler(opts.Client))
	}

	if AllowOp(opts.Mode, ops.OpProductsList) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpProductsList,
			Description: opts.Bundle.T(opts.Lang, "ops.products_list.summary", nil),
		}, makeProductsListHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpProductsGet) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpProductsGet,
			Description: opts.Bundle.T(opts.Lang, "ops.products_get.summary", nil),
		}, makeProductsGetHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpProductsCreate) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpProductsCreate,
			Description: opts.Bundle.T(opts.Lang, "ops.products_create.summary", nil),
		}, makeProductsCreateHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpProductsUpdate) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpProductsUpdate,
			Description: opts.Bundle.T(opts.Lang, "ops.products_update.summary", nil),
		}, makeProductsUpdateHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpProductsDelete) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpProductsDelete,
			Description: opts.Bundle.T(opts.Lang, "ops.products_delete.summary", nil),
		}, makeProductsDeleteHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpProductsSalesReportCreate) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpProductsSalesReportCreate,
			Description: opts.Bundle.T(opts.Lang, "ops.products_sales_report_create.summary", nil),
		}, makeProductsSalesReportCreateHandler(opts.Client))
	}

	if AllowOp(opts.Mode, ops.OpSalesList) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpSalesList,
			Description: opts.Bundle.T(opts.Lang, "ops.sales_list.summary", nil),
		}, makeSalesListHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpSalesGet) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpSalesGet,
			Description: opts.Bundle.T(opts.Lang, "ops.sales_get.summary", nil),
		}, makeSalesGetHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpSalesCreate) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpSalesCreate,
			Description: opts.Bundle.T(opts.Lang, "ops.sales_create.summary", nil),
		}, makeSalesCreateHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpSalesDelete) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpSalesDelete,
			Description: opts.Bundle.T(opts.Lang, "ops.sales_delete.summary", nil),
		}, makeSalesDeleteHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpSalesSettle) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpSalesSettle,
			Description: opts.Bundle.T(opts.Lang, "ops.sales_settle.summary", nil),
		}, makeSalesSettleHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpSalesWriteOff) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpSalesWriteOff,
			Description: opts.Bundle.T(opts.Lang, "ops.sales_write_off.summary", nil),
		}, makeSalesWriteOffHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpSalesAttachments) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpSalesAttachments,
			Description: opts.Bundle.T(opts.Lang, "ops.sales_attachments_list.summary", nil),
		}, makeSalesAttachmentsListHandler(opts.Client))
	}
	if opts.EnableAttachments && AllowOp(opts.Mode, ops.OpSalesAttach) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpSalesAttach,
			Description: opts.Bundle.T(opts.Lang, "ops.sales_attachments_attach.summary", nil),
		}, makeSalesAttachHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpSalesPaymentsList) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpSalesPaymentsList,
			Description: opts.Bundle.T(opts.Lang, "ops.sales_payments_list.summary", nil),
		}, makeSalesPaymentsListHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpSalesPaymentsGet) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpSalesPaymentsGet,
			Description: opts.Bundle.T(opts.Lang, "ops.sales_payments_get.summary", nil),
		}, makeSalesPaymentsGetHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpSalesPaymentsCreate) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpSalesPaymentsCreate,
			Description: opts.Bundle.T(opts.Lang, "ops.sales_payments_create.summary", nil),
		}, makeSalesPaymentsCreateHandler(opts.Client))
	}

	if AllowOp(opts.Mode, ops.OpSalesDraftsList) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpSalesDraftsList,
			Description: opts.Bundle.T(opts.Lang, "ops.sales_drafts_list.summary", nil),
		}, makeSaleDraftsListHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpSalesDraftsGet) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpSalesDraftsGet,
			Description: opts.Bundle.T(opts.Lang, "ops.sales_drafts_get.summary", nil),
		}, makeSaleDraftsGetHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpSalesDraftsCreate) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpSalesDraftsCreate,
			Description: opts.Bundle.T(opts.Lang, "ops.sales_drafts_create.summary", nil),
		}, makeSaleDraftsCreateHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpSalesDraftsUpdate) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpSalesDraftsUpdate,
			Description: opts.Bundle.T(opts.Lang, "ops.sales_drafts_update.summary", nil),
		}, makeSaleDraftsUpdateHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpSalesDraftsDelete) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpSalesDraftsDelete,
			Description: opts.Bundle.T(opts.Lang, "ops.sales_drafts_delete.summary", nil),
		}, makeSaleDraftsDeleteHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpSalesDraftsCreateFrom) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpSalesDraftsCreateFrom,
			Description: opts.Bundle.T(opts.Lang, "ops.sales_drafts_create_from.summary", nil),
		}, makeSaleDraftsCreateFromHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpSalesDraftsAttachmentsList) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpSalesDraftsAttachmentsList,
			Description: opts.Bundle.T(opts.Lang, "ops.sales_drafts_attachments_list.summary", nil),
		}, makeSaleDraftsAttachmentsListHandler(opts.Client))
	}
	if opts.EnableAttachments && AllowOp(opts.Mode, ops.OpSalesDraftsAttachmentsAttach) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpSalesDraftsAttachmentsAttach,
			Description: opts.Bundle.T(opts.Lang, "ops.sales_drafts_attachments_attach.summary", nil),
		}, makeSaleDraftsAttachmentsAttachHandler(opts.Client))
	}

	if AllowOp(opts.Mode, ops.OpPurchasesList) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpPurchasesList,
			Description: opts.Bundle.T(opts.Lang, "ops.purchases_list.summary", nil),
		}, makePurchasesListHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpPurchasesGet) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpPurchasesGet,
			Description: opts.Bundle.T(opts.Lang, "ops.purchases_get.summary", nil),
		}, makePurchasesGetHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpPurchasesCreate) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpPurchasesCreate,
			Description: opts.Bundle.T(opts.Lang, "ops.purchases_create.summary", nil),
		}, makePurchasesCreateHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpPurchasesDelete) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpPurchasesDelete,
			Description: opts.Bundle.T(opts.Lang, "ops.purchases_delete.summary", nil),
		}, makePurchasesDeleteHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpPurchasesAttachments) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpPurchasesAttachments,
			Description: opts.Bundle.T(opts.Lang, "ops.purchases_attachments_list.summary", nil),
		}, makePurchasesAttachmentsListHandler(opts.Client))
	}
	if opts.EnableAttachments && AllowOp(opts.Mode, ops.OpPurchasesAttach) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpPurchasesAttach,
			Description: opts.Bundle.T(opts.Lang, "ops.purchases_attachments_attach.summary", nil),
		}, makePurchasesAttachHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpPurchasesPaymentsList) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpPurchasesPaymentsList,
			Description: opts.Bundle.T(opts.Lang, "ops.purchases_payments_list.summary", nil),
		}, makePurchasesPaymentsListHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpPurchasesPaymentsGet) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpPurchasesPaymentsGet,
			Description: opts.Bundle.T(opts.Lang, "ops.purchases_payments_get.summary", nil),
		}, makePurchasesPaymentsGetHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpPurchasesPaymentsCreate) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpPurchasesPaymentsCreate,
			Description: opts.Bundle.T(opts.Lang, "ops.purchases_payments_create.summary", nil),
		}, makePurchasesPaymentsCreateHandler(opts.Client))
	}

	if AllowOp(opts.Mode, ops.OpPurchasesDraftsList) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpPurchasesDraftsList,
			Description: opts.Bundle.T(opts.Lang, "ops.purchases_drafts_list.summary", nil),
		}, makePurchaseDraftsListHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpPurchasesDraftsGet) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpPurchasesDraftsGet,
			Description: opts.Bundle.T(opts.Lang, "ops.purchases_drafts_get.summary", nil),
		}, makePurchaseDraftsGetHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpPurchasesDraftsCreate) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpPurchasesDraftsCreate,
			Description: opts.Bundle.T(opts.Lang, "ops.purchases_drafts_create.summary", nil),
		}, makePurchaseDraftsCreateHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpPurchasesDraftsUpdate) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpPurchasesDraftsUpdate,
			Description: opts.Bundle.T(opts.Lang, "ops.purchases_drafts_update.summary", nil),
		}, makePurchaseDraftsUpdateHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpPurchasesDraftsDelete) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpPurchasesDraftsDelete,
			Description: opts.Bundle.T(opts.Lang, "ops.purchases_drafts_delete.summary", nil),
		}, makePurchaseDraftsDeleteHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpPurchasesDraftsCreateFrom) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpPurchasesDraftsCreateFrom,
			Description: opts.Bundle.T(opts.Lang, "ops.purchases_drafts_create_from.summary", nil),
		}, makePurchaseDraftsCreateFromHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpPurchasesDraftsAttachmentsList) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpPurchasesDraftsAttachmentsList,
			Description: opts.Bundle.T(opts.Lang, "ops.purchases_drafts_attachments_list.summary", nil),
		}, makePurchaseDraftsAttachmentsListHandler(opts.Client))
	}
	if opts.EnableAttachments && AllowOp(opts.Mode, ops.OpPurchasesDraftsAttachmentsAttach) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpPurchasesDraftsAttachmentsAttach,
			Description: opts.Bundle.T(opts.Lang, "ops.purchases_drafts_attachments_attach.summary", nil),
		}, makePurchaseDraftsAttachmentsAttachHandler(opts.Client))
	}

	if AllowOp(opts.Mode, ops.OpInboxList) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpInboxList,
			Description: opts.Bundle.T(opts.Lang, "ops.inbox_list.summary", nil),
		}, makeInboxListHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpInboxGet) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpInboxGet,
			Description: opts.Bundle.T(opts.Lang, "ops.inbox_get.summary", nil),
		}, makeInboxGetHandler(opts.Client))
	}
	// Multipart upload is gated behind EnableAttachments — matches the
	// invoices / purchases attach pattern. Plan D wires the multipart
	// path through once EnableAttachments is the canonical toggle.
	if opts.EnableAttachments && AllowOp(opts.Mode, ops.OpInboxSend) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpInboxSend,
			Description: opts.Bundle.T(opts.Lang, "ops.inbox_send.summary", nil),
		}, makeInboxSendHandler(opts.Client))
	}

	if AllowOp(opts.Mode, ops.OpProjectsList) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpProjectsList,
			Description: opts.Bundle.T(opts.Lang, "ops.projects_list.summary", nil),
		}, makeProjectsListHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpProjectsGet) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpProjectsGet,
			Description: opts.Bundle.T(opts.Lang, "ops.projects_get.summary", nil),
		}, makeProjectsGetHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpProjectsCreate) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpProjectsCreate,
			Description: opts.Bundle.T(opts.Lang, "ops.projects_create.summary", nil),
		}, makeProjectsCreateHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpProjectsUpdate) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpProjectsUpdate,
			Description: opts.Bundle.T(opts.Lang, "ops.projects_update.summary", nil),
		}, makeProjectsUpdateHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpProjectsDelete) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpProjectsDelete,
			Description: opts.Bundle.T(opts.Lang, "ops.projects_delete.summary", nil),
		}, makeProjectsDeleteHandler(opts.Client))
	}

	if AllowOp(opts.Mode, ops.OpUserGet) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpUserGet,
			Description: opts.Bundle.T(opts.Lang, "ops.user_get.summary", nil),
		}, makeUserGetHandler(opts.Client))
	}

	if AllowOp(opts.Mode, ops.OpAccountBalancesList) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpAccountBalancesList,
			Description: opts.Bundle.T(opts.Lang, "ops.account_balances_list.summary", nil),
		}, makeAccountBalancesListHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpAccountBalancesGet) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpAccountBalancesGet,
			Description: opts.Bundle.T(opts.Lang, "ops.account_balances_get.summary", nil),
		}, makeAccountBalancesGetHandler(opts.Client))
	}

	if AllowOp(opts.Mode, ops.OpBankBalancesList) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpBankBalancesList,
			Description: opts.Bundle.T(opts.Lang, "ops.bank_balances_list.summary", nil),
		}, makeBankBalancesListHandler(opts.Client))
	}

	if AllowOp(opts.Mode, ops.OpGroupsList) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpGroupsList,
			Description: opts.Bundle.T(opts.Lang, "ops.groups_list.summary", nil),
		}, makeGroupsListHandler(opts.Client))
	}

	if AllowOp(opts.Mode, ops.OpActivitiesList) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpActivitiesList,
			Description: opts.Bundle.T(opts.Lang, "ops.activities_list.summary", nil),
		}, makeActivitiesListHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpActivitiesGet) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpActivitiesGet,
			Description: opts.Bundle.T(opts.Lang, "ops.activities_get.summary", nil),
		}, makeActivitiesGetHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpActivitiesCreate) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpActivitiesCreate,
			Description: opts.Bundle.T(opts.Lang, "ops.activities_create.summary", nil),
		}, makeActivitiesCreateHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpActivitiesUpdate) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpActivitiesUpdate,
			Description: opts.Bundle.T(opts.Lang, "ops.activities_update.summary", nil),
		}, makeActivitiesUpdateHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpActivitiesDelete) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpActivitiesDelete,
			Description: opts.Bundle.T(opts.Lang, "ops.activities_delete.summary", nil),
		}, makeActivitiesDeleteHandler(opts.Client))
	}

	if AllowOp(opts.Mode, ops.OpTimeEntriesList) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpTimeEntriesList,
			Description: opts.Bundle.T(opts.Lang, "ops.time_entries_list.summary", nil),
		}, makeTimeEntriesListHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpTimeEntriesGet) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpTimeEntriesGet,
			Description: opts.Bundle.T(opts.Lang, "ops.time_entries_get.summary", nil),
		}, makeTimeEntriesGetHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpTimeEntriesCreate) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpTimeEntriesCreate,
			Description: opts.Bundle.T(opts.Lang, "ops.time_entries_create.summary", nil),
		}, makeTimeEntriesCreateHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpTimeEntriesUpdate) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpTimeEntriesUpdate,
			Description: opts.Bundle.T(opts.Lang, "ops.time_entries_update.summary", nil),
		}, makeTimeEntriesUpdateHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpTimeEntriesDelete) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpTimeEntriesDelete,
			Description: opts.Bundle.T(opts.Lang, "ops.time_entries_delete.summary", nil),
		}, makeTimeEntriesDeleteHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpTimeEntriesInvoiceDraftFromTimes) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpTimeEntriesInvoiceDraftFromTimes,
			Description: opts.Bundle.T(opts.Lang, "ops.time_entries_invoice_draft_create.summary", nil),
		}, makeTimeEntriesInvoiceDraftFromTimesHandler(opts.Client))
	}

	if AllowOp(opts.Mode, ops.OpTimeUsersList) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpTimeUsersList,
			Description: opts.Bundle.T(opts.Lang, "ops.time_users_list.summary", nil),
		}, makeTimeUsersListHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpTimeUsersGet) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpTimeUsersGet,
			Description: opts.Bundle.T(opts.Lang, "ops.time_users_get.summary", nil),
		}, makeTimeUsersGetHandler(opts.Client))
	}

	if opts.CapGated {
		srv.AddReceivingMiddleware(capGateMiddleware)
	}

	return srv, nil
}

// makeCompaniesListHandler wires the typed companies_list tool to the
// underlying ops.Client. The Result[CompaniesListOut] envelope is the
// Out type so success/error discriminator stays uniform across
// frontends (CLI prints it, MCP serializes it as StructuredContent).
func makeCompaniesListHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.CompaniesListIn, ops.Result[ops.CompaniesListOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.CompaniesListIn) (*mcpsdk.CallToolResult, ops.Result[ops.CompaniesListOut], error) {
		res := c.CompaniesList(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeCompaniesGetHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.CompaniesGetIn, ops.Result[ops.CompanyOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.CompaniesGetIn) (*mcpsdk.CallToolResult, ops.Result[ops.CompanyOut], error) {
		res := c.CompaniesGet(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeContactsListHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.ContactsListIn, ops.Result[ops.ContactsListOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.ContactsListIn) (*mcpsdk.CallToolResult, ops.Result[ops.ContactsListOut], error) {
		res := c.ContactsList(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeContactsGetHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.ContactsGetIn, ops.Result[ops.ContactOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.ContactsGetIn) (*mcpsdk.CallToolResult, ops.Result[ops.ContactOut], error) {
		res := c.ContactsGet(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeContactsCreateHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.ContactsCreateIn, ops.Result[ops.ContactOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.ContactsCreateIn) (*mcpsdk.CallToolResult, ops.Result[ops.ContactOut], error) {
		res := c.ContactsCreate(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeContactsUpdateHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.ContactsUpdateIn, ops.Result[ops.ContactOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.ContactsUpdateIn) (*mcpsdk.CallToolResult, ops.Result[ops.ContactOut], error) {
		res := c.ContactsUpdate(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeContactsDeleteHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.ContactsDeleteIn, ops.Result[ops.ContactsDeleteOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.ContactsDeleteIn) (*mcpsdk.CallToolResult, ops.Result[ops.ContactsDeleteOut], error) {
		res := c.ContactsDelete(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeContactsPersonsListHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.ContactsPersonsListIn, ops.Result[ops.ContactPersonsListOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.ContactsPersonsListIn) (*mcpsdk.CallToolResult, ops.Result[ops.ContactPersonsListOut], error) {
		res := c.ContactsPersonsList(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeContactsPersonsGetHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.ContactsPersonsGetIn, ops.Result[ops.ContactPersonOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.ContactsPersonsGetIn) (*mcpsdk.CallToolResult, ops.Result[ops.ContactPersonOut], error) {
		res := c.ContactsPersonsGet(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeContactsPersonsCreateHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.ContactsPersonsCreateIn, ops.Result[ops.ContactPersonOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.ContactsPersonsCreateIn) (*mcpsdk.CallToolResult, ops.Result[ops.ContactPersonOut], error) {
		res := c.ContactsPersonsCreate(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeContactsPersonsUpdateHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.ContactsPersonsUpdateIn, ops.Result[ops.ContactPersonOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.ContactsPersonsUpdateIn) (*mcpsdk.CallToolResult, ops.Result[ops.ContactPersonOut], error) {
		res := c.ContactsPersonsUpdate(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeContactsPersonsDeleteHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.ContactsPersonsDeleteIn, ops.Result[ops.ContactsPersonsDeleteOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.ContactsPersonsDeleteIn) (*mcpsdk.CallToolResult, ops.Result[ops.ContactsPersonsDeleteOut], error) {
		res := c.ContactsPersonsDelete(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeAccountsListHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.AccountsListIn, ops.Result[ops.AccountsListOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.AccountsListIn) (*mcpsdk.CallToolResult, ops.Result[ops.AccountsListOut], error) {
		res := c.AccountsList(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeAccountsGetHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.AccountsGetIn, ops.Result[ops.AccountOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.AccountsGetIn) (*mcpsdk.CallToolResult, ops.Result[ops.AccountOut], error) {
		res := c.AccountsGet(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeBankAccountsListHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.BankAccountsListIn, ops.Result[ops.BankAccountsListOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.BankAccountsListIn) (*mcpsdk.CallToolResult, ops.Result[ops.BankAccountsListOut], error) {
		res := c.BankAccountsList(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeBankAccountsGetHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.BankAccountsGetIn, ops.Result[ops.BankAccountOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.BankAccountsGetIn) (*mcpsdk.CallToolResult, ops.Result[ops.BankAccountOut], error) {
		res := c.BankAccountsGet(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeBankAccountsCreateHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.BankAccountsCreateIn, ops.Result[ops.BankAccountOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.BankAccountsCreateIn) (*mcpsdk.CallToolResult, ops.Result[ops.BankAccountOut], error) {
		res := c.BankAccountsCreate(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeJournalEntriesListHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.JournalEntriesListIn, ops.Result[ops.JournalEntriesListOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.JournalEntriesListIn) (*mcpsdk.CallToolResult, ops.Result[ops.JournalEntriesListOut], error) {
		res := c.JournalEntriesList(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeJournalEntriesGetHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.JournalEntriesGetIn, ops.Result[ops.JournalEntryOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.JournalEntriesGetIn) (*mcpsdk.CallToolResult, ops.Result[ops.JournalEntryOut], error) {
		res := c.JournalEntriesGet(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeJournalEntriesCreateHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.JournalEntriesCreateIn, ops.Result[ops.JournalEntryOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.JournalEntriesCreateIn) (*mcpsdk.CallToolResult, ops.Result[ops.JournalEntryOut], error) {
		res := c.JournalEntriesCreate(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeJournalEntriesAttachmentsListHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.JournalEntriesAttachmentsListIn, ops.Result[ops.AttachmentsListOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.JournalEntriesAttachmentsListIn) (*mcpsdk.CallToolResult, ops.Result[ops.AttachmentsListOut], error) {
		res := c.JournalEntriesAttachmentsList(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeJournalEntriesAttachmentsAttachHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.JournalEntriesAttachmentsAttachIn, ops.Result[ops.JournalEntriesAttachmentsAttachOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.JournalEntriesAttachmentsAttachIn) (*mcpsdk.CallToolResult, ops.Result[ops.JournalEntriesAttachmentsAttachOut], error) {
		res := c.JournalEntriesAttachmentsAttach(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeTransactionsListHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.TransactionsListIn, ops.Result[ops.TransactionsListOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.TransactionsListIn) (*mcpsdk.CallToolResult, ops.Result[ops.TransactionsListOut], error) {
		res := c.TransactionsList(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeTransactionsGetHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.TransactionsGetIn, ops.Result[ops.TransactionOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.TransactionsGetIn) (*mcpsdk.CallToolResult, ops.Result[ops.TransactionOut], error) {
		res := c.TransactionsGet(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeInvoicesListHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.InvoicesListIn, ops.Result[ops.InvoicesListOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.InvoicesListIn) (*mcpsdk.CallToolResult, ops.Result[ops.InvoicesListOut], error) {
		res := c.InvoicesList(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeInvoicesGetHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.InvoicesGetIn, ops.Result[ops.InvoiceOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.InvoicesGetIn) (*mcpsdk.CallToolResult, ops.Result[ops.InvoiceOut], error) {
		res := c.InvoicesGet(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeInvoicesSendHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.InvoicesSendIn, ops.Result[ops.InvoicesSendOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.InvoicesSendIn) (*mcpsdk.CallToolResult, ops.Result[ops.InvoicesSendOut], error) {
		res := c.InvoicesSend(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeInvoicesCounterCreateHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.InvoicesCounterCreateIn, ops.Result[ops.InvoicesCounterCreateOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.InvoicesCounterCreateIn) (*mcpsdk.CallToolResult, ops.Result[ops.InvoicesCounterCreateOut], error) {
		res := c.InvoicesCounterCreate(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeInvoiceDraftsListHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.InvoiceDraftsListIn, ops.Result[ops.InvoiceDraftsListOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.InvoiceDraftsListIn) (*mcpsdk.CallToolResult, ops.Result[ops.InvoiceDraftsListOut], error) {
		res := c.InvoiceDraftsList(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeInvoiceDraftsGetHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.InvoiceDraftsGetIn, ops.Result[ops.InvoiceDraftOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.InvoiceDraftsGetIn) (*mcpsdk.CallToolResult, ops.Result[ops.InvoiceDraftOut], error) {
		res := c.InvoiceDraftsGet(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeInvoiceDraftsCreateHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.InvoiceDraftsCreateIn, ops.Result[ops.InvoiceDraftOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.InvoiceDraftsCreateIn) (*mcpsdk.CallToolResult, ops.Result[ops.InvoiceDraftOut], error) {
		res := c.InvoiceDraftsCreate(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeInvoiceDraftsUpdateHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.InvoiceDraftsUpdateIn, ops.Result[ops.InvoiceDraftOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.InvoiceDraftsUpdateIn) (*mcpsdk.CallToolResult, ops.Result[ops.InvoiceDraftOut], error) {
		res := c.InvoiceDraftsUpdate(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeInvoiceDraftsDeleteHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.InvoiceDraftsDeleteIn, ops.Result[ops.InvoiceDraftsDeleteOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.InvoiceDraftsDeleteIn) (*mcpsdk.CallToolResult, ops.Result[ops.InvoiceDraftsDeleteOut], error) {
		res := c.InvoiceDraftsDelete(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeInvoiceDraftsCreateFromHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.InvoiceDraftsCreateFromIn, ops.Result[ops.InvoiceDraftsCreateFromOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.InvoiceDraftsCreateFromIn) (*mcpsdk.CallToolResult, ops.Result[ops.InvoiceDraftsCreateFromOut], error) {
		res := c.InvoiceDraftsCreateFrom(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeInvoicesAttachmentsListHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.InvoicesAttachmentsListIn, ops.Result[ops.AttachmentsListOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.InvoicesAttachmentsListIn) (*mcpsdk.CallToolResult, ops.Result[ops.AttachmentsListOut], error) {
		res := c.InvoicesAttachmentsList(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeInvoicesAttachmentsAttachHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.InvoicesAttachmentsAttachIn, ops.Result[ops.InvoicesAttachmentsAttachOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.InvoicesAttachmentsAttachIn) (*mcpsdk.CallToolResult, ops.Result[ops.InvoicesAttachmentsAttachOut], error) {
		res := c.InvoicesAttachmentsAttach(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeInvoiceDraftsAttachmentsListHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.InvoiceDraftsAttachmentsListIn, ops.Result[ops.AttachmentsListOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.InvoiceDraftsAttachmentsListIn) (*mcpsdk.CallToolResult, ops.Result[ops.AttachmentsListOut], error) {
		res := c.InvoiceDraftsAttachmentsList(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeInvoiceDraftsAttachmentsAttachHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.InvoiceDraftsAttachmentsAttachIn, ops.Result[ops.InvoiceDraftsAttachmentsAttachOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.InvoiceDraftsAttachmentsAttachIn) (*mcpsdk.CallToolResult, ops.Result[ops.InvoiceDraftsAttachmentsAttachOut], error) {
		res := c.InvoiceDraftsAttachmentsAttach(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeCreditNotesListHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.CreditNotesListIn, ops.Result[ops.CreditNotesListOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.CreditNotesListIn) (*mcpsdk.CallToolResult, ops.Result[ops.CreditNotesListOut], error) {
		res := c.CreditNotesList(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeCreditNotesGetHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.CreditNotesGetIn, ops.Result[ops.CreditNoteOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.CreditNotesGetIn) (*mcpsdk.CallToolResult, ops.Result[ops.CreditNoteOut], error) {
		res := c.CreditNotesGet(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeCreditNotesSendHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.CreditNotesSendIn, ops.Result[ops.CreditNotesSendOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.CreditNotesSendIn) (*mcpsdk.CallToolResult, ops.Result[ops.CreditNotesSendOut], error) {
		res := c.CreditNotesSend(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeCreditNotesCounterCreateHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.CreditNotesCounterCreateIn, ops.Result[ops.CreditNotesCounterCreateOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.CreditNotesCounterCreateIn) (*mcpsdk.CallToolResult, ops.Result[ops.CreditNotesCounterCreateOut], error) {
		res := c.CreditNotesCounterCreate(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeCreditNotesFullCreateHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.CreditNotesFullCreateIn, ops.Result[ops.CreditNotesFullCreateOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.CreditNotesFullCreateIn) (*mcpsdk.CallToolResult, ops.Result[ops.CreditNotesFullCreateOut], error) {
		res := c.CreditNotesFullCreate(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeCreditNotesPartialCreateHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.CreditNotesPartialCreateIn, ops.Result[ops.CreditNotesPartialCreateOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.CreditNotesPartialCreateIn) (*mcpsdk.CallToolResult, ops.Result[ops.CreditNotesPartialCreateOut], error) {
		res := c.CreditNotesPartialCreate(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeCreditNoteDraftsListHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.CreditNoteDraftsListIn, ops.Result[ops.CreditNoteDraftsListOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.CreditNoteDraftsListIn) (*mcpsdk.CallToolResult, ops.Result[ops.CreditNoteDraftsListOut], error) {
		res := c.CreditNoteDraftsList(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeCreditNoteDraftsGetHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.CreditNoteDraftsGetIn, ops.Result[ops.InvoiceDraftOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.CreditNoteDraftsGetIn) (*mcpsdk.CallToolResult, ops.Result[ops.InvoiceDraftOut], error) {
		res := c.CreditNoteDraftsGet(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeCreditNoteDraftsCreateHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.CreditNoteDraftsCreateIn, ops.Result[ops.InvoiceDraftOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.CreditNoteDraftsCreateIn) (*mcpsdk.CallToolResult, ops.Result[ops.InvoiceDraftOut], error) {
		res := c.CreditNoteDraftsCreate(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeCreditNoteDraftsUpdateHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.CreditNoteDraftsUpdateIn, ops.Result[ops.InvoiceDraftOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.CreditNoteDraftsUpdateIn) (*mcpsdk.CallToolResult, ops.Result[ops.InvoiceDraftOut], error) {
		res := c.CreditNoteDraftsUpdate(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeCreditNoteDraftsDeleteHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.CreditNoteDraftsDeleteIn, ops.Result[ops.CreditNoteDraftsDeleteOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.CreditNoteDraftsDeleteIn) (*mcpsdk.CallToolResult, ops.Result[ops.CreditNoteDraftsDeleteOut], error) {
		res := c.CreditNoteDraftsDelete(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeCreditNoteDraftsCreateFromHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.CreditNoteDraftsCreateFromIn, ops.Result[ops.CreditNoteDraftsCreateFromOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.CreditNoteDraftsCreateFromIn) (*mcpsdk.CallToolResult, ops.Result[ops.CreditNoteDraftsCreateFromOut], error) {
		res := c.CreditNoteDraftsCreateFrom(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeCreditNoteDraftsAttachmentsListHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.CreditNoteDraftsAttachmentsListIn, ops.Result[ops.AttachmentsListOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.CreditNoteDraftsAttachmentsListIn) (*mcpsdk.CallToolResult, ops.Result[ops.AttachmentsListOut], error) {
		res := c.CreditNoteDraftsAttachmentsList(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeCreditNoteDraftsAttachmentsAttachHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.CreditNoteDraftsAttachmentsAttachIn, ops.Result[ops.CreditNoteDraftsAttachmentsAttachOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.CreditNoteDraftsAttachmentsAttachIn) (*mcpsdk.CallToolResult, ops.Result[ops.CreditNoteDraftsAttachmentsAttachOut], error) {
		res := c.CreditNoteDraftsAttachmentsAttach(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeOffersListHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.OffersListIn, ops.Result[ops.OffersListOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.OffersListIn) (*mcpsdk.CallToolResult, ops.Result[ops.OffersListOut], error) {
		res := c.OffersList(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeOffersGetHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.OffersGetIn, ops.Result[ops.OfferOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.OffersGetIn) (*mcpsdk.CallToolResult, ops.Result[ops.OfferOut], error) {
		res := c.OffersGet(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeOffersSendHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.OffersSendIn, ops.Result[ops.OffersSendOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.OffersSendIn) (*mcpsdk.CallToolResult, ops.Result[ops.OffersSendOut], error) {
		res := c.OffersSend(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeOffersCounterCreateHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.OffersCounterCreateIn, ops.Result[ops.OffersCounterCreateOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.OffersCounterCreateIn) (*mcpsdk.CallToolResult, ops.Result[ops.OffersCounterCreateOut], error) {
		res := c.OffersCounterCreate(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeOfferDraftsListHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.OfferDraftsListIn, ops.Result[ops.OfferDraftsListOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.OfferDraftsListIn) (*mcpsdk.CallToolResult, ops.Result[ops.OfferDraftsListOut], error) {
		res := c.OfferDraftsList(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeOfferDraftsGetHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.OfferDraftsGetIn, ops.Result[ops.InvoiceDraftOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.OfferDraftsGetIn) (*mcpsdk.CallToolResult, ops.Result[ops.InvoiceDraftOut], error) {
		res := c.OfferDraftsGet(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeOfferDraftsCreateHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.OfferDraftsCreateIn, ops.Result[ops.InvoiceDraftOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.OfferDraftsCreateIn) (*mcpsdk.CallToolResult, ops.Result[ops.InvoiceDraftOut], error) {
		res := c.OfferDraftsCreate(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeOfferDraftsUpdateHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.OfferDraftsUpdateIn, ops.Result[ops.InvoiceDraftOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.OfferDraftsUpdateIn) (*mcpsdk.CallToolResult, ops.Result[ops.InvoiceDraftOut], error) {
		res := c.OfferDraftsUpdate(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeOfferDraftsDeleteHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.OfferDraftsDeleteIn, ops.Result[ops.OfferDraftsDeleteOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.OfferDraftsDeleteIn) (*mcpsdk.CallToolResult, ops.Result[ops.OfferDraftsDeleteOut], error) {
		res := c.OfferDraftsDelete(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeOfferDraftsCreateFromHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.OfferDraftsCreateFromIn, ops.Result[ops.OfferDraftsCreateFromOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.OfferDraftsCreateFromIn) (*mcpsdk.CallToolResult, ops.Result[ops.OfferDraftsCreateFromOut], error) {
		res := c.OfferDraftsCreateFrom(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeOfferDraftsAttachmentsListHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.OfferDraftsAttachmentsListIn, ops.Result[ops.AttachmentsListOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.OfferDraftsAttachmentsListIn) (*mcpsdk.CallToolResult, ops.Result[ops.AttachmentsListOut], error) {
		res := c.OfferDraftsAttachmentsList(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeOfferDraftsAttachmentsAttachHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.OfferDraftsAttachmentsAttachIn, ops.Result[ops.OfferDraftsAttachmentsAttachOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.OfferDraftsAttachmentsAttachIn) (*mcpsdk.CallToolResult, ops.Result[ops.OfferDraftsAttachmentsAttachOut], error) {
		res := c.OfferDraftsAttachmentsAttach(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeOrderConfirmationsListHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.OrderConfirmationsListIn, ops.Result[ops.OrderConfirmationsListOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.OrderConfirmationsListIn) (*mcpsdk.CallToolResult, ops.Result[ops.OrderConfirmationsListOut], error) {
		res := c.OrderConfirmationsList(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeOrderConfirmationsGetHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.OrderConfirmationsGetIn, ops.Result[ops.OrderConfirmationOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.OrderConfirmationsGetIn) (*mcpsdk.CallToolResult, ops.Result[ops.OrderConfirmationOut], error) {
		res := c.OrderConfirmationsGet(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeOrderConfirmationsCounterCreateHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.OrderConfirmationsCounterCreateIn, ops.Result[ops.OrderConfirmationsCounterCreateOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.OrderConfirmationsCounterCreateIn) (*mcpsdk.CallToolResult, ops.Result[ops.OrderConfirmationsCounterCreateOut], error) {
		res := c.OrderConfirmationsCounterCreate(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeOrderConfirmationsCreateInvoiceDraftHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.OrderConfirmationsCreateInvoiceDraftIn, ops.Result[ops.OrderConfirmationsCreateInvoiceDraftOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.OrderConfirmationsCreateInvoiceDraftIn) (*mcpsdk.CallToolResult, ops.Result[ops.OrderConfirmationsCreateInvoiceDraftOut], error) {
		res := c.OrderConfirmationsCreateInvoiceDraft(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeOrderConfirmationDraftsListHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.OrderConfirmationDraftsListIn, ops.Result[ops.OrderConfirmationDraftsListOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.OrderConfirmationDraftsListIn) (*mcpsdk.CallToolResult, ops.Result[ops.OrderConfirmationDraftsListOut], error) {
		res := c.OrderConfirmationDraftsList(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeOrderConfirmationDraftsGetHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.OrderConfirmationDraftsGetIn, ops.Result[ops.InvoiceDraftOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.OrderConfirmationDraftsGetIn) (*mcpsdk.CallToolResult, ops.Result[ops.InvoiceDraftOut], error) {
		res := c.OrderConfirmationDraftsGet(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeOrderConfirmationDraftsCreateHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.OrderConfirmationDraftsCreateIn, ops.Result[ops.InvoiceDraftOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.OrderConfirmationDraftsCreateIn) (*mcpsdk.CallToolResult, ops.Result[ops.InvoiceDraftOut], error) {
		res := c.OrderConfirmationDraftsCreate(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeOrderConfirmationDraftsUpdateHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.OrderConfirmationDraftsUpdateIn, ops.Result[ops.InvoiceDraftOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.OrderConfirmationDraftsUpdateIn) (*mcpsdk.CallToolResult, ops.Result[ops.InvoiceDraftOut], error) {
		res := c.OrderConfirmationDraftsUpdate(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeOrderConfirmationDraftsDeleteHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.OrderConfirmationDraftsDeleteIn, ops.Result[ops.OrderConfirmationDraftsDeleteOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.OrderConfirmationDraftsDeleteIn) (*mcpsdk.CallToolResult, ops.Result[ops.OrderConfirmationDraftsDeleteOut], error) {
		res := c.OrderConfirmationDraftsDelete(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeOrderConfirmationDraftsCreateFromHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.OrderConfirmationDraftsCreateFromIn, ops.Result[ops.OrderConfirmationDraftsCreateFromOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.OrderConfirmationDraftsCreateFromIn) (*mcpsdk.CallToolResult, ops.Result[ops.OrderConfirmationDraftsCreateFromOut], error) {
		res := c.OrderConfirmationDraftsCreateFrom(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeOrderConfirmationDraftsAttachmentsListHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.OrderConfirmationDraftsAttachmentsListIn, ops.Result[ops.AttachmentsListOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.OrderConfirmationDraftsAttachmentsListIn) (*mcpsdk.CallToolResult, ops.Result[ops.AttachmentsListOut], error) {
		res := c.OrderConfirmationDraftsAttachmentsList(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeOrderConfirmationDraftsAttachmentsAttachHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.OrderConfirmationDraftsAttachmentsAttachIn, ops.Result[ops.OrderConfirmationDraftsAttachmentsAttachOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.OrderConfirmationDraftsAttachmentsAttachIn) (*mcpsdk.CallToolResult, ops.Result[ops.OrderConfirmationDraftsAttachmentsAttachOut], error) {
		res := c.OrderConfirmationDraftsAttachmentsAttach(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeProductsListHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.ProductsListIn, ops.Result[ops.ProductsListOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.ProductsListIn) (*mcpsdk.CallToolResult, ops.Result[ops.ProductsListOut], error) {
		res := c.ProductsList(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeProductsGetHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.ProductsGetIn, ops.Result[ops.ProductOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.ProductsGetIn) (*mcpsdk.CallToolResult, ops.Result[ops.ProductOut], error) {
		res := c.ProductsGet(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeProductsCreateHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.ProductsCreateIn, ops.Result[ops.ProductOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.ProductsCreateIn) (*mcpsdk.CallToolResult, ops.Result[ops.ProductOut], error) {
		res := c.ProductsCreate(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeProductsUpdateHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.ProductsUpdateIn, ops.Result[ops.ProductOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.ProductsUpdateIn) (*mcpsdk.CallToolResult, ops.Result[ops.ProductOut], error) {
		res := c.ProductsUpdate(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeProductsDeleteHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.ProductsDeleteIn, ops.Result[ops.ProductsDeleteOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.ProductsDeleteIn) (*mcpsdk.CallToolResult, ops.Result[ops.ProductsDeleteOut], error) {
		res := c.ProductsDelete(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeProductsSalesReportCreateHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.ProductsSalesReportCreateIn, ops.Result[ops.ProductsSalesReportCreateOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.ProductsSalesReportCreateIn) (*mcpsdk.CallToolResult, ops.Result[ops.ProductsSalesReportCreateOut], error) {
		res := c.ProductsSalesReportCreate(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeSalesListHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.SalesListIn, ops.Result[ops.SalesListOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.SalesListIn) (*mcpsdk.CallToolResult, ops.Result[ops.SalesListOut], error) {
		res := c.SalesList(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeSalesGetHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.SalesGetIn, ops.Result[ops.SaleOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.SalesGetIn) (*mcpsdk.CallToolResult, ops.Result[ops.SaleOut], error) {
		res := c.SalesGet(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeSalesCreateHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.SalesCreateIn, ops.Result[ops.SalesCreateOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.SalesCreateIn) (*mcpsdk.CallToolResult, ops.Result[ops.SalesCreateOut], error) {
		res := c.SalesCreate(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeSalesDeleteHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.SalesDeleteIn, ops.Result[ops.SalesDeleteOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.SalesDeleteIn) (*mcpsdk.CallToolResult, ops.Result[ops.SalesDeleteOut], error) {
		res := c.SalesDelete(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeSalesSettleHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.SalesSettleIn, ops.Result[ops.SaleOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.SalesSettleIn) (*mcpsdk.CallToolResult, ops.Result[ops.SaleOut], error) {
		res := c.SalesSettle(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeSalesWriteOffHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.SalesWriteOffIn, ops.Result[ops.SaleOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.SalesWriteOffIn) (*mcpsdk.CallToolResult, ops.Result[ops.SaleOut], error) {
		res := c.SalesWriteOff(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeSalesAttachmentsListHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.SalesAttachmentsListIn, ops.Result[ops.AttachmentsListOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.SalesAttachmentsListIn) (*mcpsdk.CallToolResult, ops.Result[ops.AttachmentsListOut], error) {
		res := c.SalesAttachmentsList(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeSalesAttachHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.SalesAttachIn, ops.Result[ops.SalesAttachOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.SalesAttachIn) (*mcpsdk.CallToolResult, ops.Result[ops.SalesAttachOut], error) {
		res := c.SalesAttach(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeSalesPaymentsListHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.SalesPaymentsListIn, ops.Result[ops.PaymentsListOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.SalesPaymentsListIn) (*mcpsdk.CallToolResult, ops.Result[ops.PaymentsListOut], error) {
		res := c.SalesPaymentsList(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeSalesPaymentsGetHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.SalesPaymentsGetIn, ops.Result[ops.PaymentOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.SalesPaymentsGetIn) (*mcpsdk.CallToolResult, ops.Result[ops.PaymentOut], error) {
		res := c.SalesPaymentsGet(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeSalesPaymentsCreateHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.SalesPaymentsCreateIn, ops.Result[ops.SalesPaymentsCreateOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.SalesPaymentsCreateIn) (*mcpsdk.CallToolResult, ops.Result[ops.SalesPaymentsCreateOut], error) {
		res := c.SalesPaymentsCreate(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makePurchasesListHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.PurchasesListIn, ops.Result[ops.PurchasesListOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.PurchasesListIn) (*mcpsdk.CallToolResult, ops.Result[ops.PurchasesListOut], error) {
		res := c.PurchasesList(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makePurchasesGetHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.PurchasesGetIn, ops.Result[ops.PurchaseOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.PurchasesGetIn) (*mcpsdk.CallToolResult, ops.Result[ops.PurchaseOut], error) {
		res := c.PurchasesGet(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makePurchasesCreateHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.PurchasesCreateIn, ops.Result[ops.PurchasesCreateOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.PurchasesCreateIn) (*mcpsdk.CallToolResult, ops.Result[ops.PurchasesCreateOut], error) {
		res := c.PurchasesCreate(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makePurchasesDeleteHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.PurchasesDeleteIn, ops.Result[ops.PurchasesDeleteOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.PurchasesDeleteIn) (*mcpsdk.CallToolResult, ops.Result[ops.PurchasesDeleteOut], error) {
		res := c.PurchasesDelete(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makePurchasesAttachmentsListHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.PurchasesAttachmentsListIn, ops.Result[ops.AttachmentsListOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.PurchasesAttachmentsListIn) (*mcpsdk.CallToolResult, ops.Result[ops.AttachmentsListOut], error) {
		res := c.PurchasesAttachmentsList(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makePurchasesAttachHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.PurchasesAttachIn, ops.Result[ops.PurchasesAttachOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.PurchasesAttachIn) (*mcpsdk.CallToolResult, ops.Result[ops.PurchasesAttachOut], error) {
		res := c.PurchasesAttach(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makePurchasesPaymentsListHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.PurchasesPaymentsListIn, ops.Result[ops.PaymentsListOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.PurchasesPaymentsListIn) (*mcpsdk.CallToolResult, ops.Result[ops.PaymentsListOut], error) {
		res := c.PurchasesPaymentsList(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makePurchasesPaymentsGetHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.PurchasesPaymentsGetIn, ops.Result[ops.PaymentOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.PurchasesPaymentsGetIn) (*mcpsdk.CallToolResult, ops.Result[ops.PaymentOut], error) {
		res := c.PurchasesPaymentsGet(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makePurchasesPaymentsCreateHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.PurchasesPaymentsCreateIn, ops.Result[ops.PurchasesPaymentsCreateOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.PurchasesPaymentsCreateIn) (*mcpsdk.CallToolResult, ops.Result[ops.PurchasesPaymentsCreateOut], error) {
		res := c.PurchasesPaymentsCreate(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeInboxListHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.InboxListIn, ops.Result[ops.InboxListOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.InboxListIn) (*mcpsdk.CallToolResult, ops.Result[ops.InboxListOut], error) {
		res := c.InboxList(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeInboxGetHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.InboxGetIn, ops.Result[ops.InboxDocumentOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.InboxGetIn) (*mcpsdk.CallToolResult, ops.Result[ops.InboxDocumentOut], error) {
		res := c.InboxGet(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeInboxSendHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.InboxSendIn, ops.Result[ops.InboxSendOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.InboxSendIn) (*mcpsdk.CallToolResult, ops.Result[ops.InboxSendOut], error) {
		res := c.InboxSend(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeProjectsListHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.ProjectsListIn, ops.Result[ops.ProjectsListOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.ProjectsListIn) (*mcpsdk.CallToolResult, ops.Result[ops.ProjectsListOut], error) {
		res := c.ProjectsList(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeProjectsGetHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.ProjectsGetIn, ops.Result[ops.ProjectOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.ProjectsGetIn) (*mcpsdk.CallToolResult, ops.Result[ops.ProjectOut], error) {
		res := c.ProjectsGet(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeProjectsCreateHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.ProjectsCreateIn, ops.Result[ops.ProjectOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.ProjectsCreateIn) (*mcpsdk.CallToolResult, ops.Result[ops.ProjectOut], error) {
		res := c.ProjectsCreate(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeProjectsUpdateHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.ProjectsUpdateIn, ops.Result[ops.ProjectOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.ProjectsUpdateIn) (*mcpsdk.CallToolResult, ops.Result[ops.ProjectOut], error) {
		res := c.ProjectsUpdate(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeProjectsDeleteHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.ProjectsDeleteIn, ops.Result[ops.ProjectsDeleteOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.ProjectsDeleteIn) (*mcpsdk.CallToolResult, ops.Result[ops.ProjectsDeleteOut], error) {
		res := c.ProjectsDelete(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeUserGetHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.UserGetIn, ops.Result[ops.UserOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.UserGetIn) (*mcpsdk.CallToolResult, ops.Result[ops.UserOut], error) {
		res := c.UserGet(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeAccountBalancesListHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.AccountBalancesListIn, ops.Result[ops.AccountBalancesListOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.AccountBalancesListIn) (*mcpsdk.CallToolResult, ops.Result[ops.AccountBalancesListOut], error) {
		res := c.AccountBalancesList(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeAccountBalancesGetHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.AccountBalancesGetIn, ops.Result[ops.AccountBalanceOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.AccountBalancesGetIn) (*mcpsdk.CallToolResult, ops.Result[ops.AccountBalanceOut], error) {
		res := c.AccountBalancesGet(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeBankBalancesListHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.BankBalancesListIn, ops.Result[ops.BankBalancesListOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.BankBalancesListIn) (*mcpsdk.CallToolResult, ops.Result[ops.BankBalancesListOut], error) {
		res := c.BankBalancesList(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeGroupsListHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.GroupsListIn, ops.Result[ops.GroupsListOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.GroupsListIn) (*mcpsdk.CallToolResult, ops.Result[ops.GroupsListOut], error) {
		res := c.GroupsList(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeActivitiesListHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.ActivitiesListIn, ops.Result[ops.ActivitiesListOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.ActivitiesListIn) (*mcpsdk.CallToolResult, ops.Result[ops.ActivitiesListOut], error) {
		res := c.ActivitiesList(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeActivitiesGetHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.ActivitiesGetIn, ops.Result[ops.ActivityOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.ActivitiesGetIn) (*mcpsdk.CallToolResult, ops.Result[ops.ActivityOut], error) {
		res := c.ActivitiesGet(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeActivitiesCreateHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.ActivitiesCreateIn, ops.Result[ops.ActivityOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.ActivitiesCreateIn) (*mcpsdk.CallToolResult, ops.Result[ops.ActivityOut], error) {
		res := c.ActivitiesCreate(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeActivitiesUpdateHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.ActivitiesUpdateIn, ops.Result[ops.ActivityOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.ActivitiesUpdateIn) (*mcpsdk.CallToolResult, ops.Result[ops.ActivityOut], error) {
		res := c.ActivitiesUpdate(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeActivitiesDeleteHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.ActivitiesDeleteIn, ops.Result[ops.ActivitiesDeleteOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.ActivitiesDeleteIn) (*mcpsdk.CallToolResult, ops.Result[ops.ActivitiesDeleteOut], error) {
		res := c.ActivitiesDelete(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeTimeEntriesListHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.TimeEntriesListIn, ops.Result[ops.TimeEntriesListOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.TimeEntriesListIn) (*mcpsdk.CallToolResult, ops.Result[ops.TimeEntriesListOut], error) {
		res := c.TimeEntriesList(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeTimeEntriesGetHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.TimeEntriesGetIn, ops.Result[ops.TimeEntryOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.TimeEntriesGetIn) (*mcpsdk.CallToolResult, ops.Result[ops.TimeEntryOut], error) {
		res := c.TimeEntriesGet(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeTimeEntriesCreateHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.TimeEntriesCreateIn, ops.Result[ops.TimeEntryOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.TimeEntriesCreateIn) (*mcpsdk.CallToolResult, ops.Result[ops.TimeEntryOut], error) {
		res := c.TimeEntriesCreate(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeTimeEntriesUpdateHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.TimeEntriesUpdateIn, ops.Result[ops.TimeEntryOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.TimeEntriesUpdateIn) (*mcpsdk.CallToolResult, ops.Result[ops.TimeEntryOut], error) {
		res := c.TimeEntriesUpdate(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeTimeEntriesDeleteHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.TimeEntriesDeleteIn, ops.Result[ops.TimeEntriesDeleteOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.TimeEntriesDeleteIn) (*mcpsdk.CallToolResult, ops.Result[ops.TimeEntriesDeleteOut], error) {
		res := c.TimeEntriesDelete(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeTimeEntriesInvoiceDraftFromTimesHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.TimeEntriesInvoiceDraftFromTimesIn, ops.Result[ops.TimeEntriesInvoiceDraftFromTimesOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.TimeEntriesInvoiceDraftFromTimesIn) (*mcpsdk.CallToolResult, ops.Result[ops.TimeEntriesInvoiceDraftFromTimesOut], error) {
		res := c.TimeEntriesInvoiceDraftFromTimes(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeTimeUsersListHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.TimeUsersListIn, ops.Result[ops.TimeUsersListOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.TimeUsersListIn) (*mcpsdk.CallToolResult, ops.Result[ops.TimeUsersListOut], error) {
		res := c.TimeUsersList(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeTimeUsersGetHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.TimeUsersGetIn, ops.Result[ops.TimeUserOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.TimeUsersGetIn) (*mcpsdk.CallToolResult, ops.Result[ops.TimeUserOut], error) {
		res := c.TimeUsersGet(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

// === Plan D / 21-op tail handlers ===

func makeInvoicesCreateHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.InvoicesCreateIn, ops.Result[ops.InvoicesCreateOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.InvoicesCreateIn) (*mcpsdk.CallToolResult, ops.Result[ops.InvoicesCreateOut], error) {
		res := c.InvoicesCreate(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeInvoicesUpdateHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.InvoicesUpdateIn, ops.Result[ops.InvoicesUpdateOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.InvoicesUpdateIn) (*mcpsdk.CallToolResult, ops.Result[ops.InvoicesUpdateOut], error) {
		res := c.InvoicesUpdate(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeInvoicesCounterGetHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.CounterGetIn, ops.Result[ops.CounterGetOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.CounterGetIn) (*mcpsdk.CallToolResult, ops.Result[ops.CounterGetOut], error) {
		res := c.InvoicesCounterGet(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeCreditNotesCounterGetHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.CounterGetIn, ops.Result[ops.CounterGetOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.CounterGetIn) (*mcpsdk.CallToolResult, ops.Result[ops.CounterGetOut], error) {
		res := c.CreditNotesCounterGet(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeOffersCounterGetHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.CounterGetIn, ops.Result[ops.CounterGetOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.CounterGetIn) (*mcpsdk.CallToolResult, ops.Result[ops.CounterGetOut], error) {
		res := c.OffersCounterGet(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeOrderConfirmationsCounterGetHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.CounterGetIn, ops.Result[ops.CounterGetOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.CounterGetIn) (*mcpsdk.CallToolResult, ops.Result[ops.CounterGetOut], error) {
		res := c.OrderConfirmationsCounterGet(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeTransactionsDeleteHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.TransactionsDeleteIn, ops.Result[ops.TransactionsDeleteOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.TransactionsDeleteIn) (*mcpsdk.CallToolResult, ops.Result[ops.TransactionsDeleteOut], error) {
		res := c.TransactionsDelete(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeContactsAttachmentsAttachHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.ContactsAttachmentsAttachIn, ops.Result[ops.ContactsAttachmentsAttachOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.ContactsAttachmentsAttachIn) (*mcpsdk.CallToolResult, ops.Result[ops.ContactsAttachmentsAttachOut], error) {
		res := c.ContactsAttachmentsAttach(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeSaleDraftsListHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.SaleDraftsListIn, ops.Result[ops.SaleDraftsListOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.SaleDraftsListIn) (*mcpsdk.CallToolResult, ops.Result[ops.SaleDraftsListOut], error) {
		res := c.SaleDraftsList(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeSaleDraftsGetHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.SaleDraftsGetIn, ops.Result[ops.SaleDraftOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.SaleDraftsGetIn) (*mcpsdk.CallToolResult, ops.Result[ops.SaleDraftOut], error) {
		res := c.SaleDraftsGet(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeSaleDraftsCreateHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.SaleDraftsCreateIn, ops.Result[ops.SaleDraftOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.SaleDraftsCreateIn) (*mcpsdk.CallToolResult, ops.Result[ops.SaleDraftOut], error) {
		res := c.SaleDraftsCreate(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeSaleDraftsUpdateHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.SaleDraftsUpdateIn, ops.Result[ops.SaleDraftOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.SaleDraftsUpdateIn) (*mcpsdk.CallToolResult, ops.Result[ops.SaleDraftOut], error) {
		res := c.SaleDraftsUpdate(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeSaleDraftsDeleteHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.SaleDraftsDeleteIn, ops.Result[ops.SaleDraftsDeleteOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.SaleDraftsDeleteIn) (*mcpsdk.CallToolResult, ops.Result[ops.SaleDraftsDeleteOut], error) {
		res := c.SaleDraftsDelete(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeSaleDraftsCreateFromHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.SaleDraftsCreateFromIn, ops.Result[ops.SaleDraftsCreateFromOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.SaleDraftsCreateFromIn) (*mcpsdk.CallToolResult, ops.Result[ops.SaleDraftsCreateFromOut], error) {
		res := c.SaleDraftsCreateFrom(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeSaleDraftsAttachmentsListHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.SaleDraftsAttachmentsListIn, ops.Result[ops.AttachmentsListOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.SaleDraftsAttachmentsListIn) (*mcpsdk.CallToolResult, ops.Result[ops.AttachmentsListOut], error) {
		res := c.SaleDraftsAttachmentsList(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makeSaleDraftsAttachmentsAttachHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.SaleDraftsAttachmentsAttachIn, ops.Result[ops.SaleDraftsAttachmentsAttachOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.SaleDraftsAttachmentsAttachIn) (*mcpsdk.CallToolResult, ops.Result[ops.SaleDraftsAttachmentsAttachOut], error) {
		res := c.SaleDraftsAttachmentsAttach(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makePurchaseDraftsListHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.PurchaseDraftsListIn, ops.Result[ops.PurchaseDraftsListOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.PurchaseDraftsListIn) (*mcpsdk.CallToolResult, ops.Result[ops.PurchaseDraftsListOut], error) {
		res := c.PurchaseDraftsList(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makePurchaseDraftsGetHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.PurchaseDraftsGetIn, ops.Result[ops.PurchaseDraftOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.PurchaseDraftsGetIn) (*mcpsdk.CallToolResult, ops.Result[ops.PurchaseDraftOut], error) {
		res := c.PurchaseDraftsGet(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makePurchaseDraftsCreateHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.PurchaseDraftsCreateIn, ops.Result[ops.PurchaseDraftOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.PurchaseDraftsCreateIn) (*mcpsdk.CallToolResult, ops.Result[ops.PurchaseDraftOut], error) {
		res := c.PurchaseDraftsCreate(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makePurchaseDraftsUpdateHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.PurchaseDraftsUpdateIn, ops.Result[ops.PurchaseDraftOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.PurchaseDraftsUpdateIn) (*mcpsdk.CallToolResult, ops.Result[ops.PurchaseDraftOut], error) {
		res := c.PurchaseDraftsUpdate(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makePurchaseDraftsDeleteHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.PurchaseDraftsDeleteIn, ops.Result[ops.PurchaseDraftsDeleteOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.PurchaseDraftsDeleteIn) (*mcpsdk.CallToolResult, ops.Result[ops.PurchaseDraftsDeleteOut], error) {
		res := c.PurchaseDraftsDelete(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makePurchaseDraftsCreateFromHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.PurchaseDraftsCreateFromIn, ops.Result[ops.PurchaseDraftsCreateFromOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.PurchaseDraftsCreateFromIn) (*mcpsdk.CallToolResult, ops.Result[ops.PurchaseDraftsCreateFromOut], error) {
		res := c.PurchaseDraftsCreateFrom(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makePurchaseDraftsAttachmentsListHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.PurchaseDraftsAttachmentsListIn, ops.Result[ops.AttachmentsListOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.PurchaseDraftsAttachmentsListIn) (*mcpsdk.CallToolResult, ops.Result[ops.AttachmentsListOut], error) {
		res := c.PurchaseDraftsAttachmentsList(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}

func makePurchaseDraftsAttachmentsAttachHandler(c *ops.Client) mcpsdk.ToolHandlerFor[ops.PurchaseDraftsAttachmentsAttachIn, ops.Result[ops.PurchaseDraftsAttachmentsAttachOut]] {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.PurchaseDraftsAttachmentsAttachIn) (*mcpsdk.CallToolResult, ops.Result[ops.PurchaseDraftsAttachmentsAttachOut], error) {
		res := c.PurchaseDraftsAttachmentsAttach(ctx, in)
		r := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			r.IsError = true
		}
		return r, res, nil
	}
}
