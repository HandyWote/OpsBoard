package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"backend/internal/domain/task"
)

// TaskFilter 控制任务查询条件。
type TaskFilter struct {
	Keyword        string
	Status         []task.Status
	SortKey        string
	Limit          int
	Offset         int
	AssignedTo     uuid.UUID
	IncludeDeleted bool
}

// TaskCreateInput 描述创建任务所需字段。
type TaskCreateInput struct {
	Title            string
	DescriptionHTML  string
	DescriptionPlain string
	Bounty           int64
	Priority         task.Priority
	Deadline         *time.Time
	CreatedBy        uuid.UUID
	Tags             []string
}

// TaskUpdateInput 描述更新任务的字段。
type TaskUpdateInput struct {
	ID               uuid.UUID
	Title            *string
	DescriptionHTML  *string
	DescriptionPlain *string
	Bounty           *int64
	Priority         *task.Priority
	Deadline         *time.Time
	Status           *task.Status
	Tags             *[]string
}

// TaskAssignmentInput 用于任务领取与释放。
type TaskAssignmentInput struct {
	TaskID uuid.UUID
	UserID uuid.UUID
}

// TaskRepository 定义任务相关数据库操作。
type TaskRepository interface {
	List(ctx context.Context, filter TaskFilter) ([]task.Task, int, error)
	Create(ctx context.Context, input TaskCreateInput) (task.Task, error)
	Update(ctx context.Context, input TaskUpdateInput) (task.Task, error)
	Delete(ctx context.Context, id uuid.UUID) error
	SetStatus(ctx context.Context, taskID uuid.UUID, status task.Status, actor uuid.UUID) (task.Task, error)
	Claim(ctx context.Context, input TaskAssignmentInput) (task.Task, error)
	Release(ctx context.Context, input TaskAssignmentInput) (task.Task, error)
	Submit(ctx context.Context, input TaskAssignmentInput) (task.Task, error)
	Reject(ctx context.Context, input TaskAssignmentInput) (task.Task, error)
	Complete(ctx context.Context, input TaskAssignmentInput) (task.Task, error)
	GetByID(ctx context.Context, id uuid.UUID) (task.Task, error)
}

type taskRepository struct {
	db *sql.DB
}

