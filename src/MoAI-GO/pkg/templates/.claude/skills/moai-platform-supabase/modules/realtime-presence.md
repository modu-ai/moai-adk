---
name: realtime-presence
description: Real-time subscriptions and presence tracking for collaborative features
parent-skill: moai-platform-supabase
version: 1.0.0
updated: 2026-01-06
---

# Real-time and Presence Module

## Overview

Supabase provides real-time capabilities through Postgres Changes (database change notifications) and Presence (user state tracking) for building collaborative applications.

## Postgres Changes Subscription

### Basic Setup

Subscribe to all changes on a table:

```typescript
import { createClient } from '@supabase/supabase-js'

const supabase = createClient(SUPABASE_URL, SUPABASE_ANON_KEY)

const channel = supabase.channel('db-changes')
  .on('postgres_changes',
    { event: '*', schema: 'public', table: 'messages' },
    (payload) => console.log('Change:', payload)
  )
  .subscribe()
```

### Event Types

Available events:
- `INSERT` - New row added
- `UPDATE` - Row modified
- `DELETE` - Row removed
- `*` - All events

### Filtered Subscriptions

Filter changes by specific conditions:

```typescript
supabase.channel('project-updates')
  .on('postgres_changes',
    {
      event: 'UPDATE',
      schema: 'public',
      table: 'projects',
      filter: `id=eq.${projectId}`
    },
    (payload) => handleProjectUpdate(payload.new)
  )
  .subscribe()
```

### Multiple Tables

Subscribe to multiple tables on one channel:

```typescript
const channel = supabase.channel('app-changes')
  .on('postgres_changes',
    { event: '*', schema: 'public', table: 'tasks' },
    handleTaskChange
  )
  .on('postgres_changes',
    { event: '*', schema: 'public', table: 'comments' },
    handleCommentChange
  )
  .subscribe()
```

## Presence Tracking

### Presence State Interface

```typescript
interface PresenceState {
  user_id: string
  online_at: string
  typing?: boolean
  cursor?: { x: number; y: number }
}
```

### Channel Setup with Presence

```typescript
const channel = supabase.channel('room:collaborative-doc', {
  config: { presence: { key: userId } }
})

channel
  .on('presence', { event: 'sync' }, () => {
    const state = channel.presenceState<PresenceState>()
    console.log('Online users:', Object.keys(state))
  })
  .on('presence', { event: 'join' }, ({ key, newPresences }) => {
    console.log('User joined:', key, newPresences)
  })
  .on('presence', { event: 'leave' }, ({ key, leftPresences }) => {
    console.log('User left:', key, leftPresences)
  })
  .subscribe(async (status) => {
    if (status === 'SUBSCRIBED') {
      await channel.track({
        user_id: userId,
        online_at: new Date().toISOString()
      })
    }
  })
```

### Update Presence State

Update user presence in real-time:

```typescript
// Track typing status
await channel.track({ typing: true })

// Track cursor position
await channel.track({ cursor: { x: 100, y: 200 } })

// Clear typing after timeout
setTimeout(async () => {
  await channel.track({ typing: false })
}, 1000)
```

## Collaborative Features

### Collaborative Cursors

```typescript
interface CursorState {
  user_id: string
  user_name: string
  cursor: { x: number; y: number }
  color: string
}

function setupCollaborativeCursors(documentId: string, userId: string, userName: string) {
  const channel = supabase.channel(`cursors:${documentId}`, {
    config: { presence: { key: userId } }
  })

  const colors = ['#FF6B6B', '#4ECDC4', '#45B7D1', '#96CEB4', '#FFEAA7']
  const userColor = colors[Math.abs(userId.hashCode()) % colors.length]

  channel
    .on('presence', { event: 'sync' }, () => {
      const state = channel.presenceState<CursorState>()
      renderCursors(Object.values(state).flat())
    })
    .subscribe(async (status) => {
      if (status === 'SUBSCRIBED') {
        await channel.track({
          user_id: userId,
          user_name: userName,
          cursor: { x: 0, y: 0 },
          color: userColor
        })
      }
    })

  // Track mouse movement
  document.addEventListener('mousemove', async (e) => {
    await channel.track({
      user_id: userId,
      user_name: userName,
      cursor: { x: e.clientX, y: e.clientY },
      color: userColor
    })
  })

  return channel
}
```

