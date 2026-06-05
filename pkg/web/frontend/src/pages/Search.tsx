import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { Search as SearchIcon } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { api } from '../hooks/useApi'
import type { SearchResult } from '../lib/api'

const DEFAULT_LIMIT = 100
const MAX_LIMIT = 500

type GroupedResult = {
  sessionId: number
  sessionPath: string
  items: SearchResult[]
}

function normalizeLimit(value: string | null) {
  const raw = Number(value ?? DEFAULT_LIMIT)
  if (!Number.isFinite(raw) || raw <= 0) return DEFAULT_LIMIT
  return Math.min(MAX_LIMIT, raw)
}

function normalizePage(value: string | null) {
  const raw = Number(value ?? 1)
  if (!Number.isFinite(raw) || raw <= 0) return 1
  return Math.floor(raw)
}

export default function SearchPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const scopedSearch = searchParams.toString()

  const [query, setQuery] = useState(() => searchParams.get('q') ?? '')
  const [regex, setRegex] = useState(() => searchParams.get('regex') === 'true')
  const [from, setFrom] = useState(() => searchParams.get('from') ?? '')
  const [to, setTo] = useState(() => searchParams.get('to') ?? '')
  const [limit, setLimit] = useState(() => normalizeLimit(searchParams.get('limit')))

  const requestParams = useMemo(() => {
    const nextQuery = (searchParams.get('q') ?? '').trim()
    if (nextQuery === '') {
      return null
    }

    const nextRegex = searchParams.get('regex') === 'true'
    const nextFrom = searchParams.get('from') ?? ''
    const nextTo = searchParams.get('to') ?? ''
    const nextLimit = normalizeLimit(searchParams.get('limit'))
    const nextPage = normalizePage(searchParams.get('page'))

    const next: Record<string, string | number | boolean> = {
      q: nextQuery,
      regex: nextRegex,
      limit: nextLimit,
      offset: (nextPage - 1) * nextLimit,
    }

    if (nextFrom) next.from = nextFrom
    if (nextTo) next.to = nextTo

    return next
  }, [searchParams])

  const currentPage = useMemo(() => normalizePage(searchParams.get('page')), [searchParams])

  const searchQuery = useQuery({
    queryKey: ['search', requestParams],
    queryFn: () => api.search.query(requestParams ?? {}),
    enabled: Boolean(requestParams),
    placeholderData: keepPreviousData,
  })

  const groupedResults = useMemo<GroupedResult[]>(() => {
    const groups = new Map<number, GroupedResult>()
    for (const result of searchQuery.data?.results ?? []) {
      const existing = groups.get(result.session_id)
      if (existing) {
        existing.items.push(result)
        continue
      }

      groups.set(result.session_id, {
        sessionId: result.session_id,
        sessionPath: result.session_path,
        items: [result],
      })
    }

    return Array.from(groups.values())
  }, [searchQuery.data])

  useEffect(() => {
    const params = new URLSearchParams(scopedSearch)
    setQuery(params.get('q') ?? '')
    setRegex(params.get('regex') === 'true')
    setFrom(params.get('from') ?? '')
    setTo(params.get('to') ?? '')
    setLimit(normalizeLimit(params.get('limit')))
  }, [scopedSearch])

  function updateParams(nextPage: number) {
    const params = new URLSearchParams()

    if (query.trim()) params.set('q', query.trim())
    if (regex) params.set('regex', 'true')
    if (from) params.set('from', from)
    if (to) params.set('to', to)
    params.set('limit', String(Math.min(MAX_LIMIT, Math.max(1, limit || DEFAULT_LIMIT))))
    params.set('page', String(Math.max(1, nextPage)))

    setSearchParams(params)
  }

  function handleSubmit(event: FormEvent) {
    event.preventDefault()
    if (!query.trim()) {
      return
    }

    updateParams(1)
  }

  function handleReset() {
    setQuery('')
    setRegex(false)
    setFrom('')
    setTo('')
    setLimit(DEFAULT_LIMIT)
    setSearchParams(new URLSearchParams())
  }

  const totalMatches = searchQuery.data?.total_matches ?? 0
  const showingFrom = totalMatches === 0 ? 0 : (searchQuery.data?.offset ?? 0) + 1
  const showingTo = Math.min(totalMatches, (searchQuery.data?.offset ?? 0) + (searchQuery.data?.results.length ?? 0))
  const totalPages = requestParams ? Math.max(1, Math.ceil(totalMatches / Number(requestParams.limit))) : 1

  return (
    <div className="page-stack">
      <section className="page-heading">
        <div>
          <div className="eyebrow">Evidence Search</div>
          <h2>Search sessions and operator notes.</h2>
          <p>Supports boolean expressions, regex mode, date scoping, and direct drill-down into matching sessions.</p>
        </div>
      </section>

      <section className="panel-card filter-card">
        <form className="search-form" onSubmit={handleSubmit}>
          <label className="field field-span-2">
            <span>Query</span>
            <div className="search-input-shell">
              <SearchIcon size={16} />
              <input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="admin OR token -cleanup" />
            </div>
          </label>

          <label className="field">
            <span>From</span>
            <input type="date" value={from} onChange={(event) => setFrom(event.target.value)} />
          </label>

          <label className="field">
            <span>To</span>
            <input type="date" value={to} onChange={(event) => setTo(event.target.value)} />
          </label>

          <label className="field">
            <span>Limit</span>
            <input type="number" min={1} max={MAX_LIMIT} value={limit} onChange={(event) => setLimit(Number(event.target.value) || DEFAULT_LIMIT)} />
          </label>

          <label className="checkbox-field">
            <input type="checkbox" checked={regex} onChange={(event) => setRegex(event.target.checked)} />
            <span>Regex mode</span>
          </label>

          <div className="row-actions">
            <button className="primary-button" type="submit" disabled={!query.trim() || searchQuery.isFetching}>
              {searchQuery.isFetching ? 'Searching…' : 'Run Search'}
            </button>
            <button className="secondary-button" type="button" onClick={handleReset}>
              Clear
            </button>
          </div>
        </form>
      </section>

      <section className="panel-card">
        <div className="panel-header">
          <div>
            <h3>Query Guide</h3>
            <p>
              Use <code>term1 term2</code> for AND, <code>term1 OR term2</code> for OR, <code>-term</code> for NOT.
            </p>
          </div>
        </div>
      </section>

      {requestParams && searchQuery.data && (
        <section className="panel-card">
          <div className="panel-header">
            <div>
              <h3>Results</h3>
              <p>
                Showing {showingFrom}-{showingTo} of {totalMatches} match{totalMatches === 1 ? '' : 'es'} for{' '}
                <code>{String(requestParams.q)}</code>
              </p>
            </div>
            <span className="pill">{regex ? 'Regex' : 'Boolean'}</span>
          </div>

          <div className="list-stack">
            {groupedResults.map((group) => (
              <article key={group.sessionId} className="search-result-card">
                <div className="search-result-header">
                  <strong>{group.sessionPath}</strong>
                  <div className="row-actions">
                    <span className="pill">{group.items.length} hits</span>
                    <Link className="inline-link" to={`/sessions/${group.sessionId}`}>
                      Open session
                    </Link>
                  </div>
                </div>
                <div className="search-result-body">
                  {group.items.map((item, index) => {
                    const itemParams = new URLSearchParams()
                    itemParams.set('matchType', item.is_note ? 'note' : 'content')
                    itemParams.set('matchText', item.content)
                    if (item.is_note && item.note_timestamp) {
                      itemParams.set('noteTs', item.note_timestamp)
                    }
                    if (!item.is_note && item.line_num > 0) {
                      itemParams.set('matchLine', String(item.line_num))
                    }

                    return (
                      <div key={`${item.session_id}-${item.line_num}-${index}`} className="search-hit">
                        <div className="search-hit-meta">
                          <span>{item.is_note ? 'Operator note' : `Line ${item.line_num}`}</span>
                          {item.is_note && item.note_timestamp && <span className="pill">{item.note_timestamp}</span>}
                          {item.is_note && <span className="badge badge-blue">Note</span>}
                        </div>

                        {!item.is_note && item.context.length > 0 ? (
                          <div className="terminal-preview content-lines">
                            {item.context.map((line, contextIndex) => {
                              const lineNumber = item.context_start_line + contextIndex
                              const isHighlightedLine = lineNumber === item.line_num

                              return (
                                <div key={`${item.session_id}-${lineNumber}-${contextIndex}`} className={`content-line ${isHighlightedLine ? 'content-line-highlight' : ''}`}>
                                  <span className="content-line-number">{lineNumber}</span>
                                  <span className="content-line-text">{line || ' '}</span>
                                </div>
                              )
                            })}
                          </div>
                        ) : (
                          <pre>{item.content}</pre>
                        )}

                        <div className="search-hit-actions">
                          <span className="subdued-text">{item.is_note ? 'Open the session note list with the matching note highlighted.' : 'Open the session transcript centered on this hit.'}</span>
                          <Link className="inline-link" to={`/sessions/${item.session_id}?${itemParams.toString()}`}>
                            Open hit
                          </Link>
                        </div>
                      </div>
                    )
                  })}
                </div>
              </article>
            ))}

            {groupedResults.length === 0 && <div className="empty-state compact">No matches found for the current query.</div>}
          </div>

          {groupedResults.length > 0 && (
            <div className="pagination-bar">
              <div className="subdued-text">
                Page {currentPage} of {totalPages}
              </div>
              <div className="row-actions">
                <button className="secondary-button" onClick={() => updateParams(currentPage - 1)} disabled={currentPage <= 1 || searchQuery.isFetching}>
                  Previous
                </button>
                <button className="secondary-button" onClick={() => updateParams(currentPage + 1)} disabled={!searchQuery.data.has_more || searchQuery.isFetching}>
                  Next
                </button>
              </div>
            </div>
          )}
        </section>
      )}
    </div>
  )
}
