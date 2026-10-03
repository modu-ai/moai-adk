# SPEC-TCD-LLM-DECIDER-001 — Acceptance Criteria

> Each criterion is Given-When-Then and binary-testable, headed by the **Covers** REQ it maps
> (1:1 with REQ-TLD-001..007). "Test" means a named Go test run phase places in
> `internal/cli` (plus the constants assertions in `internal/config`); the verdict is that
> test's actual execution output. Every LLM scenario uses an injected doer
> (`httptest.Server` or stub) — zero real network in the suite.
>
> **RED-now / green-path discipline.** Every AC carries both cells: a RED-now determination
> measured at tree `681ee30b5` (this worktree, `WT-llm-card-decider`) with the reason it is
> red, and the milestone that flips it green. An AC whose red reason is not stated is
> rejected; regression-guard rows state the mutation that proves the green is green for the
> right reason. Total 7/7 (Tier M ceiling 16).

## AC-TLD-001 — env=llm happy path records the model judgment

**Covers**: maps REQ-TLD-001

**Given** `MOAI_TODO_DECIDER=llm` and a fake endpoint injected as the decider's HTTP seam,
returning a valid closed-set judgment (`priority: high`, `blocked: false`,
`mode: parallelizable`, a one-line reason),
**When** `moai todo add "<text>"` runs,
**Then** the add exits 0 printing `<id> <position>`, and the queue record for that card carries
exactly the returned priority/blocked/mode, `decider` exactly `"llm"`, a present
`classified_at`, and a non-empty single-line `reason` — asserted via a `--json` queue read.

## AC-TLD-002 — selection matrix: unset · default · llm · invalid · jev · file precedence

**Covers**: maps REQ-TLD-002

**Given** a counting fake endpoint and the selection cases (a) env unset, (b) env `default`,
(c) env `llm`, (d) env `banana`, (e) env `jev`, (f) env `llm` plus `--classification-file`
carrying a valid judgment,
**When** `moai todo add "<text>"` runs once per case,
**Then** (a) and (b) record the deterministic default (`decider: "default"`) with ZERO endpoint
requests observed; (c) exercises the LLM path; (d) and (e) each exit 2 with a usage refusal,
nothing written, the message naming the accepted set and — for (e) — the parent SPEC's jev
refusal wording; (f) records the supplied file's judgment (its own decider identity) with ZERO
endpoint requests — the explicit file beat the standing selection.

## AC-TLD-003 — every LLM failure degrades fail-safe with exactly one notice

**Covers**: maps REQ-TLD-003

**Given** `MOAI_TODO_DECIDER=llm` and the endpoint respectively (i) unreachable (closed port),
(ii) returning HTTP 500, (iii) returning a non-JSON body, (iv) stalling past the configured
timeout, (v) returning valid JSON with `priority: "urgent"`,
**When** `moai todo add "<text>"` runs once per failure class,
**Then** every case admits the card (exit 0) with the fail-safe default values (priority
`normal`, blocked `false`, mode `serial`, `decider: "default"`), stderr carries EXACTLY ONE
fallback notice line (counted), and the `<id> <position>` line still prints.

## AC-TLD-004 — closed-set acceptance and identity overwrite

**Covers**: maps REQ-TLD-004

**Given** fake endpoints returning judgments whose `decider` field claims `"human"` and `"jev"`
respectively, with otherwise-valid fields,
**When** `moai todo add "<text>"` runs with env `llm`,
**Then** the recorded decider is exactly `"llm"` in both cases, and the value-set validation is
performed only through `kanban.ValidateCardClassification` — asserted by the behavioral cases
above plus a grep showing no re-spelled set literals in the new file.

## AC-TLD-005 — the judgment is computed outside the lock, attached inside it

**Covers**: maps REQ-TLD-005

**Given** `MOAI_TODO_DECIDER=llm` and an endpoint stalling 1500 ms per judgment,
**When** two `moai todo add` invocations run concurrently,
**Then** both succeed with correct classifications and the total wall time stays measurably
below the serialized bound (the two judgments overlapped — asserted against a bound that a
serial-in-lock placement cannot meet), and while one add's judgment is in flight a `todo list`
read completes without waiting out the stall.

## AC-TLD-006 — transport discipline: bounded timeout, credential flowed, secret never printed

**Covers**: maps REQ-TLD-006

**Given** a fake credential carrying the marker string wired through the key-loader seam and a
fake endpoint,
**When** the happy path AND one forced-failure path both run,
**Then** the marker string appears in ZERO bytes of process stdout/stderr across both runs —
with the positive control that the fake server DID receive it in the Authorization header (so
the 0-hit is meaningful, not an unflowed credential); the HTTP client's timeout equals the
`internal/config/defaults.go` constant (asserted via the seam); and greps confirm no inline
`MOAI_TODO_DECIDER` literal outside `envkeys.go` and no endpoint literal outside
`defaults.go`.

## AC-TLD-007 — default-off regression: unset behavior is unchanged

**Covers**: maps REQ-TLD-007

**Given** `MOAI_TODO_DECIDER` unset,
**When** `moai todo add "<text>"` runs,
**Then** the classification is the deterministic default with the default reason, the counting
endpoint observes zero requests, and stderr carries no line beyond the pre-SPEC expectation.

## 채택 셀 — RED-now 판별과 green 마일스톤

| AC | RED-now (`681ee30b5` 기준) | green 마일스톤 |
|---|---|---|
| AC-TLD-001 | RED — `llm` decider 구현이 없다 (seam 뒤 구현은 `default`·`static`·`unavailable` 세 개뿐) | M2 |
| AC-TLD-002 | RED — env 선택 자체가 존재하지 않는다 (`todo.go:752`는 flag 분기만 읽는다) | M1 (c 케이스는 M2) |
| AC-TLD-003 | RED — LLM 경로가 없어 실패 분류를 표현할 수 없다 | M2 (notice·잠금 재배치는 M3) |
| AC-TLD-004 | RED — 같은 이유 (구현 부재) | M2 |
| AC-TLD-005 | RED — 유일한 decider 들이 전부 즉시응답이고 분류 호출이 `Mutate` 안에 있다 — 시나리오 자체를 구성할 수 없다 | M3 |
| AC-TLD-006 | RED — 전송 계층이 없어 자격증명 흐름·시간제한을 단언할 대상이 없다 | M2 |
| AC-TLD-007 | GREEN-at-M1 회귀 가드 — 미설정 경로는 오늘도 기본 decider 로 동작 (양성 대조 = AC-TLD-002 (c) 가 같은 selector 를 뒤집음, 변이로 시험) | M1 |
