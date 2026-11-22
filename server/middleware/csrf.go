package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CrossOriginProtection rejects state-changing browser requests (anything
// but GET, HEAD, OPTIONS) that come from another origin, judged by
// Sec-Fetch-Site or, in older browsers, Origin against Host. It applies with
// and without authentication: SameSite=Strict protects the session cookie,
// but with authentication disabled there is no cookie, and a plain HTML form
// on any web page could otherwise press the power button.
//
// Requests without either header (curl, scripts) are not browser requests
// and pass, as do same-origin requests from the device's own UI.
func CrossOriginProtection() gin.HandlerFunc {
	protection := http.NewCrossOriginProtection()
	return func(c *gin.Context) {
		if err := protection.Check(c.Request); err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, "cross-origin request rejected")
			return
		}
		c.Next()
	}
}
