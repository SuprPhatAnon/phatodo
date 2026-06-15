package cli

import (
	"fmt"
	"io"
	"strings"
)

var commandGroups = []struct {
	Name     string
	Commands []string
}{
	{"setup", []string{"init", "wipe -y"}},
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

	if knownCommand(args) {
		fmt.Fprintf(stdout, "trakkr command scaffold accepted: %s\n", strings.Join(args, " "))
		return 0
	}

	fmt.Fprintf(stderr, "unknown command: %s\n\n", strings.Join(args, " "))
	printHelp(stderr)
	return 2
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
	fmt.Fprintln(w, "trakkr - centralized Trekker-compatible task tracking")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  trakkr [--toon] <command> [flags]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Command groups:")
	for _, group := range commandGroups {
		fmt.Fprintf(w, "  %-8s %s\n", group.Name, strings.Join(group.Commands, ", "))
	}
}
