import { create } from 'zustand'

interface User {
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
}

interface AuthState {
  user: User | null
  isAuthenticated: boolean
  setUser: (user: User | null) => void
  updateBalance: (balance: number, frozenBalance: number) => void
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  isAuthenticated: false,
  setUser: (user) => set({ user, isAuthenticated: !!user }),
  updateBalance: (balance, frozenBalance) =>
    set((state) => ({
      user: state.user ? { ...state.user, balance, frozen_balance: frozenBalance } : null,
    })),
}))
