import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useSessions } from '../hooks/useApi'
import { ChevronLeft, ChevronRight, Trash2, Eye, Search } from 'lucide-react'

export default function Sessions() {
  const [page, setPage] = useState(1)
  const [clientFilter, setClientFilter] = useState('')
  const limit = 15
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
    if (!confirm('Delete this session?')) return
    await fetch(`/api/sessions/${id}`, { method: 'DELETE' })
    refetch()
  }

  if (isLoading) {
    return <div className="text-center p-8">Loading...</div>
  }

  return (
    <div className="space-y-6">
      <div className="page-header">
        <h1>Sessions</h1>
        <p className="text-muted">{total} recorded sessions</p>
      </div>

      <div className="flex gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted" />
          <input
            type="text"
            placeholder="Filter by client..."
            value={clientFilter}
            onChange={(e) => { setClientFilter(e.target.value); setPage(1) }}
            className="input"
            style={{ paddingLeft: '2.5rem' }}
          />
        </div>
      </div>

      <div className="table-container">
        <table>
          <thead>
            <tr>
              <th>ID</th>
              <th>Client</th>
              <th>Engagement</th>
              <th>Phase</th>
              <th>Size</th>
              <th>Date</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {sessions.map((session) => (
              <tr key={session.id}>
                <td className="font-mono text-primary">#{session.id}</td>
                <td className="font-medium">{session.metadata.client || '-'}</td>
                <td>{session.metadata.engagement || '-'}</td>
                <td>
                  <span className="badge badge-purple">{session.metadata.phase || '-'}</span>
                </td>
                <td>{session.size_human}</td>
                <td>{session.mod_time}</td>
                <td>
                  <div className="flex gap-2">
                    <Link to={`/sessions/${session.id}`} className="p-2 hover-bg rounded">
                      <Eye className="h-4 w-4 text-muted" />
                    </Link>
                    <button onClick={() => handleDelete(session.id)} className="p-2 hover-bg rounded">
                      <Trash2 className="h-4 w-4 text-red" />
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {totalPages > 1 && (
        <div className="flex justify-between items-center">
          <span className="text-sm text-muted">Page {page} of {totalPages}</span>
          <div className="flex gap-2">
            <button 
              onClick={() => setPage(p => Math.max(1, p - 1))} 
              disabled={page === 1}
              className="p-2 border border-border rounded hover-bg disabled:opacity-50"
            >
              <ChevronLeft className="h-4 w-4" />
            </button>
            <button 
              onClick={() => setPage(p => Math.min(totalPages, p + 1))}
              disabled={page === totalPages}
              className="p-2 border border-border rounded hover-bg disabled:opacity-50"
            >
              <ChevronRight className="h-4 w-4" />
            </button>
          </div>
        </div>
      )}
    </div>
  )
}