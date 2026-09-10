import React, { useEffect, useState } from 'react';
import { Navigate, Route, Routes, useLocation, useNavigate } from 'react-router-dom';

import {
  createRunnerPipeline,
  fetchDevices,
  fetchHardwareWorkflows,
  fetchOverview,
  fetchRunnerPipelines,
  fetchSession,
  fetchSignupRequests,
  fetchSoftwareWorkflows,
  login,
  logout,
  reviewSignupRequest,
  signup,
  type AuthSession,
  type SignupRequestRecord,
} from './api';
import { PublicFooter } from './components/PublicFooter';
import { PublicHeader } from './components/PublicHeader';
import { WorkspaceLayout } from './components/WorkspaceLayout';
import { AutomationPage } from './pages/AutomationPage';
import { AccountsPage } from './pages/AccountsPage';
import { AuthPage } from './pages/AuthPage';
import { DevicesPage } from './pages/DevicesPage';
import { DashboardPage } from './pages/DashboardPage';
import { DeploymentsPage } from './pages/DeploymentsPage';
import { CommunityPage } from './pages/CommunityPage';
import { DeveloperPage } from './pages/DeveloperPage';
import { DocsPage } from './pages/DocsPage';
import { HomePage } from './pages/HomePage';
import { OrganizationsPage } from './pages/OrganizationsPage';
import { PipelinesPage } from './pages/PipelinesPage';
import { RepositoriesPage } from './pages/RepositoriesPage';
import { SettingsPage } from './pages/SettingsPage';
import { type AuthMode, type DashboardState, emptyDashboardState, type PublicRoute } from './types';

const authTokenStorageKey = 'orcastack.auth.token';

function readStoredAuthToken() {
  if (typeof window === 'undefined') {
    return null;
  }

  return window.localStorage.getItem(authTokenStorageKey);
}

function storeAuthToken(token: string | null) {
  if (typeof window === 'undefined') {
    return;
  }

  if (token) {
    window.localStorage.setItem(authTokenStorageKey, token);
    return;
  }

  window.localStorage.removeItem(authTokenStorageKey);
}

function pathToPublicRoute(pathname: string): PublicRoute {
  if (pathname === '/docs') {
    return 'docs';
  }

  if (pathname === '/developer') {
    return 'developer';
  }

  if (pathname === '/community') {
    return 'community';
  }

  return 'home';
}

