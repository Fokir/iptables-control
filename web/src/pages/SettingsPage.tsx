import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../api/client'
import type { Settings } from '../types'
import { PageHeader } from '../components/layout/PageHeader'
import { Toggle } from '../components/ui/Toggle'

export function SettingsPage() {
  const queryClient = useQueryClient()

  const { data, isLoading } = useQuery({
    queryKey: ['settings'],
    queryFn: () => api.get<Settings>('/api/settings'),
  })

  const mutation = useMutation({
    mutationFn: (forceTls12: boolean) =>
      api.put<Settings>('/api/settings', { forceTls12 }),
    onSuccess: (updated) => {
      queryClient.setQueryData(['settings'], updated)
    },
  })

  return (
    <div>
      <PageHeader title="Settings" />

      <div className="max-w-2xl">
        <div className="bg-slate-800 rounded-lg p-5">
          <div className="flex items-start justify-between gap-4">
            <div>
              <h3 className="text-sm font-medium text-slate-200">
                Только TLS 1.2 (отключить TLS 1.3)
              </h3>
              <p className="mt-1 text-sm text-slate-400">
                Помогает обойти новый способ блокировки РКН на некоторых российских
                провайдерах. HTTP/2 уже включён автоматически для всех сайтов по HTTPS.
                Изменение перестраивает конфигурацию всех доменов с SSL.
              </p>
            </div>
            <Toggle
              checked={data?.forceTls12 ?? false}
              disabled={isLoading || mutation.isPending}
              onChange={(v) => mutation.mutate(v)}
            />
          </div>
        </div>
      </div>
    </div>
  )
}
