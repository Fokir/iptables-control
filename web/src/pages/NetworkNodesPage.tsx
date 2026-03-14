import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus, Trash2 } from 'lucide-react'
import { api } from '../api/client'
import type { NetworkNode, NodeStatus } from '../types'
import { Button } from '../components/ui/Button'
import { Input } from '../components/ui/Input'
import { IpInput } from '../components/ui/IpInput'
import { PageHeader } from '../components/layout/PageHeader'

function formatTime(iso: string | null): string {
  if (!iso) return '—'
  const d = new Date(iso)
  return d.toLocaleString()
}

function StatusBadge({ status }: { status?: NodeStatus }) {
  if (!status) {
    return <span className="inline-flex items-center gap-1.5 text-slate-500 text-xs"><span className="w-2 h-2 rounded-full bg-slate-600" />—</span>
  }
  if (status.isOnline) {
    return <span className="inline-flex items-center gap-1.5 text-emerald-400 text-xs"><span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />Online</span>
  }
  return <span className="inline-flex items-center gap-1.5 text-red-400 text-xs"><span className="w-2 h-2 rounded-full bg-red-400" />Offline</span>
}

export function NetworkNodesPage() {
  const queryClient = useQueryClient()
  const [name, setName] = useState('')
  const [ip, setIp] = useState('')

  const { data: nodes = [], isLoading } = useQuery({
    queryKey: ['network-nodes'],
    queryFn: () => api.get<NetworkNode[]>('/api/network-nodes'),
  })

  const { data: statuses = [] } = useQuery({
    queryKey: ['network-nodes-status'],
    queryFn: () => api.get<NodeStatus[]>('/api/network-nodes/status'),
    refetchInterval: 10000,
  })

  const statusMap = new Map(statuses.map(s => [s.nodeId, s]))

  const addMut = useMutation({
    mutationFn: () => api.post('/api/network-nodes', { name, ip }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['network-nodes'] })
      setName('')
      setIp('')
    },
  })

  const deleteMut = useMutation({
    mutationFn: (id: number) => api.delete(`/api/network-nodes/${id}`),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['network-nodes'] }),
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    addMut.mutate()
  }

  return (
    <div>
      <PageHeader title="Network Nodes" />

      <form onSubmit={handleSubmit} className="flex flex-col sm:flex-row gap-3 mb-6">
        <Input placeholder="Node name" value={name} onChange={e => setName(e.target.value)} required />
        <IpInput placeholder="IP address" value={ip} onChange={setIp} required />
        <Button type="submit" loading={addMut.isPending} size="sm">
          <Plus size={16} className="mr-1" /> Add
        </Button>
      </form>

      {addMut.error && <p className="text-sm text-red-400 mb-4">{addMut.error.message}</p>}

      {isLoading ? (
        <p className="text-slate-400">Loading...</p>
      ) : nodes.length === 0 ? (
        <p className="text-slate-500">No network nodes configured.</p>
      ) : (
        <div className="space-y-4">
          {nodes.map(n => {
            const st = statusMap.get(n.id)
            return (
              <div key={n.id} className="bg-slate-800 rounded-xl border border-slate-700 p-4">
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center gap-3">
                    <StatusBadge status={st} />
                    <h3 className="font-semibold text-lg">{n.name}</h3>
                  </div>
                  <button
                    onClick={() => deleteMut.mutate(n.id)}
                    className="text-slate-400 hover:text-red-400 transition-colors"
                  >
                    <Trash2 size={16} />
                  </button>
                </div>

                <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 text-sm">
                  <div>
                    <span className="text-slate-500">IP Address</span>
                    <p className="text-slate-200 font-mono">{n.ip}</p>
                  </div>
                  <div>
                    <span className="text-slate-500">Ping</span>
                    <p className="text-slate-300 font-mono text-xs">
                      {st?.isOnline ? `${st.latencyMs.toFixed(1)} ms` : '—'}
                    </p>
                  </div>
                  <div>
                    <span className="text-slate-500">Online since</span>
                    <p className="text-slate-400 text-xs">
                      {st?.isOnline ? formatTime(st.firstOnline) : '—'}
                    </p>
                  </div>
                  <div>
                    <span className="text-slate-500">Last seen</span>
                    <p className="text-slate-400 text-xs">
                      {st && !st.isOnline ? formatTime(st.lastSeen) : '—'}
                    </p>
                  </div>
                </div>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}
