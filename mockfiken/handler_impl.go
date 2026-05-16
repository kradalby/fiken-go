package mockfiken

import (
	"context"

	"github.com/kradalby/fiken-go/fiken"
)

// Op-name constants. Duplicated as string literals (not imported from
// ops) to avoid an import cycle: ops tests import mockfiken, and
// mockfiken would otherwise need to import ops for these names. The
// values must stay byte-identical with ops.Op*.
const (
	opCompaniesList = "companies_list"
	opCompaniesGet  = "companies_get"

	opContactsList        = "contacts_list"
	opContactsGet         = "contacts_get"
	opContactsPersonsList = "contacts_persons_list"
	opContactsPersonsGet  = "contacts_persons_get"

	opAccountsList = "accounts_list"
	opAccountsGet  = "accounts_get"

	opBankAccountsList = "bank_accounts_list"
	opBankAccountsGet  = "bank_accounts_get"

	opJournalEntriesList            = "journal_entries_list"
	opJournalEntriesGet             = "journal_entries_get"
	opJournalEntriesAttachmentsList = "journal_entries_attachments_list"

	opTransactionsList = "transactions_list"
	opTransactionsGet  = "transactions_get"

	opInvoicesList = "invoices_list"
	opInvoicesGet  = "invoices_get"

	opInvoicesDraftsList = "invoices_drafts_list"
	opInvoicesDraftsGet  = "invoices_drafts_get"

	opInvoicesAttachmentsList       = "invoices_attachments_list"
	opInvoicesDraftsAttachmentsList = "invoices_drafts_attachments_list"

	opCreditNotesList                  = "credit_notes_list"
	opCreditNotesGet                   = "credit_notes_get"
	opCreditNotesDraftsList            = "credit_notes_drafts_list"
	opCreditNotesDraftsGet             = "credit_notes_drafts_get"
	opCreditNotesDraftsAttachmentsList = "credit_notes_drafts_attachments_list"

	opOffersList                  = "offers_list"
	opOffersGet                   = "offers_get"
	opOffersDraftsList            = "offers_drafts_list"
	opOffersDraftsGet             = "offers_drafts_get"
	opOffersDraftsAttachmentsList = "offers_drafts_attachments_list"

	opOrderConfirmationsList                  = "order_confirmations_list"
	opOrderConfirmationsGet                   = "order_confirmations_get"
	opOrderConfirmationsDraftsList            = "order_confirmations_drafts_list"
	opOrderConfirmationsDraftsGet             = "order_confirmations_drafts_get"
	opOrderConfirmationsDraftsAttachmentsList = "order_confirmations_drafts_attachments_list"

	opProductsList              = "products_list"
	opProductsGet               = "products_get"
	opProductsSalesReportCreate = "products_sales_report_create"

	opSalesList         = "sales_list"
	opSalesGet          = "sales_get"
	opSalesAttachments  = "sales_attachments_list"
	opSalesPaymentsList = "sales_payments_list"
	opSalesPaymentsGet  = "sales_payments_get"

	opPurchasesList         = "purchases_list"
	opPurchasesGet          = "purchases_get"
	opPurchasesAttachments  = "purchases_attachments_list"
	opPurchasesPaymentsList = "purchases_payments_list"
	opPurchasesPaymentsGet  = "purchases_payments_get"

	opInboxList = "inbox_list"
	opInboxGet  = "inbox_get"

	opProjectsList = "projects_list"
	opProjectsGet  = "projects_get"

	opUserGet = "user_get"

	opAccountBalancesList = "account_balances_list"
	opAccountBalancesGet  = "account_balances_get"

	opBankBalancesList = "bank_balances_list"

	opGroupsList = "groups_list"

	opActivitiesList = "activities_list"
	opActivitiesGet  = "activities_get"

	opTimeEntriesList = "time_entries_list"
	opTimeEntriesGet  = "time_entries_get"

	opTimeUsersList = "time_users_list"
	opTimeUsersGet  = "time_users_get"
)

// GetCompanies implements fiken.Handler. Returns the override registered
// for "companies_list" if any; otherwise a zero-value response with
// an empty company list.
func (h *handlerImpl) GetCompanies(_ context.Context, _ fiken.GetCompaniesParams) (*fiken.GetCompaniesOKHeaders, error) {
	v, e, hit := h.server.lookup(opCompaniesList)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.GetCompaniesOKHeaders); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.GetCompaniesOKHeaders", opCompaniesList, v)
	}
	return &fiken.GetCompaniesOKHeaders{Response: []fiken.Company{}}, nil
}

// GetCompany implements fiken.Handler. Returns the override registered
// for "companies_get" if any; otherwise a zero-value Company.
func (h *handlerImpl) GetCompany(_ context.Context, _ fiken.GetCompanyParams) (*fiken.Company, error) {
	v, e, hit := h.server.lookup(opCompaniesGet)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.Company); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.Company", opCompaniesGet, v)
	}
	return &fiken.Company{}, nil
}

// GetContacts implements fiken.Handler. Returns the override registered
// for "contacts_list" if any; otherwise an empty list.
func (h *handlerImpl) GetContacts(_ context.Context, _ fiken.GetContactsParams) (*fiken.GetContactsOKHeaders, error) {
	v, e, hit := h.server.lookup(opContactsList)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.GetContactsOKHeaders); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.GetContactsOKHeaders", opContactsList, v)
	}
	return &fiken.GetContactsOKHeaders{Response: []fiken.Contact{}}, nil
}

// GetContact implements fiken.Handler. Returns the override registered
// for "contacts_get" if any; otherwise a zero-value Contact.
func (h *handlerImpl) GetContact(_ context.Context, _ fiken.GetContactParams) (*fiken.Contact, error) {
	v, e, hit := h.server.lookup(opContactsGet)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.Contact); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.Contact", opContactsGet, v)
	}
	return &fiken.Contact{}, nil
}

// GetContactContactPerson implements fiken.Handler. Returns the
// override registered for "contacts_persons_list" if any; otherwise an
// empty slice.
func (h *handlerImpl) GetContactContactPerson(_ context.Context, _ fiken.GetContactContactPersonParams) ([]fiken.ContactPerson, error) {
	v, e, hit := h.server.lookup(opContactsPersonsList)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.([]fiken.ContactPerson); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want []fiken.ContactPerson", opContactsPersonsList, v)
	}
	return []fiken.ContactPerson{}, nil
}

// GetContactPerson implements fiken.Handler. Returns the override
// registered for "contacts_persons_get" if any; otherwise a zero-value
// ContactPerson.
func (h *handlerImpl) GetContactPerson(_ context.Context, _ fiken.GetContactPersonParams) (*fiken.ContactPerson, error) {
	v, e, hit := h.server.lookup(opContactsPersonsGet)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.ContactPerson); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.ContactPerson", opContactsPersonsGet, v)
	}
	return &fiken.ContactPerson{}, nil
}

// GetAccounts implements fiken.Handler. Returns the override registered
// for "accounts_list" if any; otherwise an empty list.
func (h *handlerImpl) GetAccounts(_ context.Context, _ fiken.GetAccountsParams) (*fiken.GetAccountsOKHeaders, error) {
	v, e, hit := h.server.lookup(opAccountsList)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.GetAccountsOKHeaders); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.GetAccountsOKHeaders", opAccountsList, v)
	}
	return &fiken.GetAccountsOKHeaders{Response: []fiken.Account{}}, nil
}

// GetAccount implements fiken.Handler. Returns the override registered
// for "accounts_get" if any; otherwise a zero-value Account.
func (h *handlerImpl) GetAccount(_ context.Context, _ fiken.GetAccountParams) (*fiken.Account, error) {
	v, e, hit := h.server.lookup(opAccountsGet)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.Account); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.Account", opAccountsGet, v)
	}
	return &fiken.Account{}, nil
}

// GetBankAccounts implements fiken.Handler. Returns the override
// registered for "bank_accounts_list" if any; otherwise an empty list.
func (h *handlerImpl) GetBankAccounts(_ context.Context, _ fiken.GetBankAccountsParams) (*fiken.GetBankAccountsOKHeaders, error) {
	v, e, hit := h.server.lookup(opBankAccountsList)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.GetBankAccountsOKHeaders); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.GetBankAccountsOKHeaders", opBankAccountsList, v)
	}
	return &fiken.GetBankAccountsOKHeaders{Response: []fiken.BankAccountResult{}}, nil
}

// GetBankAccount implements fiken.Handler. Returns the override
// registered for "bank_accounts_get" if any; otherwise a zero-value
// BankAccountResult.
func (h *handlerImpl) GetBankAccount(_ context.Context, _ fiken.GetBankAccountParams) (*fiken.BankAccountResult, error) {
	v, e, hit := h.server.lookup(opBankAccountsGet)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.BankAccountResult); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.BankAccountResult", opBankAccountsGet, v)
	}
	return &fiken.BankAccountResult{}, nil
}

// GetJournalEntries implements fiken.Handler. Returns the override
// registered for "journal_entries_list" if any; otherwise an empty
// list.
func (h *handlerImpl) GetJournalEntries(_ context.Context, _ fiken.GetJournalEntriesParams) (*fiken.GetJournalEntriesOKHeaders, error) {
	v, e, hit := h.server.lookup(opJournalEntriesList)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.GetJournalEntriesOKHeaders); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.GetJournalEntriesOKHeaders", opJournalEntriesList, v)
	}
	return &fiken.GetJournalEntriesOKHeaders{Response: []fiken.JournalEntry{}}, nil
}

// GetJournalEntry implements fiken.Handler. Returns the override
// registered for "journal_entries_get" if any; otherwise a zero-value
// JournalEntry.
func (h *handlerImpl) GetJournalEntry(_ context.Context, _ fiken.GetJournalEntryParams) (*fiken.JournalEntry, error) {
	v, e, hit := h.server.lookup(opJournalEntriesGet)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.JournalEntry); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.JournalEntry", opJournalEntriesGet, v)
	}
	return &fiken.JournalEntry{}, nil
}

// GetJournalEntryAttachments implements fiken.Handler. Returns the
// override registered for "journal_entries_attachments_list" if any;
// otherwise an empty slice.
func (h *handlerImpl) GetJournalEntryAttachments(_ context.Context, _ fiken.GetJournalEntryAttachmentsParams) ([]fiken.Attachment, error) {
	v, e, hit := h.server.lookup(opJournalEntriesAttachmentsList)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.([]fiken.Attachment); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want []fiken.Attachment", opJournalEntriesAttachmentsList, v)
	}
	return []fiken.Attachment{}, nil
}

// GetTransactions implements fiken.Handler. Returns the override
// registered for "transactions_list" if any; otherwise an empty list.
func (h *handlerImpl) GetTransactions(_ context.Context, _ fiken.GetTransactionsParams) (*fiken.GetTransactionsOKHeaders, error) {
	v, e, hit := h.server.lookup(opTransactionsList)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.GetTransactionsOKHeaders); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.GetTransactionsOKHeaders", opTransactionsList, v)
	}
	return &fiken.GetTransactionsOKHeaders{Response: []fiken.Transaction{}}, nil
}

// GetTransaction implements fiken.Handler. Returns the override
// registered for "transactions_get" if any; otherwise a zero-value
// Transaction.
func (h *handlerImpl) GetTransaction(_ context.Context, _ fiken.GetTransactionParams) (*fiken.Transaction, error) {
	v, e, hit := h.server.lookup(opTransactionsGet)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.Transaction); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.Transaction", opTransactionsGet, v)
	}
	return &fiken.Transaction{}, nil
}

// GetInvoices implements fiken.Handler. Returns the override registered
// for "invoices_list" if any; otherwise an empty list.
func (h *handlerImpl) GetInvoices(_ context.Context, _ fiken.GetInvoicesParams) (*fiken.GetInvoicesOKHeaders, error) {
	v, e, hit := h.server.lookup(opInvoicesList)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.GetInvoicesOKHeaders); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.GetInvoicesOKHeaders", opInvoicesList, v)
	}
	return &fiken.GetInvoicesOKHeaders{Response: []fiken.InvoiceResult{}}, nil
}

