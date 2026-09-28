# SPEC-LOCAL-INSTR-RECEPTION-001 — progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: pending-plan-audit
plan_complete_at: 2026-09-28
authored_by: card t1290 (manager-spec), worktree `.claude/worktrees/t1290`, HEAD `514ac7abe`
artifact_set: Tier L — spec.md, plan.md, acceptance.md, design.md, research.md (6 files with progress.md)
note: plan-audit has not run against this SPEC yet; no plan-artifact has been modified since authoring.

## §E.2 Run-phase Evidence

**M1 — the reception gate (2026-09-28 09:37–09:50 KST, caps per the §M1.4-style declaration above; GLM gateway per lead directive 2).** Raw verbatim outputs: `.moai/reports/t1290/m1-probes.md` (retained load-bearing artifact). Fixture token `TANGO9`, force-tracked per plan.md M1 step 2 (commit `6a6b32202` on WT-agents-local-migration).

| Leg | S1 file@HEAD | S2 exit 0 | S3 control (AGENTS.md loaded) | S4 local token | Verdict |
|---|---|---|---|---|---|
| RED (t1290 tree @ c05393bc2, file absent) | n/a — absent by design | ✓ | ✓ (answer describes AGENTS.md body) | `NOT_PRESENT` | **RED observed** (AC-LIR-005(a)) |
| A — materializer-created tree (`moai worktree new t1290probe-a`, base d5df9457c; fixture landed in-tree, `git show HEAD:AGENTS.local.md` exit 0) | ✓ | ✓ | ✓ | **`TANGO9`** | **PASS** (AC-LIR-001) |
| B — `claude -w t1290probe-b` native creation (ping turn `OK`; base 517ec51ba; pre-landing file@HEAD **exit 128** — the D1 geometry observed directly — fixture landed → S1 ✓) | ✓ | ✓ | ✓ | **`TANGO9`** | **PASS** (AC-LIR-002) |
| C — `EnterWorktree` re-entry into the leg-A tree | ✓ | ✓ | ✓ | **`TANGO9`** | **PASS** (AC-LIR-003) |
| D — `moai codex -w t1290probe-d -- exec` (rerun 10:38–10:40 under the operator's re-measure instruction, after the lead transplanted primary `.codex/hooks.json` under operator approval) | ✓ (fixture at the read surface: primary root) | ✓ | ✓ ("AGENTS.md content is present in my instructions") | **`TANGO9`** | **PASS (AC-LIR-004)** — real model turn, 33,049 tokens; trust prompt NOT observed; invocation pinned: the wiring gate reads the launcher-cwd project root, so the working form is from the primary cwd |
| Negative control (git-archive export of `514ac7abe`, file-less by design; one cwd-error duplicate run recorded honestly before it) | absent-by-design (the opposite of S1) | ✓ | ✓ | `NOT_PRESENT` | **PASS** — probe fails for the right reason (AC-LIR-005(b)) |

**Central measurement result:** all three Claude-side entry paths receive the force-tracked `AGENTS.local.md` content (`TANGO9` observed, control present, exit 0, file at HEAD) — the first measured confirmation that with the file tracked and present at a worktree root the `@AGENTS.local.md` import fires in a linked worktree (`AGENTS.md:262`'s negative holds for the untracked geometry only). The negative control prints `NOT_PRESENT` from the file-less geometry — the probe is failable and discriminates.

**Leg D — first pass 측정 불가 (09:38, blocker chain):** (1) the codex launcher's wiring gate refused the launch while the launcher-cwd project root lacked `.codex/hooks.json` (completing it on the primary was then outside lane authority; two refusal attempts recorded verbatim). (2) Behind the gate, then-untested: codex quota and token refresh. **Rerun (10:38, operator re-measure instruction; primary wiring transplanted by the lead under operator approval): PASS** — the exec from the primary cwd ran a real model turn and answered `TANGO9` with the control present; quota NOT blocked at 10:39 (supersedes the t1203 "until 14:37" record as a fresh observation); trust prompt NOT observed; invocation pinned (wiring gate reads the launcher-cwd root — attempt 1 from the lane worktree still refused). No pass was claimed while unprobed; §8's claim is now measured TRUE for the `-w` child. Full verbatim: `.moai/reports/t1290/m1-probes.md` (Leg D rerun section).

**M1 CLOSED: AC-LIR-001~005 all PASS** (RED + three Claude legs green + codex leg green + negative control green) — the M2 [HARD] entry condition is satisfied. M2 begins under the standing kickoff.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §M1.4-style 측정 상한 선언 (실행 전 — 2026-09-28 09:37 KST, 리드 지침 1)

| 축 | 상한 |
|---|---|
| 프로브 턴 | 총 **7회**: RED 1 + leg 4(각 1) + leg B 트리 생성 핑 1 + 부정 대조 1. 재시도 없음(원문 기록) |
| 프로브별 시간 | `timeout 180` (plan.md 레시피 고정, /opt/homebrew/bin/timeout) |
| M1 전체 벽시계 | 30분 |
| 모델 경로 | **GLM 게이트웨이**(리드 지침 2 — Anthropic 쿼터 소진). 수신 측정 대상은 하네스의 지시파일 로딩이라 모델 백엔드 독립 |
| 레시피 이탈 | 레시피의 `--model claude-haiku-4-5-20251001` 핀이 게이트웨이에서 거부되면 핀 없이 재실행 — 두 시도 모두 원문 기록, 실패 파라미터 호출은 턴 수 미소모 |
| 쿼터/자격 불가 leg | 나머지 완료 후 그 leg 만 「측정 불가·사유 명시」 기록 → M2 진입 게이트에서 정지·리드 보고 (리드 지침 3, AC-LIR-006 준수) |

킥오프 승인 근거: 운영자 09-28 전력 완수 지시 + CLAUDE.local.md §31 (리드 메시지로 수령, iter2 PASS 1.00 판정서 .moai/reports/t1290/plan-audit-iter2.md).

**M2 — step ① done, then BLOCKED on the parent's M1 verb (2026-09-28 10:4x KST).**
Step ① re-measured before-value: `git show develop:CLAUDE.local.md | wc -m` → **45,810** (the
plan's expectation at `514ac7abe`, stable at this milestone). Step ② requires `moai migrate
local-instructions` — **the verb does not exist**: `moai migrate --help` lists agency / cg /
home-state / profiles / restore-skill only, and `git grep -c "local-instructions" a6f3861cd --
internal/cli/` exits 1 with no output (checked on the parent branch tip too). The t1259 branch's
merged part (`2812287eb`, an ancestor of this tree) carried plan artifacts only, and its unmerged
tip `a6f3861cd` is docs-only (v0.3.0 scope reduction). This is the dependency manager-spec
recorded at plan time ("this SPEC's M2 needs the parent's M1 migration verb") — the parent's M1
is t1259's run phase, not started (parent status: draft, plan-audit loop). Per the recorded
dependency and scope discipline (the verb is t1259's to build; the plan's method step names the
verb — a manual migration would be an unapproved deviation), **M2 is blocked: blocker reported
to the lead** with the measured facts. M1's closed state and all its evidence stand unchanged.

**M2 resumed after t1259 landed (2026-09-28).** The isolated card branch absorbed
`develop` at `9da000bf2`; the parent migration verb is now present in
`internal/cli/migrate_local_instructions.go`. `./bin/moai migrate local-instructions`
moved the develop-identical 45,810-character `CLAUDE.local.md` and wrote a byte-identical
backup. Commit `0ab480d9d` tracks only `AGENTS.local.md` (37,061 characters) and the
relocated Jev, premise-check, LSEL, and kickoff procedures in `.moai/docs/`.
`AC-IFU-007`, `AC-IFU-024`, `AC-LIR-007`, and the file-relocation portion of
`AC-LIR-008` have committed-tree evidence in `.moai/reports/t1290/m2-verdict.md`.

**M2 reception gate remains open.** The no-tool Claude probe returned the account's
weekly-limit error; the GLM-configured retry timed out after 180 seconds
without an answer. A direct `codex -C <worktree> exec` control answered the root
`AGENTS.md` question correctly and the `AGENTS.local.md` question `ABSENT` with no
tool call. Codex's official discovery rule selects one file per directory, so this
direct entry does not provide the sibling local file. After seeding the
worktree's missing runtime `.codex/hooks.json` from the primary project's
MoAI-only hook commands, `moai codex -- exec` returned `session_drain.sh` and
`git log --all -S` with no tool call: the MoAI launcher path received the
migrated file. The template and root
`AGENTS.md`/`CLAUDE.md` text now state the measured geometry, but the live Claude
post-swap probe required by `AC-LIR-009` has not passed. Do not merge or close this
card on the structural checks alone. Raw outputs and exact baselines are in the
same M2 verdict report.

**Direct Codex follow-up (same worktree).** The root and template `AGENTS.md`
now require a direct `codex -C` session to read a present worktree-root
`AGENTS.local.md` before other project work. A new read-only Codex session
with a generic startup-check prompt ran `cat AGENTS.local.md` (exit 0,
37,061 output characters), then covered all 678 lines through `sed`, and
replied `READY`. This establishes a tool-mediated direct entry path, while
the MoAI launcher test above establishes pre-session injection. Neither
substitutes for the still-unmeasured Claude post-swap leg. A separate
minimal GLM API call returned HTTP 429 (`rate_limit_error`); its sanitized
response is in `.moai/reports/t1290/m2-glm-rate-limit.json`.
