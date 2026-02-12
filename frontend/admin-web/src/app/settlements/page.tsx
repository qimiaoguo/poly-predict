'use client'

import { useState } from 'react'
import useSWR from 'swr'
import { format } from 'date-fns'
import { adminFetchPaginated } from '@/lib/api/client'
import { useAdminAuth } from '@/hooks/use-admin-auth'
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

interface Settlement {
  id: string
  event_id: string
  event_question: string
  resolved_outcome: string
  total_bets: number
  total_payouts: number
  settled_at: string
}

interface PaginatedSettlements {
  data: Settlement[]
  pagination: {
    total: number
    page: number
    page_size: number
    pages: number
  }
}

function formatCredits(value: number): string {
  return value.toLocaleString() + ' credits'
}

export default function SettlementsPage() {
  useAdminAuth()

  const [page, setPage] = useState(1)

  const queryParams = new URLSearchParams({
    page: page.toString(),
    page_size: '20',
  })

  const { data, error, isLoading } = useSWR<PaginatedSettlements>(
    `/api/v1/settlements?${queryParams.toString()}`,
    adminFetchPaginated<Settlement>
  )

  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold">Settlements</h1>

      <Card>
        <CardHeader>
          <CardTitle>
            Settlement History
            {data?.pagination && (
              <span className="ml-2 text-sm font-normal text-muted-foreground">
                ({data.pagination.total} total)
              </span>
            )}
          </CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <p className="py-8 text-center text-muted-foreground">Loading settlements...</p>
          ) : error ? (
            <p className="py-8 text-center text-destructive">Failed to load settlements.</p>
          ) : (
            <>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Event ID</TableHead>
                    <TableHead>Event</TableHead>
                    <TableHead>Outcome</TableHead>
                    <TableHead className="text-right">Total Bets</TableHead>
                    <TableHead className="text-right">Total Payouts</TableHead>
                    <TableHead className="text-right">Settled At</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {data?.data && data.data.length > 0 ? (
                    data.data.map((settlement) => (
                      <TableRow key={settlement.id}>
                        <TableCell className="font-mono text-xs">
                          {settlement.event_id.slice(0, 8)}...
                        </TableCell>
                        <TableCell className="max-w-[200px] truncate">
                          {settlement.event_question}
                        </TableCell>
                        <TableCell>
                          <Badge
                            variant={
                              settlement.resolved_outcome === 'yes'
                                ? 'default'
                                : 'secondary'
                            }
                          >
                            {settlement.resolved_outcome.toUpperCase()}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-right">
                          {settlement.total_bets}
                        </TableCell>
                        <TableCell className="text-right">
                          {formatCredits(settlement.total_payouts)}
                        </TableCell>
                        <TableCell className="text-right text-muted-foreground">
                          {format(new Date(settlement.settled_at), 'MMM d, yyyy HH:mm')}
                        </TableCell>
                      </TableRow>
                    ))
                  ) : (
                    <TableRow>
                      <TableCell colSpan={6} className="text-center text-muted-foreground">
                        No settlements found
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
    </div>
  )
}
