package main

import (
	"net/http"
	"net/http/httptest"
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
