import { useState, useCallback, useRef, useEffect } from 'react'
import { Terminal as TerminalIcon, Trash2, Copy, Check, Plus } from 'lucide-react'
import { Button } from '@/components/ui'
import { useTerminal } from '@/hooks'
import { cn } from '@/lib/utils'

export function TerminalView() {
  const [terminalId, setTerminalId] = useState<string | null>(null)
  const [isCreating, setIsCreating] = useState(false)
  const { outputs, isConnected, sendCommand, clear } = useTerminal({ terminalId: terminalId ?? undefined })
  const [input, setInput] = useState('')
  const [copied, setCopied] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)
  const scrollRef = useRef<HTMLDivElement>(null)

  // Create a new terminal session
  const createTerminal = useCallback(async () => {
    setIsCreating(true)
    try {
      const response = await fetch('/api/terminals', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
      })
      if (response.ok) {
        const data = await response.json()
        setTerminalId(data.id)
      }
    } catch (error) {
      console.error('Failed to create terminal:', error)
    } finally {
      setIsCreating(false)
    }
  }, [])

  // Auto-scroll to bottom
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight
    }
  }, [outputs])

  const handleSubmit = useCallback(
    (e: React.FormEvent) => {
      e.preventDefault()
      if (!input.trim()) return
      sendCommand(input.trim())
      setInput('')
    },
    [input, sendCommand]
  )

  const handleCopyAll = useCallback(() => {
    const text = outputs.map((o) => o.content).join('\n')
    navigator.clipboard.writeText(text)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }, [outputs])

  return (
    <div className="flex flex-col h-full bg-zinc-950 text-zinc-100">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-2 border-b border-zinc-800">
        <div className="flex items-center gap-2">
          <TerminalIcon className="h-4 w-4" />
          <span className="text-sm font-medium">Terminal</span>
          <span
            className={cn(
              'w-2 h-2 rounded-full',
              isConnected ? 'bg-green-500' : 'bg-red-500'
            )}
          />
        </div>
        <div className="flex items-center gap-1">
          <Button
            variant="ghost"
            size="icon"
            className="h-7 w-7 text-zinc-400 hover:text-zinc-100"
            onClick={handleCopyAll}
            disabled={outputs.length === 0}
          >
            {copied ? (
              <Check className="h-3.5 w-3.5" />
            ) : (
              <Copy className="h-3.5 w-3.5" />
            )}
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="h-7 w-7 text-zinc-400 hover:text-zinc-100"
            onClick={clear}
            disabled={outputs.length === 0}
          >
            <Trash2 className="h-3.5 w-3.5" />
          </Button>
        </div>
      </div>

      {/* Output Area */}
      <div
        ref={scrollRef}
        className="flex-1 overflow-auto p-4 font-mono text-sm"
      >
        {!terminalId ? (
          <div className="flex flex-col items-center justify-center h-full text-zinc-500">
            <TerminalIcon className="h-12 w-12 mb-4 opacity-50" />
            <p className="mb-4">No terminal session active</p>
            <Button
              onClick={createTerminal}
              disabled={isCreating}
              className="gap-2"
            >
              <Plus className="h-4 w-4" />
              {isCreating ? 'Creating...' : 'Create Terminal'}
            </Button>
          </div>
        ) : outputs.length === 0 ? (
          <div className="text-zinc-500">
            <p>Terminal ready. Type a command to execute.</p>
            <p className="mt-1 text-xs">
              Note: Commands are executed in the context of the MoAI-ADK session.
            </p>
          </div>
        ) : (
          outputs.map((output) => (
            <div
              key={output.id}
              className={cn(
                'whitespace-pre-wrap break-all',
                output.type === 'stdin' && 'text-green-400',
                output.type === 'stdout' && 'text-zinc-100',
                output.type === 'stderr' && 'text-red-400'
              )}
            >
              {output.type === 'stdin' && (
                <span className="text-zinc-500">$ </span>
              )}
              {output.content}
            </div>
          ))
        )}
      </div>

      {/* Input Area */}
      <form
        onSubmit={handleSubmit}
        className="flex items-center gap-2 px-4 py-3 border-t border-zinc-800"
      >
        <span className="text-green-400 font-mono">$</span>
        <input
          ref={inputRef}
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder={!terminalId ? 'Create terminal first...' : isConnected ? 'Enter command...' : 'Connecting...'}
          disabled={!terminalId || !isConnected}
          className="flex-1 bg-transparent border-none outline-none font-mono text-sm placeholder:text-zinc-600 disabled:cursor-not-allowed"
          autoFocus
        />
      </form>
    </div>
  )
}
