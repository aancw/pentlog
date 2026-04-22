import { useEffect, useMemo, useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import {
  Activity,
  Archive,
  ChevronRight,
  Moon,
  FolderOpen,
  LayoutDashboard,
  Menu,
  Search,
  Settings,
  Shield,
  Sun,
  Terminal,
  Wrench,
  X,
} from 'lucide-react'
import { useSystemStatus } from '../hooks/useApi'

interface LayoutProps {
  children: React.ReactNode
}

const navItems = [
  { path: '/', label: 'Dashboard', icon: LayoutDashboard },
  { path: '/sessions', label: 'Sessions', icon: FolderOpen },
  { path: '/vulns', label: 'Vulnerabilities', icon: Shield },
  { path: '/reports', label: 'Reports', icon: Terminal },
  { path: '/search', label: 'Search', icon: Search },
  { path: '/archives', label: 'Archives', icon: Archive },
  { path: '/recovery', label: 'Recovery', icon: Activity },
  { path: '/settings', label: 'Settings', icon: Settings },
]

export default function Layout({ children }: LayoutProps) {
  const location = useLocation()
  const { data: status } = useSystemStatus()
  const [mobileOpen, setMobileOpen] = useState(false)
  const [theme, setTheme] = useState<'light' | 'dark'>(() => {
    const stored = window.localStorage.getItem('pentlog-theme')
    if (stored === 'light' || stored === 'dark') {
      return stored
    }

    return 'dark'
  })

  const title = useMemo(() => {
    return navItems.find((item) => item.path === location.pathname)?.label ?? 'PentLog'
  }, [location.pathname])

  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme)
    window.localStorage.setItem('pentlog-theme', theme)
  }, [theme])

  function handleToggleTheme() {
    setTheme((current) => (current === 'dark' ? 'light' : 'dark'))
  }

  return (
    <div className="app-shell">
      <a className="skip-link" href="#main-content">
        Skip to content
      </a>

      <aside className={`sidebar ${mobileOpen ? 'sidebar-open' : ''}`}>
        <div className="sidebar-header">
          <div className="brand-mark">
            <Terminal size={18} />
          </div>
          <div>
            <div className="brand-title">PentLog</div>
            <div className="brand-subtitle">Evidence dashboard</div>
          </div>
          <button className="icon-button mobile-only" onClick={() => setMobileOpen(false)} aria-label="Close navigation">
            <X size={16} />
          </button>
        </div>

        <div className="sidebar-context">
          <div className="eyebrow">Active Context</div>
          {status?.has_context && status.context ? (
            <>
              <div className="sidebar-context-title">{status.context.client}</div>
              <div className="sidebar-context-copy">
                {status.context.engagement} <ChevronRight size={12} /> {status.context.phase}
              </div>
              {status.context.target && (
                <div className="pill-row">
                  <span className="pill pill-accent">{status.context.target}</span>
                  {status.context.target_ip && <span className="pill">{status.context.target_ip}</span>}
                </div>
              )}
            </>
          ) : (
            <div className="empty-inline">No active engagement</div>
          )}
        </div>

        <nav className="sidebar-nav">
          {navItems.map((item) => {
            const Icon = item.icon
            const isActive = item.path === '/' ? location.pathname === '/' : location.pathname.startsWith(item.path)
            return (
              <Link
                key={item.path}
                to={item.path}
                className={isActive ? 'nav-link active' : 'nav-link'}
                aria-current={isActive ? 'page' : undefined}
                onClick={() => setMobileOpen(false)}
              >
                <Icon size={16} />
                <span>{item.label}</span>
              </Link>
            )
          })}
        </nav>

        <div className="sidebar-footer">
          <div className="sidebar-footer-line">
            <Wrench size={14} />
            <span>v{status?.version ?? 'dev'}</span>
          </div>
          <div className="sidebar-footer-line subdued">{status?.total_sessions ?? 0} sessions indexed</div>
        </div>
      </aside>

      {mobileOpen && <button className="sidebar-backdrop" onClick={() => setMobileOpen(false)} aria-label="Close navigation" />}

      <main className="main-shell" id="main-content" tabIndex={-1}>
        <header className="topbar">
          <div>
            <div className="eyebrow">Workspace</div>
            <h1 className="topbar-title">{title}</h1>
          </div>

          <div className="topbar-actions">
            {status?.has_context && status.context ? (
              <div className="status-chip">
                <span className="status-dot" />
                <span>{status.context.client}</span>
              </div>
            ) : (
              <div className="status-chip status-chip-muted">Context required</div>
            )}

            <button
              className={`icon-button theme-toggle ${theme === 'dark' ? 'theme-toggle-active' : ''}`}
              onClick={handleToggleTheme}
              aria-label={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
              title={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
            >
              {theme === 'dark' ? <Sun size={16} /> : <Moon size={16} />}
            </button>

            <button className="icon-button mobile-only-inline" onClick={() => setMobileOpen(true)} aria-label="Open navigation">
              <Menu size={16} />
            </button>
          </div>
        </header>

        <div className="content-wrapper">{children}</div>
      </main>
    </div>
  )
}
