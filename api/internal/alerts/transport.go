package alerts

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/franklinlem/certguard/api/internal/certificate"
	"github.com/franklinlem/certguard/api/internal/security"
	"io"
	"net"
	"net/http"
	"net/smtp"
	"net/textproto"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Payload struct {
	EventID   string    `json:"event_id"`
	Event     string    `json:"event"`
	Severity  string    `json:"severity"`
	Timestamp time.Time `json:"timestamp"`
	Endpoint  struct {
		Name     string `json:"name"`
		Hostname string `json:"hostname"`
		Port     int    `json:"port"`
	} `json:"endpoint"`
	Certificate   *certificate.Certificate `json:"certificate,omitempty"`
	DaysRemaining *int                     `json:"days_remaining,omitempty"`
	Code          string                   `json:"code"`
}
type Outcome struct {
	Accepted bool
	Retry    bool
	Code     string
}
type Send func(context.Context, string, ChannelConfig, []byte, string) Outcome

// Transport não permite alterar resolução, dial ou CAs por variáveis de ambiente.
type Transport struct {
	resolver security.Resolver
	dial     func(context.Context, string, string) (net.Conn, error)
	roots    *x509.CertPool
}

func NewTransport() *Transport {
	return &Transport{resolver: net.DefaultResolver, dial: (&net.Dialer{}).DialContext}
}

var deliveryID = regexp.MustCompile(`^[0-9a-f-]{36}$`)
var eventName = regexp.MustCompile(`^[a-z_]{1,64}$`)

func (t *Transport) Send(ctx context.Context, kind string, c ChannelConfig, body []byte, id string) Outcome {
	if c.Validate(kind) != nil || !deliveryID.MatchString(id) || len(body) > 262144 {
		return Outcome{Code: "config_error"}
	}
	var p Payload
	if json.Unmarshal(body, &p) != nil || !eventName.MatchString(p.Event) {
		return Outcome{Code: "config_error"}
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	host := c.Host
	port := c.Port
	if kind == "webhook" {
		u, _ := url.Parse(c.URL)
		host = u.Hostname()
		port = 443
	}
	approved, ips, err := security.Resolve(ctx, t.resolver, host, 443)
	if err != nil {
		if errors.Is(err, security.ErrBlocked) {
			return Outcome{Code: "blocked_target"}
		}
		return Outcome{Code: "dns_error", Retry: true}
	}
	if !security.PublicIP(ips[0]) {
		return Outcome{Code: "blocked_target"}
	}
	target := net.JoinHostPort(ips[0].String(), strconv.Itoa(port))
	tlsConfig := &tls.Config{ServerName: approved, MinVersion: tls.VersionTLS12, RootCAs: t.roots}
	if kind == "webhook" {
		return t.webhook(ctx, c, body, id, p.Event, target, tlsConfig)
	}
	return t.email(ctx, c, body, id, target, tlsConfig)
}

func signature(secret, timestamp string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(timestamp + "."))
	_, _ = mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
func (t *Transport) webhook(ctx context.Context, c ChannelConfig, body []byte, id, event, target string, config *tls.Config) Outcome {
	tr := &http.Transport{Proxy: nil, TLSClientConfig: config, DisableKeepAlives: true, MaxResponseHeaderBytes: 16384, DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) { return t.dial(ctx, network, target) }}
	defer tr.CloseIdleConnections()
	client := &http.Client{Transport: tr, Timeout: 10 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	req, err := http.NewRequestWithContext(ctx, "POST", c.URL, bytes.NewReader(body))
	if err != nil {
		return Outcome{Code: "config_error"}
	}
	stamp := strconv.FormatInt(time.Now().Unix(), 10)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CertGuard-Event", event)
	req.Header.Set("X-CertGuard-Timestamp", stamp)
	req.Header.Set("X-CertGuard-Signature", signature(c.Secret, stamp, body))
	req.Header.Set("X-CertGuard-Delivery", id)
	res, err := client.Do(req)
	if err != nil {
		return failure(err)
	}
	defer res.Body.Close()
	// Não consumir corpo remoto: confirmação depende apenas do status, sem logar conteúdo.
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		return Outcome{Accepted: true}
	}
	retry := res.StatusCode >= 500 || res.StatusCode == 408 || res.StatusCode == 429
	return Outcome{Code: "http_" + strconv.Itoa(res.StatusCode), Retry: retry}
}
func failure(err error) Outcome {
	var certErr *tls.CertificateVerificationError
	if errors.As(err, &certErr) {
		return Outcome{Code: "tls_validation_error"}
	}
	var smtpErr *textproto.Error
	if errors.As(err, &smtpErr) {
		return Outcome{Code: "smtp_" + strconv.Itoa(smtpErr.Code), Retry: smtpErr.Code >= 400 && smtpErr.Code < 500}
	}
	return Outcome{Code: "network_error", Retry: true}
}

