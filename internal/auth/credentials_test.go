package auth

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func testCreds() *Credentials {
	return &Credentials{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		Expiry:       time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
		DeviceName:   "my-laptop",
		DeviceID:     "device-abc-123",
		UplinkURL:    "https://uplink.example.com",
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	home := t.TempDir()
	creds := testCreds()

	if err := SaveCredentialsTo(home, creds); err != nil {
		t.Fatalf("SaveCredentialsTo failed: %v", err)
	}

	loaded, err := LoadCredentialsFrom(home)
	if err != nil {
		t.Fatalf("LoadCredentialsFrom failed: %v", err)
	}

	if loaded.AccessToken != creds.AccessToken {
		t.Errorf("AccessToken = %q, want %q", loaded.AccessToken, creds.AccessToken)
	}
	if loaded.RefreshToken != creds.RefreshToken {
		t.Errorf("RefreshToken = %q, want %q", loaded.RefreshToken, creds.RefreshToken)
	}
	if !loaded.Expiry.Equal(creds.Expiry) {
		t.Errorf("Expiry = %v, want %v", loaded.Expiry, creds.Expiry)
	}
	if loaded.DeviceName != creds.DeviceName {
		t.Errorf("DeviceName = %q, want %q", loaded.DeviceName, creds.DeviceName)
	}
	if loaded.DeviceID != creds.DeviceID {
		t.Errorf("DeviceID = %q, want %q", loaded.DeviceID, creds.DeviceID)
	}
	if loaded.UplinkURL != creds.UplinkURL {
		t.Errorf("UplinkURL = %q, want %q", loaded.UplinkURL, creds.UplinkURL)
	}
}

func TestFilePermissions(t *testing.T) {
	home := t.TempDir()
	creds := testCreds()

	if err := SaveCredentialsTo(home, creds); err != nil {
		t.Fatalf("SaveCredentialsTo failed: %v", err)
	}

	path := filepath.Join(home, chiefDir, credentialsFile)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0o600 {
		t.Errorf("file permissions = %o, want 600", perm)
	}
}

func TestLoadMissingFile(t *testing.T) {
	home := t.TempDir()

	_, err := LoadCredentialsFrom(home)
	if err == nil {
		t.Fatal("expected error for missing credentials file, got nil")
	}
}

func TestIsExpired(t *testing.T) {
	// Expired token
	expired := &Credentials{Expiry: time.Now().Add(-1 * time.Hour)}
	if !expired.IsExpired() {
		t.Error("expected expired token to return true")
	}

	// Valid token
	valid := &Credentials{Expiry: time.Now().Add(1 * time.Hour)}
	if valid.IsExpired() {
		t.Error("expected valid token to return false")
	}
}

func TestNeedsRefresh(t *testing.T) {
	// Expiring in 3 minutes — within the 5-minute buffer
	soon := &Credentials{Expiry: time.Now().Add(3 * time.Minute)}
	if !soon.NeedsRefresh() {
		t.Error("expected token expiring in 3m to need refresh")
	}

	// Expiring in 10 minutes — outside the buffer
	later := &Credentials{Expiry: time.Now().Add(10 * time.Minute)}
	if later.NeedsRefresh() {
		t.Error("expected token expiring in 10m to not need refresh")
	}

	// Already expired
	expired := &Credentials{Expiry: time.Now().Add(-1 * time.Hour)}
	if !expired.NeedsRefresh() {
		t.Error("expected expired token to need refresh")
	}
}

func TestDeleteCredentials(t *testing.T) {
	home := t.TempDir()
	creds := testCreds()

	if err := SaveCredentialsTo(home, creds); err != nil {
		t.Fatalf("SaveCredentialsTo failed: %v", err)
	}

	if err := DeleteCredentialsFrom(home); err != nil {
		t.Fatalf("DeleteCredentialsFrom failed: %v", err)
	}

	path := filepath.Join(home, chiefDir, credentialsFile)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("expected credentials file to be deleted")
	}
}

func TestDeleteNonExistent(t *testing.T) {
	home := t.TempDir()

	// Should not return an error when file doesn't exist
	if err := DeleteCredentialsFrom(home); err != nil {
		t.Fatalf("DeleteCredentialsFrom on non-existent file returned error: %v", err)
	}
}
