package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/SuprPhatAnon/phatodo/internal/config"
	"github.com/SuprPhatAnon/phatodo/internal/domain"
)

// Run is the CLI boundary. It currently validates the Trekker-compatible
// command shape while the API client and persistence behavior are built out.
func Run(args []string, stdout io.Writer, stderr io.Writer) int {
	setOutputMode(false)
	if len(args) == 0 {
		printHelp(stdout)
		return 0
	}
	if args[0] == "--toon" {
		setOutputMode(true)
		args = args[1:]
		if len(args) == 0 {
			fmt.Fprintln(stderr, "missing command after --toon")
			return 2
		}
	}
	if args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		printHelp(stdout)
		return 0
	}
	if helpRequested(args) {
		printCommandHelp(stdout, args)
		return 0
	}
	if len(args) == 0 {
		fmt.Fprintln(stderr, "missing command after --toon")
		return 2
	}

	if args[0] == "init" {
		return runInit(stdout, stderr)
	}
	if args[0] == "quickstart" {
		printQuickstart(stdout)
		return 0
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
	if len(args) >= 2 && args[0] == "epic" && args[1] == "create" {
		return runEpicCreate(args[2:], stdout, stderr)
	}
	if len(args) >= 2 && args[0] == "epic" && args[1] == "list" {
		return runEpicList(args[2:], stdout, stderr)
	}
	if len(args) >= 2 && args[0] == "epic" && args[1] == "show" {
		return runEpicShow(args[2:], stdout, stderr)
	}
	if len(args) >= 2 && args[0] == "epic" && args[1] == "update" {
		return runEpicUpdate(args[2:], stdout, stderr)
	}
	if len(args) >= 2 && args[0] == "epic" && args[1] == "complete" {
		return runEpicComplete(args[2:], stdout, stderr)
	}
	if len(args) >= 2 && args[0] == "epic" && args[1] == "delete" {
		return runEpicDelete(args[2:], stdout, stderr)
	}
	if len(args) >= 1 && args[0] == "ready" {
		return runReady(args[1:], stdout, stderr)
	}
	if len(args) >= 2 && args[0] == "subtask" && args[1] == "create" {
		return runSubtaskCreate(args[2:], stdout, stderr)
	}
	if len(args) >= 2 && args[0] == "subtask" && args[1] == "list" {
		return runSubtaskList(args[2:], stdout, stderr)
	}
	if len(args) >= 2 && args[0] == "subtask" && args[1] == "update" {
		return runSubtaskUpdate(args[2:], stdout, stderr)
	}
	if len(args) >= 2 && args[0] == "subtask" && args[1] == "delete" {
		return runSubtaskDelete(args[2:], stdout, stderr)
	}
	if len(args) >= 2 && args[0] == "comment" && args[1] == "add" {
		return runCommentAdd(args[2:], stdout, stderr)
	}
	if len(args) >= 2 && args[0] == "comment" && args[1] == "list" {
		return runCommentList(args[2:], stdout, stderr)
	}
	if len(args) >= 2 && args[0] == "comment" && args[1] == "update" {
		return runCommentUpdate(args[2:], stdout, stderr)
	}
	if len(args) >= 2 && args[0] == "comment" && args[1] == "delete" {
		return runCommentDelete(args[2:], stdout, stderr)
	}
	if len(args) >= 2 && args[0] == "dep" && args[1] == "add" {
		return runDepAdd(args[2:], stdout, stderr)
	}
	if len(args) >= 2 && args[0] == "dep" && args[1] == "remove" {
		return runDepRemove(args[2:], stdout, stderr)
	}
	if len(args) >= 2 && args[0] == "dep" && args[1] == "list" {
		return runDepList(args[2:], stdout, stderr)
	}
	if len(args) >= 2 && args[0] == "lock" && args[1] == "acquire" {
		return runLockAcquire(args[2:], stdout, stderr)
	}
	if len(args) >= 2 && args[0] == "lock" && args[1] == "release" {
		return runLockRelease(args[2:], stdout, stderr)
	}
	if len(args) >= 2 && args[0] == "lock" && args[1] == "list" {
		return runLockList(args[2:], stdout, stderr)
	}
	if len(args) >= 1 && args[0] == "search" {
		return runSearch(args[1:], stdout, stderr)
	}
	if len(args) >= 1 && args[0] == "history" {
		return runHistory(args[1:], stdout, stderr)
	}
	if len(args) >= 1 && args[0] == "list" {
		return runList(args[1:], stdout, stderr)
	}
	if len(args) >= 2 && args[0] == "task" && args[1] == "show" {
		return runTaskShow(args[2:], stdout, stderr)
	}
	if len(args) >= 2 && args[0] == "task" && args[1] == "update" {
		return runTaskUpdate(args[2:], stdout, stderr)
	}
	if len(args) >= 2 && args[0] == "task" && args[1] == "delete" {
		return runTaskDelete(args[2:], stdout, stderr)
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

	writeTOONListItemStart(stdout, 0, "config_path", path)
	writeTOONField(stdout, 1, "api_url", config.DefaultLocalConfig().APIURL)
	writeTOONField(stdout, 1, "workspace_id", config.DefaultLocalConfig().WorkspaceID)
	writeTOONField(stdout, 1, "project_id", config.DefaultLocalConfig().ProjectID)
	writeTOONField(stdout, 1, "access_key", config.DefaultLocalConfig().AccessKey)
	writeTOONField(stdout, 1, "access_secret", config.DefaultLocalConfig().AccessSecret)
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

	client, err := newAPIClient(cfg)
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
		return 0
	}

	for _, item := range items {
		writeProjectConfigItem(stdout, item)
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

	client, err := newAPIClient(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "failed to initialize api client: %v\n", err)
		return 1
	}

	item, err := client.SetProjectConfig(context.Background(), cfg.ProjectID, args[0], args[1])
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	writeProjectConfigItem(stdout, item)
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

	client, err := newAPIClient(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "failed to initialize api client: %v\n", err)
		return 1
	}

	item, err := client.GetProjectConfig(context.Background(), cfg.ProjectID, args[0])
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	writeProjectConfigItem(stdout, item)
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

	client, err := newAPIClient(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "failed to initialize api client: %v\n", err)
		return 1
	}

	item, err := client.UnsetProjectConfig(context.Background(), cfg.ProjectID, args[0])
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	writeProjectConfigItem(stdout, item)
	return 0
}

type epicCreateOptions struct {
	title              string
	description        string
	priority           int
	assignedTo         string
	acceptanceCriteria []string
}

