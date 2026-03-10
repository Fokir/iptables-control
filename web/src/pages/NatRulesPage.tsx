import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus, RefreshCw } from 'lucide-react'
import { api } from '../api/client'
import type { NatGroup, NetworkNode, CreateGroupRequest } from '../types'
import { Button } from '../components/ui/Button'
import { GroupCard } from '../components/nat/GroupCard'
import { GroupForm } from '../components/nat/GroupForm'
import { Modal } from '../components/ui/Modal'

export function NatRulesPage() {
  const queryClient = useQueryClient()
  const [editGroup, setEditGroup] = useState<NatGroup | null>(null)
  const [showForm, setShowForm] = useState(false)

  const { data: groups = [], isLoading } = useQuery({
    queryKey: ['nat-groups'],
    queryFn: () => api.get<NatGroup[]>('/api/nat-groups'),
  })

  const { data: nodes = [] } = useQuery({
    queryKey: ['network-nodes'],
    queryFn: () => api.get<NetworkNode[]>('/api/network-nodes'),
  })

  const deleteMut = useMutation({
    mutationFn: (id: number) => api.delete(`/api/nat-groups/${id}`),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['nat-groups'] }),
  })

  const toggleMut = useMutation({
    mutationFn: ({ id, enabled }: { id: number; enabled: boolean }) =>
      api.post(`/api/nat-groups/${id}/${enabled ? 'enable' : 'disable'}`),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['nat-groups'] }),
  })

  const syncMut = useMutation({
    mutationFn: () => api.post('/api/nat-groups/sync'),
  })

  const saveMut = useMutation({
    mutationFn: ({ id, data }: { id?: number; data: CreateGroupRequest }) =>
      id ? api.put(`/api/nat-groups/${id}`, data) : api.post('/api/nat-groups', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['nat-groups'] })
      setShowForm(false)
      setEditGroup(null)
    },
  })

  const openNew = () => {
    setEditGroup(null)
    setShowForm(true)
  }

  const openEdit = (g: NatGroup) => {
    setEditGroup(g)
    setShowForm(true)
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-semibold">NAT Rules</h2>
        <div className="flex gap-2">
          <Button variant="ghost" size="sm" onClick={() => syncMut.mutate()} loading={syncMut.isPending}>
            <RefreshCw size={16} className="mr-1.5" /> Sync
          </Button>
          <Button size="sm" onClick={openNew}>
            <Plus size={16} className="mr-1.5" /> Add Group
          </Button>
        </div>
      </div>

      {isLoading ? (
        <p className="text-slate-400">Loading...</p>
      ) : groups.length === 0 ? (
        <p className="text-slate-500">No NAT groups configured.</p>
      ) : (
        <div className="space-y-4">
          {groups.map(g => (
            <GroupCard
              key={g.id}
              group={g}
              onEdit={() => openEdit(g)}
              onDelete={() => deleteMut.mutate(g.id)}
              onToggle={enabled => toggleMut.mutate({ id: g.id, enabled })}
            />
          ))}
        </div>
      )}

      <Modal open={showForm} onClose={() => { setShowForm(false); setEditGroup(null) }} title={editGroup ? 'Edit Group' : 'New Group'}>
        <GroupForm
          group={editGroup}
          nodes={nodes}
          saving={saveMut.isPending}
          error={saveMut.error?.message}
          onSave={(data) => saveMut.mutate({ id: editGroup?.id, data })}
          onCancel={() => { setShowForm(false); setEditGroup(null) }}
        />
      </Modal>
    </div>
  )
}
