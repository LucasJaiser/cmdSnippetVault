package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/charmbracelet/huh"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/spf13/cobra"
)

type DummySnippet struct {
	Command     string
	Description string
	Tags        []string
}

var EditCommand = &cobra.Command{
	Use:   "edit <id>",
	Short: "Edit Snippet",
	Long:  "",
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

		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("id is not a number: %w", err)
		}

		snippet, err := snippetService.Get(cmd.Context(), id)
		if err != nil {
			return err
		}

		//Step 1: Copy snippet to tmp file
		file, err := os.CreateTemp("", "editTemp")
		if err != nil {
			return fmt.Errorf("could not create tmp file: %w", err)
		}

		//Creating Dummy json so the user doesent see id, use count, created_at, updated_at
		snippetJson, err := json.MarshalIndent(DummySnippet{
			Command:     snippet.Command,
			Description: snippet.Description,
			Tags:        snippet.Tags,
		}, "", "    ")

		if err != nil {
			return fmt.Errorf("could not convert snippet to json")
		}

		file.Write(snippetJson)

		filename := file.Name()

		defer os.Remove(filename)

		err = file.Close()
		if err != nil {
			return fmt.Errorf("could not close tmp file: %w", err)
		}

		//Step 2: Start Editor with tmp file
		comm := exec.Command(appCfg.Editor, filename)
		comm.Stdin = os.Stdin
		comm.Stdout = os.Stdout
		comm.Stderr = os.Stderr
		err = comm.Run()

		if err != nil {
			return fmt.Errorf("could not run editor: %w", err)
		}
		//get tmp file and diff with original
		readSnippetJson, err := os.ReadFile(filename)

		if string(readSnippetJson) == string(snippetJson) {
			fmt.Println("No Change")
			return nil
		}

		diffMatcher := diffmatchpatch.New()

		diffs := diffMatcher.DiffMain(string(snippetJson), string(readSnippetJson), true)
		diffs = diffMatcher.DiffCleanupSemantic(diffs)

		fmt.Println(diffMatcher.DiffPrettyText(diffs))

		//Step 3: Confirm Changes
		var confirm bool
		err = huh.NewConfirm().
			Title(fmt.Sprintf("Update Snippet #%d?", id)).
			Description("This action cannot be undone.").
			Affirmative("Update").
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

		//Step 4: Update snippet on yes
		var readJsonSnippet DummySnippet
		err = json.Unmarshal(readSnippetJson, &readJsonSnippet)
		if err != nil {
			return fmt.Errorf("Malformed file: %w", err)
		}

		snippet.Command = readJsonSnippet.Command
		snippet.Description = readJsonSnippet.Description
		snippet.Tags = readJsonSnippet.Tags

		err = snippetService.Update(cmd.Context(), *snippet)
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
