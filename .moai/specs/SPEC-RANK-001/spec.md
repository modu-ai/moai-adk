# SPEC-RANK-001: MoAI Token Rank Service

## Metadata

| Field | Value |
|-------|-------|
| SPEC ID | SPEC-RANK-001 |
| Title | MoAI Token Rank Leaderboard Service |
| Version | 1.0.0 |
| Status | Draft |
| Created | 2026-01-11 |
| Author | GOOS |
| Domain | Full-Stack (Frontend + Backend + CLI) |

---

## 1. Executive Summary

MoAI Token Rank는 MoAI-ADK 사용자들의 Claude Code 토큰 사용량을 수집하고,
복합 점수 기반 랭킹을 제공하는 서비스입니다.

### Goals
1. MoAI 프로젝트(.moai 폴더 존재)에서의 토큰 사용량 자동 수집
2. 복합 점수 기반 공정한 랭킹 시스템 (토큰량 40%, 효율성 25%, 세션 수 20%, 일관성 15%)
3. GitHub OAuth 인증을 통한 사용자 식별
4. 실시간 리더보드 및 개인 통계 대시보드

### Non-Goals
- 코드 내용 또는 프롬프트/응답 수집 (개인정보 보호)
- 유료 기능 또는 구독 모델
- 다른 AI 서비스 (GPT, Gemini 등) 통합

---

## 2. Requirements (EARS Format)

### 2.1 Functional Requirements

#### FR-001: Token Collection Hook
**When** a Claude Code session ends in a MoAI project (containing .moai folder),
**the system shall** automatically collect token usage data (input, output, cache tokens) and submit to the rank.mo.ai.kr API.

**Acceptance Criteria:**
- [ ] Session end hook executes within 5 seconds
- [ ] Only MoAI projects (.moai folder exists) trigger collection
- [ ] HMAC-signed request with timestamp prevents replay attacks
- [ ] Failed submissions are queued for retry (max 3 attempts)

#### FR-002: User Registration
**When** a user runs `moai-adk rank register`,
**the system shall** open a browser for GitHub OAuth authentication and generate a unique API key.

**Acceptance Criteria:**
- [ ] GitHub OAuth flow completes in browser
- [ ] API key is stored securely (OS keychain or encrypted file)
- [ ] User receives confirmation with rank.mo.ai.kr profile link

#### FR-003: Leaderboard Display
**When** a user visits rank.mo.ai.kr,
**the system shall** display a real-time leaderboard with user rankings based on composite scores.

**Acceptance Criteria:**
- [ ] Leaderboard shows: rank, username, avatar, composite score, total tokens, session count
- [ ] Period filters available: daily, weekly, monthly, all-time
- [ ] Current user highlighted if authenticated
- [ ] Privacy mode users shown as "Anonymous" with score visible

#### FR-004: User Dashboard
**When** an authenticated user visits their dashboard,
**the system shall** display personal statistics including token usage trends, ranking history, and API key management.

**Acceptance Criteria:**
- [ ] Token usage chart (daily/weekly/monthly views)
- [ ] Current rank and percentile
- [ ] API key regeneration option
- [ ] Privacy mode toggle

#### FR-005: Composite Score Calculation
**The system shall** calculate composite scores using the following weighted formula:
- Total Tokens: 40% (log10 scale)
- Token Efficiency (output/input ratio): 25%
- Session Count: 20% (log10 scale)
- Streak Days (consecutive usage): 15%

**Acceptance Criteria:**
- [ ] Score updates within 1 minute of session submission
- [ ] Daily aggregates computed at midnight UTC
- [ ] Rankings recalculated after each score update

#### FR-006: CLI Status Commands
**When** a user runs `moai-adk rank status`,
**the system shall** display current rank, total tokens, and recent session summary.

**Acceptance Criteria:**
- [ ] Shows rank position and percentile
- [ ] Shows this week's token usage
- [ ] Shows last 5 sessions summary

### 2.2 Non-Functional Requirements

#### NFR-001: Security
- API keys stored as SHA-256 hash only (never plaintext)
- HMAC signature required for all CLI-to-API requests
- Rate limiting: 100 sessions/hour, 30 leaderboard requests/minute
- OS secure storage (Keychain/libsecret/Credential Manager) for local API key

#### NFR-002: Performance
- Leaderboard page load: < 2 seconds
- API response time: < 500ms (p95)
- Session submission: < 1 second

#### NFR-003: Privacy
- No code content collection
- No prompt/response collection
- Project paths hashed with user-specific salt
- Privacy mode hides username from leaderboard

