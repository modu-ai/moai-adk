import { useState, useEffect, useCallback } from 'react'
import { FileText, List, CheckCircle, Loader2, AlertCircle } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  Tabs,
  TabsList,
  TabsTrigger,
  ScrollArea,
  Badge,
} from '@/components/ui'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import type { Spec } from '@/types'

interface SpecFileContent {
  spec: string
  plan: string
  sync: string
}

interface SpecDetailViewerProps {
  spec: Spec | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

const tabConfig = [
  { value: 'spec', label: 'Specification', icon: FileText },
  { value: 'plan', label: 'Implementation Plan', icon: List },
  { value: 'sync', label: 'Sync Documentation', icon: CheckCircle },
]

function EmptyState({ message }: { message: string }) {
  return (
    <div className="flex flex-col items-center justify-center py-16 text-center">
      <FileText className="h-12 w-12 text-muted-foreground/50 mb-3" />
      <p className="text-muted-foreground">{message}</p>
    </div>
  )
}

function LoadingState() {
  return (
    <div className="flex flex-col items-center justify-center py-16">
      <Loader2 className="h-8 w-8 animate-spin text-muted-foreground mb-3" />
      <p className="text-muted-foreground">Loading SPEC content...</p>
    </div>
  )
}

function ErrorState({ error }: { error: string }) {
  return (
    <div className="flex flex-col items-center justify-center py-16 text-center">
      <AlertCircle className="h-12 w-12 text-destructive mb-3" />
      <p className="text-destructive font-medium mb-2">Failed to load content</p>
      <p className="text-sm text-muted-foreground">{error}</p>
    </div>
  )
}

/**
 * Remove YAML frontmatter from markdown content
 * Frontmatter is enclosed between --- markers at the start of the file
 */
function removeYamlFrontmatter(content: string): string {
  const trimmed = content.trim()
  if (!trimmed.startsWith('---')) {
    return content
  }

  // Find the closing ---
  const endIndex = trimmed.indexOf('---', 3)
  if (endIndex === -1) {
    return content
  }

  // Return content after the frontmatter, trimmed
  return trimmed.slice(endIndex + 3).trim()
}

function MarkdownContent({ content }: { content: string }) {
  if (!content || content.trim() === '') {
    return <EmptyState message="No content available" />
  }

  // Remove YAML frontmatter before rendering
  const markdownContent = removeYamlFrontmatter(content)

  return (
    <div className="prose prose-sm dark:prose-invert max-w-none">
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        components={{
          h1: ({ node, ...props }) => (
            <h1 className="text-2xl font-bold mt-6 mb-4 first:mt-0" {...props} />
          ),
          h2: ({ node, ...props }) => (
            <h2 className="text-xl font-semibold mt-5 mb-3 first:mt-0" {...props} />
          ),
          h3: ({ node, ...props }) => (
            <h3 className="text-lg font-medium mt-4 mb-2 first:mt-0" {...props} />
          ),
          p: ({ node, ...props }) => (
            <p className="my-3 leading-7" {...props} />
          ),
          ul: ({ node, ...props }) => (
            <ul className="my-3 ml-6 list-disc space-y-1" {...props} />
          ),
          ol: ({ node, ...props }) => (
            <ol className="my-3 ml-6 list-decimal space-y-1" {...props} />
          ),
          li: ({ node, ...props }) => (
            <li className="mt-1" {...props} />
          ),
          code: ({ node, inline, ...props }: any) =>
            inline ? (
              <code
                className="px-1.5 py-0.5 rounded bg-muted text-sm font-mono"
                {...props}
              />
            ) : (
              <code
                className="block p-4 rounded-lg bg-muted text-sm font-mono overflow-x-auto"
                {...props}
              />
            ),
          pre: ({ node, ...props }) => (
            <pre
              className="p-4 rounded-lg bg-muted overflow-x-auto my-4"
              {...props}
            />
          ),
          a: ({ node, ...props }) => (
            <a
              className="text-primary hover:underline"
              target="_blank"
              rel="noopener noreferrer"
              {...props}
            />
          ),
          blockquote: ({ node, ...props }) => (
            <blockquote
              className="border-l-4 border-muted-foreground/20 pl-4 italic my-4"
              {...props}
            />
          ),
          table: ({ node, ...props }) => (
            <div className="my-4 overflow-x-auto">
              <table className="min-w-full divide-y divide-border" {...props} />
            </div>
          ),
          thead: ({ node, ...props }) => (
            <thead className="bg-muted" {...props} />
          ),
          tbody: ({ node, ...props }) => (
            <tbody className="divide-y divide-border" {...props} />
          ),
          tr: ({ node, ...props }) => (
            <tr className="hover:bg-muted/50" {...props} />
          ),
          th: ({ node, ...props }) => (
            <th className="px-4 py-2 text-left font-medium" {...props} />
          ),
          td: ({ node, ...props }) => (
            <td className="px-4 py-2" {...props} />
          ),
          hr: ({ node, ...props }) => (
            <hr className="my-6 border-t border-muted" {...props} />
          ),
        }}
      >
        {markdownContent}
      </ReactMarkdown>
    </div>
  )
}

export function SpecDetailViewer({ spec, open, onOpenChange }: SpecDetailViewerProps) {
  const [activeTab, setActiveTab] = useState('spec')
  const [content, setContent] = useState<SpecFileContent | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const fetchSpecContent = useCallback(async (specId: string) => {
    setLoading(true)
    setError(null)
    try {
      const response = await fetch(`/api/specs/${specId}/content`)
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      const data = await response.json()
      setContent(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch SPEC content')
      console.error('Error fetching spec content:', err)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    if (open && spec) {
      fetchSpecContent(spec.id)
    }
  }, [open, spec, fetchSpecContent])

  if (!spec) {
    return null
  }

  const currentContent = content?.[activeTab as keyof SpecFileContent] || ''

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-4xl max-h-[90vh] p-0">
        {/* Header */}
        <DialogHeader className="p-6 pb-4 border-b">
          <div className="flex items-start justify-between gap-4">
            <div className="flex-1 min-w-0">
              <DialogTitle className="text-xl font-semibold truncate">
                {spec.title}
              </DialogTitle>
              <DialogDescription className="mt-1">
                <div className="flex items-center gap-3">
                  <span className="font-mono text-xs">{spec.id}</span>
                  <Badge variant="outline" className="text-xs">
                    {spec.status}
                  </Badge>
                  {spec.priority === 'high' && (
                    <Badge variant="destructive" className="text-xs">
                      High Priority
                    </Badge>
                  )}
                </div>
              </DialogDescription>
            </div>
          </div>
        </DialogHeader>

        {/* Tabs */}
        <div className="px-6 pt-4">
          <Tabs value={activeTab} onValueChange={setActiveTab}>
            <TabsList className="w-full justify-start">
              {tabConfig.map((tab) => {
                const Icon = tab.icon
                const hasContent = content?.[tab.value as keyof SpecFileContent]

                return (
                  <TabsTrigger key={tab.value} value={tab.value} className="gap-2">
                    <Icon className="h-4 w-4" />
                    <span className="hidden sm:inline">{tab.label}</span>
                    <span className="sm:hidden">{tab.label.split(' ')[0]}</span>
                    {!hasContent && !loading && (
                      <span className="ml-1 text-xs text-muted-foreground">(empty)</span>
                    )}
                  </TabsTrigger>
                )
              })}
            </TabsList>

            {/* Tab Content */}
            <ScrollArea className="h-[60vh] mt-4 pr-4">
              {loading ? (
                <LoadingState />
              ) : error ? (
                <ErrorState error={error} />
              ) : (
                <div className="pr-4">
                  <MarkdownContent content={currentContent} />
                </div>
              )}
            </ScrollArea>
          </Tabs>
        </div>

        {/* Footer */}
        <div className="p-4 border-t bg-muted/30">
          <div className="flex items-center justify-between text-xs text-muted-foreground">
            <div>
              {spec.filePath && (
                <span className="font-mono">{spec.filePath}</span>
              )}
            </div>
            <div>
              Last updated: {new Date(spec.updatedAt).toLocaleString()}
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
