'use client'

import { ThemeProvider } from 'next-themes'
import { Navbar } from '@/components/navbar'
import { Toaster } from '@/components/ui/toaster'

export function ClientLayout({ children }: { children: React.ReactNode }) {
  return (
    <ThemeProvider attribute="class" defaultTheme="dark" enableSystem>
      <Navbar />
      <main className="min-h-[calc(100vh-3.5rem)]">{children}</main>
      <Toaster />
    </ThemeProvider>
  )
}
