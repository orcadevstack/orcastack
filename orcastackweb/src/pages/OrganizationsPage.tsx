import { useEffect, useRef, useState, type FormEvent } from 'react';
import { Activity, Building2, ExternalLink, FolderGit2, Plus, Settings2, UsersRound } from 'lucide-react';

import {
  addOrganizationMember,
  createOrganization,
  createOrganizationProject,
  createOrganizationTeam,
  fetchOrganization,
  fetchOrganizations,
  type AuthSession,
  type OrganizationDetail,
  type OrganizationRole,
  type OrganizationSummary,
} from '../api';
import { Section } from '../components/DashboardSection';
import { StatusPill } from '../components/StatusPill';

type OrganizationTab = 'overview' | 'teams' | 'people' | 'insights' | 'settings';

const tabs: { id: OrganizationTab; label: string; icon: typeof Building2 }[] = [
  { id: 'overview', label: 'Overview', icon: Building2 },
  { id: 'teams', label: 'Teams', icon: UsersRound },
  { id: 'people', label: 'People', icon: UsersRound },
  { id: 'insights', label: 'Insights', icon: Activity },
  { id: 'settings', label: 'Settings', icon: Settings2 },
];

function roleAllows(role: OrganizationRole, required: OrganizationRole) {
  const rank = { developer: 1, maintainer: 2, owner: 3 };
  return rank[role] >= rank[required];
}

