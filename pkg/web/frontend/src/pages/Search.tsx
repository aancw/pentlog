import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Search as SearchIcon, Terminal, Loader2 } from 'lucide-react'

interface SearchResult {
  session_id: number
  session_path: string
  line_num: number
  content: string
  is_note: boolean
}

export default function Search() {
  const [query, setQuery] = useState('')
  const [searched, setSearched] = useState('')
  const [isRegex, setIsRegex] = useState(false)

  const { data: resultsData, isLoading } = useQuery({
    queryKey: ['search', searched, isRegex],
    queryFn: async (): Promise<{ results: SearchResult[], total_matches: number }> => {
      const params = new URLSearchParams()
      params.set('q', searched)
      if (isRegex) params.set('regex', 'true')
      const res = await fetch(`/api/search?${params.toString()}`)
      return res.json()
    },
    enabled: searched !== '',
  })

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    if (query.trim()) {
      setSearched(query.trim())
    }
  }

  const results = resultsData?.results ?? []

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold text-[hsl(var(--foreground))]">Search</h1>
        <p className="text-[hsl(var(--muted-foreground))] mt-1">
          Search across all sessions and notes
        </p>
      </div>

      {/* Search Form */}
      <form onSubmit={handleSearch} className="space-y-4">
        <div className="flex flex-col sm:flex-row gap-4">
          <div className="relative flex-1">
            <SearchIcon className="absolute left-4 top-1/2 -translate-y-1/2 h-5 w-5 text-[hsl(var(--muted-foreground))]" />
            <input
              type="text"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Enter search query..."
              className="w-full pl-12 pr-4 py-3 bg-[hsl(var(--card))] border border-[hsl(var(--border))] rounded-xl text-[hsl(var(--foreground))] placeholder:text-[hsl(var(--muted-foreground))] focus:outline-none focus:ring-2 focus:ring-[hsl(var(--primary))] focus:border-transparent"
            />
          </div>
          <button
            type="submit"
            disabled={!query.trim() || isLoading}
            className="px-6 py-3 bg-[hsl(var(--primary))] text-[hsl(var(--primary-foreground))] rounded-xl hover:opacity-90 disabled:opacity-50 disabled:cursor-not-allowed transition-opacity font-medium flex items-center gap-2"
          >
            {isLoading ? (
              <>
                <Loader2 className="h-4 w-4 animate-spin" />
                Searching...
              </>
            ) : (
              'Search'
            )}
          </button>
        </div>

        <div className="flex items-center gap-6">
          <label className="flex items-center gap-2 cursor-pointer">
            <input
              type="checkbox"
              checked={isRegex}
              onChange={(e) => setIsRegex(e.target.checked)}
              className="w-4 h-4 rounded border-[hsl(var(--border))] text-[hsl(var(--primary))] focus:ring-[hsl(var(--primary))]"
            />
            <span className="text-sm text-[hsl(var(--muted-foreground))]">Regex mode</span>
          </label>
        </div>
      </form>

      {/* Search Tips */}
      <div className="rounded-lg bg-[hsl(var(--secondary))] p-4 text-sm text-[hsl(var(--muted-foreground))]">
        <span className="font-medium">Tips:</span> Use <code className="px-1.5 py-0.5 bg-[hsl(var(--card))] rounded text-[hsl(var(--primary))]">term1 term2</code> for AND, 
        <code className="px-1.5 py-0.5 bg-[hsl(var(--card))] rounded text-[hsl(var(--primary))]">term1 OR term2</code> for OR, 
        <code className="px-1.5 py-0.5 bg-[hsl(var(--card))] rounded text-[hsl(var(--primary))]">-term</code> for NOT
      </div>

      {/* Results */}
      {searched && !isLoading && results.length === 0 ? (
        <div className="rounded-xl border border-[hsl(var(--border))] bg-[hsl(var(--card))] p-12">
          <div className="text-center">
            <SearchIcon className="h-12 w-12 mx-auto text-[hsl(var(--muted-foreground))] mb-4" />
            <h2 className="text-xl font-semibold text-[hsl(var(--foreground))] mb-2">No Results</h2>
            <p className="text-[hsl(var(--muted-foreground))]">
              No matches found for "{searched}"
            </p>
          </div>
        </div>
      ) : results.length > 0 ? (
        <div className="space-y-4">
          <div className="flex items-center gap-2 text-sm text-[hsl(var(--muted-foreground))]">
            <span className="font-medium text-[hsl(var(--foreground))]">{resultsData?.total_matches ?? 0}</span>
            matches found
          </div>
          
          {results.map((result, idx) => (
            <div key={idx} className="rounded-xl border border-[hsl(var(--border))] bg-[hsl(var(--card))] overflow-hidden">
              <div className="flex items-center gap-3 px-5 py-3 bg-[hsl(var(--secondary))] border-b border-[hsl(var(--border))]">
                <Terminal className="h-4 w-4 text-[hsl(var(--primary))]" />
                <span className="font-medium text-sm text-[hsl(var(--foreground))]">{result.session_path}</span>
                {result.is_note && (
                  <span className="px-2 py-0.5 text-xs bg-blue-500/10 text-blue-500 rounded">Note</span>
                )}
                <span className="text-xs text-[hsl(var(--muted-foreground))]">Line {result.line_num}</span>
              </div>
              <pre className="p-4 text-sm text-[hsl(var(--foreground))] overflow-x-auto whitespace-pre-wrap font-mono">
                {result.content}
              </pre>
            </div>
          ))}
        </div>
      ) : null}
    </div>
  )
}