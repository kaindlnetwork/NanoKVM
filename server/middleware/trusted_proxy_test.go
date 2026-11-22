package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"NanoKVM-Server/config"
)

// The WebSocket origin check took the scheme from X-Forwarded-Proto of any
// client. Only a configured proxy may say the request arrived over HTTPS.
func TestWebSocketOriginIgnoresSpoofedForwardedProto(t *testing.T) {
	conf := config.GetInstance()
	original := conf.Security.TrustedProxies
	defer func() { conf.Security.TrustedProxies = original }()

	request := httptest.NewRequest(http.MethodGet, "http://kvm.local/api/ws", nil)
	request.Host = "kvm.local"
	request.RemoteAddr = "192.0.2.10:5000"
	request.Header.Set("Origin", "https://kvm.local")
	request.Header.Set("X-Forwarded-Proto", "https")

	conf.Security.TrustedProxies = nil
	if CheckWebSocketOrigin(request) {
		t.Error("https origin accepted on a plain http request because of a spoofed header")
	}
	conf.Security.TrustedProxies = []string{"192.0.2.0/24"}
	if !CheckWebSocketOrigin(request) {
		t.Error("origin rejected behind a configured TLS proxy")
	}
}
