---
id: SPEC-DOCS-SB-REMOVE-001
version: 1.0.0
status: completed
created_at: 2026-04-21
updated_at: 2026-04-21
author: moai-adk-go
priority: high
labels: [docs-site, i18n, simplify, batch, cleanup, consistency, 4-locale]
issue_number: null
depends_on: [SPEC-SKILL-GATE-001]
related_specs: []
---

# SPEC-DOCS-SB-REMOVE-001: docs-site 4개국어에서 `/simplify`·`/batch` 홍보 제거

## HISTORY

- 2026-04-21 v1.0.0: SPEC-SKILL-GATE-001로 두 스킬을 MoAI 워크플로우에서 완전 제거한 뒤에도 공식 문서 사이트(`adk.mo.ai.kr`, 4개 locale)가 여전히 두 스킬을 "핵심 자동 품질 기능"으로 홍보하고 있음을 감사 발견. CLAUDE.local.md §17.3 "4-locale 동시 업데이트" 의무 위반 상태. 본 SPEC은 공식 문서의 허위 홍보를 제거하고 ko 정본 → en/ja/zh 동기화를 달성.

## Background

### 감사 발견

`docs-site/content/` 내 공개 문서가 `/simplify` 와 `/batch` 를 여전히 "핵심 자동 품질 레이어"로 공개 홍보 중:

| Locale | 파일 | 문제 |
|--------|------|------|
| ko | `core-concepts/what-is-moai-adk.md` (line 186,197,230-237) | 표·산문 전부 /simplify + /batch 언급 |
| ko | `core-concepts/auto-quality.md` (130 라인 전체) | 두 스킬 전용 문서 |
| ko | `core-concepts/harness-engineering.md` (line 129) | auto-quality 페이지 링크 |
| ko | `core-concepts/_meta.yaml` | `auto-quality`, `harness-engineering` 네비게이션 엔트리 |
| en | `core-concepts/what-is-moai-adk.md` (line 186,197,230-237) | 동일 패턴 |
| ja | `core-concepts/what-is-moai-adk.md` (line 186,197,230-237) | 동일 패턴 |
| zh | `core-concepts/what-is-moai-adk.md` (line 186,197,230-237) | 동일 패턴 |

### 선결 i18n 기형

현재 ko만 `auto-quality.md`와 `harness-engineering.md`를 보유하고, en/ja/zh는 이 두 파일이 **부재**(`ls: No such file or directory`). CLAUDE.local.md §17.3("canonical source는 ko ... 4개 locale 해당 파일 모두 수정") 기준으로 이미 i18n 불일치 상태. 본 SPEC은 두 ko-only 파일을 삭제하여 4-locale 일관성을 회복하는 방향으로 정리 — en/ja/zh로 번역 배포하는 대안보다 스킬 자체를 제거한 현 상태와 정합적.

### 비-관련 `batch` 언급 (건드리지 말 것)

- `docs-site/content/*/advanced/pencil-guide.md`: Pencil MCP의 `batch_design`, `batch_get` tool 이름 — Claude Code `/batch`와 무관
- `docs-site/content/en/utility-commands/moai-mx.md`: mx scan 내부의 "batch processing" 표현 — 일반 명사
- `docs-site/content/en/utility-commands/moai-github.md`: `--all` 플래그의 "batch mode" 표현 — 일반 명사

본 SPEC은 Claude Code `/batch`·`/simplify` 스킬 **홍보** 제거에만 한정된다.

## Requirements (EARS)

### R1 — ko 정본 정리 (4 파일)

- **REQ-DOCS-KO-001** (Ubiquitous): `docs-site/content/ko/core-concepts/what-is-moai-adk.md`에 `/simplify` 및 `/batch` 리터럴 0건.
- **REQ-DOCS-KO-002** (Ubiquitous): `docs-site/content/ko/core-concepts/auto-quality.md` 파일 부재(삭제).
- **REQ-DOCS-KO-003** (Ubiquitous): `docs-site/content/ko/core-concepts/harness-engineering.md`에 `/simplify`·`/batch`·`auto-quality` 참조 0건.
- **REQ-DOCS-KO-004** (Ubiquitous): `docs-site/content/ko/core-concepts/_meta.yaml`에서 `auto-quality` 엔트리 제거 및 `harness-engineering` 엔트리는 en/ja/zh 일관성 평가 후 처리(en/ja/zh에 harness-engineering 부재이므로 ko에서도 일관을 위해 제거 혹은 파일 보존 판단).

### R2 — en 동기화

