import { useQuery } from '@tanstack/react-query'
import { useParams, Link } from 'react-router-dom'
import { ArrowLeft, FileText, HardDrive, Calendar } from 'lucide-react'

export default function SessionDetail() {
  const { id } = useParams()
  const { data: session, isLoading } = useQuery({
    queryKey: ['session', id],
    queryFn: async () => {
      const res = await fetch(`/api/sessions/${id}`)
      return res.json()
    },
    enabled: !!id,
  })

  if (isLoading) return <div className="text-center p-8">Loading...</div>
  if (!session) return <div className="text-center p-8">Session not found</div>

  return (
    <div className="space-y-6">
      <Link to="/sessions" className="inline-flex items-center gap-2 text-muted hover:text-white">
        <ArrowLeft className="h-4 w-4" /> Back to sessions
      </Link>

      <div className="page-header">
        <h1>Session #{session.id}</h1>
        <p className="font-mono text-sm text-muted">{session.filename}</p>
      </div>

      <div className="grid gap-4" style={{ gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))' }}>
        <div className="stat-card">
          <HardDrive className="icon h-5 w-5" />
          <div className="label">Size</div>
          <div className="value">{session.size_human}</div>
        </div>
        <div className="stat-card">
          <Calendar className="icon h-5 w-5" />
          <div className="label">Modified</div>
          <div className="value text-lg">{session.mod_time}</div>
        </div>
        <div className="stat-card">
          <FileText className="icon h-5 w-5" />
          <div className="label">Notes</div>
          <div className="value">{session.notes_count}</div>
        </div>
      </div>

      <div className="card">
        <h2 className="text-lg font-semibold mb-4">Metadata</h2>
        <div className="grid gap-3" style={{ gridTemplateColumns: 'repeat(auto-fit, minmax(150px, 1fr))' }}>
          <div>
            <div className="text-xs text-muted">Client</div>
            <div className="font-medium">{session.metadata?.client || '-'}</div>
          </div>
          <div>
            <div className="text-xs text-muted">Engagement</div>
            <div className="font-medium">{session.metadata?.engagement || '-'}</div>
          </div>
          <div>
            <div className="text-xs text-muted">Phase</div>
            <div className="font-medium capitalize">{session.metadata?.phase || '-'}</div>
          </div>
          <div>
            <div className="text-xs text-muted">Operator</div>
            <div className="font-medium">{session.metadata?.operator || '-'}</div>
          </div>
        </div>
      </div>

      <div className="card">
        <h2 className="text-lg font-semibold mb-2">File Path</h2>
        <div className="font-mono text-xs bg-secondary p-3 rounded break-all">
          {session.path}
        </div>
      </div>
    </div>
  )
}