import { NavLink, useNavigate } from 'react-router-dom'
import { Network, Globe, Server, ScrollText, LogOut, BarChart3, Shield, X } from 'lucide-react'
import { api } from '../../api/client'

const links = [
  { to: '/', label: 'NAT Rules', icon: Network },
  { to: '/nodes', label: 'Network Nodes', icon: Server },
  { to: '/traffic', label: 'Traffic', icon: BarChart3 },
  { to: '/wireguard', label: 'WireGuard', icon: Shield },
  { to: '/nginx', label: 'Nginx Domains', icon: Globe },
  { to: '/logs', label: 'Audit Log', icon: ScrollText },
]

interface Props {
  onClose?: () => void
}

export function Sidebar({ onClose }: Props) {
  const navigate = useNavigate()

  const handleLogout = async () => {
    try {
      await api.post('/api/auth/logout')
    } finally {
      navigate('/login')
    }
  }

  return (
    <aside className="w-64 bg-slate-900 border-r border-slate-800 flex flex-col h-screen sticky top-0 overflow-y-auto">
      <div className="p-4 border-b border-slate-800 flex items-center justify-between">
        <h1 className="text-lg font-bold text-blue-400">System Control</h1>
        {onClose && (
          <button onClick={onClose} className="text-slate-400 hover:text-slate-200 transition-colors md:hidden">
            <X size={20} />
          </button>
        )}
      </div>

      <nav className="flex-1 p-3 space-y-1">
        {links.map(({ to, label, icon: Icon }) => (
          <NavLink
            key={to}
            to={to}
            end={to === '/'}
            onClick={onClose}
            className={({ isActive }) =>
              `flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors ${
                isActive
                  ? 'bg-blue-600/20 text-blue-400'
                  : 'text-slate-400 hover:bg-slate-800 hover:text-slate-200'
              }`
            }
          >
            <Icon size={18} />
            {label}
          </NavLink>
        ))}
      </nav>

      <div className="p-3 border-t border-slate-800">
        <button
          onClick={handleLogout}
          className="flex items-center gap-3 px-3 py-2 rounded-lg text-sm text-slate-400 hover:bg-slate-800 hover:text-slate-200 transition-colors w-full"
        >
          <LogOut size={18} />
          Logout
        </button>
      </div>
    </aside>
  )
}
