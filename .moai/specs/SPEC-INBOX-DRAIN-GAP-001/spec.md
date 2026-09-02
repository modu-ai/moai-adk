---
id: SPEC-INBOX-DRAIN-GAP-001
title: "Distributed lessons-inbox lifecycle — collector-side write-time cap, archive rotation, and CLI drain surface"
version: "0.1.0"
status: draft
created: 2026-09-02
updated: 2026-09-02
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: internal/hook
lifecycle: spec-anchored
tags: "lessons-inbox, drain-gap, bounded-growth, write-time-cap, rotation, cli-inbox, lsel-standdown, schema-version, distributed-users"
era: V3R6
tier: M
depends_on: [SPEC-LSEL-DRAIN-STALL-001]
related_specs: [SPEC-LSEL-LOCAL-EVOLUTION-001]
---

# SPEC-INBOX-DRAIN-GAP-001 — Distributed lessons-inbox lifecycle (bounded growth without a shipped drain)

## HISTORY

- 2026-09-02 (plan-phase, v0.1.0) 최초 작성. 카드 t280 (Class C, Tier M). 카드 전문이 범위 권위: "분산 사용자 lessons-inbox 무한 증식 — 수집은 배포되고 드레인은 dev-only로 제공되는 구조 불일치". plan-phase 재검증(본 SPEC §B)으로 카드 프리미스 전건 확인: 수집 체인은 배포 템플릿에 존재하고, 드레인 체인은 dev-only이며, 배포 트리에는 인박스를 소비하는 코드가 없다(consumer vacuum). 본 SPEC은 t259(SPEC-LSEL-DRAIN-STALL-001) REQ-LDS-009가 명시적으로 범위 밖으로 이월한 분산 사용자 축을 다룬다.

## §A. User Story

**As a** MoAI-ADK distributed user whose install collects tool-failure lessons continuously via a shipped hook, **I want** the lessons inbox to be bounded by construction — never growing without limit, never depending on a drain trigger my install does not ship — **so that** `.moai/lessons-inbox.jsonl` cannot become an unbounded liability on a machine that has no consumer for it.

**결과 가설:**
- 표준 설치(로컬 LSEL curator 부재)에서 인박스 크기는 기록된 상한 아래로 수렴한다 — 사용자가 아무것도 실행하지 않아도(캡은 컬렉터 자체에 있다).
- 유지자 머신(GOOS, curator 활성)에서는 t259가 세운 드레인 소유권이 그대로 유지된다 — 본 SPEC의 상한은 curator가 있으면 움직이지 않는다(stand-down).
- 사용자는 `moai inbox status`로 인박스 상태를, `moai inbox drain`으로 수동 정리를 얻는다 — 새 세션 훅 배선도, cron/launchd/daemon도 없다.

## §B. Context and Background

### §B.1 프리미스 재검증 (2026-09-02, worktree HEAD `131daa290` 기준 실측)

**수집 체인 — 배포됨 (shipped):**

| 단계 | 근거 |
|---|---|
| `PostToolUseFailure` 훅 배선 | `internal/template/templates/.claude/settings.json.tmpl` (PostToolUseFailure 이벤트 → `handle-post-tool-failure.sh` 폴백 래퍼 포함) |
| 래퍼 스크립트 배포 | `internal/template/templates/.claude/hooks/moai/handle-post-tool-failure.sh.tmpl` 존재 실측 |
| Go 디스패치 | `internal/hook/post_tool_failure.go:76` `recordToolFailureEvent` |
| 인박스 적재 | `internal/hook/failure_observer.go:114-154` — `lessonsInboxStub{timestamp, event_key, summary, source}`, `.moai/lessons-inbox.jsonl` append-only JSONL, 0o600, fail-open, **appends open/close the file by path on every append** (line 149) |

**드레인 체인 — dev-only (NOT shipped):**

- `.claude/settings.local.json` SessionStart → `session_drain.sh` 래퍼(배타잠금 + `clusters-history/` 보존) → `drain.sh`. 오프셋은 `<state-dir>/drain-offset.json` (`drain.sh:48`).
- `internal/template/templates/.claude/hooks/moai/` 아래에 `session-drain`/`backlog` 템플릿 부재 실측 (grep 0건) — §2.3 doctrine상 local-only가 의도다.
- 소비 로직(`hns-lsel-curator` 스킬: severity 필터·클러스터링·importance 스코어링) 역시 dev-only.

**Consumer vacuum — 배포 트리에 인박스 소비자 없음 (실측):**

- `lessons-inbox` 전수 grep: 배포 트리에서 읽는 곳은 navigator-audit 통계(라인 카운트, `moai-workflow-project` 스크립트)뿐 — 드레인/소비 역할의 코드는 0건. 표준 설치 사용자에게 인박스는 **소비자 없이 무한히 자라는 파일**이다.

**`moai update`와의 관계 (실측):**

