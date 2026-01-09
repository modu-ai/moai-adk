import { useState } from 'react'
import { FileText, Loader2 } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  Button,
} from '@/components/ui'

interface SpecCreateDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onCreateSpec: (instructions: string) => Promise<void>
}

export function SpecCreateDialog({
  open,
  onOpenChange,
  onCreateSpec,
}: SpecCreateDialogProps) {
  const [instructions, setInstructions] = useState('')
  const [isCreating, setIsCreating] = useState(false)

  const handleSubmit = async () => {
    if (!instructions.trim()) return

    setIsCreating(true)
    try {
      await onCreateSpec(instructions.trim())
      setInstructions('')
      onOpenChange(false)
    } catch (error) {
      console.error('Failed to create SPEC:', error)
    } finally {
      setIsCreating(false)
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    // Cmd/Ctrl + Enter to submit
    if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
      e.preventDefault()
      handleSubmit()
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <FileText className="h-5 w-5" />
            New SPEC
          </DialogTitle>
          <DialogDescription>
            Describe the feature or requirement you want to implement. Claude will
            help you create a detailed SPEC through an interactive conversation.
          </DialogDescription>
        </DialogHeader>

        <div className="mt-4 space-y-4">
          <div className="space-y-2">
            <label
              htmlFor="instructions"
              className="text-sm font-medium leading-none"
            >
              Instructions
            </label>
            <textarea
              id="instructions"
              value={instructions}
              onChange={(e) => setInstructions(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="Describe what you want to build...&#10;&#10;Example: Implement user authentication with JWT tokens, supporting email/password login and social login via Google OAuth."
              className="flex min-h-[160px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 resize-none"
              disabled={isCreating}
              autoFocus
            />
            <p className="text-xs text-muted-foreground">
              Press Cmd+Enter (Mac) or Ctrl+Enter (Windows) to create
            </p>
          </div>

          <div className="flex justify-end gap-2">
            <Button
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={isCreating}
            >
              Cancel
            </Button>
            <Button
              onClick={handleSubmit}
              disabled={!instructions.trim() || isCreating}
            >
              {isCreating ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  Creating...
                </>
              ) : (
                'Create SPEC'
              )}
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
