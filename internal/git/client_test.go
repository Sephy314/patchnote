package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.name", "Test")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	// Create initial commit so HEAD exists
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".gitkeep"), []byte(""), 0o644))
	cmd = exec.Command("git", "add", ".gitkeep")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "commit", "-m", "init")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	return dir
}

func TestIsRepo(t *testing.T) {
	dir := setupTestRepo(t)
	assert.True(t, IsRepo(dir))
}

func TestIsRepoFalse(t *testing.T) {
	dir := t.TempDir()
	assert.False(t, IsRepo(dir))
}

func TestIsInstalled(t *testing.T) {
	assert.True(t, IsInstalled())
}

func TestNewClient(t *testing.T) {
	dir := setupTestRepo(t)
	client := NewClient(dir)
	assert.NotNil(t, client)
}

func TestRootDir(t *testing.T) {
	dir := setupTestRepo(t)
	client := NewClient(dir)

	root, err := client.RootDir()
	require.NoError(t, err)
	assert.Equal(t, dir, root)
}

func TestRootDirNotRepo(t *testing.T) {
	dir := t.TempDir()
	client := NewClient(dir)

	_, err := client.RootDir()
	assert.Error(t, err)
}

func TestStagedDiff(t *testing.T) {
	dir := setupTestRepo(t)
	client := NewClient(dir)

	// Create and stage a file
	require.NoError(t, os.WriteFile(filepath.Join(dir, "test.txt"), []byte("hello"), 0o644))

	cmd := exec.Command("git", "add", "test.txt")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	diff, err := client.StagedDiff()
	require.NoError(t, err)
	assert.Contains(t, diff, "hello")
}

func TestStagedDiffEmpty(t *testing.T) {
	dir := setupTestRepo(t)
	client := NewClient(dir)

	diff, err := client.StagedDiff()
	require.NoError(t, err)
	assert.Empty(t, diff)
}

func TestStatus(t *testing.T) {
	dir := setupTestRepo(t)
	client := NewClient(dir)

	require.NoError(t, os.WriteFile(filepath.Join(dir, "new.txt"), []byte("content"), 0o644))

	status, err := client.Status()
	require.NoError(t, err)
	assert.Contains(t, status, "new.txt")
}

func TestHasAnyChanges(t *testing.T) {
	dir := setupTestRepo(t)
	client := NewClient(dir)

	assert.False(t, client.HasAnyChanges())

	require.NoError(t, os.WriteFile(filepath.Join(dir, "file.txt"), []byte("data"), 0o644))
	assert.True(t, client.HasAnyChanges())
}

func TestHasStagedChanges(t *testing.T) {
	dir := setupTestRepo(t)
	client := NewClient(dir)

	assert.False(t, client.HasStagedChanges())

	require.NoError(t, os.WriteFile(filepath.Join(dir, "file.txt"), []byte("data"), 0o644))

	cmd := exec.Command("git", "add", "file.txt")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	assert.True(t, client.HasStagedChanges())
}

func TestChangedFiles(t *testing.T) {
	dir := setupTestRepo(t)
	client := NewClient(dir)

	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b"), 0o644))

	cmd := exec.Command("git", "add", "a.txt", "b.txt")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	files, err := client.ChangedFiles()
	require.NoError(t, err)
	assert.Len(t, files, 2)
	assert.Contains(t, files, "a.txt")
	assert.Contains(t, files, "b.txt")
}

func TestDiffStat(t *testing.T) {
	dir := setupTestRepo(t)
	client := NewClient(dir)

	require.NoError(t, os.WriteFile(filepath.Join(dir, "file.txt"), []byte("content"), 0o644))

	cmd := exec.Command("git", "add", "file.txt")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	stat, err := client.DiffStat()
	require.NoError(t, err)
	assert.Contains(t, stat, "file.txt")
}

func TestCurrentBranch(t *testing.T) {
	dir := setupTestRepo(t)
	client := NewClient(dir)

	branch, err := client.CurrentBranch()
	require.NoError(t, err)
	assert.NotEmpty(t, branch)
}

func TestCommit(t *testing.T) {
	dir := setupTestRepo(t)
	client := NewClient(dir)

	require.NoError(t, os.WriteFile(filepath.Join(dir, "test.txt"), []byte("hello"), 0o644))

	cmd := exec.Command("git", "add", "test.txt")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	result, err := client.Commit("test: initial commit")
	require.NoError(t, err)
	assert.Contains(t, result, "commit")
}

func TestUntrackedFiles(t *testing.T) {
	dir := setupTestRepo(t)
	client := NewClient(dir)

	require.NoError(t, os.WriteFile(filepath.Join(dir, "untracked.txt"), []byte("data"), 0o644))

	files, err := client.UntrackedFiles()
	require.NoError(t, err)
	assert.Contains(t, files, "untracked.txt")
}

func TestHasCommits(t *testing.T) {
	dir := setupTestRepo(t)
	client := NewClient(dir)
	assert.True(t, client.HasCommits())
}

func TestHasCommitsNoCommits(t *testing.T) {
	dir := t.TempDir()

	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	client := NewClient(dir)
	assert.False(t, client.HasCommits())
}

func setupBareRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.name", "Test")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	return dir
}

func TestChangedFilesFirstCommit(t *testing.T) {
	dir := setupBareRepo(t)
	client := NewClient(dir)

	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b"), 0o644))

	cmd := exec.Command("git", "add", "a.txt", "b.txt")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	files, err := client.ChangedFiles()
	require.NoError(t, err)
	assert.Len(t, files, 2)
	assert.Contains(t, files, "a.txt")
	assert.Contains(t, files, "b.txt")
}

func TestDiffStatFirstCommit(t *testing.T) {
	dir := setupBareRepo(t)
	client := NewClient(dir)

	require.NoError(t, os.WriteFile(filepath.Join(dir, "file.txt"), []byte("content"), 0o644))

	cmd := exec.Command("git", "add", "file.txt")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	stat, err := client.DiffStat()
	require.NoError(t, err)
	assert.Contains(t, stat, "file.txt")
}

func TestFullDiffFirstCommit(t *testing.T) {
	dir := setupBareRepo(t)
	client := NewClient(dir)

	require.NoError(t, os.WriteFile(filepath.Join(dir, "file.txt"), []byte("content"), 0o644))

	cmd := exec.Command("git", "add", "file.txt")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	diff, err := client.FullDiff()
	require.NoError(t, err)
	assert.Contains(t, diff, "content")
}
