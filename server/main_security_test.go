package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"NanoKVM-Server/config"

	"github.com/gin-gonic/gin"
)

// With authentication disabled, upstream answered every origin with
// "Access-Control-Allow-Origin: *": any web page a LAN user opened could
// then drive the KVM (keyboard, power, terminal) through their browser.
func TestNoWildcardCORSWhenAuthDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := newEngine(&config.Config{Authentication: "disable"})
	r.POST("/api/vm/gpio", func(c *gin.Context) { c.Status(http.StatusOK) })

	request := httptest.NewRequest(http.MethodOptions, "/api/vm/gpio", nil)
	request.Header.Set("Origin", "https://evil.example")
	request.Header.Set("Access-Control-Request-Method", "POST")
	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, request)

	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("cross-origin access allowed: %q", got)
	}
}

// A same-origin request from the device's own UI keeps working.
func TestSameOriginRequestUnaffected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := newEngine(&config.Config{Authentication: "disable"})
	r.POST("/api/vm/gpio", func(c *gin.Context) { c.Status(http.StatusOK) })
	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/vm/gpio", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}
}

// Removing CORS only hides responses. A plain HTML form on a foreign page
// can still POST application/x-www-form-urlencoded without a preflight, and
// with authentication disabled the action (power, reset) would run.
func TestCrossSiteFormPostIsRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := newEngine(&config.Config{Authentication: "disable"})
	ran := false
	r.POST("/api/vm/gpio", func(c *gin.Context) { ran = true; c.Status(http.StatusOK) })

	for name, set := range map[string]func(*http.Request){
		"Sec-Fetch-Site": func(r *http.Request) { r.Header.Set("Sec-Fetch-Site", "cross-site") },
		"Origin only":    func(r *http.Request) { r.Header.Set("Origin", "https://evil.example") },
	} {
		ran = false
		request := httptest.NewRequest(http.MethodPost, "http://kvm.local/api/vm/gpio",
			strings.NewReader("type=power&duration=800"))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		set(request)
		recorder := httptest.NewRecorder()
		r.ServeHTTP(recorder, request)
		if ran || recorder.Code != http.StatusForbidden {
			t.Errorf("%s: handler ran=%v, status=%d, want 403", name, ran, recorder.Code)
		}
	}
}

// The device's own UI sends Sec-Fetch-Site: same-origin and must pass.
func TestSameOriginBrowserPostAllowed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := newEngine(&config.Config{Authentication: "enable"})
	r.POST("/api/vm/gpio", func(c *gin.Context) { c.Status(http.StatusOK) })
	request := httptest.NewRequest(http.MethodPost, "http://kvm.local/api/vm/gpio", nil)
	request.Header.Set("Sec-Fetch-Site", "same-origin")
	request.Header.Set("Origin", "http://kvm.local")
	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}
}
