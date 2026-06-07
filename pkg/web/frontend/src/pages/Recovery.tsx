import { AlertTriangle, LifeBuoy, PauseCircle, ShieldAlert, Trash2 } from 'lucide-react'
import { formatDate } from '../lib/api'
import { api, useRecoveryStatus } from '../hooks/useApi'
import type { RecoverySessionStatus } from '../lib/api'

export default function Recovery() {
  const recoveryQuery = useRecoveryStatus()

  async function refresh() {
    await recoveryQuery.refetch()
  }

  if (recoveryQuery.isLoading) {
    return <div className="loading-state">Loading recovery data…</div>
  }

  const data = recoveryQuery.data
  const active = data?.active ?? []
  const paused = data?.paused ?? []
  const reviewNeeded = data?.review_needed ?? []
  const stale = data?.stale ?? []
  const crashed = data?.crashed ?? []
  const orphaned = data?.orphaned ?? []
  const timeoutMinutes = data?.stale_timeout_minutes ?? 0

  return (
    <div className="page-stack">
      <section className="page-heading">
        <div>
          <div className="eyebrow">Recovery</div>
          <h2>Inspect live, stale, crashed, or orphaned sessions.</h2>
          <p>
            PentLog only auto-crashes sessions it can prove are dead. The current stale timeout is {timeoutMinutes} minute{timeoutMinutes === 1 ? '' : 's'}.
          </p>
        </div>
      </section>

      <section className="row-actions">
        <button className="primary-button" onClick={() => api.recovery.markStale().then(refresh)}>
          Mark Definitely Stale
        </button>
        <button className="secondary-button" onClick={() => api.recovery.recoverAll().then(refresh)}>
          Recover All Crashed
        </button>
        <button className="secondary-button danger" onClick={() => api.recovery.deleteOrphans().then(refresh)}>
          Remove Orphans
        </button>
      </section>

      <section className="stats-grid">
        <article className="stat-card stat-card-green">
          <div className="stat-card-header">
            <span>Likely Live</span>
            <LifeBuoy size={18} />
          </div>
          <div className="stat-card-value">{active.length}</div>
        </article>
        <article className="stat-card stat-card-sand">
          <div className="stat-card-header">
            <span>Paused</span>
            <PauseCircle size={18} />
          </div>
          <div className="stat-card-value">{paused.length}</div>
        </article>
        <article className="stat-card stat-card-red">
          <div className="stat-card-header">
            <span>Crashed</span>
            <AlertTriangle size={18} />
          </div>
          <div className="stat-card-value">{crashed.length}</div>
        </article>
        <article className="stat-card stat-card-sand">
          <div className="stat-card-header">
            <span>Needs Review</span>
            <ShieldAlert size={18} />
          </div>
          <div className="stat-card-value">{reviewNeeded.length}</div>
        </article>
        <article className="stat-card stat-card-red">
          <div className="stat-card-header">
            <span>Definitely Stale</span>
            <AlertTriangle size={18} />
          </div>
          <div className="stat-card-value">{stale.length}</div>
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
        <StatusColumn title="Likely-Live Active Sessions" sessions={active} emptyLabel="No likely-live active sessions." />
        <StatusColumn title="Paused Sessions" sessions={paused} emptyLabel="No paused sessions." />
      </section>

      <section className="panel-grid two-columns">
        <StatusColumn title="Review-Needed Sessions" sessions={reviewNeeded} emptyLabel="No review-needed sessions." />
        <StatusColumn title="Definitely Stale Sessions" sessions={stale} emptyLabel="No definitely stale sessions." />
      </section>

      <section className="panel-card">
        <div className="panel-header">
          <div>
            <h3>Crashed Sessions</h3>
            <p>Sessions already marked crashed and ready to recover.</p>
          </div>
        </div>
        <div className="list-stack">
          {crashed.map((entry) => (
            <div key={entry.session.id} className="list-row">
              <SessionSummary entry={entry} />
              <div className="row-actions">
                <button className="secondary-button" onClick={() => api.recovery.recoverOne(entry.session.id).then(refresh)}>
                  Recover
                </button>
              </div>
            </div>
          ))}
          {crashed.length === 0 && <div className="empty-state compact">No crashed sessions.</div>}
        </div>
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
}: {
  title: string
  sessions: RecoverySessionStatus[]
  emptyLabel: string
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
        {sessions.map((entry) => (
          <div key={entry.session.id} className="list-row">
            <SessionSummary entry={entry} />
          </div>
        ))}
        {sessions.length === 0 && <div className="empty-state compact">{emptyLabel}</div>}
      </div>
    </article>
  )
}

function SessionSummary({ entry }: { entry: RecoverySessionStatus }) {
  const session = entry.session

  return (
    <div>
      <strong>#{session.id} · {session.metadata.client || 'Unknown client'}</strong>
      <div className="subdued-text">
        {session.metadata.engagement || 'No engagement'} · {session.metadata.phase || 'No phase'} · {session.state}
      </div>
      <div className="subdued-text">
        {entry.reason}
      </div>
      <div className="subdued-text">
        Last heartbeat: {entry.last_seen_age || formatDate(entry.last_seen_at)} · Size: {session.size_human}
      </div>
      {(session.recorder_pid || session.hostname || session.resume_count) && (
        <div className="subdued-text">
          {session.recorder_pid ? `PID ${session.recorder_pid}` : 'No PID'}
          {session.hostname ? ` · ${session.hostname}` : ''}
          {session.resume_count ? ` · resumed ${session.resume_count}x` : ''}
        </div>
      )}
    </div>
  )
}