func parseEpicCreateArgs(args []string) (epicCreateOptions, error) {
	fs := flag.NewFlagSet("epic create", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var title string
	var description string
	var assignedTo string
	var criteriaJSON string
	priority := int(domain.PriorityMedium)

	fs.StringVar(&title, "t", "", "")
	fs.StringVar(&title, "title", "", "")
	fs.StringVar(&description, "d", "", "")
	fs.StringVar(&description, "description", "", "")
	fs.IntVar(&priority, "p", int(domain.PriorityMedium), "")
	fs.IntVar(&priority, "priority", int(domain.PriorityMedium), "")
	fs.StringVar(&assignedTo, "a", "", "")
	fs.StringVar(&assignedTo, "assigned-to", "", "")
	fs.StringVar(&criteriaJSON, "criteria-json", "", "")

	if err := fs.Parse(args); err != nil {
		return epicCreateOptions{}, fmt.Errorf("invalid epic create flags: %w", err)
	}
	if fs.NArg() > 0 {
		return epicCreateOptions{}, fmt.Errorf("epic create does not accept positional arguments")
	}
	if title == "" {
		return epicCreateOptions{}, fmt.Errorf("epic create requires -t <title>")
	}

	criteria, err := parseJSONStringList(criteriaJSON)
	if err != nil {
		return epicCreateOptions{}, err
	}

	return epicCreateOptions{
		title:              title,
		description:        description,
		priority:           priority,
		assignedTo:         assignedTo,
		acceptanceCriteria: criteria,
	}, nil
}

type epicUpdateOptions struct {
	epicID          string
	title           string
	description     string
	priority        int
	status          string
	assignedTo      string
	criteriaJSON    string
	summary         string
	evidenceJSON    string
	hasTitle        bool
	hasDescription  bool
	hasPriority     bool
	hasStatus       bool
	hasAssignedTo   bool
	hasCriteriaJSON bool
	hasSummary      bool
	hasEvidenceJSON bool
}

func parseEpicUpdateArgs(args []string) (epicUpdateOptions, error) {
	fs := flag.NewFlagSet("epic update", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var opts epicUpdateOptions
	fs.StringVar(&opts.title, "t", "", "")
	fs.StringVar(&opts.title, "title", "", "")
	fs.StringVar(&opts.description, "d", "", "")
	fs.StringVar(&opts.description, "description", "", "")
	fs.IntVar(&opts.priority, "p", int(domain.PriorityMedium), "")
	fs.IntVar(&opts.priority, "priority", int(domain.PriorityMedium), "")
	fs.StringVar(&opts.status, "s", "", "")
	fs.StringVar(&opts.status, "status", "", "")
	fs.StringVar(&opts.assignedTo, "a", "", "")
	fs.StringVar(&opts.assignedTo, "assigned-to", "", "")
	fs.StringVar(&opts.criteriaJSON, "criteria-json", "", "")
	fs.StringVar(&opts.summary, "summary", "", "")
	fs.StringVar(&opts.evidenceJSON, "evidence-json", "", "")

	if len(args) == 0 {
		return epicUpdateOptions{}, fmt.Errorf("epic update requires <epic-id>")
	}
	opts.epicID = args[0]

	if err := fs.Parse(args[1:]); err != nil {
		return epicUpdateOptions{}, fmt.Errorf("invalid epic update flags: %w", err)
	}
	if fs.NArg() > 0 {
		return epicUpdateOptions{}, fmt.Errorf("epic update does not accept positional arguments after <epic-id>")
	}

	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "t", "title":
			opts.hasTitle = true
		case "d", "description":
			opts.hasDescription = true
		case "p", "priority":
			opts.hasPriority = true
		case "s", "status":
			opts.hasStatus = true
		case "a", "assigned-to":
			opts.hasAssignedTo = true
		case "criteria-json":
			opts.hasCriteriaJSON = true
		case "summary":
			opts.hasSummary = true
		case "evidence-json":
			opts.hasEvidenceJSON = true
		}
	})

	if !opts.hasTitle && !opts.hasDescription && !opts.hasPriority && !opts.hasStatus && !opts.hasAssignedTo && !opts.hasCriteriaJSON && !opts.hasSummary && !opts.hasEvidenceJSON {
		return epicUpdateOptions{}, fmt.Errorf("epic update requires at least one change flag")
	}
	if opts.hasStatus && !isAllowedEpicStatus(opts.status) {
		return epicUpdateOptions{}, fmt.Errorf("epic update status must be todo, in_progress, completed, or archived")
	}

	return opts, nil
}

type epicListOptions struct {
	status string
	limit  int
}

func parseEpicListArgs(args []string) (epicListOptions, error) {
	fs := flag.NewFlagSet("epic list", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var status string
	var limit int
	fs.StringVar(&status, "status", "", "")
	fs.IntVar(&limit, "limit", 20, "")

	if err := fs.Parse(args); err != nil {
		return epicListOptions{}, fmt.Errorf("invalid epic list flags: %w", err)
	}
	if fs.NArg() > 0 {
		return epicListOptions{}, fmt.Errorf("epic list does not accept positional arguments")
	}
	if status != "" && !isAllowedEpicStatus(status) {
		return epicListOptions{}, fmt.Errorf("epic list status must be todo, in_progress, completed, or archived")
	}

	return epicListOptions{status: status, limit: limit}, nil
}

func runEpicCreate(args []string, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseEpicCreateArgs(args)
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

	client, err := newAPIClient(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "failed to initialize api client: %v\n", err)
		return 1
	}

	priority := domain.Priority(opts.priority)
	req := domain.EpicCreateRequest{
		Title:              opts.title,
		Description:        opts.description,
		Priority:           &priority,
		AssignedTo:         opts.assignedTo,
		AcceptanceCriteria: opts.acceptanceCriteria,
	}

	resp, err := client.CreateEpic(context.Background(), cfg.ProjectID, req)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	writeEpic(stdout, 0, resp)
	return 0
}

func runEpicList(args []string, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseEpicListArgs(args)
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

	client, err := newAPIClient(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "failed to initialize api client: %v\n", err)
		return 1
	}

	resp, err := client.ListEpics(context.Background(), cfg.ProjectID, opts.status, opts.limit)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	writeTOONArrayHeader(stdout, 0, "epics", len(resp.Items))
	for _, item := range resp.Items {
		writeEpic(stdout, 1, item)
	}
	return 0
}

func runEpicShow(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: ptodo epic show <epic-id>")
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

	client, err := newAPIClient(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "failed to initialize api client: %v\n", err)
		return 1
	}

	resp, err := client.GetEpic(context.Background(), cfg.ProjectID, args[0])
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	writeEpic(stdout, 0, resp)
	return 0
}