func (t *Transport) email(ctx context.Context, c ChannelConfig, body []byte, id, target string, config *tls.Config) Outcome {
	conn, err := t.dial(ctx, "tcp", target)
	if err != nil {
		return failure(err)
	}
	defer conn.Close()
	deadline, _ := ctx.Deadline()
	_ = conn.SetDeadline(deadline)
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer stop()
	var wire net.Conn = conn
	if c.Port == 465 {
		secure := tls.Client(conn, config)
		if err = secure.HandshakeContext(ctx); err != nil {
			return failure(err)
		}
		wire = secure
	}
	client, err := smtp.NewClient(wire, c.Host)
	if err != nil {
		return failure(err)
	}
	defer client.Close()
	if c.Port == 587 {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			return Outcome{Code: "starttls_required"}
		}
		if err = client.StartTLS(config); err != nil {
			return failure(err)
		}
	}
	if state, ok := client.TLSConnectionState(); !ok || !state.HandshakeComplete {
		return Outcome{Code: "starttls_required"}
	}
	if err = client.Auth(smtp.PlainAuth("", c.Username, c.Password, c.Host)); err != nil {
		return failure(err)
	}
	if err = client.Mail(c.From); err != nil {
		return failure(err)
	}
	if err = client.Rcpt(c.To); err != nil {
		return failure(err)
	}
	writer, err := client.Data()
	if err != nil {
		return failure(err)
	}
	message := mailMessage(c, body, id)
	if _, err = io.WriteString(writer, message); err != nil {
		return failure(err)
	}
	if err = writer.Close(); err != nil {
		return failure(err)
	}
	// DATA aceito já confirma entrega ao servidor; falha no QUIT não deve duplicar.
	_ = client.Quit()
	return Outcome{Accepted: true}
}
func mailMessage(c ChannelConfig, body []byte, id string) string {
	var p Payload
	_ = json.Unmarshal(body, &p)
	text := fmt.Sprintf("Alerta CertGuard\nEvento: %s\nSeveridade: %s\nDestino: %s:%d\nObservado em: %s\n\nObservação preservada em JSON:\n%s", p.Event, p.Severity, p.Endpoint.Hostname, p.Endpoint.Port, p.Timestamp.Format(time.RFC3339), body)
	encoded := base64.StdEncoding.EncodeToString([]byte(text))
	var lines strings.Builder
	for len(encoded) > 76 {
		lines.WriteString(encoded[:76] + "\r\n")
		encoded = encoded[76:]
	}
	lines.WriteString(encoded + "\r\n")
	domain := c.From[strings.LastIndex(c.From, "@")+1:]
	return fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: Alerta CertGuard\r\nDate: %s\r\nMessage-ID: <%s@%s>\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\nContent-Transfer-Encoding: base64\r\n\r\n%s", c.From, c.To, time.Now().UTC().Format(time.RFC1123Z), id, domain, lines.String())
}

func SafeCode(code string) string {
	switch code {
	case "config_error", "blocked_target", "dns_error", "network_error", "tls_validation_error", "starttls_required", "channel_paused", "channel_archived":
		return code
	}
	if matched, _ := regexp.MatchString(`^(http|smtp)_[1-5][0-9]{2}$`, code); matched {
		return code
	}
	return "delivery_error"
}
