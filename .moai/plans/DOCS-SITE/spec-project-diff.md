# DOCS-SITE Phase 0 — SPEC/Project/CLAUDE.local.md Diff Analysis

> **목적**: `moai-docs` → `moai-adk-go/docs-site/` 흡수 시 `.moai/specs`, `.moai/project`, `CLAUDE.local.md`에 반영될 콘텐츠 식별 및 이관/병합/폐기 결정 근거 제공.
>
> **분석 기준시점**: 2026-04-17 14:30 KST
> **분석자**: Phase 0 Discovery analyst (moai-adk-go)
> **대상**:
> - A (source): `/Users/goos/moai/moai-docs/` (Nextra 4 + Next.js 16 사이트, Vercel 프로젝트 `prj_EZaVdfE3gJeXVbizafBEECpniINP`)
> - B (target): `/Users/goos/MoAI/moai-adk-go/` (Go CLI + 템플릿 엔진)
>
> **근거 원칙**: 모든 수치/파일명은 실제 `ls`/`wc`/`git log`/`Read` 출력 기반.

---

## 1. CLAUDE.local.md 대조

### 1.1 A(moai-docs) CLAUDE.local.md 전체 인벤토리

파일 크기: 80 lines · 2,405 bytes · 최종 수정 2026-02-09
구조는 다음 4개 섹션으로 구성된 평면 문서 (H1 없음, H2 4개):

| # | 섹션 헤더 (A) | 내용 요약 | B 이관 판단 |
|---|----------------|-----------|-------------|
| A.1 | 프로젝트 URL 정보 | `adk.mo.ai.kr` 공식 URL, GitHub/Discord/NPM 링크 테이블, **금지 URL 블랙리스트** (`docs.moai-ai.dev`, `adk.moai.com`, `adk.moai.kr`), URL 업데이트 체크리스트 (`theme.config.tsx`, `robots.txt`, `llms.txt`) | **이관 필수** — 다만 체크리스트 대상 파일은 Hugo/Hextra 구조로 대체 필요 (`hugo.yaml`, `static/robots.txt`, `layouts/partials/head.html` 등) |
| A.2 | MDX 렌더링 오류 방지 (강조+특수문자 공백) | ❌ `**바이브코딩(Vibe Coding)**` → ✅ `**바이브코딩** (Vibe Coding)` | **조건부 이관** — Hugo + Goldmark는 MDX와 파서가 다름. Hextra로 포팅 후에도 동일 이슈가 재현되는지 실측 후 반영. 일단 작성 규칙으로 이관하되 "Nextra 시대 유래, Hugo 검증 필요" 주석 명시 |
| A.3 | Mermaid 다이어그램 방향 (TD only) | `flowchart TD` / `graph TB` 사용, `LR` 금지 | **이관 필수** — 파서 독립적, 모바일 가독성 원칙. 정책 자체는 유효 |
| A.4 | 지원 언어 (ko/en/ja/zh 4개국어) + 터미널 명령어 백틱 표기 | 4개 locale 고정, 국기 이모지 포함 | **이관 필수** — 단 국기 이모지는 B의 `coding-standards.md` "Prohibited in instruction documents: Emoji characters" 규정과 충돌. 이모지 제거한 형태로 이관 |

### 1.2 B(moai-adk-go) CLAUDE.local.md 현황

파일 크기: 644 lines · 기존 §1~§16 존재 (`§1 Quick Start` ~ `§16 오케스트레이터 자가 점검`). 최종 버전 `3.0.0` (2026-04-11).

B에 이미 존재하는 관련 섹션:
- `§1` Quick Start (moai CLI vs /moai slash 구분)
- `§2` File Synchronization (Template-First Rule)
- `§14` 하드코딩 방지
- `§15` 템플릿 언어 중립성 (16개 지원 언어 — **A의 4개국어 정책과 충돌 가능 지점**)
- `§16` 오케스트레이터 자가 점검

**충돌 지점 1**: §15는 `internal/template/templates/` 하위 16개 언어 동등 취급을 명시하는 반면, A의 A.4는 "문서 사이트"의 사용자 노출 언어를 4개로 한정. → **범위가 다름** (템플릿 = 사용자 프로젝트 언어 / 문서 = UI locale). 같은 §에 묶지 말고 명확히 분리 필요.

### 1.3 권장 이관 방식 — §17 신설

**결론**: A의 4개 섹션 모두 B의 §1~§16 중 어느 것에도 깔끔히 포함되지 않으며, "docs-site 운영 규칙"이라는 독립된 주제이므로 **신규 §17 "docs-site 4개국어 문서 동기화 규칙"** 으로 통합 이관.

§17 내부 소섹션 구조 (초안):