func runEpicUpdate(args []string, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseEpicUpdateArgs(args)
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

	client, err := newAPIClient(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "failed to initialize api client: %v\n", err)
		return 1
	}

	req := domain.EpicUpdateRequest{}
	if opts.hasTitle {
		req.Title = &opts.title
	}
	if opts.hasDescription {
		req.Description = &opts.description
	}
	if opts.hasPriority {
		priority := domain.Priority(opts.priority)
		req.Priority = &priority
	}
	if opts.hasStatus {
		status := domain.Status(opts.status)
		req.Status = &status
	}
	if opts.hasAssignedTo {
		req.AssignedTo = &opts.assignedTo
	}
	if opts.hasCriteriaJSON {
		criteria, err := parseJSONStringList(opts.criteriaJSON)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 2
		}
		req.AcceptanceCriteria = &criteria
	}
	if opts.hasSummary {
		req.CompletionSummary = &opts.summary
	}
	if opts.hasEvidenceJSON {
		evidence, err := parseJSONStringList(opts.evidenceJSON)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 2
		}
		req.CompletionEvidence = &evidence
	}

	resp, err := client.UpdateEpic(context.Background(), cfg.ProjectID, opts.epicID, req)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	writeEpic(stdout, 0, resp)
	return 0
}

func runEpicComplete(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: ptodo epic complete <epic-id>")
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

	client, err := newAPIClient(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "failed to initialize api client: %v\n", err)
		return 1
	}

	resp, err := client.CompleteEpic(context.Background(), cfg.ProjectID, args[0])
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	writeEpic(stdout, 0, resp)
	return 0
}

func runEpicDelete(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: ptodo epic delete <epic-id>")
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

	client, err := newAPIClient(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "failed to initialize api client: %v\n", err)
		return 1
	}

	resp, err := client.DeleteEpic(context.Background(), cfg.ProjectID, args[0])
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	writeEpic(stdout, 0, resp)
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

	client, err := newAPIClient(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "failed to initialize api client: %v\n", err)
		return 1
	}

	priority := domain.Priority(opts.priority)
	req := domain.TaskCreateRequest{
		Title:              opts.title,
		Kind:               domain.TaskKind(opts.kind),
		IssuePrefix:        opts.issuePrefix,
		Description:        opts.description,
		Priority:           &priority,
		EpicID:             opts.epicID,
		RootCauseAnalysis:  opts.rootCauseAnalysis,
		Tags:               opts.tags,
		PlannedFiles:       opts.plannedFiles,
		AssignedTo:         opts.assignedTo,
		AcceptanceCriteria: opts.acceptanceCriteria,
	}

	resp, err := client.CreateTask(context.Background(), cfg.ProjectID, req)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	writeTaskCreateResponse(stdout, resp)
	return 0
}

func runSubtaskCreate(args []string, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseSubtaskCreateArgs(args)
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

	client, err := newAPIClient(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "failed to initialize api client: %v\n", err)
		return 1
	}

	priority := domain.Priority(opts.priority)
	req := domain.TaskCreateRequest{
		Title:              opts.title,
		Kind:               domain.TaskKind(opts.kind),
		Description:        opts.description,
		Priority:           &priority,
		AssignedTo:         opts.assignedTo,
		RootCauseAnalysis:  opts.rootCauseAnalysis,
		PlannedFiles:       opts.plannedFiles,
		AcceptanceCriteria: opts.acceptanceCriteria,
	}

	resp, err := client.CreateSubtask(context.Background(), cfg.ProjectID, opts.taskID, req)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	writeTaskCreateResponse(stdout, resp)
	return 0
}

func runTaskShow(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: ptodo task show <task-id>")
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

	client, err := newAPIClient(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "failed to initialize api client: %v\n", err)
		return 1
	}

	resp, err := client.GetTask(context.Background(), cfg.ProjectID, args[0])
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	writeTaskDetail(stdout, resp)
	return 0
}

func runSubtaskList(args []string, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseSubtaskListArgs(args)
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

	client, err := newAPIClient(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "failed to initialize api client: %v\n", err)
		return 1
	}

	resp, err := client.ListSubtasks(context.Background(), cfg.ProjectID, opts.taskID, opts.limit)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	writeTOONArrayHeader(stdout, 0, "subtasks", len(resp.Items))
	for _, item := range resp.Items {
		writeTaskListItem(stdout, 1, item)
	}
	return 0
}

func runTaskUpdate(args []string, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseTaskUpdateArgs(args)
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

	client, err := newAPIClient(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "failed to initialize api client: %v\n", err)
		return 1
	}

	req := domain.TaskUpdateRequest{}
	if opts.hasTitle {
		req.Title = &opts.title
	}
	if opts.hasDescription {
		req.Description = &opts.description
	}
	if opts.hasPriority {
		priority := domain.Priority(opts.priority)
		req.Priority = &priority
	}
	if opts.hasStatus {
		status := domain.Status(opts.status)
		req.Status = &status
	}
	if opts.hasKind {
		kind := domain.TaskKind(opts.kind)
		req.Kind = &kind
	}
	if opts.hasTags {
		tags := parseCSVList(opts.tagsValue)
		req.Tags = &tags
	}
	if opts.noEpic {
		req.NoEpic = true
	} else if opts.hasEpicID {
		req.EpicID = &opts.epicID
	}
	if opts.hasAssignedTo {
		req.AssignedTo = &opts.assignedTo
	}
	if opts.hasRootCause {
		req.RootCauseAnalysis = &opts.rootCause
	}
	if opts.hasChangedFilesJSON {
		changedFiles, err := parseJSONStringList(opts.changedFilesJSON)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 2
		}
		req.ChangedFiles = &changedFiles
	}
	if opts.hasCriteriaJSON {
		criteria, err := parseJSONStringList(opts.criteriaJSON)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 2
		}
		req.AcceptanceCriteria = &criteria
	}
	if opts.hasSummary {
		req.CompletionSummary = &opts.summary
	}
	if opts.hasEvidenceJSON {
		evidence, err := parseJSONStringList(opts.evidenceJSON)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 2
		}
		req.CompletionEvidence = &evidence
	}

	resp, err := client.UpdateTask(context.Background(), cfg.ProjectID, opts.taskID, req)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	writeTaskDetail(stdout, resp)
	return 0
}

