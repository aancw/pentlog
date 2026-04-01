import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useSessions } from '../hooks/useApi'
import { ChevronLeft, ChevronRight, Trash2, Eye } from 'lucide-react'

export default function Sessions() {
  const [page, setPage] = useState(1)
  const [clientFilter, setClientFilter] = useState('')
  const limit = 20
  const offset = (page - 1) * limit

  const { data, isLoading, refetch } = useSessions({ 
    limit, 
    offset,
    client: clientFilter || undefined,
  })

  const sessions = data?.sessions ?? []
  const total = data?.total ?? 0
  const totalPages = Math.ceil(total / limit)

  const handleDelete = async (id: number) => {
    if (!confirm('Are you sure you want to delete this session?')) return
    try {
      const response = await fetch(`/api/sessions/${id}`, { method: 'DELETE' })
      if (response.ok) {
        refetch()
      }
    } catch (error) {
      console.error('Failed to delete session:', error)
    }
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-muted-foreground">Loading sessions...</div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Sessions</h1>
          <p className="text-muted-foreground mt-1">
            {total} recorded session{total !== 1 ? 's' : ''}
          </p>
        </div>
        <div className="flex items-center gap-4">
          <input
            type="text"
            placeholder="Filter by client..."
            value={clientFilter}
            onChange={(e) => {
              setClientFilter(e.target.value)
              setPage(1)
            }}
            className="px-3 py-2 border border-border rounded-lg bg-background text-sm"
          />
        </div>
      </div>

      <div className="rounded-lg border border-border overflow-hidden">
        <table className="w-full">
          <thead className="bg-muted/50">
            <tr>
              <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">ID</th>
              <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">Client</th>
              <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">Engagement</th>
              <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">Phase</th>
              <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">Size</th>
              <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">Date</th>
              <th className="px-4 py-3 text-right text-sm font-medium text-muted-foreground">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border">
            {sessions.map((session) => (
              <tr key={session.id} className="hover:bg-muted/30 transition-colors">
                <td className="px-4 py-3 text-sm font-mono">{session.id}</td>
                <td className="px-4 py-3 text-sm">{session.metadata.client || '-'}</td>
                <td className="px-4 py-3 text-sm">{session.metadata.engagement || '-'}</td>
                <td className="px-4 py-3 text-sm capitalize">{session.metadata.phase || '-'}</td>
                <td className="px-4 py-3 text-sm text-muted-foreground">{session.size_human}</td>
                <td className="px-4 py-3 text-sm text-muted-foreground">{session.mod_time}</td>
                <td className="px-4 py-3 text-right">
                  <div className="flex items-center justify-end gap-2">
                    <Link
                      to={`/sessions/${session.id}`}
                      className="p-1 hover:bg-accent rounded"
                      title="View details"
                    >
                      <Eye className="h-4 w-4 text-muted-foreground" />
                    </Link>
                    <button
                      onClick={() => handleDelete(session.id)}
                      className="p-1 hover:bg-destructive/10 rounded"
                      title="Delete session"
                    >
                      <Trash2 className="h-4 w-4 text-destructive" />
                    </button>
                  </div>
                </td>
              </tr>
            ))}
            {sessions.length === 0 && (
              <tr>
                <td colSpan={7} className="px-4 py-8 text-center text-muted-foreground">
                  No sessions found
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      {totalPages > 1 && (
        <div className="flex items-center justify-between">
          <p className="text-sm text-muted-foreground">
            Page {page} of {totalPages}
          </p>
          <div className="flex items-center gap-2">
            <button
              onClick={() => setPage(p => Math.max(1, p - 1))}
              disabled={page === 1}
              className="p-2 border border-border rounded-lg hover:bg-accent disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <ChevronLeft className="h-4 w-4" />
            </button>
            <button
              onClick={() => setPage(p => Math.min(totalPages, p + 1))}
              disabled={page === totalPages}
              className="p-2 border border-border rounded-lg hover:bg-accent disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <ChevronRight className="h-4 w-4" />
            </button>
          </div>
        </div>
      )}
    </div>
  )
}