import { Archive as ArchiveIcon, Lock, Download } from 'lucide-react'

export default function Archives() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Archives</h1>
        <p className="text-muted-foreground mt-1">
          Create and manage session archives
        </p>
      </div>

      <div className="rounded-lg border border-border bg-card p-12">
        <div className="text-center">
          <ArchiveIcon className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
          <h2 className="text-xl font-semibold mb-2">Archive Management</h2>
          <p className="text-muted-foreground max-w-md mx-auto">
            Create encrypted ZIP archives of your session data for backup or sharing.
            Archives can be password-protected with AES-256 encryption.
          </p>
          <div className="flex items-center justify-center gap-4 mt-6">
            <div className="flex items-center gap-2 px-4 py-2 border border-border rounded-lg">
              <Download className="h-4 w-4" />
              <span className="text-sm">Export</span>
            </div>
            <div className="flex items-center gap-2 px-4 py-2 border border-border rounded-lg">
              <Lock className="h-4 w-4" />
              <span className="text-sm">Encrypt</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}