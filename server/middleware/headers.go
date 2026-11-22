package middleware

import "github.com/gin-gonic/gin"

// SecurityHeaders sets defensive response headers on every response.
//
// The policy is deliberately narrow: it forbids
// framing (clickjacking of the remote console), MIME sniffing and referrer
// leakage. Content loading itself (scripts, media, WebRTC) is not
// restricted here.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Frame-Options", "DENY")
		h.Set("Content-Security-Policy", "frame-ancestors 'none'; base-uri 'self'; form-action 'self'; object-src 'none'")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "no-referrer")
		c.Next()
	}
}
