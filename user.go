package types

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
)

const HeaderUser = "X-User"

type User struct {
	ID           int64             `json:"id"`
	Email        string            `json:"email,omitempty"`
	Name         string            `json:"name,omitempty"`
	Roles        []string          `json:"roles,omitempty"`
	Features     []string          `json:"features,omitempty"`
	Tenant       string            `json:"tenant,omitempty"`
	Impersonator *int64            `json:"impersonator,omitempty"`
	Claims       map[string]string `json:"claims,omitempty"`
}

func (u User) Has(role string) bool { return slices.Contains(u.Roles, role) }

func (u User) HasAny(roles ...string) bool {
	for _, r := range roles {
		if u.Has(r) {
			return true
		}
	}
	return false
}

func (u User) Can(feature string) bool { return slices.Contains(u.Features, feature) }

func (u User) Impersonated() bool { return u.Impersonator != nil }

var ErrNoUser = errors.New("no user in request")

func ParseUser(header string) (User, error) {
	if header == "" {
		return User{}, ErrNoUser
	}
	var u User
	if err := json.Unmarshal([]byte(header), &u); err != nil {
		return User{}, err
	}
	if u.ID == 0 {
		return User{}, ErrNoUser
	}
	return u, nil
}

func (u User) Header() string {
	b, _ := json.Marshal(u)
	return string(b)
}

type userKey struct{}

func WithUser(ctx context.Context, u User) context.Context {
	return context.WithValue(ctx, userKey{}, u)
}

func UserFrom(ctx context.Context) (User, bool) {
	u, ok := ctx.Value(userKey{}).(User)
	return u, ok
}

func MustUser(ctx context.Context) User {
	u, ok := UserFrom(ctx)
	if !ok {
		panic(ErrNoUser)
	}
	return u
}
