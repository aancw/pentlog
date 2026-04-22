import { useMemo, useState } from 'react'
import type { FormEvent } from 'react'
import { Database, FolderOpen, RotateCcw, Server, Target } from 'lucide-react'
import { formatDate } from '../lib/api'
import { api, useContextHistory, useCurrentContext, useSystemInfo, useSystemStatus, useTargets } from '../hooks/useApi'

type ContextDraft = {
  client: string
  engagement: string
  scope: string
  operator: string
  phase: string
  target: string
  target_ip: string
  type: string
}

export default function Settings() {
  const statusQuery = useSystemStatus()
  const infoQuery = useSystemInfo()
  const contextQuery = useCurrentContext()
  const historyQuery = useContextHistory()
  const targetsQuery = useTargets()

  const activeContext = contextQuery.data?.context ?? statusQuery.data?.context ?? null
  const contextDefaults = useMemo<ContextDraft>(() => ({
    client: activeContext?.client ?? '',
    engagement: activeContext?.engagement ?? '',
    scope: activeContext?.scope ?? '',
    operator: activeContext?.operator ?? '',
    phase: activeContext?.phase ?? '',
    target: activeContext?.target ?? '',
    target_ip: activeContext?.target_ip ?? '',
    type: activeContext?.type ?? 'Client',
  }), [activeContext])

  const [contextOverrides, setContextOverrides] = useState<Partial<ContextDraft>>({})
  const [targetForm, setTargetForm] = useState({ name: '', ip: '' })

  function contextValue(field: keyof ContextDraft) {
    return contextOverrides[field] ?? contextDefaults[field]
  }

  function updateContext(field: keyof ContextDraft, value: string) {
    setContextOverrides((current) => ({ ...current, [field]: value }))
  }

  async function refreshAll() {
    await Promise.all([
      statusQuery.refetch(),
      contextQuery.refetch(),
      historyQuery.refetch(),
      targetsQuery.refetch(),
    ])
  }

  async function handleContextSubmit(event: FormEvent) {
    event.preventDefault()
    const payload: ContextDraft = {
      client: contextValue('client'),
      engagement: contextValue('engagement'),
      scope: contextValue('scope'),
      operator: contextValue('operator'),
      phase: contextValue('phase'),
      target: contextValue('target'),
      target_ip: contextValue('target_ip'),
      type: contextValue('type'),
    }

    if (!payload.client || !payload.engagement || !payload.phase) {
      return
    }

    if (activeContext) {
      await api.contexts.update(payload)
    } else {
      await api.contexts.create(payload)
    }

    setContextOverrides({})
    await refreshAll()
  }

  async function handleContextReset() {
    if (!window.confirm('Reset the active engagement context?')) {
      return
    }
    await api.contexts.reset()
    setContextOverrides({})
    await refreshAll()
  }

  async function handleCreateTarget(event: FormEvent) {
    event.preventDefault()
    if (!targetForm.name.trim()) {
      return
    }
    await api.targets.create({ name: targetForm.name.trim(), ip: targetForm.ip.trim() })
    setTargetForm({ name: '', ip: '' })
    await targetsQuery.refetch()
  }

  const paths = useMemo(
    () => [
      ['Home', infoQuery.data?.paths.home],
      ['Logs', infoQuery.data?.paths.logs_dir],
      ['Reports', infoQuery.data?.paths.reports_dir],
      ['Archives', infoQuery.data?.paths.archive_dir],
      ['Database', infoQuery.data?.paths.database_file],
    ],
    [infoQuery.data],
  )

  if (statusQuery.isLoading || infoQuery.isLoading || historyQuery.isLoading || targetsQuery.isLoading) {
    return <div className="loading-state">Loading settings…</div>
  }

  return (
    <div className="page-stack">
      <section className="page-heading">
        <div>
          <div className="eyebrow">Settings</div>
          <h2>System, context, and target management.</h2>
          <p>Web support now covers context updates, target switching, and core environment visibility.</p>
        </div>
      </section>

      <section className="panel-grid two-columns">
        <article className="panel-card">
          <div className="panel-header">
            <div>
              <h3>System Status</h3>
              <p>Runtime information for the local dashboard service.</p>
            </div>
            <Server size={18} />
          </div>
          <div className="meta-grid compact-meta-grid">
            <div>
              <span>Version</span>
              <strong>{statusQuery.data?.version ?? '-'}</strong>
            </div>
            <div>
              <span>Context</span>
              <strong>{statusQuery.data?.has_context ? 'Active' : 'Not set'}</strong>
            </div>
            <div>
              <span>Sessions</span>
              <strong>{statusQuery.data?.total_sessions ?? 0}</strong>
            </div>
            <div>
              <span>Uptime</span>
              <strong>{infoQuery.data?.uptime ?? '-'}</strong>
            </div>
          </div>
        </article>

        <article className="panel-card">
          <div className="panel-header">
            <div>
              <h3>Storage Paths</h3>
              <p>Resolved filesystem locations from the config manager.</p>
            </div>
            <Database size={18} />
          </div>
          <div className="list-stack">
            {paths.map(([label, value]) => (
              <div key={label} className="list-row stacked-row">
                <strong>{label}</strong>
                <code>{value || '-'}</code>
              </div>
            ))}
          </div>
        </article>
      </section>

      <section className="panel-grid two-columns">
        <article className="panel-card">
          <div className="panel-header">
            <div>
              <h3>{activeContext ? 'Update Context' : 'Create Context'}</h3>
              <p>Manage the current engagement from the web UI.</p>
            </div>
            <FolderOpen size={18} />
          </div>
          <form className="form-grid" onSubmit={handleContextSubmit}>
            <label className="field">
              <span>Client</span>
              <input value={contextValue('client')} onChange={(event) => updateContext('client', event.target.value)} />
            </label>
            <label className="field">
              <span>Engagement</span>
              <input value={contextValue('engagement')} onChange={(event) => updateContext('engagement', event.target.value)} />
            </label>
            <label className="field">
              <span>Phase</span>
              <input value={contextValue('phase')} onChange={(event) => updateContext('phase', event.target.value)} />
            </label>
            <label className="field">
              <span>Operator</span>
              <input value={contextValue('operator')} onChange={(event) => updateContext('operator', event.target.value)} />
            </label>
            <label className="field field-span-2">
              <span>Scope</span>
              <input value={contextValue('scope')} onChange={(event) => updateContext('scope', event.target.value)} />
            </label>
            <label className="field">
              <span>Target</span>
              <input value={contextValue('target')} onChange={(event) => updateContext('target', event.target.value)} />
            </label>
            <label className="field">
              <span>Target IP</span>
              <input value={contextValue('target_ip')} onChange={(event) => updateContext('target_ip', event.target.value)} />
            </label>
            <div className="row-actions">
              <button className="primary-button" type="submit">
                {activeContext ? 'Save Context' : 'Create Context'}
              </button>
              {activeContext && (
                <button className="secondary-button" type="button" onClick={handleContextReset}>
                  Reset Context
                </button>
              )}
            </div>
          </form>
        </article>

        <article className="panel-card">
          <div className="panel-header">
            <div>
              <h3>Targets</h3>
              <p>Manage stored targets and switch the active one.</p>
            </div>
            <Target size={18} />
          </div>
          <form className="form-grid tight-form" onSubmit={handleCreateTarget}>
            <label className="field">
              <span>Name</span>
              <input value={targetForm.name} onChange={(event) => setTargetForm({ ...targetForm, name: event.target.value })} />
            </label>
            <label className="field">
              <span>IP</span>
              <input value={targetForm.ip} onChange={(event) => setTargetForm({ ...targetForm, ip: event.target.value })} />
            </label>
            <button className="primary-button" type="submit">
              Add Target
            </button>
          </form>

          <div className="list-stack">
            {(targetsQuery.data?.targets ?? []).map((target) => (
              <div key={target.name} className="list-row">
                <div>
                  <strong>{target.name}</strong>
                  <div className="subdued-text">{target.ip || 'No IP set'}</div>
                </div>
                <div className="row-actions">
                  {target.is_current ? (
                    <span className="badge badge-green">Active</span>
                  ) : (
                    <button className="secondary-button" onClick={() => api.targets.switch(target.name).then(() => refreshAll())}>
                      Switch
                    </button>
                  )}
                  <button className="secondary-button" onClick={() => api.targets.delete(target.name).then(() => refreshAll())}>
                    Delete
                  </button>
                </div>
              </div>
            ))}
            {(targetsQuery.data?.targets ?? []).length === 0 && <div className="empty-state compact">No saved targets.</div>}
          </div>
          {targetsQuery.data?.current && (
            <button className="secondary-button" onClick={() => api.targets.clear().then(() => refreshAll())}>
              Clear Active Target
            </button>
          )}
        </article>
      </section>

      <section className="panel-card">
        <div className="panel-header">
          <div>
            <h3>Context History</h3>
            <p>Recent engagement states stored in `history.jsonl`.</p>
          </div>
          <RotateCcw size={18} />
        </div>
        <div className="table-shell compact-table">
          <table>
            <thead>
              <tr>
                <th>Client</th>
                <th>Engagement</th>
                <th>Phase</th>
                <th>Target</th>
                <th>Timestamp</th>
              </tr>
            </thead>
            <tbody>
              {(historyQuery.data?.history ?? [])
                .slice(-10)
                .reverse()
                .map((entry, index) => (
                  <tr key={`${entry.client}-${entry.engagement}-${index}`}>
                    <td>{entry.client}</td>
                    <td>{entry.engagement}</td>
                    <td>{entry.phase}</td>
                    <td>{entry.target || '-'}</td>
                    <td>{formatDate(entry.timestamp)}</td>
                  </tr>
                ))}
            </tbody>
          </table>
          {(historyQuery.data?.history ?? []).length === 0 && <div className="empty-state compact">No context history recorded yet.</div>}
        </div>
      </section>
    </div>
  )
}
