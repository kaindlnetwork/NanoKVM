package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// The UI and the API must not be embeddable in a foreign page (clickjacking
// of the remote console), must not be MIME-sniffed and must not leak the
// URL as referrer.
func TestResponsesCarrySecurityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	web(r)
	r.GET("/api/probe", func(c *gin.Context) { c.Status(http.StatusOK) })

	for _, path := range []string{"/", "/api/probe"} {
		recorder := httptest.NewRecorder()
		r.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		for header, want := range map[string]string{
			"X-Frame-Options":        "DENY",
			"X-Content-Type-Options": "nosniff",
			"Referrer-Policy":        "no-referrer",
		} {
			if got := recorder.Header().Get(header); got != want {
				t.Errorf("%s: %s = %q, want %q", path, header, got, want)
			}
		}
		if csp := recorder.Header().Get("Content-Security-Policy"); csp == "" {
			t.Errorf("%s: Content-Security-Policy missing", path)
		}
	}
}
