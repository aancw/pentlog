import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Search as SearchIcon, FileText } from 'lucide-react'

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
      <div>
        <h1 className="text-3xl font-bold">Search</h1>
        <p className="text-muted-foreground mt-1">
          Search across all sessions and notes
        </p>
      </div>

      <form onSubmit={handleSearch} className="flex gap-4">
        <div className="flex-1 relative">
          <SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Enter search query..."
            className="w-full pl-10 pr-4 py-2 border border-border rounded-lg bg-background text-sm"
          />
        </div>
        <label className="flex items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={isRegex}
            onChange={(e) => setIsRegex(e.target.checked)}
            className="rounded"
          />
          Regex
        </label>
        <button
          type="submit"
          className="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90"
        >
          Search
        </button>
      </form>

      <div className="text-xs text-muted-foreground">
        Use boolean operators: <code className="text-primary">term1 term2</code> (AND), <code className="text-primary">term1 OR term2</code>, <code className="text-primary">-term</code> (NOT)
      </div>

      {isLoading ? (
        <div className="text-center py-12 text-muted-foreground">Searching...</div>
      ) : searched && results.length === 0 ? (
        <div className="text-center py-12">
          <SearchIcon className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
          <p className="text-muted-foreground">No results found for "{searched}"</p>
        </div>
      ) : results.length > 0 ? (
        <div className="space-y-4">
          <p className="text-sm text-muted-foreground">
            {resultsData?.total_matches ?? 0} matches found
          </p>
          {results.map((result, idx) => (
            <div key={idx} className="rounded-lg border border-border bg-card p-4">
              <div className="flex items-center gap-2 mb-2 text-sm">
                <FileText className="h-4 w-4 text-muted-foreground" />
                <span className="font-medium">{result.session_path}</span>
                {result.is_note && (
                  <span className="px-2 py-0.5 text-xs bg-blue-500/10 text-blue-500 rounded">Note</span>
                )}
                <span className="text-muted-foreground">Line {result.line_num}</span>
              </div>
              <pre className="text-sm bg-muted p-3 rounded overflow-x-auto whitespace-pre-wrap">
                {result.content}
              </pre>
            </div>
          ))}
        </div>
      ) : null}
    </div>
  )
}