package alerts

import (
	"bufio"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const testDelivery = "11111111-1111-4111-8111-111111111111"

type testDNS struct {
	ips   []netip.Addr
	calls atomic.Int32
}

func (r *testDNS) LookupNetIP(context.Context, string, string) ([]netip.Addr, error) {
	r.calls.Add(1)
	return r.ips, nil
}
func tlsFixture(t *testing.T) (*tls.Config, *x509.CertPool) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "notify.example.com"}, DNSNames: []string{"notify.example.com"}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour), IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}
	der, err := x509.CreateCertificate(rand.Reader, template, template, pub, priv)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	roots := x509.NewCertPool()
	roots.AddCert(cert)
	return &tls.Config{MinVersion: tls.VersionTLS12, Certificates: []tls.Certificate{{Certificate: [][]byte{der}, PrivateKey: priv}}}, roots
}
func payloadBytes() []byte {
	p := Payload{EventID: testDelivery, Event: "certificate_expiry_warning", Severity: "warning", Timestamp: time.Now().UTC()}
	p.Endpoint.Hostname = "example.com"
	p.Endpoint.Port = 443
	b, _ := json.Marshal(p)
	return b
}
func transportFor(t *testing.T, target string, roots *x509.CertPool, port string) (*Transport, *testDNS) {
	t.Helper()
	dns := &testDNS{ips: []netip.Addr{netip.MustParseAddr("93.184.216.34")}}
	tr := NewTransport()
	tr.resolver = dns
	tr.roots = roots
	tr.dial = func(ctx context.Context, network, addr string) (net.Conn, error) {
		if addr != "93.184.216.34:"+port {
			t.Error("dial não fixado", addr)
		}
		return (&net.Dialer{}).DialContext(ctx, network, target)
	}
	return tr, dns
}
func TestWebhookTLSAndSigning(t *testing.T) {
	serverTLS, roots := tlsFixture(t)
	cfg := hookConfig()
	body := payloadBytes()
	var status atomic.Int32
	status.Store(204)
	var requests atomic.Int32
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		raw, _ := io.ReadAll(r.Body)
		if r.Host != "notify.example.com" || r.TLS.ServerName != "notify.example.com" || string(raw) != string(body) || r.Header.Get("X-CertGuard-Delivery") != testDelivery || r.Header.Get("X-CertGuard-Signature") != signature(cfg.Secret, r.Header.Get("X-CertGuard-Timestamp"), raw) {
			t.Error("payload, assinatura, Host ou SNI incorreto")
		}
		w.Header().Set("Location", "https://127.0.0.1/private")
		w.WriteHeader(int(status.Load()))
	}))
	server.TLS = serverTLS
	server.StartTLS()
	defer server.Close()
	tr, dns := transportFor(t, server.Listener.Addr().String(), roots, "443")
	for _, v := range []struct {
		status    int
		ok, retry bool
	}{{204, true, false}, {503, false, true}, {429, false, true}, {400, false, false}, {302, false, false}} {
		status.Store(int32(v.status))
		out := tr.Send(context.Background(), "webhook", cfg, body, testDelivery)
		if out.Accepted != v.ok || out.Retry != v.retry {
			t.Fatal(v, out)
		}
	}
	if requests.Load() != 5 || dns.calls.Load() != 5 {
		t.Fatal("redirect ou resolução adicional", requests.Load(), dns.calls.Load())
	}
	tr.roots = nil
	if out := tr.Send(context.Background(), "webhook", cfg, body, testDelivery); out.Code != "tls_validation_error" || out.Retry {
		t.Fatal("TLS inválido aceito", out)
	}
}
func TestWebhookSSRF(t *testing.T) {
	tr := NewTransport()
	dns := &testDNS{ips: []netip.Addr{netip.MustParseAddr("93.184.216.34"), netip.MustParseAddr("127.0.0.1")}}
	tr.resolver = dns
	tr.dial = func(context.Context, string, string) (net.Conn, error) {
		t.Fatal("DNS misto chegou ao dial")
		return nil, nil
	}
	if out := tr.Send(context.Background(), "webhook", hookConfig(), payloadBytes(), testDelivery); out.Code != "blocked_target" {
		t.Fatal(out)
	}
	dns.ips = []netip.Addr{netip.MustParseAddr("169.254.169.254")}
	if out := tr.Send(context.Background(), "smtp", smtpConfig(587), payloadBytes(), testDelivery); out.Code != "blocked_target" {
		t.Fatal(out)
	}
}
func TestSMTPMandatoryTLSAndData(t *testing.T) {
	for _, port := range []int{465, 587, 25} {
		t.Run(fmt.Sprint(port), func(t *testing.T) {
			config, roots := tlsFixture(t)
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			defer listener.Close()
			received := make(chan string, 1)
			done := make(chan struct{})
			var authenticated atomic.Bool
			go func() {
				defer close(done)
				conn, err := listener.Accept()
				if err != nil {
					return
				}
				defer conn.Close()
				_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
				secure := port == 465
				if secure {
					c := tls.Server(conn, config)
					if c.Handshake() != nil {
						return
					}
					conn = c
				}
				reader := bufio.NewReader(conn)
				_, _ = io.WriteString(conn, "220 notify.example.com ESMTP\r\n")
				for {
					line, err := reader.ReadString('\n')
					if err != nil {
						return
					}
					switch {
					case strings.HasPrefix(line, "EHLO"):
						if port == 587 && !secure {
							_, _ = io.WriteString(conn, "250-notify.example.com\r\n250-STARTTLS\r\n250 AUTH PLAIN\r\n")
						} else {
							_, _ = io.WriteString(conn, "250-notify.example.com\r\n250 AUTH PLAIN\r\n")
						}
					case strings.HasPrefix(line, "STARTTLS"):
						_, _ = io.WriteString(conn, "220 Ready\r\n")
						c := tls.Server(conn, config)
						if c.Handshake() != nil {
							return
						}
						conn = c
						reader = bufio.NewReader(c)
						secure = true
					case strings.HasPrefix(line, "AUTH"):
						if !secure {
							t.Error("credenciais antes de TLS")
						}
						authenticated.Store(true)
						_, _ = io.WriteString(conn, "235 Authenticated\r\n")
					case strings.HasPrefix(line, "MAIL"), strings.HasPrefix(line, "RCPT"):
						_, _ = io.WriteString(conn, "250 OK\r\n")
					case strings.HasPrefix(line, "DATA"):
						_, _ = io.WriteString(conn, "354 Data\r\n")
						var data strings.Builder
						for {
							l, e := reader.ReadString('\n')
							if e != nil {
								return
							}
							if l == ".\r\n" {
								break
							}
							data.WriteString(l)
						}
						received <- data.String()
						_, _ = io.WriteString(conn, "250 Accepted\r\n")
					case strings.HasPrefix(line, "QUIT"):
						_, _ = io.WriteString(conn, "221 Bye\r\n")
						return
					}
				}
			}()
			effectivePort := port
			if port == 25 {
				effectivePort = 587
			} // Servidor sem STARTTLS na porta aprovada de teste.
			tr, _ := transportFor(t, listener.Addr().String(), roots, fmt.Sprint(effectivePort))
			out := tr.Send(context.Background(), "smtp", smtpConfig(effectivePort), payloadBytes(), testDelivery)
			if port == 25 {
				if out.Code != "starttls_required" || authenticated.Load() {
					t.Fatal(out)
				}
			} else {
				if !out.Accepted || !authenticated.Load() {
					t.Fatal(out)
				}
				select {
				case msg := <-received:
					if !strings.Contains(msg, "Message-ID: <"+testDelivery+"@example.com>") || !strings.Contains(msg, "Content-Transfer-Encoding: base64") {
						t.Fatal("mensagem sem identidade/encoding")
					}
				case <-time.After(time.Second):
					t.Fatal("DATA não recebido")
				}
			}
			select {
			case <-done:
			case <-time.After(time.Second):
				t.Fatal("SMTP não encerrou")
			}
		})
	}
}
