import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus, Trash2 } from 'lucide-react'
import { api } from '../api/client'
import type { NetworkNode } from '../types'
import { Button } from '../components/ui/Button'
import { Input } from '../components/ui/Input'
import { IpInput } from '../components/ui/IpInput'

export function NetworkNodesPage() {
  const queryClient = useQueryClient()
  const [name, setName] = useState('')
  const [ip, setIp] = useState('')

  const { data: nodes = [], isLoading } = useQuery({
    queryKey: ['network-nodes'],
    queryFn: () => api.get<NetworkNode[]>('/api/network-nodes'),
  })

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
      <h2 className="text-xl font-semibold mb-6">Network Nodes</h2>

      <form onSubmit={handleSubmit} className="flex gap-3 mb-6">
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
        <div className="bg-slate-800 rounded-xl border border-slate-700 overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-slate-700 text-slate-400">
                <th className="text-left p-3 font-medium">Name</th>
                <th className="text-left p-3 font-medium">IP Address</th>
                <th className="p-3 w-12"></th>
              </tr>
            </thead>
            <tbody>
              {nodes.map(n => (
                <tr key={n.id} className="border-b border-slate-700/50 last:border-0">
                  <td className="p-3 text-slate-200">{n.name}</td>
                  <td className="p-3 text-slate-300 font-mono">{n.ip}</td>
                  <td className="p-3">
                    <button
                      onClick={() => deleteMut.mutate(n.id)}
                      className="text-slate-400 hover:text-red-400 transition-colors"
                    >
                      <Trash2 size={16} />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
