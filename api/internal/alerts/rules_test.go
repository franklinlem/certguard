package alerts

import (
	"github.com/franklinlem/certguard/api/internal/certificate"
	"testing"
)

func observation(status string, days int) certificate.Result {
	yes := true
	return certificate.Result{Hostname: "example.com", Port: 443, Code: "success", Status: status, DaysRemaining: &days, HostnameValid: &yes, ChainValid: &yes, Certificate: &certificate.Certificate{Fingerprint: "a"}}
}
func TestRules(t *testing.T) {
	if Evaluate(observation("healthy", 90), 30, 7) != nil {
		t.Fatal("saudável gerou alerta")
	}
	for _, v := range []struct {
		status         string
		days, critical int
		kind           string
		threshold      int
	}{
		{"warning", 20, 7, "certificate_expiry_warning", 45},
		{"critical", 5, 7, "certificate_expiry_critical", 7},
		{"critical", 3, 7, "certificate_expiry_reminder", 3},
		{"critical", 0, 7, "certificate_expiry_reminder", 3},
		{"critical", -1, 7, "certificate_expired", 0},
		{"critical", 2, 2, "certificate_expiry_critical", 2},
	} {
		rule := Evaluate(observation(v.status, v.days), 45, v.critical)
		if rule == nil || rule.Type != v.kind || rule.Threshold != v.threshold {
			t.Fatal(v, rule)
		}
	}
	for _, code := range []string{"certificate_expired", "certificate_not_yet_valid", "hostname_mismatch", "untrusted_chain", "blocked_target", "dns_error", "connection_timeout", "connection_refused", "connection_error", "tls_error", "internal_error"} {
		r := observation("error", 0)
		r.Code = code
		if rule := Evaluate(r, 30, 7); rule == nil || rule.Type != code || rule.Severity != "critical" {
			t.Fatal(code, rule)
		}
	}
	r := observation("healthy", 90)
	r.ChainValid = nil
	if rule := Evaluate(r, 30, 7); rule == nil || rule.Type != "invalid_validation" {
		t.Fatal(rule)
	}
	r = observation("warning", 20)
	r.DaysRemaining = nil
	if rule := Evaluate(r, 30, 7); rule == nil || rule.Type != "invalid_validity" {
		t.Fatal(rule)
	}
}
func TestKey(t *testing.T) {
	a := observation("warning", 20)
	rule := *Evaluate(a, 30, 7)
	key := Key(a, rule)
	b := observation("warning", 19)
	if Key(b, rule) != key {
		t.Fatal("cada dia gera repetição")
	}
	b.Certificate.Fingerprint = "b"
	if Key(b, rule) == key {
		t.Fatal("certificado novo ignorado")
	}
	b = observation("warning", 20)
	b.Hostname = "other.com"
	if Key(b, rule) == key {
		t.Fatal("destino ignorado")
	}
	rule.Threshold = 45
	if Key(a, rule) == key {
		t.Fatal("limiar ignorado")
	}
}
