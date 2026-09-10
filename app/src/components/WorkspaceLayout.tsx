import type React from 'react';

import type { AuthSession } from '../api';
import type { DashboardState } from '../types';
import { EnterpriseHeader } from './EnterpriseHeader';

type WorkspaceLayoutProps = {
  authSession: AuthSession;
  dashboard: DashboardState;
  title: string;
  summary: string;
  loading: boolean;
  error: string | null;
  onRefresh: () => void;
  onLogout: () => void;
  children: React.ReactNode;
};

export function WorkspaceLayout({
  authSession,
  dashboard,
  title,
  summary,
  loading,
  error,
  onRefresh,
  onLogout,
  children,
}: WorkspaceLayoutProps) {
  const pendingApprovals = dashboard.signupRequests.filter((request) => request.status === 'pending').length;

  return (
    <main className="dashboard-shell">
      <EnterpriseHeader authSession={authSession} onLogout={onLogout} summary={summary} title={title} />
      <div className="workspace-shell">
        <div className="workspace-main">
          <section aria-label="Live workspace status" className="workspace-toolbar">
            <div><span>Pending approvals</span><strong>{pendingApprovals}</strong></div>
            <div><span>Queued jobs</span><strong>{dashboard.runner?.summary.queued_jobs ?? 0}</strong></div>
            <button className="workspace-banner__refresh" disabled={loading} onClick={onRefresh} type="button">Refresh data</button>
          </section>

          {error ? <div className="banner banner--error">{error}</div> : null}
          {children}
          {loading ? <div className="banner banner--info">Loading live platform data...</div> : null}
        </div>
      </div>
    </main>
  );
}