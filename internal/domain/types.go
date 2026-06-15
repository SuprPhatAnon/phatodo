package domain

import "time"

type Status string

const (
	StatusTodo       Status = "todo"
	StatusInProgress Status = "in_progress"
	StatusCompleted  Status = "completed"
	StatusWontFix    Status = "wont_fix"
	StatusArchived   Status = "archived"
)

type Priority int

const (
	PriorityCritical Priority = iota
	PriorityHigh
	PriorityMedium
	PriorityLow
	PriorityBacklog
	PrioritySomeday
)

type UserRole string

const (
	UserRoleAdmin UserRole = "admin"
	UserRoleUser  UserRole = "user"
)

type User struct {
	ID               string
	DisplayName      string
	Role             UserRole
	AccessKey        string
	AccessSecretHash string
	Username         string
	PasswordHash     string
	DisabledAt       *time.Time
	LastSeenAt       *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type Task struct {
	ID           string
	WorkspaceID  string
	ProjectID    string
	EpicID       string
	ParentTaskID string
	Title        string
	Description  string
	Status       Status
	Priority     Priority
	Tags         []string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type TaskCreateRequest struct {
	Title              string    `json:"title"`
	IssuePrefix        string    `json:"issue_prefix"`
	Description        string    `json:"description"`
	Priority           *Priority `json:"priority,omitempty"`
	EpicID             string    `json:"epic_id"`
	Tags               []string  `json:"tags"`
	AssignedTo         string    `json:"assigned_to"`
	AcceptanceCriteria []string  `json:"acceptance_criteria"`
}

type TaskCreateResponse struct {
	ID          string   `json:"id"`
	IssuePrefix string   `json:"issue_prefix"`
	Title       string   `json:"title"`
	Status      Status   `json:"status"`
	Priority    Priority `json:"priority"`
	ProjectID   string   `json:"project_id"`
	WorkspaceID string   `json:"workspace_id"`
}

type Epic struct {
	ID          string
	WorkspaceID string
	ProjectID   string
	Title       string
	Description string
	Status      Status
	Priority    Priority
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Comment struct {
	ID          string
	WorkspaceID string
	ProjectID   string
	TaskID      string
	Author      string
	Content     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ProjectConfig struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ProjectConfigSetRequest struct {
	Value string `json:"value"`
}

type AdminInitRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AdminInitResponse struct {
	UserID       string `json:"user_id"`
	Username     string `json:"username"`
	AccessKey    string `json:"access_key"`
	AccessSecret string `json:"access_secret"`
}

type AdminBootstrapRequest struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	WorkspaceName string `json:"workspace_name"`
	ProjectName   string `json:"project_name"`
}

type AdminBootstrapResponse struct {
	WorkspaceID  string `json:"workspace_id"`
	ProjectID    string `json:"project_id"`
	AccessKey    string `json:"access_key"`
	AccessSecret string `json:"access_secret"`
}
