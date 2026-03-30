package cli

import (
	"fmt"
	templatevar "lucasjaiser/goSnippetVault/pkg"
	"os"
	"os/exec"
	"strconv"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

// ExecCommand executes a snippet directly in the shell.
var ExecCommand = &cobra.Command{
	Use:   "exec <id>",
	Short: "Execute a Snippet directly",
	Long: `Execute a snippet directly in your shell by its ID. If the snippet contains
template variables (e.g. {{host}}), you will be prompted to fill them in
before execution.

You will be asked to confirm before the command runs unless confirm_execute
is disabled in your config.

Examples:
  cmdSnippetVault exec 42`,
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
			return fmt.Errorf("missing id argument")
		}

		//Step 1: Get Snippet
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("id is not a number: %w", err)
		}

		snippet, err := snippetService.Get(cmd.Context(), id)

		if err != nil {
			return fmt.Errorf("could not get snippet: %w", err)
		}

		//Step 2: Template if needed
		templ, values, err := templatevar.Parse(snippet.Command)
		if err != nil {
			return err
		}

		if len(values) > 0 {
			//Prompt for template values which we will use to template the command
			inputValues := make([]string, len(values))
			inputs := make([]huh.Field, len(values))
			for i, key := range values {
				inputs[i] = huh.NewInput().
					Title(key).
					Prompt("> ").
					Value(&inputValues[i])
			}

			err = huh.NewForm(huh.NewGroup(inputs...)).Run()
			if err != nil {
				return fmt.Errorf("cancelled")
			}

			templateValues := make(map[string]string, len(values))
			for i, key := range values {
				templateValues[key] = inputValues[i]
			}

			templatedCommand, err := templatevar.Resolve(templ, &templateValues)
			if err != nil {
				return err
			}

			snippet.Command = templatedCommand
		}

		//Step 3: Ask for confirmation
		PrintSnippetDetail(snippet)

		var confirm bool
		err = huh.NewConfirm().
			Title(fmt.Sprintf("Execute Snippet #%d?", id)).
			Description("This action cannot be undone.").
			Affirmative("Execute").
			Negative("Cancel").
			Value(&confirm).
			Run()
		if err != nil {
			return fmt.Errorf("cancelled")
		}
		if !confirm {
			fmt.Println("Aborted.")
			return nil
		}

		//Step 4: Execute Command
		comm := exec.Command("sh", "-c", snippet.Command)
		comm.Stdin = os.Stdin
		comm.Stdout = os.Stdout
		comm.Stderr = os.Stderr
		err = comm.Run()

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
