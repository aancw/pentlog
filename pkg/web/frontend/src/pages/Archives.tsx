import { useQuery } from '@tanstack/react-query'
import { Archive, Download, Lock, FileArchive } from 'lucide-react'

interface ArchiveItem {
  name: string
  client: string
  path: string
  size: number
  size_human: string
  mod_time: string
  encrypted: boolean
}

export default function Archives() {
  const { data: archivesData, isLoading } = useQuery({
    queryKey: ['archives'],
    queryFn: async (): Promise<{ archives: ArchiveItem[] }> => {
      const res = await fetch('/api/archives')
      return res.json()
    },
  })

  const archives = archivesData?.archives ?? []

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Archives</h1>
          <p className="text-muted-foreground mt-1">
            Session archives and backups
          </p>
        </div>
      </div>

      {isLoading ? (
        <div className="text-center py-12 text-muted-foreground">Loading...</div>
      ) : archives.length === 0 ? (
        <div className="rounded-lg border border-border bg-card p-12">
          <div className="text-center">
            <Archive className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
            <h2 className="text-xl font-semibold mb-2">No Archives</h2>
            <p className="text-muted-foreground max-w-md mx-auto">
              Create archives from your session data using the CLI: <code className="text-primary">pentlog archive</code>
            </p>
            <div className="flex items-center justify-center gap-4 mt-6">
              <div className="flex items-center gap-2 px-4 py-2 border border-border rounded-lg">
                <Download className="h-4 w-4" />
                <span className="text-sm">Export</span>
              </div>
              <div className="flex items-center gap-2 px-4 py-2 border border-border rounded-lg">
                <Lock className="h-4 w-4" />
                <span className="text-sm">Encrypt</span>
              </div>
            </div>
          </div>
        </div>
      ) : (
        <div className="rounded-lg border border-border overflow-hidden">
          <table className="w-full">
            <thead className="bg-muted/50">
              <tr>
                <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">Name</th>
                <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">Client</th>
                <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">Size</th>
                <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">Date</th>
                <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">Encrypted</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {archives.map((archive, idx) => (
                <tr key={idx} className="hover:bg-muted/30 transition-colors">
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2">
                      <FileArchive className="h-4 w-4 text-muted-foreground" />
                      <span className="font-medium">{archive.name}</span>
                    </div>
                  </td>
                  <td className="px-4 py-3 text-sm">{archive.client}</td>
                  <td className="px-4 py-3 text-sm text-muted-foreground">{archive.size_human}</td>
                  <td className="px-4 py-3 text-sm text-muted-foreground">{archive.mod_time}</td>
                  <td className="px-4 py-3">
                    {archive.encrypted ? (
                      <span className="flex items-center gap-1 text-xs text-orange-500">
                        <Lock className="h-3 w-3" />
                        Yes
                      </span>
                    ) : (
                      <span className="text-xs text-muted-foreground">No</span>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}