import { useDashboardStats, useDashboardActivity } from '../hooks/useApi'
import { 
  Activity, 
  Database, 
  FileText, 
  Shield, 
  Users,
  FolderOpen
} from 'lucide-react'

export default function Dashboard() {
  const { data: stats, isLoading: statsLoading } = useDashboardStats()
  const { data: activity, isLoading: activityLoading } = useDashboardActivity()

  if (statsLoading || activityLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-muted-foreground">Loading...</div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Dashboard</h1>
        <p className="text-muted-foreground mt-1">
          Overview of your penetration testing sessions
        </p>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <StatCard
          title="Total Sessions"
          value={stats?.total_sessions ?? 0}
          icon={FolderOpen}
          description="Recorded sessions"
        />
        <StatCard
          title="Evidence Size"
          value={stats?.total_size_human ?? '0 B'}
          icon={Database}
          description="Total data captured"
        />
        <StatCard
          title="Clients"
          value={stats?.unique_clients ?? 0}
          icon={Users}
          description="Unique clients"
        />
        <StatCard
          title="Vulnerabilities"
          value={stats?.total_vulns ?? 0}
          icon={Shield}
          description="Total findings"
        />
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        <div className="rounded-lg border border-border bg-card p-6">
          <h2 className="text-lg font-semibold mb-4">Phase Distribution</h2>
          {stats?.phase_counts && Object.keys(stats.phase_counts).length > 0 ? (
            <div className="space-y-3">
              {Object.entries(stats.phase_counts).map(([phase, count]) => (
                <div key={phase} className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground capitalize">{phase}</span>
                  <div className="flex items-center gap-2">
                    <div className="h-2 w-32 bg-secondary rounded-full overflow-hidden">
                      <div 
                        className="h-full bg-primary rounded-full" 
                        style={{ width: `${Math.min((count / stats.total_sessions) * 100, 100)}%` }}
                      />
                    </div>
                    <span className="text-sm font-medium w-8 text-right">{count}</span>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-muted-foreground text-sm">No phase data available</p>
          )}
        </div>

        <div className="rounded-lg border border-border bg-card p-6">
          <h2 className="text-lg font-semibold mb-4">Severity Distribution</h2>
          {stats?.severity_counts && Object.keys(stats.severity_counts).length > 0 ? (
            <div className="space-y-3">
              {Object.entries(stats.severity_counts).map(([severity, count]) => (
                <div key={severity} className="flex items-center justify-between">
                  <span className={`text-sm font-medium ${getSeverityColor(severity)}`}>
                    {severity}
                  </span>
                  <span className="text-sm text-muted-foreground">{count}</span>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-muted-foreground text-sm">No vulnerability data available</p>
          )}
        </div>
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        <div className="rounded-lg border border-border bg-card p-6">
          <div className="flex items-center gap-2 mb-4">
            <Activity className="h-5 w-5 text-muted-foreground" />
            <h2 className="text-lg font-semibold">Recent Sessions</h2>
          </div>
          {activity?.recent_sessions && activity.recent_sessions.length > 0 ? (
            <div className="space-y-3">
              {activity.recent_sessions.slice(0, 5).map((session) => (
                <div key={session.id} className="flex items-center justify-between text-sm">
                  <div className="truncate flex-1">
                    <span className="font-medium">{session.metadata.client}</span>
                    <span className="text-muted-foreground"> / {session.metadata.phase}</span>
                  </div>
                  <span className="text-muted-foreground ml-4">{session.size_human}</span>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-muted-foreground text-sm">No recent sessions</p>
          )}
        </div>

        <div className="rounded-lg border border-border bg-card p-6">
          <div className="flex items-center gap-2 mb-4">
            <FileText className="h-5 w-5 text-muted-foreground" />
            <h2 className="text-lg font-semibold">Recent Notes</h2>
          </div>
          <p className="text-muted-foreground text-sm">{stats?.total_notes ?? 0} total notes recorded</p>
        </div>
      </div>
    </div>
  )
}

function StatCard({ title, value, icon: Icon, description }: {
  title: string
  value: string | number
  icon: React.ComponentType<{ className?: string }>
  description: string
}) {
  return (
    <div className="rounded-lg border border-border bg-card p-6">
      <div className="flex items-center justify-between">
        <span className="text-sm font-medium text-muted-foreground">{title}</span>
        <Icon className="h-4 w-4 text-muted-foreground" />
      </div>
      <div className="mt-2">
        <span className="text-2xl font-bold">{value}</span>
      </div>
      <p className="text-xs text-muted-foreground mt-1">{description}</p>
    </div>
  )
}

function getSeverityColor(severity: string): string {
  switch (severity.toLowerCase()) {
    case 'critical':
      return 'text-red-500'
    case 'high':
      return 'text-orange-500'
    case 'medium':
      return 'text-yellow-500'
    case 'low':
      return 'text-green-500'
    default:
      return 'text-blue-500'
  }
}