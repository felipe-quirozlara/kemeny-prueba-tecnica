package model

import (
	"time"
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	AvatarURL    *string   `json:"avatar_url"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Task struct {
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	Description    *string   `json:"description"`
	Status         string    `json:"status"`
	Priority       string    `json:"priority"`
	Category       *string   `json:"category"`
	Summary        *string   `json:"summary"`
	CreatorID      string    `json:"creator_id"`
	AssigneeID     *string   `json:"assignee_id"`
	DueDate        *time.Time `json:"due_date"`
	EstimatedHours *float64  `json:"estimated_hours"`
	ActualHours    *float64  `json:"actual_hours"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// Joined fields
	Creator  *User  `json:"creator,omitempty"`
	Assignee *User  `json:"assignee,omitempty"`
	Tags     []Tag  `json:"tags,omitempty"`
}

type EditHistory struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"task_id"`
	UserID    string    `json:"user_id"`
	FieldName string    `json:"field_name"`
	OldValue  *string   `json:"old_value"`
	NewValue  *string   `json:"new_value"`
	EditedAt  time.Time `json:"edited_at"`
}

type Tag struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
}

type TaskTag struct {
	TaskID     string    `json:"task_id"`
	TagID      string    `json:"tag_id"`
	AssignedBy string    `json:"assigned_by"` // "manual" or "ai"
	AssignedAt time.Time `json:"assigned_at"`
}

type CreateTaskRequest struct {
	Title          string  `json:"title"`
	Description    *string `json:"description"`
	Status         string  `json:"status"`
	Priority       string  `json:"priority"`
	AssigneeID     *string `json:"assignee_id"`
	DueDate        *string `json:"due_date"`
	EstimatedHours *float64 `json:"estimated_hours"`
}

type UpdateTaskRequest struct {
	Title          *string  `json:"title"`
	Description    *string  `json:"description"`
	Status         *string  `json:"status"`
	Priority       *string  `json:"priority"`
	AssigneeID     *string  `json:"assignee_id"`
	DueDate        *string  `json:"due_date"`
	EstimatedHours *float64 `json:"estimated_hours"`
	ActualHours    *float64 `json:"actual_hours"`
}
