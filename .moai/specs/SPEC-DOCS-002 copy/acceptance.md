# @SPEC:DOCS-002 인수 테스트 시나리오

## 개요

본 문서는 SPEC-DOCS-002 (VitePress Phase 2 - 핵심 개념 페이지 3개 추가) 완료 기준을 정의합니다.

---

## 인수 기준 (Given-When-Then)

### AC-001: Alfred 10개 AI 에이전트 팀 페이지 표시

**Given**: VitePress 개발 서버가 실행 중이고
**When**: 사용자가 `http://localhost:5173/concepts/alfred-agents` 페이지를 방문하면
**Then**:
- [ ] 페이지 타이틀 "Alfred 10개 AI 에이전트 팀"이 표시됨
- [ ] Alfred SuperAgent 소개 섹션이 표시됨 (~50 LOC)
- [ ] 9개 전문 에이전트 표가 표시됨 (페르소나, 전문 영역, 커맨드, 위임 시점 포함)
- [ ] Mermaid 오케스트레이션 다이어그램이 렌더링됨
- [ ] 에이전트 협업 원칙 섹션이 표시됨
- [ ] 페이지 내 모든 내부 링크가 유효함
- [ ] 페이지 로딩 시간 < 2초

**검증 방법**:
```bash
# 페이지 존재 확인
test -f docs/concepts/alfred-agents.md

# 빌드 성공 확인
bun run docs:build

# 페이지 크기 확인 (300 LOC 목표)
wc -l docs/concepts/alfred-agents.md | awk '{print $1}'
```

---

### AC-002: @TAG 추적성 시스템 페이지 표시

**Given**: VitePress 개발 서버가 실행 중이고
**When**: 사용자가 `http://localhost:5173/concepts/tag-system` 페이지를 방문하면
**Then**:
- [ ] 페이지 타이틀 "@TAG 추적성 시스템"이 표시됨
- [ ] CODE-FIRST 원칙 섹션이 표시됨 (~50 LOC)
- [ ] TAG 체인 다이어그램이 렌더링됨 (@SPEC → @TEST → @CODE → @DOC)
- [ ] TAG ID 규칙 설명이 표시됨 (`<도메인>-<3자리>`)
- [ ] 언어별 TAG 사용 예시 (TypeScript, Python, Dart) 표시됨
- [ ] TAG 검증 방법 (rg 명령어) 표시됨
- [ ] 페이지 내 모든 코드 블록이 올바르게 렌더링됨

**검증 방법**:
```bash
# 페이지 존재 확인
test -f docs/concepts/tag-system.md

# TAG 예시 코드 블록 개수 확인 (3개 이상)
grep -c '```' docs/concepts/tag-system.md

# 페이지 크기 확인 (250 LOC 목표)
wc -l docs/concepts/tag-system.md | awk '{print $1}'
```

---

### AC-003: TRUST 5원칙 페이지 표시

**Given**: VitePress 개발 서버가 실행 중이고
**When**: 사용자가 `http://localhost:5173/concepts/trust-principles` 페이지를 방문하면
**Then**:
- [ ] 페이지 타이틀 "TRUST 5원칙"이 표시됨
- [ ] 5개 원칙 섹션 (T-R-U-S-T)이 각각 표시됨
- [ ] Test First: 테스트 커버리지 ≥85%, TDD Red-Green-Refactor 설명
- [ ] Readable: 파일 ≤300 LOC, 함수 ≤50 LOC, 복잡도 ≤10 기준
- [ ] Unified: 아키텍처 통합성, 언어별 표준 도구 설명
- [ ] Secured: SQL Injection, XSS, CSRF 방어 전략
- [ ] Trackable: @TAG 체인 무결성, 고아 TAG 탐지
- [ ] 언어별 도구 비교 표 (Jest/pytest/go test/cargo test 등)

**검증 방법**:
```bash
# 페이지 존재 확인
test -f docs/concepts/trust-principles.md

# TRUST 각 원칙 섹션 존재 확인
grep -E '## (T|R|U|S|T) -' docs/concepts/trust-principles.md

# 페이지 크기 확인 (200 LOC 목표)
wc -l docs/concepts/trust-principles.md | awk '{print $1}'
```

---

### AC-004: VitePress 빌드 성공

