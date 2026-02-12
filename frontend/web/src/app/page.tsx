'use client'

import { useState } from 'react'
import useSWR from 'swr'
import { Search, SlidersHorizontal, BarChart3 } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { EventCard } from '@/components/event-card'
import { apiGet, apiFetchPaginated } from '@/lib/api/client'

interface Event {
  id: string
  question: string
  category: string
  status: 'open' | 'closed' | 'resolved'
  outcome_prices: string[]
  volume_24h: number
  end_date: string
}

interface Category {
  id: string
  name: string
  slug: string
}

const SORT_OPTIONS = [
  { value: 'trending', label: 'Trending' },
  { value: 'volume', label: 'Volume' },
  { value: 'ending_soon', label: 'Ending Soon' },
]

export default function HomePage() {
  const [category, setCategory] = useState('all')
  const [search, setSearch] = useState('')
  const [sort, setSort] = useState('trending')
  const [page, setPage] = useState(1)

  const { data: categories } = useSWR<Category[]>(
    '/api/v1/categories',
    (url: string) => apiGet<Category[]>(url)
  )

  const queryParams = new URLSearchParams()
  if (category !== 'all') queryParams.set('category', category)
  if (search) queryParams.set('q', search)
  queryParams.set('sort', sort)
  queryParams.set('page', page.toString())
  queryParams.set('page_size', '12')

  const { data: eventsResponse, isLoading } = useSWR(
    `/api/v1/events?${queryParams.toString()}`,
    (url: string) => apiFetchPaginated<Event>(url),
    { refreshInterval: 15000 }
  )

  const events = eventsResponse?.data || []
  const pagination = eventsResponse?.pagination

  return (
    <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold tracking-tight sm:text-4xl">
          Prediction Markets
        </h1>
        <p className="mt-2 text-muted-foreground">
          Bet on the outcome of real-world events with virtual credits
        </p>
      </div>

      {/* Category Tabs */}
      <Tabs
        value={category}
        onValueChange={(val) => {
          setCategory(val)
          setPage(1)
        }}
        className="mb-5"
      >
        <TabsList className="flex-wrap">
          <TabsTrigger value="all">All</TabsTrigger>
          {categories?.map((cat) => (
            <TabsTrigger key={cat.id} value={cat.slug}>
              {cat.name}
            </TabsTrigger>
          ))}
        </TabsList>
      </Tabs>

      {/* Search and Sort */}
      <div className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Search events..."
            value={search}
            onChange={(e) => {
              setSearch(e.target.value)
              setPage(1)
            }}
            className="pl-9 bg-card border-border/50 focus-visible:ring-primary/30"
          />
        </div>
        <div className="flex items-center gap-2">
          <SlidersHorizontal className="h-4 w-4 text-muted-foreground" />
          <Select
            value={sort}
            onValueChange={(val) => {
              setSort(val)
              setPage(1)
            }}
          >
            <SelectTrigger className="w-[160px] border-border/50">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {SORT_OPTIONS.map((opt) => (
                <SelectItem key={opt.value} value={opt.value}>
                  {opt.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>

      {/* Events Grid */}
      {isLoading ? (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 6 }).map((_, i) => (
            <div
              key={i}
              className="h-64 rounded-xl border border-border/50 bg-card shimmer"
            />
          ))}
        </div>
      ) : events.length === 0 ? (
        <div className="flex h-64 flex-col items-center justify-center gap-3 rounded-xl border border-dashed border-border/50">
          <BarChart3 className="h-10 w-10 text-muted-foreground/50" />
          <p className="text-muted-foreground">
            No events found. Try adjusting your filters.
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
          {events.map((event) => (
            <EventCard key={event.id} event={event} />
          ))}
        </div>
      )}

      {/* Pagination */}
      {pagination && pagination.pages > 1 && (
        <div className="mt-8 flex items-center justify-center gap-2">
          <Button
            variant="outline"
            size="sm"
            disabled={page <= 1}
            onClick={() => setPage(page - 1)}
            className="border-border/50"
          >
            Previous
          </Button>
          <span className="px-3 text-sm text-muted-foreground">
            Page {pagination.page} of {pagination.pages}
          </span>
          <Button
            variant="outline"
            size="sm"
            disabled={page >= pagination.pages}
            onClick={() => setPage(page + 1)}
            className="border-border/50"
          >
            Next
          </Button>
        </div>
      )}
    </div>
  )
}
