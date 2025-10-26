# `.moai/memory/` 파일 사용처 분석

**작성일**: 2025-10-27
**상태**: 완료
**버전**: 1.0.0

---

## 📋 개요

`.moai/memory/` 디렉토리는 **Alfred와 Sub-agents가 공유하는 지식 저장소**입니다. 이 문서는 각 메모리 파일이 누가(Who), 언제(When), 어디서(Where), 왜(Why) 사용되는지 분석합니다.

---

## 📁 메모리 파일 목록

### 핵심 가이드 문서 (3개)

| 파일명 | 크기 | 용도 | 카테고리 |
|--------|------|------|----------|
| `CLAUDE-AGENTS-GUIDE.md` | 15KB | Sub-agent 선택 및 협업 가이드 | Agent Orchestration |
| `CLAUDE-PRACTICES.md` | 12KB | 실전 워크플로우 예제 및 패턴 | Best Practices |
| `CLAUDE-RULES.md` | 19KB | Skill/TAG/Git 규칙 및 원칙 | Rules & Standards |

### 표준 정의 문서 (4개)

| 파일명 | 크기 | 용도 | 카테고리 |
|--------|------|------|----------|
| `CONFIG-SCHEMA.md` | 12KB | `.moai/config.json` 스키마 정의 | Configuration |
| `DEVELOPMENT-GUIDE.md` | 14KB | SPEC-First TDD 워크플로우 가이드 | Development |
| `GITFLOW-PROTECTION-POLICY.md` | 6KB | Git 브랜치 보호 정책 | Git Workflow |
| `SPEC-METADATA.md` | 9KB | SPEC YAML frontmatter 표준 | Specification |

### 구현 분석 문서 (7개)

| 파일명 | 크기 | 용도 | 카테고리 |
|--------|------|------|----------|
| `IMPLEMENTATION-GUIDE-PHASE3-4.md` | 8KB | Skills Progressive Disclosure 구현 | Implementation |
| `PHASE4-VALIDATION.md` | 4KB | Phase 4 검증 체크리스트 | Validation |
| `SKILLS-DESCRIPTION-POLICY.md` | 7KB | Skills Description 작성 표준 | Skills Management |
| `TEAM_MODE_GITHUB_INTEGRATION_ANALYSIS.md` | 25KB | Team 모드 GitHub 통합 분석 | Team Collaboration |
| `WORKFLOW-IMPLEMENTATION-REPORT.md` | 10KB | Commands→Agents→Skills 통합 보고서 | Implementation Report |

---

## 🎯 파일별 사용처 분석

### 1. `CLAUDE-AGENTS-GUIDE.md`

**누가 사용하나?**
- 🎩 **Alfred (Main)**: Sub-agent 선택 시 결정 트리 참조
- 👥 **개발자**: Agent 아키텍처 이해

**언제 사용하나?**
- `/alfred:1-plan` 시작 시 → spec-builder 선택
- `/alfred:2-run` 시작 시 → implementation-planner/tdd-implementer 선택
- `/alfred:3-sync` 시작 시 → doc-syncer/tag-agent 선택
- Sub-agent 간 핸드오프 결정 시

**어디서 참조하나?**
- `CLAUDE.md` Line 31: "See [CLAUDE-AGENTS-GUIDE.md]"
- `CLAUDE.md` 문서 라우팅 맵: "Sub-agent 선택 방법"

**왜 필요한가?**
- 19개 팀원(Alfred + 10 core agents + 6 specialists + 2 built-in) 역할 명확화
- Model 선택 기준 (Sonnet vs Haiku)
- Agent 협업 원칙 정의

---

### 2. `CLAUDE-PRACTICES.md`

**누가 사용하나?**
- 🎩 **Alfred**: 실전 워크플로우 실행 시 패턴 참조
- 🤖 **Sub-agents**: 구체적 작업 예제 학습
- 👥 **개발자**: Best practices 학습

**언제 사용하나?**
- 복잡한 워크플로우 실행 시
- Context Engineering 전략 필요 시
- 실제 작업 예제 참조 시

**어디서 참조하나?**
- `CLAUDE.md` 문서 라우팅 맵: "실전 워크플로우 예제", "Context Engineering Strategy"

**왜 필요한가?**
- 이론(CLAUDE-RULES.md)과 실전(CLAUDE-PRACTICES.md) 분리
- 구체적인 실행 패턴 제공
- Context 효율성 최적화 전략

---

