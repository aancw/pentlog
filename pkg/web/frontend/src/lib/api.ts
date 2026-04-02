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

export interface SessionMetadata {
  client: string
  engagement: string
  scope: string
  operator: string
  phase: string
  target: string
  target_ip: string
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
  metadata: SessionMetadata
  tags: string[]
  notes_count: number
  has_gif: boolean
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

export const api = {
  dashboard: {
    stats: async (): Promise<DashboardStats> => {
      const res = await fetch('/api/dashboard/stats')
      return res.json()
    },
    activity: async (): Promise<ActivityResponse> => {
      const res = await fetch('/api/dashboard/activity')
      return res.json()
    },
    clients: async (): Promise<ClientsResponse> => {
      const res = await fetch('/api/dashboard/clients')
      return res.json()
    },
  },
  sessions: {
    list: async (params?: { limit?: number; offset?: number; tag?: string; client?: string; phase?: string }): Promise<SessionsListResponse> => {
      const searchParams = new URLSearchParams()
      if (params?.limit) searchParams.set('limit', String(params.limit))
      if (params?.offset) searchParams.set('offset', String(params.offset))
      if (params?.tag) searchParams.set('tag', params.tag)
      if (params?.client) searchParams.set('client', params.client)
      if (params?.phase) searchParams.set('phase', params.phase)
      const query = searchParams.toString()
      const res = await fetch(`/api/sessions${query ? `?${query}` : ''}`)
      return res.json()
    },
    get: async (id: number): Promise<SessionResponse> => {
      const res = await fetch(`/api/sessions/${id}`)
      return res.json()
    },
    delete: async (id: number): Promise<{ message: string }> => {
      const res = await fetch(`/api/sessions/${id}`, { method: 'DELETE' })
      return res.json()
    },
  },
  system: {
    status: async (): Promise<SystemStatus> => {
      const res = await fetch('/api/system/status')
      return res.json()
    },
    info: async (): Promise<SystemInfo> => {
      const res = await fetch('/api/system/info')
      return res.json()
    },
  },
}