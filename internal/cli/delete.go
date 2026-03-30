package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

// DeleteCommand removes a snippet by ID, with optional force flag.
var DeleteCommand = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a Snippet",
	Long: `Delete a snippet from your collection by its ID. You will be prompted
for confirmation unless the --force flag is provided.

Examples:
  cmdSnipperVault delete 42
  cmdSnipperVault delete 42 --force`,
	Args:  cobra.ExactArgs(1),
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := getService()
		if err != nil {
			return err
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("missing id argmument")
		}

		force, _ := cmd.Flags().GetBool("force")

		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("id is not a number: %w", err)
		}

		if !force {
			var confirm bool
			err := huh.NewConfirm().
				Title(fmt.Sprintf("Delete snippet #%d?", id)).
				Description("This action cannot be undone.").
				Affirmative("Delete").
				Negative("Cancel").
				Value(&confirm).
				Run()
			if err != nil {
				return fmt.Errorf("cancelled")
			}
			if !confirm {
				fmt.Println("Aborted.")
				return nil
			}
		}

		err = snippetService.Delete(context.Background(), id)

		if err != nil {
			return err
		}

		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		Cleanup()

		return nil
	},
}

func init() {

}
