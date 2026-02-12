'use client'

import { useState } from 'react'
import useSWR from 'swr'
import { Trophy, Medal, TrendingUp } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { apiGet } from '@/lib/api/client'
import { useAuthStore } from '@/lib/store'

interface RankingEntry {
  rank: number
  user_id: string
  display_name: string
  avatar_url: string | null
  total_profit: number
  win_rate: number
  roi: number
  total_bets: number
}

const PERIODS = [
  { value: 'all_time', label: 'All Time' },
  { value: 'weekly', label: 'Weekly' },
  { value: 'monthly', label: 'Monthly' },
]

function formatCredits(amount: number): string {
  return (amount / 100).toLocaleString()
}

function getRankIcon(rank: number) {
  if (rank === 1) return <Trophy className="h-5 w-5 text-yellow-500" />
  if (rank === 2) return <Medal className="h-5 w-5 text-gray-400" />
  if (rank === 3) return <Medal className="h-5 w-5 text-amber-600" />
  return <span className="flex h-5 w-5 items-center justify-center text-xs font-bold text-muted-foreground">{rank}</span>
}

function rankBg(rank: number) {
  if (rank === 1) return 'bg-yellow-500/5 border-l-2 border-l-yellow-500'
  if (rank === 2) return 'bg-gray-500/5 border-l-2 border-l-gray-400'
  if (rank === 3) return 'bg-amber-500/5 border-l-2 border-l-amber-600'
  return ''
}

export default function LeaderboardPage() {
  const [period, setPeriod] = useState('all_time')
  const { user } = useAuthStore()

  const { data: rankings, isLoading } = useSWR<RankingEntry[]>(
    `/api/v1/rankings?period=${period}`,
    (url: string) => apiGet<RankingEntry[]>(url),
    { refreshInterval: 60000 }
  )

  return (
    <div className="mx-auto max-w-4xl px-4 py-8 sm:px-6 lg:px-8">
      {/* Header */}
      <div className="mb-8 flex items-center gap-3">
        <div className="rounded-xl bg-yellow-500/10 p-2.5">
          <Trophy className="h-7 w-7 text-yellow-500" />
        </div>
        <div>
          <h1 className="text-2xl font-bold tracking-tight sm:text-3xl">Leaderboard</h1>
          <p className="text-sm text-muted-foreground">
            Top predictors ranked by performance
          </p>
        </div>
      </div>

      {/* Period Selector */}
      <Tabs
        value={period}
        onValueChange={setPeriod}
        className="mb-6"
      >
        <TabsList>
          {PERIODS.map((p) => (
            <TabsTrigger key={p.value} value={p.value}>
              {p.label}
            </TabsTrigger>
          ))}
        </TabsList>
      </Tabs>

      {/* Rankings Table */}
      <Card className="border-border/50">
        <CardHeader className="pb-3">
          <CardTitle className="text-lg">Rankings</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="space-y-3">
              {Array.from({ length: 10 }).map((_, i) => (
                <div
                  key={i}
                  className="h-14 rounded-lg bg-muted shimmer"
                />
              ))}
            </div>
          ) : !rankings || rankings.length === 0 ? (
            <div className="flex h-40 items-center justify-center text-sm text-muted-foreground">
              No rankings available for this period.
            </div>
          ) : (
            <>
              {/* Table Header */}
              <div className="mb-2 hidden grid-cols-12 gap-4 px-4 text-xs font-medium uppercase tracking-wider text-muted-foreground sm:grid">
                <div className="col-span-1">Rank</div>
                <div className="col-span-3">Player</div>
                <div className="col-span-2 text-right">Profit</div>
                <div className="col-span-2 text-right">Win Rate</div>
                <div className="col-span-2 text-right">ROI</div>
                <div className="col-span-2 text-right">Bets</div>
              </div>

              <div className="space-y-1">
                {rankings.map((entry) => {
                  const isCurrentUser = user?.id === entry.user_id
                  return (
                    <div
                      key={entry.user_id}
                      className={`grid grid-cols-12 items-center gap-4 rounded-lg px-4 py-3 text-sm transition-colors ${
                        isCurrentUser
                          ? 'bg-primary/5 ring-1 ring-primary/20'
                          : `hover:bg-muted/50 ${rankBg(entry.rank)}`
                      }`}
                    >
                      {/* Rank */}
                      <div className="col-span-2 sm:col-span-1">
                        {getRankIcon(entry.rank)}
                      </div>

                      {/* Player */}
                      <div className="col-span-5 sm:col-span-3">
                        <p className="truncate font-medium">
                          {entry.display_name}
                          {isCurrentUser && (
                            <span className="ml-1 text-xs text-primary">(you)</span>
                          )}
                        </p>
                      </div>

                      {/* Profit */}
                      <div className="col-span-5 text-right sm:col-span-2">
                        <span
                          className={`font-semibold ${
                            entry.total_profit >= 0
                              ? 'text-green-600 dark:text-green-400'
                              : 'text-red-600 dark:text-red-400'
                          }`}
                        >
                          {entry.total_profit >= 0 ? '+' : ''}
                          {formatCredits(entry.total_profit)}
                        </span>
                      </div>

                      {/* Win Rate - hidden on mobile */}
                      <div className="col-span-2 hidden text-right sm:block">
                        {Math.round(entry.win_rate * 100)}%
                      </div>

                      {/* ROI - hidden on mobile */}
                      <div className="col-span-2 hidden text-right sm:block">
                        <span className="flex items-center justify-end gap-1">
                          <TrendingUp className="h-3 w-3" />
                          {Math.round(entry.roi * 100)}%
                        </span>
                      </div>

                      {/* Bets - hidden on mobile */}
                      <div className="col-span-2 hidden text-right sm:block">
                        {entry.total_bets}
                      </div>
                    </div>
                  )
                })}
              </div>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
