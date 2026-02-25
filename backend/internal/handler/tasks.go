package handler

import (
	"encoding/json"
	// "fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"

	"github.com/KemenyStudio/task-manager/internal/db"
	"github.com/KemenyStudio/task-manager/internal/middleware"
	"github.com/KemenyStudio/task-manager/internal/llm"
	"github.com/KemenyStudio/task-manager/internal/model"
)

var jwtSecret []byte
var llmClient llm.LLMClient

// SetLLMClient allows wiring a chosen LLM implementation at startup.
func SetLLMClient(c llm.LLMClient) {
    llmClient = c
}

func init() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default-secret-change-in-production"
	}
	jwtSecret = []byte(secret)
}

// ListTasks returns all tasks, optionally filtered by status and with assignee data.
func ListTasks(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	include := r.URL.Query().Get("include")

	// Improved: Use JOIN to fetch tasks and tags in one query
	var query string
	var args []interface{}
	if status != "" {
		query = `SELECT 
			tasks.id, tasks.title, tasks.description, tasks.status, tasks.priority, tasks.category, tasks.summary, tasks.creator_id, tasks.assignee_id, tasks.due_date, tasks.estimated_hours, tasks.actual_hours, tasks.created_at, tasks.updated_at,
			tags.id, tags.name, tags.color, tags.created_at
		FROM tasks
		LEFT JOIN task_tags ON tasks.id = task_tags.task_id
		LEFT JOIN tags ON task_tags.tag_id = tags.id
		WHERE tasks.status = $1
		ORDER BY tasks.created_at DESC`
		args = append(args, status)
	} else {
		query = `SELECT 
			tasks.id, tasks.title, tasks.description, tasks.status, tasks.priority, tasks.category, tasks.summary, tasks.creator_id, tasks.assignee_id, tasks.due_date, tasks.estimated_hours, tasks.actual_hours, tasks.created_at, tasks.updated_at,
			tags.id, tags.name, tags.color, tags.created_at
		FROM tasks
		LEFT JOIN task_tags ON tasks.id = task_tags.task_id
		LEFT JOIN tags ON task_tags.tag_id = tags.id
		ORDER BY tasks.created_at DESC`
	}

    rows, err := db.Pool.Query(r.Context(), query, args...)
    if err != nil {
        Error(w, r, http.StatusInternalServerError, "failed to query tasks", err, 0)
        return
    }
	defer rows.Close()

	taskMap := make(map[string]*model.Task)
	for rows.Next() {
		var (
			t model.Task
			tagID *string
			tagName *string
			tagColor *string
			tagCreatedAt *time.Time
		)
		err := rows.Scan(
			&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority,
			&t.Category, &t.Summary, &t.CreatorID, &t.AssigneeID,
			&t.DueDate, &t.EstimatedHours, &t.ActualHours,
			&t.CreatedAt, &t.UpdatedAt,
			&tagID, &tagName, &tagColor, &tagCreatedAt,
		)
		if err != nil {
			log.Printf("error scanning task+tag: %v", err)
			continue
		}
		task, exists := taskMap[t.ID]
		if !exists {
			taskCopy := t
			taskCopy.Tags = []model.Tag{}
			taskMap[t.ID] = &taskCopy
			task = &taskCopy
		}
		if tagID != nil && tagName != nil && tagColor != nil && tagCreatedAt != nil {
			task.Tags = append(task.Tags, model.Tag{
				ID:         *tagID,
				Name:       *tagName,
				Color:      *tagColor,
				CreatedAt:  *tagCreatedAt,
			})
		}
	}

	tasks := make([]model.Task, 0, len(taskMap))
	for _, t := range taskMap {
		tasks = append(tasks, *t)
	}

	// Load assignee for each task
	if include == "assignee" {
		for i, t := range tasks {
			if t.AssigneeID != nil {
				var user model.User
				err := db.Pool.QueryRow(r.Context(),
					"SELECT id, email, name, role, avatar_url, created_at, updated_at FROM users WHERE id = $1",
					*t.AssigneeID,
				).Scan(&user.ID, &user.Email, &user.Name, &user.Role, &user.AvatarURL, &user.CreatedAt, &user.UpdatedAt)
				if err == nil {
					tasks[i].Assignee = &user
				}
			}
		}
	}

    if err := JSON(w, http.StatusOK, tasks); err != nil {
        Error(w, r, http.StatusInternalServerError, "failed to encode tasks", err, 0)
    }
}

