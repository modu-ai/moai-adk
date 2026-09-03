---
id: SPEC-CODEX-SKILL-PATH-001
title: "Codex stale-skill checker path-shape resolution — home-relative expansion, relative/odd-form classification, and an isolated reproduction fixture"
version: "0.1.0"
status: completed
created: 2026-09-03
updated: 2026-09-03
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: internal/cli
lifecycle: spec-anchored
era: V3R6
tier: S
tags: "codex, doctor, stale-skill, path-shape, home-expansion, false-positive, isolation-fixture"
related_specs: [SPEC-CODEX-WIRING-001]
---

# SPEC-CODEX-SKILL-PATH-001 — Codex stale-skill checker path-shape resolution

## HISTORY

- 2026-09-03 (plan-phase, v0.1.0) 최초 작성. 카드 t468 (Class B defect-adjacent, Tier S). lane-12가 t451의 후속 후보로 올리고 t451 범위에 흡수하지 않은 잔여 위험 3종. 본 SPEC 작성 중 실측으로 확인된 정정 1건: 카드가 "error taxonomy 미분류"로 서술한 t451 상속 제약 2항(`os.Stat` 오류 접기 금지)은 **이미 구현돼 있다**(`internal/cli/doctor_codex.go` `codexStaleSkillFinding` — `errors.Is(serr, fs.ErrNotExist)`만 missing으로, 나머지는 `indeterminate`로 분리). 따라서 해당 항목은 신규 요구가 아니라 PRESERVE 요구(REQ-CSP-005)로 반영한다. 실제 잔여 결함은 경로 형태 처리 3종(§B.2)이다.

## §A. User Story

**As a** MoAI-ADK user whose Codex `config.toml` registers skills with `~`-relative, relative, or Windows-shaped paths, **I want** `moai doctor`'s stale-skill check to either resolve my path correctly or admit it cannot check it, **so that** the doctor never tells me to delete a registration that is actually valid.

**결과 가설:**
- 어떤 경로 형태의 등록도 "존재하는데 stale로 보고"되지 않는다 — missing 판정은 결정론적으로 해석 가능한 경로의 실제 부재에서만 나온다.
- 확인 불가능한 경로 형태는 별도 분류로 보고되고, "remove the stale entries" 지시는 절대 그 분류에 붙지 않는다.
- 개발 머신에서 관측되는 거짓 지적은 0이다(잠재 결함) — 재현은 오직 격리 fixture를 통해서만 이루어진다(REQ-CSP-007).

## §B. Context and Background

### §B.1 결함 표면 (실측, worktree HEAD `d592b0551`)

| 항목 | 위치 |
|---|---|
| 판정 함수 | `internal/cli/doctor_codex.go` `codexStaleSkillFinding()` — `[[skills.config]]` 항목의 `e.Path`를 원문 그대로 `os.Stat`(line ~338) |
| 파서 | `internal/codexwiring/skills.go` `ParseSkillEntries()` — `SkillEntry.Path`는 TOML 기본 문자열의 원문 토큰(이스케이프 복호화 없음) |
| 홈 해석 | `internal/cli/mcp_codex.go` `resolveCodexHomeDir()` + seam `codexUserHomeDir = os.UserHomeDir` (t451) |
| 테스트 기반 | `internal/cli/doctor_codex_test.go` — `stubCodexHome`(CODEX_HOME + seam 동시 고정), `writeCodexHomeConfig`(합성 codex home fixture), `codexSkillEntrySpec` |

`SkillEntry.Path`의 주석은 "absolute, as Codex writes it"이라 단언하지만, 사용자가 손으로 편집한 config에는 이 전제가 성립하지 않는다 — 그 전제 위반 3종이 본 카드다.

### §B.2 잔여 위험 3종 (각각 실측 근거 포함, tree `d592b0551`)

