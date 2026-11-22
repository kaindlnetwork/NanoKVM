package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// An https update source must not be redirected to plain http: on that hop
// an on-path attacker could swap latest.json and the package together.
// Upstream only refused the downgrade when credentials were attached.
func TestUpdateClientRefusesHTTPSToHTTPRedirect(t *testing.T) {
	reached := false
	plain := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		_, _ = w.Write([]byte(`{"version":"9.9.9"}`))
	}))
	defer plain.Close()
	secure := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, plain.URL+"/latest.json", http.StatusFound)
	}))
	defer secure.Close()

	client := NewUpdateHTTPClient(5 * time.Second)
	client.Transport = secure.Client().Transport

	request, err := NewAuthenticatedRequest(http.MethodGet, secure.URL+"/latest.json", nil)
	if err != nil {
		t.Fatal(err)
	}
	response, err := client.Do(request)
	if err == nil {
		_ = response.Body.Close()
	}
	if reached {
		t.Fatal("followed https -> http redirect without credentials")
	}
	if err == nil && response.StatusCode == http.StatusOK {
		t.Fatal("downgrade answered with 200")
	}
}

// Redirects that stay on https keep working (CDNs use them).
func TestUpdateClientFollowsHTTPSRedirect(t *testing.T) {
	target := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer target.Close()
	secure := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL+"/pkg", http.StatusFound)
	}))
	defer secure.Close()

	client := NewUpdateHTTPClient(5 * time.Second)
	client.Transport = secure.Client().Transport // same test CA for both
	response, err := client.Get(secure.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", response.StatusCode)
	}
}
