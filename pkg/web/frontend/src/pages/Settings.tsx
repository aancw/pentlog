import { useQuery } from '@tanstack/react-query'
import { Database, Folder, Server, HardDrive } from 'lucide-react'

interface Context {
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

interface SystemStatus {
  has_context: boolean
  context: Context | null
  version: string
  db_path: string
  total_sessions: number
}

interface SystemInfo {
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

export default function Settings() {
  const { data: status } = useQuery({
    queryKey: ['system', 'status'],
    queryFn: async (): Promise<SystemStatus> => {
      const res = await fetch('/api/system/status')
      return res.json()
    },
  })

  const { data: info } = useQuery({
    queryKey: ['system', 'info'],
    queryFn: async (): Promise<SystemInfo> => {
      const res = await fetch('/api/system/info')
      return res.json()
    },
  })

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold text-[hsl(var(--foreground))]">Settings</h1>
        <p className="text-[hsl(var(--muted-foreground))] mt-1">
          System configuration and information
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* System Status */}
        <div className="rounded-xl border border-[hsl(var(--border))] bg-[hsl(var(--card))] p-6">
          <div className="flex items-center gap-3 mb-6">
            <div className="p-2 rounded-lg bg-[hsl(var(--primary))]/10">
              <Server className="h-5 w-5 text-[hsl(var(--primary))]" />
            </div>
            <h2 className="text-lg font-semibold text-[hsl(var(--foreground))]">System Status</h2>
          </div>
          <div className="space-y-4">
            <div className="flex items-center justify-between py-2 border-b border-[hsl(var(--border))]">
              <span className="text-sm text-[hsl(var(--muted-foreground))]">Version</span>
              <span className="font-mono text-sm text-[hsl(var(--foreground))]">{status?.version ?? '-'}</span>
            </div>
            <div className="flex items-center justify-between py-2 border-b border-[hsl(var(--border))]">
              <span className="text-sm text-[hsl(var(--muted-foreground))]">Context Status</span>
              <span className={`inline-flex items-center gap-1 text-sm ${status?.has_context ? 'text-green-500' : 'text-yellow-500'}`}>
                {status?.has_context ? '●' : '○'} {status?.has_context ? 'Active' : 'Not set'}
              </span>
            </div>
            <div className="flex items-center justify-between py-2">
              <span className="text-sm text-[hsl(var(--muted-foreground))]">Total Sessions</span>
              <span className="font-mono text-sm text-[hsl(var(--foreground))]">{status?.total_sessions ?? 0}</span>
            </div>
          </div>
        </div>

        {/* Database */}
        <div className="rounded-xl border border-[hsl(var(--border))] bg-[hsl(var(--card))] p-6">
          <div className="flex items-center gap-3 mb-6">
            <div className="p-2 rounded-lg bg-blue-500/10">
              <Database className="h-5 w-5 text-blue-500" />
            </div>
            <h2 className="text-lg font-semibold text-[hsl(var(--foreground))]">Database</h2>
          </div>
          <div>
            <span className="text-xs text-[hsl(var(--muted-foreground))] uppercase tracking-wider">Path</span>
            <p className="mt-2 font-mono text-xs text-[hsl(var(--foreground))] bg-[hsl(var(--secondary))] p-3 rounded-lg break-all">
              {status?.db_path ?? '-'}
            </p>
          </div>
        </div>

        {/* Storage Paths */}
        <div className="rounded-xl border border-[hsl(var(--border))] bg-[hsl(var(--card))] p-6">
          <div className="flex items-center gap-3 mb-6">
            <div className="p-2 rounded-lg bg-green-500/10">
              <HardDrive className="h-5 w-5 text-green-500" />
            </div>
            <h2 className="text-lg font-semibold text-[hsl(var(--foreground))]">Storage Paths</h2>
          </div>
          <div className="space-y-3">
            {info?.paths && Object.entries({
              'Home': info.paths.home,
              'Logs': info.paths.logs_dir,
              'Reports': info.paths.reports_dir,
              'Archives': info.paths.archive_dir,
            }).map(([label, path]) => (
              <div key={label} className="flex items-center justify-between">
                <span className="text-sm text-[hsl(var(--muted-foreground))]">{label}</span>
                <span className="font-mono text-xs text-[hsl(var(--foreground))] truncate max-w-48" title={path}>{path}</span>
              </div>
            ))}
          </div>
        </div>

        {/* Current Context */}
        {status?.context && (
          <div className="rounded-xl border border-[hsl(var(--border))] bg-[hsl(var(--card))] p-6">
            <div className="flex items-center gap-3 mb-6">
              <div className="p-2 rounded-lg bg-purple-500/10">
                <Folder className="h-5 w-5 text-purple-500" />
              </div>
              <h2 className="text-lg font-semibold text-[hsl(var(--foreground))]">Current Context</h2>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <span className="text-xs text-[hsl(var(--muted-foreground))]">Client</span>
                <p className="font-medium text-[hsl(var(--foreground))]">{status.context.client}</p>
              </div>
              <div>
                <span className="text-xs text-[hsl(var(--muted-foreground))]">Engagement</span>
                <p className="font-medium text-[hsl(var(--foreground))]">{status.context.engagement}</p>
              </div>
              <div>
                <span className="text-xs text-[hsl(var(--muted-foreground))]">Phase</span>
                <p className="font-medium text-[hsl(var(--foreground))] capitalize">{status.context.phase}</p>
              </div>
              <div>
                <span className="text-xs text-[hsl(var(--muted-foreground))]">Operator</span>
                <p className="font-medium text-[hsl(var(--foreground))]">{status.context.operator}</p>
              </div>
              <div>
                <span className="text-xs text-[hsl(var(--muted-foreground))]">Scope</span>
                <p className="font-medium text-[hsl(var(--foreground))]">{status.context.scope}</p>
              </div>
              <div>
                <span className="text-xs text-[hsl(var(--muted-foreground))]">Type</span>
                <p className="font-medium text-[hsl(var(--foreground))]">{status.context.type}</p>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}