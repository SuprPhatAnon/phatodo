package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/SuprPhatAnon/phatodo/internal/config"
)

var commandGroups = []struct {
	Name     string
	Commands []string
}{
	{"setup", []string{"init", "wipe -y"}},
	{"admin", []string{"admin init", "admin bootstrap"}},
	{"epic", []string{"epic create", "epic list", "epic show", "epic update", "epic complete", "epic delete"}},
	{"task", []string{"task create", "task list", "task show", "task update", "task delete"}},
	{"subtask", []string{"subtask create", "subtask list", "subtask update", "subtask delete"}},
	{"comment", []string{"comment add", "comment list", "comment update", "comment delete"}},
	{"dep", []string{"dep add", "dep remove", "dep list"}},
	{"config", []string{"config list", "config get", "config set", "config unset"}},
	{"query", []string{"search", "history", "list"}},
}

// Run is the CLI boundary. It currently validates the Trekker-compatible
// command shape while the API client and persistence behavior are built out.
func Run(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		printHelp(stdout)
		return 0
	}

	if args[0] == "--toon" {
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintln(stderr, "missing command after --toon")
		return 2
	}

	if args[0] == "init" {
		return runInit(stdout, stderr)
	}

	if len(args) >= 2 && args[0] == "admin" {
		switch args[1] {
		case "init":
			return runAdminInit(args[2:], stdout, stderr)
		case "bootstrap":
			return runAdminBootstrap(args[2:], stdout, stderr)
		}
	}

	if len(args) >= 2 && args[0] == "config" && args[1] == "list" {
		return runConfigList(stdout, stderr)
	}
	if len(args) >= 2 && args[0] == "config" && args[1] == "set" {
		return runConfigSet(args[2:], stdout, stderr)
	}

	if knownCommand(args) {
		fmt.Fprintf(stdout, "ptodo command scaffold accepted: %s\n", strings.Join(args, " "))
		return 0
	}

	fmt.Fprintf(stderr, "unknown command: %s\n\n", strings.Join(args, " "))
	printHelp(stderr)
	return 2
}

func runInit(stdout io.Writer, stderr io.Writer) int {
	workdir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "failed to determine working directory: %v\n", err)
		return 1
	}

	path, err := config.WriteLocal(workdir, config.DefaultLocalConfig())
	if err != nil {
		fmt.Fprintf(stderr, "failed to initialize local config: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "initialized local ptodo config at %s\n", path)
	fmt.Fprintln(stdout, "edit api_url, workspace_id, project_id, access_key, and access_secret before using server-backed commands")
	return 0
}

func runConfigList(stdout io.Writer, stderr io.Writer) int {
	workdir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "failed to determine working directory: %v\n", err)
		return 1
	}

	cfg, _, err := config.ReadLocal(workdir)
	if err != nil {
		fmt.Fprintf(stderr, "failed to read local config: %v\n", err)
		return 1
	}

	client, err := NewAPIClient(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "failed to initialize api client: %v\n", err)
		return 1
	}

	items, err := client.ListProjectConfig(context.Background(), cfg.ProjectID)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	if len(items) == 0 {
		fmt.Fprintln(stdout, "no project config entries")
		return 0
	}

	for _, item := range items {
		fmt.Fprintf(stdout, "%s=%s\n", item.Key, item.Value)
	}
	return 0
}

func runConfigSet(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) != 2 {
		fmt.Fprintln(stderr, "usage: ptodo config set <key> <value>")
		return 2
	}

	workdir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "failed to determine working directory: %v\n", err)
		return 1
	}

	cfg, _, err := config.ReadLocal(workdir)
	if err != nil {
		fmt.Fprintf(stderr, "failed to read local config: %v\n", err)
		return 1
	}

	client, err := NewAPIClient(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "failed to initialize api client: %v\n", err)
		return 1
	}

	item, err := client.SetProjectConfig(context.Background(), cfg.ProjectID, args[0], args[1])
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "%s=%s\n", item.Key, item.Value)
	return 0
}

func knownCommand(args []string) bool {
	for _, group := range commandGroups {
		for _, command := range group.Commands {
			parts := strings.Fields(command)
			if len(args) >= len(parts) && equalPrefix(args, parts) {
				return true
			}
		}
	}
	return false
}

func equalPrefix(args []string, parts []string) bool {
	for i := range parts {
		if args[i] != parts[i] {
			return false
		}
	}
	return true
}

func printHelp(w io.Writer) {
	fmt.Fprintln(w, "ptodo - centralized Trekker-compatible task tracking")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  ptodo [--toon] <command> [flags]")
	fmt.Fprintln(w, "  ptodo init")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Command groups:")
	for _, group := range commandGroups {
		fmt.Fprintf(w, "  %-8s %s\n", group.Name, strings.Join(group.Commands, ", "))
	}
}
