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
		reflect.TypeFor[CompanyOut](),
		reflect.TypeFor[CompaniesListOut](),
		reflect.TypeFor[ContactOut](),
		reflect.TypeFor[ContactsListOut](),
		reflect.TypeFor[ContactPersonOut](),
		reflect.TypeFor[ContactPersonsListOut](),
		reflect.TypeFor[ContactsDeleteOut](),
		reflect.TypeFor[ContactsPersonsDeleteOut](),
		reflect.TypeFor[AccountOut](),
		reflect.TypeFor[AccountsListOut](),
		reflect.TypeFor[BankAccountOut](),
		reflect.TypeFor[BankAccountsListOut](),
		reflect.TypeFor[JournalEntryOut](),
		reflect.TypeFor[JournalEntryLineOut](),
		reflect.TypeFor[JournalEntriesListOut](),
		reflect.TypeFor[AttachmentOut](),
		reflect.TypeFor[AttachmentsListOut](),
		reflect.TypeFor[JournalEntriesAttachmentsAttachOut](),
		reflect.TypeFor[TransactionOut](),
		reflect.TypeFor[TransactionsListOut](),
		reflect.TypeFor[InvoiceOut](),
		reflect.TypeFor[InvoiceLineOut](),
		reflect.TypeFor[InvoicesListOut](),
		reflect.TypeFor[InvoicesSendOut](),
		reflect.TypeFor[InvoicesCounterCreateOut](),
		reflect.TypeFor[InvoiceDraftOut](),
		reflect.TypeFor[InvoiceDraftLineOut](),
		reflect.TypeFor[InvoiceDraftsListOut](),
		reflect.TypeFor[InvoiceDraftsDeleteOut](),
		reflect.TypeFor[InvoiceDraftsCreateFromOut](),
		reflect.TypeFor[InvoicesAttachmentsAttachOut](),
		reflect.TypeFor[InvoiceDraftsAttachmentsAttachOut](),
		reflect.TypeFor[CreditNoteOut](),
		reflect.TypeFor[CreditNotesListOut](),
		reflect.TypeFor[CreditNotesSendOut](),
		reflect.TypeFor[CreditNotesCounterCreateOut](),
		reflect.TypeFor[CreditNotesFullCreateOut](),
		reflect.TypeFor[CreditNotesPartialCreateOut](),
		reflect.TypeFor[CreditNoteDraftsListOut](),
		reflect.TypeFor[CreditNoteDraftsDeleteOut](),
		reflect.TypeFor[CreditNoteDraftsCreateFromOut](),
		reflect.TypeFor[CreditNoteDraftsAttachmentsAttachOut](),
	}
}

func TestOutFieldUnits(t *testing.T) {
	var failures []string
	for _, st := range outStructs() {
		for f := range st.Fields() {
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
	if t == reflect.TypeFor[time.Time]() {
		return "time.Time"
	}
	if t.Kind() == reflect.String && t.Name() == "Date" {
		return "Date"
	}
	return t.Kind().String()
}
