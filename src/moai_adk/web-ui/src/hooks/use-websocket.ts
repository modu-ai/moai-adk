import { useEffect, useRef, useCallback, useState } from 'react'
import type { WSMessage, WSMessageType } from '@/types'

interface UseWebSocketOptions {
  url: string | null
  onMessage?: (message: WSMessage) => void
  onOpen?: () => void
  onClose?: () => void
  onError?: (error: Event) => void
  reconnect?: boolean
  reconnectInterval?: number
  maxReconnectAttempts?: number
}

interface UseWebSocketReturn {
  isConnected: boolean
  isConnecting: boolean
  send: (type: WSMessageType, payload: unknown) => void
  disconnect: () => void
  reconnect: () => void
}

export function useWebSocket(options: UseWebSocketOptions): UseWebSocketReturn {
  const {
    url,
    onMessage,
    onOpen,
    onClose,
    onError,
    reconnect: shouldReconnect = true,
    reconnectInterval = 3000,
    maxReconnectAttempts = 5,
  } = options

  const wsRef = useRef<WebSocket | null>(null)
  const reconnectAttemptsRef = useRef(0)
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const [isConnected, setIsConnected] = useState(false)
  const [isConnecting, setIsConnecting] = useState(false)

  const clearReconnectTimeout = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current)
      reconnectTimeoutRef.current = null
    }
  }, [])

  const connect = useCallback(() => {
    if (!url || wsRef.current?.readyState === WebSocket.OPEN) {
      return
    }

    setIsConnecting(true)
    clearReconnectTimeout()

    try {
      const ws = new WebSocket(url)

      ws.onopen = () => {
        setIsConnected(true)
        setIsConnecting(false)
        reconnectAttemptsRef.current = 0
        onOpen?.()
      }

      ws.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data) as WSMessage
          onMessage?.(message)
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error)
        }
      }

      ws.onclose = () => {
        setIsConnected(false)
        setIsConnecting(false)
        onClose?.()

        if (
          shouldReconnect &&
          reconnectAttemptsRef.current < maxReconnectAttempts
        ) {
          reconnectAttemptsRef.current += 1
          reconnectTimeoutRef.current = setTimeout(() => {
            // eslint-disable-next-line react-hooks/immutability
            connect()
          }, reconnectInterval)
        }
      }

      ws.onerror = (error) => {
        setIsConnecting(false)
        onError?.(error)
      }

      wsRef.current = ws
    } catch (error) {
      setIsConnecting(false)
      console.error('Failed to create WebSocket connection:', error)
    }
  }, [
    url,
    onMessage,
    onOpen,
    onClose,
    onError,
    shouldReconnect,
    reconnectInterval,
    maxReconnectAttempts,
    clearReconnectTimeout,
  ])

  const disconnect = useCallback(() => {
    clearReconnectTimeout()
    reconnectAttemptsRef.current = maxReconnectAttempts // Prevent reconnect
    if (wsRef.current) {
      wsRef.current.close()
      wsRef.current = null
    }
  }, [clearReconnectTimeout, maxReconnectAttempts])

  const reconnectFn = useCallback(() => {
    reconnectAttemptsRef.current = 0
    disconnect()
    connect()
  }, [connect, disconnect])

  const send = useCallback((type: WSMessageType, payload: unknown) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      // For terminal messages, send in backend format: { type, data, ... }
      // For other messages, send in standard format: { type, payload, timestamp }
      const payloadObj = payload as Record<string, unknown>
      const message = payloadObj?.data !== undefined
        ? { type, data: payloadObj.data }  // Terminal format
        : { type, payload, timestamp: new Date().toISOString() }  // Standard format
      wsRef.current.send(JSON.stringify(message))
    } else {
      console.warn('WebSocket is not connected')
    }
  }, [])

  useEffect(() => {
    connect()
    return () => {
      disconnect()
    }
  }, [connect, disconnect])

  return {
    isConnected,
    isConnecting,
    send,
    disconnect,
    reconnect: reconnectFn,
  }
}
