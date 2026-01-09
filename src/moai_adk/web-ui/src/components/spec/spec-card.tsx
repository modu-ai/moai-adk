import { FileText, Clock, CheckCircle, AlertCircle, Loader2 } from 'lucide-react'
import { Card, CardHeader, CardTitle, CardContent, Badge } from '@/components/ui'
import { cn } from '@/lib/utils'
import type { Spec, SpecStatus } from '@/types'

interface SpecCardProps {
  spec: Spec
  onClick?: () => void
}

const statusConfig: Record<
  SpecStatus,
  { icon: typeof CheckCircle; color: string; label: string }
> = {
  draft: { icon: FileText, color: 'text-muted-foreground', label: 'Draft' },
  planned: { icon: Clock, color: 'text-blue-500', label: 'Planned' },
  in_progress: { icon: Loader2, color: 'text-yellow-500', label: 'In Progress' },
  completed: { icon: CheckCircle, color: 'text-green-500', label: 'Completed' },
  blocked: { icon: AlertCircle, color: 'text-red-500', label: 'Blocked' },
}

export function SpecCard({ spec, onClick }: SpecCardProps) {
  const status = statusConfig[spec.status]
  const StatusIcon = status.icon

  return (
    <Card
      className={cn(
        'cursor-pointer transition-colors hover:bg-muted/50',
        onClick && 'hover:border-primary/50'
      )}
      onClick={onClick}
    >
      <CardHeader className="pb-2">
        <div className="flex items-start justify-between gap-2">
          <CardTitle className="text-base line-clamp-1">{spec.title}</CardTitle>
          <StatusIcon
            className={cn(
              'h-4 w-4 shrink-0',
              status.color,
              spec.status === 'in_progress' && 'animate-spin'
            )}
          />
        </div>
      </CardHeader>
      <CardContent>
        <p className="text-sm text-muted-foreground line-clamp-2 mb-3">
          {spec.description}
        </p>

        {/* Progress Bar */}
        <div className="mb-3">
          <div className="flex items-center justify-between text-xs mb-1">
            <span className="text-muted-foreground">Progress</span>
            <span className="font-medium">{spec.progress}%</span>
          </div>
          <div className="h-1.5 bg-muted rounded-full overflow-hidden">
            <div
              className="h-full bg-primary transition-all"
              style={{ width: `${spec.progress}%` }}
            />
          </div>
        </div>

        {/* Tags */}
        <div className="flex flex-wrap gap-1">
          {spec.tags.slice(0, 3).map((tag) => (
            <Badge key={tag} variant="secondary" className="text-[10px]">
              {tag}
            </Badge>
          ))}
          {spec.tags.length > 3 && (
            <Badge variant="outline" className="text-[10px]">
              +{spec.tags.length - 3}
            </Badge>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
