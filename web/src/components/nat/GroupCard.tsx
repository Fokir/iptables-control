import { Pencil, Trash2 } from 'lucide-react'
import type { NatGroup } from '../../types'
import { Toggle } from '../ui/Toggle'

interface Props {
  group: NatGroup
  onEdit: () => void
  onDelete: () => void
  onToggle: (enabled: boolean) => void
}

export function GroupCard({ group, onEdit, onDelete, onToggle }: Props) {
  return (
    <div className="bg-slate-800 rounded-xl border border-slate-700 p-4">
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-3">
          <Toggle checked={group.enabled} onChange={onToggle} />
          <h3 className="font-semibold text-lg">{group.name}</h3>
        </div>
        <div className="flex gap-2">
          <button onClick={onEdit} className="text-slate-400 hover:text-blue-400 transition-colors">
            <Pencil size={16} />
          </button>
          <button onClick={onDelete} className="text-slate-400 hover:text-red-400 transition-colors">
            <Trash2 size={16} />
          </button>
        </div>
      </div>

      <div className="grid grid-cols-3 gap-4 text-sm mb-3">
        <div>
          <span className="text-slate-500">Target IP</span>
          <p className="text-slate-200 font-mono">{group.targetIp}</p>
        </div>
        <div>
          <span className="text-slate-500">Destination IP</span>
          <p className={`font-mono ${group.destinationIp ? 'text-slate-200' : 'text-slate-500 italic'}`}>{group.destinationIp || 'Any'}</p>
        </div>
        <div>
          <span className="text-slate-500">SNAT IP</span>
          <p className={`font-mono ${group.targetReverseIp ? 'text-slate-200' : 'text-slate-500 italic'}`}>{group.targetReverseIp || 'Masquerade'}</p>
        </div>
      </div>

      {group.ports.length > 0 && (
        <div className="border-t border-slate-700 pt-3">
          <table className="w-full text-sm">
            <thead>
              <tr className="text-slate-500 text-left">
                <th className="pb-1 font-medium">External Port</th>
                <th className="pb-1 font-medium">Internal Port</th>
                <th className="pb-1 font-medium">Protocols</th>
              </tr>
            </thead>
            <tbody>
              {group.ports.map(p => (
                <tr key={p.id} className="text-slate-300">
                  <td className="py-0.5 font-mono">{p.externalPort}</td>
                  <td className="py-0.5 font-mono">{p.internalPort}</td>
                  <td className="py-0.5">
                    {[p.protocolTcp && 'TCP', p.protocolUdp && 'UDP'].filter(Boolean).join(', ')}
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
