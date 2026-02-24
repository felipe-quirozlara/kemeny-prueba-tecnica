import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Task Manager',
  description: 'Project task management application',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body style={{ margin: 0, fontFamily: 'system-ui, -apple-system, sans-serif', backgroundColor: '#f9fafb' }}>
        <nav style={{
          backgroundColor: '#1f2937',
          color: 'white',
          padding: '1rem 2rem',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
        }}>
          <h1 style={{ margin: 0, fontSize: '1.25rem' }}>Task Manager</h1>
          <div style={{ display: 'flex', gap: '1.5rem' }}>
            <a href="/" style={{ color: '#d1d5db', textDecoration: 'none' }}>Dashboard</a>
            <a href="/tasks" style={{ color: '#d1d5db', textDecoration: 'none' }}>Tasks</a>
          </div>
        </nav>
        <main style={{ padding: '2rem', maxWidth: '1200px', margin: '0 auto' }}>
          {children}
        </main>
      </body>
    </html>
  );
}