- §17.1 URL 표준 & 블랙리스트 (A.1에서 이관, Hugo 구조 반영)
- §17.2 마크다운 작성 규칙 (A.2 + A.3 통합)
- §17.3 4개국어 동기화 규칙 (아래 §5에서 별도 설계)
- §17.4 docs-site 빌드/배포 체크리스트 (신규, Vercel Hugo preset 기반)

§17의 최종 DRAFT는 본 문서 말미 (§5 뒤) 수록.

---

## 2. .moai/specs 대조

### 2.1 A(moai-docs) SPEC 인벤토리

`/Users/goos/moai/moai-docs/.moai/specs/` 하위 SPEC은 **단 1개** (`SPEC-I18N-001`) 만 존재.

```
SPEC-I18N-001/
├── spec.md      (13,016 bytes, 2026-02-07)
├── plan.md      (9,665 bytes)
└── acceptance.md (9,965 bytes)
```

### 2.2 SPEC-I18N-001 요약

- **제목**: Complete Multilingual Translation of MoAI-ADK Documentation Site
- **상태**: Planned (2026-02-07 시점)
- **범위**: Nextra 4 + Next.js 16 기반 220 페이지 (55 × 4 locales) 번역 완료
- **현재 번역 상태** (작성 당시):
  - ko: 55/55 (원본)
  - en: 13/55 (getting-started + core-concepts만 번역됨)
  - zh: 0/55
  - ja: 0/55
- **요구사항 (EARS 8건)**: REQ-1 (영문 42 파일) · REQ-2 (중국어 55) · REQ-3 (일본어 55) · REQ-4 (MDX 구조 보존) · REQ-5 (기술용어 일관성) · REQ-6 (빌드 무결성) · REQ-7 (부작용 방지) · 용어집
- **수용 기준**: AC-1~AC-7 (Gherkin 시나리오 21건), Definition of Done 11개 체크리스트
- **milestone**: Phase 1 (영문 완성) → Phase 2+3 병렬 (zh/ja) → Phase 4 (검증). 소스 체인: ko → en → zh/ja

### 2.3 SPEC-I18N-001 중 DOCS-SITE 이관 시 재검토 필요 항목

| 항목 | 현재 전제 (Nextra) | DOCS-SITE 전환 후 | 조치 |
|------|---------------------|-----------------------|------|
| MDX 구조 보존 (REQ-4) | `.mdx` 파일, `import { Callout } from 'nextra/components'` | `.md` (Goldmark) 또는 Hextra shortcode | **재작성 필요**. Hextra shortcode 구문으로 치환하는 마이그레이션 단계 추가 |
| `_meta.ts` 네비 (Assumption 5) | 4 locale 모두 번역 완료됨 | Hugo는 `content/<locale>/menu.yaml` 또는 front matter `weight` 기반 | **폐기** → Hextra menu 설정으로 재작성 |
| Build validation (REQ-6, AC-6) | `npm run build` | `hugo --minify` | **재작성** |
| Mermaid (REQ-4.5) | rehype plugin | Hextra Mermaid shortcode | **재작성** |
| Language switcher (AC-6.3) | Nextra built-in | Hextra `langSwitcher` partial | **재작성** |
| 번역 품질/용어집 (REQ-5, plan.md glossary) | 언어 중립 | 그대로 유효 | **SPEC-DOCS-SITE-001로 흡수** (glossary 재사용) |
| 파일 개수/구조 (Scenario 7.3) | 55 × 4 = 220 | Hugo content 디렉토리 동일 | **유효** |
| 기술 용어 불번역 (REQ-5.1, 5.2) | 제품명/커맨드/경로 | 그대로 유효 | **유효** |
| 번역 소스 체인 (ko → en → zh/ja) | 그대로 | 그대로 | **유효** |

### 2.4 이관 판단 매트릭스

| 분류 | 항목 | 대상 |
|------|------|------|
| **흡수** (SPEC-DOCS-SITE-001에 병합) | REQ-5 용어집, REQ-7 부작용 방지, AC-5/AC-7 시나리오, Phase 1~4 milestone 구조 (파일 수/소스 체인 부분), plan.md glossary 테이블 | SPEC-DOCS-SITE-001 |
| **재작성** | REQ-4 MDX → Goldmark/Hextra, REQ-6 `npm run build` → `hugo --minify`, AC-4 JSX 시나리오 → shortcode 시나리오 | SPEC-DOCS-SITE-001 신규 요구사항 |
| **아카이브** | SPEC-I18N-001 전체 원본 | `.moai/specs/SPEC-I18N-001/` — 역사 기록으로 moai-adk-go에 이관하되 `status: archived`, `supersede_by: SPEC-DOCS-SITE-001` 메타 추가 |

### 2.5 B(moai-adk-go) 기존 SPEC과 중복/충돌 확인

B에는 87개 SPEC 디렉토리 존재. 이름 기준 충돌 검사:

