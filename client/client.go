package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/ahmedbaig/go-backend-types"
)

type IDsFunc func(ctx context.Context) (requestID, correlationID string)

type Client struct {
	base    string
	http    *http.Client
	headers http.Header
	ids     IDsFunc
	forward bool
}

type Option func(*Client)

func WithHTTPClient(h *http.Client) Option { return func(c *Client) { c.http = h } }

func WithHeader(k, v string) Option { return func(c *Client) { c.headers.Set(k, v) } }

func WithIDs(fn IDsFunc) Option { return func(c *Client) { c.ids = fn } }

func ForwardUser() Option { return func(c *Client) { c.forward = true } }

func New(baseURL string, opts ...Option) *Client {
	c := &Client{base: strings.TrimRight(baseURL, "/"), http: &http.Client{Timeout: 30 * time.Second}, headers: http.Header{}}
	for _, o := range opts {
		o(c)
	}
	return c
}

func (c *Client) BaseURL() string { return c.base }

func Do[Res any](ctx context.Context, c *Client, method, path string, body any) (Res, error) {
	var out Res
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return out, err
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.base+path, rdr)
	if err != nil {
		return out, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	for k, vs := range c.headers {
		req.Header[k] = vs
	}
	if c.ids != nil {
		if _, cid := c.ids(ctx); cid != "" {
			req.Header.Set(types.HeaderCorrelationID, cid)
		}
	}
	if c.forward {
		if u, ok := types.UserFrom(ctx); ok {
			req.Header.Set(types.HeaderUser, u.Header())
		}
	}
	res, err := c.http.Do(req)
	if err != nil {
		return out, err
	}
	defer res.Body.Close()
	data, err := io.ReadAll(io.LimitReader(res.Body, 32<<20))
	if err != nil {
		return out, err
	}
	if res.StatusCode >= 300 {
		ae := &types.APIError{Status: res.StatusCode, Body: string(data)}
		_ = json.Unmarshal(data, &ae.Envelope)
		return out, ae
	}
	if res.StatusCode == http.StatusNoContent || len(data) == 0 {
		return out, nil
	}
	return out, json.Unmarshal(data, &out)
}

func Get[Res any](ctx context.Context, c *Client, path string, query any) (Res, error) {
	return Do[Res](ctx, c, http.MethodGet, path+Query(query), nil)
}

func Post[Res any](ctx context.Context, c *Client, path string, body any) (Res, error) {
	return Do[Res](ctx, c, http.MethodPost, path, body)
}

func Put[Res any](ctx context.Context, c *Client, path string, body any) (Res, error) {
	return Do[Res](ctx, c, http.MethodPut, path, body)
}

func Patch[Res any](ctx context.Context, c *Client, path string, body any) (Res, error) {
	return Do[Res](ctx, c, http.MethodPatch, path, body)
}

func Delete(ctx context.Context, c *Client, path string) error {
	_, err := Do[types.Empty](ctx, c, http.MethodDelete, path, nil)
	return err
}

func Path(pattern string, params ...any) string {
	out := pattern
	for _, p := range params {
		i := strings.Index(out, "{")
		j := strings.Index(out, "}")
		if i < 0 || j < i {
			break
		}
		out = out[:i] + url.PathEscape(toString(reflect.ValueOf(p))) + out[j+1:]
	}
	return out
}

func Query(v any) string {
	if v == nil {
		return ""
	}
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return ""
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return ""
	}
	q := url.Values{}
	addQuery(q, rv)
	if len(q) == 0 {
		return ""
	}
	return "?" + q.Encode()
}

func addQuery(q url.Values, rv reflect.Value) {
	t := rv.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fv := rv.Field(i)
		if f.Anonymous && fv.Kind() == reflect.Struct {
			addQuery(q, fv)
			continue
		}
		name, _, _ := strings.Cut(f.Tag.Get("query"), ",")
		if name == "" || name == "-" || !f.IsExported() {
			continue
		}
		if fv.Kind() == reflect.Pointer {
			if fv.IsNil() {
				continue
			}
			fv = fv.Elem()
		}
		if fv.Kind() == reflect.Slice {
			for j := 0; j < fv.Len(); j++ {
				q.Add(name, toString(fv.Index(j)))
			}
			continue
		}
		if fv.IsZero() {
			continue
		}
		q.Set(name, toString(fv))
	}
}

func toString(v reflect.Value) string {
	if t, ok := v.Interface().(time.Time); ok {
		return t.Format(time.RFC3339)
	}
	if d, ok := v.Interface().(time.Duration); ok {
		return d.String()
	}
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64)
	}
	b, _ := json.Marshal(v.Interface())
	return string(b)
}
