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
import { Button } from '@/components/ui/button'
import { apiGet } from '@/lib/api/client'

interface PricePoint {
  timestamp: string
  yes_price: number
  no_price: number
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
  if (period === '1h' || period === '6h') {
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  }
  if (period === '24h') {
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  }
  return date.toLocaleDateString([], { month: 'short', day: 'numeric' })
}

export function PriceChart({ eventId }: PriceChartProps) {
  const [period, setPeriod] = useState('24h')

  const { data: prices, isLoading } = useSWR<PricePoint[]>(
    `/api/v1/events/${eventId}/prices?period=${period}`,
    (url: string) => apiGet<PricePoint[]>(url),
    {
      refreshInterval: 30000,
    }
  )

  const chartData = prices?.map((point) => ({
    ...point,
    time: formatTime(point.timestamp, period),
    yes: Math.round(point.yes_price * 100),
    no: Math.round(point.no_price * 100),
  })) || []

  return (
    <Card>
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg">Price History</CardTitle>
          <div className="flex gap-1">
            {PERIODS.map((p) => (
              <Button
                key={p.value}
                variant={period === p.value ? 'default' : 'ghost'}
                size="sm"
                className="h-7 px-2 text-xs"
                onClick={() => setPeriod(p.value)}
              >
                {p.label}
              </Button>
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
              <CartesianGrid strokeDasharray="3 3" className="opacity-30" />
              <XAxis
                dataKey="time"
                tick={{ fontSize: 12 }}
                tickLine={false}
                axisLine={false}
              />
              <YAxis
                domain={[0, 100]}
                tick={{ fontSize: 12 }}
                tickLine={false}
                axisLine={false}
                tickFormatter={(value) => `${value}%`}
              />
              <Tooltip
                formatter={(value: number) => [`${value}%`]}
                contentStyle={{
                  backgroundColor: 'hsl(var(--card))',
                  border: '1px solid hsl(var(--border))',
                  borderRadius: '6px',
                  fontSize: '12px',
                }}
              />
              <Legend />
              <Line
                type="monotone"
                dataKey="yes"
                stroke="#22c55e"
                strokeWidth={2}
                dot={false}
                name="Yes"
              />
              <Line
                type="monotone"
                dataKey="no"
                stroke="#ef4444"
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