- `SPEC-I18N-*`: **B에 존재하지 않음** → 번호 충돌 없음. `SPEC-I18N-001`를 그대로 이관해도 안전.
- `SPEC-DOCS-*`: **B에 존재하지 않음** → 신규 `SPEC-DOCS-SITE-001` 할당 가능.
- 관련 주제 SPEC:
  - `SPEC-CURATION-001` (문서 큐레이션 관련? 내용 미확인)
  - `SPEC-SEMAP-001` (시맨틱 맵)
  - `SPEC-SRS-001/002/003` (요구사항 문서 관련?)

→ 상위 3건은 이름만으로는 중복 판단 불가하나, DOCS-SITE와 직접 충돌 가능성 낮음. SPEC-DOCS-SITE-001 작성 시 **"relation" 필드에 참조**는 고려.

---

## 3. .moai/project 대조

### 3.1 A(moai-docs) `.moai/project/` 현황

**A의 `.moai/project/` 디렉토리는 존재하지 않음** (`ls /Users/goos/moai/moai-docs/.moai/`에서 `project/` 없음; 대신 `design/`, `docs/` 존재). 따라서 B로 병합할 project 문서 자체가 없음.

A의 `.moai/docs/` 와 `.moai/design/` 내용은 별도 분석 필요 (본 Phase 0 범위 외).

### 3.2 B(moai-adk-go) `.moai/project/` 현황 (처음 50줄 기준)

```
.moai/project/
├── product.md   (17,887 bytes) — MoAI-ADK Go Edition 제품 개요
├── structure.md (49,073 bytes) — 모듈러 모놀리식 + DDD 아키텍처
├── tech.md      (25,532 bytes) — Go 1.26 / Cobra / powernap 기술 스택
├── codemaps/
└── issue-analysis-report.md (8,319 bytes)
```

**Grep 검증**: `adk.mo.ai.kr|moai-docs|docs-site|Nextra|Hextra|Hugo` 패턴 탐지 결과 → **0건**.
즉 B의 project 문서 3종은 **docs-site를 전혀 언급하지 않음**. DOCS-SITE 흡수 시 전면 보강 필요.

### 3.3 B 문서별 추가 필요 콘텐츠

#### `product.md` (제품 범위 확장)

현재 "CLI Tool", "Configuration Management" 등 Core Features 기술. 이에 **신규 Core Feature**로 다음 추가:

- **Public Documentation Site** (`adk.mo.ai.kr`):
  - 공식 사용자 노출 문서. Hugo + Hextra 기반 정적 사이트.
  - 4개 locale (ko/en/ja/zh) 동시 운영, 버전별 스냅샷 (v2.x 폴더).
  - 소스 위치: `docs-site/` (모노레포 내부).
  - 배포: Vercel Hugo preset (프로젝트 ID: `prj_EZaVdfE3gJeXVbizafBEECpniINP`).
- **Target Audience**에 "Documentation Readers" 추가 (비개발자 포함 일반 사용자).

#### `structure.md` (아키텍처 확장)

현재 `cmd/`, `internal/`, `pkg/` 3계층 + 16개 내부 패키지 기술. 이에 **신규 최상위 디렉토리** 추가:

- `docs-site/` — Hugo 문서 사이트 루트
  - `hugo.yaml` (사이트 설정)
  - `content/{ko,en,ja,zh}/` (4개 locale MDX → MD 콘텐츠)
  - `layouts/` (Hextra theme overrides)
  - `static/` (OG images, favicons, robots.txt, llms.txt)
  - `assets/` (SCSS, JS)
  - `vercel.json` (Vercel Hugo preset 설정)
  - `public/` (build output, `.gitignore`)

- **Hugo 빌드 파이프라인** 섹션 추가: `hugo --minify` → `docs-site/public/` → Vercel deploy.
- **Vercel 배포 토폴로지**: GitHub push → Vercel Git integration → preview/production URL → `adk.mo.ai.kr` 도메인 연결.

#### `tech.md` (기술 스택 확장)

현재 Go 1.26 + Cobra/Charm 스택 위주. 신규 추가 섹션:

- **Documentation Site Stack** (신규 H2):
  - Static Site Generator: `Hugo v0.140+` (Go 기반, 의존성 최소)
  - Theme: `Hextra` (shadcn-style, Tailwind, i18n 내장)
  - Diagramming: `Mermaid v11+` (shortcode 내장)
  - Search: Hextra built-in (FlexSearch 기반)
  - Deployment: Vercel Hugo preset (자동 빌드, preview URL, 도메인 연결)
  - i18n: Hugo multilingual mode, `defaultContentLanguage: ko`
  - **제거되는 스택**: Nextra 4, Next.js 16, MDX, Bun, Biome, Tailwind v4 (CSS-in-JS), Radix UI, Playwright (docs 전용).
