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
    // Handle backend message format: { type: "output" | "error" | "close", data: string }
    const msgType = wsMessage.type
    const data = (wsMessage as unknown as { data?: string }).data ||
                 (wsMessage.payload as { data?: string })?.data || ''

    if (msgType === 'output' && data) {
      const output: TerminalOutput = {
        id: `${Date.now()}-${Math.random().toString(36).slice(2)}`,
        type: 'stdout',
        content: data,
        timestamp: wsMessage.timestamp || new Date().toISOString(),
      }
      setOutputs((prev) => [...prev, output])
    } else if (msgType === 'error' && data) {
      const output: TerminalOutput = {
        id: `${Date.now()}-${Math.random().toString(36).slice(2)}`,
        type: 'stderr',
        content: data,
        timestamp: wsMessage.timestamp || new Date().toISOString(),
      }
      setOutputs((prev) => [...prev, output])
    } else if (msgType === 'close') {
      const output: TerminalOutput = {
        id: `${Date.now()}-close`,
        type: 'stderr',
        content: `[Terminal closed: ${data}]`,
        timestamp: wsMessage.timestamp || new Date().toISOString(),
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

      // Send command in backend format: { type: "input", data: "command\n" }
      send('input', { data: command + '\n' })
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
