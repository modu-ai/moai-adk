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
  approved: { icon: CheckCircle, color: 'text-blue-600', label: 'Approved' },
  in_progress: { icon: Loader2, color: 'text-yellow-500', label: 'In Progress' },
  implementing: { icon: Loader2, color: 'text-orange-500', label: 'Implementing' },
  testing: { icon: Loader2, color: 'text-purple-500', label: 'Testing' },
  completed: { icon: CheckCircle, color: 'text-green-500', label: 'Completed' },
  blocked: { icon: AlertCircle, color: 'text-red-500', label: 'Blocked' },
}

export function SpecCard({ spec, onClick }: SpecCardProps) {
  const status = statusConfig[spec.status]
  const StatusIcon = status.icon

  // Format dates
  const formatDate = (dateString: string) => {
    const date = new Date(dateString)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24))

    if (diffDays === 0) return 'Today'
    if (diffDays === 1) return 'Yesterday'
    if (diffDays < 7) return `${diffDays}d ago`
    if (diffDays < 30) return `${Math.floor(diffDays / 7)}w ago`
    return date.toLocaleDateString()
  }

  return (
    <Card
      className={cn(
        'group cursor-pointer transition-all hover:shadow-md hover:border-primary/30',
        onClick && 'active:scale-[0.99]'
      )}
      onClick={onClick}
    >
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between gap-3">
          <div className="flex-1 min-w-0">
            <CardTitle className="text-base font-semibold line-clamp-1 mb-1">
              {spec.title}
            </CardTitle>
            <div className="flex items-center gap-2 text-xs text-muted-foreground">
              <span>{spec.id}</span>
              <span>•</span>
              <span>{formatDate(spec.updatedAt)}</span>
            </div>
          </div>
          <StatusIcon
            className={cn(
              'h-5 w-5 shrink-0',
              status.color,
              spec.status === 'in_progress' && 'animate-spin'
            )}
          />
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Description */}
        <p className="text-sm text-muted-foreground line-clamp-2 leading-relaxed">
          {spec.description}
        </p>

        {/* Progress Bar */}
        {spec.progress > 0 && (
          <div className="space-y-1.5">
            <div className="flex items-center justify-between text-xs">
              <span className="text-muted-foreground">Progress</span>
              <span className="font-medium">{spec.progress}%</span>
            </div>
            <div className="h-2 bg-muted rounded-full overflow-hidden">
              <div
                className={cn(
                  'h-full rounded-full transition-all duration-500',
                  spec.progress === 100 ? 'bg-green-500' : 'bg-primary'
                )}
                style={{ width: `${spec.progress}%` }}
              />
            </div>
          </div>
        )}

        {/* Footer */}
        <div className="flex items-center justify-between pt-2 border-t">
          {/* Priority Badge */}
          <Badge
            variant={spec.priority === 'high' ? 'destructive' : spec.priority === 'medium' ? 'default' : 'secondary'}
            className="text-[10px] font-medium"
          >
            {spec.priority}
          </Badge>

          {/* Tags */}
          <div className="flex items-center gap-1">
            {spec.tags.slice(0, 2).map((tag) => (
              <Badge key={tag} variant="outline" className="text-[10px]">
                {tag}
              </Badge>
            ))}
            {spec.tags.length > 2 && (
              <Badge variant="outline" className="text-[10px]">
                +{spec.tags.length - 2}
              </Badge>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