// GetInvoice implements fiken.Handler. Returns the override registered
// for "invoices_get" if any; otherwise a zero-value InvoiceResult.
func (h *handlerImpl) GetInvoice(_ context.Context, _ fiken.GetInvoiceParams) (*fiken.InvoiceResult, error) {
	v, e, hit := h.server.lookup(opInvoicesGet)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.InvoiceResult); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.InvoiceResult", opInvoicesGet, v)
	}
	return &fiken.InvoiceResult{}, nil
}

// GetInvoiceDrafts implements fiken.Handler. Returns the override
// registered for "invoices_drafts_list" if any; otherwise an empty list.
func (h *handlerImpl) GetInvoiceDrafts(_ context.Context, _ fiken.GetInvoiceDraftsParams) (*fiken.GetInvoiceDraftsOKHeaders, error) {
	v, e, hit := h.server.lookup(opInvoicesDraftsList)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.GetInvoiceDraftsOKHeaders); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.GetInvoiceDraftsOKHeaders", opInvoicesDraftsList, v)
	}
	return &fiken.GetInvoiceDraftsOKHeaders{Response: []fiken.InvoiceishDraftResult{}}, nil
}

// GetInvoiceDraft implements fiken.Handler. Returns the override
// registered for "invoices_drafts_get" if any; otherwise a zero-value
// InvoiceishDraftResult.
func (h *handlerImpl) GetInvoiceDraft(_ context.Context, _ fiken.GetInvoiceDraftParams) (*fiken.InvoiceishDraftResult, error) {
	v, e, hit := h.server.lookup(opInvoicesDraftsGet)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.InvoiceishDraftResult); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.InvoiceishDraftResult", opInvoicesDraftsGet, v)
	}
	return &fiken.InvoiceishDraftResult{}, nil
}

// GetInvoiceAttachments implements fiken.Handler. Returns the override
// registered for "invoices_attachments_list" if any; otherwise an
// empty slice.
func (h *handlerImpl) GetInvoiceAttachments(_ context.Context, _ fiken.GetInvoiceAttachmentsParams) ([]fiken.Attachment, error) {
	v, e, hit := h.server.lookup(opInvoicesAttachmentsList)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.([]fiken.Attachment); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want []fiken.Attachment", opInvoicesAttachmentsList, v)
	}
	return []fiken.Attachment{}, nil
}

// GetInvoiceDraftAttachments implements fiken.Handler. Returns the
// override registered for "invoices_drafts_attachments_list" if any;
// otherwise an empty slice.
func (h *handlerImpl) GetInvoiceDraftAttachments(_ context.Context, _ fiken.GetInvoiceDraftAttachmentsParams) ([]fiken.Attachment, error) {
	v, e, hit := h.server.lookup(opInvoicesDraftsAttachmentsList)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.([]fiken.Attachment); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want []fiken.Attachment", opInvoicesDraftsAttachmentsList, v)
	}
	return []fiken.Attachment{}, nil
}

// GetCreditNotes implements fiken.Handler. Returns the override
// registered for "credit_notes_list" if any; otherwise an empty list.
func (h *handlerImpl) GetCreditNotes(_ context.Context, _ fiken.GetCreditNotesParams) (*fiken.GetCreditNotesOKHeaders, error) {
	v, e, hit := h.server.lookup(opCreditNotesList)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.GetCreditNotesOKHeaders); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.GetCreditNotesOKHeaders", opCreditNotesList, v)
	}
	return &fiken.GetCreditNotesOKHeaders{Response: []fiken.CreditNoteResult{}}, nil
}

// GetCreditNote implements fiken.Handler. Returns the override
// registered for "credit_notes_get" if any; otherwise a zero-value
// CreditNoteResult.
func (h *handlerImpl) GetCreditNote(_ context.Context, _ fiken.GetCreditNoteParams) (*fiken.CreditNoteResult, error) {
	v, e, hit := h.server.lookup(opCreditNotesGet)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.CreditNoteResult); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.CreditNoteResult", opCreditNotesGet, v)
	}
	return &fiken.CreditNoteResult{}, nil
}

// GetCreditNoteDrafts implements fiken.Handler. Returns the override
// registered for "credit_notes_drafts_list" if any; otherwise an
// empty list.
func (h *handlerImpl) GetCreditNoteDrafts(_ context.Context, _ fiken.GetCreditNoteDraftsParams) (*fiken.GetCreditNoteDraftsOKHeaders, error) {
	v, e, hit := h.server.lookup(opCreditNotesDraftsList)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.GetCreditNoteDraftsOKHeaders); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.GetCreditNoteDraftsOKHeaders", opCreditNotesDraftsList, v)
	}
	return &fiken.GetCreditNoteDraftsOKHeaders{Response: []fiken.InvoiceishDraftResult{}}, nil
}

// GetCreditNoteDraft implements fiken.Handler. Returns the override
// registered for "credit_notes_drafts_get" if any; otherwise a
// zero-value InvoiceishDraftResult.
func (h *handlerImpl) GetCreditNoteDraft(_ context.Context, _ fiken.GetCreditNoteDraftParams) (*fiken.InvoiceishDraftResult, error) {
	v, e, hit := h.server.lookup(opCreditNotesDraftsGet)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.InvoiceishDraftResult); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.InvoiceishDraftResult", opCreditNotesDraftsGet, v)
	}
	return &fiken.InvoiceishDraftResult{}, nil
}

// GetCreditNoteDraftAttachments implements fiken.Handler. Returns the
// override registered for "credit_notes_drafts_attachments_list" if
// any; otherwise an empty slice.
func (h *handlerImpl) GetCreditNoteDraftAttachments(_ context.Context, _ fiken.GetCreditNoteDraftAttachmentsParams) ([]fiken.Attachment, error) {
	v, e, hit := h.server.lookup(opCreditNotesDraftsAttachmentsList)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.([]fiken.Attachment); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want []fiken.Attachment", opCreditNotesDraftsAttachmentsList, v)
	}
	return []fiken.Attachment{}, nil
}

// GetOffers implements fiken.Handler. Returns the override registered
// for "offers_list" if any; otherwise an empty list.
func (h *handlerImpl) GetOffers(_ context.Context, _ fiken.GetOffersParams) (*fiken.GetOffersOKHeaders, error) {
	v, e, hit := h.server.lookup(opOffersList)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.GetOffersOKHeaders); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.GetOffersOKHeaders", opOffersList, v)
	}
	return &fiken.GetOffersOKHeaders{Response: []fiken.Offer{}}, nil
}

// GetOffer implements fiken.Handler. Returns the override registered
// for "offers_get" if any; otherwise a zero-value Offer.
func (h *handlerImpl) GetOffer(_ context.Context, _ fiken.GetOfferParams) (*fiken.Offer, error) {
	v, e, hit := h.server.lookup(opOffersGet)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.Offer); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.Offer", opOffersGet, v)
	}
	return &fiken.Offer{}, nil
}

// GetOfferDrafts implements fiken.Handler. Returns the override
// registered for "offers_drafts_list" if any; otherwise an empty list.
func (h *handlerImpl) GetOfferDrafts(_ context.Context, _ fiken.GetOfferDraftsParams) (*fiken.GetOfferDraftsOKHeaders, error) {
	v, e, hit := h.server.lookup(opOffersDraftsList)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.GetOfferDraftsOKHeaders); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.GetOfferDraftsOKHeaders", opOffersDraftsList, v)
	}
	return &fiken.GetOfferDraftsOKHeaders{Response: []fiken.InvoiceishDraftResult{}}, nil
}

// GetOfferDraft implements fiken.Handler. Returns the override
// registered for "offers_drafts_get" if any; otherwise a zero-value
// InvoiceishDraftResult.
func (h *handlerImpl) GetOfferDraft(_ context.Context, _ fiken.GetOfferDraftParams) (*fiken.InvoiceishDraftResult, error) {
	v, e, hit := h.server.lookup(opOffersDraftsGet)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.InvoiceishDraftResult); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.InvoiceishDraftResult", opOffersDraftsGet, v)
	}
	return &fiken.InvoiceishDraftResult{}, nil
}

// GetOfferDraftAttachments implements fiken.Handler. Returns the
// override registered for "offers_drafts_attachments_list" if any;
// otherwise an empty slice.
func (h *handlerImpl) GetOfferDraftAttachments(_ context.Context, _ fiken.GetOfferDraftAttachmentsParams) ([]fiken.Attachment, error) {
	v, e, hit := h.server.lookup(opOffersDraftsAttachmentsList)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.([]fiken.Attachment); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want []fiken.Attachment", opOffersDraftsAttachmentsList, v)
	}
	return []fiken.Attachment{}, nil
}

// GetOrderConfirmations implements fiken.Handler. Returns the override
// registered for "order_confirmations_list" if any; otherwise an empty
// list.
func (h *handlerImpl) GetOrderConfirmations(_ context.Context, _ fiken.GetOrderConfirmationsParams) (*fiken.GetOrderConfirmationsOKHeaders, error) {
	v, e, hit := h.server.lookup(opOrderConfirmationsList)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.GetOrderConfirmationsOKHeaders); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.GetOrderConfirmationsOKHeaders", opOrderConfirmationsList, v)
	}
	return &fiken.GetOrderConfirmationsOKHeaders{Response: []fiken.OrderConfirmation{}}, nil
}

// GetOrderConfirmation implements fiken.Handler. Returns the override
// registered for "order_confirmations_get" if any; otherwise a
// zero-value OrderConfirmation.
func (h *handlerImpl) GetOrderConfirmation(_ context.Context, _ fiken.GetOrderConfirmationParams) (*fiken.OrderConfirmation, error) {
	v, e, hit := h.server.lookup(opOrderConfirmationsGet)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.OrderConfirmation); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.OrderConfirmation", opOrderConfirmationsGet, v)
	}
	return &fiken.OrderConfirmation{}, nil
}

// GetOrderConfirmationDrafts implements fiken.Handler. Returns the
// override registered for "order_confirmations_drafts_list" if any;
// otherwise an empty list.
func (h *handlerImpl) GetOrderConfirmationDrafts(_ context.Context, _ fiken.GetOrderConfirmationDraftsParams) (*fiken.GetOrderConfirmationDraftsOKHeaders, error) {
	v, e, hit := h.server.lookup(opOrderConfirmationsDraftsList)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.GetOrderConfirmationDraftsOKHeaders); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.GetOrderConfirmationDraftsOKHeaders", opOrderConfirmationsDraftsList, v)
	}
	return &fiken.GetOrderConfirmationDraftsOKHeaders{Response: []fiken.InvoiceishDraftResult{}}, nil
}

// GetOrderConfirmationDraft implements fiken.Handler. Returns the
// override registered for "order_confirmations_drafts_get" if any;
// otherwise a zero-value InvoiceishDraftResult.
func (h *handlerImpl) GetOrderConfirmationDraft(_ context.Context, _ fiken.GetOrderConfirmationDraftParams) (*fiken.InvoiceishDraftResult, error) {
	v, e, hit := h.server.lookup(opOrderConfirmationsDraftsGet)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.InvoiceishDraftResult); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.InvoiceishDraftResult", opOrderConfirmationsDraftsGet, v)
	}
	return &fiken.InvoiceishDraftResult{}, nil
}

// GetOrderConfirmationDraftAttachments implements fiken.Handler. Returns
// the override registered for "order_confirmations_drafts_attachments_list"
// if any; otherwise an empty slice.
func (h *handlerImpl) GetOrderConfirmationDraftAttachments(_ context.Context, _ fiken.GetOrderConfirmationDraftAttachmentsParams) ([]fiken.Attachment, error) {
	v, e, hit := h.server.lookup(opOrderConfirmationsDraftsAttachmentsList)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.([]fiken.Attachment); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want []fiken.Attachment", opOrderConfirmationsDraftsAttachmentsList, v)
	}
	return []fiken.Attachment{}, nil
}

// GetProducts implements fiken.Handler. Returns the override registered
// for "products_list" if any; otherwise an empty list.
func (h *handlerImpl) GetProducts(_ context.Context, _ fiken.GetProductsParams) (*fiken.GetProductsOKHeaders, error) {
	v, e, hit := h.server.lookup(opProductsList)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.GetProductsOKHeaders); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.GetProductsOKHeaders", opProductsList, v)
	}
	return &fiken.GetProductsOKHeaders{Response: []fiken.Product{}}, nil
}

