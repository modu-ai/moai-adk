import { forwardRef, type HTMLAttributes } from 'react'
import { cn } from '@/lib/utils'

const ScrollArea = forwardRef<HTMLDivElement, HTMLAttributes<HTMLDivElement>>(
  ({ className, children, ...props }, ref) => (
    <div
      ref={ref}
      className={cn('relative overflow-hidden', className)}
      {...props}
    >
      <div className="h-full w-full overflow-auto scrollbar-thin scrollbar-thumb-muted scrollbar-track-transparent">
        {children}
      </div>
    </div>
  )
)
ScrollArea.displayName = 'ScrollArea'

export { ScrollArea }
