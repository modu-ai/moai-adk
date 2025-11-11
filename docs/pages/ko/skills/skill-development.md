# Skill 개발 가이드

## 개요

MoAI-ADK의 Skills는 재사용 가능한 지식 캡슐입니다. 이 가이드는 새로운 Skill을 처음부터 만드는 방법, 표준 구조, 베스트 프랙티스, 그리고 Context7 통합 패턴을 다룹니다.

**Skill 철학**:
- **작고 집중적**: 하나의 명확한 목적
- **재사용 가능**: 여러 상황에서 활용
- **자기 문서화**: 명확한 사용 예제 포함
- **검증 가능**: 자동 검증 통과
- **진화 가능**: Context7로 최신 상태 유지

## Skill 구조 표준

### 1. 필수 구성 요소

```yaml
---
# YAML Frontmatter (필수)
name: skill-name                    # 고유 식별자
version: 1.0.0                      # 시맨틱 버전
status: production                  # production | beta | experimental
description: Brief description      # 한 줄 설명
category: foundation               # foundation | workflow | baas | pattern | advanced
tags: [tag1, tag2]                 # 분류 태그
author: MoAI Team                  # 작성자
created: 2024-01-01                # 생성일
updated: 2024-01-15                # 최종 수정일
dependencies: []                    # 의존 Skills
context7:
  enabled: true                     # Context7 통합 여부
  references: []                    # Context7 라이브러리 참조
---

# Skill Title

## @TAG-XXXX-001: Purpose

명확한 목적 설명 (2-3 문장)

## @TAG-XXXX-002: Core Concepts

핵심 개념 설명 (3-5개의 주요 개념)

## @TAG-XXXX-003: Usage Examples

실제 사용 예제 (3-5개의 시나리오)

## @TAG-XXXX-004: Best Practices

권장 사항 (5-7개의 베스트 프랙티스)

## @TAG-XXXX-005: Common Pitfalls

주의사항 및 안티패턴

## @TAG-XXXX-006: Integration

다른 Skills/도구와의 통합 방법

## @TAG-XXXX-007: Troubleshooting

일반적인 문제 및 해결 방법

## @TAG-XXXX-008: References

관련 문서 및 리소스
```

### 2. YAML Frontmatter 상세

```yaml
---
# 기본 정보
name: moai-pattern-serverless-api       # kebab-case, 고유
version: 2.1.0                          # major.minor.patch
status: production                      # 상태
description: >                          # 여러 줄 설명 가능
  Serverless API 패턴 구현 가이드.
  AWS Lambda, Vercel Functions, Cloudflare Workers 지원.

# 분류
category: pattern                       # 카테고리
tags:
  - serverless
  - api
  - cloud
  - scalability
subcategory: backend                    # 하위 카테고리

# 메타데이터
author: MoAI Team
contributors:
  - Jane Doe
  - John Smith
created: 2024-01-01
updated: 2024-03-15
license: MIT

# 의존성
dependencies:
  - moai-foundation-tags               # 필수 Skill
  - moai-workflow-tdd                  # 권장 Skill
related:
  - moai-pattern-microservices         # 관련 Skill
  - moai-baas-vercel

# Context7 통합
context7:
  enabled: true
  references:
    - library: /vercel/next.js
      topics: [serverless-functions, api-routes]
      tokens: 3000
    - library: /aws/lambda
      topics: [runtime, cold-start]
      tokens: 2000
  update_frequency: weekly              # 업데이트 빈도

# 검증 설정
validation:
  strict: true                          # 엄격 모드
  min_score: 85                         # 최소 점수
  auto_fix: safe                        # 자동 수정 레벨
  trust5_required: true                 # TRUST 5 필수

# 사용 통계 (자동 수집)
usage:
  invocations: 1234
  success_rate: 0.98
  avg_execution_time: 1.2s
---
```

## 새 Skill 만들기

### 1. CLI로 생성

```bash
# 기본 템플릿으로 생성
moai-adk create skill \
  --name moai-custom-api-gateway \
  --category pattern \
  --tags api,gateway,routing

# Context7 통합 포함
moai-adk create skill \
  --name moai-custom-graphql \
  --category pattern \
  --context7 /graphql-foundation/graphql-js \
  --template advanced

# 대화형 생성
moai-adk create skill --interactive
```

### 2. 수동으로 생성

```bash
# 1. 파일 생성
touch .claude/skills/moai-custom-websocket-server.md

# 2. 템플릿 복사
moai-adk template skill > .claude/skills/moai-custom-websocket-server.md

# 3. 편집
vim .claude/skills/moai-custom-websocket-server.md
```

### 3. 실제 예제: WebSocket Skill 만들기

