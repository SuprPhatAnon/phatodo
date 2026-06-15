package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/SuprPhatAnon/phatodo/internal/config"
	"github.com/SuprPhatAnon/phatodo/internal/domain"
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
	{"workflow", []string{"ready"}},
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
	if len(args) >= 2 && args[0] == "config" && args[1] == "get" {
		return runConfigGet(args[2:], stdout, stderr)
	}
	if len(args) >= 2 && args[0] == "config" && args[1] == "set" {
		return runConfigSet(args[2:], stdout, stderr)
	}
	if len(args) >= 2 && args[0] == "config" && args[1] == "unset" {
		return runConfigUnset(args[2:], stdout, stderr)
	}
	if len(args) >= 1 && args[0] == "ready" {
		return runReady(args[1:], stdout, stderr)
	}
	if len(args) >= 2 && args[0] == "task" && args[1] == "list" {
		return runTaskList(args[2:], stdout, stderr)
	}
	if len(args) >= 2 && args[0] == "task" && args[1] == "create" {
		return runTaskCreate(args[2:], stdout, stderr)
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

func runConfigGet(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: ptodo config get <key>")
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

	item, err := client.GetProjectConfig(context.Background(), cfg.ProjectID, args[0])
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "%s=%s\n", item.Key, item.Value)
	return 0
}

func runConfigUnset(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: ptodo config unset <key>")
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

	item, err := client.UnsetProjectConfig(context.Background(), cfg.ProjectID, args[0])
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "%s=%s\n", item.Key, item.Value)
	return 0
}

func runTaskCreate(args []string, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseTaskCreateArgs(args)
	if err != nil {
		fmt.Fprintln(stderr, err)
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

	priority := domain.Priority(opts.priority)
	req := domain.TaskCreateRequest{
		Title:              opts.title,
		IssuePrefix:        opts.issuePrefix,
		Description:        opts.description,
		Priority:           &priority,
		EpicID:             opts.epicID,
		Tags:               opts.tags,
		AssignedTo:         opts.assignedTo,
		AcceptanceCriteria: opts.acceptanceCriteria,
	}

	resp, err := client.CreateTask(context.Background(), cfg.ProjectID, req)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "id=%s\n", resp.ID)
	fmt.Fprintf(stdout, "issue_prefix=%s\n", resp.IssuePrefix)
	fmt.Fprintf(stdout, "title=%s\n", resp.Title)
	return 0
}

