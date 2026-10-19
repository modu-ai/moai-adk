# SPEC-LANE-READ-COMPACT-001 — Decision Index

This index lists the decisions of card t1810, carried from `SPEC-LANE-READ-VERBS-001` when board ruling `d-20261015T122748Z-6d6b` split card t1777 by risk. It is stateless on the status axis: it carries no `status:` field. The SPEC lifecycle lives in `spec.md` alone.

Labels use only the four routing labels: `DECIDED`, `POLICY-COVERED`, `EVIDENCE-NEEDED`, `FOUNDER`. The Q numbers are the parent's, so that each row still matches the board record that decides it.

## Routing

Board record `d-20261015T113856Z-6f23` states that card-scoped brain rulings under the operator-proxy standing count as decision-board records, with the board id as the anchor. Each DECIDED row cites `board:<id>` and pins its record line and body line verbatim, as copied in `decision-anchor.md`. Rows here: Q1, Q2, Q3, Q6, Q7, Q8, and Q9 (its todo and factory half).

---

### Q1: Should the default `moai todo list --json` stay the full backlog record, with a compact form added behind an opt-in flag?

Label: DECIDED
Authority anchor: board:d-20261015T031405Z-dfd8 (anchor = board id, for card-scoped brain rulings under the operator-proxy standing, as record d-20261015T113856Z-6f23 states). Pinned record, verbatim from `moai decision read --scope card:t1777`:
```text
decision record: decided_by=brain 영실이 (operator proxy, directive 2026-10-15) evidence_refs=lane-23 plan draft SPEC-LANE-READ-VERBS-001 open decisions Q1/Q4; frozen goldens REQ-TCL-014, AC-TST-012, AC-TSP-021 ladder_path=② id=d-20261015T031405Z-dfd8 scope=card:t1777 kind=ruling
  body: t1777 Q1 = (b): moai todo list --json default stays full (frozen goldens untouched); add opt-in --compact; the effect lives in consumers — lane cycle (--auto path), lane guidance text and factory lane prompts default to --compact, and that consumer change is an AC (a flag nobody uses misses the card's intent). Q4 = S-2: session list default shows live sessions only, cap exempt; dead rows behind --all (or equivalent). Liveness by PID probe (kill -0 etc.), not registry row presence; indeterminate counts as live. The AC change from the 2KB cap to 'live entries complete, dead rows excluded by default' is a scope change: record it in decision-index with this decision id and rewrite the AC. Tier unchanged (not S-3).
```
Why unresolved: The current default is a frozen contract. `TestTodoListJSON_GoldenByteIdentity` (`internal/cli/todo_json_golden_test.go:65`) pins it, and SPEC-TODO-CLAIM-LEASE-001 REQ-TCL-014 and SPEC-TODO-TRANSITION-STAMPS-001 AC-TST-012 name it as frozen. SPEC-TODO-SURFACE-POLISH-001 AC-TSP-021 says `--json` ignores `--limit`. The record keeps the default and adds `--compact` (option b). The consumer clause is measured as follows: the `--auto` cycle reads the store in process (`internal/cli/todo_auto.go:280`) and has no CLI JSON call site, so the lane-facing reads are the doctrine sentence and the gtd.md documentation (plan § D.4). The 0.1.0 bound on the full default is withdrawn; the bound moves to `--compact` (row Q9).
Class: product-level
Default: (b) (rule: preserves current behavior)
Alternate: (a)
Operator verdict: DECIDED by board:d-20261015T031405Z-dfd8: (b): the default `moai todo list --json` stays the full record, an opt-in `--compact` is added, and the consumers default to `--compact`.

---

### Q2: What is the row cap of the compact todo form?

