package commands

import (
	"github.com/minicodemonkey/chief/internal/cmd"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Disconnect this machine from the Uplink server",
	Long:  "Revoke device authorization on the server and delete local credentials.",
	Args:  cobra.NoArgs,
	RunE: func(c *cobra.Command, args []string) error {
		return cmd.RunLogout()
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