func runSubtaskUpdate(args []string, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseSubtaskUpdateArgs(args)
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

	client, err := newAPIClient(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "failed to initialize api client: %v\n", err)
		return 1
	}

	req := domain.TaskUpdateRequest{}
	if opts.hasTitle {
		req.Title = &opts.title
	}
	if opts.hasDescription {
		req.Description = &opts.description
	}
	if opts.hasPriority {
		priority := domain.Priority(opts.priority)
		req.Priority = &priority
	}
	if opts.hasStatus {
		status := domain.Status(opts.status)
		req.Status = &status
	}
	if opts.hasAssignedTo {
		req.AssignedTo = &opts.assignedTo
	}
	if opts.hasChangedFilesJSON {
		changedFiles, err := parseJSONStringList(opts.changedFilesJSON)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 2
		}
		req.ChangedFiles = &changedFiles
	}
	if opts.hasCriteriaJSON {
		criteria, err := parseJSONStringList(opts.criteriaJSON)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 2
		}
		req.AcceptanceCriteria = &criteria
	}
	if opts.hasSummary {
		req.CompletionSummary = &opts.summary
	}
	if opts.hasEvidenceJSON {
		evidence, err := parseJSONStringList(opts.evidenceJSON)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 2
		}
		req.CompletionEvidence = &evidence
	}

	resp, err := client.UpdateTask(context.Background(), cfg.ProjectID, opts.taskID, req)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	writeTaskDetail(stdout, resp)
	return 0
}

func runTaskDelete(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: ptodo task delete <task-id>")
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

	client, err := newAPIClient(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "failed to initialize api client: %v\n", err)
		return 1
	}

	resp, err := client.DeleteTask(context.Background(), cfg.ProjectID, args[0])
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	writeTaskDetail(stdout, resp)
	return 0
}

func runSubtaskDelete(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: ptodo subtask delete <subtask-id>")
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

	client, err := newAPIClient(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "failed to initialize api client: %v\n", err)
		return 1
	}

	resp, err := client.DeleteTask(context.Background(), cfg.ProjectID, args[0])
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	writeTaskDetail(stdout, resp)
	return 0
}

func runCommentAdd(args []string, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseCommentAddArgs(args)
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

	client, err := newAPIClient(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "failed to initialize api client: %v\n", err)
		return 1
	}

	req := domain.CommentCreateRequest{
		Author:  opts.author,
		Kind:    opts.kind,
		Content: opts.content,
	}

	resp, err := client.AddComment(context.Background(), cfg.ProjectID, opts.taskID, req)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	writeComment(stdout, 0, resp)
	return 0
}

func runCommentList(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: ptodo comment list <task-id>")
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

	client, err := newAPIClient(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "failed to initialize api client: %v\n", err)
		return 1
	}

	resp, err := client.ListComments(context.Background(), cfg.ProjectID, args[0])
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	writeTOONArrayHeader(stdout, 0, "comments", len(resp.Items))
	for _, item := range resp.Items {
		writeComment(stdout, 1, item)
	}
	return 0
}

func runCommentUpdate(args []string, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseCommentUpdateArgs(args)
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

	client, err := newAPIClient(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "failed to initialize api client: %v\n", err)
		return 1
	}

	resp, err := client.UpdateComment(context.Background(), cfg.ProjectID, opts.commentID, domain.CommentUpdateRequest{Content: opts.content})
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	writeComment(stdout, 0, resp)
	return 0
}

func runCommentDelete(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: ptodo comment delete <comment-id>")
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

	client, err := newAPIClient(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "failed to initialize api client: %v\n", err)
		return 1
	}

	resp, err := client.DeleteComment(context.Background(), cfg.ProjectID, args[0])
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	writeComment(stdout, 0, resp)
	return 0
}

func runDepAdd(args []string, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseDependencyPairArgs("dep add", args)
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

	client, err := newAPIClient(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "failed to initialize api client: %v\n", err)
		return 1
	}

	resp, err := client.AddDependency(context.Background(), cfg.ProjectID, opts.taskID, opts.dependsOnID)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	writeDependency(stdout, 0, resp)
	return 0
}

func runDepRemove(args []string, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseDependencyPairArgs("dep remove", args)
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

	client, err := newAPIClient(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "failed to initialize api client: %v\n", err)
		return 1
	}

	resp, err := client.RemoveDependency(context.Background(), cfg.ProjectID, opts.taskID, opts.dependsOnID)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	writeDependency(stdout, 0, resp)
	return 0
}

func runDepList(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: ptodo dep list <task-id>")
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

	client, err := newAPIClient(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "failed to initialize api client: %v\n", err)
		return 1
	}

	resp, err := client.ListDependencies(context.Background(), cfg.ProjectID, args[0])
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	writeTOONArrayHeader(stdout, 0, "dependencies", len(resp.Items))
	for _, item := range resp.Items {
		writeDependency(stdout, 1, item)
	}
	return 0
}

type lockAcquireOptions struct {
	entityType string
	entityID   string
	reason     string
	expires    string
}

func parseLockAcquireArgs(args []string) (lockAcquireOptions, error) {
	fs := flag.NewFlagSet("lock acquire", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var opts lockAcquireOptions
	fs.StringVar(&opts.reason, "reason", "", "")
	fs.StringVar(&opts.expires, "expires", "", "")

	if len(args) < 2 {
		return lockAcquireOptions{}, fmt.Errorf("lock acquire requires <entity-type> <entity-id>")
	}
	opts.entityType = args[0]
	opts.entityID = args[1]

	if err := fs.Parse(args[2:]); err != nil {
		return lockAcquireOptions{}, fmt.Errorf("invalid lock acquire flags: %w", err)
	}
	if fs.NArg() > 0 {
		return lockAcquireOptions{}, fmt.Errorf("lock acquire does not accept positional arguments after <entity-id>")
	}
	if !isAllowedLockEntityType(opts.entityType) {
		return lockAcquireOptions{}, fmt.Errorf("lock acquire entity type must be epic, task, or subtask")
	}
	if opts.expires != "" {
		if _, err := time.ParseDuration(opts.expires); err != nil {
			return lockAcquireOptions{}, fmt.Errorf("invalid lock acquire expires duration: %w", err)
		}
	}
	return opts, nil
}

type lockListOptions struct {
	entityTypes string
	entityID    string
	active      bool
}

func parseLockListArgs(args []string) (lockListOptions, error) {
	fs := flag.NewFlagSet("lock list", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var opts lockListOptions
	fs.StringVar(&opts.entityTypes, "type", "", "")
	fs.StringVar(&opts.entityID, "entity", "", "")
	fs.BoolVar(&opts.active, "active", false, "")

	if err := fs.Parse(args); err != nil {
		return lockListOptions{}, fmt.Errorf("invalid lock list flags: %w", err)
	}
	if fs.NArg() > 0 {
		return lockListOptions{}, fmt.Errorf("lock list does not accept positional arguments")
	}
	if opts.entityTypes != "" {
		for _, entityType := range strings.Split(opts.entityTypes, ",") {
			if !isAllowedLockEntityType(strings.TrimSpace(entityType)) {
				return lockListOptions{}, fmt.Errorf("lock list type must be epic, task, or subtask")
			}
		}
	}
	return opts, nil
}

func runLockAcquire(args []string, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseLockAcquireArgs(args)
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

	client, err := newAPIClient(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "failed to initialize api client: %v\n", err)
		return 1
	}

	req := domain.LockAcquireRequest{
		EntityType: opts.entityType,
		EntityID:   opts.entityID,
		Reason:     opts.reason,
		TTL:        opts.expires,
	}

	resp, err := client.AcquireLock(context.Background(), cfg.ProjectID, req)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	writeLock(stdout, 0, resp)
	return 0
}

func runLockRelease(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: ptodo lock release <lock-id>")
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

	client, err := newAPIClient(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "failed to initialize api client: %v\n", err)
		return 1
	}

	resp, err := client.ReleaseLock(context.Background(), cfg.ProjectID, args[0])
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	writeLock(stdout, 0, resp)
	return 0
}

