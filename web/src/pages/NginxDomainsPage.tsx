import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus, Trash2, Shield, ShieldOff, Lock, Download, ExternalLink } from 'lucide-react'
import { api } from '../api/client'
import type { NginxDomain, ExternalNginxDomain } from '../types'
import { Button } from '../components/ui/Button'
import { Input } from '../components/ui/Input'
import { IpInput } from '../components/ui/IpInput'
import { Modal } from '../components/ui/Modal'
import { Toggle } from '../components/ui/Toggle'

export function NginxDomainsPage() {
  const queryClient = useQueryClient()
  const [showForm, setShowForm] = useState(false)
  const [domain, setDomain] = useState('')
  const [upstreamIp, setUpstreamIp] = useState('')
  const [upstreamPort, setUpstreamPort] = useState('80')
  const [basicAuthUser, setBasicAuthUser] = useState('')
  const [basicAuthPassword, setBasicAuthPassword] = useState('')
  const [sslEmail, setSslEmail] = useState('')
  const [sslDomainId, setSslDomainId] = useState<number | null>(null)
  const [importFilename, setImportFilename] = useState<string | null>(null)

  const { data: domains = [], isLoading } = useQuery({
    queryKey: ['nginx-domains'],
    queryFn: () => api.get<NginxDomain[]>('/api/nginx-domains'),
  })

  const { data: externalDomains = [] } = useQuery({
    queryKey: ['nginx-domains-external'],
    queryFn: () => api.get<ExternalNginxDomain[]>('/api/nginx-domains/external'),
  })

  const createMut = useMutation({
    mutationFn: () => api.post('/api/nginx-domains', {
      domain,
      upstreamIp,
      upstreamPort: +upstreamPort,
      basicAuthUser: basicAuthUser || undefined,
      basicAuthPassword: basicAuthPassword || undefined,
    }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['nginx-domains'] })
      setShowForm(false)
      setDomain('')
      setUpstreamIp('')
      setUpstreamPort('80')
      setBasicAuthUser('')
      setBasicAuthPassword('')
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

  const importMut = useMutation({
    mutationFn: (filename: string) => api.post('/api/nginx-domains/import', { filename }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['nginx-domains'] })
      queryClient.invalidateQueries({ queryKey: ['nginx-domains-external'] })
      setImportFilename(null)
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
      ) : domains.length === 0 && externalDomains.length === 0 ? (
        <p className="text-slate-500">No domains configured.</p>
      ) : (
        <div className="bg-slate-800 rounded-xl border border-slate-700 overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-slate-700 text-slate-400">
                <th className="text-left p-3 font-medium">Status</th>
                <th className="text-left p-3 font-medium">Domain</th>
                <th className="text-left p-3 font-medium">Upstream</th>
                <th className="text-left p-3 font-medium">Auth</th>
                <th className="text-left p-3 font-medium">SSL</th>
                <th className="p-3 w-24"></th>
              </tr>
            </thead>
            <tbody>
              {domains.map(d => (
                <tr key={`managed-${d.id}`} className="border-b border-slate-700/50 last:border-0">
                  <td className="p-3">
                    <Toggle checked={d.enabled} onChange={enabled => toggleMut.mutate({ id: d.id, enabled })} />
                  </td>
                  <td className="p-3 text-slate-200 font-mono">{d.domain}</td>
                  <td className="p-3 text-slate-300 font-mono">{d.upstreamIp}:{d.upstreamPort}</td>
                  <td className="p-3">
                    {d.basicAuthUser ? (
                      <span className="text-blue-400 flex items-center gap-1"><Lock size={14} /> {d.basicAuthUser}</span>
                    ) : (
                      <span className="text-slate-500">-</span>
                    )}
                  </td>
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
              {externalDomains.map(d => (
                <tr key={`ext-${d.filename}`} className="border-b border-slate-700/50 last:border-0 bg-slate-800/50">
                  <td className="p-3">
                    <span className="inline-flex items-center gap-1 text-xs font-medium text-amber-400 bg-amber-400/10 px-2 py-0.5 rounded-full">
                      <ExternalLink size={12} /> External
                    </span>
                  </td>
                  <td className="p-3 text-slate-200 font-mono">{d.domain}</td>
                  <td className="p-3 text-slate-300 font-mono">
                    {d.upstreamIp && d.upstreamPort ? `${d.upstreamIp}:${d.upstreamPort}` : <span className="text-slate-500">unknown</span>}
                  </td>
                  <td className="p-3">
                    {d.hasBasicAuth ? (
                      <span className="text-blue-400 flex items-center gap-1"><Lock size={14} /> Yes</span>
                    ) : (
                      <span className="text-slate-500">-</span>
                    )}
                  </td>
                  <td className="p-3">
                    {d.sslEnabled ? (
                      <span className="text-green-400 flex items-center gap-1"><Shield size={14} /> Active</span>
                    ) : (
                      <span className="text-slate-500">-</span>
                    )}
                  </td>
                  <td className="p-3 text-right">
                    <button
                      onClick={() => setImportFilename(d.filename)}
                      className="text-slate-400 hover:text-emerald-400 transition-colors"
                      title="Import to managed"
                    >
                      <Download size={16} />
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
          <IpInput label="Upstream IP" placeholder="10.7.0.3" value={upstreamIp} onChange={setUpstreamIp} required />
          <Input label="Upstream Port" type="number" value={upstreamPort} onChange={e => setUpstreamPort(e.target.value)} />
          <div className="border-t border-slate-700 pt-4">
            <p className="text-sm text-slate-400 mb-3">Basic Auth (optional)</p>
            <div className="space-y-3">
              <Input label="Username" placeholder="admin" value={basicAuthUser} onChange={e => setBasicAuthUser(e.target.value)} />
              <Input label="Password" type="password" placeholder="password" value={basicAuthPassword} onChange={e => setBasicAuthPassword(e.target.value)} />
            </div>
          </div>
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

      <Modal open={importFilename !== null} onClose={() => setImportFilename(null)} title="Import External Config">
        <div className="space-y-4">
          <p className="text-sm text-slate-300">
            Config <span className="font-mono text-amber-400">{importFilename}</span> will be imported and managed by system-control.
          </p>
          <p className="text-sm text-slate-400">
            A backup of the original config will be created (<span className="font-mono">.bak</span>).
            The config will be regenerated using the standard template.
          </p>
          <p className="text-sm text-yellow-400/80">
            Note: Basic Auth credentials cannot be imported. You can set them after import.
          </p>
          {importMut.error && <p className="text-sm text-red-400">{importMut.error.message}</p>}
          <div className="flex justify-end gap-2">
            <Button variant="ghost" onClick={() => setImportFilename(null)}>Cancel</Button>
            <Button
              onClick={() => importFilename && importMut.mutate(importFilename)}
              loading={importMut.isPending}
            >
              Import
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  )
}
