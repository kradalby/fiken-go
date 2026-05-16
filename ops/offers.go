package ops

import (
	"context"
	"strconv"

	"github.com/kradalby/fiken-go/fiken"
)

// OffersListIn carries paged-list input for offers. The upstream
// endpoint exposes a narrow filter surface (only page/page-size); no
// content-axis filters are surfaced.
type OffersListIn struct {
	Company  string `json:"company"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
}

// OfferOut is the canonical offer shape exposed to CLI/MCP. Monetary
// fields stay int64 øre per the global convention. The upstream
// schema flattens contactId/contactPersonId (no nested customer); the
// translation keeps the flat shape and exposes ContactID + ContactName
// only when a customer is attached. Lines reuse InvoiceLineOut since
// the upstream type is invoiceLineResult — identical to invoices.
type OfferOut struct {
	OfferID         int64            `json:"offer_id,omitempty"`
	OfferNumber     int32            `json:"offer_number,omitempty"`
	OfferDraftUUID  string           `json:"offer_draft_uuid,omitempty"`
	Date            Date             `json:"date,omitempty"`
	Net             int64            `json:"net,omitempty"`
	Vat             int64            `json:"vat,omitempty"`
	Gross           int64            `json:"gross,omitempty"`
	Comment         string           `json:"comment,omitempty"`
	YourReference   string           `json:"your_reference,omitempty"`
	OurReference    string           `json:"our_reference,omitempty"`
	OrderReference  string           `json:"order_reference,omitempty"`
	Discount        float64          `json:"discount,omitempty"`
	Currency        string           `json:"currency,omitempty"`
	ContactID       int64            `json:"contact_id,omitempty"`
	ContactPersonID int64            `json:"contact_person_id,omitempty"`
	ProjectID       int64            `json:"project_id,omitempty"`
	Archived        bool             `json:"archived,omitempty"`
	Accepted        Date             `json:"accepted,omitempty"`
	Lines           []InvoiceLineOut `json:"lines,omitempty"`
}

// TableHeader implements output.tableRow.
func (OfferOut) TableHeader() []string {
	return []string{"ID", "NUMBER", "DATE", "GROSS", "ACCEPTED"}
}

// TableRow implements output.tableRow.
func (o OfferOut) TableRow() []string {
	return []string{
		strconv.FormatInt(o.OfferID, 10),
		strconv.Itoa(int(o.OfferNumber)),
		string(o.Date),
		strconv.FormatInt(o.Gross, 10),
		string(o.Accepted),
	}
}

// OffersListOut is the paged response.
type OffersListOut = ListOut[OfferOut]

// OffersList returns offers for the specified company.
func (c *Client) OffersList(ctx context.Context, in OffersListIn) Result[OffersListOut] {
	if in.Company == "" {
		return Err[OffersListOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpOffersList,
		})
	}
	params := fiken.GetOffersParams{CompanySlug: in.Company}
	if in.Page > 0 {
		params.Page.SetTo(in.Page)
	}
	if in.PageSize > 0 {
		params.PageSize.SetTo(in.PageSize)
	}
	resp, err := c.gen.GetOffers(ctx, params)
	if err != nil {
		return Err[OffersListOut](MapErr(OpOffersList, err))
	}
	return Ok[OffersListOut](translateOffersList(resp))
}

// translateOffersList converts the ogen response into the canonical
// ListOut[OfferOut] envelope, including paging meta.
func translateOffersList(resp *fiken.GetOffersOKHeaders) OffersListOut {
	if resp == nil {
		return OffersListOut{Items: []OfferOut{}, Meta: ListMeta{}}
	}
	items := make([]OfferOut, 0, len(resp.Response))
	for _, o := range resp.Response {
		items = append(items, offerToOut(o))
	}
	meta := ListMeta{Returned: len(items)}
	if page, ok := resp.FikenAPIPage.Get(); ok {
		if pageCount, ok2 := resp.FikenAPIPageCount.Get(); ok2 && page+1 < pageCount {
			meta.NextPage = page + 1
		}
	}
	return OffersListOut{Items: items, Meta: meta}
}

// offerToOut maps fiken.Offer into the canonical OfferOut. Address is
// left out for now — the upstream type is OptAddress and the table
// summary stays focused on id/number/date/gross/accepted. Dispatches
// are reachable through the underlying ogen Result if needed.
func offerToOut(o fiken.Offer) OfferOut {
	out := OfferOut{
		OfferID:         o.OfferId.Or(0),
		OfferNumber:     o.OfferNumber.Or(0),
		OfferDraftUUID:  o.OfferDraftUuid.Or(""),
		Net:             o.Net.Or(0),
		Vat:             o.Vat.Or(0),
		Gross:           o.Gross.Or(0),
		Comment:         o.Comment.Or(""),
		YourReference:   o.YourReference.Or(""),
		OurReference:    o.OurReference.Or(""),
		OrderReference:  o.OrderReference.Or(""),
		Discount:        o.Discount.Or(0),
		Currency:        o.Currency.Or(""),
		ContactID:       o.ContactId.Or(0),
		ContactPersonID: o.ContactPersonId.Or(0),
		ProjectID:       o.ProjectId.Or(0),
		Archived:        o.Archived.Or(false),
	}
	if d, ok := o.Date.Get(); ok {
		out.Date = Date(d.Format("2006-01-02"))
	}
	if d, ok := o.Accepted.Get(); ok {
		out.Accepted = Date(d.Format("2006-01-02"))
	}
	if len(o.Lines) > 0 {
		out.Lines = make([]InvoiceLineOut, 0, len(o.Lines))
		for _, ln := range o.Lines {
			out.Lines = append(out.Lines, invoiceLineToOut(ln))
		}
	}
	return out
}

// OffersGetIn requires company + offer id. OfferID is a string here
// because the upstream path-param is `string` (not int) in the spec —
// keeps the call-site faithful to the OAS shape.
type OffersGetIn struct {
	Company string `json:"company"`
	OfferID string `json:"offer_id"`
}

// OffersGet returns a single offer by id.
func (c *Client) OffersGet(ctx context.Context, in OffersGetIn) Result[OfferOut] {
	if in.Company == "" {
		return Err[OfferOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpOffersGet,
		})
	}
	if in.OfferID == "" {
		return Err[OfferOut](&Error{
			Code: CodeValidation, Message: "offer_id is required", Op: OpOffersGet,
		})
	}
	resp, err := c.gen.GetOffer(ctx, fiken.GetOfferParams{
		CompanySlug: in.Company,
		OfferId:     in.OfferID,
	})
	if err != nil {
		return Err[OfferOut](MapErr(OpOffersGet, err))
	}
	if resp == nil {
		return Ok[OfferOut](OfferOut{})
	}
	return Ok[OfferOut](offerToOut(*resp))
}

// OffersSendIn carries the send-offer payload. Body mirrors the
// upstream SendOfferRequest so the field surface stays in lock-step
// with the spec.
type OffersSendIn struct {
	Company string                  `json:"company"`
	Body    *fiken.SendOfferRequest `json:"body"`
}

// OffersSendOut is the canonical success shape for sendOffer. The
// upstream endpoint returns 200 with no body; surfacing the offer_id
// keeps the success/error envelope meaningful.
type OffersSendOut struct {
	OfferID int64 `json:"offer_id,omitempty"`
}

// TableHeader implements output.tableRow.
func (OffersSendOut) TableHeader() []string { return []string{"OFFER_ID"} }

// TableRow implements output.tableRow.
func (o OffersSendOut) TableRow() []string {
	return []string{strconv.FormatInt(o.OfferID, 10)}
}

// OffersSend posts a send-offer request. Upstream returns 200 with no
// body; we echo the body's OfferId back so the success envelope stays
// meaningful.
func (c *Client) OffersSend(ctx context.Context, in OffersSendIn) Result[OffersSendOut] {
	if in.Company == "" {
		return Err[OffersSendOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpOffersSend,
		})
	}
	if in.Body == nil {
		return Err[OffersSendOut](&Error{
			Code: CodeValidation, Message: "body is required", Op: OpOffersSend,
		})
	}
	if err := c.gen.SendOffer(ctx, in.Body, fiken.SendOfferParams{
		CompanySlug: in.Company,
	}); err != nil {
		return Err[OffersSendOut](MapErr(OpOffersSend, err))
	}
	return Ok[OffersSendOut](OffersSendOut{OfferID: in.Body.OfferId})
}

// OffersCounterCreateIn carries the create-counter payload. The
// upstream endpoint sets the offer-counter starting value for the
// fiscal year.
type OffersCounterCreateIn struct {
	Company string `json:"company"`
	Value   int32  `json:"value"`
}

// OffersCounterCreateOut surfaces the new counter value.
type OffersCounterCreateOut struct {
	Value int32 `json:"value,omitempty"`
}

// TableHeader implements output.tableRow.
func (OffersCounterCreateOut) TableHeader() []string { return []string{"VALUE"} }

// TableRow implements output.tableRow.
func (o OffersCounterCreateOut) TableRow() []string {
	return []string{strconv.Itoa(int(o.Value))}
}

// OffersCounterCreate sets the offer-counter starting value for the
// current fiscal year. Upstream returns 201 with no body; we echo the
// requested value back as confirmation.
func (c *Client) OffersCounterCreate(ctx context.Context, in OffersCounterCreateIn) Result[OffersCounterCreateOut] {
	if in.Company == "" {
		return Err[OffersCounterCreateOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpOffersCounterCreate,
		})
	}
	if in.Value <= 0 {
		return Err[OffersCounterCreateOut](&Error{
			Code: CodeValidation, Message: "value must be > 0", Op: OpOffersCounterCreate,
		})
	}
	counter := fiken.NewOptCounter(fiken.Counter{Value: fiken.NewOptInt32(in.Value)})
	if err := c.gen.CreateOfferCounter(ctx, counter, fiken.CreateOfferCounterParams{
		CompanySlug: in.Company,
	}); err != nil {
		return Err[OffersCounterCreateOut](MapErr(OpOffersCounterCreate, err))
	}
	return Ok[OffersCounterCreateOut](OffersCounterCreateOut{Value: in.Value})
}

// OffersCounterGet returns the current offer counter value for the
// fiscal year. Read-only. Reuses CounterGetIn / CounterGetOut from
// invoices.go.
func (c *Client) OffersCounterGet(ctx context.Context, in CounterGetIn) Result[CounterGetOut] {
	if in.Company == "" {
		return Err[CounterGetOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpOffersCounterGet,
		})
	}
	resp, err := c.gen.GetOfferCounter(ctx, fiken.GetOfferCounterParams{
		CompanySlug: in.Company,
	})
	if err != nil {
		return Err[CounterGetOut](MapErr(OpOffersCounterGet, err))
	}
	if resp == nil {
		return Ok[CounterGetOut](CounterGetOut{})
	}
	return Ok[CounterGetOut](CounterGetOut{Value: resp.Value.Or(0)})
}
