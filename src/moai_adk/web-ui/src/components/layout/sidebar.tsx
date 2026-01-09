import {
  MessageSquare,
  FileText,
  Terminal,
  DollarSign,
  Plus,
  Trash2,
} from 'lucide-react'
import { Button, ScrollArea } from '@/components/ui'
import { useUIStore, useSessionStore } from '@/stores'
import { cn } from '@/lib/utils'
import { formatTimestamp } from '@/lib/utils'

const navItems = [
  { id: 'chat' as const, icon: MessageSquare, label: 'Chat' },
  { id: 'specs' as const, icon: FileText, label: 'SPECs' },
  { id: 'terminal' as const, icon: Terminal, label: 'Terminal' },
  { id: 'costs' as const, icon: DollarSign, label: 'Costs' },
]

export function Sidebar() {
  const { sidebarOpen, activeTab, setActiveTab } = useUIStore()
  const { sessions, activeSessionId, setActiveSession, addSession, removeSession } =
    useSessionStore()

  const handleNewSession = () => {
    const newSession = {
      id: `session-${Date.now()}`,
      name: `Session ${sessions.length + 1}`,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      provider: 'anthropic' as const,
      model: 'claude-sonnet-4-20250514',
      messages: [],
      totalCost: 0,
      totalTokens: { input: 0, output: 0 },
    }
    addSession(newSession)
  }

  if (!sidebarOpen) {
    return (
      <aside className="hidden md:flex flex-col w-16 border-r bg-muted/30">
        <nav className="flex flex-col items-center gap-1 p-2">
          {navItems.map((item) => (
            <Button
              key={item.id}
              variant={activeTab === item.id ? 'secondary' : 'ghost'}
              size="icon"
              onClick={() => setActiveTab(item.id)}
              aria-label={item.label}
              className="relative"
            >
              <item.icon className="h-5 w-5" />
              {activeTab === item.id && (
                <span className="absolute left-0 top-1/2 -translate-y-1/2 w-0.5 h-5 bg-primary rounded-r-full" />
              )}
            </Button>
          ))}
        </nav>
      </aside>
    )
  }

  return (
    <aside className="hidden md:flex flex-col w-64 border-r bg-muted/30">
      {/* Navigation */}
      <nav className="flex flex-col gap-1 p-2 border-b">
        {navItems.map((item) => (
          <Button
            key={item.id}
            variant={activeTab === item.id ? 'secondary' : 'ghost'}
            size="sm"
            className="w-full justify-start"
            onClick={() => setActiveTab(item.id)}
          >
            <item.icon className="h-4 w-4 mr-2" />
            {item.label}
          </Button>
        ))}
      </nav>

      {/* Sessions List (shown for chat tab) */}
      {activeTab === 'chat' && (
        <div className="flex-1 flex flex-col min-h-0">
          <div className="flex items-center justify-between px-3 py-2 border-b">
            <span className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
              Sessions
            </span>
            <Button
              variant="ghost"
              size="icon"
              className="h-7 w-7 shrink-0"
              onClick={handleNewSession}
              aria-label="New session"
            >
              <Plus className="h-4 w-4" />
            </Button>
          </div>

          <ScrollArea className="flex-1">
            <div className="p-2 space-y-1">
              {sessions.length === 0 ? (
                <div className="flex flex-col items-center justify-center py-8 px-2 text-center">
                  <MessageSquare className="h-8 w-8 text-muted-foreground/50 mb-2" />
                  <p className="text-sm text-muted-foreground">
                    No sessions yet
                  </p>
                </div>
              ) : (
                sessions.map((session) => (
                  <button
                    key={session.id}
                    className={cn(
                      'group w-full flex items-center gap-2 p-2.5 rounded-md text-left transition-all',
                      activeSessionId === session.id
                        ? 'bg-secondary shadow-sm'
                        : 'hover:bg-muted/50'
                    )}
                    onClick={() => setActiveSession(session.id)}
                  >
                    <MessageSquare className={cn(
                      'h-4 w-4 shrink-0',
                      activeSessionId === session.id ? 'text-foreground' : 'text-muted-foreground'
                    )} />
                    <div className="flex-1 min-w-0">
                      <p className={cn(
                        'text-sm truncate',
                        activeSessionId === session.id ? 'font-medium' : 'font-normal'
                      )}>
                        {session.name}
                      </p>
                      <p className="text-xs text-muted-foreground">
                        {formatTimestamp(session.updatedAt)}
                      </p>
                    </div>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-6 w-6 shrink-0 opacity-0 group-hover:opacity-100 transition-opacity"
                      onClick={(e) => {
                        e.stopPropagation()
                        removeSession(session.id)
                      }}
                      aria-label="Delete session"
                    >
                      <Trash2 className="h-3.5 w-3.5" />
                    </Button>
                  </button>
                ))
              )}
            </div>
          </ScrollArea>
        </div>
      )}
    </aside>
  )
}
