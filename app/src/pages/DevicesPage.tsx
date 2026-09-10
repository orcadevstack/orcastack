import { DataTable, Section } from '../components/DashboardSection';
import { StatusPill } from '../components/StatusPill';
import type { DashboardState } from '../types';

type DevicesPageProps = {
  dashboard: DashboardState;
};

export function DevicesPage({ dashboard }: DevicesPageProps) {
  const devices = dashboard.devices?.devices ?? [];
  const healthyDevices = devices.filter((device) => device.health.toLowerCase().includes('healthy')).length;
  const assignedDevices = devices.filter((device) => device.assigned_run).length;

  return (
    <>
      <section className="workspace-summary-grid">
        <article className="workspace-summary-tile">
          <span>Total targets</span>
          <strong>{devices.length}</strong>
        </article>
        <article className="workspace-summary-tile">
          <span>Healthy devices</span>
          <strong>{healthyDevices}</strong>
        </article>
        <article className="workspace-summary-tile">
          <span>Assigned runs</span>
          <strong>{assignedDevices}</strong>
        </article>
      </section>

      <section className="dashboard-grid">
        <Section compact description="Every connected device target, its operating health, and location context." title="Device inventory">
          <DataTable>
            <table>
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Location</th>
                  <th>Health</th>
                  <th>Status</th>
                  <th>Capabilities</th>
                </tr>
              </thead>
              <tbody>
                {devices.map((device) => (
                  <tr key={device.id}>
                    <td>{device.name}</td>
                    <td>{device.location}</td>
                    <td>{device.health}</td>
                    <td><StatusPill value={device.status} /></td>
                    <td>{device.capabilities.join(', ') || 'Unavailable'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </DataTable>
        </Section>

        <Section description="Tagging and assignment state for device-backed validation lanes." title="Allocation and tags">
          <ul className="stack-list">
            {devices.map((device) => (
              <li key={device.id}>
                <div>
                  <strong>{device.name}</strong>
                  <span>{device.tags.join(', ') || 'No tags'} · {device.assigned_run ?? 'Unassigned'}</span>
                </div>
                <StatusPill value={device.status} />
              </li>
            ))}
          </ul>
        </Section>
      </section>
    </>
  );
}