package auth

import (
	"strings"
	"sync"
	"testing"
	"time"
)

func TestCredentialsAndTokens(t *testing.T) {
	for _, raw := range []string{"bad", "A <a@example.com>", strings.Repeat("a", 255) + "@x.com"} {
		if _, e := Email(raw); e == nil {
			t.Fatal(raw)
		}
	}
	if v, e := Email(" ADMIN@EXAMPLE.COM "); e != nil || v != "admin@example.com" {
		t.Fatal(v, e)
	}
	for _, pw := range []string{"short", strings.Repeat("a", 73)} {
		if _, e := HashPassword(pw); e == nil {
			t.Fatal("senha inválida aceita")
		}
	}
	h, e := HashPassword("Teste-local-12345")
	if e != nil || !Verify(h, "Teste-local-12345") || Verify(h, "errada") || Verify(h, strings.Repeat("a", 73)) {
		t.Fatal("verificação de senha")
	}
	a, b := Token(), Token()
	if a == b || !ValidToken(a) || Hash(a) == a || CSRF(a) == Hash(a) || ValidToken("invalid") || ValidToken(strings.Repeat("z", 64)) {
		t.Fatal("token inválido")
	}
}
func TestLimiter(t *testing.T) {
	l := NewLimiter()
	now := time.Now()
	l.now = func() time.Time { return now }
	for range 2 {
		if !l.Allow("a", 2, time.Minute) {
			t.Fatal("bloqueio antecipado")
		}
	}
	if l.Allow("a", 2, time.Minute) {
		t.Fatal("limite ignorado")
	}

	now = now.Add(time.Minute)
	if !l.Allow("a", 2, time.Minute) {
		t.Fatal("janela não expirou")
	}
	for i := range 4096 {
		l.entries[string(rune(i+100))] = entry{until: now.Add(time.Hour)}
	}
	if l.Allow("new", 2, time.Minute) {
		t.Fatal("mapa sem limite")
	}
}
func TestLimiterConcurrent(t *testing.T) {
	l := NewLimiter()
	var wg sync.WaitGroup
	var mu sync.Mutex
	allowed := 0
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if l.Allow("ip", 5, time.Minute) {
				mu.Lock()
				allowed++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if allowed != 5 {
		t.Fatal(allowed)
	}
}
