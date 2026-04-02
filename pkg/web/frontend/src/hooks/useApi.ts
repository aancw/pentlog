import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { api } from '../lib/api'

export function useDashboardStats() {
  return useQuery({
    queryKey: ['dashboard', 'stats'],
    queryFn: api.dashboard.stats,
  })
}

export function useDashboardActivity() {
  return useQuery({
    queryKey: ['dashboard', 'activity'],
    queryFn: api.dashboard.activity,
  })
}

export function useDashboardClients() {
  return useQuery({
    queryKey: ['dashboard', 'clients'],
    queryFn: api.dashboard.clients,
  })
}

export function useDashboardEngagements(client?: string) {
  return useQuery({
    queryKey: ['dashboard', 'engagements', client],
    queryFn: () => api.dashboard.engagements(client),
  })
}

export function useSessions(params?: Record<string, string | number | undefined>) {
  return useQuery({
    queryKey: ['sessions', params],
    queryFn: () => api.sessions.list(params),
    placeholderData: keepPreviousData,
  })
}

export function useSession(id?: number | string) {
  return useQuery({
    queryKey: ['session', id],
    queryFn: () => api.sessions.get(id as number | string),
    enabled: Boolean(id),
  })
}

export function useSessionNotes(id?: number | string) {
  return useQuery({
    queryKey: ['session', id, 'notes'],
    queryFn: () => api.sessions.notes(id as number | string),
    enabled: Boolean(id),
  })
}

export function useSessionTimeline(id?: number | string) {
  return useQuery({
    queryKey: ['session', id, 'timeline'],
    queryFn: () => api.sessions.timeline(id as number | string),
    enabled: Boolean(id),
  })
}

export function useSessionContent(id?: number | string) {
  return useQuery({
    queryKey: ['session', id, 'content'],
    queryFn: () => api.sessions.content(id as number | string),
    enabled: Boolean(id),
  })
}

export function useSessionTags() {
  return useQuery({
    queryKey: ['sessions', 'tags'],
    queryFn: api.sessions.tags,
  })
}

export function useVulns() {
  return useQuery({
    queryKey: ['vulns'],
    queryFn: api.vulns.list,
  })
}

export function useReports() {
  return useQuery({
    queryKey: ['reports'],
    queryFn: api.reports.list,
  })
}

export function useArchives() {
  return useQuery({
    queryKey: ['archives'],
    queryFn: api.archives.list,
  })
}

export function useShareStatus() {
  return useQuery({
    queryKey: ['share', 'status'],
    queryFn: api.share.status,
    refetchInterval: 5000,
  })
}

export function useSystemStatus() {
  return useQuery({
    queryKey: ['system', 'status'],
    queryFn: api.system.status,
  })
}

export function useSystemInfo() {
  return useQuery({
    queryKey: ['system', 'info'],
    queryFn: api.system.info,
  })
}

export function useCurrentContext() {
  return useQuery({
    queryKey: ['contexts', 'current'],
    queryFn: api.contexts.current,
  })
}

export function useContextHistory() {
  return useQuery({
    queryKey: ['contexts', 'history'],
    queryFn: api.contexts.history,
  })
}

export function useTargets() {
  return useQuery({
    queryKey: ['targets'],
    queryFn: api.targets.list,
  })
}

export function useRecoveryStatus() {
  return useQuery({
    queryKey: ['recovery', 'status'],
    queryFn: api.recovery.status,
  })
}

export { api }
export type {
  ActivityResponse,
  ArchiveRecord,
  ContextEntry,
  DashboardStats,
  RecoveryStatus,
  ReportRecord,
  SearchResponse,
  SessionContentResponse,
  SessionNote,
  SessionResponse,
  ShareStatus,
  SessionTimelineEntry,
  SessionsListResponse,
  SystemInfo,
  SystemStatus,
  TargetRecord,
  Vulnerability,
} from '../lib/api'
