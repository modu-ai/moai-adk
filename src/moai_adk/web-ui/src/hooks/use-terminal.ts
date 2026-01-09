import { useCallback, useRef, useState, useMemo } from 'react'
import { useWebSocket } from './use-websocket'
import type { WSMessage, TerminalOutput } from '@/types'

interface UseTerminalOptions {
  terminalId?: string
}

interface UseTerminalReturn {
  outputs: TerminalOutput[]
  isConnected: boolean
  sendCommand: (command: string) => void
  clear: () => void
  terminalRef: React.RefObject<HTMLDivElement | null>
}

export function useTerminal(options: UseTerminalOptions = {}): UseTerminalReturn {
  const { terminalId } = options
  const [outputs, setOutputs] = useState<TerminalOutput[]>([])
  const terminalRef = useRef<HTMLDivElement>(null)

  const handleMessage = useCallback((wsMessage: WSMessage) => {
    if (wsMessage.type === 'tool.update') {
      const payload = wsMessage.payload as {
        type: 'stdout' | 'stderr'
        content: string
      }
      const output: TerminalOutput = {
        id: `${Date.now()}-${Math.random().toString(36).slice(2)}`,
        type: payload.type,
        content: payload.content,
        timestamp: wsMessage.timestamp,
      }
      setOutputs((prev) => [...prev, output])
    }
  }, [])

  const wsUrl = useMemo(() => {
    if (!terminalId) return null
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    return `${protocol}//${host}/ws/terminal/${terminalId}`
  }, [terminalId])

  const { isConnected, send } = useWebSocket({
    url: wsUrl,
    onMessage: handleMessage,
  })

  const sendCommand = useCallback(
    (command: string) => {
      if (!command.trim()) return

      // Add input to outputs
      const inputOutput: TerminalOutput = {
        id: `${Date.now()}-input`,
        type: 'stdin',
        content: command,
        timestamp: new Date().toISOString(),
      }
      setOutputs((prev) => [...prev, inputOutput])

      // Send command
      send('tool.start', { command })
    },
    [send]
  )

  const clear = useCallback(() => {
    setOutputs([])
  }, [])

  return {
    outputs,
    isConnected,
    sendCommand,
    clear,
    terminalRef,
  }
}
