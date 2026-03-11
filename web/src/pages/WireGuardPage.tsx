import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus, Trash2, Download, QrCode, Pencil, Copy, Check, Shield, ShieldOff, RefreshCw } from 'lucide-react'
import { api } from '../api/client'
import type { WireguardStatus, WireguardPeer, WireguardPeerConfig } from '../types'
import { Button } from '../components/ui/Button'
import { Input } from '../components/ui/Input'
import { Modal } from '../components/ui/Modal'

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KiB', 'MiB', 'GiB', 'TiB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

function formatHandshake(ts?: string): string {
  if (!ts) return 'Never'
  const seconds = Math.floor(Date.now() / 1000 - parseInt(ts))
  if (seconds < 60) return `${seconds}s ago`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`
  return `${Math.floor(seconds / 86400)}d ago`
}

export function WireGuardPage() {
  const queryClient = useQueryClient()
  const [showAddForm, setShowAddForm] = useState(false)
  const [peerName, setPeerName] = useState('')
  const [peerDns, setPeerDns] = useState('')
  const [peerAllowedIps, setPeerAllowedIps] = useState('')
  const [configPeer, setConfigPeer] = useState<WireguardPeer | null>(null)
  const [qrPeer, setQrPeer] = useState<number | null>(null)
  const [qrData, setQrData] = useState<string | null>(null)
  const [copied, setCopied] = useState(false)
  const [editPeer, setEditPeer] = useState<WireguardPeer | null>(null)
  const [editForm, setEditForm] = useState({ name: '', dns: '', allowedIps: '' })

  // Enable form
  const [showEnableForm, setShowEnableForm] = useState(false)
  const [enableEndpoint, setEnableEndpoint] = useState('')
  const [enablePort, setEnablePort] = useState('51820')
  const [enableAddress, setEnableAddress] = useState('10.7.0.1/24')

  const { data: status, isLoading: statusLoading } = useQuery({
    queryKey: ['wg-status'],
    queryFn: () => api.get<WireguardStatus>('/api/wireguard/status'),
    refetchInterval: 10000,
  })

  const { data: peers = [], isLoading: peersLoading } = useQuery({
    queryKey: ['wg-peers'],
    queryFn: () => api.get<WireguardPeer[]>('/api/wireguard/peers'),
    enabled: !!status?.running,
    refetchInterval: 10000,
  })

  const enableMut = useMutation({
    mutationFn: (data: { endpoint: string; listenPort: number; address: string }) =>
      api.post('/api/wireguard/enable', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['wg-status'] })
      queryClient.invalidateQueries({ queryKey: ['wg-peers'] })
      setShowEnableForm(false)
    },
  })

  const disableMut = useMutation({
    mutationFn: () => api.post('/api/wireguard/disable'),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['wg-status'] })
      queryClient.invalidateQueries({ queryKey: ['wg-peers'] })
    },
  })

  const createMut = useMutation({
    mutationFn: () => api.post('/api/wireguard/peers', {
      name: peerName,
      dns: peerDns || undefined,
      allowedIps: peerAllowedIps || undefined,
    }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['wg-peers'] })
      queryClient.invalidateQueries({ queryKey: ['wg-status'] })
      setShowAddForm(false)
      setPeerName('')
      setPeerDns('')
      setPeerAllowedIps('')
    },
  })

  const deleteMut = useMutation({
    mutationFn: (id: number) => api.delete(`/api/wireguard/peers/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['wg-peers'] })
      queryClient.invalidateQueries({ queryKey: ['wg-status'] })
    },
  })

  const syncMut = useMutation({
    mutationFn: () => api.post('/api/wireguard/sync'),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['wg-peers'] })
      queryClient.invalidateQueries({ queryKey: ['wg-status'] })
    },
  })

  const updateMut = useMutation({
    mutationFn: () => api.put(`/api/wireguard/peers/${editPeer!.id}`, {
      name: editForm.name,
      dns: editForm.dns || undefined,
      allowedIps: editForm.allowedIps || undefined,
    }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['wg-peers'] })
      setEditPeer(null)
    },
  })

  const { data: peerConfig } = useQuery({
    queryKey: ['wg-peer-config', configPeer?.id],
    queryFn: () => api.get<WireguardPeerConfig>(`/api/wireguard/peers/${configPeer!.id}/config`),
    enabled: !!configPeer,
  })

  const handleShowQR = async (peerId: number) => {
    setQrPeer(peerId)
    setQrData(null)
    try {
      const res = await fetch(`/api/wireguard/peers/${peerId}/qrcode`, { credentials: 'include' })
      if (!res.ok) throw new Error('Failed to get QR code')
      const blob = await res.blob()
      setQrData(URL.createObjectURL(blob))
    } catch {
      setQrData(null)
    }
  }

  const handleCopyConfig = async () => {
    if (peerConfig?.config) {
      await navigator.clipboard.writeText(peerConfig.config)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    }
  }

  const handleDownloadConfig = () => {
    if (peerConfig?.config && configPeer) {
      const blob = new Blob([peerConfig.config], { type: 'text/plain' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `${configPeer.name}.conf`
      a.click()
      URL.revokeObjectURL(url)
    }
  }

  const openEdit = (p: WireguardPeer) => {
    setEditForm({ name: p.name, dns: p.dns, allowedIps: p.allowedIps })
    setEditPeer(p)
  }

  const handleEnable = () => {
    if (status?.running) {
      // Already configured, just start
      enableMut.mutate({ endpoint: '', listenPort: 0, address: '' })
    } else if (status?.installed) {
      // Installed but not running, start
      enableMut.mutate({ endpoint: '', listenPort: 0, address: '' })
    } else {
      // Need initial setup
      setShowEnableForm(true)
    }
  }

  if (statusLoading) {
    return <p className="text-slate-400">Loading...</p>
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-3">
          <h2 className="text-xl font-semibold">WireGuard</h2>
          {status?.running ? (
            <span className="inline-flex items-center gap-1.5 text-xs font-medium text-green-400 bg-green-400/10 px-2.5 py-1 rounded-full">
              <span className="w-1.5 h-1.5 rounded-full bg-green-400 animate-pulse" />
              Running
            </span>
          ) : status?.installed ? (
            <span className="inline-flex items-center gap-1.5 text-xs font-medium text-yellow-400 bg-yellow-400/10 px-2.5 py-1 rounded-full">
              Stopped
            </span>
          ) : (
            <span className="inline-flex items-center gap-1.5 text-xs font-medium text-slate-400 bg-slate-400/10 px-2.5 py-1 rounded-full">
              Not installed
            </span>
          )}
        </div>
        <div className="flex gap-2">
          {status?.running ? (
            <>
              <Button size="sm" variant="ghost" onClick={() => disableMut.mutate()} loading={disableMut.isPending}>
                <ShieldOff size={16} className="mr-1.5" /> Stop
              </Button>
              <Button size="sm" variant="ghost" onClick={() => syncMut.mutate()} loading={syncMut.isPending}>
                <RefreshCw size={16} className="mr-1.5" /> Sync
              </Button>
              <Button size="sm" onClick={() => setShowAddForm(true)}>
                <Plus size={16} className="mr-1.5" /> Add Peer
              </Button>
            </>
          ) : (
            <Button size="sm" onClick={handleEnable} loading={enableMut.isPending}>
              <Shield size={16} className="mr-1.5" /> Enable WireGuard
            </Button>
          )}
        </div>
      </div>

      {/* Server info */}
      {status?.running && (
        <div className="bg-slate-800 rounded-xl border border-slate-700 p-4 mb-6">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
            <div>
              <span className="text-slate-400 block">Endpoint</span>
              <span className="text-slate-200 font-mono">{status.endpoint}:{status.listenPort}</span>
            </div>
            <div>
              <span className="text-slate-400 block">Address</span>
              <span className="text-slate-200 font-mono">{status.address}</span>
            </div>
            <div>
              <span className="text-slate-400 block">Public Key</span>
              <span className="text-slate-200 font-mono text-xs break-all">{status.publicKey?.slice(0, 20)}...</span>
            </div>
            <div>
              <span className="text-slate-400 block">Peers</span>
              <span className="text-slate-200">{status.peerCount}</span>
            </div>
          </div>
        </div>
      )}

      {/* Peers table */}
      {status?.running && (
        peersLoading ? (
          <p className="text-slate-400">Loading peers...</p>
        ) : peers.length === 0 ? (
          <p className="text-slate-500">No peers configured.</p>
        ) : (
          <div className="bg-slate-800 rounded-xl border border-slate-700 overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-slate-700 text-slate-400">
                  <th className="text-left p-3 font-medium">Name</th>
                  <th className="text-left p-3 font-medium">Address</th>
                  <th className="text-left p-3 font-medium">Status</th>
                  <th className="text-left p-3 font-medium">Transfer</th>
                  <th className="p-3 w-36"></th>
                </tr>
              </thead>
              <tbody>
                {peers.map(p => {
                  const isOnline = p.latestHandshake && (Date.now() / 1000 - parseInt(p.latestHandshake)) < 180
                  return (
                    <tr key={p.id} className="border-b border-slate-700/50 last:border-0">
                      <td className="p-3">
                        <div className="flex items-center gap-2">
                          <span className="text-slate-200 font-medium">{p.name}</span>
                          {p.imported && (
                            <span className="text-xs font-medium text-amber-400 bg-amber-400/10 px-1.5 py-0.5 rounded">
                              Imported
                            </span>
                          )}
                        </div>
                        <span className="text-slate-500 text-xs block font-mono mt-0.5">{p.publicKey.slice(0, 20)}...</span>
                      </td>
                      <td className="p-3 text-slate-300 font-mono">{p.address}</td>
                      <td className="p-3">
                        {isOnline ? (
                          <span className="inline-flex items-center gap-1.5 text-green-400">
                            <span className="w-1.5 h-1.5 rounded-full bg-green-400" />
                            {formatHandshake(p.latestHandshake)}
                          </span>
                        ) : (
                          <span className="text-slate-500">
                            {p.latestHandshake ? formatHandshake(p.latestHandshake) : 'Never connected'}
                          </span>
                        )}
                        {p.endpoint && <span className="text-slate-500 text-xs block font-mono mt-0.5">{p.endpoint}</span>}
                      </td>
                      <td className="p-3 text-slate-400 text-xs">
                        {(p.transferRx || p.transferTx) ? (
                          <>
                            <span className="text-green-400/70">↓{formatBytes(p.transferRx || 0)}</span>
                            {' / '}
                            <span className="text-blue-400/70">↑{formatBytes(p.transferTx || 0)}</span>
                          </>
                        ) : '-'}
                      </td>
                      <td className="p-3 text-right">
                        <div className="flex justify-end gap-1.5">
                          {!p.imported && (
                            <>
                              <button
                                onClick={() => setConfigPeer(p)}
                                className="text-slate-400 hover:text-blue-400 transition-colors p-1"
                                title="Download config"
                              >
                                <Download size={15} />
                              </button>
                              <button
                                onClick={() => handleShowQR(p.id)}
                                className="text-slate-400 hover:text-purple-400 transition-colors p-1"
                                title="QR Code"
                              >
                                <QrCode size={15} />
                              </button>
                            </>
                          )}
                          <button
                            onClick={() => openEdit(p)}
                            className="text-slate-400 hover:text-blue-400 transition-colors p-1"
                            title="Edit"
                          >
                            <Pencil size={15} />
                          </button>
                          <button
                            onClick={() => deleteMut.mutate(p.id)}
                            className="text-slate-400 hover:text-red-400 transition-colors p-1"
                            title="Delete"
                          >
                            <Trash2 size={15} />
                          </button>
                        </div>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        )
      )}

      {/* Enable WireGuard form */}
      <Modal open={showEnableForm} onClose={() => setShowEnableForm(false)} title="Enable WireGuard">
        <form onSubmit={e => { e.preventDefault(); enableMut.mutate({ endpoint: enableEndpoint, listenPort: +enablePort, address: enableAddress }) }} className="space-y-4">
          <Input label="Server Endpoint (public IP)" placeholder="212.193.55.91" value={enableEndpoint} onChange={e => setEnableEndpoint(e.target.value)} required />
          <Input label="Listen Port" type="number" value={enablePort} onChange={e => setEnablePort(e.target.value)} />
          <Input label="Address (CIDR)" placeholder="10.7.0.1/24" value={enableAddress} onChange={e => setEnableAddress(e.target.value)} />
          {enableMut.error && <p className="text-sm text-red-400">{enableMut.error.message}</p>}
          <div className="flex justify-end gap-2">
            <Button variant="ghost" type="button" onClick={() => setShowEnableForm(false)}>Cancel</Button>
            <Button type="submit" loading={enableMut.isPending}>Enable</Button>
          </div>
        </form>
      </Modal>

      {/* Add peer modal */}
      <Modal open={showAddForm} onClose={() => setShowAddForm(false)} title="Add Peer">
        <form onSubmit={e => { e.preventDefault(); createMut.mutate() }} className="space-y-4">
          <Input label="Name" placeholder="my-device" value={peerName} onChange={e => setPeerName(e.target.value)} required />
          <Input label="DNS (optional)" placeholder="1.1.1.1, 1.0.0.1" value={peerDns} onChange={e => setPeerDns(e.target.value)} />
          <Input label="Allowed IPs (optional)" placeholder="auto" value={peerAllowedIps} onChange={e => setPeerAllowedIps(e.target.value)} />
          {createMut.error && <p className="text-sm text-red-400">{createMut.error.message}</p>}
          <div className="flex justify-end gap-2">
            <Button variant="ghost" type="button" onClick={() => setShowAddForm(false)}>Cancel</Button>
            <Button type="submit" loading={createMut.isPending}>Create</Button>
          </div>
        </form>
      </Modal>

      {/* Edit peer modal */}
      <Modal open={editPeer !== null} onClose={() => setEditPeer(null)} title="Edit Peer">
        <form onSubmit={e => { e.preventDefault(); updateMut.mutate() }} className="space-y-4">
          <Input label="Name" value={editForm.name} onChange={e => setEditForm(f => ({ ...f, name: e.target.value }))} required />
          <Input label="DNS" value={editForm.dns} onChange={e => setEditForm(f => ({ ...f, dns: e.target.value }))} />
          <Input label="Allowed IPs" value={editForm.allowedIps} onChange={e => setEditForm(f => ({ ...f, allowedIps: e.target.value }))} />
          {updateMut.error && <p className="text-sm text-red-400">{updateMut.error.message}</p>}
          <div className="flex justify-end gap-2">
            <Button variant="ghost" type="button" onClick={() => setEditPeer(null)}>Cancel</Button>
            <Button type="submit" loading={updateMut.isPending}>Save</Button>
          </div>
        </form>
      </Modal>

      {/* Config modal */}
      <Modal open={configPeer !== null} onClose={() => { setConfigPeer(null); setCopied(false) }} title={`Config: ${configPeer?.name}`}>
        <div className="space-y-4">
          {peerConfig ? (
            <>
              <pre className="bg-slate-900 rounded-lg p-4 text-xs text-slate-300 font-mono overflow-x-auto whitespace-pre-wrap border border-slate-700">
                {peerConfig.config}
              </pre>
              <div className="flex justify-end gap-2">
                <Button variant="ghost" size="sm" onClick={handleCopyConfig}>
                  {copied ? <Check size={14} className="mr-1.5" /> : <Copy size={14} className="mr-1.5" />}
                  {copied ? 'Copied' : 'Copy'}
                </Button>
                <Button size="sm" onClick={handleDownloadConfig}>
                  <Download size={14} className="mr-1.5" /> Download .conf
                </Button>
              </div>
            </>
          ) : (
            <p className="text-slate-400">Loading config...</p>
          )}
        </div>
      </Modal>

      {/* QR code modal */}
      <Modal open={qrPeer !== null} onClose={() => { setQrPeer(null); setQrData(null) }} title="QR Code">
        <div className="flex justify-center p-4">
          {qrData ? (
            <img src={qrData} alt="WireGuard QR Code" className="max-w-[280px] rounded-lg" />
          ) : (
            <p className="text-slate-400">Generating QR code...</p>
          )}
        </div>
      </Modal>
    </div>
  )
}
