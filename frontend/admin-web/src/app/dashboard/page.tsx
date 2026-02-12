'use client'

import useSWR from 'swr'
import { Users, TrendingUp, Calendar, DollarSign } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import { adminGet } from '@/lib/api/client'
import { useAdminAuth } from '@/hooks/use-admin-auth'
import { StatCard } from '@/components/stat-card'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'

interface DashboardData {
  total_users: number
  total_bets: number
  active_events: number
  total_volume: number
  recent_bets: {
    id: string
    user_display_name: string
    event_question: string
    side: string
    amount: number
    created_at: string
  }[]
}

function formatCredits(value: number): string {
  return (value / 100).toLocaleString() + ' credits'
}

export default function DashboardPage() {
  useAdminAuth()

  const { data, error, isLoading } = useSWR<DashboardData>(
    '/api/v1/dashboard',
    adminGet,
    { refreshInterval: 30000 }
  )

  if (isLoading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <p className="text-muted-foreground">Loading dashboard...</p>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex h-64 items-center justify-center">
        <p className="text-destructive">Failed to load dashboard data.</p>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold">Dashboard</h1>

      {/* Stat cards */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard
          icon={Users}
          label="Total Users"
          value={data?.total_users?.toLocaleString() ?? '0'}
        />
        <StatCard
          icon={TrendingUp}
          label="Total Bets"
          value={data?.total_bets?.toLocaleString() ?? '0'}
        />
        <StatCard
          icon={Calendar}
          label="Active Events"
          value={data?.active_events?.toLocaleString() ?? '0'}
        />
        <StatCard
          icon={DollarSign}
          label="Total Volume"
          value={formatCredits(data?.total_volume ?? 0)}
        />
      </div>

      {/* Recent bets table */}
      <Card>
        <CardHeader>
          <CardTitle>Recent Bets</CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>User</TableHead>
                <TableHead>Event</TableHead>
                <TableHead>Side</TableHead>
                <TableHead className="text-right">Amount</TableHead>
                <TableHead className="text-right">Time</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {data?.recent_bets && data.recent_bets.length > 0 ? (
                data.recent_bets.map((bet) => (
                  <TableRow key={bet.id}>
                    <TableCell className="font-medium">
                      {bet.user_display_name}
                    </TableCell>
                    <TableCell className="max-w-[200px] truncate">
                      {bet.event_question}
                    </TableCell>
                    <TableCell>
                      <Badge variant={bet.side === 'yes' ? 'default' : 'secondary'}>
                        {bet.side.toUpperCase()}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-right">
                      {formatCredits(bet.amount)}
                    </TableCell>
                    <TableCell className="text-right text-muted-foreground">
                      {formatDistanceToNow(new Date(bet.created_at), { addSuffix: true })}
                    </TableCell>
                  </TableRow>
                ))
              ) : (
                <TableRow>
                  <TableCell colSpan={5} className="text-center text-muted-foreground">
                    No recent bets
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  )
}