// GetProduct implements fiken.Handler. Returns the override registered
// for "products_get" if any; otherwise a zero-value Product.
func (h *handlerImpl) GetProduct(_ context.Context, _ fiken.GetProductParams) (*fiken.Product, error) {
	v, e, hit := h.server.lookup(opProductsGet)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.Product); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.Product", opProductsGet, v)
	}
	return &fiken.Product{}, nil
}

// CreateProductSalesReport implements fiken.Handler. Returns the
// override registered for "products_sales_report_create" if any;
// otherwise an empty slice.
func (h *handlerImpl) CreateProductSalesReport(_ context.Context, _ *fiken.ProductSalesReportRequest, _ fiken.CreateProductSalesReportParams) ([]fiken.ProductSalesReportResult, error) {
	v, e, hit := h.server.lookup(opProductsSalesReportCreate)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.([]fiken.ProductSalesReportResult); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want []fiken.ProductSalesReportResult", opProductsSalesReportCreate, v)
	}
	return []fiken.ProductSalesReportResult{}, nil
}

// GetSales implements fiken.Handler. Returns the override registered
// for "sales_list" if any; otherwise an empty list.
func (h *handlerImpl) GetSales(_ context.Context, _ fiken.GetSalesParams) (*fiken.GetSalesOKHeaders, error) {
	v, e, hit := h.server.lookup(opSalesList)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.GetSalesOKHeaders); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.GetSalesOKHeaders", opSalesList, v)
	}
	return &fiken.GetSalesOKHeaders{Response: []fiken.SaleResult{}}, nil
}

// GetSale implements fiken.Handler. Returns the override registered
// for "sales_get" if any; otherwise a zero-value SaleResult.
func (h *handlerImpl) GetSale(_ context.Context, _ fiken.GetSaleParams) (*fiken.SaleResult, error) {
	v, e, hit := h.server.lookup(opSalesGet)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.SaleResult); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.SaleResult", opSalesGet, v)
	}
	return &fiken.SaleResult{}, nil
}

// GetSaleAttachments implements fiken.Handler. Returns the override
// registered for "sales_attachments_list" if any; otherwise an empty
// slice.
func (h *handlerImpl) GetSaleAttachments(_ context.Context, _ fiken.GetSaleAttachmentsParams) ([]fiken.Attachment, error) {
	v, e, hit := h.server.lookup(opSalesAttachments)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.([]fiken.Attachment); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want []fiken.Attachment", opSalesAttachments, v)
	}
	return []fiken.Attachment{}, nil
}

// GetSalePayments implements fiken.Handler. Returns the override
// registered for "sales_payments_list" if any; otherwise an empty slice.
func (h *handlerImpl) GetSalePayments(_ context.Context, _ fiken.GetSalePaymentsParams) ([]fiken.Payment, error) {
	v, e, hit := h.server.lookup(opSalesPaymentsList)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.([]fiken.Payment); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want []fiken.Payment", opSalesPaymentsList, v)
	}
	return []fiken.Payment{}, nil
}

// GetSalePayment implements fiken.Handler. Returns the override
// registered for "sales_payments_get" if any; otherwise a zero-value
// Payment.
func (h *handlerImpl) GetSalePayment(_ context.Context, _ fiken.GetSalePaymentParams) (*fiken.Payment, error) {
	v, e, hit := h.server.lookup(opSalesPaymentsGet)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.Payment); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.Payment", opSalesPaymentsGet, v)
	}
	return &fiken.Payment{}, nil
}

// GetPurchases implements fiken.Handler. Returns the override registered
// for "purchases_list" if any; otherwise an empty list.
func (h *handlerImpl) GetPurchases(_ context.Context, _ fiken.GetPurchasesParams) (*fiken.GetPurchasesOKHeaders, error) {
	v, e, hit := h.server.lookup(opPurchasesList)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.GetPurchasesOKHeaders); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.GetPurchasesOKHeaders", opPurchasesList, v)
	}
	return &fiken.GetPurchasesOKHeaders{Response: []fiken.PurchaseResult{}}, nil
}

// GetPurchase implements fiken.Handler. Returns the override registered
// for "purchases_get" if any; otherwise a zero-value PurchaseResult.
func (h *handlerImpl) GetPurchase(_ context.Context, _ fiken.GetPurchaseParams) (*fiken.PurchaseResult, error) {
	v, e, hit := h.server.lookup(opPurchasesGet)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.PurchaseResult); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.PurchaseResult", opPurchasesGet, v)
	}
	return &fiken.PurchaseResult{}, nil
}

// GetPurchaseAttachments implements fiken.Handler. Returns the override
// registered for "purchases_attachments_list" if any; otherwise an
// empty slice.
func (h *handlerImpl) GetPurchaseAttachments(_ context.Context, _ fiken.GetPurchaseAttachmentsParams) ([]fiken.Attachment, error) {
	v, e, hit := h.server.lookup(opPurchasesAttachments)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.([]fiken.Attachment); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want []fiken.Attachment", opPurchasesAttachments, v)
	}
	return []fiken.Attachment{}, nil
}

// GetPurchasePayments implements fiken.Handler. Returns the override
// registered for "purchases_payments_list" if any; otherwise an empty
// slice.
func (h *handlerImpl) GetPurchasePayments(_ context.Context, _ fiken.GetPurchasePaymentsParams) ([]fiken.Payment, error) {
	v, e, hit := h.server.lookup(opPurchasesPaymentsList)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.([]fiken.Payment); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want []fiken.Payment", opPurchasesPaymentsList, v)
	}
	return []fiken.Payment{}, nil
}

// GetPurchasePayment implements fiken.Handler. Returns the override
// registered for "purchases_payments_get" if any; otherwise a
// zero-value Payment.
func (h *handlerImpl) GetPurchasePayment(_ context.Context, _ fiken.GetPurchasePaymentParams) (*fiken.Payment, error) {
	v, e, hit := h.server.lookup(opPurchasesPaymentsGet)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.Payment); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.Payment", opPurchasesPaymentsGet, v)
	}
	return &fiken.Payment{}, nil
}

// GetInbox implements fiken.Handler. Returns the override registered
// for "inbox_list" if any; otherwise an empty list.
func (h *handlerImpl) GetInbox(_ context.Context, _ fiken.GetInboxParams) (*fiken.GetInboxOKHeaders, error) {
	v, e, hit := h.server.lookup(opInboxList)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.GetInboxOKHeaders); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.GetInboxOKHeaders", opInboxList, v)
	}
	return &fiken.GetInboxOKHeaders{Response: []fiken.InboxResult{}}, nil
}

// GetInboxDocument implements fiken.Handler. Returns the override
// registered for "inbox_get" if any; otherwise a zero-value
// InboxResult.
func (h *handlerImpl) GetInboxDocument(_ context.Context, _ fiken.GetInboxDocumentParams) (*fiken.InboxResult, error) {
	v, e, hit := h.server.lookup(opInboxGet)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.InboxResult); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.InboxResult", opInboxGet, v)
	}
	return &fiken.InboxResult{}, nil
}

// GetProjects implements fiken.Handler. Returns the override registered
// for "projects_list" if any; otherwise an empty list.
func (h *handlerImpl) GetProjects(_ context.Context, _ fiken.GetProjectsParams) (*fiken.GetProjectsOKHeaders, error) {
	v, e, hit := h.server.lookup(opProjectsList)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.GetProjectsOKHeaders); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.GetProjectsOKHeaders", opProjectsList, v)
	}
	return &fiken.GetProjectsOKHeaders{Response: []fiken.ProjectResult{}}, nil
}

// GetProject implements fiken.Handler. Returns the override registered
// for "projects_get" if any; otherwise a zero-value ProjectResult.
func (h *handlerImpl) GetProject(_ context.Context, _ fiken.GetProjectParams) (*fiken.ProjectResult, error) {
	v, e, hit := h.server.lookup(opProjectsGet)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.ProjectResult); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.ProjectResult", opProjectsGet, v)
	}
	return &fiken.ProjectResult{}, nil
}

// GetUser implements fiken.Handler. Returns the override registered for
// "user_get" if any; otherwise a zero-value Userinfo.
func (h *handlerImpl) GetUser(_ context.Context) (*fiken.Userinfo, error) {
	v, e, hit := h.server.lookup(opUserGet)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.Userinfo); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.Userinfo", opUserGet, v)
	}
	return &fiken.Userinfo{}, nil
}

// GetAccountBalances implements fiken.Handler. Returns the override
// registered for "account_balances_list" if any; otherwise an empty
// list.
func (h *handlerImpl) GetAccountBalances(_ context.Context, _ fiken.GetAccountBalancesParams) (*fiken.GetAccountBalancesOKHeaders, error) {
	v, e, hit := h.server.lookup(opAccountBalancesList)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.GetAccountBalancesOKHeaders); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.GetAccountBalancesOKHeaders", opAccountBalancesList, v)
	}
	return &fiken.GetAccountBalancesOKHeaders{Response: []fiken.AccountBalance{}}, nil
}

// GetAccountBalance implements fiken.Handler. Returns the override
// registered for "account_balances_get" if any; otherwise a zero-value
// AccountBalance.
func (h *handlerImpl) GetAccountBalance(_ context.Context, _ fiken.GetAccountBalanceParams) (*fiken.AccountBalance, error) {
	v, e, hit := h.server.lookup(opAccountBalancesGet)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.AccountBalance); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.AccountBalance", opAccountBalancesGet, v)
	}
	return &fiken.AccountBalance{}, nil
}

// GetBankBalances implements fiken.Handler. Returns the override
// registered for "bank_balances_list" if any; otherwise an empty list.
func (h *handlerImpl) GetBankBalances(_ context.Context, _ fiken.GetBankBalancesParams) (*fiken.GetBankBalancesOKHeaders, error) {
	v, e, hit := h.server.lookup(opBankBalancesList)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.GetBankBalancesOKHeaders); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.GetBankBalancesOKHeaders", opBankBalancesList, v)
	}
	return &fiken.GetBankBalancesOKHeaders{Response: []fiken.BankBalanceResult{}}, nil
}

// GetGroups implements fiken.Handler. Returns the override registered
// for "groups_list" if any; otherwise an empty list of group names.
func (h *handlerImpl) GetGroups(_ context.Context, _ fiken.GetGroupsParams) (*fiken.GetGroupsOKHeaders, error) {
	v, e, hit := h.server.lookup(opGroupsList)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.GetGroupsOKHeaders); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.GetGroupsOKHeaders", opGroupsList, v)
	}
	return &fiken.GetGroupsOKHeaders{Response: []string{}}, nil
}

// GetActivities implements fiken.Handler. Returns the override
// registered for "activities_list" if any; otherwise an empty list.
func (h *handlerImpl) GetActivities(_ context.Context, _ fiken.GetActivitiesParams) (*fiken.GetActivitiesOKHeaders, error) {
	v, e, hit := h.server.lookup(opActivitiesList)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.GetActivitiesOKHeaders); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.GetActivitiesOKHeaders", opActivitiesList, v)
	}
	return &fiken.GetActivitiesOKHeaders{Response: []fiken.ActivityResult{}}, nil
}

// GetActivity implements fiken.Handler. Returns the override registered
// for "activities_get" if any; otherwise a zero-value ActivityResult.
func (h *handlerImpl) GetActivity(_ context.Context, _ fiken.GetActivityParams) (*fiken.ActivityResult, error) {
	v, e, hit := h.server.lookup(opActivitiesGet)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.ActivityResult); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.ActivityResult", opActivitiesGet, v)
	}
	return &fiken.ActivityResult{}, nil
}

// GetTimeEntries implements fiken.Handler. Returns the override
// registered for "time_entries_list" if any; otherwise an empty list.
func (h *handlerImpl) GetTimeEntries(_ context.Context, _ fiken.GetTimeEntriesParams) (*fiken.GetTimeEntriesOKHeaders, error) {
	v, e, hit := h.server.lookup(opTimeEntriesList)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.GetTimeEntriesOKHeaders); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.GetTimeEntriesOKHeaders", opTimeEntriesList, v)
	}
	return &fiken.GetTimeEntriesOKHeaders{Response: []fiken.TimeEntryResult{}}, nil
}

