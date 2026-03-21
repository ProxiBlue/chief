package commands

import (
	"github.com/minicodemonkey/chief/internal/cmd"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Chief to the latest version",
	Args:  cobra.NoArgs,
	RunE: func(c *cobra.Command, args []string) error {
		return cmd.RunUpdate(cmd.UpdateOptions{
			Version: Version,
		})
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
