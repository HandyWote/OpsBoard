package task

import (
	"time"

	"github.com/google/uuid"
)

// Priority 优先级枚举。
type Priority string

// Status 任务状态枚举。
type Status string

const (
	PriorityCritical Priority = "critical"
	PriorityHigh     Priority = "high"
	PriorityMedium   Priority = "medium"
	PriorityLow      Priority = "low"

	StatusDraft     Status = "draft"
	StatusAvailable Status = "available"
	StatusClaimed   Status = "claimed"
	StatusSubmitted Status = "submitted"
	StatusCompleted Status = "completed"
	StatusArchived  Status = "archived"
)

// Task 描述任务主体。
type Task struct {
	ID               uuid.UUID
	Title            string
	DescriptionHTML  string
	DescriptionPlain string
	Bounty           int64
	Priority         Priority
	Status           Status
	Deadline         *time.Time
	CreatedBy        uuid.UUID
	PublishedBy      *uuid.UUID
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Tags             []Tag
	CurrentAssignee  *Assignment
}

// Tag 为任务分类标签。
type Tag struct {
	ID        int64
	Name      string
	CreatedAt time.Time
}

// Assignment 记录领取信息。
type Assignment struct {
	ID          int64
	TaskID      uuid.UUID
	UserID      uuid.UUID
	Username    string
	Status      Status
	CreatedAt   time.Time
	CompletedAt *time.Time
	ReleasedAt  *time.Time
}
