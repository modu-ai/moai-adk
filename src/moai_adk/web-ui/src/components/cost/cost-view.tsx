import { CostSummary } from './cost-summary'
import { CostChart } from './cost-chart'
import { CostByProvider } from './cost-by-provider'
import { ScrollArea } from '@/components/ui'
import { useCostStore } from '@/stores'
import type { CostSummary as CostSummaryType, Provider } from '@/types'

// Demo data
const demoSummary: CostSummaryType = {
  daily: 1.2345,
  weekly: 8.7654,
  monthly: 42.5678,
  byProvider: {
    anthropic: 25.5,
    openai: 12.3,
    google: 4.7678,
  } as Record<Provider, number>,
  byModel: {
    'claude-sonnet-4-20250514': 20.5,
    'claude-3-5-haiku': 5.0,
    'gpt-4o': 12.3,
    'gemini-2.0-flash': 4.7678,
  },
}

const demoChartData = Array.from({ length: 14 }, (_, i) => {
  const date = new Date()
  date.setDate(date.getDate() - (13 - i))
  return {
    date: date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
    cost: Math.random() * 5 + 1,
  }
})

export function CostView() {
  const { summary } = useCostStore()
  const displaySummary = summary ?? demoSummary

  return (
    <ScrollArea className="h-full">
      <div className="p-6 space-y-6">
        <div>
          <h2 className="text-2xl font-bold mb-2">Cost Analytics</h2>
          <p className="text-muted-foreground">
            Track and analyze your AI API usage costs across providers and models.
          </p>
        </div>

        <CostSummary summary={displaySummary} />

        <div className="grid gap-6 lg:grid-cols-2">
          <CostChart data={demoChartData} title="Daily Costs (Last 14 Days)" />
          <CostByProvider data={displaySummary.byProvider} />
        </div>

        {/* Model Breakdown */}
        <div className="rounded-lg border">
          <div className="p-4 border-b">
            <h3 className="font-semibold">Cost by Model</h3>
          </div>
          <div className="divide-y">
            {Object.entries(displaySummary.byModel).map(([model, cost]) => (
              <div
                key={model}
                className="flex items-center justify-between p-4"
              >
                <span className="text-sm font-medium">{model}</span>
                <span className="text-sm text-muted-foreground">
                  ${cost.toFixed(4)}
                </span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </ScrollArea>
  )
}
