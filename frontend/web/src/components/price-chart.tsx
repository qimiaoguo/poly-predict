'use client'

import { useState } from 'react'
import useSWR from 'swr'
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from 'recharts'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { apiGet } from '@/lib/api/client'

interface PriceHistoryRecord {
  id: number
  event_id: string
  outcome_label: string
  price: number
  recorded_at: string
}

interface PriceChartProps {
  eventId: string
}

const PERIODS = [
  { label: '1H', value: '1h' },
  { label: '6H', value: '6h' },
  { label: '24H', value: '24h' },
  { label: '7D', value: '7d' },
  { label: '30D', value: '30d' },
]

function formatTime(timestamp: string, period: string): string {
  const date = new Date(timestamp)
  if (period === '1h' || period === '6h' || period === '24h') {
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  }
  return date.toLocaleDateString([], { month: 'short', day: 'numeric' })
}

function groupByTimestamp(records: PriceHistoryRecord[]) {
  const map = new Map<string, { yes: number; no: number; recorded_at: string }>()
  for (const r of records) {
    const key = r.recorded_at
    const entry = map.get(key) ?? { yes: 0, no: 0, recorded_at: key }
    if (r.outcome_label === 'Yes') {
      entry.yes = Math.round(r.price * 100)
    } else if (r.outcome_label === 'No') {
      entry.no = Math.round(r.price * 100)
    }
    map.set(key, entry)
  }
  return Array.from(map.values()).sort(
    (a, b) => new Date(a.recorded_at).getTime() - new Date(b.recorded_at).getTime()
  )
}

export function PriceChart({ eventId }: PriceChartProps) {
  const [period, setPeriod] = useState('24h')

  const { data: prices, isLoading } = useSWR<PriceHistoryRecord[]>(
    `/api/v1/events/${eventId}/prices?period=${period}`,
    (url: string) => apiGet<PriceHistoryRecord[]>(url),
    {
      refreshInterval: 30000,
    }
  )

  const chartData = groupByTimestamp(prices ?? []).map((point) => ({
    ...point,
    time: formatTime(point.recorded_at, period),
  }))

  return (
    <Card className="border-border/50">
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg">Price History</CardTitle>
          <div className="flex gap-0.5 rounded-lg bg-muted p-0.5">
            {PERIODS.map((p) => (
              <button
                key={p.value}
                className={`rounded-md px-2.5 py-1 text-xs font-medium transition-all ${
                  period === p.value
                    ? 'bg-background text-foreground shadow-sm'
                    : 'text-muted-foreground hover:text-foreground'
                }`}
                onClick={() => setPeriod(p.value)}
              >
                {p.label}
              </button>
            ))}
          </div>
        </div>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <div className="flex h-[300px] items-center justify-center text-sm text-muted-foreground">
            Loading chart...
          </div>
        ) : chartData.length === 0 ? (
          <div className="flex h-[300px] items-center justify-center text-sm text-muted-foreground">
            No price data available
          </div>
        ) : (
          <ResponsiveContainer width="100%" height={300}>
            <LineChart data={chartData}>
              <CartesianGrid
                strokeDasharray="3 3"
                stroke="var(--border)"
                strokeOpacity={0.5}
              />
              <XAxis
                dataKey="time"
                tick={{ fontSize: 11, fill: 'var(--muted-foreground)' }}
                tickLine={false}
                axisLine={false}
              />
              <YAxis
                domain={[0, 100]}
                tick={{ fontSize: 11, fill: 'var(--muted-foreground)' }}
                tickLine={false}
                axisLine={false}
                tickFormatter={(value) => `${value}%`}
              />
              <Tooltip
                formatter={(value: number) => [`${value}%`]}
                contentStyle={{
                  backgroundColor: 'var(--card)',
                  border: '1px solid var(--border)',
                  borderRadius: '8px',
                  fontSize: '12px',
                  boxShadow: '0 4px 12px rgba(0,0,0,0.1)',
                }}
                labelStyle={{ color: 'var(--muted-foreground)', fontSize: '11px' }}
              />
              <Legend
                wrapperStyle={{ fontSize: '12px', paddingTop: '8px' }}
              />
              <Line
                type="monotone"
                dataKey="yes"
                stroke="var(--chart-1)"
                strokeWidth={2}
                dot={false}
                name="Yes"
              />
              <Line
                type="monotone"
                dataKey="no"
                stroke="var(--chart-2)"
                strokeWidth={2}
                dot={false}
                name="No"
              />
            </LineChart>
          </ResponsiveContainer>
        )}
      </CardContent>
    </Card>
  )
}
