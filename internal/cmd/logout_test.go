package cmd

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/minicodemonkey/chief/internal/auth"
)

func TestRunLogoutSuccess(t *testing.T) {
	revokeCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete && r.URL.Path == "/api/auth/device/dev-123" {
			revokeCalled = true
			if r.Header.Get("Authorization") != "Bearer test-access-token" {
				t.Errorf("expected Bearer token, got %q", r.Header.Get("Authorization"))
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	tmpHome := t.TempDir()
	creds := &auth.Credentials{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		Expiry:       time.Now().Add(time.Hour),
		DeviceName:   "test-machine",
		DeviceID:     "dev-123",
		UplinkURL:    server.URL,
	}
	if err := auth.SaveCredentialsTo(tmpHome, creds); err != nil {
		t.Fatalf("failed to save test credentials: %v", err)
	}

	err := RunLogoutFrom(tmpHome)
	if err != nil {
		t.Fatalf("RunLogoutFrom returned error: %v", err)
	}

	if !revokeCalled {
		t.Error("expected revoke endpoint to be called")
	}

	// Verify credentials were deleted
	_, err = auth.LoadCredentialsFrom(tmpHome)
	if err == nil {
		t.Error("expected credentials to be deleted")
	}
}

func TestRunLogoutServerRevocationFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	tmpHome := t.TempDir()
	creds := &auth.Credentials{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		Expiry:       time.Now().Add(time.Hour),
		DeviceName:   "test-machine",
		DeviceID:     "dev-456",
		UplinkURL:    server.URL,
	}
	if err := auth.SaveCredentialsTo(tmpHome, creds); err != nil {
		t.Fatalf("failed to save test credentials: %v", err)
	}

	// Should succeed even though server revocation fails
	err := RunLogoutFrom(tmpHome)
	if err != nil {
		t.Fatalf("RunLogoutFrom should not return error when server revocation fails: %v", err)
	}

	// Verify credentials were still deleted
	_, err = auth.LoadCredentialsFrom(tmpHome)
	if err == nil {
		t.Error("expected credentials to be deleted even when server revocation fails")
	}
}

func TestRunLogoutNotLoggedIn(t *testing.T) {
	tmpHome := t.TempDir()

	err := RunLogoutFrom(tmpHome)
	if err == nil {
		t.Fatal("expected error when not logged in")
	}
	if err.Error() != "not logged in: no credentials found" {
		t.Errorf("expected 'not logged in: no credentials found', got %q", err.Error())
	}
}

func TestRunLogoutUsesDefaultHome(t *testing.T) {
	// Ensure RunLogout() calls RunLogoutFrom("") which uses real home dir
	// We test this by setting HOME to a temp dir with no credentials
	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	err := RunLogout()
	if err == nil {
		t.Fatal("expected error when not logged in")
	}
}
