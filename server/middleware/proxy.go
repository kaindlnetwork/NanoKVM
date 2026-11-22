package middleware

import (
	"net"
	"net/http"
	"strconv"
	"strings"

	"NanoKVM-Server/config"
)

// trustedProxyNets parses security.trustedProxies (IPs or CIDRs). Invalid
// entries are ignored. An empty list means no proxy is trusted.
func trustedProxyNets() []*net.IPNet {
	entries := config.GetInstance().Security.TrustedProxies
	nets := make([]*net.IPNet, 0, len(entries))
	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if !strings.Contains(entry, "/") {
			if ip := net.ParseIP(entry); ip != nil {
				bits := 32
				if ip.To4() == nil {
					bits = 128
				}
				entry = ip.String() + "/" + strconv.Itoa(bits)
			}
		}
		if _, n, err := net.ParseCIDR(entry); err == nil {
			nets = append(nets, n)
		}
	}
	return nets
}

// IsTrustedProxy reports whether the direct peer of the request is a proxy
// listed in security.trustedProxies. Only then are X-Forwarded-* headers
// honoured; otherwise any client could claim to have arrived over HTTPS.
func IsTrustedProxy(remoteAddr string) bool {
	host := strings.TrimSpace(remoteAddr)
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	ip := net.ParseIP(strings.Trim(host, "[]"))
	if ip == nil {
		return false
	}
	for _, n := range trustedProxyNets() {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

// ForwardedProto returns the first X-Forwarded-Proto value if the request
// came through a trusted proxy, or "" otherwise.
func ForwardedProto(request *http.Request) string {
	if request == nil || !IsTrustedProxy(request.RemoteAddr) {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(strings.Split(request.Header.Get("X-Forwarded-Proto"), ",")[0]))
}
