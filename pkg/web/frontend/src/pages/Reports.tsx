import { ExternalLink, FileCode2, FileText } from 'lucide-react'
import { formatDate } from '../lib/api'
import { useReports } from '../hooks/useApi'

export default function Reports() {
  const { data, isLoading } = useReports()
  const reports = data?.reports ?? []
  const htmlReports = reports.filter((report) => report.type === 'html').length

  if (isLoading) {
    return <div className="loading-state">Loading reports…</div>
  }

  return (
    <div className="page-stack">
      <section className="page-heading">
        <div>
          <div className="eyebrow">Reports</div>
          <h2>Review exported engagement artifacts.</h2>
          <p>The web UI now opens generated reports directly. Creation still happens through the CLI export flow.</p>
        </div>
      </section>

      <section className="stats-grid">
        <article className="stat-card stat-card-sand">
          <div className="stat-card-header">
            <span>Total Reports</span>
            <FileText size={18} />
          </div>
          <div className="stat-card-value">{reports.length}</div>
        </article>
        <article className="stat-card stat-card-blue">
          <div className="stat-card-header">
            <span>HTML Reports</span>
            <FileCode2 size={18} />
          </div>
          <div className="stat-card-value">{htmlReports}</div>
        </article>
      </section>

      <section className="panel-card">
        <div className="table-shell">
          <table>
            <thead>
              <tr>
                <th>Name</th>
                <th>Client</th>
                <th>Type</th>
                <th>Size</th>
                <th>Modified</th>
                <th>Stored As</th>
                <th>Action</th>
              </tr>
            </thead>
            <tbody>
              {reports.map((report) => (
                <tr key={`${report.client}-${report.name}`}>
                  <td>
                    <div className="table-title-cell">
                      {report.type === 'html' ? <FileCode2 size={16} /> : <FileText size={16} />}
                      <span>{report.name}</span>
                    </div>
                  </td>
                  <td>{report.client}</td>
                  <td>
                    <span className={`badge ${report.type === 'html' ? 'badge-blue' : 'badge-sand'}`}>{report.type.toUpperCase()}</span>
                  </td>
                  <td>{report.size_human}</td>
                  <td>{formatDate(report.mod_time)}</td>
                  <td className="mono-text">{report.relative_path}</td>
                  <td>
                    <a className="inline-link" href={report.view_url} target="_blank" rel="noreferrer">
                      {report.type === 'html' ? 'Open report' : 'View file'} <ExternalLink size={14} />
                    </a>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {reports.length === 0 && <div className="empty-state compact">No exported reports found.</div>}
        </div>
      </section>
    </div>
  )
}
