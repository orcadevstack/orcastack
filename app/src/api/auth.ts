import { requestGateway } from './client';
import type { AuthSession, SignupRequestInput, SignupRequestRecord, SignupRequestResult } from './types';

export async function login(username: string, password: string, signal?: AbortSignal): Promise<AuthSession> {
  return requestGateway<AuthSession>('/api/auth/login', {
    signal,
    method: 'POST',
    body: { username, password },
  });
}

export async function signup(input: SignupRequestInput, signal?: AbortSignal): Promise<SignupRequestResult> {
  const response = await requestGateway<SignupRequestResult>('/api/auth/signup', {
    signal,
    method: 'POST',
    body: {
      username: input.username,
      email: input.email,
      password: input.password,
    },
  });

  return {
    request_id: response.request_id ?? response.id ?? 'pending-request',
    status: response.status,
    message: response.message ?? 'Signup request submitted for review.',
    id: response.id,
    created_at: response.created_at,
  };
}

export async function fetchSignupRequests(token: string, signal?: AbortSignal): Promise<SignupRequestRecord[]> {
  return requestGateway<SignupRequestRecord[]>('/api/auth/signup-requests', {
    signal,
    token,
  });
}

export async function reviewSignupRequest(
  id: string,
  status: 'approved' | 'rejected',
  note: string,
  token: string,
  signal?: AbortSignal,
): Promise<SignupRequestRecord> {
  return requestGateway<SignupRequestRecord>(`/api/auth/signup-requests/${id}`, {
    signal,
    token,
    method: 'POST',
    body: { status, note },
  });
}
