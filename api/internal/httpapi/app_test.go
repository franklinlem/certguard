package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/franklinlem/certguard/api/internal/auth"
	"github.com/franklinlem/certguard/api/internal/certificate"
	"github.com/franklinlem/certguard/api/internal/scheduler"
	"github.com/franklinlem/certguard/api/internal/storage"
	"github.com/franklinlem/certguard/api/internal/testutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func request(h http.Handler, method, path, body string, cookie *http.Cookie, csrf, origin string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	if origin != "" {
		r.Header.Set("Origin", origin)
	}
	if csrf != "" {
		r.Header.Set("X-CSRF-Token", csrf)
	}
	if cookie != nil {
		r.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}
func TestOriginAndUnauthorized(t *testing.T) {
	a, e := NewApp(&storage.Store{}, "https://certguard.example.com", false)
	if e != nil {
		t.Fatal(e)
	}
	h := a.Handler()
	for _, path := range []string{"/api/v1/alerts", "/api/v1/dashboard", "/api/v1/auth/me", "/api/v1/endpoints", "/api/v1/endpoints/00000000-0000-0000-0000-000000000000/checks"} {
		if w := request(h, "GET", path, "", nil, "", ""); w.Code != 401 {
			t.Fatal(w.Code)
		}
	}
	for _, origin := range []string{"", "null", "https://evil.example.com"} {
		if w := request(h, "POST", "/api/v1/auth/login", "{}", nil, "", origin); w.Code != 403 {
			t.Fatal(w.Code)
		}
	}
	for _, cfg := range []struct {
		url   string
		local bool
	}{{"http://public.example.com", false}, {"http://public.example.com", true}, {"https://u:p@example.com", false}, {"https://example.com/path", false}, {"", false}} {
		if _, e := NewApp(&storage.Store{}, cfg.url, cfg.local); e == nil {
			t.Fatal(cfg)
		}
	}
	w := httptest.NewRecorder()
	a.setCookie(w, auth.Token(), 3600)
	c := w.Result().Cookies()[0]
	if c.Name != "__Host-certguard_session" || !c.Secure || !c.HttpOnly || c.SameSite != http.SameSiteStrictMode || c.Domain != "" {
		t.Fatal(c)
	}
}
func TestDecodeAndPagination(t *testing.T) {
	for _, body := range []string{`{"unknown":1}`, `{} {}`, strings.Repeat("x", 9000)} {
		r := httptest.NewRequest("POST", "/", strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		var v struct {
			Name string `json:"name"`
		}
		if decode(httptest.NewRecorder(), r, &v) == nil {
			t.Fatal("payload aceito")
		}
	}
	for _, q := range []string{"?limit=101", "?limit=0", "?offset=-1", "?limit=bad"} {
		if _, _, ok := pagination(httptest.NewRecorder(), httptest.NewRequest("GET", "/"+q, nil)); ok {
			t.Fatal(q)
		}
	}
}
func TestPostgresHTTPFlow(t *testing.T) {
	pool := testutil.Database(t)
	s := &storage.Store{Pool: pool}
	if e := s.Bootstrap(context.Background(), "admin@example.com", "Teste-local-12345"); e != nil {
		t.Fatal(e)
	}
	a, e := NewApp(s, "http://localhost:8088", true)
	if e != nil {
		t.Fatal(e)
	}
	a.Background = scheduler.Config{Enabled: true, Workers: 5, RetentionDays: 180}
	a.inspect = func(_ context.Context, e storage.Endpoint) certificate.Result {
		return certificate.Result{Hostname: e.Hostname, Port: e.Port, Status: "error", Code: "blocked_target", CheckedAt: time.Now().UTC()}
	}
	h := a.Handler()
	origin := "http://localhost:8088"
	login := request(h, "POST", "/api/v1/auth/login", `{"email":"admin@example.com","password":"Teste-local-12345"}`, nil, "", origin)
	if login.Code != 200 {
		t.Fatal(login.Code, login.Body.String())
	}
	var session struct {
		CSRF       string           `json:"csrf_token"`
		Background scheduler.Config `json:"background"`
	}
	if e = json.Unmarshal(login.Body.Bytes(), &session); e != nil {
		t.Fatal(e)
	}
	if !session.Background.Enabled || session.Background.Workers != 5 {
		t.Fatal("configuração operacional ausente")
	}
	cookie := login.Result().Cookies()[0]
	for _, q := range []string{"?state=invalid", "?endpoint_id=bad", "?limit=101"} {
		if w := request(h, "GET", "/api/v1/alerts"+q, "", cookie, "", ""); w.Code != 400 {
			t.Fatal(w.Code)
		}
	}
	if w := request(h, "GET", "/api/v1/alerts", "", cookie, "", ""); w.Code != 200 || !strings.Contains(w.Body.String(), `"active":0`) {
		t.Fatal(w.Code, w.Body.String())
	}

	if w := request(h, "GET", "/api/v1/dashboard?limit=101", "", cookie, "", ""); w.Code != 400 {
		t.Fatal(w.Code)
	}
	if w := request(h, "GET", "/api/v1/dashboard?limit=20", "", cookie, "", ""); w.Code != 200 || !strings.Contains(w.Body.String(), `"total":0`) {
		t.Fatal(w.Code, w.Body.String())
	}

	if w := request(h, "POST", "/api/v1/endpoints", `{"name":"Teste","hostname":"example.com"}`, cookie, "", origin); w.Code != 403 {
		t.Fatal("CSRF ausente aceito", w.Code)
	}
	if w := request(h, "GET", "/api/v1/auth/me", "", cookie, "", ""); w.Code != 200 {
		t.Fatal(w.Code)
	}
	for _, host := range []string{"127.0.0.1", "localhost", "http://example.com"} {
		body, _ := json.Marshal(map[string]string{"name": "Teste", "hostname": host})
		if w := request(h, "POST", "/api/v1/endpoints", string(body), cookie, session.CSRF, origin); w.Code != 400 {
			t.Fatal(w.Code)
		}
	}
	created := request(h, "POST", "/api/v1/endpoints", `{"name":"Teste","hostname":"EXAMPLE.COM."}`, cookie, session.CSRF, origin)
	if created.Code != 201 {
		t.Fatal(created.Code, created.Body.String())
	}
	var ep storage.Endpoint
	if e = json.Unmarshal(created.Body.Bytes(), &ep); e != nil {
		t.Fatal(e)
	}
	if ep.Hostname != "example.com" || ep.NextCheck == nil {
		t.Fatal(ep)
	}
	path := "/api/v1/endpoints/" + ep.ID
	if w := request(h, "PATCH", path, `{"name":"Renomeado","warning_days":31}`, cookie, session.CSRF, origin); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if w := request(h, "PATCH", path, `{"enabled":null}`, cookie, session.CSRF, origin); w.Code != 400 {
		t.Fatal(w.Code)
	}
	if w := request(h, "PATCH", path, `{"name|hostname":"invalid"}`, cookie, session.CSRF, origin); w.Code != 400 {
		t.Fatal("campo desconhecido aceito", w.Code)
	}
	if w := request(h, "POST", path+"/check", "", cookie, session.CSRF, origin); w.Code != 201 || !bytes.Contains(w.Body.Bytes(), []byte("blocked_target")) {
		t.Fatal(w.Code, w.Body.String())
	}
	if w := request(h, "POST", path+"/check", "", cookie, session.CSRF, origin); w.Code != 409 {
		t.Fatal(w.Code)
	}
	if w := request(h, "GET", path+"/checks", "", cookie, "", ""); w.Code != 200 || !strings.Contains(w.Body.String(), "blocked_target") {
		t.Fatal(w.Code)
	}
	if w := request(h, "DELETE", path, "", cookie, session.CSRF, origin); w.Code != 204 {
		t.Fatal(w.Code)
	}
	if w := request(h, "GET", path+"/checks", "", cookie, "", ""); w.Code != 200 {
		t.Fatal("histórico arquivado", w.Code)
	}
	if w := request(h, "POST", "/api/v1/auth/logout", "", cookie, session.CSRF, origin); w.Code != 204 {
		t.Fatal(w.Code)
	}
	if w := request(h, "GET", "/api/v1/auth/me", "", cookie, "", ""); w.Code != 401 {
		t.Fatal("sessão não revogada", w.Code)
	}
	for i := 0; i < 5; i++ {
		w := request(h, "POST", "/api/v1/auth/login", `{"email":"admin@example.com","password":"errada"}`, nil, "", origin)
		if i == 4 && w.Code != 429 {
			t.Fatal("rate limit ausente", w.Code)
		}
	}
}
