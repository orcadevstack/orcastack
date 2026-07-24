import { DataTable, Section } from '../components/DashboardSection';
import { StatusPill } from '../components/StatusPill';
import type { DashboardState } from '../types';

export function DeploymentsPage({ dashboard }: { dashboard: DashboardState }) {
  const deployments = dashboard.overview?.deployments ?? [];
  const environments = new Set(deployments.map((deployment) => deployment.environment)).size;
  const verified = deployments.filter((deployment) => deployment.security.verified).length;

  return (
    <>
      <section className="workspace-summary-grid">
        <article className="workspace-summary-tile"><span>Active releases</span><strong>{deployments.length}</strong></article>
        <article className="workspace-summary-tile"><span>Environments</span><strong>{environments}</strong></article>
        <article className="workspace-summary-tile"><span>Verified artifacts</span><strong>{verified}</strong></article>
      </section>
      <Section compact description="Release state, target clusters, artifacts, and identity verification from the live control-plane inventory." title="Deployment environments">
        <DataTable>
          <table>
            <thead><tr><th>Service</th><th>Version</th><th>Environment</th><th>Cluster</th><th>Status</th><th>Security</th></tr></thead>
            <tbody>
              {deployments.map((deployment) => (
                <tr key={deployment.id}>
                  <td>{deployment.service_name}</td><td>{deployment.version}</td><td>{deployment.environment}</td><td>{deployment.cluster}</td>
                  <td><StatusPill value={deployment.status} /></td><td><StatusPill value={deployment.security.verified ? 'verified' : 'pending'} /></td>
                </tr>
              ))}
            </tbody>
          </table>
        </DataTable>
      </Section>
    </>
  );
}