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
  timestamp?: string
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

export interface SessionsListResponse {
  sessions: SessionResponse[]
  total: number
  page: number
  has_more: boolean
}

export interface Vulnerability {
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

export interface ActivityResponse {
  recent_sessions: SessionResponse[]
  recent_vulns: Vulnerability[]
}

export interface ClientSummary {
  name: string
  sessions_count: number
  total_size: number
  total_size_human: string
}

export interface EngagementSummary {
  name: string
  sessions_count: number
  total_size: number
  total_size_human: string
}

export interface PhaseSummary {
  name: string
  sessions_count: number
}

export interface SessionNote {
  timestamp: string
  content: string
  byte_offset: number
}

export interface SessionTimelineEntry {
  timestamp: string
  command: string
  output: string
}

export interface SessionContentResponse {
  content: string
  total_bytes: number
  has_more: boolean
}

export interface SearchResult {
  session_id: number
  session_path: string
  line_num: number
  content: string
  is_note: boolean
}

export interface SearchResponse {
  results: SearchResult[]
  total_matches: number
  query: string
  is_regex: boolean
}

export interface SystemStatus {
  has_context: boolean
  context: ContextEntry | null
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

export interface ContextEntry {
  client: string
  engagement: string
  scope: string
  operator: string
  phase: string
  target: string
  target_ip: string
  timestamp: string
  type: string
}

export interface ContextResponse {
  has_context: boolean
  context: ContextEntry | null
}

export interface TargetRecord {
  name: string
  ip: string
  is_current: boolean
}

export interface TargetsResponse {
  targets: TargetRecord[]
  current: TargetRecord | null
}

export interface ReportRecord {
  name: string
  client: string
  path: string
  relative_path: string
  size: number
  size_human: string
  mod_time: string
  type: string
  view_url: string
}

export interface ReportGenerateRequest {
  client?: string
  engagement?: string
  phase?: string
  format?: 'html' | 'md'
  include_gifs?: boolean
  gif_resolution?: '720p' | '1080p'
  output_name?: string
}

export interface ReportGenerateJob {
  id: string
  status: 'queued' | 'running' | 'completed' | 'failed'
  message: string
  error?: string
  client: string
  engagement?: string
  phase?: string
  format: 'html' | 'md'
  include_gifs: boolean
  gif_resolution?: '720p' | '1080p'
  output_name: string
  report_path?: string
  relative_path?: string
  view_url?: string
  sessions_count: number
  gif_generated: number
  gif_failed: number
  avg_time_per_session_secs?: number
  est_time_remaining_secs?: number
  created_at: string
  updated_at: string
}

export interface ArchiveRecord {
  name: string
  client: string
  path: string
  relative_path: string
  size: number
  size_human: string
  mod_time: string
  encrypted: boolean
  download_url: string
}

export interface ShareStatus {
  active: boolean
  pid?: number
  log_file?: string
  watch_url?: string
  status_url?: string
  port?: number
  reachable?: boolean
  clients?: number
  client_ips?: string[]
}

export interface RecoveryStatus {
  crashed: SessionResponse[]
  active: SessionResponse[]
  orphaned: SessionResponse[]
}

export interface TagsResponse {
  tags: Array<{ name: string; count: number }>
}

async function fetchJson<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, {
    headers: {
      'Content-Type': 'application/json',
      ...(init?.headers ?? {}),
    },
    ...init,
  })

  if (!response.ok) {
    const text = await response.text()
    throw new Error(text || `Request failed: ${response.status}`)
  }

  return response.json() as Promise<T>
}

export function formatDate(value?: string) {
  if (!value) return '-'
  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) return value
  return parsed.toLocaleString()
}

export function formatDuration(seconds?: number) {
  if (!seconds || seconds <= 0) return '-'
  if (seconds < 60) return `${seconds}s`
  const mins = Math.floor(seconds / 60)
  const secs = seconds % 60
  if (mins < 60) return secs > 0 ? `${mins}m ${secs}s` : `${mins}m`
  const hours = Math.floor(mins / 60)
  const remainMins = mins % 60
  return remainMins > 0 ? `${hours}h ${remainMins}m` : `${hours}h`
}

export function formatListLabel(label: string) {
  return label.replace(/[-_]/g, ' ')
}

