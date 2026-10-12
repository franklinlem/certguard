// Package certificate inspeciona TLS sem enviar dados de aplicação.
package certificate

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"math"
	"net"
	"strconv"
	"syscall"
	"time"

	"github.com/franklinlem/certguard/api/internal/security"
)

type Certificate struct {
	EmailAddresses []string  `json:"email_addresses"`
	URIs           []string  `json:"uris"`
	Subject        string    `json:"subject"`
	Issuer         string    `json:"issuer"`
	Serial         string    `json:"serial_number"`
	DNSNames       []string  `json:"dns_names"`
	IPAddresses    []string  `json:"ip_addresses"`
	NotBefore      time.Time `json:"not_before"`
	NotAfter       time.Time `json:"not_after"`
	Fingerprint    string    `json:"fingerprint_sha256"`
}
type Result struct {
	Hostname      string        `json:"hostname"`
	Port          int           `json:"port"`
	CheckedAt     time.Time     `json:"checked_at"`
	DurationMS    int64         `json:"duration_ms"`
	Status        string        `json:"status"`
	Code          string        `json:"code"`
	ResolvedIP    string        `json:"resolved_ip,omitempty"`
	TLSVersion    string        `json:"tls_version,omitempty"`
	CipherSuite   string        `json:"cipher_suite,omitempty"`
	Certificate   *Certificate  `json:"certificate,omitempty"`
	Chain         []Certificate `json:"chain,omitempty"`
	DaysRemaining *int          `json:"days_remaining,omitempty"`
	HostnameValid *bool         `json:"hostname_valid,omitempty"`
	ChainValid    *bool         `json:"chain_valid,omitempty"`
}

type Checker struct {
	timeout      time.Duration
	warningDays  int
	criticalDays int
	resolver     security.Resolver
	dial         func(context.Context, string, string) (net.Conn, error)
	roots        *x509.CertPool
	now          func() time.Time
}

func New(timeout time.Duration, warningDays, criticalDays int) (*Checker, error) {
	if timeout <= 0 || timeout > 30*time.Second || criticalDays < 0 || warningDays <= criticalDays || warningDays > 36500 {
		return nil, errors.New("timeout ou limiares inválidos")
	}
	return &Checker{timeout: timeout, warningDays: warningDays, criticalDays: criticalDays, resolver: net.DefaultResolver, dial: (&net.Dialer{}).DialContext, now: time.Now}, nil
}

func metadata(c *x509.Certificate) Certificate {
	sum := sha256.Sum256(c.Raw)
	ips := make([]string, 0, len(c.IPAddresses))
	for _, ip := range c.IPAddresses {
		ips = append(ips, ip.String())
	}
	uris := make([]string, 0, len(c.URIs))
	for _, uri := range c.URIs {
		uris = append(uris, uri.String())
	}
	return Certificate{EmailAddresses: append([]string{}, c.EmailAddresses...), URIs: uris, Subject: c.Subject.String(), Issuer: c.Issuer.String(), Serial: c.SerialNumber.Text(16), DNSNames: append([]string{}, c.DNSNames...), IPAddresses: ips, NotBefore: c.NotBefore, NotAfter: c.NotAfter, Fingerprint: hex.EncodeToString(sum[:])}
}

func (c *Checker) Check(ctx context.Context, hostname string, port int) (r Result) {
	start := time.Now()
	now := c.now().UTC()
	r = Result{Hostname: hostname, Port: port, CheckedAt: now, Status: "error", Code: "internal_error"}
	defer func() { r.DurationMS = time.Since(start).Milliseconds() }()
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	host, ips, err := security.Resolve(ctx, c.resolver, hostname, port)
	if err != nil {
		r.Code = "dns_error"
		if errors.Is(err, security.ErrBlocked) {
			r.Code = "blocked_target"
		}
		return
	}
	r.Hostname = host
	// Um IP por check. Nenhuma nova resolução; novo check revalida todos os registros.
	ip := ips[0]
	if !security.PublicIP(ip) {
		r.Code = "blocked_target"
		return
	}
	r.ResolvedIP = ip.String()
	raw, err := c.dial(ctx, "tcp", net.JoinHostPort(ip.String(), strconv.Itoa(port)))
	if err != nil {
		r.Code = "connection_error"
		if errors.Is(err, syscall.ECONNREFUSED) {
			r.Code = "connection_refused"
		}
		var ne net.Error
		if errors.As(err, &ne) && ne.Timeout() {
			r.Code = "connection_timeout"
		}
		return
	}
	defer raw.Close()
	deadline, _ := ctx.Deadline()
	if err = raw.SetDeadline(deadline); err != nil {
		return
	}
	// Inspeção isolada: não há HTTP, credenciais, sessões ou dados de aplicação.
	// InsecureSkipVerify permite coletar certificados inválidos. A confiança é
	// calculada OBRIGATORIAMENTE abaixo, por VerifyHostname e x509.Verify.
	conn := tls.Client(raw, &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12, InsecureSkipVerify: true}) // #nosec G402 -- ADR-003
	if err = conn.HandshakeContext(ctx); err != nil {
		r.Code = "tls_error"
		var ne net.Error
		if errors.As(err, &ne) && ne.Timeout() {
			r.Code = "connection_timeout"
		}
		return
	}
	state := conn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		r.Code = "tls_error"
		return
	}
	leaf := state.PeerCertificates[0]
	r.TLSVersion = tls.VersionName(state.Version)
	r.CipherSuite = tls.CipherSuiteName(state.CipherSuite)
	m := metadata(leaf)
	r.Certificate = &m
	for _, cert := range state.PeerCertificates {
		r.Chain = append(r.Chain, metadata(cert))
	}
	days := int(math.Floor(leaf.NotAfter.Sub(now).Hours() / 24))
	r.DaysRemaining = &days
	hv := leaf.VerifyHostname(host) == nil
	r.HostnameValid = &hv
	pool := x509.NewCertPool()
	for _, cert := range state.PeerCertificates[1:] {
		pool.AddCert(cert)
	}
	_, err = leaf.Verify(x509.VerifyOptions{Roots: c.roots, Intermediates: pool, CurrentTime: now, KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}})
	cv := err == nil
	r.ChainValid = &cv
	r.Status = "critical"
	switch {
	case now.Before(leaf.NotBefore):
		r.Code = "certificate_not_yet_valid"
	case now.After(leaf.NotAfter):
		r.Code = "certificate_expired"
	case !hv:
		r.Code = "hostname_mismatch"
	case !cv:
		r.Code = "untrusted_chain"
	default:
		r.Code = "success"
		r.Status = "healthy"
		remaining := leaf.NotAfter.Sub(now)
		if remaining <= time.Duration(c.criticalDays)*24*time.Hour {
			r.Status = "critical"
		} else if remaining <= time.Duration(c.warningDays)*24*time.Hour {
			r.Status = "warning"
		}
	}
	return
}
