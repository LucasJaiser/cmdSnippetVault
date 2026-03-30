package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// SearchCommand finds snippets matching a query across commands, descriptions, and tags.
var SearchCommand = &cobra.Command{
	Use:   "search <query>",
	Short: "search for snippets in your database",
	Long: `Search for snippets by matching against command, description, and tag names.
The search is case-insensitive and results are ranked by relevance, with exact
command matches ranked highest.

Results can be output as JSON for scripting and limited to a maximum count.

Examples:
  cmdSnipperVault search docker
  cmdSnipperVault search "git log" --json --pretty
  cmdSnipperVault search deploy -l 5`,
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
			return fmt.Errorf("missing query argument")
		}

		limit, _ := cmd.Flags().GetInt("limit")
		json, _ := cmd.Flags().GetBool("json")
		pretty, _ := cmd.Flags().GetBool("pretty")

		snippets, err := snippetService.Search(cmd.Context(), args[0])

		if err != nil {
			return err
		}

		err = PrintSnippetList(snippets, limit, json, pretty)
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
