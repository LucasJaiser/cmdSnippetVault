package cli

import (
	"github.com/spf13/cobra"
)

// ListTagsCommand displays all tags with their snippet counts.
var ListTagsCommand = &cobra.Command{
	Use:   "tags",
	Short: "Lists available tags",
	Long: `List all tags in your snippet collection along with the number of snippets
associated with each tag. Tags are displayed in alphabetical order.

Examples:
  cmdSnippetVault list tags`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := getService()
		if err != nil {
			return err
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {

		tags, err := snippetService.ListTags(cmd.Context())
		if err != nil {
			return err
		}

		PrintTags(tags)

		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		Cleanup()

		return nil
	},
}
