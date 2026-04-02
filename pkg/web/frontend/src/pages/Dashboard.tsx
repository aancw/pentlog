import { useDashboardStats, useDashboardActivity } from '../hooks/useApi'
import { 
  Activity, 
  Database, 
  FileText, 
  Shield, 
  Users,
  FolderOpen,
  TrendingUp,
  Clock
} from 'lucide-react'

export default function Dashboard() {
  const { data: stats, isLoading: statsLoading } = useDashboardStats()
  const { data: activity, isLoading: activityLoading } = useDashboardActivity()

  if (statsLoading || activityLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-[hsl(var(--primary))]"></div>
      </div>
    )
  }

  return (
    <div className="space-y-8">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold text-[hsl(var(--foreground))]">Dashboard</h1>
        <p className="text-[hsl(var(--muted-foreground))] mt-1">
          Overview of your penetration testing activity
        </p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          title="Total Sessions"
          value={stats?.total_sessions ?? 0}
          icon={FolderOpen}
          color="purple"
        />
        <StatCard
          title="Evidence Size"
          value={stats?.total_size_human ?? '0 B'}
          icon={Database}
          color="blue"
        />
        <StatCard
          title="Clients"
          value={stats?.unique_clients ?? 0}
          icon={Users}
          color="green"
        />
        <StatCard
          title="Vulnerabilities"
          value={stats?.total_vulns ?? 0}
          icon={Shield}
          color="red"
        />
      </div>

      {/* Charts Row */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Phase Distribution */}
        <div className="rounded-xl border border-[hsl(var(--border))] bg-[hsl(var(--card))] p-6">
          <h2 className="text-lg font-semibold text-[hsl(var(--foreground))] mb-6">Phase Distribution</h2>
          {stats?.phase_counts && Object.keys(stats.phase_counts).length > 0 ? (
            <div className="space-y-4">
              {Object.entries(stats.phase_counts).map(([phase, count]) => {
                const percentage = stats.total_sessions > 0 ? (count / stats.total_sessions) * 100 : 0
                return (
                  <div key={phase}>
                    <div className="flex items-center justify-between mb-2">
                      <span className="text-sm text-[hsl(var(--muted-foreground))] capitalize">{phase}</span>
                      <span className="text-sm font-medium text-[hsl(var(--foreground))]">{count}</span>
                    </div>
                    <div className="h-2 bg-[hsl(var(--secondary))] rounded-full overflow-hidden">
                      <div 
                        className="h-full bg-[hsl(var(--primary))] rounded-full transition-all duration-500" 
                        style={{ width: `${percentage}%` }}
                      />
                    </div>
                  </div>
                )
              })}
            </div>
          ) : (
            <p className="text-[hsl(var(--muted-foreground))] text-sm">No phase data available</p>
          )}
        </div>

        {/* Severity Distribution */}
        <div className="rounded-xl border border-[hsl(var(--border))] bg-[hsl(var(--card))] p-6">
          <h2 className="text-lg font-semibold text-[hsl(var(--foreground))] mb-6">Vulnerability Severity</h2>
          {stats?.severity_counts && Object.keys(stats.severity_counts).length > 0 ? (
            <div className="grid grid-cols-2 gap-4">
              {Object.entries(stats.severity_counts).map(([severity, count]) => (
                <div 
                  key={severity}
                  className={`p-4 rounded-lg border ${
                    severity === 'Critical' ? 'bg-red-500/10 border-red-500/30' :
                    severity === 'High' ? 'bg-orange-500/10 border-orange-500/30' :
                    severity === 'Medium' ? 'bg-yellow-500/10 border-yellow-500/30' :
                    severity === 'Low' ? 'bg-green-500/10 border-green-500/30' :
                    'bg-blue-500/10 border-blue-500/30'
                  }`}
                >
                  <div className={`text-2xl font-bold ${
                    severity === 'Critical' ? 'text-red-500' :
                    severity === 'High' ? 'text-orange-500' :
                    severity === 'Medium' ? 'text-yellow-500' :
                    severity === 'Low' ? 'text-green-500' :
                    'text-blue-500'
                  }`}>{count}</div>
                  <div className="text-sm text-[hsl(var(--muted-foreground))]">{severity}</div>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-[hsl(var(--muted-foreground))] text-sm">No vulnerability data</p>
          )}
        </div>
      </div>

      {/* Recent Activity */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Recent Sessions */}
        <div className="rounded-xl border border-[hsl(var(--border))] bg-[hsl(var(--card))] p-6">
          <div className="flex items-center gap-2 mb-6">
            <Clock className="h-5 w-5 text-[hsl(var(--muted-foreground))]" />
            <h2 className="text-lg font-semibold text-[hsl(var(--foreground))]">Recent Sessions</h2>
          </div>
          {activity?.recent_sessions && activity.recent_sessions.length > 0 ? (
            <div className="space-y-3">
              {activity.recent_sessions.slice(0, 5).map((session) => (
                <div key={session.id} className="flex items-center justify-between p-3 rounded-lg bg-[hsl(var(--secondary))]">
                  <div className="flex-1 min-w-0">
                    <div className="font-medium text-[hsl(var(--foreground))] truncate">{session.metadata.client}</div>
                    <div className="text-sm text-[hsl(var(--muted-foreground))]">
                      {session.metadata.engagement} / {session.metadata.phase}
                    </div>
                  </div>
                  <div className="text-sm text-[hsl(var(--muted-foreground))] ml-4">{session.size_human}</div>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-[hsl(var(--muted-foreground))] text-sm">No recent sessions</p>
          )}
        </div>

        {/* Quick Stats */}
        <div className="rounded-xl border border-[hsl(var(--border))] bg-[hsl(var(--card))] p-6">
          <div className="flex items-center gap-2 mb-6">
            <TrendingUp className="h-5 w-5 text-[hsl(var(--muted-foreground))]" />
            <h2 className="text-lg font-semibold text-[hsl(var(--foreground))]">Quick Stats</h2>
          </div>
          <div className="space-y-4">
            <div className="flex items-center justify-between p-4 rounded-lg bg-[hsl(var(--secondary))]">
              <div className="flex items-center gap-3">
                <FileText className="h-5 w-5 text-[hsl(var(--primary))]" />
                <span className="text-[hsl(var(--muted-foreground))]">Notes Recorded</span>
              </div>
              <span className="text-xl font-bold text-[hsl(var(--foreground))]">{stats?.total_notes ?? 0}</span>
            </div>
            <div className="flex items-center justify-between p-4 rounded-lg bg-[hsl(var(--secondary))]">
              <div className="flex items-center gap-3">
                <Shield className="h-5 w-5 text-[hsl(var(--primary))]" />
                <span className="text-[hsl(var(--muted-foreground))]">Total Findings</span>
              </div>
              <span className="text-xl font-bold text-[hsl(var(--foreground))]">{stats?.total_vulns ?? 0}</span>
            </div>
            <div className="flex items-center justify-between p-4 rounded-lg bg-[hsl(var(--secondary))]">
              <div className="flex items-center gap-3">
                <Activity className="h-5 w-5 text-[hsl(var(--primary))]" />
                <span className="text-[hsl(var(--muted-foreground))]">Engagements</span>
              </div>
              <span className="text-xl font-bold text-[hsl(var(--foreground))]">{stats?.unique_engagements ?? 0}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

function StatCard({ title, value, icon: Icon, color }: {
  title: string
  value: string | number
  icon: React.ComponentType<{ className?: string }>
  color: 'purple' | 'blue' | 'green' | 'red'
}) {
  const colorClasses = {
    purple: 'from-purple-500/20 to-purple-500/5 border-purple-500/30',
    blue: 'from-blue-500/20 to-blue-500/5 border-blue-500/30',
    green: 'from-green-500/20 to-green-500/5 border-green-500/30',
    red: 'from-red-500/20 to-red-500/5 border-red-500/30',
  }
  
  const iconColors = {
    purple: 'text-purple-400',
    blue: 'text-blue-400',
    green: 'text-green-400',
    red: 'text-red-400',
  }

  return (
    <div className={`rounded-xl border bg-gradient-to-br ${colorClasses[color]} p-6`}>
      <div className="flex items-center justify-between">
        <Icon className={`h-5 w-5 ${iconColors[color]}`} />
      </div>
      <div className="mt-4">
        <span className="text-3xl font-bold text-[hsl(var(--foreground))]">{value}</span>
      </div>
      <p className="text-sm text-[hsl(var(--muted-foreground))] mt-1">{title}</p>
    </div>
  )
}