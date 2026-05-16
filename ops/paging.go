package ops

// ListMeta carries pagination metadata accompanying a List operation's
// Items slice. Truncated and Returned are required. NextPage and
// Cancelled are omitempty so default success cases stay terse.
type ListMeta struct {
	Truncated bool `json:"truncated"`
	NextPage  int  `json:"next_page,omitempty"`
	Returned  int  `json:"returned"`
	Cancelled bool `json:"cancelled,omitempty"`
}

// ListOut is the canonical Ok shape for any paged operation.
type ListOut[T any] struct {
	Items []T      `json:"items"`
	Meta  ListMeta `json:"meta"`
}