// GetTimeEntry implements fiken.Handler. Returns the override
// registered for "time_entries_get" if any; otherwise a zero-value
// TimeEntryResult.
func (h *handlerImpl) GetTimeEntry(_ context.Context, _ fiken.GetTimeEntryParams) (*fiken.TimeEntryResult, error) {
	v, e, hit := h.server.lookup(opTimeEntriesGet)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.TimeEntryResult); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.TimeEntryResult", opTimeEntriesGet, v)
	}
	return &fiken.TimeEntryResult{}, nil
}

// GetTimeUsers implements fiken.Handler. Returns the override
// registered for "time_users_list" if any; otherwise an empty list.
func (h *handlerImpl) GetTimeUsers(_ context.Context, _ fiken.GetTimeUsersParams) (*fiken.GetTimeUsersOKHeaders, error) {
	v, e, hit := h.server.lookup(opTimeUsersList)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.GetTimeUsersOKHeaders); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.GetTimeUsersOKHeaders", opTimeUsersList, v)
	}
	return &fiken.GetTimeUsersOKHeaders{Response: []fiken.TimeUserResult{}}, nil
}

// GetTimeUser implements fiken.Handler. Returns the override registered
// for "time_users_get" if any; otherwise a zero-value TimeUserResult.
func (h *handlerImpl) GetTimeUser(_ context.Context, _ fiken.GetTimeUserParams) (*fiken.TimeUserResult, error) {
	v, e, hit := h.server.lookup(opTimeUsersGet)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.TimeUserResult); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.TimeUserResult", opTimeUsersGet, v)
	}
	return &fiken.TimeUserResult{}, nil
}

// === A3 mutating-handler additions ===

// Op-name constants for the mutating handlers added in Phase A3.
// Values stay byte-identical with ops.Op* — same import-cycle
// avoidance as the original block above.
const (
	opContactsPersonsCreate                     = "contacts_persons_create"
	opContactsPersonsUpdate                     = "contacts_persons_update"
	opContactsPersonsDelete                     = "contacts_persons_delete"
	opContactsCreate                            = "contacts_create"
	opContactsUpdate                            = "contacts_update"
	opContactsDelete                            = "contacts_delete"
	opBankAccountsCreate                        = "bank_accounts_create"
	opJournalEntriesCreate                      = "journal_entries_create"
	opInvoicesSend                              = "invoices_send"
	opInvoicesCounterCreate                     = "invoices_counter_create"
	opInvoicesDraftsCreate                      = "invoices_drafts_create"
	opInvoicesDraftsUpdate                      = "invoices_drafts_update"
	opInvoicesDraftsDelete                      = "invoices_drafts_delete"
	opInvoicesDraftsCreateFrom                  = "invoices_drafts_create_from"
	opCreditNotesSend                           = "credit_notes_send"
	opCreditNotesCounterCreate                  = "credit_notes_counter_create"
	opCreditNotesFullCreate                     = "credit_notes_full_create"
	opCreditNotesPartialCreate                  = "credit_notes_partial_create"
	opCreditNotesDraftsCreate                   = "credit_notes_drafts_create"
	opCreditNotesDraftsUpdate                   = "credit_notes_drafts_update"
	opCreditNotesDraftsDelete                   = "credit_notes_drafts_delete"
	opCreditNotesDraftsCreateFrom               = "credit_notes_drafts_create_from"
	opOffersSend                                = "offers_send"
	opOffersCounterCreate                       = "offers_counter_create"
	opOffersDraftsCreate                        = "offers_drafts_create"
	opOffersDraftsUpdate                        = "offers_drafts_update"
	opOffersDraftsDelete                        = "offers_drafts_delete"
	opOffersDraftsCreateFrom                    = "offers_drafts_create_from"
	opOrderConfirmationsCounterCreate           = "order_confirmations_counter_create"
	opOrderConfirmationsCreateInvoiceDraft      = "order_confirmations_create_invoice_draft"
	opOrderConfirmationsDraftsCreate            = "order_confirmations_drafts_create"
	opOrderConfirmationsDraftsUpdate            = "order_confirmations_drafts_update"
	opOrderConfirmationsDraftsDelete            = "order_confirmations_drafts_delete"
	opOrderConfirmationsDraftsCreateFrom        = "order_confirmations_drafts_create_from"
	opProductsCreate                            = "products_create"
	opProductsUpdate                            = "products_update"
	opProductsDelete                            = "products_delete"
	opSalesCreate                               = "sales_create"
	opSalesDelete                               = "sales_delete"
	opSalesSettle                               = "sales_settle"
	opSalesWriteOff                             = "sales_write_off"
	opSalesPaymentsCreate                       = "sales_payments_create"
	opPurchasesCreate                           = "purchases_create"
	opPurchasesDelete                           = "purchases_delete"
	opPurchasesPaymentsCreate                   = "purchases_payments_create"
	opInboxSend                                 = "inbox_send"
	opProjectsCreate                            = "projects_create"
	opProjectsUpdate                            = "projects_update"
	opProjectsDelete                            = "projects_delete"
	opActivitiesCreate                          = "activities_create"
	opActivitiesUpdate                          = "activities_update"
	opActivitiesDelete                          = "activities_delete"
	opTimeEntriesCreate                         = "time_entries_create"
	opTimeEntriesUpdate                         = "time_entries_update"
	opTimeEntriesDelete                         = "time_entries_delete"
	opTimeEntriesInvoiceDraftFromTimes          = "time_entries_invoice_draft_create"
	opInvoicesAttachmentsAttach                 = "invoices_attachments_attach"
	opJournalEntriesAttachmentsAttach           = "journal_entries_attachments_attach"
	opInvoicesDraftsAttachmentsAttach           = "invoices_drafts_attachments_attach"
	opCreditNotesDraftsAttachmentsAttach        = "credit_notes_drafts_attachments_attach"
	opOffersDraftsAttachmentsAttach             = "offers_drafts_attachments_attach"
	opOrderConfirmationsDraftsAttachmentsAttach = "order_confirmations_drafts_attachments_attach"
	opSalesAttach                               = "sales_attachments_attach"
	opPurchasesAttach                           = "purchases_attachments_attach"

	// Plan D / 21-op tail additions.
	opInvoicesCreate                   = "invoices_create"
	opInvoicesUpdate                   = "invoices_update"
	opInvoicesCounterGet               = "invoices_counter_get"
	opOffersCounterGet                 = "offers_counter_get"
	opOrderConfirmationsCounterGet     = "order_confirmations_counter_get"
	opCreditNotesCounterGet            = "credit_notes_counter_get"
	opTransactionsDelete               = "transactions_delete"
	opContactsAttachmentsAttach        = "contacts_attachments_attach"
	opSalesDraftsList                  = "sales_drafts_list"
	opSalesDraftsGet                   = "sales_drafts_get"
	opSalesDraftsCreate                = "sales_drafts_create"
	opSalesDraftsUpdate                = "sales_drafts_update"
	opSalesDraftsDelete                = "sales_drafts_delete"
	opSalesDraftsCreateFrom            = "sales_drafts_create_from"
	opSalesDraftsAttachmentsList       = "sales_drafts_attachments_list"
	opSalesDraftsAttachmentsAttach     = "sales_drafts_attachments_attach"
	opPurchasesDraftsList              = "purchases_drafts_list"
	opPurchasesDraftsGet               = "purchases_drafts_get"
	opPurchasesDraftsCreate            = "purchases_drafts_create"
	opPurchasesDraftsUpdate            = "purchases_drafts_update"
	opPurchasesDraftsDelete            = "purchases_drafts_delete"
	opPurchasesDraftsCreateFrom        = "purchases_drafts_create_from"
	opPurchasesDraftsAttachmentsList   = "purchases_drafts_attachments_list"
	opPurchasesDraftsAttachmentsAttach = "purchases_drafts_attachments_attach"
)

// AddContactPersonToContact implements fiken.Handler. Returns the override registered
// for 'contacts_persons_create' if any; otherwise a zero-value *fiken.AddContactPersonToContactOK.
func (h *handlerImpl) AddContactPersonToContact(_ context.Context, _ *fiken.ContactPerson, _ fiken.AddContactPersonToContactParams) (*fiken.AddContactPersonToContactOK, error) {
	v, e, hit := h.server.lookup(opContactsPersonsCreate)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.AddContactPersonToContactOK); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.AddContactPersonToContactOK", opContactsPersonsCreate, v)
	}
	return &fiken.AddContactPersonToContactOK{}, nil
}

// UpdateContactContactPerson implements fiken.Handler. Returns the override registered
// for 'contacts_persons_update' if any; otherwise a zero-value *fiken.UpdateContactContactPersonOK.
func (h *handlerImpl) UpdateContactContactPerson(_ context.Context, _ *fiken.ContactPerson, _ fiken.UpdateContactContactPersonParams) (*fiken.UpdateContactContactPersonOK, error) {
	v, e, hit := h.server.lookup(opContactsPersonsUpdate)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.UpdateContactContactPersonOK); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.UpdateContactContactPersonOK", opContactsPersonsUpdate, v)
	}
	return &fiken.UpdateContactContactPersonOK{}, nil
}

// DeleteContactContactPerson implements fiken.Handler. Returns the override registered
// for 'contacts_persons_delete' if any; otherwise nil (success, no body).
func (h *handlerImpl) DeleteContactContactPerson(_ context.Context, _ fiken.DeleteContactContactPersonParams) error {
	_, e, hit := h.server.lookup(opContactsPersonsDelete)
	if hit {
		if e != nil {
			return e
		}
	}
	return nil
}

// CreateContact implements fiken.Handler. Returns the override registered
// for 'contacts_create' if any; otherwise a zero-value *fiken.CreateContactCreated.
func (h *handlerImpl) CreateContact(_ context.Context, _ *fiken.Contact, _ fiken.CreateContactParams) (*fiken.CreateContactCreated, error) {
	v, e, hit := h.server.lookup(opContactsCreate)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.CreateContactCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.CreateContactCreated", opContactsCreate, v)
	}
	return &fiken.CreateContactCreated{}, nil
}

// UpdateContact implements fiken.Handler. Returns the override registered
// for 'contacts_update' if any; otherwise a zero-value *fiken.UpdateContactOK.
func (h *handlerImpl) UpdateContact(_ context.Context, _ *fiken.Contact, _ fiken.UpdateContactParams) (*fiken.UpdateContactOK, error) {
	v, e, hit := h.server.lookup(opContactsUpdate)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.UpdateContactOK); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.UpdateContactOK", opContactsUpdate, v)
	}
	return &fiken.UpdateContactOK{}, nil
}

// DeleteContact implements fiken.Handler. Returns the override registered
// for 'contacts_delete' if any; otherwise nil (no-body NoContent path).
func (h *handlerImpl) DeleteContact(_ context.Context, _ fiken.DeleteContactParams) (fiken.DeleteContactRes, error) {
	v, e, hit := h.server.lookup(opContactsDelete)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(fiken.DeleteContactRes); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want fiken.DeleteContactRes", opContactsDelete, v)
	}
	return nil, nil
}

// CreateBankAccount implements fiken.Handler. Returns the override registered
// for 'bank_accounts_create' if any; otherwise a zero-value *fiken.CreateBankAccountCreated.
func (h *handlerImpl) CreateBankAccount(_ context.Context, _ *fiken.BankAccountRequest, _ fiken.CreateBankAccountParams) (*fiken.CreateBankAccountCreated, error) {
	v, e, hit := h.server.lookup(opBankAccountsCreate)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.CreateBankAccountCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.CreateBankAccountCreated", opBankAccountsCreate, v)
	}
	return &fiken.CreateBankAccountCreated{}, nil
}

// CreateGeneralJournalEntry implements fiken.Handler. Returns the override registered
// for 'journal_entries_create' if any; otherwise a zero-value *fiken.CreateGeneralJournalEntryCreated.
func (h *handlerImpl) CreateGeneralJournalEntry(_ context.Context, _ *fiken.GeneralJournalEntryRequest, _ fiken.CreateGeneralJournalEntryParams) (*fiken.CreateGeneralJournalEntryCreated, error) {
	v, e, hit := h.server.lookup(opJournalEntriesCreate)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.CreateGeneralJournalEntryCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.CreateGeneralJournalEntryCreated", opJournalEntriesCreate, v)
	}
	return &fiken.CreateGeneralJournalEntryCreated{}, nil
}

