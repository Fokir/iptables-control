import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus, Trash2, Shield, ShieldOff, Lock, Download, ExternalLink, Pencil } from 'lucide-react'
import { api } from '../api/client'
import type { NginxDomain, ExternalNginxDomain, NetworkNode } from '../types'
import { Button } from '../components/ui/Button'
import { Input } from '../components/ui/Input'
import { IpInput } from '../components/ui/IpInput'
import { Modal } from '../components/ui/Modal'
import { Toggle } from '../components/ui/Toggle'
import { PageHeader } from '../components/layout/PageHeader'
import { IpWithNodeName } from '../components/ui/IpWithNodeName'

export function NginxDomainsPage() {
  const queryClient = useQueryClient()
  const [showForm, setShowForm] = useState(false)
  const [domain, setDomain] = useState('')
  const [upstreamIp, setUpstreamIp] = useState('')
  const [upstreamPort, setUpstreamPort] = useState('80')
  const [upstreamScheme, setUpstreamScheme] = useState('http')
  const [upstreamSslVerify, setUpstreamSslVerify] = useState(true)
  const [basicAuthUser, setBasicAuthUser] = useState('')
  const [basicAuthPassword, setBasicAuthPassword] = useState('')
  const [sslEmail, setSslEmail] = useState('')
  const [sslDomainId, setSslDomainId] = useState<number | null>(null)
  const [importFilename, setImportFilename] = useState<string | null>(null)
  const [editDomain, setEditDomain] = useState<NginxDomain | null>(null)
  const [editForm, setEditForm] = useState({ domain: '', upstreamIp: '', upstreamPort: '80', upstreamScheme: 'http', upstreamSslVerify: true, basicAuthUser: '', basicAuthPassword: '' })

  const { data: domains = [], isLoading } = useQuery({
    queryKey: ['nginx-domains'],
    queryFn: () => api.get<NginxDomain[]>('/api/nginx-domains'),
  })

  const { data: nodes = [] } = useQuery({
    queryKey: ['network-nodes'],
    queryFn: () => api.get<NetworkNode[]>('/api/network-nodes'),
  })

  const ipOptions = nodes.map(n => ({ label: `${n.name} (${n.ip})`, value: n.ip }))

  const { data: externalDomains = [] } = useQuery({
    queryKey: ['nginx-domains-external'],
    queryFn: () => api.get<ExternalNginxDomain[]>('/api/nginx-domains/external'),
  })

  const createMut = useMutation({
    mutationFn: () => api.post('/api/nginx-domains', {
      domain,
      upstreamIp,
      upstreamPort: +upstreamPort,
      upstreamScheme,
      upstreamSslVerify,
      basicAuthUser: basicAuthUser || undefined,
      basicAuthPassword: basicAuthPassword || undefined,
    }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['nginx-domains'] })
      setShowForm(false)
      setDomain('')
      setUpstreamIp('')
      setUpstreamPort('80')
      setUpstreamScheme('http')
      setUpstreamSslVerify(true)
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

  const updateMut = useMutation({
    mutationFn: () => api.put(`/api/nginx-domains/${editDomain!.id}`, {
      domain: editForm.domain,
      upstreamIp: editForm.upstreamIp,
      upstreamPort: +editForm.upstreamPort,
      upstreamScheme: editForm.upstreamScheme,
      upstreamSslVerify: editForm.upstreamSslVerify,
      basicAuthUser: editForm.basicAuthUser || undefined,
      basicAuthPassword: editForm.basicAuthPassword || undefined,
    }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['nginx-domains'] })
      setEditDomain(null)
    },
  })

  const openEdit = (d: NginxDomain) => {
    setEditForm({
      domain: d.domain,
      upstreamIp: d.upstreamIp,
      upstreamPort: String(d.upstreamPort),
      upstreamScheme: d.upstreamScheme || 'http',
      upstreamSslVerify: d.upstreamSslVerify,
      basicAuthUser: d.basicAuthUser || '',
      basicAuthPassword: '',
    })
    setEditDomain(d)
  }

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
      <PageHeader title="Nginx Domains">
        <Button size="sm" onClick={() => setShowForm(true)}>
          <Plus size={16} className="mr-1.5" /> Add Domain
        </Button>
      </PageHeader>

      {isLoading ? (
        <p className="text-slate-400">Loading...</p>
      ) : domains.length === 0 && externalDomains.length === 0 ? (
        <p className="text-slate-500">No domains configured.</p>
      ) : (
        <div className="space-y-4">
          {domains.map(d => (
            <div key={`managed-${d.id}`} className="bg-slate-800 rounded-xl border border-slate-700 p-4">
              <div className="flex items-center justify-between mb-3">
                <div className="flex items-center gap-3">
                  <Toggle checked={d.enabled} onChange={enabled => toggleMut.mutate({ id: d.id, enabled })} />
                  <h3 className="font-semibold text-lg font-mono">{d.domain}</h3>
                </div>
                <div className="flex gap-2">
                  <button onClick={() => openEdit(d)} className="text-slate-400 hover:text-blue-400 transition-colors">
                    <Pencil size={16} />
                  </button>
                  <button onClick={() => deleteMut.mutate(d.id)} className="text-slate-400 hover:text-red-400 transition-colors">
                    <Trash2 size={16} />
                  </button>
                </div>
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 text-sm">
                <div>
                  <span className="text-slate-500">Upstream</span>
                  <p className="text-slate-200">
                    <span className="font-mono">{d.upstreamScheme || 'http'}://</span>
                    <IpWithNodeName ip={d.upstreamIp} />
                    <span className="font-mono">:{d.upstreamPort}</span>
                  </p>
                </div>
                <div>
                  <span className="text-slate-500">Auth</span>
                  {d.basicAuthUser ? (
                    <p className="text-blue-400 flex items-center gap-1"><Lock size={14} /> {d.basicAuthUser}</p>
                  ) : (
                    <p className="text-slate-500">—</p>
                  )}
                </div>
                <div>
                  <span className="text-slate-500">SSL</span>
                  {d.sslEnabled ? (
                    <p className="text-green-400 flex items-center gap-1"><Shield size={14} /> Active</p>
                  ) : (
                    <button
                      onClick={() => setSslDomainId(d.id)}
                      className="text-slate-400 hover:text-yellow-400 flex items-center gap-1 text-sm transition-colors"
                    >
                      <ShieldOff size={14} /> Enable SSL
                    </button>
                  )}
                </div>
              </div>
            </div>
          ))}

          {externalDomains.map(d => (
            <div key={`ext-${d.filename}`} className="bg-slate-800/50 rounded-xl border border-slate-700 p-4">
              <div className="flex items-center justify-between mb-3">
                <div className="flex items-center gap-3">
                  <span className="inline-flex items-center gap-1 text-xs font-medium text-amber-400 bg-amber-400/10 px-2 py-0.5 rounded-full">
                    <ExternalLink size={12} /> External
                  </span>
                  <h3 className="font-semibold text-lg font-mono">{d.domain}</h3>
                </div>
                <button
                  onClick={() => setImportFilename(d.filename)}
                  className="text-slate-400 hover:text-emerald-400 transition-colors"
                  title="Import to managed"
                >
                  <Download size={16} />
                </button>
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 text-sm">
                <div>
                  <span className="text-slate-500">Upstream</span>
                  {d.upstreamIp && d.upstreamPort ? (
                    <p className="text-slate-200">
                      <IpWithNodeName ip={d.upstreamIp} />
                      <span className="font-mono">:{d.upstreamPort}</span>
                    </p>
                  ) : (
                    <p className="text-slate-500">unknown</p>
                  )}
                </div>
                <div>
                  <span className="text-slate-500">Auth</span>
                  {d.hasBasicAuth ? (
                    <p className="text-blue-400 flex items-center gap-1"><Lock size={14} /> Yes</p>
                  ) : (
                    <p className="text-slate-500">—</p>
                  )}
                </div>
                <div>
                  <span className="text-slate-500">SSL</span>
                  {d.sslEnabled ? (
                    <p className="text-green-400 flex items-center gap-1"><Shield size={14} /> Active</p>
                  ) : (
                    <p className="text-slate-500">—</p>
                  )}
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      <Modal open={showForm} onClose={() => setShowForm(false)} title="Add Domain">
        <form onSubmit={e => { e.preventDefault(); createMut.mutate() }} className="space-y-4">
          <Input label="Domain" placeholder="example.com" value={domain} onChange={e => setDomain(e.target.value)} required />
          <div className="flex flex-col gap-1">
            <label className="text-sm text-slate-400">Upstream IP</label>
            <div className="flex gap-2">
              <IpInput value={upstreamIp} onChange={setUpstreamIp} className="flex-1" placeholder="10.7.0.3" required />
              {ipOptions.length > 0 && (
                <select
                  value=""
                  onChange={e => { if (e.target.value) setUpstreamIp(e.target.value) }}
                  className="bg-slate-700 border border-slate-600 rounded-lg px-2 py-2 text-slate-100 text-sm"
                >
                  <option value="">Select node</option>
                  {ipOptions.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
                </select>
              )}
            </div>
          </div>
          <div className="flex gap-4">
            <Input label="Upstream Port" type="number" value={upstreamPort} onChange={e => setUpstreamPort(e.target.value)} />
            <div className="flex flex-col gap-1">
              <label className="text-sm text-slate-400">Scheme</label>
              <select
                value={upstreamScheme}
                onChange={e => setUpstreamScheme(e.target.value)}
                className="bg-slate-700 border border-slate-600 rounded-lg px-3 py-2 text-slate-100 text-sm"
              >
                <option value="http">HTTP</option>
                <option value="https">HTTPS</option>
              </select>
            </div>
          </div>
          {upstreamScheme === 'https' && (
            <label className="flex items-center gap-2 text-sm text-slate-300 cursor-pointer">
              <input
                type="checkbox"
                checked={!upstreamSslVerify}
                onChange={e => setUpstreamSslVerify(!e.target.checked)}
                className="rounded border-slate-600 bg-slate-700 text-blue-500"
              />
              Skip upstream SSL certificate verification
            </label>
          )}
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

      <Modal open={editDomain !== null} onClose={() => setEditDomain(null)} title="Edit Domain">
        <form onSubmit={e => { e.preventDefault(); updateMut.mutate() }} className="space-y-4">
          <Input label="Domain" placeholder="example.com" value={editForm.domain} onChange={e => setEditForm(f => ({ ...f, domain: e.target.value }))} required />
          <div className="flex flex-col gap-1">
            <label className="text-sm text-slate-400">Upstream IP</label>
            <div className="flex gap-2">
              <IpInput value={editForm.upstreamIp} onChange={v => setEditForm(f => ({ ...f, upstreamIp: v }))} className="flex-1" placeholder="10.7.0.3" required />
              {ipOptions.length > 0 && (
                <select
                  value=""
                  onChange={e => { if (e.target.value) setEditForm(f => ({ ...f, upstreamIp: e.target.value })) }}
                  className="bg-slate-700 border border-slate-600 rounded-lg px-2 py-2 text-slate-100 text-sm"
                >
                  <option value="">Select node</option>
                  {ipOptions.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
                </select>
              )}
            </div>
          </div>
          <div className="flex gap-4">
            <Input label="Upstream Port" type="number" value={editForm.upstreamPort} onChange={e => setEditForm(f => ({ ...f, upstreamPort: e.target.value }))} />
            <div className="flex flex-col gap-1">
              <label className="text-sm text-slate-400">Scheme</label>
              <select
                value={editForm.upstreamScheme}
                onChange={e => setEditForm(f => ({ ...f, upstreamScheme: e.target.value }))}
                className="bg-slate-700 border border-slate-600 rounded-lg px-3 py-2 text-slate-100 text-sm"
              >
                <option value="http">HTTP</option>
                <option value="https">HTTPS</option>
              </select>
            </div>
          </div>
          {editForm.upstreamScheme === 'https' && (
            <label className="flex items-center gap-2 text-sm text-slate-300 cursor-pointer">
              <input
                type="checkbox"
                checked={!editForm.upstreamSslVerify}
                onChange={e => setEditForm(f => ({ ...f, upstreamSslVerify: !e.target.checked }))}
                className="rounded border-slate-600 bg-slate-700 text-blue-500"
              />
              Skip upstream SSL certificate verification
            </label>
          )}
          <div className="border-t border-slate-700 pt-4">
            <p className="text-sm text-slate-400 mb-3">Basic Auth (optional, leave password empty to keep current)</p>
            <div className="space-y-3">
              <Input label="Username" placeholder="admin" value={editForm.basicAuthUser} onChange={e => setEditForm(f => ({ ...f, basicAuthUser: e.target.value }))} />
              <Input label="Password" type="password" placeholder="new password" value={editForm.basicAuthPassword} onChange={e => setEditForm(f => ({ ...f, basicAuthPassword: e.target.value }))} />
            </div>
          </div>
          {updateMut.error && <p className="text-sm text-red-400">{updateMut.error.message}</p>}
          <div className="flex justify-end gap-2">
            <Button variant="ghost" type="button" onClick={() => setEditDomain(null)}>Cancel</Button>
            <Button type="submit" loading={updateMut.isPending}>Save</Button>
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
