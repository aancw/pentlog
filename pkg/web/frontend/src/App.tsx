import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { Suspense, lazy } from 'react'
import { BrowserRouter, Route, Routes } from 'react-router-dom'
import Layout from './components/Layout'
import './index.css'

const Dashboard = lazy(() => import('./pages/Dashboard'))
const Sessions = lazy(() => import('./pages/Sessions'))
const SessionDetail = lazy(() => import('./pages/SessionDetail'))
const Vulns = lazy(() => import('./pages/Vulns'))
const Reports = lazy(() => import('./pages/Reports'))
const Search = lazy(() => import('./pages/Search'))
const Archives = lazy(() => import('./pages/Archives'))
const Recovery = lazy(() => import('./pages/Recovery'))
const Settings = lazy(() => import('./pages/Settings'))

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30000,
      refetchOnWindowFocus: false,
      retry: 1,
    },
  },
})

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Layout>
          <Suspense fallback={<div className="loading-state">Loading view…</div>}>
            <Routes>
              <Route path="/" element={<Dashboard />} />
              <Route path="/sessions" element={<Sessions />} />
              <Route path="/sessions/:id" element={<SessionDetail />} />
              <Route path="/vulns" element={<Vulns />} />
              <Route path="/reports" element={<Reports />} />
              <Route path="/search" element={<Search />} />
              <Route path="/archives" element={<Archives />} />
              <Route path="/recovery" element={<Recovery />} />
              <Route path="/settings" element={<Settings />} />
            </Routes>
          </Suspense>
        </Layout>
      </BrowserRouter>
    </QueryClientProvider>
  )
}
