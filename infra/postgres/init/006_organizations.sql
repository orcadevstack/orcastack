create table if not exists organizations (
    id uuid primary key,
    slug varchar(120) not null unique,
    name varchar(255) not null,
    description text not null default '',
    website text not null default '',
    created_by varchar(80) not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table if not exists organization_members (
    id uuid primary key,
    organization_id uuid not null references organizations(id) on delete cascade,
    username varchar(80) not null,
    role varchar(32) not null check (role in ('owner', 'maintainer', 'developer')),
    created_at timestamptz not null default now(),
    unique (organization_id, username)
);

create index if not exists organization_members_username_idx
    on organization_members (username, organization_id);

create table if not exists organization_teams (
    id uuid primary key,
    organization_id uuid not null references organizations(id) on delete cascade,
    slug varchar(120) not null,
    name varchar(255) not null,
    description text not null default '',
    created_by varchar(80) not null,
    created_at timestamptz not null default now(),
    unique (organization_id, slug)
);

create table if not exists organization_team_members (
    team_id uuid not null references organization_teams(id) on delete cascade,
    username varchar(80) not null,
    created_at timestamptz not null default now(),
    primary key (team_id, username)
);

create table if not exists organization_projects (
    id uuid primary key,
    organization_id uuid not null references organizations(id) on delete cascade,
    team_id uuid references organization_teams(id) on delete set null,
    repository_name varchar(255) not null unique,
    name varchar(255) not null,
    description text not null default '',
    branch_strategy varchar(32) not null check (branch_strategy in ('feature', 'release', 'hotfix')),
    default_branch varchar(255) not null default 'main',
    status varchar(32) not null default 'provisioning',
    created_by varchar(80) not null,
    created_at timestamptz not null default now(),
    unique (organization_id, name)
);

create table if not exists organization_activity (
    id uuid primary key,
    organization_id uuid not null references organizations(id) on delete cascade,
    actor varchar(80) not null,
    action varchar(64) not null,
    resource_type varchar(48) not null,
    resource_id varchar(255) not null,
    summary text not null,
    occurred_at timestamptz not null default now()
);

create index if not exists organization_activity_org_time_idx
    on organization_activity (organization_id, occurred_at desc);

insert into dashboard_header_items (id, section, label, description, path, icon, action, required_permission, sort_order)
values ('projects-organizations', 'projects', 'Organizations', 'Manage organizations, teams, people, and grouped projects.', '/app/organizations', 'building-2', 'navigate', 'repositories:read', 5)
on conflict (id) do update set
    label = excluded.label,
    description = excluded.description,
    path = excluded.path,
    icon = excluded.icon,
    action = excluded.action,
    required_permission = excluded.required_permission,
    sort_order = excluded.sort_order,
    enabled = true,
    updated_at = now();