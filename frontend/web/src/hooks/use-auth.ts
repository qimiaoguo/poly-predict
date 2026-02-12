'use client'
import { useEffect } from 'react'
import { supabase } from '@/lib/supabase'
import { useAuthStore } from '@/lib/store'
import { apiGet } from '@/lib/api/client'

export function useAuth() {
  const { user, isAuthenticated, setUser } = useAuthStore()

  useEffect(() => {
    supabase.auth.getSession().then(({ data: { session } }) => {
      if (session) {
        fetchProfile()
      }
    })

    const { data: { subscription } } = supabase.auth.onAuthStateChange((_event, session) => {
      if (session) {
        fetchProfile()
      } else {
        setUser(null)
      }
    })

    return () => subscription.unsubscribe()
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  async function fetchProfile() {
    try {
      const profile = await apiGet<{
        id: string
        display_name: string
        avatar_url: string | null
        balance: number
        frozen_balance: number
        level: number
        xp: number
        current_streak: number
        max_streak: number
        total_bets: number
        total_wins: number
      }>('/api/v1/users/me')
      setUser(profile)
    } catch {
      // User might not exist yet
    }
  }

  async function signInWithEmail(email: string, password: string) {
    const { error } = await supabase.auth.signInWithPassword({ email, password })
    if (error) throw error
  }

  async function signUpWithEmail(email: string, password: string) {
    const { error } = await supabase.auth.signUp({ email, password })
    if (error) throw error
  }

  async function signOut() {
    await supabase.auth.signOut()
    setUser(null)
  }

  return { user, isAuthenticated, signInWithEmail, signUpWithEmail, signOut, fetchProfile }
}
