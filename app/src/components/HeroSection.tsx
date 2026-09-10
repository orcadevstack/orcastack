import { lazy, Suspense } from 'react';
import { ArrowRight, BookOpenText, Rocket, UsersRound } from 'lucide-react';

const PublicHeroScene = lazy(() => import('./PublicHeroScene').then((module) => ({ default: module.PublicHeroScene })));

type HeroSectionProps = {
  onDeploy: () => void;
  onCommunity: () => void;
  onDocs: () => void;
};

export function HeroSection({ onDeploy, onCommunity, onDocs }: HeroSectionProps) {
  return (
    <section className="public-hero" id="platform">
      <Suspense fallback={<div aria-hidden="true" className="public-hero-scene public-hero-scene--loading" />}><PublicHeroScene /></Suspense>
      <div className="public-hero__shade" />
      <div className="public-hero__content">
        <span className="public-hero__signal"><span /> Developer control plane</span>
        <h1>OrcaStack</h1>
        <p className="public-hero__tagline">Build. Secure. Ship. Operate.</p>
        <p className="public-hero__lede">One governed system for source, CI/CD, private-cloud delivery, and hardware-backed automation.</p>
        <div className="public-hero__actions">
          <button className="primary-button primary-button--warm" onClick={onDeploy} type="button"><Rocket aria-hidden="true" size={17} /> Deploy <ArrowRight aria-hidden="true" size={15} /></button>
          <button className="secondary-button" onClick={onCommunity} type="button"><UsersRound aria-hidden="true" size={17} /> Join Community</button>
          <button className="secondary-button secondary-button--ghost" onClick={onDocs} type="button"><BookOpenText aria-hidden="true" size={17} /> Explore Docs</button>
        </div>
        <div className="public-hero__proof" aria-label="Platform capabilities">
          <span><strong>Source</strong> governed repositories</span>
          <span><strong>Delivery</strong> policy-aware releases</span>
          <span><strong>Automation</strong> software and devices</span>
        </div>
      </div>
      <a className="public-hero__next" href="#deployments">Live platform integration <ArrowRight aria-hidden="true" size={15} /></a>
    </section>
  );
}