// GetTask returns a single task by ID with full details.
func GetTask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")

	var t model.Task
	err := db.Pool.QueryRow(r.Context(),
		`SELECT id, title, description, status, priority, category, summary,
		        creator_id, assignee_id, due_date, estimated_hours, actual_hours,
		        created_at, updated_at
		 FROM tasks WHERE id = $1`, taskID,
	).Scan(
		&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority,
		&t.Category, &t.Summary, &t.CreatorID, &t.AssigneeID,
		&t.DueDate, &t.EstimatedHours, &t.ActualHours,
		&t.CreatedAt, &t.UpdatedAt,
	)

    if err == pgx.ErrNoRows {
        Error(w, r, http.StatusNotFound, "task not found", nil, 0)
        return
    }
    if err != nil {
        Error(w, r, http.StatusInternalServerError, "failed to get task", err, 0)
        return
    }

	// Load creator
	var creator model.User
	_ = db.Pool.QueryRow(r.Context(),
		"SELECT id, email, name, role, avatar_url, created_at, updated_at FROM users WHERE id = $1",
		t.CreatorID,
	).Scan(&creator.ID, &creator.Email, &creator.Name, &creator.Role, &creator.AvatarURL, &creator.CreatedAt, &creator.UpdatedAt)
	t.Creator = &creator

	// Load assignee
	if t.AssigneeID != nil {
		var assignee model.User
		_ = db.Pool.QueryRow(r.Context(),
			"SELECT id, email, name, role, avatar_url, created_at, updated_at FROM users WHERE id = $1",
			*t.AssigneeID,
		).Scan(&assignee.ID, &assignee.Email, &assignee.Name, &assignee.Role, &assignee.AvatarURL, &assignee.CreatedAt, &assignee.UpdatedAt)
		t.Assignee = &assignee
	}

	// Load tags
	tagRows, err := db.Pool.Query(r.Context(),
		`SELECT t.id, t.name, t.color, t.created_at
		 FROM tags t
		 INNER JOIN task_tags tt ON t.id = tt.tag_id
		 WHERE tt.task_id = $1`, t.ID)
	if err == nil {
		defer tagRows.Close()
		for tagRows.Next() {
			var tag model.Tag
			_ = tagRows.Scan(&tag.ID, &tag.Name, &tag.Color, &tag.CreatedAt)
			t.Tags = append(t.Tags, tag)
		}
	}

    if err := JSON(w, http.StatusOK, t); err != nil {
        Error(w, r, http.StatusInternalServerError, "failed to encode task", err, 0)
    }
}

