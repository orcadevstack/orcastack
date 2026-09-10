import type { AuthSession } from './types';

type RequestOptions = {
  signal?: AbortSignal;
  token?: string | null;
  method?: 'GET' | 'POST' | 'PATCH';
  body?: unknown;
  keepalive?: boolean;
};

const configuredGatewayBase = import.meta.env.VITE_ORCASTACK_GATEWAY_URL;
const configuredAppBase = import.meta.env.VITE_ORCASTACK_APP_URL;
const configuredRunnerBase = import.meta.env.VITE_ORCASTACK_RUNNER_URL;
const configuredDeviceBase = import.meta.env.VITE_ORCASTACK_DEVICE_ORCH_URL;
const configuredHWAutomationBase = import.meta.env.VITE_ORCASTACK_HW_AUTOMATION_URL;
const configuredSWAutomationBase = import.meta.env.VITE_ORCASTACK_SW_AUTOMATION_URL;

function isLocalHostname(hostname: string) {
  return hostname === 'localhost' || hostname === '127.0.0.1' || hostname === '::1';
}

function isGitHubPagesHostname(hostname: string) {
  return hostname.endsWith('.github.io');
}

function normalizeConfiguredBase(value?: string | null) {
  if (!value) {
    return null;
  }

  if (typeof window === 'undefined') {
    return value;
  }

  try {
    return new URL(value, window.location.origin).toString().replace(/\/$/, '');
  } catch {
    return value;
  }
}

function shouldUseConfiguredBase(value?: string | null) {
  const resolvedBase = normalizeConfiguredBase(value);
  if (!resolvedBase) {
    return false;
  }

  if (typeof window === 'undefined') {
    return true;
  }

  try {
    const serviceUrl = new URL(resolvedBase, window.location.origin);

    if (isGitHubPagesHostname(window.location.hostname)) {
      if (isLocalHostname(serviceUrl.hostname)) {
        return false;
      }

      if (serviceUrl.origin === window.location.origin) {
        return false;
      }
    }

    return true;
  } catch {
    return true;
  }
}

function candidatesFor(configuredBase: string | undefined, localPort: number, pathPrefix = '') {
  const resolvedBase = normalizeConfiguredBase(configuredBase);

  if (resolvedBase && shouldUseConfiguredBase(configuredBase)) {
    return [resolvedBase];
  }

  if (typeof window === 'undefined') {
    return [`http://localhost:${localPort}${pathPrefix}`];
  }

  const sameOriginBase = `${window.location.origin}${pathPrefix}`.replace(/\/$/, '');

  if (isLocalHostname(window.location.hostname)) {
    return [sameOriginBase, `http://localhost:${localPort}${pathPrefix}`, `http://localhost:18080${pathPrefix}`];
  }

  if (!isGitHubPagesHostname(window.location.hostname)) {
    return [sameOriginBase];
  }

  return [];
}

const gatewayCandidates = candidatesFor(configuredGatewayBase, 8080);
const appCandidates = candidatesFor(configuredAppBase, 3000);
const runnerCandidates = candidatesFor(configuredRunnerBase, 8086, '/runner');
const deviceCandidates = candidatesFor(configuredDeviceBase, 8089, '/device-orch');
const hwAutomationCandidates = candidatesFor(configuredHWAutomationBase, 8087, '/hw-automation');
const swAutomationCandidates = candidatesFor(configuredSWAutomationBase, 8088, '/sw-automation');

let lastResolvedGatewayBase = gatewayCandidates[0] ?? 'gateway unavailable';

async function requestJSON<T>(candidates: string[], path: string, options: RequestOptions = {}): Promise<T> {
  let lastError: Error | null = null;

  for (const base of candidates) {
    try {
      const headers = new Headers();
      if (options.body !== undefined) {
        headers.set('Content-Type', 'application/json');
      }
      if (options.token) {
        headers.set('Authorization', `Bearer ${options.token}`);
      }

      const response = await fetch(`${base}${path}`, {
        signal: options.signal,
        method: options.method || 'GET',
        headers,
        body: options.body !== undefined ? JSON.stringify(options.body) : undefined,
        keepalive: options.keepalive,
      });

      if (!response.ok) {
        let message = `Service returned ${response.status}`;
        try {
          const errorPayload = (await response.json()) as { error?: string };
          if (errorPayload.error) {
            message = errorPayload.error;
          }
        } catch {
          // ignore parse failure and preserve default message
        }
        throw new Error(message);
      }

      if (response.status === 204) {
        return undefined as T;
      }

      if (candidates === gatewayCandidates) {
        lastResolvedGatewayBase = base;
      }

      return response.json() as Promise<T>;
    } catch (error) {
      if (options.signal?.aborted) {
        throw error;
      }
      lastError = error instanceof Error ? error : new Error('Unknown service error');
    }
  }

  throw lastError ?? new Error('Failed to reach any configured service endpoint');
}

export function shouldUseStaticOverview() {
  if (shouldUseConfiguredBase(configuredGatewayBase)) {
    return false;
  }

  if (typeof window === 'undefined') {
    return false;
  }

  return !isLocalHostname(window.location.hostname);
}

export async function requestGateway<T>(path: string, options: RequestOptions = {}) {
  return requestJSON<T>(gatewayCandidates, path, options);
}

export async function requestRubyApp<T>(path: string, options: RequestOptions = {}) {
  return requestJSON<T>(appCandidates, path, options);
}

export async function requestRunner<T>(path: string, options: RequestOptions = {}) {
  return requestJSON<T>(runnerCandidates, path, options);
}

export async function requestDeviceOrch<T>(path: string, options: RequestOptions = {}) {
  return requestJSON<T>(deviceCandidates, path, options);
}

export async function requestHWAutomation<T>(path: string, options: RequestOptions = {}) {
  return requestJSON<T>(hwAutomationCandidates, path, options);
}

export async function requestSWAutomation<T>(path: string, options: RequestOptions = {}) {
  return requestJSON<T>(swAutomationCandidates, path, options);
}

export function getGatewayBase() {
  return lastResolvedGatewayBase;
}

export function isStaticOverviewMode() {
  return shouldUseStaticOverview();
}

export async function fetchSession(token: string, signal?: AbortSignal): Promise<AuthSession> {
  return requestGateway<AuthSession>('/api/auth/session', { signal, token });
}

export async function logout(token: string, signal?: AbortSignal): Promise<void> {
  await requestGateway<void>('/api/auth/logout', {
    signal,
    token,
    method: 'POST',
  });
}

export {
  configuredAppBase,
  configuredDeviceBase,
  configuredGatewayBase,
  configuredHWAutomationBase,
  configuredRunnerBase,
  configuredSWAutomationBase,
  gatewayCandidates,
};