// NewTaskRepository 构造任务仓储实例。
func NewTaskRepository(db *sql.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) List(ctx context.Context, filter TaskFilter) ([]task.Task, int, error) {
	args := make([]any, 0)
	conditions := make([]string, 0)

	if !filter.IncludeDeleted {
		conditions = append(conditions, "t.deleted_at IS NULL")
	}

	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		args = append(args, "%"+strings.ToLower(keyword)+"%")
		placeholder := fmt.Sprintf("$%d", len(args))
		conditions = append(conditions, fmt.Sprintf("(LOWER(t.title) LIKE %s OR LOWER(t.description_plain) LIKE %s)", placeholder, placeholder))
	}

	if len(filter.Status) > 0 {
		placeholders := make([]string, 0, len(filter.Status))
		for _, st := range filter.Status {
			args = append(args, string(st))
			placeholders = append(placeholders, fmt.Sprintf("$%d", len(args)))
		}
		conditions = append(conditions, fmt.Sprintf("t.status IN (%s)", strings.Join(placeholders, ", ")))
	}

	if filter.AssignedTo != uuid.Nil {
		args = append(args, filter.AssignedTo)
		placeholder := fmt.Sprintf("$%d", len(args))
		conditions = append(conditions, fmt.Sprintf(`
EXISTS (
	SELECT 1
	FROM task_assignments ta
	WHERE ta.task_id = t.id
		AND ta.user_id = %s
		AND ta.status = t.status
)`, placeholder))
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	sortKey := strings.ToLower(strings.TrimSpace(filter.SortKey))
	sortClause := "ORDER BY t.created_at DESC"
	switch sortKey {
	case "deadline":
		sortClause = "ORDER BY t.deadline NULLS LAST, t.created_at DESC"
	case "priority":
		sortClause = `
ORDER BY
	CASE t.priority
		WHEN 'critical' THEN 1
		WHEN 'high' THEN 2
		WHEN 'medium' THEN 3
		WHEN 'low' THEN 4
		ELSE 5
	END,
	t.created_at DESC`
	case "status_priority":
		sortClause = `
ORDER BY
	CASE
		WHEN t.status = 'available' THEN 1
		WHEN t.status = 'claimed' THEN 2
		WHEN t.status = 'submitted' THEN 3
		WHEN t.status = 'completed' THEN 4
		ELSE 5
	END,
	CASE t.priority
		WHEN 'critical' THEN 1
		WHEN 'high' THEN 2
		WHEN 'medium' THEN 3
		WHEN 'low' THEN 4
		ELSE 5
	END,
	t.created_at DESC`
	case "bounty_desc":
		sortClause = "ORDER BY t.bounty DESC, t.created_at DESC"
	case "created_desc":
		sortClause = "ORDER BY t.created_at DESC"
	case "created_asc":
		sortClause = "ORDER BY t.created_at ASC"
	}

	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	argsWithPagination := append(args, limit, offset)

	query := fmt.Sprintf(`
SELECT
	t.id,
	t.title,
	t.description_html,
	t.description_plain,
	t.bounty,
	t.priority,
	t.status,
	t.deadline,
	t.created_by,
	t.published_by,
	t.created_at,
	t.updated_at,
	t.deleted_at,
	la.assignment_id,
	la.user_id,
	la.display_name,
	la.assignment_status,
	la.assigned_at,
	la.completed_at,
	la.released_at
FROM tasks t
LEFT JOIN LATERAL (
	SELECT
		ta.id AS assignment_id,
		ta.user_id,
		u.display_name,
		ta.status AS assignment_status,
		ta.created_at AS assigned_at,
		ta.completed_at,
		ta.released_at
	FROM task_assignments ta
	JOIN users u ON u.id = ta.user_id
	WHERE ta.task_id = t.id
	ORDER BY ta.created_at DESC
	LIMIT 1
) la ON true
%s
%s
LIMIT $%d OFFSET $%d
`, where, sortClause, len(argsWithPagination)-1, len(argsWithPagination))

	rows, err := r.db.QueryContext(ctx, query, argsWithPagination...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	taskList := make([]task.Task, 0)
	taskIndex := make(map[uuid.UUID]int)
	taskIDs := make([]uuid.UUID, 0)

	for rows.Next() {
		var (
			tk               task.Task
			publishedByNull  sql.NullString
			deadlineNull     sql.NullTime
			deletedNull      sql.NullTime
			assignmentID     sql.NullInt64
			assignmentUser   sql.NullString
			assignmentName   sql.NullString
			assignmentStatus sql.NullString
			assignedAt       sql.NullTime
			completedAt      sql.NullTime
			releasedAt       sql.NullTime
		)

		err := rows.Scan(
			&tk.ID,
			&tk.Title,
			&tk.DescriptionHTML,
			&tk.DescriptionPlain,
			&tk.Bounty,
			&tk.Priority,
			&tk.Status,
			&deadlineNull,
			&tk.CreatedBy,
			&publishedByNull,
			&tk.CreatedAt,
			&tk.UpdatedAt,
			&deletedNull,
			&assignmentID,
			&assignmentUser,
			&assignmentName,
			&assignmentStatus,
			&assignedAt,
			&completedAt,
			&releasedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		if deadlineNull.Valid {
			deadline := deadlineNull.Time
			tk.Deadline = &deadline
		}

		if deletedNull.Valid {
			del := deletedNull.Time
			tk.DeletedAt = &del
		}

		if publishedByNull.Valid {
			if pubID, err := uuid.Parse(publishedByNull.String); err == nil {
				tk.PublishedBy = &pubID
			}
		}

		if assignmentID.Valid {
			assign := &task.Assignment{
				ID:     assignmentID.Int64,
				TaskID: tk.ID,
				Status: task.Status(assignmentStatus.String),
			}
			if assignmentUser.Valid {
				if uid, err := uuid.Parse(assignmentUser.String); err == nil {
					assign.UserID = uid
				}
			}
			if assignmentName.Valid {
				assign.Username = assignmentName.String
			}
			if assignedAt.Valid {
				assign.CreatedAt = assignedAt.Time
			}
			if completedAt.Valid {
				t := completedAt.Time
				assign.CompletedAt = &t
			}
			if releasedAt.Valid {
				t := releasedAt.Time
				assign.ReleasedAt = &t
			}
			tk.CurrentAssignee = assign
		}

		taskIndex[tk.ID] = len(taskList)
		taskList = append(taskList, tk)
		taskIDs = append(taskIDs, tk.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	if len(taskIDs) > 0 {
		if err := r.attachTagsToTasks(ctx, taskIDs, taskList, taskIndex); err != nil {
			return nil, 0, err
		}
	}

	countQuery := "SELECT COUNT(*) FROM tasks t " + where
	var total int
	if len(conditions) == 0 {
		if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
			return nil, 0, err
		}
	} else {
		if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
			return nil, 0, err
		}
	}

	return taskList, total, nil
}

func (r *taskRepository) Create(ctx context.Context, input TaskCreateInput) (task.Task, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return task.Task{}, err
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	id := uuid.New()

	const insertTask = `
INSERT INTO tasks (id, title, description_html, description_plain, bounty, priority, status, deadline, created_by, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, 'draft', $7, $8, $9, $9)
RETURNING id, title, description_html, description_plain, bounty, priority, status, deadline, created_by, published_by, created_at, updated_at
`

	var (
		tk           task.Task
		deadlineNull sql.NullTime
		pubNull      sql.NullString
	)
	err = tx.QueryRowContext(ctx, insertTask,
		id,
		input.Title,
		input.DescriptionHTML,
		input.DescriptionPlain,
		input.Bounty,
		input.Priority,
		input.Deadline,
		input.CreatedBy,
		now,
	).Scan(
		&tk.ID,
		&tk.Title,
		&tk.DescriptionHTML,
		&tk.DescriptionPlain,
		&tk.Bounty,
		&tk.Priority,
		&tk.Status,
		&deadlineNull,
		&tk.CreatedBy,
		&pubNull,
		&tk.CreatedAt,
		&tk.UpdatedAt,
	)
	if err != nil {
		return task.Task{}, err
	}

	if deadlineNull.Valid {
		dl := deadlineNull.Time
		tk.Deadline = &dl
	}
	if pubNull.Valid {
		if id, err := uuid.Parse(pubNull.String); err == nil {
			tk.PublishedBy = &id
		}
	}

	if err := r.attachTags(ctx, tx, tk.ID, input.Tags); err != nil {
		return task.Task{}, err
	}

	if err := tx.Commit(); err != nil {
		return task.Task{}, err
	}

	tk.Tags = make([]task.Tag, 0, len(input.Tags))
	for _, name := range input.Tags {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		tk.Tags = append(tk.Tags, task.Tag{Name: name})
	}
	return tk, nil
}

func (r *taskRepository) Update(ctx context.Context, input TaskUpdateInput) (task.Task, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return task.Task{}, err
	}
	defer tx.Rollback()

	setParts := make([]string, 0)
	args := make([]any, 0)

	if input.Title != nil {
		args = append(args, *input.Title)
		setParts = append(setParts, fmt.Sprintf("title = $%d", len(args)))
	}
	if input.DescriptionHTML != nil {
		args = append(args, *input.DescriptionHTML)
		setParts = append(setParts, fmt.Sprintf("description_html = $%d", len(args)))
	}
	if input.DescriptionPlain != nil {
		args = append(args, *input.DescriptionPlain)
		setParts = append(setParts, fmt.Sprintf("description_plain = $%d", len(args)))
	}
	if input.Bounty != nil {
		args = append(args, *input.Bounty)
		setParts = append(setParts, fmt.Sprintf("bounty = $%d", len(args)))
	}
	if input.Priority != nil {
		args = append(args, *input.Priority)
		setParts = append(setParts, fmt.Sprintf("priority = $%d", len(args)))
	}
	if input.Deadline != nil {
		args = append(args, *input.Deadline)
		setParts = append(setParts, fmt.Sprintf("deadline = $%d", len(args)))
	}
	if input.Status != nil {
		args = append(args, *input.Status)
		setParts = append(setParts, fmt.Sprintf("status = $%d", len(args)))
	}

	if len(setParts) > 0 {
		args = append(args, time.Now().UTC(), input.ID)
		query := fmt.Sprintf(`
UPDATE tasks
SET %s,
	updated_at = $%d
WHERE id = $%d
	AND deleted_at IS NULL
`, strings.Join(setParts, ", "), len(args)-1, len(args))
		result, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			return task.Task{}, err
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return task.Task{}, err
		}
		if affected == 0 {
			return task.Task{}, ErrNotFound
		}
	}

	if input.Tags != nil {
		if err := r.attachTags(ctx, tx, input.ID, *input.Tags); err != nil {
			return task.Task{}, err
		}
	}

	tk, err := r.fetchTaskTx(ctx, tx, input.ID)
	if err != nil {
		return task.Task{}, err
	}

	if err := tx.Commit(); err != nil {
		return task.Task{}, err
	}

	return tk, nil
}