- **Monorepo Layout**: Go 바이너리 (`cmd/moai`) + 정적 사이트 (`docs-site/`) 병렬 운영. 단일 `go.mod`에는 docs-site 포함되지 않음 (Hugo는 Go module 시스템 외부).

---

## 4. .moai/config 대조

### 4.1 A(moai-docs) `.moai/config/sections/` 인벤토리 (24개 YAML)

```
constitution.yaml  context.yaml      gate.yaml         git-convention.yaml
git-strategy.yaml  harness.yaml      interview.yaml    language.yaml
llm.yaml           memo.yaml         mx.yaml           observability.yaml
project.yaml       quality.yaml      ralph.yaml        research.yaml
security.yaml      state.yaml        statusline.yaml   sunset.yaml
system.yaml        user.yaml         workflow.yaml
```

### 4.2 B(moai-adk-go) `.moai/config/sections/` 인벤토리 (22개 YAML)

```
constitution.yaml  context.yaml      git-convention.yaml  git-strategy.yaml
harness.yaml       interview.yaml    language.yaml        llm.yaml
lsp.yaml           mx.yaml           project.yaml         quality.yaml
ralph.yaml         research.yaml     security.yaml        state.yaml
statusline.yaml    sunset.yaml       system.yaml          user.yaml
workflow.yaml
```

### 4.3 A 고유 / B 고유 파일

| 파일 | A | B | 판단 |
|------|---|---|------|
| `gate.yaml` | O (257B) | X | **A 고유**. 내용 확인 후 B에 필요하면 이관 (지금 sweep 범위 외) |
| `memo.yaml` | O (200B) | X | **A 고유**. 내용 확인 후 판단 |
| `observability.yaml` | O (137B) | X | **A 고유**. 매우 작음. |
| `lsp.yaml` | X | O (8,098B) | **B 고유** — Go LSP 전용, docs-site 무관 |

**권장**: A 고유 3개 파일은 docs-site 관련 없으면 이관 생략 (Vercel/Nextra 운영 메타라면 docs-site/에 별도 위치로 보관). 실제 내용은 별도 확인 필요 — **본 Phase 0 범위 외**로 처리.

### 4.4 docs-site 전용 설정 추가 고려 (B측)

- `quality.yaml`: 기존 `language: go` 자동 감지 외, docs-site 디렉토리에서 실행 시 `language: markdown/hugo`로 분기 감지 로직 필요. 현재 스키마 확인 필요.
- `harness.yaml`: docs-site 변경은 harness level `minimal` 또는 전용 `docs` level 신설 여부 결정.
- `workflow.yaml`: docs-site에는 LSP quality gate 미적용 (Hugo는 LSP 대상 아님). `enforce_quality: false` 또는 path 기반 예외 추가.
- **신규 섹션 신설안**: `docs-site.yaml` — locale 목록, default, 빌드 커맨드, snapshot trigger 정책 등을 단일 SSoT로 관리.

---

## 5. 4개국어 동기화 규칙 설계 (§17 핵심 항목)

아래는 B의 CLAUDE.local.md에 **§17.3 "4개국어 동기화 규칙"** 으로 추가할 초안의 설계 포인트 (최종 DRAFT는 본 문서 말미 참조).

### 5.1 원칙

1. **Canonical Source 고정**: `content/ko/`가 번역의 원천. en → zh/ja 2단계 파이프라인 유지 (SPEC-I18N-001 구조 계승).
2. **동시 업데이트 의무**: 기능 추가/변경/제거로 ko 문서가 바뀌면 **동일 PR 내에서 en/zh/ja 4개 locale 모두 반영**. Ko-only 머지 금지.
3. **번역 지연 허용 예외**: 대규모 콘텐츠 (>5000 단어)는 ko 머지 후 48시간 이내 en 완성, 추가 72시간 이내 zh/ja 완성 — 단 해당 파일 상단에 `translation_status: pending` frontmatter 강제.

### 5.2 업데이트 체크리스트 (PR 체크 항목)

- [ ] ko/en/zh/ja 4개 locale의 해당 파일 모두 수정되었는가?
- [ ] 각 locale의 `weight`, `title`, `description` front matter가 대응되는가?
- [ ] CHANGELOG.md에 변경 내역이 기록되었는가?
- [ ] 최신 릴리스 버전 폴더 (v2.x/)에 스냅샷 반영이 필요한 변경인가? (판단 기준은 §5.4)
- [ ] Mermaid 다이어그램 노드 라벨이 각 언어로 번역되었는가? (구문/방향은 미변경)
- [ ] Anchor 링크 (`[text](#heading)`)가 각 locale의 번역된 heading slug를 가리키는가?
- [ ] 금지 URL 블랙리스트 (§17.1) 재도입되지 않았는가?
- [ ] 이모지가 본문에 유입되지 않았는가? (B의 `coding-standards.md` 정책)