### 3. `CLAUDE-RULES.md`

**누가 사용하나?**
- 🎩 **Alfred**: 모든 의사결정 시 규칙 검증
- 🤖 **All Sub-agents**: Skill 호출, TAG 생성, Git 커밋 시
- 👥 **개발자**: 표준 준수 검증

**언제 사용하나?**
- Skill 호출 전 (invocation rules 확인)
- @TAG 생성/검증 시 (TAG lifecycle 확인)
- Git 커밋 메시지 작성 시 (commit standard 확인)
- TRUST 5 원칙 검증 시
- AskUserQuestion 호출 기준 판단 시

**어디서 참조하나?**
- `CLAUDE.md` 문서 라우팅 맵: 5개 항목
  - "Skill 호출 규칙"
  - "AskUserQuestion 기준"
  - "Git 커밋 메시지 형식"
  - "@TAG 규칙과 검증"
  - "TRUST 5 원칙"

**왜 필요한가?**
- MoAI-ADK의 핵심 규칙 단일 저장소 (SSOT)
- 일관성 있는 의사결정 보장
- 품질 기준 명확화

---

### 4. `CONFIG-SCHEMA.md`

**누가 사용하나?**
- 🎩 **Alfred**: `/alfred:0-project` 실행 시
- 🤖 **project-manager**: 프로젝트 초기화 시
- 👥 **개발자**: `.moai/config.json` 수정 시

**언제 사용하나?**
- 새 프로젝트 생성 시
- Config 스키마 검증 시
- 다국어 설정 변경 시

**어디서 참조하나?**
- `CLAUDE.md` Line 7: "Config: `.moai/config.json`"
- `.claude/agents/alfred/project-manager.md`: Config validation

**왜 필요한가?**
- `.moai/config.json` 스키마 명세서 (SSOT)
- 다국어 지원 설정 표준화
- Project Owner, Mode(personal/team) 정의

---

### 5. `DEVELOPMENT-GUIDE.md`

**누가 사용하나?**
- 🎩 **Alfred**: 전체 워크플로우 오케스트레이션
- 🤖 **All Sub-agents**: SPEC-First TDD 실행
- 👥 **개발자**: MoAI-ADK 학습 및 실행

**언제 사용하나?**
- `/alfred:1-plan` → EARS 요구사항 작성 시
- `/alfred:2-run` → TDD 루프 실행 시
- `/alfred:3-sync` → @TAG 체인 검증 시
- Context Engineering 전략 참조 시

**어디서 참조하나?**
- `WORKFLOW-IMPLEMENTATION-REPORT.md` Line 250
- `/alfred:1-plan.md` Line 558: "Context Engineering"
- `/alfred:3-sync.md` Line 592: "Context Engineering"

**왜 필요한가?**
- SPEC-First TDD의 완전한 가이드
- @TAG 시스템 사용법
- TRUST 5 원칙 실전 적용
- Context Engineering 전략

---

### 6. `GITFLOW-PROTECTION-POLICY.md`

**누가 사용하나?**
- 🎩 **Alfred**: Git 작업 오케스트레이션
- 🤖 **git-manager**: 브랜치/PR 관리
- 👥 **개발자**: Git 정책 이해

**언제 사용하나?**
- Feature 브랜치 생성 시
- PR 생성/머지 시
- Main/Develop 브랜치 보호 정책 적용 시

**어디서 참조하나?**
- `.claude/agents/alfred/git-manager.md`: Git workflow

**왜 필요한가?**
- Git 브랜치 전략 표준화
- 실수 방지 (force push to main 등)
- PR 정책 명확화

---

### 7. `SPEC-METADATA.md`

**누가 사용하나?**
- 🎩 **Alfred**: SPEC 검증 시
- 🤖 **spec-builder**: SPEC 작성 시
- 🤖 **doc-syncer**: SPEC 메타데이터 동기화 시
- 👥 **개발자**: SPEC 작성 시

**언제 사용하나?**
- `/alfred:1-plan` → SPEC 작성 시
- SPEC YAML frontmatter 검증 시
- HISTORY 섹션 작성 시
- 버전 관리 정책 참조 시

**어디서 참조하나?**
- `DEVELOPMENT-GUIDE.md` Line 189: "SPEC Metadata Standard (SSOT)"
- `DEVELOPMENT-GUIDE.md` Line 196: "validation commands in SPEC-METADATA.md"
- `DEVELOPMENT-GUIDE.md` Line 247: "versioning"
- `/alfred:1-plan.md` Line 333: "SPEC Metadata Standard"
- `/alfred:1-plan.md` Line 360: "Full field description"
- `/alfred:1-plan.md` Line 394: "version-system"

