package middleware

import (
	"net/http"
	"path/filepath"
	"testing"

	"NanoKVM-Server/authn"

	"github.com/gin-gonic/gin"
)

// With the factory password (admin/admin) a session may open the password
// page and nothing else. Upstream only showed a dismissible hint in the UI;
// the whole API, terminal included, was usable.
func TestFactoryPasswordOnlyReachesPasswordPage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := authn.NewStore(filepath.Join(t.TempDir(), "pwd"))
	admin, ok, err := store.Authenticate("admin", "admin")
	if err != nil || !ok || !admin.MustChangePassword {
		t.Fatalf("default login: ok=%v err=%v", ok, err)
	}
	restore := useTestAuthStore(t, store)
	defer restore()
	token, err := GenerateJWT(admin.Username, admin.TokenVersion)
	if err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	ok204 := func(c *gin.Context) { c.Status(http.StatusNoContent) }
	router.GET("/api/auth/password", CheckToken(), ok204)
	router.GET("/api/auth/account", CheckToken(), ok204)
	router.GET("/api/vm/terminal", CheckToken(), ok204)
	router.GET("/api/stream/mjpeg", CheckTokenOrLoopbackInternalToken(), ok204)

	for path, want := range map[string]int{
		"/api/auth/password": http.StatusNoContent,
		"/api/auth/account":  http.StatusNoContent,
		"/api/vm/terminal":   http.StatusPreconditionRequired,
		"/api/stream/mjpeg":  http.StatusPreconditionRequired,
	} {
		if status := requestWithToken(router, path, token); status != want {
			t.Errorf("%s status = %d, want %d", path, status, want)
		}
	}
}

// After the password was changed, the same account reaches everything again.
func TestChangedPasswordLiftsTheRestriction(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := authn.NewStore(filepath.Join(t.TempDir(), "pwd"))
	user, err := store.SetPassword("admin", "admin-password-1")
	if err != nil {
		t.Fatal(err)
	}
	restore := useTestAuthStore(t, store)
	defer restore()
	token, err := GenerateJWT(user.Username, user.TokenVersion)
	if err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.GET("/api/vm/terminal", CheckToken(), func(c *gin.Context) { c.Status(http.StatusNoContent) })
	if status := requestWithToken(router, "/api/vm/terminal", token); status != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", status)
	}
}
