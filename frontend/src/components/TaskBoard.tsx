"use client";

import { Task } from '@/types';
import { TaskCard } from './TaskCard';

interface TaskBoardProps {
  tasks: Task[];
  onTaskUpdate: () => void;
}

const columns = [
  { key: 'todo', label: 'To Do', color: '#6b7280' },
  { key: 'in_progress', label: 'In Progress', color: '#3b82f6' },
  { key: 'review', label: 'Review', color: '#8b5cf6' },
  { key: 'done', label: 'Done', color: '#10b981' },
];

export function TaskBoard({ tasks }: TaskBoardProps) {
  const tasksByStatus = columns.map((col) => ({
    ...col,
    tasks: tasks.filter((t) => t.status === col.key),
  }));

  return (
    <div style={{
      display: 'grid',
      gridTemplateColumns: 'repeat(4, 1fr)',
      gap: '1rem',
      minHeight: '60vh',
    }}>
      {tasksByStatus.map((column) => (
        <div
          key={column.key}
          style={{
            backgroundColor: '#f3f4f6',
            borderRadius: '0.5rem',
            padding: '1rem',
          }}
        >
          <div style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: '1rem',
          }}>
            <h3 style={{
              margin: 0,
              fontSize: '0.875rem',
              fontWeight: 600,
              color: column.color,
            }}>
              {column.label}
            </h3>
            <span style={{
              backgroundColor: `${column.color}20`,
              color: column.color,
              padding: '0.125rem 0.5rem',
              borderRadius: '9999px',
              fontSize: '0.75rem',
              fontWeight: 600,
            }}>
              {column.tasks.length}
            </span>
          </div>

          <div>
            {column.tasks.map((task) => (
              <TaskCard key={task.id} task={task} />
            ))}
            {column.tasks.length === 0 && (
              <p style={{
                textAlign: 'center',
                color: '#9ca3af',
                fontSize: '0.875rem',
                padding: '2rem 0',
              }}>
                No tasks
              </p>
            )}
          </div>
        </div>
      ))}
    </div>
  );
}
