package commands

import (
	"strings"

	"github.com/minicodemonkey/chief/internal/cmd"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new [name] [context...]",
	Short: "Create a new PRD interactively",
	Args:  cobra.ArbitraryArgs,
	RunE: func(c *cobra.Command, args []string) error {
		opts := cmd.NewOptions{}
		if len(args) > 0 {
			opts.Name = args[0]
		}
		if len(args) > 1 {
			opts.Context = strings.Join(args[1:], " ")
		}

		provider, err := resolveProvider()
		if err != nil {
			return err
		}
		opts.Provider = provider

		return cmd.RunNew(opts)
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
}
