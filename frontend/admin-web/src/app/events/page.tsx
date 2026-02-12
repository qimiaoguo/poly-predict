'use client'

import { useState, useCallback } from 'react'
import useSWR from 'swr'
import { Search } from 'lucide-react'
import { format } from 'date-fns'
import { adminFetchPaginated, adminPost, adminPatch } from '@/lib/api/client'
import { useAdminAuth } from '@/hooks/use-admin-auth'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Label } from '@/components/ui/label'

interface Event {
  id: string
  question: string
  description: string
  category: string
  status: string
  yes_price: number
  no_price: number
  volume: number
  end_date: string
  resolved_outcome: string | null
  created_at: string
}

interface PaginatedEvents {
  data: Event[]
  pagination: {
    total: number
    page: number
    page_size: number
    pages: number
  }
}

function formatCredits(value: number): string {
  return (value / 100).toLocaleString() + ' credits'
}

function statusBadgeVariant(status: string): 'default' | 'secondary' | 'destructive' | 'outline' {
  switch (status.toLowerCase()) {
    case 'open':
      return 'default'
    case 'closed':
      return 'secondary'
    case 'resolved':
      return 'outline'
    default:
      return 'secondary'
  }
}

export default function EventsPage() {
  useAdminAuth()

  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [searchQuery, setSearchQuery] = useState('')
  const [statusFilter, setStatusFilter] = useState('all')
  const [selectedEvent, setSelectedEvent] = useState<Event | null>(null)
  const [settleOutcome, setSettleOutcome] = useState<string>('')
  const [settleLoading, setSettleLoading] = useState(false)
  const [settleError, setSettleError] = useState('')
  const [showSettleDialog, setShowSettleDialog] = useState(false)

  const queryParams = new URLSearchParams({
    page: page.toString(),
    page_size: '20',
    ...(searchQuery && { search: searchQuery }),
    ...(statusFilter !== 'all' && { status: statusFilter }),
  })

  const { data, error, isLoading, mutate } = useSWR<PaginatedEvents>(
    `/api/v1/events?${queryParams.toString()}`,
    adminFetchPaginated<Event>
  )

  const handleSearch = useCallback(() => {
    setPage(1)
    setSearchQuery(search)
  }, [search])

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === 'Enter') handleSearch()
    },
    [handleSearch]
  )

  function handleStatusFilter(value: string) {
    setStatusFilter(value)
    setPage(1)
  }

  async function handleSettle() {
    if (!selectedEvent || !settleOutcome) return
    setSettleLoading(true)
    setSettleError('')

    try {
      await adminPost(`/api/v1/events/${selectedEvent.id}/settle`, {
        outcome: settleOutcome,
      })
      setShowSettleDialog(false)
      setSelectedEvent(null)
      setSettleOutcome('')
      mutate()
    } catch (err) {
      setSettleError(err instanceof Error ? err.message : 'Failed to settle event')
    } finally {
      setSettleLoading(false)
    }
  }

  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold">Events</h1>

      {/* Filters */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Search events..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            onKeyDown={handleKeyDown}
            className="pl-9"
          />
        </div>
        <Button onClick={handleSearch} variant="secondary">
          Search
        </Button>
        <Tabs value={statusFilter} onValueChange={handleStatusFilter}>
          <TabsList>
            <TabsTrigger value="all">All</TabsTrigger>
            <TabsTrigger value="open">Open</TabsTrigger>
            <TabsTrigger value="closed">Closed</TabsTrigger>
            <TabsTrigger value="resolved">Resolved</TabsTrigger>
          </TabsList>
        </Tabs>
      </div>

      {/* Events table */}
      <Card>
        <CardHeader>
          <CardTitle>
            Events
            {data?.pagination && (
              <span className="ml-2 text-sm font-normal text-muted-foreground">
                ({data.pagination.total} total)
              </span>
            )}
          </CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <p className="py-8 text-center text-muted-foreground">Loading events...</p>
          ) : error ? (
            <p className="py-8 text-center text-destructive">Failed to load events.</p>
          ) : (
            <>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[300px]">Question</TableHead>
                    <TableHead>Category</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead className="text-right">Volume</TableHead>
                    <TableHead className="text-right">End Date</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {data?.data && data.data.length > 0 ? (
                    data.data.map((event) => (
                      <TableRow
                        key={event.id}
                        className="cursor-pointer"
                        onClick={() => {
                          setSelectedEvent(event)
                          setSettleOutcome('')
                          setSettleError('')
                        }}
                      >
                        <TableCell className="max-w-[300px] truncate font-medium">
                          {event.question}
                        </TableCell>
                        <TableCell>{event.category}</TableCell>
                        <TableCell>
                          <Badge variant={statusBadgeVariant(event.status)}>
                            {event.status}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-right">
                          {formatCredits(event.volume)}
                        </TableCell>
                        <TableCell className="text-right text-muted-foreground">
                          {format(new Date(event.end_date), 'MMM d, yyyy')}
                        </TableCell>
                      </TableRow>
                    ))
                  ) : (
                    <TableRow>
                      <TableCell colSpan={5} className="text-center text-muted-foreground">
                        No events found
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>

              {/* Pagination */}
              {data?.pagination && data.pagination.pages > 1 && (
                <div className="flex items-center justify-between pt-4">
                  <p className="text-sm text-muted-foreground">
                    Page {data.pagination.page} of {data.pagination.pages}
                  </p>
                  <div className="flex gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setPage((p) => Math.max(1, p - 1))}
                      disabled={page <= 1}
                    >
                      Previous
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setPage((p) => Math.min(data.pagination.pages, p + 1))}
                      disabled={page >= data.pagination.pages}
                    >
                      Next
                    </Button>
                  </div>
                </div>
              )}
            </>
          )}
        </CardContent>
      </Card>

      {/* Event detail dialog */}
      <Dialog
        open={!!selectedEvent && !showSettleDialog}
        onOpenChange={(open) => {
          if (!open) setSelectedEvent(null)
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Event Details</DialogTitle>
            <DialogDescription>{selectedEvent?.id}</DialogDescription>
          </DialogHeader>

          <div className="space-y-3 text-sm">
            <div>
              <span className="font-medium">Question:</span>
              <p className="mt-1">{selectedEvent?.question}</p>
            </div>
            {selectedEvent?.description && (
              <div>
                <span className="font-medium">Description:</span>
                <p className="mt-1 text-muted-foreground">{selectedEvent.description}</p>
              </div>
            )}
            <div className="grid grid-cols-2 gap-3">
              <div>
                <span className="text-muted-foreground">Category</span>
                <p className="font-medium">{selectedEvent?.category}</p>
              </div>
              <div>
                <span className="text-muted-foreground">Status</span>
                <p>
                  <Badge variant={statusBadgeVariant(selectedEvent?.status ?? '')}>
                    {selectedEvent?.status}
                  </Badge>
                </p>
              </div>
              <div>
                <span className="text-muted-foreground">Yes Price</span>
                <p className="font-medium">{selectedEvent?.yes_price}</p>
              </div>
              <div>
                <span className="text-muted-foreground">No Price</span>
                <p className="font-medium">{selectedEvent?.no_price}</p>
              </div>
              <div>
                <span className="text-muted-foreground">Volume</span>
                <p className="font-medium">{formatCredits(selectedEvent?.volume ?? 0)}</p>
              </div>
              <div>
                <span className="text-muted-foreground">End Date</span>
                <p className="font-medium">
                  {selectedEvent?.end_date
                    ? format(new Date(selectedEvent.end_date), 'MMM d, yyyy HH:mm')
                    : '-'}
                </p>
              </div>
              {selectedEvent?.resolved_outcome && (
                <div className="col-span-2">
                  <span className="text-muted-foreground">Resolved Outcome</span>
                  <p className="font-medium">{selectedEvent.resolved_outcome}</p>
                </div>
              )}
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setSelectedEvent(null)}>
              Close
            </Button>
            {selectedEvent?.status !== 'resolved' && (
              <Button
                variant="destructive"
                onClick={() => setShowSettleDialog(true)}
              >
                Force Settle
              </Button>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Force settle dialog */}
      <Dialog
        open={showSettleDialog}
        onOpenChange={(open) => {
          if (!open) {
            setShowSettleDialog(false)
            setSettleOutcome('')
            setSettleError('')
          }
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Force Settle Event</DialogTitle>
            <DialogDescription>
              This will resolve the event and settle all bets. This action cannot be undone.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-3">
            <p className="text-sm font-medium">{selectedEvent?.question}</p>

            <div className="space-y-2">
              <Label>Select Outcome</Label>
              <Select value={settleOutcome} onValueChange={setSettleOutcome}>
                <SelectTrigger>
                  <SelectValue placeholder="Choose outcome..." />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="yes">Yes</SelectItem>
                  <SelectItem value="no">No</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {settleError && (
              <p className="text-sm text-destructive">{settleError}</p>
            )}
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => {
                setShowSettleDialog(false)
                setSettleOutcome('')
                setSettleError('')
              }}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleSettle}
              disabled={!settleOutcome || settleLoading}
            >
              {settleLoading ? 'Settling...' : 'Confirm Settlement'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
