package cli

import (
	"errors"
	"fmt"
	"lucasjaiser/goSnippetVault/internal/domain"
	"strconv"

	"github.com/spf13/cobra"
)

// GetCommand retrieves a snippet by ID and optionally copies it to the clipboard.
var GetCommand = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a Snippet and or Copy it to the Clipboard",
	Long: `Retrieve a snippet by its ID, display its details, and copy the command
to the clipboard. The snippet's use count is incremented on each retrieval.

Use --no-copy to display the snippet without copying to the clipboard.

Examples:
  cmdSnippetVault get 42
  cmdSnippetVault get 42 --no-copy`,
	Args: cobra.ExactArgs(1),
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := getService()
		if err != nil {
			return err
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("missing id argument")
		}

		noCopy, _ := cmd.Flags().GetBool("no-copy")
		var snippet *domain.Snippet

		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("id is not a number: %w", err)
		}

		if !noCopy {
			err = snippetService.GetAndCopy(cmd.Context(), id)
		} else {
			snippet, err = snippetService.Get(cmd.Context(), id)
		}

		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				fmt.Printf("no Snippets found with id %d", id)
				return nil
			}

			return err
		}

		if snippet != nil {
			PrintSnippetDetail(snippet)
		}

		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		Cleanup()

		return nil
	},
}