**Given**: 3개 페이지 파일이 생성되고
**When**: `bun run docs:build` 명령어를 실행하면
**Then**:
- [ ] 빌드가 에러 없이 완료됨
- [ ] `docs/.vitepress/dist/` 디렉토리가 생성됨
- [ ] 3개 페이지 HTML 파일이 생성됨 (alfred-agents.html, tag-system.html, trust-principles.html)
- [ ] 빌드 시간 < 10초
- [ ] 빌드 로그에 에러/경고 없음

**검증 방법**:
```bash
# 빌드 실행
bun run docs:build

# HTML 파일 생성 확인
test -f docs/.vitepress/dist/concepts/alfred-agents.html
test -f docs/.vitepress/dist/concepts/tag-system.html
test -f docs/.vitepress/dist/concepts/trust-principles.html
```

---

### AC-005: Sidebar 네비게이션 확장

**Given**: VitePress 개발 서버가 실행 중이고
**When**: 사용자가 Sidebar의 "핵심 개념" 섹션을 클릭하면
**Then**:
- [ ] Sidebar에 "핵심 개념" 섹션이 표시됨
- [ ] 4개 항목이 표시됨:
  - SPEC-First TDD 개발 방법 (Phase 1 기존)
  - Alfred 10개 AI 에이전트 팀 (Phase 2 신규)
  - @TAG 추적성 시스템 (Phase 2 신규)
  - TRUST 5원칙 (Phase 2 신규)
- [ ] 현재 페이지가 하이라이트됨 (파란색)
- [ ] 각 링크 클릭 시 해당 페이지로 이동함
- [ ] SPA 방식 전환 (새로고침 없음)

**검증 방법**:
```bash
# config.mts에서 Sidebar 설정 확인
grep -A 20 "'핵심 개념'" docs/.vitepress/config.mts

# 4개 항목 존재 확인
grep -c "text: '" docs/.vitepress/config.mts | awk '{if($1>=4) print "PASS"}'
```

---

### AC-006: 내부 링크 유효성 검증

**Given**: 3개 페이지에 내부 링크가 포함되어 있고
**When**: VitePress 빌드 또는 개발 서버 실행 시
**Then**:
- [ ] 모든 내부 링크가 유효함 (깨진 링크 0개)
- [ ] 상대 경로 링크 (`./`, `../`)가 올바르게 해석됨
- [ ] 절대 경로 링크 (`/guide/`, `/concepts/`)가 올바르게 해석됨
- [ ] 앵커 링크 (`#section`)가 올바르게 작동함

**검증 방법**:
```bash
# 링크 패턴 추출
rg '\[.*\]\(.*\)' docs/concepts/alfred-agents.md docs/concepts/tag-system.md docs/concepts/trust-principles.md

# VitePress 빌드 시 링크 체크 (에러 없음 확인)
bun run docs:build 2>&1 | grep -i "broken link" || echo "No broken links"
```

---

### AC-007: 테스트 통과 확인

**Given**: Phase 2 테스트가 작성되어 있고
**When**: `bun test` 명령어를 실행하면
**Then**:
- [ ] 모든 Phase 1 테스트 통과 (9개)
- [ ] 모든 Phase 2 테스트 통과 (예상 3개 추가, 총 12개)
- [ ] 테스트 커버리지 ≥85%
- [ ] 테스트 실행 시간 < 30초

**검증 방법**:
```bash
# 테스트 실행
bun test moai-adk-ts/__tests__/docs/

# 테스트 개수 확인
bun test moai-adk-ts/__tests__/docs/ --reporter=verbose | grep -c "✓"

# 커버리지 확인
bun test moai-adk-ts/__tests__/docs/ --coverage
```

---

### AC-008: Mermaid 다이어그램 렌더링

**Given**: alfred-agents.md에 Mermaid 다이어그램이 포함되어 있고
**When**: 사용자가 페이지를 방문하면
**Then**:
- [ ] Mermaid 다이어그램이 SVG로 렌더링됨
- [ ] 다이어그램에 노드 및 엣지가 표시됨
- [ ] 다이어그램이 읽기 쉬운 크기로 표시됨
- [ ] Dark Mode에서도 올바르게 렌더링됨

**검증 방법**:
```bash
# Mermaid 코드 블록 존재 확인
grep -A 10 '```mermaid' docs/concepts/alfred-agents.md

