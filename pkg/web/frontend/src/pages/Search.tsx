import { Search as SearchIcon } from 'lucide-react'

export default function Search() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Search</h1>
        <p className="text-muted-foreground mt-1">
          Search across all sessions and notes
        </p>
      </div>

      <div className="rounded-lg border border-border bg-card p-12">
        <div className="text-center">
          <SearchIcon className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
          <h2 className="text-xl font-semibold mb-2">Full-Text Search</h2>
          <p className="text-muted-foreground max-w-md mx-auto">
            Search through all recorded sessions and notes using boolean operators 
            (AND, OR, NOT) and regular expressions.
          </p>
          <div className="max-w-md mx-auto mt-6">
            <div className="relative">
              <SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <input
                type="text"
                placeholder="Enter search query..."
                className="w-full pl-10 pr-4 py-2 border border-border rounded-lg bg-background text-sm"
              />
            </div>
            <p className="text-xs text-muted-foreground mt-2 text-left">
              Example: <code className="text-primary">nmap AND ssh</code> or <code className="text-primary">exploit -http</code>
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}