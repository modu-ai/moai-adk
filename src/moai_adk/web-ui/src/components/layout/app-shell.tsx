import { type ReactNode, useEffect } from 'react'
import { Header } from './header'
import { Sidebar } from './sidebar'
import { useUIStore } from '@/stores'

interface AppShellProps {
  children: ReactNode
  isConnected: boolean
}

export function AppShell({ children, isConnected }: AppShellProps) {
  const { theme } = useUIStore()

  // Apply theme to document
  useEffect(() => {
    const root = window.document.documentElement

    if (theme === 'system') {
      const systemTheme = window.matchMedia('(prefers-color-scheme: dark)')
        .matches
        ? 'dark'
        : 'light'
      root.classList.remove('light', 'dark')
      root.classList.add(systemTheme)
    } else {
      root.classList.remove('light', 'dark')
      root.classList.add(theme)
    }
  }, [theme])

  // Listen for system theme changes
  useEffect(() => {
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    const handleChange = () => {
      const { theme } = useUIStore.getState()
      if (theme === 'system') {
        const root = window.document.documentElement
        root.classList.remove('light', 'dark')
        root.classList.add(mediaQuery.matches ? 'dark' : 'light')
      }
    }

    mediaQuery.addEventListener('change', handleChange)
    return () => mediaQuery.removeEventListener('change', handleChange)
  }, [])

  return (
    <div className="h-screen flex flex-col bg-background text-foreground overflow-hidden">
      <Header isConnected={isConnected} />
      <div className="flex flex-1 overflow-hidden">
        <Sidebar />
        <main className="flex-1 overflow-hidden">{children}</main>
      </div>
    </div>
  )
}
