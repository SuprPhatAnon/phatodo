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
	ParentTaskID       string    `json:"parent_task_id,omitempty"`
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

type TaskListItem struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description,omitempty"`
	Status       Status    `json:"status"`
	Priority     Priority  `json:"priority"`
	EpicID       string    `json:"epic_id,omitempty"`
	ParentTaskID string    `json:"parent_task_id,omitempty"`
	Tags         []string  `json:"tags,omitempty"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
}

type TaskListResponse struct {
	ProjectID string         `json:"project_id"`
	Items     []TaskListItem `json:"items"`
}

type TaskDetail struct {
	ID                 string    `json:"id"`
	WorkspaceID        string    `json:"workspace_id,omitempty"`
	ProjectID          string    `json:"project_id,omitempty"`
	EpicID             string    `json:"epic_id,omitempty"`
	ParentTaskID       string    `json:"parent_task_id,omitempty"`
	AssignedTo         string    `json:"assigned_to,omitempty"`
	CreatedBy          string    `json:"created_by,omitempty"`
	UpdatedBy          string    `json:"updated_by,omitempty"`
	CompletedBy        string    `json:"completed_by,omitempty"`
	Title              string    `json:"title"`
	Description        string    `json:"description,omitempty"`
	Status             Status    `json:"status"`
	Priority           Priority  `json:"priority"`
	Tags               []string  `json:"tags,omitempty"`
	AcceptanceCriteria []string  `json:"acceptance_criteria,omitempty"`
	CompletionEvidence []string  `json:"completion_evidence,omitempty"`
	CompletionSummary  string    `json:"completion_summary,omitempty"`
	CompletedAt        time.Time `json:"completed_at,omitempty"`
	CreatedAt          time.Time `json:"created_at,omitempty"`
	UpdatedAt          time.Time `json:"updated_at,omitempty"`
}

type TaskUpdateRequest struct {
	Title              *string   `json:"title,omitempty"`
	Description        *string   `json:"description,omitempty"`
	Priority           *Priority `json:"priority,omitempty"`
	Status             *Status   `json:"status,omitempty"`
	Tags               *[]string `json:"tags,omitempty"`
	EpicID             *string   `json:"epic_id,omitempty"`
	NoEpic             bool      `json:"no_epic,omitempty"`
	AssignedTo         *string   `json:"assigned_to,omitempty"`
	AcceptanceCriteria *[]string `json:"acceptance_criteria,omitempty"`
	CompletionSummary  *string   `json:"completion_summary,omitempty"`
	CompletionEvidence *[]string `json:"completion_evidence,omitempty"`
}

type ReadyListItem struct {
	ID           string         `json:"id"`
	Title        string         `json:"title"`
	Description  string         `json:"description,omitempty"`
	Status       Status         `json:"status"`
	Priority     Priority       `json:"priority"`
	EpicID       string         `json:"epic_id,omitempty"`
	ParentTaskID string         `json:"parent_task_id,omitempty"`
	Tags         []string       `json:"tags,omitempty"`
	Unblocks     []TaskListItem `json:"unblocks,omitempty"`
	CreatedAt    time.Time      `json:"created_at,omitempty"`
	UpdatedAt    time.Time      `json:"updated_at,omitempty"`
}

type ReadyListResponse struct {
	ProjectID string          `json:"project_id"`
	Items     []ReadyListItem `json:"items"`
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
	ID           string    `json:"id"`
	WorkspaceID  string    `json:"workspace_id,omitempty"`
	ProjectID    string    `json:"project_id,omitempty"`
	TaskID       string    `json:"task_id,omitempty"`
	AuthorUserID string    `json:"author_user_id,omitempty"`
	Author       string    `json:"author"`
	Kind         string    `json:"kind"`
	Content      string    `json:"content"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
}

type CommentCreateRequest struct {
	Author  string `json:"author"`
	Kind    string `json:"kind"`
	Content string `json:"content"`
}

type CommentUpdateRequest struct {
	Content string `json:"content"`
}

type CommentListResponse struct {
	ProjectID string    `json:"project_id"`
	TaskID    string    `json:"task_id"`
	Items     []Comment `json:"items"`
}

type Dependency struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id,omitempty"`
	ProjectID   string    `json:"project_id,omitempty"`
	TaskID      string    `json:"task_id"`
	DependsOnID string    `json:"depends_on_id"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
}

type DependencyListResponse struct {
	ProjectID string       `json:"project_id"`
	TaskID    string       `json:"task_id"`
	Items     []Dependency `json:"items"`
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
