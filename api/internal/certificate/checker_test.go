package certificate

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"errors"
	"math/big"
	"net"
	"net/netip"
	"sync"
	"syscall"
	"testing"
	"time"
)

type resolverFunc func(context.Context, string, string) ([]netip.Addr, error)

func (f resolverFunc) LookupNetIP(c context.Context, n, h string) ([]netip.Addr, error) {
	return f(c, n, h)
}

var fixedNow = time.Date(2026, 10, 12, 12, 0, 0, 0, time.UTC)

func fixture(t *testing.T, names []string, before, after time.Time) (tls.Certificate, *x509.CertPool) {
	t.Helper()
	key, e := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if e != nil {
		t.Fatal(e)
	}
	ca := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "CA de teste"}, NotBefore: fixedNow.AddDate(-1, 0, 0), NotAfter: fixedNow.AddDate(1, 0, 0), IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign}
	raw, e := x509.CreateCertificate(rand.Reader, ca, ca, &key.PublicKey, key)
	if e != nil {
		t.Fatal(e)
	}
	root, e := x509.ParseCertificate(raw)
	if e != nil {
		t.Fatal(e)
	}
	leaf := &x509.Certificate{SerialNumber: big.NewInt(2), DNSNames: names, NotBefore: before, NotAfter: after, KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}
	der, e := x509.CreateCertificate(rand.Reader, leaf, root, &key.PublicKey, key)
	if e != nil {
		t.Fatal(e)
	}
	pool := x509.NewCertPool()
	pool.AddCert(root)
	return tls.Certificate{Certificate: [][]byte{der, raw}, PrivateKey: key}, pool
}

