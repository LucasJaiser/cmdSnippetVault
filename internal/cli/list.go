package cli

import (
	"lucasjaiser/goSnippetVault/internal/domain"

	"github.com/spf13/cobra"
)

// ListCommand displays snippets with optional tag filter and pagination.
var ListCommand = &cobra.Command{
	Use:   "list",
	Short: "List available snippets",
	Long: `List snippets from your collection, ordered by most recently updated.
Results can be filtered by tag, paginated with limit and offset, and
output as JSON for scripting.

Examples:
  cmdSnippetVault list
  cmdSnippetVault list -t docker -l 10
  cmdSnippetVault list --json --pretty`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := getService()
		if err != nil {
			return err
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {

		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")
		tag, _ := cmd.Flags().GetString("tag")
		json, _ := cmd.Flags().GetBool("json")
		pretty, _ := cmd.Flags().GetBool("pretty")

		snippets, err := snippetService.List(cmd.Context(), domain.ListFilter{
			Tag:    tag,
			Limit:  limit,
			Offset: offset,
		})

		if err != nil {
			return err
		}

		if err != PrintSnippetList(snippets, limit, json, pretty) {
			return err
		}
		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		Cleanup()

		return nil
	},
}
