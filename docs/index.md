---
layout: home

hero:
  name: "MoAI-ADK"
  text: "Agentic Development Kit for Agents"
  tagline: Claude Code 기반 SPEC-First TDD 범용 개발 툴킷
  actions:
    - theme: brand
      text: 시작하기
      link: /getting-started/installation
    - theme: alt
      text: GitHub 보기
      link: https://github.com/modu-ai/moai-adk
  image:
    light: /moai-tui_screen-light.png
    dark: /moai-tui_screen-dark.png
    alt: MoAI-ADK CLI
---

<div class="stats-section">
  <div class="stats-grid">
    <div class="stat-card">
      <div class="stat-icon">🎩</div>
      <div class="stat-number">1개</div>
      <div class="stat-label">SuperAgent Alfred</div>
      <div class="stat-desc">AI 오케스트레이터</div>
    </div>

    <div class="stat-card">
      <div class="stat-icon">🏷️</div>
      <div class="stat-number">4-Core</div>
      <div class="stat-label">@TAG 시스템</div>
      <div class="stat-desc">CODE-FIRST 추적성</div>
    </div>

    <div class="stat-card">
      <div class="stat-icon">📊</div>
      <div class="stat-number">85%+</div>
      <div class="stat-label">테스트 커버리지</div>
      <div class="stat-desc">품질 보증</div>
    </div>
  </div>
</div>

<style>
.stats-section {
  padding: 4rem 2rem;
  background: hsl(var(--background));
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 2rem;
  max-width: 1200px;
  margin: 0 auto;
}

.stat-card {
  background: hsl(var(--card));
  border: 1px solid hsl(var(--border));
  border-radius: calc(var(--radius) + 4px);
  padding: 2rem;
  text-align: center;
  transition: all 0.3s ease;
}

.stat-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.1);
  border-color: hsl(var(--ring));
}

.stat-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
}

.stat-number {
  font-size: 2.5rem;
  font-weight: 900;
  color: hsl(var(--foreground));
  margin-bottom: 0.5rem;
}

.stat-label {
  font-size: 1.125rem;
  font-weight: 600;
  color: hsl(var(--foreground));
  margin-bottom: 0.25rem;
}

.stat-desc {
  font-size: 0.875rem;
  color: hsl(var(--muted-foreground));
}
</style>

## 왜 MoAI-ADK인가?

<div class="value-props">
  <div class="value-card">
    <div class="value-icon">🏷️</div>
    <h3>CODE-FIRST @TAG 추적성</h3>
    <p>요구사항부터 문서까지 완벽한 추적성</p>
    <ul>
      <li>@SPEC:ID → @TEST:ID → @CODE:ID → @DOC:ID</li>
      <li>코드 직접 스캔 방식 (ripgrep)</li>
      <li>고아 TAG 자동 탐지</li>
    </ul>
  </div>

  <div class="value-card">
    <div class="value-icon">📐</div>
    <h3>SPEC 우선 TDD 워크플로우</h3>
    <p>명세 없이는 코드 없음</p>
    <ul>
      <li>EARS 방법론 기반 요구사항 작성</li>
      <li>Red-Green-Refactor 자동화</li>
      <li>Living Document 동기화</li>
    </ul>
  </div>

  <div class="value-card">
    <div class="value-icon">🎩</div>
    <h3>SuperAgent Alfred 오케스트레이션</h3>
    <p>9개 전문 에이전트 통합 관리</p>
    <ul>
      <li>요청 분석 및 라우팅</li>
      <li>병렬/순차 작업 조율</li>
      <li>TRUST 품질 게이트 검증</li>
    </ul>
  </div>
</div>

<style>
.value-props {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 2rem;
  margin: 2rem 0;
}

.value-card {
  background: hsl(var(--card));
  border: 1px solid hsl(var(--border));
  border-radius: calc(var(--radius) + 4px);
  padding: 2rem;
  transition: all 0.3s ease;
}

.value-card:hover {
  border-color: hsl(var(--ring));
  transform: translateY(-2px);
}

.value-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
}

.value-card h3 {
  font-size: 1.5rem;
  font-weight: 700;
  margin-bottom: 0.5rem;
  color: hsl(var(--foreground));
}

.value-card p {
  color: hsl(var(--muted-foreground));
  margin-bottom: 1rem;
}

.value-card ul {
  list-style: none;
  padding: 0;
}

