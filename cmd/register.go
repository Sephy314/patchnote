package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Sephy314/patchnote/internal/ai"
	"github.com/Sephy314/patchnote/internal/config"
	"github.com/Sephy314/patchnote/internal/ui"
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register an AI provider API key",
	Long:  `Verify and save your Groq API key for use with PatchNote.`,
	RunE:  runRegister,
}

func runRegister(cmd *cobra.Command, _ []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("cannot load config: %w", err)
	}

	apiKey := ui.PromptString("Enter Groq API Key:")
	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	ui.PrintInfo("Verifying API key...")

	client := ai.NewGroqClient(cfg)
	if err := client.ValidateKey(cmd.Context(), apiKey); err != nil {
		return fmt.Errorf("API key verification failed: %w", err)
	}

	cfg.APIKey = apiKey
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	ui.PrintSuccess("API key verified.")
	ui.PrintSuccess("Configuration saved.")

	return nil
}

func init() {
	rootCmd.AddCommand(registerCmd)
}