- `ManagedCleanTargets` (`internal/cli/update/deploy/deploy.go:56`)이 지우는 뿌리는 `.claude/settings.json`, `.claude/hooks/moai`, `.moai/config` 등 — `.moai/lessons-inbox.jsonl`은 `.moai` **루트**에 있어 어떤 와이프 대상도 아니다. 즉 인박스는 업데이트를 무한히 생존하며, 상한이 없으면 증식도 무한하다.

### §B.2 계보 (t259 lineage)

- SPEC-LSEL-DRAIN-STALL-001 (completed) REQ-LDS-009: SessionStart 배선을 `.claude/settings.local.json`에 두는 것은 유지자 머신 로컬 산출물이며, "tracked `.claude/settings.json` entry would be wiped by every `moai update` (§2.3), so tracked wiring is affirmatively wrong" — 당시 판정은 **유지자 로컬 드레인**에 관한 것이고, 분산 사용자 축은 명시적으로 이월됐다. 본 SPEC은 그 이월분만을 다루며 t259의 결정(wrapper 경유 위생 REQ-LDS-010, offset 역학, clusters-history 보존)을 재판정하지 않는다.
- CLAUDE.local.md §28 (dev-only 문서 — 내부 맥락 인용이며 배포 동작이 아님): 드레인 운영은 `session_drain.sh` 래퍼 경유로만 한다. 본 SPEC의 curator refusal 설계(REQ-IBX-007)는 이 위생을 CLI 표면에서도 강제한다.

## §C. Requirements (GEARS)

### §C.1 Bounded growth (collector-side cap)

- **REQ-IBX-001 (ubiquitous)** — The lessons-inbox collector shall enforce a hard size cap on `.moai/lessons-inbox.jsonl` on every append, for installs where the LSEL drain-ownership marker is absent. The cap value is a single-source default (threshold constant in `internal/config/defaults.go`, no duplicate literals).
- **REQ-IBX-002 (state-driven, curator stand-down)** — While the LSEL drain-ownership marker (the `.moai/state/lsel/` directory) is present, the collector shall not rotate, trim, or otherwise mutate the inbox beyond its own append — the local drain owns the inbox lifecycle on that machine (t259 lineage preserved, not re-decided).
- **REQ-IBX-003 (event-driven, rotation)** — When an append observes the live inbox at or over the size cap, the collector shall rotate the live file into a bounded archive generation before continuing the append on a fresh live file.
- **REQ-IBX-004 (unwanted, retention bound)** — The collector shall not retain more than two rotated archive generations (`.lessons-inbox.jsonl.1`, `.2`); older generations shall be deleted at rotation time. This deletion is the documented eviction policy, not an error path.

### §C.2 CLI lifecycle surface

- **REQ-IBX-005 (capability gate)** — Where a user invokes `moai inbox status`, the CLI shall report the live inbox size, line count, cap distance, archive generations present, and the active ownership regime (`curator` vs `cap-managed`), exiting 0.
- **REQ-IBX-006 (capability gate)** — Where the LSEL drain-ownership marker is absent, `moai inbox drain` shall perform the same bounded rotation as the write-time cap and report the rotation statistics, exiting 0.
- **REQ-IBX-007 (capability gate + unwanted)** — Where the LSEL drain-ownership marker is present, `moai inbox drain` shall exit non-zero with a notice pointing at the wrapper-mediated drain (`.claude/skills/hns-lsel-curator/session_drain.sh`) and shall not rotate, trim, or mutate the inbox.

### §C.3 Format and failure discipline

- **REQ-IBX-008 (ubiquitous, schema version)** — Every stub the collector appends shall carry an integer schema-version field, and every inbox reader shipped by this SPEC shall tolerate the field's absence (a pre-upgrade line parses as version 1).
- **REQ-IBX-009 (unwanted, fail-open)** — The collector shall not block, fail, or delay the session on cap-check or rotation errors; a failed rotation degrades to a logged warning and the append proceeds best-effort on the existing file.
- **REQ-IBX-010 (unwanted, no scheduling surface)** — The design shall not introduce cron, launchd, daemon, or any new SessionStart hook wiring; the inbox lifecycle surfaces are exactly the collector's write-time cap and the two manual CLI verbs. The design shall not edit any file under `internal/cli/update/`.

## §D. Non-Functional Constraints

- **NFC-1 (concurrency tolerance)** — Rotation races a concurrent appender only within one in-flight append (each append opens/closes by path; `failure_observer.go:149`); a stub may land in the rotated generation. This is acceptable and documented — the same interleaving tolerance the inbox already grants concurrent Stop hooks (EC-4 lineage).
- **NFC-2 (rotation cost)** — Rotation is a rename (atomic on POSIX) plus at most two renames and one delete for the retention chain; the steady-state cost of the cap is one file-size stat per append.
- **NFC-3 (fail-open parity)** — The cap path holds the same fail-open discipline as the existing append (errors logged and swallowed; a learning-loop write must never block the session).
- **NFC-4 (curator-first precedence)** — On any machine where the local drain exists, local behavior is byte-identical to pre-SPEC behavior. No t259 offset semantics, wrapper hygiene, or clusters-history preservation changes.
- **NFC-5 (platform)** — The retention chain must hold on Windows, where `os.Rename` fails over an existing destination; the implementation removes the destination before renaming (documented in plan.md risks; cross-platform CI is the verdict).

