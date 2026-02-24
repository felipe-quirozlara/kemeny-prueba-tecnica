"use client";

import { Task } from '@/types';

const priorityColors: Record<string, string> = {
  low: '#6b7280',
  medium: '#3b82f6',
  high: '#f59e0b',
  urgent: '#ef4444',
};

interface TaskCardProps {
  task: Task;
}

export function TaskCard({ task }: TaskCardProps) {
  const isOverdue = task.due_date && new Date(task.due_date) < new Date() && task.status !== 'done';

  return (
    <a
      href={`/tasks/${task.id}`}
      style={{
        display: 'block',
        backgroundColor: 'white',
        borderRadius: '0.5rem',
        padding: '1rem',
        marginBottom: '0.5rem',
        boxShadow: '0 1px 2px rgba(0,0,0,0.05)',
        textDecoration: 'none',
        color: 'inherit',
        border: isOverdue ? '1px solid #fca5a5' : '1px solid #e5e7eb',
      }}
    >
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '0.5rem' }}>
        <h4 style={{
          margin: 0,
          fontSize: '0.875rem',
          fontWeight: 600,
          color: '#111827',
          lineHeight: 1.4,
          flex: 1,
        }}>
          {task.title}
        </h4>
        <span style={{
          padding: '0.125rem 0.5rem',
          borderRadius: '9999px',
          fontSize: '0.625rem',
          fontWeight: 600,
          backgroundColor: `${priorityColors[task.priority]}15`,
          color: priorityColors[task.priority],
          marginLeft: '0.5rem',
          whiteSpace: 'nowrap',
        }}>
          {task.priority}
        </span>
      </div>

      {task.description && (
        <p style={{
          margin: '0 0 0.5rem',
          fontSize: '0.75rem',
          color: '#6b7280',
          lineHeight: 1.4,
          overflow: 'hidden',
          textOverflow: 'ellipsis',
          display: '-webkit-box',
          WebkitLineClamp: 2,
          WebkitBoxOrient: 'vertical',
        }}>
          {task.description}
        </p>
      )}

      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div style={{ display: 'flex', gap: '0.25rem', flexWrap: 'wrap' }}>
          {task.tags?.slice(0, 3).map((tag) => (
            <span
              key={tag.id}
              style={{
                padding: '0.125rem 0.375rem',
                borderRadius: '0.25rem',
                fontSize: '0.625rem',
                backgroundColor: `${tag.color}15`,
                color: tag.color,
              }}
            >
              {tag.name}
            </span>
          ))}
        </div>
        <div style={{ fontSize: '0.75rem', color: '#9ca3af' }}>
          {task.assignee?.name || 'Unassigned'}
        </div>
      </div>

      {isOverdue && (
        <div style={{
          marginTop: '0.5rem',
          fontSize: '0.625rem',
          color: '#ef4444',
          fontWeight: 500,
        }}>
          Overdue
        </div>
      )}
    </a>
  );
}
