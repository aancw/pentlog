import { useQuery } from '@tanstack/react-query'
import { FileText, FileCode } from 'lucide-react'

interface Report {
  name: string
  client: string
  path: string
  size: number
  size_human: string
  mod_time: string
  type: string
}

export default function Reports() {
  const { data: reportsData, isLoading } = useQuery({
    queryKey: ['reports'],
    queryFn: async (): Promise<{ reports: Report[] }> => {
      const res = await fetch('/api/reports')
      return res.json()
    },
  })

  const reports = reportsData?.reports ?? []

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold text-[hsl(var(--foreground))]">Reports</h1>
        <p className="text-[hsl(var(--muted-foreground))] mt-1">
          Generated engagement reports
        </p>
      </div>

      {isLoading ? (
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-[hsl(var(--primary))]"></div>
        </div>
      ) : reports.length === 0 ? (
        <div className="rounded-xl border border-[hsl(var(--border))] bg-[hsl(var(--card))] p-12">
          <div className="text-center">
            <FileText className="h-12 w-12 mx-auto text-[hsl(var(--muted-foreground))] mb-4" />
            <h2 className="text-xl font-semibold text-[hsl(var(--foreground))] mb-2">No Reports Generated</h2>
            <p className="text-[hsl(var(--muted-foreground))] max-w-md mx-auto mb-6">
              Generate reports from your session data using the CLI:
            </p>
            <code className="inline-block px-4 py-2 bg-[hsl(var(--secondary))] rounded-lg text-sm text-[hsl(var(--primary))]">
              pentlog export
            </code>
          </div>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {reports.map((report, idx) => (
            <div key={idx} className="rounded-xl border border-[hsl(var(--border))] bg-[hsl(var(--card))] p-5 hover:border-[hsl(var(--primary))]/50 transition-colors group">
              <div className="flex items-start gap-4">
                <div className={`p-3 rounded-lg ${report.type === 'html' ? 'bg-blue-500/10' : 'bg-[hsl(var(--secondary))]'}`}>
                  {report.type === 'html' ? (
                    <FileCode className="h-6 w-6 text-blue-500" />
                  ) : (
                    <FileText className="h-6 w-6 text-[hsl(var(--muted-foreground))]" />
                  )}
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-1">
                    <span className={`px-2 py-0.5 text-xs font-medium rounded ${
                      report.type === 'html' ? 'bg-blue-500/10 text-blue-500' : 'bg-[hsl(var(--secondary))] text-[hsl(var(--muted-foreground))]'
                    }`}>
                      {report.type.toUpperCase()}
                    </span>
                  </div>
                  <h3 className="font-medium text-[hsl(var(--foreground))] truncate">{report.name}</h3>
                  <div className="flex items-center gap-3 mt-2 text-xs text-[hsl(var(--muted-foreground))]">
                    <span>{report.client}</span>
                    <span>•</span>
                    <span>{report.size_human}</span>
                  </div>
                  <div className="text-xs text-[hsl(var(--muted-foreground))] mt-1">{report.mod_time}</div>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}