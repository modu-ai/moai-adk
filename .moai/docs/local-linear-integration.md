# Linear 연동 (개인/로컬 전용)

> Extracted from `CLAUDE.local.md` §26 for context-budget diet. Canonical reference for Linear ↔ SPEC bridge rules, workspace mapping, and idea-capture triggers. Read on demand.

## 26. Linear 연동 (개인/로컬 전용)

> [ZONE:Local-Only] 본 섹션은 GOOS 개인 개발 전용이며 로컬 전용이다. `internal/template/templates/`에 절대 미러 금지 (§25 isolation + §14 하드코딩 허용 영역). **범용 배포/제품화 계획 없음** — Linear 연동은 개인 개발 워크플로우로만 사용한다(배포용 스킬/MCP 프로비저닝 제품화는 하지 않기로 결정, 2026-07-12). **주의: `CLAUDE.local.md`는 gitignored가 아니라 git에 추적·커밋되어 공개 저장소(`origin/main`)에 공유된다** — 이름과 달리 로컬 전용 파일이 아니므로, 이 파일에 적는 모든 내용(워크스페이스 매핑·개인 경로·운영 방침)은 공개를 전제로 작성한다. "로컬 전용"은 *템플릿에 미러하지 않는다*는 뜻이지 *공유되지 않는다*는 뜻이 아니다.

### §26.1 워크스페이스 매핑

- **Linear 워크스페이스**: `모두의AI` (`linear.app/goos`)
- **Team**: `모두의AI` · **이슈 접두사**: `MOAI-`
- **Project ↔ 리포 매핑** (동시 진행 7개):

| Linear Project | 리포 | 진행중 이슈 |
|---|---|---|
| moai-adk-go | /Users/goos/MoAI/moai-adk-go | MOAI-13/14/15 (Project-Harness Epic, plan) |
| claude.mo.ai.kr | claude.mo.ai.kr | — |
| mo.ai.kr | mo.ai.kr | MOAI-10 결제 오류 (Bug/High) |
| MINK | MINK | MOAI-11 harness:mink-e2e |
| copythat | copythat | — |
| moai-stock | moai-stock | — |
| academy | academy | MOAI-12 feat/design-system |

- **SSOT 문서**: Linear 팀 문서 "SPEC ↔ Linear 브리지 규칙" (`linear.app/goos/document/spec-linear-브리지-규칙-fed04d006656`)

### §26.2 운영 규칙 (2계층 모델)

- **Linear 계층 (제품/기획)**: 무엇을/언제. 아이디어 저수지 + 우선순위 + 크로스 프로젝트 조율.
- **SPEC 계층 (기술 SSOT)**: 정확히 어떻게. 각 리포 `.moai/specs/`. SPEC 저작은 **항상 리포 파일에서만** — Linear로 옮기지 않는다 (git·에이전트 파이프라인·frontmatter 결합).

흐름:
1. 아이디어·버그 → Linear 이슈로 캡처 (`Idea` 라벨 → Triage). 아직 SPEC 아님.
2. 착수 결정 → 해당 리포에서 `/moai plan`으로 SPEC 승격 + frontmatter `linear_issue: MOAI-N`.
3. `/moai run` → `/moai sync`로 구현·클로즈 → Linear 이슈 Done + PR/커밋 링크.

상태 매핑 (느슨한 결합 — 정밀 상태의 진실은 SPEC frontmatter):

| Linear 상태 | SPEC status |
|---|---|
| Backlog / Todo | draft |
| In Progress | in-progress |
| Done | implemented / completed |

### §26.3 오케스트레이터 지침

- 세션에서 Linear 작업 시 `mcp__claude_ai_Linear__*` 도구 사용 (deferred → ToolSearch preload 필요).
- 팀/사이클 생성·rename은 MCP 미지원 → Linear UI 전용.
- `linear_issue` frontmatter 필드는 아직 SPEC 스키마 정식 필드 아님 (도입은 SPEC-HARNESS-MCP-PROVISION-001 소관). 현재는 서술적 참조로만 사용.

### §26.4 아이디어 기록 방법

- **A. 자연어 지시**: "아이디어 Linear에 기록해줘: <제목> — <설명>" → 이 프로젝트(moai-adk-go)의 Linear Project에 `Idea` 라벨 + Backlog(Triage) 이슈 생성. 어느 Project인지는 §26.1 매핑으로 자동 인식.
- **B. 빠른 트리거**: 사용자 메시지가 `아이디어:` 또는 `💡`로 시작하면, 나머지 내용을 이 Project의 Linear 이슈로 **즉시 기록**한다 (Team `모두의AI`, 라벨 `Idea`, 상태 Backlog/Triage). 확인 질문 없이 기록 후 생성된 `MOAI-N` 이슈 URL을 보고. SPEC 승격은 하지 않음(아이디어 저수지 단계 — 착수 결정 시 `/moai plan`으로 승격).
- 전제: 세션에 Linear MCP가 로드/인증돼 있어야 실제 기록됨(미인증 시 `/mcp`로 Linear 인증 후 재시도).
