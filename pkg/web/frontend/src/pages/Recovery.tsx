import { AlertTriangle, LifeBuoy, Trash2 } from 'lucide-react'
import { formatDate } from '../lib/api'
import { api, useRecoveryStatus } from '../hooks/useApi'

export default function Recovery() {
  const recoveryQuery = useRecoveryStatus()

  async function refresh() {
    await recoveryQuery.refetch()
  }

  if (recoveryQuery.isLoading) {
    return <div className="loading-state">Loading recovery data…</div>
  }

  const data = recoveryQuery.data
  const crashed = data?.crashed ?? []
  const active = data?.active ?? []
  const orphaned = data?.orphaned ?? []

  return (
    <div className="page-stack">
      <section className="page-heading">
        <div>
          <div className="eyebrow">Recovery</div>
          <h2>Inspect unhealthy or orphaned sessions.</h2>
          <p>Use the web UI to mark stale sessions, recover crashed entries, and clean orphaned records.</p>
        </div>
      </section>

      <section className="row-actions">
        <button className="primary-button" onClick={() => api.recovery.markStale().then(refresh)}>
          Mark Stale Sessions
        </button>
        <button className="secondary-button" onClick={() => api.recovery.recoverAll().then(refresh)}>
          Recover All Crashed
        </button>
        <button className="secondary-button danger" onClick={() => api.recovery.deleteOrphans().then(refresh)}>
          Remove Orphans
        </button>
      </section>

      <section className="stats-grid">
        <article className="stat-card stat-card-red">
          <div className="stat-card-header">
            <span>Crashed</span>
            <AlertTriangle size={18} />
          </div>
          <div className="stat-card-value">{crashed.length}</div>
        </article>
        <article className="stat-card stat-card-green">
          <div className="stat-card-header">
            <span>Active</span>
            <LifeBuoy size={18} />
          </div>
          <div className="stat-card-value">{active.length}</div>
        </article>
        <article className="stat-card stat-card-sand">
          <div className="stat-card-header">
            <span>Orphaned</span>
            <Trash2 size={18} />
          </div>
          <div className="stat-card-value">{orphaned.length}</div>
        </article>
      </section>

      <section className="panel-grid two-columns">
        <StatusColumn
          title="Crashed Sessions"
          sessions={crashed}
          emptyLabel="No crashed sessions."
          actionLabel="Recover"
          onAction={(id) => api.recovery.recoverOne(id).then(refresh)}
        />
        <StatusColumn title="Active Sessions" sessions={active} emptyLabel="No active sessions." />
      </section>

      <section className="panel-card">
        <div className="panel-header">
          <div>
            <h3>Orphaned Sessions</h3>
            <p>Database records with missing files.</p>
          </div>
        </div>
        <div className="list-stack">
          {orphaned.map((session) => (
            <div key={session.id} className="list-row">
              <div>
                <strong>#{session.id} · {session.metadata.client || 'Unknown client'}</strong>
                <div className="subdued-text">{session.display_path || session.path}</div>
              </div>
              <div className="list-row-meta">
                <span>{formatDate(session.mod_time)}</span>
              </div>
            </div>
          ))}
          {orphaned.length === 0 && <div className="empty-state compact">No orphaned sessions.</div>}
        </div>
      </section>
    </div>
  )
}

function StatusColumn({
  title,
  sessions,
  emptyLabel,
  actionLabel,
  onAction,
}: {
  title: string
  sessions: Array<{
    id: number
    mod_time: string
    metadata: { client: string; engagement: string; phase: string }
    size_human: string
  }>
  emptyLabel: string
  actionLabel?: string
  onAction?: (id: number) => void
}) {
  return (
    <article className="panel-card">
      <div className="panel-header">
        <div>
          <h3>{title}</h3>
          <p>{sessions.length} session{sessions.length === 1 ? '' : 's'}</p>
        </div>
      </div>
      <div className="list-stack">
        {sessions.map((session) => (
          <div key={session.id} className="list-row">
            <div>
              <strong>#{session.id} · {session.metadata.client || 'Unknown client'}</strong>
              <div className="subdued-text">
                {session.metadata.engagement || 'No engagement'} · {session.metadata.phase || 'No phase'}
              </div>
            </div>
            <div className="row-actions">
              <span className="pill">{session.size_human}</span>
              <span className="subdued-text">{formatDate(session.mod_time)}</span>
              {actionLabel && onAction && (
                <button className="secondary-button" onClick={() => onAction(session.id)}>
                  {actionLabel}
                </button>
              )}
            </div>
          </div>
        ))}
        {sessions.length === 0 && <div className="empty-state compact">{emptyLabel}</div>}
      </div>
    </article>
  )
}