.value-card li {
  padding: 0.5rem 0;
  color: hsl(var(--foreground));
  border-bottom: 1px solid hsl(var(--border));
}

.value-card li:last-child {
  border-bottom: none;
}

.value-card li::before {
  content: "✓ ";
  color: hsl(var(--primary));
  font-weight: bold;
  margin-right: 0.5rem;
}
</style>

---

## 🏷️ @TAG 4-Core 추적성 시스템

**CODE-FIRST 원칙**: TAG의 진실은 코드 자체에만 존재합니다.

### TAG 체인 흐름

```mermaid
graph LR
  A[@SPEC:ID] -->|TDD RED| B[@TEST:ID]
  B -->|TDD GREEN| C[@CODE:ID]
  C -->|문서화| D[@DOC:ID]

  A -.->|.moai/specs/| SPEC[📋 명세 문서]
  B -.->|tests/| TEST[🧪 테스트 코드]
  C -.->|src/| CODE[💻 구현 코드]
  D -.->|docs/| DOC[📖 Living Doc]

  style A fill:#f9f,stroke:#333,stroke-width:2px
  style B fill:#bbf,stroke:#333,stroke-width:2px
  style C fill:#bfb,stroke:#333,stroke-width:2px
  style D fill:#fbb,stroke:#333,stroke-width:2px
```

### TAG 적용 예시

**1. SPEC 문서** (`.moai/specs/SPEC-AUTH-001.md`)
```markdown
---
id: AUTH-001
version: 2.1.0
status: active
created: 2025-09-15
updated: 2025-10-01
---

# @SPEC:AUTH-001: JWT 인증 시스템

## HISTORY

### v2.1.0 (2025-10-01)
- **CHANGED**: 토큰 만료 시간 15분 → 30분으로 변경
- **ADDED**: 리프레시 토큰 자동 갱신 요구사항 추가
- **AUTHOR**: @goos

### v1.0.0 (2025-09-15)
- **INITIAL**: 기본 JWT 인증 명세 작성
- **AUTHOR**: @goos

## EARS 요구사항

### Ubiquitous Requirements
- 시스템은 JWT 기반 인증 기능을 제공해야 한다

### Event-driven Requirements
- WHEN 유효한 자격증명으로 로그인하면, JWT 토큰을 발급해야 한다
```

**2. 테스트 코드** (`tests/auth/service.test.ts`)
```typescript
// @TEST:AUTH-001 | SPEC: SPEC-AUTH-001.md

describe('AuthService', () => {
  test('should authenticate valid user', () => {
    const service = new AuthService();
    const result = await service.authenticate('user', 'pass');
    expect(result.success).toBe(true);
  });
});
```

**3. 구현 코드** (`src/auth/service.ts`)
```typescript
// @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: tests/auth/service.test.ts

export class AuthService {
  // @CODE:AUTH-001:API - 인증 API 엔드포인트
  async authenticate(username: string, password: string) {
    // @CODE:AUTH-001:DOMAIN - 입력 검증
    this.validateInput(username, password);

    // @CODE:AUTH-001:DATA - 사용자 조회
    const user = await this.userRepo.findByUsername(username);

    return this.verifyCredentials(user, password);
  }
}
```

### @CODE 서브 카테고리

구현 세부사항은 `@CODE:ID` 내부에 주석으로 표기:

- `@CODE:ID:API` - REST API, GraphQL 엔드포인트
- `@CODE:ID:UI` - 컴포넌트, 뷰, 화면
- `@CODE:ID:DATA` - 데이터 모델, 스키마, 타입
- `@CODE:ID:DOMAIN` - 비즈니스 로직, 도메인 규칙
- `@CODE:ID:INFRA` - 인프라, 데이터베이스, 외부 연동

### TAG 검증 및 무결성

```bash
# 전체 TAG 스캔
rg '@(SPEC|TEST|CODE|DOC):' -n

# 고아 TAG 탐지 (SPEC 없는 CODE)
rg '@CODE:AUTH-001' -n src/          # CODE는 있는데
rg '@SPEC:AUTH-001' -n .moai/specs/  # SPEC이 없으면 고아

# TAG 체인 무결성 확인
/moai:3-sync  # 자동 검증
```

---

## 🚀 3단계 개발 워크플로우

### ① /moai:1-spec - SPEC 작성

