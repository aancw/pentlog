import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Plus, Shield, Edit2, Trash2 } from 'lucide-react'

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

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Vulnerabilities</h1>
          <p className="text-muted-foreground mt-1">
            {vulnsData?.total ?? 0} findings tracked
          </p>
        </div>
        <button className="flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90">
          <Plus className="h-4 w-4" />
          Add Finding
        </button>
      </div>

      <div className="flex gap-4">
        <select
          value={severityFilter}
          onChange={(e) => setSeverityFilter(e.target.value)}
          className="px-3 py-2 border border-border rounded-lg bg-background text-sm"
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
          className="px-3 py-2 border border-border rounded-lg bg-background text-sm"
        >
          <option value="">All Statuses</option>
          <option value="Open">Open</option>
          <option value="Closed">Closed</option>
          <option value="Verified">Verified</option>
        </select>
      </div>

      {isLoading ? (
        <div className="text-center py-12 text-muted-foreground">Loading...</div>
      ) : vulns.length === 0 ? (
        <div className="text-center py-12">
          <Shield className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
          <p className="text-muted-foreground">No vulnerabilities found</p>
        </div>
      ) : (
        <div className="space-y-4">
          {vulns.map((vuln) => (
            <div key={vuln.id} className="rounded-lg border border-border bg-card p-4">
              <div className="flex items-start justify-between">
                <div className="flex-1">
                  <div className="flex items-center gap-3 mb-2">
                    <span className={`px-2 py-0.5 text-xs rounded font-medium ${
                      vuln.severity === 'Critical' ? 'bg-red-500/10 text-red-500' :
                      vuln.severity === 'High' ? 'bg-orange-500/10 text-orange-500' :
                      vuln.severity === 'Medium' ? 'bg-yellow-500/10 text-yellow-500' :
                      vuln.severity === 'Low' ? 'bg-green-500/10 text-green-500' :
                      'bg-blue-500/10 text-blue-500'
                    }`}>
                      {vuln.severity}
                    </span>
                    <span className={`px-2 py-0.5 text-xs rounded ${
                      vuln.status === 'Open' ? 'bg-yellow-500/10 text-yellow-600' :
                      vuln.status === 'Verified' ? 'bg-green-500/10 text-green-600' :
                      'bg-gray-500/10 text-gray-600'
                    }`}>
                      {vuln.status}
                    </span>
                    <span className="text-xs text-muted-foreground">{vuln.id}</span>
                  </div>
                  <h3 className="font-semibold text-lg mb-1">{vuln.title}</h3>
                  {vuln.description && (
                    <p className="text-sm text-muted-foreground mb-2">{vuln.description}</p>
                  )}
                  <div className="flex items-center gap-4 text-xs text-muted-foreground">
                    {vuln.phase && <span>Phase: {vuln.phase}</span>}
                    <span>Created: {new Date(vuln.created_at).toLocaleDateString()}</span>
                  </div>
                </div>
                <div className="flex items-center gap-2 ml-4">
                  <button className="p-2 hover:bg-accent rounded" title="Edit">
                    <Edit2 className="h-4 w-4 text-muted-foreground" />
                  </button>
                  <button className="p-2 hover:bg-destructive/10 rounded" title="Delete">
                    <Trash2 className="h-4 w-4 text-destructive" />
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}