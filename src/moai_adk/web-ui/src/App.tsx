import { useState, useEffect } from 'react'
import { AppShell } from '@/components/layout'
import { ChatView } from '@/components/chat'
import { SpecList } from '@/components/spec'
import { TerminalView } from '@/components/terminal'
import { CostView } from '@/components/cost'
import { useUIStore } from '@/stores'
import { useWebSocket } from '@/hooks'

function App() {
  const { activeTab } = useUIStore()
  const [isConnected, setIsConnected] = useState(false)

  // Main WebSocket connection for app-wide events
  const wsUrl = `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/ws`

  const { isConnected: wsConnected } = useWebSocket({
    url: wsUrl,
    onOpen: () => setIsConnected(true),
    onClose: () => setIsConnected(false),
    reconnect: true,
    reconnectInterval: 3000,
  })

  useEffect(() => {
    setIsConnected(wsConnected)
  }, [wsConnected])

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
