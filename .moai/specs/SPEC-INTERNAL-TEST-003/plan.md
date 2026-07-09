---
id: SPEC-INTERNAL-TEST-003
title: "Add missing i18n dictionary entries for workflow.agentic_loop.max_iterations"
version: "0.1.0"
status: completed
created: 2026-07-09
updated: 2026-07-09
author: manager-spec
priority: P1
phase: "v3.x target"
module: "internal/web/assets"
lifecycle: spec-anchored
tags: "i18n, web-console, test-fix, debt-cleanup"
tier: S
depends_on: []
related_specs: [SPEC-INTERNAL-TEST-002, SPEC-INTERNAL-ARCH-001]
---

# SPEC-INTERNAL-TEST-003 — Implementation Plan

## §A. Context (invariants run-phase MUST preserve)

run-phase agent가 다음 불변조건을 반드시 보존:

1. **i18n.js 4-locale block 구조**: `window.MOAI_I18N = { en: {...}, ko: {...}, ja: {...}, zh: {...} };` — 4개 블록 순서 고정. 새 키는 각 블록에 평행하게 추가.
2. **위젯-스키마-i18n 3-표면 동기화 계약**: schema `FieldDef.I18nKey` → templ 위젯 data-i18n 방출 → i18n.js 사전 항목. 사전만 신규 추가하며, 다른 두 표면은 이미 올바르게 동작 중 (실패 메시지가 "rendered page"의 키를 감지한 것 자체가 위젯 방출 정상의 증거).
3. **R6 boundary contract**: `TestDataI18nKeysSubsetOfDictionary`는 "렌더된 모든 data-i18n 키가 사전에 존재"를 강제. 새 항목 추가 후에도 위반 없음.
4. **4-locale parity contract**: `TestI18nKeySetParity`는 모든 schema 필드의 `.title`/`.desc` 키가 4 locale 모두에 존재(`>= 4` occurrences)할 것을 강제.
5. **code chip 비번역 계약**: `<code class="field__key">`는 data-i18n 없이 영문 코드 토큰으로 렌더 — 본 SPEC은 data-i18n이 있는 title/desc만 추가하므로 이 계약에 무관하지만, 실수로 key chip까지 번역하는 일은 없어야 한다.

## §B. Known Issues

| ID | Issue | Impact | Mitigation |
|----|-------|--------|-----------|
| KI-1 | 번역 품질의 주관성 | "agentic loop" / "completion-loop" / "pipeline-level"의 번역이 locale마다 자연스럽지 않을 수 있음 | `loop_prevention.max_iterations`의 각 locale 번역을 semantic anchor로 사용 + 반드시 의미 구분 (agentic=pipeline completion vs loop_prevention=diagnostic per-op) |
| KI-2 | 공유 checkout race | 작업 tree에 unrelated uncommitted changes가 다수 존재 (architecture redesign reports, effort/model edits, SUBCOMMAND-RETIRE-001 등). pathspec commit 또는 worktree 격리 필요 | run-phase 진입 전 `git status`로 unrelated changes 확인; i18n.js만 단일 파일 stage (`git add internal/web/assets/i18n.js`) |
| KI-3 | `loop_prevention` vs `agentic_loop` 의미 혼동 | 둘 다 "max_iterations" 필드명을 공유하므로 copy-paste drift 위험 | REQ-I18N-004 semantic distinctness 강제; run-phase에서 4-locale 각각 두 키의 값이 다름을 grep diff로 확인 |

## §C. Pre-flight (before any write)

run-phase 첫 액션 전 수행:

```bash
# 1. 현재 HEAD 확인 (unrelated uncommitted changes 보존 전제)
git status --short
git log --oneline -1

# 2. 실패 재실측 (plan-phase evidence와 동일한 축어 출력이再现되는지)
go test ./internal/web/ -run 'TestDataI18nKeysSubsetOfDictionary|TestI18nKeySetParity' -count=1 -v 2>&1 | tee /tmp/moai-verify/run-entry-baseline.log

# 3. i18n.js의 4-locale block 구조와 삽입 지점 확인
grep -n 'en:\|ko:\|ja:\|zh:\|^}' internal/web/assets/i18n.js | head -10
grep -n '"f.workflow.loop_prevention.max_iterations' internal/web/assets/i18n.js
grep -n '"f.workflow.agentic_loop.max_iterations' internal/web/assets/i18n.js   # expect 0 matches (confirmed absent)
```

## §D. Constraints (carry from spec.md §D)

- CON-001 locale 순서 보존 (en → ko → ja → zh)
- CON-002 JSON 형식 일관성 (`"<key>": "<value>",`)
- CON-003 번역 품질 최소 bar (en humanized baseline + ko/ja/zh semantic distinctness)
- CON-004 회귀 없음 (`TestI18nSegmentKeysRemovedFromWebDictionary`, `TestNoReviewKeys`, `TestSchemaEmptyLabelParity`, 전체 `internal/web/` 패키지 테스트)

## §E. Self-Verification (run-phase 의무)

run-phase 종료 전 다음을 실행하고 축어 출력을 보존:

```bash
# E1. 두 실패 테스트 해소
go test ./internal/web/ -run 'TestDataI18nKeysSubsetOfDictionary|TestI18nKeySetParity' -count=1 -v 2>&1 | tee /tmp/moai-verify/test-003-e1.log
# expected: exit 0, both PASS

# E2. internal/web/ 패키지 전체 회귀
go test ./internal/web/ -count=1 2>&1 | tee /tmp/moai-verify/test-003-e2.log
# expected: exit 0

# E3. whole-repo 선결 조건 (ARCH-001 unblock)
go test ./... 2>&1 | tee /tmp/moai-verify/test-003-e3.log
# expected: exit 0

# E4. 새 항목 4-locale 존재 및 형식 일관성
grep -c '"f.workflow.agentic_loop.max_iterations.title":' internal/web/assets/i18n.js  # expect 4
grep -c '"f.workflow.agentic_loop.max_iterations.desc":' internal/web/assets/i18n.js   # expect 4

# E5. semantic distinctness (agentic_loop vs loop_prevention — copy-paste drift 탐지)
for locale in en ko ja zh; do
  echo "--- $locale ---"
  grep -A0 '"f.workflow.\(agentic_loop\|loop_prevention\)\.max_iterations\.\(title\|desc\)":' internal/web/assets/i18n.js
done
```

`verification-claim-integrity.md §1.1 surface 2` 의거: E1-E5의 출력은 반드시 run-phase에서 실제로 실행·관측한 축어 결과여야 하며, carry-over 또는 가정 금지.

## §F. Milestones (priority-based, no time estimates)

### M1 — i18n.js 4-locale 항목 추가 (P1)

- 각 locale block(en/ko/ja/zh)에 2개 항목씩, 총 8개 추가.
- 삽입 위치 권장: `auto_clear.token_threshold.{title,desc}` 이후, `loop_prevention.*` 블록 이전 (schema `schema_sections.go:187-208`의 top-down 순서와 일관).
- 번역은 `loop_prevention.max_iterations` 각 locale 항목을 semantic anchor로 활용하되, REQ-I18N-004 distinctness 조건 충족.

**완료 기준**: `grep -c '"f.workflow.agentic_loop.max_iterations\.\(title\|desc\)":' internal/web/assets/i18n.js` returns `8` (4 locales × 2 keys).

### M2 — 검증 및 ARCH-001 unblock 확인 (P1)

- §E. Self-Verification의 E1-E5를 모두 실행.
- whole-repo `go test ./...` exit 0 확인 (이것이 `SPEC-INTERNAL-ARCH-001` plan-audit D1 재진입의 전제).
- `internal/web/` 패키지 테스트만 별도 exit 0 확인 (회귀 없음).

**완료 기준**: E1, E2, E3 모두 exit 0; E4 counts = 4/4; E5 grep diff가 4-locale 모두에서 agentic_loop와 loop_prevention 값이 다름을 보임.

## §G. Anti-Patterns (run-phase가 피해야 할 패턴)

| ID | Anti-pattern | Correct approach |
|----|--------------|------------------|
| AP-1 | 두 sibling의 번역이 동일 (copy-paste drift) | REQ-I18N-004 의거해 반드시 의미 구분 — KO/JA/ZH에서 "pipeline-level completion loop" vs "diagnostic per-operation loop"가 구분되는 자연어 표현 사용 |
| AP-2 | 4-locale 중 일부만 추가 (예: en/ko만, ja/zh 누락) | 4-locale 모두에 추가하지 않으면 `TestI18nKeySetParity`가 여전히 FAIL (`>= 4` 조건 위반) |
| AP-3 | `git add -A`로 unrelated uncommitted changes까지 stage | 반드시 `git add internal/web/assets/i18n.js`로 단일 파일만 stage |
| AP-4 | i18n.js 대신 schema나 테스트 코드를 수정하려는 유혹 | 원인은 사전 누락이지 스키마/테스트 결함이 아님; spec.md §B Out of Scope 참조 |
| AP-5 | 사전 항목 추가 후 테스트를 실행하지 않고 "PASS 주장" | `verification-claim-integrity.md` 위반; 반드시 §E의 명령을 실행하고 축어 출력 보존 |
| AP-6 | trailing comma 누락으로 JSON 형식 훼손 | i18n.js는 JS object literal이지만 관례상 모든 항목이 trailing comma를 가짐; 기존 항목 형식 준수 |

## §H. Cross-References

- **spec.md 본 SPEC**: `.moai/specs/SPEC-INTERNAL-TEST-003/spec.md`
- **acceptance.md 본 SPEC**: `.moai/specs/SPEC-INTERNAL-TEST-003/acceptance.md`
- **선행 SPEC**: `.moai/specs/SPEC-INTERNAL-TEST-002/` (계보상 선행; M1/M3 잔여 부채가 본 SPEC)
- **후행 SPEC (unblock 대상)**: `.moai/specs/SPEC-INTERNAL-ARCH-001/` (plan-audit D1-D9 재진입)
- **스키마 소스**: `internal/settings/schema_sections.go:193` (`workflow.agentic_loop.max_iterations` 필드 정의)
- **사전 소스**: `internal/web/assets/i18n.js` (수정 대상)
- **계약 테스트**: `internal/web/i18n_test.go:267` (R6), `internal/web/schema_label_test.go:101` (4-locale parity)
- **Doctrines**: `.claude/rules/moai/core/verification-claim-integrity.md` (evidence-claim integrity), `.claude/rules/moai/quality/boundary-verification.md` (R6)
