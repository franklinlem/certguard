package httpapi

import (
	"context"
	"errors"
	"github.com/franklinlem/certguard/api/internal/telemetry"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMetricsAccess(t *testing.T) {
	token := strings.Repeat("x", 32)
	calls := 0
	collect := func(context.Context) (telemetry.Snapshot, error) {
		calls++
		return telemetry.Snapshot{Checks: 2, Buckets: []int64{0, 1, 1, 2, 2, 2, 2, 2}}, nil
	}
	for _, tc := range []struct {
		token, auth string
		status      int
	}{{"", "", 404}, {token, "", 401}, {token, "Bearer wrong", 401}, {token, "Bearer " + token, 200}, {token, "Basic " + token, 401}} {
		h := MetricsHandler(tc.token, collect)
		r := httptest.NewRequest("GET", "/metrics", nil)
		r.Header.Set("Authorization", tc.auth)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != tc.status || w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal(w.Code, w.Header())
		}
		if tc.status == 200 && !strings.Contains(w.Body.String(), "certguard_checks_total 2\n") {
			t.Fatal(w.Body.String())
		}
	}
	if calls != 1 {
		t.Fatal("coleta sem autenticação", calls)
	}
	h := MetricsHandler(token, func(context.Context) (telemetry.Snapshot, error) {
		return telemetry.Snapshot{}, errors.New("postgres://secret@example.com/private")
	})
	r := httptest.NewRequest("GET", "/metrics", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 503 || strings.Contains(w.Body.String(), "secret") {
		t.Fatal(w.Code, w.Body.String())
	}
	for _, v := range []string{"short", strings.Repeat("x", 257), strings.Repeat("x", 32) + "\n", strings.Repeat("á", 32)} {
		if ValidateMetricsToken(v) == nil {
			t.Fatal("token inválido aceito")
		}
	}
}