Label: DECIDED
Authority anchor: board:d-20261015T113856Z-6f23 (anchor = board id, for card-scoped brain rulings under the operator-proxy standing, as record d-20261015T113856Z-6f23 states). Pinned record, verbatim from `moai decision read --scope card:t1777`:
```text
decision record: decided_by=brain 영실이 (operator proxy standing bb5e) evidence_refs=.moai/worktrees/t1777 decision-index.md 0.2.3; plan-audit-iter3.md FAIL 0.8125 blocking 1 (B1R) codex fail; d-20261015T031405Z-dfd8 ladder_path=brain ruling under bb5e; standing C run-transition (9e76/75ca) id=d-20261015T113856Z-6f23 scope=card:t1777 kind=ruling
  body: t1777 FOUNDER 행 확정: Q2 = compact todo 형식 행 상한 6 (최악 1,963B ≤ 2048, 실측 L-15), 인코딩 폭 접두 규칙(상한 20) 미채택. Q3 = factory status 기본 출력 행 상한 3 + quota 후보 2 (최악 1,899B, 실측 L-16), 전체는 --run <id>(Q8), 대안 A1·A2·B·C 미채택(REQ-TCD-010 유지). Q5 = L-A: PID 탐침, 판정 불가는 살아 있음(dfd8과 같음); JSON 기본 경로에서 REQ-SCV-006 '탐침 없음'을 뒤집는 것을 포함; 텍스트 뷰·--all·필드 집합은 REQ-SCV-006 그대로. Q9 = 범위 변경 확정: 세션 기본값은 2KB 상한 대신 '살아 있는 항목 완전·죽은 행 기본 제외', todo 기본값은 상한 없음(상한은 --compact로), factory 기본값은 Q3 상한 유지. Q7 = (a), Q8 = (a) DEFAULT-APPLIED 확정. Q6 = run 단계 측정 그대로. 앵커: 운영자 대리 standing 아래 카드 범위 브레인 ruling(dfd8과 이 기록)은 결정 보드 기록으로 인정, 앵커 = 보드 id. 진행: run 전환 — 정정 커밋 1개(B1R: windows-latest CI 잡에서 Windows 생존 판정 함수를 실제 실행하는 AC 1개 + M2 Exit 1줄 + decision-index 해당 행 DECIDED 전환, 레코드 줄 인용) → 그 커밋만 델타 1회(codex 필수 pass) → kickoff; FAIL이면 hold + t1777a·t1777b 분할.
```
Why unresolved: The record does not settle the cap, so the operator decides it on the measured bound. Command: `python3 -I .moai/specs/SPEC-LANE-READ-COMPACT-001/measure-caps.py <BIN>` (read-only; verbatim output in `acceptance.md` § D, ledger L-12). The script validated its encoder byte-for-byte against the three live outputs (todo default, factory status, session list) before it printed.
Measured, worst case (declared widths; each prefix rune at its 6-byte JSON width): cap 5 = 1,673 bytes; **cap 6 = 1,963 (within 2048; proposed)**; cap 7 = 2,253; cap 8 = 2,543; cap 10 = 3,123.
Measured, live queue (163 live items at the L-12 run): cap 6 = 817; cap 10 = 1,273; cap 15 = 1,833; cap 20 = 2,355 (over the bound).
Hypothesis check: with 3-byte CJK prefixes and no escapes, cap 10 = 1,923 (within) and cap 11 = 2,093 (over). Escapable characters push cap 10 over the bound, so the proposal is cap 6.
Bound test at run-phase: `TestTodoListCompactByteBound` (AC-LRC-003).
Class: product-level
Alternate (operator option, not chosen): an encoded-width prefix rule (the first 40 JSON-encoded bytes, then `...`) allows cap 20 (worst case 2,023 bytes; cap 21 gives 2,113).
Operator verdict: DECIDED by board:d-20261015T113856Z-6f23: row cap 6 for the compact todo form (worst case 1,963 B, L-12); the encoded-width prefix option (cap 20) is not adopted.

---

### Q3: What is the row cap of the active-run listing of `moai factory status --json`?

