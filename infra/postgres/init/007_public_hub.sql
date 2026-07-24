create table if not exists community_posts (
    id uuid primary key,
    slug varchar(160) not null unique,
    kind varchar(32) not null check (kind in ('discussion', 'tutorial', 'announcement', 'project')),
    title varchar(255) not null,
    excerpt text not null default '',
    markdown text not null,
    media_url text not null default '',
    author_username varchar(80) not null,
    visibility varchar(24) not null default 'public' check (visibility in ('public', 'members')),
    featured boolean not null default false,
    published boolean not null default true,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create index if not exists community_posts_published_idx
    on community_posts (featured desc, updated_at desc)
    where published = true;

create table if not exists community_events (
    id uuid primary key,
    title varchar(255) not null,
    summary text not null default '',
    starts_at timestamptz not null,
    ends_at timestamptz,
    location varchar(255) not null default '',
    event_url text not null default '',
    published boolean not null default true,
    created_at timestamptz not null default now()
);

create index if not exists community_events_upcoming_idx
    on community_events (starts_at)
    where published = true;

create table if not exists public_hub_audit_log (
    id uuid primary key,
    actor varchar(80) not null,
    role varchar(64) not null,
    action varchar(48) not null,
    resource_type varchar(48) not null,
    resource_id varchar(255) not null default '',
    occurred_at timestamptz not null default now()
);

create index if not exists public_hub_audit_log_time_idx
    on public_hub_audit_log (occurred_at desc);

insert into community_posts (id, slug, kind, title, excerpt, markdown, author_username, featured)
values
    ('a2b182d8-834a-45d6-afde-07c39fb17801', 'welcome-to-orcastack', 'announcement', 'Welcome to the OrcaStack developer community', 'The public hub for platform engineering knowledge, project updates, and contributor collaboration.', E'OrcaStack brings source governance, delivery automation, and device orchestration into one control plane.\n\nUse this hub to follow product announcements, publish implementation notes, and exchange operational knowledge.', 'orcastack', true),
    ('a2b182d8-834a-45d6-afde-07c39fb17802', 'contributor-workflow', 'tutorial', 'Contributor workflow', 'A concise path from local checkout to a reviewed OrcaStack change.', E'## Start with a focused change\n\n```bash\ngit clone https://github.com/orcastack/orcastack.git\ncd orcastack\n```\n\nOpen a focused change, run the relevant test suite, and document the operational impact for reviewers.', 'orcastack', true),
    ('a2b182d8-834a-45d6-afde-07c39fb17803', 'api-first-automation', 'discussion', 'Designing API-first automation lanes', 'Discuss reliable contracts for software delivery and hardware-backed validation.', E'How should a delivery contract represent approvals, device reservations, artifacts, and rollback checkpoints? Share concrete API patterns and implementation experience.', 'orcastack', false)
on conflict (slug) do nothing;
