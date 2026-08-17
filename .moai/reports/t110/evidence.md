# t110 — 공장장 역할 규율 명문화: 증거

Date: 2026-08-17 · Card: t110 (lead dispatch) · Branch: WT-t110 · Base: 0ede5db6a

## Card

> 공장장 역할 규율 명문화 (t84 재정의, §6 최종안 반영 — t106 착지로 전제 해제).
> 범위: 플래그 신설 없음. 본 카드는 역할 규율만 문서화:
> [HARD] ① 위임 채널은 큐(디스크), 세션 간 메시지는 nudge 알림만(배달 비보장·수신 측 할당량 소모)
> ② 백로그 승격은 운영자 행위 유지(kanban-dispatch.md [HARD])
> ③ 공장장은 큐의 유일한 생산자 + 최종 PASS/FAIL 판정(GLM 반장이 Claude 작업자를 못 만드는 매핑 제약 보완)
> 주의: always-loaded 예산 초과 중 — kanban-dispatch.md 편집 시 순증가 금지

## Claim

`kanban-dispatch.md` now states the three [HARD] discipline rules of the
board's lead role (the card's "공장장"; the `chief` NAMING and the `-k`
entry guidance remain t97's scope, so this card documents the role under the
existing "lead" term): the queue on disk is the delegation channel and
messages are nudges; the lead is the queue's sole producer while promotion
stays the operator's act; the final PASS/FAIL verdict is the lead's and is
never delegated to the lane that produced the work. The file SHRANK despite
the additions.

## Change — additions (kanban-dispatch.md, mirrored to the template)

1. **New subsection "The delegation channel is the queue"** (inside The
   dispatch cycle): work is delegated through the queue on disk — the file
   resolves against the primary checkout from every linked worktree (one
   repository, one queue; the t106-landed property that made this principle
   viable), so a card admitted from anywhere is visible everywhere. A
   cross-session message is a nudge, never the delegation: delivery is not
   guaranteed and a delivered message consumes the recipient's quota; no
   dispatch may depend on a message arriving.
2. **"Entry into the board is an operator act" rewritten** with two [HARD]
   blocks: **the lead is the queue's sole producer** (the operator asks, the
   lead turns the request into a card via `moai todo add`; production is
   translation, not invention — nothing enters the queue the operator did
   not ask for), and **promotion is the operator's act, always** (the lead
   presents the queue, never picks, never reorders by inferred priority; an
   empty queue is a state to report). Supersedes the looser "the mechanism
   is /moai todo" listing while keeping the operator-origin principle intact.
3. **"Completion is read, never trusted" gains a verdict paragraph**: the
   final PASS/FAIL is the lead's, read from the evidence on disk, never
   delegated to the producing lane — and the division is structural, not
   ceremonial: on a mixed-backend board the lane sessions cannot commission
   judgment work onto the lead's backend, so the verdict has a home in the
   lead even when the execution has none (the card's GLM-lead/Claude-worker
   mapping constraint, stated backend-neutrally for the distributed
   template).

**Consistency fix surfaced by principle ①**: the Scope section claimed the
"dispatch cycle rides entirely on cross-session messaging" — directly
contradicted by the queue-as-channel rule. Rewritten: nudge delivery rides
on messaging (unavailable on native Windows / some providers — see
cross-session-messaging.md); the queue itself keeps working without it.

## Change — compensating reductions (same file)

Eight compressions, wording-only, no rule weakened: dispatch-language
rationale merged into one paragraph; the card-classes meta comment; two
CodeRabbit rationale passages; the verification-incident narrative; the
integration closing sentence (duplicated the push bullet); the Boundaries
no-occupant bullet; dispatch-cycle naming paragraph; `/clear` and
`WT-` wording tightened.

## Evidence (verbatim)

Size discipline — the card's no-net-growth constraint:

```
$ wc -c .claude/rules/moai/workflow/kanban-dispatch.md
   27764   (base 0ede5db6a)
   ... after additions only ...
   28183   (+419 — rejected interim state, more trimming applied)
   ... after compensating reductions ...
   27709   (final: −55 vs base; ~1,050 B added, ~1,105 B trimmed)
$ diff .claude/rules/moai/workflow/kanban-dispatch.md \
       internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md
MIRROR SYNCED (byte-identical after cp; make build re-ran)
```

Template guards (neutrality + mirror parity; strict leak mode):

```
$ MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/template	24.066s
```

Always-loaded budget guard — the pre-existing overage, NOT a t110 failure:

```
$ go test ./internal/config/ -run 'TestAlwaysLoadedTokenBudget$' -count=1 -v
    token_budget_guard_test.go:67: always-loaded surface = 76666 tokens,
    exceeds budget 76000 (overflow 666); surface has 17 entries
--- FAIL: TestAlwaysLoadedTokenBudget
```

Attribution that this failure predates t110: the guard landed in
SPEC-TOKEN-EFFICIENCY-001 (single commit, present in base 0ede5db6a), and
the lead's dispatch measured the overage at ~680 BEFORE this card (t114 is
the diet card in flight). t110's only always-loaded edit shrinks its file by
55 bytes (≈ −14 tokens at 4 B/token), so the overflow could only have been
equal or larger at base — arithmetic proof, no re-run needed. The card's
constraint ("no net growth in kanban-dispatch.md") is met with margin.

## Baseline attribution

All measurements were taken on this branch's tree (WT-t110) against base
0ede5db6a's file (27,764 B) on the same morning; the budget figure 76,666 is
the guard's own output on the current tree.

## Gaps

- The byte→token conversion (4 B/token) is an approximation; the guard's own
  666-token figure is the authoritative overage measurement.
- Whether the eight compressions preserve every nuance is the hub review's
  call — each was wording-only, but review should read the trimmed passages
  (dispatch-language rationale and the two CodeRabbit passages carry the
  most argument structure).
- The `chief` naming, the `-k` entry-point guidance, and the SessionStart
  notice overhaul remain t97's scope; this card deliberately introduces no
  new term.
- Full-suite / cross-platform verdict deferred to CI per the lane-local
  discipline.

## Residual risk

- The "sole producer" discipline is a documented norm, not a mechanical
  gate: an operator typing `moai todo add` directly still works, and the
  rules do not (and should not) prevent it — the norm binds the lead's
  behavior, not the operator's tools.
- The mixed-backend verdict paragraph states the constraint generically; if
  the backend set changes, the sentence survives but its motivation should
  be re-read.
