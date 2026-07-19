package prompts

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/patchnote/patchnote/internal/config"
	"github.com/patchnote/patchnote/internal/git"
)

func TestBuildCommitMessages(t *testing.T) {
	input := &CommitInput{
		Diff:     "--- a/file.go\n+++ b/file.go\n@@ -1 +1 @@\n-old\n+new",
		Status:   "M file.go",
		Files:    []string{"file.go"},
		DiffStat: "1 file changed, 1 insertion(+), 1 deletion(-)",
		Branch:   "main",
		Language: "english",
	}

	messages := BuildCommitMessages(input)

	require.Len(t, messages, 2)
	assert.Equal(t, "system", messages[0].Role)
	assert.Equal(t, "user", messages[1].Role)
	assert.Contains(t, messages[0].Content, "Conventional Commit")
	assert.Contains(t, messages[0].Content, "english")
	assert.Contains(t, messages[1].Content, "main")
	assert.Contains(t, messages[1].Content, "file.go")
	assert.Contains(t, messages[1].Content, "M file.go")
}

func TestBuildReviewMessages(t *testing.T) {
	input := &ReviewInput{
		Diff:     "diff content",
		Status:   "M file.go",
		Files:    []string{"file.go"},
		DiffStat: "1 file changed",
		Branch:   "feature/x",
		Language: "english",
	}

	messages := BuildReviewMessages(input)

	require.Len(t, messages, 2)
	assert.Equal(t, "system", messages[0].Role)
	assert.Equal(t, "user", messages[1].Role)
	assert.Contains(t, messages[0].Content, "code review")
	assert.Contains(t, messages[0].Content, "Security")
	assert.Contains(t, messages[1].Content, "feature/x")
}

func TestBuildPRMessages(t *testing.T) {
	input := &PRInput{
		Diff:     "diff content",
		Status:   "M file.go",
		Files:    []string{"file.go"},
		DiffStat: "1 file changed",
		Branch:   "feature/y",
		Language: "english",
	}

	messages := BuildPRMessages(input)

	require.Len(t, messages, 2)
	assert.Equal(t, "system", messages[0].Role)
	assert.Equal(t, "user", messages[1].Role)
	assert.Contains(t, messages[0].Content, "PR description")
	assert.Contains(t, messages[0].Content, "Breaking Changes")
	assert.Contains(t, messages[1].Content, "feature/y")
}

func TestBuildCommitMessagesFrench(t *testing.T) {
	input := &CommitInput{
		Diff:     "diff",
		Status:   "",
		Files:    []string{},
		DiffStat: "",
		Branch:   "main",
		Language: "french",
	}

	messages := BuildCommitMessages(input)
	assert.Contains(t, messages[0].Content, "french")
}

func TestCollectCommitInput(t *testing.T) {
	dir := setupTestGitRepo(t)
	client := git.NewClient(dir)
	cfg := &config.Config{Language: "english"}

	require.NoError(t, os.WriteFile(filepath.Join(dir, "test.go"), []byte("package main\n"), 0o644))
	gitAdd(t, dir, "test.go")

	input, err := CollectCommitInput(client, cfg)
	require.NoError(t, err)
	assert.NotEmpty(t, input.Diff)
	assert.NotEmpty(t, input.Files)
	assert.Equal(t, "english", input.Language)
}

func TestCollectReviewInput(t *testing.T) {
	dir := setupTestGitRepo(t)
	client := git.NewClient(dir)
	cfg := &config.Config{Language: "english"}

	require.NoError(t, os.WriteFile(filepath.Join(dir, "test.go"), []byte("package main\n"), 0o644))
	gitAdd(t, dir, "test.go")

	input, err := CollectReviewInput(client, cfg)
	require.NoError(t, err)
	assert.NotEmpty(t, input.Diff)
	assert.Equal(t, "english", input.Language)
}

func TestCollectPRInput(t *testing.T) {
	dir := setupTestGitRepo(t)
	client := git.NewClient(dir)
	cfg := &config.Config{Language: "english"}

	require.NoError(t, os.WriteFile(filepath.Join(dir, "test.go"), []byte("package main\n"), 0o644))
	gitAdd(t, dir, "test.go")

	input, err := CollectPRInput(client, cfg)
	require.NoError(t, err)
	assert.NotEmpty(t, input.Diff)
	assert.Equal(t, "english", input.Language)
}

func TestBuildCommitMessagesContainsAllInput(t *testing.T) {
	input := &CommitInput{
		Diff:     "some diff content here",
		Status:   "M main.go",
		Files:    []string{"main.go", "utils.go"},
		DiffStat: "2 files changed",
		Branch:   "develop",
		Language: "english",
	}

	messages := BuildCommitMessages(input)

	userContent := messages[1].Content
	assert.Contains(t, userContent, "some diff content here")
	assert.Contains(t, userContent, "M main.go")
	assert.Contains(t, userContent, "main.go")
	assert.Contains(t, userContent, "utils.go")
	assert.Contains(t, userContent, "2 files changed")
	assert.Contains(t, userContent, "develop")
}

func TestBuildReviewMessagesAllSections(t *testing.T) {
	input := &ReviewInput{
		Diff: "test diff",
	}

	messages := BuildReviewMessages(input)
	systemContent := messages[0].Content

	assert.Contains(t, systemContent, "Summary")
	assert.Contains(t, systemContent, "Positive Observations")
	assert.Contains(t, systemContent, "Potential Bugs")
	assert.Contains(t, systemContent, "Security")
	assert.Contains(t, systemContent, "Performance")
	assert.Contains(t, systemContent, "Suggestions")
}

func TestBuildPRMessagesAllSections(t *testing.T) {
	input := &PRInput{
		Diff: "test diff",
	}

	messages := BuildPRMessages(input)
	systemContent := messages[0].Content

	assert.Contains(t, systemContent, "Title")
	assert.Contains(t, systemContent, "Summary")
	assert.Contains(t, systemContent, "Motivation")
	assert.Contains(t, systemContent, "Main Changes")
	assert.Contains(t, systemContent, "Testing")
	assert.Contains(t, systemContent, "Breaking Changes")
	assert.Contains(t, systemContent, "Checklist")
}

func setupTestGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	gitRun(t, dir, "init")
	gitRun(t, dir, "config", "user.email", "test@test.com")
	gitRun(t, dir, "config", "user.name", "Test")

	// Create initial commit so HEAD exists
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".gitkeep"), []byte(""), 0o644))
	gitRun(t, dir, "add", ".gitkeep")
	gitRun(t, dir, "commit", "-m", "init")

	return dir
}

func gitAdd(t *testing.T, dir string, files ...string) {
	t.Helper()
	args := append([]string{"add"}, files...)
	gitRun(t, dir, args...)
}

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v failed: %s", args, out)
}
