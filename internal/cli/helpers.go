package cli

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"lucasjaiser/goSnippetVault/internal/clipboard"
	"lucasjaiser/goSnippetVault/internal/domain"
	"lucasjaiser/goSnippetVault/internal/exporter"
	"lucasjaiser/goSnippetVault/internal/importer"
	"lucasjaiser/goSnippetVault/internal/service"
	"lucasjaiser/goSnippetVault/internal/storage/sqlite"

	"github.com/charmbracelet/lipgloss"
)

func getService() error {
	if snippetService != nil {
		return nil
	}

	if appCfg == nil {
		InitEssential()
	}

	repo, err := sqlite.New(appCfg.DatabasePath)
	if err != nil {
		return err
	}

	Cleanup = func() { repo.Close() }

	if appCfg.Clipboard {
		snippetService = service.NewSnippetService(repo, clipboard.NewSystemClipboard())
	} else {
		snippetService = service.NewSnippetService(repo, clipboard.NewNoopClipboard())
	}

	return nil
}

func getImportForFileType(filename string, formatOverride string) domain.Importer {

	if formatOverride == "" {
		formatOverride = filepath.Ext(filename)
	}

	switch formatOverride {
	case "yaml", ".yaml":
		return importer.NewYAMLImporter()
	case "yml", ".yml":
		return importer.NewYAMLImporter()
	case "json", ".json":
		return importer.NewJSONImporter()
	default:
		return importer.NewJSONImporter()

	}

}

func getExporterForType(typeString string) domain.Exporter {

	switch typeString {
	case "yaml", ".yaml":
		return exporter.NewYAMLExporter()
	case "yml", ".yml":
		return exporter.NewYAMLExporter()
	case "json", ".json":
		return exporter.NewJSONExporter()
	default:
		return exporter.NewJSONExporter()
	}

}

// PrintSnippetDetail prints a single snippet in full detail view.
func PrintSnippetDetail(snippet *domain.Snippet) {
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

// PrintSnippetList prints a list of snippets in a compact table view or as JSON.
func PrintSnippetList(snippets []*domain.Snippet, limit int, jsonFlag bool, pretty bool) error {
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

// PrintTags prints a list of tags with their snippet counts.
func PrintTags(tags []*domain.TagWithCount) {
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
