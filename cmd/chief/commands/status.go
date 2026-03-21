package commands

import (
	"github.com/minicodemonkey/chief/internal/cmd"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [name]",
	Short: "Show progress for a PRD",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		opts := cmd.StatusOptions{}
		if len(args) > 0 {
			opts.Name = args[0]
		}
		return cmd.RunStatus(opts)
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
