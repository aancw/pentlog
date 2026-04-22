import { useEffect, useMemo, useState } from 'react'
import type { FormEvent } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Search as SearchIcon } from 'lucide-react'
import { useSearchParams } from 'react-router-dom'
import { api } from '../hooks/useApi'
import type { SearchResult } from '../lib/api'

export default function SearchPage() {
  const [searchParams] = useSearchParams()
  const scopedSearch = searchParams.toString()
  const [query, setQuery] = useState(() => searchParams.get('q') ?? '')
  const [regex, setRegex] = useState(() => searchParams.get('regex') === 'true')
  const [from, setFrom] = useState(() => searchParams.get('from') ?? '')
  const [to, setTo] = useState(() => searchParams.get('to') ?? '')
  const [limit, setLimit] = useState(() => {
    const raw = Number(searchParams.get('limit') ?? 100)
    if (!Number.isFinite(raw) || raw <= 0) return 100
    return Math.min(500, raw)
  })
  const [submitted, setSubmitted] = useState<Record<string, string | number | boolean> | null>(null)

  const searchQuery = useQuery({
    queryKey: ['search', submitted],
    queryFn: () => api.search.query(submitted ?? {}),
    enabled: Boolean(submitted),
  })

  const groupedResults = useMemo(() => {
    const groups = new Map<string, SearchResult[]>()
    for (const result of searchQuery.data?.results ?? []) {
      const existing = groups.get(result.session_path) ?? []
      existing.push(result)
      groups.set(result.session_path, existing)
    }
    return Array.from(groups.entries())
  }, [searchQuery.data])

  useEffect(() => {
    const params = new URLSearchParams(scopedSearch)
    const nextQuery = params.get('q') ?? ''
    const nextRegex = params.get('regex') === 'true'
    const nextFrom = params.get('from') ?? ''
    const nextTo = params.get('to') ?? ''
    const rawLimit = Number(params.get('limit') ?? 100)
    const nextLimit = Number.isFinite(rawLimit) && rawLimit > 0 ? Math.min(500, rawLimit) : 100

    setQuery(nextQuery)
    setRegex(nextRegex)
    setFrom(nextFrom)
    setTo(nextTo)
    setLimit(nextLimit)

    if (nextQuery.trim() === '') {
      setSubmitted(null)
      return
    }

    const next: Record<string, string | number | boolean> = {
      q: nextQuery.trim(),
      regex: nextRegex,
      limit: nextLimit,
    }
    if (nextFrom) next.from = nextFrom
    if (nextTo) next.to = nextTo
    setSubmitted(next)
  }, [scopedSearch])

  function handleSubmit(event: FormEvent) {
    event.preventDefault()
    if (!query.trim()) {
      return
    }

    const next: Record<string, string | number | boolean> = {
      q: query.trim(),
      regex,
      limit,
    }
    if (from) next.from = from
    if (to) next.to = to
    setSubmitted(next)
  }

  return (
    <div className="page-stack">
      <section className="page-heading">
        <div>
          <div className="eyebrow">Evidence Search</div>
          <h2>Search sessions and operator notes.</h2>
          <p>Supports boolean expressions, regex mode, and date scoping.</p>
        </div>
      </section>

      <section className="panel-card filter-card">
        <form className="search-form" onSubmit={handleSubmit}>
          <label className="field field-search field-span-2">
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
            <input type="number" min={1} max={500} value={limit} onChange={(event) => setLimit(Number(event.target.value) || 100)} />
          </label>

          <label className="checkbox-field">
            <input type="checkbox" checked={regex} onChange={(event) => setRegex(event.target.checked)} />
            <span>Regex mode</span>
          </label>

          <button className="primary-button" type="submit" disabled={!query.trim() || searchQuery.isFetching}>
            {searchQuery.isFetching ? 'Searching…' : 'Run Search'}
          </button>
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

      {submitted && searchQuery.data && (
        <section className="panel-card">
          <div className="panel-header">
            <div>
              <h3>Results</h3>
              <p>
                {searchQuery.data.total_matches} match{searchQuery.data.total_matches === 1 ? '' : 'es'} for{' '}
                <code>{String(submitted.q)}</code>
              </p>
            </div>
            <span className="pill">{regex ? 'Regex' : 'Boolean'}</span>
          </div>

          <div className="list-stack">
            {groupedResults.map(([sessionPath, items]) => (
              <article key={sessionPath} className="search-result-card">
                <div className="search-result-header">
                  <strong>{sessionPath}</strong>
                  <span className="pill">{items.length} hits</span>
                </div>
                <div className="search-result-body">
                  {items.map((item, index) => (
                    <div key={`${item.line_num}-${index}`} className="search-hit">
                      <div className="search-hit-meta">
                        <span>Line {item.line_num}</span>
                        {item.is_note && <span className="badge badge-blue">Note</span>}
                      </div>
                      <pre>{item.content}</pre>
                    </div>
                  ))}
                </div>
              </article>
            ))}
            {groupedResults.length === 0 && <div className="empty-state compact">No matches found for the current query.</div>}
          </div>
        </section>
      )}
    </div>
  )
}