export const api = {
  dashboard: {
    stats: () => fetchJson<DashboardStats>('/api/dashboard/stats'),
    activity: () => fetchJson<ActivityResponse>('/api/dashboard/activity'),
    clients: () => fetchJson<{ clients: ClientSummary[] }>('/api/dashboard/clients'),
    engagements: (client?: string) => {
      const query = client ? `?client=${encodeURIComponent(client)}` : ''
      return fetchJson<{ engagements: EngagementSummary[] }>(`/api/dashboard/engagements${query}`)
    },
    phases: (client?: string, engagement?: string) => {
      const params = new URLSearchParams()
      if (client) params.set('client', client)
      if (engagement) params.set('engagement', engagement)
      const query = params.toString()
      return fetchJson<{ phases: PhaseSummary[] }>(`/api/dashboard/phases${query ? `?${query}` : ''}`)
    },
  },
  sessions: {
    list: (params?: Record<string, string | number | undefined>) => {
      const search = new URLSearchParams()
      Object.entries(params ?? {}).forEach(([key, value]) => {
        if (value !== undefined && value !== '') {
          search.set(key, String(value))
        }
      })
      const query = search.toString()
      return fetchJson<SessionsListResponse>(`/api/sessions${query ? `?${query}` : ''}`)
    },
    get: (id: number | string) => fetchJson<SessionResponse>(`/api/sessions/${id}`),
    delete: (id: number | string) => fetchJson<{ message: string }>(`/api/sessions/${id}`, { method: 'DELETE' }),
    notes: (id: number | string) => fetchJson<{ notes: SessionNote[] }>(`/api/sessions/${id}/notes`),
    addNote: (id: number | string, content: string) => fetchJson<{ message: string }>(`/api/sessions/${id}/notes`, {
      method: 'POST',
      body: JSON.stringify({ content }),
    }),
    timeline: (id: number | string) => fetchJson<{ session_id: number; commands: SessionTimelineEntry[] }>(`/api/sessions/${id}/timeline`),
    content: (id: number | string) => fetchJson<SessionContentResponse>(`/api/sessions/${id}/content`),
    tags: () => fetchJson<TagsResponse>('/api/sessions/tags'),
  },
  vulns: {
    list: () => fetchJson<{ vulns: Vulnerability[]; total: number }>('/api/vulns'),
  },
  search: {
    query: (params: Record<string, string | number | boolean | undefined>) => {
      const search = new URLSearchParams()
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined && value !== '') {
          search.set(key, String(value))
        }
      })
      return fetchJson<SearchResponse>(`/api/search?${search.toString()}`)
    },
  },
  reports: {
    list: () => fetchJson<{ reports: ReportRecord[] }>('/api/reports'),
    generate: (payload: ReportGenerateRequest) =>
      fetchJson<{ job: ReportGenerateJob }>('/api/reports/generate', {
        method: 'POST',
        body: JSON.stringify(payload),
      }),
    job: (id: string) => fetchJson<{ job: ReportGenerateJob }>(`/api/reports/jobs/${encodeURIComponent(id)}`),
    activeJob: () => fetchJson<{ job: ReportGenerateJob | null }>('/api/reports/jobs/active'),
  },
  archives: {
    list: () => fetchJson<{ archives: ArchiveRecord[] }>('/api/archives'),
  },
  share: {
    status: () => fetchJson<ShareStatus>('/api/share/status'),
  },
  system: {
    status: () => fetchJson<SystemStatus>('/api/system/status'),
    info: () => fetchJson<SystemInfo>('/api/system/info'),
  },
  contexts: {
    current: () => fetchJson<ContextResponse>('/api/contexts/current'),
    history: () => fetchJson<{ history: ContextEntry[] }>('/api/contexts/history'),
    create: (payload: Partial<ContextEntry>) => fetchJson<{ message: string; context: ContextEntry }>('/api/contexts/create', {
      method: 'POST',
      body: JSON.stringify(payload),
    }),
    update: (payload: Partial<ContextEntry>) => fetchJson<{ message: string; context: ContextEntry }>('/api/contexts/current', {
      method: 'PUT',
      body: JSON.stringify(payload),
    }),
    reset: () => fetchJson<{ message: string }>('/api/contexts/reset', { method: 'DELETE' }),
  },
  targets: {
    list: () => fetchJson<TargetsResponse>('/api/targets'),
    create: (payload: { name: string; ip: string }) => fetchJson<{ message: string }>('/api/targets', {
      method: 'POST',
      body: JSON.stringify(payload),
    }),
    switch: (name: string) => fetchJson<{ message: string }>(`/api/targets/${encodeURIComponent(name)}/switch`, { method: 'PUT' }),
    delete: (name: string) => fetchJson<{ message: string }>(`/api/targets/${encodeURIComponent(name)}`, { method: 'DELETE' }),
    clear: () => fetchJson<{ message: string }>('/api/targets/clear', { method: 'DELETE' }),
  },
  recovery: {
    status: () => fetchJson<RecoveryStatus>('/api/recovery/status'),
    markStale: (timeoutMinutes = 5) => fetchJson<{ message: string }>('/api/recovery/mark-stale', {
      method: 'POST',
      body: JSON.stringify({ timeout_minutes: timeoutMinutes }),
    }),
    recoverAll: () => fetchJson<{ message: string }>('/api/recovery/recover-all', { method: 'POST' }),
    recoverOne: (id: number) => fetchJson<{ message: string }>(`/api/recovery/recover/${id}`, { method: 'POST' }),
    deleteOrphans: () => fetchJson<{ message: string }>('/api/recovery/orphans', { method: 'DELETE' }),
  },
}
