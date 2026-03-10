import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus, Trash2, Shield, ShieldOff } from 'lucide-react'
import { api } from '../api/client'
import type { NginxDomain } from '../types'
import { Button } from '../components/ui/Button'
import { Input } from '../components/ui/Input'
import { Modal } from '../components/ui/Modal'
import { Toggle } from '../components/ui/Toggle'

export function NginxDomainsPage() {
  const queryClient = useQueryClient()
  const [showForm, setShowForm] = useState(false)
  const [domain, setDomain] = useState('')
  const [upstreamIp, setUpstreamIp] = useState('')
  const [upstreamPort, setUpstreamPort] = useState('80')
  const [sslEmail, setSslEmail] = useState('')
  const [sslDomainId, setSslDomainId] = useState<number | null>(null)

  const { data: domains = [], isLoading } = useQuery({
    queryKey: ['nginx-domains'],
    queryFn: () => api.get<NginxDomain[]>('/api/nginx-domains'),
  })

  const createMut = useMutation({
    mutationFn: () => api.post('/api/nginx-domains', { domain, upstreamIp, upstreamPort: +upstreamPort }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['nginx-domains'] })
      setShowForm(false)
      setDomain('')
      setUpstreamIp('')
      setUpstreamPort('80')
    },
  })

  const deleteMut = useMutation({
    mutationFn: (id: number) => api.delete(`/api/nginx-domains/${id}`),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['nginx-domains'] }),
  })

  const toggleMut = useMutation({
    mutationFn: ({ id, enabled }: { id: number; enabled: boolean }) =>
      api.post(`/api/nginx-domains/${id}/${enabled ? 'enable' : 'disable'}`),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['nginx-domains'] }),
  })

  const sslMut = useMutation({
    mutationFn: ({ id, email }: { id: number; email: string }) =>
      api.post(`/api/nginx-domains/${id}/ssl`, { email }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['nginx-domains'] })
      setSslDomainId(null)
      setSslEmail('')
    },
  })

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-semibold">Nginx Domains</h2>
        <Button size="sm" onClick={() => setShowForm(true)}>
          <Plus size={16} className="mr-1.5" /> Add Domain
        </Button>
      </div>

      {isLoading ? (
        <p className="text-slate-400">Loading...</p>
      ) : domains.length === 0 ? (
        <p className="text-slate-500">No domains configured.</p>
      ) : (
        <div className="bg-slate-800 rounded-xl border border-slate-700 overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-slate-700 text-slate-400">
                <th className="text-left p-3 font-medium">Status</th>
                <th className="text-left p-3 font-medium">Domain</th>
                <th className="text-left p-3 font-medium">Upstream</th>
                <th className="text-left p-3 font-medium">SSL</th>
                <th className="p-3 w-24"></th>
              </tr>
            </thead>
            <tbody>
              {domains.map(d => (
                <tr key={d.id} className="border-b border-slate-700/50 last:border-0">
                  <td className="p-3">
                    <Toggle checked={d.enabled} onChange={enabled => toggleMut.mutate({ id: d.id, enabled })} />
                  </td>
                  <td className="p-3 text-slate-200 font-mono">{d.domain}</td>
                  <td className="p-3 text-slate-300 font-mono">{d.upstreamIp}:{d.upstreamPort}</td>
                  <td className="p-3">
                    {d.sslEnabled ? (
                      <span className="text-green-400 flex items-center gap-1"><Shield size={14} /> Active</span>
                    ) : (
                      <button
                        onClick={() => setSslDomainId(d.id)}
                        className="text-slate-400 hover:text-yellow-400 flex items-center gap-1 text-sm transition-colors"
                      >
                        <ShieldOff size={14} /> Enable SSL
                      </button>
                    )}
                  </td>
                  <td className="p-3 text-right">
                    <button onClick={() => deleteMut.mutate(d.id)} className="text-slate-400 hover:text-red-400 transition-colors">
                      <Trash2 size={16} />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <Modal open={showForm} onClose={() => setShowForm(false)} title="Add Domain">
        <form onSubmit={e => { e.preventDefault(); createMut.mutate() }} className="space-y-4">
          <Input label="Domain" placeholder="example.com" value={domain} onChange={e => setDomain(e.target.value)} required />
          <Input label="Upstream IP" placeholder="10.7.0.3" value={upstreamIp} onChange={e => setUpstreamIp(e.target.value)} required />
          <Input label="Upstream Port" type="number" value={upstreamPort} onChange={e => setUpstreamPort(e.target.value)} />
          {createMut.error && <p className="text-sm text-red-400">{createMut.error.message}</p>}
          <div className="flex justify-end gap-2">
            <Button variant="ghost" type="button" onClick={() => setShowForm(false)}>Cancel</Button>
            <Button type="submit" loading={createMut.isPending}>Create</Button>
          </div>
        </form>
      </Modal>

      <Modal open={sslDomainId !== null} onClose={() => setSslDomainId(null)} title="Enable SSL (Let's Encrypt)">
        <form onSubmit={e => { e.preventDefault(); if (sslDomainId) sslMut.mutate({ id: sslDomainId, email: sslEmail }) }} className="space-y-4">
          <Input label="Email for Let's Encrypt" placeholder="admin@example.com" value={sslEmail} onChange={e => setSslEmail(e.target.value)} required type="email" />
          {sslMut.error && <p className="text-sm text-red-400">{sslMut.error.message}</p>}
          <div className="flex justify-end gap-2">
            <Button variant="ghost" type="button" onClick={() => setSslDomainId(null)}>Cancel</Button>
            <Button type="submit" loading={sslMut.isPending}>Request Certificate</Button>
          </div>
        </form>
      </Modal>
    </div>
  )
}
