package main

import (
	"errors"
	"strings"
	"testing"
)

type brokenReader struct{}

func (brokenReader) Read([]byte) (int, error) { return 0, errors.New("segredo do leitor") }

func TestReadBootstrapPassword(t *testing.T) {
	for _, tc := range []struct {
		name, input, want string
		valid             bool
	}{
		{"sem quebra", "Abc1234567", "Abc1234567", true},
		{"LF", "Abc1234567\n", "Abc1234567", true},
		{"CRLF no limite", strings.Repeat("a", 72) + "\r\n", strings.Repeat("a", 72), true},
		{"Unicode curto", strings.Repeat("á", 9), "", false},
		{"Unicode válido", strings.Repeat("á", 10), strings.Repeat("á", 10), true},
		{"curta", "Abc123456", "", false},
		{"longa", strings.Repeat("a", 73), "", false},
		{"entrada limitada", strings.Repeat("a", 10000), "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := readBootstrapPassword(strings.NewReader(tc.input))
			if (err == nil) != tc.valid || got != tc.want {
				t.Fatal("leitura ou validação inesperada")
			}
			if err != nil && strings.Contains(err.Error(), tc.input) {
				t.Fatal("senha exposta no erro")
			}
		})
	}
	if _, err := readBootstrapPassword(strings.NewReader("Abc123456")); err == nil || !strings.Contains(err.Error(), "10 caracteres") {
		t.Fatal("orientação de tamanho ausente")
	}
	if _, err := readBootstrapPassword(brokenReader{}); err == nil || strings.Contains(err.Error(), "segredo") {
		t.Fatal("erro de leitura deve ser seguro")
	}
}
