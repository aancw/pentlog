import { useQuery } from '@tanstack/react-query'
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

export function useSessions(params?: { limit?: number; offset?: number; tag?: string; client?: string }) {
  return useQuery({
    queryKey: ['sessions', params],
    queryFn: () => api.sessions.list(params),
  })
}

export function useSession(id: number) {
  return useQuery({
    queryKey: ['session', id],
    queryFn: () => api.sessions.get(id),
    enabled: id > 0,
  })
}

export function useSystemStatus() {
  return useQuery({
    queryKey: ['system', 'status'],
    queryFn: api.system.status,
  })
}