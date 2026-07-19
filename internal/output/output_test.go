package output

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCommitMessage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		title    string
		hasBody  bool
	}{
		{
			name:    "simple message",
			input:   "feat(auth): add login",
			title:   "feat(auth): add login",
			hasBody: false,
		},
		{
			name: "message with body",
			input: `feat(auth): add login

- Add login endpoint
- Add token validation`,
			title:   "feat(auth): add login",
			hasBody: true,
		},
		{
			name:    "empty message",
			input:   "",
			title:   "",
			hasBody: false,
		},
		{
			name:    "whitespace only",
			input:   "  \n  \n",
			title:   "",
			hasBody: false,
		},
		{
			name: "multiple body lines",
			input: `fix: resolve memory leak

- Fix buffer pool not releasing
- Add cleanup on shutdown
- Update tests`,
			title:   "fix: resolve memory leak",
			hasBody: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := ParseCommitMessage(tt.input)
			assert.Equal(t, tt.title, msg.Title)

			if tt.hasBody {
				assert.NotEmpty(t, msg.Body)
			} else {
				assert.Empty(t, msg.Body)
			}
		})
	}
}

func TestFullMessage(t *testing.T) {
	msg := &CommitMessage{
		Title: "feat: add feature",
		Body:  "- Add feature A\n- Add feature B",
	}

	full := msg.FullMessage()
	assert.Contains(t, full, "feat: add feature")
	assert.Contains(t, full, "- Add feature A")
}

func TestFullMessageNoBody(t *testing.T) {
	msg := &CommitMessage{
		Title: "feat: add feature",
		Body:  "",
	}

	full := msg.FullMessage()
	assert.Equal(t, "feat: add feature", full)
}

func TestFormatCommitMessage(t *testing.T) {
	msg := &CommitMessage{
		Title: "feat(auth): implement login",
		Body:  "- Add login endpoint\n- Add validation",
	}

	formatted := FormatCommitMessage(msg)
	assert.Contains(t, formatted, "feat(auth): implement login")
	assert.Contains(t, formatted, "- Add login endpoint")
	assert.Contains(t, formatted, "- Add validation")
}

func TestFormatCommitMessageNoBody(t *testing.T) {
	msg := &CommitMessage{
		Title: "chore: update deps",
		Body:  "",
	}

	formatted := FormatCommitMessage(msg)
	assert.Contains(t, formatted, "chore: update deps")
}

func TestSeparator(t *testing.T) {
	sep := Separator()
	assert.True(t, strings.HasPrefix(sep, "\n"))
	assert.Contains(t, sep, "─")
}

func TestPrompt(t *testing.T) {
	result := Prompt("Commit", "Y")
	assert.Equal(t, "  [Y] Commit", result)
}

func TestFormatReview(t *testing.T) {
	review := "## Summary\nThis is a review."
	formatted := FormatReview(review)
	assert.Contains(t, formatted, "## Summary")
	assert.Contains(t, formatted, "  ## Summary")
}

func TestFormatPRDescription(t *testing.T) {
	desc := "## Title\nMy PR"
	formatted := FormatPRDescription(desc)
	assert.Contains(t, formatted, "## Title")
	assert.Contains(t, formatted, "  ## Title")
}