export function OrganizationsPage({ authSession }: { authSession: AuthSession }) {
  const [organizations, setOrganizations] = useState<OrganizationSummary[]>([]);
  const [selectedSlug, setSelectedSlug] = useState('');
  const [detail, setDetail] = useState<OrganizationDetail | null>(null);
  const [activeTab, setActiveTab] = useState<OrganizationTab>('overview');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [organizationForm, setOrganizationForm] = useState({ name: '', slug: '', description: '', website: '' });
  const [teamForm, setTeamForm] = useState({ name: '', description: '' });
  const [memberForm, setMemberForm] = useState({ username: '', role: 'developer', team_id: '' });
  const [projectForm, setProjectForm] = useState({ name: '', description: '', repository_name: '', team_id: '', branch_strategy: 'feature', default_branch: 'main' });
  const selectionController = useRef<AbortController | null>(null);

  async function loadCatalog(preferredSlug?: string) {
    const catalog = await fetchOrganizations(undefined, authSession.token);
    setOrganizations(catalog);
    const nextSlug = preferredSlug || selectedSlug || catalog[0]?.slug || '';
    setSelectedSlug(nextSlug);
    if (nextSlug) {
      setDetail(await fetchOrganization(nextSlug, undefined, authSession.token));
    } else {
      setDetail(null);
    }
  }

  useEffect(() => {
    const controller = new AbortController();
    void fetchOrganizations(controller.signal, authSession.token)
      .then(async (catalog) => {
        setOrganizations(catalog);
        const slug = catalog[0]?.slug ?? '';
        setSelectedSlug(slug);
        if (slug) setDetail(await fetchOrganization(slug, controller.signal, authSession.token));
      })
      .catch((loadError) => {
        if (!controller.signal.aborted) setError(loadError instanceof Error ? loadError.message : 'Organization load failed');
      });
    return () => controller.abort();
  }, [authSession.token]);

  async function runMutation(operation: () => Promise<string | void>, reset: () => void) {
    setBusy(true);
    setError(null);
    try {
      const preferredSlug = await operation();
      reset();
      await loadCatalog(typeof preferredSlug === 'string' ? preferredSlug : undefined);
    } catch (mutationError) {
      setError(mutationError instanceof Error ? mutationError.message : 'Organization update failed');
    } finally {
      setBusy(false);
    }
  }

  async function selectOrganization(slug: string) {
    selectionController.current?.abort();
    const controller = new AbortController();
    selectionController.current = controller;
    setSelectedSlug(slug);
    setError(null);
    try {
      setDetail(await fetchOrganization(slug, controller.signal, authSession.token));
    } catch (loadError) {
      if (!controller.signal.aborted) setError(loadError instanceof Error ? loadError.message : 'Organization load failed');
    }
  }

  function submitOrganization(event: FormEvent) {
    event.preventDefault();
    void runMutation(async () => {
      const created = await createOrganization(organizationForm, authSession.token);
      return created.slug;
    }, () => setOrganizationForm({ name: '', slug: '', description: '', website: '' }));
  }

  function submitTeam(event: FormEvent) {
    event.preventDefault();
    if (!detail) return;
    void runMutation(async () => { await createOrganizationTeam(detail.organization.slug, teamForm, authSession.token); }, () => setTeamForm({ name: '', description: '' }));
  }

  function submitMember(event: FormEvent) {
    event.preventDefault();
    if (!detail) return;
    void runMutation(async () => { await addOrganizationMember(detail.organization.slug, memberForm, authSession.token); }, () => setMemberForm({ username: '', role: 'developer', team_id: '' }));
  }

  function submitProject(event: FormEvent) {
    event.preventDefault();
    if (!detail) return;
    void runMutation(async () => { await createOrganizationProject(detail.organization.slug, projectForm, authSession.token); }, () => setProjectForm({ name: '', description: '', repository_name: '', team_id: '', branch_strategy: 'feature', default_branch: 'main' }));
  }

  const canCreateOrganization = authSession.user.role === 'platform-admin';
  const canMaintain = detail ? roleAllows(detail.organization.role, 'maintainer') : false;
  const canOwn = detail ? roleAllows(detail.organization.role, 'owner') : false;

  return (
    <div className="organization-workspace">
      <aside className="organization-switcher">
        <div className="organization-switcher__header"><span>Organizations</span><strong>{organizations.length}</strong></div>
        <div className="organization-switcher__list">
          {organizations.map((organization) => (
            <button className={selectedSlug === organization.slug ? 'organization-switcher__item organization-switcher__item--active' : 'organization-switcher__item'} key={organization.id} onClick={() => void selectOrganization(organization.slug)} type="button">
              <span className="organization-avatar">{organization.name.slice(0, 2).toUpperCase()}</span>
              <span><strong>{organization.name}</strong><small>{organization.role} · {organization.project_count} projects</small></span>
            </button>
          ))}
        </div>
        {canCreateOrganization ? (
          <form className="organization-form organization-form--compact" onSubmit={submitOrganization}>
            <h3><Plus aria-hidden="true" size={16} /> New organization</h3>
            <label>Name<input onChange={(event) => setOrganizationForm((current) => ({ ...current, name: event.target.value }))} required value={organizationForm.name} /></label>
            <label>Slug<input onChange={(event) => setOrganizationForm((current) => ({ ...current, slug: event.target.value }))} placeholder="generated from name" value={organizationForm.slug} /></label>
            <label>Description<textarea onChange={(event) => setOrganizationForm((current) => ({ ...current, description: event.target.value }))} value={organizationForm.description} /></label>
            <label>Website<input onChange={(event) => setOrganizationForm((current) => ({ ...current, website: event.target.value }))} placeholder="https://" type="url" value={organizationForm.website} /></label>
            <button className="primary-button primary-button--warm" disabled={busy} type="submit">Create organization</button>
          </form>
        ) : null}
      </aside>

      <div className="organization-main">
        {error ? <div className="banner banner--error">{error}</div> : null}
        {!detail ? <Section title="No organizations"><p>Create an organization to group teams, people, repositories, and delivery work.</p></Section> : (
          <>
            <section className="organization-hero">
              <span className="organization-avatar organization-avatar--large">{detail.organization.name.slice(0, 2).toUpperCase()}</span>
              <div><span className="eyebrow">Organization · {detail.organization.role}</span><h2>{detail.organization.name}</h2><p>{detail.organization.description}</p></div>
              {detail.organization.website ? <a href={detail.organization.website} rel="noreferrer" target="_blank">Website <ExternalLink aria-hidden="true" size={14} /></a> : null}
            </section>
            <nav aria-label="Organization sections" className="organization-tabs">
              {tabs.map(({ id, label, icon: Icon }) => <button className={activeTab === id ? 'organization-tab organization-tab--active' : 'organization-tab'} key={id} onClick={() => setActiveTab(id)} type="button"><Icon aria-hidden="true" size={16} />{label}</button>)}
            </nav>

            {activeTab === 'overview' ? <OverviewTab canMaintain={canMaintain} detail={detail} busy={busy} form={projectForm} setForm={setProjectForm} submit={submitProject} /> : null}
            {activeTab === 'teams' ? <TeamsTab canMaintain={canMaintain} detail={detail} busy={busy} form={teamForm} setForm={setTeamForm} submit={submitTeam} /> : null}
            {activeTab === 'people' ? <PeopleTab canOwn={canOwn} detail={detail} busy={busy} form={memberForm} setForm={setMemberForm} submit={submitMember} /> : null}
            {activeTab === 'insights' ? <InsightsTab detail={detail} /> : null}
            {activeTab === 'settings' ? <SettingsTab detail={detail} /> : null}
          </>
        )}
      </div>
    </div>
  );
}

