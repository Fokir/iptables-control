import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { api } from '../api/client'
import type { TrafficSeries } from '../types'
import { TrafficChart } from '../components/traffic/TrafficChart'
import { PageHeader } from '../components/layout/PageHeader'

const intervals = [
  { value: '1m', label: '1 min' },
  { value: '5m', label: '5 min' },
  { value: '30m', label: '30 min' },
  { value: '1h', label: '1 hour' },
  { value: '1d', label: '1 day' },
  { value: '1w', label: '1 week' },
]

function refetchMs(interval: string): number {
  if (interval === '1m' || interval === '5m') return 10_000
  if (interval === '30m' || interval === '1h') return 60_000
  return 300_000
}

export function TrafficPage() {
  const [interval, setInterval] = useState('5m')

  const { data: nodeSeries = [], isLoading: nodesLoading } = useQuery({
    queryKey: ['traffic-nodes', interval],
    queryFn: () => api.get<TrafficSeries[]>(`/api/traffic/nodes?interval=${interval}`),
    refetchInterval: refetchMs(interval),
  })

  const { data: ifaceSeries = [], isLoading: ifaceLoading } = useQuery({
    queryKey: ['traffic-interfaces', interval],
    queryFn: () => api.get<TrafficSeries[]>(`/api/traffic/interfaces?interval=${interval}`),
    refetchInterval: refetchMs(interval),
  })

  return (
    <div>
      <PageHeader title="Traffic Statistics" />

      {/* Network Nodes */}
      <section className="mb-8">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 mb-4">
          <h3 className="text-sm font-medium text-slate-400 uppercase tracking-wider">Network Nodes</h3>
          <div className="flex gap-1 bg-slate-800 rounded-lg p-1 self-start sm:self-auto">
            {intervals.map(({ value, label }) => (
              <button
                key={value}
                onClick={() => setInterval(value)}
                className={`px-3 py-1.5 rounded-md text-xs font-medium transition-colors ${
                  interval === value
                    ? 'bg-blue-600 text-white'
                    : 'text-slate-400 hover:text-slate-200 hover:bg-slate-700'
                }`}
              >
                {label}
              </button>
            ))}
          </div>
        </div>
        {nodesLoading ? (
          <p className="text-slate-400">Loading...</p>
        ) : nodeSeries.length === 0 ? (
          <p className="text-slate-500 text-sm">No node traffic data yet. Data will appear after collection starts.</p>
        ) : (
          <div className="grid gap-4">
            {nodeSeries.map(s => (
              <TrafficChart
                key={s.nodeId}
                title={s.nodeName || `Node #${s.nodeId}`}
                points={s.points}
                interval={interval}
              />
            ))}
          </div>
        )}
      </section>

      {/* Interfaces */}
      <section>
        <h3 className="text-sm font-medium text-slate-400 uppercase tracking-wider mb-4">Network Interfaces</h3>
        {ifaceLoading ? (
          <p className="text-slate-400">Loading...</p>
        ) : ifaceSeries.length === 0 ? (
          <p className="text-slate-500 text-sm">No interface traffic data yet. Set TRAFFIC_INTERFACES env variable to monitor interfaces.</p>
        ) : (
          <div className="grid gap-4">
            {ifaceSeries.map(s => (
              <TrafficChart
                key={s.interfaceName}
                title={s.interfaceName || 'Unknown'}
                points={s.points}
                interval={interval}
              />
            ))}
          </div>
        )}
      </section>
    </div>
  )
}
