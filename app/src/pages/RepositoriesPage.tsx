import { DataTable, Section } from '../components/DashboardSection';
import { StatusPill } from '../components/StatusPill';
import type { DashboardState } from '../types';

type RepositoriesPageProps = {
  dashboard: DashboardState;
};

export function RepositoriesPage({ dashboard }: RepositoriesPageProps) {
  const overview = dashboard.overview;
  const repositories = overview?.repositories ?? [];
  const reviews = overview?.reviews ?? [];
  const verifiedCount = repositories.filter((repository) => repository.security.verified).length;

  return (
    <>
      <section className="workspace-summary-grid">
        <article className="workspace-summary-tile">
          <span>Repositories</span>
          <strong>{repositories.length}</strong>
        </article>
        <article className="workspace-summary-tile">
          <span>Verified projects</span>
          <strong>{verifiedCount}</strong>
        </article>
        <article className="workspace-summary-tile">
          <span>Open reviews</span>
          <strong>{reviews.length}</strong>
        </article>
      </section>

      <section className="dashboard-grid">
        <Section compact description="Governed repositories, default branches, identities, and review ownership." title="Repository inventory">
          <DataTable>
            <table>
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Branch</th>
                  <th>Reviewer</th>
                  <th>Identity</th>
                  <th>Status</th>
                </tr>
              </thead>
              <tbody>
                {repositories.map((repository) => (
                  <tr key={repository.id}>
                    <td>{repository.name}</td>
                    <td>{repository.default_branch}</td>
                    <td>{repository.reviewer}</td>
                    <td>{repository.identity}</td>
                    <td><StatusPill value={repository.security.verified ? 'verified' : 'pending'} /></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </DataTable>
        </Section>

        <Section description="Review activity and approval posture across projects." title="Review flow">
          <ul className="stack-list">
            {reviews.map((review) => (
              <li key={review.id}>
                <div>
                  <strong>{review.title}</strong>
                  <span>{review.approvals}/{review.required_approvals} approvals</span>
                </div>
                <StatusPill value={review.status} />
              </li>
            ))}
          </ul>
        </Section>
      </section>
    </>
  );
}