package cli

import (
	"fmt"
	"os"

	"lucasjaiser/goSnipperVault/internal/config"
	"lucasjaiser/goSnipperVault/internal/service"

	"github.com/spf13/cobra"
)

var (
	cfgFile        string
	noColor        bool
	appCfg         *config.Config
	snippetService *service.SnippetService
	Cleanup        func()

	rootCmd = &cobra.Command{
		Use:     "cmdSnipperVault",
		Short:   "A CLI snippet manager for developers",
		Long:    "",
		Version: "1.0",
	}
)

func Init() {
	cobra.OnInitialize(InitEssential)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path which will be used")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable color output")

	AddCommand.Flags().StringP("command", "c", "", "command to save")
	AddCommand.Flags().StringP("description", "d", "", "description of the command")
	AddCommand.Flags().StringSliceP("tags", "t", nil, "tags of the snippet, comma separated")
	rootCmd.AddCommand(AddCommand)

	rootCmd.AddCommand(EditCommand)
	rootCmd.AddCommand(GetCommand)
	rootCmd.AddCommand(ExecCommand)
	rootCmd.AddCommand(ListCommand)
	rootCmd.AddCommand(SearchCommand)
	rootCmd.AddCommand(DeleteCommand)

}

func Execute() error {
	return rootCmd.Execute()
}

func InitEssential() {
	cfg, err := config.InitConfig(rootCmd.Flags(), cfgFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}
	appCfg = cfg
}
