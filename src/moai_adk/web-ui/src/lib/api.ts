import type {
  ApiResponse,
  Session,
  Spec,
  CostSummary,
  CostRecord,
  ProviderConfig,
  Message,
} from '@/types'

const API_BASE = '/api'

async function fetchApi<T>(
  endpoint: string,
  options?: RequestInit
): Promise<ApiResponse<T>> {
  try {
    const response = await fetch(`${API_BASE}${endpoint}`, {
      headers: {
        'Content-Type': 'application/json',
        ...options?.headers,
      },
      ...options,
    })

    if (!response.ok) {
      const error = await response.json().catch(() => ({
        code: 'UNKNOWN_ERROR',
        message: response.statusText,
      }))
      return { success: false, error }
    }

    const data = await response.json()
    return { success: true, data }
  } catch (error) {
    return {
      success: false,
      error: {
        code: 'NETWORK_ERROR',
        message: error instanceof Error ? error.message : 'Network error',
      },
    }
  }
}

// Provider API
export const providerApi = {
  list: () => fetchApi<ProviderConfig[]>('/providers'),

  switch: (provider: string, model: string) =>
    fetchApi<{ success: boolean }>('/providers/switch', {
      method: 'POST',
      body: JSON.stringify({ provider, model }),
    }),
}

// Session API
export const sessionApi = {
  list: () => fetchApi<Session[]>('/sessions'),

  get: (id: string) => fetchApi<Session>(`/sessions/${id}`),

  create: (name?: string) =>
    fetchApi<Session>('/sessions', {
      method: 'POST',
      body: JSON.stringify({ name }),
    }),

  delete: (id: string) =>
    fetchApi<void>(`/sessions/${id}`, {
      method: 'DELETE',
    }),

  sendMessage: (sessionId: string, content: string) =>
    fetchApi<Message>(`/sessions/${sessionId}/messages`, {
      method: 'POST',
      body: JSON.stringify({ content }),
    }),
}

// SPEC API
export const specApi = {
  list: () => fetchApi<Spec[]>('/specs'),

  get: (id: string) => fetchApi<Spec>(`/specs/${id}`),

  create: (spec: Partial<Spec>) =>
    fetchApi<Spec>('/specs', {
      method: 'POST',
      body: JSON.stringify(spec),
    }),

  update: (id: string, spec: Partial<Spec>) =>
    fetchApi<Spec>(`/specs/${id}`, {
      method: 'PUT',
      body: JSON.stringify(spec),
    }),

  delete: (id: string) =>
    fetchApi<void>(`/specs/${id}`, {
      method: 'DELETE',
    }),
}

// Cost API
export const costApi = {
  summary: () => fetchApi<CostSummary>('/costs/summary'),

  history: (days?: number) =>
    fetchApi<CostRecord[]>(`/costs/history${days ? `?days=${days}` : ''}`),

  export: (format: 'csv' | 'json') =>
    fetchApi<Blob>(`/costs/export?format=${format}`),
}

// Health check
export const healthApi = {
  check: () => fetchApi<{ status: string; version: string }>('/health'),
}
