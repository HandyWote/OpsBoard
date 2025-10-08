package transporthttp

import (
	"net/http"
	"strings"
	"time"

	"backend/internal/domain/task"
	"backend/internal/service"
)

type createTaskRequest struct {
	Title           string   `json:"title"`
	DescriptionHTML string   `json:"descriptionHtml"`
	Bounty          int64    `json:"bounty"`
	Priority        string   `json:"priority"`
	Deadline        string   `json:"deadline"`
	Tags            []string `json:"tags"`
	TagsText        string   `json:"tagsText"`
	Publish         bool     `json:"publish"`
}

type updateTaskRequest struct {
	Title           *string   `json:"title"`
	DescriptionHTML *string   `json:"descriptionHtml"`
	Bounty          *int64    `json:"bounty"`
	Priority        *string   `json:"priority"`
	Deadline        *string   `json:"deadline"`
	Tags            *[]string `json:"tags"`
	TagsText        *string   `json:"tagsText"`
	Status          *string   `json:"status"`
}

func (h *Handler) handleListTasks(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	sortKey := r.URL.Query().Get("sort")
	page := queryInt(r, "page", 1)
	pageSize := queryInt(r, "pageSize", 20)

	statusParam := r.URL.Query().Get("status")
	statuses := make([]task.Status, 0)
	if strings.TrimSpace(statusParam) != "" {
		for _, part := range strings.Split(statusParam, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			statuses = append(statuses, task.Status(part))
		}
	}

	result, err := h.services.Tasks.ListTasks(r.Context(), service.TaskListInput{
		Keyword:  keyword,
		Status:   statuses,
		SortKey:  sortKey,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		h.respondServiceError(w, err)
		return
	}

	tasks := make([]taskDTO, 0, len(result.Items))
	for _, item := range result.Items {
		tasks = append(tasks, mapTask(item))
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"items":    tasks,
		"total":    result.Total,
		"page":     result.Page,
		"pageSize": result.PageSize,
	})
}

func (h *Handler) handleGetTask(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_id", "任务 ID 不合法")
		return
	}

	t, err := h.services.Tasks.GetTask(r.Context(), id)
	if err != nil {
		h.respondServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, mapTask(t))
}

func (h *Handler) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := CurrentUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "未授权访问")
		return
	}

	var req createTaskRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_payload", "请求格式不正确")
		return
	}

	deadline := parseTime(req.Deadline)
	tags := mergeTags(req.Tags, req.TagsText)

	priority := task.Priority(strings.TrimSpace(req.Priority))

	created, err := h.services.Tasks.CreateTask(r.Context(), service.TaskCreateInput{
		Title:           req.Title,
		DescriptionHTML: req.DescriptionHTML,
		Bounty:          req.Bounty,
		Priority:        priority,
		Deadline:        deadline,
		Tags:            tags,
		CreatedBy:       userID,
		Publish:         req.Publish,
	})
	if err != nil {
		h.respondServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, mapTask(created))
}

func (h *Handler) handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_id", "任务 ID 不合法")
		return
	}

	var req updateTaskRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_payload", "请求格式不正确")
		return
	}

	input := service.TaskUpdateInput{ID: id}
	if req.Title != nil {
		input.Title = req.Title
	}
	if req.DescriptionHTML != nil {
		input.DescriptionHTML = req.DescriptionHTML
	}
	if req.Bounty != nil {
		input.Bounty = req.Bounty
	}
	if req.Priority != nil {
		priority := task.Priority(strings.TrimSpace(*req.Priority))
		input.Priority = &priority
	}
	if req.Deadline != nil {
		parsed := parseTime(*req.Deadline)
		input.Deadline = parsed
	}
	if req.Tags != nil || req.TagsText != nil {
		tags := make([]string, 0)
		if req.Tags != nil {
			tags = mergeTags(*req.Tags, "")
		}
		if req.TagsText != nil {
			tags = mergeTags(tags, *req.TagsText)
		}
		input.Tags = &tags
	}
	if req.Status != nil {
		status := task.Status(strings.TrimSpace(*req.Status))
		input.Status = &status
	}

	updated, err := h.services.Tasks.UpdateTask(r.Context(), input)
	if err != nil {
		h.respondServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, mapTask(updated))
}

func (h *Handler) handlePublishTask(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_id", "任务 ID 不合法")
		return
	}
	actor, ok := CurrentUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "未授权访问")
		return
	}

	updated, err := h.services.Tasks.PublishTask(r.Context(), id, actor)
	if err != nil {
		h.respondServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, mapTask(updated))
}

func (h *Handler) handleArchiveTask(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_id", "任务 ID 不合法")
		return
	}
	actor, ok := CurrentUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "未授权访问")
		return
	}

	updated, err := h.services.Tasks.ArchiveTask(r.Context(), id, actor)
	if err != nil {
		h.respondServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, mapTask(updated))
}

func (h *Handler) handleClaimTask(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_id", "任务 ID 不合法")
		return
	}
	userID, ok := CurrentUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "未授权访问")
		return
	}

	updated, err := h.services.Tasks.ClaimTask(r.Context(), id, userID)
	if err != nil {
		h.respondServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, mapTask(updated))
}

func (h *Handler) handleReleaseTask(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_id", "任务 ID 不合法")
		return
	}
	userID, ok := CurrentUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "未授权访问")
		return
	}

	updated, err := h.services.Tasks.ReleaseTask(r.Context(), id, userID)
	if err != nil {
		h.respondServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, mapTask(updated))
}

func (h *Handler) handleCompleteTask(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_id", "任务 ID 不合法")
		return
	}
	userID, ok := CurrentUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "未授权访问")
		return
	}

	updated, err := h.services.Tasks.CompleteTask(r.Context(), id, userID)
	if err != nil {
		h.respondServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, mapTask(updated))
}

func parseTime(value string) *time.Time {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	if t, err := time.Parse(time.RFC3339, trimmed); err == nil {
		return &t
	}
	if t, err := time.Parse("2006-01-02T15:04", trimmed); err == nil {
		return &t
	}
	return nil
}

func mergeTags(existing []string, tagsText string) []string {
	result := make([]string, 0, len(existing))
	seen := make(map[string]struct{})
	for _, tagName := range existing {
		tagName = strings.TrimSpace(tagName)
		if tagName == "" {
			continue
		}
		lower := strings.ToLower(tagName)
		if _, ok := seen[lower]; ok {
			continue
		}
		seen[lower] = struct{}{}
		result = append(result, tagName)
	}
	for _, tagName := range strings.Split(tagsText, ",") {
		tagName = strings.TrimSpace(tagName)
		if tagName == "" {
			continue
		}
		lower := strings.ToLower(tagName)
		if _, ok := seen[lower]; ok {
			continue
		}
		seen[lower] = struct{}{}
		result = append(result, tagName)
	}
	return result
}
