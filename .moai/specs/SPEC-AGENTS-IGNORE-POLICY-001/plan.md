---
id: SPEC-AGENTS-IGNORE-POLICY-001
title: "plan — .agents .gitignore policy ruling"
version: "0.1.0"
created: 2026-09-14
updated: 2026-09-14
author: manager-spec
tier: M
---

# SPEC-AGENTS-IGNORE-POLICY-001 — Implementation Plan

## §A Context

Card t738 (Class C, policy adjudication). The dev repo's root `.gitignore` (lines 133-170) and the distributed template's `.gitignore` (lines 193-226) carry mutually exclusive `.agents` ignore policies. The ruling (spec.md §D.4) names the template policy canonical. This plan lays out the remediation options with blast radius, the operator-gate question, and the deferred application path.

## §B Known Issues

- The dev-repo whitelist chain (L133-170) hides every `.agents` artifact except 16 published SKILL.md files — measured via `git check-ignore -v` on 2026-09-14 (evidence in spec.md §D.1).
- The template policy's own comment block (L193-203) documents the intended semantics; the dev repo never adopted them.
- Card t498's mirror-absence investigation was muddied by the whitelist chain: ignored local state is indistinguishable from absent state in `git status` and grep.
- t701 measured the root `.agents/*` pattern is root-anchored and never covered the template subtree (card-supplied background; consistent with the anchoring semantics of patterns containing a non-trailing `/`, not independently re-derived here).

## §C Pre-flight

- [x] Both `.gitignore` files read side by side (2026-09-14, HEAD `99e02ac52`).
- [x] `git check-ignore -v` probes run on both trees; outputs recorded verbatim (spec.md §D.1/§D.2).
- [x] Present-state measurement: `.agents/skills/` holds exactly the 16 tracked published dirs; `git status --porcelain .agents/` empty; no ignored-but-present content.
- [x] SPEC ID `SPEC-AGENTS-IGNORE-POLICY-001` regex PASS; no collision with existing SPEC ids (`SPEC-AGENTS-MD-CANON-001` etc. are distinct).
- [x] Operator adjudication recorded: **Option A chosen** — the template policy is confirmed canonical (default-allow, ignore only the regenerated `moai*` mirrors). Decider: operator; channel: the lead's AskUserQuestion round, 2026-09-14. The dev-repo root `.gitignore` alignment work is split into a follow-up card (draft text in §H) and executes only after cards t498 and t510 close; THIS card applies nothing.

## §D Constraints

- NO `.gitignore` rule may be edited by this SPEC (card HARD constraint 2; REQ-POL-004).
- Application timing belongs to t498/t510; this SPEC delivers the ruling + application plan only.
- If the operator selects a template-side option (all-user impact), REQ-POL-005's migration obligation binds the follow-up card.
- No Go code changes; no publisher logic changes.

## §E Self-Verification

Plan-phase verification performed during authoring (all commands run in this worktree, HEAD `99e02ac52`):

| # | Command | Observed result |
|---|---------|-----------------|
| 1 | `git check-ignore -v .agents/skills/moai-clean/SKILL.md .agents/skills/moai-clean/manifest.json .agents/skills/moai-workflow-tdd/SKILL.md .agents/skills/my-custom/SKILL.md .agents/skills/moai-clean/extra/SKILL.md .agents/notes.md` | 5 lines of ignore hits; `moai-clean/SKILL.md` absent from output (= tracked). Verbatim output in spec.md §D.1. |
| 2 | `git -C internal/template/templates check-ignore -v ...` (same path set + `moai-todo/manifest.json`) | single hit on `moai-workflow-tdd/SKILL.md` via L204. Verbatim output in spec.md §D.2. |
| 3 | `git ls-files .agents \| wc -l` | `16` |
| 4 | `git status --porcelain .agents/` | empty (clean) |
| 5 | `ls .agents/skills/` | exactly the 16 `moai-<command>` directories |

Gaps: the t701 "removed 171-174 re-include was a no-op" archaeology was NOT re-derived (card-supplied background); no behavioral test of `moai init` output was run (template-side, out of scope for the ruling).

## §F Remediation Options — blast radius (operator-gate material)

### Option A — align dev repo to template policy (RECOMMENDED)

- **What changes** (dev-side only, in the follow-up card): replace root `.gitignore` lines 133-170 with the template-equivalent rule set — `.agents/skills/moai*` (ignore the regenerated mirror) + the 16 `!.agents/skills/moai-<command>/` re-includes (full directory contents, no content re-exclusion). Net effect identical to template lines 204-226.
- **Blast radius**: dev repo only. Zero user impact — the template is untouched. Verified: no ignored-but-present `.agents` content exists today (§E row 4/5), so the swap surfaces no untracked noise. Nothing currently tracked becomes untracked.
- **Behavioral gain**: future files inside the 16 published dirs become tracked (dev policy ignored them via L154); user-authored `.agents` entries become visible; dogfooding of user-side `.agents` state becomes possible.
- **Migration**: none.

### Option B — template adopts the dev whitelist policy (all-user impact)

