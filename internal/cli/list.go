package cli

import (
	"lucasjaiser/goSnipperVault/internal/domain"

	"github.com/spf13/cobra"
)

var ListCommand = &cobra.Command{
	Use:   "list",
	Short: "List available snippets",
	Long:  "",
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

		if err != SearchCommand_PrintSnippets(snippets, limit, json, pretty) {
			return err
		}
		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		Cleanup()

		return nil
	},
}