type ProjectForm = { name: string; description: string; repository_name: string; team_id: string; branch_strategy: string; default_branch: string };
function OverviewTab({ canMaintain, detail, busy, form, setForm, submit }: { canMaintain: boolean; detail: OrganizationDetail; busy: boolean; form: ProjectForm; setForm: React.Dispatch<React.SetStateAction<ProjectForm>>; submit: (event: FormEvent) => void }) {
  return <section className="organization-content-grid"><Section description="Projects grouped under this organization with real Git repositories and governed branching." title="Projects"><div className="organization-projects">{detail.projects.map((project) => <article className="organization-project" key={project.id}><div><FolderGit2 aria-hidden="true" size={18} /><StatusPill value={project.status} /></div><h3>{project.name}</h3><p>{project.description}</p><dl><div><dt>Repository</dt><dd>{project.repository_name}</dd></div><div><dt>Team</dt><dd>{project.team_name || 'Unassigned'}</dd></div><div><dt>Strategy</dt><dd>{project.branch_strategy}</dd></div></dl><a href={project.clone_url}>{project.clone_url}</a></article>)}</div></Section>{canMaintain ? <Section description="Provision a grouped project and its bare Git repository in one controlled operation." title="Create project"><form className="organization-form" onSubmit={submit}><label>Project name<input required value={form.name} onChange={(event) => setForm((current) => ({ ...current, name: event.target.value }))} /></label><label>Description<textarea value={form.description} onChange={(event) => setForm((current) => ({ ...current, description: event.target.value }))} /></label><label>Repository name<input placeholder="generated when empty" value={form.repository_name} onChange={(event) => setForm((current) => ({ ...current, repository_name: event.target.value }))} /></label><div className="organization-form__row"><label>Owning team<select value={form.team_id} onChange={(event) => setForm((current) => ({ ...current, team_id: event.target.value }))}><option value="">Unassigned</option>{detail.teams.map((team) => <option key={team.id} value={team.id}>{team.name}</option>)}</select></label><label>Branch strategy<select value={form.branch_strategy} onChange={(event) => setForm((current) => ({ ...current, branch_strategy: event.target.value }))}><option value="feature">Feature</option><option value="release">Release</option><option value="hotfix">Hotfix</option></select></label></div><label>Default branch<input required value={form.default_branch} onChange={(event) => setForm((current) => ({ ...current, default_branch: event.target.value }))} /></label><button className="primary-button primary-button--warm" disabled={busy} type="submit">Provision project</button></form></Section> : null}</section>;
}

