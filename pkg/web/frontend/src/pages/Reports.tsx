import { useEffect, useState, type FormEvent } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { ExternalLink, FileCode2, FileText } from 'lucide-react'
import { formatDate, formatDuration } from '../lib/api'
import { api } from '../hooks/useApi'
import { useCurrentContext, useReportJob, useReports, useActiveReportJob, useDashboardClients, useDashboardEngagements, useDashboardPhases } from '../hooks/useApi'

export default function Reports() {
  const queryClient = useQueryClient()
  const { data, isLoading, refetch } = useReports()
  const contextQuery = useCurrentContext()
  const clientsQuery = useDashboardClients()

  const [client, setClient] = useState('')
  const [engagement, setEngagement] = useState('')
  const [phase, setPhase] = useState('')
  const [format, setFormat] = useState<'html' | 'md'>('html')
  const [includeGifs, setIncludeGifs] = useState(false)
  const [gifResolution, setGifResolution] = useState<'720p' | '1080p'>('720p')
  const [outputName, setOutputName] = useState('')
  const [submitError, setSubmitError] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [jobId, setJobId] = useState(() => localStorage.getItem('reportJobId') || '')

  const engagementsQuery = useDashboardEngagements(client)
  const phasesQuery = useDashboardPhases(client, engagement)
  const activeJobQuery = useActiveReportJob()
  const jobQuery = useReportJob(jobId || undefined)
  const currentJob = jobQuery.data?.job

  const reports = data?.reports ?? []
  const htmlReports = reports.filter((report) => report.type === 'html').length
  const clients = clientsQuery.data?.clients ?? []
  const engagements = engagementsQuery.data?.engagements ?? []
  const phases = phasesQuery.data?.phases ?? []

  useEffect(() => {
    if (contextQuery.data?.context && !client) {
      setClient(contextQuery.data.context.client ?? '')
      setEngagement(contextQuery.data.context.engagement ?? '')
      setPhase(contextQuery.data.context.phase ?? '')
    }
  }, [contextQuery.data, client])

  useEffect(() => {
    setEngagement('')
    setPhase('')
  }, [client])

  useEffect(() => {
    setPhase('')
  }, [engagement])

  useEffect(() => {
    if (format !== 'html') {
      setIncludeGifs(false)
    }
  }, [format])

  useEffect(() => {
    if (currentJob?.status === 'completed') {
      void refetch()
      void queryClient.invalidateQueries({ queryKey: ['reports'] })
    }
    if (currentJob?.status === 'completed' || currentJob?.status === 'failed') {
      localStorage.removeItem('reportJobId')
    }
  }, [currentJob?.status, queryClient, refetch])

  useEffect(() => {
    if (!jobId && activeJobQuery.data?.job) {
      const activeJob = activeJobQuery.data.job
      if (activeJob.status === 'queued' || activeJob.status === 'running') {
        setJobId(activeJob.id)
        localStorage.setItem('reportJobId', activeJob.id)
      }
    }
  }, [activeJobQuery.data, jobId])

  useEffect(() => {
    if (jobQuery.isError) {
      localStorage.removeItem('reportJobId')
      setJobId('')
    }
  }, [jobQuery.isError])

  async function handleGenerate(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setSubmitError('')
    setIsSubmitting(true)

    try {
      const response = await api.reports.generate({
        client: client || undefined,
        engagement: engagement || undefined,
        phase: phase || undefined,
        format,
        include_gifs: includeGifs,
        gif_resolution: gifResolution,
        output_name: outputName || undefined,
      })
      setJobId(response.job.id)
      localStorage.setItem('reportJobId', response.job.id)
    } catch (error) {
      setSubmitError(error instanceof Error ? error.message : 'Failed to start report generation')
    } finally {
      setIsSubmitting(false)
    }
  }

  if (isLoading) {
    return <div className="loading-state">Loading reports…</div>
  }

  return (
    <div className="page-stack">
      <section className="page-heading">
        <div>
          <div className="eyebrow">Reports</div>
          <h2>Review and generate engagement artifacts.</h2>
          <p>Generate Markdown or HTML reports directly from the web dashboard.</p>
        </div>
      </section>

      <section className="panel-card">
        <div className="panel-header">
          <div>
            <h3>Generate Report</h3>
            <p>Choose scope and output format. GIF capture is available for HTML exports.</p>
          </div>
        </div>
        <form className="search-form" onSubmit={handleGenerate}>
          <div className="form-grid">
            <label className="field">
              <span>Client</span>
              <select value={client} onChange={(event) => setClient(event.target.value)}>
                <option value="">Select client</option>
                {clients.map((c) => (
                  <option key={c.name} value={c.name}>{c.name}</option>
                ))}
              </select>
            </label>
            <label className="field">
              <span>Engagement (optional)</span>
              <select value={engagement} onChange={(event) => setEngagement(event.target.value)} disabled={!client}>
                <option value="">All engagements</option>
                {engagements.map((e) => (
                  <option key={e.name} value={e.name}>{e.name}</option>
                ))}
              </select>
            </label>
            <label className="field">
              <span>Phase (optional)</span>
              <select value={phase} onChange={(event) => setPhase(event.target.value)} disabled={!client}>
                <option value="">All phases</option>
                {phases.map((p) => (
                  <option key={p.name} value={p.name}>{p.name}</option>
                ))}
              </select>
            </label>
            <label className="field">
              <span>Format</span>
              <select value={format} onChange={(event) => setFormat(event.target.value as 'html' | 'md')}>
                <option value="html">HTML</option>
                <option value="md">Markdown</option>
              </select>
            </label>
            <label className="field field-span-2">
              <span>Output filename (optional)</span>
              <input value={outputName} onChange={(event) => setOutputName(event.target.value)} placeholder="custom_report.html" />
            </label>
            <label className="checkbox-field field-span-2">
              <input type="checkbox" checked={includeGifs} disabled={format !== 'html'} onChange={(event) => setIncludeGifs(event.target.checked)} />
              <span>Generate GIFs and embed them in HTML report</span>
            </label>
            {format === 'html' && includeGifs && (
              <label className="field">
                <span>GIF Resolution</span>
                <select value={gifResolution} onChange={(event) => setGifResolution(event.target.value as '720p' | '1080p')}>
                  <option value="720p">720p</option>
                  <option value="1080p">1080p</option>
                </select>
              </label>
            )}
          </div>
          {submitError && <div className="empty-state compact">{submitError}</div>}
          <div className="row-actions">
            <button className="primary-button" type="submit" disabled={isSubmitting || jobQuery.isFetching}>
              {isSubmitting ? 'Starting…' : 'Generate report'}
            </button>
            {currentJob && (
              <span className={`badge ${currentJob.status === 'completed' ? 'badge-green' : currentJob.status === 'failed' ? 'badge-red' : 'badge-amber'}`}>
                {currentJob.status}
              </span>
            )}
          </div>
        </form>
      </section>

      {currentJob && (
        <section className="panel-card">
          <div className="panel-header">
            <div>
              <h3>Latest Generation Job</h3>
              <p>{currentJob.message}</p>
            </div>
          </div>
          <div className="meta-grid compact-meta-grid">
            <div>
              <span>Job ID</span>
              <strong className="mono-text">{currentJob.id}</strong>
            </div>
            <div>
              <span>Scope</span>
              <strong>
                {currentJob.client}
                {currentJob.engagement ? ` / ${currentJob.engagement}` : ''}
                {currentJob.phase ? ` / ${currentJob.phase}` : ''}
              </strong>
            </div>
            <div>
              <span>Sessions</span>
              <strong>{currentJob.sessions_count}</strong>
            </div>
            <div>
              <span>GIFs</span>
              <strong>
                {currentJob.gif_generated} generated / {currentJob.gif_failed} failed
              </strong>
            </div>
            {currentJob.include_gifs && currentJob.status === 'running' && currentJob.est_time_remaining_secs && currentJob.est_time_remaining_secs > 0 && (
              <div>
                <span>Est. Time Remaining</span>
                <strong>{formatDuration(currentJob.est_time_remaining_secs)}</strong>
              </div>
            )}
            {currentJob.include_gifs && currentJob.avg_time_per_session_secs && currentJob.avg_time_per_session_secs > 0 && (
              <div>
                <span>Avg. per Session</span>
                <strong>{formatDuration(currentJob.avg_time_per_session_secs)}</strong>
              </div>
            )}
            <div>
              <span>Updated</span>
              <strong>{formatDate(currentJob.updated_at)}</strong>
            </div>
            <div>
              <span>Output</span>
              <strong className="mono-text">{currentJob.relative_path || '-'}</strong>
            </div>
          </div>
          {currentJob.error && <div className="empty-state compact">{currentJob.error}</div>}
          {currentJob.status === 'completed' && currentJob.view_url && (
            <div className="row-actions">
              <a className="inline-link" href={currentJob.view_url} target="_blank" rel="noreferrer">
                Open generated report <ExternalLink size={14} />
              </a>
            </div>
          )}
        </section>
      )}

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