**왜 필요한가?**
- SPEC 메타데이터의 단일 진실 공급원 (SSOT)
- Required Fields 7개 정의
- HISTORY 작성 표준
- Semantic Versioning 규칙

---

### 8. `SKILLS-DESCRIPTION-POLICY.md`

**누가 사용하나?**
- 🎩 **Alfred**: Skills description 검증 시
- 🤖 **skill-factory**: 새 Skill 생성 시
- 👥 **개발자**: Skills 유지보수 시

**언제 사용하나?**
- 새 Skill 작성 시
- 기존 Skill description 개선 시
- Skills 품질 검증 시

**어디서 참조하나?**
- `WORKFLOW-IMPLEMENTATION-REPORT.md` Line 45: "60개+ Skills 개선 가이드"
- `.claude/skills/moai-skill-factory/`: Skill 생성 템플릿

**왜 필요한가?**
- 55개 Skills description 일관성 유지
- "Use when" 패턴 표준화
- Progressive Disclosure 구조 가이드

---

### 9. 기타 구현 문서

#### `IMPLEMENTATION-GUIDE-PHASE3-4.md`
- **사용자**: Alfred, skill-factory
- **시점**: Progressive Disclosure 구현 시
- **목적**: 대형 Skills 분리 전략

#### `PHASE4-VALIDATION.md`
- **사용자**: quality-gate, trust-checker
- **시점**: Phase 4 검증 시
- **목적**: 구현 완료 검증 체크리스트

#### `TEAM_MODE_GITHUB_INTEGRATION_ANALYSIS.md`
- **사용자**: Alfred, project-manager
- **시점**: Team 모드 전환 시
- **목적**: GitHub 통합 전략 분석

#### `WORKFLOW-IMPLEMENTATION-REPORT.md`
- **사용자**: 개발자, Alfred
- **시점**: 시스템 아키텍처 이해 시
- **목적**: Commands→Agents→Skills 통합 상태 보고

---

## 🔄 파일 간 의존성 그래프

```
CLAUDE.md (Root)
├─ CLAUDE-AGENTS-GUIDE.md (Agent 선택)
│  └─ CLAUDE-PRACTICES.md (실전 패턴)
├─ CLAUDE-RULES.md (규칙)
│  ├─ SPEC-METADATA.md (SPEC 표준)
│  ├─ GITFLOW-PROTECTION-POLICY.md (Git 정책)
│  └─ DEVELOPMENT-GUIDE.md (워크플로우)
└─ CONFIG-SCHEMA.md (설정)

Commands
├─ /alfred:1-plan
│  ├─ SPEC-METADATA.md ← YAML frontmatter
│  └─ DEVELOPMENT-GUIDE.md ← Context Engineering
├─ /alfred:2-run
│  └─ DEVELOPMENT-GUIDE.md ← TDD workflow
└─ /alfred:3-sync
   └─ DEVELOPMENT-GUIDE.md ← TAG validation

Sub-agents
├─ spec-builder → SPEC-METADATA.md, CLAUDE-RULES.md
├─ tdd-implementer → DEVELOPMENT-GUIDE.md, CLAUDE-RULES.md
├─ doc-syncer → DEVELOPMENT-GUIDE.md, SPEC-METADATA.md
├─ git-manager → GITFLOW-PROTECTION-POLICY.md, CLAUDE-RULES.md
└─ skill-factory → SKILLS-DESCRIPTION-POLICY.md
```

---

## ⏱️ 로딩 타이밍 전략

### 세션 시작 시 (SessionStart Hook)
- ✅ `CLAUDE.md` (항상)
- ✅ `CLAUDE-AGENTS-GUIDE.md` (Alfred 초기화)
- ✅ `CLAUDE-RULES.md` (규칙 로드)

### Just-In-Time 로딩 (Command 실행 시)
- 📄 `/alfred:1-plan` → `SPEC-METADATA.md`, `DEVELOPMENT-GUIDE.md`
- 📄 `/alfred:2-run` → `DEVELOPMENT-GUIDE.md`
- 📄 `/alfred:3-sync` → `DEVELOPMENT-GUIDE.md`

