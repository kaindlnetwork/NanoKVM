package vm

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func scriptRouter(t *testing.T) (*gin.Engine, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	original := ScriptDirectory
	ScriptDirectory = dir
	t.Cleanup(func() { ScriptDirectory = original })

	service := NewService()
	router := gin.New()
	router.POST("/run", service.RunScript)
	router.DELETE("/script", service.DeleteScript)
	return router, dir
}

func postJSON(router http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	data, _ := json.Marshal(body)
	request := httptest.NewRequest(method, path, bytes.NewReader(data))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

// The script name used to be pasted into "sh -c", so a name with shell
// syntax ran arbitrary commands.
func TestRunScriptDoesNotPassNameToShell(t *testing.T) {
	router, dir := scriptRouter(t)
	script := filepath.Join(dir, "ok.sh")
	if err := os.WriteFile(script, []byte("#!/bin/sh\necho ran\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(t.TempDir(), "pwned")

	postJSON(router, http.MethodPost, "/run", map[string]string{
		"name": "ok.sh; touch " + marker + " #.sh",
		"type": "foreground",
	})

	if _, err := os.Stat(marker); err == nil {
		t.Fatal("injected command was executed")
	}
}

func TestRunScriptRejectsPathTraversal(t *testing.T) {
	router, dir := scriptRouter(t)
	outside := filepath.Join(filepath.Dir(dir), "outside-"+filepath.Base(dir)+".sh")
	marker := outside + ".ran"
	if err := os.WriteFile(outside, []byte("#!/bin/sh\ntouch "+marker+"\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Remove(outside); _ = os.Remove(marker) })

	postJSON(router, http.MethodPost, "/run", map[string]string{
		"name": "../" + filepath.Base(outside),
		"type": "foreground",
	})

	if _, err := os.Stat(marker); err == nil {
		t.Fatal("script outside the script directory was executed")
	}
}

func TestRunScriptStillRunsAValidScript(t *testing.T) {
	router, dir := scriptRouter(t)
	if err := os.WriteFile(filepath.Join(dir, "hello.sh"), []byte("echo hello-from-script\n"), 0o700); err != nil {
		t.Fatal(err)
	}

	recorder := postJSON(router, http.MethodPost, "/run", map[string]string{
		"name": "hello.sh",
		"type": "foreground",
	})

	if !bytes.Contains(recorder.Body.Bytes(), []byte("hello-from-script")) {
		t.Fatalf("script output missing: %s", recorder.Body.String())
	}
}

func TestDeleteScriptRejectsPathTraversal(t *testing.T) {
	router, dir := scriptRouter(t)
	victim := filepath.Join(filepath.Dir(dir), "victim-"+filepath.Base(dir))
	if err := os.WriteFile(victim, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Remove(victim) })

	postJSON(router, http.MethodDelete, "/script", map[string]string{
		"name": "../" + filepath.Base(victim),
	})

	if _, err := os.Stat(victim); err != nil {
		t.Fatal("file outside the script directory was deleted")
	}
}
