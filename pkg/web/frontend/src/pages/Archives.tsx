import { Archive, Download, Lock } from 'lucide-react'
import { formatDate } from '../lib/api'
import { useArchives } from '../hooks/useApi'

export default function Archives() {
  const { data, isLoading } = useArchives()
  const archives = data?.archives ?? []
  const encryptedCount = archives.filter((archive) => archive.encrypted).length

  if (isLoading) {
    return <div className="loading-state">Loading archives…</div>
  }

  return (
    <div className="page-stack">
      <section className="page-heading">
        <div>
          <div className="eyebrow">Archives</div>
          <h2>Review packaged evidence bundles.</h2>
          <p>The web UI now provides direct archive downloads; creation still uses the CLI archive flow.</p>
        </div>
      </section>

      <section className="stats-grid">
        <article className="stat-card stat-card-blue">
          <div className="stat-card-header">
            <span>Total Archives</span>
            <Archive size={18} />
          </div>
          <div className="stat-card-value">{archives.length}</div>
        </article>
        <article className="stat-card stat-card-red">
          <div className="stat-card-header">
            <span>Encrypted</span>
            <Lock size={18} />
          </div>
          <div className="stat-card-value">{encryptedCount}</div>
        </article>
      </section>

      <section className="panel-card">
        <div className="table-shell">
          <table>
            <thead>
              <tr>
                <th>Name</th>
                <th>Client</th>
                <th>Protected</th>
                <th>Size</th>
                <th>Modified</th>
                <th>Stored As</th>
                <th>Action</th>
              </tr>
            </thead>
            <tbody>
              {archives.map((archive) => (
                <tr key={`${archive.client}-${archive.name}`}>
                  <td>{archive.name}</td>
                  <td>{archive.client}</td>
                  <td>
                    {archive.encrypted ? <span className="badge badge-red">Encrypted</span> : <span className="badge badge-green">Standard</span>}
                  </td>
                  <td>{archive.size_human}</td>
                  <td>{formatDate(archive.mod_time)}</td>
                  <td className="mono-text">{archive.relative_path}</td>
                  <td>
                    <a className="inline-link" href={archive.download_url} download>
                      Download <Download size={14} />
                    </a>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {archives.length === 0 && <div className="empty-state compact">No archives found.</div>}
        </div>
      </section>
    </div>
  )
}