type TeamForm = { name: string; description: string };
function TeamsTab({ canMaintain, detail, busy, form, setForm, submit }: { canMaintain: boolean; detail: OrganizationDetail; busy: boolean; form: TeamForm; setForm: React.Dispatch<React.SetStateAction<TeamForm>>; submit: (event: FormEvent) => void }) {
  return <section className="organization-content-grid"><Section title="Teams"><div className="organization-list">{detail.teams.map((team) => <article key={team.id}><span className="organization-avatar">{team.name.slice(0, 2).toUpperCase()}</span><div><h3>{team.name}</h3><p>{team.description}</p><small>{team.member_count} members</small></div></article>)}</div></Section>{canMaintain ? <Section title="Create team"><form className="organization-form" onSubmit={submit}><label>Team name<input required value={form.name} onChange={(event) => setForm((current) => ({ ...current, name: event.target.value }))} /></label><label>Description<textarea value={form.description} onChange={(event) => setForm((current) => ({ ...current, description: event.target.value }))} /></label><button className="primary-button primary-button--warm" disabled={busy} type="submit">Create team</button></form></Section> : null}</section>;
}

type MemberForm = { username: string; role: string; team_id: string };
function PeopleTab({ canOwn, detail, busy, form, setForm, submit }: { canOwn: boolean; detail: OrganizationDetail; busy: boolean; form: MemberForm; setForm: React.Dispatch<React.SetStateAction<MemberForm>>; submit: (event: FormEvent) => void }) {
  return <section className="organization-content-grid"><Section title="People"><div className="organization-list">{detail.members.map((member) => <article key={member.username}><span className="organization-avatar">{member.username.slice(0, 2).toUpperCase()}</span><div><h3>{member.username}</h3><p>{member.role}</p></div></article>)}</div></Section>{canOwn ? <Section description="Owners can assign organization roles and optional team membership." title="Assign member"><form className="organization-form" onSubmit={submit}><label>Username<input required value={form.username} onChange={(event) => setForm((current) => ({ ...current, username: event.target.value }))} /></label><div className="organization-form__row"><label>Role<select value={form.role} onChange={(event) => setForm((current) => ({ ...current, role: event.target.value }))}><option value="developer">Developer</option><option value="maintainer">Maintainer</option><option value="owner">Owner</option></select></label><label>Team<select value={form.team_id} onChange={(event) => setForm((current) => ({ ...current, team_id: event.target.value }))}><option value="">No team</option>{detail.teams.map((team) => <option key={team.id} value={team.id}>{team.name}</option>)}</select></label></div><button className="primary-button primary-button--warm" disabled={busy} type="submit">Assign member</button></form></Section> : null}</section>;
}

function InsightsTab({ detail }: { detail: OrganizationDetail }) {
  return <Section description="Audited organization, team, membership, and project operations." title="Activity"><ul className="organization-activity">{detail.activity.map((event) => <li key={event.id}><span className="organization-activity__dot" /><div><strong>{event.summary}</strong><span>{event.actor} · {event.resource_type} · {new Date(event.occurred_at).toLocaleString()}</span></div></li>)}</ul></Section>;
}

function SettingsTab({ detail }: { detail: OrganizationDetail }) {
  return <section className="organization-content-grid"><Section title="Organization settings"><div className="settings-list"><div className="settings-list__item"><span>Slug</span><strong>{detail.organization.slug}</strong></div><div className="settings-list__item"><span>Your role</span><strong>{detail.organization.role}</strong></div><div className="settings-list__item"><span>Visibility</span><strong>Private</strong></div><div className="settings-list__item"><span>Audit retention</span><strong>Enabled</strong></div></div></Section><Section description="Capabilities are enforced by the gateway for every mutation." title="Role policy"><div className="settings-list"><div className="settings-list__item"><span>Owner</span><strong>People, teams, projects</strong></div><div className="settings-list__item"><span>Maintainer</span><strong>Teams and projects</strong></div><div className="settings-list__item"><span>Developer</span><strong>Read and collaborate</strong></div></div></Section></section>;
}