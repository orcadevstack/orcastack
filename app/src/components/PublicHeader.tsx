import { Layers3 } from 'lucide-react';

import type { PublicRoute } from '../types';

type PublicHeaderProps = {
  currentPage: PublicRoute;
  onNavigate: (route: PublicRoute) => void;
  onLogin: () => void;
  onSignup: () => void;
};

export function PublicHeader({ currentPage, onNavigate, onLogin, onSignup }: PublicHeaderProps) {
  const routes: Array<{ route: PublicRoute; label: string }> = [
    { route: 'home', label: 'Home' },
    { route: 'docs', label: 'Docs' },
    { route: 'developer', label: 'Developer' },
    { route: 'community', label: 'Community' },
  ];

  return (
    <header className="landing-header">
      <div className="landing-header__left">
        <div className="brand-lockup">
          <span className="brand-mark"><Layers3 aria-hidden="true" size={20} /></span>
          <div>
            <span className="eyebrow">OrcaStack</span>
          </div>
        </div>
        <strong>
        <span className="brand-badge">DevSecOps Platform</span>
        </strong>
      </div>
      <nav className="landing-nav" aria-label="Primary">
        {routes.map((item) => (
          <button
            className={currentPage === item.route ? 'landing-nav__link landing-nav__link--active' : 'landing-nav__link'}
            key={item.route}
            onClick={() => onNavigate(item.route)}
            type="button"
          >
            {item.label}
          </button>
        ))}
      </nav>
      <div className="landing-actions">
        <button className="secondary-button" onClick={onLogin} type="button">
          Sign in
        </button>
        <button className="primary-button primary-button--warm" onClick={onSignup} type="button">
          Request access
        </button>
      </div>
    </header>
  );
}