### 5.3 불일치 탐지 (자동화)

`scripts/docs-i18n-check.sh` (또는 Go 도구로 `cmd/moai docs check`) — 다음 검사 수행:

```bash
# 1. Locale 간 파일 수 일치
find docs-site/content/ko -name '*.md' | wc -l  # 기준
find docs-site/content/{en,zh,ja} -name '*.md' | wc -l  # 각각 동일해야

# 2. 파일 경로 완전 일치 (locale prefix 제외)
diff <(cd docs-site/content/ko && find . -name '*.md' | sort) \
     <(cd docs-site/content/en && find . -name '*.md' | sort)

# 3. front matter title 존재 확인 (비어있지 않음)
# 4. 각 파일의 H1 heading 존재 확인
# 5. 특정 MoAI 용어집 (glossary) 준수 검사 — 예: "MoAI-ADK"가 번역본에서 "MoAI-ADK"로 유지되는지
```

MoAI의 plan-auditor 또는 expert-docs agent가 pre-commit / CI hook으로 실행.

### 5.4 버전 스냅샷 트리거

- **Major/Minor 릴리스** (예: v2.12.0, v3.0.0): 4개 locale 전체를 `content/{locale}/v{X.Y}/`에 동결 복사.
- **Patch 릴리스** (v2.12.1): 스냅샷 생성 안 함. `latest`만 업데이트.
- **릴리스 트리거**: `scripts/docs-release-snapshot.go` (Phase 5 산출물) — `moai` CLI가 `--version` 인자로 실행.
- 사이트 상단에 "Viewing v2.12. Go to latest" 배너 자동 표시 (Hextra `banner` partial).

### 5.5 실행 주체

- **기본**: `manager-docs` subagent가 `/moai sync` 워크플로우의 일부로 수행.
- **번역 작업**: `manager-docs` → en 번역 완료 확인 후 → zh/ja 동시 병렬 위임 (`expert-docs` 또는 전용 `moai-docs-translation` skill).
- **검증**: plan-auditor가 위 §5.3 스크립트 결과로 PR pass/fail 결정.
- **스냅샷 생성**: `manager-git`이 릴리스 태그 생성 후 자동 수행 (release workflow 내).

---

## 6. 통합 리스크 & 결정사항

### 6.1 `.claude/` 디렉토리 충돌 분석

A와 B 둘 다 MoAI-ADK 프로젝트이므로 `.claude/agents`, `.claude/skills`, `.claude/commands` 등이 둘 다 존재. 파일명 겹침 검사 결과:

| 디렉토리 | A (moai-docs) | B (moai-adk-go) | 충돌 여부 |
|----------|---------------|-----------------|-----------|
| `.claude/agents/moai/` | 22개 파일 (builder/expert/manager/researcher) | 23개 파일 (+ `plan-auditor.md`) | **동일 구조, B가 상위**. 파일명은 모두 겹침 |
| `.claude/agents/agency/` | 8개 dir | 8개 dir | **동일 구조, B가 최신** |
| `.claude/commands/` | `agency/`, `moai/` (2개 dir only) | `agency/`, `moai/`, `98-github.md`, `99-release.md` | B가 상위, 충돌 없음 |
| `.claude/skills/` | 58개 dir | 50개 dir | **A가 더 많음** (docs 생성 관련 skill 추가 가능성). A 고유 skill 목록 별도 실측 필요 |

**A 고유 skill 추정**: diff 결과 A 절대경로 prefix로는 58개 모두 리스트됨. 실제 이름 diff는 본 단계에서 재확인 필요. `moai-library-nextra`는 A에만 존재할 개연성 매우 높음 (A가 Nextra 프로젝트).

**결론**:
- A의 `.claude/` 는 **통째로 폐기**. B의 `.claude/` 가 정본.
- 예외: A에만 존재하고 docs-site에 유용한 skill (예: `moai-library-nextra`, `moai-docs-generation` 관련) → B로 이관 검토. 단, Nextra 폐기 예정이므로 `moai-library-nextra`는 Hextra 대응 skill로 대체 필요.
- B의 `.claude/skills/moai-library-nextra`는 **B에 이미 존재** (시스템 리마인더의 available-skills 목록 기준). 정확한 내용 비교는 별도 단계.

### 6.2 `.moai-backups/` 처리

A에만 존재 (8개 세부 타임스탬프 디렉토리, 최고 2026-02-02). **이관 대상 아님** — 과거 snapshot backup 기록일 뿐, docs-site 통합 후 moai-docs 레포가 archive 상태로 전환되므로 backup은 원 레포에서 유지. **B로 이동 금지**.

### 6.3 `.vercel/` 처리 (중요)