- **What changes**: `internal/template/templates/.gitignore` lines 193-226 replaced with a default-deny whitelist; every user project receives it on the next `moai update` (`.gitignore` is template-managed).
- **Blast radius**: all user projects. User-authored entries under `.agents/` that are not yet tracked become silently ignored on the next update; tracked files stay tracked but every new file under `.agents/` becomes invisible. Directly contradicts the template's own documented intent (L202-203). Sidecars inside published dirs become invisible.
- **Migration required**: CHANGELOG entry + release-note warning; users with untracked user-authored `.agents` content must add local exceptions or commit before updating. This is the option whose conclusion triggers the operator gate per the card's HARD constraint 3.
- **Recommendation**: rejected — it manufactures a regression for every user to preserve a dev-repo convenience that itself caused the dogfooding defect.

### Option C — keep both policies, document the divergence

- **What changes**: documentation only; the divergence persists.
- **Blast radius**: none immediately; both failure axes stay open, and the next `.agents` investigation repeats t498's muddied-evidence experience.
- **Recommendation**: rejected — the card exists precisely because an unruled divergence is the defect.

## §G Milestones (priority-ordered; no time estimates)

- **M1 (High) — Record the ruling**: spec.md §D carries the ruling with both policies, line citations, and probe evidence. DONE at plan-phase (this artifact set).
- **M2 (High) — Operator adjudication**: DONE 2026-09-14. Operator chose Option A via the lead's question round; decision recorded in progress.md §E.1 and the follow-up card drafted (§H).
- **M3 (Medium) — Application plan hand-off**: on operator approval of Option A, register the follow-up card that applies the dev-side alignment AFTER t498 and t510 close; the follow-up card owns the `.gitignore` edit, `make build` (embed check), and the post-change `git check-ignore` regression probes.

## §H Operator Gate — question text (Korean, operator-readable)

> **질문**: `.agents` gitignore 정책 판정을 승인합니까? 판정 내용 — 배포 템플릿 쪽 정책(기본 허용, 재생성되는 `moai*` 미러만 무시, `templates/.gitignore` 193-226행)이 정책으로 맞고, 개발 저장소 루트 `.gitignore`의 화이트리스트 체인(133-170행)이 정합화 대상입니다.
>
> - **옵션 A (권장)**: 템플릿 정책 채택 확정. 개발 저장소 정합화는 후속 카드로, t498/t510 종결 이후 적용. 사용자 영향 0, 현재 저장소에 무시되고 있는 `.agents` 내용물이 없어 즉시 노이즈도 없음 (실측).
> - **옵션 B**: 반대로 템플릿을 화이트리스트 정책으로 변경. `moai update`로 전 사용자 프로젝트에 배포되며, 사용자가 만든 `.agents` 항목이 조용히 무시됨. 마이그레이션 고지 필요 — 전 사용자 영향.
> - **옵션 C**: 현상 유지 + 분기 문서화만. 양쪽 실패 축이 열린 채로 남음.
> - **보류**: 판정에 필요한 추가 조사를 지시.

### §H.1 후속 카드 텍스트 (안) — 옵션 A 적용 전용 카드 (리드가 운영자 승인 후 발행)

> **카드 제목**: `.agents gitignore 정합화 — 개발 저장소를 템플릿 정책으로 (적용 전용)`
>
> **내용**: 루트 `.gitignore` 133-170행의 화이트리스트 체인(`.agents/*` → `!.agents/skills/` → `.agents/skills/*` → 16개 `!dir` → `.agents/skills/*/*` → 16개 `!SKILL.md`)을 제거하고, 템플릿 정책과 동일한 규칙으로 교체한다 — `.agents/skills/moai*`(재생성 미러만 무시) + 16개 `!.agents/skills/moai-<command>/` 재포함(디렉터리 전체 내용 추적, 내용 재제외 없음), `.agents/` 루트 자체는 추적 영역으로 둔다.
>
> **판정 근거**: SPEC-AGENTS-IGNORE-POLICY-001 (운영자 옵션 A 확정, 2026-09-14). 이 카드는 판정을 다루지 않는다 — 적용만 한다.
>
> **전제 (게이트)**: t498, t510 이 모두 종결된 뒤에만 착수한다. 그 전에는 어떤 `.gitignore` 수정도 금지다 — 두 카드가 현재의 `.agents` 부재 상태를 관측 대상으로 사용한다.
>
> **범위**: 루트 `.gitignore` 한 파일. 템플릿 측(`internal/template/templates/.gitignore`) 변경 없음 — 사용자 영향 0. Go 코드·퍼블리셔 로직 변경 없음.
>
> **검증**: `make build`(embed 재생성) 후 `git check-ignore -v` 프로브로 spec.md §D.3 표의 기대값을 재현한다 — (1) 16개 SKILL.md 추적 유지, (2) `moai-clean/manifest.json` 류 사이드카 추적 전환, (3) `my-custom/`·루트 `notes.md` 추적 전환, (4) `moai-workflow-*` 미러는 계속 무시. `git ls-files .agents | wc -l`는 16 유지.

## §I Anti-Patterns

- Un-ignoring `.agents` files now to "make them appear" — forbidden by card constraint 2; t498/t510 use the absence as an observation target.
- Treating the dev-repo whitelist as intentional policy just because it predates the ruling — it has no decision record, and its observed semantics (spec.md §D.1) hide user-authored and sidecar artifacts.
- Editing the template "while we're in the file" under Option A — Option A is dev-side only by construction.

## §J Cross-References

- spec.md §D (ruling + evidence), acceptance.md (AC matrix)
- Cards: t738, t498/t510 (application timing), t701 (anchoring background)
