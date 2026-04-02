import { useParams, Link } from 'react-router-dom'
import { useSession } from '../hooks/useApi'
import { ArrowLeft, Tag, FileText, HardDrive, Calendar, User, Target, Globe } from 'lucide-react'

export default function SessionDetail() {
  const { id } = useParams<{ id: string }>()
  const sessionId = parseInt(id || '0', 10)
  const { data: session, isLoading } = useSession(sessionId)

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-[hsl(var(--primary))]"></div>
      </div>
    )
  }

  if (!session) {
    return (
      <div className="space-y-6">
        <Link to="/sessions" className="inline-flex items-center gap-2 text-[hsl(var(--muted-foreground))] hover:text-[hsl(var(--foreground))] transition-colors">
          <ArrowLeft className="h-4 w-4" />
          Back to sessions
        </Link>
        <div className="text-center py-12">
          <p className="text-[hsl(var(--muted-foreground))]">Session not found</p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Back Link */}
      <Link to="/sessions" className="inline-flex items-center gap-2 text-[hsl(var(--muted-foreground))] hover:text-[hsl(var(--foreground))] transition-colors">
        <ArrowLeft className="h-4 w-4" />
        Back to sessions
      </Link>

      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-3xl font-bold text-[hsl(var(--foreground))]">Session #{session.id}</h1>
          <p className="text-[hsl(var(--muted-foreground))] mt-1 font-mono text-sm">{session.filename}</p>
        </div>
        {session.state && (
          <span className={`px-3 py-1 text-sm rounded-full font-medium ${
            session.state === 'active' ? 'bg-green-500/10 text-green-500' :
            session.state === 'crashed' ? 'bg-red-500/10 text-red-500' :
            'bg-[hsl(var(--secondary))] text-[hsl(var(--muted-foreground))]'
          }`}>
            {session.state}
          </span>
        )}
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div className="rounded-xl border border-[hsl(var(--border))] bg-[hsl(var(--card))] p-6">
          <div className="flex items-center gap-3">
            <HardDrive className="h-5 w-5 text-[hsl(var(--primary))]" />
            <span className="text-sm text-[hsl(var(--muted-foreground))]">Size</span>
          </div>
          <div className="mt-2 text-2xl font-bold text-[hsl(var(--foreground))]">{session.size_human}</div>
          <div className="text-xs text-[hsl(var(--muted-foreground))]">{session.size.toLocaleString()} bytes</div>
        </div>

        <div className="rounded-xl border border-[hsl(var(--border))] bg-[hsl(var(--card))] p-6">
          <div className="flex items-center gap-3">
            <Calendar className="h-5 w-5 text-[hsl(var(--primary))]" />
            <span className="text-sm text-[hsl(var(--muted-foreground))]">Modified</span>
          </div>
          <div className="mt-2 text-lg font-semibold text-[hsl(var(--foreground))]">{session.mod_time}</div>
        </div>

        <div className="rounded-xl border border-[hsl(var(--border))] bg-[hsl(var(--card))] p-6">
          <div className="flex items-center gap-3">
            <FileText className="h-5 w-5 text-[hsl(var(--primary))]" />
            <span className="text-sm text-[hsl(var(--muted-foreground))]">Notes</span>
          </div>
          <div className="mt-2 text-2xl font-bold text-[hsl(var(--foreground))]">{session.notes_count}</div>
        </div>
      </div>

      {/* Metadata */}
      <div className="rounded-xl border border-[hsl(var(--border))] bg-[hsl(var(--card))] p-6">
        <h2 className="text-lg font-semibold text-[hsl(var(--foreground))] mb-6">Metadata</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          <MetadataItem icon={User} label="Client" value={session.metadata.client} />
          <MetadataItem icon={FileText} label="Engagement" value={session.metadata.engagement} />
          <MetadataItem icon={Target} label="Phase" value={session.metadata.phase} />
          <MetadataItem icon={User} label="Operator" value={session.metadata.operator} />
          <MetadataItem icon={Globe} label="Scope" value={session.metadata.scope} />
          <MetadataItem icon={Target} label="Target" value={session.metadata.target} />
        </div>
      </div>

      {/* Tags */}
      {session.tags && session.tags.length > 0 && (
        <div className="rounded-xl border border-[hsl(var(--border))] bg-[hsl(var(--card))] p-6">
          <div className="flex items-center gap-2 mb-4">
            <Tag className="h-5 w-5 text-[hsl(var(--primary))]" />
            <h2 className="text-lg font-semibold text-[hsl(var(--foreground))]">Tags</h2>
          </div>
          <div className="flex flex-wrap gap-2">
            {session.tags.map((tag) => (
              <span key={tag} className="px-3 py-1 text-sm bg-[hsl(var(--primary))]/10 text-[hsl(var(--primary))] rounded-full">
                {tag}
              </span>
            ))}
          </div>
        </div>
      )}

      {/* File Path */}
      <div className="rounded-xl border border-[hsl(var(--border))] bg-[hsl(var(--card))] p-6">
        <h2 className="text-lg font-semibold text-[hsl(var(--foreground))] mb-4">File Path</h2>
        <div className="font-mono text-sm text-[hsl(var(--muted-foreground))] bg-[hsl(var(--secondary))] p-4 rounded-lg break-all">
          {session.path}
        </div>
      </div>
    </div>
  )
}

function MetadataItem({ icon: Icon, label, value }: { icon: React.ComponentType<{ className?: string }>, label: string, value: string }) {
  return (
    <div className="flex items-start gap-3">
      <Icon className="h-4 w-4 text-[hsl(var(--muted-foreground))] mt-0.5" />
      <div>
        <div className="text-xs text-[hsl(var(--muted-foreground))]">{label}</div>
        <div className="font-medium text-[hsl(var(--foreground))] capitalize">{value || '-'}</div>
      </div>
    </div>
  )
}