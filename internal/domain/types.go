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
	IssuePrefix   string `json:"issue_prefix"`
}

type AdminBootstrapResponse struct {
	WorkspaceID  string `json:"workspace_id"`
	ProjectID    string `json:"project_id"`
	AccessKey    string `json:"access_key"`
	AccessSecret string `json:"access_secret"`
}
