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

var rootCmd = &cobra.Command{
	Use:   "patchnote",
	Short: "AI-powered Git assistant",
	Long: `PatchNote analyses repository changes and generates
high-quality commit messages, patch notes, PR descriptions,
and code reviews using AI.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          runRoot,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		ui.PrintError(err.Error())
		os.Exit(1)
	}
}

func runRoot(cmd *cobra.Command, _ []string) error {
	// Pre-checks
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

	input, err := prompts.CollectCommitInput(gitClient, cfg)
	if err != nil {
		return fmt.Errorf("failed to collect git data: %w", err)
	}

	messages := prompts.BuildCommitMessages(input)

	client, err := ai.New(cfg)
	if err != nil {
		return err
	}

	fmt.Println()
	ui.PrintInfo("Generating commit message...")

	resp, err := client.Complete(cmd.Context(), ai.Request{
		Messages:    messages,
		Temperature: cfg.Temperature,
	})
	if err != nil {
		return fmt.Errorf("AI generation failed: %w", err)
	}

	msg := output.ParseCommitMessage(resp.Content)

	ui.PrintInfo("Generated commit message")
	fmt.Println(output.FormatCommitMessage(msg))
	fmt.Print(output.Separator())

	action := ui.PromptAction()

	switch action {
	case ui.ActionCommit:
		result, err := gitClient.Commit(msg.FullMessage())
		if err != nil {
			return err
		}
		ui.PrintSuccess(result)
	case ui.ActionEdit:
		edited := ui.PromptString("Enter edited commit message")
		if edited != "" {
			msg = output.ParseCommitMessage(edited)
			result, err := gitClient.Commit(msg.FullMessage())
			if err != nil {
				return err
			}
			ui.PrintSuccess(result)
		} else {
			ui.PrintInfo("Edit cancelled")
		}
	case ui.ActionCopy:
		if err := copyToClipboard(msg.FullMessage()); err != nil {
			return fmt.Errorf("failed to copy: %w", err)
		}
		ui.PrintSuccess("Copied to clipboard")
	case ui.ActionCancel:
		ui.PrintInfo("Cancelled")
	}

	return nil
}

func copyToClipboard(text string) error {
	return clipboardWrite(text)
}

//
//func getHomeDir() string {
//	home, err := os.UserHomeDir()
//	if err != nil {
//		return filepath.Join(os.TempDir(), "patchnote")
//	}
//	return filepath.Join(home, ".config", "patchnote")
//}
