package cli

import (
	"fmt"
	"lucasjaiser/goSnippetVault/internal/domain"

	"github.com/spf13/cobra"
)

// ExportCommand exports snippets to a JSON or YAML file.
var ExportCommand = &cobra.Command{
	Use:   "export",
	Short: "Export your snippets",
	Long: `Export snippets from your collection to a JSON or YAML file. You can filter
by tag to export only specific snippets, and specify the output file path.

The format defaults to JSON and can be changed with the --format flag.

Examples:
  cmdSnippetVault export -o backup.json
  cmdSnippetVault export -o docker.yaml -f yaml -t docker`,
	Args: cobra.ExactArgs(1),
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := getService()
		if err != nil {
			return err
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {

		format, _ := cmd.Flags().GetString("format")
		tag, _ := cmd.Flags().GetString("tag")
		output, _ := cmd.Flags().GetString("output")

		snippets, err := snippetService.List(cmd.Context(), domain.ListFilter{
			Tag:    tag,
			Limit:  100, //TODO: Configurable?
			Offset: 0,
		})

		if err != nil {
			return err
		}

		exporter := getExporterForType(format)

		err = exporter.Write(snippets, output)
		if err != nil {
			return fmt.Errorf("could not export snippets: %w", err)
		}

		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		Cleanup()

		return nil
	},
}
