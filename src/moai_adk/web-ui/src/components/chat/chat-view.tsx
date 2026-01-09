import { MessageList } from './message-list'
import { ChatInput } from './chat-input'
import { useChat } from '@/hooks'
import { useActiveSession } from '@/stores'

export function ChatView() {
  const activeSession = useActiveSession()
  const { messages, isConnected, isStreaming, sendMessage, stopGeneration } =
    useChat()

  if (!activeSession) {
    return (
      <div className="flex-1 flex flex-col items-center justify-center p-8 text-center">
        <div className="w-20 h-20 rounded-full bg-muted flex items-center justify-center mb-4">
          <span className="text-3xl">+</span>
        </div>
        <h2 className="text-xl font-semibold mb-2">No Session Selected</h2>
        <p className="text-muted-foreground max-w-md">
          Create a new session or select an existing one from the sidebar to
          start chatting.
        </p>
      </div>
    )
  }

  return (
    <div className="flex flex-col h-full">
      <MessageList messages={messages} isStreaming={isStreaming} />
      <ChatInput
        onSend={sendMessage}
        onStop={stopGeneration}
        isStreaming={isStreaming}
        isConnected={isConnected}
      />
    </div>
  )
}
