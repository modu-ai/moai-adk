# SPEC-LANE-READ-COMPACT-001 — Decision anchor (verbatim copies of the board rulings for card t1810)

Source: `moai decision read --scope card:t1777`, run from the worktree root on 2026-10-15 (first line `board=ok unparseable=0 records=100`). Each record line and the body line that follows it are copied without edits. The rulings are card-scoped brain rulings under the operator-proxy standing; record d-20261015T113856Z-6f23 sets their anchor to the board id.

## Verbatim record and body: d-20261015T031405Z-dfd8

```text
decision record: decided_by=brain 영실이 (operator proxy, directive 2026-10-15) evidence_refs=lane-23 plan draft SPEC-LANE-READ-VERBS-001 open decisions Q1/Q4; frozen goldens REQ-TCL-014, AC-TST-012, AC-TSP-021 ladder_path=② id=d-20261015T031405Z-dfd8 scope=card:t1777 kind=ruling
  body: t1777 Q1 = (b): moai todo list --json default stays full (frozen goldens untouched); add opt-in --compact; the effect lives in consumers — lane cycle (--auto path), lane guidance text and factory lane prompts default to --compact, and that consumer change is an AC (a flag nobody uses misses the card's intent). Q4 = S-2: session list default shows live sessions only, cap exempt; dead rows behind --all (or equivalent). Liveness by PID probe (kill -0 etc.), not registry row presence; indeterminate counts as live. The AC change from the 2KB cap to 'live entries complete, dead rows excluded by default' is a scope change: record it in decision-index with this decision id and rewrite the AC. Tier unchanged (not S-3).
```

## Verbatim record and body: d-20261015T113856Z-6f23

```text
decision record: decided_by=brain 영실이 (operator proxy standing bb5e) evidence_refs=.moai/worktrees/t1777 decision-index.md 0.2.3; plan-audit-iter3.md FAIL 0.8125 blocking 1 (B1R) codex fail; d-20261015T031405Z-dfd8 ladder_path=brain ruling under bb5e; standing C run-transition (9e76/75ca) id=d-20261015T113856Z-6f23 scope=card:t1777 kind=ruling
  body: t1777 FOUNDER 행 확정: Q2 = compact todo 형식 행 상한 6 (최악 1,963B ≤ 2048, 실측 L-15), 인코딩 폭 접두 규칙(상한 20) 미채택. Q3 = factory status 기본 출력 행 상한 3 + quota 후보 2 (최악 1,899B, 실측 L-16), 전체는 --run <id>(Q8), 대안 A1·A2·B·C 미채택(REQ-TCD-010 유지). Q5 = L-A: PID 탐침, 판정 불가는 살아 있음(dfd8과 같음); JSON 기본 경로에서 REQ-SCV-006 '탐침 없음'을 뒤집는 것을 포함; 텍스트 뷰·--all·필드 집합은 REQ-SCV-006 그대로. Q9 = 범위 변경 확정: 세션 기본값은 2KB 상한 대신 '살아 있는 항목 완전·죽은 행 기본 제외', todo 기본값은 상한 없음(상한은 --compact로), factory 기본값은 Q3 상한 유지. Q7 = (a), Q8 = (a) DEFAULT-APPLIED 확정. Q6 = run 단계 측정 그대로. 앵커: 운영자 대리 standing 아래 카드 범위 브레인 ruling(dfd8과 이 기록)은 결정 보드 기록으로 인정, 앵커 = 보드 id. 진행: run 전환 — 정정 커밋 1개(B1R: windows-latest CI 잡에서 Windows 생존 판정 함수를 실제 실행하는 AC 1개 + M2 Exit 1줄 + decision-index 해당 행 DECIDED 전환, 레코드 줄 인용) → 그 커밋만 델타 1회(codex 필수 pass) → kickoff; FAIL이면 hold + t1777a·t1777b 분할.
```

## Verbatim record and body: d-20261015T122748Z-6d6b

```text
decision record: decided_by=brain 영실이 + Jev(yes 0.68) via leader 5a9c6c18 evidence_refs=t1777 plan-audit-iter4.md FAIL 0.8125 blocking 3 (B1R2, B2, B3), codex fail rcpt-9dda0cf60e478c9ce0db1e6e; 6f23 ladder_path=brain ruling (Jev agree) id=d-20261015T122748Z-6d6b scope=card:t1777 kind=ruling
  body: t1777 = hold + 위험 기준 분할: t1810(a) = todo --compact(Q1·Q2) + factory status 기본값(Q3·Q8), Windows 없음·저위험; t1811(b) = session list 생존 판정(Q4·Q5, Windows 탐침, REQ-SCV-006 반전), B1R2(AC-LRV-014 OpenProcess seam + '알 수 없는 오류 = 판정 불가 → 항목 유지' 4번째 케이스)와 B3(DoD에 AC-LRV-014) 수리. B2는 두 plan 모두 한 줄: '운영자 대리 standing 아래 카드 범위 브레인 ruling의 앵커 = 보드 id(dfd8, 6f23); moai decision read 다이제스트 결함은 후속 카드 t1809'. t1809(Class A, 다이제스트 출력)는 제안 묶음에서 앞당겨 발행.
```