# VitePress 개발 서버 실행 후 수동 확인
bun run docs:dev
# 브라우저에서 http://localhost:5173/concepts/alfred-agents 접속
```

---

## 콘텐츠 소스 비율 검증

**목표**: README 60% + dev-guide 30% + 신규 10%

| 페이지 | README 소스 | dev-guide 소스 | 신규 내용 | 합계 |
|--------|------------|---------------|-----------|------|
| alfred-agents.md | 270 LOC (90%) | 0 LOC | 30 LOC (10%) | 300 LOC |
| tag-system.md | 180 LOC (72%) | 20 LOC (8%) | 50 LOC (20%) | 250 LOC |
| trust-principles.md | 50 LOC (25%) | 120 LOC (60%) | 30 LOC (15%) | 200 LOC |
| **합계** | 500 LOC (67%) | 140 LOC (19%) | 110 LOC (14%) | 750 LOC |

**검증 방법**:
- README.md 608~932줄 (Alfred 섹션)을 alfred-agents.md에 반영 확인
- README.md 813~931줄 (TAG 섹션)을 tag-system.md에 반영 확인
- development-guide.md TRUST 원칙 섹션을 trust-principles.md에 반영 확인

---

## Phase 2 완료 정의 (Definition of Done)

- [x] SPEC-DOCS-002 문서 작성 완료 (spec.md, plan.md, acceptance.md)
- [ ] 3개 페이지 파일 생성 (alfred-agents.md, tag-system.md, trust-principles.md)
- [ ] VitePress Sidebar 설정 업데이트 (config.mts)
- [ ] VitePress 빌드 성공 (에러 0개)
- [ ] 모든 테스트 통과 (12개 이상)
- [ ] 내부 링크 유효성 100%
- [ ] Mermaid 다이어그램 렌더링 확인
- [ ] 콘텐츠 소스 비율 검증 (README 67%)
- [ ] CHANGELOG 업데이트
- [ ] Git 커밋 (TDD 사이클: RED → GREEN → REFACTOR)
- [ ] 문서 동기화 (`/alfred:3-sync`)

---

## 비기능 요구사항 검증

### 성능
- [ ] 페이지 로딩 시간 < 2초 (개별 페이지)
- [ ] 빌드 시간 < 10초 (전체 사이트)
- [ ] 핫 리로드 시간 < 3초 (개발 모드)

### 접근성
- [ ] 의미 있는 HTML 구조 (h1 → h2 → h3)
- [ ] Alt 텍스트 (이미지 포함 시)
- [ ] 키보드 네비게이션 가능

### 유지보수성
- [ ] 파일 크기 적절 (300 LOC 이하 권장)
- [ ] Markdown 문법 오류 없음
- [ ] 코드 블록 언어 명시 (```typescript, ```python 등)

---

## 리스크 및 대응

### 기술적 리스크

1. **Mermaid 렌더링 실패**
   - **확률**: 낮음 (VitePress 기본 지원)
   - **영향**: 중간 (다이어그램 표시 불가)
   - **대응**: 다이어그램 문법 사전 검증, 대체 이미지 준비

2. **README 콘텐츠 중복**
   - **확률**: 중간 (Phase 1과 중복 가능)
   - **영향**: 낮음 (중복 정리 필요)
   - **대응**: Phase 1 페이지 검토 후 차별화

3. **빌드 시간 증가**
   - **확률**: 낮음 (페이지 3개 추가)
   - **영향**: 낮음 (개발 경험 약간 저하)
   - **대응**: 빌드 시간 모니터링 (< 10초 목표)

### 콘텐츠 리스크

1. **깊이 부족**
   - **확률**: 낮음 (README 상세 내용 충분)
   - **영향**: 중간 (사용자 이해도 저하)
   - **대응**: dev-guide.md 심화 내용 추가, 예시 코드 확충

2. **일관성 부족**
   - **확률**: 중간 (3개의 에이전트가 각각 작성)
   - **영향**: 중간 (사용자 혼란)
   - **대응**: trust-checker로 문서 스타일 일관성 검증

---

## 다음 단계 (Phase 3-4)

**Phase 3** (Priority 2):
- Output Styles (4가지 스타일: professional, pair-collab, beginner-learning, study-deep)
- CLI Reference (10개 페이지)
- 문제 해결 (4개 페이지)

**Phase 4** (Priority 3):
- 언어별 가이드 (TypeScript, Python, Flutter, Go, Rust - 10개 페이지)
- 고급 주제 (4개 페이지)
- 기여 가이드 (5개 페이지)

---

**작성자**: spec-builder 에이전트
**검토자**: trust-checker 에이전트
**승인**: @Goos
