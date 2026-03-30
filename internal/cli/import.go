package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ImportCommand imports snippets from a JSON or YAML file.
var ImportCommand = &cobra.Command{
	Use:   "import <import_file>",
	Short: "Import snippet from a file",
	Long: `Import snippets from a JSON or YAML file into your collection. Duplicate
snippets (matching by command) are skipped, and invalid entries are rejected.

The file format is detected from the extension, or can be overridden with
the --format flag. Use --dry-run to preview the import without saving.

Examples:
  cmdSnippetVault import snippets.yaml
  cmdSnippetVault import snippets.txt --format json
  cmdSnippetVault import snippets.yaml --dry-run`,
	Args: cobra.ExactArgs(1),
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := getService()
		if err != nil {
			return err
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("missing impport filename argmument")
		}

		formatOverride, _ := cmd.Flags().GetString("format")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		importFile := args[0]

		importer := getImportForFileType(importFile, formatOverride)

		if importer == nil {
			return fmt.Errorf("wrong file type, only support yaml, yml, json")
		}

		snippets, err := importer.Read(importFile)
		if err != nil {
			return err
		}

		stats, err := snippetService.CreateBatch(cmd.Context(), snippets, dryRun)
		if err != nil {
			return err
		}

		fmt.Printf("import report: Imported %d, Skipped %d duplicates, %d invalid\n", stats.Created, stats.Duplicates, stats.Rejected)

		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		Cleanup()

		return nil
	},
}