// SendInvoice implements fiken.Handler. Returns the override registered
// for 'invoices_send' if any; otherwise nil (success, no body).
func (h *handlerImpl) SendInvoice(_ context.Context, _ *fiken.SendInvoiceRequest, _ fiken.SendInvoiceParams) error {
	_, e, hit := h.server.lookup(opInvoicesSend)
	if hit {
		if e != nil {
			return e
		}
	}
	return nil
}

// CreateInvoiceCounter implements fiken.Handler. Returns the override registered
// for 'invoices_counter_create' if any; otherwise nil (success, no body).
func (h *handlerImpl) CreateInvoiceCounter(_ context.Context, _ fiken.OptCounter, _ fiken.CreateInvoiceCounterParams) error {
	_, e, hit := h.server.lookup(opInvoicesCounterCreate)
	if hit {
		if e != nil {
			return e
		}
	}
	return nil
}

// CreateInvoiceDraft implements fiken.Handler. Returns the override registered
// for 'invoices_drafts_create' if any; otherwise a zero-value *fiken.CreateInvoiceDraftCreated.
func (h *handlerImpl) CreateInvoiceDraft(_ context.Context, _ *fiken.InvoiceishDraftRequest, _ fiken.CreateInvoiceDraftParams) (*fiken.CreateInvoiceDraftCreated, error) {
	v, e, hit := h.server.lookup(opInvoicesDraftsCreate)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.CreateInvoiceDraftCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.CreateInvoiceDraftCreated", opInvoicesDraftsCreate, v)
	}
	return &fiken.CreateInvoiceDraftCreated{}, nil
}

// UpdateInvoiceDraft implements fiken.Handler. Returns the override registered
// for 'invoices_drafts_update' if any; otherwise a zero-value *fiken.UpdateInvoiceDraftCreated.
func (h *handlerImpl) UpdateInvoiceDraft(_ context.Context, _ *fiken.InvoiceishDraftRequest, _ fiken.UpdateInvoiceDraftParams) (*fiken.UpdateInvoiceDraftCreated, error) {
	v, e, hit := h.server.lookup(opInvoicesDraftsUpdate)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.UpdateInvoiceDraftCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.UpdateInvoiceDraftCreated", opInvoicesDraftsUpdate, v)
	}
	return &fiken.UpdateInvoiceDraftCreated{}, nil
}

// DeleteInvoiceDraft implements fiken.Handler. Returns the override registered
// for 'invoices_drafts_delete' if any; otherwise nil (success, no body).
func (h *handlerImpl) DeleteInvoiceDraft(_ context.Context, _ fiken.DeleteInvoiceDraftParams) error {
	_, e, hit := h.server.lookup(opInvoicesDraftsDelete)
	if hit {
		if e != nil {
			return e
		}
	}
	return nil
}

// CreateInvoiceFromDraft implements fiken.Handler. Returns the override registered
// for 'invoices_drafts_create_from' if any; otherwise a zero-value *fiken.CreateInvoiceFromDraftCreated.
func (h *handlerImpl) CreateInvoiceFromDraft(_ context.Context, _ fiken.CreateInvoiceFromDraftParams) (*fiken.CreateInvoiceFromDraftCreated, error) {
	v, e, hit := h.server.lookup(opInvoicesDraftsCreateFrom)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.CreateInvoiceFromDraftCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.CreateInvoiceFromDraftCreated", opInvoicesDraftsCreateFrom, v)
	}
	return &fiken.CreateInvoiceFromDraftCreated{}, nil
}

// SendCreditNote implements fiken.Handler. Returns the override registered
// for 'credit_notes_send' if any; otherwise nil (success, no body).
func (h *handlerImpl) SendCreditNote(_ context.Context, _ *fiken.SendCreditNoteRequest, _ fiken.SendCreditNoteParams) error {
	_, e, hit := h.server.lookup(opCreditNotesSend)
	if hit {
		if e != nil {
			return e
		}
	}
	return nil
}

// CreateCreditNoteCounter implements fiken.Handler. Returns the override registered
// for 'credit_notes_counter_create' if any; otherwise nil (success, no body).
func (h *handlerImpl) CreateCreditNoteCounter(_ context.Context, _ fiken.OptCounter, _ fiken.CreateCreditNoteCounterParams) error {
	_, e, hit := h.server.lookup(opCreditNotesCounterCreate)
	if hit {
		if e != nil {
			return e
		}
	}
	return nil
}

// CreateFullCreditNote implements fiken.Handler. Returns the override registered
// for 'credit_notes_full_create' if any; otherwise a zero-value *fiken.CreateFullCreditNoteCreated.
func (h *handlerImpl) CreateFullCreditNote(_ context.Context, _ *fiken.FullCreditNoteRequest, _ fiken.CreateFullCreditNoteParams) (*fiken.CreateFullCreditNoteCreated, error) {
	v, e, hit := h.server.lookup(opCreditNotesFullCreate)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.CreateFullCreditNoteCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.CreateFullCreditNoteCreated", opCreditNotesFullCreate, v)
	}
	return &fiken.CreateFullCreditNoteCreated{}, nil
}

// CreatePartialCreditNote implements fiken.Handler. Returns the override registered
// for 'credit_notes_partial_create' if any; otherwise a zero-value *fiken.CreatePartialCreditNoteCreated.
func (h *handlerImpl) CreatePartialCreditNote(_ context.Context, _ *fiken.PartialCreditNoteRequest, _ fiken.CreatePartialCreditNoteParams) (*fiken.CreatePartialCreditNoteCreated, error) {
	v, e, hit := h.server.lookup(opCreditNotesPartialCreate)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.CreatePartialCreditNoteCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.CreatePartialCreditNoteCreated", opCreditNotesPartialCreate, v)
	}
	return &fiken.CreatePartialCreditNoteCreated{}, nil
}

// CreateCreditNoteDraft implements fiken.Handler. Returns the override registered
// for 'credit_notes_drafts_create' if any; otherwise a zero-value *fiken.CreateCreditNoteDraftCreated.
func (h *handlerImpl) CreateCreditNoteDraft(_ context.Context, _ *fiken.InvoiceishDraftRequest, _ fiken.CreateCreditNoteDraftParams) (*fiken.CreateCreditNoteDraftCreated, error) {
	v, e, hit := h.server.lookup(opCreditNotesDraftsCreate)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.CreateCreditNoteDraftCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.CreateCreditNoteDraftCreated", opCreditNotesDraftsCreate, v)
	}
	return &fiken.CreateCreditNoteDraftCreated{}, nil
}

// UpdateCreditNoteDraft implements fiken.Handler. Returns the override registered
// for 'credit_notes_drafts_update' if any; otherwise a zero-value *fiken.UpdateCreditNoteDraftCreated.
func (h *handlerImpl) UpdateCreditNoteDraft(_ context.Context, _ *fiken.InvoiceishDraftRequest, _ fiken.UpdateCreditNoteDraftParams) (*fiken.UpdateCreditNoteDraftCreated, error) {
	v, e, hit := h.server.lookup(opCreditNotesDraftsUpdate)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.UpdateCreditNoteDraftCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.UpdateCreditNoteDraftCreated", opCreditNotesDraftsUpdate, v)
	}
	return &fiken.UpdateCreditNoteDraftCreated{}, nil
}

// DeleteCreditNoteDraft implements fiken.Handler. Returns the override registered
// for 'credit_notes_drafts_delete' if any; otherwise nil (success, no body).
func (h *handlerImpl) DeleteCreditNoteDraft(_ context.Context, _ fiken.DeleteCreditNoteDraftParams) error {
	_, e, hit := h.server.lookup(opCreditNotesDraftsDelete)
	if hit {
		if e != nil {
			return e
		}
	}
	return nil
}

// CreateCreditNoteFromDraft implements fiken.Handler. Returns the override registered
// for 'credit_notes_drafts_create_from' if any; otherwise a zero-value *fiken.CreateCreditNoteFromDraftCreated.
func (h *handlerImpl) CreateCreditNoteFromDraft(_ context.Context, _ fiken.CreateCreditNoteFromDraftParams) (*fiken.CreateCreditNoteFromDraftCreated, error) {
	v, e, hit := h.server.lookup(opCreditNotesDraftsCreateFrom)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.CreateCreditNoteFromDraftCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.CreateCreditNoteFromDraftCreated", opCreditNotesDraftsCreateFrom, v)
	}
	return &fiken.CreateCreditNoteFromDraftCreated{}, nil
}

// SendOffer implements fiken.Handler. Returns the override registered
// for 'offers_send' if any; otherwise nil (success, no body).
func (h *handlerImpl) SendOffer(_ context.Context, _ *fiken.SendOfferRequest, _ fiken.SendOfferParams) error {
	_, e, hit := h.server.lookup(opOffersSend)
	if hit {
		if e != nil {
			return e
		}
	}
	return nil
}

// CreateOfferCounter implements fiken.Handler. Returns the override registered
// for 'offers_counter_create' if any; otherwise nil (success, no body).
func (h *handlerImpl) CreateOfferCounter(_ context.Context, _ fiken.OptCounter, _ fiken.CreateOfferCounterParams) error {
	_, e, hit := h.server.lookup(opOffersCounterCreate)
	if hit {
		if e != nil {
			return e
		}
	}
	return nil
}

// CreateOfferDraft implements fiken.Handler. Returns the override registered
// for 'offers_drafts_create' if any; otherwise a zero-value *fiken.CreateOfferDraftCreated.
func (h *handlerImpl) CreateOfferDraft(_ context.Context, _ *fiken.InvoiceishDraftRequest, _ fiken.CreateOfferDraftParams) (*fiken.CreateOfferDraftCreated, error) {
	v, e, hit := h.server.lookup(opOffersDraftsCreate)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.CreateOfferDraftCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.CreateOfferDraftCreated", opOffersDraftsCreate, v)
	}
	return &fiken.CreateOfferDraftCreated{}, nil
}

// UpdateOfferDraft implements fiken.Handler. Returns the override registered
// for 'offers_drafts_update' if any; otherwise a zero-value *fiken.UpdateOfferDraftCreated.
func (h *handlerImpl) UpdateOfferDraft(_ context.Context, _ *fiken.InvoiceishDraftRequest, _ fiken.UpdateOfferDraftParams) (*fiken.UpdateOfferDraftCreated, error) {
	v, e, hit := h.server.lookup(opOffersDraftsUpdate)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.UpdateOfferDraftCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.UpdateOfferDraftCreated", opOffersDraftsUpdate, v)
	}
	return &fiken.UpdateOfferDraftCreated{}, nil
}

// DeleteOfferDraft implements fiken.Handler. Returns the override registered
// for 'offers_drafts_delete' if any; otherwise nil (success, no body).
func (h *handlerImpl) DeleteOfferDraft(_ context.Context, _ fiken.DeleteOfferDraftParams) error {
	_, e, hit := h.server.lookup(opOffersDraftsDelete)
	if hit {
		if e != nil {
			return e
		}
	}
	return nil
}

// CreateOfferFromDraft implements fiken.Handler. Returns the override registered
// for 'offers_drafts_create_from' if any; otherwise a zero-value *fiken.CreateOfferFromDraftCreated.
func (h *handlerImpl) CreateOfferFromDraft(_ context.Context, _ fiken.CreateOfferFromDraftParams) (*fiken.CreateOfferFromDraftCreated, error) {
	v, e, hit := h.server.lookup(opOffersDraftsCreateFrom)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.CreateOfferFromDraftCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.CreateOfferFromDraftCreated", opOffersDraftsCreateFrom, v)
	}
	return &fiken.CreateOfferFromDraftCreated{}, nil
}

// CreateOrderConfirmationCounter implements fiken.Handler. Returns the override registered
// for 'order_confirmations_counter_create' if any; otherwise nil (success, no body).
func (h *handlerImpl) CreateOrderConfirmationCounter(_ context.Context, _ fiken.OptCounter, _ fiken.CreateOrderConfirmationCounterParams) error {
	_, e, hit := h.server.lookup(opOrderConfirmationsCounterCreate)
	if hit {
		if e != nil {
			return e
		}
	}
	return nil
}

