package cli

import (
	"fmt"
	"lucasjaiser/goSnipperVault/internal/domain"
	"strconv"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var ListTagsCommand = &cobra.Command{
	Use:   "tags",
	Short: "Lists available tags",
	Long:  "",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := getService()
		if err != nil {
			return err
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {

		tags, err := snippetService.ListTags(cmd.Context())
		if err != nil {
			return err
		}

		ListTagsCommand_PrintTags(tags)

		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		Cleanup()

		return nil
	},
}

func ListTagsCommand_PrintTags(tags []*domain.TagWithCount) {
	if len(tags) == 0 {
		fmt.Println("No tags found.")
		return
	}

	name := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	count := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	fmt.Printf("\n%s\n\n", lipgloss.NewStyle().Bold(true).Render(
		fmt.Sprintf("%d tag(s)", len(tags)),
	))

	for _, t := range tags {
		fmt.Printf("  %s  %s\n",
			name.Render(t.Name),
			count.Render(strconv.Itoa(t.Count)+" snippet(s)"),
		)
	}
	fmt.Println()
}
