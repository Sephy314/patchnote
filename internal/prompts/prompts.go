package prompts

import (
	"fmt"
	"strings"

	"github.com/Sephy314/patchnote/internal/ai"
	"github.com/Sephy314/patchnote/internal/config"
	"github.com/Sephy314/patchnote/internal/git"
)

const maxDiffLength = 5000

// CommitInput holds all data needed to generate a commit message.
type CommitInput struct {
	Diff        string
	Status      string
	Files       []string
	DiffStat    string
	Branch      string
	Language    string
}

// ReviewInput holds all data needed to generate a code review.
type ReviewInput struct {
	Diff     string
	Status   string
	Files    []string
	DiffStat string
	Branch   string
	Language string
}

// PRInput holds all data needed to generate a PR description.
type PRInput struct {
	Diff     string
	Status   string
	Files    []string
	DiffStat string
	Branch   string
	Language string
}

// CollectCommitInput gathers the required git data for commit message generation.
func CollectCommitInput(gitClient *git.Client, cfg *config.Config) (*CommitInput, error) {
	diff, err := combinedDiff(gitClient)
	if err != nil {
		return nil, err
	}

	status, err := gitClient.StatusPorcelain()
	if err != nil {
		return nil, err
	}

	files, err := gitClient.ChangedFiles()
	if err != nil {
		return nil, err
	}

	stat, err := gitClient.DiffStat()
	if err != nil {
		return nil, err
	}

	branch, err := gitClient.CurrentBranch()
	if err != nil {
		branch = "unknown"
	}

	files = gitClient.FilterIgnoredFiles(files)
	status = gitClient.FilterIgnoredLines(status)
	diff = gitClient.FilterIgnoredDiff(diff)

	return &CommitInput{
		Diff:     truncateDiff(diff),
		Status:   status,
		Files:    files,
		DiffStat: stat,
		Branch:   branch,
		Language: cfg.Language,
	}, nil
}

// CollectReviewInput gathers the required git data for code review generation.
func CollectReviewInput(gitClient *git.Client, cfg *config.Config) (*ReviewInput, error) {
	diff, err := combinedDiff(gitClient)
	if err != nil {
		return nil, err
	}

	status, err := gitClient.StatusPorcelain()
	if err != nil {
		return nil, err
	}

	files, err := gitClient.ChangedFiles()
	if err != nil {
		return nil, err
	}

	stat, err := gitClient.DiffStat()
	if err != nil {
		return nil, err
	}

	branch, err := gitClient.CurrentBranch()
	if err != nil {
		branch = "unknown"
	}

	files = gitClient.FilterIgnoredFiles(files)
	status = gitClient.FilterIgnoredLines(status)
	diff = gitClient.FilterIgnoredDiff(diff)

	return &ReviewInput{
		Diff:     truncateDiff(diff),
		Status:   status,
		Files:    files,
		DiffStat: stat,
		Branch:   branch,
		Language: cfg.Language,
	}, nil
}

// CollectPRInput gathers the required git data for PR description generation.
func CollectPRInput(gitClient *git.Client, cfg *config.Config) (*PRInput, error) {
	diff, err := combinedDiff(gitClient)
	if err != nil {
		return nil, err
	}

	status, err := gitClient.StatusPorcelain()
	if err != nil {
		return nil, err
	}

	files, err := gitClient.ChangedFiles()
	if err != nil {
		return nil, err
	}

	stat, err := gitClient.DiffStat()
	if err != nil {
		return nil, err
	}

	branch, err := gitClient.CurrentBranch()
	if err != nil {
		branch = "unknown"
	}

	files = gitClient.FilterIgnoredFiles(files)
	status = gitClient.FilterIgnoredLines(status)
	diff = gitClient.FilterIgnoredDiff(diff)

	return &PRInput{
		Diff:     truncateDiff(diff),
		Status:   status,
		Files:    files,
		DiffStat: stat,
		Branch:   branch,
		Language: cfg.Language,
	}, nil
}

// BuildCommitMessages returns the system and user messages for commit generation.
func BuildCommitMessages(input *CommitInput) []ai.Message {
	system := `Generate a Conventional Commit message: type(scope): description.
Use imperative mood. After title, blank line then bullet points for changes.
Respond in: ` + input.Language

	user := fmt.Sprintf(`Branch: %s

Changed files:
%s

Diff statistics:
%s

Git status:
%s

Diff:
%s`,
		input.Branch,
		strings.Join(input.Files, "\n"),
		input.DiffStat,
		input.Status,
		input.Diff,
	)

	return []ai.Message{
		{Role: "system", Content: system},
		{Role: "user", Content: user},
	}
}

// BuildReviewMessages returns the system and user messages for code review.
func BuildReviewMessages(input *ReviewInput) []ai.Message {
	system := `Generate a concise code review with: Summary, Positive Observations, Potential Bugs, Security, Performance, Suggestions.
Be specific. Respond in: ` + input.Language

	user := fmt.Sprintf(`Branch: %s

Changed files:
%s

Diff statistics:
%s

Git status:
%s

Diff:
%s`,
		input.Branch,
		strings.Join(input.Files, "\n"),
		input.DiffStat,
		input.Status,
		input.Diff,
	)

	return []ai.Message{
		{Role: "system", Content: system},
		{Role: "user", Content: user},
	}
}

// BuildPRMessages returns the system and user messages for PR description.
func BuildPRMessages(input *PRInput) []ai.Message {
	system := `Generate a Markdown PR description with: Title, Summary, Motivation, Main Changes, Testing, Breaking Changes, Checklist.
Use Conventional Commits for title. Respond in: ` + input.Language

	user := fmt.Sprintf(`Branch: %s

Changed files:
%s

Diff statistics:
%s

Git status:
%s

Diff:
%s`,
		input.Branch,
		strings.Join(input.Files, "\n"),
		input.DiffStat,
		input.Status,
		input.Diff,
	)

	return []ai.Message{
		{Role: "system", Content: system},
		{Role: "user", Content: user},
	}
}

func combinedDiff(gitClient *git.Client) (string, error) {
	staged, err := gitClient.StagedDiff()
	if err != nil {
		return "", err
	}

	unstaged, err := gitClient.UnstagedDiff()
	if err != nil {
		return "", err
	}

	if staged == "" && unstaged == "" {
		return "", fmt.Errorf("no changes detected. Stage or modify files first")
	}

	var parts []string
	if staged != "" {
		parts = append(parts, "--- Staged Changes ---\n"+staged)
	}
	if unstaged != "" {
		parts = append(parts, "--- Unstaged Changes ---\n"+unstaged)
	}

	return strings.Join(parts, "\n\n"), nil
}

func truncateDiff(diff string) string {
	if len(diff) <= maxDiffLength {
		return diff
	}

	truncated := diff[:maxDiffLength]
	lastNewline := strings.LastIndex(truncated, "\n")
	if lastNewline > 0 {
		truncated = truncated[:lastNewline]
	}

	return truncated + "\n\n[... diff truncated for brevity ...]"
}
