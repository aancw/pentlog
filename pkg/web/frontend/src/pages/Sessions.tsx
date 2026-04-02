import { useState } from 'react'
import { Link } from 'react-router-dom'
import { Eye, Search, Trash2 } from 'lucide-react'
import { formatDate, formatListLabel } from '../lib/api'
import { useSessionTags, useSessions } from '../hooks/useApi'
import { api } from '../hooks/useApi'

const PAGE_SIZE = 20
const phaseOptions = ['recon', 'enumeration', 'initial-access', 'exploitation', 'post-exploitation', 'reporting']
const stateOptions = ['completed', 'active', 'crashed', 'paused']

export default function Sessions() {
  const [page, setPage] = useState(1)
  const [query, setQuery] = useState('')
  const [client, setClient] = useState('')
  const [phase, setPhase] = useState('')
  const [state, setState] = useState('')
  const [tag, setTag] = useState('')

  const { data, isLoading, refetch } = useSessions({
    limit: PAGE_SIZE,
    offset: (page - 1) * PAGE_SIZE,
    q: query || undefined,
    client: client || undefined,
    phase: phase || undefined,
    state: state || undefined,
    tag: tag || undefined,
  })
  const { data: tags } = useSessionTags()

  const total = data?.total ?? 0
  const sessions = data?.sessions ?? []
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))

  async function handleDelete(id: number) {
    if (!window.confirm(`Delete session #${id}?`)) {
      return
    }
    await api.sessions.delete(id)
    await refetch()
  }

  function resetPage() {
    setPage(1)
  }

  if (isLoading) {
    return <div className="loading-state">Loading sessions…</div>
  }

  return (
    <div className="page-stack">
      <section className="page-heading">
        <div>
          <div className="eyebrow">Session Browser</div>
          <h2>Review captured shell recordings.</h2>
          <p>Filter by client, phase, state, or tags to narrow the evidence trail.</p>
        </div>
        <div className="page-heading-meta">
          <span className="pill">{total} matches</span>
        </div>
      </section>

      <section className="panel-card filter-card">
        <div className="filter-grid">
          <label className="field field-search">
            <span>Search</span>
            <div className="search-input-shell">
              <Search size={16} />
              <input
                value={query}
                onChange={(event) => {
                  setQuery(event.target.value)
                  resetPage()
                }}
                placeholder="Filename, client, engagement…"
              />
            </div>
          </label>

          <label className="field">
            <span>Client</span>
            <input
              value={client}
              onChange={(event) => {
                setClient(event.target.value)
                resetPage()
              }}
              placeholder="Contains client name"
            />
          </label>

          <label className="field">
            <span>Phase</span>
            <select
              value={phase}
              onChange={(event) => {
                setPhase(event.target.value)
                resetPage()
              }}
            >
              <option value="">All phases</option>
              {phaseOptions.map((option) => (
                <option key={option} value={option}>
                  {formatListLabel(option)}
                </option>
              ))}
            </select>
          </label>

          <label className="field">
            <span>State</span>
            <select
              value={state}
              onChange={(event) => {
                setState(event.target.value)
                resetPage()
              }}
            >
              <option value="">All states</option>
              {stateOptions.map((option) => (
                <option key={option} value={option}>
                  {formatListLabel(option)}
                </option>
              ))}
            </select>
          </label>

          <label className="field">
            <span>Tag</span>
            <select
              value={tag}
              onChange={(event) => {
                setTag(event.target.value)
                resetPage()
              }}
            >
              <option value="">All tags</option>
              {(tags?.tags ?? []).map((item) => (
                <option key={item.name} value={item.name}>
                  {item.name} ({item.count})
                </option>
              ))}
            </select>
          </label>
        </div>
      </section>

      <section className="panel-card">
        <div className="table-shell">
          <table>
            <thead>
              <tr>
                <th>ID</th>
                <th>Client</th>
                <th>Engagement</th>
                <th>Phase</th>
                <th>State</th>
                <th>Notes</th>
                <th>Size</th>
                <th>Recorded</th>
                <th aria-label="Actions" />
              </tr>
            </thead>
            <tbody>
              {sessions.map((session) => (
                <tr key={session.id}>
                  <td className="mono-text">#{session.id}</td>
                  <td>{session.metadata.client || '-'}</td>
                  <td>{session.metadata.engagement || '-'}</td>
                  <td>
                    <span className="pill">{session.metadata.phase || '-'}</span>
                  </td>
                  <td>
                    <span className={`badge badge-${session.state.toLowerCase()}`}>{formatListLabel(session.state)}</span>
                  </td>
                  <td>{session.notes_count}</td>
                  <td>{session.size_human}</td>
                  <td>{formatDate(session.mod_time)}</td>
                  <td>
                    <div className="row-actions">
                      <Link className="icon-button" to={`/sessions/${session.id}`} aria-label={`Open session ${session.id}`}>
                        <Eye size={16} />
                      </Link>
                      <button className="icon-button danger" onClick={() => handleDelete(session.id)} aria-label={`Delete session ${session.id}`}>
                        <Trash2 size={16} />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {sessions.length === 0 && <div className="empty-state compact">No sessions match the current filters.</div>}
        </div>

        <div className="pagination-bar">
          <div className="subdued-text">
            Page {page} of {totalPages}
          </div>
          <div className="row-actions">
            <button className="secondary-button" onClick={() => setPage((current) => Math.max(1, current - 1))} disabled={page === 1}>
              Previous
            </button>
            <button className="secondary-button" onClick={() => setPage((current) => Math.min(totalPages, current + 1))} disabled={page >= totalPages}>
              Next
            </button>
          </div>
        </div>
      </section>
    </div>
  )
}
