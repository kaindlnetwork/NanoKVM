package auth

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/config"

	"github.com/gin-gonic/gin"
)

// Parallel wrong guesses from one client must not slip past the lockout.
func TestParallelLoginGuessesAreLockedOut(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := authn.NewStore(filepath.Join(t.TempDir(), "pwd"))
	restore := useTestStore(store)
	defer restore()

	conf := config.GetInstance()
	origDuration, origFailures := conf.Security.LoginLockoutDuration, conf.Security.LoginMaxFailures
	conf.Security.LoginLockoutDuration = 60
	conf.Security.LoginMaxFailures = 2
	defer func() {
		conf.Security.LoginLockoutDuration, conf.Security.LoginMaxFailures = origDuration, origFailures
		loginMutex.Lock()
		delete(loginAttempts, "198.51.100.7")
		loginMutex.Unlock()
	}()

	service := NewService()
	router := gin.New()
	router.POST("/login", service.Login)

	const attempts = 6
	var wg sync.WaitGroup
	var mu sync.Mutex
	checked := 0
	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			recorder := httptest.NewRecorder()
			request := jsonRequest(http.MethodPost, "/login", map[string]any{
				"username": "admin",
				"password": encryptForRequest("wrong-guess"),
			})
			request.RemoteAddr = "198.51.100.7:4000"
			router.ServeHTTP(recorder, request)
			// -2 = "invalid username or password": the guess was checked.
			if strings.Contains(recorder.Body.String(), `"code":-2`) {
				mu.Lock()
				checked++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	// Only LoginMaxFailures guesses may reach the password check; the rest
	// must be answered by the lockout.
	if checked > conf.Security.LoginMaxFailures {
		t.Fatalf("%d of %d parallel guesses were checked, want <= %d", checked, attempts, conf.Security.LoginMaxFailures)
	}
}

// A single host usually controls a whole IPv6 /64 and could rotate its
// address after every lockout. The lockout key therefore covers the /64.
func TestLockoutKeyGroupsIPv6By64(t *testing.T) {
	key := func(remote string) string {
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = httptest.NewRequest(http.MethodPost, "/login", nil)
		c.Request.RemoteAddr = remote
		return GetClientIP(c)
	}
	a := key("[2001:db8:1:2::1]:4000")
	b := key("[2001:db8:1:2:ffff::9]:4000")
	if a != b {
		t.Errorf("same /64 got different keys: %q %q", a, b)
	}
	if key("[2001:db8:1:3::1]:4000") == a {
		t.Error("different /64 share a key")
	}
	if got := key("192.0.2.1:4000"); got != "192.0.2.1" {
		t.Errorf("IPv4 key = %q", got)
	}
}
