import { useQuery } from '@tanstack/react-query'
import { FileText } from 'lucide-react'

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
      <div>
        <h1 className="text-3xl font-bold">Reports</h1>
        <p className="text-muted-foreground mt-1">
          Generated engagement reports
        </p>
      </div>

      {isLoading ? (
        <div className="text-center py-12 text-muted-foreground">Loading...</div>
      ) : reports.length === 0 ? (
        <div className="rounded-lg border border-border bg-card p-12">
          <div className="text-center">
            <FileText className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
            <h2 className="text-xl font-semibold mb-2">No Reports Generated</h2>
            <p className="text-muted-foreground max-w-md mx-auto">
              Generate reports from your session data using the CLI: <code className="text-primary">pentlog export</code>
            </p>
          </div>
        </div>
      ) : (
        <div className="rounded-lg border border-border overflow-hidden">
          <table className="w-full">
            <thead className="bg-muted/50">
              <tr>
                <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">Name</th>
                <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">Client</th>
                <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">Type</th>
                <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">Size</th>
                <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">Date</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {reports.map((report, idx) => (
                <tr key={idx} className="hover:bg-muted/30 transition-colors">
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2">
                      <FileText className="h-4 w-4 text-muted-foreground" />
                      <span className="font-medium">{report.name}</span>
                    </div>
                  </td>
                  <td className="px-4 py-3 text-sm">{report.client}</td>
                  <td className="px-4 py-3">
                    <span className={`px-2 py-0.5 text-xs rounded ${
                      report.type === 'html' ? 'bg-blue-500/10 text-blue-500' : 'bg-gray-500/10 text-gray-500'
                    }`}>
                      {report.type.toUpperCase()}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-sm text-muted-foreground">{report.size_human}</td>
                  <td className="px-4 py-3 text-sm text-muted-foreground">{report.mod_time}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}