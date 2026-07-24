import { requestDeviceOrch, requestHWAutomation, requestRubyApp, requestRunner, requestSWAutomation } from './client';
import type {
  AdminSummary,
  AutomationWorkflow,
  BillingPlan,
  BillingUsage,
  Device,
  DeviceListResponse,
  RunnerJob,
  RunnerPipeline,
  RunnerSummary,
  UserProfile,
  WikiPage,
} from './types';

export async function fetchProfile(signal?: AbortSignal): Promise<UserProfile> {
  return requestRubyApp<UserProfile>('/api/v1/users/profile', { signal });
}

export async function updatePreferences(payload: Partial<UserProfile['preferences']> & { fullName?: string; email?: string; identity?: string }, signal?: AbortSignal): Promise<UserProfile> {
  return requestRubyApp<UserProfile>('/api/v1/users/preferences', {
    signal,
    method: 'PATCH',
    body: payload,
  });
}

export async function fetchBillingPlans(signal?: AbortSignal): Promise<BillingPlan[]> {
  const response = await requestRubyApp<{ plans: BillingPlan[] }>('/api/v1/billing/plans', { signal });
  return response.plans;
}

export async function fetchBillingUsage(signal?: AbortSignal): Promise<BillingUsage> {
  return requestRubyApp<BillingUsage>('/api/v1/billing/usage', { signal });
}

export async function fetchWikiPages(signal?: AbortSignal): Promise<WikiPage[]> {
  const response = await requestRubyApp<{ pages: WikiPage[] }>('/api/v1/wiki/pages', { signal });
  return response.pages;
}

export async function fetchAdminSummary(signal?: AbortSignal): Promise<AdminSummary> {
  return requestRubyApp<AdminSummary>('/api/v1/admin/summary', { signal });
}

export async function fetchRunnerPipelines(signal?: AbortSignal): Promise<{ pipelines: RunnerPipeline[]; summary: RunnerSummary }> {
  return requestRunner<{ pipelines: RunnerPipeline[]; summary: RunnerSummary }>('/pipelines', { signal });
}

export async function createRunnerPipeline(): Promise<RunnerPipeline> {
  return requestRunner<RunnerPipeline>('/pipelines', {
    method: 'POST',
    body: {
      project: 'orcastack/platform',
      ref: 'main',
      trigger: 'manual',
      workflow: 'api-test',
    },
  });
}

export async function fetchRunnerJobs(signal?: AbortSignal): Promise<{ jobs: RunnerJob[]; summary: RunnerSummary }> {
  return requestRunner<{ jobs: RunnerJob[]; summary: RunnerSummary }>('/jobs', { signal });
}

export async function fetchDevices(signal?: AbortSignal): Promise<DeviceListResponse> {
  return requestDeviceOrch<DeviceListResponse>('/devices', { signal });
}

export async function fetchDevice(id: string, signal?: AbortSignal): Promise<Device> {
  return requestDeviceOrch<Device>(`/devices/${id}`, { signal });
}

export async function fetchHardwareWorkflows(signal?: AbortSignal): Promise<AutomationWorkflow[]> {
  const response = await requestHWAutomation<{ workflows: AutomationWorkflow[] }>('/workflows', { signal });
  return response.workflows;
}

export async function fetchSoftwareWorkflows(signal?: AbortSignal): Promise<AutomationWorkflow[]> {
  const response = await requestSWAutomation<{ workflows: AutomationWorkflow[] }>('/workflows', { signal });
  return response.workflows;
}