// ClassifyTask calls the configured LLM to classify a task and persists results.
func ClassifyTask(w http.ResponseWriter, r *http.Request) {
    if llmClient == nil {
        Error(w, r, http.StatusInternalServerError, "LLM client not configured", nil, 0)
        return
    }

    taskID := chi.URLParam(r, "id")
    userID := middleware.GetUserID(r)
    if userID == "" {
        Error(w, r, http.StatusUnauthorized, "unauthorized", nil, 0)
        return
    }

    // Fetch task
    var t model.Task
    err := db.Pool.QueryRow(r.Context(),
        `SELECT id, title, COALESCE(description, '') FROM tasks WHERE id = $1`, taskID,
    ).Scan(&t.ID, &t.Title, &t.Description)
    if err == pgx.ErrNoRows {
        Error(w, r, http.StatusNotFound, "task not found", nil, 0)
        return
    }
    if err != nil {
        Error(w, r, http.StatusInternalServerError, "failed to get task", err, 0)
        return
    }

    desc := ""
    if t.Description != nil {
        desc = *t.Description
    }

    // Call LLM with timeout
    ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
    defer cancel()

    classification, err := llmClient.ClassifyTask(ctx, t.Title, desc)
    if err != nil {
        Error(w, r, http.StatusBadGateway, "llm classification failed", err, 0)
        return
    }

    // Normalize and validate
    // priority and category allowed values
    validPriorities := map[string]bool{"low":true, "medium":true, "high":true, "urgent":true}
    if !validPriorities[classification.Priority] {
        // fallback
        classification.Priority = "medium"
    }
    validCategories := map[string]bool{"bug":true, "feature":true, "improvement":true, "research":true}
    if !validCategories[classification.Category] {
        classification.Category = "feature"
    }
    if len(classification.Summary) > 140 {
        classification.Summary = classification.Summary[:137] + "..."
    }
    // sanitize tags
    tags := []string{}
    for i, tag := range classification.Tags {
        if i>=8 { break }
        t2 := strings.TrimSpace(strings.ToLower(tag))
        t2 = strings.ReplaceAll(t2, " ", "-")
        if t2=="" { continue }
        tags = append(tags, t2)
    }

    // Persist in transaction
    tx, err := db.Pool.Begin(r.Context())
    if err != nil {
        Error(w, r, http.StatusInternalServerError, "failed to begin tx", err, 0)
        return
    }
    defer tx.Rollback(r.Context())

    // Update task
    _, err = tx.Exec(r.Context(), `UPDATE tasks SET category=$1, priority=$2, summary=$3, updated_at=NOW() WHERE id=$4`,
        classification.Category, classification.Priority, classification.Summary, taskID)
    if err != nil {
        Error(w, r, http.StatusInternalServerError, "failed to update task", err, 0)
        return
    }

    // Ensure tags exist and collect ids
    tagIDs := []string{}
    for _, name := range tags {
        // insert tag if not exists
        _, _ = tx.Exec(r.Context(), `INSERT INTO tags (name, color) VALUES ($1, '#6B7280') ON CONFLICT (name) DO NOTHING`, name)
        var id string
        err := tx.QueryRow(r.Context(), `SELECT id FROM tags WHERE name=$1`, name).Scan(&id)
        if err != nil {
            Error(w, r, http.StatusInternalServerError, "failed to get tag id", err, 0)
            return
        }
        tagIDs = append(tagIDs, id)
    }

    // Remove previous AI-assigned tags for this task
    _, err = tx.Exec(r.Context(), `DELETE FROM task_tags WHERE task_id=$1 AND assigned_by='ai'`, taskID)
    if err != nil {
        Error(w, r, http.StatusInternalServerError, "failed to delete old task_tags", err, 0)
        return
    }

    // Insert new task_tags
    for _, tid := range tagIDs {
        _, err := tx.Exec(r.Context(), `INSERT INTO task_tags (task_id, tag_id, assigned_by) VALUES ($1,$2,'ai') ON CONFLICT DO NOTHING`, taskID, tid)
        if err != nil {
            Error(w, r, http.StatusInternalServerError, "failed to insert task_tag", err, 0)
            return
        }
    }

    // Record edit history for fields that changed
    // Fetch old values to compare
    var oldPriority, oldCategory, oldSummary *string
    _ = tx.QueryRow(r.Context(), `SELECT priority, category, summary FROM tasks WHERE id=$1`, taskID).Scan(&oldPriority, &oldCategory, &oldSummary)
    if oldPriority == nil || *oldPriority != classification.Priority {
        _, _ = tx.Exec(r.Context(), `INSERT INTO edit_history (task_id, user_id, field_name, old_value, new_value) VALUES ($1,$2,'priority',$3,$4)`, taskID, userID, oldPriority, classification.Priority+" (AI suggestion)")
    }
    if oldCategory == nil || *oldCategory != classification.Category {
        _, _ = tx.Exec(r.Context(), `INSERT INTO edit_history (task_id, user_id, field_name, old_value, new_value) VALUES ($1,$2,'category',$3,$4)`, taskID, userID, oldCategory, classification.Category+" (AI suggestion)")
    }
    if oldSummary == nil || *oldSummary != classification.Summary {
        _, _ = tx.Exec(r.Context(), `INSERT INTO edit_history (task_id, user_id, field_name, old_value, new_value) VALUES ($1,$2,'summary',$3,$4)`, taskID, userID, oldSummary, classification.Summary+" (AI suggestion)")
    }

    if err := tx.Commit(r.Context()); err != nil {
        Error(w, r, http.StatusInternalServerError, "failed to commit tx", err, 0)
        return
    }

    // Return updated task (reuse GetTask logic by calling DB again)
    GetTask(w, r)
}

