import { useSystemStatus } from '../hooks/useApi'
import { Database, Folder, Cpu } from 'lucide-react'

export default function Settings() {
  const { data: status } = useSystemStatus()

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
            <Cpu className="h-5 w-5 text-muted-foreground" />
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

        {status?.context && (
          <div className="rounded-lg border border-border bg-card p-6 md:col-span-2">
            <div className="flex items-center gap-2 mb-4">
              <Folder className="h-5 w-5 text-muted-foreground" />
              <h2 className="text-lg font-semibold">Current Context</h2>
            </div>
            <div className="grid gap-4 md:grid-cols-3">
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