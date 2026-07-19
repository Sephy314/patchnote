package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Action represents the user's choice after generation.
type Action int

const (
	ActionCommit Action = iota
	ActionEdit
	ActionCopy
	ActionCancel
)

// PromptAction asks the user to choose an action after generation.
func PromptAction() Action {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Println("  [Y] Commit  [E] Edit  [C] Copy  [N] Cancel")
	fmt.Print("\n> ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	switch input {
	case "y", "yes":
		return ActionCommit
	case "e", "edit":
		return ActionEdit
	case "c", "copy":
		return ActionCopy
	default:
		return ActionCancel
	}
}

// PromptString asks the user for a string input with a prompt.
func PromptString(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt + "\n> ")
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// PromptConfirm asks the user a yes/no question.
func PromptConfirm(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt + " [y/N] > ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}

// PrintSuccess prints a success message.
func PrintSuccess(msg string) {
	fmt.Printf("  ✓ %s\n", msg)
}

// PrintError prints an error message.
func PrintError(msg string) {
	fmt.Printf("  ✗ %s\n", msg)
}

// PrintInfo prints an informational message.
func PrintInfo(msg string) {
	fmt.Printf("  %s\n", msg)
}
