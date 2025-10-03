---
layout: home

hero:
  name: "MoAI-ADK"
  text: "MoAI's Agentic Development Kit"
  tagline: 명세 없으면 코드 없다. 테스트 없으면 구현 없다.
  actions:
    - theme: brand
      text: 시작하기
      link: /getting-started/installation
    - theme: alt
      text: GitHub
      link: https://github.com/modu-ai/moai-adk
  image:
    light: /moai-tui_screen-light.png
    dark: /moai-tui_screen-dark.png
    alt: MoAI-ADK

features:
  - icon:
      src: /icons/sparkles.svg
    title: SuperAgent Alfred
    details: 9개 전문 에이전트 통합 오케스트레이션
  - icon:
      src: /icons/workflow.svg
    title: 3단계 워크플로우
    details: SPEC → BUILD → SYNC
  - icon:
      src: /icons/tag.svg
    title: CODE-FIRST @TAG
    details: 코드 직접 스캔 방식 추적성
  - icon:
      src: /icons/check-circle.svg
    title: TRUST 5원칙
    details: Test, Readable, Unified, Secured, Trackable
  - icon:
      src: /icons/spec.svg
    title: EARS 방법론
    details: 체계적 요구사항 작성
  - icon:
      src: /icons/language.svg
    title: 멀티 언어
    details: Python, TypeScript, Java, Go, Rust

---

## SPEC-First TDD

**"명세 없으면 코드 없다. 테스트 없으면 구현 없다."**

MoAI-ADK는 Claude Code 기반 개발 프레임워크로, `@TAG` 추적성 시스템을 통해 **요구사항부터 문서까지 완벽한 추적성**을 보장합니다.

### 3단계 워크플로우

```bash
# 1. SPEC 작성
/alfred:1-spec "사용자 인증 시스템"

# 2. TDD 구현
/alfred:2-build SPEC-001

# 3. 문서 동기화
/alfred:3-sync
```

### `@TAG` 추적성

`@SPEC:ID` → `@TEST:ID` → `@CODE:ID` → `@DOC:ID`

코드 직접 스캔 방식으로 고아 TAG 자동 탐지, 체인 무결성 검증.

---

## 시작하기

<div class="getting-started-grid">
  <a href="/getting-started/installation" class="card-link">
    <div class="start-card">
      <div class="card-icon-wrapper">
        <IconPackage :size="24" :stroke-width="1.5" class="card-icon" />
      </div>
      <h3 class="card-title">설치</h3>
      <p class="card-desc">npm/bun으로 빠른 설치</p>
    </div>
  </a>
  <a href="/guide/workflow" class="card-link">
    <div class="start-card">
      <div class="card-icon-wrapper">
        <IconWorkflow :size="24" :stroke-width="1.5" class="card-icon" />
      </div>
      <h3 class="card-title">워크플로우</h3>
      <p class="card-desc">3단계 개발 사이클</p>
    </div>
  </a>
  <a href="/claude/agents" class="card-link">
    <div class="start-card">
      <div class="card-icon-wrapper">
        <IconUsers :size="24" :stroke-width="1.5" class="card-icon" />
      </div>
      <h3 class="card-title">에이전트</h3>
      <p class="card-desc">9개 전문 에이전트</p>
    </div>
  </a>
  <a href="/guide/tag-system" class="card-link">
    <div class="start-card">
      <div class="card-icon-wrapper">
        <IconTag :size="24" :stroke-width="1.5" class="card-icon" />
      </div>
      <h3 class="card-title">TAG 시스템</h3>
      <p class="card-desc">CODE-FIRST 추적성</p>
    </div>
  </a>
</div>

---

<div class="cta-section">
  <h2 class="cta-title">지금 시작하세요</h2>
  <p class="cta-subtitle">TypeScript 기반 범용 언어 지원 SPEC-First TDD 프레임워크</p>
  <div class="cta-buttons">
    <a href="/getting-started/installation" class="cta-primary">시작하기 →</a>
    <a href="https://github.com/modu-ai/moai-adk" class="cta-secondary">GitHub →</a>
  </div>
</div>

<style scoped>
/* 시작하기 그리드 */
.getting-started-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 32px;
  margin-top: 48px;
}

.card-link {
  text-decoration: none;
  color: inherit;
}

.start-card {
  padding: 32px;
  border: 1px solid var(--vp-c-border);
  border-radius: 8px;
  background: var(--vp-c-bg);
  transition: all 0.3s ease;
}

.start-card:hover {
  transform: translateY(-4px);
  border-color: var(--vp-c-brand-1);
  box-shadow: 0 8px 24px rgba(24, 24, 27, 0.1);
}

:root.dark .start-card:hover {
  box-shadow: 0 8px 24px rgba(250, 250, 250, 0.05);
}

.card-icon-wrapper {
  width: 48px;
  height: 48px;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  background: var(--vp-c-brand-1);
}

.card-icon {
  color: var(--vp-c-bg);
}

:root.dark .card-icon {
  color: #18181b;
}

.card-title {
  margin-top: 0;
  font-weight: 700;
  font-size: 1.25rem;
  color: var(--vp-c-text-1);
}

.card-desc {
  font-weight: 400;
  color: var(--vp-c-text-3);
}

/* CTA 섹션 */
.cta-section {
  text-align: center;
  padding: 64px 0;
  background: linear-gradient(180deg, var(--vp-c-bg-soft) 0%, var(--vp-c-bg-mute) 100%);
}

.cta-title {
  font-weight: 900;
  font-size: 2.5rem;
  letter-spacing: -0.05em;
  color: var(--vp-c-text-1);
  margin-bottom: 16px;
}

.cta-subtitle {
  font-weight: 300;
  font-size: 1.25rem;
  letter-spacing: -0.03em;
  color: var(--vp-c-text-3);
  margin-bottom: 32px;
}

.cta-buttons {
  display: flex;
  gap: 16px;
  justify-content: center;
  flex-wrap: wrap;
}

.cta-primary {
  padding: 16px 32px;
  background: var(--vp-c-brand-1) !important;
  color: var(--vp-c-bg) !important;
  border-radius: 8px;
  text-decoration: none;
  font-weight: 700;
  letter-spacing: -0.03em;
  transition: all 0.3s ease;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.cta-primary:hover {
  background: var(--vp-c-brand-2) !important;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.2);
}

:root.dark .cta-primary {
  background: #fafafa !important;
  color: #18181b !important;
  box-shadow: 0 2px 8px rgba(250, 250, 250, 0.15) !important;
}

:root.dark .cta-primary:hover {
  background: #e4e4e7 !important;
  box-shadow: 0 4px 16px rgba(250, 250, 250, 0.25) !important;
}

.cta-secondary {
  padding: 16px 32px;
  border: 1px solid var(--vp-c-brand-1);
  color: var(--vp-c-brand-1);
  background: transparent;
  border-radius: 8px;
  text-decoration: none;
  font-weight: 700;
  letter-spacing: -0.03em;
  transition: all 0.3s ease;
}

.cta-secondary:hover {
  background: var(--vp-c-brand-1);
  color: var(--vp-c-bg);
}
</style>
