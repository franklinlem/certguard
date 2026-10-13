// Package auth implementa segredos de sessão e validação de credenciais.
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"net/mail"
	"strings"
	"sync"
	"time"
)

const SessionAge = 8 * time.Hour

func Email(raw string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(raw))
	a, e := mail.ParseAddress(email)
	if e != nil || a.Address != email || len(email) > 254 {
		return "", errors.New("e-mail inválido")
	}
	return email, nil
}
func HashPassword(password string) (string, error) {
	if len(password) < 12 || len(password) > 72 {
		return "", errors.New("senha deve ter entre 12 e 72 bytes")
	}
	hash, e := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(hash), e
}
func Verify(hash, password string) bool {
	if len(password) > 72 {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
func Token() string            { b := make([]byte, 32); _, _ = rand.Read(b); return hex.EncodeToString(b) }
func Hash(token string) string { s := sha256.Sum256([]byte(token)); return hex.EncodeToString(s[:]) }
func CSRF(token string) string { return Hash("certguard-csrf:" + token) }
func ValidToken(token string) bool {
	if len(token) != 64 {
		return false
	}
	_, e := hex.DecodeString(token)
	return e == nil
}

type entry struct {
	count int
	until time.Time
}

// Limiter não confia em headers de proxy. O limite global evita churn de chaves.
type Limiter struct {
	mu      sync.Mutex
	entries map[string]entry
	now     func() time.Time
}

func NewLimiter() *Limiter { return &Limiter{entries: make(map[string]entry), now: time.Now} }
func (l *Limiter) Allow(key string, limit int, window time.Duration) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	for k, v := range l.entries {
		if !now.Before(v.until) {
			delete(l.entries, k)
		}
	}
	e, ok := l.entries[key]
	if !ok {
		if len(l.entries) >= 4096 {
			return false
		}
		e = entry{until: now.Add(window)}
	}
	if e.count >= limit {
		return false
	}
	e.count++
	l.entries[key] = e
	return true
}