export function App() {
  const [authToken, setAuthToken] = useState<string | null>(() => readStoredAuthToken());
  const [authSession, setAuthSession] = useState<AuthSession | null>(null);
  const [dashboard, setDashboard] = useState<DashboardState>(emptyDashboardState);
  const [loading, setLoading] = useState(() => readStoredAuthToken() !== null);
  const [authChecking, setAuthChecking] = useState(() => readStoredAuthToken() !== null);
  const [error, setError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);
  const [loginForm, setLoginForm] = useState({ username: 'admin', password: 'admin12345' });
  const [signupForm, setSignupForm] = useState({ username: '', email: '', password: '' });
  const [reviewBusyId, setReviewBusyId] = useState<string | null>(null);
  const [pipelineBusy, setPipelineBusy] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();

  const isPlatformAdmin = authSession?.user.role === 'platform-admin';
  const publicRoute = pathToPublicRoute(location.pathname);

  useEffect(() => {
    if (!authToken) {
      setAuthChecking(false);
      setAuthSession(null);
      setDashboard(emptyDashboardState);
      return;
    }

    const controller = new AbortController();
    setAuthChecking(true);

    void fetchSession(authToken, controller.signal)
      .then((session) => {
        setAuthSession({ ...session, token: authToken });
        setError(null);
      })
      .catch((sessionError) => {
        const message = sessionError instanceof Error ? sessionError.message : 'Session check failed';
        setError(message);
        setAuthSession(null);
        setAuthToken(null);
        storeAuthToken(null);
        setDashboard(emptyDashboardState);
      })
      .finally(() => setAuthChecking(false));

    return () => controller.abort();
  }, [authToken]);

  useEffect(() => {
    if (!authSession?.token) {
      return;
    }

    const controller = new AbortController();
    setLoading(true);

    const signupRequestsPromise = isPlatformAdmin
      ? fetchSignupRequests(authSession.token, controller.signal)
      : Promise.resolve<SignupRequestRecord[]>([]);

    void Promise.all([
      fetchOverview(controller.signal, authSession.token),
      fetchRunnerPipelines(controller.signal),
      fetchDevices(controller.signal),
      fetchHardwareWorkflows(controller.signal),
      fetchSoftwareWorkflows(controller.signal),
      signupRequestsPromise,
    ])
      .then(([overview, runner, devices, hardwareWorkflows, softwareWorkflows, signupRequests]) => {
        setDashboard({
          overview,
          runner,
          devices,
          hardwareWorkflows,
          softwareWorkflows,
          signupRequests,
        });
        setError(null);
      })
      .catch((loadError) => {
        const message = loadError instanceof Error ? loadError.message : 'Dashboard load failed';
        setError(message);
      })
      .finally(() => setLoading(false));

    return () => controller.abort();
  }, [authSession, isPlatformAdmin]);

  async function handleLogin(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setNotice(null);
    setLoading(true);

    try {
      const session = await login(loginForm.username.trim(), loginForm.password, undefined);
      if (!session.token) {
        throw new Error('Missing session token');
      }

      storeAuthToken(session.token);
      setAuthToken(session.token);
      setAuthSession(session);
      navigate('/app/dashboard', { replace: true });
    } catch (loginError) {
      const message = loginError instanceof Error ? loginError.message : 'Login failed';
      setError(message);
    } finally {
      setLoading(false);
    }
  }

  async function handleSignup(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setNotice(null);

    try {
      const result = await signup(signupForm, undefined);
      setNotice(result.message);
      setSignupForm({ username: '', email: '', password: '' });
      navigate('/login', { replace: true });
    } catch (signupError) {
      const message = signupError instanceof Error ? signupError.message : 'Signup failed';
      setError(message);
    }
  }

  async function handleLogout() {
    const token = authSession?.token ?? authToken;

    try {
      if (token) {
        await logout(token);
      }
    } catch {
      // Ignore logout transport failures and clear local state.
    }

    storeAuthToken(null);
    setAuthToken(null);
    setAuthSession(null);
    setDashboard(emptyDashboardState);
    setNotice(null);
    navigate('/', { replace: true });
  }

  async function handleReview(id: string, status: 'approved' | 'rejected') {
    if (!authSession?.token) {
      return;
    }

    setReviewBusyId(id);
    setError(null);

    try {
      const updated = await reviewSignupRequest(id, status, '', authSession.token);
      setDashboard((current) => ({
        ...current,
        signupRequests: current.signupRequests.map((request) => (request.id === id ? updated : request)),
      }));
    } catch (reviewError) {
      const message = reviewError instanceof Error ? reviewError.message : 'Review update failed';
      setError(message);
    } finally {
      setReviewBusyId(null);
    }
  }

  async function handleRunPipeline() {
    setPipelineBusy(true);
    setError(null);

    try {
      const pipeline = await createRunnerPipeline();
      setDashboard((current) => ({
        ...current,
        runner: current.runner
          ? {
              ...current.runner,
              pipelines: [pipeline, ...current.runner.pipelines],
              summary: {
                ...current.runner.summary,
                queued_jobs: current.runner.summary.queued_jobs + 1,
              },
            }
          : current.runner,
      }));

      for (let attempt = 0; attempt < 20; attempt += 1) {
        await new Promise((resolve) => window.setTimeout(resolve, 500));
        const runner = await fetchRunnerPipelines();
        setDashboard((current) => ({ ...current, runner }));

        const updated = runner.pipelines.find((candidate) => candidate.id === pipeline.id);
        if (updated && updated.status !== 'queued' && updated.status !== 'running') {
          break;
        }
      }
    } catch (pipelineError) {
      const message = pipelineError instanceof Error ? pipelineError.message : 'Pipeline execution failed';
      setError(message);
    } finally {
      setPipelineBusy(false);
    }
  }

  function navigatePublic(route: PublicRoute) {
    const nextPath = route === 'home' ? '/' : `/${route}`;
    navigate(nextPath);
    window.scrollTo({ top: 0, behavior: 'smooth' });
  }

  function openAuthPage(mode: AuthMode) {
    setError(null);
    setNotice(null);
    navigate(mode === 'login' ? '/login' : '/request-access');
  }

  if (authToken && authChecking && !authSession) {
    return (
      <main className="public-shell public-shell--loading">
        <section className="loading-card">
          <span className="eyebrow">ORCASTACK</span>
          <h1>Restoring your platform session</h1>
          <p>Checking access and loading your workspace.</p>
        </section>
      </main>
    );
  }

  if (!authToken || !authSession) {
    return (
      <Routes>
        <Route
          path="/"
          element={
            <main className="public-shell">
              <PublicHeader currentPage="home" onLogin={() => openAuthPage('login')} onNavigate={navigatePublic} onSignup={() => openAuthPage('signup')} />
              <HomePage onLogin={() => openAuthPage('login')} />
              <PublicFooter />
            </main>
          }
        />
        <Route
          path="/docs"
          element={
            <main className="public-shell">
              <PublicHeader currentPage="docs" onLogin={() => openAuthPage('login')} onNavigate={navigatePublic} onSignup={() => openAuthPage('signup')} />
              <DocsPage onLogin={() => openAuthPage('login')} onSignup={() => openAuthPage('signup')} />
              <PublicFooter />
            </main>
          }
        />
        <Route
          path="/developer"
          element={
            <main className="public-shell">
              <PublicHeader currentPage="developer" onLogin={() => openAuthPage('login')} onNavigate={navigatePublic} onSignup={() => openAuthPage('signup')} />
              <DeveloperPage onLogin={() => openAuthPage('login')} onSignup={() => openAuthPage('signup')} />
              <PublicFooter />
            </main>
          }
        />
        <Route
          path="/community"
          element={
            <main className="public-shell">
              <PublicHeader currentPage="community" onLogin={() => openAuthPage('login')} onNavigate={navigatePublic} onSignup={() => openAuthPage('signup')} />
              <CommunityPage onLogin={() => openAuthPage('login')} onSignup={() => openAuthPage('signup')} />
              <PublicFooter />
            </main>
          }
        />
        <Route
          path="/login"
          element={
            <AuthPage
              authChecking={authChecking}
              authMode="login"
              error={error}
              loading={loading}
              loginForm={loginForm}
              notice={notice}
              onAuthModeChange={openAuthPage}
              onBack={() => navigate('/')}
              onLoginFieldChange={(field, value) => setLoginForm((current) => ({ ...current, [field]: value }))}
              onLoginSubmit={handleLogin}
              onSignupFieldChange={(field, value) => setSignupForm((current) => ({ ...current, [field]: value }))}
              onSignupSubmit={handleSignup}
              signupForm={signupForm}
            />
          }
        />
        <Route
          path="/request-access"
          element={
            <AuthPage
              authChecking={authChecking}
              authMode="signup"
              error={error}
              loading={loading}
              loginForm={loginForm}
              notice={notice}
              onAuthModeChange={openAuthPage}
              onBack={() => navigate('/')}
              onLoginFieldChange={(field, value) => setLoginForm((current) => ({ ...current, [field]: value }))}
              onLoginSubmit={handleLogin}
              onSignupFieldChange={(field, value) => setSignupForm((current) => ({ ...current, [field]: value }))}
              onSignupSubmit={handleSignup}
              signupForm={signupForm}
            />
          }
        />
        <Route path="/app/*" element={<Navigate replace to="/" />} />
        <Route path="*" element={<Navigate replace to="/" />} />
      </Routes>
    );
  }

  return (
    <Routes>
      <Route path="/" element={<Navigate replace to="/app/dashboard" />} />
      <Route path="/docs" element={<Navigate replace to="/app/dashboard" />} />
      <Route path="/developer" element={<Navigate replace to="/app/dashboard" />} />
      <Route path="/community" element={<Navigate replace to="/app/dashboard" />} />
      <Route
        path="/app/dashboard"
        element={
          <WorkspaceLayout
            authSession={authSession}
            dashboard={dashboard}
            error={error}
            loading={loading}
            onLogout={handleLogout}
            onRefresh={() => window.location.reload()}
            summary="Monitor repositories, CI/CD execution, deployments, labs, and approval workflows from one production-grade engineering surface."
            title="Platform dashboard"
          >
            <DashboardPage dashboard={dashboard} isPlatformAdmin={isPlatformAdmin} onReview={handleReview} reviewBusyId={reviewBusyId} />
          </WorkspaceLayout>
        }
      />
      <Route
        path="/app/organizations"
        element={
          <WorkspaceLayout
            authSession={authSession}
            dashboard={dashboard}
            error={error}
            loading={loading}
            onLogout={handleLogout}
            onRefresh={() => window.location.reload()}
            summary="Group projects under organizations, assign teams and roles, provision repositories, and inspect audited engineering activity."
            title="Organizations"
          >
            <OrganizationsPage authSession={authSession} />
          </WorkspaceLayout>
        }
      />
      <Route
        path="/app/repositories"
        element={
          <WorkspaceLayout
            authSession={authSession}
            dashboard={dashboard}
            error={error}
            loading={loading}
            onLogout={handleLogout}
            onRefresh={() => window.location.reload()}
            summary="Review governed repositories, approval flow, and identity posture with the same structured workspace shell used across the platform."
            title="Repositories"
          >
            <RepositoriesPage dashboard={dashboard} />
          </WorkspaceLayout>
        }
      />
      <Route
        path="/app/pipelines"
        element={
          <WorkspaceLayout
            authSession={authSession}
            dashboard={dashboard}
            error={error}
            loading={loading}
            onLogout={handleLogout}
            onRefresh={() => window.location.reload()}
            summary="Track pipeline inventory, execution state, and runner utilization through dedicated route-based navigation."
            title="Pipelines"
          >
            <PipelinesPage dashboard={dashboard} onRunPipeline={handleRunPipeline} pipelineBusy={pipelineBusy} />
          </WorkspaceLayout>
        }
      />
      <Route
        path="/app/devices"
        element={
          <WorkspaceLayout
            authSession={authSession}
            dashboard={dashboard}
            error={error}
            loading={loading}
            onLogout={handleLogout}
            onRefresh={() => window.location.reload()}
            summary="Inspect every connected device target, assignment state, and hardware health from a dedicated internal workspace page."
            title="Devices"
          >
            <DevicesPage dashboard={dashboard} />
          </WorkspaceLayout>
        }
      />
      <Route
        path="/app/automation"
        element={
          <WorkspaceLayout
            authSession={authSession}
            dashboard={dashboard}
            error={error}
            loading={loading}
            onLogout={handleLogout}
            onRefresh={() => window.location.reload()}
            summary="Coordinate hardware and software automation lanes with one consistent navigation, layout, and card system."
            title="Automation"
          >
            <AutomationPage dashboard={dashboard} />
          </WorkspaceLayout>
        }
      />
      <Route
        path="/app/deployments"
        element={
          <WorkspaceLayout
            authSession={authSession}
            dashboard={dashboard}
            error={error}
            loading={loading}
            onLogout={handleLogout}
            onRefresh={() => window.location.reload()}
            summary="Track controlled releases, target environments, artifact identity, and deployment health across the platform."
            title="Deployments"
          >
            <DeploymentsPage dashboard={dashboard} />
          </WorkspaceLayout>
        }
      />
      <Route
        path="/app/accounts"
        element={
          <WorkspaceLayout
            authSession={authSession}
            dashboard={dashboard}
            error={error}
            loading={loading}
            onLogout={handleLogout}
            onRefresh={() => window.location.reload()}
            summary="Review your authenticated identity, RBAC scope, session lifetime, and authorized account governance."
            title="Accounts"
          >
            <AccountsPage authSession={authSession} dashboard={dashboard} isPlatformAdmin={isPlatformAdmin} />
          </WorkspaceLayout>
        }
      />
      <Route
        path="/app/settings"
        element={
          <WorkspaceLayout
            authSession={authSession}
            dashboard={dashboard}
            error={error}
            loading={loading}
            onLogout={handleLogout}
            onRefresh={() => window.location.reload()}
            summary="Manage access posture, identity, and workspace-level settings using the same design system as the rest of the platform."
            title="Settings"
          >
            <SettingsPage authSession={authSession} dashboard={dashboard} />
          </WorkspaceLayout>
        }
      />
      <Route path="/app" element={<Navigate replace to="/app/dashboard" />} />
      <Route path="*" element={<Navigate replace to="/app/dashboard" />} />
    </Routes>
  );
}