func (r *taskRepository) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	result, err := r.db.ExecContext(ctx, `
UPDATE tasks
SET deleted_at = $2,
    status = CASE WHEN status = 'archived' THEN status ELSE 'archived' END,
    updated_at = $2
WHERE id = $1
	AND deleted_at IS NULL
`, id, now)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *taskRepository) SetStatus(ctx context.Context, taskID uuid.UUID, status task.Status, actor uuid.UUID) (task.Task, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return task.Task{}, err
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	const query = `
UPDATE tasks
SET status = $2,
	published_by = CASE WHEN $2 = 'available' THEN $3 ELSE published_by END,
	updated_at = $4
WHERE id = $1
	AND deleted_at IS NULL
`
	if _, err := tx.ExecContext(ctx, query, taskID, status, actor, now); err != nil {
		return task.Task{}, err
	}

	tk, err := r.fetchTaskTx(ctx, tx, taskID)
	if err != nil {
		return task.Task{}, err
	}

	if err := tx.Commit(); err != nil {
		return task.Task{}, err
	}

	return tk, nil
}

func (r *taskRepository) Claim(ctx context.Context, input TaskAssignmentInput) (task.Task, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return task.Task{}, err
	}
	defer tx.Rollback()

	var currentStatus string
	if err := tx.QueryRowContext(ctx, `SELECT status FROM tasks WHERE id = $1 AND deleted_at IS NULL`, input.TaskID).Scan(&currentStatus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return task.Task{}, ErrNotFound
		}
		return task.Task{}, err
	}
	if currentStatus != string(task.StatusAvailable) {
		return task.Task{}, fmt.Errorf("task not available for claim")
	}

	now := time.Now().UTC()
	if _, err := tx.ExecContext(ctx, `
INSERT INTO task_assignments (task_id, user_id, status, created_at)
VALUES ($1, $2, 'claimed', $3)
`, input.TaskID, input.UserID, now); err != nil {
		return task.Task{}, err
	}

	if _, err := tx.ExecContext(ctx, `
UPDATE tasks SET status = 'claimed', updated_at = $2 WHERE id = $1 AND deleted_at IS NULL
`, input.TaskID, now); err != nil {
		return task.Task{}, err
	}

	tk, err := r.fetchTaskTx(ctx, tx, input.TaskID)
	if err != nil {
		return task.Task{}, err
	}

	if err := tx.Commit(); err != nil {
		return task.Task{}, err
	}

	return tk, nil
}

