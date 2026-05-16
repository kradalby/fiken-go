package ops

// Op*: stable op-name constants. The string value is the canonical
// shared name used by:
//   - the CLI subcommand path (e.g. "companies_list" ↔ `fiken companies list`)
//   - the MCP tool name in tools/list
//   - the i18n key prefix (e.g. ops.companies_list.summary)
//   - the key into Registry
//
// The value is INDEPENDENT of the upstream OpenAPI operationId; see
// RegistryEntry.OperationID for the OAS-side name (used to query
// IsMutating, which is sourced from the spec's HTTP method).
const (
	OpCompaniesList = "companies_list"
	OpCompaniesGet  = "companies_get"

	OpContactsList              = "contacts_list"
	OpContactsGet               = "contacts_get"
	OpContactsCreate            = "contacts_create"
	OpContactsUpdate            = "contacts_update"
	OpContactsDelete            = "contacts_delete"
	OpContactsAttachmentsAttach = "contacts_attachments_attach"

	OpContactsPersonsList   = "contacts_persons_list"
	OpContactsPersonsGet    = "contacts_persons_get"
	OpContactsPersonsCreate = "contacts_persons_create"
	OpContactsPersonsUpdate = "contacts_persons_update"
	OpContactsPersonsDelete = "contacts_persons_delete"

	OpAccountsList = "accounts_list"
	OpAccountsGet  = "accounts_get"

	OpBankAccountsList   = "bank_accounts_list"
	OpBankAccountsGet    = "bank_accounts_get"
	OpBankAccountsCreate = "bank_accounts_create"

	OpJournalEntriesList              = "journal_entries_list"
	OpJournalEntriesGet               = "journal_entries_get"
	OpJournalEntriesCreate            = "journal_entries_create"
	OpJournalEntriesAttachmentsList   = "journal_entries_attachments_list"
	OpJournalEntriesAttachmentsAttach = "journal_entries_attachments_attach"

	OpTransactionsList   = "transactions_list"
	OpTransactionsGet    = "transactions_get"
	OpTransactionsDelete = "transactions_delete"

	OpInvoicesList          = "invoices_list"
	OpInvoicesGet           = "invoices_get"
	OpInvoicesCreate        = "invoices_create"
	OpInvoicesUpdate        = "invoices_update"
	OpInvoicesSend          = "invoices_send"
	OpInvoicesCounterCreate = "invoices_counter_create"
	OpInvoicesCounterGet    = "invoices_counter_get"

	OpInvoicesDraftsList       = "invoices_drafts_list"
	OpInvoicesDraftsGet        = "invoices_drafts_get"
	OpInvoicesDraftsCreate     = "invoices_drafts_create"
	OpInvoicesDraftsUpdate     = "invoices_drafts_update"
	OpInvoicesDraftsDelete     = "invoices_drafts_delete"
	OpInvoicesDraftsCreateFrom = "invoices_drafts_create_from"

	OpInvoicesAttachmentsList         = "invoices_attachments_list"
	OpInvoicesAttachmentsAttach       = "invoices_attachments_attach"
	OpInvoicesDraftsAttachmentsList   = "invoices_drafts_attachments_list"
	OpInvoicesDraftsAttachmentsAttach = "invoices_drafts_attachments_attach"

	// gosec G101 flags some of these op-name constants as possible
	// hardcoded credentials (the heuristic trips on the "Credit"
	// substring + "Create" suffix combinations). They are op names,
	// not secrets — silence the false positives.
	OpCreditNotesList          = "credit_notes_list"
	OpCreditNotesGet           = "credit_notes_get"
	OpCreditNotesSend          = "credit_notes_send"
	OpCreditNotesCounterCreate = "credit_notes_counter_create" //nolint:gosec // op name, not a credential
	OpCreditNotesCounterGet    = "credit_notes_counter_get"    //nolint:gosec // op name, not a credential
	OpCreditNotesFullCreate    = "credit_notes_full_create"    //nolint:gosec // op name, not a credential
	OpCreditNotesPartialCreate = "credit_notes_partial_create"

	OpCreditNotesDraftsList       = "credit_notes_drafts_list"
	OpCreditNotesDraftsGet        = "credit_notes_drafts_get"
	OpCreditNotesDraftsCreate     = "credit_notes_drafts_create"
	OpCreditNotesDraftsUpdate     = "credit_notes_drafts_update"
	OpCreditNotesDraftsDelete     = "credit_notes_drafts_delete"
	OpCreditNotesDraftsCreateFrom = "credit_notes_drafts_create_from"

	OpCreditNotesDraftsAttachmentsList   = "credit_notes_drafts_attachments_list"
	OpCreditNotesDraftsAttachmentsAttach = "credit_notes_drafts_attachments_attach"

	OpOffersList          = "offers_list"
	OpOffersGet           = "offers_get"
	OpOffersSend          = "offers_send"
	OpOffersCounterCreate = "offers_counter_create"
	OpOffersCounterGet    = "offers_counter_get"

	OpOffersDraftsList       = "offers_drafts_list"
	OpOffersDraftsGet        = "offers_drafts_get"
	OpOffersDraftsCreate     = "offers_drafts_create"
	OpOffersDraftsUpdate     = "offers_drafts_update"
	OpOffersDraftsDelete     = "offers_drafts_delete"
	OpOffersDraftsCreateFrom = "offers_drafts_create_from"

	OpOffersDraftsAttachmentsList   = "offers_drafts_attachments_list"
	OpOffersDraftsAttachmentsAttach = "offers_drafts_attachments_attach"

	OpOrderConfirmationsList                    = "order_confirmations_list"
	OpOrderConfirmationsGet                     = "order_confirmations_get"
	OpOrderConfirmationsCounterCreate           = "order_confirmations_counter_create"
	OpOrderConfirmationsCounterGet              = "order_confirmations_counter_get"
	OpOrderConfirmationsCreateInvoiceDraft      = "order_confirmations_create_invoice_draft"
	OpOrderConfirmationsDraftsList              = "order_confirmations_drafts_list"
	OpOrderConfirmationsDraftsGet               = "order_confirmations_drafts_get"
	OpOrderConfirmationsDraftsCreate            = "order_confirmations_drafts_create"
	OpOrderConfirmationsDraftsUpdate            = "order_confirmations_drafts_update"
	OpOrderConfirmationsDraftsDelete            = "order_confirmations_drafts_delete"
	OpOrderConfirmationsDraftsCreateFrom        = "order_confirmations_drafts_create_from"
	OpOrderConfirmationsDraftsAttachmentsList   = "order_confirmations_drafts_attachments_list"
	OpOrderConfirmationsDraftsAttachmentsAttach = "order_confirmations_drafts_attachments_attach"

	OpProductsList              = "products_list"
	OpProductsGet               = "products_get"
	OpProductsCreate            = "products_create"
	OpProductsUpdate            = "products_update"
	OpProductsDelete            = "products_delete"
	OpProductsSalesReportCreate = "products_sales_report_create"

	OpSalesList        = "sales_list"
	OpSalesGet         = "sales_get"
	OpSalesCreate      = "sales_create"
	OpSalesDelete      = "sales_delete"
	OpSalesSettle      = "sales_settle"
	OpSalesWriteOff    = "sales_write_off"
	OpSalesAttachments = "sales_attachments_list"
	OpSalesAttach      = "sales_attachments_attach"

	OpSalesPaymentsList   = "sales_payments_list"
	OpSalesPaymentsGet    = "sales_payments_get"
	OpSalesPaymentsCreate = "sales_payments_create"

	OpSalesDraftsList              = "sales_drafts_list"
	OpSalesDraftsGet               = "sales_drafts_get"
	OpSalesDraftsCreate            = "sales_drafts_create"
	OpSalesDraftsUpdate            = "sales_drafts_update"
	OpSalesDraftsDelete            = "sales_drafts_delete"
	OpSalesDraftsCreateFrom        = "sales_drafts_create_from"
	OpSalesDraftsAttachmentsList   = "sales_drafts_attachments_list"
	OpSalesDraftsAttachmentsAttach = "sales_drafts_attachments_attach"

	OpPurchasesList        = "purchases_list"
	OpPurchasesGet         = "purchases_get"
	OpPurchasesCreate      = "purchases_create"
	OpPurchasesDelete      = "purchases_delete"
	OpPurchasesAttachments = "purchases_attachments_list"
	OpPurchasesAttach      = "purchases_attachments_attach"

	OpPurchasesPaymentsList   = "purchases_payments_list"
	OpPurchasesPaymentsGet    = "purchases_payments_get"
	OpPurchasesPaymentsCreate = "purchases_payments_create"

	OpPurchasesDraftsList              = "purchases_drafts_list"
	OpPurchasesDraftsGet               = "purchases_drafts_get"
	OpPurchasesDraftsCreate            = "purchases_drafts_create"
	OpPurchasesDraftsUpdate            = "purchases_drafts_update"
	OpPurchasesDraftsDelete            = "purchases_drafts_delete"
	OpPurchasesDraftsCreateFrom        = "purchases_drafts_create_from"
	OpPurchasesDraftsAttachmentsList   = "purchases_drafts_attachments_list"
	OpPurchasesDraftsAttachmentsAttach = "purchases_drafts_attachments_attach"

	OpInboxList = "inbox_list"
	OpInboxGet  = "inbox_get"
	OpInboxSend = "inbox_send"

	OpProjectsList   = "projects_list"
	OpProjectsGet    = "projects_get"
	OpProjectsCreate = "projects_create"
	OpProjectsUpdate = "projects_update"
	OpProjectsDelete = "projects_delete"

	OpUserGet = "user_get"

	OpAccountBalancesList = "account_balances_list"
	OpAccountBalancesGet  = "account_balances_get"

	OpBankBalancesList = "bank_balances_list"

	OpGroupsList = "groups_list"

	OpActivitiesList   = "activities_list"
	OpActivitiesGet    = "activities_get"
	OpActivitiesCreate = "activities_create"
	OpActivitiesUpdate = "activities_update"
	OpActivitiesDelete = "activities_delete"

	OpTimeEntriesList                  = "time_entries_list"
	OpTimeEntriesGet                   = "time_entries_get"
	OpTimeEntriesCreate                = "time_entries_create"
	OpTimeEntriesUpdate                = "time_entries_update"
	OpTimeEntriesDelete                = "time_entries_delete"
	OpTimeEntriesInvoiceDraftFromTimes = "time_entries_invoice_draft_create"

	OpTimeUsersList = "time_users_list"
	OpTimeUsersGet  = "time_users_get"
)

