package types

import (
	"context"
	"errors"
	"testing"
)

func TestParseUser(t *testing.T) {
	u, err := ParseUser(`{"id":7,"email":"a@example.com","roles":["admin"],"features":["beta"]}`)
	if err != nil || u.ID != 7 || !u.Has("admin") || u.Has("owner") || !u.HasAny("x", "admin") || !u.Can("beta") {
		t.Fatalf("%+v %v", u, err)
	}
	if _, err := ParseUser(""); !errors.Is(err, ErrNoUser) {
		t.Fatal("empty header")
	}
	if _, err := ParseUser(`{"id":0}`); !errors.Is(err, ErrNoUser) {
		t.Fatal("zero id")
	}
	if _, err := ParseUser(`{`); err == nil {
		t.Fatal("bad json")
	}
	again, _ := ParseUser(u.Header())
	if again.ID != 7 || again.Email != u.Email {
		t.Fatal("header round trip")
	}
}

func TestUserContext(t *testing.T) {
	ctx := WithUser(context.Background(), User{ID: 3})
	if u, ok := UserFrom(ctx); !ok || u.ID != 3 || MustUser(ctx).ID != 3 {
		t.Fatal("context")
	}
	defer func() {
		if recover() == nil {
			t.Fatal("MustUser should panic without a user")
		}
	}()
	MustUser(context.Background())
}

func TestPage(t *testing.T) {
	if (PageRequest{}).Limit() != DefaultTake || (PageRequest{Take: 500}).Limit() != MaxTake || (PageRequest{Skip: -3}).Offset() != 0 {
		t.Fatal("limits")
	}
	p := NewPage([]int{1, 2}, 5, PageRequest{Take: 2})
	if !p.HasMore() || p.Take != 2 {
		t.Fatalf("%+v", p)
	}
	s := MapPage(p, func(i int) string { return string(rune('a' + i)) })
	if s.Items[1] != "c" || s.Total != 5 {
		t.Fatalf("%+v", s)
	}
	if NewPage[int](nil, 0, PageRequest{}).Items == nil {
		t.Fatal("nil items should become empty slice")
	}
}

func TestAPIError(t *testing.T) {
	e := &APIError{Status: 404, Envelope: ErrorEnvelope{Errors: []ErrorItem{{Code: "E404-1", Message: "nope"}}}}
	if !e.NotFound() || e.Code() != "E404-1" || e.Error() != "404 E404-1: nope" {
		t.Fatal(e.Error())
	}
	if (&APIError{Status: 502}).Error() != "502 Bad Gateway" {
		t.Fatal("no envelope")
	}
}
