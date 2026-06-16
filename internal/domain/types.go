package domain

import (
	"encoding/json"
	"time"
)

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
	Kind         TaskKind
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
	Kind               TaskKind  `json:"kind"`
	IssuePrefix        string    `json:"issue_prefix"`
	ParentTaskID       string    `json:"parent_task_id,omitempty"`
	Description        string    `json:"description"`
	Priority           *Priority `json:"priority,omitempty"`
	EpicID             string    `json:"epic_id"`
	RootCauseAnalysis  string    `json:"root_cause_analysis,omitempty"`
	Tags               []string  `json:"tags"`
	PlannedFiles       []string  `json:"planned_files,omitempty"`
	AssignedTo         string    `json:"assigned_to"`
	AcceptanceCriteria []string  `json:"acceptance_criteria"`
}

type TaskCreateResponse struct {
	ID           string   `json:"id"`
	IssuePrefix  string   `json:"issue_prefix"`
	Title        string   `json:"title"`
	Kind         TaskKind `json:"kind"`
	Status       Status   `json:"status"`
	RootCause    string   `json:"root_cause_analysis,omitempty"`
	PlannedFiles []string `json:"planned_files,omitempty"`
	Priority     Priority `json:"priority"`
	ProjectID    string   `json:"project_id"`
	WorkspaceID  string   `json:"workspace_id"`
}

type TaskListItem struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Kind         TaskKind  `json:"kind,omitempty"`
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
	Kind               TaskKind  `json:"kind,omitempty"`
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
	PlannedFiles       []string  `json:"planned_files,omitempty"`
	ChangedFiles       []string  `json:"changed_files,omitempty"`
	AcceptanceCriteria []string  `json:"acceptance_criteria,omitempty"`
	CompletionEvidence []string  `json:"completion_evidence,omitempty"`
	CompletionSummary  string    `json:"completion_summary,omitempty"`
	RootCauseAnalysis  string    `json:"root_cause_analysis,omitempty"`
	CompletedAt        time.Time `json:"completed_at,omitempty"`
	CreatedAt          time.Time `json:"created_at,omitempty"`
	UpdatedAt          time.Time `json:"updated_at,omitempty"`
}

type TaskUpdateRequest struct {
	Title              *string   `json:"title,omitempty"`
	Description        *string   `json:"description,omitempty"`
	Kind               *TaskKind `json:"kind,omitempty"`
	Priority           *Priority `json:"priority,omitempty"`
	Status             *Status   `json:"status,omitempty"`
	Tags               *[]string `json:"tags,omitempty"`
	EpicID             *string   `json:"epic_id,omitempty"`
	NoEpic             bool      `json:"no_epic,omitempty"`
	AssignedTo         *string   `json:"assigned_to,omitempty"`
	RootCauseAnalysis  *string   `json:"root_cause_analysis,omitempty"`
	ChangedFiles       *[]string `json:"changed_files,omitempty"`
	AcceptanceCriteria *[]string `json:"acceptance_criteria,omitempty"`
	CompletionSummary  *string   `json:"completion_summary,omitempty"`
	CompletionEvidence *[]string `json:"completion_evidence,omitempty"`
}

type TaskKind string

const (
	TaskKindTask    TaskKind = "task"
	TaskKindBug     TaskKind = "bug"
	TaskKindFeature TaskKind = "feature"
	TaskKindChore   TaskKind = "chore"
	TaskKindSpike   TaskKind = "spike"
)

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
	ID                 string    `json:"id"`
	WorkspaceID        string    `json:"workspace_id,omitempty"`
	ProjectID          string    `json:"project_id,omitempty"`
	AssignedTo         string    `json:"assigned_to,omitempty"`
	CreatedBy          string    `json:"created_by,omitempty"`
	UpdatedBy          string    `json:"updated_by,omitempty"`
	CompletedBy        string    `json:"completed_by,omitempty"`
	Title              string    `json:"title"`
	Description        string    `json:"description,omitempty"`
	Status             Status    `json:"status"`
	Priority           Priority  `json:"priority"`
	AcceptanceCriteria []string  `json:"acceptance_criteria,omitempty"`
	CompletionEvidence []string  `json:"completion_evidence,omitempty"`
	CompletionSummary  string    `json:"completion_summary,omitempty"`
	CompletedAt        time.Time `json:"completed_at,omitempty"`
	CreatedAt          time.Time `json:"created_at,omitempty"`
	UpdatedAt          time.Time `json:"updated_at,omitempty"`
}

