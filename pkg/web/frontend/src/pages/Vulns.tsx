import { Shield } from 'lucide-react'

export default function Vulns() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Vulnerabilities</h1>
        <p className="text-muted-foreground mt-1">
          Manage and track security findings
        </p>
      </div>

      <div className="rounded-lg border border-border bg-card p-12">
        <div className="text-center">
          <Shield className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
          <h2 className="text-xl font-semibold mb-2">Vulnerability Management</h2>
          <p className="text-muted-foreground max-w-md mx-auto">
            This page will display and manage security findings from your penetration testing sessions.
            Vulnerabilities can be created, categorized by severity, and tracked through remediation.
          </p>
          <div className="flex items-center justify-center gap-6 mt-6">
            <div className="flex items-center gap-2">
              <div className="w-3 h-3 rounded-full bg-red-500" />
              <span className="text-sm">Critical</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="w-3 h-3 rounded-full bg-orange-500" />
              <span className="text-sm">High</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="w-3 h-3 rounded-full bg-yellow-500" />
              <span className="text-sm">Medium</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="w-3 h-3 rounded-full bg-green-500" />
              <span className="text-sm">Low</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}