func runLockList(args []string, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseLockListArgs(args)
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

	client, err := newAPIClient(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "failed to initialize api client: %v\n", err)
		return 1
	}

	resp, err := client.ListLocks(context.Background(), cfg.ProjectID, opts.entityTypes, opts.entityID, opts.active)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	writeTOONArrayHeader(stdout, 0, "locks", len(resp.Items))
	for _, item := range resp.Items {
		writeLock(stdout, 1, item)
	}
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

	client, err := newAPIClient(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "failed to initialize api client: %v\n", err)
		return 1
	}

	resp, err := client.ListTasks(context.Background(), cfg.ProjectID, opts.status, opts.epicID, opts.limit)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	writeTOONArrayHeader(stdout, 0, "tasks", len(resp.Items))
	for _, item := range resp.Items {
		writeTaskListItem(stdout, 1, item)
	}
	return 0
}

type taskUpdateOptions struct {
	taskID              string
	title               string
	description         string
	priority            int
	status              string
	kind                string
	rootCause           string
	tagsValue           string
	epicID              string
	noEpic              bool
	assignedTo          string
	criteriaJSON        string
	changedFilesJSON    string
	summary             string
	evidenceJSON        string
	hasTitle            bool
	hasDescription      bool
	hasPriority         bool
	hasStatus           bool
	hasKind             bool
	hasTags             bool
	hasEpicID           bool
	hasAssignedTo       bool
	hasRootCause        bool
	hasCriteriaJSON     bool
	hasChangedFilesJSON bool
	hasSummary          bool
	hasEvidenceJSON     bool
}

func parseTaskUpdateArgs(args []string) (taskUpdateOptions, error) {
	fs := flag.NewFlagSet("task update", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var opts taskUpdateOptions
	fs.StringVar(&opts.title, "t", "", "")
	fs.StringVar(&opts.title, "title", "", "")
	fs.StringVar(&opts.description, "d", "", "")
	fs.StringVar(&opts.description, "description", "", "")
	fs.IntVar(&opts.priority, "p", int(domain.PriorityMedium), "")
	fs.IntVar(&opts.priority, "priority", int(domain.PriorityMedium), "")
	fs.StringVar(&opts.status, "s", "", "")
	fs.StringVar(&opts.status, "status", "", "")
	fs.StringVar(&opts.kind, "k", "", "")
	fs.StringVar(&opts.kind, "kind", "", "")
	fs.StringVar(&opts.tagsValue, "tags", "", "")
	fs.StringVar(&opts.epicID, "e", "", "")
	fs.StringVar(&opts.epicID, "epic", "", "")
	fs.BoolVar(&opts.noEpic, "no-epic", false, "")
	fs.StringVar(&opts.assignedTo, "a", "", "")
	fs.StringVar(&opts.assignedTo, "assigned-to", "", "")
	fs.StringVar(&opts.rootCause, "root-cause", "", "")
	fs.StringVar(&opts.rootCause, "root-cause-analysis", "", "")
	fs.StringVar(&opts.changedFilesJSON, "changed-files-json", "", "")
	fs.StringVar(&opts.criteriaJSON, "criteria-json", "", "")
	fs.StringVar(&opts.summary, "summary", "", "")
	fs.StringVar(&opts.evidenceJSON, "evidence-json", "", "")

	if len(args) == 0 {
		return taskUpdateOptions{}, fmt.Errorf("task update requires <task-id>")
	}
	opts.taskID = args[0]

	if err := fs.Parse(args[1:]); err != nil {
		return taskUpdateOptions{}, fmt.Errorf("invalid task update flags: %w", err)
	}
	if fs.NArg() > 0 {
		return taskUpdateOptions{}, fmt.Errorf("task update does not accept positional arguments after <task-id>")
	}

	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "t", "title":
			opts.hasTitle = true
		case "d", "description":
			opts.hasDescription = true
		case "p", "priority":
			opts.hasPriority = true
		case "s", "status":
			opts.hasStatus = true
		case "k", "kind":
			opts.hasKind = true
		case "tags":
			opts.hasTags = true
		case "e", "epic":
			opts.hasEpicID = true
		case "a", "assigned-to":
			opts.hasAssignedTo = true
		case "root-cause", "root-cause-analysis":
			opts.hasRootCause = true
		case "changed-files-json":
			opts.hasChangedFilesJSON = true
		case "criteria-json":
			opts.hasCriteriaJSON = true
		case "summary":
			opts.hasSummary = true
		case "evidence-json":
			opts.hasEvidenceJSON = true
		}
	})

	if !opts.hasTitle && !opts.hasDescription && !opts.hasPriority && !opts.hasStatus && !opts.hasKind && !opts.hasTags && !opts.hasEpicID && !opts.noEpic && !opts.hasAssignedTo && !opts.hasRootCause && !opts.hasChangedFilesJSON && !opts.hasCriteriaJSON && !opts.hasSummary && !opts.hasEvidenceJSON {
		return taskUpdateOptions{}, fmt.Errorf("task update requires at least one change flag")
	}
	if opts.noEpic && opts.hasEpicID {
		return taskUpdateOptions{}, fmt.Errorf("task update cannot combine --no-epic with -e/--epic")
	}
	if opts.hasStatus && !isAllowedTaskStatus(opts.status) {
		return taskUpdateOptions{}, fmt.Errorf("task update status must be todo, in_progress, completed, wont_fix, or archived")
	}
	if opts.hasKind && !isAllowedTaskKind(opts.kind) {
		return taskUpdateOptions{}, fmt.Errorf("task update kind must be task, bug, feature, chore, or spike")
	}

	return opts, nil
}

type subtaskUpdateOptions struct {
	taskID              string
	title               string
	description         string
	priority            int
	status              string
	assignedTo          string
	criteriaJSON        string
	changedFilesJSON    string
	summary             string
	evidenceJSON        string
	hasTitle            bool
	hasDescription      bool
	hasPriority         bool
	hasStatus           bool
	hasAssignedTo       bool
	hasCriteriaJSON     bool
	hasChangedFilesJSON bool
	hasSummary          bool
	hasEvidenceJSON     bool
}

