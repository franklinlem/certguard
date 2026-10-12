package httpapi

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"
)

type ping func(context.Context) error

func (f ping) Ping(c context.Context) error { return f(c) }
func TestHealth(t *testing.T) {
	for _, tc := range []struct {
		path string
		db   Pinger
		code int
	}{{"/health", nil, 200}, {"/ready", nil, 503}, {"/ready", ping(func(context.Context) error { return errors.New("segredo não deve vazar") }), 503}, {"/ready", ping(func(context.Context) error { return nil }), 200}, {"/api/v1/check", nil, 404}} {
		r := httptest.NewRecorder()
		Handler(tc.db).ServeHTTP(r, httptest.NewRequest("GET", tc.path, nil))
		if r.Code != tc.code {
			t.Fatal(r.Code)
		}
		if r.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("cache inseguro")
		}
	}
}
