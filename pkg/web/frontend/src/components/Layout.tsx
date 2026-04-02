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
    <div className="flex h-screen">
      <aside className="sidebar">
        <div className="sidebar-header">
          <Terminal className="h-6 w-6 text-primary mr-2" />
          <span className="text-lg font-bold">PentLog</span>
          <span className="ml-2 text-xs px-1.5 py-0.5 rounded bg-primary text-white">
            v{status?.version || '0.16'}
          </span>
        </div>

        {status?.has_context && status.context && (
          <div className="context-badge">
            <div className="text-xs text-muted mb-1">Active Context</div>
            <div className="font-medium text-sm truncate">{status.context.client}</div>
            <div className="text-xs text-muted truncate">
              {status.context.engagement} / {status.context.phase}
            </div>
          </div>
        )}

        <nav className="sidebar-nav">
          {navItems.map((item) => {
            const Icon = item.icon
            const isActive = location.pathname === item.path
            return (
              <Link
                key={item.path}
                to={item.path}
                className={isActive ? 'active' : ''}
              >
                <Icon className="h-4 w-4" />
                {item.label}
              </Link>
            )
          })}
        </nav>

        <div className="sidebar-footer">
          {status?.total_sessions ?? 0} sessions recorded
        </div>
      </aside>

      <main className="main-content">
        <div className="content-wrapper">
          {children}
        </div>
      </main>
    </div>
  )
}