```markdown
---
name: moai-pattern-websocket-server
version: 1.0.0
status: production
description: Real-time WebSocket 서버 구현 가이드
category: pattern
tags: [websocket, real-time, communication]
author: MoAI Team
created: 2024-01-20
context7:
  enabled: true
  references:
    - library: /ws/ws
      topics: [websocket-server, client-connection]
      tokens: 2500
---

# WebSocket Server Pattern

## @TAG-WS-001: Purpose

실시간 양방향 통신을 위한 WebSocket 서버 구현 패턴입니다. 채팅, 알림, 협업 도구 등 실시간 기능이 필요한 애플리케이션에 적합합니다.

## @TAG-WS-002: Core Concepts

### 1. Connection Management
```typescript
// 연결 관리 시스템
class ConnectionManager {
  private connections: Map<string, WebSocket> = new Map()

  register(id: string, ws: WebSocket) {
    this.connections.set(id, ws)
    console.log(`Client ${id} connected. Total: ${this.connections.size}`)
  }

  unregister(id: string) {
    this.connections.delete(id)
    console.log(`Client ${id} disconnected`)
  }

  broadcast(message: string, exclude?: string) {
    this.connections.forEach((ws, id) => {
      if (id !== exclude && ws.readyState === WebSocket.OPEN) {
        ws.send(message)
      }
    })
  }
}
```

### 2. Message Protocol
```typescript
// 메시지 프로토콜 정의
interface WSMessage {
  type: 'message' | 'command' | 'event'
  payload: any
  timestamp: number
  sender: string
}

function parseMessage(data: string): WSMessage {
  const parsed = JSON.parse(data)
  return {
    type: parsed.type || 'message',
    payload: parsed.payload,
    timestamp: Date.now(),
    sender: parsed.sender,
  }
}
```

### 3. Heartbeat Mechanism
```typescript
// 연결 상태 확인
function setupHeartbeat(ws: WebSocket) {
  const interval = setInterval(() => {
    if (ws.readyState === WebSocket.OPEN) {
      ws.ping()
    } else {
      clearInterval(interval)
    }
  }, 30000) // 30초마다

  ws.on('pong', () => {
    console.log('Client is alive')
  })
}
```

## @TAG-WS-003: Usage Examples

### Example 1: 기본 WebSocket 서버
```typescript
import { WebSocketServer } from 'ws'

const wss = new WebSocketServer({ port: 8080 })
const connectionManager = new ConnectionManager()

wss.on('connection', (ws, req) => {
  const clientId = generateClientId()

  connectionManager.register(clientId, ws)
  setupHeartbeat(ws)

  ws.on('message', (data) => {
    const message = parseMessage(data.toString())

    // 브로드캐스트
    connectionManager.broadcast(
      JSON.stringify(message),
      clientId
    )
  })

  ws.on('close', () => {
    connectionManager.unregister(clientId)
  })
})
```

### Example 2: 채팅 애플리케이션
```typescript
// 채팅방 관리
class ChatRoom {
  private rooms: Map<string, Set<string>> = new Map()

  join(roomId: string, clientId: string) {
    if (!this.rooms.has(roomId)) {
      this.rooms.set(roomId, new Set())
    }
    this.rooms.get(roomId)!.add(clientId)
  }

  leave(roomId: string, clientId: string) {
    this.rooms.get(roomId)?.delete(clientId)
  }

  broadcast(roomId: string, message: string, exclude?: string) {
    const room = this.rooms.get(roomId)
    if (room) {
      room.forEach((clientId) => {
        if (clientId !== exclude) {
          connectionManager.send(clientId, message)
        }
      })
    }
  }
}

// 사용
wss.on('connection', (ws) => {
  ws.on('message', (data) => {
    const { type, payload } = parseMessage(data.toString())

    switch (type) {
      case 'join':
        chatRoom.join(payload.roomId, clientId)
        break
      case 'message':
        chatRoom.broadcast(payload.roomId, data.toString(), clientId)
        break
      case 'leave':
        chatRoom.leave(payload.roomId, clientId)
        break
    }
  })
})
```

### Example 3: 인증 및 권한
```typescript
import jwt from 'jsonwebtoken'

