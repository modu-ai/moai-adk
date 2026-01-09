import { useCallback, useRef, useState } from 'react'
import { useWebSocket } from './use-websocket'
import type { WSMessage, TerminalOutput } from '@/types'

interface UseTerminalReturn {
  outputs: TerminalOutput[]
  isConnected: boolean
  sendCommand: (command: string) => void
  clear: () => void
  terminalRef: React.RefObject<HTMLDivElement | null>
}

export function useTerminal(): UseTerminalReturn {
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

  const wsUrl = `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/ws/terminal`

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