### Live Editing Indicators

```typescript
interface EditingState {
  user_id: string
  user_name: string
  editing_field: string | null
}

function setupFieldLocking(formId: string) {
  const channel = supabase.channel(`form:${formId}`, {
    config: { presence: { key: currentUserId } }
  })

  channel
    .on('presence', { event: 'sync' }, () => {
      const state = channel.presenceState<EditingState>()
      updateFieldLocks(Object.values(state).flat())
    })
    .subscribe()

  return {
    startEditing: async (fieldName: string) => {
      await channel.track({
        user_id: currentUserId,
        user_name: currentUserName,
        editing_field: fieldName
      })
    },
    stopEditing: async () => {
      await channel.track({
        user_id: currentUserId,
        user_name: currentUserName,
        editing_field: null
      })
    }
  }
}
```

## Broadcast Messages

Send arbitrary messages to channel subscribers:

```typescript
const channel = supabase.channel('room:chat')

// Subscribe to broadcasts
channel
  .on('broadcast', { event: 'message' }, ({ payload }) => {
    console.log('Received:', payload)
  })
  .subscribe()

// Send broadcast
await channel.send({
  type: 'broadcast',
  event: 'message',
  payload: { text: 'Hello everyone!', sender: userId }
})
```

## Subscription Management

### Unsubscribe

```typescript
// Unsubscribe from specific channel
await supabase.removeChannel(channel)

// Unsubscribe from all channels
await supabase.removeAllChannels()
```

### Subscription Status

```typescript
channel.subscribe((status) => {
  switch (status) {
    case 'SUBSCRIBED':
      console.log('Connected to channel')
      break
    case 'CLOSED':
      console.log('Channel closed')
      break
    case 'CHANNEL_ERROR':
      console.log('Channel error')
      break
    case 'TIMED_OUT':
      console.log('Connection timed out')
      break
  }
})
```

## React Integration

### Custom Hook for Presence

```typescript
import { useEffect, useState } from 'react'
import { supabase } from './supabase/client'

export function usePresence<T>(channelName: string, userId: string, initialState: T) {
  const [presences, setPresences] = useState<Record<string, T[]>>({})

  useEffect(() => {
    const channel = supabase.channel(channelName, {
      config: { presence: { key: userId } }
    })

    channel
      .on('presence', { event: 'sync' }, () => {
        setPresences(channel.presenceState<T>())
      })
      .subscribe(async (status) => {
        if (status === 'SUBSCRIBED') {
          await channel.track(initialState)
        }
      })

    return () => {
      supabase.removeChannel(channel)
    }
  }, [channelName, userId])

  const updatePresence = async (state: Partial<T>) => {
    const channel = supabase.getChannels().find(c => c.topic === channelName)
    if (channel) {
      await channel.track({ ...initialState, ...state } as T)
    }
  }

  return { presences, updatePresence }
}
```

### Usage

```typescript
function CollaborativeEditor({ documentId, userId }) {
  const { presences, updatePresence } = usePresence(
    `doc:${documentId}`,
    userId,
    { user_id: userId, typing: false, cursor: null }
  )

  return (
    <div>
      {Object.values(presences).flat().map(p => (
        <Cursor key={p.user_id} position={p.cursor} />
      ))}
    </div>
  )
}
```

## Context7 Query Examples

For latest real-time documentation:

Topic: "realtime postgres_changes subscription"
Topic: "presence tracking channel"
Topic: "broadcast messages supabase"

---

Related Modules:
- typescript-patterns.md - Client architecture
- auth-integration.md - Authenticated subscriptions
