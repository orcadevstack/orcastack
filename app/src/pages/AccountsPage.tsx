import { Section } from '../components/DashboardSection';
import type { AuthSession } from '../api';
import type { DashboardState } from '../types';

type AccountsPageProps = { authSession: AuthSession; dashboard: DashboardState; isPlatformAdmin: boolean };

export function AccountsPage({ authSession, dashboard, isPlatformAdmin }: AccountsPageProps) {
  const pendingRequests = dashboard.signupRequests.filter((request) => request.status === 'pending');
  return (
    <>
      <section className="workspace-summary-grid">
        <article className="workspace-summary-tile"><span>Role</span><strong>{authSession.user.role}</strong></article>
        <article className="workspace-summary-tile"><span>Granted permissions</span><strong>{authSession.user.permissions.length}</strong></article>
        <article className="workspace-summary-tile"><span>Pending access</span><strong>{isPlatformAdmin ? pendingRequests.length : 'Restricted'}</strong></article>
      </section>
      <section className="dashboard-grid">
        <Section compact description="Session-aware identity data returned by the authenticated gateway." title="Account identity">
          <div className="settings-list" id="profile">
            <div className="settings-list__item"><span>Name</span><strong>{authSession.user.full_name || authSession.user.username}</strong></div>
            <div className="settings-list__item"><span>Email</span><strong>{authSession.user.email}</strong></div>
            <div className="settings-list__item"><span>RBAC realm</span><strong>{authSession.user.rbac_realm}</strong></div>
            <div className="settings-list__item"><span>Session expires</span><strong>{authSession.expires_at}</strong></div>
          </div>
        </Section>
        <Section description="Recent platform activity available to your authenticated workspace session." title="Notifications">
          <ul className="event-list" id="notifications">
            {(dashboard.overview?.events ?? []).slice(0, 5).map((event) => (
              <li key={event.id}><div><strong>{event.summary}</strong><span>{event.component} · {event.result}</span></div><time>{event.time}</time></li>
            ))}
          </ul>
        </Section>
        <Section description="Access request visibility is restricted to platform administrators by the server-issued session role." title="Access governance">
          {isPlatformAdmin ? (
            <ul className="stack-list" id="access-requests">
              {pendingRequests.length > 0 ? pendingRequests.map((request) => <li key={request.id}><div><strong>{request.username}</strong><span>{request.email}</span></div><span>{request.status}</span></li>) : <li><div><strong>No pending requests</strong><span>The access queue is clear.</span></div></li>}
            </ul>
          ) : <p className="restricted-copy">Administrator permission is required to view workspace access requests.</p>}
        </Section>
      </section>
    </>
  );
}