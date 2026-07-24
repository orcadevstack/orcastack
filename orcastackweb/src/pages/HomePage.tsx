import { useEffect, useState } from 'react';
import { GitPullRequestArrow, Network, ShieldCheck, TerminalSquare } from 'lucide-react';
import { useNavigate } from 'react-router-dom';

import { fetchPublicHome, type PublicHome } from '../api';
import { DeploymentShowcase } from '../components/DeploymentShowcase';
import { DeveloperResourceDeck } from '../components/DeveloperResourceDeck';
import { HeroSection } from '../components/HeroSection';

type HomePageProps = {
  onLogin: () => void;
};

const capabilityLanes = [
  { title: 'Governed source', description: 'Repositories, reviews, teams, and policy share one ownership model.', icon: GitPullRequestArrow },
  { title: 'Continuous delivery', description: 'Build evidence moves through controlled environments with explicit status.', icon: Network },
  { title: 'Security in context', description: 'Identity, approvals, and audit records stay attached to every operation.', icon: ShieldCheck },
  { title: 'Programmable platform', description: 'APIs and SDKs extend software delivery and physical lab workflows.', icon: TerminalSquare },
];

export function HomePage(props: HomePageProps) {
  const navigate = useNavigate();
  const [home, setHome] = useState<PublicHome | null>(null);
  const [homeError, setHomeError] = useState<string | null>(null);

  useEffect(() => {
    const controller = new AbortController();
    void fetchPublicHome(controller.signal)
      .then((response) => { setHome(response); setHomeError(null); })
      .catch((error) => { if (!controller.signal.aborted) setHomeError(error instanceof Error ? error.message : 'Public platform data unavailable'); });
    return () => controller.abort();
  }, []);

  return (
    <>
      <HeroSection onCommunity={() => navigate('/community')} onDeploy={props.onLogin} onDocs={() => navigate('/docs')} />

      <section className="live-proof-strip" aria-label="Live public platform data">
        <div><span>Deployment records</span><strong>{home?.deployments.length ?? 0}</strong></div>
        <div><span>Community posts</span><strong>{home?.community.posts ?? 0}</strong></div>
        <div><span>Contributors</span><strong>{home?.community.contributors ?? 0}</strong></div>
        <div><span>Upcoming events</span><strong>{home?.community.events ?? 0}</strong></div>
      </section>
      {homeError ? <div className="public-data-notice">Live integration unavailable: {homeError}</div> : null}

      <DeploymentShowcase deployments={home?.deployments ?? []} onDeploy={props.onLogin} viewerRole={home?.viewer_role ?? 'viewer'} />

      <section className="platform-lanes">
        <div className="landing-section__heading"><div><span className="eyebrow">One operating model</span><h2>From first commit to controlled runtime.</h2></div><p>Each lane is a real platform boundary, connected by identity and auditability.</p></div>
        <div className="platform-lanes__grid">
          {capabilityLanes.map(({ title, description, icon: Icon }, index) => (
            <article key={title}><span>0{index + 1}</span><Icon aria-hidden="true" size={23} /><h3>{title}</h3><p>{description}</p></article>
          ))}
        </div>
      </section>

      <DeveloperResourceDeck />
    </>
  );
}
