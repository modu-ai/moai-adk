import { useState, useCallback, useRef, useEffect } from 'react'
import { Terminal as TerminalIcon, Trash2, Plus } from 'lucide-react'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import '@xterm/xterm/css/xterm.css'
import { Button } from '@/components/ui'
import { cn } from '@/lib/utils'

export function TerminalView() {
  const [terminalId, setTerminalId] = useState<string | null>(null)
  const [isCreating, setIsCreating] = useState(false)
  const [isConnected, setIsConnected] = useState(false)
  const terminalRef = useRef<HTMLDivElement>(null)
  const xtermRef = useRef<Terminal | null>(null)
  const fitAddonRef = useRef<FitAddon | null>(null)
  const wsRef = useRef<WebSocket | null>(null)

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

  // Initialize xterm.js when terminal container is ready
  useEffect(() => {
    if (!terminalRef.current || !terminalId) return

    // Create xterm instance
    const term = new Terminal({
      cursorBlink: true,
      fontSize: 14,
      fontFamily: 'Menlo, Monaco, "Courier New", monospace',
      theme: {
        background: '#09090b',
        foreground: '#fafafa',
        cursor: '#fafafa',
        cursorAccent: '#09090b',
        selectionBackground: '#3f3f46',
        black: '#09090b',
        red: '#ef4444',
        green: '#22c55e',
        yellow: '#eab308',
        blue: '#3b82f6',
        magenta: '#a855f7',
        cyan: '#06b6d4',
        white: '#fafafa',
        brightBlack: '#52525b',
        brightRed: '#f87171',
        brightGreen: '#4ade80',
        brightYellow: '#facc15',
        brightBlue: '#60a5fa',
        brightMagenta: '#c084fc',
        brightCyan: '#22d3ee',
        brightWhite: '#ffffff',
      },
    })

    // Add addons
    const fitAddon = new FitAddon()
    const webLinksAddon = new WebLinksAddon()
    term.loadAddon(fitAddon)
    term.loadAddon(webLinksAddon)

    // Open terminal in DOM
    term.open(terminalRef.current)
    fitAddon.fit()

    // Store refs
    xtermRef.current = term
    fitAddonRef.current = fitAddon

    // Connect WebSocket
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    const wsUrl = `${protocol}//${host}/ws/terminal/${terminalId}`
    const ws = new WebSocket(wsUrl)

    ws.onopen = () => {
      setIsConnected(true)
      term.writeln('\x1b[32m[Connected to terminal]\x1b[0m')
    }

    ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data)
        if (message.type === 'output' && message.data) {
          term.write(message.data)
        } else if (message.type === 'error' && message.data) {
          term.write(`\x1b[31m${message.data}\x1b[0m`)
        } else if (message.type === 'close') {
          term.writeln(`\x1b[33m[Terminal closed: ${message.data}]\x1b[0m`)
          setIsConnected(false)
        }
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error)
      }
    }

    ws.onclose = () => {
      setIsConnected(false)
      term.writeln('\x1b[33m[Disconnected]\x1b[0m')
    }

    ws.onerror = (error) => {
      console.error('WebSocket error:', error)
      term.writeln('\x1b[31m[Connection error]\x1b[0m')
    }

    wsRef.current = ws

    // Handle terminal input
    term.onData((data) => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type: 'input', data }))
      }
    })

    // Handle resize
    const handleResize = () => {
      fitAddon.fit()
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({
          type: 'resize',
          cols: term.cols,
          rows: term.rows,
        }))
      }
    }

    window.addEventListener('resize', handleResize)

    // Cleanup
    return () => {
      window.removeEventListener('resize', handleResize)
      ws.close()
      term.dispose()
      xtermRef.current = null
      fitAddonRef.current = null
      wsRef.current = null
    }
  }, [terminalId])

  // Resize terminal when container size changes
  useEffect(() => {
    if (!fitAddonRef.current) return

    const resizeObserver = new ResizeObserver(() => {
      fitAddonRef.current?.fit()
    })

    if (terminalRef.current) {
      resizeObserver.observe(terminalRef.current)
    }

    return () => {
      resizeObserver.disconnect()
    }
  }, [terminalId])

  const handleClear = useCallback(() => {
    xtermRef.current?.clear()
  }, [])

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
            onClick={handleClear}
            disabled={!terminalId}
          >
            <Trash2 className="h-3.5 w-3.5" />
          </Button>
        </div>
      </div>

      {/* Terminal Area */}
      <div className="flex-1 overflow-hidden">
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
        ) : (
          <div
            ref={terminalRef}
            className="h-full w-full p-2"
          />
        )}
      </div>
    </div>
  )
}