func (r *taskRepository) Release(ctx context.Context, input TaskAssignmentInput) (task.Task, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return task.Task{}, err
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	res, err := tx.ExecContext(ctx, `
UPDATE task_assignments
SET status = 'released',
	released_at = $3
WHERE task_id = $1 AND user_id = $2 AND status = 'claimed'
`, input.TaskID, input.UserID, now)
	if err != nil {
		return task.Task{}, err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return task.Task{}, fmt.Errorf("no active claim")
	}

	if _, err := tx.ExecContext(ctx, `
UPDATE tasks SET status = 'available', updated_at = $2 WHERE id = $1 AND deleted_at IS NULL
`, input.TaskID, now); err != nil {
		return task.Task{}, err
	}

	tk, err := r.fetchTaskTx(ctx, tx, input.TaskID)
	if err != nil {
		return task.Task{}, err
	}

	if err := tx.Commit(); err != nil {
		return task.Task{}, err
	}
	return tk, nil
}

func (r *taskRepository) Submit(ctx context.Context, input TaskAssignmentInput) (task.Task, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return task.Task{}, err
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	res, err := tx.ExecContext(ctx, `
UPDATE task_assignments
SET status = 'submitted',
	completed_at = NULL,
	released_at = NULL
WHERE task_id = $1 AND user_id = $2 AND status = 'claimed'
`, input.TaskID, input.UserID)
	if err != nil {
		return task.Task{}, err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return task.Task{}, fmt.Errorf("no active claim to submit")
	}

	if _, err := tx.ExecContext(ctx, `
UPDATE tasks SET status = 'submitted', updated_at = $2 WHERE id = $1 AND deleted_at IS NULL
`, input.TaskID, now); err != nil {
		return task.Task{}, err
	}

	tk, err := r.fetchTaskTx(ctx, tx, input.TaskID)
	if err != nil {
		return task.Task{}, err
	}

	if err := tx.Commit(); err != nil {
		return task.Task{}, err
	}

	return tk, nil
}

