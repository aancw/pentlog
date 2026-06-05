import { useEffect, useMemo, useState } from 'react'
import { Link, useParams, useSearchParams } from 'react-router-dom'
import { ArrowLeft, Clock3, ExternalLink, HardDrive, Maximize2, Radio, StickyNote, Terminal, X } from 'lucide-react'
import { formatDate } from '../lib/api'
import { useSession, useSessionContent, useSessionNotes, useSessionTimeline, useShareStatus } from '../hooks/useApi'
import { api } from '../hooks/useApi'

type FocusedLine = {
  number: number
  text: string
}

export default function SessionDetail() {
  const { id } = useParams()
  const [searchParams] = useSearchParams()
  const [noteDraft, setNoteDraft] = useState('')
  const [showFullContent, setShowFullContent] = useState(false)
  const [showFullTimeline, setShowFullTimeline] = useState(false)
  const sessionQuery = useSession(id)
  const notesQuery = useSessionNotes(id)
  const timelineQuery = useSessionTimeline(id)
  const contentQuery = useSessionContent(id)
  const shareQuery = useShareStatus()

  const matchType = searchParams.get('matchType') ?? ''
  const matchText = searchParams.get('matchText') ?? ''
  const noteTimestamp = searchParams.get('noteTs') ?? ''
  const matchLine = Number(searchParams.get('matchLine') ?? 0)

  useEffect(() => {
    function handleEscape(event: KeyboardEvent) {
      if (event.key === 'Escape') {
        setShowFullContent(false)
        setShowFullTimeline(false)
      }
    }

    window.addEventListener('keydown', handleEscape)
    return () => window.removeEventListener('keydown', handleEscape)
  }, [])

  async function handleAddNote() {
    if (!id || !noteDraft.trim()) {
      return
    }
    await api.sessions.addNote(id, noteDraft.trim())
    setNoteDraft('')
    await notesQuery.refetch()
    await sessionQuery.refetch()
  }

  if (sessionQuery.isLoading) {
    return <div className="loading-state">Loading session…</div>
  }

  const session = sessionQuery.data
  if (!session) {
    return <div className="empty-state">Session not found.</div>
  }

  const notes = notesQuery.data?.notes ?? []
  const commands = timelineQuery.data?.commands ?? []
  const fullContent = contentQuery.data?.content ?? ''
  const contentPreview = contentQuery.data?.content?.slice(0, 2400) ?? ''
  const share = shareQuery.data
  const isLiveShared = Boolean(share?.active && share.log_file && share.log_file === session.path)

  const focusedContentLines = useMemo<FocusedLine[]>(() => {
    if (matchType === 'note' || matchLine <= 0 || fullContent === '') {
      return []
    }

    const lines = fullContent.split('\n')
    const start = Math.max(0, matchLine - 3)
    const end = Math.min(lines.length, matchLine + 2)

    return lines.slice(start, end).map((line, index) => ({
      number: start + index + 1,
      text: line,
    }))
  }, [fullContent, matchLine, matchType])

  const highlightedNoteIndex = useMemo(() => {
    if (notes.length === 0 || matchType !== 'note') {
      return -1
    }

    return notes.findIndex((note) => {
      if (noteTimestamp && note.timestamp === noteTimestamp) {
        return true
      }
      return matchText !== '' && note.content === matchText
    })
  }, [matchText, matchType, noteTimestamp, notes])

  const hasSearchFocus = focusedContentLines.length > 0 || highlightedNoteIndex >= 0

  return (
    <div className="page-stack">
      <Link to="/sessions" className="inline-link back-link">
        <ArrowLeft size={14} /> Back to sessions
      </Link>

      <section className="page-heading">
        <div>
          <div className="eyebrow">Session Detail</div>
          <h2>Session #{session.id}</h2>
          <p>{session.display_path || session.filename}</p>
        </div>
        <div className="page-heading-meta">
          <span className={`badge badge-${session.state.toLowerCase()}`}>{session.state}</span>
        </div>
      </section>

      {hasSearchFocus && (
        <section className="panel-card">
          <div className="panel-header">
            <div>
              <h3>Search Focus</h3>
              <p>{highlightedNoteIndex >= 0 ? 'This page opened from a matching operator note.' : 'This page opened from a matching transcript line.'}</p>
            </div>
          </div>
          {highlightedNoteIndex >= 0 ? (
            <div className="list-row list-row-highlight stacked-row">
              <strong>{formatDate(notes[highlightedNoteIndex]?.timestamp)}</strong>
              <p>{notes[highlightedNoteIndex]?.content}</p>
            </div>
          ) : (
            <div className="terminal-preview content-lines">
              {focusedContentLines.map((line) => (
                <div key={line.number} className={`content-line ${line.number === matchLine ? 'content-line-highlight' : ''}`}>
                  <span className="content-line-number">{line.number}</span>
                  <span className="content-line-text">{line.text || ' '}</span>
                </div>
              ))}
            </div>
          )}
        </section>
      )}

      <section className="stats-grid">
        <article className="stat-card stat-card-blue">
          <div className="stat-card-header">
            <span>Size</span>
            <HardDrive size={18} />
          </div>
          <div className="stat-card-value">{session.size_human}</div>
        </article>
        <article className="stat-card stat-card-sand">
          <div className="stat-card-header">
            <span>Notes</span>
            <StickyNote size={18} />
          </div>
          <div className="stat-card-value">{session.notes_count}</div>
        </article>
        <article className="stat-card stat-card-amber">
          <div className="stat-card-header">
            <span>Commands</span>
            <Terminal size={18} />
          </div>
          <div className="stat-card-value">{commands.length}</div>
        </article>
        <article className="stat-card stat-card-green">
          <div className="stat-card-header">
            <span>Recorded</span>
            <Clock3 size={18} />
          </div>
          <div className="stat-card-value text-compact">{formatDate(session.mod_time)}</div>
        </article>
      </section>

      <section className="panel-grid two-columns">
        <article className="panel-card">
          <div className="panel-header">
            <div>
              <h3>Metadata</h3>
              <p>Context captured with this recording.</p>
            </div>
          </div>
          <div className="meta-grid">
            <div>
              <span>Client</span>
              <strong>{session.metadata.client || '-'}</strong>
            </div>
            <div>
              <span>Engagement</span>
              <strong>{session.metadata.engagement || '-'}</strong>
            </div>
            <div>
              <span>Phase</span>
              <strong>{session.metadata.phase || '-'}</strong>
            </div>
            <div>
              <span>Operator</span>
              <strong>{session.metadata.operator || '-'}</strong>
            </div>
            <div>
              <span>Target</span>
              <strong>{session.metadata.target || '-'}</strong>
            </div>
            <div>
              <span>Target IP</span>
              <strong>{session.metadata.target_ip || '-'}</strong>
            </div>
          </div>
        </article>

        <article className="panel-card">
          <div className="panel-header">
            <div>
              <h3>Content Preview</h3>
              <p>{focusedContentLines.length > 0 ? 'Transcript excerpt centered on the matching line.' : 'Cleaned terminal output excerpt.'}</p>
            </div>
            <button className="secondary-button" onClick={() => setShowFullContent(true)} disabled={!fullContent}>
              <Maximize2 size={14} /> View full
            </button>
          </div>
          {contentQuery.isLoading ? (
            <div className="loading-state compact">Loading content…</div>
          ) : focusedContentLines.length > 0 ? (
            <div className="terminal-preview content-lines">
              {focusedContentLines.map((line) => (
                <div key={line.number} className={`content-line ${line.number === matchLine ? 'content-line-highlight' : ''}`}>
                  <span className="content-line-number">{line.number}</span>
                  <span className="content-line-text">{line.text || ' '}</span>
                </div>
              ))}
            </div>
          ) : (
            <pre className="terminal-preview">{contentPreview || 'No content available.'}</pre>
          )}
        </article>
      </section>

      <section className="panel-card">
        <div className="panel-header">
          <div>
            <h3>Replay & Share</h3>
            <p>Session replay already exists outside the dashboard; live share status is surfaced here.</p>
          </div>
          {isLiveShared && share?.watch_url ? (
            <a className="inline-link" href={share.watch_url} target="_blank" rel="noreferrer">
              Open watch <ExternalLink size={14} />
            </a>
          ) : null}
        </div>
        {isLiveShared ? (
          <div className="list-row">
            <div>
              <strong>This session is currently being shared live.</strong>
              <div className="subdued-text mono-text">{share?.log_file}</div>
            </div>
            <div className="list-row-meta">
              <span className={`badge ${share?.reachable ? 'badge-green' : 'badge-amber'}`}>{share?.reachable ? 'Online' : 'Pending'}</span>
              <span className="pill">
                <Radio size={14} /> {share?.clients ?? 0} viewers
              </span>
            </div>
          </div>
        ) : (
          <div className="empty-state compact">No active live share bound to this session.</div>
        )}
      </section>

      <section className="panel-grid two-columns">
        <article className="panel-card">
          <div className="panel-header">
            <div>
              <h3>Notes</h3>
              <p>Operator annotations for this session.</p>
            </div>
          </div>
          <div className="note-composer">
            <textarea value={noteDraft} onChange={(event) => setNoteDraft(event.target.value)} placeholder="Add note…" rows={4} />
            <button className="primary-button" onClick={handleAddNote} disabled={!noteDraft.trim()}>
              Save note
            </button>
          </div>
          <div className="list-stack">
            {notes.map((note, index) => (
              <div key={`${note.timestamp}-${index}`} className={`list-row stacked-row ${index === highlightedNoteIndex ? 'list-row-highlight' : ''}`}>
                <strong>{formatDate(note.timestamp)}</strong>
                <p>{note.content}</p>
              </div>
            ))}
            {notes.length === 0 && <div className="empty-state compact">No notes for this session.</div>}
          </div>
        </article>

        <article className="panel-card">
          <div className="panel-header">
            <div>
              <h3>Timeline</h3>
              <p>Parsed commands from the ttyrec stream.</p>
            </div>
            <button className="secondary-button" onClick={() => setShowFullTimeline(true)} disabled={commands.length === 0}>
              <Maximize2 size={14} /> View full
            </button>
          </div>
          <div className="list-stack timeline-stack">
            {commands.slice(0, 20).map((entry, index) => (
              <div key={`${entry.timestamp}-${index}`} className="timeline-entry">
                <div className="timeline-entry-header">
                  <span className="pill">{formatDate(entry.timestamp)}</span>
                  <code>{entry.command}</code>
                </div>
                {entry.output && <pre>{entry.output.slice(0, 320)}</pre>}
              </div>
            ))}
            {commands.length === 0 && <div className="empty-state compact">No commands extracted yet.</div>}
          </div>
        </article>
      </section>

      <section className="panel-card">
        <div className="panel-header">
          <div>
            <h3>Filesystem Path</h3>
            <p>Resolved session log location.</p>
          </div>
        </div>
        <div className="code-block">{session.path}</div>
      </section>

      {showFullContent && (
        <div className="modal-backdrop" onClick={() => setShowFullContent(false)}>
          <section className="modal-card" onClick={(event) => event.stopPropagation()}>
            <div className="modal-header">
              <h3>Full Content Preview</h3>
              <button className="icon-button" onClick={() => setShowFullContent(false)} aria-label="Close full content preview">
                <X size={16} />
              </button>
            </div>
            <div className="modal-body">
              <pre className="terminal-preview modal-terminal">{fullContent || 'No content available.'}</pre>
            </div>
          </section>
        </div>
      )}

      {showFullTimeline && (
        <div className="modal-backdrop" onClick={() => setShowFullTimeline(false)}>
          <section className="modal-card" onClick={(event) => event.stopPropagation()}>
            <div className="modal-header">
              <h3>Full Timeline</h3>
              <button className="icon-button" onClick={() => setShowFullTimeline(false)} aria-label="Close full timeline">
                <X size={16} />
              </button>
            </div>
            <div className="modal-body list-stack timeline-stack">
              {commands.map((entry, index) => (
                <div key={`${entry.timestamp}-${index}`} className="timeline-entry">
                  <div className="timeline-entry-header">
                    <span className="pill">{formatDate(entry.timestamp)}</span>
                    <code>{entry.command}</code>
                  </div>
                  {entry.output && <pre>{entry.output}</pre>}
                </div>
              ))}
            </div>
          </section>
        </div>
      )}
    </div>
  )
}