wss.on('connection', async (ws, req) => {
  // URL에서 토큰 추출
  const token = new URL(req.url!, 'ws://localhost').searchParams.get('token')

  try {
    // JWT 검증
    const payload = jwt.verify(token!, process.env.JWT_SECRET!)

    // 사용자 정보 저장
    ws.userId = payload.sub
    ws.role = payload.role

    // 권한 확인
    ws.on('message', (data) => {
      const message = parseMessage(data.toString())

      if (requiresAdmin(message.type) && ws.role !== 'admin') {
        ws.send(JSON.stringify({ error: 'Unauthorized' }))
        return
      }

      // 메시지 처리...
    })
  } catch (error) {
    ws.close(1008, 'Unauthorized')
  }
})
```

## @TAG-WS-004: Best Practices

1. **연결 관리**
   - 클라이언트 ID 생성 및 추적
   - 연결 풀 크기 제한
   - 재연결 로직 구현

2. **메시지 검증**
   - 모든 메시지 타입 검증
   - 페이로드 크기 제한
   - Rate limiting 적용

3. **에러 처리**
   - Graceful degradation
   - 자동 재연결
   - 에러 로깅 및 모니터링

4. **보안**
   - TLS/SSL 사용 (wss://)
   - 인증 토큰 검증
   - CORS 설정

5. **확장성**
   - Redis Pub/Sub for clustering
   - Load balancer 설정
   - Horizontal scaling

6. **성능 최적화**
   - 메시지 압축
   - Binary protocol 고려
   - Connection pooling

7. **모니터링**
   - 연결 수 추적
   - 메시지 처리 시간 측정
   - 에러율 모니터링

## @TAG-WS-005: Common Pitfalls

### 1. 메모리 누수
```typescript
// ❌ 나쁜 예
ws.on('message', (data) => {
  // 이벤트 리스너가 누적됨
  ws.on('close', () => {
    cleanup()
  })
})

// ✅ 좋은 예
ws.once('close', () => {
  cleanup()
})
```

### 2. Heartbeat 미구현
```typescript
// ❌ Dead connections 방치
// 클라이언트가 비정상 종료 시 연결이 계속 유지됨

// ✅ Heartbeat으로 감지
setupHeartbeat(ws)
```

### 3. 에러 처리 부재
```typescript
// ❌ 에러 무시
ws.on('message', (data) => {
  const message = JSON.parse(data) // 파싱 에러 가능
})

// ✅ 에러 처리
ws.on('message', (data) => {
  try {
    const message = JSON.parse(data.toString())
    processMessage(message)
  } catch (error) {
    ws.send(JSON.stringify({ error: 'Invalid message format' }))
  }
})

ws.on('error', (error) => {
  console.error('WebSocket error:', error)
})
```

## @TAG-WS-006: Integration

### Next.js API Routes
```typescript
// pages/api/socket.ts
import { NextApiRequest } from 'next'
import { WebSocketServer } from 'ws'

export const config = {
  api: {
    bodyParser: false,
  },
}

const wss = new WebSocketServer({ noServer: true })

export default function handler(req: NextApiRequest, res: any) {
  if (req.socket.server.wss) {
    console.log('WebSocket server already running')
    res.end()
    return
  }

  req.socket.server.wss = wss

  req.socket.server.on('upgrade', (request, socket, head) => {
    wss.handleUpgrade(request, socket, head, (ws) => {
      wss.emit('connection', ws, request)
    })
  })

  res.end()
}
```

### Redis Pub/Sub (클러스터링)
```typescript
import Redis from 'ioredis'

const pub = new Redis()
const sub = new Redis()

// 메시지 구독
sub.subscribe('chat-messages')

sub.on('message', (channel, message) => {
  // 모든 서버 인스턴스에 브로드캐스트
  connectionManager.broadcast(message)
})

// 메시지 발행
wss.on('connection', (ws) => {
  ws.on('message', (data) => {
    pub.publish('chat-messages', data.toString())
  })
})
```

## @TAG-WS-007: Troubleshooting

### 문제: Connection timeout
```typescript
// 해결: Keep-alive 설정
const server = http.createServer()
server.keepAliveTimeout = 65000
server.headersTimeout = 66000
```

### 문제: Too many connections
```typescript
// 해결: Connection limiting
const MAX_CONNECTIONS = 10000

wss.on('connection', (ws) => {
  if (connectionManager.size >= MAX_CONNECTIONS) {
    ws.close(1008, 'Server full')
    return
  }
  // 연결 처리...
})
```

### 문제: Memory leak
```typescript
// 해결: Proper cleanup
ws.on('close', () => {
  clearInterval(heartbeatInterval)
  connectionManager.unregister(clientId)
  ws.removeAllListeners()
})
```

## @TAG-WS-008: References

- [WebSocket RFC 6455](https://datatracker.ietf.org/doc/html/rfc6455)
- [ws Library Documentation](https://github.com/websockets/ws)
- [Socket.IO Documentation](https://socket.io/docs/)
- [Context7: /ws/ws](https://context7.com/ws/ws)

## Related Skills

- @REF(moai-pattern-real-time-backend)
- @REF(moai-baas-convex)
- @REF(moai-workflow-testing)
```