func (r *taskRepository) Reject(ctx context.Context, input TaskAssignmentInput) (task.Task, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return task.Task{}, err
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	res, err := tx.ExecContext(ctx, `
UPDATE task_assignments
SET status = 'claimed',
	completed_at = NULL,
	released_at = NULL
WHERE task_id = $1 AND user_id = $2 AND status = 'submitted'
`, input.TaskID, input.UserID)
	if err != nil {
		return task.Task{}, err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return task.Task{}, fmt.Errorf("no submission to reject")
	}

	if _, err := tx.ExecContext(ctx, `
UPDATE tasks SET status = 'claimed', updated_at = $2 WHERE id = $1 AND deleted_at IS NULL
`, input.TaskID, now); err != nil {
		return task.Task{}, err
	}

	tk, err := r.fetchTaskTx(ctx, tx, input.TaskID)
	if err != nil {
		return task.Task{}, err
	}

	if err := tx.Commit(); err != nil {
		return task.Task{}, err
	}

	return tk, nil
}

func (r *taskRepository) Complete(ctx context.Context, input TaskAssignmentInput) (task.Task, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return task.Task{}, err
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	res, err := tx.ExecContext(ctx, `
UPDATE task_assignments
SET status = 'completed',
	completed_at = $3
WHERE task_id = $1 AND user_id = $2 AND status = 'submitted'
`, input.TaskID, input.UserID, now)
	if err != nil {
		return task.Task{}, err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return task.Task{}, fmt.Errorf("no submission to approve")
	}

	if _, err := tx.ExecContext(ctx, `
UPDATE tasks SET status = 'completed', updated_at = $2 WHERE id = $1 AND deleted_at IS NULL
`, input.TaskID, now); err != nil {
		return task.Task{}, err
	}

	tk, err := r.fetchTaskTx(ctx, tx, input.TaskID)
	if err != nil {
		return task.Task{}, err
	}

	if err := tx.Commit(); err != nil {
		return task.Task{}, err
	}
	return tk, nil
}

func (r *taskRepository) GetByID(ctx context.Context, id uuid.UUID) (task.Task, error) {
	return r.fetchTask(ctx, id)
}

func (r *taskRepository) fetchTask(ctx context.Context, id uuid.UUID) (task.Task, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return task.Task{}, err
	}
	defer tx.Rollback()

	tk, err := r.fetchTaskTx(ctx, tx, id)
	if err != nil {
		return task.Task{}, err
	}
	if err := tx.Commit(); err != nil {
		return task.Task{}, err
	}
	return tk, nil
}

