'use client'

import { use } from 'react'
import useSWR from 'swr'
import { format, formatDistanceToNow } from 'date-fns'
import { ArrowLeft, Calendar, TrendingUp, BarChart3 } from 'lucide-react'
import Link from 'next/link'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { PriceChart } from '@/components/price-chart'
import { BetPanel } from '@/components/bet-panel'
import { apiGet } from '@/lib/api/client'

interface EventDetail {
  id: string
  question: string
  description: string
  category: string
  status: 'open' | 'closed' | 'resolved'
  outcome_prices: string[]
  volume_24h: number
  volume: number
  end_date: string
  created_at: string
  resolved_outcome?: string
}

function statusColor(status: string) {
  switch (status) {
    case 'open':
      return 'bg-green-500/10 text-green-700 dark:text-green-400 border-green-500/20'
    case 'closed':
      return 'bg-yellow-500/10 text-yellow-700 dark:text-yellow-400 border-yellow-500/20'
    case 'resolved':
      return 'bg-gray-500/10 text-gray-700 dark:text-gray-400 border-gray-500/20'
    default:
      return ''
  }
}

function formatCredits(amount: number): string {
  return amount.toLocaleString()
}

export default function EventDetailPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = use(params)

  const { data: event, isLoading, error } = useSWR<EventDetail>(
    `/api/v1/events/${id}`,
    (url: string) => apiGet<EventDetail>(url),
    { refreshInterval: 10000 }
  )

  if (isLoading) {
    return (
      <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
        <div className="space-y-4">
          <div className="h-8 w-64 rounded-lg bg-muted shimmer" />
          <div className="h-4 w-full max-w-xl rounded-lg bg-muted shimmer" />
          <div className="h-[400px] rounded-xl bg-card border border-border/50 shimmer" />
        </div>
      </div>
    )
  }

  if (error || !event) {
    return (
      <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
        <div className="flex flex-col items-center justify-center py-20 text-center">
          <div className="mb-4 rounded-full bg-muted p-4">
            <BarChart3 className="h-8 w-8 text-muted-foreground" />
          </div>
          <h2 className="text-xl font-semibold">Event not found</h2>
          <p className="mt-2 text-muted-foreground">
            This event may have been removed or doesn&apos;t exist.
          </p>
          <Button asChild className="mt-6">
            <Link href="/">Back to Events</Link>
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
      {/* Back Link */}
      <Link
        href="/"
        className="mb-6 inline-flex items-center gap-1.5 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
      >
        <ArrowLeft className="h-4 w-4" />
        Back to Events
      </Link>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        {/* Main Content */}
        <div className="space-y-6 lg:col-span-2">
          {/* Event Header */}
          <div>
            <div className="mb-3 flex flex-wrap items-center gap-2">
              <Badge variant="secondary" className="font-medium">{event.category}</Badge>
              <Badge variant="outline" className={statusColor(event.status)}>
                {event.status}
              </Badge>
              {event.resolved_outcome && (
                <Badge className="bg-blue-500/10 text-blue-700 dark:text-blue-400 border-blue-500/20">
                  Resolved: {event.resolved_outcome.toUpperCase()}
                </Badge>
              )}
            </div>
            <h1 className="text-2xl font-bold tracking-tight sm:text-3xl">{event.question}</h1>
            {event.description && (
              <p className="mt-3 leading-relaxed text-muted-foreground">{event.description}</p>
            )}
          </div>

          {/* Price Chart */}
          <PriceChart eventId={id} />

          {/* Event Metadata */}
          <Card className="border-border/50">
            <CardHeader>
              <CardTitle className="text-lg">Event Details</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
                <div className="space-y-1.5">
                  <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                    <div className="rounded-md bg-muted p-1">
                      <TrendingUp className="h-3 w-3" />
                    </div>
                    24h Volume
                  </div>
                  <p className="text-sm font-semibold">
                    {formatCredits(event.volume_24h)} credits
                  </p>
                </div>
                <div className="space-y-1.5">
                  <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                    <div className="rounded-md bg-muted p-1">
                      <BarChart3 className="h-3 w-3" />
                    </div>
                    Total Volume
                  </div>
                  <p className="text-sm font-semibold">
                    {formatCredits(event.volume)} credits
                  </p>
                </div>
                <div className="space-y-1.5">
                  <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                    <div className="rounded-md bg-muted p-1">
                      <Calendar className="h-3 w-3" />
                    </div>
                    End Date
                  </div>
                  <p className="text-sm font-semibold">
                    {format(new Date(event.end_date), 'MMM d, yyyy')}
                  </p>
                  <p className="text-xs text-muted-foreground">
                    {formatDistanceToNow(new Date(event.end_date), { addSuffix: true })}
                  </p>
                </div>
                <div className="space-y-1.5">
                  <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                    <div className="rounded-md bg-muted p-1">
                      <Calendar className="h-3 w-3" />
                    </div>
                    Created
                  </div>
                  <p className="text-sm font-semibold">
                    {format(new Date(event.created_at), 'MMM d, yyyy')}
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Sidebar - Bet Panel */}
        <div className="lg:sticky lg:top-20 lg:self-start">
          <BetPanel
            eventId={id}
            yesPrice={parseFloat(event.outcome_prices?.[0] ?? '0')}
            noPrice={parseFloat(event.outcome_prices?.[1] ?? '0')}
            status={event.status}
          />
        </div>
      </div>
    </div>
  )
}
