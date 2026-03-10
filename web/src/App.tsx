import { Routes, Route } from 'react-router-dom'
import { AppShell } from './components/layout/AppShell'
import { AuthGuard } from './components/layout/AuthGuard'
import { LoginPage } from './pages/LoginPage'
import { NatRulesPage } from './pages/NatRulesPage'
import { NetworkNodesPage } from './pages/NetworkNodesPage'
import { NginxDomainsPage } from './pages/NginxDomainsPage'
import { AuditLogPage } from './pages/AuditLogPage'

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route element={<AuthGuard><AppShell /></AuthGuard>}>
        <Route path="/" element={<NatRulesPage />} />
        <Route path="/nodes" element={<NetworkNodesPage />} />
        <Route path="/nginx" element={<NginxDomainsPage />} />
        <Route path="/logs" element={<AuditLogPage />} />
      </Route>
    </Routes>
  )
}
