package types

import "time"

const (
	DefaultTake = 20
	MaxTake     = 100
)

type PageRequest struct {
	Take int `query:"take" json:"take,omitempty" validate:"omitempty,min=1,max=100"`
	Skip int `query:"skip" json:"skip,omitempty" validate:"omitempty,min=0"`
}

func (p PageRequest) Limit() int {
	switch {
	case p.Take <= 0:
		return DefaultTake
	case p.Take > MaxTake:
		return MaxTake
	}
	return p.Take
}

func (p PageRequest) Offset() int {
	if p.Skip < 0 {
		return 0
	}
	return p.Skip
}

type Page[T any] struct {
	Items []T   `json:"items"`
	Total int64 `json:"total"`
	Take  int   `json:"take"`
	Skip  int   `json:"skip"`
}

func NewPage[T any](items []T, total int64, req PageRequest) Page[T] {
	if items == nil {
		items = []T{}
	}
	return Page[T]{Items: items, Total: total, Take: req.Limit(), Skip: req.Offset()}
}

func (p Page[T]) HasMore() bool { return int64(p.Skip+len(p.Items)) < p.Total }

func MapPage[A, B any](p Page[A], fn func(A) B) Page[B] {
	out := make([]B, len(p.Items))
	for i, v := range p.Items {
		out[i] = fn(v)
	}
	return Page[B]{Items: out, Total: p.Total, Take: p.Take, Skip: p.Skip}
}

type Direction string

const (
	Asc  Direction = "asc"
	Desc Direction = "desc"
)

type Sort struct {
	Field     string    `query:"sort" json:"field"`
	Direction Direction `query:"order" json:"direction,omitempty" validate:"omitempty,oneof=asc desc"`
}

type Timestamps struct {
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type IDRequest struct {
	ID int64 `path:"id" validate:"required,min=1"`
}

type UUIDRequest struct {
	ID string `path:"id" validate:"required,uuid"`
}

type Empty struct{}
