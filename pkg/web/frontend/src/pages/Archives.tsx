import { useQuery } from '@tanstack/react-query'
import { Archive, Lock, FileArchive, Folder } from 'lucide-react'

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
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold text-[hsl(var(--foreground))]">Archives</h1>
        <p className="text-[hsl(var(--muted-foreground))] mt-1">
          Session archives and backups
        </p>
      </div>

      {isLoading ? (
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-[hsl(var(--primary))]"></div>
        </div>
      ) : archives.length === 0 ? (
        <div className="rounded-xl border border-[hsl(var(--border))] bg-[hsl(var(--card))] p-12">
          <div className="text-center">
            <Archive className="h-12 w-12 mx-auto text-[hsl(var(--muted-foreground))] mb-4" />
            <h2 className="text-xl font-semibold text-[hsl(var(--foreground))] mb-2">No Archives</h2>
            <p className="text-[hsl(var(--muted-foreground))] max-w-md mx-auto mb-6">
              Create archives from your session data using the CLI:
            </p>
            <code className="inline-block px-4 py-2 bg-[hsl(var(--secondary))] rounded-lg text-sm text-[hsl(var(--primary))]">
              pentlog archive
            </code>
          </div>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {archives.map((archive, idx) => (
            <div key={idx} className="rounded-xl border border-[hsl(var(--border))] bg-[hsl(var(--card))] p-5 hover:border-[hsl(var(--primary))]/50 transition-colors">
              <div className="flex items-start gap-4">
                <div className="p-3 rounded-lg bg-[hsl(var(--secondary))]">
                  <FileArchive className="h-6 w-6 text-[hsl(var(--primary))]" />
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-1">
                    {archive.encrypted && (
                      <span className="inline-flex items-center gap-1 px-2 py-0.5 text-xs bg-orange-500/10 text-orange-500 rounded">
                        <Lock className="h-3 w-3" />
                        Encrypted
                      </span>
                    )}
                  </div>
                  <h3 className="font-medium text-[hsl(var(--foreground))] truncate">{archive.name}</h3>
                  <div className="flex items-center gap-2 mt-1 text-xs text-[hsl(var(--muted-foreground))]">
                    <Folder className="h-3 w-3" />
                    <span>{archive.client}</span>
                    <span>•</span>
                    <span>{archive.size_human}</span>
                  </div>
                  <div className="text-xs text-[hsl(var(--muted-foreground))] mt-1">{archive.mod_time}</div>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}