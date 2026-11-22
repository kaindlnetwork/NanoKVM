package auth

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/config"

	"github.com/gin-gonic/gin"
)

// Any client could send "X-Forwarded-Proto: https" and get a session cookie
// marked Secure (and, with a real proxy missing, be told it is on HTTPS).
// The header only counts from a configured proxy.
func TestSpoofedForwardedProtoDoesNotChangeCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := authn.NewStore(filepath.Join(t.TempDir(), "pwd"))
	restore := useTestStore(store)
	defer restore()
	if _, err := store.SetPassword("admin", "admin-password-1"); err != nil {
		t.Fatal(err)
	}
	conf := config.GetInstance()
	origProto, origProxies := conf.Proto, conf.Security.TrustedProxies
	conf.Proto = "http"
	defer func() { conf.Proto, conf.Security.TrustedProxies = origProto, origProxies }()

	service := NewService()
	router := gin.New()
	router.POST("/login", service.Login)

	login := func(remote string) *http.Cookie {
		recorder := httptest.NewRecorder()
		request := jsonRequest(http.MethodPost, "/login", map[string]any{
			"username": "admin",
			"password": encryptForRequest("admin-password-1"),
		})
		request.RemoteAddr = remote
		request.Header.Set("X-Forwarded-Proto", "https")
		router.ServeHTTP(recorder, request)
		for _, cookie := range recorder.Result().Cookies() {
			if cookie.Name == "nano-kvm-token" {
				return cookie
			}
		}
		t.Fatalf("no session cookie: %s", recorder.Body.String())
		return nil
	}

	conf.Security.TrustedProxies = nil
	if login("203.0.113.5:1000").Secure {
		t.Fatal("untrusted X-Forwarded-Proto honoured")
	}
	conf.Security.TrustedProxies = []string{"203.0.113.5"}
	if !login("203.0.113.5:1001").Secure {
		t.Fatal("trusted proxy's X-Forwarded-Proto ignored")
	}
}