func parseSubtaskUpdateArgs(args []string) (subtaskUpdateOptions, error) {
	fs := flag.NewFlagSet("subtask update", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var opts subtaskUpdateOptions
	fs.StringVar(&opts.title, "t", "", "")
	fs.StringVar(&opts.title, "title", "", "")
	fs.StringVar(&opts.description, "d", "", "")
	fs.StringVar(&opts.description, "description", "", "")
	fs.IntVar(&opts.priority, "p", int(domain.PriorityMedium), "")
	fs.IntVar(&opts.priority, "priority", int(domain.PriorityMedium), "")
	fs.StringVar(&opts.status, "s", "", "")
	fs.StringVar(&opts.status, "status", "", "")
	fs.StringVar(&opts.assignedTo, "a", "", "")
	fs.StringVar(&opts.assignedTo, "assigned-to", "", "")
	fs.StringVar(&opts.changedFilesJSON, "changed-files-json", "", "")
	fs.StringVar(&opts.criteriaJSON, "criteria-json", "", "")
	fs.StringVar(&opts.summary, "summary", "", "")
	fs.StringVar(&opts.evidenceJSON, "evidence-json", "", "")

	if len(args) == 0 {
		return subtaskUpdateOptions{}, fmt.Errorf("subtask update requires <subtask-id>")
	}
	opts.taskID = args[0]

	if err := fs.Parse(args[1:]); err != nil {
		return subtaskUpdateOptions{}, fmt.Errorf("invalid subtask update flags: %w", err)
	}
	if fs.NArg() > 0 {
		return subtaskUpdateOptions{}, fmt.Errorf("subtask update does not accept positional arguments after <subtask-id>")
	}

	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "t", "title":
			opts.hasTitle = true
		case "d", "description":
			opts.hasDescription = true
		case "p", "priority":
			opts.hasPriority = true
		case "s", "status":
			opts.hasStatus = true
		case "a", "assigned-to":
			opts.hasAssignedTo = true
		case "changed-files-json":
			opts.hasChangedFilesJSON = true
		case "criteria-json":
			opts.hasCriteriaJSON = true
		case "summary":
			opts.hasSummary = true
		case "evidence-json":
			opts.hasEvidenceJSON = true
		}
	})

	if !opts.hasTitle && !opts.hasDescription && !opts.hasPriority && !opts.hasStatus && !opts.hasAssignedTo && !opts.hasChangedFilesJSON && !opts.hasCriteriaJSON && !opts.hasSummary && !opts.hasEvidenceJSON {
		return subtaskUpdateOptions{}, fmt.Errorf("subtask update requires at least one change flag")
	}
	if opts.hasStatus && !isAllowedTaskStatus(opts.status) {
		return subtaskUpdateOptions{}, fmt.Errorf("subtask update status must be todo, in_progress, completed, wont_fix, or archived")
	}

	return opts, nil
}

type commentAddOptions struct {
	taskID  string
	author  string
	kind    string
	content string
}

