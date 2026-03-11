import { useMemo } from 'react'
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts'
import type { TrafficPoint } from '../../types'

interface Props {
  title: string
  points: TrafficPoint[]
  interval: string
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  const val = bytes / Math.pow(1024, i)
  return `${val.toFixed(val < 10 ? 1 : 0)} ${units[i]}`
}

function formatTime(iso: string, interval: string): string {
  const d = new Date(iso)
  if (interval === '1d' || interval === '1w') {
    return d.toLocaleDateString(undefined, { month: 'short', day: 'numeric' })
  }
  return d.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit', second: interval === '1m' || interval === '5m' ? '2-digit' : undefined })
}

function CustomTooltip({ active, payload, label }: { active?: boolean; payload?: Array<{ value: number; name: string; color: string }>; label?: string }) {
  if (!active || !payload?.length) return null
  return (
    <div className="bg-slate-900 border border-slate-700 rounded-lg p-2.5 text-xs shadow-xl">
      <p className="text-slate-400 mb-1.5">{label}</p>
      {payload.map(p => (
        <p key={p.name} style={{ color: p.color }} className="flex justify-between gap-4">
          <span>{p.name === 'in' ? 'In' : 'Out'}:</span>
          <span className="font-mono">{formatBytes(p.value)}</span>
        </p>
      ))}
    </div>
  )
}

export function TrafficChart({ title, points, interval }: Props) {
  const chartData = useMemo(() =>
    points.map(p => ({
      time: formatTime(p.t, interval),
      in: p.in,
      out: p.out,
    })),
    [points, interval],
  )

  const totalIn = points.reduce((sum, p) => sum + p.in, 0)
  const totalOut = points.reduce((sum, p) => sum + p.out, 0)

  return (
    <div className="bg-slate-800 rounded-xl border border-slate-700 p-4">
      <div className="flex items-center justify-between mb-3">
        <h4 className="text-sm font-medium text-slate-200">{title}</h4>
        <div className="flex gap-4 text-xs">
          <span className="text-blue-400">In: {formatBytes(totalIn)}</span>
          <span className="text-emerald-400">Out: {formatBytes(totalOut)}</span>
        </div>
      </div>
      {chartData.length === 0 ? (
        <div className="h-48 flex items-center justify-center text-slate-500 text-sm">
          No data for this interval
        </div>
      ) : (
        <ResponsiveContainer width="100%" height={200}>
          <AreaChart data={chartData} margin={{ top: 5, right: 5, left: 0, bottom: 0 }}>
            <defs>
              <linearGradient id={`gradIn-${title}`} x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.3} />
                <stop offset="95%" stopColor="#3b82f6" stopOpacity={0} />
              </linearGradient>
              <linearGradient id={`gradOut-${title}`} x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="#10b981" stopOpacity={0.3} />
                <stop offset="95%" stopColor="#10b981" stopOpacity={0} />
              </linearGradient>
            </defs>
            <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
            <XAxis
              dataKey="time"
              tick={{ fill: '#94a3b8', fontSize: 11 }}
              axisLine={{ stroke: '#475569' }}
              tickLine={false}
            />
            <YAxis
              tickFormatter={formatBytes}
              tick={{ fill: '#94a3b8', fontSize: 11 }}
              axisLine={{ stroke: '#475569' }}
              tickLine={false}
              width={60}
            />
            <Tooltip content={<CustomTooltip />} />
            <Area
              type="monotone"
              dataKey="in"
              stroke="#3b82f6"
              strokeWidth={2}
              fill={`url(#gradIn-${title})`}
            />
            <Area
              type="monotone"
              dataKey="out"
              stroke="#10b981"
              strokeWidth={2}
              fill={`url(#gradOut-${title})`}
            />
          </AreaChart>
        </ResponsiveContainer>
      )}
    </div>
  )
}
