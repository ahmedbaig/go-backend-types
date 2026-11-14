# go-backend-types

Types shared by every service in the `go-backend-*` family, plus a small typed JSON client for service contracts. Standard library only, so any module can import it cheaply.

```sh
go get github.com/ahmedbaig/go-backend-types
```

| Type | |
|---|---|
| `types.User` | the gateway's `X-User` session: id, email, name, roles, features, tenant, impersonator. `ParseUser`, `Header()`, `Has`, `HasAny`, `Can`, `WithUser` / `UserFrom` / `MustUser` |
| `types.PageRequest` | `take`/`skip` from query or JSON with validation tags; `Limit()` (default 20, max 100), `Offset()` |
| `types.Page[T]` | `{items, total, take, skip}`; `NewPage`, `MapPage`, `HasMore` |
| `types.Sort`, `Direction` | `?sort=field&order=desc` |
| `types.Timestamps` | embed for `createdAt` / `updatedAt` |
| `types.IDRequest`, `UUIDRequest` | `{id}` path parameter with validation |
| `types.ErrorEnvelope`, `ErrorItem`, `FieldError` | the error response every service returns |
| `types.APIError` | what a client gets back for a non-2xx: status, envelope, `Code()`, `NotFound()`, `Conflict()`, `Invalid()`, `Unauthorized()` |

## Contract clients

A service publishes an `api` module with its request/response types and a client built on `client`:

```go
type Client struct{ c *client.Client }

func NewClient(baseURL string, opts ...client.Option) *Client {
	return &Client{c: client.New(baseURL, opts...)}
}

func (c *Client) GetOrder(ctx context.Context, id int64) (Order, error) {
	return client.Get[Order](ctx, c.c, client.Path("/api/v1/orders/{id}", id), nil)
}
```

`client.Get` / `Post` / `Put` / `Patch` / `Delete` encode JSON, turn `query:"..."` struct fields into the query string, and decode the error envelope into `*types.APIError`. Options: `WithHeader`, `WithHTTPClient`, `WithIDs(httpx.IDs)` (forward the correlation ID) and `ForwardUser()` (pass the caller's `X-User` on).

## License

MIT
