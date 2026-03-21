package commands

import (
	"github.com/minicodemonkey/chief/internal/cmd"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all PRDs with progress",
	Args:  cobra.NoArgs,
	RunE: func(c *cobra.Command, args []string) error {
		return cmd.RunList(cmd.ListOptions{})
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
