import { requestGateway } from './client';
import type { CommunityPost, CreateRepositoryInput, HeaderNavigation, HeaderSection, ImportRepositoryInput, OrganizationDetail, OrganizationMember, OrganizationProject, OrganizationSummary, OrganizationTeam, Overview, PublicCommunity, PublicHome, RepositoryMutationResult } from './types';

export async function fetchOverview(signal?: AbortSignal, token?: string | null): Promise<Overview> {
  return requestGateway<Overview>('/api/overview', { signal, token });
}

export async function fetchHeaderNavigation(signal?: AbortSignal, token?: string | null): Promise<HeaderNavigation> {
  return requestGateway<HeaderNavigation>('/api/header/navigation', { signal, token });
}

export async function auditHeaderInteraction(
  action: 'open' | 'navigate',
  section: HeaderSection,
  targetPath: string,
  token?: string | null,
): Promise<void> {
  return requestGateway<void>('/api/header/audit', {
    token,
    method: 'POST',
    keepalive: true,
    body: { action, section, target_path: targetPath },
  });
}

export async function fetchOrganizations(signal?: AbortSignal, token?: string | null): Promise<OrganizationSummary[]> {
  const response = await requestGateway<{ organizations: OrganizationSummary[] }>('/api/organizations', { signal, token });
  return response.organizations;
}

export async function fetchPublicHome(signal?: AbortSignal, token?: string | null): Promise<PublicHome> {
  return requestGateway<PublicHome>('/api/public/home', { signal, token });
}

export async function fetchPublicCommunity(signal?: AbortSignal, token?: string | null): Promise<PublicCommunity> {
  return requestGateway<PublicCommunity>('/api/public/community', { signal, token });
}

export async function auditPublicHub(resourceType: 'deployment' | 'community-post' | 'resource', resourceId: string, token?: string | null): Promise<void> {
  return requestGateway<void>('/api/public/audit', { token, method: 'POST', keepalive: true, body: { action: 'open', resource_type: resourceType, resource_id: resourceId } });
}

export async function createCommunityPost(input: { kind: string; title: string; excerpt: string; markdown: string; media_url: string }, token: string): Promise<CommunityPost> {
  return requestGateway<CommunityPost>('/api/community/posts', { token, method: 'POST', body: input });
}

export async function fetchOrganization(slug: string, signal?: AbortSignal, token?: string | null): Promise<OrganizationDetail> {
  return requestGateway<OrganizationDetail>(`/api/organizations/${slug}`, { signal, token });
}

export async function createOrganization(input: { name: string; slug: string; description: string; website: string }, token?: string | null): Promise<OrganizationSummary> {
  return requestGateway<OrganizationSummary>('/api/organizations', { token, method: 'POST', body: input });
}

export async function createOrganizationTeam(slug: string, input: { name: string; description: string }, token?: string | null): Promise<OrganizationTeam> {
  return requestGateway<OrganizationTeam>(`/api/organizations/${slug}/teams`, { token, method: 'POST', body: input });
}

export async function addOrganizationMember(slug: string, input: { username: string; role: string; team_id: string }, token?: string | null): Promise<OrganizationMember> {
  return requestGateway<OrganizationMember>(`/api/organizations/${slug}/members`, { token, method: 'POST', body: input });
}

export async function createOrganizationProject(slug: string, input: { name: string; description: string; repository_name: string; team_id: string; branch_strategy: string; default_branch: string }, token?: string | null): Promise<OrganizationProject> {
  return requestGateway<OrganizationProject>(`/api/organizations/${slug}/projects`, { token, method: 'POST', body: input });
}

export async function createRepository(input: CreateRepositoryInput, token?: string | null): Promise<RepositoryMutationResult> {
  return requestGateway<RepositoryMutationResult>('/api/repositories', {
    token,
    method: 'POST',
    body: {
      name: input.name,
      summary: input.summary,
      default_branch: input.defaultBranch,
    },
  });
}

export async function importRepository(input: ImportRepositoryInput, token?: string | null): Promise<RepositoryMutationResult> {
  return requestGateway<RepositoryMutationResult>('/api/repositories/import', {
    token,
    method: 'POST',
    body: {
      name: input.name,
      summary: input.summary,
      default_branch: input.defaultBranch,
      source_url: input.sourceUrl,
    },
  });
}
