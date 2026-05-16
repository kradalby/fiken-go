package ops

import (
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"
)

// unitRules maps a field-name regex to the Go type that field MUST
// have. Add patterns here as the domain grows.
var unitRules = []struct {
	pattern  *regexp.Regexp
	wantKind string // "int64", "int", "Date", "time.Time"
}{
	{regexp.MustCompile(`(?i)(amount|price|total|sum|net|gross|balance|paid|due)$`), "int64"},
	{regexp.MustCompile(`(?i)(rate|percent)$`), "int"},
	{regexp.MustCompile(`(?i)date$`), "Date"},
	{regexp.MustCompile(`([a-z]At|(?i)datetime)$`), "time.Time"},
}

// outStructs returns every exported type in this package whose name
// starts with "Out" or ends with "Out" — i.e. the canonical
// per-operation response structs. Populated incrementally; Task 12
// (CompanyOut) is the first addition.
func outStructs() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf(CompanyOut{}),
		reflect.TypeOf(CompaniesListOut{}),
		reflect.TypeOf(ContactOut{}),
		reflect.TypeOf(ContactsListOut{}),
		reflect.TypeOf(ContactPersonOut{}),
		reflect.TypeOf(ContactPersonsListOut{}),
		reflect.TypeOf(ContactsDeleteOut{}),
		reflect.TypeOf(ContactsPersonsDeleteOut{}),
		reflect.TypeOf(AccountOut{}),
		reflect.TypeOf(AccountsListOut{}),
		reflect.TypeOf(BankAccountOut{}),
		reflect.TypeOf(BankAccountsListOut{}),
		reflect.TypeOf(JournalEntryOut{}),
		reflect.TypeOf(JournalEntryLineOut{}),
		reflect.TypeOf(JournalEntriesListOut{}),
		reflect.TypeOf(AttachmentOut{}),
		reflect.TypeOf(AttachmentsListOut{}),
		reflect.TypeOf(JournalEntriesAttachmentsAttachOut{}),
		reflect.TypeOf(TransactionOut{}),
		reflect.TypeOf(TransactionsListOut{}),
		reflect.TypeOf(InvoiceOut{}),
		reflect.TypeOf(InvoiceLineOut{}),
		reflect.TypeOf(InvoicesListOut{}),
		reflect.TypeOf(InvoicesSendOut{}),
		reflect.TypeOf(InvoicesCounterCreateOut{}),
		reflect.TypeOf(InvoiceDraftOut{}),
		reflect.TypeOf(InvoiceDraftLineOut{}),
		reflect.TypeOf(InvoiceDraftsListOut{}),
		reflect.TypeOf(InvoiceDraftsDeleteOut{}),
		reflect.TypeOf(InvoiceDraftsCreateFromOut{}),
		reflect.TypeOf(InvoicesAttachmentsAttachOut{}),
		reflect.TypeOf(InvoiceDraftsAttachmentsAttachOut{}),
		reflect.TypeOf(CreditNoteOut{}),
		reflect.TypeOf(CreditNotesListOut{}),
		reflect.TypeOf(CreditNotesSendOut{}),
		reflect.TypeOf(CreditNotesCounterCreateOut{}),
		reflect.TypeOf(CreditNotesFullCreateOut{}),
		reflect.TypeOf(CreditNotesPartialCreateOut{}),
		reflect.TypeOf(CreditNoteDraftsListOut{}),
		reflect.TypeOf(CreditNoteDraftsDeleteOut{}),
		reflect.TypeOf(CreditNoteDraftsCreateFromOut{}),
		reflect.TypeOf(CreditNoteDraftsAttachmentsAttachOut{}),
	}
}

func TestOutFieldUnits(t *testing.T) {
	var failures []string
	for _, st := range outStructs() {
		for i := 0; i < st.NumField(); i++ {
			f := st.Field(i)
			if !f.IsExported() {
				continue
			}
			for _, rule := range unitRules {
				if !rule.pattern.MatchString(f.Name) {
					continue
				}
				gotKind := goKindFor(f.Type)
				if gotKind != rule.wantKind {
					failures = append(failures,
						strings.Join([]string{st.Name(), ".", f.Name, ": got ", gotKind, " want ", rule.wantKind}, ""))
				}
			}
		}
	}
	if len(failures) > 0 {
		t.Fatalf("unit-type violations:\n  %s", strings.Join(failures, "\n  "))
	}
}

func goKindFor(t reflect.Type) string {
	if t == reflect.TypeOf(time.Time{}) {
		return "time.Time"
	}
	if t.Kind() == reflect.String && t.Name() == "Date" {
		return "Date"
	}
	return t.Kind().String()
}