Label: DECIDED
Authority anchor: board:d-20261015T113856Z-6f23 (anchor = board id, for card-scoped brain rulings under the operator-proxy standing, as record d-20261015T113856Z-6f23 states). Pinned record, verbatim from `moai decision read --scope card:t1777`:
```text
decision record: decided_by=brain 영실이 (operator proxy standing bb5e) evidence_refs=.moai/worktrees/t1777 decision-index.md 0.2.3; plan-audit-iter3.md FAIL 0.8125 blocking 1 (B1R) codex fail; d-20261015T031405Z-dfd8 ladder_path=brain ruling under bb5e; standing C run-transition (9e76/75ca) id=d-20261015T113856Z-6f23 scope=card:t1777 kind=ruling
  body: t1777 FOUNDER 행 확정: Q2 = compact todo 형식 행 상한 6 (최악 1,963B ≤ 2048, 실측 L-15), 인코딩 폭 접두 규칙(상한 20) 미채택. Q3 = factory status 기본 출력 행 상한 3 + quota 후보 2 (최악 1,899B, 실측 L-16), 전체는 --run <id>(Q8), 대안 A1·A2·B·C 미채택(REQ-TCD-010 유지). Q5 = L-A: PID 탐침, 판정 불가는 살아 있음(dfd8과 같음); JSON 기본 경로에서 REQ-SCV-006 '탐침 없음'을 뒤집는 것을 포함; 텍스트 뷰·--all·필드 집합은 REQ-SCV-006 그대로. Q9 = 범위 변경 확정: 세션 기본값은 2KB 상한 대신 '살아 있는 항목 완전·죽은 행 기본 제외', todo 기본값은 상한 없음(상한은 --compact로), factory 기본값은 Q3 상한 유지. Q7 = (a), Q8 = (a) DEFAULT-APPLIED 확정. Q6 = run 단계 측정 그대로. 앵커: 운영자 대리 standing 아래 카드 범위 브레인 ruling(dfd8과 이 기록)은 결정 보드 기록으로 인정, 앵커 = 보드 id. 진행: run 전환 — 정정 커밋 1개(B1R: windows-latest CI 잡에서 Windows 생존 판정 함수를 실제 실행하는 AC 1개 + M2 Exit 1줄 + decision-index 해당 행 DECIDED 전환, 레코드 줄 인용) → 그 커밋만 델타 1회(codex 필수 pass) → kickoff; FAIL이면 hold + t1777a·t1777b 분할.
```
Why unresolved: As Q2, and two completed SPECs bear on the shape. The default lists the active run's cards in abbreviated rows (`run_id`, `card_id`, `state`, `stage`, `mode`, `priority`, `lease_holder`, `lease_expired`). REQ-TCD-010 (SPEC-TODO-CLASSIFY-DISPATCH-001) asks the factory status surface for each card's lane, mode, and priority, so the rows keep `mode` and `priority`, and REQ-TCD-010 holds on the default. Under quota pressure the `quota` key carries one candidate per live GLM or GPT lane, which no row cap bounds, so the proposal also caps the default's candidates at 2, with a `candidates_withheld` count, and cuts `lease_holder` and candidate labels to 16 encoded bytes. The candidate cap departs from REQ-QAS-019 (SPEC-QUOTA-AWARE-SCHEDULING-001), which asks the status JSON for the candidate lanes, on the default JSON only: `--run <id>`, `--full`, and the text view keep the whole list, and the AC-QAS-019 test reads `--run`. The verdict on this row covers that departure.
Command and validation: ledger L-13 (`acceptance.md` § D), which loads `measure-caps.py` unchanged and validates its encoder byte-exact against the three live outputs in the same run.
Measured, worst case under quota pressure (both windows fresh with every optional field set, holder and labels at 16 encoded bytes, the longest state `needs-decision`): **cap 3 = 1,899 bytes (within 2048; proposed)**; cap 4 = 2,161. With three rows, three candidates also fit (1,978) and four do not (2,057).
Why the 0.2.0 proposal (cap 6 = 1,924, ledger L-12) is withdrawn: it counted the quota block with pressure off and the holder cut at 16 raw bytes. Measured again in L-13, cap 6 reaches 2,113 bytes with live windows, pressure on, and two candidate lanes, and 2,404 bytes with a holder of 16 `<` characters. Both candidates in the card (8 and 10) fail by more.
Measured, live active run `tmvgvn` (68 cards) at cap 3: 1,338 bytes with the live quota block; 1,557 with live windows, pressure on, and two candidates.
Bound test at run-phase: `TestFactoryStatusDefaultBoundedBytes` (AC-LRC-006), on a pressure-on fixture and on a report value at the declared maxima.
Class: product-level
Alternate (operator options, not chosen; same worst case): (A1) candidates as a count only: cap 4 = 2,036 (within), cap 5 = 2,298 (over). (A2) rows without `mode` and `priority`, which would need a REQ-TCD-010 carve-out: cap 4 = 1,921 (within), cap 5 = 2,123 (over). (B) drop the per-row `run_id` and name the single active run once at the top: cap 3 = 1,821 (within), cap 4 = 2,057 (over), so no gain; valid only while exactly one run is active. (C) a compact JSON encoding for the default: cap 6 = 1,926 (within), cap 7 = 2,114 (over); this changes the default's formatting.
Operator verdict: DECIDED by board:d-20261015T113856Z-6f23: row cap 3 plus 2 quota candidates (worst case 1,899 B, L-13); the full output comes through `--run <id>` (Q8); alternatives A1, A2, B, and C are not adopted, and REQ-TCD-010 holds.

