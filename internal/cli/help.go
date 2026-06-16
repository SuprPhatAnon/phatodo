package cli

import (
	"fmt"
	"io"
	"strings"
)

type commandSpec struct {
	path  []string
	usage string
}

var commandSpecs = []commandSpec{
	{path: []string{"init"}, usage: "ptodo init"},
	{path: []string{"quickstart"}, usage: "ptodo quickstart"},
	{path: []string{"wipe"}, usage: "ptodo wipe -y"},
	{path: []string{"admin", "init"}, usage: "ptodo admin init"},
	{path: []string{"admin", "bootstrap"}, usage: "ptodo admin bootstrap"},
	{path: []string{"epic", "create"}, usage: `ptodo epic create -t "Title" [-d "desc"] [-p 0-5] [-a <user-id>] [--criteria-json '["criterion 1","criterion 2"]']`},
	{path: []string{"epic", "list"}, usage: "ptodo epic list [--status <status>] [--limit <n>]"},
	{path: []string{"epic", "show"}, usage: "ptodo epic show <epic-id>"},
	{path: []string{"epic", "update"}, usage: `ptodo epic update <epic-id> [-t "Title"] [-d "desc"] [-p 0-5] [-s <status>] [-a <user-id>] [--criteria-json '["criterion 1","criterion 2"]'] [--summary "completion summary"] [--evidence-json '["link-or-note"]']`},
	{path: []string{"epic", "complete"}, usage: "ptodo epic complete <epic-id>"},
	{path: []string{"epic", "delete"}, usage: "ptodo epic delete <epic-id>"},
	{path: []string{"task", "create"}, usage: `ptodo task create -t "Title" --prefix <prefix> (or --issue-prefix <prefix>) [-d "desc"] [-p 0-5] [-e <epic-id>] [--tags "a,b"] [-a <user-id>] [--criteria-json '["criterion 1","criterion 2"]']`},
	{path: []string{"task", "list"}, usage: "ptodo task list [--status <status>] [--epic <epic-id>] [--limit <n>]"},
	{path: []string{"task", "show"}, usage: "ptodo task show <task-id>"},
	{path: []string{"task", "update"}, usage: `ptodo task update <task-id> [-t "Title"] [-d "desc"] [-p 0-5] [-s <status>] [--tags "a,b"] [-e <epic-id>] [--no-epic] [-a <user-id>] [--criteria-json '["criterion 1","criterion 2"]'] [--summary "completion summary"] [--evidence-json '["link-or-note"]']`},
	{path: []string{"task", "delete"}, usage: "ptodo task delete <task-id>"},
	{path: []string{"subtask", "create"}, usage: `ptodo subtask create <task-id> -t "Title" [-d "desc"] [-p 0-5] [-a <user-id>] [--criteria-json '["criterion 1","criterion 2"]']`},
	{path: []string{"subtask", "list"}, usage: "ptodo subtask list <task-id> [--limit <n>]"},
	{path: []string{"subtask", "update"}, usage: `ptodo subtask update <subtask-id> [-t "Title"] [-d "desc"] [-p 0-5] [-s <status>] [-a <user-id>] [--criteria-json '["criterion 1","criterion 2"]'] [--summary "completion summary"] [--evidence-json '["link-or-note"]']`},
	{path: []string{"subtask", "delete"}, usage: "ptodo subtask delete <subtask-id>"},
	{path: []string{"comment", "add"}, usage: `ptodo comment add <task-id> -a "agent" -c "content" [-k comment|analysis|summary|checkpoint|handoff]`},
	{path: []string{"comment", "list"}, usage: "ptodo comment list <task-id>"},
	{path: []string{"comment", "update"}, usage: `ptodo comment update <comment-id> -c "new content"`},
	{path: []string{"comment", "delete"}, usage: "ptodo comment delete <comment-id>"},
	{path: []string{"dep", "add"}, usage: "ptodo dep add <task-id> <depends-on-id>"},
	{path: []string{"dep", "remove"}, usage: "ptodo dep remove <task-id> <depends-on-id>"},
	{path: []string{"dep", "list"}, usage: "ptodo dep list <task-id>"},
	{path: []string{"lock", "acquire"}, usage: `ptodo lock acquire <entity-type> <entity-id> [--reason "why"] [--expires 1h]`},
	{path: []string{"lock", "release"}, usage: "ptodo lock release <lock-id>"},
	{path: []string{"lock", "list"}, usage: "ptodo lock list [--type epic,task,subtask] [--entity <entity-id>] [--active]"},
	{path: []string{"ready"}, usage: "ptodo ready [--epic <epic-id>] [--limit <n>]"},
	{path: []string{"config", "list"}, usage: "ptodo config list"},
	{path: []string{"config", "get"}, usage: "ptodo config get <key>"},
	{path: []string{"config", "set"}, usage: "ptodo config set <key> <value>"},
	{path: []string{"config", "unset"}, usage: "ptodo config unset <key>"},
	{path: []string{"search"}, usage: `ptodo search "query" [--type epic,task,subtask,comment] [--status <status>] [--limit <n>]`},
	{path: []string{"history"}, usage: "ptodo history [--entity <entity-id>] [--type task] [--action create,update,delete] [--since <date>] [--limit <n>]"},
	{path: []string{"list"}, usage: "ptodo list [--type epic,task,subtask] [--status <status>] [--priority 0,1] [--sort priority:asc,created:desc] [--limit <n>]"},
}

