package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ahmedbaig/go-backend-types"
)

type item struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type listReq struct {
	types.PageRequest
	Q    string   `query:"q"`
	Tags []string `query:"tag"`
	Pub  *bool    `query:"published"`
}

func TestClient(t *testing.T) {
	var seen *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r
		switch {
		case r.Method == "GET" && r.URL.Path == "/items":
			_ = json.NewEncoder(w).Encode(types.Page[item]{Items: []item{{ID: 1, Name: "a"}}, Total: 1})
		case r.Method == "POST":
			var in item
			_ = json.NewDecoder(r.Body).Decode(&in)
			in.ID = 9
			w.WriteHeader(201)
			_ = json.NewEncoder(w).Encode(in)
		case r.Method == "DELETE":
			w.WriteHeader(204)
		default:
			w.WriteHeader(404)
			_, _ = w.Write([]byte(`{"requestId":"r","errors":[{"code":"E404-1","message":"missing"}]}`))
		}
	}))
	defer srv.Close()

	c := New(srv.URL, WithHeader("X-Api-Key", "k"), ForwardUser(),
		WithIDs(func(context.Context) (string, string) { return "r1", "c1" }))
	ctx := types.WithUser(context.Background(), types.User{ID: 5})

	pub := true
	page, err := Get[types.Page[item]](ctx, c, "/items", listReq{PageRequest: types.PageRequest{Take: 5}, Q: "x y", Tags: []string{"a", "b"}, Pub: &pub})
	if err != nil || page.Total != 1 || page.Items[0].Name != "a" {
		t.Fatalf("%+v %v", page, err)
	}
	if got := seen.URL.RawQuery; got != "published=true&q=x+y&tag=a&tag=b&take=5" {
		t.Fatalf("query %q", got)
	}
	if seen.Header.Get("X-Api-Key") != "k" || seen.Header.Get(types.HeaderCorrelationID) != "c1" {
		t.Fatal("headers")
	}
	if u, _ := types.ParseUser(seen.Header.Get(types.HeaderUser)); u.ID != 5 {
		t.Fatal("user not forwarded")
	}

	created, err := Post[item](ctx, c, "/items", item{Name: "n"})
	if err != nil || created.ID != 9 {
		t.Fatalf("%+v %v", created, err)
	}
	if err := Delete(ctx, c, Path("/items/{id}", 9)); err != nil || seen.URL.Path != "/items/9" {
		t.Fatalf("%v %s", err, seen.URL.Path)
	}
	_, err = Get[item](ctx, c, Path("/other/{a}/{b}", "x/y", 2), nil)
	var ae *types.APIError
	if !errors.As(err, &ae) || !ae.NotFound() || ae.Envelope.RequestID != "r" || seen.URL.RawPath != "/other/x%2Fy/2" {
		t.Fatalf("%v %s", err, seen.URL.RawPath)
	}
}