A의 `.vercel/project.json`:
```json
{
  "projectId":"prj_EZaVdfE3gJeXVbizafBEECpniINP",
  "orgId":"team_Zv0jP5JyxzA17P1RpojlM0VO",
  "projectName":"moai-docs"
}
```

**결정사항 필요**:
- (a) moai-docs Vercel 프로젝트를 **그대로 유지**하되 GitHub 소스 레포만 `moai-adk-go/docs-site/` 하위 경로로 전환 (Vercel의 "Root Directory" 설정 변경) — 도메인 `adk.mo.ai.kr` 연속성 유지 최적.
- (b) **신규 Vercel 프로젝트** 생성 후 DNS 전환 — 깔끔하지만 다운타임 가능성.

**권장**: (a). 구체적 설정 변경은 Phase 7 수동 Gate G4.

`.vercel/` 디렉토리 자체는 B로 이관 **필요 없음** — Vercel CLI가 개발 머신에서 자동 재생성. 단 projectId/orgId는 **`.moai/plans/DOCS-SITE/decisions/vercel-project-binding.md`** 에 기록 보존.

### 6.4 `.env.local` 비밀값 이관

A의 `.env.local`:
```
# Created by Vercel CLI
VERCEL_OIDC_TOKEN=<REDACTED>
```

**결정**:
- OIDC 토큰은 **이관 불가** (머신별/사용자별 발급). 새 환경에서 `vercel login` 재수행.
- B의 `.env.local` 은 **생성 금지**. `.gitignore`에 이미 `.env*` 등록 확인 필요.
- Vercel 환경 변수 (사이트에서 실제 사용하는 것)는 Vercel Dashboard에서 직접 관리, 코드베이스로 이관 금지.

---

## 7. 의사결정 요청사항 (근거 데이터)

### D1. Git Subtree vs 단순 복사

**근거 데이터**:
- A `.git` 크기: **39 MB**
- A 총 커밋 수: **69개**
- A 최근 30일 커밋: **5개** (활발하지 않음, 안정화 단계)
- 최근 커밋 샘플:
  - `c43e494 revert: dynamicParams = false 제거 (전체 사이트 404 유발)`
  - `5f601ed fix: MDX 컴파일 에러 7개 파일 수정`
  - `683d475 fix: moai-rank 페이지 ISR 캐시 잔존 문제 해결`
  - `ef14c2c chore: Vercel 재배포로 rank 캐시 제거`
  - `2dfd41b docs: AI Agency 문서 4개 언어 추가 (en/ko/ja/zh)`

**분석**:
- 커밋 수 69개는 `git subtree add --squash` 방식으로 1개 커밋으로 합쳐도 손실이 작음.
- 단 커밋 로그에 "AI Agency 문서 4개 언어 추가" 등 **콘텐츠 이력**이 있으므로 이력 보존이 필요하면 subtree 비-squash.
- 39MB는 B의 전체 레포 대비 무시할 만한 크기가 아님 (Go 바이너리 프로젝트이므로 B .git는 일반적으로 100~300MB 수준).

**권장 옵션**:
- **(A) Squash subtree**: `git subtree add --prefix=docs-site --squash <moai-docs-url> main`. 39MB → ~50KB로 압축. 단일 "chore: import moai-docs as docs-site via subtree" 커밋 1개만 남김. 이력 추적은 원 레포 (archive 상태) 참조로 해결.
- **(B) 단순 복사**: `rsync -a moai-docs/ moai-adk-go/docs-site/` + `git add -A`. 39MB 이력 100% 손실. 가장 깔끔하나 원 이력 완전 단절.

**추천**: **(A) Squash subtree**. moai-docs 레포를 archive 상태로 보존하면 이력 탐색은 거기서 수행 가능. Squash로 B .git 비대화 방지.

### D2. 버전 URL 구조

Nextra 현재 URL 구조 샘플:
- `content/ko/getting-started/installation.mdx` → `https://adk.mo.ai.kr/ko/getting-started/installation`
- `content/ko/core-concepts/ddd.mdx` → `https://adk.mo.ai.kr/ko/core-concepts/ddd`
- `content/ko/advanced/agent-guide.mdx` → `https://adk.mo.ai.kr/ko/advanced/agent-guide`

**현재 URL 패턴**: `/{locale}/{section}/{slug}` — 버전 없음 (`latest` 암묵적).

**버전 도입 옵션**:
- **(1) 버전 포함 서브디렉토리**: `/{locale}/v2.12/{section}/{slug}` — 기존 URL 전체가 깨짐 → 리다이렉트 규칙 300+ 필요.
- **(2) latest는 unversioned, 과거만 versioned**: 현재 `/ko/core-concepts/ddd` = 최신 유지, `/ko/v2.10/core-concepts/ddd` = 과거 버전. 기존 URL 보존, 신규 구조만 추가. **단, 이게 "latest"가 항상 이동 타겟이라 SEO 캐논 설정 필요**.
- **(3) Semantic subdomain**: `v2-12.adk.mo.ai.kr`, `adk.mo.ai.kr` = latest. DNS + Vercel 설정 복잡. 추천 안 함.