func printHelp(w io.Writer) {
	fmt.Fprintln(w, "ptodo - centralized Trekker-compatible task tracking")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  ptodo [--toon] <command> [args]")
	fmt.Fprintln(w, "  ptodo <command> --help")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Commands:")
	for _, spec := range commandSpecs {
		fmt.Fprintf(w, "  %s\n", spec.usage)
	}
}

func printCommandHelp(w io.Writer, args []string) {
	if len(args) == 0 {
		printHelp(w)
		return
	}

	prefix := commandHelpPrefix(args)
	if len(prefix) == 0 {
		printHelp(w)
		return
	}

	matches := commandSpecsWithPrefix(prefix)
	if len(matches) == 0 {
		printHelp(w)
		return
	}

	if len(matches) == 1 && len(matches[0].path) == len(prefix) {
		fmt.Fprintln(w, "Usage:")
		fmt.Fprintf(w, "  %s\n", matches[0].usage)
		if len(prefix) == 1 && prefix[0] == "quickstart" {
			fmt.Fprintln(w)
			fmt.Fprintln(w, "Show quick reference for AI agents")
		}
		return
	}

	fmt.Fprintf(w, "%s\n\nCommands:\n", strings.Join(append([]string{"ptodo"}, prefix...), " "))
	for _, spec := range matches {
		fmt.Fprintf(w, "  %s\n", spec.usage)
	}
}

func commandHelpPrefix(args []string) []string {
	longest := 0
	for _, spec := range commandSpecs {
		if len(args) < len(spec.path) {
			continue
		}
		if equalPrefix(args, spec.path) && len(spec.path) > longest {
			longest = len(spec.path)
		}
	}
	if longest > 0 {
		return append([]string(nil), commandSpecsWithExactLength(args, longest)...)
	}

	if len(args) == 0 {
		return nil
	}
	if hasCommandFamily(args[0]) {
		return []string{args[0]}
	}
	return nil
}

func commandSpecsWithExactLength(args []string, length int) []string {
	for _, spec := range commandSpecs {
		if len(spec.path) == length && equalPrefix(args, spec.path) {
			return spec.path
		}
	}
	return nil
}

func commandSpecsWithPrefix(prefix []string) []commandSpec {
	var matches []commandSpec
	for _, spec := range commandSpecs {
		if len(spec.path) < len(prefix) {
			continue
		}
		if equalPrefix(spec.path, prefix) {
			matches = append(matches, spec)
		}
	}
	return matches
}

func hasCommandFamily(prefix string) bool {
	for _, spec := range commandSpecs {
		if len(spec.path) > 0 && spec.path[0] == prefix {
			return true
		}
	}
	return false
}

func helpRequested(args []string) bool {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return true
		}
	}
	return false
}
