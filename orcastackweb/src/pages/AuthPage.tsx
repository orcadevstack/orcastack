import type React from 'react';
import { ArrowLeft, Layers3, ShieldCheck } from 'lucide-react';

import { AccessPanel } from '../components/AccessPanel';
import type { AuthMode } from '../types';

type AuthPageProps = {
  authMode: AuthMode;
  error: string | null;
  notice: string | null;
  loading: boolean;
  authChecking: boolean;
  loginForm: { username: string; password: string };
  signupForm: { username: string; email: string; password: string };
  onBack: () => void;
  onAuthModeChange: (mode: AuthMode) => void;
  onLoginFieldChange: (field: 'username' | 'password', value: string) => void;
  onSignupFieldChange: (field: 'username' | 'email' | 'password', value: string) => void;
  onLoginSubmit: (event: React.FormEvent<HTMLFormElement>) => void;
  onSignupSubmit: (event: React.FormEvent<HTMLFormElement>) => void;
};

export function AuthPage({ onBack, ...accessProps }: AuthPageProps) {
  return (
    <main className="auth-page">
      <header className="auth-page__header">
        <button className="auth-page__brand" onClick={onBack} type="button">
          <span className="brand-mark"><Layers3 aria-hidden="true" size={20} /></span>
          <strong>OrcaStack</strong>
        </button>
        <button className="secondary-button" onClick={onBack} type="button"><ArrowLeft aria-hidden="true" size={16} /> Back to platform</button>
      </header>
      <div className="auth-page__layout">
        <section className="auth-page__context">
          <span className="eyebrow">Governed workspace</span>
          <h1>Continue to the OrcaStack control plane.</h1>
          <p>Use your managed identity to access projects, delivery pipelines, deployments, and automation workflows.</p>
          <div className="auth-page__assurance"><ShieldCheck aria-hidden="true" size={19} /><span>Sessions and permissions are enforced by the platform RBAC service.</span></div>
        </section>
        <AccessPanel {...accessProps} />
      </div>
    </main>
  );
}