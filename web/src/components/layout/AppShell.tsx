import { useState } from 'react'
import { Outlet } from 'react-router-dom'
import { Sidebar } from './Sidebar'
import { NetworkNodesProvider } from '../../contexts/NetworkNodesContext'

export function AppShell() {
  const [sidebarOpen, setSidebarOpen] = useState(false)

  return (
    <NetworkNodesProvider>
      <div className="flex min-h-screen">
        {/* Desktop sidebar */}
        <div className="hidden md:block">
          <Sidebar />
        </div>

        {/* Mobile sidebar overlay */}
        {sidebarOpen && (
          <div className="fixed inset-0 z-40 md:hidden" onClick={() => setSidebarOpen(false)}>
            <div className="absolute inset-0 bg-black/50" />
            <div className="relative z-50" onClick={e => e.stopPropagation()}>
              <Sidebar onClose={() => setSidebarOpen(false)} />
            </div>
          </div>
        )}

        <main className="flex-1 p-4 md:p-6 overflow-auto">
          <div className="mx-auto w-full max-w-[900px]">
            <Outlet context={{ openSidebar: () => setSidebarOpen(true) }} />
          </div>
        </main>
      </div>
    </NetworkNodesProvider>
  )
}