func parseCommentAddArgs(args []string) (commentAddOptions, error) {
	fs := flag.NewFlagSet("comment add", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var author string
	var kind string
	var content string
	fs.StringVar(&author, "a", "", "")
	fs.StringVar(&author, "author", "", "")
	fs.StringVar(&kind, "k", "comment", "")
	fs.StringVar(&kind, "kind", "comment", "")
	fs.StringVar(&content, "c", "", "")
	fs.StringVar(&content, "content", "", "")

	if len(args) == 0 {
		return commentAddOptions{}, fmt.Errorf("comment add requires <task-id>")
	}
	taskID := args[0]

	if err := fs.Parse(args[1:]); err != nil {
		return commentAddOptions{}, fmt.Errorf("invalid comment add flags: %w", err)
	}
	if fs.NArg() > 0 {
		return commentAddOptions{}, fmt.Errorf("comment add does not accept positional arguments after <task-id>")
	}
	if author == "" {
		return commentAddOptions{}, fmt.Errorf("comment add requires -a <author>")
	}
	if content == "" {
		return commentAddOptions{}, fmt.Errorf("comment add requires -c <content>")
	}
	if !isAllowedCommentKind(kind) {
		return commentAddOptions{}, fmt.Errorf("comment add kind must be comment, analysis, summary, checkpoint, or handoff")
	}

	return commentAddOptions{taskID: taskID, author: author, kind: kind, content: content}, nil
}

type commentUpdateOptions struct {
	commentID string
	content   string
}

func parseCommentUpdateArgs(args []string) (commentUpdateOptions, error) {
	fs := flag.NewFlagSet("comment update", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var content string
	fs.StringVar(&content, "c", "", "")
	fs.StringVar(&content, "content", "", "")

	if len(args) == 0 {
		return commentUpdateOptions{}, fmt.Errorf("comment update requires <comment-id>")
	}
	commentID := args[0]

	if err := fs.Parse(args[1:]); err != nil {
		return commentUpdateOptions{}, fmt.Errorf("invalid comment update flags: %w", err)
	}
	if fs.NArg() > 0 {
		return commentUpdateOptions{}, fmt.Errorf("comment update does not accept positional arguments after <comment-id>")
	}
	if content == "" {
		return commentUpdateOptions{}, fmt.Errorf("comment update requires -c <content>")
	}

	return commentUpdateOptions{commentID: commentID, content: content}, nil
}

type dependencyPairOptions struct {
	taskID      string
	dependsOnID string
}

func parseDependencyPairArgs(command string, args []string) (dependencyPairOptions, error) {
	if len(args) != 2 {
		return dependencyPairOptions{}, fmt.Errorf("usage: ptodo %s <task-id> <depends-on-id>", command)
	}
	return dependencyPairOptions{taskID: args[0], dependsOnID: args[1]}, nil
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

	client, err := newAPIClient(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "failed to initialize api client: %v\n", err)
		return 1
	}

	resp, err := client.ListReadyTasks(context.Background(), cfg.ProjectID, opts.epicID, opts.limit)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	if currentOutputMode == outputTOON {
		writeTOONArrayHeader(stdout, 0, "ready", len(resp.Items))
		for _, item := range resp.Items {
			writeReadyListItem(stdout, 1, item)
		}
		return 0
	}

	writeReadyHumanList(stdout, resp)
	return 0
}

type searchOptions struct {
	query      string
	entityType string
	status     string
	limit      int
}

func parseSearchArgs(args []string) (searchOptions, error) {
	fs := flag.NewFlagSet("search", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var opts searchOptions
	fs.StringVar(&opts.entityType, "type", "", "")
	fs.StringVar(&opts.status, "status", "", "")
	fs.IntVar(&opts.limit, "limit", 20, "")

	if len(args) == 0 {
		return searchOptions{}, fmt.Errorf("search requires <query>")
	}
	opts.query = args[0]

	if err := fs.Parse(args[1:]); err != nil {
		return searchOptions{}, fmt.Errorf("invalid search flags: %w", err)
	}
	if fs.NArg() > 0 {
		return searchOptions{}, fmt.Errorf("search does not accept positional arguments after <query>")
	}
	if strings.TrimSpace(opts.query) == "" {
		return searchOptions{}, fmt.Errorf("search requires <query>")
	}

	return opts, nil
}

type historyOptions struct {
	entityID   string
	entityType string
	action     string
	since      string
	limit      int
}

func parseHistoryArgs(args []string) (historyOptions, error) {
	fs := flag.NewFlagSet("history", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var opts historyOptions
	fs.StringVar(&opts.entityID, "entity", "", "")
	fs.StringVar(&opts.entityType, "type", "", "")
	fs.StringVar(&opts.action, "action", "", "")
	fs.StringVar(&opts.since, "since", "", "")
	fs.IntVar(&opts.limit, "limit", 50, "")

	if err := fs.Parse(args); err != nil {
		return historyOptions{}, fmt.Errorf("invalid history flags: %w", err)
	}
	if fs.NArg() > 0 {
		return historyOptions{}, fmt.Errorf("history does not accept positional arguments")
	}

	return opts, nil
}

type listOptions struct {
	entityType string
	status     string
	priority   string
	sortSpec   string
	limit      int
}

func parseListArgs(args []string) (listOptions, error) {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var opts listOptions
	fs.StringVar(&opts.entityType, "type", "", "")
	fs.StringVar(&opts.status, "status", "", "")
	fs.StringVar(&opts.priority, "priority", "", "")
	fs.StringVar(&opts.sortSpec, "sort", "", "")
	fs.IntVar(&opts.limit, "limit", 50, "")

	if err := fs.Parse(args); err != nil {
		return listOptions{}, fmt.Errorf("invalid list flags: %w", err)
	}
	if fs.NArg() > 0 {
		return listOptions{}, fmt.Errorf("list does not accept positional arguments")
	}

	return opts, nil
}

func runSearch(args []string, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseSearchArgs(args)
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

	client, err := newAPIClient(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "failed to initialize api client: %v\n", err)
		return 1
	}

	resp, err := client.Search(context.Background(), cfg.ProjectID, opts.query, opts.entityType, opts.status, opts.limit)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	writeTOONArrayHeader(stdout, 0, "search", len(resp.Items))
	for _, item := range resp.Items {
		writeSearchItem(stdout, 1, item)
	}
	return 0
}

func runHistory(args []string, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseHistoryArgs(args)
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

	client, err := newAPIClient(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "failed to initialize api client: %v\n", err)
		return 1
	}

	resp, err := client.History(context.Background(), cfg.ProjectID, opts.entityID, opts.entityType, opts.action, opts.since, opts.limit)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	writeTOONArrayHeader(stdout, 0, "history", len(resp.Items))
	for _, item := range resp.Items {
		writeHistoryEvent(stdout, 1, item)
	}
	return 0
}

func runList(args []string, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseListArgs(args)
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

	client, err := newAPIClient(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "failed to initialize api client: %v\n", err)
		return 1
	}

	resp, err := client.ListUnified(context.Background(), cfg.ProjectID, opts.entityType, opts.status, opts.priority, opts.sortSpec, opts.limit)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	writeTOONArrayHeader(stdout, 0, "list", len(resp.Items))
	for _, item := range resp.Items {
		writeUnifiedListItem(stdout, 1, item)
	}
	return 0
}

type taskCreateOptions struct {
	title              string
	issuePrefix        string
	description        string
	kind               string
	priority           int
	epicID             string
	tags               []string
	assignedTo         string
	rootCauseAnalysis  string
	plannedFiles       []string
	acceptanceCriteria []string
}