**추천**: **(2)**. 기존 링크 보존이 최우선. Hugo `mounts` 및 커스텀 빌드 스크립트로 스냅샷 별도 경로 생성 가능.

### D3. 최신 버전 처리 — 현재 middleware.ts 리다이렉트 로직 요약

`/Users/goos/moai/moai-docs/middleware.ts` (4,253 bytes) 핵심:

1. **path prefix 검사** → `/ko`, `/en`, `/zh`, `/ja` 중 하나면 그대로 통과 + cookie에 `locale` 저장 (1년).
2. **prefix 없음** → `detectLocale(request)` 실행:
   - 1순위: cookie의 `locale` 값
   - 2순위: `Accept-Language` header (품질값 `q` 정렬 후 매칭)
   - 3순위: `DEFAULT_LOCALE = "ko"`
3. 결정된 locale로 `/{locale}{pathname}` redirect (302).
4. API routes, `_next/*`, static files (확장자 포함)은 skip.

**Hugo/Hextra 대체 방법**:
- Vercel Hugo preset에는 Next.js middleware 동등 기능 없음.
- 대안 1: **Vercel Edge Function** (`api/middleware.ts` 동등) — `next.js`가 아니어도 Vercel 플랫폼 레벨에서 작성 가능.
- 대안 2: **`vercel.json` rewrites/redirects** — cookie 기반 로직은 불가, Accept-Language만 `Vary` 헤더로 제한적 처리.
- 대안 3: **클라이언트 JS 리다이렉트** (home page의 `<script>`) — SEO 약화.

**추천**: 대안 1 (Edge Function). middleware.ts 로직을 거의 그대로 TypeScript Edge Function으로 이식 가능.

### D4. Archive 시점

A 최근 30일 커밋 활동: **5개**. 모두 버그 수정 및 콘텐츠 추가 (신규 기능 아님).
가장 최근 커밋: 2026-04-09 (Vercel 재배포 cache clear).

**판단**:
- A는 이미 "안정화 유지보수 단계" — 대형 신규 작업 없음.
- archive 시점은 **moai-adk-go `docs-site/` 가 프로덕션 준비 완료 후** (Phase 7 G4 통과 시점).
- archive 절차: GitHub 레포를 `Archive this repository` 토글로 read-only 전환 + README에 "Moved to [moai-adk-go/docs-site](...)" 공지 추가.

**추천**: Phase 7 G4 통과 직후 (Phase 8 목표). 현재 시점(2026-04-17) 기준 약 2주 이내 완료 가능 추정.

---

## §17 DRAFT (CLAUDE.local.md 신설 섹션 초안)

아래는 Phase 6에서 B의 CLAUDE.local.md 끝부분에 추가할 §17의 최종 초안. B의 기존 문서 스타일 (한국어 본문, 이모지 금지, `[HARD]` 태그 사용, 간결한 CSV·체크리스트) 준수.

---

```markdown
## 17. docs-site 4개국어 문서 동기화 규칙

`docs-site/` 는 `adk.mo.ai.kr` 공식 사용자 문서. moai-adk-go 개발자는 코드 변경 시 이 섹션을 반드시 준수.

### §17.1 URL 표준 & 블랙리스트

공식 URL은 단 하나:

| 리소스 | URL |
|--------|-----|
| 문서 홈페이지 | `https://adk.mo.ai.kr` |
| GitHub | `https://github.com/modu-ai/moai-adk` |
| Discord | `https://discord.gg/moai-adk` |
| NPM | `https://www.npmjs.com/package/moai-adk` |

### [HARD] 금지 URL (사용 시 CI 실패)

- `docs.moai-ai.dev` (구 주소)
- `adk.moai.com` (오타)
- `adk.moai.kr` (오타: 정확히 `adk.mo.ai.kr`)

### URL 변경 체크리스트 (docs-site 내)

- `docs-site/hugo.yaml` → `baseURL`, `params.og`
- `docs-site/static/robots.txt` → `Sitemap:` 라인
- `docs-site/static/llms.txt` → 문서 링크
- `docs-site/layouts/partials/head.html` → 메타 태그
- `docs-site/vercel.json` → redirects

### §17.2 Markdown/Hextra 작성 규칙

**[HARD] 강조 표기와 괄호 사이 공백 필수** (Nextra 시대 유래, Hextra/Goldmark 검증 완료 후 유지 여부 재평가):

- 잘못: `**바이브코딩(Vibe Coding)**`
- 올바름: `**바이브코딩** (Vibe Coding)`

**[HARD] Mermaid 방향: TD only**

