import { lazy, Suspense, useEffect, useRef, useState } from 'react';
import { Building2, ChevronDown, ExternalLink, FolderGit2, Layers3, Menu, Rocket, UserRound, X } from 'lucide-react';

import { auditHeaderInteraction, fetchHeaderNavigation, type AuthSession, type HeaderNavigation, type HeaderSection } from '../api';

const ProjectsDropdown = lazy(() => import('./header/ProjectsDropdown'));
const DeploymentsDropdown = lazy(() => import('./header/DeploymentsDropdown'));
const ProfileDropdown = lazy(() => import('./header/ProfileDropdown'));

const sections = [
  { id: 'projects' as const, label: 'Projects', icon: FolderGit2, component: ProjectsDropdown },
  { id: 'deployments' as const, label: 'Deployments', icon: Rocket, component: DeploymentsDropdown },
];

type EnterpriseHeaderProps = {
  authSession: AuthSession;
  title: string;
  summary: string;
  onLogout: () => void;
};

export function EnterpriseHeader({ authSession, title, summary, onLogout }: EnterpriseHeaderProps) {
  const [navigation, setNavigation] = useState<HeaderNavigation | null>(null);
  const [activeSection, setActiveSection] = useState<HeaderSection | null>(null);
  const [tripPanelOpen, setTripPanelOpen] = useState(false);
  const [profileMenuOpen, setProfileMenuOpen] = useState(false);
  const [loadError, setLoadError] = useState<string | null>(null);
  const headerRef = useRef<HTMLElement>(null);

  useEffect(() => {
    const controller = new AbortController();
    void fetchHeaderNavigation(controller.signal, authSession.token)
      .then((response) => {
        setNavigation(response);
        setLoadError(null);
      })
      .catch((error) => {
        if (!controller.signal.aborted) {
          setLoadError(error instanceof Error ? error.message : 'Header navigation unavailable');
        }
      });
    return () => controller.abort();
  }, [authSession.token]);

  useEffect(() => {
    function closeOnOutsidePointer(event: PointerEvent) {
      if (headerRef.current && !headerRef.current.contains(event.target as Node)) {
        setActiveSection(null);
        setProfileMenuOpen(false);
      }
    }
    document.addEventListener('pointerdown', closeOnOutsidePointer);
    return () => document.removeEventListener('pointerdown', closeOnOutsidePointer);
  }, []);

  function openSection(section: HeaderSection) {
    setActiveSection(section);
    setProfileMenuOpen(false);
    void auditHeaderInteraction('open', section, '', authSession.token).catch(() => undefined);
  }

  function navigateFromSection(section: HeaderSection, path: string) {
    setActiveSection(null);
    setTripPanelOpen(false);
    setProfileMenuOpen(false);
    void auditHeaderInteraction('navigate', section, path, authSession.token).catch(() => undefined);
  }

  function toggleProfileMenu() {
    const nextOpen = !profileMenuOpen;
    setProfileMenuOpen(nextOpen);
    setActiveSection(null);
    if (nextOpen) {
      void auditHeaderInteraction('open', 'profile', '', authSession.token).catch(() => undefined);
    }
  }

  async function handleProfileLogout() {
    setProfileMenuOpen(false);
    await auditHeaderInteraction('navigate', 'profile', '', authSession.token).catch(() => undefined);
    onLogout();
  }

  function renderDropdown(section: HeaderSection) {
    const definition = sections.find((candidate) => candidate.id === section);
    if (!definition) {
      return null;
    }
    const Dropdown = definition.component;
    const items = navigation?.items.filter((item) => item.section === section) ?? [];
    return (
      <Suspense fallback={<div className="enterprise-dropdown__loading">Loading live workspace data...</div>}>
        <Dropdown items={items} onNavigate={(path) => navigateFromSection(section, path)} />
      </Suspense>
    );
  }

  return (
    <header className="enterprise-header" onKeyDown={(event) => event.key === 'Escape' && (setActiveSection(null), setTripPanelOpen(false), setProfileMenuOpen(false))} ref={headerRef}>
      <div className="enterprise-header__primary">
        <a aria-label="OrcaStack dashboard" className="enterprise-brand" href="/app/dashboard">
          <span className="enterprise-brand__mark"><Layers3 aria-hidden="true" size={24} /></span>
          <span><strong>{navigation?.organization ?? 'OrcaStack'}</strong><small>{navigation?.tagline ?? 'Build. Secure. Ship. Operate.'}</small></span>
        </a>

        <nav aria-label="Dashboard modules" className="enterprise-header__nav" onMouseLeave={() => setActiveSection(null)}>
          {sections.map(({ id, label, icon: Icon }) => {
            const visible = navigation?.items.some((item) => item.section === id) ?? true;
            if (!visible) return null;
            return (
              <div className="enterprise-menu" key={id} onMouseEnter={() => openSection(id)}>
                <button aria-expanded={activeSection === id} aria-haspopup="menu" className={activeSection === id ? 'enterprise-menu__trigger enterprise-menu__trigger--active' : 'enterprise-menu__trigger'} onClick={() => activeSection === id ? setActiveSection(null) : openSection(id)} type="button">
                  <Icon aria-hidden="true" size={17} />
                  <span>{label}</span>
                  <ChevronDown aria-hidden="true" size={14} />
                </button>
                {activeSection === id ? <div className="enterprise-dropdown" role="menu">{renderDropdown(id)}</div> : null}
              </div>
            );
          })}
        </nav>

        <div className="enterprise-header__utility">
          <button aria-label="Open dashboard navigation" className="enterprise-trip-trigger" onClick={() => setTripPanelOpen(true)} type="button"><Menu aria-hidden="true" size={20} /><span>Menu</span></button>
          <div className="profile-menu">
            <button aria-expanded={profileMenuOpen} aria-haspopup="menu" aria-label="Open user profile menu" className={profileMenuOpen ? 'profile-menu__trigger profile-menu__trigger--active' : 'profile-menu__trigger'} onClick={toggleProfileMenu} title="User profile" type="button"><UserRound aria-hidden="true" size={20} /></button>
            {profileMenuOpen ? (
              <div className="profile-dropdown" role="menu">
                <div className="profile-dropdown__context"><span>Signed in</span><strong>{authSession.user.role}</strong></div>
                <Suspense fallback={<div className="enterprise-dropdown__loading">Loading profile actions...</div>}>
                  <ProfileDropdown items={navigation?.items.filter((item) => item.section === 'profile') ?? []} onLogout={handleProfileLogout} onNavigate={(path) => navigateFromSection('profile', path)} />
                </Suspense>
              </div>
            ) : null}
          </div>
        </div>
      </div>

      <div className="enterprise-header__context">
        <div><span className="eyebrow">Control center</span><h1>{title}</h1><p>{summary}</p></div>
        <div className="enterprise-header__links" aria-label="OrcaStack links">
          <a href="https://orcastack.org" rel="noreferrer" target="_blank">OrcaStack.org <ExternalLink aria-hidden="true" size={13} /></a>
          <a href="https://github.com/orcastack" rel="noreferrer" target="_blank">GitHub</a>
          <a href="mailto:support@orcastack.org">Support</a>
        </div>
      </div>
      {loadError ? <div className="enterprise-header__error" role="status">{loadError}</div> : null}

      {tripPanelOpen ? (
        <>
          <div aria-hidden="true" className="trip-panel-backdrop trip-panel-backdrop--open" onClick={() => setTripPanelOpen(false)} />
          <aside aria-label="Dashboard navigation panel" aria-modal="true" className="trip-panel trip-panel--open" role="dialog">
            <div className="trip-panel__header"><span><Building2 aria-hidden="true" size={18} /> Workspace navigation</span><button aria-label="Close navigation" onClick={() => setTripPanelOpen(false)} type="button"><X aria-hidden="true" size={20} /></button></div>
            <div className="trip-panel__triggers">
              {sections.map(({ id, label, icon: Icon }) => {
                const visible = navigation?.items.some((item) => item.section === id) ?? true;
                return visible ? <button aria-expanded={activeSection === id} className={activeSection === id ? 'trip-button trip-button--active' : 'trip-button'} key={id} onClick={() => activeSection === id ? setActiveSection(null) : openSection(id)} type="button"><Icon aria-hidden="true" size={18} /><span>{label}</span><ChevronDown aria-hidden="true" size={15} /></button> : null;
              })}
            </div>
            {activeSection ? <div className="trip-panel__content">{renderDropdown(activeSection)}</div> : null}
          </aside>
        </>
      ) : null}
    </header>
  );
}