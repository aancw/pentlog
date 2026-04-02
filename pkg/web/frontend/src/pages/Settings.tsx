import { useQuery } from '@tanstack/react-query'
import { Database, Folder, Server, HardDrive } from 'lucide-react'

export default function Settings() {
  const { data: status } = useQuery({
    queryKey: ['system', 'status'],
    queryFn: async () => {
      const res = await fetch('/api/system/status')
      return res.json()
    },
  })

  const { data: info } = useQuery({
    queryKey: ['system', 'info'],
    queryFn: async () => {
      const res = await fetch('/api/system/info')
      return res.json()
    },
  })

  return (
    <div className="space-y-6">
      <div className="page-header">
        <h1>Settings</h1>
        <p className="text-muted">System configuration</p>
      </div>

      <div className="grid gap-6" style={{ gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))' }}>
        <div className="card">
          <div className="flex items-center gap-3 mb-4">
            <Server className="h-5 w-5 text-primary" />
            <h2 className="text-lg font-semibold">System Status</h2>
          </div>
          <div className="space-y-3">
            <div className="flex justify-between py-2 border-b border-border">
              <span className="text-muted">Version</span>
              <span className="font-mono">{status?.version || '-'}</span>
            </div>
            <div className="flex justify-between py-2 border-b border-border">
              <span className="text-muted">Context</span>
              <span className={status?.has_context ? 'text-green' : 'text-yellow'}>
                {status?.has_context ? '● Active' : '○ Not set'}
              </span>
            </div>
            <div className="flex justify-between py-2">
              <span className="text-muted">Total Sessions</span>
              <span className="font-mono">{status?.total_sessions || 0}</span>
            </div>
          </div>
        </div>

        <div className="card">
          <div className="flex items-center gap-3 mb-4">
            <Database className="h-5 w-5 text-blue" />
            <h2 className="text-lg font-semibold">Database</h2>
          </div>
          <div className="text-xs text-muted mb-2">Path</div>
          <div className="font-mono text-xs bg-secondary p-3 rounded break-all">
            {status?.db_path || '-'}
          </div>
        </div>

        <div className="card">
          <div className="flex items-center gap-3 mb-4">
            <HardDrive className="h-5 w-5 text-green" />
            <h2 className="text-lg font-semibold">Storage Paths</h2>
          </div>
          <div className="space-y-2 text-sm">
            <div className="flex justify-between">
              <span className="text-muted">Home</span>
              <span className="font-mono text-xs truncate max-w-48">{info?.paths?.home}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted">Logs</span>
              <span className="font-mono text-xs truncate max-w-48">{info?.paths?.logs_dir}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted">Reports</span>
              <span className="font-mono text-xs truncate max-w-48">{info?.paths?.reports_dir}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted">Archives</span>
              <span className="font-mono text-xs truncate max-w-48">{info?.paths?.archive_dir}</span>
            </div>
          </div>
        </div>

        {status?.context && (
          <div className="card">
            <div className="flex items-center gap-3 mb-4">
              <Folder className="h-5 w-5 text-purple-400" />
              <h2 className="text-lg font-semibold">Current Context</h2>
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <div className="text-xs text-muted">Client</div>
                <div className="font-medium">{status.context.client}</div>
              </div>
              <div>
                <div className="text-xs text-muted">Engagement</div>
                <div className="font-medium">{status.context.engagement}</div>
              </div>
              <div>
                <div className="text-xs text-muted">Phase</div>
                <div className="font-medium capitalize">{status.context.phase}</div>
              </div>
              <div>
                <div className="text-xs text-muted">Operator</div>
                <div className="font-medium">{status.context.operator}</div>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}