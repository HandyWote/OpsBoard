package service

import (
	"context"
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"backend/internal/domain/task"
	"backend/internal/repository"

	"go.uber.org/zap"
)

// TaskService 管理任务的业务逻辑。
type TaskService struct {
	repo repository.TaskRepository
	log  *zap.Logger
}

// TaskListInput 控制任务查询条件。
type TaskListInput struct {
	Keyword  string
	Status   []task.Status
	SortKey  string
	Page     int
	PageSize int
}

// TaskListResult 是分页返回结果。
type TaskListResult struct {
	Items    []task.Task
	Total    int
	Page     int
	PageSize int
}

// TaskCreateInput 描述新任务的字段。
type TaskCreateInput struct {
	Title           string
	DescriptionHTML string
	Bounty          int64
	Priority        task.Priority
	Deadline        *time.Time
	Tags            []string
	CreatedBy       uuid.UUID
	Publish         bool
}

// TaskUpdateInput 描述任务更新字段。
type TaskUpdateInput struct {
	ID              uuid.UUID
	Title           *string
	DescriptionHTML *string
	Bounty          *int64
	Priority        *task.Priority
	Deadline        *time.Time
	Tags            *[]string
	Status          *task.Status
}

// NewTaskService 构造任务服务。
func NewTaskService(repo repository.TaskRepository, log *zap.Logger) *TaskService {
	if log == nil {
		log = zap.NewNop()
	}
	return &TaskService{repo: repo, log: log}
}

// ListTasks 返回分页任务数据。
func (s *TaskService) ListTasks(ctx context.Context, input TaskListInput) (TaskListResult, error) {
	pageSize := input.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	page := input.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	tasks, total, err := s.repo.List(ctx, repository.TaskFilter{
		Keyword: strings.TrimSpace(input.Keyword),
		Status:  input.Status,
		SortKey: input.SortKey,
		Limit:   pageSize,
		Offset:  offset,
	})
	if err != nil {
		return TaskListResult{}, err
	}

	return TaskListResult{
		Items:    tasks,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// CreateTask 新建任务，可选择立即发布。
func (s *TaskService) CreateTask(ctx context.Context, input TaskCreateInput) (task.Task, error) {
	if strings.TrimSpace(input.Title) == "" {
		return task.Task{}, fmt.Errorf("%w: title required", ErrValidation)
	}
	if len([]rune(input.Title)) > 120 {
		return task.Task{}, fmt.Errorf("%w: title too long", ErrValidation)
	}
	if input.Bounty < 0 {
		return task.Task{}, fmt.Errorf("%w: bounty must be non-negative", ErrValidation)
	}
	priority := input.Priority
	if priority == "" {
		priority = task.PriorityMedium
	}

	plain := s.extractPlainText(input.DescriptionHTML)
	if strings.TrimSpace(plain) == "" {
		return task.Task{}, fmt.Errorf("%w: description required", ErrValidation)
	}

	cleanedTags := s.normalizeTags(input.Tags)

	created, err := s.repo.Create(ctx, repository.TaskCreateInput{
		Title:            strings.TrimSpace(input.Title),
		DescriptionHTML:  input.DescriptionHTML,
		DescriptionPlain: plain,
		Bounty:           input.Bounty,
		Priority:         priority,
		Deadline:         input.Deadline,
		CreatedBy:        input.CreatedBy,
		Tags:             cleanedTags,
	})
	if err != nil {
		return task.Task{}, err
	}

	if input.Publish {
		created, err = s.repo.SetStatus(ctx, created.ID, task.StatusAvailable, input.CreatedBy)
		if err != nil {
			return task.Task{}, err
		}
	}

	return created, nil
}

// UpdateTask 更新任务字段。
func (s *TaskService) UpdateTask(ctx context.Context, input TaskUpdateInput) (task.Task, error) {
	update := repository.TaskUpdateInput{ID: input.ID}
	if input.Title != nil {
		title := strings.TrimSpace(*input.Title)
		if title == "" {
			return task.Task{}, fmt.Errorf("%w: title required", ErrValidation)
		}
		update.Title = &title
	}
	if input.DescriptionHTML != nil {
		htmlText := *input.DescriptionHTML
		plain := s.extractPlainText(htmlText)
		if strings.TrimSpace(plain) == "" {
			return task.Task{}, fmt.Errorf("%w: description required", ErrValidation)
		}
		update.DescriptionHTML = &htmlText
		update.DescriptionPlain = &plain
	}
	if input.Bounty != nil {
		if *input.Bounty < 0 {
			return task.Task{}, fmt.Errorf("%w: bounty negative", ErrValidation)
		}
		update.Bounty = input.Bounty
	}
	if input.Priority != nil {
		update.Priority = input.Priority
	}
	if input.Deadline != nil {
		update.Deadline = input.Deadline
	}
	if input.Tags != nil {
		tags := s.normalizeTags(*input.Tags)
		update.Tags = &tags
	}
	if input.Status != nil {
		update.Status = input.Status
	}

	return s.repo.Update(ctx, update)
}

// PublishTask 将任务状态切换为可领取。
func (s *TaskService) PublishTask(ctx context.Context, taskID uuid.UUID, actor uuid.UUID) (task.Task, error) {
	return s.repo.SetStatus(ctx, taskID, task.StatusAvailable, actor)
}

// ArchiveTask 将任务归档。
func (s *TaskService) ArchiveTask(ctx context.Context, taskID uuid.UUID, actor uuid.UUID) (task.Task, error) {
	return s.repo.SetStatus(ctx, taskID, task.StatusArchived, actor)
}

// ClaimTask 领取任务。
func (s *TaskService) ClaimTask(ctx context.Context, taskID, userID uuid.UUID) (task.Task, error) {
	return s.repo.Claim(ctx, repository.TaskAssignmentInput{TaskID: taskID, UserID: userID})
}

// ReleaseTask 释放任务。
func (s *TaskService) ReleaseTask(ctx context.Context, taskID, userID uuid.UUID) (task.Task, error) {
	return s.repo.Release(ctx, repository.TaskAssignmentInput{TaskID: taskID, UserID: userID})
}

// CompleteTask 完成任务。
func (s *TaskService) CompleteTask(ctx context.Context, taskID, userID uuid.UUID) (task.Task, error) {
	return s.repo.Complete(ctx, repository.TaskAssignmentInput{TaskID: taskID, UserID: userID})
}

// GetTask 返回任务详情。
func (s *TaskService) GetTask(ctx context.Context, taskID uuid.UUID) (task.Task, error) {
	return s.repo.GetByID(ctx, taskID)
}

var tagCleaner = regexp.MustCompile(`<[^>]*>`)

func (s *TaskService) extractPlainText(htmlSource string) string {
	noTags := tagCleaner.ReplaceAllString(htmlSource, " ")
	noTags = strings.ReplaceAll(noTags, "\n", " ")
	collapsed := strings.Join(strings.Fields(noTags), " ")
	return html.UnescapeString(collapsed)
}

func (s *TaskService) normalizeTags(tags []string) []string {
	uniq := make(map[string]struct{})
	cleaned := make([]string, 0, len(tags))
	for _, tagName := range tags {
		tagName = strings.TrimSpace(tagName)
		if tagName == "" {
			continue
		}
		lowered := strings.ToLower(tagName)
		if _, exists := uniq[lowered]; exists {
			continue
		}
		uniq[lowered] = struct{}{}
		cleaned = append(cleaned, tagName)
	}
	return cleaned
}
