package httpapi

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"github.com/franklinlem/certguard/api/internal/auth"
	"github.com/franklinlem/certguard/api/internal/certificate"
	"github.com/franklinlem/certguard/api/internal/storage"
	"github.com/jackc/pgx/v5"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func (a *App) cookieName() string {
	if a.secure {
		return "__Host-certguard_session"
	}
	return "certguard_session"
}

type App struct {
	store   *storage.Store
	origin  string
	secure  bool
	limiter *auth.Limiter
	logins  chan struct{}
	checks  chan struct{}
	dummy   string
	inspect storage.Inspect
}

func NewApp(store *storage.Store, origin string, local bool) (*App, error) {
	u, e := url.Parse(origin)
	if e != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return nil, errors.New("PUBLIC_APP_URL inválida")
	}
	if local {
		if u.Scheme != "http" && u.Scheme != "https" {
			return nil, errors.New("origem inválida")
		}
		if u.Hostname() != "localhost" && u.Hostname() != "127.0.0.1" {
			return nil, errors.New("HTTP local restrito a loopback")
		}
	} else if u.Scheme != "https" {
		return nil, errors.New("HTTPS obrigatório fora do desenvolvimento")
	}
	dummy, e := auth.HashPassword(auth.Token())
	if e != nil {
		return nil, e
	}
	a := &App{store: store, origin: u.Scheme + "://" + u.Host, secure: !local || u.Scheme == "https", limiter: auth.NewLimiter(), logins: make(chan struct{}, 2), checks: make(chan struct{}, 4), dummy: dummy}
	a.inspect = func(ctx context.Context, e storage.Endpoint) certificate.Result {
		c, _ := certificate.New(10*time.Second, e.Warning, e.Critical)
		return c.Check(ctx, e.Hostname, e.Port)
	}
	return a, nil
}
func send(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
func fail(w http.ResponseWriter, code int, message string) {
	send(w, code, map[string]string{"error": message})
}
func decode(w http.ResponseWriter, r *http.Request, v any) error {
	media, _, e := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if e != nil || media != "application/json" {
		return storage.ErrInvalid
	}
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192))
	d.DisallowUnknownFields()
	if e = d.Decode(v); e != nil {
		return storage.ErrInvalid
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return storage.ErrInvalid
	}
	return nil
}
func (a *App) setCookie(w http.ResponseWriter, token string, age int) {
	http.SetCookie(w, &http.Cookie{Name: a.cookieName(), Value: token, Path: "/", HttpOnly: true, Secure: a.secure, SameSite: http.SameSiteStrictMode, MaxAge: age})
}
func (a *App) tokenFrom(r *http.Request) string {
	c, e := r.Cookie(a.cookieName())
	if e != nil || !auth.ValidToken(c.Value) {
		return ""
	}
	return c.Value
}
func (a *App) session(w http.ResponseWriter, r *http.Request) (storage.User, string, bool) {
	token := a.tokenFrom(r)
	if token == "" {
		fail(w, 401, "Sessão ausente ou expirada.")
		return storage.User{}, "", false
	}
	u, e := a.store.Session(r.Context(), auth.Hash(token))
	if errors.Is(e, pgx.ErrNoRows) {
		fail(w, 401, "Sessão ausente ou expirada.")
		return u, "", false
	}
	if e != nil {
		fail(w, 503, "Serviço temporariamente indisponível.")
		return u, "", false
	}
	return u, token, true
}
func (a *App) protected(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, token, ok := a.session(w, r)
		if !ok {
			return
		}
		if r.Method != "GET" && r.Method != "HEAD" {
			expected := auth.CSRF(token)
			if subtle.ConstantTimeCompare([]byte(r.Header.Get("X-CSRF-Token")), []byte(expected)) != 1 {
				fail(w, 403, "Token de proteção inválido.")
				return
			}
		}
		next(w, r)
	}
}
func (a *App) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/", Handler(a.store.Pool))
	mux.HandleFunc("POST /api/v1/auth/login", a.login)
	mux.HandleFunc("GET /api/v1/auth/me", func(w http.ResponseWriter, r *http.Request) {
		u, t, ok := a.session(w, r)
		if ok {
			send(w, 200, map[string]any{"user": u, "csrf_token": auth.CSRF(t)})
		}
	})
	mux.HandleFunc("POST /api/v1/auth/logout", a.protected(func(w http.ResponseWriter, r *http.Request) {
		if e := a.store.Logout(r.Context(), auth.Hash(a.tokenFrom(r))); e != nil {
			fail(w, 503, "Não foi possível encerrar a sessão.")
			return
		}
		a.setCookie(w, "", -1)
		w.WriteHeader(204)
	}))
	mux.HandleFunc("GET /api/v1/endpoints", a.protected(a.list))
	mux.HandleFunc("POST /api/v1/endpoints", a.protected(a.create))
	mux.HandleFunc("GET /api/v1/endpoints/{id}", a.protected(a.get))
	mux.HandleFunc("PATCH /api/v1/endpoints/{id}", a.protected(a.update))
	mux.HandleFunc("DELETE /api/v1/endpoints/{id}", a.protected(a.archive))
	mux.HandleFunc("POST /api/v1/endpoints/{id}/check", a.protected(a.check))
	mux.HandleFunc("GET /api/v1/endpoints/{id}/checks", a.protected(a.history))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		if r.Method != "GET" && r.Method != "HEAD" {
			if r.Header.Get("Origin") != a.origin {
				fail(w, 403, "Origem não permitida.")
				return
			}
		}
		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
}
func (a *App) login(w http.ResponseWriter, r *http.Request) {
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	if !a.limiter.Allow("global", 60, time.Minute) || !a.limiter.Allow("ip:"+ip, 10, 15*time.Minute) {
		w.Header().Set("Retry-After", "900")
		fail(w, 429, "Muitas tentativas. Aguarde antes de tentar novamente.")
		return
	}
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if decode(w, r, &input) != nil {
		fail(w, 400, "Dados de acesso inválidos.")
		return
	}
	email, e := auth.Email(input.Email)
	if e != nil {
		fail(w, 401, "E-mail ou senha inválidos.")
		return
	}
	if !a.limiter.Allow("email:"+auth.Hash(email), 5, 15*time.Minute) {
		w.Header().Set("Retry-After", "900")
		fail(w, 429, "Muitas tentativas. Aguarde antes de tentar novamente.")
		return
	}
	select {
	case a.logins <- struct{}{}:
		defer func() { <-a.logins }()
	default:
		fail(w, 429, "Aguarde e tente novamente.")
		return
	}
	u, hash, e := a.store.Credentials(r.Context(), email)
	if e != nil && !errors.Is(e, pgx.ErrNoRows) {
		fail(w, 503, "Serviço temporariamente indisponível.")
		return
	}
	if e != nil {
		hash = a.dummy
	}
	valid := auth.Verify(hash, input.Password)
	if !valid || e != nil {
		fail(w, 401, "E-mail ou senha inválidos.")
		return
	}
	token := auth.Token()
	if e = a.store.NewSession(r.Context(), u.ID, auth.Hash(token)); e != nil {
		fail(w, 503, "Não foi possível iniciar a sessão.")
		return
	}
	a.setCookie(w, token, int(auth.SessionAge.Seconds()))
	send(w, 200, map[string]any{"user": u, "csrf_token": auth.CSRF(token)})
}

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func idFrom(w http.ResponseWriter, r *http.Request) (string, bool) {
	id := r.PathValue("id")
	if !uuidPattern.MatchString(id) {
		fail(w, 400, "Identificador inválido.")
		return "", false
	}
	return id, true
}
func pagination(w http.ResponseWriter, r *http.Request) (int, int, bool) {
	limit, offset := 50, 0
	for key, dest := range map[string]*int{"limit": &limit, "offset": &offset} {
		if raw := r.URL.Query().Get(key); raw != "" {
			n, e := strconv.Atoi(raw)
			if e != nil {
				fail(w, 400, "Paginação inválida.")
				return 0, 0, false
			}
			*dest = n
		}
	}
	if limit < 1 || limit > 100 || offset < 0 || offset > 100000 {
		fail(w, 400, "Paginação inválida.")
		return 0, 0, false
	}
	return limit, offset, true
}
func resultError(w http.ResponseWriter, e error) {
	switch {
	case errors.Is(e, pgx.ErrNoRows):
		fail(w, 404, "Registro não encontrado.")
	case errors.Is(e, storage.ErrInvalid):
		fail(w, 400, "Confira hostname, porta, nome e limiares.")
	case errors.Is(e, storage.ErrBusy), storage.Conflict(e):
		fail(w, 409, "Registro duplicado, em uso ou verificado há menos de 30 segundos.")
	case errors.Is(e, storage.ErrDisabled):
		fail(w, 409, "Ative o endpoint antes de verificar.")
	default:
		fail(w, 503, "Não foi possível concluir a operação.")
	}
}
func (a *App) list(w http.ResponseWriter, r *http.Request) {
	l, o, ok := pagination(w, r)
	if !ok {
		return
	}
	items, e := a.store.List(r.Context(), l, o)
	if e != nil {
		resultError(w, e)
		return
	}
	send(w, 200, map[string]any{"items": items, "limit": l, "offset": o})
}
func (a *App) create(w http.ResponseWriter, r *http.Request) {
	i := storage.DefaultInput()
	if decode(w, r, &i) != nil {
		resultError(w, storage.ErrInvalid)
		return
	}
	v, e := a.store.Create(r.Context(), i)
	if e != nil {
		resultError(w, e)
		return
	}
	send(w, 201, v)
}
func (a *App) get(w http.ResponseWriter, r *http.Request) {
	id, ok := idFrom(w, r)
	if !ok {
		return
	}
	v, e := a.store.Get(r.Context(), id)
	if e != nil {
		resultError(w, e)
		return
	}
	send(w, 200, v)
}
func (a *App) update(w http.ResponseWriter, r *http.Request) {
	id, ok := idFrom(w, r)
	if !ok {
		return
	}
	var patch map[string]json.RawMessage
	if decode(w, r, &patch) != nil || len(patch) == 0 {
		resultError(w, storage.ErrInvalid)
		return
	}
	for k, v := range patch {
		if !strings.Contains("|name|hostname|port|enabled|interval_minutes|warning_days|critical_days|", "|"+k+"|") || string(v) == "null" {
			resultError(w, storage.ErrInvalid)
			return
		}
	}
	v, e := a.store.Update(r.Context(), id, func(i *storage.Input) error {
		raw, _ := json.Marshal(patch)
		if json.Unmarshal(raw, i) != nil {
			return storage.ErrInvalid
		}
		return nil
	})
	if e != nil {
		resultError(w, e)
		return
	}
	send(w, 200, v)
}
func (a *App) archive(w http.ResponseWriter, r *http.Request) {
	id, ok := idFrom(w, r)
	if !ok {
		return
	}
	if e := a.store.Archive(r.Context(), id); e != nil {
		resultError(w, e)
		return
	}
	w.WriteHeader(204)
}
func (a *App) check(w http.ResponseWriter, r *http.Request) {
	id, ok := idFrom(w, r)
	if !ok {
		return
	}
	select {
	case a.checks <- struct{}{}:
		defer func() { <-a.checks }()
	default:
		fail(w, 429, "Limite de verificações simultâneas atingido.")
		return
	}
	v, e := a.store.Check(r.Context(), id, a.inspect)
	if e != nil {
		resultError(w, e)
		return
	}
	send(w, 201, v)
}
func (a *App) history(w http.ResponseWriter, r *http.Request) {
	id, ok := idFrom(w, r)
	if !ok {
		return
	}
	l, o, ok := pagination(w, r)
	if !ok {
		return
	}
	v, e := a.store.History(r.Context(), id, l, o)
	if e != nil {
		resultError(w, e)
		return
	}
	send(w, 200, map[string]any{"items": v, "limit": l, "offset": o})
}
