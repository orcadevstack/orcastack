export type Provider = {
  id: string;
  name: string;
  status: string;
  repos: number;
  latency: string;
  identity: string;
  connected: boolean;
};

export type Repository = {
  id: string;
  provider_id: string;
  name: string;
  branch: string;
  default_branch: string;
  commit: string;
  last_commit_at: string;
  reviewer: string;
  summary: string;
  clone_url: string;
  identity: string;
  security: SecurityState;
};

export type SecurityState = {
  ldap_registered: boolean;
  rbac_verified: boolean;
  attestation_signed: boolean;
  verified: boolean;
};

export type CloneOperation = {
  repository_id: string;
  status: string;
  clone_url: string;
  command: string;
  upi: string;
  updated_at: string;
};

export type Review = {
  id: string;
  repository_id: string;
  title: string;
  status: string;
  required_approvals: number;
  approvals: number;
  reviewers: string[];
  last_updated: string;
};

export type Pipeline = {
  id: string;
  repository_id: string;
  name: string;
  branch: string;
  last_run: string;
  status: string;
  stages: PipelineStage[];
  run_history: PipelineRun[];
  log_channel: string;
  upi: string;
  security: SecurityState;
  updated_at: string;
};

export type PipelineStage = {
  name: string;
  status: string;
};

export type PipelineRun = {
  id: string;
  started_at: string;
  status: string;
  trigger: string;
};

export type Deployment = {
  id: string;
  repository_id: string;
  service_name: string;
  version: string;
  environment: string;
  status: string;
  cluster: string;
  artifact: string;
  target_commit: string;
  previous_version: string;
  log_channel: string;
  upi: string;
  security: SecurityState;
};

export type Container = {
  name: string;
  upi: string;
  state: string;
  host: string;
  actions: string[];
  cpu: string;
  memory: string;
  restarts: number;
  metrics_url: string;
  log_channel: string;
  security: SecurityState;
};

export type DashboardSecurity = {
  repository_identity: string;
  ui_process_identity: string;
  directory: SecurityState;
};

export type EventEntry = {
  id: string;
  time: string;
  component: string;
  kind: string;
  repository_id: string;
  action: string;
  result: string;
  upi: string;
  summary: string;
};

export type CloudLayer = {
  name: string;
  platform: string;
  status: string;
  endpoint: string;
  identity: string;
  summary: string;
  coverage: string[];
};

export type Cluster = {
  id: string;
  name: string;
  provider: string;
  status: string;
  version: string;
  control_planes: number;
  workers: number;
  gpu_workers: number;
  rancher_project: string;
  registration_status: string;
  upgrade_policy: string;
  api_endpoint: string;
};

export type AutomationLane = {
  name: string;
  type: string;
  status: string;
  entrypoint: string;
  target: string;
  last_run: string;
};

export type ObservabilitySurface = {
  name: string;
  kind: string;
  status: string;
  endpoint: string;
  backing: string;
};

export type SelfManagementCapability = {
  name: string;
  status: string;
  workflow: string;
  summary: string;
};

export type Metric = {
  label: string;
  value: string;
  hint: string;
};

export type Overview = {
  providers: Provider[];
  repositories: Repository[];
  clone_operations: CloneOperation[];
  reviews: Review[];
  pipelines: Pipeline[];
  deployments: Deployment[];
  containers: Container[];
  security: DashboardSecurity;
  events: EventEntry[];
  updated_at: string;
  metrics: Metric[];
  activity: string[];
  cloud_layers: CloudLayer[];
  clusters: Cluster[];
  automation_lanes: AutomationLane[];
  observability: ObservabilitySurface[];
  self_management: SelfManagementCapability[];
};

export type AuthUser = {
  username: string;
  full_name: string;
  email: string;
  role: string;
  identity: string;
  rbac_realm: string;
  permissions: string[];
};

export type AuthSession = {
  token?: string;
  user: AuthUser;
  expires_at: string;
};

export type HeaderSection = 'projects' | 'deployments' | 'accounts' | 'settings' | 'profile';

export type HeaderNavigationItem = {
  id: string;
  section: HeaderSection;
  label: string;
  description: string;
  path: string;
  icon: string;
  action: 'navigate' | 'logout';
  count?: number;
};

export type HeaderNavigation = {
  organization: string;
  tagline: string;
  items: HeaderNavigationItem[];
  updated_at: string;
};

export type OrganizationRole = 'owner' | 'maintainer' | 'developer';

export type OrganizationSummary = {
  id: string;
  slug: string;
  name: string;
  description: string;
  website: string;
  role: OrganizationRole;
  member_count: number;
  team_count: number;
  project_count: number;
  created_at: string;
};

export type OrganizationMember = {
  username: string;
  role: OrganizationRole;
  created_at: string;
};

