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

func TestPasswordCreationPolicy(t *testing.T) {
	for _, tc := range []struct {
		name     string
		password string
		valid    bool
	}{
		{"nove caracteres", "Abc123456", false},
		{"dez caracteres", "Abc1234567", true},
		{"nove caracteres Unicode", strings.Repeat("á", 9), false},
		{"dez caracteres Unicode", strings.Repeat("á", 10), true},
		{"símbolos permitidos", "Teste-123!", true},
		{"limite bcrypt", strings.Repeat("a", 72), true},
		{"acima do limite bcrypt", strings.Repeat("a", 73), false},
		{"Unicode no limite bcrypt", strings.Repeat("á", 36), true},
		{"Unicode acima do limite bcrypt", strings.Repeat("á", 37), false},
		{"UTF-8 inválido", strings.Repeat("a", 10) + "\xff", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePassword(tc.password)
			if (err == nil) != tc.valid {
				t.Fatal("resultado inesperado de validação")
			}
			if err != nil && strings.Contains(err.Error(), "bytes") {
				t.Fatal("mensagem técnica exposta ao usuário")
			}
		})
	}
	// Exercise hashing as well as validation for the newly accepted minimum.
	for _, password := range []string{"Abc1234567", strings.Repeat("á", 10)} {
		hash, err := HashPassword(password)
		if err != nil || !Verify(hash, password) {
			t.Fatal("senha válida não pôde ser usada")
		}
	}
}
