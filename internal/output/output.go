package output

import (
	"fmt"
	"strings"
)

// CommitMessage represents a parsed AI-generated commit message.
type CommitMessage struct {
	Title string
	Body  string
}

// ParseCommitMessage splits an AI response into title and body.
func ParseCommitMessage(raw string) *CommitMessage {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return &CommitMessage{}
	}

	lines := strings.SplitN(raw, "\n", 2)
	title := strings.TrimSpace(lines[0])

	var body string
	if len(lines) > 1 {
		body = strings.TrimSpace(lines[1])
	}

	return &CommitMessage{
		Title: title,
		Body:  body,
	}
}

// FullMessage returns the complete commit message (title + body).
func (cm *CommitMessage) FullMessage() string {
	if cm.Body == "" {
		return cm.Title
	}
	return cm.Title + "\n\n" + cm.Body
}

// FormatCommitMessage renders the commit message for display.
func FormatCommitMessage(msg *CommitMessage) string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("  " + msg.Title + "\n")

	if msg.Body != "" {
		sb.WriteString("\n")
		for _, line := range strings.Split(msg.Body, "\n") {
			sb.WriteString("  " + line + "\n")
		}
	}

	return sb.String()
}

// FormatReview renders a code review for display.
func FormatReview(review string) string {
	return "\n" + indent(review, "  ") + "\n"
}

// FormatPRDescription renders a PR description for display.
func FormatPRDescription(desc string) string {
	return "\n" + indent(desc, "  ") + "\n"
}

// Separator renders a horizontal rule.
func Separator() string {
	return "\n" + strings.Repeat("─", 50) + "\n"
}

// Prompt renders an interactive prompt choice line.
func Prompt(label, shortcut string) string {
	return fmt.Sprintf("  [%s] %s", shortcut, label)
}

func indent(s, prefix string) string {
	var sb strings.Builder
	for _, line := range strings.Split(s, "\n") {
		if line == "" {
			sb.WriteString("\n")
		} else {
			sb.WriteString(prefix + line + "\n")
		}
	}
	return strings.TrimRight(sb.String(), "\n")
}
