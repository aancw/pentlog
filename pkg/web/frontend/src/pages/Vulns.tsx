import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Plus, Shield, Edit2, Trash2, AlertCircle, AlertTriangle, Info, CheckCircle, XCircle } from 'lucide-react'

interface Vuln {
  id: string
  title: string
  severity: string
  severity_color: string
  status: string
  description: string
  phase: string
  created_at: string
}

export default function Vulns() {
  const [severityFilter, setSeverityFilter] = useState<string>('')
  const [statusFilter, setStatusFilter] = useState<string>('')

  const { data: vulnsData, isLoading } = useQuery({
    queryKey: ['vulns', severityFilter, statusFilter],
    queryFn: async (): Promise<{ vulns: Vuln[], total: number }> => {
      const params = new URLSearchParams()
      if (severityFilter) params.set('severity', severityFilter)
      if (statusFilter) params.set('status', statusFilter)
      const res = await fetch(`/api/vulns?${params.toString()}`)
      return res.json()
    },
  })

  const vulns = vulnsData?.vulns ?? []

  const getSeverityConfig = (severity: string) => {
    const configs: Record<string, { bg: string; text: string; border: string; icon: typeof AlertCircle }> = {
      Critical: { bg: 'bg-red-500/10', text: 'text-red-500', border: 'border-red-500/30', icon: AlertCircle },
      High: { bg: 'bg-orange-500/10', text: 'text-orange-500', border: 'border-orange-500/30', icon: AlertTriangle },
      Medium: { bg: 'bg-yellow-500/10', text: 'text-yellow-500', border: 'border-yellow-500/30', icon: AlertTriangle },
      Low: { bg: 'bg-green-500/10', text: 'text-green-500', border: 'border-green-500/30', icon: Info },
      Info: { bg: 'bg-blue-500/10', text: 'text-blue-500', border: 'border-blue-500/30', icon: Info },
    }
    return configs[severity] || configs.Info
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold text-[hsl(var(--foreground))]">Vulnerabilities</h1>
          <p className="text-[hsl(var(--muted-foreground))] mt-1">
            {vulnsData?.total ?? 0} findings tracked
          </p>
        </div>
        <button className="inline-flex items-center gap-2 px-4 py-2.5 bg-[hsl(var(--primary))] text-[hsl(var(--primary-foreground))] rounded-lg hover:opacity-90 transition-opacity font-medium">
          <Plus className="h-4 w-4" />
          Add Finding
        </button>
      </div>

      {/* Filters */}
      <div className="flex flex-wrap gap-3">
        <select
          value={severityFilter}
          onChange={(e) => setSeverityFilter(e.target.value)}
          className="px-4 py-2 bg-[hsl(var(--card))] border border-[hsl(var(--border))] rounded-lg text-sm text-[hsl(var(--foreground))] focus:outline-none focus:ring-2 focus:ring-[hsl(var(--primary))]"
        >
          <option value="">All Severities</option>
          <option value="Critical">Critical</option>
          <option value="High">High</option>
          <option value="Medium">Medium</option>
          <option value="Low">Low</option>
          <option value="Info">Info</option>
        </select>
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          className="px-4 py-2 bg-[hsl(var(--card))] border border-[hsl(var(--border))] rounded-lg text-sm text-[hsl(var(--foreground))] focus:outline-none focus:ring-2 focus:ring-[hsl(var(--primary))]"
        >
          <option value="">All Statuses</option>
          <option value="Open">Open</option>
          <option value="Closed">Closed</option>
          <option value="Verified">Verified</option>
        </select>
      </div>

      {/* Content */}
      {isLoading ? (
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-[hsl(var(--primary))]"></div>
        </div>
      ) : vulns.length === 0 ? (
        <div className="rounded-xl border border-[hsl(var(--border))] bg-[hsl(var(--card))] p-12">
          <div className="text-center">
            <Shield className="h-12 w-12 mx-auto text-[hsl(var(--muted-foreground))] mb-4" />
            <h2 className="text-xl font-semibold text-[hsl(var(--foreground))] mb-2">No Vulnerabilities</h2>
            <p className="text-[hsl(var(--muted-foreground))]">
              {severityFilter || statusFilter ? 'No findings match your filters' : 'Start tracking security findings'}
            </p>
          </div>
        </div>
      ) : (
        <div className="space-y-4">
          {vulns.map((vuln) => {
            const config = getSeverityConfig(vuln.severity)
            const Icon = config.icon
            return (
              <div key={vuln.id} className={`rounded-xl border ${config.border} bg-[hsl(var(--card))] p-5 hover:border-[hsl(var(--primary))]/50 transition-colors`}>
                <div className="flex items-start gap-4">
                  <div className={`p-2 rounded-lg ${config.bg}`}>
                    <Icon className={`h-5 w-5 ${config.text}`} />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-3 mb-2 flex-wrap">
                      <span className={`px-2.5 py-0.5 text-xs font-semibold rounded ${config.bg} ${config.text}`}>
                        {vuln.severity}
                      </span>
                      <span className={`inline-flex items-center gap-1 px-2.5 py-0.5 text-xs rounded ${
                        vuln.status === 'Open' ? 'bg-yellow-500/10 text-yellow-500' :
                        vuln.status === 'Verified' ? 'bg-green-500/10 text-green-500' :
                        'bg-[hsl(var(--secondary))] text-[hsl(var(--muted-foreground))]'
                      }`}>
                        {vuln.status === 'Verified' && <CheckCircle className="h-3 w-3" />}
                        {vuln.status === 'Closed' && <XCircle className="h-3 w-3" />}
                        {vuln.status}
                      </span>
                      <span className="text-xs text-[hsl(var(--muted-foreground))]">{vuln.id}</span>
                    </div>
                    <h3 className="font-semibold text-lg text-[hsl(var(--foreground))] mb-1">{vuln.title}</h3>
                    {vuln.description && (
                      <p className="text-sm text-[hsl(var(--muted-foreground))] mb-3 line-clamp-2">{vuln.description}</p>
                    )}
                    <div className="flex items-center gap-4 text-xs text-[hsl(var(--muted-foreground))]">
                      {vuln.phase && (
                        <span className="inline-flex items-center gap-1 px-2 py-0.5 bg-[hsl(var(--secondary))] rounded">
                          {vuln.phase}
                        </span>
                      )}
                      <span>{new Date(vuln.created_at).toLocaleDateString()}</span>
                    </div>
                  </div>
                  <div className="flex items-center gap-1">
                    <button className="p-2 hover:bg-[hsl(var(--accent))] rounded-lg transition-colors" title="Edit">
                      <Edit2 className="h-4 w-4 text-[hsl(var(--muted-foreground))]" />
                    </button>
                    <button className="p-2 hover:bg-red-500/10 rounded-lg transition-colors" title="Delete">
                      <Trash2 className="h-4 w-4 text-red-500" />
                    </button>
                  </div>
                </div>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}