<div class="workflow-card">
  <div class="workflow-header">
    <span class="workflow-number">1</span>
    <span class="workflow-title">SPEC 작성</span>
  </div>
  <div class="workflow-body">
    <p><strong>담당</strong>: spec-builder 🏗️ (시스템 아키텍트)</p>
    <ul>
      <li>EARS 방법론 기반 요구사항 작성</li>
      <li><code>@SPEC:ID</code> TAG 생성</li>
      <li>브랜치/PR 자동 생성 (Team 모드)</li>
    </ul>
    <div class="workflow-principle">
      <strong>핵심 원칙:</strong> 명세 없이는 코드 없음
    </div>
  </div>
</div>

### ② /moai:2-build - TDD 구현

<div class="workflow-card">
  <div class="workflow-header">
    <span class="workflow-number">2</span>
    <span class="workflow-title">TDD 구현</span>
  </div>
  <div class="workflow-body">
    <p><strong>담당</strong>: code-builder 💎 (수석 개발자)</p>
    <ul>
      <li><strong>RED</strong>: <code>@TEST:ID</code> - 실패하는 테스트 작성</li>
      <li><strong>GREEN</strong>: <code>@CODE:ID</code> - 최소 구현</li>
      <li><strong>REFACTOR</strong>: 코드 품질 개선</li>
    </ul>
    <div class="workflow-principle">
      <strong>핵심 원칙:</strong> 테스트 없이는 구현 없음
    </div>
  </div>
</div>

### ③ /moai:3-sync - 문서 동기화

<div class="workflow-card">
  <div class="workflow-header">
    <span class="workflow-number">3</span>
    <span class="workflow-title">문서 동기화</span>
  </div>
  <div class="workflow-body">
    <p><strong>담당</strong>: doc-syncer 📖 (테크니컬 라이터)</p>
    <ul>
      <li>Living Document 자동 갱신</li>
      <li>TAG 체인 무결성 검증</li>
      <li>PR Draft → Ready 전환</li>
    </ul>
    <div class="workflow-principle">
      <strong>핵심 원칙:</strong> 추적성 없이는 완성 없음
    </div>
  </div>
</div>

<style>
.workflow-card {
  background: hsl(var(--card));
  border: 1px solid hsl(var(--border));
  border-radius: calc(var(--radius) + 4px);
  margin: 1.5rem 0;
  overflow: hidden;
}

.workflow-header {
  background: hsl(var(--muted));
  padding: 1rem 1.5rem;
  display: flex;
  align-items: center;
  gap: 1rem;
}

.workflow-number {
  width: 2.5rem;
  height: 2.5rem;
  background: hsl(var(--primary));
  color: hsl(var(--primary-foreground));
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.25rem;
  font-weight: 900;
}

.workflow-title {
  font-size: 1.25rem;
  font-weight: 700;
  color: hsl(var(--foreground));
}

.workflow-body {
  padding: 1.5rem;
}

.workflow-body ul {
  margin: 1rem 0;
}

.workflow-principle {
  margin-top: 1rem;
  padding: 1rem;
  background: hsl(var(--accent));
  border-left: 4px solid hsl(var(--primary));
  border-radius: var(--radius);
  color: hsl(var(--accent-foreground));
}
</style>

---

## 🎩 SuperAgent Alfred

<div class="alfred-hero">
  <div class="alfred-icon">🎩</div>
  <h3>모두의 AI 집사 - 당신의 개발 오케스트레이터</h3>
  <p>정확하고 예의 바르며, 모든 요청을 체계적으로 처리하는 전문 지휘자</p>
</div>

### 핵심 오케스트레이션 역할

**1. 요청 분석 및 라우팅**
- 사용자 의도 파악
- 적절한 Sub-Agent 식별
- 복합 작업 단계별 분해

**2. Sub-Agent 위임 전략**
- **직접 처리**: 간단한 조회/분석
- **Single Agent**: 단일 완결 작업
- **Sequential**: `/moai:1-spec` → `/moai:2-build` → `/moai:3-sync`
- **Parallel**: 테스트 + 린트 + 빌드 동시 실행

**3. 품질 게이트 검증**
- TRUST 5원칙 준수 확인
- @TAG 체인 무결성 검증
- debug-helper 자동 호출

<style>
.alfred-hero {
  text-align: center;
  padding: 3rem 2rem;
  background: linear-gradient(135deg, hsl(var(--muted)) 0%, hsl(var(--card)) 100%);
  border-radius: calc(var(--radius) + 8px);
  margin: 2rem 0;
}

.alfred-icon {
  font-size: 5rem;
  margin-bottom: 1rem;
}

