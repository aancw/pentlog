import { Link, useLocation } from 'react-router-dom'
import { useSystemStatus } from '../hooks/useApi'
import { 
  LayoutDashboard, 
  FileText, 
  Search, 
  Archive, 
  Settings,
  Shield,
  FolderOpen,
  Terminal
} from 'lucide-react'

interface LayoutProps {
  children: React.ReactNode
}

const navItems = [
  { path: '/', label: 'Dashboard', icon: LayoutDashboard },
  { path: '/sessions', label: 'Sessions', icon: FolderOpen },
  { path: '/vulns', label: 'Vulnerabilities', icon: Shield },
  { path: '/reports', label: 'Reports', icon: FileText },
  { path: '/search', label: 'Search', icon: Search },
  { path: '/archives', label: 'Archives', icon: Archive },
  { path: '/settings', label: 'Settings', icon: Settings },
]

export default function Layout({ children }: LayoutProps) {
  const location = useLocation()
  const { data: status } = useSystemStatus()

  return (
    <div className="flex h-screen bg-[hsl(var(--background))]">
      {/* Sidebar */}
      <aside className="w-64 border-r border-[hsl(var(--border))] bg-[hsl(var(--card))] flex flex-col">
        {/* Logo */}
        <div className="h-16 flex items-center px-6 border-b border-[hsl(var(--border))]">
          <Terminal className="h-6 w-6 text-[hsl(var(--primary))] mr-2" />
          <span className="text-lg font-bold text-[hsl(var(--foreground))]">PentLog</span>
          <span className="ml-2 text-xs px-1.5 py-0.5 rounded bg-[hsl(var(--primary))] text-[hsl(var(--primary-foreground))]">v{status?.version || '0.16'}</span>
        </div>

        {/* Context Badge */}
        {status?.has_context && status.context && (
          <div className="mx-3 mt-3 p-3 rounded-lg bg-[hsl(var(--secondary))]">
            <div className="text-xs text-[hsl(var(--muted-foreground))] mb-1">Active Context</div>
            <div className="font-medium text-sm text-[hsl(var(--foreground))] truncate">{status.context.client}</div>
            <div className="text-xs text-[hsl(var(--muted-foreground))] truncate">{status.context.engagement} / {status.context.phase}</div>
          </div>
        )}

        {/* Navigation */}
        <nav className="flex-1 p-3 overflow-y-auto">
          <ul className="space-y-1">
            {navItems.map((item) => {
              const Icon = item.icon
              const isActive = location.pathname === item.path
              return (
                <li key={item.path}>
                  <Link
                    to={item.path}
                    className={`flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-all duration-200 ${
                      isActive
                        ? 'bg-[hsl(var(--primary))] text-[hsl(var(--primary-foreground))] shadow-lg shadow-[hsl(var(--primary))]/20'
                        : 'text-[hsl(var(--muted-foreground))] hover:bg-[hsl(var(--accent))] hover:text-[hsl(var(--accent-foreground))]'
                    }`}
                  >
                    <Icon className="h-4 w-4" />
                    {item.label}
                  </Link>
                </li>
              )
            })}
          </ul>
        </nav>

        {/* Footer */}
        <div className="p-4 border-t border-[hsl(var(--border))]">
          <div className="text-xs text-[hsl(var(--muted-foreground))]">
            {status?.total_sessions ?? 0} sessions recorded
          </div>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 overflow-auto bg-[hsl(var(--background))]">
        <div className="max-w-7xl mx-auto p-8">
          {children}
        </div>
      </main>
    </div>
  )
}