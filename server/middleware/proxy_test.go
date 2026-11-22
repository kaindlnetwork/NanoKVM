package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"NanoKVM-Server/config"
)

func TestTrustedProxyMatching(t *testing.T) {
	conf := config.GetInstance()
	original := conf.Security.TrustedProxies
	defer func() { conf.Security.TrustedProxies = original }()

	conf.Security.TrustedProxies = []string{"192.0.2.10", "2001:db8::/64", "not-an-ip"}
	for remote, want := range map[string]bool{
		"192.0.2.10:443":     true,
		"192.0.2.11:443":     false,
		"[2001:db8::7]:443":  true,
		"[2001:db8:1::]:443": false,
		"garbage":            false,
	} {
		if got := IsTrustedProxy(remote); got != want {
			t.Errorf("IsTrustedProxy(%q) = %v, want %v", remote, got, want)
		}
	}
}

func TestForwardedProtoTakesFirstValueOnly(t *testing.T) {
	conf := config.GetInstance()
	original := conf.Security.TrustedProxies
	defer func() { conf.Security.TrustedProxies = original }()
	conf.Security.TrustedProxies = []string{"192.0.2.10"}

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.RemoteAddr = "192.0.2.10:1"
	request.Header.Set("X-Forwarded-Proto", "HTTPS, http")
	if got := ForwardedProto(request); got != "https" {
		t.Fatalf("got %q", got)
	}
}