---

### Q6: Does the factory dispatch role-core injection have room for the added sentence?

Label: EVIDENCE-NEEDED
Authority anchor: none. The records do not decide Q6; record d-20261015T113856Z-6f23 leaves it to the run-phase measurement. The run-phase routing came from the lane's brief to the SPEC author.
Why unresolved: The role core is injected at SessionStart for lanes, and its headroom is not measured. Run-phase measurement: `internal/hook/role_composite_budget_meas_test.go`, before and after this SPEC's queue-and-factory sentence lands (plan § B.5, milestone M4). If there is no room, the placement in Q7 is decided again, with the fallback `.claude/rules/moai/core/agent-common-protocol.md`. SPEC-LANE-READ-SESSION-001 measures again for its own session sentence.

---

### Q7: Where does the lane doctrine sentence go?

Label: DECIDED
Authority anchor: board:d-20261015T113856Z-6f23 (anchor = board id, for card-scoped brain rulings under the operator-proxy standing, as record d-20261015T113856Z-6f23 states). Pinned record, verbatim from `moai decision read --scope card:t1777`:
```text
decision record: decided_by=brain 영실이 (operator proxy standing bb5e) evidence_refs=.moai/worktrees/t1777 decision-index.md 0.2.3; plan-audit-iter3.md FAIL 0.8125 blocking 1 (B1R) codex fail; d-20261015T031405Z-dfd8 ladder_path=brain ruling under bb5e; standing C run-transition (9e76/75ca) id=d-20261015T113856Z-6f23 scope=card:t1777 kind=ruling
  body: t1777 FOUNDER 행 확정: Q2 = compact todo 형식 행 상한 6 (최악 1,963B ≤ 2048, 실측 L-15), 인코딩 폭 접두 규칙(상한 20) 미채택. Q3 = factory status 기본 출력 행 상한 3 + quota 후보 2 (최악 1,899B, 실측 L-16), 전체는 --run <id>(Q8), 대안 A1·A2·B·C 미채택(REQ-TCD-010 유지). Q5 = L-A: PID 탐침, 판정 불가는 살아 있음(dfd8과 같음); JSON 기본 경로에서 REQ-SCV-006 '탐침 없음'을 뒤집는 것을 포함; 텍스트 뷰·--all·필드 집합은 REQ-SCV-006 그대로. Q9 = 범위 변경 확정: 세션 기본값은 2KB 상한 대신 '살아 있는 항목 완전·죽은 행 기본 제외', todo 기본값은 상한 없음(상한은 --compact로), factory 기본값은 Q3 상한 유지. Q7 = (a), Q8 = (a) DEFAULT-APPLIED 확정. Q6 = run 단계 측정 그대로. 앵커: 운영자 대리 standing 아래 카드 범위 브레인 ruling(dfd8과 이 기록)은 결정 보드 기록으로 인정, 앵커 = 보드 id. 진행: run 전환 — 정정 커밋 1개(B1R: windows-latest CI 잡에서 Windows 생존 판정 함수를 실제 실행하는 AC 1개 + M2 Exit 1줄 + decision-index 해당 행 DECIDED 전환, 레코드 줄 인용) → 그 커밋만 델타 1회(codex 필수 pass) → kickoff; FAIL이면 hold + t1777a·t1777b 분할.
```
Why unresolved: The sentence must reach lanes and must be mirrored in three copies. Options:
  - (a): the role-core region of `.claude/rules/moai/workflow/factory-dispatch.md`, with the identical sentence in `.moai/policies/workflow/factory-dispatch.md` and in its template copy. The file is path-scoped; lanes receive the role core at SessionStart; `TestAutoPickMirrorParity` keeps its one-line difference.
  - (b): `.claude/rules/moai/core/agent-common-protocol.md`. The file is always loaded and byte-identical with its template, so it is enrollable in the mirror guard, but every session pays the bytes.
