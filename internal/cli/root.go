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
		Short: "A CLI snippet manager for developers",
		Long: `cmdSnipperVault is a CLI snippet manager for saving, tagging, searching,
and executing shell commands. Store frequently used commands with descriptions
and tags, then quickly find and run them when needed.

Snippets are stored in a local SQLite database and can be imported or exported
as JSON or YAML for sharing and backup.`,
		Version: "dev",
	}
)

// SetVersion configures the version string displayed by the --version flag.
func SetVersion(version, commit, date string) {
	rootCmd.Version = fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date)
}

// Init registers all commands and flags on the root command.
func Init() {
	cobra.OnInitialize(InitEssential)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path which will be used")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable color output")

	AddCommand.Flags().StringP("command", "c", "", "command to save")
	AddCommand.Flags().StringP("description", "d", "", "description of the command")
	AddCommand.Flags().StringSliceP("tags", "t", nil, "tags of the snippet, comma separated")
	rootCmd.AddCommand(AddCommand)

	rootCmd.AddCommand(EditCommand)

	GetCommand.Flags().BoolP("no-copy", "n", false, "Dont copy command to Clipboard")
	rootCmd.AddCommand(GetCommand)

	rootCmd.AddCommand(ExecCommand)

	ListCommand.Flags().IntP("limit", "l", 20, "limits the max showed snippets")
	ListCommand.Flags().IntP("offset", "o", 0, "offsets start of limit count in the list")
	ListCommand.Flags().StringP("tag", "t", "", "filter for tag")
	ListCommand.Flags().BoolP("json", "j", false, "outputs snippets as JSON")
	ListCommand.Flags().BoolP("pretty", "p", false, "outputs JSON in a pretty format")

	ListCommand.AddCommand(ListTagsCommand)

	rootCmd.AddCommand(ListCommand)

	SearchCommand.Flags().BoolP("json", "j", false, "Output list of snippets as JSON")
	SearchCommand.Flags().BoolP("pretty", "p", false, "Outputs json in a pretty format")
	SearchCommand.Flags().IntP("limit", "l", 20, "Max number of shown Snippets")
	rootCmd.AddCommand(SearchCommand)

	DeleteCommand.Flags().BoolP("force", "f", false, "Delete without confirming")
	rootCmd.AddCommand(DeleteCommand)

	ImportCommand.Flags().StringP("format", "f", "json", "Format of the input file")
	ImportCommand.Flags().BoolP("dry-run", "d", false, "dont save just execute")
	rootCmd.AddCommand(ImportCommand)

	ExportCommand.Flags().StringP("format", "f", "", "Format of the output")
	ExportCommand.Flags().StringP("tag", "t", "", "Export specific snippet with tag")
	ExportCommand.Flags().StringP("output", "o", "", "define output file")
	rootCmd.AddCommand(ExportCommand)

	rootCmd.AddCommand(CompletionCommand)

}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

// InitEssential loads the application configuration.
func InitEssential() {
	cfg, err := config.InitConfig(rootCmd.Flags(), cfgFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}
	appCfg = cfg
}