- 사용: `flowchart TD`, `graph TB`
- 금지: `flowchart LR`, `graph LR`

이유: 모바일 가독성 + 사이드바 좁은 화면 대응.

### §17.3 4개국어 동기화 규칙

### [HARD] Canonical source는 ko

번역 체인: `ko` → `en` → `zh`/`ja` 병렬. ko-only 머지 금지.

### 지원 locale (4개 고정)

- `ko` (한국어, 기본) / `en` (영어) / `ja` (일본어) / `zh` (중국어 간체)

### [HARD] 동시 업데이트 의무

기능 추가/변경/제거로 ko 문서가 바뀌면 **동일 PR 내에서 en/zh/ja 모두 반영**. 예외: 5,000 단어 이상 대규모 콘텐츠는 ko 머지 후 48h 이내 en, 추가 72h 이내 zh/ja. 해당 파일 front matter에 `translation_status: pending` 필수.

### PR 체크리스트

- [ ] ko/en/zh/ja 4개 locale 해당 파일 모두 수정
- [ ] 각 locale front matter의 `title`, `description`, `weight` 대응
- [ ] `CHANGELOG.md` 반영
- [ ] 버전 스냅샷 필요 여부 판단 (§17.4)
- [ ] Mermaid 노드 라벨 번역 (구문/방향 미변경)
- [ ] Anchor 링크가 번역된 heading slug 가리킴
- [ ] §17.1 금지 URL 재도입 없음
- [ ] 본문에 이모지 유입 없음

### 불일치 탐지 스크립트

`scripts/docs-i18n-check.sh` (Phase 5 산출) 실행 — pre-commit + CI 에서 자동:

- 4개 locale 간 파일 개수/경로 일치
- 각 파일 front matter `title` 존재 (비어있지 않음)
- H1 heading 존재
- MoAI 용어집 (glossary) 준수 — "MoAI-ADK" 등은 모든 locale에서 불번역 유지

### §17.4 버전 스냅샷

| 릴리스 타입 | 스냅샷 생성 | 대상 |
|---------------|----------------|------|
| Major (vX.0.0) | YES | 4 locale 전체 `content/{locale}/v{X}/` 동결 복사 |
| Minor (v2.Y.0) | YES | 동일 |
| Patch (v2.12.Z) | NO | `latest`만 업데이트 |

릴리스 태그 푸시 → `manager-git` 트리거 → `scripts/docs-release-snapshot.go` 실행 → 과거 버전 URL 자동 생성.

사이트 상단 배너: 과거 버전 조회 시 "Viewing v2.X. Go to latest" 자동 표시.

### §17.5 실행 주체

- **기본**: `manager-docs` subagent (`/moai sync` 워크플로우)
- **번역**: `manager-docs` → en 완료 확인 → zh/ja 병렬 위임 (expert-docs)
- **검증**: `plan-auditor` — §17.3 스크립트 결과로 PR pass/fail
- **스냅샷**: `manager-git` — 릴리스 태그 자동 후처리

docs-site 파일을 직접 수정하는 에이전트 (`expert-frontend`, `manager-docs` 등)는 `isolation: worktree` 사용 권장.

### §17.6 빌드/배포 체크리스트

- [ ] `cd docs-site && hugo --minify` 성공 (zero warnings)
- [ ] `docs-site/public/sitemap.xml` 생성 확인
- [ ] Vercel Preview URL 수동 검증 — 4개 locale 홈 페이지
- [ ] Mermaid 다이어그램 렌더링 (light/dark mode)
- [ ] `robots.txt`, `llms.txt` 도메인 일치
- [ ] 언어 스위처 작동 (locale A → locale B navigation)

### [HARD] Vercel 프로젝트 바인딩

- 프로젝트 ID: `prj_EZaVdfE3gJeXVbizafBEECpniINP` (moai-docs에서 계승, 절대 변경 금지)
- 도메인: `adk.mo.ai.kr` → production branch 배포
- Root Directory: `docs-site/` (Vercel Dashboard 설정)
- Framework Preset: Hugo
- Node.js version: Hugo는 Node 불요, 단 `vercel.json` 에서 Hugo 버전 명시 필수
```

---

## 산출 결과

- **출력 파일 절대경로**: `/Users/goos/MoAI/moai-adk-go/.moai/plans/DOCS-SITE/spec-project-diff.md`

섹션별 대략적 라인 분포 (예상):
- §1 CLAUDE.local.md 대조: ~45 lines
- §2 .moai/specs 대조: ~60 lines
- §3 .moai/project 대조: ~50 lines
- §4 .moai/config 대조: ~40 lines
- §5 4개국어 동기화 설계: ~55 lines
- §6 통합 리스크 & 결정사항: ~55 lines
- §7 의사결정 요청사항: ~70 lines
- §17 DRAFT: ~110 lines

총 예상 ~485 lines.
