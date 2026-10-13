package alerts

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/franklinlem/certguard/api/internal/security"
	"net/mail"
	"net/url"
	"strings"
)

var ErrConfig = errors.New("configuração de canal inválida")

type ChannelConfig struct {
	URL      string `json:"url,omitempty"`
	Secret   string `json:"secret,omitempty"`
	Host     string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	From     string `json:"from,omitempty"`
	To       string `json:"to,omitempty"`
}

func (c *ChannelConfig) Validate(kind string) error {
	if kind == "webhook" {
		u, err := url.Parse(c.URL)
		if err != nil || u.Scheme != "https" || u.User != nil || u.Fragment != "" || u.Host == "" || (u.Port() != "" && u.Port() != "443") || len(c.URL) > 2048 || len(c.Secret) < 32 || len(c.Secret) > 1024 || c.Host != "" || c.Port != 0 || c.Username != "" || c.Password != "" || c.From != "" || c.To != "" {
			return ErrConfig
		}
		h, err := security.Normalize(u.Hostname())
		if err != nil {
			return ErrConfig
		}
		u.Host = h
		c.URL = u.String()
		return nil
	}
	if kind != "smtp" || c.URL != "" || c.Secret != "" || (c.Port != 465 && c.Port != 587) || len(c.Username) > 254 || len(c.Password) > 1024 || c.Username == "" || c.Password == "" || strings.ContainsAny(c.Username, "\r\n\x00") {
		return ErrConfig
	}
	host, err := security.Normalize(c.Host)
	if err != nil {
		return ErrConfig
	}
	c.Host = host
	for _, address := range []string{c.From, c.To} {
		a, e := mail.ParseAddress(address)
		if e != nil || a.Address != address || a.Name != "" || len(address) > 254 || strings.ContainsAny(address, "\r\n\x00") || strings.ContainsAny(address, "<>\" ") {
			return ErrConfig
		}
		for _, r := range address {
			if r > 127 {
				return ErrConfig
			}
		}
	}
	return nil
}

type Cipher struct{ aead cipher.AEAD }

func NewCipher(encoded string) (*Cipher, error) {
	key, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(key) != 32 {
		return nil, ErrConfig
	}
	block, err := aes.NewCipher(key)
	clear(key)
	if err != nil {
		return nil, ErrConfig
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, ErrConfig
	}
	return &Cipher{aead}, nil
}
func (c *Cipher) Seal(id, kind string, config ChannelConfig) ([]byte, error) {
	if c == nil {
		return nil, ErrConfig
	}
	raw, err := json.Marshal(config)
	if err != nil {
		return nil, ErrConfig
	}
	defer clear(raw)
	nonce := make([]byte, c.aead.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, ErrConfig
	}
	return c.aead.Seal(nonce, nonce, raw, []byte(id+":"+kind)), nil
}
func (c *Cipher) Open(id, kind string, sealed []byte) (ChannelConfig, error) {
	var config ChannelConfig
	if c == nil || len(sealed) < c.aead.NonceSize()+c.aead.Overhead() {
		return config, ErrConfig
	}
	n := c.aead.NonceSize()
	raw, err := c.aead.Open(nil, sealed[:n], sealed[n:], []byte(id+":"+kind))
	if err != nil || json.Unmarshal(raw, &config) != nil {
		return config, ErrConfig
	}
	if config.Validate(kind) != nil {
		return ChannelConfig{}, ErrConfig
	}
	return config, nil
}
