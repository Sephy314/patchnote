package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// Client provides operations for interacting with a Git repository.
type Client struct {
	repoDir string
}

// NewClient creates a Git client rooted at the given directory.
func NewClient(repoDir string) *Client {
	return &Client{repoDir: repoDir}
}

// IsRepo checks if the given directory is inside a Git repository.
func IsRepo(dir string) bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = dir
	return cmd.Run() == nil
}

// IsInstalled checks if Git is available on the system.
func IsInstalled() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

// RootDir returns the root directory of the Git repository.
func (c *Client) RootDir() (string, error) {
	out, err := c.git("rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("not inside a git repository: %w", err)
	}
	return strings.TrimSpace(out), nil
}

// StagedDiff returns the diff of staged changes.
func (c *Client) StagedDiff() (string, error) {
	out, err := c.git("diff", "--cached")
	if err != nil {
		return "", fmt.Errorf("failed to get staged diff: %w", err)
	}
	return out, nil
}

// UnstagedDiff returns the diff of unstaged changes.
func (c *Client) UnstagedDiff() (string, error) {
	out, err := c.git("diff")
	if err != nil {
		return "", fmt.Errorf("failed to get unstaged diff: %w", err)
	}
	return out, nil
}

// FullDiff returns the combined diff of all changes (staged + unstaged).
// On the first commit (no HEAD), returns only staged diff.
func (c *Client) FullDiff() (string, error) {
	if c.HasCommits() {
		out, err := c.git("diff", "HEAD")
		if err != nil {
			return "", fmt.Errorf("failed to get full diff: %w", err)
		}
		return out, nil
	}

	staged, err := c.StagedDiff()
	if err != nil {
		return "", err
	}
	return staged, nil
}

// Status returns the raw git status output.
func (c *Client) Status() (string, error) {
	out, err := c.git("status", "--short")
	if err != nil {
		return "", fmt.Errorf("failed to get git status: %w", err)
	}
	return out, nil
}

// StatusPorcelain returns porcelain-format status output.
func (c *Client) StatusPorcelain() (string, error) {
	out, err := c.git("status", "--porcelain")
	if err != nil {
		return "", fmt.Errorf("failed to get porcelain status: %w", err)
	}
	return out, nil
}

// ChangedFiles returns the list of changed files (staged + unstaged).
// On the first commit (no HEAD), returns only staged files.
func (c *Client) ChangedFiles() ([]string, error) {
	var out string
	var err error

	if c.HasCommits() {
		out, err = c.git("diff", "--name-only", "HEAD")
	} else {
		out, err = c.git("diff", "--cached", "--name-only")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to list changed files: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	var files []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

// DiffStat returns diffstat output for the current changes.
// On the first commit (no HEAD), returns stat for staged files only.
func (c *Client) DiffStat() (string, error) {
	var out string
	var err error

	if c.HasCommits() {
		out, err = c.git("diff", "--stat", "HEAD")
	} else {
		out, err = c.git("diff", "--cached", "--stat")
	}
	if err != nil {
		return "", fmt.Errorf("failed to get diff stat: %w", err)
	}
	return out, nil
}

// HasCommits checks if the repository has at least one commit.
func (c *Client) HasCommits() bool {
	_, err := c.git("rev-parse", "HEAD")
	return err == nil
}

// HasStagedChanges checks if there are any staged changes.
func (c *Client) HasStagedChanges() bool {
	out, err := c.git("diff", "--cached", "--name-only")
	if err != nil {
		return false
	}
	return strings.TrimSpace(out) != ""
}

// HasAnyChanges checks if there are any staged or unstaged changes.
func (c *Client) HasAnyChanges() bool {
	out, err := c.git("status", "--porcelain")
	if err != nil {
		return false
	}
	return strings.TrimSpace(out) != ""
}

// Commit creates a commit with the given message.
func (c *Client) Commit(message string) (string, error) {
	out, err := c.git("commit", "-m", message)
	if err != nil {
		return "", fmt.Errorf("git commit failed: %w", err)
	}
	return strings.TrimSpace(out), nil
}

// CurrentBranch returns the name of the current branch.
func (c *Client) CurrentBranch() (string, error) {
	out, err := c.git("branch", "--show-current")
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(out), nil
}

// UntrackedFiles returns a list of untracked files.
func (c *Client) UntrackedFiles() ([]string, error) {
	out, err := c.git("ls-files", "--others", "--exclude-standard")
	if err != nil {
		return nil, fmt.Errorf("failed to list untracked files: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	var files []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

// IsIgnored checks if a file path is ignored by .gitignore rules.
func (c *Client) IsIgnored(path string) bool {
	_, err := c.git("check-ignore", "-q", path)
	return err == nil
}

// FilterIgnoredFiles returns only the files that are not ignored by .gitignore.
func (c *Client) FilterIgnoredFiles(files []string) []string {
	var filtered []string
	for _, f := range files {
		if !c.IsIgnored(f) {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

// FilterIgnoredLines filters git status/diff output lines, removing entries for ignored files.
func (c *Client) FilterIgnoredLines(output string) string {
	lines := strings.Split(output, "\n")
	var filtered []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		path := extractPathFromStatusLine(trimmed)
		if path != "" && c.IsIgnored(path) {
			continue
		}
		filtered = append(filtered, line)
	}
	return strings.Join(filtered, "\n")
}

// FilterIgnoredDiff filters a full diff output, removing sections for ignored files.
func (c *Client) FilterIgnoredDiff(diff string) string {
	lines := strings.Split(diff, "\n")
	var filtered []string
	skipSection := false

	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git ") {
			// Extract file path from "diff --git a/path b/path"
			parts := strings.Split(line, " b/")
			if len(parts) == 2 {
				filePath := parts[1]
				skipSection = c.IsIgnored(filePath)
			} else {
				skipSection = false
			}
		}

		if !skipSection {
			filtered = append(filtered, line)
		}
	}
	return strings.Join(filtered, "\n")
}

// extractPathFromStatusLine extracts the file path from a git status porcelain line.
func extractPathFromStatusLine(line string) string {
	if len(line) > 3 && line[2] == ' ' {
		return strings.TrimSpace(line[3:])
	}
	return ""
}

// git runs a git command in the repository directory and returns stdout.
func (c *Client) git(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = c.repoDir

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %s: %w", args[0], strings.TrimSpace(stderr.String()), err)
	}

	return stdout.String(), nil
}
