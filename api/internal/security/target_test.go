package security

import (
	"context"
	"errors"
	"net/netip"
	"strings"
	"testing"
)

type resolverFunc func(context.Context, string, string) ([]netip.Addr, error)

func (f resolverFunc) LookupNetIP(c context.Context, n, h string) ([]netip.Addr, error) {
	return f(c, n, h)
}

func TestNormalize(t *testing.T) {
	for _, s := range []string{"example.com", "EXAMPLE.COM.", "xn--bcher-kva.de", "api.example.com"} {
		if _, err := Normalize(s); err != nil {
			t.Errorf("%s: %v", s, err)
		}
	}
	for _, s := range []string{"", "localhost", "foo.localhost", "foo.local", "metadata.google.internal", "host.home.arpa", "foo.test", "127.0.0.1", "::1", "[::1]", "2130706433", "0177.0.0.1", "0x7f000001", "https://example.com", "example.com:443", "u@example.com", " example.com", "example.com\n", "example.com/", "a..com", "-a.com", "a-.com", "*.com", "bücher.de", "a.123", strings.Repeat("a", 64) + ".com", strings.Repeat("a.", 127) + "com"} {
		if _, err := Normalize(s); err == nil {
			t.Errorf("aceitou %q", s)
		}
	}
	h, _ := Normalize("EXAMPLE.COM.")
	if h != "example.com" {
		t.Fatal(h)
	}
}
func TestPublicIP(t *testing.T) {
	for _, s := range []string{"8.8.8.8", "1.1.1.1", "2606:4700:4700::1111", "2001:4860:4860::8888"} {
		if !PublicIP(netip.MustParseAddr(s)) {
			t.Errorf("bloqueou %s", s)
		}
	}
	for _, s := range []string{"0.0.0.0", "10.1.2.3", "127.0.0.1", "172.16.0.1", "172.31.255.255", "192.168.1.1", "169.254.169.254", "100.100.100.200", "192.0.0.9", "192.0.2.1", "192.31.196.1", "192.52.193.1", "192.88.99.1", "192.175.48.1", "198.18.0.1", "198.51.100.1", "203.0.113.1", "224.0.0.1", "240.0.0.1", "255.255.255.255", "168.63.129.16", "::", "::1", "::ffff:127.0.0.1", "::ffff:8.8.8.8", "fe80::1", "fe80::1%eth0", "fc00::1", "ff02::1", "64:ff9b::808:808", "2001::1", "2001:db8::1", "2002:808:808::1", "3fff::1", "2620:4f:8000::1", "4000::1"} {
		if PublicIP(netip.MustParseAddr(s)) {
			t.Errorf("aceitou %s", s)
		}
	}
	if PublicIP(netip.Addr{}) {
		t.Fatal("IP vazio permitido")
	}
}
func TestResolve(t *testing.T) {
	for _, tc := range []struct {
		name    string
		ips     []netip.Addr
		err     error
		blocked bool
	}{
		{"público", []netip.Addr{netip.MustParseAddr("1.1.1.1")}, nil, false},
		{"misto", []netip.Addr{netip.MustParseAddr("1.1.1.1"), netip.MustParseAddr("10.0.0.1")}, nil, true},
		{"misto ipv6", []netip.Addr{netip.MustParseAddr("1.1.1.1"), netip.MustParseAddr("::1")}, nil, true},
		{"vazio", nil, nil, true}, {"erro", nil, errors.New("DNS falhou"), true},
		{"excesso", make([]netip.Addr, 65), nil, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := resolverFunc(func(_ context.Context, n, h string) ([]netip.Addr, error) {
				if n != "ip" || h != "example.com." {
					t.Fatal(n, h)
				}
				return tc.ips, tc.err
			})
			_, _, err := Resolve(context.Background(), r, "EXAMPLE.COM", 443)
			if (err != nil) != tc.blocked {
				t.Fatal(err)
			}
		})
	}
	for _, p := range []int{-1, 0, 80, 444, 65536} {
		_, _, err := Resolve(context.Background(), nil, "example.com", p)
		if !errors.Is(err, ErrBlocked) {
			t.Fatal(p)
		}
	}
	_, _, err := Resolve(context.Background(), nil, "localhost", 443)
	if !errors.Is(err, ErrBlocked) {
		t.Fatal(err)
	}
}
func FuzzNormalize(f *testing.F) {
	for _, s := range []string{"example.com", "::1", "a\x00.com", "a..com"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		h, e := Normalize(s)
		if e == nil {
			again, e := Normalize(h)
			if e != nil || again != h {
				t.Fatal("normalização instável")
			}
		}
	})
}