// CreateInvoiceDraftFromOrderConfirmation implements fiken.Handler. Returns the override registered
// for 'order_confirmations_create_invoice_draft' if any; otherwise a zero-value *fiken.CreateInvoiceDraftFromOrderConfirmationCreated.
func (h *handlerImpl) CreateInvoiceDraftFromOrderConfirmation(_ context.Context, _ fiken.CreateInvoiceDraftFromOrderConfirmationParams) (*fiken.CreateInvoiceDraftFromOrderConfirmationCreated, error) {
	v, e, hit := h.server.lookup(opOrderConfirmationsCreateInvoiceDraft)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.CreateInvoiceDraftFromOrderConfirmationCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.CreateInvoiceDraftFromOrderConfirmationCreated", opOrderConfirmationsCreateInvoiceDraft, v)
	}
	return &fiken.CreateInvoiceDraftFromOrderConfirmationCreated{}, nil
}

// CreateOrderConfirmationDraft implements fiken.Handler. Returns the override registered
// for 'order_confirmations_drafts_create' if any; otherwise a zero-value *fiken.CreateOrderConfirmationDraftCreated.
func (h *handlerImpl) CreateOrderConfirmationDraft(_ context.Context, _ *fiken.InvoiceishDraftRequest, _ fiken.CreateOrderConfirmationDraftParams) (*fiken.CreateOrderConfirmationDraftCreated, error) {
	v, e, hit := h.server.lookup(opOrderConfirmationsDraftsCreate)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.CreateOrderConfirmationDraftCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.CreateOrderConfirmationDraftCreated", opOrderConfirmationsDraftsCreate, v)
	}
	return &fiken.CreateOrderConfirmationDraftCreated{}, nil
}

// UpdateOrderConfirmationDraft implements fiken.Handler. Returns the override registered
// for 'order_confirmations_drafts_update' if any; otherwise a zero-value *fiken.UpdateOrderConfirmationDraftCreated.
func (h *handlerImpl) UpdateOrderConfirmationDraft(_ context.Context, _ *fiken.InvoiceishDraftRequest, _ fiken.UpdateOrderConfirmationDraftParams) (*fiken.UpdateOrderConfirmationDraftCreated, error) {
	v, e, hit := h.server.lookup(opOrderConfirmationsDraftsUpdate)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.UpdateOrderConfirmationDraftCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.UpdateOrderConfirmationDraftCreated", opOrderConfirmationsDraftsUpdate, v)
	}
	return &fiken.UpdateOrderConfirmationDraftCreated{}, nil
}

// DeleteOrderConfirmationDraft implements fiken.Handler. Returns the override registered
// for 'order_confirmations_drafts_delete' if any; otherwise nil (success, no body).
func (h *handlerImpl) DeleteOrderConfirmationDraft(_ context.Context, _ fiken.DeleteOrderConfirmationDraftParams) error {
	_, e, hit := h.server.lookup(opOrderConfirmationsDraftsDelete)
	if hit {
		if e != nil {
			return e
		}
	}
	return nil
}

// CreateOrderConfirmationFromDraft implements fiken.Handler. Returns the override registered
// for 'order_confirmations_drafts_create_from' if any; otherwise a zero-value *fiken.CreateOrderConfirmationFromDraftCreated.
func (h *handlerImpl) CreateOrderConfirmationFromDraft(_ context.Context, _ fiken.CreateOrderConfirmationFromDraftParams) (*fiken.CreateOrderConfirmationFromDraftCreated, error) {
	v, e, hit := h.server.lookup(opOrderConfirmationsDraftsCreateFrom)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.CreateOrderConfirmationFromDraftCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.CreateOrderConfirmationFromDraftCreated", opOrderConfirmationsDraftsCreateFrom, v)
	}
	return &fiken.CreateOrderConfirmationFromDraftCreated{}, nil
}

// CreateProduct implements fiken.Handler. Returns the override registered
// for 'products_create' if any; otherwise a zero-value *fiken.CreateProductCreated.
func (h *handlerImpl) CreateProduct(_ context.Context, _ *fiken.Product, _ fiken.CreateProductParams) (*fiken.CreateProductCreated, error) {
	v, e, hit := h.server.lookup(opProductsCreate)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.CreateProductCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.CreateProductCreated", opProductsCreate, v)
	}
	return &fiken.CreateProductCreated{}, nil
}

// UpdateProduct implements fiken.Handler. Returns the override registered
// for 'products_update' if any; otherwise a zero-value *fiken.UpdateProductOK.
func (h *handlerImpl) UpdateProduct(_ context.Context, _ *fiken.Product, _ fiken.UpdateProductParams) (*fiken.UpdateProductOK, error) {
	v, e, hit := h.server.lookup(opProductsUpdate)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.UpdateProductOK); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.UpdateProductOK", opProductsUpdate, v)
	}
	return &fiken.UpdateProductOK{}, nil
}

// DeleteProduct implements fiken.Handler. Returns the override registered
// for 'products_delete' if any; otherwise nil (success, no body).
func (h *handlerImpl) DeleteProduct(_ context.Context, _ fiken.DeleteProductParams) error {
	_, e, hit := h.server.lookup(opProductsDelete)
	if hit {
		if e != nil {
			return e
		}
	}
	return nil
}

// CreateSale implements fiken.Handler. Returns the override registered
// for 'sales_create' if any; otherwise a zero-value *fiken.CreateSaleCreated.
func (h *handlerImpl) CreateSale(_ context.Context, _ *fiken.SaleRequest, _ fiken.CreateSaleParams) (*fiken.CreateSaleCreated, error) {
	v, e, hit := h.server.lookup(opSalesCreate)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.CreateSaleCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.CreateSaleCreated", opSalesCreate, v)
	}
	return &fiken.CreateSaleCreated{}, nil
}

// DeleteSale implements fiken.Handler. Returns the override registered
// for 'sales_delete' if any; otherwise a zero-value *fiken.SaleResult.
func (h *handlerImpl) DeleteSale(_ context.Context, _ fiken.DeleteSaleParams) (*fiken.SaleResult, error) {
	v, e, hit := h.server.lookup(opSalesDelete)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.SaleResult); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.SaleResult", opSalesDelete, v)
	}
	return &fiken.SaleResult{}, nil
}

// SettledSale implements fiken.Handler. Returns the override registered
// for 'sales_settle' if any; otherwise a zero-value *fiken.SaleResult.
func (h *handlerImpl) SettledSale(_ context.Context, _ fiken.SettledSaleParams) (*fiken.SaleResult, error) {
	v, e, hit := h.server.lookup(opSalesSettle)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.SaleResult); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.SaleResult", opSalesSettle, v)
	}
	return &fiken.SaleResult{}, nil
}

// WriteOffSale implements fiken.Handler. Returns the override registered
// for 'sales_write_off' if any; otherwise a zero-value *fiken.SaleResult.
func (h *handlerImpl) WriteOffSale(_ context.Context, _ *fiken.WriteOffRequest, _ fiken.WriteOffSaleParams) (*fiken.SaleResult, error) {
	v, e, hit := h.server.lookup(opSalesWriteOff)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.SaleResult); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.SaleResult", opSalesWriteOff, v)
	}
	return &fiken.SaleResult{}, nil
}

// CreateSalePayment implements fiken.Handler. Returns the override registered
// for 'sales_payments_create' if any; otherwise a zero-value *fiken.CreateSalePaymentCreated.
func (h *handlerImpl) CreateSalePayment(_ context.Context, _ *fiken.Payment, _ fiken.CreateSalePaymentParams) (*fiken.CreateSalePaymentCreated, error) {
	v, e, hit := h.server.lookup(opSalesPaymentsCreate)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.CreateSalePaymentCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.CreateSalePaymentCreated", opSalesPaymentsCreate, v)
	}
	return &fiken.CreateSalePaymentCreated{}, nil
}

// CreatePurchase implements fiken.Handler. Returns the override registered
// for 'purchases_create' if any; otherwise a zero-value *fiken.CreatePurchaseCreated.
func (h *handlerImpl) CreatePurchase(_ context.Context, _ *fiken.PurchaseRequest, _ fiken.CreatePurchaseParams) (*fiken.CreatePurchaseCreated, error) {
	v, e, hit := h.server.lookup(opPurchasesCreate)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.CreatePurchaseCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.CreatePurchaseCreated", opPurchasesCreate, v)
	}
	return &fiken.CreatePurchaseCreated{}, nil
}

// DeletePurchase implements fiken.Handler. Returns the override registered
// for 'purchases_delete' if any; otherwise a zero-value *fiken.PurchaseResult.
func (h *handlerImpl) DeletePurchase(_ context.Context, _ fiken.DeletePurchaseParams) (*fiken.PurchaseResult, error) {
	v, e, hit := h.server.lookup(opPurchasesDelete)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.PurchaseResult); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.PurchaseResult", opPurchasesDelete, v)
	}
	return &fiken.PurchaseResult{}, nil
}

// CreatePurchasePayment implements fiken.Handler. Returns the override registered
// for 'purchases_payments_create' if any; otherwise a zero-value *fiken.CreatePurchasePaymentCreated.
func (h *handlerImpl) CreatePurchasePayment(_ context.Context, _ *fiken.Payment, _ fiken.CreatePurchasePaymentParams) (*fiken.CreatePurchasePaymentCreated, error) {
	v, e, hit := h.server.lookup(opPurchasesPaymentsCreate)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.CreatePurchasePaymentCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.CreatePurchasePaymentCreated", opPurchasesPaymentsCreate, v)
	}
	return &fiken.CreatePurchasePaymentCreated{}, nil
}

// CreateInboxDocument implements fiken.Handler. Returns the override registered
// for 'inbox_send' if any; otherwise a zero-value *fiken.CreateInboxDocumentCreated.
func (h *handlerImpl) CreateInboxDocument(_ context.Context, _ *fiken.CreateInboxDocumentReq, _ fiken.CreateInboxDocumentParams) (*fiken.CreateInboxDocumentCreated, error) {
	v, e, hit := h.server.lookup(opInboxSend)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.CreateInboxDocumentCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.CreateInboxDocumentCreated", opInboxSend, v)
	}
	return &fiken.CreateInboxDocumentCreated{}, nil
}

// CreateProject implements fiken.Handler. Returns the override registered
// for 'projects_create' if any; otherwise a zero-value *fiken.CreateProjectCreated.
func (h *handlerImpl) CreateProject(_ context.Context, _ *fiken.ProjectRequest, _ fiken.CreateProjectParams) (*fiken.CreateProjectCreated, error) {
	v, e, hit := h.server.lookup(opProjectsCreate)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.CreateProjectCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.CreateProjectCreated", opProjectsCreate, v)
	}
	return &fiken.CreateProjectCreated{}, nil
}

// UpdateProject implements fiken.Handler. Returns the override registered
// for 'projects_update' if any; otherwise a zero-value *fiken.UpdateProjectCreated.
func (h *handlerImpl) UpdateProject(_ context.Context, _ *fiken.UpdateProjectRequest, _ fiken.UpdateProjectParams) (*fiken.UpdateProjectCreated, error) {
	v, e, hit := h.server.lookup(opProjectsUpdate)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.UpdateProjectCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.UpdateProjectCreated", opProjectsUpdate, v)
	}
	return &fiken.UpdateProjectCreated{}, nil
}

// DeleteProject implements fiken.Handler. Returns the override registered
// for 'projects_delete' if any; otherwise nil (success, no body).
func (h *handlerImpl) DeleteProject(_ context.Context, _ fiken.DeleteProjectParams) error {
	_, e, hit := h.server.lookup(opProjectsDelete)
	if hit {
		if e != nil {
			return e
		}
	}
	return nil
}

// CreateActivity implements fiken.Handler. Returns the override registered
// for 'activities_create' if any; otherwise a zero-value *fiken.CreateActivityCreated.
func (h *handlerImpl) CreateActivity(_ context.Context, _ *fiken.ActivityRequest, _ fiken.CreateActivityParams) (*fiken.CreateActivityCreated, error) {
	v, e, hit := h.server.lookup(opActivitiesCreate)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.CreateActivityCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.CreateActivityCreated", opActivitiesCreate, v)
	}
	return &fiken.CreateActivityCreated{}, nil
}

