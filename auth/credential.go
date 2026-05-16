package auth

// Credential is the storage value for a single profile's auth state.
// Phase 1 holds only a personal API token (Kind="personal"); phase 2
// will widen to include refresh/expires for OAuth.
type Credential struct {
	Kind  string `json:"kind"`
	Token string `json:"token"`
}

// NewPersonal builds a Credential for a personal API token.
func NewPersonal(token string) Credential {
	return Credential{Kind: "personal", Token: token}
}
