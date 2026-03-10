import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { api } from '../api/client'
import type { AuditLogResult } from '../types'
import { Button } from '../components/ui/Button'

export function AuditLogPage() {
  const [page, setPage] = useState(1)
  const limit = 30

  const { data, isLoading } = useQuery({
    queryKey: ['audit-logs', page],
    queryFn: () => api.get<AuditLogResult>(`/api/audit-logs?page=${page}&limit=${limit}`),
  })

  const totalPages = data ? Math.ceil(data.total / limit) : 0

  const formatDate = (s: string) => {
    const d = new Date(s)
    return d.toLocaleString()
  }

  return (
    <div>
      <h2 className="text-xl font-semibold mb-6">Audit Log</h2>

      {isLoading ? (
        <p className="text-slate-400">Loading...</p>
      ) : !data || data.entries.length === 0 ? (
        <p className="text-slate-500">No audit log entries.</p>
      ) : (
        <>
          <div className="bg-slate-800 rounded-xl border border-slate-700 overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-slate-700 text-slate-400">
                  <th className="text-left p-3 font-medium">Time</th>
                  <th className="text-left p-3 font-medium">User</th>
                  <th className="text-left p-3 font-medium">Action</th>
                  <th className="text-left p-3 font-medium">Details</th>
                  <th className="text-left p-3 font-medium">IP</th>
                </tr>
              </thead>
              <tbody>
                {data.entries.map(e => (
                  <tr key={e.id} className="border-b border-slate-700/50 last:border-0">
                    <td className="p-3 text-slate-400 text-xs whitespace-nowrap">{formatDate(e.createdAt)}</td>
                    <td className="p-3 text-slate-300">{e.username}</td>
                    <td className="p-3">
                      <span className="bg-slate-700 text-slate-300 px-2 py-0.5 rounded text-xs font-mono">{e.action}</span>
                    </td>
                    <td className="p-3 text-slate-400 text-xs max-w-xs truncate">{e.details}</td>
                    <td className="p-3 text-slate-500 font-mono text-xs">{e.ipAddress}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {totalPages > 1 && (
            <div className="flex items-center justify-center gap-2 mt-4">
              <Button variant="ghost" size="sm" disabled={page <= 1} onClick={() => setPage(p => p - 1)}>
                Previous
              </Button>
              <span className="text-sm text-slate-400">
                Page {page} of {totalPages}
              </span>
              <Button variant="ghost" size="sm" disabled={page >= totalPages} onClick={() => setPage(p => p + 1)}>
                Next
              </Button>
            </div>
          )}
        </>
      )}
    </div>
  )
}
