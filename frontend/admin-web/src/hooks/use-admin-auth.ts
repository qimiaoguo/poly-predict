'use client'
import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useAdminStore } from '@/lib/store'
import { adminPost, setToken, clearToken } from '@/lib/api/client'

export function useAdminAuth(requireAuth = true) {
  const { admin, isAuthenticated, setAdmin, logout } = useAdminStore()
  const router = useRouter()

  useEffect(() => {
    if (requireAuth && !isAuthenticated) {
      const token = typeof window !== 'undefined' ? localStorage.getItem('admin_token') : null
      if (!token) {
        router.push('/login')
      }
    }
  }, [isAuthenticated, requireAuth, router])

  async function login(email: string, password: string) {
    const result = await adminPost<{ token: string; admin: { id: string; email: string; role: string } }>('/api/v1/auth/login', { email, password })
    setToken(result.token)
    setAdmin(result.admin)
    router.push('/dashboard')
  }

  function handleLogout() {
    clearToken()
    logout()
    router.push('/login')
  }

  return { admin, isAuthenticated, login, logout: handleLogout }
}
