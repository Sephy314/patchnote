package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, "groq", cfg.Provider)
	assert.Equal(t, "llama-3.3-70b-versatile", cfg.Model)
	assert.Equal(t, 0.2, cfg.Temperature)
	assert.Equal(t, "english", cfg.Language)
}

func TestDir(t *testing.T) {
	dir, err := Dir()
	require.NoError(t, err)

	home, err := os.UserHomeDir()
	require.NoError(t, err)

	expected := filepath.Join(home, ".config", "patchnote")
	assert.Equal(t, expected, dir)
}

func TestPath(t *testing.T) {
	path, err := Path()
	require.NoError(t, err)

	assert.True(t, filepath.IsAbs(path))
	assert.True(t, filepath.Ext(path) == ".yaml")
}

func TestLoadDefault(t *testing.T) {
	// When no config file exists, should return defaults
	// We use a temp dir to simulate this
	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpHome)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "groq", cfg.Provider)
	assert.Equal(t, "llama-3.3-70b-versatile", cfg.Model)
}

func TestSaveAndLoad(t *testing.T) {
	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	original := &Config{
		Provider:    "groq",
		Model:       "custom-model",
		Temperature: 0.5,
		Language:    "french",
		APIKey:      "test-key-123",
	}

	err := Save(original)
	require.NoError(t, err)

	loaded, err := Load()
	require.NoError(t, err)

	assert.Equal(t, original.Provider, loaded.Provider)
	assert.Equal(t, original.Model, loaded.Model)
	assert.Equal(t, original.Temperature, loaded.Temperature)
	assert.Equal(t, original.Language, loaded.Language)
	assert.Equal(t, original.APIKey, loaded.APIKey)
}

func TestSaveCreatesDir(t *testing.T) {
	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	cfg := DefaultConfig()
	cfg.APIKey = "test"

	err := Save(cfg)
	require.NoError(t, err)

	dir, err := Dir()
	require.NoError(t, err)

	info, err := os.Stat(dir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}
