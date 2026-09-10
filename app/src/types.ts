import type {
  AutomationWorkflow,
  DeviceListResponse,
  Overview,
  RunnerPipeline,
  RunnerSummary,
  SignupRequestRecord,
} from './api';

export type RunnerData = {
  pipelines: RunnerPipeline[];
  summary: RunnerSummary;
};

export type DashboardState = {
  overview: Overview | null;
  runner: RunnerData | null;
  devices: DeviceListResponse | null;
  hardwareWorkflows: AutomationWorkflow[];
  softwareWorkflows: AutomationWorkflow[];
  signupRequests: SignupRequestRecord[];
};

export type AuthMode = 'login' | 'signup';

export type PublicRoute = 'home' | 'docs' | 'developer' | 'community';

export const emptyDashboardState: DashboardState = {
  overview: null,
  runner: null,
  devices: null,
  hardwareWorkflows: [],
  softwareWorkflows: [],
  signupRequests: [],
};