.alfred-hero h3 {
  font-size: 2rem;
  font-weight: 900;
  color: hsl(var(--foreground));
  margin-bottom: 0.5rem;
}

.alfred-hero p {
  font-size: 1.125rem;
  color: hsl(var(--muted-foreground));
}
</style>

---

## 📋 9개 전문 에이전트 생태계

<div class="agent-grid">
  <div class="agent-card">
    <div class="agent-icon">🏗️</div>
    <h4>spec-builder</h4>
    <p class="agent-role">시스템 아키텍트</p>
    <span class="agent-tag">SPEC 작성</span>
  </div>

  <div class="agent-card">
    <div class="agent-icon">💎</div>
    <h4>code-builder</h4>
    <p class="agent-role">수석 개발자</p>
    <span class="agent-tag">TDD 구현</span>
  </div>

  <div class="agent-card">
    <div class="agent-icon">📖</div>
    <h4>doc-syncer</h4>
    <p class="agent-role">테크니컬 라이터</p>
    <span class="agent-tag">문서 동기화</span>
  </div>

  <div class="agent-card">
    <div class="agent-icon">🏷️</div>
    <h4>tag-agent</h4>
    <p class="agent-role">지식 관리자</p>
    <span class="agent-tag">TAG 관리</span>
  </div>

  <div class="agent-card">
    <div class="agent-icon">🚀</div>
    <h4>git-manager</h4>
    <p class="agent-role">릴리스 엔지니어</p>
    <span class="agent-tag">Git 관리</span>
  </div>

  <div class="agent-card">
    <div class="agent-icon">🔬</div>
    <h4>debug-helper</h4>
    <p class="agent-role">트러블슈터</p>
    <span class="agent-tag">문제 해결</span>
  </div>

  <div class="agent-card">
    <div class="agent-icon">✅</div>
    <h4>trust-checker</h4>
    <p class="agent-role">품질 보증 리드</p>
    <span class="agent-tag">품질 검증</span>
  </div>

  <div class="agent-card">
    <div class="agent-icon">🛠️</div>
    <h4>cc-manager</h4>
    <p class="agent-role">데브옵스 엔지니어</p>
    <span class="agent-tag">환경 설정</span>
  </div>

  <div class="agent-card">
    <div class="agent-icon">📋</div>
    <h4>project-manager</h4>
    <p class="agent-role">프로젝트 매니저</p>
    <span class="agent-tag">프로젝트 초기화</span>
  </div>
</div>

<style>
.agent-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1.5rem;
  margin: 2rem 0;
}

.agent-card {
  background: hsl(var(--card));
  border: 1px solid hsl(var(--border));
  border-radius: calc(var(--radius) + 4px);
  padding: 1.5rem;
  text-align: center;
  transition: all 0.3s ease;
}

.agent-card:hover {
  transform: translateY(-4px);
  border-color: hsl(var(--ring));
  box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.1);
}

.agent-icon {
  font-size: 2.5rem;
  margin-bottom: 0.75rem;
}

.agent-card h4 {
  font-size: 1.125rem;
  font-weight: 700;
  color: hsl(var(--foreground));
  margin-bottom: 0.5rem;
}

.agent-role {
  font-size: 0.875rem;
  color: hsl(var(--muted-foreground));
  margin-bottom: 0.75rem;
}

.agent-tag {
  display: inline-block;
  padding: 0.25rem 0.75rem;
  background: hsl(var(--secondary));
  color: hsl(var(--secondary-foreground));
  border-radius: calc(var(--radius) - 2px);
  font-size: 0.75rem;
  font-weight: 600;
}
</style>

---

## ✅ TRUST 5원칙

<div class="trust-grid">
  <div class="trust-card">
    <div class="trust-letter">T</div>
    <div class="trust-word">est First</div>
    <p>SPEC 기반 TDD - 테스트 없이는 구현 없음</p>
  </div>

  <div class="trust-card">
    <div class="trust-letter">R</div>
    <div class="trust-word">eadable</div>
    <p>요구사항 주도 가독성 - SPEC 정렬 클린 코드</p>
  </div>

  <div class="trust-card">
    <div class="trust-letter">U</div>
    <div class="trust-word">nified</div>
    <p>통합 SPEC 아키텍처 - 언어 간 일관성</p>
  </div>

  <div class="trust-card">
    <div class="trust-letter">S</div>
    <div class="trust-word">ecured</div>
    <p>SPEC 준수 보안 - 설계 단계 보안</p>
  </div>

  <div class="trust-card">
    <div class="trust-letter">T</div>
    <div class="trust-word">rackable</div>
    <p>@TAG 추적성 - CODE-FIRST 방식</p>
  </div>
