"use client";

import { useEffect, useState } from 'react';
import { DashboardStats } from '@/types';
import { api } from '@/lib/api';

function validateStats(stats: DashboardStats | null): boolean {
  if (!stats) return false;
  if (typeof stats.total_tasks !== 'number') return false;
  if (!stats.by_status || typeof stats.by_status !== 'object') return false;
  if (!stats.by_priority || typeof stats.by_priority !== 'object') return false;
  if (typeof stats.overdue_tasks !== 'number') return false;
  return true;
}

const statusLabels: Record<string, string> = {
  todo: 'To Do',
  in_progress: 'In Progress',
  review: 'Review',
  done: 'Done',
};

const priorityLabels: Record<string, string> = {
  low: 'Low',
  medium: 'Medium',
  high: 'High',
  urgent: 'Urgent',
};

const statusColors: Record<string, string> = {
  todo: '#6b7280',
  in_progress: '#3b82f6',
  review: '#8b5cf6',
  done: '#10b981',
};

const priorityColors: Record<string, string> = {
  low: '#6b7280',
  medium: '#3b82f6',
  high: '#f59e0b',
  urgent: '#ef4444',
};

export function Dashboard() {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadStats();
  }, []);

  async function loadStats() {
    try {
      setLoading(true);
      const data = await api.getDashboardStats();

      if (!validateStats(data)) {
        throw new Error('Invalid dashboard data');
      }

      setStats(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load stats');
    } finally {
      setLoading(false);
    }
  }

  if (loading) {
    return <div style={{ textAlign: 'center', padding: '2rem' }}>Loading dashboard...</div>;
  }

  if (error) {
    return (
      <div style={{ textAlign: 'center', padding: '2rem', color: '#ef4444' }}>
        Error: {error}
      </div>
    );
  }

  if (!validateStats(stats)) {
    return (
      <div style={{ textAlign: 'center', padding: '2rem', color: '#ef4444' }}>
        Invalid dashboard data
      </div>
    );
  }

  // At this point stats is guaranteed non-null by validateStats, but TypeScript
  // doesn't know that — so we need the ! operator below. This is a code smell.

  return (
    <div>
      {/* Summary Cards */}
      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(4, 1fr)',
        gap: '1rem',
        marginBottom: '2rem',
      }}>
        <StatCard
          label="Total Tasks"
          value={stats!.total_tasks}
          color="#3b82f6"
        />
        <StatCard
          label="Completed"
          value={stats!.by_status['done'] || 0}
          color="#10b981"
        />
        <StatCard
          label="In Progress"
          value={(stats!.by_status['in_progress'] || 0) + (stats!.by_status['review'] || 0)}
          color="#8b5cf6"
        />
        <StatCard
          label="Overdue"
          value={stats!.overdue_tasks}
          color="#ef4444"
        />
      </div>

      {/* Status Breakdown */}
      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(2, 1fr)',
        gap: '1.5rem',
      }}>
        <div style={{
          backgroundColor: 'white',
          borderRadius: '0.5rem',
          padding: '1.5rem',
          boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
        }}>
          <h3 style={{ margin: '0 0 1rem', fontSize: '1rem', color: '#374151' }}>By Status</h3>
          {Object.entries(stats!.by_status).map(([status, count]) => (
            <div
              key={status}
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                padding: '0.5rem 0',
                borderBottom: '1px solid #f3f4f6',
              }}
            >
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <div style={{
                  width: '0.5rem',
                  height: '0.5rem',
                  borderRadius: '50%',
                  backgroundColor: statusColors[status] || '#6b7280',
                }} />
                <span style={{ fontSize: '0.875rem', color: '#374151' }}>
                  {statusLabels[status] || status}
                </span>
              </div>
              <span style={{ fontWeight: 600, color: '#111827' }}>{count}</span>
            </div>
          ))}
        </div>

        <div style={{
          backgroundColor: 'white',
          borderRadius: '0.5rem',
          padding: '1.5rem',
          boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
        }}>
          <h3 style={{ margin: '0 0 1rem', fontSize: '1rem', color: '#374151' }}>By Priority</h3>
          {Object.entries(stats!.by_priority).map(([priority, count]) => (
            <div
              key={priority}
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                padding: '0.5rem 0',
                borderBottom: '1px solid #f3f4f6',
              }}
            >
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <div style={{
                  width: '0.5rem',
                  height: '0.5rem',
                  borderRadius: '50%',
                  backgroundColor: priorityColors[priority] || '#6b7280',
                }} />
                <span style={{ fontSize: '0.875rem', color: '#374151' }}>
                  {priorityLabels[priority] || priority}
                </span>
              </div>
              <span style={{ fontWeight: 600, color: '#111827' }}>{count}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

function StatCard({ label, value, color }: { label: string; value: number; color: string }) {
  return (
    <div style={{
      backgroundColor: 'white',
      borderRadius: '0.5rem',
      padding: '1.5rem',
      boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
      borderLeft: `4px solid ${color}`,
    }}>
      <p style={{ margin: 0, fontSize: '0.875rem', color: '#6b7280' }}>{label}</p>
      <p style={{ margin: '0.25rem 0 0', fontSize: '2rem', fontWeight: 700, color: '#111827' }}>
        {value}
      </p>
    </div>
  );
}
