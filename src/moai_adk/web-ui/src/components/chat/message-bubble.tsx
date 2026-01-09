import { User, Bot, Wrench, Loader2 } from 'lucide-react'
import { Badge } from '@/components/ui'
import { cn, formatTokens, formatCost, formatTimestamp } from '@/lib/utils'
import type { Message, ToolCall } from '@/types'

interface MessageBubbleProps {
  message: Message
}

function ToolCallDisplay({ toolCall }: { toolCall: ToolCall }) {
  return (
    <div className="mt-2 p-2 rounded-md bg-muted/50 border text-sm">
      <div className="flex items-center gap-2 mb-1">
        <Wrench className="h-3.5 w-3.5" />
        <span className="font-medium">{toolCall.name}</span>
        <Badge
          variant={
            toolCall.status === 'completed'
              ? 'success'
              : toolCall.status === 'error'
                ? 'destructive'
                : toolCall.status === 'running'
                  ? 'warning'
                  : 'secondary'
          }
          className="text-[10px] px-1.5 py-0"
        >
          {toolCall.status}
        </Badge>
        {toolCall.duration && (
          <span className="text-xs text-muted-foreground ml-auto">
            {toolCall.duration}ms
          </span>
        )}
      </div>
      {toolCall.result && (
        <pre className="text-xs text-muted-foreground overflow-x-auto whitespace-pre-wrap">
          {toolCall.result.slice(0, 500)}
          {toolCall.result.length > 500 && '...'}
        </pre>
      )}
    </div>
  )
}

export function MessageBubble({ message }: MessageBubbleProps) {
  const isUser = message.role === 'user'
  const isAssistant = message.role === 'assistant'

  return (
    <div
      className={cn(
        'flex gap-3 p-4',
        isUser ? 'bg-muted/30' : 'bg-background'
      )}
    >
      {/* Avatar */}
      <div
        className={cn(
          'flex-shrink-0 w-8 h-8 rounded-full flex items-center justify-center',
          isUser ? 'bg-primary text-primary-foreground' : 'bg-secondary'
        )}
      >
        {isUser ? (
          <User className="h-4 w-4" />
        ) : (
          <Bot className="h-4 w-4" />
        )}
      </div>

      {/* Content */}
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 mb-1">
          <span className="font-medium text-sm">
            {isUser ? 'You' : 'Assistant'}
          </span>
          <span className="text-xs text-muted-foreground">
            {formatTimestamp(message.timestamp)}
          </span>
          {message.isStreaming && (
            <Loader2 className="h-3 w-3 animate-spin text-muted-foreground" />
          )}
        </div>

        {/* Message Content */}
        <div className="prose prose-sm dark:prose-invert max-w-none">
          <p className="whitespace-pre-wrap break-words">{message.content}</p>
        </div>

        {/* Tool Calls */}
        {message.toolCalls && message.toolCalls.length > 0 && (
          <div className="mt-2 space-y-2">
            {message.toolCalls.map((toolCall) => (
              <ToolCallDisplay key={toolCall.id} toolCall={toolCall} />
            ))}
          </div>
        )}

        {/* Token/Cost Info */}
        {isAssistant && message.tokens && (
          <div className="flex items-center gap-3 mt-2 text-xs text-muted-foreground">
            <span>
              Tokens: {formatTokens(message.tokens.input)} in /{' '}
              {formatTokens(message.tokens.output)} out
            </span>
            {message.cost !== undefined && (
              <span>Cost: {formatCost(message.cost)}</span>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