## §E. Design Decision Record (Class C)

### §E.1 Decision taken — collector-side cap + manual CLI surface (Options 3+2 composite)

The bound lives in the collector itself: a write-time size cap with archive-rotation (REQ-IBX-001..004), armed only when `.moai/state/lsel/` is absent (REQ-IBX-002), plus a schema version field (REQ-IBX-008), plus a manual CLI surface `moai inbox status` / `moai inbox drain` with curator-refusal (REQ-IBX-005..007). No new hook wiring, no daemon, no update-path edits (REQ-IBX-010).

**Why this wins:**
1. **The cap cannot be orphaned.** The binary is the one artifact guaranteed present and version-consistent on every install and every update — the cap rides the same binary as the collector itself. A user who runs nothing but Claude Code sessions is still bounded.
2. **Zero new wiring = zero calm-first-update exposure.** The t230/t255-series calm-first-update constraint applies to NEW shipped hook wiring; this design ships none. No first-update misfire class exists because there is no wiring to misfire.
3. **The consumer vacuum is respected, not papered over.** A drain for users can only trim (verified: nothing shipped consumes the stubs). Trimming does not need session-start latency billed to every user every session — it needs a size check at write time.
4. **Curator precedence is structural, not conventional.** The stand-down marker makes the maintainer machine's t259 semantics a hard invariant (NFC-4), so this SPEC re-decides nothing it must not.

### §E.2 Runner-up — Option (1), template-shipped SessionStart drain trigger — and why it lost

**What redistribution actually does to tracked wiring (measured, §B.1):** `ManagedCleanTargets` wipes `.claude/settings.json` and `.claude/hooks/moai`, then redeploys both from the embedded template FS. Therefore (a) hand-added tracked-settings entries are wiped every update — the t259 REQ-LDS-009 grounding, fatal for wiring that references non-templated scripts; (b) a template-shipped wiring + template-shipped script pair is version-consistent by construction — both are re-laid from the same binary snapshot, so Option (1) is structurally viable on survival grounds, and the stale-wiring hazard reduces to cross-version field renames, which the wholesale redeploy also resolves.

**Why it lost anyway (cost and semantics, not survival):**
1. **Consumer vacuum.** A shipped trigger can only trim, and the trim payload needs no per-session hook. Option (1) bills session-start latency to every user, every session, for work a one-sysstat cap does at write time.
2. **Config hostility.** Template-shipped wiring cannot be disabled per-user except by editing a file every update re-lays — the same fight-the-redeploy class the card's §2.3 concern names.
3. **Calm-first-update coupling.** New SessionStart wiring + a new hook script must ship atomically and must not misfire on first update for existing installs — an entire hazard class the collector-side cap avoids by shipping no wiring at all.
4. **Scope.** Shipping a drain trigger implies deciding what it invokes; shipping the curator body (clustering/scoring) to every user is a scope and noise-profile explosion the card flags, and a reduced trim-only hook duplicates what the cap already does better.

**Option (2) partial adoption.** The CLI verb is adopted (REQ-IBX-005..007) as the user-agency surface, but the card's opportunistic-invocation variant (auto-drain inside other CLI flows) is rejected: it would couple the inbox lifecycle into `internal/cli/update/` and other command paths for no bounded-growth benefit (the cap already bounds growth), and it would move the merge-judgment surface for no reason.

**Option (3) adoption.** The schema version field and the hard cap with a documented eviction policy are adopted wholesale — they are the core of the decision, not an accessory. The version field is introduced at `1` simultaneously with the first consumer (the rotation/CLI code); absence reads as `1`.

## §F. Out of Scope

### Out of Scope — LSEL curator distribution

- The clustering/scoring curator (`hns-lsel-curator`), its severity filter, its PROPOSE/APPLY seam, and `clusters.json` staging remain dev-only (GOOS-local). This SPEC ships no curator logic to users; user installs get bounded growth, not lesson mining.

### Out of Scope — tracked settings.json drain wiring

- Option (1) is rejected as the mechanism (§E.2). No SessionStart entry is added to tracked or template `settings.json`, and no drain hook script is added to the template tree.

### Out of Scope — update-path changes

- No file under `internal/cli/update/` is edited; no opportunistic drain is added to `moai update`, `moai doctor`, or any other command. (Merge-judgment boundary with card t239/SPEC-LLMCFG-PRESERVE-001: that SPEC is test-only in a different subsystem — `internal/config` llm.yaml preservation — and shares no code path with this one; the boundary is recorded here for the lane's batch judgment.)

### Out of Scope — scheduling surfaces

- No cron, launchd, daemon, background worker, or timer. The card excludes them explicitly; the binary-carried write-time cap makes them unnecessary for the bounded-growth goal.

### Out of Scope — stub content quality

- What the stubs contain (severity classification, summary truncation at the collector) is REQ-HRR-006 territory, already landed. This SPEC governs the inbox's size and lifecycle only, not the content's quality or any future user-facing consumer of it.
