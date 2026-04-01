const API_BASE = '/api'

export async function fetchApi<T>(path: string, options?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  })

  if (!response.ok) {
    throw new Error(`API Error: ${response.status} ${response.statusText}`)
  }

  return response.json()
}

export const api = {
  dashboard: {
    stats: () => fetchApi<DashboardStats>('/dashboard/stats'),
    activity: () => fetchApi<ActivityResponse>('/dashboard/activity'),
    clients: () => fetchApi<ClientsResponse>('/dashboard/clients'),
  },
  sessions: {
    list: (params?: { limit?: number; offset?: number; tag?: string; client?: string; phase?: string }) => {
      const searchParams = new URLSearchParams()
      if (params?.limit) searchParams.set('limit', String(params.limit))
      if (params?.offset) searchParams.set('offset', String(params.offset))
      if (params?.tag) searchParams.set('tag', params.tag)
      if (params?.client) searchParams.set('client', params.client)
      if (params?.phase) searchParams.set('phase', params.phase)
      const query = searchParams.toString()
      return fetchApi<SessionsListResponse>(`/sessions${query ? `?${query}` : ''}`)
    },
    get: (id: number) => fetchApi<SessionResponse>(`/sessions/${id}`),
    delete: (id: number) => 
      fetchApi<{ message: string }>(`/sessions/${id}`, { method: 'DELETE' }),
  },
  system: {
    status: () => fetchApi<SystemStatus>('/system/status'),
    info: () => fetchApi<SystemInfo>('/system/info'),
  },
}

export interface DashboardStats {
  total_sessions: number
  total_size: number
  total_size_human: string
  unique_clients: number
  unique_engagements: number
  total_notes: number
  total_vulns: number
  phase_counts: Record<string, number>
  severity_counts: Record<string, number>
}

export interface ActivityResponse {
  recent_sessions: SessionResponse[]
  recent_vulns: VulnResponse[]
}

export interface ClientsResponse {
  clients: Array<{
    name: string
    sessions_count: number
    total_size: number
    total_size_human: string
  }>
}

export interface SessionResponse {
  id: number
  filename: string
  path: string
  display_path: string
  size: number
  size_human: string
  mod_time: string
  state: string
  metadata: {
    client: string
    engagement: string
    scope: string
    operator: string
    phase: string
    target: string
    target_ip: string
  }
  tags: string[]
  notes_count: number
  has_gif: boolean
}

export interface SessionsListResponse {
  sessions: SessionResponse[]
  total: number
  page: number
  has_more: boolean
}

export interface VulnResponse {
  id: string
  title: string
  severity: string
  severity_color: string
  status: string
  description: string
  remediation: string
  references: string[]
  evidence: string[]
  phase: string
  created_at: string
  updated_at: string
}

export interface SystemStatus {
  has_context: boolean
  context: {
    client: string
    engagement: string
    scope: string
    operator: string
    phase: string
    target: string
    target_ip: string
    timestamp: string
    type: string
  } | null
  version: string
  db_path: string
  total_sessions: number
}

export interface SystemInfo {
  version: string
  paths: {
    home: string
    logs_dir: string
    reports_dir: string
    archive_dir: string
    database_file: string
  }
  uptime: string
}