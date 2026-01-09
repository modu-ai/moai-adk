import { useEffect, useRef } from 'react'
import { MessageBubble } from './message-bubble'
import { ScrollArea } from '@/components/ui'
import type { Message } from '@/types'

interface MessageListProps {
  messages: Message[]
  isStreaming: boolean
}

export function MessageList({ messages, isStreaming }: MessageListProps) {
  const bottomRef = useRef<HTMLDivElement>(null)

  // Auto-scroll to bottom when new messages arrive
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages, isStreaming])

  if (messages.length === 0) {
    return (
      <div className="flex-1 flex flex-col items-center justify-center p-8 text-center">
        <div className="w-16 h-16 rounded-full bg-muted flex items-center justify-center mb-4">
          <span className="text-2xl">AI</span>
        </div>
        <h2 className="text-xl font-semibold mb-2">Start a Conversation</h2>
        <p className="text-muted-foreground max-w-md">
          Send a message to begin chatting with the AI assistant. You can ask
          questions, get help with code, or explore ideas.
        </p>
      </div>
    )
  }

  return (
    <ScrollArea className="flex-1">
      <div className="divide-y">
        {messages.map((message) => (
          <MessageBubble key={message.id} message={message} />
        ))}
      </div>
      <div ref={bottomRef} />
    </ScrollArea>
  )
}
