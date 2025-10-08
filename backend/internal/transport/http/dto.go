package transporthttp

import (
	"time"

	"backend/internal/domain/task"
	"backend/internal/domain/user"

	"github.com/google/uuid"
)

type userDTO struct {
	ID          string   `json:"id"`
	Username    string   `json:"username"`
	DisplayName string   `json:"displayName"`
	Name        string   `json:"name"`
	Email       string   `json:"email,omitempty"`
	Headline    string   `json:"headline,omitempty"`
	Bio         string   `json:"bio,omitempty"`
	Roles       []string `json:"roles"`
}

type taskDTO struct {
	ID               string         `json:"id"`
	Title            string         `json:"title"`
	DescriptionHTML  string         `json:"descriptionHtml"`
	DescriptionPlain string         `json:"descriptionPlain"`
	Bounty           int64          `json:"bounty"`
	Priority         string         `json:"priority"`
	Status           string         `json:"status"`
	Deadline         *string        `json:"deadline,omitempty"`
	CreatedBy        string         `json:"createdBy"`
	PublishedBy      *string        `json:"publishedBy,omitempty"`
	CreatedAt        string         `json:"createdAt"`
	UpdatedAt        string         `json:"updatedAt"`
	Tags             []string       `json:"tags"`
	CurrentAssignee  *assignmentDTO `json:"currentAssignee,omitempty"`
}

type assignmentDTO struct {
	ID          int64   `json:"id"`
	UserID      string  `json:"userId"`
	Username    string  `json:"username"`
	Status      string  `json:"status"`
	AssignedAt  string  `json:"assignedAt"`
	CompletedAt *string `json:"completedAt,omitempty"`
	ReleasedAt  *string `json:"releasedAt,omitempty"`
}

func mapUser(u user.User) userDTO {
	roles := make([]string, 0, len(u.Roles))
	for _, role := range u.Roles {
		roles = append(roles, string(role))
	}
	return userDTO{
		ID:          u.ID.String(),
		Username:    u.Username,
		DisplayName: u.DisplayName,
		Name:        u.DisplayName,
		Email:       u.Email,
		Headline:    u.Headline,
		Bio:         u.Bio,
		Roles:       roles,
	}
}

func mapTask(t task.Task) taskDTO {
	dto := taskDTO{
		ID:               t.ID.String(),
		Title:            t.Title,
		DescriptionHTML:  t.DescriptionHTML,
		DescriptionPlain: t.DescriptionPlain,
		Bounty:           t.Bounty,
		Priority:         string(t.Priority),
		Status:           string(t.Status),
		CreatedBy:        t.CreatedBy.String(),
		CreatedAt:        t.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        t.UpdatedAt.Format(time.RFC3339),
		Tags:             make([]string, 0, len(t.Tags)),
	}
	if t.Deadline != nil {
		formatted := t.Deadline.Format(time.RFC3339)
		dto.Deadline = &formatted
	}
	if t.PublishedBy != nil {
		val := t.PublishedBy.String()
		dto.PublishedBy = &val
	}
	for _, tagItem := range t.Tags {
		dto.Tags = append(dto.Tags, tagItem.Name)
	}
	if t.CurrentAssignee != nil {
		assignee := assignmentDTO{
			ID:         t.CurrentAssignee.ID,
			Status:     string(t.CurrentAssignee.Status),
			Username:   t.CurrentAssignee.Username,
			AssignedAt: t.CurrentAssignee.CreatedAt.Format(time.RFC3339),
		}
		if t.CurrentAssignee.UserID != uuid.Nil {
			assignee.UserID = t.CurrentAssignee.UserID.String()
		}
		if t.CurrentAssignee.CompletedAt != nil {
			val := t.CurrentAssignee.CompletedAt.Format(time.RFC3339)
			assignee.CompletedAt = &val
		}
		if t.CurrentAssignee.ReleasedAt != nil {
			val := t.CurrentAssignee.ReleasedAt.Format(time.RFC3339)
			assignee.ReleasedAt = &val
		}
		dto.CurrentAssignee = &assignee
	}
	return dto
}