// CreateTask creates a new task.
func CreateTask(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
    if userID == "" {
        Error(w, r, http.StatusUnauthorized, "unauthorized", nil, 0)
        return
    }

	var req model.CreateTaskRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        Error(w, r, http.StatusBadRequest, "invalid request body", err, 0)
        return
    }

    if req.Title == "" {
        Error(w, r, http.StatusBadRequest, "title is required", nil, 0)
        return
    }

    if len(req.Title) > 500 {
        Error(w, r, http.StatusBadRequest, "title too long", nil, 0)
        return
    }

	validStatuses := map[string]bool{"todo": true, "in_progress": true, "review": true, "done": true}
	if req.Status == "" {
		req.Status = "todo"
	}
    if !validStatuses[req.Status] {
        Error(w, r, http.StatusBadRequest, "invalid status", nil, 0)
        return
    }

	validPriorities := map[string]bool{"low": true, "medium": true, "high": true, "urgent": true}
	if req.Priority == "" {
		req.Priority = "medium"
	}
    if !validPriorities[req.Priority] {
        Error(w, r, http.StatusBadRequest, "invalid priority", nil, 0)
        return
    }

	// Validate assignee exists if provided
	if req.AssigneeID != nil && *req.AssigneeID != "" {
		var exists bool
		err := db.Pool.QueryRow(r.Context(),
			"SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", *req.AssigneeID,
		).Scan(&exists)
        if err != nil || !exists {
            Error(w, r, http.StatusBadRequest, "assignee not found", err, 0)
            return
        }
	}

	// Parse due date if provided
	var dueDate *time.Time
	if req.DueDate != nil && *req.DueDate != "" {
		parsed, err := time.Parse(time.RFC3339, *req.DueDate)
        if err != nil {
            Error(w, r, http.StatusBadRequest, "invalid due_date format, use RFC3339", err, 0)
            return
        }
		dueDate = &parsed
	}

	var task model.Task
	err := db.Pool.QueryRow(r.Context(),
		`INSERT INTO tasks (title, description, status, priority, creator_id, assignee_id, due_date, estimated_hours)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, title, description, status, priority, category, summary, creator_id, assignee_id, due_date, estimated_hours, actual_hours, created_at, updated_at`,
		req.Title, req.Description, req.Status, req.Priority, userID, req.AssigneeID, dueDate, req.EstimatedHours,
	).Scan(
		&task.ID, &task.Title, &task.Description, &task.Status, &task.Priority,
		&task.Category, &task.Summary, &task.CreatorID, &task.AssigneeID,
		&task.DueDate, &task.EstimatedHours, &task.ActualHours,
		&task.CreatedAt, &task.UpdatedAt,
	)

    if err != nil {
        log.Printf("error creating task: %v", err)
        Error(w, r, http.StatusInternalServerError, "failed to create task", err, 0)
        return
    }

    if err := JSON(w, http.StatusCreated, task); err != nil {
        Error(w, r, http.StatusInternalServerError, "failed to encode created task", err, 0)
    }
}

