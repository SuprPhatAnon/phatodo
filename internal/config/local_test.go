package config

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultLocalConfig(t *testing.T) {
	cfg := DefaultLocalConfig()

	require.Equal(t, "http://localhost:8080", cfg.APIURL)
	require.Equal(t, "default", cfg.WorkspaceID)
	require.Equal(t, "default", cfg.ProjectID)
	require.Equal(t, "replace-me", cfg.AccessKey)
	require.Equal(t, "replace-me", cfg.AccessSecret)
}

func TestWriteAndReadLocalRoundTrip(t *testing.T) {
	workdir := t.TempDir()
	cfg := LocalConfig{
		APIURL:       "https://api.example.test",
		WorkspaceID:  "workspace-1",
		ProjectID:    "project-1",
		AccessKey:    "key-1",
		AccessSecret: "secret-1",
	}

	path, err := WriteLocal(workdir, cfg)
	require.NoError(t, err)
	require.Equal(t, LocalPath(workdir), path)

	info, err := os.Stat(path)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o600), info.Mode().Perm())

	got, readPath, err := ReadLocal(workdir)
	require.NoError(t, err)
	require.Equal(t, path, readPath)
	require.Equal(t, cfg, got)
}

func TestWriteLocalRejectsExistingConfig(t *testing.T) {
	workdir := t.TempDir()
	cfg := DefaultLocalConfig()

	path, err := WriteLocal(workdir, cfg)
	require.NoError(t, err)

	againPath, err := WriteLocal(workdir, cfg)
	require.Error(t, err)
	require.Equal(t, path, againPath)
	require.Contains(t, err.Error(), "already exists")
}

func TestReadLocalReturnsPathOnReadAndDecodeErrors(t *testing.T) {
	workdir := t.TempDir()

	_, missingPath, err := ReadLocal(workdir)
	require.Error(t, err)
	require.Equal(t, LocalPath(workdir), missingPath)

	require.NoError(t, os.MkdirAll(strings.TrimSuffix(LocalPath(workdir), LocalFileName), 0o700))
	require.NoError(t, os.WriteFile(LocalPath(workdir), []byte("{not-json"), 0o600))

	_, invalidPath, err := ReadLocal(workdir)
	require.Error(t, err)
	require.Equal(t, LocalPath(workdir), invalidPath)
}
