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
        <nav className="flex flex-col items-center gap-2 p-2">
          {navItems.map((item) => (
            <Button
              key={item.id}
              variant={activeTab === item.id ? 'secondary' : 'ghost'}
              size="icon"
              onClick={() => setActiveTab(item.id)}
              aria-label={item.label}
            >
              <item.icon className="h-5 w-5" />
            </Button>
          ))}
        </nav>
      </aside>
    )
  }

  return (
    <aside className="hidden md:flex flex-col w-64 border-r bg-muted/30">
      {/* Navigation */}
      <nav className="flex items-center gap-1 p-2 border-b">
        {navItems.map((item) => (
          <Button
            key={item.id}
            variant={activeTab === item.id ? 'secondary' : 'ghost'}
            size="sm"
            className="flex-1"
            onClick={() => setActiveTab(item.id)}
          >
            <item.icon className="h-4 w-4 mr-1.5" />
            {item.label}
          </Button>
        ))}
      </nav>

      {/* Sessions List (shown for chat tab) */}
      {activeTab === 'chat' && (
        <div className="flex-1 flex flex-col">
          <div className="flex items-center justify-between p-2 border-b">
            <span className="text-sm font-medium">Sessions</span>
            <Button
              variant="ghost"
              size="icon"
              className="h-7 w-7"
              onClick={handleNewSession}
              aria-label="New session"
            >
              <Plus className="h-4 w-4" />
            </Button>
          </div>

          <ScrollArea className="flex-1">
            <div className="p-2 space-y-1">
              {sessions.length === 0 ? (
                <p className="text-sm text-muted-foreground text-center py-4">
                  No sessions yet
                </p>
              ) : (
                sessions.map((session) => (
                  <div
                    key={session.id}
                    className={cn(
                      'group flex items-center gap-2 p-2 rounded-md cursor-pointer transition-colors',
                      activeSessionId === session.id
                        ? 'bg-secondary'
                        : 'hover:bg-muted'
                    )}
                    onClick={() => setActiveSession(session.id)}
                  >
                    <MessageSquare className="h-4 w-4 shrink-0 text-muted-foreground" />
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium truncate">
                        {session.name}
                      </p>
                      <p className="text-xs text-muted-foreground">
                        {formatTimestamp(session.updatedAt)}
                      </p>
                    </div>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-6 w-6 opacity-0 group-hover:opacity-100 transition-opacity"
                      onClick={(e) => {
                        e.stopPropagation()
                        removeSession(session.id)
                      }}
                      aria-label="Delete session"
                    >
                      <Trash2 className="h-3.5 w-3.5" />
                    </Button>
                  </div>
                ))
              )}
            </div>
          </ScrollArea>
        </div>
      )}
    </aside>
  )
}