## Context7 통합 패턴

### 1. 기본 통합

```yaml
---
context7:
  enabled: true
  references:
    - library: /vercel/next.js
      topics: [api-routes, middleware]
      tokens: 3000
---
```

### 2. 다중 라이브러리 참조

```yaml
context7:
  enabled: true
  references:
    - library: /supabase/supabase
      topics: [authentication, database]
      tokens: 2500
    - library: /stripe/stripe-node
      topics: [payments, webhooks]
      tokens: 2000
    - library: /react/react
      topics: [hooks, components]
      tokens: 1500
  update_frequency: weekly
  cache_duration: 3600
```

### 3. 동적 Context7 로딩

```typescript
// Skill 내에서 Context7 활용
class SkillExecutor {
  async execute(skill: Skill) {
    if (skill.context7?.enabled) {
      // Context7에서 최신 정보 가져오기
      for (const ref of skill.context7.references) {
        const docs = await context7.getDocs(
          ref.library,
          ref.topics,
          ref.tokens
        )

        // Skill에 최신 정보 주입
        skill.updateWithContext(docs)
      }
    }

    return skill.render()
  }
}
```

## 테스팅

### 1. Skill 검증

```bash
# 단일 Skill 검증
moai-adk validate skill \
  --path .claude/skills/moai-custom-websocket-server.md

# 결과:
# ✓ YAML frontmatter valid
# ✓ All required sections present
# ✓ TAG chain complete
# ✓ TRUST 5 compliant
# ✓ No duplicate TAGs
# Score: 94.5 (Grade A)
```

### 2. 사용 테스트

```typescript
// tests/skills/websocket-server.test.ts
import { SkillExecutor } from '@moai/core'

describe('WebSocket Server Skill', () => {
  it('should execute successfully', async () => {
    const skill = await SkillExecutor.load(
      'moai-pattern-websocket-server'
    )

    const result = await skill.execute({
      context: {
        framework: 'next.js',
        deployment: 'vercel',
      },
    })

    expect(result.status).toBe('success')
    expect(result.output).toContain('WebSocketServer')
  })

  it('should integrate with Context7', async () => {
    const skill = await SkillExecutor.load(
      'moai-pattern-websocket-server'
    )

    // Context7 정보 포함 확인
    expect(skill.context7?.enabled).toBe(true)
    expect(skill.context7?.references).toHaveLength(1)
  })
})
```

## 배포 및 공유

### 1. Skill 등록

```bash
# MoAI Skills Registry에 등록
moai-adk publish skill \
  --path .claude/skills/moai-custom-websocket-server.md \
  --registry official \
  --public

# 비공개 레지스트리
moai-adk publish skill \
  --path .claude/skills/moai-custom-api.md \
  --registry private \
  --organization my-company
```

### 2. 버전 관리

```bash
# 버전 업데이트
moai-adk bump skill \
  --name moai-custom-websocket-server \
  --type minor  # major | minor | patch

# 변경 이력 생성
moai-adk changelog skill \
  --name moai-custom-websocket-server \
  --from v1.0.0 \
  --to v1.1.0
```

## Best Practices 요약

### 1. 명확한 목적
- 하나의 Skill은 하나의 명확한 목적
- 복잡한 경우 여러 Skill로 분리

### 2. 풍부한 예제
- 최소 3개의 실제 사용 예제
- 다양한 난이도 커버

### 3. 완벽한 문서화
- 모든 TAG에 컨텍스트 제공
- 크로스 레퍼런스 명확히

### 4. TRUST 5 준수
- Test First: 테스트 예제 포함
- Readable: 명확한 구조와 주석
- Unified: 일관된 스타일
- Secured: 보안 베스트 프랙티스
- Trackable: 완전한 TAG 체인

### 5. Context7 활용
- 최신 라이브러리 패턴 참조
- 정기적인 업데이트

### 6. 검증 통과
- 자동 검증 85점 이상
- A등급 이상 유지

## 다음 단계

- [Validation System](/ko/skills/validation-system) - 품질 검증
- [Advanced Skills](/ko/skills/advanced-skills) - 고급 기능
- [Foundation Skills](/ko/skills/foundation) - 기본 Skills 참고
- [Context7 통합](/ko/skills/context7-integration) - 최신 패턴 활용

## 참고 자료

- [MoAI-ADK Skill API](https://moai-adk.dev/api/skills)
- [Skills Registry](https://skills.moai-adk.dev)
- [Context7 가이드](https://context7.com/docs)
- [YAML Specification](https://yaml.org/spec/)