// UpdateActivity implements fiken.Handler. Returns the override registered
// for 'activities_update' if any; otherwise a zero-value *fiken.UpdateActivityOK.
func (h *handlerImpl) UpdateActivity(_ context.Context, _ *fiken.UpdateActivityRequest, _ fiken.UpdateActivityParams) (*fiken.UpdateActivityOK, error) {
	v, e, hit := h.server.lookup(opActivitiesUpdate)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.UpdateActivityOK); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.UpdateActivityOK", opActivitiesUpdate, v)
	}
	return &fiken.UpdateActivityOK{}, nil
}

// DeleteActivity implements fiken.Handler. Returns the override registered
// for 'activities_delete' if any; otherwise nil (success, no body).
func (h *handlerImpl) DeleteActivity(_ context.Context, _ fiken.DeleteActivityParams) error {
	_, e, hit := h.server.lookup(opActivitiesDelete)
	if hit {
		if e != nil {
			return e
		}
	}
	return nil
}

// CreateTimeEntry implements fiken.Handler. Returns the override registered
// for 'time_entries_create' if any; otherwise a zero-value *fiken.CreateTimeEntryCreated.
func (h *handlerImpl) CreateTimeEntry(_ context.Context, _ *fiken.TimeEntryRequest, _ fiken.CreateTimeEntryParams) (*fiken.CreateTimeEntryCreated, error) {
	v, e, hit := h.server.lookup(opTimeEntriesCreate)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.CreateTimeEntryCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.CreateTimeEntryCreated", opTimeEntriesCreate, v)
	}
	return &fiken.CreateTimeEntryCreated{}, nil
}

// UpdateTimeEntry implements fiken.Handler. Returns the override registered
// for 'time_entries_update' if any; otherwise a zero-value *fiken.UpdateTimeEntryOK.
func (h *handlerImpl) UpdateTimeEntry(_ context.Context, _ *fiken.UpdateTimeEntryRequest, _ fiken.UpdateTimeEntryParams) (*fiken.UpdateTimeEntryOK, error) {
	v, e, hit := h.server.lookup(opTimeEntriesUpdate)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.UpdateTimeEntryOK); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.UpdateTimeEntryOK", opTimeEntriesUpdate, v)
	}
	return &fiken.UpdateTimeEntryOK{}, nil
}

// DeleteTimeEntry implements fiken.Handler. Returns the override registered
// for 'time_entries_delete' if any; otherwise nil (no-body NoContent path).
func (h *handlerImpl) DeleteTimeEntry(_ context.Context, _ fiken.DeleteTimeEntryParams) (fiken.DeleteTimeEntryRes, error) {
	v, e, hit := h.server.lookup(opTimeEntriesDelete)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(fiken.DeleteTimeEntryRes); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want fiken.DeleteTimeEntryRes", opTimeEntriesDelete, v)
	}
	return nil, nil
}

// CreateInvoiceDraftFromTimeEntries implements fiken.Handler. Returns the override registered
// for 'time_entries_invoice_draft_create' if any; otherwise a zero-value *fiken.InvoiceishDraftResultHeaders.
func (h *handlerImpl) CreateInvoiceDraftFromTimeEntries(_ context.Context, _ *fiken.TimeEntryInvoiceDraftRequest, _ fiken.CreateInvoiceDraftFromTimeEntriesParams) (*fiken.InvoiceishDraftResultHeaders, error) {
	v, e, hit := h.server.lookup(opTimeEntriesInvoiceDraftFromTimes)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.InvoiceishDraftResultHeaders); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.InvoiceishDraftResultHeaders", opTimeEntriesInvoiceDraftFromTimes, v)
	}
	return &fiken.InvoiceishDraftResultHeaders{}, nil
}

// AddAttachmentToInvoice implements fiken.Handler. Returns the override registered
// for 'invoices_attachments_attach' if any; otherwise a zero-value *fiken.AddAttachmentToInvoiceCreated.
func (h *handlerImpl) AddAttachmentToInvoice(_ context.Context, _ fiken.OptAddAttachmentToInvoiceReq, _ fiken.AddAttachmentToInvoiceParams) (*fiken.AddAttachmentToInvoiceCreated, error) {
	v, e, hit := h.server.lookup(opInvoicesAttachmentsAttach)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.AddAttachmentToInvoiceCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.AddAttachmentToInvoiceCreated", opInvoicesAttachmentsAttach, v)
	}
	return &fiken.AddAttachmentToInvoiceCreated{}, nil
}

// AddAttachmentToJournalEntry implements fiken.Handler. Returns the override registered
// for 'journal_entries_attachments_attach' if any; otherwise a zero-value *fiken.AddAttachmentToJournalEntryCreated.
func (h *handlerImpl) AddAttachmentToJournalEntry(_ context.Context, _ fiken.OptAddAttachmentToJournalEntryReq, _ fiken.AddAttachmentToJournalEntryParams) (*fiken.AddAttachmentToJournalEntryCreated, error) {
	v, e, hit := h.server.lookup(opJournalEntriesAttachmentsAttach)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.AddAttachmentToJournalEntryCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.AddAttachmentToJournalEntryCreated", opJournalEntriesAttachmentsAttach, v)
	}
	return &fiken.AddAttachmentToJournalEntryCreated{}, nil
}

// AddAttachmentToInvoiceDraft implements fiken.Handler. Returns the override registered
// for 'invoices_drafts_attachments_attach' if any; otherwise a zero-value *fiken.AddAttachmentToInvoiceDraftCreated.
func (h *handlerImpl) AddAttachmentToInvoiceDraft(_ context.Context, _ fiken.OptAddAttachmentToInvoiceDraftReq, _ fiken.AddAttachmentToInvoiceDraftParams) (*fiken.AddAttachmentToInvoiceDraftCreated, error) {
	v, e, hit := h.server.lookup(opInvoicesDraftsAttachmentsAttach)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.AddAttachmentToInvoiceDraftCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.AddAttachmentToInvoiceDraftCreated", opInvoicesDraftsAttachmentsAttach, v)
	}
	return &fiken.AddAttachmentToInvoiceDraftCreated{}, nil
}

// AddAttachmentToCreditNoteDraft implements fiken.Handler. Returns the override registered
// for 'credit_notes_drafts_attachments_attach' if any; otherwise a zero-value *fiken.AddAttachmentToCreditNoteDraftCreated.
func (h *handlerImpl) AddAttachmentToCreditNoteDraft(_ context.Context, _ fiken.OptAddAttachmentToCreditNoteDraftReq, _ fiken.AddAttachmentToCreditNoteDraftParams) (*fiken.AddAttachmentToCreditNoteDraftCreated, error) {
	v, e, hit := h.server.lookup(opCreditNotesDraftsAttachmentsAttach)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.AddAttachmentToCreditNoteDraftCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.AddAttachmentToCreditNoteDraftCreated", opCreditNotesDraftsAttachmentsAttach, v)
	}
	return &fiken.AddAttachmentToCreditNoteDraftCreated{}, nil
}

// AddAttachmentToOfferDraft implements fiken.Handler. Returns the override registered
// for 'offers_drafts_attachments_attach' if any; otherwise a zero-value *fiken.AddAttachmentToOfferDraftCreated.
func (h *handlerImpl) AddAttachmentToOfferDraft(_ context.Context, _ fiken.OptAddAttachmentToOfferDraftReq, _ fiken.AddAttachmentToOfferDraftParams) (*fiken.AddAttachmentToOfferDraftCreated, error) {
	v, e, hit := h.server.lookup(opOffersDraftsAttachmentsAttach)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.AddAttachmentToOfferDraftCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.AddAttachmentToOfferDraftCreated", opOffersDraftsAttachmentsAttach, v)
	}
	return &fiken.AddAttachmentToOfferDraftCreated{}, nil
}

// AddAttachmentToOrderConfirmationDraft implements fiken.Handler. Returns the override registered
// for 'order_confirmations_drafts_attachments_attach' if any; otherwise a zero-value *fiken.AddAttachmentToOrderConfirmationDraftCreated.
func (h *handlerImpl) AddAttachmentToOrderConfirmationDraft(_ context.Context, _ fiken.OptAddAttachmentToOrderConfirmationDraftReq, _ fiken.AddAttachmentToOrderConfirmationDraftParams) (*fiken.AddAttachmentToOrderConfirmationDraftCreated, error) {
	v, e, hit := h.server.lookup(opOrderConfirmationsDraftsAttachmentsAttach)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.AddAttachmentToOrderConfirmationDraftCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.AddAttachmentToOrderConfirmationDraftCreated", opOrderConfirmationsDraftsAttachmentsAttach, v)
	}
	return &fiken.AddAttachmentToOrderConfirmationDraftCreated{}, nil
}

// AddAttachmentToSale implements fiken.Handler. Returns the override registered
// for 'sales_attachments_attach' if any; otherwise a zero-value *fiken.AddAttachmentToSaleCreated.
func (h *handlerImpl) AddAttachmentToSale(_ context.Context, _ fiken.OptAddAttachmentToSaleReq, _ fiken.AddAttachmentToSaleParams) (*fiken.AddAttachmentToSaleCreated, error) {
	v, e, hit := h.server.lookup(opSalesAttach)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.AddAttachmentToSaleCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.AddAttachmentToSaleCreated", opSalesAttach, v)
	}
	return &fiken.AddAttachmentToSaleCreated{}, nil
}

// AddAttachmentToPurchase implements fiken.Handler. Returns the override registered
// for 'purchases_attachments_attach' if any; otherwise a zero-value *fiken.AddAttachmentToPurchaseCreated.
func (h *handlerImpl) AddAttachmentToPurchase(_ context.Context, _ fiken.OptAddAttachmentToPurchaseReq, _ fiken.AddAttachmentToPurchaseParams) (*fiken.AddAttachmentToPurchaseCreated, error) {
	v, e, hit := h.server.lookup(opPurchasesAttach)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.AddAttachmentToPurchaseCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.AddAttachmentToPurchaseCreated", opPurchasesAttach, v)
	}
	return &fiken.AddAttachmentToPurchaseCreated{}, nil
}

// === Plan D / 21-op tail handlers ===

// CreateInvoice implements fiken.Handler. Returns the override registered
// for 'invoices_create' if any; otherwise a zero-value *fiken.CreateInvoiceCreated.
func (h *handlerImpl) CreateInvoice(_ context.Context, _ *fiken.InvoiceRequest, _ fiken.CreateInvoiceParams) (*fiken.CreateInvoiceCreated, error) {
	v, e, hit := h.server.lookup(opInvoicesCreate)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.CreateInvoiceCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.CreateInvoiceCreated", opInvoicesCreate, v)
	}
	return &fiken.CreateInvoiceCreated{}, nil
}

// UpdateInvoice implements fiken.Handler. Returns the override registered
// for 'invoices_update' if any; otherwise a zero-value *fiken.UpdateInvoiceOK.
func (h *handlerImpl) UpdateInvoice(_ context.Context, _ *fiken.UpdateInvoiceRequest, _ fiken.UpdateInvoiceParams) (*fiken.UpdateInvoiceOK, error) {
	v, e, hit := h.server.lookup(opInvoicesUpdate)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.UpdateInvoiceOK); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.UpdateInvoiceOK", opInvoicesUpdate, v)
	}
	return &fiken.UpdateInvoiceOK{}, nil
}

// GetInvoiceCounter implements fiken.Handler. Returns the override registered
// for 'invoices_counter_get' if any; otherwise a zero-value *fiken.Counter.
func (h *handlerImpl) GetInvoiceCounter(_ context.Context, _ fiken.GetInvoiceCounterParams) (*fiken.Counter, error) {
	v, e, hit := h.server.lookup(opInvoicesCounterGet)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.Counter); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.Counter", opInvoicesCounterGet, v)
	}
	return &fiken.Counter{}, nil
}

// GetOfferCounter implements fiken.Handler. Returns the override registered
// for 'offers_counter_get' if any; otherwise a zero-value *fiken.Counter.
func (h *handlerImpl) GetOfferCounter(_ context.Context, _ fiken.GetOfferCounterParams) (*fiken.Counter, error) {
	v, e, hit := h.server.lookup(opOffersCounterGet)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.Counter); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.Counter", opOffersCounterGet, v)
	}
	return &fiken.Counter{}, nil
}

