import {
  Menu,
  Moon,
  Sun,
  Settings,
  Wifi,
  WifiOff,
} from 'lucide-react'
import { Button } from '@/components/ui'
import { useUIStore, useProviderStore } from '@/stores'
import { cn } from '@/lib/utils'

interface HeaderProps {
  isConnected: boolean
}

export function Header({ isConnected }: HeaderProps) {
  const { toggleSidebar, theme, setTheme } = useUIStore()
  const { activeProvider, activeModel } = useProviderStore()

  const toggleTheme = () => {
    if (theme === 'dark') {
      setTheme('light')
    } else {
      setTheme('dark')
    }
  }

  return (
    <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="flex h-14 items-center px-4 gap-4">
        <Button
          variant="ghost"
          size="icon"
          onClick={toggleSidebar}
          aria-label="Toggle sidebar"
        >
          <Menu className="h-5 w-5" />
        </Button>

        <div className="flex items-center gap-2">
          <span className="font-semibold text-lg">MoAI-ADK</span>
          <span className="text-xs text-muted-foreground">Web UI</span>
        </div>

        <div className="flex-1" />

        <div className="flex items-center gap-2">
          {/* Connection Status */}
          <div
            className={cn(
              'flex items-center gap-1.5 px-2 py-1 rounded-md text-xs',
              isConnected
                ? 'bg-green-500/10 text-green-600 dark:text-green-400'
                : 'bg-red-500/10 text-red-600 dark:text-red-400'
            )}
          >
            {isConnected ? (
              <Wifi className="h-3.5 w-3.5" />
            ) : (
              <WifiOff className="h-3.5 w-3.5" />
            )}
            <span>{isConnected ? 'Connected' : 'Disconnected'}</span>
          </div>

          {/* Provider Info */}
          {activeProvider && activeModel && (
            <div className="hidden sm:flex items-center gap-1.5 px-2 py-1 rounded-md bg-muted text-xs">
              <span className="capitalize">{activeProvider}</span>
              <span className="text-muted-foreground">/</span>
              <span className="text-muted-foreground truncate max-w-[120px]">
                {activeModel}
              </span>
            </div>
          )}

          {/* Theme Toggle */}
          <Button
            variant="ghost"
            size="icon"
            onClick={toggleTheme}
            aria-label="Toggle theme"
          >
            {theme === 'dark' ? (
              <Sun className="h-5 w-5" />
            ) : (
              <Moon className="h-5 w-5" />
            )}
          </Button>

          {/* Settings */}
          <Button variant="ghost" size="icon" aria-label="Settings">
            <Settings className="h-5 w-5" />
          </Button>
        </div>
      </div>
    </header>
  )
}
