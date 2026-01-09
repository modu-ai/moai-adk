import { useState } from 'react'
import { Plus, Search, Filter } from 'lucide-react'
import { SpecCard } from './spec-card'
import { Button, Input, ScrollArea } from '@/components/ui'
import type { Spec, SpecStatus } from '@/types'

// Demo data for display
const demoSpecs: Spec[] = [
  {
    id: 'SPEC-001',
    title: 'User Authentication System',
    description: 'Implement OAuth2 authentication with JWT tokens and refresh token rotation.',
    status: 'in_progress',
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
    tags: ['auth', 'security', 'backend'],
    progress: 65,
    tasks: [],
  },
  {
    id: 'SPEC-002',
    title: 'API Rate Limiting',
    description: 'Add rate limiting middleware with configurable thresholds per endpoint.',
    status: 'planned',
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
    tags: ['api', 'performance'],
    progress: 0,
    tasks: [],
  },
  {
    id: 'SPEC-003',
    title: 'Dashboard Analytics',
    description: 'Create analytics dashboard with real-time charts and metrics.',
    status: 'completed',
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
    tags: ['frontend', 'charts', 'analytics'],
    progress: 100,
    tasks: [],
  },
]

const statusFilters: { value: SpecStatus | 'all'; label: string }[] = [
  { value: 'all', label: 'All' },
  { value: 'draft', label: 'Draft' },
  { value: 'planned', label: 'Planned' },
  { value: 'in_progress', label: 'In Progress' },
  { value: 'completed', label: 'Completed' },
  { value: 'blocked', label: 'Blocked' },
]

export function SpecList() {
  const [search, setSearch] = useState('')
  const [statusFilter, setStatusFilter] = useState<SpecStatus | 'all'>('all')

  const filteredSpecs = demoSpecs.filter((spec) => {
    const matchesSearch =
      spec.title.toLowerCase().includes(search.toLowerCase()) ||
      spec.description.toLowerCase().includes(search.toLowerCase())
    const matchesStatus =
      statusFilter === 'all' || spec.status === statusFilter
    return matchesSearch && matchesStatus
  })

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="p-4 border-b space-y-3">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold">SPEC Monitor</h2>
          <Button size="sm">
            <Plus className="h-4 w-4 mr-1" />
            New SPEC
          </Button>
        </div>

        {/* Search */}
        <div className="relative">
          <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search SPECs..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-9"
          />
        </div>

        {/* Filters */}
        <div className="flex items-center gap-2 overflow-x-auto">
          <Filter className="h-4 w-4 text-muted-foreground shrink-0" />
          {statusFilters.map((filter) => (
            <Button
              key={filter.value}
              variant={statusFilter === filter.value ? 'secondary' : 'ghost'}
              size="sm"
              onClick={() => setStatusFilter(filter.value)}
              className="shrink-0"
            >
              {filter.label}
            </Button>
          ))}
        </div>
      </div>

      {/* Spec List */}
      <ScrollArea className="flex-1">
        <div className="p-4 grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {filteredSpecs.length === 0 ? (
            <div className="col-span-full text-center py-8">
              <p className="text-muted-foreground">No SPECs found</p>
            </div>
          ) : (
            filteredSpecs.map((spec) => (
              <SpecCard key={spec.id} spec={spec} />
            ))
          )}
        </div>
      </ScrollArea>
    </div>
  )
}
