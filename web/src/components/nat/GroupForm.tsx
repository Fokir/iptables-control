import { useState } from 'react'
import { Plus, Trash2 } from 'lucide-react'
import type { NatGroup, NetworkNode, CreateGroupRequest, CreatePortRequest } from '../../types'
import { Button } from '../ui/Button'
import { Input } from '../ui/Input'
import { IpInput } from '../ui/IpInput'
import { Toggle } from '../ui/Toggle'

interface Props {
  group: NatGroup | null
  nodes: NetworkNode[]
  saving: boolean
  error?: string
  onSave: (data: CreateGroupRequest) => void
  onCancel: () => void
}

const emptyPort: CreatePortRequest = { externalPort: 0, internalPort: 0, protocolTcp: true, protocolUdp: false }

export function GroupForm({ group, nodes, saving, error, onSave, onCancel }: Props) {
  const [name, setName] = useState(group?.name ?? '')
  const [enabled, setEnabled] = useState(group?.enabled ?? true)
  const [targetIp, setTargetIp] = useState(group?.targetIp ?? '')
  const [targetReverseIp, setTargetReverseIp] = useState(group?.targetReverseIp ?? '')
  const [destinationIp, setDestinationIp] = useState(group?.destinationIp ?? '')
  const [ports, setPorts] = useState<CreatePortRequest[]>(
    group?.ports.map(p => ({
      externalPort: p.externalPort,
      internalPort: p.internalPort,
      protocolTcp: p.protocolTcp,
      protocolUdp: p.protocolUdp,
    })) ?? [{ ...emptyPort }]
  )

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    onSave({ name, enabled, targetIp, targetReverseIp, destinationIp, ports })
  }

  const updatePort = (idx: number, patch: Partial<CreatePortRequest>) => {
    setPorts(prev => prev.map((p, i) => i === idx ? { ...p, ...patch } : p))
  }

  const ipOptions = nodes.map(n => ({ label: `${n.name} (${n.ip})`, value: n.ip }))

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <Input label="Name" value={name} onChange={e => setName(e.target.value)} required />

      <div className="flex items-center gap-3">
        <Toggle checked={enabled} onChange={setEnabled} />
        <span className="text-sm text-slate-400">{enabled ? 'Enabled' : 'Disabled'}</span>
      </div>

      <IpField label="Target IP" value={targetIp} onChange={setTargetIp} options={ipOptions} required />
      <IpField label="Destination IP" value={destinationIp} onChange={setDestinationIp} options={ipOptions} placeholder="Any (match all incoming)" />
      <IpField label="SNAT IP" value={targetReverseIp} onChange={setTargetReverseIp} options={ipOptions} placeholder="Auto (masquerade)" />

      <div>
        <div className="flex items-center justify-between mb-2">
          <span className="text-sm text-slate-400">Ports</span>
          <button type="button" onClick={() => setPorts([...ports, { ...emptyPort }])} className="text-blue-400 hover:text-blue-300 text-sm flex items-center gap-1">
            <Plus size={14} /> Add port
          </button>
        </div>
        <div className="space-y-2">
          {ports.map((p, i) => (
            <div key={i} className="flex items-center gap-2 bg-slate-700/50 rounded-lg p-2">
              <input
                type="number"
                placeholder="External"
                value={p.externalPort || ''}
                onChange={e => updatePort(i, { externalPort: +e.target.value })}
                className="bg-slate-700 border border-slate-600 rounded px-2 py-1 text-sm w-24 text-slate-100"
                min={1}
                max={65535}
              />
              <span className="text-slate-500">→</span>
              <input
                type="number"
                placeholder="Internal"
                value={p.internalPort || ''}
                onChange={e => updatePort(i, { internalPort: +e.target.value })}
                className="bg-slate-700 border border-slate-600 rounded px-2 py-1 text-sm w-24 text-slate-100"
                min={1}
                max={65535}
              />
              <label className="flex items-center gap-1 text-sm text-slate-300">
                <input type="checkbox" checked={p.protocolTcp} onChange={e => updatePort(i, { protocolTcp: e.target.checked })} /> TCP
              </label>
              <label className="flex items-center gap-1 text-sm text-slate-300">
                <input type="checkbox" checked={p.protocolUdp} onChange={e => updatePort(i, { protocolUdp: e.target.checked })} /> UDP
              </label>
              {ports.length > 1 && (
                <button type="button" onClick={() => setPorts(ports.filter((_, j) => j !== i))} className="text-red-400 hover:text-red-300 ml-auto">
                  <Trash2 size={14} />
                </button>
              )}
            </div>
          ))}
        </div>
      </div>

      {error && <p className="text-sm text-red-400">{error}</p>}

      <div className="flex justify-end gap-2 pt-2">
        <Button variant="ghost" type="button" onClick={onCancel}>Cancel</Button>
        <Button type="submit" loading={saving}>Save</Button>
      </div>
    </form>
  )
}

function IpField({ label, value, onChange, options, required, placeholder }: { label: string; value: string; onChange: (v: string) => void; options: { label: string; value: string }[]; required?: boolean; placeholder?: string }) {
  return (
    <div className="flex flex-col gap-1">
      <label className="text-sm text-slate-400">{label}{!required && <span className="text-slate-600 ml-1">(optional)</span>}</label>
      <div className="flex gap-2">
        <IpInput
          value={value}
          onChange={onChange}
          className="flex-1"
          placeholder={placeholder ?? "0.0.0.0"}
          required={required}
        />
        {options.length > 0 && (
          <select
            value=""
            onChange={e => { if (e.target.value) onChange(e.target.value) }}
            className="bg-slate-700 border border-slate-600 rounded-lg px-2 py-2 text-slate-100 text-sm"
          >
            <option value="">Select node</option>
            {options.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
          </select>
        )}
      </div>
    </div>
  )
}
