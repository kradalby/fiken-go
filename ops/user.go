package ops

import (
	"context"
)

// UserGetIn is intentionally empty. /user takes no parameters — it
// returns the user identified by the bearer token. The struct exists
// so the MCP tool registration has a typed In to derive InputSchema
// from, matching the convention used by every other op.
type UserGetIn struct{}

// UserOut is the canonical user shape exposed to CLI/MCP. Mirrors the
// upstream userinfo schema: name + email, both optional in the spec.
type UserOut struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

// TableHeader implements output.tableRow.
func (UserOut) TableHeader() []string {
	return []string{"NAME", "EMAIL"}
}

// TableRow implements output.tableRow.
func (u UserOut) TableRow() []string {
	return []string{u.Name, u.Email}
}

// UserGet returns information about the authenticated user. Not
// company-scoped — the endpoint resolves the caller from the bearer
// token, so no slug is needed.
func (c *Client) UserGet(ctx context.Context, _ UserGetIn) Result[UserOut] {
	resp, err := c.gen.GetUser(ctx)
	if err != nil {
		return Err[UserOut](MapErr(OpUserGet, err))
	}
	if resp == nil {
		return Ok[UserOut](UserOut{})
	}
	return Ok[UserOut](UserOut{
		Name:  resp.Name.Or(""),
		Email: resp.Email.Or(""),
	})
}
