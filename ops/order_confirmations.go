package ops

import (
	"context"
	"net/url"
	"strconv"

	"github.com/kradalby/fiken-go/fiken"
)

// OrderConfirmationsListIn carries paged-list input for order
// confirmations. The upstream endpoint exposes a narrow filter surface
// (only page/page-size); no content-axis filters are surfaced.
type OrderConfirmationsListIn struct {
	Company  string `json:"company"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
}

// OrderConfirmationOut is the canonical order-confirmation shape
// exposed to CLI/MCP. Monetary fields stay int64 øre per the global
// convention. The upstream schema flattens contactId/contactPersonId
// (no nested customer); the translation keeps the flat shape and
// exposes ContactID + ContactPersonID only when set. Lines reuse
// InvoiceLineOut since the upstream type is invoiceLineResult —
// identical to invoices and offers.
type OrderConfirmationOut struct {
	ConfirmationID        int64            `json:"confirmation_id,omitempty"`
	ConfirmationNumber    int32            `json:"confirmation_number,omitempty"`
	ConfirmationDraftUUID string           `json:"confirmation_draft_uuid,omitempty"`
	Date                  Date             `json:"date,omitempty"`
	Net                   int64            `json:"net,omitempty"`
	Vat                   int64            `json:"vat,omitempty"`
	Gross                 int64            `json:"gross,omitempty"`
	Comment               string           `json:"comment,omitempty"`
	YourReference         string           `json:"your_reference,omitempty"`
	OurReference          string           `json:"our_reference,omitempty"`
	OrderReference        string           `json:"order_reference,omitempty"`
	Discount              float64          `json:"discount,omitempty"`
	Currency              string           `json:"currency,omitempty"`
	ContactID             int64            `json:"contact_id,omitempty"`
	ContactPersonID       int64            `json:"contact_person_id,omitempty"`
	ProjectID             int64            `json:"project_id,omitempty"`
	CreatedInvoice        int64            `json:"created_invoice,omitempty"`
	Archived              bool             `json:"archived,omitempty"`
	InternalComment       string           `json:"internal_comment,omitempty"`
	Lines                 []InvoiceLineOut `json:"lines,omitempty"`
}

// TableHeader implements output.tableRow.
func (OrderConfirmationOut) TableHeader() []string {
	return []string{"ID", "NUMBER", "DATE", "GROSS", "INVOICE_ID"}
}

// TableRow implements output.tableRow.
func (o OrderConfirmationOut) TableRow() []string {
	return []string{
		strconv.FormatInt(o.ConfirmationID, 10),
		strconv.Itoa(int(o.ConfirmationNumber)),
		string(o.Date),
		strconv.FormatInt(o.Gross, 10),
		strconv.FormatInt(o.CreatedInvoice, 10),
	}
}

// OrderConfirmationsListOut is the paged response.
type OrderConfirmationsListOut = ListOut[OrderConfirmationOut]

// OrderConfirmationsList returns order confirmations for the specified
// company.
func (c *Client) OrderConfirmationsList(ctx context.Context, in OrderConfirmationsListIn) Result[OrderConfirmationsListOut] {
	if in.Company == "" {
		return Err[OrderConfirmationsListOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpOrderConfirmationsList,
		})
	}
	params := fiken.GetOrderConfirmationsParams{CompanySlug: in.Company}
	if in.Page > 0 {
		params.Page.SetTo(in.Page)
	}
	if in.PageSize > 0 {
		params.PageSize.SetTo(in.PageSize)
	}
	resp, err := c.gen.GetOrderConfirmations(ctx, params)
	if err != nil {
		return Err[OrderConfirmationsListOut](MapErr(OpOrderConfirmationsList, err))
	}
	return Ok[OrderConfirmationsListOut](translateOrderConfirmationsList(resp))
}

// translateOrderConfirmationsList converts the ogen response into the
// canonical ListOut[OrderConfirmationOut] envelope, including paging
// meta.
func translateOrderConfirmationsList(resp *fiken.GetOrderConfirmationsOKHeaders) OrderConfirmationsListOut {
	if resp == nil {
		return OrderConfirmationsListOut{Items: []OrderConfirmationOut{}, Meta: ListMeta{}}
	}
	items := make([]OrderConfirmationOut, 0, len(resp.Response))
	for _, o := range resp.Response {
		items = append(items, orderConfirmationToOut(o))
	}
	meta := ListMeta{Returned: len(items)}
	if page, ok := resp.FikenAPIPage.Get(); ok {
		if pageCount, ok2 := resp.FikenAPIPageCount.Get(); ok2 && page+1 < pageCount {
			meta.NextPage = page + 1
		}
	}
	return OrderConfirmationsListOut{Items: items, Meta: meta}
}

// orderConfirmationToOut maps fiken.OrderConfirmation into the
// canonical OrderConfirmationOut. Address is left out for now — the
// upstream type is OptAddress and the table summary stays focused on
// id/number/date/gross/created-invoice.
func orderConfirmationToOut(o fiken.OrderConfirmation) OrderConfirmationOut {
	out := OrderConfirmationOut{
		ConfirmationID:        o.ConfirmationId.Or(0),
		ConfirmationNumber:    o.ConfirmationNumber.Or(0),
		ConfirmationDraftUUID: o.ConfirmationDraftUuid.Or(""),
		Net:                   o.Net.Or(0),
		Vat:                   o.Vat.Or(0),
		Gross:                 o.Gross.Or(0),
		Comment:               o.Comment.Or(""),
		YourReference:         o.YourReference.Or(""),
		OurReference:          o.OurReference.Or(""),
		OrderReference:        o.OrderReference.Or(""),
		Discount:              o.Discount.Or(0),
		Currency:              o.Currency.Or(""),
		ContactID:             o.ContactId.Or(0),
		ContactPersonID:       o.ContactPersonId.Or(0),
		ProjectID:             o.ProjectId.Or(0),
		CreatedInvoice:        o.CreatedInvoice.Or(0),
		Archived:              o.Archived.Or(false),
		InternalComment:       o.InternalComment.Or(""),
	}
	if d, ok := o.Date.Get(); ok {
		out.Date = Date(d.Format("2006-01-02"))
	}
	if len(o.Lines) > 0 {
		out.Lines = make([]InvoiceLineOut, 0, len(o.Lines))
		for _, ln := range o.Lines {
			out.Lines = append(out.Lines, invoiceLineToOut(ln))
		}
	}
	return out
}

// OrderConfirmationsGetIn requires company + confirmation id. The
// upstream path-param is `string` (not int) so it stays a string here
// for faithfulness with the OAS shape.
type OrderConfirmationsGetIn struct {
	Company        string `json:"company"`
	ConfirmationID string `json:"confirmation_id"`
}

// OrderConfirmationsGet returns a single order confirmation by id.
func (c *Client) OrderConfirmationsGet(ctx context.Context, in OrderConfirmationsGetIn) Result[OrderConfirmationOut] {
	if in.Company == "" {
		return Err[OrderConfirmationOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpOrderConfirmationsGet,
		})
	}
	if in.ConfirmationID == "" {
		return Err[OrderConfirmationOut](&Error{
			Code: CodeValidation, Message: "confirmation_id is required", Op: OpOrderConfirmationsGet,
		})
	}
	resp, err := c.gen.GetOrderConfirmation(ctx, fiken.GetOrderConfirmationParams{
		CompanySlug:    in.Company,
		ConfirmationId: in.ConfirmationID,
	})
	if err != nil {
		return Err[OrderConfirmationOut](MapErr(OpOrderConfirmationsGet, err))
	}
	if resp == nil {
		return Ok[OrderConfirmationOut](OrderConfirmationOut{})
	}
	return Ok[OrderConfirmationOut](orderConfirmationToOut(*resp))
}

// OrderConfirmationsCounterCreateIn carries the create-counter payload.
// The upstream endpoint sets the order-confirmation counter starting
// value for the fiscal year.
type OrderConfirmationsCounterCreateIn struct {
	Company string `json:"company"`
	Value   int32  `json:"value"`
}

// OrderConfirmationsCounterCreateOut surfaces the new counter value.
type OrderConfirmationsCounterCreateOut struct {
	Value int32 `json:"value,omitempty"`
}

// TableHeader implements output.tableRow.
func (OrderConfirmationsCounterCreateOut) TableHeader() []string { return []string{"VALUE"} }

// TableRow implements output.tableRow.
func (o OrderConfirmationsCounterCreateOut) TableRow() []string {
	return []string{strconv.Itoa(int(o.Value))}
}

// OrderConfirmationsCounterCreate sets the order-confirmation counter
// starting value for the current fiscal year. Upstream returns 201
// with no body; we echo the requested value back as confirmation.
func (c *Client) OrderConfirmationsCounterCreate(ctx context.Context, in OrderConfirmationsCounterCreateIn) Result[OrderConfirmationsCounterCreateOut] {
	if in.Company == "" {
		return Err[OrderConfirmationsCounterCreateOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpOrderConfirmationsCounterCreate,
		})
	}
	if in.Value <= 0 {
		return Err[OrderConfirmationsCounterCreateOut](&Error{
			Code: CodeValidation, Message: "value must be > 0", Op: OpOrderConfirmationsCounterCreate,
		})
	}
	counter := fiken.NewOptCounter(fiken.Counter{Value: fiken.NewOptInt32(in.Value)})
	if err := c.gen.CreateOrderConfirmationCounter(ctx, counter, fiken.CreateOrderConfirmationCounterParams{
		CompanySlug: in.Company,
	}); err != nil {
		return Err[OrderConfirmationsCounterCreateOut](MapErr(OpOrderConfirmationsCounterCreate, err))
	}
	return Ok[OrderConfirmationsCounterCreateOut](OrderConfirmationsCounterCreateOut{Value: in.Value})
}

// OrderConfirmationsCreateInvoiceDraftIn turns an order confirmation
// into an invoice draft via
// POST /orderConfirmations/{confirmationId}/createInvoiceDraft.
type OrderConfirmationsCreateInvoiceDraftIn struct {
	Company        string `json:"company"`
	ConfirmationID string `json:"confirmation_id"`
}

// OrderConfirmationsCreateInvoiceDraftOut surfaces the Location header
// pointing at the newly-created invoice draft resource.
type OrderConfirmationsCreateInvoiceDraftOut struct {
	Location string `json:"location,omitempty"`
}

// TableHeader implements output.tableRow.
func (OrderConfirmationsCreateInvoiceDraftOut) TableHeader() []string { return []string{"LOCATION"} }

// TableRow implements output.tableRow.
func (o OrderConfirmationsCreateInvoiceDraftOut) TableRow() []string { return []string{o.Location} }

// OrderConfirmationsCreateInvoiceDraft promotes an order confirmation
// into a new invoice draft. Upstream returns 201 with the new draft
// URL in the Location header.
func (c *Client) OrderConfirmationsCreateInvoiceDraft(ctx context.Context, in OrderConfirmationsCreateInvoiceDraftIn) Result[OrderConfirmationsCreateInvoiceDraftOut] {
	if in.Company == "" {
		return Err[OrderConfirmationsCreateInvoiceDraftOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpOrderConfirmationsCreateInvoiceDraft,
		})
	}
	if in.ConfirmationID == "" {
		return Err[OrderConfirmationsCreateInvoiceDraftOut](&Error{
			Code: CodeValidation, Message: "confirmation_id is required", Op: OpOrderConfirmationsCreateInvoiceDraft,
		})
	}
	resp, err := c.gen.CreateInvoiceDraftFromOrderConfirmation(ctx, fiken.CreateInvoiceDraftFromOrderConfirmationParams{
		CompanySlug:    in.Company,
		ConfirmationId: in.ConfirmationID,
	})
	if err != nil {
		return Err[OrderConfirmationsCreateInvoiceDraftOut](MapErr(OpOrderConfirmationsCreateInvoiceDraft, err))
	}
	out := OrderConfirmationsCreateInvoiceDraftOut{}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[OrderConfirmationsCreateInvoiceDraftOut](out)
}

// OrderConfirmationsCounterGet returns the current order-confirmation
// counter value for the fiscal year. Read-only. Reuses CounterGetIn /
// CounterGetOut from invoices.go.
func (c *Client) OrderConfirmationsCounterGet(ctx context.Context, in CounterGetIn) Result[CounterGetOut] {
	if in.Company == "" {
		return Err[CounterGetOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpOrderConfirmationsCounterGet,
		})
	}
	resp, err := c.gen.GetOrderConfirmationCounter(ctx, fiken.GetOrderConfirmationCounterParams{
		CompanySlug: in.Company,
	})
	if err != nil {
		return Err[CounterGetOut](MapErr(OpOrderConfirmationsCounterGet, err))
	}
	if resp == nil {
		return Ok[CounterGetOut](CounterGetOut{})
	}
	return Ok[CounterGetOut](CounterGetOut{Value: resp.Value.Or(0)})
}
