import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useSessions } from '../hooks/useApi'
import { ChevronLeft, ChevronRight, Trash2, Eye, Search, Filter } from 'lucide-react'

export default function Sessions() {
  const [page, setPage] = useState(1)
  const [clientFilter, setClientFilter] = useState('')
  const [searchQuery, setSearchQuery] = useState('')
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

  const filteredSessions = searchQuery 
    ? sessions.filter(s => 
        s.filename.toLowerCase().includes(searchQuery.toLowerCase()) ||
        s.metadata.client.toLowerCase().includes(searchQuery.toLowerCase()) ||
        s.metadata.engagement.toLowerCase().includes(searchQuery.toLowerCase())
      )
    : sessions

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-[hsl(var(--foreground))]">Sessions</h1>
          <p className="text-[hsl(var(--muted-foreground))] mt-1">
            {total} recorded session{total !== 1 ? 's' : ''}
          </p>
        </div>
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-[hsl(var(--muted-foreground))]" />
          <input
            type="text"
            placeholder="Search sessions..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pl-10 pr-4 py-2.5 bg-[hsl(var(--card))] border border-[hsl(var(--border))] rounded-lg text-sm text-[hsl(var(--foreground))] placeholder:text-[hsl(var(--muted-foreground))] focus:outline-none focus:ring-2 focus:ring-[hsl(var(--primary))]"
          />
        </div>
        <div className="relative">
          <Filter className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-[hsl(var(--muted-foreground))]" />
          <input
            type="text"
            placeholder="Filter by client..."
            value={clientFilter}
            onChange={(e) => {
              setClientFilter(e.target.value)
              setPage(1)
            }}
            className="w-full sm:w-48 pl-10 pr-4 py-2.5 bg-[hsl(var(--card))] border border-[hsl(var(--border))] rounded-lg text-sm text-[hsl(var(--foreground))] placeholder:text-[hsl(var(--muted-foreground))] focus:outline-none focus:ring-2 focus:ring-[hsl(var(--primary))]"
          />
        </div>
      </div>

      {/* Table */}
      {isLoading ? (
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-[hsl(var(--primary))]"></div>
        </div>
      ) : (
        <div className="rounded-xl border border-[hsl(var(--border))] bg-[hsl(var(--card))] overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-[hsl(var(--border))] bg-[hsl(var(--secondary))]">
                  <th className="px-4 py-3 text-left text-xs font-semibold text-[hsl(var(--muted-foreground))] uppercase tracking-wider">ID</th>
                  <th className="px-4 py-3 text-left text-xs font-semibold text-[hsl(var(--muted-foreground))] uppercase tracking-wider">Client</th>
                  <th className="px-4 py-3 text-left text-xs font-semibold text-[hsl(var(--muted-foreground))] uppercase tracking-wider">Engagement</th>
                  <th className="px-4 py-3 text-left text-xs font-semibold text-[hsl(var(--muted-foreground))] uppercase tracking-wider">Phase</th>
                  <th className="px-4 py-3 text-left text-xs font-semibold text-[hsl(var(--muted-foreground))] uppercase tracking-wider">Size</th>
                  <th className="px-4 py-3 text-left text-xs font-semibold text-[hsl(var(--muted-foreground))] uppercase tracking-wider">Date</th>
                  <th className="px-4 py-3 text-right text-xs font-semibold text-[hsl(var(--muted-foreground))] uppercase tracking-wider">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[hsl(var(--border))]">
                {filteredSessions.map((session) => (
                  <tr key={session.id} className="hover:bg-[hsl(var(--secondary))] transition-colors">
                    <td className="px-4 py-4">
                      <span className="font-mono text-sm text-[hsl(var(--primary))]">#{session.id}</span>
                    </td>
                    <td className="px-4 py-4">
                      <span className="font-medium text-[hsl(var(--foreground))]">{session.metadata.client || '-'}</span>
                    </td>
                    <td className="px-4 py-4 text-sm text-[hsl(var(--muted-foreground))]">{session.metadata.engagement || '-'}</td>
                    <td className="px-4 py-4">
                      <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-[hsl(var(--primary))]/10 text-[hsl(var(--primary))] capitalize">
                        {session.metadata.phase || '-'}
                      </span>
                    </td>
                    <td className="px-4 py-4 text-sm text-[hsl(var(--muted-foreground))]">{session.size_human}</td>
                    <td className="px-4 py-4 text-sm text-[hsl(var(--muted-foreground))]">{session.mod_time}</td>
                    <td className="px-4 py-4">
                      <div className="flex items-center justify-end gap-2">
                        <Link
                          to={`/sessions/${session.id}`}
                          className="p-2 hover:bg-[hsl(var(--accent))] rounded-lg transition-colors"
                          title="View details"
                        >
                          <Eye className="h-4 w-4 text-[hsl(var(--muted-foreground))]" />
                        </Link>
                        <button
                          onClick={() => handleDelete(session.id)}
                          className="p-2 hover:bg-red-500/10 rounded-lg transition-colors"
                          title="Delete session"
                        >
                          <Trash2 className="h-4 w-4 text-red-500" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
                {filteredSessions.length === 0 && (
                  <tr>
                    <td colSpan={7} className="px-4 py-12 text-center text-[hsl(var(--muted-foreground))]">
                      No sessions found
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between">
          <p className="text-sm text-[hsl(var(--muted-foreground))]">
            Page {page} of {totalPages}
          </p>
          <div className="flex items-center gap-2">
            <button
              onClick={() => setPage(p => Math.max(1, p - 1))}
              disabled={page === 1}
              className="p-2 border border-[hsl(var(--border))] rounded-lg hover:bg-[hsl(var(--accent))] disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              <ChevronLeft className="h-4 w-4" />
            </button>
            <button
              onClick={() => setPage(p => Math.min(totalPages, p + 1))}
              disabled={page === totalPages}
              className="p-2 border border-[hsl(var(--border))] rounded-lg hover:bg-[hsl(var(--accent))] disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              <ChevronRight className="h-4 w-4" />
            </button>
          </div>
        </div>
      )}
    </div>
  )
}