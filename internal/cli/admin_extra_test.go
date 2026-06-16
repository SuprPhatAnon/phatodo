package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/SuprPhatAnon/phatodo/internal/config"
	"github.com/SuprPhatAnon/phatodo/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestPromptPasswordReadsNonTerminalStdin(t *testing.T) {
	oldStdin := os.Stdin
	readEnd, writeEnd, err := os.Pipe()
	require.NoError(t, err)
	t.Cleanup(func() {
		os.Stdin = oldStdin
		_ = readEnd.Close()
	})

	_, err = writeEnd.WriteString("secret\n")
	require.NoError(t, err)
	require.NoError(t, writeEnd.Close())
	os.Stdin = readEnd

	var stderr bytes.Buffer
	password, err := promptPassword("password: ", &stderr)
	require.NoError(t, err)
	require.Equal(t, "secret", password)
	require.Equal(t, "password: ", stderr.String())
}

func TestRunAdminInitRejectsMismatchedPasswords(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	prompts := []string{"secret", "different"}

	oldPrompt := readPasswordPrompt
	readPasswordPrompt = func(prompt string, _ io.Writer) (string, error) {
		password := prompts[0]
		prompts = prompts[1:]
		return password, nil
	}
	t.Cleanup(func() { readPasswordPrompt = oldPrompt })

	code := runAdminInit([]string{"-u", "alice", "--url", "http://example.invalid"}, &stdout, &stderr)
	require.Equal(t, 1, code)
	require.Empty(t, stdout.String())
	require.Contains(t, stderr.String(), "passwords do not match")
}

func TestRunAdminInitHandlesPromptAndAPIErrors(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	oldPrompt := readPasswordPrompt
	readPasswordPrompt = func(prompt string, _ io.Writer) (string, error) {
		return "", errors.New("no tty")
	}
	t.Cleanup(func() { readPasswordPrompt = oldPrompt })

	code := runAdminInit([]string{"-u", "alice", "--url", "http://example.invalid"}, &stdout, &stderr)
	require.Equal(t, 1, code)
	require.Contains(t, stderr.String(), "failed to read password")

	stderr.Reset()
	readPasswordPrompt = func(prompt string, _ io.Writer) (string, error) {
		return "secret", nil
	}

	oldFactory := newAPIClient
	newAPIClient = func(cfg config.LocalConfig) (apiClient, error) {
		return &fakeAPIClient{
			initAdminFn: func(ctx context.Context, req domain.AdminInitRequest) (domain.AdminInitResponse, error) {
				return domain.AdminInitResponse{}, errors.New("admin exists")
			},
		}, nil
	}
	t.Cleanup(func() { newAPIClient = oldFactory })

	code = runAdminInit([]string{"-u", "alice", "--url", "http://example.invalid"}, &stdout, &stderr)
	require.Equal(t, 1, code)
	require.Contains(t, stderr.String(), "admin exists")
}

func TestRunAdminInitHandlesConfirmationPromptError(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	call := 0

	oldPrompt := readPasswordPrompt
	readPasswordPrompt = func(prompt string, _ io.Writer) (string, error) {
		call++
		if call == 2 {
			return "", errors.New("confirm failed")
		}
		return "secret", nil
	}
	t.Cleanup(func() { readPasswordPrompt = oldPrompt })

	code := runAdminInit([]string{"-u", "alice", "--url", "http://example.invalid"}, &stdout, &stderr)
	require.Equal(t, 1, code)
	require.Contains(t, stderr.String(), "failed to read password confirmation")
}

func TestRunAdminBootstrapHandlesPromptAndAPIErrors(t *testing.T) {
	workdir := t.TempDir()
	oldwd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(oldwd))
	})
	require.NoError(t, os.Chdir(workdir))

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	oldPrompt := readPasswordPrompt
	readPasswordPrompt = func(prompt string, _ io.Writer) (string, error) {
		return "", errors.New("no tty")
	}
	t.Cleanup(func() { readPasswordPrompt = oldPrompt })

	code := runAdminBootstrap([]string{"-u", "alice", "--url", "http://example.invalid"}, &stdout, &stderr)
	require.Equal(t, 1, code)
	require.Contains(t, stderr.String(), "failed to read password")

	stderr.Reset()
	readPasswordPrompt = func(prompt string, _ io.Writer) (string, error) {
		return "secret", nil
	}

	oldFactory := newAPIClient
	newAPIClient = func(cfg config.LocalConfig) (apiClient, error) {
		return &fakeAPIClient{
			bootstrapAdminFn: func(ctx context.Context, req domain.AdminBootstrapRequest) (domain.AdminBootstrapResponse, error) {
				return domain.AdminBootstrapResponse{}, errors.New("invalid credentials")
			},
		}, nil
	}
	t.Cleanup(func() { newAPIClient = oldFactory })

	code = runAdminBootstrap([]string{"-u", "alice", "--url", "http://example.invalid"}, &stdout, &stderr)
	require.Equal(t, 1, code)
	require.Contains(t, stderr.String(), "invalid credentials")
}

func TestRepoBaseNameFallback(t *testing.T) {
	require.Equal(t, "phatodo", repoBaseName("/"))
	require.Equal(t, "repo", repoBaseName("/tmp/repo/"))
}