- **REQ-DOCS-EN-001** (Ubiquitous): `docs-site/content/en/core-concepts/what-is-moai-adk.md`에 `/simplify` 및 `/batch` 리터럴 0건.

### R3 — ja 동기화

- **REQ-DOCS-JA-001** (Ubiquitous): `docs-site/content/ja/core-concepts/what-is-moai-adk.md`에 `/simplify` 및 `/batch` 리터럴 0건.

### R4 — zh 동기화

- **REQ-DOCS-ZH-001** (Ubiquitous): `docs-site/content/zh/core-concepts/what-is-moai-adk.md`에 `/simplify` 및 `/batch` 리터럴 0건.

### R5 — 빌드 회귀 방어

- **REQ-DOCS-BUILD-001** (Ubiquitous): `cd docs-site && hugo --minify` 명령이 성공한다 (exit 0).
- **REQ-DOCS-BUILD-002** (Ubiquitous): Hugo 빌드 출력에 깨진 링크(broken reference) 경고 0건.

## Acceptance Criteria

- **AC-1**: `grep -rn '/simplify\|/batch' docs-site/content/{ko,en,ja,zh}/core-concepts/` 결과 0 라인.
- **AC-2**: `docs-site/content/ko/core-concepts/auto-quality.md` 파일 부재(`test ! -f`).
- **AC-3**: `grep -rn 'auto-quality' docs-site/content/ko/` 결과 0 라인 (링크·메타 참조 전수 제거).
- **AC-4**: `cd docs-site && hugo --minify` exit 0, stderr에 "broken" 또는 "ref not found" 워닝 0건.
- **AC-5**: 4개 locale의 `core-concepts/what-is-moai-adk.md`에서 TDD RED-GREEN-REFACTOR 및 DDD ANALYZE-PRESERVE-IMPROVE 사이클 설명은 **보존**(스킬 참조 제거 후에도 핵심 방법론 설명은 유지).
- **AC-6**: `_meta.yaml` 4개 locale 구조 일관성(`auto-quality` 엔트리 ko에서 제거, en/ja/zh는 이미 부재).
- **AC-7**: Pencil MCP의 `batch_design`·`batch_get` 및 moai-mx/moai-github의 "batch mode" 일반 명사 언급은 **건드리지 않음** — `grep -rn 'batch_design\|batch_get' docs-site/content/` 결과 pre/post 동일.

## Exclusions

- Pencil MCP 관련 `batch_design`, `batch_get` tool 이름 변경 없음.
- `utility-commands/moai-mx.md`, `moai-github.md`의 "batch" 일반 명사 언급 변경 없음.
- `adk.mo.ai.kr` 도메인·Vercel 설정 변경 없음.
- 4개 locale 외 다른 언어 추가 없음.

## Target Files

**수정**:

1. `docs-site/content/ko/core-concepts/what-is-moai-adk.md`
2. `docs-site/content/ko/core-concepts/harness-engineering.md`
3. `docs-site/content/ko/core-concepts/_meta.yaml`
4. `docs-site/content/en/core-concepts/what-is-moai-adk.md`
5. `docs-site/content/ja/core-concepts/what-is-moai-adk.md`
6. `docs-site/content/zh/core-concepts/what-is-moai-adk.md`

**삭제**:

7. `docs-site/content/ko/core-concepts/auto-quality.md`

## Risks

- **R-1 (Hextra 빌드 깨짐)**: auto-quality.md 삭제 후 _meta.yaml에서 참조가 남으면 빌드 실패. **완화**: _meta.yaml에서 엔트리 제거 + hugo build 검증.
- **R-2 (번역 품질 불일치)**: en/ja/zh에서 같은 문장을 제거할 때 근접 단락이 어색해질 수 있음. **완화**: 제거 후 각 locale의 REFACTOR/IMPROVE 단락이 독립적으로 읽히는지 수동 검토.
- **R-3 (harness-engineering.md 처리)**: ko-only 파일이므로 본 SPEC은 **내부 링크만 정리**(파일 자체는 보존). en/ja/zh에 번역 배포는 별도 SPEC 대상.

## Implementation Notes

v1.0.0 단일 세션 구현:
1. ko/en/ja/zh what-is-moai-adk.md의 3개 구간(line ~186, ~197, ~230-237) 동시 편집
2. ko auto-quality.md 파일 삭제
3. ko harness-engineering.md 내 auto-quality 링크 제거
4. ko _meta.yaml에서 `auto-quality` 엔트리 제거
5. `cd docs-site && hugo --minify` 검증
6. AC-1 ~ AC-7 grep 전수 통과 확인
7. 단일 `docs(site)` 커밋
