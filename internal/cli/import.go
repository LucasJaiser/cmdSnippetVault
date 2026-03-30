package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var ImportCommand = &cobra.Command{
	Use:   "import <import_file>",
	Short: "Import snippet from a file",
	Long:  "",
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