export type OrganizationTeam = {
  id: string;
  slug: string;
  name: string;
  description: string;
  member_count: number;
  created_at: string;
};

export type OrganizationProject = {
  id: string;
  name: string;
  description: string;
  repository_name: string;
  team_id?: string;
  team_name?: string;
  branch_strategy: 'feature' | 'release' | 'hotfix';
  default_branch: string;
  status: string;
  clone_url: string;
  created_at: string;
};

export type OrganizationActivity = {
  id: string;
  actor: string;
  action: string;
  resource_type: string;
  resource_id: string;
  summary: string;
  occurred_at: string;
};

export type OrganizationDetail = {
  organization: OrganizationSummary;
  members: OrganizationMember[];
  teams: OrganizationTeam[];
  projects: OrganizationProject[];
  activity: OrganizationActivity[];
};

export type PublicDeployment = {
  id: string;
  project_slug: string;
  project_name: string;
  repository_name: string;
  environment: string;
  target_kind: string;
  status: string;
  build_status: string;
  commit_sha: string;
  updated_at: string;
  dashboard_path: string;
  can_access_dashboard: boolean;
};

export type PublicHome = {
  deployments: PublicDeployment[];
  recent_builds: PublicDeployment[];
  community: { posts: number; contributors: number; events: number };
  viewer_role: 'admin' | 'developer' | 'viewer';
  updated_at: string;
};

export type CommunityPost = {
  id: string;
  slug: string;
  kind: 'discussion' | 'tutorial' | 'announcement' | 'project';
  title: string;
  excerpt: string;
  markdown: string;
  media_url: string;
  author_username: string;
  featured: boolean;
  created_at: string;
  updated_at: string;
};

export type CommunityContributor = {
  username: string;
  display_name: string;
  contributions: number;
};

export type CommunityEvent = {
  id: string;
  title: string;
  summary: string;
  starts_at: string;
  ends_at: string;
  location: string;
  event_url: string;
};

export type PublicCommunity = {
  posts: CommunityPost[];
  contributors: CommunityContributor[];
  events: CommunityEvent[];
  can_publish: boolean;
  viewer_role: 'admin' | 'developer' | 'viewer';
};

export type SignupRequestInput = {
  username: string;
  email: string;
  password: string;
};

export type SignupRequestResult = {
  request_id: string;
  status: string;
  message: string;
  id?: string;
  created_at?: string;
};

export type SignupRequestRecord = {
  id: string;
  username: string;
  email: string;
  status: string;
  created_at: string;
  reviewed_at?: string;
  reviewed_by?: string;
  review_note?: string;
};

export type RepositoryMutationResult = {
  repository: Repository;
  clone_operation: CloneOperation;
};

export type CreateRepositoryInput = {
  name: string;
  summary: string;
  defaultBranch: string;
};

export type ImportRepositoryInput = {
  name: string;
  summary: string;
  defaultBranch: string;
  sourceUrl: string;
};

export type BillingPlan = {
  id: string;
  name: string;
  price_monthly: number;
  quotas: Record<string, number>;
};

export type BillingUsage = {
  active_plan: string;
  metrics: Record<string, number>;
};

export type UserProfile = {
  id: string;
  username: string;
  email: string;
  full_name: string;
  role: string;
  identity: string;
  rbac_realm: string;
  preferences: {
    theme: string;
    notifications: string[];
    timezone: string;
    bio: string;
    avatar: string;
  };
  ssh_keys: Array<{
    id: string;
    title: string;
    fingerprint: string;
  }>;
  permissions: string[];
};

export type WikiPage = {
  id: string;
  project_id: string;
  slug: string;
  title: string;
  markdown: string;
  updated_at: string;
};

export type AdminSummary = {
  users: number;
  projects: number;
  pending_notifications: number;
  feature_flags: Record<string, boolean>;
};

export type RunnerJob = {
  id: string;
  name: string;
  stage: string;
  status: string;
  executor: string;
  started_at: string;
  finished_at?: string;
  log_endpoint: string;
};

export type RunnerPipeline = {
  id: string;
  project: string;
  ref: string;
  status: string;
  started_at: string;
  trigger: string;
  jobs: RunnerJob[];
  artifact_url: string;
  labels: string[];
};

export type RunnerSummary = {
  capacity: number;
  busy_executors: number;
  queued_jobs: number;
  live_updates_mode: string;
};

export type Device = {
  id: string;
  name: string;
  status: string;
  health: string;
  tags: string[];
  location: string;
  capabilities: string[];
  assigned_run?: string;
};

export type DeviceListResponse = {
  devices: Device[];
  summary: Record<string, number>;
};

export type AutomationWorkflow = {
  id: string;
  name: string;
  status: string;
  target_pool?: string;
  required_tags?: string[];
  last_run?: string;
  firmware_image?: string;
  target?: string;
  integrations?: string[];
};