// O dial de teste é privado ao pacote. Produção não pode habilitar destinos locais.
func localChecker(t *testing.T, cert tls.Certificate, roots *x509.CertPool) *Checker {
	t.Helper()
	ln, e := net.Listen("tcp", "127.0.0.1:0")
	if e != nil {
		t.Fatal(e)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			raw, e := ln.Accept()
			if e != nil {
				return
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer raw.Close()
				_ = raw.SetDeadline(time.Now().Add(2 * time.Second))
				conn := tls.Server(raw, &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12})
				_ = conn.Handshake()
			}()
		}
	}()
	t.Cleanup(func() { _ = ln.Close(); wg.Wait() })
	c, _ := New(time.Second, 30, 7)
	c.roots = roots
	c.now = func() time.Time { return fixedNow }
	c.resolver = resolverFunc(func(context.Context, string, string) ([]netip.Addr, error) {
		return []netip.Addr{netip.MustParseAddr("1.1.1.1")}, nil
	})
	c.dial = func(ctx context.Context, network, address string) (net.Conn, error) {
		if address != "1.1.1.1:443" {
			t.Errorf("dial não fixado: %s", address)
		}
		return (&net.Dialer{}).DialContext(ctx, network, ln.Addr().String())
	}
	return c
}
func TestTLSCertificates(t *testing.T) {
	for _, tc := range []struct {
		name          string
		names         []string
		before, after time.Duration
		untrusted     bool
		code, status  string
		hv, cv        bool
	}{
		{"válido", []string{"example.com"}, -time.Hour, 60 * 24 * time.Hour, false, "success", "healthy", true, true},
		{"expirado", []string{"example.com"}, -48 * time.Hour, -time.Second, false, "certificate_expired", "critical", true, false},
		{"futuro", []string{"example.com"}, time.Hour, 60 * 24 * time.Hour, false, "certificate_not_yet_valid", "critical", true, false},
		{"hostname", []string{"other.com"}, -time.Hour, 60 * 24 * time.Hour, false, "hostname_mismatch", "critical", false, true},
		{"CA desconhecida", []string{"example.com"}, -time.Hour, 60 * 24 * time.Hour, true, "untrusted_chain", "critical", true, false},
		{"SAN múltiplo", []string{"other.com", "example.com"}, -time.Hour, 60 * 24 * time.Hour, false, "success", "healthy", true, true},
		{"wildcard", []string{"*.example.com"}, -time.Hour, 60 * 24 * time.Hour, false, "success", "healthy", true, true},
		{"warning limite", []string{"example.com"}, -time.Hour, 30 * 24 * time.Hour, false, "success", "warning", true, true},
		{"critical limite", []string{"example.com"}, -time.Hour, 7 * 24 * time.Hour, false, "success", "critical", true, true},
		{"fora warning", []string{"example.com"}, -time.Hour, 30*24*time.Hour + time.Second, false, "success", "healthy", true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cert, roots := fixture(t, tc.names, fixedNow.Add(tc.before), fixedNow.Add(tc.after))
			if tc.untrusted {
				roots = x509.NewCertPool()
			}
			c := localChecker(t, cert, roots)
			host := "example.com"
			if tc.name == "wildcard" {
				host = "api.example.com"
			}
			r := c.Check(context.Background(), host, 443)
			if r.Code != tc.code || r.Status != tc.status {
				t.Fatalf("%+v", r)
			}
			if r.HostnameValid == nil || *r.HostnameValid != tc.hv || r.ChainValid == nil || *r.ChainValid != tc.cv {
				t.Fatalf("validação: %+v", r)
			}
			sum := sha256.Sum256(cert.Certificate[0])
			if r.Certificate.Fingerprint != hex.EncodeToString(sum[:]) || len(r.Chain) != 2 || r.TLSVersion == "" {
				t.Fatal("metadados incompletos")
			}
			if tc.name == "expirado" && *r.DaysRemaining != -1 {
				t.Fatal(*r.DaysRemaining)
			}
		})
	}
}
func TestBlockedNeverDials(t *testing.T) {
	c, _ := New(time.Second, 30, 7)
	c.dial = func(context.Context, string, string) (net.Conn, error) {
		t.Fatal("SSRF alcançou dial")
		return nil, nil
	}
	for _, ip := range []string{"127.0.0.1", "10.0.0.1", "::1", "fc00::1", "169.254.169.254"} {
		c.resolver = resolverFunc(func(context.Context, string, string) ([]netip.Addr, error) {
			return []netip.Addr{netip.MustParseAddr("1.1.1.1"), netip.MustParseAddr(ip)}, nil
		})
		if r := c.Check(context.Background(), "example.com", 443); r.Code != "blocked_target" {
			t.Fatal(r)
		}
	}
	if r := c.Check(context.Background(), "example.com", 80); r.Code != "blocked_target" {
		t.Fatal(r)
	}
}
func TestRebindingAndSNI(t *testing.T) {
	cert, roots := fixture(t, []string{"example.com"}, fixedNow.Add(-time.Hour), fixedNow.Add(60*24*time.Hour))
	c := localChecker(t, cert, roots)
	calls := 0
	c.resolver = resolverFunc(func(context.Context, string, string) ([]netip.Addr, error) {
		calls++
		ip := "1.1.1.1"
		if calls > 1 {
			ip = "127.0.0.1"
		}
		return []netip.Addr{netip.MustParseAddr(ip)}, nil
	})
	c.dial = func(ctx context.Context, n, a string) (net.Conn, error) {
		client, server := net.Pipe()
		go func() {
			defer server.Close()
			_ = server.SetDeadline(time.Now().Add(time.Second))
			s := tls.Server(server, &tls.Config{Certificates: []tls.Certificate{cert}, GetConfigForClient: func(hello *tls.ClientHelloInfo) (*tls.Config, error) {
				if hello.ServerName != "example.com" {
					t.Errorf("SNI: %s", hello.ServerName)
				}
				return nil, nil
			}})
			_ = s.Handshake()
		}()
		if a != "1.1.1.1:443" {
			t.Fatal(a)
		}
		return client, nil
	}
	if r := c.Check(context.Background(), "example.com", 443); r.Code != "success" || calls != 1 {
		t.Fatal(r, calls)
	}
	if r := c.Check(context.Background(), "example.com", 443); r.Code != "blocked_target" || calls != 2 {
		t.Fatal(r, calls)
	}
}
func TestFailuresAndTimeout(t *testing.T) {
	c, _ := New(30*time.Millisecond, 30, 7)
	c.resolver = resolverFunc(func(ctx context.Context, _, _ string) ([]netip.Addr, error) { <-ctx.Done(); return nil, ctx.Err() })
	start := time.Now()
	if r := c.Check(context.Background(), "example.com", 443); r.Code != "dns_error" {
		t.Fatal(r)
	}
	if time.Since(start) > time.Second {
		t.Fatal("timeout DNS não aplicado")
	}
	c.resolver = resolverFunc(func(context.Context, string, string) ([]netip.Addr, error) {
		return []netip.Addr{netip.MustParseAddr("1.1.1.1")}, nil
	})
	for _, tc := range []struct {
		err  error
		code string
	}{{syscall.ECONNREFUSED, "connection_refused"}, {context.DeadlineExceeded, "connection_timeout"}, {errors.New("rede indisponível"), "connection_error"}} {
		c.dial = func(context.Context, string, string) (net.Conn, error) { return nil, tc.err }
		if r := c.Check(context.Background(), "example.com", 443); r.Code != tc.code {
			t.Fatal(r)
		}
	}
	c.dial = func(context.Context, string, string) (net.Conn, error) {
		a, b := net.Pipe()
		go func() { defer b.Close(); time.Sleep(100 * time.Millisecond) }()
		return a, nil
	}
	if r := c.Check(context.Background(), "example.com", 443); r.Code != "connection_timeout" {
		t.Fatal(r)
	}
	c.dial = func(context.Context, string, string) (net.Conn, error) {
		a, b := net.Pipe()
		go func() {
			defer b.Close()
			_ = b.SetDeadline(time.Now().Add(time.Second))
			buf := make([]byte, 4096)
			_, _ = b.Read(buf)
			_, _ = b.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
		}()
		return a, nil
	}
	if r := c.Check(context.Background(), "example.com", 443); r.Code != "tls_error" {
		t.Fatal(r)
	}
}
func TestConfig(t *testing.T) {
	for _, tc := range []struct {
		d    time.Duration
		w, k int
	}{{0, 30, 7}, {31 * time.Second, 30, 7}, {time.Second, 7, 7}, {time.Second, 30, -1}} {
		if _, e := New(tc.d, tc.w, tc.k); e == nil {
			t.Fatal(tc)
		}
	}
}
