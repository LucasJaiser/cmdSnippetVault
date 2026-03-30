package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// CompletionCommand generates shell completion scripts.
var CompletionCommand = &cobra.Command{
	Use:   "completion <shell>",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for bash, zsh, fish, or powershell.

To load completions:

Bash:
  $ source <(cmdSnipperVault completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ cmdSnipperVault completion bash > /etc/bash_completion.d/cmdSnipperVault
  # macOS:
  $ cmdSnipperVault completion bash > $(brew --prefix)/etc/bash_completion.d/cmdSnipperVault

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc
  # To load completions for each session, execute once:
  $ cmdSnipperVault completion zsh > "${fpath[1]}/_cmdSnipperVault"
  # You will need to start a new shell for this setup to take effect.

Fish:
  $ cmdSnipperVault completion fish | source
  # To load completions for each session, execute once:
  $ cmdSnipperVault completion fish > ~/.config/fish/completions/cmdSnipperVault.fish

PowerShell:
  PS> cmdSnipperVault completion powershell | Out-String | Invoke-Expression
  # To load completions for every new session, run:
  PS> cmdSnipperVault completion powershell > cmdSnipperVault.ps1
  # and source this file from your PowerShell profile.
`,
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletionV2(os.Stdout, true)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		default:
			return fmt.Errorf("unsupported shell: %s (valid: bash, zsh, fish, powershell)", args[0])
		}
	},
}
