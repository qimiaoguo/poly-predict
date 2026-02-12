'use client'

import { usePathname } from 'next/navigation'
import { Sidebar } from '@/components/sidebar'

const PUBLIC_ROUTES = ['/login']

export function LayoutShell({ children }: { children: React.ReactNode }) {
  const pathname = usePathname()
  const isPublicRoute = PUBLIC_ROUTES.some((route) => pathname.startsWith(route))

  if (isPublicRoute) {
    return <>{children}</>
  }

  return (
    <div className="flex min-h-screen">
      <Sidebar />
      {/* Main content area: offset by sidebar width on desktop, offset by mobile header on mobile */}
      <main className="flex-1 pt-14 lg:ml-64 lg:pt-0">
        <div className="p-6">{children}</div>
      </main>
    </div>
  )
}