func (r *taskRepository) fetchTaskTx(ctx context.Context, tx *sql.Tx, id uuid.UUID) (task.Task, error) {
	const query = `
SELECT
	t.id,
	t.title,
	t.description_html,
	t.description_plain,
	t.bounty,
	t.priority,
	t.status,
	t.deadline,
	t.created_by,
	t.published_by,
	t.created_at,
	t.updated_at,
	t.deleted_at
FROM tasks t
WHERE t.id = $1
	AND t.deleted_at IS NULL
`
	var (
		tk           task.Task
		deadlineNull sql.NullTime
		pubNull      sql.NullString
		deletedNull  sql.NullTime
	)
	err := tx.QueryRowContext(ctx, query, id).Scan(
		&tk.ID,
		&tk.Title,
		&tk.DescriptionHTML,
		&tk.DescriptionPlain,
		&tk.Bounty,
		&tk.Priority,
		&tk.Status,
		&deadlineNull,
		&tk.CreatedBy,
		&pubNull,
		&tk.CreatedAt,
		&tk.UpdatedAt,
		&deletedNull,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return task.Task{}, ErrNotFound
	}
	if err != nil {
		return task.Task{}, err
	}

	if deadlineNull.Valid {
		dl := deadlineNull.Time
		tk.Deadline = &dl
	}
	if pubNull.Valid {
		if id, err := uuid.Parse(pubNull.String); err == nil {
			tk.PublishedBy = &id
		}
	}
	if deletedNull.Valid {
		del := deletedNull.Time
		tk.DeletedAt = &del
	}

	if err := r.attachTagsForTask(ctx, tx, &tk); err != nil {
		return task.Task{}, err
	}

	const assignmentQuery = `
SELECT ta.id, ta.user_id, u.display_name, ta.status, ta.created_at, ta.completed_at, ta.released_at
FROM task_assignments ta
JOIN users u ON u.id = ta.user_id
WHERE ta.task_id = $1
ORDER BY ta.created_at DESC
LIMIT 1
`
	var (
		assignID        sql.NullInt64
		assignUser      sql.NullString
		assignName      sql.NullString
		assignStatus    sql.NullString
		assignCreated   sql.NullTime
		assignCompleted sql.NullTime
		assignReleased  sql.NullTime
	)
	err = tx.QueryRowContext(ctx, assignmentQuery, id).Scan(
		&assignID,
		&assignUser,
		&assignName,
		&assignStatus,
		&assignCreated,
		&assignCompleted,
		&assignReleased,
	)
	if err == nil && assignID.Valid {
		assign := &task.Assignment{
			ID:     assignID.Int64,
			TaskID: id,
			Status: task.Status(assignStatus.String),
		}
		if assignUser.Valid {
			if uid, parseErr := uuid.Parse(assignUser.String); parseErr == nil {
				assign.UserID = uid
			}
		}
		if assignName.Valid {
			assign.Username = assignName.String
		}
		if assignCreated.Valid {
			assign.CreatedAt = assignCreated.Time
		}
		if assignCompleted.Valid {
			t := assignCompleted.Time
			assign.CompletedAt = &t
		}
		if assignReleased.Valid {
			t := assignReleased.Time
			assign.ReleasedAt = &t
		}
		tk.CurrentAssignee = assign
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return task.Task{}, err
	}

	return tk, nil
}

func (r *taskRepository) attachTags(ctx context.Context, tx *sql.Tx, taskID uuid.UUID, tags []string) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM task_tag_map WHERE task_id = $1`, taskID); err != nil {
		return err
	}

	for _, name := range tags {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		var tagID int64
		if err := tx.QueryRowContext(ctx, `
INSERT INTO task_tags (name)
VALUES ($1)
ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
RETURNING id
`, name).Scan(&tagID); err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, `
INSERT INTO task_tag_map (task_id, tag_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING
`, taskID, tagID); err != nil {
			return err
		}
	}
	return nil
}

func (r *taskRepository) attachTagsForTask(ctx context.Context, tx *sql.Tx, tk *task.Task) error {
	rows, err := tx.QueryContext(ctx, `
SELECT tt.id, tt.name, tt.created_at
FROM task_tag_map tm
JOIN task_tags tt ON tt.id = tm.tag_id
WHERE tm.task_id = $1
ORDER BY tt.name ASC
`, tk.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	tk.Tags = make([]task.Tag, 0)
	for rows.Next() {
		var tagItem task.Tag
		if err := rows.Scan(&tagItem.ID, &tagItem.Name, &tagItem.CreatedAt); err != nil {
			return err
		}
		tk.Tags = append(tk.Tags, tagItem)
	}
	return rows.Err()
}

func (r *taskRepository) attachTagsToTasks(ctx context.Context, ids []uuid.UUID, tasks []task.Task, index map[uuid.UUID]int) error {
	if len(ids) == 0 {
		return nil
	}

	placeholders := make([]string, len(ids))
	args := make([]any, 0, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args = append(args, id)
	}

	query := fmt.Sprintf(`
SELECT tm.task_id, tt.id, tt.name, tt.created_at
FROM task_tag_map tm
JOIN task_tags tt ON tt.id = tm.tag_id
WHERE tm.task_id IN (%s)
ORDER BY tt.name ASC
`, strings.Join(placeholders, ", "))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			taskID  uuid.UUID
			tagItem task.Tag
		)
		if err := rows.Scan(&taskID, &tagItem.ID, &tagItem.Name, &tagItem.CreatedAt); err != nil {
			return err
		}
		if idx, ok := index[taskID]; ok {
			tasks[idx].Tags = append(tasks[idx].Tags, tagItem)
		}
	}
	return rows.Err()
}
