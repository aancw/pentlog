import { Archive, Database, ExternalLink, FileCode2, FolderOpen, Radio, Shield, StickyNote, Users } from 'lucide-react'
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
import { useArchives, useDashboardActivity, useDashboardClients, useDashboardStats, useReports, useShareStatus } from '../hooks/useApi'

const severityPalette = ['#d9485f', '#ef8f56', '#e5ba52', '#48a36d', '#4b88a2']

export default function Dashboard() {
  const { data: stats, isLoading: statsLoading } = useDashboardStats()
  const { data: activity, isLoading: activityLoading } = useDashboardActivity()
  const { data: clients, isLoading: clientsLoading } = useDashboardClients()
  const { data: reports } = useReports()
  const { data: archives } = useArchives()
  const { data: share } = useShareStatus()

  if (statsLoading || activityLoading || clientsLoading) {
    return <div className="loading-state">Loading dashboard…</div>
  }

  const phaseData = Object.entries(stats?.phase_counts ?? {}).map(([name, value]) => ({
    name: formatListLabel(name),
    value,
  }))

  const severityData = Object.entries(stats?.severity_counts ?? {}).map(([name, value]) => ({
    name,
    value,
  }))

  const statCards = [
    {
      label: 'Sessions',
      value: stats?.total_sessions ?? 0,
      icon: FolderOpen,
      tone: 'amber',
    },
    {
      label: 'Evidence Size',
      value: stats?.total_size_human ?? '0 B',
      icon: Database,
      tone: 'blue',
    },
    {
      label: 'Clients',
      value: stats?.unique_clients ?? 0,
      icon: Users,
      tone: 'green',
    },
    {
      label: 'Notes',
      value: stats?.total_notes ?? 0,
      icon: StickyNote,
      tone: 'sand',
    },
    {
      label: 'Findings',
      value: stats?.total_vulns ?? 0,
      icon: Shield,
      tone: 'red',
    },
  ]

  const reportCount = reports?.reports.length ?? 0
  const archiveCount = archives?.archives.length ?? 0
  const latestReport = reports?.reports[0]

  return (
    <div className="page-stack">
      <section className="hero-card">
        <div>
          <div className="eyebrow">Web Dashboard</div>
          <h2 className="hero-title">Operational visibility for your evidence trail.</h2>
          <p className="hero-copy">
            Track recorded shell sessions, review findings, and move from raw evidence to reporting without leaving the browser.
          </p>
        </div>
        <div className="hero-metrics">
          <div>
            <span className="hero-metric-label">Engagements</span>
            <strong>{stats?.unique_engagements ?? 0}</strong>
          </div>
          <div>
            <span className="hero-metric-label">Latest activity</span>
            <strong>{activity?.recent_sessions?.[0] ? formatDate(activity.recent_sessions[0].mod_time) : 'No data'}</strong>
          </div>
        </div>
      </section>

      <section className="stats-grid stats-grid-wide">
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

      <section className="panel-grid two-columns">
        <article className="panel-card">
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
              <div className="pill-row">
                <span className="pill">Port {share.port ?? '-'}</span>
                {(share.client_ips ?? []).slice(0, 3).map((ip) => (
                  <span key={ip} className="pill pill-accent">{ip}</span>
                ))}
              </div>
            </div>
          ) : (
            <div className="empty-state compact">No active live share session.</div>
          )}
        </article>

        <article className="panel-card">
          <div className="panel-header">
            <div>
              <h3>Artifact Pipeline</h3>
              <p>Existing reports and archives surfaced directly from disk.</p>
            </div>
          </div>
          <div className="stats-grid">
            <article className="stat-card stat-card-blue">
              <div className="stat-card-header">
                <span>Reports</span>
                <FileCode2 size={18} />
              </div>
              <div className="stat-card-value">{reportCount}</div>
            </article>
            <article className="stat-card stat-card-sand">
              <div className="stat-card-header">
                <span>Archives</span>
                <Archive size={18} />
              </div>
              <div className="stat-card-value">{archiveCount}</div>
            </article>
          </div>
          <div className="list-stack">
            <div className="list-row">
              <div>
                <strong>{latestReport?.name ?? 'No report exported yet'}</strong>
                <div className="subdued-text">{latestReport ? formatDate(latestReport.mod_time) : 'Run the CLI export flow to generate the first report.'}</div>
              </div>
              <div className="row-actions">
                <Link className="inline-link" to="/reports">
                  Reports
                </Link>
                <Link className="inline-link" to="/archives">
                  Archives
                </Link>
              </div>
            </div>
          </div>
        </article>
      </section>

      <section className="panel-grid two-columns">
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

      <section className="panel-grid two-columns">
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
            {(activity?.recent_sessions ?? []).slice(0, 6).map((session) => (
              <Link key={session.id} to={`/sessions/${session.id}`} className="list-row interactive-row">
                <div>
                  <strong>{session.metadata.client || 'Unknown client'}</strong>
                  <div className="subdued-text">
                    {session.metadata.engagement || 'No engagement'} · {session.metadata.phase || 'No phase'}
                  </div>
                </div>
                <div className="list-row-meta">
                  <span className="pill">{session.size_human}</span>
                  <span>{formatDate(session.mod_time)}</span>
                </div>
              </Link>
            ))}
            {activity?.recent_sessions?.length === 0 && <div className="empty-state compact">No sessions recorded yet.</div>}
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
            {(activity?.recent_vulns ?? []).slice(0, 6).map((vuln) => (
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
            {activity?.recent_vulns?.length === 0 && <div className="empty-state compact">No findings tracked yet.</div>}
          </div>
        </article>
      </section>

      <section className="panel-card">
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
              {(clients?.clients ?? []).slice(0, 8).map((client) => (
                <tr key={client.name}>
                  <td>{client.name}</td>
                  <td>{client.sessions_count}</td>
                  <td>{client.total_size_human}</td>
                </tr>
              ))}
            </tbody>
          </table>
          {clients?.clients?.length === 0 && <div className="empty-state compact">No client data available.</div>}
        </div>
      </section>
    </div>
  )
}
