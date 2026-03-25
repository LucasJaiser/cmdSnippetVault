package cli

import (
	"context"
	"errors"
	"fmt"
	"lucasjaiser/goSnipperVault/internal/domain"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var GetCommand = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a Snippet and or Copy it to the Clipboard",
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
		if len(args) != 1 {
			return fmt.Errorf("missing id argument")
		}

		noCopy, _ := cmd.Flags().GetBool("no-copy")
		var snippet *domain.Snippet

		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("id is not a number: %w", err)
		}

		if noCopy {
			err = snippetService.GetAndCopy(context.Background(), id)
		} else {
			snippet, err = snippetService.Get(context.Background(), id)
		}

		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				fmt.Printf("no Snippets found with id %d", id)
				return nil
			}

			return err
		}

		if snippet != nil {
			GetCommand_PrintSnippet(snippet)
		}

		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		Cleanup()

		return nil
	},
}

func GetCommand_PrintSnippet(snippet *domain.Snippet) {
	label := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	command := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2")).PaddingLeft(2)
	tag := lipgloss.NewStyle().Background(lipgloss.Color("8")).Foreground(lipgloss.Color("15")).Padding(0, 1)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	fmt.Println()
	fmt.Println(label.Render("Command"))
	fmt.Println(command.Render(snippet.Command))
	fmt.Println()

	if snippet.Description != "" {
		fmt.Println(label.Render("Description"))
		fmt.Printf("  %s\n\n", snippet.Description)
	}

	if len(snippet.Tags) > 0 {
		var tags []string
		for _, t := range snippet.Tags {
			tags = append(tags, tag.Render(t))
		}
		fmt.Printf("%s %s\n\n", label.Render("Tags"), strings.Join(tags, " "))
	}

	fmt.Printf("%s %s    %s %s\n",
		label.Render("ID"), dim.Render(strconv.FormatInt(snippet.ID, 10)),
		label.Render("Uses"), dim.Render(strconv.Itoa(snippet.UseCount)),
	)
	fmt.Println()
}