#### NFR-004: Availability
- 99.9% uptime target
- Graceful degradation if database unavailable
- Offline-first CLI (queue submissions for later)

---

## 3. Technical Architecture

### 3.1 Technology Stack

| Component | Technology |
|-----------|------------|
| Frontend | Next.js 16, React 19, Tailwind CSS, shadcn/ui |
| Backend | Vercel Serverless Functions (Edge Runtime) |
| Database | Neon PostgreSQL (Serverless) |
| Auth | NextAuth.js v5 + GitHub OAuth |
| Rate Limiting | Upstash Redis |
| Hosting | Vercel (rank.mo.ai.kr) |
| CLI | Python (moai-adk package) |

### 3.2 Repository Structure

```
moai-rank/                      # ~/MoAI/moai-rank
├── apps/
│   └── web/                    # Next.js 16 application
│       ├── src/
│       │   ├── app/            # App Router pages
│       │   ├── components/     # React components
│       │   └── lib/            # Utilities
│       ├── drizzle/            # DB migrations
│       └── package.json
├── packages/
│   └── shared/                 # Shared types/schemas
├── .moai/
│   └── specs/                  # SPEC documents
├── pnpm-workspace.yaml
├── turbo.json
└── vercel.json
```

### 3.3 Database Schema

See design document for full schema. Key tables:
- `users`: GitHub info, API key hash, user salt
- `sessions`: Server-calculated hash, anonymous project ID
- `token_usage`: Input/output/cache token counts
- `daily_aggregates`: Pre-computed daily stats
- `rankings`: Period-based rankings
- `security_audit_log`: Security events

### 3.4 API Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | /api/leaderboard | Public | Get rankings |
| GET | /api/users/[username] | Public | User profile |
| POST | /api/v1/sessions | API Key + HMAC | Submit session |
| GET | /api/v1/rank | API Key | Current rank |
| GET | /api/me | Session | User dashboard data |
| POST | /api/me/regenerate-key | Session | Regenerate API key |

---

## 4. Security Measures

### 4.1 Critical (Must Have)
- [x] API key generation: 256-bit entropy (HMAC-SHA256)
- [x] Local storage: OS secure keychain integration
- [x] Request signing: HMAC with timestamp (5-minute window)
- [x] Server-side hash recalculation (don't trust client)

### 4.2 High Priority
- [x] Rate limiting with burst protection
- [x] Security audit logging
- [x] Session timeout and re-authentication
- [x] API key regeneration mechanism

### 4.3 Standard
- [x] Security headers (CSP, HSTS, X-Frame-Options)
- [x] Row Level Security in PostgreSQL
- [x] Input validation with Zod schemas

---

## 5. Implementation Phases

### Phase 1: Infrastructure (Week 1)
- [ ] Neon DB setup + schema migration
- [ ] Next.js 16 project initialization
- [ ] NextAuth.js + GitHub OAuth
- [ ] Upstash Redis for rate limiting
- [ ] Vercel deployment

### Phase 2: Backend API (Week 2)
- [ ] Session submission endpoint with HMAC validation
- [ ] Leaderboard query with period filters
- [ ] User profile and dashboard APIs
- [ ] Composite score calculation job

### Phase 3: Frontend (Week 3)
- [ ] Leaderboard page with real-time updates
- [ ] User dashboard with charts
- [ ] Authentication flow
- [ ] Responsive design

### Phase 4: CLI Integration (Week 4)
- [ ] moai-adk rank commands (in MoAI-ADK repo)
- [ ] Secure storage integration
- [ ] Session collection hook
- [ ] Status and leaderboard CLI views

### Phase 5: Polish & Launch (Week 5)
- [ ] Security testing
- [ ] Performance optimization
- [ ] Documentation
- [ ] Domain configuration (rank.mo.ai.kr)

---

## 6. Success Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| Active Users | 100+ in first month | Weekly active users |
| Session Submissions | 1000+ sessions/week | Database count |
| API Uptime | 99.9% | Vercel analytics |
| Page Load Time | < 2s | Lighthouse score |

---

## 7. Open Questions

1. **Leaderboard Reset**: Should rankings reset periodically (monthly)?
2. **Badges/Achievements**: Add gamification elements later?
3. **Team Rankings**: Support for organization/team leaderboards?

---

## 8. References

- [claude-code-leaderboard](https://github.com/grp06/claude-code-leaderboard) - Reference implementation
- [Security Review Report](./security-review.md) - 12 findings addressed
- [Design Document](./design.md) - Full architecture details

---

**Status**: Ready for Implementation
**Approved By**: GOOS (2026-01-11)
