// Package security aplica a política de saída de rede do CertGuard.
package security

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"strings"
)

var ErrBlocked = errors.New("destino bloqueado pela política SSRF")

// Normalize aceita somente nomes DNS ASCII completos. IPs literais e URLs são rejeitados.
// IDNs devem ser fornecidos como A-label (punycode). Não aceita porta junto ao nome.
func Normalize(raw string) (string, error) {
	h := strings.ToLower(strings.TrimSuffix(raw, "."))
	if len(h) > 253 || len(h) == 0 || !strings.Contains(h, ".") {
		return "", ErrBlocked
	}
	if _, err := netip.ParseAddr(h); err == nil {
		return "", ErrBlocked
	}
	for _, label := range strings.Split(h, ".") {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return "", ErrBlocked
		}
		for _, c := range label {
			if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-') {
				return "", ErrBlocked
			}
		}
	}
	tld := h[strings.LastIndex(h, ".")+1:]
	letters := false
	for _, c := range tld {
		if c >= 'a' && c <= 'z' {
			letters = true
		}
	}
	if !letters {
		return "", ErrBlocked
	}
	for _, suffix := range []string{"localhost", "local", "internal", "home.arpa", "test", "invalid", "example", "onion"} {
		if h == suffix || strings.HasSuffix(h, "."+suffix) {
			return "", ErrBlocked
		}
	}
	return h, nil
}

// Política conservadora: bloqueia também blocos especiais globalmente alcançáveis.
// Revisar junto aos registros IANA quando atualizar a política (ADR-002).
var blocked = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"), netip.MustParsePrefix("10.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"), netip.MustParsePrefix("127.0.0.0/8"),
	netip.MustParsePrefix("169.254.0.0/16"), netip.MustParsePrefix("172.16.0.0/12"),
	netip.MustParsePrefix("192.0.0.0/24"), netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("192.31.196.0/24"), netip.MustParsePrefix("192.52.193.0/24"),
	netip.MustParsePrefix("192.88.99.0/24"), netip.MustParsePrefix("192.168.0.0/16"),
	netip.MustParsePrefix("192.175.48.0/24"), netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"), netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("224.0.0.0/4"), netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("2001::/23"), netip.MustParsePrefix("2001:db8::/32"),
	netip.MustParsePrefix("2002::/16"), netip.MustParsePrefix("2620:4f:8000::/48"),
	netip.MustParsePrefix("3fff::/20"),
	// Endpoint de plataforma Azure, embora fora de um bloco privado.
	netip.MustParsePrefix("168.63.129.16/32"),
}

func PublicIP(ip netip.Addr) bool {
	if !ip.IsValid() || ip.Zone() != "" || ip.Is4In6() || !ip.IsGlobalUnicast() || ip.IsPrivate() {
		return false
	}
	if ip.Is6() && !netip.MustParsePrefix("2000::/3").Contains(ip) {
		return false
	}
	for _, p := range blocked {
		if p.Contains(ip) {
			return false
		}
	}
	return true
}

type Resolver interface {
	LookupNetIP(context.Context, string, string) ([]netip.Addr, error)
}

// Resolve valida o conjunto completo ANTES de liberar qualquer endereço para conexão.
func Resolve(ctx context.Context, resolver Resolver, host string, port int) (string, []netip.Addr, error) {
	h, err := Normalize(host)
	if err != nil || port != 443 {
		return "", nil, ErrBlocked
	}
	if resolver == nil {
		resolver = net.DefaultResolver
	}
	ips, err := resolver.LookupNetIP(ctx, "ip", h+".")
	if err != nil {
		return h, nil, fmt.Errorf("resolução DNS: %w", err)
	}
	if len(ips) == 0 {
		return h, nil, errors.New("DNS sem endereços")
	}
	if len(ips) > 64 {
		return h, nil, ErrBlocked
	}
	for _, ip := range ips {
		if !PublicIP(ip) {
			return h, nil, ErrBlocked
		}
	}
	return h, append([]netip.Addr(nil), ips...), nil
}
