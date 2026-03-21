package commands

import (
	"github.com/minicodemonkey/chief/internal/cmd"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit [name]",
	Short: "Edit an existing PRD interactively",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		opts := cmd.EditOptions{}
		if len(args) > 0 {
			opts.Name = args[0]
		}

		provider, err := resolveProvider()
		if err != nil {
			return err
		}
		opts.Provider = provider

		return cmd.RunEdit(opts)
	},
}

func init() {
	// Register --merge and --force on edit too (advertised in current help).
	// These use local variables since EditOptions doesn't consume them yet —
	// they exist purely for backwards compatibility with scripts that pass them.
	var editMerge, editForce bool
	editCmd.Flags().BoolVar(&editMerge, "merge", false, "Auto-merge progress on conversion conflicts")
	editCmd.Flags().BoolVar(&editForce, "force", false, "Auto-overwrite on conversion conflicts")
	rootCmd.AddCommand(editCmd)
}
