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
      <div>
        <h1 className="text-3xl font-bold">Settings</h1>
        <p className="text-muted-foreground mt-1">
          System configuration and information
        </p>
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        <div className="rounded-lg border border-border bg-card p-6">
          <div className="flex items-center gap-2 mb-4">
            <Server className="h-5 w-5 text-muted-foreground" />
            <h2 className="text-lg font-semibold">System Status</h2>
          </div>
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">Version</span>
              <span className="font-mono text-sm">{status?.version ?? '-'}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">Context</span>
              <span className={`text-sm ${status?.has_context ? 'text-green-500' : 'text-yellow-500'}`}>
                {status?.has_context ? 'Active' : 'Not set'}
              </span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">Total Sessions</span>
              <span className="font-mono text-sm">{status?.total_sessions ?? 0}</span>
            </div>
          </div>
        </div>

        <div className="rounded-lg border border-border bg-card p-6">
          <div className="flex items-center gap-2 mb-4">
            <Database className="h-5 w-5 text-muted-foreground" />
            <h2 className="text-lg font-semibold">Database</h2>
          </div>
          <div className="space-y-3">
            <div>
              <span className="text-sm text-muted-foreground">Path</span>
              <p className="font-mono text-xs bg-muted p-2 rounded mt-1 break-all">
                {status?.db_path ?? '-'}
              </p>
            </div>
          </div>
        </div>

        <div className="rounded-lg border border-border bg-card p-6">
          <div className="flex items-center gap-2 mb-4">
            <HardDrive className="h-5 w-5 text-muted-foreground" />
            <h2 className="text-lg font-semibold">Storage Paths</h2>
          </div>
          <div className="space-y-2">
            <div className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">Home</span>
              <span className="font-mono text-xs truncate max-w-48">{info?.paths.home}</span>
            </div>
            <div className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">Logs</span>
              <span className="font-mono text-xs truncate max-w-48">{info?.paths.logs_dir}</span>
            </div>
            <div className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">Reports</span>
              <span className="font-mono text-xs truncate max-w-48">{info?.paths.reports_dir}</span>
            </div>
            <div className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">Archives</span>
              <span className="font-mono text-xs truncate max-w-48">{info?.paths.archive_dir}</span>
            </div>
          </div>
        </div>

        {status?.context && (
          <div className="rounded-lg border border-border bg-card p-6">
            <div className="flex items-center gap-2 mb-4">
              <Folder className="h-5 w-5 text-muted-foreground" />
              <h2 className="text-lg font-semibold">Current Context</h2>
            </div>
            <div className="grid gap-3 grid-cols-2">
              <div>
                <span className="text-sm text-muted-foreground">Client</span>
                <p className="font-medium">{status.context.client}</p>
              </div>
              <div>
                <span className="text-sm text-muted-foreground">Engagement</span>
                <p className="font-medium">{status.context.engagement}</p>
              </div>
              <div>
                <span className="text-sm text-muted-foreground">Phase</span>
                <p className="font-medium capitalize">{status.context.phase}</p>
              </div>
              <div>
                <span className="text-sm text-muted-foreground">Operator</span>
                <p className="font-medium">{status.context.operator}</p>
              </div>
              <div>
                <span className="text-sm text-muted-foreground">Scope</span>
                <p className="font-medium">{status.context.scope}</p>
              </div>
              <div>
                <span className="text-sm text-muted-foreground">Type</span>
                <p className="font-medium">{status.context.type}</p>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}