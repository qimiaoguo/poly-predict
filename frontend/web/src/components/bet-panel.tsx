'use client'

import { useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { apiPost } from '@/lib/api/client'
import { useAuthStore } from '@/lib/store'
import { useToast } from '@/hooks/use-toast'

interface BetPanelProps {
  eventId: string
  yesPrice: number
  noPrice: number
  status: string
}

const QUICK_AMOUNTS = [100, 500, 1000, 5000]

export function BetPanel({ eventId, yesPrice, noPrice, status }: BetPanelProps) {
  const [outcome, setOutcome] = useState<'yes' | 'no'>('yes')
  const [amount, setAmount] = useState<string>('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const { user, isAuthenticated, updateBalance } = useAuthStore()
  const { toast } = useToast()

  const currentPrice = outcome === 'yes' ? yesPrice : noPrice
  const amountNum = parseInt(amount) || 0
  const potentialPayout = amountNum > 0 && currentPrice > 0
    ? Math.round(amountNum / currentPrice)
    : 0
  const potentialProfit = potentialPayout - amountNum

  async function handlePlaceBet() {
    if (!isAuthenticated) {
      setError('Please sign in to place a bet')
      return
    }
    if (status !== 'open') {
      setError('This event is not open for betting')
      return
    }
    if (amountNum <= 0) {
      setError('Please enter a valid amount')
      return
    }
    if (user && amountNum > user.balance) {
      setError('Insufficient balance')
      return
    }

    setError(null)
    setLoading(true)

    try {
      const result = await apiPost<{
        balance: number
        frozen_balance: number
      }>('/api/v1/bets', {
        event_id: eventId,
        outcome,
        amount: amountNum,
      })
      updateBalance(result.balance, result.frozen_balance)
      setAmount('')
      toast({
        title: 'Bet placed!',
        description: `You bet ${(amountNum / 100).toLocaleString()} credits on ${outcome.toUpperCase()}`,
      })
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to place bet'
      setError(message)
    } finally {
      setLoading(false)
    }
  }

  const isReady = isAuthenticated && status === 'open' && amountNum > 0

  return (
    <Card className="border-border/50 transition-all duration-200 hover:shadow-lg hover:shadow-primary/5">
      <CardHeader className="pb-4">
        <CardTitle className="text-lg">Place a Bet</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Outcome Toggle */}
        <div className="grid grid-cols-2 gap-2">
          <button
            className={`flex items-center justify-center rounded-lg border-2 px-4 py-2.5 text-sm font-semibold transition-all ${
              outcome === 'yes'
                ? 'border-green-500 bg-green-500/10 text-green-600 dark:text-green-400 glow-success'
                : 'border-border/50 text-muted-foreground hover:border-green-500/50 hover:text-green-600 dark:hover:text-green-400'
            }`}
            onClick={() => setOutcome('yes')}
          >
            Yes {Math.round(yesPrice * 100)}%
          </button>
          <button
            className={`flex items-center justify-center rounded-lg border-2 px-4 py-2.5 text-sm font-semibold transition-all ${
              outcome === 'no'
                ? 'border-red-500 bg-red-500/10 text-red-600 dark:text-red-400 glow-danger'
                : 'border-border/50 text-muted-foreground hover:border-red-500/50 hover:text-red-600 dark:hover:text-red-400'
            }`}
            onClick={() => setOutcome('no')}
          >
            No {Math.round(noPrice * 100)}%
          </button>
        </div>

        {/* Amount Input */}
        <div className="space-y-2">
          <Label htmlFor="amount" className="text-xs text-muted-foreground">Amount (credits)</Label>
          <div className="relative">
            <Input
              id="amount"
              type="number"
              placeholder="0"
              value={amount}
              onChange={(e) => {
                setAmount(e.target.value)
                setError(null)
              }}
              min={1}
              disabled={!isAuthenticated || status !== 'open'}
              className="border-border/50 pr-16 text-lg font-semibold"
            />
            <span className="absolute right-3 top-1/2 -translate-y-1/2 text-xs text-muted-foreground">
              credits
            </span>
          </div>
        </div>

        {/* Quick Amount Buttons */}
        <div className="grid grid-cols-4 gap-2">
          {QUICK_AMOUNTS.map((quickAmount) => (
            <Button
              key={quickAmount}
              variant="outline"
              size="sm"
              className="border-border/50 text-xs font-medium hover:bg-accent"
              onClick={() => {
                setAmount(quickAmount.toString())
                setError(null)
              }}
              disabled={!isAuthenticated || status !== 'open'}
            >
              {(quickAmount / 100).toLocaleString()}
            </Button>
          ))}
        </div>

        {/* Divider */}
        <div className="border-t border-border/50" />

        {/* Payout Info */}
        <div className="space-y-2.5 text-sm">
          <div className="flex justify-between">
            <span className="text-muted-foreground">Current odds</span>
            <span className="font-semibold">{Math.round(currentPrice * 100)}%</span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted-foreground">Potential payout</span>
            <span className="font-semibold">
              {(potentialPayout / 100).toLocaleString()} credits
            </span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted-foreground">Potential profit</span>
            <span className="font-semibold text-green-600 dark:text-green-400">
              +{(potentialProfit / 100).toLocaleString()} credits
            </span>
          </div>
        </div>

        {/* Error */}
        {error && (
          <p className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">{error}</p>
        )}

        {/* Place Bet Button */}
        <Button
          className={`w-full text-sm font-semibold ${
            isReady
              ? 'bg-green-600 hover:bg-green-700 text-white'
              : ''
          }`}
          size="lg"
          onClick={handlePlaceBet}
          disabled={loading || !isReady}
        >
          {loading
            ? 'Placing bet...'
            : !isAuthenticated
              ? 'Sign in to bet'
              : status !== 'open'
                ? 'Event not open'
                : 'Place Bet'}
        </Button>

        {/* Balance */}
        {isAuthenticated && user && (
          <p className="text-center text-xs text-muted-foreground">
            Your balance: {(user.balance / 100).toLocaleString()} credits
          </p>
        )}
      </CardContent>
    </Card>
  )
}
