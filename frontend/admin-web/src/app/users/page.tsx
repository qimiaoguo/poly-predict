'use client'

import { useState, useCallback } from 'react'
import useSWR from 'swr'
import { Search } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import { adminFetchPaginated, adminPost } from '@/lib/api/client'
import { useAdminAuth } from '@/hooks/use-admin-auth'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
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
import { Label } from '@/components/ui/label'

interface User {
  id: string
  display_name: string
  email: string
  balance: number
  frozen_balance: number
  total_bets: number
  total_wins: number
  created_at: string
}

interface PaginatedUsers {
  data: User[]
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

export default function UsersPage() {
  useAdminAuth()

  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedUser, setSelectedUser] = useState<User | null>(null)
  const [adjustAmount, setAdjustAmount] = useState('')
  const [adjustLoading, setAdjustLoading] = useState(false)
  const [adjustError, setAdjustError] = useState('')

  const queryParams = new URLSearchParams({
    page: page.toString(),
    page_size: '20',
    ...(searchQuery && { search: searchQuery }),
  })

  const { data, error, isLoading, mutate } = useSWR<PaginatedUsers>(
    `/api/v1/users?${queryParams.toString()}`,
    adminFetchPaginated<User>
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

  async function handleAdjustBalance() {
    if (!selectedUser || !adjustAmount) return
    setAdjustLoading(true)
    setAdjustError('')

    try {
      const amount = Math.round(parseFloat(adjustAmount))
      await adminPost(`/api/v1/users/${selectedUser.id}/balance`, { amount })
      setSelectedUser(null)
      setAdjustAmount('')
      mutate()
    } catch (err) {
      setAdjustError(err instanceof Error ? err.message : 'Failed to adjust balance')
    } finally {
      setAdjustLoading(false)
    }
  }

  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold">Users</h1>

      {/* Search */}
      <div className="flex gap-2">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Search users..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            onKeyDown={handleKeyDown}
            className="pl-9"
          />
        </div>
        <Button onClick={handleSearch} variant="secondary">
          Search
        </Button>
      </div>

      {/* Users table */}
      <Card>
        <CardHeader>
          <CardTitle>
            All Users
            {data?.pagination && (
              <span className="ml-2 text-sm font-normal text-muted-foreground">
                ({data.pagination.total} total)
              </span>
            )}
          </CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <p className="py-8 text-center text-muted-foreground">Loading users...</p>
          ) : error ? (
            <p className="py-8 text-center text-destructive">Failed to load users.</p>
          ) : (
            <>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Display Name</TableHead>
                    <TableHead className="text-right">Balance</TableHead>
                    <TableHead className="text-right">Frozen</TableHead>
                    <TableHead className="text-right">Total Bets</TableHead>
                    <TableHead className="text-right">Total Wins</TableHead>
                    <TableHead className="text-right">Joined</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {data?.data && data.data.length > 0 ? (
                    data.data.map((user) => (
                      <TableRow
                        key={user.id}
                        className="cursor-pointer"
                        onClick={() => {
                          setSelectedUser(user)
                          setAdjustAmount('')
                          setAdjustError('')
                        }}
                      >
                        <TableCell className="font-medium">
                          {user.display_name}
                        </TableCell>
                        <TableCell className="text-right">
                          {formatCredits(user.balance)}
                        </TableCell>
                        <TableCell className="text-right">
                          {formatCredits(user.frozen_balance)}
                        </TableCell>
                        <TableCell className="text-right">
                          {user.total_bets}
                        </TableCell>
                        <TableCell className="text-right">
                          {user.total_wins}
                        </TableCell>
                        <TableCell className="text-right text-muted-foreground">
                          {formatDistanceToNow(new Date(user.created_at), { addSuffix: true })}
                        </TableCell>
                      </TableRow>
                    ))
                  ) : (
                    <TableRow>
                      <TableCell colSpan={6} className="text-center text-muted-foreground">
                        No users found
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

      {/* User detail / Balance adjustment dialog */}
      <Dialog
        open={!!selectedUser}
        onOpenChange={(open) => {
          if (!open) setSelectedUser(null)
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{selectedUser?.display_name}</DialogTitle>
            <DialogDescription>
              User ID: {selectedUser?.id}
            </DialogDescription>
          </DialogHeader>

          <div className="grid gap-3 py-2 text-sm">
            <div className="flex justify-between">
              <span className="text-muted-foreground">Email</span>
              <span>{selectedUser?.email}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">Balance</span>
              <span>{formatCredits(selectedUser?.balance ?? 0)}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">Frozen Balance</span>
              <span>{formatCredits(selectedUser?.frozen_balance ?? 0)}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">Total Bets</span>
              <span>{selectedUser?.total_bets}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">Total Wins</span>
              <span>{selectedUser?.total_wins}</span>
            </div>
          </div>

          <div className="space-y-2 border-t pt-4">
            <Label htmlFor="adjust-amount">Adjust Balance (credits)</Label>
            <p className="text-xs text-muted-foreground">
              Enter a positive number to add credits, or a negative number to subtract.
            </p>
            <Input
              id="adjust-amount"
              type="number"
              step="0.01"
              placeholder="e.g. 100 or -50"
              value={adjustAmount}
              onChange={(e) => setAdjustAmount(e.target.value)}
            />
            {adjustError && (
              <p className="text-sm text-destructive">{adjustError}</p>
            )}
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setSelectedUser(null)}>
              Cancel
            </Button>
            <Button
              onClick={handleAdjustBalance}
              disabled={!adjustAmount || adjustLoading}
            >
              {adjustLoading ? 'Adjusting...' : 'Adjust Balance'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
