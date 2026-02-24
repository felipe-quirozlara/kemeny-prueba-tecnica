"use client";

import { useEffect, useState } from 'react';
import { Task } from '@/types';
import { api } from '@/lib/api';
import { TaskBoard } from '@/components/TaskBoard';

export default function TasksPage() {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [statusFilter, setStatusFilter] = useState<string>('');

  useEffect(() => {
    loadTasks();
  }, [statusFilter]);

  async function loadTasks() {
    try {
      setLoading(true);
      const data = await api.getTasks(statusFilter || undefined, 'assignee');
      setTasks(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load tasks');
    } finally {
      setLoading(false);
    }
  }

  if (loading) {
    return <div style={{ textAlign: 'center', padding: '2rem' }}>Loading tasks...</div>;
  }

  if (error) {
    return (
      <div style={{ textAlign: 'center', padding: '2rem', color: '#ef4444' }}>
        Error: {error}
      </div>
    );
  }

  return (
    <div>
      <div style={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        marginBottom: '1.5rem',
      }}>
        <h2 style={{ margin: 0 }}>Tasks</h2>
        <div style={{ display: 'flex', gap: '0.5rem' }}>
          {['', 'todo', 'in_progress', 'review', 'done'].map((status) => (
            <button
              key={status}
              onClick={() => setStatusFilter(status)}
              style={{
                padding: '0.5rem 1rem',
                borderRadius: '0.375rem',
                border: 'none',
                cursor: 'pointer',
                backgroundColor: statusFilter === status ? '#3b82f6' : '#e5e7eb',
                color: statusFilter === status ? 'white' : '#374151',
                fontSize: '0.875rem',
              }}
            >
              {status === '' ? 'All' : status.replace('_', ' ')}
            </button>
          ))}
        </div>
      </div>
      <TaskBoard tasks={tasks} onTaskUpdate={loadTasks} />
    </div>
  );
}
