import { DataTable, Section } from '../components/DashboardSection';
import { StatusPill } from '../components/StatusPill';
import type { DashboardState } from '../types';

type PipelinesPageProps = {
  dashboard: DashboardState;
  onRunPipeline: () => void;
  pipelineBusy: boolean;
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

export function PipelinesPage({ dashboard, onRunPipeline, pipelineBusy }: PipelinesPageProps) {
  const pipelines = dashboard.overview?.pipelines ?? [];
  const runnerPipelines = dashboard.runner?.pipelines ?? [];

  return (
    <>
      <section className="workspace-summary-grid">
        <article className="workspace-summary-tile">
          <span>Tracked pipelines</span>
          <strong>{pipelines.length}</strong>
        </article>
        <article className="workspace-summary-tile">
          <span>Active runner queues</span>
          <strong>{dashboard.runner?.summary.queued_jobs ?? 0}</strong>
        </article>
        <article className="workspace-summary-tile">
          <span>Busy executors</span>
          <strong>{dashboard.runner?.summary.busy_executors ?? 0}</strong>
        </article>
      </section>

      <section className="dashboard-grid">
        <Section compact description="Pipeline state, branch target, and latest execution update." title="Pipeline catalogue">
          <DataTable>
            <table>
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Branch</th>
                  <th>Stages</th>
                  <th>Status</th>
                  <th>Updated</th>
                </tr>
              </thead>
              <tbody>
                {pipelines.map((pipeline) => (
                  <tr key={pipeline.id}>
                    <td>{pipeline.name}</td>
                    <td>{pipeline.branch}</td>
                    <td>{pipeline.stages.length}</td>
                    <td><StatusPill value={pipeline.status} /></td>
                    <td>{formatDateTime(pipeline.updated_at)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </DataTable>
        </Section>

        <Section
          actions={(
            <button className="primary-button primary-button--warm" disabled={pipelineBusy} onClick={onRunPipeline} type="button">
              {pipelineBusy ? 'Running tests...' : 'Run API tests'}
            </button>
          )}
          description="Runner fleet execution view with refs and current trigger path."
          title="Runner execution"
        >
          <ul className="stack-list">
            {runnerPipelines.map((pipeline) => (
              <li key={pipeline.id}>
                <div>
                  <strong>{pipeline.project}</strong>
                  <span>{pipeline.ref} · {pipeline.trigger}</span>
                </div>
                <StatusPill value={pipeline.status} />
              </li>
            ))}
          </ul>
        </Section>
      </section>
    </>
  );
}