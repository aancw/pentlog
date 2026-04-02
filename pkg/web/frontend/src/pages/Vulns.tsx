import { useQuery } from '@tanstack/react-query'
import { Shield, Plus } from 'lucide-react'

export default function Vulns() {
  const { data, isLoading } = useQuery({
    queryKey: ['vulns'],
    queryFn: async () => {
      const res = await fetch('/api/vulns')
      return res.json()
    },
  })

  const vulns = data?.vulns ?? []

  if (isLoading) return <div className="text-center p-8">Loading...</div>

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div className="page-header mb-0">
          <h1>Vulnerabilities</h1>
          <p className="text-muted">{data?.total ?? 0} findings tracked</p>
        </div>
        <button className="btn">
          <Plus className="h-4 w-4" />
          Add Finding
        </button>
      </div>

      {vulns.length === 0 ? (
        <div className="card text-center p-12">
          <Shield className="h-12 w-12 mx-auto text-muted mb-4" />
          <h2 className="text-xl font-semibold mb-2">No Vulnerabilities</h2>
          <p className="text-muted">Start tracking security findings</p>
        </div>
      ) : (
        <div className="space-y-4">
          {vulns.map((vuln: any) => {
            const severityColors: Record<string, string> = {
              Critical: 'badge-red',
              High: 'badge-orange',
              Medium: 'badge-yellow',
              Low: 'badge-green',
              Info: 'badge-blue',
            }
            return (
              <div key={vuln.id} className="card">
                <div className="flex items-start gap-4">
                  <div className="p-2 rounded bg-secondary">
                    <Shield className="h-5 w-5 text-primary" />
                  </div>
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-2">
                      <span className={`badge ${severityColors[vuln.severity] || 'badge-purple'}`}>
                        {vuln.severity}
                      </span>
                      <span className={`badge ${vuln.status === 'Open' ? 'badge-yellow' : vuln.status === 'Verified' ? 'badge-green' : 'badge-purple'}`}>
                        {vuln.status}
                      </span>
                      <span className="text-xs text-muted">{vuln.id}</span>
                    </div>
                    <h3 className="font-semibold text-lg">{vuln.title}</h3>
                    {vuln.description && (
                      <p className="text-sm text-muted mt-1">{vuln.description}</p>
                    )}
                    <div className="text-xs text-muted mt-2">
                      {vuln.phase && <span>Phase: {vuln.phase} · </span>}
                      {new Date(vuln.created_at).toLocaleDateString()}
                    </div>
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