type EpicCreateRequest struct {
	Title              string    `json:"title"`
	Description        string    `json:"description,omitempty"`
	Priority           *Priority `json:"priority,omitempty"`
	AssignedTo         string    `json:"assigned_to,omitempty"`
	AcceptanceCriteria []string  `json:"acceptance_criteria,omitempty"`
}

type EpicUpdateRequest struct {
	Title              *string   `json:"title,omitempty"`
	Description        *string   `json:"description,omitempty"`
	Priority           *Priority `json:"priority,omitempty"`
	Status             *Status   `json:"status,omitempty"`
	AssignedTo         *string   `json:"assigned_to,omitempty"`
	AcceptanceCriteria *[]string `json:"acceptance_criteria,omitempty"`
	CompletionSummary  *string   `json:"completion_summary,omitempty"`
	CompletionEvidence *[]string `json:"completion_evidence,omitempty"`
}

type EpicListResponse struct {
	ProjectID string `json:"project_id"`
	Items     []Epic `json:"items"`
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

type WorkItemLock struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id,omitempty"`
	ProjectID   string    `json:"project_id,omitempty"`
	EntityType  string    `json:"entity_type"`
	EntityID    string    `json:"entity_id"`
	LockedBy    string    `json:"locked_by"`
	Reason      string    `json:"reason,omitempty"`
	LeasedAt    time.Time `json:"leased_at,omitempty"`
	ExpiresAt   time.Time `json:"expires_at,omitempty"`
	ReleasedAt  time.Time `json:"released_at,omitempty"`
}

type LockAcquireRequest struct {
	EntityType string     `json:"entity_type"`
	EntityID   string     `json:"entity_id"`
	Reason     string     `json:"reason,omitempty"`
	TTL        string     `json:"ttl,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

type LockListResponse struct {
	ProjectID string         `json:"project_id"`
	Items     []WorkItemLock `json:"items"`
}

type SearchItem struct {
	EntityType   string    `json:"entity_type"`
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description,omitempty"`
	Content      string    `json:"content,omitempty"`
	Status       Status    `json:"status,omitempty"`
	Priority     Priority  `json:"priority,omitempty"`
	EpicID       string    `json:"epic_id,omitempty"`
	ParentTaskID string    `json:"parent_task_id,omitempty"`
	Author       string    `json:"author,omitempty"`
	Kind         string    `json:"kind,omitempty"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
}

type SearchResponse struct {
	ProjectID string       `json:"project_id"`
	Query     string       `json:"query"`
	Items     []SearchItem `json:"items"`
}

type UnifiedListItem struct {
	EntityType   string    `json:"entity_type"`
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description,omitempty"`
	Status       Status    `json:"status,omitempty"`
	Priority     Priority  `json:"priority,omitempty"`
	EpicID       string    `json:"epic_id,omitempty"`
	ParentTaskID string    `json:"parent_task_id,omitempty"`
	Tags         []string  `json:"tags,omitempty"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
}

type ListResponse struct {
	ProjectID string            `json:"project_id"`
	Items     []UnifiedListItem `json:"items"`
}

type HistoryEvent struct {
	ID          int64           `json:"id"`
	WorkspaceID string          `json:"workspace_id,omitempty"`
	ProjectID   string          `json:"project_id,omitempty"`
	Action      string          `json:"action"`
	EntityType  string          `json:"entity_type"`
	EntityID    string          `json:"entity_id"`
	ActorUserID string          `json:"actor_user_id,omitempty"`
	ActorLabel  string          `json:"actor_label,omitempty"`
	BeforeState json.RawMessage `json:"before_state,omitempty"`
	AfterState  json.RawMessage `json:"after_state,omitempty"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`
	CreatedAt   time.Time       `json:"created_at,omitempty"`
}

type HistoryResponse struct {
	ProjectID string         `json:"project_id"`
	Items     []HistoryEvent `json:"items"`
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
