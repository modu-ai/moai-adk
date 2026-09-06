---
id: SPEC-PLAN-AUDITOR-RESIDUE-001
title: "plan-auditor 조항 잔여 — t387 곁말 규약 반영 + t386 반출 조항 plan-auditor 측 적용 (카드 t450)"
version: "1.0.0"
status: completed
created: 2026-09-03
updated: 2026-09-03
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: ".claude/agents/moai/plan-auditor.md"
lifecycle: spec-anchored
tier: M
tags: "plan-auditor, audit-artifact-convention, export-mandate, side-talk, template-mirror, neutrality"
---

# SPEC-PLAN-AUDITOR-RESIDUE-001 — plan-auditor 조항 잔여 (카드 t450)

## HISTORY

| 날짜 | 내용 |
|------|------|
| 2026-09-03 | 최초 작성 (manager-spec, plan-phase, 카드 t450) |

## WHY (배경)

t387(f47d7f5a9)은 감사 판정문에 붙는 곁말(side-talk)의 형태를 규약 문서(`audit-artifact-convention.md` § Side-talk)로 성문화했다. 당시 에이전트 정의 반영은 lane-9 카드(t367 루브릭 정비, t302 판정 소유권)와 같은 파일 충돌을 피하려고 보류됐다 — t387 판정서 Gaps 절과 f47d7f5a9 커밋 본문이 이 보류를 명시한다.

t386 반출 조항(export mandate)은 sync-auditor 쪽만 4244c4a06에서 먼저 착지했고, 그 커밋 본문은 "plan-auditor counterpart stays deferred until t367 closes"라고 못박았다.

전제가 이제 충족됐다: t367(18fc2c9ef, 머지 a3f4bc617)은 origin/develop 7835148d3에 이미 착지해 있다(이번 실행 merge-base로 측정). 같은 파일의 루브릭 영역(:72 부근)은 t367이 고쳤으므로 이 SPEC은 그 위에 쌓는다 — 되돌리지 않는다.

**현재 상태의 결함 (이번 실행 직접 측정, 트리 7835148d3):**

1. `plan-auditor.md` § Output Format(양쪽 쌍둥이 모두)이 감사 보고서 쓰기 위치를 `.moai/reports/plan-audit/{SPEC-ID}-review-{iteration}.md`로 명령하는데, 이 디렉터리는 의도적으로 gitignored다(`.gitignore:230` `.moai/reports/plan-audit/*.md`, `git check-ignore` exit=0 관측). 규약 문서는 그 디렉터리를 FORBIDDEN으로 선언하고 카드 범위 패밀리 `.moai/reports/<card-id>/plan-audit.md` / `plan-audit-iter<N>.md`를 정본으로 정했다. 즉 plan-auditor는 지금 규약이 금지한 위치에 쓰도록 지시받고 있다 — 쓰면 반출이 아니라 폐기다.
2. 규약 문서 § What makes the convention stick는 "The plan-auditor and sync-auditor agent definitions carry the export as a HARD completion condition"이라고 서술하는데, plan-auditor 쪽으로는 현재 거짓이다(sync-auditor만 4244c4a06에서 반영됨). 규약의 교차참조 절도 `plan-auditor.md`를 반출 조항 절로 명시한다.
3. 곁말 규약은 어느 감사자 정의에도 반영돼 있지 않다(양쪽 plan-auditor·sync-auditor에 `Side-talk` 0매치 관측). t450은 plan-auditor 쪽만 반영한다 — 카드 범위.

## WHAT (범위)

정확히 세 축만 고친다(카드 [HARD] 범위 — 확장 금지):

1. **plan-auditor.md 쌍둥이(템플릿 + 로컬)** — `.claude/agents/moai/plan-auditor.md`(로컬 도그푸드)와 `internal/template/templates/.claude/agents/moai/plan-auditor.md`(템플릿 미러)를 **함께** 편집한다. codex 방출물(`internal/template/templates/.codex/agents/moai/plan-auditor.toml`)은 템플릿 쪽에서 `make agents-emit`으로 기계 생성한다(C3, CLAUDE.local.md §2.0 — 손편집 금지).
2. **규약 문서 크로스레퍼런스** — 적용하는 조항을 참조하는 규약/룰 문서의 교차참조를 일관되게 갱신한다. 측정된 대상: `spec-workflow.md` § Report Persistence(양쪽 미러, "Two report streams coexist deliberately in `.moai/reports/plan-audit/`" — 금지 디렉터리를 의도된 보관처로 서술)와 `audit-artifact-convention.md` § What makes the convention stick / § Cross-references(양쪽 미러).
3. **중립성 0매치** — 템플릿 쪽 편집은 템플릿 중립성 금지 클래스(C1-C8, `.moai/docs/template-internal-isolation-doctrine.md` §25.1; CI 가드 `.github/workflows/template-neutrality-check.yaml`)에 0매치여야 한다. 형제 가드 `internal_content_leak_test.go`(C3 날짜 + C7 커밋 해시)도 함께 녹색이어야 한다.

## REQUIREMENTS

