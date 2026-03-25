package cli

import (
	"encoding/json"
	"fmt"
	"lucasjaiser/goSnipperVault/internal/domain"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var SearchCommand = &cobra.Command{
	Use:   "search <query>",
	Short: "search for snippets in your database",
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
			return fmt.Errorf("missing query argument")
		}

		limit, _ := cmd.Flags().GetInt("limit")
		json, _ := cmd.Flags().GetBool("json")
		pretty, _ := cmd.Flags().GetBool("pretty")

		snippets, err := snippetService.Search(cmd.Context(), args[0])

		if err != nil {
			return err
		}

		err = SearchCommand_PrintSnippets(snippets, limit, json, pretty)
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

func SearchCommand_PrintSnippets(snippets []*domain.Snippet, limit int, jsonFlag bool, pretty bool) error {
	if len(snippets) == 0 {
		fmt.Println("No snippets found.")
		return nil
	}

	if jsonFlag {
		var bytes []byte
		var err error

		if pretty {

			bytes, err = json.MarshalIndent(snippets, "", "    ")
		} else {

			bytes, err = json.Marshal(snippets)
		}
		if err != nil {
			return fmt.Errorf("could not convert snippets to json: %w", err)
		}

		fmt.Println(string(bytes))
		return nil
	}

	id := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6")).Width(6)
	cmd := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2"))
	desc := lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	tag := lipgloss.NewStyle().Background(lipgloss.Color("8")).Foreground(lipgloss.Color("15")).Padding(0, 1)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	divider := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(strings.Repeat("─", 50))

	fmt.Printf("\n%s\n\n", lipgloss.NewStyle().Bold(true).Render(
		fmt.Sprintf("Found %d snippet(s)", len(snippets)),
	))

	for i, s := range snippets {
		if i == limit {
			break
		}
		fmt.Printf("%s %s\n", id.Render("#"+strconv.FormatInt(s.ID, 10)), cmd.Render(s.Command))

		fmt.Println()
		if s.Description != "" {
			fmt.Printf("       %s\n", desc.Render(s.Description))
		}

		if len(s.Tags) > 0 {
			var tags []string
			for _, t := range s.Tags {
				tags = append(tags, tag.Render(t))
			}
			fmt.Printf("       %s\n", strings.Join(tags, " "))
		}

		fmt.Printf("       %s\n", dim.Render(fmt.Sprintf("used %d time(s)", s.UseCount)))

		if i < len(snippets)-1 {
			fmt.Printf("  %s\n", divider)
		}
	}
	fmt.Println()

	return nil
}
