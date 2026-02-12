'use client'

import Link from 'next/link'
import { formatDistanceToNow } from 'date-fns'
import { TrendingUp, Clock } from 'lucide-react'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'

interface Event {
  id: string
  question: string
  category: string
  status: 'open' | 'closed' | 'resolved'
  outcome_prices: string[]
  volume_24h: number
  end_date: string
}

interface EventCardProps {
  event: Event
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
  return (amount / 100).toLocaleString()
}

export function EventCard({ event }: EventCardProps) {
  const yesPercent = Math.round(parseFloat(event.outcome_prices?.[0] ?? '0') * 100)
  const noPercent = Math.round(parseFloat(event.outcome_prices?.[1] ?? '0') * 100)

  return (
    <Link href={`/events/${event.id}`} className="group cursor-pointer">
      <Card className="h-full border-border/50 transition-all duration-200 hover:shadow-lg hover:border-primary/20 group-hover:shadow-primary/5">
        <CardHeader className="pb-3">
          <div className="flex items-start justify-between gap-2">
            <h3 className="line-clamp-2 text-sm font-semibold leading-snug tracking-tight">
              {event.question}
            </h3>
          </div>
          <div className="flex items-center gap-2 pt-1">
            <Badge variant="secondary" className="text-xs font-medium">
              {event.category}
            </Badge>
            <Badge
              variant="outline"
              className={statusColor(event.status)}
            >
              {event.status}
            </Badge>
          </div>
        </CardHeader>
        <CardContent className="space-y-3">
          {/* Price bars */}
          <div className="space-y-2.5">
            <div className="space-y-1.5">
              <div className="flex items-center justify-between text-xs">
                <span className="font-semibold text-green-600 dark:text-green-400">Yes</span>
                <span className="font-mono font-semibold">{yesPercent}%</span>
              </div>
              <div className="h-2.5 w-full overflow-hidden rounded-full bg-muted">
                <div
                  className="h-full rounded-full bg-green-500 transition-all duration-500"
                  style={{ width: `${yesPercent}%` }}
                />
              </div>
            </div>
            <div className="space-y-1.5">
              <div className="flex items-center justify-between text-xs">
                <span className="font-semibold text-red-600 dark:text-red-400">No</span>
                <span className="font-mono font-semibold">{noPercent}%</span>
              </div>
              <div className="h-2.5 w-full overflow-hidden rounded-full bg-muted">
                <div
                  className="h-full rounded-full bg-red-500 transition-all duration-500"
                  style={{ width: `${noPercent}%` }}
                />
              </div>
            </div>
          </div>

          {/* Metadata */}
          <div className="flex items-center justify-between border-t border-border/50 pt-3 text-xs text-muted-foreground">
            <div className="flex items-center gap-1">
              <TrendingUp className="h-3 w-3" />
              <span>{formatCredits(event.volume_24h)} vol</span>
            </div>
            <div className="flex items-center gap-1">
              <Clock className="h-3 w-3" />
              <span>
                {formatDistanceToNow(new Date(event.end_date), { addSuffix: true })}
              </span>
            </div>
          </div>
        </CardContent>
      </Card>
    </Link>
  )
}