// UpdateTask updates an existing task.
func UpdateTask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")

	var req model.UpdateTaskRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        Error(w, r, http.StatusBadRequest, "invalid request body", err, 0)
        return
    }

	// Fetch current task state
	var existing model.Task
	err := db.Pool.QueryRow(r.Context(),
		`SELECT id, title, description, status, priority, creator_id, assignee_id,
		        due_date, estimated_hours, actual_hours
		 FROM tasks WHERE id = $1`, taskID,
	).Scan(
		&existing.ID, &existing.Title, &existing.Description, &existing.Status,
		&existing.Priority, &existing.CreatorID, &existing.AssigneeID,
		&existing.DueDate, &existing.EstimatedHours, &existing.ActualHours,
	)

    if err == pgx.ErrNoRows {
        Error(w, r, http.StatusNotFound, "task not found", nil, 0)
        return
    }
    if err != nil {
        Error(w, r, http.StatusInternalServerError, "failed to get task", err, 0)
        return
    }

	// Build update fields
	if req.Title != nil {
		existing.Title = *req.Title
	}
	if req.Description != nil {
		existing.Description = req.Description
	}
	if req.Status != nil {
		validStatuses := map[string]bool{"todo": true, "in_progress": true, "review": true, "done": true}
        if !validStatuses[*req.Status] {
            Error(w, r, http.StatusBadRequest, "invalid status", nil, 0)
            return
        }
		existing.Status = *req.Status
	}
	if req.Priority != nil {
		validPriorities := map[string]bool{"low": true, "medium": true, "high": true, "urgent": true}
        if !validPriorities[*req.Priority] {
            Error(w, r, http.StatusBadRequest, "invalid priority", nil, 0)
            return
        }
		existing.Priority = *req.Priority
	}
	if req.AssigneeID != nil {
		existing.AssigneeID = req.AssigneeID
	}
	if req.EstimatedHours != nil {
		existing.EstimatedHours = req.EstimatedHours
	}
	if req.ActualHours != nil {
		existing.ActualHours = req.ActualHours
	}

	_, _ = db.Pool.Exec(r.Context(),
		`UPDATE tasks SET title=$1, description=$2, status=$3, priority=$4,
		 assignee_id=$5, estimated_hours=$6, actual_hours=$7, updated_at=NOW()
		 WHERE id=$8`,
		existing.Title, existing.Description, existing.Status, existing.Priority,
		existing.AssigneeID, existing.EstimatedHours, existing.ActualHours,
		taskID,
	)

	// Record edit history
	userID := middleware.GetUserID(r)
	if req.Status != nil {
		_, _ = db.Pool.Exec(r.Context(),
			`INSERT INTO edit_history (task_id, user_id, field_name, old_value, new_value)
			 VALUES ($1, $2, 'status', $3, $4)`,
			taskID, userID, existing.Status, *req.Status,
		)
	}

	// Return updated task
	var updated model.Task
	err = db.Pool.QueryRow(r.Context(),
		`SELECT id, title, description, status, priority, category, summary,
		        creator_id, assignee_id, due_date, estimated_hours, actual_hours,
		        created_at, updated_at
		 FROM tasks WHERE id = $1`, taskID,
	).Scan(
		&updated.ID, &updated.Title, &updated.Description, &updated.Status,
		&updated.Priority, &updated.Category, &updated.Summary,
		&updated.CreatorID, &updated.AssigneeID,
		&updated.DueDate, &updated.EstimatedHours, &updated.ActualHours,
		&updated.CreatedAt, &updated.UpdatedAt,
	)

    if err != nil {
        Error(w, r, http.StatusInternalServerError, "failed to retrieve updated task", err, 0)
        return
    }

    if err := JSON(w, http.StatusOK, updated); err != nil {
        Error(w, r, http.StatusInternalServerError, "failed to encode updated task", err, 0)
    }
}

// DeleteTask deletes a task by ID.
func DeleteTask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")

	result, err := db.Pool.Exec(r.Context(),
		"DELETE FROM tasks WHERE id = $1", taskID,
	)

    if err != nil {
        Error(w, r, http.StatusInternalServerError, "failed to delete task", err, 0)
        return
    }

    if result.RowsAffected() == 0 {
        Error(w, r, http.StatusNotFound, "task not found", nil, 0)
        return
    }

	w.WriteHeader(http.StatusNoContent)
}

// GetTaskHistory returns the edit history for a task.
func GetTaskHistory(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")

    rows, err := db.Pool.Query(r.Context(),
        `SELECT id, task_id, user_id, field_name, old_value, new_value, edited_at
         FROM edit_history WHERE task_id = $1 ORDER BY edited_at DESC`, taskID)
    if err != nil {
        Error(w, r, http.StatusInternalServerError, "failed to get history", err, 0)
        return
    }
	defer rows.Close()

	history := []model.EditHistory{}
	for rows.Next() {
		var h model.EditHistory
		_ = rows.Scan(&h.ID, &h.TaskID, &h.UserID, &h.FieldName, &h.OldValue, &h.NewValue, &h.EditedAt)
		history = append(history, h)
	}

    if err := JSON(w, http.StatusOK, history); err != nil {
        Error(w, r, http.StatusInternalServerError, "failed to encode history", err, 0)
    }
}

