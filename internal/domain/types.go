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

type Task struct {
	ID          string
	EpicID      string
	Title       string
	Description string
	Status      Status
	Priority    Priority
	Tags        []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Epic struct {
	ID          string
	Title       string
	Description string
	Status      Status
	Priority    Priority
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Comment struct {
	ID        string
	EntityID  string
	Author    string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}
