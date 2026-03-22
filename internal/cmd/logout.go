package cmd

import (
	"fmt"

	"github.com/minicodemonkey/chief/internal/auth"
)

// RunLogout revokes the device on the server and deletes local credentials.
func RunLogout() error {
	return RunLogoutFrom("")
}

// RunLogoutFrom revokes the device and deletes credentials from the specified home directory.
func RunLogoutFrom(homeDir string) error {
	creds, err := auth.LoadCredentialsFrom(homeDir)
	if err != nil {
		return fmt.Errorf("not logged in: no credentials found")
	}

	// Try to revoke the device on the server
	flow := auth.NewDeviceFlow(creds.UplinkURL)
	if err := flow.RevokeDevice(creds.DeviceID, creds.AccessToken); err != nil {
		fmt.Printf("Warning: failed to revoke device on server: %v\n", err)
	}

	// Always delete local credentials
	if err := auth.DeleteCredentialsFrom(homeDir); err != nil {
		return fmt.Errorf("failed to delete credentials: %w", err)
	}

	fmt.Println("Logged out successfully.")
	return nil
}
