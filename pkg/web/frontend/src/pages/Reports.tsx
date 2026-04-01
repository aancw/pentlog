import { FileText, Download, Sparkles } from 'lucide-react'

export default function Reports() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Reports</h1>
        <p className="text-muted-foreground mt-1">
          Generate and manage engagement reports
        </p>
      </div>

      <div className="rounded-lg border border-border bg-card p-12">
        <div className="text-center">
          <FileText className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
          <h2 className="text-xl font-semibold mb-2">Report Generation</h2>
          <p className="text-muted-foreground max-w-md mx-auto">
            Generate Markdown or HTML reports from your session data. 
            Include AI-powered analysis for executive summaries and technical findings.
          </p>
          <div className="flex items-center justify-center gap-4 mt-6">
            <div className="flex items-center gap-2 px-4 py-2 border border-border rounded-lg">
              <Download className="h-4 w-4" />
              <span className="text-sm">Markdown</span>
            </div>
            <div className="flex items-center gap-2 px-4 py-2 border border-border rounded-lg">
              <Download className="h-4 w-4" />
              <span className="text-sm">HTML</span>
            </div>
            <div className="flex items-center gap-2 px-4 py-2 border border-border rounded-lg">
              <Sparkles className="h-4 w-4" />
              <span className="text-sm">AI Analysis</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}