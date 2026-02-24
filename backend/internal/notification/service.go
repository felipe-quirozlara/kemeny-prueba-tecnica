package notification

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/KemenyStudio/task-manager/internal/db"
)

type TaskNotification struct {
	TaskID     string
	TaskTitle  string
	AssigneeID string
	Email      string
	DueDate    time.Time
}

// GetUpcomingDeadlines finds tasks with deadlines approaching within the next 24 hours.
func GetUpcomingDeadlines(ctx context.Context) ([]TaskNotification, error) {
	now := time.Now()
	tomorrow := now.Add(24 * time.Hour)

	rows, err := db.Pool.Query(ctx,
		`SELECT t.id, t.title, t.assignee_id, u.email, t.due_date
		 FROM tasks t
		 JOIN users u ON t.assignee_id = u.id
		 WHERE t.due_date > $1
		   AND t.due_date < $2
		   AND t.status != 'done'
		   AND t.assignee_id IS NOT NULL`,
		now, tomorrow,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query upcoming deadlines: %w", err)
	}
	defer rows.Close()

	var notifications []TaskNotification
	for rows.Next() {
		var n TaskNotification
		if err := rows.Scan(&n.TaskID, &n.TaskTitle, &n.AssigneeID, &n.Email, &n.DueDate); err != nil {
			log.Printf("error scanning notification: %v", err)
			continue
		}
		notifications = append(notifications, n)
	}

	return notifications, nil
}

// SendDeadlineNotifications sends email notifications for upcoming deadlines.
// In a real app, this would integrate with SendGrid, SES, etc.
func SendDeadlineNotifications(ctx context.Context) error {
	notifications, err := GetUpcomingDeadlines(ctx)
	if err != nil {
		return err
	}

	for _, n := range notifications {
		// Simulate sending email
		log.Printf("NOTIFICATION: Task '%s' (ID: %s) is due on %s. Notifying %s",
			n.TaskTitle, n.TaskID, n.DueDate.Format(time.RFC3339), n.Email)

		// In production: send actual email via SendGrid/SES
		// sendEmail(n.Email, "Task Deadline Approaching", ...)
	}

	log.Printf("Sent %d deadline notifications", len(notifications))
	return nil
}
