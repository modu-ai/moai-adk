import { DollarSign, TrendingUp, Calendar, Activity } from 'lucide-react'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui'
import { formatCost } from '@/lib/utils'
import type { CostSummary as CostSummaryType } from '@/types'

interface CostSummaryProps {
  summary: CostSummaryType
}

export function CostSummary({ summary }: CostSummaryProps) {
  const cards = [
    {
      title: 'Today',
      value: summary.daily,
      icon: Calendar,
      color: 'text-blue-500',
    },
    {
      title: 'This Week',
      value: summary.weekly,
      icon: TrendingUp,
      color: 'text-green-500',
    },
    {
      title: 'This Month',
      value: summary.monthly,
      icon: DollarSign,
      color: 'text-purple-500',
    },
    {
      title: 'Active Models',
      value: Object.keys(summary.byModel).length,
      icon: Activity,
      color: 'text-orange-500',
      isCount: true,
    },
  ]

  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      {cards.map((card) => (
        <Card key={card.title}>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">{card.title}</CardTitle>
            <card.icon className={`h-4 w-4 ${card.color}`} />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {card.isCount ? card.value : formatCost(card.value as number)}
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  )
}