**REQ-001**: `plan-auditor.md`의 양쪽 쌍둥이 사본은 감사 판정문의 반출을 완료 조건으로 요구하는 조항을 품어야 한다 — 판정문은 렌더된 같은 턴에 규약이 정한 카드 범위 패밀리 위치(`.moai/reports/<card-id>/plan-audit.md` 또는 `plan-audit-iter<N>.md`; SPEC 범위 감사는 `.moai/reports/<SPEC-ID>/`)에 파일로 기록되며, 반출 파일 없는 감사 응답은 불완전 감사다.

**REQ-002**: `plan-auditor.md`의 양쪽 쌍둥이 사본은 감사 보고서의 쓰기 대상으로 gitignored 위치 `.moai/reports/plan-audit/`를 더 이상 명령해서는 안 된다 — 기존 § Output Format의 경로 지시는 REQ-001의 반출 조항과 일치하도록 갱신된다.

**REQ-003**: `plan-auditor.md`의 양쪽 쌍둥이 사본은 곁말 규약(`audit-artifact-convention.md` § Side-talk)을 반영해야 한다 — 판정에 덧붙는 조언은 미검증으로 제목 붙은 별도 절에 두고, 결론이 아니라 측정 지시 형태로 쓰며, 각 조언 줄에 상태 라벨(`measured` / `inferred` / `assumption`)을 단다.

**REQ-004**: REQ-001 내지 REQ-003의 조항 문면은 로컬 사본과 템플릿 사본에서 서로 동일해야 한다 — 조항 자체의 쌍둥이 일치는 이 SPEC의 [HARD] 축이다. (기존에 존재하던 두 곳의 쌍둥이 드리프트 — D7-1 예시 식별자와 Tier-resolved ceiling 문단 — 는 이 SPEC이 만든 것이 아니므로 수리 대상이 아니다. 기록만 한다.)

**REQ-005**: 규약·룰 문서의 교차참조는 적용된 조항과 일치해야 한다 — `spec-workflow.md` § Report Persistence와 `audit-artifact-convention.md`의 감사자 측 서술이 plan-auditor의 실제 반출 경로·조항을 정확히 가리키도록 갱신된다(양쪽 템플릿 미러 포함).

**REQ-006**: 템플릿 쪽 편집 결과는 템플릿 중립성 금지 클래스(C1-C8)에 0매치여야 하며, codex 방출물은 템플릿 쪽 `plan-auditor.md`에서 재생성돼야 한다.

**REQ-007**: **When** t367이 착지시킨 :72 부근 루브릭 문면(Event-detected 제5패턴 불릿, Unwanted는 legacy-only 불릿)이 편집 과정에서 손상될 것으로 관측되면, 편집은 그 문면을 원문 그대로 보존하도록 수정돼야 한다 — 이 SPEC은 t367의 착지 내용 위에 쌓이는 것이지 되돌리는 것이 아니다.

**REQ-008**: **When** `make agents-emit` 전체 실행이 무관한 t443 소관 드리프트(sync-auditor.toml sha256 불일치 — 이번 실행에서 `go test ./internal/template/agentemit/...` FAIL로 재측정됨)로 막혀 있을 때, run-phase는 t367 선례(커밋 2549f775f)를 따라 plan-auditor 항목만 범위 재생성하고 sync-auditor.toml은 develop 값으로 복원한 뒤, 그 드리프트를 t443 소관 미수리로 기록해야 한다 — t461·t367 선례의 record-and-not-repair다.

## ACCEPTANCE CRITERIA (요약 — 상세는 acceptance.md)

- AC-001 반출 조항 존재(양쪽 사본) — release-blocking, RED-now 있음
- AC-002 금지 경로 제거(양쪽 사본) — release-blocking, RED-now 있음
- AC-003 곁말 규약 반영(양쪽 사본) — release-blocking, RED-now 있음
- AC-004 조항 쌍둥이 일치 — release-blocking
- AC-005 교차참조 일치 — release-blocking, RED-now 있음
- AC-006 방출물+카탈로그 해시 재생성 — release-blocking, RED-now 있음
- AC-007 t367 :72 문면 보존 — 회귀 가드 (preservation)
- AC-008 템플릿 중립성 0매치 — 회귀 가드 (preservation)

## OUT OF SCOPE

### Out of Scope — sync-auditor 쪽 잔여

- sync-auditor.md에 곁말 규약 반영은 t387 Gaps가 함께 보류로 적은 항목이지만, 카드 t450의 범위는 plan-auditor 축 3가지로 한정된다. sync-auditor 곁말 반영은 별도 카드 소관으로 남는다.

### Out of Scope — 기존 쌍둥이 드리프트 수리

- 로컬 `plan-auditor.md`와 템플릿 미러 사이에 이미 존재하는 두 hunk(D7-1 예시 식별자 `SPEC-DOMAIN-WO-001` vs 중립화된 `SPEC-EXAMPLE-DOMAIN-001`; Tier-resolved ceiling 문단이 로컬에만 존재)는 이 SPEC이 만든 드리프트가 아니므로 수리하지 않는다 — 기록만 하고, drive-by 수리는 하지 않는다.

### Out of Scope — 루브릭·판정 로직 변경

- 감사 루브릭 점수대, must-pass 목록(MP-1~8), 판정 절차의 변경은 하지 않는다. t367이 정비한 GEARS 루브릭은 그대로다.
