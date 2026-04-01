import { useParams, Link } from 'react-router-dom'
import { useSession } from '../hooks/useApi'
import { ArrowLeft, Tag, FileText, Clock, HardDrive } from 'lucide-react'

export default function SessionDetail() {
  const { id } = useParams<{ id: string }>()
  const sessionId = parseInt(id || '0', 10)
  const { data: session, isLoading } = useSession(sessionId)

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-muted-foreground">Loading session...</div>
      </div>
    )
  }

  if (!session) {
    return (
      <div className="space-y-6">
        <Link to="/sessions" className="flex items-center gap-2 text-muted-foreground hover:text-foreground">
          <ArrowLeft className="h-4 w-4" />
          Back to sessions
        </Link>
        <div className="text-center py-12">
          <p className="text-muted-foreground">Session not found</p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <Link to="/sessions" className="flex items-center gap-2 text-muted-foreground hover:text-foreground">
        <ArrowLeft className="h-4 w-4" />
        Back to sessions
      </Link>

      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-3xl font-bold">Session #{session.id}</h1>
          <p className="text-muted-foreground mt-1 font-mono text-sm">{session.filename}</p>
        </div>
        <div className="flex items-center gap-2">
          {session.state && (
            <span className={`px-2 py-1 text-xs rounded-full ${
              session.state === 'active' ? 'bg-green-500/10 text-green-500' :
              session.state === 'crashed' ? 'bg-red-500/10 text-red-500' :
              'bg-secondary text-secondary-foreground'
            }`}>
              {session.state}
            </span>
          )}
        </div>
      </div>

      <div className="grid gap-6 md:grid-cols-3">
        <div className="rounded-lg border border-border bg-card p-6">
          <div className="flex items-center gap-2 text-muted-foreground mb-2">
            <HardDrive className="h-4 w-4" />
            <span className="text-sm font-medium">Size</span>
          </div>
          <p className="text-2xl font-bold">{session.size_human}</p>
          <p className="text-xs text-muted-foreground mt-1">{session.size.toLocaleString()} bytes</p>
        </div>

        <div className="rounded-lg border border-border bg-card p-6">
          <div className="flex items-center gap-2 text-muted-foreground mb-2">
            <Clock className="h-4 w-4" />
            <span className="text-sm font-medium">Modified</span>
          </div>
          <p className="text-lg font-semibold">{session.mod_time}</p>
        </div>

        <div className="rounded-lg border border-border bg-card p-6">
          <div className="flex items-center gap-2 text-muted-foreground mb-2">
            <FileText className="h-4 w-4" />
            <span className="text-sm font-medium">Notes</span>
          </div>
          <p className="text-2xl font-bold">{session.notes_count}</p>
        </div>
      </div>

      <div className="rounded-lg border border-border bg-card p-6">
        <h2 className="text-lg font-semibold mb-4">Metadata</h2>
        <div className="grid gap-4 md:grid-cols-2">
          <MetadataItem label="Client" value={session.metadata.client} />
          <MetadataItem label="Engagement" value={session.metadata.engagement} />
          <MetadataItem label="Phase" value={session.metadata.phase} />
          <MetadataItem label="Operator" value={session.metadata.operator} />
          <MetadataItem label="Scope" value={session.metadata.scope} />
          <MetadataItem label="Target" value={session.metadata.target || '-'} />
          <MetadataItem label="Target IP" value={session.metadata.target_ip || '-'} />
        </div>
      </div>

      {session.tags && session.tags.length > 0 && (
        <div className="rounded-lg border border-border bg-card p-6">
          <div className="flex items-center gap-2 mb-4">
            <Tag className="h-4 w-4 text-muted-foreground" />
            <h2 className="text-lg font-semibold">Tags</h2>
          </div>
          <div className="flex flex-wrap gap-2">
            {session.tags.map((tag) => (
              <span key={tag} className="px-2 py-1 text-sm bg-secondary rounded-md">
                {tag}
              </span>
            ))}
          </div>
        </div>
      )}

      <div className="rounded-lg border border-border bg-card p-6">
        <h2 className="text-lg font-semibold mb-4">File Path</h2>
        <p className="font-mono text-sm text-muted-foreground bg-muted p-3 rounded-md break-all">
          {session.path}
        </p>
      </div>
    </div>
  )
}

function MetadataItem({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <span className="text-sm text-muted-foreground">{label}</span>
      <p className="font-medium capitalize">{value || '-'}</p>
    </div>
  )
}