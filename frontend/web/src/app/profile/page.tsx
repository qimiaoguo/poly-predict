'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import useSWR from 'swr'
import { format } from 'date-fns'
import {
  Wallet,
  Lock,
  Star,
  Zap,
  Flame,
  Target,
  Pencil,
  Check,
  X,
} from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useAuth } from '@/hooks/use-auth'
import { apiPatch, apiFetchPaginated } from '@/lib/api/client'
import { useToast } from '@/hooks/use-toast'

interface Bet {
  id: string
  event_id: string
  event_question: string
  outcome: 'yes' | 'no'
  amount: number
  odds: number
  status: 'pending' | 'won' | 'lost' | 'cancelled'
  potential_payout: number
  created_at: string
}

interface Transaction {
  id: string
  type: string
  amount: number
  description: string
  created_at: string
}

function formatCredits(amount: number): string {
  return amount.toLocaleString()
}

function betStatusColor(status: string) {
  switch (status) {
    case 'pending':
      return 'bg-yellow-500/10 text-yellow-700 dark:text-yellow-400 border-yellow-500/20'
    case 'won':
      return 'bg-green-500/10 text-green-700 dark:text-green-400 border-green-500/20'
    case 'lost':
      return 'bg-red-500/10 text-red-700 dark:text-red-400 border-red-500/20'
    case 'cancelled':
      return 'bg-gray-500/10 text-gray-700 dark:text-gray-400 border-gray-500/20'
    default:
      return ''
  }
}

const STAT_CARDS = [
  { key: 'balance', label: 'Balance', icon: Wallet, color: 'text-green-500', borderColor: 'border-l-green-500' },
  { key: 'frozen', label: 'Frozen', icon: Lock, color: 'text-blue-500', borderColor: 'border-l-blue-500' },
  { key: 'level', label: 'Level', icon: Star, color: 'text-yellow-500', borderColor: 'border-l-yellow-500' },
  { key: 'xp', label: 'XP', icon: Zap, color: 'text-purple-500', borderColor: 'border-l-purple-500' },
  { key: 'streak', label: 'Streak', icon: Flame, color: 'text-orange-500', borderColor: 'border-l-orange-500' },
  { key: 'winRate', label: 'Win Rate', icon: Target, color: 'text-cyan-500', borderColor: 'border-l-cyan-500' },
] as const

