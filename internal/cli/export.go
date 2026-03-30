package cli

import (
	"fmt"
	"lucasjaiser/goSnipperVault/internal/domain"

	"github.com/spf13/cobra"
)

var ExportCommand = &cobra.Command{
	Use:   "export",
	Short: "Export your snippets",
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
