import { useQuery } from '@tanstack/react-query'
import { Archive, FileArchive, Folder, Lock } from 'lucide-react'

export default function Archives() {
  const { data, isLoading } = useQuery({
    queryKey: ['archives'],
    queryFn: async () => {
      const res = await fetch('/api/archives')
      return res.json()
    },
  })

  const archives = data?.archives ?? []

  if (isLoading) return <div className="text-center p-8">Loading...</div>

  return (
    <div className="space-y-6">
      <div className="page-header">
        <h1>Archives</h1>
        <p className="text-muted">Session archives and backups</p>
      </div>

      {archives.length === 0 ? (
        <div className="card text-center p-12">
          <Archive className="h-12 w-12 mx-auto text-muted mb-4" />
          <h2 className="text-xl font-semibold mb-2">No Archives</h2>
          <p className="text-muted mb-4">Create archives using CLI:</p>
          <code className="text-sm text-primary bg-secondary px-4 py-2 rounded">pentlog archive</code>
        </div>
      ) : (
        <div className="grid gap-4" style={{ gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))' }}>
          {archives.map((archive: any, idx: number) => (
            <div key={idx} className="card">
              <div className="flex items-start gap-4">
                <div className="p-3 rounded bg-secondary">
                  <FileArchive className="h-6 w-6 text-primary" />
                </div>
                <div className="flex-1 min-w-0">
                  {archive.encrypted && (
                    <span className="badge badge-orange mb-1">
                      <Lock className="h-3 w-3 mr-1" /> Encrypted
                    </span>
                  )}
                  <div className="font-medium truncate">{archive.name}</div>
                  <div className="text-xs text-muted mt-1">
                    <Folder className="inline h-3 w-3 mr-1" />
                    {archive.client} · {archive.size_human}
                  </div>
                  <div className="text-xs text-muted">{archive.mod_time}</div>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}