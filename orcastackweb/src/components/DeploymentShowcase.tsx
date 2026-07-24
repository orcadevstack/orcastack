import { ArrowUpRight, GitCommitHorizontal, Rocket, ShieldCheck } from 'lucide-react';
import { Link } from 'react-router-dom';

import { auditPublicHub, type PublicDeployment } from '../api';
import { StatusPill } from './StatusPill';

type DeploymentShowcaseProps = {
  deployments: PublicDeployment[];
  viewerRole: string;
  onDeploy: () => void;
};

export function DeploymentShowcase({ deployments, viewerRole, onDeploy }: DeploymentShowcaseProps) {
  return (
    <section className="deployment-showcase" id="deployments">
      <div className="landing-section__heading">
        <div><span className="eyebrow">Delivery command</span><h2>Deployment evidence, not presentation data.</h2></div>
        <span className="viewer-badge"><ShieldCheck aria-hidden="true" size={15} /> {viewerRole} visibility</span>
      </div>
      {deployments.length === 0 ? (
        <div className="deployment-empty">
          <Rocket aria-hidden="true" size={28} />
          <div><strong>No public deployments recorded</strong><span>Deployment cards appear here when governed pipeline runs publish release records.</span></div>
          <button className="secondary-button" onClick={onDeploy} type="button">Open deployment workspace</button>
        </div>
      ) : (
        <div className="deployment-rail">
          {deployments.slice(0, 6).map((deployment) => {
            const content = (
              <>
                <div className="deployment-card__top"><span>{deployment.environment}</span><StatusPill value={deployment.status} /></div>
                <h3>{deployment.project_name}</h3>
                <p>{deployment.repository_name}</p>
                <dl>
                  <div><dt>Build</dt><dd><StatusPill value={deployment.build_status} /></dd></div>
                  <div><dt>Target</dt><dd>{deployment.target_kind}</dd></div>
                  <div><dt>Commit</dt><dd><GitCommitHorizontal aria-hidden="true" size={13} /> {deployment.commit_sha.slice(0, 8)}</dd></div>
                </dl>
                <span className="deployment-card__action">Open project <ArrowUpRight aria-hidden="true" size={15} /></span>
              </>
            );
            return deployment.can_access_dashboard ? (
              <Link className="deployment-card" key={deployment.id} onClick={() => void auditPublicHub('deployment', deployment.id)} to={deployment.dashboard_path}>{content}</Link>
            ) : (
              <button className="deployment-card" key={deployment.id} onClick={() => { void auditPublicHub('deployment', deployment.id); onDeploy(); }} type="button">{content}</button>
            );
          })}
        </div>
      )}
    </section>
  );
}