</div>

<style>
.trust-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1.5rem;
  margin: 2rem 0;
}

.trust-card {
  background: hsl(var(--card));
  border: 1px solid hsl(var(--border));
  border-radius: calc(var(--radius) + 4px);
  padding: 2rem 1.5rem;
  text-align: center;
  transition: all 0.3s ease;
}

.trust-card:hover {
  transform: translateY(-2px);
  border-color: hsl(var(--primary));
}

.trust-letter {
  font-size: 3rem;
  font-weight: 900;
  color: hsl(var(--primary));
  line-height: 1;
  margin-bottom: 0.5rem;
}

.trust-word {
  font-size: 1.25rem;
  font-weight: 700;
  color: hsl(var(--foreground));
  margin-bottom: 0.75rem;
}

.trust-card p {
  font-size: 0.875rem;
  color: hsl(var(--muted-foreground));
  line-height: 1.5;
}
</style>

---

## 지금 바로 MoAI-ADK를 시작하세요

<div class="cta-section">
  <h3>SPEC-First TDD로 완벽한 코드 품질을 경험하세요</h3>
  <p>9개 전문 에이전트와 함께 개발 생산성을 혁신하세요</p>
  <div class="cta-buttons">
    <a href="/getting-started/installation" class="cta-button primary">시작하기 →</a>
    <a href="/guide/workflow" class="cta-button secondary">문서 보기 →</a>
  </div>
</div>

<style>
.cta-section {
  text-align: center;
  padding: 4rem 2rem;
  background: linear-gradient(135deg, hsl(var(--primary)) 0%, hsl(var(--secondary)) 100%);
  border-radius: calc(var(--radius) + 8px);
  margin: 3rem 0;
}

.cta-section h3 {
  font-size: 2rem;
  font-weight: 900;
  color: hsl(var(--primary-foreground));
  margin-bottom: 1rem;
}

.cta-section p {
  font-size: 1.125rem;
  color: hsl(var(--primary-foreground));
  opacity: 0.9;
  margin-bottom: 2rem;
}

.cta-buttons {
  display: flex;
  gap: 1rem;
  justify-content: center;
  flex-wrap: wrap;
}

.cta-button {
  display: inline-flex;
  align-items: center;
  padding: 0.75rem 2rem;
  border-radius: var(--radius);
  font-weight: 600;
  text-decoration: none;
  transition: all 0.3s ease;
}

.cta-button.primary {
  background: hsl(var(--background));
  color: hsl(var(--foreground));
}

.cta-button.primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.3);
}

.cta-button.secondary {
  background: transparent;
  color: hsl(var(--primary-foreground));
  border: 2px solid hsl(var(--primary-foreground));
}

.cta-button.secondary:hover {
  background: hsl(var(--primary-foreground));
  color: hsl(var(--primary));
}
</style>

---

## 더 알아보기

### 핵심 가이드
- [3단계 워크플로우](/guide/workflow) - `/moai:1-spec` → `/moai:2-build` → `/moai:3-sync`
- [SPEC-First TDD](/guide/spec-first-tdd) - EARS 방식 명세 작성법
- [TAG 시스템](/guide/tag-system) - CODE-FIRST 추적성 관리
- [TRUST 5원칙](/concepts/trust-principles) - 코드 품질 보증 원칙

### Claude Code 에이전트
- [에이전트 개요](/claude/agents) - 9개 전문 에이전트 소개
- [SuperAgent Alfred](/claude/agents/alfred) - AI 오케스트레이터
- [워크플로우 명령어](/claude/commands) - `/moai:` 명령어 상세
- [이벤트 훅](/claude/hooks) - 자동화 시스템

### 시작하기
- [설치 가이드](/getting-started/installation) - 시스템 요구사항 및 설치
- [빠른 시작](/getting-started/quick-start) - 5분 안에 첫 프로젝트
- [프로젝트 설정](/getting-started/project-setup) - 언어별 설정 가이드

### CLI 명령어
- `moai init` - 프로젝트 초기화
- `moai doctor` - 지능형 시스템 진단
- `moai status` - 프로젝트 상태 확인
- `moai update` - 업데이트 관리
- `moai restore` - 백업 복원
