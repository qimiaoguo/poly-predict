'use client'

import { useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Separator } from '@/components/ui/separator'
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

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">Place a Bet</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Outcome Toggle */}
        <div className="grid grid-cols-2 gap-2">
          <Button
            variant={outcome === 'yes' ? 'default' : 'outline'}
            className={
              outcome === 'yes'
                ? 'bg-green-600 hover:bg-green-700 text-white'
                : 'border-green-600 text-green-600 hover:bg-green-50 dark:hover:bg-green-950'
            }
            onClick={() => setOutcome('yes')}
          >
            Yes ({Math.round(yesPrice * 100)}%)
          </Button>
          <Button
            variant={outcome === 'no' ? 'default' : 'outline'}
            className={
              outcome === 'no'
                ? 'bg-red-600 hover:bg-red-700 text-white'
                : 'border-red-600 text-red-600 hover:bg-red-50 dark:hover:bg-red-950'
            }
            onClick={() => setOutcome('no')}
          >
            No ({Math.round(noPrice * 100)}%)
          </Button>
        </div>

        {/* Amount Input */}
        <div className="space-y-2">
          <Label htmlFor="amount">Amount (credits)</Label>
          <Input
            id="amount"
            type="number"
            placeholder="Enter amount"
            value={amount}
            onChange={(e) => {
              setAmount(e.target.value)
              setError(null)
            }}
            min={1}
            disabled={!isAuthenticated || status !== 'open'}
          />
        </div>

        {/* Quick Amount Buttons */}
        <div className="grid grid-cols-4 gap-2">
          {QUICK_AMOUNTS.map((quickAmount) => (
            <Button
              key={quickAmount}
              variant="outline"
              size="sm"
              className="text-xs"
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

        <Separator />

        {/* Payout Info */}
        <div className="space-y-2 text-sm">
          <div className="flex justify-between">
            <span className="text-muted-foreground">Current odds</span>
            <span className="font-medium">{Math.round(currentPrice * 100)}%</span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted-foreground">Potential payout</span>
            <span className="font-medium">
              {(potentialPayout / 100).toLocaleString()} credits
            </span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted-foreground">Potential profit</span>
            <span className="font-medium text-green-600 dark:text-green-400">
              +{(potentialProfit / 100).toLocaleString()} credits
            </span>
          </div>
        </div>

        {/* Error */}
        {error && (
          <p className="text-sm text-destructive">{error}</p>
        )}

        {/* Place Bet Button */}
        <Button
          className="w-full"
          size="lg"
          onClick={handlePlaceBet}
          disabled={loading || !isAuthenticated || status !== 'open' || amountNum <= 0}
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
