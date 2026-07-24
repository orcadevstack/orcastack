create table if not exists dashboard_header_items (
    id varchar(80) primary key,
    section varchar(32) not null,
    label varchar(80) not null,
    description varchar(255) not null,
    path varchar(255) not null,
    icon varchar(48) not null,
    action varchar(32) not null default 'navigate',
    required_permission varchar(80) not null default '',
    sort_order integer not null default 0,
    enabled boolean not null default true,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

alter table dashboard_header_items
    add column if not exists action varchar(32) not null default 'navigate';

create index if not exists dashboard_header_items_section_idx
    on dashboard_header_items (section, sort_order)
    where enabled = true;

insert into dashboard_header_items (id, section, label, description, path, icon, required_permission, sort_order)
values
    ('projects-overview', 'projects', 'Project catalogue', 'Browse governed source repositories.', '/app/repositories', 'folder-git-2', 'repositories:read', 10),
    ('projects-reviews', 'projects', 'Review activity', 'Inspect repository review and approval state.', '/app/repositories#reviews', 'git-pull-request', 'repositories:read', 20),
    ('deployments-pipelines', 'deployments', 'Delivery pipelines', 'Run and inspect controlled CI workflows.', '/app/pipelines', 'workflow', 'pipelines:write', 10),
    ('deployments-environments', 'deployments', 'Environments', 'Track releases across managed environments.', '/app/deployments', 'rocket', 'deployments:write', 20),
    ('accounts-profile', 'accounts', 'My account', 'Review your active identity and access realm.', '/app/accounts', 'user-round', 'control-panel:read', 10),
    ('accounts-access', 'accounts', 'Access requests', 'Review pending workspace access requests.', '/app/accounts#access-requests', 'users-round', 'control-panel:admin', 20),
    ('settings-platform', 'settings', 'Platform settings', 'Inspect workspace security and service posture.', '/app/settings', 'settings-2', 'control-panel:read', 10),
    ('settings-security', 'settings', 'Security controls', 'Review permissions and identity enforcement.', '/app/settings#security', 'shield-check', 'control-panel:read', 20),
    ('profile-account', 'profile', 'Profile', 'Review your authenticated account identity.', '/app/accounts#profile', 'user-round', 'control-panel:read', 10),
    ('profile-settings', 'profile', 'Settings', 'Manage workspace identity and security preferences.', '/app/settings', 'settings-2', 'control-panel:read', 20),
    ('profile-notifications', 'profile', 'Notifications', 'Review recent platform activity and alerts.', '/app/accounts#notifications', 'bell', 'control-panel:read', 30),
    ('profile-access-admin', 'profile', 'Access administration', 'Review workspace account access requests.', '/app/accounts#access-requests', 'shield-check', 'control-panel:admin', 40),
    ('profile-signout', 'profile', 'Sign out', 'Securely end this workspace session.', '', 'log-out', '', 50)
on conflict (id) do update set
    section = excluded.section,
    label = excluded.label,
    description = excluded.description,
    path = excluded.path,
    icon = excluded.icon,
    action = excluded.action,
    required_permission = excluded.required_permission,
    sort_order = excluded.sort_order,
    enabled = excluded.enabled,
    updated_at = now();

update dashboard_header_items
set action = 'logout'
where id = 'profile-signout';

create table if not exists dashboard_header_audit_log (
    id uuid primary key,
    username varchar(80) not null,
    role varchar(64) not null,
    action varchar(48) not null,
    section varchar(32) not null,
    target_path varchar(255) not null default '',
    occurred_at timestamptz not null default now()
);

create index if not exists dashboard_header_audit_log_user_time_idx
    on dashboard_header_audit_log (username, occurred_at desc);

create index if not exists dashboard_header_audit_log_time_idx
    on dashboard_header_audit_log (occurred_at desc);