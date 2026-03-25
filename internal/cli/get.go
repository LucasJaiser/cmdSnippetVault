package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var GetCommand = &cobra.Command{
	Use:   "get",
	Short: "Get a Snippet and or Copy it to the Clipboard",
	Long:  "",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := getService()
		if err != nil {
			return err
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Hello World!")
		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		Cleanup()

		return nil
	},
}