1. **`~/` 접두 경로 미전개** — `os.Stat("~/...")`는 셸이 아니므로 `~`를 전개하지 않고 항상 ENOENT. 실측: `os.Stat("~/definitely-not-a-real-path-t468")` → `no such file or directory`, `errors.Is(err, fs.ErrNotExist) == true`. 따라서 존재하는 스킬의 `~/`-등록도 missing으로 집계되고, finding의 "remove the stale entries" 지시가 **유효한 등록의 삭제를 권고**한다.
2. **상대 경로 무기준** — 상대 경로는 어떤 기준(cwd / codex home / 기타)으로 해석되는지 이 리포지토리에서 관측되지 않았다. 임의의 기준을 추측해 resolve하는 것은 미관측 전제를 사실로 만드는 것(`verification-claim-integrity.md` §1)이므로 금지하고, 분류로만 보고한다(REQ-CSP-003).
3. **이스케이프/기괴 형태 경로** — 백슬래시 포함(Windows 형태 또는 미복호화 TOML 이스케이프 잔존) 경로. 실측: darwin에서 `os.Stat("C:\\Users\\goos\\SKILL.md")` → ENOENT, `isNotExist=true`. 비-Windows 호스트에서 무조건 missing으로 떨어진다.

### §B.3 확인된 사실 (리드 제공, 본 SPEC이 재검증하지 않는 것)

- 이 개발 머신의 현재 거짓 지적 수: **0** — 실제 등록 항목 전부 절대 경로. 결함은 **잠재적**이며, 다른 경로 형태를 쓰는 사용자에게만 발현한다. 그러므로 본 카드는 재현 fixture를 요구사항으로 명시한다(REQ-CSP-007). fixture 없이 관측 가능한 실패를 가정하는 plan/AC는 공허하다.
- 기존 검사의 초록 baseline (실측, tree `d592b0551`): `go test ./internal/cli/ -run 'TestCheckCodexWiring_StaleHomeSkillsReported|TestCheckCodexWiring_HealthyHomeSkillsNoFinding' -count=1 -v` → both PASS.

### §B.4 계보와 범위 경계

- **t451 lineage**: `resolveCodexHomeDir` 경유 `CODEX_HOME` 존중, 단일 홈 해석 경로, 읽기 전용/fail-open 검사, stat 오류 분류(taxonomy) — 모두 본 SPEC이 PRESERVE한다(REQ-CSP-005/006).
- **t452 (lane-12, codex skill wiring)** — 스킬 배선 표면. 본 SPEC은 건드리지 않는다.
- **t462 (lane-6, codex e2e exhaustive sweep)** — e2e 소진 스윕. 본 SPEC은 격리 단위/검사 테스트에 머문다.
- `moai doctor`의 "Codex Wiring" 검사 자체(SPEC-CODEX-WIRING-001 REQ-CW-010)는 배경이지 대상이 아니다 — 본 SPEC은 그 안의 stale-skill 서브검사의 경로 해석만 다룬다.

## §C. Requirements (GEARS)

