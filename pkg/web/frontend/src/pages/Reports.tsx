import { useQuery } from '@tanstack/react-query'
import { FileText, FileCode } from 'lucide-react'

export default function Reports() {
  const { data, isLoading } = useQuery({
    queryKey: ['reports'],
    queryFn: async () => {
      const res = await fetch('/api/reports')
      return res.json()
    },
  })

  const reports = data?.reports ?? []

  if (isLoading) return <div className="text-center p-8">Loading...</div>

  return (
    <div className="space-y-6">
      <div className="page-header">
        <h1>Reports</h1>
        <p className="text-muted">Generated engagement reports</p>
      </div>

      {reports.length === 0 ? (
        <div className="card text-center p-12">
          <FileText className="h-12 w-12 mx-auto text-muted mb-4" />
          <h2 className="text-xl font-semibold mb-2">No Reports</h2>
          <p className="text-muted mb-4">Generate reports using CLI:</p>
          <code className="text-sm text-primary bg-secondary px-4 py-2 rounded">pentlog export</code>
        </div>
      ) : (
        <div className="grid gap-4" style={{ gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))' }}>
          {reports.map((report: any, idx: number) => (
            <div key={idx} className="card">
              <div className="flex items-start gap-4">
                <div className={`p-3 rounded ${report.type === 'html' ? 'bg-blue/10' : 'bg-secondary'}`}>
                  {report.type === 'html' 
                    ? <FileCode className="h-6 w-6 text-blue" />
                    : <FileText className="h-6 w-6 text-muted" />
                  }
                </div>
                <div className="flex-1 min-w-0">
                  <span className={`badge ${report.type === 'html' ? 'badge-blue' : 'badge-purple'} mb-1`}>
                    {report.type.toUpperCase()}
                  </span>
                  <div className="font-medium truncate">{report.name}</div>
                  <div className="text-xs text-muted mt-1">
                    {report.client} · {report.size_human}
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}