export default function ProfilePage() {
  const router = useRouter()
  const { user, isAuthenticated, fetchProfile } = useAuth()
  const { toast } = useToast()
  const [editingName, setEditingName] = useState(false)
  const [displayName, setDisplayName] = useState('')

  const { data: betsResponse } = useSWR(
    isAuthenticated ? '/api/v1/bets?page_size=20' : null,
    (url: string) => apiFetchPaginated<Bet>(url)
  )

  const { data: txResponse } = useSWR(
    isAuthenticated ? '/api/v1/users/me/transactions?page_size=20' : null,
    (url: string) => apiFetchPaginated<Transaction>(url)
  )

  // Redirect if not authenticated
  if (!isAuthenticated) {
    router.push('/auth')
    return null
  }

  if (!user) {
    return (
      <div className="flex min-h-[calc(100vh-3.5rem)] items-center justify-center">
        <div className="text-muted-foreground">Loading profile...</div>
      </div>
    )
  }

  const bets = betsResponse?.data || []
  const transactions = txResponse?.data || []

  async function handleSaveName() {
    if (!displayName.trim()) return
    try {
      await apiPatch('/api/v1/users/me', { display_name: displayName.trim() })
      await fetchProfile()
      setEditingName(false)
      toast({ title: 'Display name updated' })
    } catch {
      toast({ title: 'Failed to update display name', variant: 'destructive' })
    }
  }

  const winRate = user.total_bets > 0
    ? Math.round((user.total_wins / user.total_bets) * 100)
    : 0

  function getStatValue(key: string) {
    switch (key) {
      case 'balance': return formatCredits(user!.balance)
      case 'frozen': return formatCredits(user!.frozen_balance)
      case 'level': return user!.level
      case 'xp': return user!.xp.toLocaleString()
      case 'streak': return `${user!.current_streak}/${user!.max_streak}`
      case 'winRate': return `${winRate}%`
      default: return ''
    }
  }

  return (
    <div className="mx-auto max-w-4xl px-4 py-8 sm:px-6 lg:px-8">
      {/* Profile Header */}
      <div className="mb-8">
        <div className="flex items-center gap-3">
          {editingName ? (
            <div className="flex items-center gap-2">
              <Input
                value={displayName}
                onChange={(e) => setDisplayName(e.target.value)}
                className="max-w-xs border-border/50"
                autoFocus
              />
              <Button size="icon" variant="ghost" onClick={handleSaveName}>
                <Check className="h-4 w-4" />
              </Button>
              <Button size="icon" variant="ghost" onClick={() => setEditingName(false)}>
                <X className="h-4 w-4" />
              </Button>
            </div>
          ) : (
            <div className="flex items-center gap-2">
              <h1 className="text-2xl font-bold tracking-tight">{user.display_name}</h1>
              <Button
                size="icon"
                variant="ghost"
                className="h-8 w-8 text-muted-foreground hover:text-foreground"
                onClick={() => {
                  setDisplayName(user.display_name)
                  setEditingName(true)
                }}
              >
                <Pencil className="h-4 w-4" />
              </Button>
            </div>
          )}
        </div>
        <p className="mt-1 text-sm text-muted-foreground">
          Level {user.level} Predictor
        </p>
      </div>

      {/* Stats Grid */}
      <div className="mb-6 grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-6">
        {STAT_CARDS.map((stat) => {
          const Icon = stat.icon
          return (
            <Card key={stat.key} className={`border-border/50 border-l-2 ${stat.borderColor}`}>
              <CardContent className="p-4">
                <div className="flex items-center gap-2 text-xs text-muted-foreground">
                  <Icon className={`h-3.5 w-3.5 ${stat.color}`} />
                  {stat.label}
                </div>
                <p className="mt-1 text-lg font-bold">{getStatValue(stat.key)}</p>
              </CardContent>
            </Card>
          )
        })}
      </div>

      {/* XP Progress */}
      <Card className="mb-6 border-border/50">
        <CardContent className="p-4">
          <div className="flex items-center justify-between text-sm">
            <span className="font-medium">Level {user.level}</span>
            <span className="text-muted-foreground">
              {user.xp.toLocaleString()} XP
            </span>
          </div>
          <div className="mt-2 h-2.5 w-full overflow-hidden rounded-full bg-muted">
            <div
              className="h-full rounded-full bg-gradient-to-r from-primary to-primary/70 transition-all"
              style={{ width: `${(user.xp % 1000) / 10}%` }}
            />
          </div>
          <p className="mt-1.5 text-xs text-muted-foreground">
            {1000 - (user.xp % 1000)} XP to next level
          </p>
        </CardContent>
      </Card>

      {/* Tabs: Bets and Transactions */}
      <Tabs defaultValue="bets">
        <TabsList>
          <TabsTrigger value="bets" className="gap-1.5">
            <Target className="h-4 w-4" />
            Recent Bets
          </TabsTrigger>
          <TabsTrigger value="transactions" className="gap-1.5">
            <Wallet className="h-4 w-4" />
            Transactions
          </TabsTrigger>
        </TabsList>

        <TabsContent value="bets">
          <Card className="border-border/50">
            <CardHeader>
              <CardTitle className="text-lg">Your Bets</CardTitle>
            </CardHeader>
            <CardContent>
              {bets.length === 0 ? (
                <div className="flex flex-col items-center justify-center py-12 text-center">
                  <Target className="mb-3 h-8 w-8 text-muted-foreground/50" />
                  <p className="text-sm text-muted-foreground">
                    No bets placed yet. Start predicting!
                  </p>
                </div>
              ) : (
                <div className="space-y-2">
                  {bets.map((bet) => (
                    <div
                      key={bet.id}
                      className="flex flex-col gap-2 rounded-lg border border-border/50 p-4 transition-colors hover:bg-muted/30 sm:flex-row sm:items-center sm:justify-between"
                    >
                      <div className="min-w-0 flex-1">
                        <p className="truncate text-sm font-medium">
                          {bet.event_question}
                        </p>
                        <div className="mt-1.5 flex items-center gap-2">
                          <Badge
                            variant="outline"
                            className={
                              bet.outcome === 'yes'
                                ? 'text-green-600 border-green-500/30 bg-green-500/5'
                                : 'text-red-600 border-red-500/30 bg-red-500/5'
                            }
                          >
                            {bet.outcome.toUpperCase()}
                          </Badge>
                          <Badge variant="outline" className={betStatusColor(bet.status)}>
                            {bet.status}
                          </Badge>
                          <span className="text-xs text-muted-foreground">
                            {format(new Date(bet.created_at), 'MMM d, yyyy')}
                          </span>
                        </div>
                      </div>
                      <div className="text-right">
                        <p className="text-sm font-semibold">
                          {formatCredits(bet.amount)} credits
                        </p>
                        <p className="text-xs text-muted-foreground">
                          Payout: {formatCredits(bet.potential_payout)}
                        </p>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="transactions">
          <Card className="border-border/50">
            <CardHeader>
              <CardTitle className="text-lg">Transaction History</CardTitle>
            </CardHeader>
            <CardContent>
              {transactions.length === 0 ? (
                <div className="flex flex-col items-center justify-center py-12 text-center">
                  <Wallet className="mb-3 h-8 w-8 text-muted-foreground/50" />
                  <p className="text-sm text-muted-foreground">
                    No transactions yet.
                  </p>
                </div>
              ) : (
                <div className="space-y-1">
                  {transactions.map((tx) => (
                    <div
                      key={tx.id}
                      className="flex items-center justify-between rounded-lg border border-border/50 p-4 transition-colors hover:bg-muted/30"
                    >
                      <div>
                        <p className="text-sm font-medium">{tx.description}</p>
                        <p className="text-xs text-muted-foreground">
                          {format(new Date(tx.created_at), 'MMM d, yyyy h:mm a')}
                        </p>
                      </div>
                      <span
                        className={`text-sm font-semibold ${
                          tx.amount >= 0
                            ? 'text-green-600 dark:text-green-400'
                            : 'text-red-600 dark:text-red-400'
                        }`}
                      >
                        {tx.amount >= 0 ? '+' : ''}
                        {formatCredits(tx.amount)}
                      </span>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}
