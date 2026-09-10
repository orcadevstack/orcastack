import { Section } from '../components/DashboardSection';
import { StatusPill } from '../components/StatusPill';
import type { DashboardState } from '../types';

type AutomationPageProps = {
  dashboard: DashboardState;
};

export function AutomationPage({ dashboard }: AutomationPageProps) {
  const hardware = dashboard.hardwareWorkflows;
  const software = dashboard.softwareWorkflows;

  return (
    <>
      <section className="workspace-summary-grid">
        <article className="workspace-summary-tile">
          <span>Hardware workflows</span>
          <strong>{hardware.length}</strong>
        </article>
        <article className="workspace-summary-tile">
          <span>Software workflows</span>
          <strong>{software.length}</strong>
        </article>
        <article className="workspace-summary-tile">
          <span>Total automations</span>
          <strong>{hardware.length + software.length}</strong>
        </article>
      </section>

      <section className="dashboard-grid">
        <Section compact description="Device-oriented automation lanes, target pools, and execution state." title="Hardware automation">
          <ul className="stack-list">
            {hardware.map((workflow) => (
              <li key={workflow.id}>
                <div>
                  <strong>{workflow.name}</strong>
                  <span>{workflow.target_pool ?? 'hardware'} · {workflow.required_tags?.join(', ') || 'No required tags'}</span>
                </div>
                <StatusPill value={workflow.status} />
              </li>
            ))}
          </ul>
        </Section>

        <Section description="Software delivery lanes, integration hooks, and execution posture." title="Software automation">
          <ul className="stack-list">
            {software.map((workflow) => (
              <li key={workflow.id}>
                <div>
                  <strong>{workflow.name}</strong>
                  <span>{workflow.target ?? 'software'} · {workflow.integrations?.join(', ') || 'No integrations'}</span>
                </div>
                <StatusPill value={workflow.status} />
              </li>
            ))}
          </ul>
        </Section>
      </section>
    </>
  );
}