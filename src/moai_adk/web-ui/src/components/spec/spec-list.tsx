import { useState, useEffect, useCallback } from 'react'
import { Plus, Search, Filter, RefreshCw, Play, FileText } from 'lucide-react'
import { SpecCard } from './spec-card'
import { SpecDetailViewer } from './spec-detail-viewer'
import { SpecCreateDialog } from './spec-create-dialog'
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
  const [selectedSpec, setSelectedSpec] = useState<Spec | null>(null)
  const [isViewerOpen, setIsViewerOpen] = useState(false)
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false)

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
        const errorData = await response.json().catch(() => ({}))
        throw new Error(errorData.detail || `Failed to run SPEC: ${response.status}`)
      }
      const data = await response.json()
      console.log('SPEC run initiated:', data)

      // Emit event for terminal panel to connect
      const event = new CustomEvent('spec-terminal-created', {
        detail: {
          specId: data.spec_id,
          terminalId: data.terminal_id,
          websocketUrl: data.websocket_url,
          command: data.command,
          worktreePath: data.worktree_path,
        },
      })
      window.dispatchEvent(event)

      // Refresh spec list to update status
      await fetchSpecs()
    } catch (err) {
      console.error('Error running SPEC:', err)
      alert(err instanceof Error ? err.message : 'Failed to run SPEC')
    }
  }

  const handleCreateSpec = async (instructions: string) => {
    try {
      const response = await fetch('/api/specs/create', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ instructions }),
      })

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}))
        throw new Error(errorData.detail || `Failed to create SPEC: ${response.status}`)
      }

      const data = await response.json()
      console.log('SPEC creation initiated:', data)

      // Emit event for terminal panel to connect
      const event = new CustomEvent('spec-terminal-created', {
        detail: {
          specId: null, // New SPEC, ID will be generated
          terminalId: data.terminal_id,
          websocketUrl: data.websocket_url,
          command: data.command,
          worktreePath: null,
        },
      })
      window.dispatchEvent(event)

      // Refresh spec list after a delay to allow SPEC creation
      setTimeout(() => fetchSpecs(), 3000)
    } catch (err) {
      console.error('Error creating SPEC:', err)
      alert(err instanceof Error ? err.message : 'Failed to create SPEC')
      throw err // Re-throw to let dialog handle loading state
    }
  }

  return (
    <div className="flex flex-col h-full bg-background">
      {/* Page Header */}
      <div className="border-b bg-card">
        <div className="px-6 py-4">
          <div className="flex items-center justify-between mb-4">
            <div>
              <h1 className="text-xl font-semibold tracking-tight">SPEC Monitor</h1>
              <p className="text-sm text-muted-foreground mt-1">
                {specs.length} {specs.length === 1 ? 'specification' : 'specifications'}
              </p>
            </div>
            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={fetchSpecs}
                disabled={loading}
                className="gap-2"
              >
                <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
                <span className="hidden sm:inline">Refresh</span>
              </Button>
              <Button size="sm" className="gap-2" onClick={() => setIsCreateDialogOpen(true)}>
                <Plus className="h-4 w-4" />
                <span className="hidden sm:inline">New SPEC</span>
              </Button>
            </div>
          </div>

          {/* Search */}
          <div className="relative max-w-md">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search specifications..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="pl-10"
            />
          </div>
        </div>

        {/* Status Filters */}
        <div className="px-6 pb-3">
          <div className="flex items-center gap-2 overflow-x-auto pb-1">
            <Filter className="h-4 w-4 text-muted-foreground shrink-0" />
            <div className="h-4 w-px bg-border shrink-0" />
            {statusFilters.map((filter) => (
              <Button
                key={filter.value}
                variant={statusFilter === filter.value ? 'secondary' : 'ghost'}
                size="sm"
                onClick={() => setStatusFilter(filter.value)}
                className="shrink-0 h-7"
              >
                {filter.label}
              </Button>
            ))}
          </div>
        </div>
      </div>

      {/* Spec List */}
      <ScrollArea className="flex-1">
        <div className="p-6">
          <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
            {loading && specs.length === 0 ? (
              <div className="col-span-full flex flex-col items-center justify-center py-16">
                <RefreshCw className="h-8 w-8 animate-spin text-muted-foreground mb-3" />
                <p className="text-muted-foreground">Loading specifications...</p>
              </div>
            ) : error ? (
              <div className="col-span-full flex flex-col items-center justify-center py-16">
                <div className="text-destructive mb-3">{error}</div>
                <Button variant="outline" size="sm" onClick={fetchSpecs}>
                  <RefreshCw className="h-4 w-4 mr-2" />
                  Retry
                </Button>
              </div>
            ) : filteredSpecs.length === 0 ? (
              <div className="col-span-full flex flex-col items-center justify-center py-16">
                <FileText className="h-12 w-12 text-muted-foreground/50 mb-3" />
                <p className="text-muted-foreground">
                  {specs.length === 0
                    ? 'No specifications found. Create one using /moai:1-plan in chat.'
                    : 'No specifications match your search'}
                </p>
              </div>
            ) : (
              filteredSpecs.map((spec) => (
                <div key={spec.id} className="relative group">
                  <SpecCard spec={spec} onClick={() => {
                    setSelectedSpec(spec)
                    setIsViewerOpen(true)
                  }} />
                  {/* Run button overlay */}
                  <div className="absolute top-3 right-3 opacity-0 group-hover:opacity-100 transition-opacity">
                    <Button
                      size="sm"
                      onClick={(e) => {
                        e.stopPropagation()
                        handleRunSpec(spec.id)
                      }}
                      title="Run SPEC with /moai:all-is-well"
                      className="shadow-sm"
                    >
                      <Play className="h-3 w-3 mr-1" />
                      Run
                    </Button>
                  </div>
                  {/* Worktree indicator */}
                  {spec.worktreePath && (
                    <div className="absolute bottom-3 right-3">
                      <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400">
                        Worktree
                      </span>
                    </div>
                  )}
                </div>
              ))
            )}
          </div>
        </div>
      </ScrollArea>

      {/* SPEC Detail Viewer */}
      <SpecDetailViewer
        spec={selectedSpec}
        open={isViewerOpen}
        onOpenChange={setIsViewerOpen}
      />

      {/* SPEC Create Dialog */}
      <SpecCreateDialog
        open={isCreateDialogOpen}
        onOpenChange={setIsCreateDialogOpen}
        onCreateSpec={handleCreateSpec}
      />
    </div>
  )
}
