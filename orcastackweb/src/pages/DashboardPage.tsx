import { useMemo } from 'react';

import type { SignupRequestRecord } from '../api';
import { DataTable, Section } from '../components/DashboardSection';
import { StatusPill } from '../components/StatusPill';
import type { DashboardState } from '../types';

type DashboardPageProps = {
  dashboard: DashboardState;
  isPlatformAdmin: boolean;
  reviewBusyId: string | null;
  onReview: (id: string, status: 'approved' | 'rejected') => Promise<void>;
};

function formatDateTime(value?: string) {
  if (!value) {
    return 'Unavailable';
  }

  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) {
    return value;
  }

  return new Intl.DateTimeFormat(undefined, {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(parsed);
}

export function DashboardPage({
  dashboard,
  isPlatformAdmin,
  reviewBusyId,
  onReview,
}: DashboardPageProps) {
  const summaryCards = useMemo(() => {
    const overview = dashboard.overview;
    const deviceSummary = dashboard.devices?.summary ?? {};

    return [
      { label: 'Repositories', value: String(overview?.repositories.length ?? 0) },
      { label: 'Pipelines', value: String(overview?.pipelines.length ?? 0) },
      { label: 'Devices', value: String(deviceSummary.total ?? dashboard.devices?.devices.length ?? 0) },
      { label: 'Deployments', value: String(overview?.deployments.length ?? 0) },
    ];
  }, [dashboard.devices, dashboard.overview]);

  const automationRows = useMemo(() => {
    return [
      ...dashboard.hardwareWorkflows.map((workflow) => ({
        id: workflow.id,
        name: workflow.name,
        scope: workflow.target_pool ?? 'hardware',
        status: workflow.status,
      })),
      ...dashboard.softwareWorkflows.map((workflow) => ({
        id: workflow.id,
        name: workflow.name,
        scope: workflow.target ?? 'software',
        status: workflow.status,
      })),
    ];
  }, [dashboard.hardwareWorkflows, dashboard.softwareWorkflows]);

  const overview = dashboard.overview;

  return (
    <>
      <section className="stats-grid">
        {summaryCards.map((card) => (
          <article className="stat-card stat-card--premium" key={card.label}>
            <span>{card.label}</span>
            <strong>{card.value}</strong>
          </article>
        ))}
      </section>

      <section className="dashboard-grid">
            <Section compact description="Live platform posture, identity, and core KPIs." title="Overview">
              <div className="overview-meta">
                <div>
                  <span>Updated</span>
                  <strong>{formatDateTime(overview?.updated_at)}</strong>
                </div>
                <div>
                  <span>Identity</span>
                  <strong>{overview?.security.repository_identity ?? 'Unavailable'}</strong>
                </div>
              </div>
              <div className="metric-list">
                {(overview?.metrics ?? []).slice(0, 4).map((metric) => (
                  <div className="metric-list__item" key={metric.label}>
                    <span>{metric.label}</span>
                    <strong>{metric.value}</strong>
                  </div>
                ))}
              </div>
            </Section>

            <Section description="Governed projects, default branches, and review assignment." title="Repositories">
              <DataTable>
                <table>
                  <thead>
                    <tr>
                      <th>Name</th>
                      <th>Branch</th>
                      <th>Reviewer</th>
                      <th>Status</th>
                    </tr>
                  </thead>
                  <tbody>
                    {(overview?.repositories ?? []).length === 0 ? <tr><td className="dashboard-empty" colSpan={4}>No repositories have been created or imported.</td></tr> : (overview?.repositories ?? []).map((repository) => (
                      <tr key={repository.id}>
                        <td>{repository.name}</td>
                        <td>{repository.default_branch}</td>
                        <td>{repository.reviewer}</td>
                        <td><StatusPill value={repository.security.verified ? 'verified' : 'pending'} /></td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </DataTable>
            </Section>

            <Section description="Current delivery runs, branch targets, and execution state." title="Pipelines">
              <DataTable>
                <table>
                  <thead>
                    <tr>
                      <th>Name</th>
                      <th>Branch</th>
                      <th>Status</th>
                      <th>Updated</th>
                    </tr>
                  </thead>
                  <tbody>
                    {(overview?.pipelines ?? []).length === 0 ? <tr><td className="dashboard-empty" colSpan={4}>No pipeline runs have been recorded.</td></tr> : (overview?.pipelines ?? []).map((pipeline) => (
                      <tr key={pipeline.id}>
                        <td>{pipeline.name}</td>
                        <td>{pipeline.branch}</td>
                        <td><StatusPill value={pipeline.status} /></td>
                        <td>{formatDateTime(pipeline.updated_at)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </DataTable>
            </Section>

            <Section description="Release posture by service, environment, and target cluster." title="Deployments">
              <DataTable>
                <table>
                  <thead>
                    <tr>
                      <th>Service</th>
                      <th>Environment</th>
                      <th>Status</th>
                      <th>Cluster</th>
                    </tr>
                  </thead>
                  <tbody>
                    {(overview?.deployments ?? []).length === 0 ? <tr><td className="dashboard-empty" colSpan={4}>No deployments have been recorded.</td></tr> : (overview?.deployments ?? []).map((deployment) => (
                      <tr key={deployment.id}>
                        <td>{deployment.service_name}</td>
                        <td>{deployment.environment}</td>
                        <td><StatusPill value={deployment.status} /></td>
                        <td>{deployment.cluster}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </DataTable>
            </Section>

            <Section description="Executor capacity, queue depth, and active pipeline execution." title="Runner fleet">
              <div className="runner-summary">
                <div>
                  <span>Capacity</span>
                  <strong>{dashboard.runner?.summary.capacity ?? 0}</strong>
                </div>
                <div>
                  <span>Busy</span>
                  <strong>{dashboard.runner?.summary.busy_executors ?? 0}</strong>
                </div>
                <div>
                  <span>Queued</span>
                  <strong>{dashboard.runner?.summary.queued_jobs ?? 0}</strong>
                </div>
              </div>
              <ul className="stack-list">
                {(dashboard.runner?.pipelines ?? []).length === 0 ? <li className="dashboard-empty">No runner pipelines are active.</li> : (dashboard.runner?.pipelines ?? []).map((pipeline) => (
                  <li key={pipeline.id}>
                    <div>
                      <strong>{pipeline.project}</strong>
                      <span>{pipeline.ref}</span>
                    </div>
                    <StatusPill value={pipeline.status} />
                  </li>
                ))}
              </ul>
            </Section>

            <Section description="Connected endpoints, locations, and device health." title="Devices">
              <ul className="stack-list">
                {(dashboard.devices?.devices ?? []).length === 0 ? <li className="dashboard-empty">No devices are registered.</li> : (dashboard.devices?.devices ?? []).map((device) => (
                  <li key={device.id}>
                    <div>
                      <strong>{device.name}</strong>
                      <span>{device.location}</span>
                    </div>
                    <StatusPill value={device.status} />
                  </li>
                ))}
              </ul>
            </Section>

            <Section description="Hardware and software workflows coordinated by one platform." title="Automation lanes">
              <ul className="stack-list">
                {automationRows.length === 0 ? <li className="dashboard-empty">No automation workflows are registered.</li> : automationRows.map((workflow) => (
                  <li key={workflow.id}>
                    <div>
                      <strong>{workflow.name}</strong>
                      <span>{workflow.scope}</span>
                    </div>
                    <StatusPill value={workflow.status} />
                  </li>
                ))}
              </ul>
            </Section>

            <Section description="Recent platform events, results, and affected components." title="Activity stream">
              <ul className="event-list">
                {(overview?.events ?? []).length === 0 ? <li className="dashboard-empty">No platform activity has been recorded.</li> : (overview?.events ?? []).slice(0, 6).map((event) => (
                  <li key={event.id}>
                    <div>
                      <strong>{event.action}</strong>
                      <span>{event.component}</span>
                    </div>
                    <div>
                      <StatusPill value={event.result} />
                      <time>{formatDateTime(event.time)}</time>
                    </div>
                  </li>
                ))}
              </ul>
            </Section>

            {isPlatformAdmin ? (
              <Section description="Pending managed-account requests awaiting administrative action." title="Access requests">
                <DataTable>
                  <table>
                    <thead>
                      <tr>
                        <th>User</th>
                        <th>Email</th>
                        <th>Status</th>
                        <th>Action</th>
                      </tr>
                    </thead>
                    <tbody>
                      {dashboard.signupRequests.length === 0 ? <tr><td className="dashboard-empty" colSpan={4}>No access requests are pending.</td></tr> : dashboard.signupRequests.map((request: SignupRequestRecord) => (
                        <tr key={request.id}>
                          <td>{request.username}</td>
                          <td>{request.email}</td>
                          <td><StatusPill value={request.status} /></td>
                          <td>
                            <div className="table-actions">
                              <button className="table-button" disabled={reviewBusyId === request.id || request.status !== 'pending'} onClick={() => void onReview(request.id, 'approved')} type="button">
                                Approve
                              </button>
                              <button className="table-button" disabled={reviewBusyId === request.id || request.status !== 'pending'} onClick={() => void onReview(request.id, 'rejected')} type="button">
                                Reject
                              </button>
                            </div>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </DataTable>
              </Section>
            ) : null}
      </section>
    </>
  );
}