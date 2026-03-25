package cli

import (
	"fmt"
	"lucasjaiser/goSnipperVault/internal/domain"
	"slices"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var AddCommand = &cobra.Command{
	Use:   "add",
	Short: "Add a new Snippet",
	Long:  "",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := getService()
		if err != nil {
			return err
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {

		command, _ := cmd.Flags().GetString("command")
		description, _ := cmd.Flags().GetString("description")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		if command == "" && description == "" && len(tags) == 0 {
			err := AddCommand_Interactive(&command, &description, &tags)

			if err != nil {
				return err
			}
		}

		err := snippetService.Create(cmd.Context(), domain.Snippet{
			Command:     command,
			Description: description,
			Tags:        tags,
		})

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

func AddCommand_Interactive(command *string, description *string, tags *[]string) error {
	var tagsInput string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Command").
				Prompt("> ").
				Value(command),

			huh.NewInput().
				Title("Description").
				Prompt("> ").
				Placeholder("optional").
				Value(description),

			huh.NewInput().
				Title("Tags").
				Prompt("> ").
				Placeholder("comma separated").
				Value(&tagsInput),
		),
	)

	err := form.Run()
	if err != nil {
		return fmt.Errorf("Cancelled")
	}

	tagsInput = strings.TrimSpace(tagsInput)
	if tagsInput != "" {

		for t := range strings.SplitSeq(tagsInput, ",") {
			t = strings.TrimSpace(t)
			t = strings.ToLower(t)

			if t != "" && !slices.Contains(*tags, t) {
				*tags = append(*tags, t)
			}
		}
	}

	return nil
}
