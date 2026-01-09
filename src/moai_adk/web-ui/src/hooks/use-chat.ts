import { useCallback, useMemo } from 'react'
import { useWebSocket } from './use-websocket'
import { useSessionStore, useActiveSession } from '@/stores'
import type { WSMessage, Message } from '@/types'

interface UseChatOptions {
  sessionId?: string
}

interface UseChatReturn {
  messages: Message[]
  isConnected: boolean
  isStreaming: boolean
  sendMessage: (content: string) => void
  stopGeneration: () => void
}

export function useChat(options: UseChatOptions = {}): UseChatReturn {
  const { sessionId } = options
  const activeSession = useActiveSession()
  const currentSessionId = sessionId ?? activeSession?.id

  const { addMessage, updateMessage } = useSessionStore()

  const handleMessage = useCallback(
    (wsMessage: WSMessage) => {
      if (!currentSessionId) return

      switch (wsMessage.type) {
        case 'message.create': {
          const message = wsMessage.payload as Message
          addMessage(currentSessionId, message)
          break
        }
        case 'message.update': {
          const { id, ...updates } = wsMessage.payload as Partial<Message> & {
            id: string
          }
          updateMessage(currentSessionId, id, updates)
          break
        }
        case 'stream.chunk': {
          const { messageId, content } = wsMessage.payload as {
            messageId: string
            content: string
          }
          const session = useSessionStore
            .getState()
            .sessions.find((s) => s.id === currentSessionId)
          const existingMessage = session?.messages.find(
            (m) => m.id === messageId
          )
          if (existingMessage) {
            updateMessage(currentSessionId, messageId, {
              content: existingMessage.content + content,
            })
          }
          break
        }
        case 'message.complete': {
          const { id, tokens, cost } = wsMessage.payload as {
            id: string
            tokens: { input: number; output: number }
            cost: number
          }
          updateMessage(currentSessionId, id, {
            isStreaming: false,
            tokens,
            cost,
          })
          break
        }
        case 'error': {
          console.error('Chat error:', wsMessage.payload)
          break
        }
      }
    },
    [currentSessionId, addMessage, updateMessage]
  )

  const wsUrl = useMemo(() => {
    if (!currentSessionId) return null
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    return `${protocol}//${host}/ws/chat/${currentSessionId}`
  }, [currentSessionId])

  const { isConnected, send } = useWebSocket({
    url: wsUrl,
    onMessage: handleMessage,
  })

  const messages = useMemo(() => {
    return activeSession?.messages ?? []
  }, [activeSession?.messages])

  const isStreaming = useMemo(() => {
    return messages.some((m) => m.isStreaming)
  }, [messages])

  const sendMessage = useCallback(
    (content: string) => {
      if (!content.trim() || !currentSessionId) return

      // Optimistically add user message
      const userMessage: Message = {
        id: `temp-${Date.now()}`,
        role: 'user',
        content: content.trim(),
        timestamp: new Date().toISOString(),
      }
      addMessage(currentSessionId, userMessage)

      // Send via WebSocket
      send('message.create', {
        sessionId: currentSessionId,
        content: content.trim(),
      })
    },
    [currentSessionId, send, addMessage]
  )

  const stopGeneration = useCallback(() => {
    send('message.update', { action: 'stop' })
  }, [send])

  return {
    messages,
    isConnected,
    isStreaming,
    sendMessage,
    stopGeneration,
  }
}