- **REQ-CSP-001 (ubiquitous, classification gate)** — The stale-skill checker shall classify every declared `[[skills.config]]` path by shape (absolute / home-relative / relative / oddly-formed) before any filesystem check, and only paths whose shape is determinately resolvable in this environment (absolute, and home-relative after expansion) shall be candidates for the missing-path verdict.
- **REQ-CSP-002 (event-driven, home-relative expansion)** — **When** a declared path is exactly `~` or `~/`-prefixed, the checker shall expand it against the user home directory resolved through the existing `codexUserHomeDir` seam (`os.UserHomeDir`) — the USER home, distinct from the codex home — and stat the expanded path; an expanded path that exists shall not be counted missing. A `~user`-prefixed path (another user's home) is not expandable by this rule and shall fall under REQ-CSP-004's oddly-formed classification.
- **REQ-CSP-003 (unwanted, relative paths)** — The checker shall not count a relative path toward the missing/stale verdict and shall not resolve it against any guessed base; relative entries shall be reported as a distinct declared classification with a non-destructive description (the resolution base is unobserved in this repository and must not be asserted).
- **REQ-CSP-004 (unwanted, oddly-formed paths)** — The checker shall not count an oddly-formed path toward the missing/stale verdict; such entries shall be reported as a distinct declared classification. Oddly-formed means: a path that is NOT absolute per `filepath.IsAbs` (evaluated first, so a Windows host classifies its own native `C:\...` absolutes as absolute, never as odd) AND either contains a backslash (a Windows-shaped separator in a non-absolute position, or a residual un-decoded TOML escape) or carries another user's `~user` home form.
- **REQ-CSP-005 (unwanted, taxonomy preserved from t451)** — The checker shall feed the missing/stale verdict exclusively from `os.IsNotExist`-class failures; permission denied, symlink loops, and every other stat error shall remain a separate indeterminate classification surfaced in Detail and shall never produce a stale-entry recommendation. No implementation under this SPEC shall collapse stat error classes into "path missing".
- **REQ-CSP-006 (ubiquitous, home-resolution and posture preserved)** — The checker shall continue to locate the config through `resolveCodexHomeDir()` (so `CODEX_HOME` decides which file is read), shall introduce no second home-resolution path, and shall remain read-only and fail-open in every direction (unresolvable home / absent config / no entries → silent skip).
- **REQ-CSP-007 (capability gate, reproduction fixture)** — **Where** the defect is reproduced or verified, tests shall construct a synthetic codex home under the system temp root (`t.TempDir()`), isolate `CODEX_HOME` to that fixture (pinning the `codexUserHomeDir` seam as the existing `stubCodexHome` helper does), and cover each path shape — absolute-existing, absolute-missing, home-relative-existing, home-relative-missing, bare `~`, `~user` (unexpandable), relative, oddly-formed, and an indeterminate stat failure — without reading the developer's real `~/.codex` or executing codex against the dev project.

## §D. Acceptance Criteria (inline, Tier S — Given-When-Then)

실행 명령(전 AC 공통): `go test ./internal/cli/ -run 'CodexSkillPath' -count=1 -v` (신규 테스트명 접두) 및 기존 회귀 `go test ./internal/cli/ -run 'TestCheckCodexWiring_StaleHomeSkillsReported|TestCheckCodexWiring_HealthyHomeSkillsNoFinding' -count=1`.

- AC-CSP-001-01: Given a fixture codex home whose config declares a `~/`-prefixed entry whose expansion target EXISTS under the pinned user home, When `codexStaleSkillFinding` runs, Then the entry is NOT counted missing and no stale finding is emitted for it (maps REQ-CSP-002)
- AC-CSP-001-02: Given a `~/`-prefixed entry whose expansion target does NOT exist under the pinned user home, When the checker runs, Then the entry IS counted missing — expansion must not create false negatives (maps REQ-CSP-002)
- AC-CSP-001-03: Given an entry with a relative path (e.g. `skills/foo/SKILL.md`), When the checker runs, Then the entry is NOT counted missing, is reported as a distinct relative classification in Detail, and the "remove the stale entries" directive does not apply to it (maps REQ-CSP-003)
- AC-CSP-001-04: Given an entry with a backslash-bearing NON-absolute path (e.g. `skills\foo\SKILL.md` — a backslash-bearing relative fragment, oddly-formed on BOTH darwin and windows: `filepath.IsAbs` is false on both hosts for this shape), When the checker runs, Then the entry is NOT counted missing and is reported as a distinct oddly-formed classification; on a Windows host, a native absolute `C:\...` entry is classified ABSOLUTE (per `filepath.IsAbs`) and is never oddly-formed; an entry `~otheruser/skills/x/SKILL.md` (another user's home form) is likewise NOT counted missing and NOT expanded — reported under this oddly-formed classification (maps REQ-CSP-004)
- AC-CSP-001-05: Given an absolute-path entry that exists AND an absolute-path entry that does not, When the checker runs, Then the existing entry yields no finding and the missing one IS counted missing with the remove directive — regression guard; the two existing doctor tests stay green (maps REQ-CSP-001)
- AC-CSP-001-06: Given a fixture containing BOTH a real missing entry (an absolute path under the temp root that does not exist — so a finding IS emitted) AND an indeterminate entry created as a symlink loop (`a → b → a`; stat returns ELOOP, not ErrNotExist), When the checker runs, Then the finding's missing count includes ONLY the real missing entry and the indeterminate count for the loop entry is surfaced in Detail — the entry is never folded into missing (preserving the t451 taxonomy verbatim, including the no-finding-when-missing==0 posture at doctor_codex.go:365-366). Where the host cannot create symlink loops (a Windows runner without symlink privilege), the indeterminate entry is substituted with a permission-denied parent, or the sub-assertion skips with an explicit reason — it never silently passes (maps REQ-CSP-005)
- AC-CSP-001-07: Given the new fixture tests, When their sources are scanned for home resolution, Then every new test uses `t.TempDir()` + `stubCodexHome`/seam pinning with zero references to the developer's real home and zero writes outside the temp root, and the config file is still located through `CODEX_HOME` via `stubCodexHome` (maps REQ-CSP-006, REQ-CSP-007)
- AC-CSP-001-08: Given a pinned `codexUserHomeDir` seam returning a temp home that CONTAINS the target file while the real user home does not, When the checker evaluates a `~/`-entry, Then the expansion observably used the seam (entry not missing), proving the expansion routes through `os.UserHomeDir` and not through `CODEX_HOME` or a hardcoded home; a bare `~` entry expands to the seam-pinned user home itself and is likewise NOT counted missing (maps REQ-CSP-002)

### §D.1 RED-now / green-path cells (two-cell adoption, tree `d592b0551`)

| AC | Cell | Observation |
|---|---|---|
| AC-CSP-001-01 | RED-now (right reason) | raw `os.Stat("~/...")` ENOENT, `isNotExist=true` (probe on this tree) → entry counted missing today; flips at plan M2 |
| AC-CSP-001-02 | vacuous-green today (same raw-stat path) | must become determinately green via expansion at M2 |
| AC-CSP-001-03 | RED-now | relative stat against process cwd ENOENT → counted missing today; flips at M2 |
| AC-CSP-001-04 | RED-now | `os.Stat` of any backslash-bearing non-absolute shape ENOENT `isNotExist=true` (probe on `C:\\...` form) → counted missing today; the folded `~user` claim is RED the same way; flips at M2 |
| AC-CSP-001-05 | GREEN now (measured: both existing tests PASS) | regression guard — must stay green through M1-M3 |
| AC-CSP-001-06 | GREEN now (existing indeterminate branch) — the REAL-missing+symlink-loop fixture makes the Detail surfacing assertable without changing the no-finding posture | regression guard — must stay green through M1-M3 |
| AC-CSP-001-07 | structural | verified at M1 by source scan |
| AC-CSP-001-08 | RED-now (no expansion exists — covers both `~/`-prefixed and the folded bare-`~` claim) | flips at M2 |

## §E. Non-Functional Constraints

- **NFC-1 (advisory posture)** — The checker stays advisory and fail-open; no new failure path may block `moai doctor`.
- **NFC-2 (no new dependency)** — Path classification uses the standard library only; no TOML dependency is introduced (the hand-rolled parser posture of `internal/codexwiring` is preserved).
- **NFC-3 (cross-platform)** — The backslash classification must hold on both darwin/linux (oddly-formed) and windows (where a backslash path may be a legitimate absolute path — `filepath.IsAbs` decides, so a Windows-shaped path on a Windows host is absolute, not odd). CI matrix is the verdict.
- **NFC-4 (read-only)** — The user-layer config.toml stays byte-invariant; nothing in this SPEC writes to any codex-owned surface.

## §F. Exclusions

### Out of Scope — sibling card surfaces (t452 / t462)

- t452 (lane-12) codex skill wiring generation/deployment surface — not touched.
- t462 (lane-6) codex e2e exhaustive sweep — no e2e harness work here; verification stays at the unit/checker level under an isolated fixture.

### Out of Scope — parser and semantics deliberately not changed

- Full TOML basic-string escape-sequence decoding in `ParseSkillEntries` (e.g. decoding `\\` to `\`). This SPEC classifies oddly-shaped paths instead of re-interpreting them; a decoder rewrite is a separate decision with its own false-finding risk profile.
- Deciding (or guessing) the resolution base for relative paths — explicitly prohibited by REQ-CSP-003.
- Any auto-repair / auto-removal of stale entries; the checker reports only.

### Out of Scope — doctor surface beyond the stale-skill sub-check

- The broader "Codex Wiring" doctor check (SPEC-CODEX-WIRING-001) — its other sub-checks (binary freshness, PATH lookup, hooks/config parity) are untouched.
