// Package alerts classifica observações TLS sem executar entregas externas.
package alerts

import (
	"crypto/sha256"
	"fmt"
	"github.com/franklinlem/certguard/api/internal/certificate"
)

type Rule struct {
	Type      string `json:"event_type"`
	Severity  string `json:"severity"`
	Threshold int    `json:"threshold_days"`
}

// Evaluate retorna nil somente quando a inspeção confirmou saúde e confiança.
func Evaluate(r certificate.Result, warning, critical int) *Rule {
	if r.Code == "success" {
		if r.HostnameValid == nil || r.ChainValid == nil || !*r.HostnameValid || !*r.ChainValid {
			return &Rule{Type: "invalid_validation", Severity: "critical"}
		}
		if r.Status == "healthy" {
			return nil
		}
		if r.DaysRemaining == nil {
			return &Rule{Type: "invalid_validity", Severity: "critical"}
		}
		if r.Status == "warning" {
			return &Rule{Type: "certificate_expiry_warning", Severity: "warning", Threshold: warning}
		}
		if r.Status == "critical" {
			if *r.DaysRemaining < 0 {
				return &Rule{Type: "certificate_expired", Severity: "critical"}
			}
			if *r.DaysRemaining <= 3 && critical >= 3 {
				return &Rule{Type: "certificate_expiry_reminder", Severity: "critical", Threshold: 3}
			}
			return &Rule{Type: "certificate_expiry_critical", Severity: "critical", Threshold: critical}
		}
	}
	kind := r.Code
	if kind == "" || kind == "success" {
		kind = "invalid_result"
	}
	return &Rule{Type: kind, Severity: "critical"}
}

func Key(r certificate.Result, rule Rule) string {
	fingerprint := ""
	if r.Certificate != nil {
		fingerprint = r.Certificate.Fingerprint
	}
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s:%d:%s:%s:%d:%s", r.Hostname, r.Port, rule.Type, rule.Severity, rule.Threshold, fingerprint)))
	return fmt.Sprintf("%x", sum)
}
