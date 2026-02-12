'use client'

import { useState } from 'react'
import useSWR from 'swr'
import { Search, SlidersHorizontal } from 'lucide-react'
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
    <div className="mx-auto max-w-7xl px-4 py-6 sm:px-6 lg:px-8">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold sm:text-3xl">Prediction Markets</h1>
        <p className="mt-1 text-sm text-muted-foreground">
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
        className="mb-4"
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
            className="pl-9"
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
            <SelectTrigger className="w-[160px]">
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
              className="h-64 animate-pulse rounded-xl border bg-muted"
            />
          ))}
        </div>
      ) : events.length === 0 ? (
        <div className="flex h-64 items-center justify-center text-muted-foreground">
          No events found. Try adjusting your filters.
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
          >
            Previous
          </Button>
          <span className="text-sm text-muted-foreground">
            Page {pagination.page} of {pagination.pages}
          </span>
          <Button
            variant="outline"
            size="sm"
            disabled={page >= pagination.pages}
            onClick={() => setPage(page + 1)}
          >
            Next
          </Button>
        </div>
      )}
    </div>
  )
}