func runTaskList(args []string, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseTaskListArgs(args)
	if err != nil {
		fmt.Fprintln(stderr, err)
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

	resp, err := client.ListTasks(context.Background(), cfg.ProjectID, opts.status, opts.epicID)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	for _, item := range resp.Items {
		fmt.Fprintf(stdout, "id=%s title=%s status=%s priority=%d", item.ID, item.Title, item.Status, item.Priority)
		if item.EpicID != "" {
			fmt.Fprintf(stdout, " epic_id=%s", item.EpicID)
		}
		if item.ParentTaskID != "" {
			fmt.Fprintf(stdout, " parent_task_id=%s", item.ParentTaskID)
		}
		if len(item.Tags) > 0 {
			fmt.Fprintf(stdout, " tags=%s", strings.Join(item.Tags, ","))
		}
		fmt.Fprintln(stdout)
	}
	return 0
}

func runReady(args []string, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseReadyArgs(args)
	if err != nil {
		fmt.Fprintln(stderr, err)
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

	resp, err := client.ListReadyTasks(context.Background(), cfg.ProjectID, opts.epicID)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	for _, item := range resp.Items {
		fmt.Fprintf(stdout, "%s | P%d | %s", item.ID, item.Priority, item.Title)
		if item.EpicID != "" {
			fmt.Fprintf(stdout, " (%s)", item.EpicID)
		}
		if len(item.Tags) > 0 {
			fmt.Fprintf(stdout, " [%s]", strings.Join(item.Tags, ","))
		}
		fmt.Fprintln(stdout)
		for _, blocked := range item.Unblocks {
			fmt.Fprintf(stdout, "  -> unblocks %s | %s | P%d | %s", blocked.ID, blocked.Status, blocked.Priority, blocked.Title)
			if blocked.EpicID != "" {
				fmt.Fprintf(stdout, " (%s)", blocked.EpicID)
			}
			if len(blocked.Tags) > 0 {
				fmt.Fprintf(stdout, " [%s]", strings.Join(blocked.Tags, ","))
			}
			fmt.Fprintln(stdout)
		}
	}
	return 0
}

type taskCreateOptions struct {
	title              string
	issuePrefix        string
	description        string
	priority           int
	epicID             string
	tags               []string
	assignedTo         string
	acceptanceCriteria []string
}

func parseTaskCreateArgs(args []string) (taskCreateOptions, error) {
	fs := flag.NewFlagSet("task create", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var title string
	var issuePrefix string
	var description string
	var epicID string
	var tagsValue string
	var assignedTo string
	var criteriaJSON string
	priority := int(domain.PriorityMedium)

	fs.StringVar(&title, "t", "", "")
	fs.StringVar(&title, "title", "", "")
	fs.StringVar(&issuePrefix, "issue-prefix", "", "")
	fs.StringVar(&description, "d", "", "")
	fs.StringVar(&description, "description", "", "")
	fs.IntVar(&priority, "p", int(domain.PriorityMedium), "")
	fs.IntVar(&priority, "priority", int(domain.PriorityMedium), "")
	fs.StringVar(&epicID, "e", "", "")
	fs.StringVar(&epicID, "epic", "", "")
	fs.StringVar(&tagsValue, "tags", "", "")
	fs.StringVar(&assignedTo, "a", "", "")
	fs.StringVar(&assignedTo, "assigned-to", "", "")
	fs.StringVar(&criteriaJSON, "criteria-json", "", "")

	if err := fs.Parse(args); err != nil {
		return taskCreateOptions{}, fmt.Errorf("invalid task create flags: %w", err)
	}
	if fs.NArg() > 0 {
		return taskCreateOptions{}, fmt.Errorf("task create does not accept positional arguments")
	}
	if title == "" {
		return taskCreateOptions{}, fmt.Errorf("task create requires -t <title>")
	}
	if issuePrefix == "" {
		return taskCreateOptions{}, fmt.Errorf("task create requires --issue-prefix <prefix>")
	}

	tags := parseCSVList(tagsValue)
	criteria, err := parseJSONStringList(criteriaJSON)
	if err != nil {
		return taskCreateOptions{}, err
	}

	return taskCreateOptions{
		title:              title,
		issuePrefix:        issuePrefix,
		description:        description,
		priority:           priority,
		epicID:             epicID,
		tags:               tags,
		assignedTo:         assignedTo,
		acceptanceCriteria: criteria,
	}, nil
}

func parseCSVList(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			items = append(items, trimmed)
		}
	}
	return items
}

func parseJSONStringList(value string) ([]string, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	var items []string
	if err := json.Unmarshal([]byte(value), &items); err != nil {
		return nil, fmt.Errorf("invalid criteria json: %w", err)
	}
	return items, nil
}

func isAllowedTaskStatus(value string) bool {
	switch domain.Status(value) {
	case domain.StatusTodo, domain.StatusInProgress, domain.StatusCompleted, domain.StatusWontFix, domain.StatusArchived:
		return true
	default:
		return false
	}
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

type taskListOptions struct {
	status string
	epicID string
}

func parseTaskListArgs(args []string) (taskListOptions, error) {
	fs := flag.NewFlagSet("task list", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var status string
	var epicID string
	fs.StringVar(&status, "status", "", "")
	fs.StringVar(&epicID, "epic", "", "")

	if err := fs.Parse(args); err != nil {
		return taskListOptions{}, fmt.Errorf("invalid task list flags: %w", err)
	}
	if fs.NArg() > 0 {
		return taskListOptions{}, fmt.Errorf("task list does not accept positional arguments")
	}
	if status != "" && !isAllowedTaskStatus(status) {
		return taskListOptions{}, fmt.Errorf("task list status must be todo, in_progress, completed, wont_fix, or archived")
	}

	return taskListOptions{
		status: status,
		epicID: epicID,
	}, nil
}

type readyOptions struct {
	epicID string
}

func parseReadyArgs(args []string) (readyOptions, error) {
	fs := flag.NewFlagSet("ready", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var epicID string
	fs.StringVar(&epicID, "epic", "", "")

	if err := fs.Parse(args); err != nil {
		return readyOptions{}, fmt.Errorf("invalid ready flags: %w", err)
	}
	if fs.NArg() > 0 {
		return readyOptions{}, fmt.Errorf("ready does not accept positional arguments")
	}

	return readyOptions{epicID: epicID}, nil
}
