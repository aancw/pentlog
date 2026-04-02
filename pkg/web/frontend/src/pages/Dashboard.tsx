import { useDashboardStats, useDashboardActivity } from '../hooks/useApi'
import { FolderOpen, Database, Users, Shield } from 'lucide-react'

export default function Dashboard() {
  const { data: stats, isLoading: statsLoading } = useDashboardStats()
  const { data: activity, isLoading: activityLoading } = useDashboardActivity()

  if (statsLoading || activityLoading) {
    return (
      <div className="text-center p-8">
        <div className="animate-spin inline-block w-8 h-8 border-2 border-primary rounded-full" style={{ borderTopColor: 'transparent' }}></div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="page-header">
        <h1>Dashboard</h1>
        <p className="text-muted">Overview of your penetration testing activity</p>
      </div>

      <div className="stats-grid">
        <div className="stat-card">
          <div className="flex items-center justify-between">
            <span className="label">Total Sessions</span>
            <FolderOpen className="icon h-5 w-5" />
          </div>
          <div className="value">{stats?.total_sessions ?? 0}</div>
        </div>

        <div className="stat-card">
          <div className="flex items-center justify-between">
            <span className="label">Evidence Size</span>
            <Database className="icon h-5 w-5" />
          </div>
          <div className="value">{stats?.total_size_human ?? '0 B'}</div>
        </div>

        <div className="stat-card">
          <div className="flex items-center justify-between">
            <span className="label">Clients</span>
            <Users className="icon h-5 w-5" />
          </div>
          <div className="value">{stats?.unique_clients ?? 0}</div>
        </div>

        <div className="stat-card">
          <div className="flex items-center justify-between">
            <span className="label">Vulnerabilities</span>
            <Shield className="icon h-5 w-5" />
          </div>
          <div className="value">{stats?.total_vulns ?? 0}</div>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-6" style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))' }}>
        <div className="card">
          <h2 className="text-lg font-semibold mb-4">Phase Distribution</h2>
          {stats?.phase_counts && Object.keys(stats.phase_counts).length > 0 ? (
            <div className="space-y-3">
              {Object.entries(stats.phase_counts).map(([phase, count]) => (
                <div key={phase}>
                  <div className="flex justify-between text-sm mb-1">
                    <span className="text-muted capitalize">{phase}</span>
                    <span className="font-medium">{count}</span>
                  </div>
                  <div className="h-2 bg-secondary rounded">
                    <div 
                      className="h-full bg-primary rounded"
                      style={{ width: `${Math.min((count / (stats?.total_sessions || 1)) * 100, 100)}%` }}
                    />
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-muted text-sm">No phase data available</p>
          )}
        </div>

        <div className="card">
          <h2 className="text-lg font-semibold mb-4">Recent Sessions</h2>
          {activity?.recent_sessions && activity.recent_sessions.length > 0 ? (
            <div className="space-y-3">
              {activity.recent_sessions.slice(0, 5).map((session) => (
                <div key={session.id} className="p-3 bg-secondary rounded">
                  <div className="font-medium truncate">{session.metadata.client}</div>
                  <div className="text-sm text-muted">
                    {session.metadata.engagement} / {session.metadata.phase}
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-muted text-sm">No recent sessions</p>
          )}
        </div>
      </div>
    </div>
  )
}