Class: implementation-level
Default: (a) (rule: smaller user-visible surface)
Alternate: (b)
Operator verdict: DECIDED by board:d-20261015T113856Z-6f23: (a) DEFAULT-APPLIED, confirmed.

---

### Q8: Does `moai factory status --json --run <id>` keep its pre-change behavior?

Label: DECIDED
Authority anchor: board:d-20261015T113856Z-6f23 (anchor = board id, for card-scoped brain rulings under the operator-proxy standing, as record d-20261015T113856Z-6f23 states). Pinned record, verbatim from `moai decision read --scope card:t1777`:
```text
decision record: decided_by=brain 영실이 (operator proxy standing bb5e) evidence_refs=.moai/worktrees/t1777 decision-index.md 0.2.3; plan-audit-iter3.md FAIL 0.8125 blocking 1 (B1R) codex fail; d-20261015T031405Z-dfd8 ladder_path=brain ruling under bb5e; standing C run-transition (9e76/75ca) id=d-20261015T113856Z-6f23 scope=card:t1777 kind=ruling
  body: t1777 FOUNDER 행 확정: Q2 = compact todo 형식 행 상한 6 (최악 1,963B ≤ 2048, 실측 L-15), 인코딩 폭 접두 규칙(상한 20) 미채택. Q3 = factory status 기본 출력 행 상한 3 + quota 후보 2 (최악 1,899B, 실측 L-16), 전체는 --run <id>(Q8), 대안 A1·A2·B·C 미채택(REQ-TCD-010 유지). Q5 = L-A: PID 탐침, 판정 불가는 살아 있음(dfd8과 같음); JSON 기본 경로에서 REQ-SCV-006 '탐침 없음'을 뒤집는 것을 포함; 텍스트 뷰·--all·필드 집합은 REQ-SCV-006 그대로. Q9 = 범위 변경 확정: 세션 기본값은 2KB 상한 대신 '살아 있는 항목 완전·죽은 행 기본 제외', todo 기본값은 상한 없음(상한은 --compact로), factory 기본값은 Q3 상한 유지. Q7 = (a), Q8 = (a) DEFAULT-APPLIED 확정. Q6 = run 단계 측정 그대로. 앵커: 운영자 대리 standing 아래 카드 범위 브레인 ruling(dfd8과 이 기록)은 결정 보드 기록으로 인정, 앵커 = 보드 id. 진행: run 전환 — 정정 커밋 1개(B1R: windows-latest CI 잡에서 Windows 생존 판정 함수를 실제 실행하는 AC 1개 + M2 Exit 1줄 + decision-index 해당 행 DECIDED 전환, 레코드 줄 인용) → 그 커밋만 델타 1회(codex 필수 pass) → kickoff; FAIL이면 hold + t1777a·t1777b 분할.
```
Why unresolved: The quiet defaults are defined for the invocation without `--run`. Options:
  - (a): `--run <id>` keeps every card field and no row cap, which is the pre-change behavior. The existing callers of `status --json --run` stay unchanged.
  - (b): the quiet defaults also apply to `--run <id>`.
