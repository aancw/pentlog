import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Search, Terminal } from 'lucide-react'

export default function SearchPage() {
  const [query, setQuery] = useState('')
  const [searched, setSearched] = useState('')

  const { data, isLoading } = useQuery({
    queryKey: ['search', searched],
    queryFn: async () => {
      const res = await fetch(`/api/search?q=${encodeURIComponent(searched)}`)
      return res.json()
    },
    enabled: !!searched,
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (query.trim()) setSearched(query.trim())
  }

  return (
    <div className="space-y-6">
      <div className="page-header">
        <h1>Search</h1>
        <p className="text-muted">Search across all sessions and notes</p>
      </div>

      <form onSubmit={handleSubmit} className="flex gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-4 top-1/2 -translate-y-1/2 h-5 w-5 text-muted" />
          <input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Enter search query..."
            className="input"
            style={{ paddingLeft: '3rem' }}
          />
        </div>
        <button type="submit" className="btn" disabled={!query.trim() || isLoading}>
          {isLoading ? 'Searching...' : 'Search'}
        </button>
      </form>

      <div className="p-4 bg-secondary rounded text-sm text-muted">
        Use <code className="text-primary">term1 term2</code> (AND), 
        <code className="text-primary"> term1 OR term2</code>, 
        <code className="text-primary"> -term</code> (NOT)
      </div>

      {searched && !isLoading && data?.results?.length === 0 && (
        <div className="card text-center p-12">
          <Search className="h-12 w-12 mx-auto text-muted mb-4" />
          <h2 className="text-xl font-semibold mb-2">No Results</h2>
          <p className="text-muted">No matches for "{searched}"</p>
        </div>
      )}

      {data?.results?.length > 0 && (
        <div className="space-y-4">
          <div className="text-sm text-muted">
            {data.total_matches} matches found
          </div>
          {data.results.map((result: any, idx: number) => (
            <div key={idx} className="card">
              <div className="flex items-center gap-3 p-3 bg-secondary rounded-t border-b border-border">
                <Terminal className="h-4 w-4 text-primary" />
                <span className="font-medium text-sm">{result.session_path}</span>
                {result.is_note && <span className="badge badge-blue">Note</span>}
                <span className="text-xs text-muted">Line {result.line_num}</span>
              </div>
              <pre className="p-4 text-sm overflow-x-auto font-mono">
                {result.content}
              </pre>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}