import { useState, useCallback, useRef, useEffect } from 'react'
import { Send, Square, Paperclip } from 'lucide-react'
import { Button } from '@/components/ui'
import { cn } from '@/lib/utils'

interface ChatInputProps {
  onSend: (content: string) => void
  onStop: () => void
  isStreaming: boolean
  isConnected: boolean
  disabled?: boolean
}

export function ChatInput({
  onSend,
  onStop,
  isStreaming,
  isConnected,
  disabled = false,
}: ChatInputProps) {
  const [value, setValue] = useState('')
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  // Auto-resize textarea
  useEffect(() => {
    const textarea = textareaRef.current
    if (textarea) {
      textarea.style.height = 'auto'
      textarea.style.height = `${Math.min(textarea.scrollHeight, 200)}px`
    }
  }, [value])

  const handleSubmit = useCallback(() => {
    if (!value.trim() || isStreaming || !isConnected || disabled) return
    onSend(value.trim())
    setValue('')
  }, [value, isStreaming, isConnected, disabled, onSend])

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
      if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault()
        handleSubmit()
      }
    },
    [handleSubmit]
  )

  return (
    <div className="border-t bg-background p-4">
      <div className="max-w-4xl mx-auto">
        <div
          className={cn(
            'flex items-end gap-2 p-2 rounded-lg border bg-muted/30',
            !isConnected && 'opacity-50'
          )}
        >
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 shrink-0"
            disabled={!isConnected || isStreaming}
            aria-label="Attach file"
          >
            <Paperclip className="h-4 w-4" />
          </Button>

          <textarea
            ref={textareaRef}
            value={value}
            onChange={(e) => setValue(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder={
              !isConnected
                ? 'Connecting...'
                : isStreaming
                  ? 'Waiting for response...'
                  : 'Type a message... (Shift+Enter for new line)'
            }
            disabled={!isConnected || disabled}
            className="flex-1 min-h-[40px] max-h-[200px] bg-transparent border-0 resize-none focus:outline-none text-sm placeholder:text-muted-foreground disabled:cursor-not-allowed"
            rows={1}
          />

          {isStreaming ? (
            <Button
              variant="destructive"
              size="icon"
              className="h-8 w-8 shrink-0"
              onClick={onStop}
              aria-label="Stop generation"
            >
              <Square className="h-4 w-4" />
            </Button>
          ) : (
            <Button
              variant="default"
              size="icon"
              className="h-8 w-8 shrink-0"
              onClick={handleSubmit}
              disabled={!value.trim() || !isConnected || disabled}
              aria-label="Send message"
            >
              <Send className="h-4 w-4" />
            </Button>
          )}
        </div>

        <p className="text-xs text-muted-foreground text-center mt-2">
          Press Enter to send, Shift+Enter for new line
        </p>
      </div>
    </div>
  )
}
