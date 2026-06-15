package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/SuprPhatAnon/phatodo/internal/config"
	"github.com/SuprPhatAnon/phatodo/internal/domain"
	"golang.org/x/term"
)

var readPasswordPrompt = promptPassword

func runAdminInit(args []string, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseAdminInitArgs(args)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}

	password, err := readPasswordPrompt("admin password: ", stderr)
	if err != nil {
		fmt.Fprintf(stderr, "failed to read password: %v\n", err)
		return 1
	}
	confirm, err := readPasswordPrompt("confirm password: ", stderr)
	if err != nil {
		fmt.Fprintf(stderr, "failed to read password confirmation: %v\n", err)
		return 1
	}
	if password != confirm {
		fmt.Fprintln(stderr, "passwords do not match")
		return 1
	}

	client, err := newAPIClient(config.LocalConfig{APIURL: opts.apiURL})
	if err != nil {
		fmt.Fprintf(stderr, "failed to initialize api client: %v\n", err)
		return 1
	}

	resp, err := client.InitAdmin(context.Background(), domain.AdminInitRequest{
		Username: opts.username,
		Password: password,
	})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	writeTOONListItemStart(stdout, 0, "user_id", resp.UserID)
	writeTOONField(stdout, 1, "username", resp.Username)
	writeTOONField(stdout, 1, "access_key", resp.AccessKey)
	writeTOONField(stdout, 1, "access_secret", resp.AccessSecret)
	return 0
}

func runAdminBootstrap(args []string, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseAdminBootstrapArgs(args)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}

	workdir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "failed to determine working directory: %v\n", err)
		return 1
	}

	projectName := opts.projectName
	if projectName == "" {
		projectName = repoBaseName(workdir)
	}
	workspaceName := opts.workspaceName
	if workspaceName == "" {
		workspaceName = repoBaseName(workdir)
	}

	password, err := readPasswordPrompt("admin password: ", stderr)
	if err != nil {
		fmt.Fprintf(stderr, "failed to read password: %v\n", err)
		return 1
	}

	client, err := newAPIClient(config.LocalConfig{APIURL: opts.apiURL})
	if err != nil {
		fmt.Fprintf(stderr, "failed to initialize api client: %v\n", err)
		return 1
	}

	resp, err := client.BootstrapAdmin(context.Background(), domain.AdminBootstrapRequest{
		Username:      opts.username,
		Password:      password,
		WorkspaceName: workspaceName,
		ProjectName:   projectName,
	})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	path, err := config.WriteLocal(workdir, config.LocalConfig{
		APIURL:       opts.apiURL,
		WorkspaceID:  resp.WorkspaceID,
		ProjectID:    resp.ProjectID,
		AccessKey:    resp.AccessKey,
		AccessSecret: resp.AccessSecret,
	})
	if err != nil {
		fmt.Fprintf(stderr, "failed to write local config: %v\n", err)
		return 1
	}

	writeTOONListItemStart(stdout, 0, "workspace_id", resp.WorkspaceID)
	writeTOONField(stdout, 1, "project_id", resp.ProjectID)
	writeTOONField(stdout, 1, "access_key", resp.AccessKey)
	writeTOONField(stdout, 1, "access_secret", resp.AccessSecret)
	writeTOONField(stdout, 1, "config_path", path)
	return 0
}

type adminInitOptions struct {
	username string
	apiURL   string
}

type adminBootstrapOptions struct {
	username      string
	apiURL        string
	workspaceName string
	projectName   string
}

func parseAdminInitArgs(args []string) (adminInitOptions, error) {
	fs := flag.NewFlagSet("admin init", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var usernameLong string
	var apiURLLong string
	var usernameShort string
	var apiURLShort string

	fs.StringVar(&usernameShort, "u", "", "")
	fs.StringVar(&usernameLong, "username", "", "")
	fs.StringVar(&apiURLShort, "url", "", "")
	fs.StringVar(&apiURLLong, "api-url", "", "")

	if err := fs.Parse(args); err != nil {
		return adminInitOptions{}, fmt.Errorf("invalid admin init flags: %w", err)
	}
	if fs.NArg() > 0 {
		return adminInitOptions{}, fmt.Errorf("admin init does not accept positional arguments")
	}

	username := firstNonEmpty(usernameShort, usernameLong)
	apiURL := firstNonEmpty(apiURLShort, apiURLLong)
	if username == "" {
		return adminInitOptions{}, fmt.Errorf("admin init requires -u <username>")
	}
	if apiURL == "" {
		return adminInitOptions{}, fmt.Errorf("admin init requires --url <api-server-url>")
	}

	return adminInitOptions{username: username, apiURL: apiURL}, nil
}

func parseAdminBootstrapArgs(args []string) (adminBootstrapOptions, error) {
	fs := flag.NewFlagSet("admin bootstrap", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var usernameShort string
	var usernameLong string
	var apiURLShort string
	var apiURLLong string
	var workspaceName string
	var projectName string

	fs.StringVar(&usernameShort, "u", "", "")
	fs.StringVar(&usernameLong, "username", "", "")
	fs.StringVar(&apiURLShort, "url", "", "")
	fs.StringVar(&apiURLLong, "api-url", "", "")
	fs.StringVar(&workspaceName, "workspace-name", "", "")
	fs.StringVar(&projectName, "project-name", "", "")

	if err := fs.Parse(args); err != nil {
		return adminBootstrapOptions{}, fmt.Errorf("invalid admin bootstrap flags: %w", err)
	}
	if fs.NArg() > 0 {
		return adminBootstrapOptions{}, fmt.Errorf("admin bootstrap does not accept positional arguments")
	}

	username := firstNonEmpty(usernameShort, usernameLong)
	apiURL := firstNonEmpty(apiURLShort, apiURLLong)
	if username == "" {
		return adminBootstrapOptions{}, fmt.Errorf("admin bootstrap requires -u <username>")
	}
	if apiURL == "" {
		return adminBootstrapOptions{}, fmt.Errorf("admin bootstrap requires --url <api-server-url>")
	}

	return adminBootstrapOptions{
		username:      username,
		apiURL:        apiURL,
		workspaceName: workspaceName,
		projectName:   projectName,
	}, nil
}

func promptPassword(prompt string, stderr io.Writer) (string, error) {
	fmt.Fprint(stderr, prompt)
	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		password, err := term.ReadPassword(fd)
		fmt.Fprintln(stderr)
		if err != nil {
			return "", err
		}
		return string(password), nil
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func repoBaseName(workdir string) string {
	name := filepath.Base(filepath.Clean(workdir))
	if name == "." || name == string(filepath.Separator) || name == "" {
		return "phatodo"
	}
	return name
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