// RegistryEntry holds per-op metadata that BOTH frontends consume.
// OperationID is the snake_case OAS operationId — looked up against
// ops/mutating.gen.go to get the Mutating bit.
// CompanyScoped is true when the op acts on `/companies/{slug}/...`.
type RegistryEntry struct {
	OperationID   string
	Mutating      bool
	CompanyScoped bool
}

// mustEntry builds a RegistryEntry, sourcing Mutating from
// mutating.gen.go via the OAS operationId. Panics at package-init
// time if opID is unknown — catches typos at startup, not later.
func mustEntry(opID string, companyScoped bool) RegistryEntry {
	if _, ok := mutating[opID]; !ok {
		panic("ops.Registry: unknown OperationID " + opID + " (not in mutating.gen.go)")
	}
	return RegistryEntry{
		OperationID:   opID,
		Mutating:      mutating[opID],
		CompanyScoped: companyScoped,
	}
}

// Registry indexed by Op* const value (human-friendly).
var Registry = map[string]RegistryEntry{
	OpCompaniesList: mustEntry("get_companies", false),
	OpCompaniesGet:  mustEntry("get_company", true),

	OpContactsList:              mustEntry("get_contacts", true),
	OpContactsGet:               mustEntry("get_contact", true),
	OpContactsCreate:            mustEntry("create_contact", true),
	OpContactsUpdate:            mustEntry("update_contact", true),
	OpContactsDelete:            mustEntry("delete_contact", true),
	OpContactsAttachmentsAttach: mustEntry("add_attachment_to_contact", true),

	OpContactsPersonsList:   mustEntry("get_contact_contact_person", true),
	OpContactsPersonsGet:    mustEntry("get_contact_person", true),
	OpContactsPersonsCreate: mustEntry("add_contact_person_to_contact", true),
	OpContactsPersonsUpdate: mustEntry("update_contact_contact_person", true),
	OpContactsPersonsDelete: mustEntry("delete_contact_contact_person", true),

	OpAccountsList: mustEntry("get_accounts", true),
	OpAccountsGet:  mustEntry("get_account", true),

	OpBankAccountsList:   mustEntry("get_bank_accounts", true),
	OpBankAccountsGet:    mustEntry("get_bank_account", true),
	OpBankAccountsCreate: mustEntry("create_bank_account", true),

	OpJournalEntriesList:              mustEntry("get_journal_entries", true),
	OpJournalEntriesGet:               mustEntry("get_journal_entry", true),
	OpJournalEntriesCreate:            mustEntry("create_general_journal_entry", true),
	OpJournalEntriesAttachmentsList:   mustEntry("get_journal_entry_attachments", true),
	OpJournalEntriesAttachmentsAttach: mustEntry("add_attachment_to_journal_entry", true),

	OpTransactionsList:   mustEntry("get_transactions", true),
	OpTransactionsGet:    mustEntry("get_transaction", true),
	OpTransactionsDelete: mustEntry("delete_transaction", true),

	OpInvoicesList:          mustEntry("get_invoices", true),
	OpInvoicesGet:           mustEntry("get_invoice", true),
	OpInvoicesCreate:        mustEntry("create_invoice", true),
	OpInvoicesUpdate:        mustEntry("update_invoice", true),
	OpInvoicesSend:          mustEntry("send_invoice", true),
	OpInvoicesCounterCreate: mustEntry("create_invoice_counter", true),
	OpInvoicesCounterGet:    mustEntry("get_invoice_counter", true),

	OpInvoicesDraftsList:       mustEntry("get_invoice_drafts", true),
	OpInvoicesDraftsGet:        mustEntry("get_invoice_draft", true),
	OpInvoicesDraftsCreate:     mustEntry("create_invoice_draft", true),
	OpInvoicesDraftsUpdate:     mustEntry("update_invoice_draft", true),
	OpInvoicesDraftsDelete:     mustEntry("delete_invoice_draft", true),
	OpInvoicesDraftsCreateFrom: mustEntry("create_invoice_from_draft", true),

	OpInvoicesAttachmentsList:         mustEntry("get_invoice_attachments", true),
	OpInvoicesAttachmentsAttach:       mustEntry("add_attachment_to_invoice", true),
	OpInvoicesDraftsAttachmentsList:   mustEntry("get_invoice_draft_attachments", true),
	OpInvoicesDraftsAttachmentsAttach: mustEntry("add_attachment_to_invoice_draft", true),

	OpCreditNotesList:          mustEntry("get_credit_notes", true),
	OpCreditNotesGet:           mustEntry("get_credit_note", true),
	OpCreditNotesSend:          mustEntry("send_credit_note", true),
	OpCreditNotesCounterCreate: mustEntry("create_credit_note_counter", true),
	OpCreditNotesCounterGet:    mustEntry("get_credit_note_counter", true),
	OpCreditNotesFullCreate:    mustEntry("create_full_credit_note", true),
	OpCreditNotesPartialCreate: mustEntry("create_partial_credit_note", true),

	OpCreditNotesDraftsList:       mustEntry("get_credit_note_drafts", true),
	OpCreditNotesDraftsGet:        mustEntry("get_credit_note_draft", true),
	OpCreditNotesDraftsCreate:     mustEntry("create_credit_note_draft", true),
	OpCreditNotesDraftsUpdate:     mustEntry("update_credit_note_draft", true),
	OpCreditNotesDraftsDelete:     mustEntry("delete_credit_note_draft", true),
	OpCreditNotesDraftsCreateFrom: mustEntry("create_credit_note_from_draft", true),

	OpCreditNotesDraftsAttachmentsList:   mustEntry("get_credit_note_draft_attachments", true),
	OpCreditNotesDraftsAttachmentsAttach: mustEntry("add_attachment_to_credit_note_draft", true),

	OpOffersList:          mustEntry("get_offers", true),
	OpOffersGet:           mustEntry("get_offer", true),
	OpOffersSend:          mustEntry("send_offer", true),
	OpOffersCounterCreate: mustEntry("create_offer_counter", true),
	OpOffersCounterGet:    mustEntry("get_offer_counter", true),

	OpOffersDraftsList:       mustEntry("get_offer_drafts", true),
	OpOffersDraftsGet:        mustEntry("get_offer_draft", true),
	OpOffersDraftsCreate:     mustEntry("create_offer_draft", true),
	OpOffersDraftsUpdate:     mustEntry("update_offer_draft", true),
	OpOffersDraftsDelete:     mustEntry("delete_offer_draft", true),
	OpOffersDraftsCreateFrom: mustEntry("create_offer_from_draft", true),

	OpOffersDraftsAttachmentsList:   mustEntry("get_offer_draft_attachments", true),
	OpOffersDraftsAttachmentsAttach: mustEntry("add_attachment_to_offer_draft", true),

	OpOrderConfirmationsList:               mustEntry("get_order_confirmations", true),
	OpOrderConfirmationsGet:                mustEntry("get_order_confirmation", true),
	OpOrderConfirmationsCounterCreate:      mustEntry("create_order_confirmation_counter", true),
	OpOrderConfirmationsCounterGet:         mustEntry("get_order_confirmation_counter", true),
	OpOrderConfirmationsCreateInvoiceDraft: mustEntry("create_invoice_draft_from_order_confirmation", true),

	OpOrderConfirmationsDraftsList:       mustEntry("get_order_confirmation_drafts", true),
	OpOrderConfirmationsDraftsGet:        mustEntry("get_order_confirmation_draft", true),
	OpOrderConfirmationsDraftsCreate:     mustEntry("create_order_confirmation_draft", true),
	OpOrderConfirmationsDraftsUpdate:     mustEntry("update_order_confirmation_draft", true),
	OpOrderConfirmationsDraftsDelete:     mustEntry("delete_order_confirmation_draft", true),
	OpOrderConfirmationsDraftsCreateFrom: mustEntry("create_order_confirmation_from_draft", true),

	OpOrderConfirmationsDraftsAttachmentsList:   mustEntry("get_order_confirmation_draft_attachments", true),
	OpOrderConfirmationsDraftsAttachmentsAttach: mustEntry("add_attachment_to_order_confirmation_draft", true),

	OpProductsList:              mustEntry("get_products", true),
	OpProductsGet:               mustEntry("get_product", true),
	OpProductsCreate:            mustEntry("create_product", true),
	OpProductsUpdate:            mustEntry("update_product", true),
	OpProductsDelete:            mustEntry("delete_product", true),
	OpProductsSalesReportCreate: mustEntry("create_product_sales_report", true),

	OpSalesList:        mustEntry("get_sales", true),
	OpSalesGet:         mustEntry("get_sale", true),
	OpSalesCreate:      mustEntry("create_sale", true),
	OpSalesDelete:      mustEntry("delete_sale", true),
	OpSalesSettle:      mustEntry("settled_sale", true),
	OpSalesWriteOff:    mustEntry("write_off_sale", true),
	OpSalesAttachments: mustEntry("get_sale_attachments", true),
	OpSalesAttach:      mustEntry("add_attachment_to_sale", true),

	OpSalesPaymentsList:   mustEntry("get_sale_payments", true),
	OpSalesPaymentsGet:    mustEntry("get_sale_payment", true),
	OpSalesPaymentsCreate: mustEntry("create_sale_payment", true),

	OpSalesDraftsList:              mustEntry("get_sale_drafts", true),
	OpSalesDraftsGet:               mustEntry("get_sale_draft", true),
	OpSalesDraftsCreate:            mustEntry("create_sale_draft", true),
	OpSalesDraftsUpdate:            mustEntry("update_sale_draft", true),
	OpSalesDraftsDelete:            mustEntry("delete_sale_draft", true),
	OpSalesDraftsCreateFrom:        mustEntry("create_sale_from_draft", true),
	OpSalesDraftsAttachmentsList:   mustEntry("get_sale_draft_attachments", true),
	OpSalesDraftsAttachmentsAttach: mustEntry("add_attachment_to_sale_draft", true),

	OpPurchasesList:        mustEntry("get_purchases", true),
	OpPurchasesGet:         mustEntry("get_purchase", true),
	OpPurchasesCreate:      mustEntry("create_purchase", true),
	OpPurchasesDelete:      mustEntry("delete_purchase", true),
	OpPurchasesAttachments: mustEntry("get_purchase_attachments", true),
	OpPurchasesAttach:      mustEntry("add_attachment_to_purchase", true),

	OpPurchasesPaymentsList:   mustEntry("get_purchase_payments", true),
	OpPurchasesPaymentsGet:    mustEntry("get_purchase_payment", true),
	OpPurchasesPaymentsCreate: mustEntry("create_purchase_payment", true),

	OpPurchasesDraftsList:              mustEntry("get_purchase_drafts", true),
	OpPurchasesDraftsGet:               mustEntry("get_purchase_draft", true),
	OpPurchasesDraftsCreate:            mustEntry("create_purchase_draft", true),
	OpPurchasesDraftsUpdate:            mustEntry("update_purchase_draft", true),
	OpPurchasesDraftsDelete:            mustEntry("delete_purchase_draft", true),
	OpPurchasesDraftsCreateFrom:        mustEntry("create_purchase_from_draft", true),
	OpPurchasesDraftsAttachmentsList:   mustEntry("get_purchase_draft_attachments", true),
	OpPurchasesDraftsAttachmentsAttach: mustEntry("add_attachment_to_purchase_draft", true),

	OpInboxList: mustEntry("get_inbox", true),
	OpInboxGet:  mustEntry("get_inbox_document", true),
	OpInboxSend: mustEntry("create_inbox_document", true),

	OpProjectsList:   mustEntry("get_projects", true),
	OpProjectsGet:    mustEntry("get_project", true),
	OpProjectsCreate: mustEntry("create_project", true),
	OpProjectsUpdate: mustEntry("update_project", true),
	OpProjectsDelete: mustEntry("delete_project", true),

	OpUserGet: mustEntry("get_user", false),

	OpAccountBalancesList: mustEntry("get_account_balances", true),
	OpAccountBalancesGet:  mustEntry("get_account_balance", true),

	OpBankBalancesList: mustEntry("get_bank_balances", true),

	OpGroupsList: mustEntry("get_groups", true),

	OpActivitiesList:   mustEntry("get_activities", true),
	OpActivitiesGet:    mustEntry("get_activity", true),
	OpActivitiesCreate: mustEntry("create_activity", true),
	OpActivitiesUpdate: mustEntry("update_activity", true),
	OpActivitiesDelete: mustEntry("delete_activity", true),

	OpTimeEntriesList:                  mustEntry("get_time_entries", true),
	OpTimeEntriesGet:                   mustEntry("get_time_entry", true),
	OpTimeEntriesCreate:                mustEntry("create_time_entry", true),
	OpTimeEntriesUpdate:                mustEntry("update_time_entry", true),
	OpTimeEntriesDelete:                mustEntry("delete_time_entry", true),
	OpTimeEntriesInvoiceDraftFromTimes: mustEntry("create_invoice_draft_from_time_entries", true),

	OpTimeUsersList: mustEntry("get_time_users", true),
	OpTimeUsersGet:  mustEntry("get_time_user", true),
}