func parseTaskCreateArgs(args []string) (taskCreateOptions, error) {
	fs := flag.NewFlagSet("task create", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var title string
	var issuePrefixShort string
	var issuePrefixLong string
	var description string
	var epicID string
	var tagsValue string
	var assignedTo string
	var criteriaJSON string
	var plannedFilesJSON string
	var kind string
	var rootCause string
	priority := int(domain.PriorityMedium)

	fs.StringVar(&title, "t", "", "")
	fs.StringVar(&title, "title", "", "")
	fs.StringVar(&issuePrefixShort, "prefix", "", "")
	fs.StringVar(&issuePrefixLong, "issue-prefix", "", "")
	fs.StringVar(&description, "d", "", "")
	fs.StringVar(&description, "description", "", "")
	fs.StringVar(&kind, "k", string(domain.TaskKindTask), "")
	fs.StringVar(&kind, "kind", string(domain.TaskKindTask), "")
	fs.IntVar(&priority, "p", int(domain.PriorityMedium), "")
	fs.IntVar(&priority, "priority", int(domain.PriorityMedium), "")
	fs.StringVar(&epicID, "e", "", "")
	fs.StringVar(&epicID, "epic", "", "")
	fs.StringVar(&tagsValue, "tags", "", "")
	fs.StringVar(&assignedTo, "a", "", "")
	fs.StringVar(&assignedTo, "assigned-to", "", "")
	fs.StringVar(&plannedFilesJSON, "planned-files-json", "", "")
	fs.StringVar(&criteriaJSON, "criteria-json", "", "")
	fs.StringVar(&rootCause, "root-cause", "", "")
	fs.StringVar(&rootCause, "root-cause-analysis", "", "")

	if err := fs.Parse(args); err != nil {
		return taskCreateOptions{}, fmt.Errorf("invalid task create flags: %w", err)
	}
	if fs.NArg() > 0 {
		return taskCreateOptions{}, fmt.Errorf("task create does not accept positional arguments")
	}
	if title == "" {
		return taskCreateOptions{}, fmt.Errorf("task create requires -t <title>")
	}
	issuePrefix := firstNonEmpty(issuePrefixShort, issuePrefixLong)
	if issuePrefix == "" {
		return taskCreateOptions{}, fmt.Errorf("task create requires --prefix <prefix>")
	}
	if !isAllowedTaskKind(kind) {
		return taskCreateOptions{}, fmt.Errorf("task create kind must be task, bug, feature, chore, or spike")
	}
	if kind == string(domain.TaskKindBug) && strings.TrimSpace(rootCause) == "" {
		return taskCreateOptions{}, fmt.Errorf("task create requires --root-cause-analysis when kind is bug")
	}

	tags := parseCSVList(tagsValue)
	plannedFiles, err := parseJSONStringList(plannedFilesJSON)
	if err != nil {
		return taskCreateOptions{}, err
	}
	criteria, err := parseJSONStringList(criteriaJSON)
	if err != nil {
		return taskCreateOptions{}, err
	}

	return taskCreateOptions{
		title:              title,
		issuePrefix:        issuePrefix,
		description:        description,
		kind:               kind,
		priority:           priority,
		epicID:             epicID,
		tags:               tags,
		assignedTo:         assignedTo,
		rootCauseAnalysis:  rootCause,
		plannedFiles:       plannedFiles,
		acceptanceCriteria: criteria,
	}, nil
}

type subtaskCreateOptions struct {
	taskID             string
	title              string
	description        string
	kind               string
	priority           int
	rootCauseAnalysis  string
	assignedTo         string
	plannedFiles       []string
	acceptanceCriteria []string
}

func parseSubtaskCreateArgs(args []string) (subtaskCreateOptions, error) {
	fs := flag.NewFlagSet("subtask create", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var title string
	var description string
	var kind string
	var rootCause string
	var assignedTo string
	var plannedFilesJSON string
	var criteriaJSON string
	priority := int(domain.PriorityMedium)

	fs.StringVar(&title, "t", "", "")
	fs.StringVar(&title, "title", "", "")
	fs.StringVar(&description, "d", "", "")
	fs.StringVar(&description, "description", "", "")
	fs.StringVar(&kind, "k", string(domain.TaskKindTask), "")
	fs.StringVar(&kind, "kind", string(domain.TaskKindTask), "")
	fs.IntVar(&priority, "p", int(domain.PriorityMedium), "")
	fs.IntVar(&priority, "priority", int(domain.PriorityMedium), "")
	fs.StringVar(&assignedTo, "a", "", "")
	fs.StringVar(&assignedTo, "assigned-to", "", "")
	fs.StringVar(&plannedFilesJSON, "planned-files-json", "", "")
	fs.StringVar(&criteriaJSON, "criteria-json", "", "")
	fs.StringVar(&rootCause, "root-cause", "", "")
	fs.StringVar(&rootCause, "root-cause-analysis", "", "")

	if len(args) == 0 {
		return subtaskCreateOptions{}, fmt.Errorf("subtask create requires <task-id>")
	}
	taskID := args[0]
	if err := fs.Parse(args[1:]); err != nil {
		return subtaskCreateOptions{}, fmt.Errorf("invalid subtask create flags: %w", err)
	}
	if fs.NArg() > 0 {
		return subtaskCreateOptions{}, fmt.Errorf("subtask create does not accept positional arguments after <task-id>")
	}
	if title == "" {
		return subtaskCreateOptions{}, fmt.Errorf("subtask create requires -t <title>")
	}
	if !isAllowedTaskKind(kind) {
		return subtaskCreateOptions{}, fmt.Errorf("subtask create kind must be task, bug, feature, chore, or spike")
	}
	if kind == string(domain.TaskKindBug) && strings.TrimSpace(rootCause) == "" {
		return subtaskCreateOptions{}, fmt.Errorf("subtask create requires --root-cause-analysis when kind is bug")
	}

	plannedFiles, err := parseJSONStringList(plannedFilesJSON)
	if err != nil {
		return subtaskCreateOptions{}, err
	}
	criteria, err := parseJSONStringList(criteriaJSON)
	if err != nil {
		return subtaskCreateOptions{}, err
	}

	return subtaskCreateOptions{
		taskID:             taskID,
		title:              title,
		description:        description,
		kind:               kind,
		priority:           priority,
		assignedTo:         assignedTo,
		rootCauseAnalysis:  rootCause,
		plannedFiles:       plannedFiles,
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

func isAllowedTaskKind(value string) bool {
	switch domain.TaskKind(strings.TrimSpace(value)) {
	case domain.TaskKindTask, domain.TaskKindBug, domain.TaskKindFeature, domain.TaskKindChore, domain.TaskKindSpike:
		return true
	default:
		return false
	}
}

func isAllowedEpicStatus(value string) bool {
	switch domain.Status(value) {
	case domain.StatusTodo, domain.StatusInProgress, domain.StatusCompleted, domain.StatusArchived:
		return true
	default:
		return false
	}
}

func isAllowedLockEntityType(value string) bool {
	switch strings.TrimSpace(value) {
	case "epic", "task", "subtask":
		return true
	default:
		return false
	}
}

func isAllowedCommentKind(value string) bool {
	switch value {
	case "comment", "analysis", "summary", "checkpoint", "handoff":
		return true
	default:
		return false
	}
}

func knownCommand(args []string) bool {
	for _, spec := range commandSpecs {
		if len(args) >= len(spec.path) && equalPrefix(args, spec.path) {
			return true
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

type taskListOptions struct {
	status string
	epicID string
	limit  int
}

func parseTaskListArgs(args []string) (taskListOptions, error) {
	fs := flag.NewFlagSet("task list", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var status string
	var epicID string
	var limit int
	fs.StringVar(&status, "status", "", "")
	fs.StringVar(&epicID, "epic", "", "")
	fs.IntVar(&limit, "limit", 20, "")

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
		limit:  limit,
	}, nil
}

type subtaskListOptions struct {
	taskID string
	limit  int
}

func parseSubtaskListArgs(args []string) (subtaskListOptions, error) {
	fs := flag.NewFlagSet("subtask list", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var limit int
	fs.IntVar(&limit, "limit", 20, "")

	if len(args) == 0 {
		return subtaskListOptions{}, fmt.Errorf("subtask list requires <task-id>")
	}
	taskID := args[0]

	if err := fs.Parse(args[1:]); err != nil {
		return subtaskListOptions{}, fmt.Errorf("invalid subtask list flags: %w", err)
	}
	if fs.NArg() > 0 {
		return subtaskListOptions{}, fmt.Errorf("subtask list does not accept positional arguments after <task-id>")
	}

	return subtaskListOptions{taskID: taskID, limit: limit}, nil
}

type readyOptions struct {
	epicID string
	limit  int
}

func parseReadyArgs(args []string) (readyOptions, error) {
	fs := flag.NewFlagSet("ready", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var epicID string
	var limit int
	fs.StringVar(&epicID, "epic", "", "")
	fs.IntVar(&limit, "limit", 20, "")

	if err := fs.Parse(args); err != nil {
		return readyOptions{}, fmt.Errorf("invalid ready flags: %w", err)
	}
	if fs.NArg() > 0 {
		return readyOptions{}, fmt.Errorf("ready does not accept positional arguments")
	}

	return readyOptions{epicID: epicID, limit: limit}, nil
}