Class: implementation-level
Default: (a) (rule: preserves current behavior)
Alternate: (b)
Operator verdict: DECIDED by board:d-20261015T113856Z-6f23: (a) DEFAULT-APPLIED, confirmed.

---

### Q9: Scope change record, todo and factory half — the 2 KB bound leaves the todo default and moves to the compact form; the factory default keeps a bound.

Label: DECIDED
Authority anchor: board:d-20261015T113856Z-6f23 (anchor = board id, for card-scoped brain rulings under the operator-proxy standing, as record d-20261015T113856Z-6f23 states). Pinned record, verbatim from `moai decision read --scope card:t1777`:
```text
decision record: decided_by=brain 영실이 (operator proxy standing bb5e) evidence_refs=.moai/worktrees/t1777 decision-index.md 0.2.3; plan-audit-iter3.md FAIL 0.8125 blocking 1 (B1R) codex fail; d-20261015T031405Z-dfd8 ladder_path=brain ruling under bb5e; standing C run-transition (9e76/75ca) id=d-20261015T113856Z-6f23 scope=card:t1777 kind=ruling
  body: t1777 FOUNDER 행 확정: Q2 = compact todo 형식 행 상한 6 (최악 1,963B ≤ 2048, 실측 L-15), 인코딩 폭 접두 규칙(상한 20) 미채택. Q3 = factory status 기본 출력 행 상한 3 + quota 후보 2 (최악 1,899B, 실측 L-16), 전체는 --run <id>(Q8), 대안 A1·A2·B·C 미채택(REQ-TCD-010 유지). Q5 = L-A: PID 탐침, 판정 불가는 살아 있음(dfd8과 같음); JSON 기본 경로에서 REQ-SCV-006 '탐침 없음'을 뒤집는 것을 포함; 텍스트 뷰·--all·필드 집합은 REQ-SCV-006 그대로. Q9 = 범위 변경 확정: 세션 기본값은 2KB 상한 대신 '살아 있는 항목 완전·죽은 행 기본 제외', todo 기본값은 상한 없음(상한은 --compact로), factory 기본값은 Q3 상한 유지. Q7 = (a), Q8 = (a) DEFAULT-APPLIED 확정. Q6 = run 단계 측정 그대로. 앵커: 운영자 대리 standing 아래 카드 범위 브레인 ruling(dfd8과 이 기록)은 결정 보드 기록으로 인정, 앵커 = 보드 id. 진행: run 전환 — 정정 커밋 1개(B1R: windows-latest CI 잡에서 Windows 생존 판정 함수를 실제 실행하는 AC 1개 + M2 Exit 1줄 + decision-index 해당 행 DECIDED 전환, 레코드 줄 인용) → 그 커밋만 델타 1회(codex 필수 pass) → kickoff; FAIL이면 hold + t1777a·t1777b 분할.
```
Why unresolved: The parent's 0.1.0 SPEC bounded the todo default and the factory default at 2048 bytes. The record keeps the full todo default without a bound and moves the bound to the new `--compact` form (Q1, option b; REQ-LRC-001, REQ-LRC-003), and the factory default keeps a bound with the row cap of Q3 (REQ-LRC-005, REQ-LRC-006). The session half of this row is in SPEC-LANE-READ-SESSION-001.
Class: product-level (changes the default user-visible output of a shipped command: the todo default is no longer bounded, and the factory default is abbreviated)
Operator verdict: DECIDED by board:d-20261015T113856Z-6f23: scope change confirmed: the todo default has no cap (the cap is on `--compact`); the factory default keeps the Q3 cap.

Version: 0.1.0 (2026-10-15). Carried from SPEC-LANE-READ-VERBS-001 decision-index 0.2.4 by the split of ruling d-20261015T122748Z-6d6b.
