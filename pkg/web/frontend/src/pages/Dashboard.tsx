import {
  AlertTriangle,
  Archive,
  Database,
  ExternalLink,
  FileCode2,
  FolderOpen,
  Radio,
  Search,
  Shield,
  StickyNote,
  Users,
} from 'lucide-react'
import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  Pie,
  PieChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'
import { Link } from 'react-router-dom'
import { formatDate, formatListLabel } from '../lib/api'
import { useDashboardOverview, useShareStatus } from '../hooks/useApi'

const severityPalette = ['#d9485f', '#ef8f56', '#e5ba52', '#48a36d', '#4b88a2']

function withQuery(path: string, params: Record<string, string | undefined>) {
  const search = new URLSearchParams()
  Object.entries(params).forEach(([key, value]) => {
    if (value && value.trim() !== '') {
      search.set(key, value)
    }
  })
  const query = search.toString()
  return query ? `${path}?${query}` : path
}

export default function Dashboard() {
  const overviewQuery = useDashboardOverview()
  const { data: share } = useShareStatus()

  if (overviewQuery.isLoading) {
    return <div className="loading-state">Loading dashboard…</div>
  }

  if (overviewQuery.isError || !overviewQuery.data) {
    return (
      <div className="page-stack">
        <section className="panel-card" aria-live="polite">
          <div className="panel-header">
            <div>
              <h3>Dashboard Unavailable</h3>
              <p>The dashboard snapshot failed to load. Retry after confirming local service and storage health.</p>
            </div>
            <AlertTriangle size={18} />
          </div>
          <div className="row-actions">
            <button className="primary-button" onClick={() => overviewQuery.refetch()}>
              Retry
            </button>
            <Link className="secondary-button" to="/recovery">
              Open recovery
            </Link>
          </div>
        </section>
      </div>
    )
  }

  const overview = overviewQuery.data
  const stats = overview.stats
  const activity = overview.activity
  const clients = overview.clients
  const context = overview.context.context
  const hasContext = overview.context.has_context && context !== null
  const artifacts = overview.artifacts

  const phaseData = Object.entries(stats.phase_counts ?? {}).map(([name, value]) => ({
    name: formatListLabel(name),
    value,
  }))

  const severityData = Object.entries(stats.severity_counts ?? {}).map(([name, value]) => ({
    name,
    value,
  }))

  const stateEntries = Object.entries(stats.state_counts ?? {})
    .map(([name, value]) => ({ name: formatListLabel(name), raw: name, value }))
    .sort((a, b) => b.value - a.value)

  const scopedSessionsLink = withQuery('/sessions', {
    client: context?.client,
    phase: context?.phase,
  })
  const scopedSearchLink = withQuery('/search', {
    q: [context?.client, context?.phase].filter(Boolean).join(' ') || undefined,
  })
  const scopedReportsLink = withQuery('/reports', {
    client: context?.client,
    engagement: context?.engagement,
    phase: context?.phase,
  })
  const scopedArchivesLink = withQuery('/archives', {
    client: context?.client,
  })

  const statCards = [
    { label: 'Sessions', value: stats.total_sessions ?? 0, icon: FolderOpen, tone: 'amber' },
    { label: 'Evidence Size', value: stats.total_size_human ?? '0 B', icon: Database, tone: 'blue' },
    { label: 'Clients', value: stats.unique_clients ?? 0, icon: Users, tone: 'green' },
    { label: 'Notes', value: stats.total_notes ?? 0, icon: StickyNote, tone: 'sand' },
    { label: 'Findings', value: stats.total_vulns ?? 0, icon: Shield, tone: 'red' },
  ]

  return (
    <div className="page-stack dashboard-page">
      <section className="hero-card" aria-labelledby="dashboard-hero-heading">
        <div>
          <div className="eyebrow">Web Dashboard</div>
          <h2 id="dashboard-hero-heading" className="hero-title">
            Mission control for session evidence and reporting flow.
          </h2>
          <p className="hero-copy">
            Monitor active context, session health, findings, and artifact readiness from one browser view.
          </p>
        </div>
        <div className="hero-metrics">
          <div>
            <span className="hero-metric-label">Engagements</span>
            <strong>{stats.unique_engagements ?? 0}</strong>
          </div>
          <div>
            <span className="hero-metric-label">Latest activity</span>
            <strong>{formatDate(stats.last_activity)}</strong>
          </div>
        </div>
      </section>

      <section className="stats-grid stats-grid-wide" aria-label="Primary metrics">
        {statCards.map((card) => {
          const Icon = card.icon
          return (
            <article key={card.label} className={`stat-card stat-card-${card.tone}`}>
              <div className="stat-card-header">
                <span>{card.label}</span>
                <Icon size={18} />
              </div>
              <div className="stat-card-value">{card.value}</div>
            </article>
          )
        })}
      </section>

      <section className="panel-grid two-columns" aria-label="Operational status">
        <article className="panel-card">
          <div className="panel-header">
            <div>
              <h3>Active Context</h3>
              <p>Current execution scope driving session capture and reporting.</p>
            </div>
          </div>
          {hasContext ? (
            <div className="list-stack">
              <div className="list-row">
                <div>
                  <strong>{context?.client || 'Unknown client'}</strong>
                  <div className="subdued-text">
                    {context?.engagement || 'No engagement'} · {context?.phase || 'No phase'}
                  </div>
                </div>
                <span className="badge badge-active">In Progress</span>
              </div>
              <div className="pill-row">
                {context?.operator && <span className="pill">Operator: {context.operator}</span>}
                {context?.target && <span className="pill pill-accent">{context.target}</span>}
                {context?.target_ip && <span className="pill">{context.target_ip}</span>}
              </div>
              <div className="row-actions">
                <Link className="inline-link" to={scopedSessionsLink}>
                  Open scoped sessions
                </Link>
                <Link className="inline-link" to="/settings">
                  Edit context
                </Link>
              </div>
            </div>
          ) : (
            <div className="empty-state compact">
              No active context. Open Settings to create one, then run shell capture to populate the dashboard pipeline.
            </div>
          )}
        </article>

        <article className="panel-card">
          <div className="panel-header">
            <div>
              <h3>Session Health</h3>
              <p>State distribution with non-color labels for triage.</p>
            </div>
          </div>
          {stateEntries.length > 0 ? (
            <div className="list-stack">
              {stateEntries.map((state) => (
                <div key={state.raw} className="list-row">
                  <div>
                    <strong>{state.name}</strong>
                    <div className="subdued-text">State: {state.raw}</div>
                  </div>
                  <span className={`badge badge-${state.raw.toLowerCase()}`}>{state.value} sessions</span>
                </div>
              ))}
            </div>
          ) : (
            <div className="empty-state compact">No session states available yet.</div>
          )}
        </article>
      </section>

      <section className="panel-grid two-columns" aria-label="Workflow actions">
        <article className="panel-card">
          <div className="panel-header">
            <div>
              <h3>Workflow Actions</h3>
              <p>Jump directly into the next step of the `create → shell → search/timeline → export → archive` flow.</p>
            </div>
          </div>
          <div className="list-stack">
            <Link className="list-row interactive-row" to={scopedSessionsLink}>
              <div>
                <strong>Review sessions</strong>
                <div className="subdued-text">Inspect current scope recordings and state.</div>
              </div>
            </Link>
            <Link className="list-row interactive-row" to={scopedSearchLink}>
              <div>
                <strong>Search evidence</strong>
                <div className="subdued-text">Open search with context-seeded query terms.</div>
              </div>
              <Search size={16} />
            </Link>
            <Link className="list-row interactive-row" to={scopedReportsLink}>
              <div>
                <strong>Generate report</strong>
                <div className="subdued-text">Reports page opens with context prefilled.</div>
              </div>
              <FileCode2 size={16} />
            </Link>
            <Link className="list-row interactive-row" to={scopedArchivesLink}>
              <div>
                <strong>Review archives</strong>
                <div className="subdued-text">Validate package readiness before delivery.</div>
              </div>
              <Archive size={16} />
            </Link>
          </div>
        </article>

        <article className="panel-card">
          <div className="panel-header">
            <div>
              <h3>Artifact Pipeline</h3>
              <p>Report and archive readiness from disk.</p>
            </div>
          </div>
          <div className="stats-grid">
            <article className="stat-card stat-card-blue">
              <div className="stat-card-header">
                <span>Reports</span>
                <FileCode2 size={18} />
              </div>
              <div className="stat-card-value">{artifacts.reports_total}</div>
            </article>
            <article className="stat-card stat-card-sand">
              <div className="stat-card-header">
                <span>Archives</span>
                <Archive size={18} />
              </div>
              <div className="stat-card-value">{artifacts.archives_total}</div>
            </article>
          </div>
          <div className="list-stack">
            <div className="list-row">
              <div>
                <strong>{artifacts.latest_report?.name ?? 'No report generated yet'}</strong>
                <div className="subdued-text">{formatDate(artifacts.latest_report?.mod_time)}</div>
              </div>
              <div className="row-actions">
                {artifacts.latest_report?.url && (
                  <a className="inline-link" href={artifacts.latest_report.url} target="_blank" rel="noreferrer">
                    Open <ExternalLink size={14} />
                  </a>
                )}
                <Link className="inline-link" to="/reports">
                  Reports
                </Link>
              </div>
            </div>
            <div className="list-row">
              <div>
                <strong>{artifacts.latest_archive?.name ?? 'No archive generated yet'}</strong>
                <div className="subdued-text">{formatDate(artifacts.latest_archive?.mod_time)}</div>
              </div>
              <div className="row-actions">
                {artifacts.latest_archive?.url && (
                  <a className="inline-link" href={artifacts.latest_archive.url}>
                    Download
                  </a>
                )}
                <Link className="inline-link" to="/archives">
                  Archives
                </Link>
              </div>
            </div>
          </div>
        </article>
      </section>

      <section className="panel-grid two-columns" aria-label="Charts">
        <article className="panel-card chart-card">
          <div className="panel-header">
            <div>
              <h3>Phase Distribution</h3>
              <p>Session count by engagement phase.</p>
            </div>
          </div>
          {phaseData.length > 0 ? (
            <div className="chart-area">
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={phaseData} margin={{ top: 4, right: 8, bottom: 4, left: -16 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke="rgba(75, 103, 108, 0.14)" vertical={false} />
                  <XAxis dataKey="name" tickLine={false} axisLine={false} tick={{ fill: '#496267', fontSize: 12 }} />
                  <YAxis tickLine={false} axisLine={false} tick={{ fill: '#496267', fontSize: 12 }} allowDecimals={false} />
                  <Tooltip cursor={{ fill: 'rgba(194, 142, 74, 0.08)' }} />
                  <Bar dataKey="value" fill="#c28e4a" radius={[8, 8, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            </div>
          ) : (
            <div className="empty-state compact">No phase data available yet.</div>
          )}
        </article>

        <article className="panel-card chart-card">
          <div className="panel-header">
            <div>
              <h3>Severity Mix</h3>
              <p>Current vulnerability distribution.</p>
            </div>
          </div>
          {severityData.length > 0 ? (
            <div className="chart-area pie-chart-area">
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie data={severityData} dataKey="value" nameKey="name" innerRadius={55} outerRadius={86} paddingAngle={4}>
                    {severityData.map((entry, index) => (
                      <Cell key={entry.name} fill={severityPalette[index % severityPalette.length]} />
                    ))}
                  </Pie>
                  <Tooltip />
                </PieChart>
              </ResponsiveContainer>
              <div className="legend-list">
                {severityData.map((entry, index) => (
                  <div key={entry.name} className="legend-row">
                    <span className="legend-swatch" style={{ backgroundColor: severityPalette[index % severityPalette.length] }} />
                    <span>{entry.name}</span>
                    <strong>{entry.value}</strong>
                  </div>
                ))}
              </div>
            </div>
          ) : (
            <div className="empty-state compact">No findings recorded yet.</div>
          )}
        </article>
      </section>

      <section className="panel-grid two-columns" aria-label="Recent activity">
        <article className="panel-card">
          <div className="panel-header">
            <div>
              <h3>Recent Sessions</h3>
              <p>Latest captured shell recordings.</p>
            </div>
            <Link className="inline-link" to="/sessions">
              View all
            </Link>
          </div>
          <div className="list-stack">
            {(activity.recent_sessions ?? []).slice(0, 6).map((session) => (
              <Link key={session.id} to={`/sessions/${session.id}`} className="list-row interactive-row">
                <div>
                  <strong>{session.metadata.client || 'Unknown client'}</strong>
                  <div className="subdued-text">
                    {session.metadata.engagement || 'No engagement'} · {session.metadata.phase || 'No phase'}
                  </div>
                </div>
                <div className="list-row-meta">
                  <span className={`badge badge-${session.state.toLowerCase()}`}>{formatListLabel(session.state)}</span>
                  <span>{formatDate(session.mod_time)}</span>
                </div>
              </Link>
            ))}
            {(activity.recent_sessions ?? []).length === 0 && <div className="empty-state compact">No sessions recorded yet.</div>}
          </div>
        </article>

        <article className="panel-card">
          <div className="panel-header">
            <div>
              <h3>Recent Findings</h3>
              <p>Newest vulnerability entries.</p>
            </div>
            <Link className="inline-link" to="/vulns">
              Open tracker
            </Link>
          </div>
          <div className="list-stack">
            {(activity.recent_vulns ?? []).slice(0, 6).map((vuln) => (
              <div key={vuln.id} className="list-row">
                <div>
                  <strong>{vuln.title}</strong>
                  <div className="subdued-text">{vuln.phase || 'Unassigned phase'}</div>
                </div>
                <div className="list-row-meta">
                  <span className={`badge badge-${vuln.severity.toLowerCase()}`}>{vuln.severity}</span>
                  <span>{formatDate(vuln.created_at)}</span>
                </div>
              </div>
            ))}
            {(activity.recent_vulns ?? []).length === 0 && <div className="empty-state compact">No findings tracked yet.</div>}
          </div>
        </article>
      </section>

      <section className="panel-card" aria-label="Live share status">
        <div className="panel-header">
          <div>
            <h3>Live Share</h3>
            <p>Current browser streaming status from the existing share service.</p>
          </div>
          {share?.active && share.watch_url ? (
            <a className="inline-link" href={share.watch_url} target="_blank" rel="noreferrer">
              Open watch <ExternalLink size={14} />
            </a>
          ) : null}
        </div>
        {share?.active ? (
          <div className="list-stack">
            <div className="list-row">
              <div>
                <strong>{share.reachable ? 'Live session reachable' : 'Live session registered'}</strong>
                <div className="subdued-text mono-text">{share.log_file ?? '-'}</div>
              </div>
              <div className="list-row-meta">
                <span className={`badge ${share.reachable ? 'badge-green' : 'badge-amber'}`}>{share.reachable ? 'Online' : 'Pending'}</span>
                <span className="pill">
                  <Radio size={14} /> {share.clients ?? 0} viewers
                </span>
              </div>
            </div>
          </div>
        ) : (
          <div className="empty-state compact">No active live share session.</div>
        )}
      </section>

      <section className="panel-card" aria-label="Client footprint">
        <div className="panel-header">
          <div>
            <h3>Client Footprint</h3>
            <p>Largest evidence collections by client.</p>
          </div>
        </div>
        <div className="table-shell compact-table">
          <table>
            <thead>
              <tr>
                <th>Client</th>
                <th>Sessions</th>
                <th>Total Size</th>
              </tr>
            </thead>
            <tbody>
              {clients.slice(0, 8).map((client) => (
                <tr key={client.name}>
                  <td>{client.name}</td>
                  <td>{client.sessions_count}</td>
                  <td>{client.total_size_human}</td>
                </tr>
              ))}
            </tbody>
          </table>
          {clients.length === 0 && <div className="empty-state compact">No client data available.</div>}
        </div>
      </section>
    </div>
  )
}
