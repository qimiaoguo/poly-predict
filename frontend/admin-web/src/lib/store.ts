import { create } from 'zustand'

interface AdminUser {
  id: string
  email: string
  role: string
}

interface AdminAuthState {
  admin: AdminUser | null
  isAuthenticated: boolean
  setAdmin: (admin: AdminUser | null) => void
  logout: () => void
}

export const useAdminStore = create<AdminAuthState>((set) => ({
  admin: null,
  isAuthenticated: false,
  setAdmin: (admin) => set({ admin, isAuthenticated: !!admin }),
  logout: () => {
    if (typeof window !== 'undefined') localStorage.removeItem('admin_token')
    set({ admin: null, isAuthenticated: false })
  },
}))
