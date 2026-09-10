import type React from 'react';

import type { AuthMode } from '../types';

type AccessPanelProps = {
  authMode: AuthMode;
  error: string | null;
  notice: string | null;
  loading: boolean;
  authChecking: boolean;
  loginForm: { username: string; password: string };
  signupForm: { username: string; email: string; password: string };
  onAuthModeChange: (mode: AuthMode) => void;
  onLoginFieldChange: (field: 'username' | 'password', value: string) => void;
  onSignupFieldChange: (field: 'username' | 'email' | 'password', value: string) => void;
  onLoginSubmit: (event: React.FormEvent<HTMLFormElement>) => void;
  onSignupSubmit: (event: React.FormEvent<HTMLFormElement>) => void;
};

export function AccessPanel({
  authMode,
  error,
  notice,
  loading,
  authChecking,
  loginForm,
  signupForm,
  onAuthModeChange,
  onLoginFieldChange,
  onSignupFieldChange,
  onLoginSubmit,
  onSignupSubmit,
}: AccessPanelProps) {
  return (
    <section className="auth-card auth-card--landing">
      <div className="auth-card__header">
        <div>
          <span className="eyebrow">Workspace entry</span>
          <h2>{authMode === 'login' ? 'Sign in to the dashboard' : 'Request a managed account'}</h2>
        </div>
      </div>
      <div className="auth-toggle">
        <button className={authMode === 'login' ? 'auth-toggle__button auth-toggle__button--active' : 'auth-toggle__button'} onClick={() => onAuthModeChange('login')} type="button">
          Sign in
        </button>
        <button className={authMode === 'signup' ? 'auth-toggle__button auth-toggle__button--active' : 'auth-toggle__button'} onClick={() => onAuthModeChange('signup')} type="button">
          Request access
        </button>
      </div>
      {error ? <div className="banner banner--error">{error}</div> : null}
      {notice ? <div className="banner banner--info">{notice}</div> : null}
      {authMode === 'login' ? (
        <form className="auth-form" onSubmit={onLoginSubmit}>
          <label>
            <span>Username</span>
            <input value={loginForm.username} onChange={(event) => onLoginFieldChange('username', event.target.value)} />
          </label>
          <label>
            <span>Password</span>
            <input type="password" value={loginForm.password} onChange={(event) => onLoginFieldChange('password', event.target.value)} />
          </label>
          <button className="primary-button primary-button--warm" disabled={loading || authChecking} type="submit">
            {loading || authChecking ? 'Signing in...' : 'Sign in'}
          </button>
        </form>
      ) : (
        <form className="auth-form" onSubmit={onSignupSubmit}>
          <label>
            <span>Username</span>
            <input value={signupForm.username} onChange={(event) => onSignupFieldChange('username', event.target.value)} />
          </label>
          <label>
            <span>Email</span>
            <input type="email" value={signupForm.email} onChange={(event) => onSignupFieldChange('email', event.target.value)} />
          </label>
          <label>
            <span>Password</span>
            <input type="password" value={signupForm.password} onChange={(event) => onSignupFieldChange('password', event.target.value)} />
          </label>
          <button className="primary-button primary-button--warm" type="submit">Submit</button>
        </form>
      )}
    </section>
  );
}