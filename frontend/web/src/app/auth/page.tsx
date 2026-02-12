'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { TrendingUp } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { useAuth } from '@/hooks/use-auth'

export default function AuthPage() {
  const router = useRouter()
  const { signInWithEmail, signUpWithEmail, isAuthenticated } = useAuth()
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  // Redirect if already authenticated
  if (isAuthenticated) {
    router.push('/')
    return null
  }

  async function handleSignIn(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    setError(null)
    setLoading(true)

    const formData = new FormData(e.currentTarget)
    const email = formData.get('email') as string
    const password = formData.get('password') as string

    try {
      await signInWithEmail(email, password)
      router.push('/')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to sign in')
    } finally {
      setLoading(false)
    }
  }

  async function handleSignUp(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    setError(null)
    setLoading(true)

    const formData = new FormData(e.currentTarget)
    const email = formData.get('email') as string
    const password = formData.get('password') as string
    const confirmPassword = formData.get('confirmPassword') as string

    if (password !== confirmPassword) {
      setError('Passwords do not match')
      setLoading(false)
      return
    }

    if (password.length < 6) {
      setError('Password must be at least 6 characters')
      setLoading(false)
      return
    }

    try {
      await signUpWithEmail(email, password)
      router.push('/')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to sign up')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex min-h-[calc(100vh-3.5rem)] items-center justify-center px-4 py-12">
      <div className="relative w-full max-w-md">
        {/* Background glow */}
        <div className="absolute -inset-1 rounded-2xl bg-gradient-to-r from-primary/20 via-transparent to-primary/20 opacity-50 blur-xl" />

        <Card className="relative border-border/50">
          <CardHeader className="text-center pb-2">
            <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-xl bg-primary/10">
              <TrendingUp className="h-6 w-6 text-primary" />
            </div>
            <CardTitle className="text-2xl tracking-tight">Welcome to Poly-Predict</CardTitle>
            <CardDescription className="text-muted-foreground">
              Sign in or create an account to start predicting
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Tabs defaultValue="signin" className="w-full" onValueChange={() => setError(null)}>
              <TabsList className="grid w-full grid-cols-2">
                <TabsTrigger value="signin">Sign In</TabsTrigger>
                <TabsTrigger value="signup">Sign Up</TabsTrigger>
              </TabsList>

              <TabsContent value="signin">
                <form onSubmit={handleSignIn} className="space-y-4">
                  <div className="space-y-2">
                    <Label htmlFor="signin-email">Email</Label>
                    <Input
                      id="signin-email"
                      name="email"
                      type="email"
                      placeholder="you@example.com"
                      required
                      autoComplete="email"
                      className="border-border/50"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="signin-password">Password</Label>
                    <Input
                      id="signin-password"
                      name="password"
                      type="password"
                      placeholder="Enter your password"
                      required
                      autoComplete="current-password"
                      className="border-border/50"
                    />
                  </div>
                  {error && (
                    <p className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">{error}</p>
                  )}
                  <Button type="submit" className="w-full font-semibold" disabled={loading}>
                    {loading ? 'Signing in...' : 'Sign In'}
                  </Button>
                </form>
              </TabsContent>

              <TabsContent value="signup">
                <form onSubmit={handleSignUp} className="space-y-4">
                  <div className="space-y-2">
                    <Label htmlFor="signup-email">Email</Label>
                    <Input
                      id="signup-email"
                      name="email"
                      type="email"
                      placeholder="you@example.com"
                      required
                      autoComplete="email"
                      className="border-border/50"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="signup-password">Password</Label>
                    <Input
                      id="signup-password"
                      name="password"
                      type="password"
                      placeholder="Create a password"
                      required
                      autoComplete="new-password"
                      minLength={6}
                      className="border-border/50"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="signup-confirm">Confirm Password</Label>
                    <Input
                      id="signup-confirm"
                      name="confirmPassword"
                      type="password"
                      placeholder="Confirm your password"
                      required
                      autoComplete="new-password"
                      minLength={6}
                      className="border-border/50"
                    />
                  </div>
                  {error && (
                    <p className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">{error}</p>
                  )}
                  <Button type="submit" className="w-full font-semibold" disabled={loading}>
                    {loading ? 'Creating account...' : 'Create Account'}
                  </Button>
                </form>
              </TabsContent>
            </Tabs>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
