import { useState, useEffect, useCallback } from 'react'
import { Plus, Search, Filter, RefreshCw, Play } from 'lucide-react'
import { SpecCard } from './spec-card'
import { Button, Input, ScrollArea } from '@/components/ui'
import type { Spec, SpecStatus, SpecListResponse } from '@/types'

const statusFilters: { value: SpecStatus | 'all'; label: string }[] = [
  { value: 'all', label: 'All' },
  { value: 'draft', label: 'Draft' },
  { value: 'planned', label: 'Planned' },
  { value: 'in_progress', label: 'In Progress' },
  { value: 'completed', label: 'Completed' },
  { value: 'blocked', label: 'Blocked' },
]

// Transform API response to frontend Spec type
function transformApiSpec(apiSpec: Record<string, unknown>): Spec {
  return {
    id: (apiSpec.spec_id as string) || '',
    title: (apiSpec.title as string) || '',
    description: (apiSpec.description as string) || '',
    status: (apiSpec.status as SpecStatus) || 'draft',
    priority: (apiSpec.priority as 'low' | 'medium' | 'high') || 'medium',
    createdAt: (apiSpec.created_at as string) || new Date().toISOString(),
    updatedAt: (apiSpec.updated_at as string) || new Date().toISOString(),
    tags: (apiSpec.tags as string[]) || [],
    progress: (apiSpec.progress as number) || 0,
    tasks: [],
    filePath: apiSpec.file_path as string | undefined,
    worktreePath: apiSpec.worktree_path as string | undefined,
  }
}

export function SpecList() {
  const [search, setSearch] = useState('')
  const [statusFilter, setStatusFilter] = useState<SpecStatus | 'all'>('all')
  const [specs, setSpecs] = useState<Spec[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchSpecs = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const response = await fetch('/api/specs')
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      const data: SpecListResponse = await response.json()
      // Transform API response to match frontend types
      const transformedSpecs = data.specs.map((spec) =>
        transformApiSpec(spec as unknown as Record<string, unknown>)
      )
      setSpecs(transformedSpecs)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch SPECs')
      console.error('Error fetching specs:', err)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchSpecs()
  }, [fetchSpecs])

  const filteredSpecs = specs.filter((spec) => {
    const matchesSearch =
      spec.title.toLowerCase().includes(search.toLowerCase()) ||
      spec.description.toLowerCase().includes(search.toLowerCase())
    const matchesStatus =
      statusFilter === 'all' || spec.status === statusFilter
    return matchesSearch && matchesStatus
  })

  const handleRunSpec = async (specId: string) => {
    try {
      const response = await fetch(`/api/specs/${specId}/run`, {
        method: 'POST',
      })
      if (!response.ok) {
        throw new Error(`Failed to run SPEC: ${response.status}`)
      }
      const data = await response.json()
      console.log('SPEC run initiated:', data)
      // TODO: Open terminal with the command
      // For now, show alert with command
      alert(`Ready to run: ${data.command}\n\nWorktree: ${data.worktree_path || 'None'}`)
    } catch (err) {
      console.error('Error running SPEC:', err)
      alert(err instanceof Error ? err.message : 'Failed to run SPEC')
    }
  }

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="p-4 border-b space-y-3">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold">SPEC Monitor</h2>
          <div className="flex items-center gap-2">
            <Button
              variant="ghost"
              size="sm"
              onClick={fetchSpecs}
              disabled={loading}
            >
              <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
            </Button>
            <Button size="sm">
              <Plus className="h-4 w-4 mr-1" />
              New SPEC
            </Button>
          </div>
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
          {loading ? (
            <div className="col-span-full text-center py-8">
              <RefreshCw className="h-6 w-6 animate-spin mx-auto mb-2 text-muted-foreground" />
              <p className="text-muted-foreground">Loading SPECs...</p>
            </div>
          ) : error ? (
            <div className="col-span-full text-center py-8">
              <p className="text-destructive mb-2">{error}</p>
              <Button variant="outline" size="sm" onClick={fetchSpecs}>
                Retry
              </Button>
            </div>
          ) : filteredSpecs.length === 0 ? (
            <div className="col-span-full text-center py-8">
              <p className="text-muted-foreground">
                {specs.length === 0
                  ? 'No SPECs found. Create one using /moai:1-plan in chat.'
                  : 'No SPECs match your search'}
              </p>
            </div>
          ) : (
            filteredSpecs.map((spec) => (
              <div key={spec.id} className="relative group">
                <SpecCard spec={spec} />
                {/* Run button overlay */}
                <div className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity">
                  <Button
                    size="sm"
                    variant="secondary"
                    onClick={() => handleRunSpec(spec.id)}
                    title="Run SPEC with /moai:all-is-well"
                  >
                    <Play className="h-3 w-3 mr-1" />
                    Run
                  </Button>
                </div>
                {/* Worktree indicator */}
                {spec.worktreePath && (
                  <div className="absolute bottom-2 right-2">
                    <span className="text-xs bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200 px-2 py-0.5 rounded">
                      Worktree
                    </span>
                  </div>
                )}
              </div>
            ))
          )}
        </div>
      </ScrollArea>
    </div>
  )
}
