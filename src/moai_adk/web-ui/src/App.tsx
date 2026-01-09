import { useState, useEffect, useCallback } from 'react'
import { AppShell } from '@/components/layout'
import { ChatView } from '@/components/chat'
import { SpecList } from '@/components/spec'
import { TerminalView } from '@/components/terminal'
import { CostView } from '@/components/cost'
import { useUIStore } from '@/stores'

function App() {
  const { activeTab } = useUIStore()
  const [isConnected, setIsConnected] = useState(false)

  // Check server health via REST API polling
  const checkHealth = useCallback(async () => {
    try {
      const response = await fetch('/api/health')
      setIsConnected(response.ok)
    } catch {
      setIsConnected(false)
    }
  }, [])

  useEffect(() => {
    // Initial health check
    checkHealth()

    // Poll health every 5 seconds
    const interval = setInterval(checkHealth, 5000)
    return () => clearInterval(interval)
  }, [checkHealth])

  const renderContent = () => {
    switch (activeTab) {
      case 'chat':
        return <ChatView />
      case 'specs':
        return <SpecList />
      case 'terminal':
        return <TerminalView />
      case 'costs':
        return <CostView />
      default:
        return <ChatView />
    }
  }

  return (
    <AppShell isConnected={isConnected}>
      {renderContent()}
    </AppShell>
  )
}

export default App
