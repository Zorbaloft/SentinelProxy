// Browser must use localhost, not Docker service name
const API_URL = 'http://localhost:8090';
const ADMIN_TOKEN = import.meta.env.VITE_ADMIN_TOKEN || 'changeme';

async function fetchAPI(endpoint: string, options: RequestInit = {}) {
  const response = await fetch(`${API_URL}${endpoint}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      'X-Sentinel-Admin-Token': ADMIN_TOKEN,
      ...options.headers,
    },
  });

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(`API error: ${response.status} ${response.statusText} - ${errorText}`);
  }

  return response.json();
}

export const api = {
  logs: {
    get: (params?: { from?: string; to?: string; ip?: string; path?: string; status?: string; limit?: number; cursor?: string }) => {
      const query = new URLSearchParams();
      if (params) {
        Object.entries(params).forEach(([key, value]) => {
          if (value !== undefined && value !== null) {
            query.append(key, String(value));
          }
        });
      }
      return fetchAPI(`/logs?${query.toString()}`);
    },
  },
  incidents: {
    get: (status?: string) => fetchAPI(`/incidents${status ? `?status=${status}` : ''}`),
    close: (id: string) => fetchAPI(`/incidents/${id}/close`, { method: 'POST' }),
  },
  rules: {
    get: () => fetchAPI('/rules'),
    create: (rule: any) => fetchAPI('/rules', { method: 'POST', body: JSON.stringify(rule) }),
    update: (id: string, rule: any) => fetchAPI(`/rules/${id}`, { method: 'PUT', body: JSON.stringify(rule) }),
    delete: (id: string) => fetchAPI(`/rules/${id}`, { method: 'DELETE' }),
  },
  actions: {
    block: (ip: string, ttlSec: number, reason: string) => 
      fetchAPI('/block', { method: 'POST', body: JSON.stringify({ ip, ttlSec, reason }) }),
    unblock: (ip: string) => 
      fetchAPI('/unblock', { method: 'POST', body: JSON.stringify({ ip }) }),
    redirect: (ip: string, targetUrl: string, ttlSec: number, reason: string) => 
      fetchAPI('/redirect', { method: 'POST', body: JSON.stringify({ ip, targetUrl, ttlSec, reason }) }),
    unredirect: (ip: string) => 
      fetchAPI('/unredirect', { method: 'POST', body: JSON.stringify({ ip }) }),
  },
  ai: {
    analyze: (questionId: string, timeRange?: string) =>
      fetchAPI('/ai/analyze', {
        method: 'POST',
        body: JSON.stringify({ questionId, timeRange }),
      }),
  },
};