### Sub-agent 핸드오프 시
- 🤖 spec-builder → `SPEC-METADATA.md`, `CLAUDE-RULES.md`
- 🤖 git-manager → `GITFLOW-PROTECTION-POLICY.md`
- 🤖 skill-factory → `SKILLS-DESCRIPTION-POLICY.md`

### 조건부 로딩
- 🔧 Config 수정 시 → `CONFIG-SCHEMA.md`
- 📊 Skills 생성 시 → `SKILLS-DESCRIPTION-POLICY.md`
- 🔍 구현 분석 필요 시 → `IMPLEMENTATION-GUIDE-*.md`

---

## 📊 사용 빈도 순위

| 순위 | 파일명 | 사용 빈도 | 주 사용자 |
|------|--------|----------|----------|
| 1 | `CLAUDE-RULES.md` | 매우 높음 | Alfred, All Sub-agents |
| 2 | `DEVELOPMENT-GUIDE.md` | 높음 | Alfred, All Sub-agents |
| 3 | `SPEC-METADATA.md` | 높음 | spec-builder, doc-syncer |
| 4 | `CLAUDE-AGENTS-GUIDE.md` | 중간 | Alfred |
| 5 | `CLAUDE-PRACTICES.md` | 중간 | Alfred, Developers |
| 6 | `CONFIG-SCHEMA.md` | 낮음 | project-manager |
| 7 | `GITFLOW-PROTECTION-POLICY.md` | 낮음 | git-manager |
| 8 | `SKILLS-DESCRIPTION-POLICY.md` | 낮음 | skill-factory |
| 9 | 구현 분석 문서 (5개) | 매우 낮음 | Developers |

---

## 🎯 최적화 권장사항

### 1. Core Files (항상 메모리에 유지)
```
CLAUDE.md
CLAUDE-AGENTS-GUIDE.md
CLAUDE-RULES.md
```

### 2. JIT Load Files (명령어 시작 시 로드)
```
DEVELOPMENT-GUIDE.md (3개 명령어 공통)
SPEC-METADATA.md (/alfred:1-plan, /alfred:3-sync)
```

### 3. Conditional Load Files (필요 시에만)
```
CONFIG-SCHEMA.md (프로젝트 초기화)
GITFLOW-PROTECTION-POLICY.md (Git 작업)
SKILLS-DESCRIPTION-POLICY.md (Skill 생성)
```

### 4. Archive Files (참고용)
```
IMPLEMENTATION-GUIDE-PHASE3-4.md
PHASE4-VALIDATION.md
TEAM_MODE_GITHUB_INTEGRATION_ANALYSIS.md
WORKFLOW-IMPLEMENTATION-REPORT.md
```

---

## 🔍 검증 명령어

### 메모리 파일 참조 확인
```bash
# 모든 메모리 파일 참조 찾기
rg "\.moai/memory|CLAUDE-|CONFIG-|DEVELOPMENT-|GITFLOW-|SPEC-METADATA|SKILLS-" \
   -g "*.md" \
   -n

# 특정 파일 참조 추적
rg "SPEC-METADATA\.md" -g "*.md" -n

# 깨진 링크 확인
rg "\[.*\.md\]" -g "*.md" -n | grep -v "http"
```

### 파일명 일관성 확인
```bash
# 소문자 파일명 찾기 (모두 대문자여야 함)
ls .moai/memory/ | grep -v "^[A-Z]"

# 대문자 파일명 확인
ls .moai/memory/ | grep "^[A-Z]"
```

---

## 📝 유지보수 체크리스트

- [ ] 모든 메모리 파일명이 대문자인가?
- [ ] `CLAUDE.md`에서 모든 메모리 파일 참조가 올바른가?
- [ ] Commands에서 메모리 파일 참조가 정확한가?
- [ ] Sub-agents가 필요한 메모리 파일에 접근 가능한가?
- [ ] 깨진 링크가 없는가?
- [ ] JIT 로딩 전략이 적용되었는가?
- [ ] 중복 내용이 없는가? (SSOT 원칙)

---

## 🔗 관련 문서

- `CLAUDE.md`: 문서 라우팅 맵
- `CLAUDE-AGENTS-GUIDE.md`: Agent 선택 가이드
- `CLAUDE-RULES.md`: Skill/TAG/Git 규칙
- `DEVELOPMENT-GUIDE.md`: SPEC-First TDD 워크플로우

---

**작성**: Alfred (MoAI-ADK SuperAgent)
**최종 업데이트**: 2025-10-27
**상태**: 완료 ✅