// SearchTasks searches tasks by title or description.
func SearchTasks(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
    if q == "" {
        Error(w, r, http.StatusBadRequest, "query parameter q is required", nil, 0)
        return
    }

	searchTerm := "%" + strings.ToLower(q) + "%"

	rows, err := db.Pool.Query(r.Context(),
		`SELECT id, title, description, status, priority, category, summary,
		        creator_id, assignee_id, due_date, estimated_hours, actual_hours,
		        created_at, updated_at
		 FROM tasks
		 WHERE LOWER(title) LIKE $1 OR LOWER(COALESCE(description, '')) LIKE $1
		 ORDER BY created_at DESC`,
		searchTerm,
	)
    if err != nil {
        Error(w, r, http.StatusInternalServerError, "search failed", err, 0)
        return
    }
	defer rows.Close()

	tasks := []model.Task{}
	for rows.Next() {
		var t model.Task
		err := rows.Scan(
			&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority,
			&t.Category, &t.Summary, &t.CreatorID, &t.AssigneeID,
			&t.DueDate, &t.EstimatedHours, &t.ActualHours,
			&t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			continue
		}
		tasks = append(tasks, t)
	}

    if err := JSON(w, http.StatusOK, tasks); err != nil {
        Error(w, r, http.StatusInternalServerError, "failed to encode search results", err, 0)
    }
}

// GetDashboardStats returns summary statistics for the dashboard.
func GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	type Stats struct {
		TotalTasks   int            `json:"total_tasks"`
		ByStatus     map[string]int `json:"by_status"`
		ByPriority   map[string]int `json:"by_priority"`
		OverdueTasks int            `json:"overdue_tasks"`
	}

	var stats Stats
	stats.ByStatus = make(map[string]int)
	stats.ByPriority = make(map[string]int)

	// Total
	err := db.Pool.QueryRow(r.Context(), "SELECT COUNT(*) FROM tasks").Scan(&stats.TotalTasks)
    if err != nil {
        Error(w, r, http.StatusInternalServerError, "failed to get stats", err, 0)
        return
    }

	// By status
	rows, err := db.Pool.Query(r.Context(), "SELECT status, COUNT(*) FROM tasks GROUP BY status")
    if err != nil {
        Error(w, r, http.StatusInternalServerError, "failed to get stats", err, 0)
        return
    }
	defer rows.Close()
	for rows.Next() {
		var status string
		var count int
		_ = rows.Scan(&status, &count)
		stats.ByStatus[status] = count
	}

	// By priority
	rows2, err := db.Pool.Query(r.Context(), "SELECT priority, COUNT(*) FROM tasks GROUP BY priority")
    if err != nil {
        Error(w, r, http.StatusInternalServerError, "failed to get stats", err, 0)
        return
    }
	defer rows2.Close()
	for rows2.Next() {
		var priority string
		var count int
		_ = rows2.Scan(&priority, &count)
		stats.ByPriority[priority] = count
	}

	// Overdue
	_ = db.Pool.QueryRow(r.Context(),
		"SELECT COUNT(*) FROM tasks WHERE due_date < NOW() AND status != 'done'",
	).Scan(&stats.OverdueTasks)

    if err := JSON(w, http.StatusOK, stats); err != nil {
        Error(w, r, http.StatusInternalServerError, "failed to encode stats", err, 0)
    }
}

// LoginHandler handles user authentication and returns a JWT token.
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        Error(w, r, http.StatusBadRequest, "invalid request", err, 0)
        return
    }

	var user model.User
	err := db.Pool.QueryRow(r.Context(),
		"SELECT id, email, name, password_hash, role FROM users WHERE email = $1",
		req.Email,
	).Scan(&user.ID, &user.Email, &user.Name, &user.PasswordHash, &user.Role)

    if err != nil {
        Error(w, r, http.StatusUnauthorized, "invalid credentials", err, 0)
        return
    }

	// NOTE: In a real app, we'd use bcrypt.CompareHashAndPassword here.
	// For the assessment, we accept any password for seeded users.
	// This is intentional to simplify testing.
	_ = user.PasswordHash

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id": user.ID,
        "email":   user.Email,
        "role":    user.Role,
        "exp": time.Now().Add(24 * time.Hour).Unix(),
    })

    tokenString, err := token.SignedString(jwtSecret)
    if err != nil {
        Error(w, r, http.StatusInternalServerError, "failed to generate token", err, 0)
        return
    }

    if err := JSON(w, http.StatusOK, map[string]interface{}{
        "token": tokenString,
        "user": map[string]string{
            "id":    user.ID,
            "email": user.Email,
            "name":  user.Name,
            "role":  user.Role,
        },
    }); err != nil {
        Error(w, r, http.StatusInternalServerError, "failed to encode token response", err, 0)
    }
}