// GetOrderConfirmationCounter implements fiken.Handler. Returns the override registered
// for 'order_confirmations_counter_get' if any; otherwise a zero-value *fiken.Counter.
func (h *handlerImpl) GetOrderConfirmationCounter(_ context.Context, _ fiken.GetOrderConfirmationCounterParams) (*fiken.Counter, error) {
	v, e, hit := h.server.lookup(opOrderConfirmationsCounterGet)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.Counter); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.Counter", opOrderConfirmationsCounterGet, v)
	}
	return &fiken.Counter{}, nil
}

// GetCreditNoteCounter implements fiken.Handler. Returns the override registered
// for 'credit_notes_counter_get' if any; otherwise a zero-value *fiken.Counter.
func (h *handlerImpl) GetCreditNoteCounter(_ context.Context, _ fiken.GetCreditNoteCounterParams) (*fiken.Counter, error) {
	v, e, hit := h.server.lookup(opCreditNotesCounterGet)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.Counter); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.Counter", opCreditNotesCounterGet, v)
	}
	return &fiken.Counter{}, nil
}

// DeleteTransaction implements fiken.Handler. Returns the override registered
// for 'transactions_delete' if any; otherwise a zero-value *fiken.Transaction.
func (h *handlerImpl) DeleteTransaction(_ context.Context, _ fiken.DeleteTransactionParams) (*fiken.Transaction, error) {
	v, e, hit := h.server.lookup(opTransactionsDelete)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.Transaction); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.Transaction", opTransactionsDelete, v)
	}
	return &fiken.Transaction{}, nil
}

// AddAttachmentToContact implements fiken.Handler. Returns the override registered
// for 'contacts_attachments_attach' if any; otherwise a zero-value *fiken.AddAttachmentToContactCreated.
func (h *handlerImpl) AddAttachmentToContact(_ context.Context, _ fiken.OptAddAttachmentToContactReq, _ fiken.AddAttachmentToContactParams) (*fiken.AddAttachmentToContactCreated, error) {
	v, e, hit := h.server.lookup(opContactsAttachmentsAttach)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.AddAttachmentToContactCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.AddAttachmentToContactCreated", opContactsAttachmentsAttach, v)
	}
	return &fiken.AddAttachmentToContactCreated{}, nil
}

// GetSaleDrafts implements fiken.Handler. Returns the override registered
// for 'sales_drafts_list' if any; otherwise a zero-value *fiken.GetSaleDraftsOKHeaders.
func (h *handlerImpl) GetSaleDrafts(_ context.Context, _ fiken.GetSaleDraftsParams) (*fiken.GetSaleDraftsOKHeaders, error) {
	v, e, hit := h.server.lookup(opSalesDraftsList)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.GetSaleDraftsOKHeaders); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.GetSaleDraftsOKHeaders", opSalesDraftsList, v)
	}
	return &fiken.GetSaleDraftsOKHeaders{Response: []fiken.DraftResult{}}, nil
}

// GetSaleDraft implements fiken.Handler. Returns the override registered
// for 'sales_drafts_get' if any; otherwise a zero-value *fiken.DraftResult.
func (h *handlerImpl) GetSaleDraft(_ context.Context, _ fiken.GetSaleDraftParams) (*fiken.DraftResult, error) {
	v, e, hit := h.server.lookup(opSalesDraftsGet)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.DraftResult); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.DraftResult", opSalesDraftsGet, v)
	}
	return &fiken.DraftResult{}, nil
}

// CreateSaleDraft implements fiken.Handler. Returns the override registered
// for 'sales_drafts_create' if any; otherwise a zero-value *fiken.CreateSaleDraftCreated.
func (h *handlerImpl) CreateSaleDraft(_ context.Context, _ *fiken.DraftRequest, _ fiken.CreateSaleDraftParams) (*fiken.CreateSaleDraftCreated, error) {
	v, e, hit := h.server.lookup(opSalesDraftsCreate)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.CreateSaleDraftCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.CreateSaleDraftCreated", opSalesDraftsCreate, v)
	}
	return &fiken.CreateSaleDraftCreated{}, nil
}

// UpdateSaleDraft implements fiken.Handler. Returns the override registered
// for 'sales_drafts_update' if any; otherwise a zero-value *fiken.UpdateSaleDraftCreated.
func (h *handlerImpl) UpdateSaleDraft(_ context.Context, _ *fiken.DraftRequest, _ fiken.UpdateSaleDraftParams) (*fiken.UpdateSaleDraftCreated, error) {
	v, e, hit := h.server.lookup(opSalesDraftsUpdate)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.UpdateSaleDraftCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.UpdateSaleDraftCreated", opSalesDraftsUpdate, v)
	}
	return &fiken.UpdateSaleDraftCreated{}, nil
}

// DeleteSaleDraft implements fiken.Handler. Returns the override registered
// for 'sales_drafts_delete' if any; otherwise nil (success, no body).
func (h *handlerImpl) DeleteSaleDraft(_ context.Context, _ fiken.DeleteSaleDraftParams) error {
	_, e, hit := h.server.lookup(opSalesDraftsDelete)
	if hit {
		if e != nil {
			return e
		}
	}
	return nil
}

// CreateSaleFromDraft implements fiken.Handler. Returns the override registered
// for 'sales_drafts_create_from' if any; otherwise a zero-value *fiken.CreateSaleFromDraftCreated.
func (h *handlerImpl) CreateSaleFromDraft(_ context.Context, _ fiken.CreateSaleFromDraftParams) (*fiken.CreateSaleFromDraftCreated, error) {
	v, e, hit := h.server.lookup(opSalesDraftsCreateFrom)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.CreateSaleFromDraftCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.CreateSaleFromDraftCreated", opSalesDraftsCreateFrom, v)
	}
	return &fiken.CreateSaleFromDraftCreated{}, nil
}

// GetSaleDraftAttachments implements fiken.Handler. Returns the override registered
// for 'sales_drafts_attachments_list' if any; otherwise an empty slice.
func (h *handlerImpl) GetSaleDraftAttachments(_ context.Context, _ fiken.GetSaleDraftAttachmentsParams) ([]fiken.Attachment, error) {
	v, e, hit := h.server.lookup(opSalesDraftsAttachmentsList)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.([]fiken.Attachment); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want []fiken.Attachment", opSalesDraftsAttachmentsList, v)
	}
	return []fiken.Attachment{}, nil
}

// AddAttachmentToSaleDraft implements fiken.Handler. Returns the override registered
// for 'sales_drafts_attachments_attach' if any; otherwise a zero-value response.
func (h *handlerImpl) AddAttachmentToSaleDraft(_ context.Context, _ fiken.OptAddAttachmentToSaleDraftReq, _ fiken.AddAttachmentToSaleDraftParams) (*fiken.AddAttachmentToSaleDraftCreated, error) {
	v, e, hit := h.server.lookup(opSalesDraftsAttachmentsAttach)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.AddAttachmentToSaleDraftCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.AddAttachmentToSaleDraftCreated", opSalesDraftsAttachmentsAttach, v)
	}
	return &fiken.AddAttachmentToSaleDraftCreated{}, nil
}

// GetPurchaseDrafts implements fiken.Handler. Returns the override registered
// for 'purchases_drafts_list' if any; otherwise a zero-value response.
func (h *handlerImpl) GetPurchaseDrafts(_ context.Context, _ fiken.GetPurchaseDraftsParams) (*fiken.GetPurchaseDraftsOKHeaders, error) {
	v, e, hit := h.server.lookup(opPurchasesDraftsList)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.GetPurchaseDraftsOKHeaders); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.GetPurchaseDraftsOKHeaders", opPurchasesDraftsList, v)
	}
	return &fiken.GetPurchaseDraftsOKHeaders{Response: []fiken.DraftResult{}}, nil
}

// GetPurchaseDraft implements fiken.Handler. Returns the override registered
// for 'purchases_drafts_get' if any; otherwise a zero-value *fiken.DraftResult.
func (h *handlerImpl) GetPurchaseDraft(_ context.Context, _ fiken.GetPurchaseDraftParams) (*fiken.DraftResult, error) {
	v, e, hit := h.server.lookup(opPurchasesDraftsGet)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.DraftResult); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.DraftResult", opPurchasesDraftsGet, v)
	}
	return &fiken.DraftResult{}, nil
}

// CreatePurchaseDraft implements fiken.Handler. Returns the override registered
// for 'purchases_drafts_create' if any; otherwise a zero-value response.
func (h *handlerImpl) CreatePurchaseDraft(_ context.Context, _ *fiken.DraftRequest, _ fiken.CreatePurchaseDraftParams) (*fiken.CreatePurchaseDraftCreated, error) {
	v, e, hit := h.server.lookup(opPurchasesDraftsCreate)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.CreatePurchaseDraftCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.CreatePurchaseDraftCreated", opPurchasesDraftsCreate, v)
	}
	return &fiken.CreatePurchaseDraftCreated{}, nil
}

// UpdatePurchaseDraft implements fiken.Handler. Returns the override registered
// for 'purchases_drafts_update' if any; otherwise a zero-value response.
func (h *handlerImpl) UpdatePurchaseDraft(_ context.Context, _ *fiken.DraftRequest, _ fiken.UpdatePurchaseDraftParams) (*fiken.UpdatePurchaseDraftCreated, error) {
	v, e, hit := h.server.lookup(opPurchasesDraftsUpdate)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.UpdatePurchaseDraftCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.UpdatePurchaseDraftCreated", opPurchasesDraftsUpdate, v)
	}
	return &fiken.UpdatePurchaseDraftCreated{}, nil
}

// DeletePurchaseDraft implements fiken.Handler. Returns the override registered
// for 'purchases_drafts_delete' if any; otherwise nil (success).
func (h *handlerImpl) DeletePurchaseDraft(_ context.Context, _ fiken.DeletePurchaseDraftParams) error {
	_, e, hit := h.server.lookup(opPurchasesDraftsDelete)
	if hit {
		if e != nil {
			return e
		}
	}
	return nil
}

// CreatePurchaseFromDraft implements fiken.Handler. Returns the override registered
// for 'purchases_drafts_create_from' if any; otherwise a zero-value response.
func (h *handlerImpl) CreatePurchaseFromDraft(_ context.Context, _ fiken.CreatePurchaseFromDraftParams) (*fiken.CreatePurchaseFromDraftCreated, error) {
	v, e, hit := h.server.lookup(opPurchasesDraftsCreateFrom)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.CreatePurchaseFromDraftCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.CreatePurchaseFromDraftCreated", opPurchasesDraftsCreateFrom, v)
	}
	return &fiken.CreatePurchaseFromDraftCreated{}, nil
}

// GetPurchaseDraftAttachments implements fiken.Handler. Returns the override registered
// for 'purchases_drafts_attachments_list' if any; otherwise an empty slice.
func (h *handlerImpl) GetPurchaseDraftAttachments(_ context.Context, _ fiken.GetPurchaseDraftAttachmentsParams) ([]fiken.Attachment, error) {
	v, e, hit := h.server.lookup(opPurchasesDraftsAttachmentsList)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.([]fiken.Attachment); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want []fiken.Attachment", opPurchasesDraftsAttachmentsList, v)
	}
	return []fiken.Attachment{}, nil
}

// AddAttachmentToPurchaseDraft implements fiken.Handler. Returns the override registered
// for 'purchases_drafts_attachments_attach' if any; otherwise a zero-value response.
func (h *handlerImpl) AddAttachmentToPurchaseDraft(_ context.Context, _ fiken.OptAddAttachmentToPurchaseDraftReq, _ fiken.AddAttachmentToPurchaseDraftParams) (*fiken.AddAttachmentToPurchaseDraftCreated, error) {
	v, e, hit := h.server.lookup(opPurchasesDraftsAttachmentsAttach)
	if hit {
		if e != nil {
			return nil, e
		}
		if resp, ok := v.(*fiken.AddAttachmentToPurchaseDraftCreated); ok {
			return resp, nil
		}
		h.server.t.Fatalf("mockfiken: Set(%s) got %T, want *fiken.AddAttachmentToPurchaseDraftCreated", opPurchasesDraftsAttachmentsAttach, v)
	}
	return &fiken.AddAttachmentToPurchaseDraftCreated{}, nil
}
