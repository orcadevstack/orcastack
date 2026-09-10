import { Section } from '../components/DashboardSection';
import type { AuthSession } from '../api';
import type { DashboardState } from '../types';

type SettingsPageProps = {
  authSession: AuthSession;
  dashboard: DashboardState;
};

export function SettingsPage({ authSession, dashboard }: SettingsPageProps) {
  const overview = dashboard.overview;

  return (
    <>
      <section className="workspace-summary-grid">
        <article className="workspace-summary-tile">
          <span>Permissions</span>
          <strong>{authSession.user.permissions.length}</strong>
        </article>
        <article className="workspace-summary-tile">
          <span>RBAC realm</span>
          <strong>{authSession.user.rbac_realm}</strong>
        </article>
        <article className="workspace-summary-tile">
          <span>Identity</span>
          <strong>{authSession.user.identity}</strong>
        </article>
      </section>

      <section className="dashboard-grid">
        <Section compact description="User identity, access scope, and managed workspace session attributes." title="Access settings">
          <div className="settings-list">
            <div className="settings-list__item">
              <span>Name</span>
              <strong>{authSession.user.full_name || authSession.user.username}</strong>
            </div>
            <div className="settings-list__item">
              <span>Email</span>
              <strong>{authSession.user.email}</strong>
            </div>
            <div className="settings-list__item">
              <span>Role</span>
              <strong>{authSession.user.role}</strong>
            </div>
            <div className="settings-list__item">
              <span>Session expires</span>
              <strong>{authSession.expires_at}</strong>
            </div>
          </div>
        </Section>

        <Section description="Repository and directory controls currently shaping the platform security posture." title="Platform posture">
          <div className="settings-list">
            <div className="settings-list__item">
              <span>Repository identity</span>
              <strong>{overview?.security.repository_identity ?? 'Unavailable'}</strong>
            </div>
            <div className="settings-list__item">
              <span>UI process identity</span>
              <strong>{overview?.security.ui_process_identity ?? 'Unavailable'}</strong>
            </div>
            <div className="settings-list__item">
              <span>Directory verification</span>
              <strong>{overview?.security.directory.verified ? 'Verified' : 'Pending'}</strong>
            </div>
          </div>
        </Section>

        <Section description="Granted permissions exposed by the current authenticated workspace session." title="Permissions">
          <ul className="tag-list">
            {authSession.user.permissions.map((permission) => (
              <li key={permission}>{permission}</li>
            ))}
          </ul>
        </Section>
      </section>
    </>
  );
}