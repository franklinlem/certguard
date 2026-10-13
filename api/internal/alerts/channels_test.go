package alerts

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"
	"time"
)

func testCipher(t *testing.T) *Cipher {
	t.Helper()
	c, err := NewCipher(base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{7}, 32)))
	if err != nil {
		t.Fatal(err)
	}
	return c
}
func hookConfig() ChannelConfig {
	return ChannelConfig{URL: "https://notify.example.com/hook", Secret: strings.Repeat("s", 32)}
}
func smtpConfig(port int) ChannelConfig {
	return ChannelConfig{Host: "notify.example.com", Port: port, Username: "usuario", Password: "senha-somente-teste", From: "alertas@example.com", To: "admin@example.com"}
}
func TestChannelValidationAndCipher(t *testing.T) {
	for _, mutate := range []func(*ChannelConfig){func(c *ChannelConfig) { c.URL = "http://notify.example.com/hook" }, func(c *ChannelConfig) { c.URL = "https://user:pass@notify.example.com/" }, func(c *ChannelConfig) { c.URL = "https://notify.example.com:8443/" }, func(c *ChannelConfig) { c.URL = "https://127.0.0.1/" }, func(c *ChannelConfig) { c.URL += "#secret" }, func(c *ChannelConfig) { c.Secret = "short" }, func(c *ChannelConfig) { c.Host = "notify.example.com" }} {
		c := hookConfig()
		mutate(&c)
		if c.Validate("webhook") == nil {
			t.Fatal("config inválida aceita")
		}
	}
	for _, mutate := range []func(*ChannelConfig){func(c *ChannelConfig) { c.Port = 25 }, func(c *ChannelConfig) { c.Host = "localhost" }, func(c *ChannelConfig) { c.From = "a@example.com\r\nBcc: b@example.com" }, func(c *ChannelConfig) { c.To = "Nome <b@example.com>" }, func(c *ChannelConfig) { c.Username = "bad\nuser" }, func(c *ChannelConfig) { c.Password = "" }, func(c *ChannelConfig) { c.URL = "https://notify.example.com/" }} {
		c := smtpConfig(587)
		mutate(&c)
		if c.Validate("smtp") == nil {
			t.Fatal("SMTP inválido aceito")
		}
	}
	c := testCipher(t)
	cfg := hookConfig()
	a, err := c.Seal("id", "webhook", cfg)
	if err != nil {
		t.Fatal(err)
	}
	b, err := c.Seal("id", "webhook", cfg)
	if err != nil || bytes.Equal(a, b) || bytes.Contains(a, []byte(cfg.Secret)) {
		t.Fatal("nonce repetido ou segredo exposto")
	}
	if got, err := c.Open("id", "webhook", a); err != nil || got != cfg {
		t.Fatal(got, err)
	}
	for _, scope := range [][2]string{{"other", "webhook"}, {"id", "smtp"}} {
		if _, err = c.Open(scope[0], scope[1], a); err == nil {
			t.Fatal("AAD ignorado")
		}
	}
	a[len(a)-1] ^= 1
	if _, err = c.Open("id", "webhook", a); err == nil {
		t.Fatal("adulteração aceita")
	}
	if _, err = NewCipher("invalid"); err == nil {
		t.Fatal("chave inválida aceita")
	}
	for n := 1; n <= 4; n++ {
		d := RetryDelay(n)
		base := []time.Duration{time.Minute, 5 * time.Minute, 15 * time.Minute, time.Hour}[n-1]
		if d < base || d > base+base/10 {
			t.Fatal(d)
		}
	}
	if SafeCode("remote: segredo\n") != "delivery_error" {
		t.Fatal("erro bruto persistido")
	}
}
