import { Shield } from 'lucide-react'
import { formatDate } from '../lib/api'
import { useVulns } from '../hooks/useApi'

const severityClass: Record<string, string> = {
  critical: 'badge-red',
  high: 'badge-red',
  medium: 'badge-amber',
  low: 'badge-green',
  info: 'badge-blue',
}

export default function Vulns() {
  const { data, isLoading } = useVulns()
  const vulns = data?.vulns ?? []

  if (isLoading) {
    return <div className="loading-state">Loading vulnerabilities…</div>
  }

  return (
    <div className="page-stack">
      <section className="page-heading">
        <div>
          <div className="eyebrow">Vulnerability Tracker</div>
          <h2>Review findings captured for the active context.</h2>
          <p>The current dashboard supports listing and inspection. Editing still relies on the CLI flow.</p>
        </div>
        <div className="page-heading-meta">
          <span className="pill">{data?.total ?? 0} findings</span>
        </div>
      </section>

      <section className="card-grid">
        {vulns.map((vuln) => (
          <article key={vuln.id} className="panel-card vuln-card">
            <div className="panel-header">
              <div>
                <div className="badge-row">
                  <span className={`badge ${severityClass[vuln.severity.toLowerCase()] ?? 'badge-blue'}`}>{vuln.severity}</span>
                  <span className="badge badge-sand">{vuln.status}</span>
                </div>
                <h3>{vuln.title}</h3>
                <p>{vuln.phase || 'Unassigned phase'}</p>
              </div>
              <Shield size={18} />
            </div>

            {vuln.description && <p className="body-copy">{vuln.description}</p>}

            <div className="meta-grid compact-meta-grid">
              <div>
                <span>Created</span>
                <strong>{formatDate(vuln.created_at)}</strong>
              </div>
              <div>
                <span>ID</span>
                <strong className="mono-text">{vuln.id}</strong>
              </div>
            </div>

            {vuln.evidence?.length > 0 && (
              <div className="tag-cloud">
                {vuln.evidence.slice(0, 3).map((evidence) => (
                  <span key={evidence} className="pill">
                    {evidence}
                  </span>
                ))}
              </div>
            )}
          </article>
        ))}
        {vulns.length === 0 && <div className="empty-state">No vulnerabilities tracked yet.</div>}
      </section>
    </div>
  )
}
