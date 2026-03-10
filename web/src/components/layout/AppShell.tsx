import { Outlet } from 'react-router-dom'
import { Sidebar } from './Sidebar'

export function AppShell() {
  return (
    <div className="flex min-h-screen">
      <Sidebar />
      <main className="flex-1 p-6 overflow-auto">
        <div className="mx-auto w-full max-w-[900px]">
          <Outlet />
        </div>
      </main>
    </div>
  )
}
