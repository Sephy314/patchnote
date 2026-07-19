package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/patchnote/patchnote/internal/ai"
	"github.com/patchnote/patchnote/internal/config"
	"github.com/patchnote/patchnote/internal/git"
	"github.com/patchnote/patchnote/internal/output"
	"github.com/patchnote/patchnote/internal/prompts"
	"github.com/patchnote/patchnote/internal/ui"
)

var prCmd = &cobra.Command{
	Use:   "pr",
	Short: "Generate a PR description",
	Long: `Analyse changes and generate a Markdown PR description
with summary, motivation, testing, and checklist.`,
	RunE: runPR,
}

func runPR(cmd *cobra.Command, _ []string) error {
	if !git.IsInstalled() {
		return fmt.Errorf("git is not installed. Please install git first: https://git-scm.com")
	}

	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot determine current directory: %w", err)
	}

	if !git.IsRepo(dir) {
		return fmt.Errorf("not inside a git repository. Navigate to a repo or run 'git init' first")
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	if cfg.APIKey == "" {
		return fmt.Errorf("no API key configured. Run 'patchnote register' to set it up")
	}

	gitClient := git.NewClient(dir)

	if !gitClient.HasAnyChanges() {
		return fmt.Errorf("no changes detected. Stage or modify files first")
	}

	input, err := prompts.CollectPRInput(gitClient, cfg)
	if err != nil {
		return fmt.Errorf("failed to collect git data: %w", err)
	}

	messages := prompts.BuildPRMessages(input)

	client, err := ai.New(cfg)
	if err != nil {
		return err
	}

	fmt.Println()
	ui.PrintInfo("Generating PR description...")

	resp, err := client.Complete(cmd.Context(), ai.Request{
		Messages:    messages,
		Temperature: cfg.Temperature,
	})
	if err != nil {
		return fmt.Errorf("AI generation failed: %w", err)
	}

	ui.PrintInfo("PR Description")
	fmt.Println(output.FormatPRDescription(resp.Content))
	fmt.Print(output.Separator())

	action := ui.PromptAction()

	switch action {
	case ui.ActionCommit:
		if err := copyToClipboard(resp.Content); err != nil {
			return fmt.Errorf("failed to copy: %w", err)
		}
		ui.PrintSuccess("Copied PR description to clipboard")
	case ui.ActionCopy:
		if err := copyToClipboard(resp.Content); err != nil {
			return fmt.Errorf("failed to copy: %w", err)
		}
		ui.PrintSuccess("Copied to clipboard")
	default:
		ui.PrintInfo("Cancelled")
	}

	return nil
}

func init() {
	rootCmd.AddCommand(prCmd)
}
