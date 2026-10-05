# Card t835 판정서 — t657 병합 후 검증 FAIL(1건)의 원인 판정과 검증기 수리

- Date: 2026-09-13 · Lane: t835 전용 세션 · Branch: `WT-queue-verifier` (기점 develop `d4c45bdcc`)
- 선행 판정: `.moai/reports/t657-queue-merge/verdict.md` + `SPEC-TODO-QUEUE-HOME-MERGE-001` progress.md §M4
- 결론 한 줄: **데이터 결함 없음 — 검증기가 duplicate 흡수 경계를 모르던 결함. 수리로 RED→GREEN. t204/t718 동일 본문 쌍은 의도된 결과(재확인).**

## Claim

1. `project archived finding t204->t538 not rewritten` FAIL의 원인은 검증기 결함이다 — 흡수(duplicate)로 의도적으로 미반영된 항목의 finding까지 재작성 대상으로 세었다.
2. 수리(검증기 판정식 1개 추가)로 RED→GREEN을 실측했다.
3. 홈 원본 t204(queued)와 재번호 이관본 t718(queued)의 동일 본문 공존은 REQ-TQM-006 v2 판별식의 정상 작동이다.

## Evidence (이번 세션 실측, 전부 읽기전용)

**실데이터 대조 (sqlite3 -readonly):**

| 질의 | 결과 |
|---|---|
| 병합 후 홈 스토어(`~/.moai/db/moai-adk-go-1bd3d038/todo/backlog.db`) `archived_findings` | seq 217, 호스트 **t538**, `t204→t538 contains` — **미재작성 원본 그대로 존재** |
| 프로젝트 백업(store-1) seq 217 vs 홈 백업(store-2) seq 217 | subject·related·relation·source·score·**note·at까지 바이트 동일** |
| store-1·store-2 양쪽 `archived_items` t538 | 동일 seq(217)·동일 state(picked)·동일 본문 → 병합 시 프로젝트 t538은 duplicate 흡수(항목+finding 통째로 미반영 — `todo_queue_merge.go` 296-308행의 설계된 동작) |
| `id-mapping.tsv` | 127행(헤더+126), 11행 `t204→t718` 존재, **`t538` 0행(부재)** — 재번호 대상 아님 |
| 라이브 홈 `items` | `t204` queued + `t718` queued, 동일 본문 공존 |

**원인 기제:** 검증기(`VerifyMergedRecord`)의 archived finding 루프는 host 항목의 처분을 보지 않고 `touched(subject=t204 ∈ 매핑)`만으로 대상화 → 재작성 튜플(t718→t538)이 merged에 없다고 판정. 그러나 host t538은 duplicate 흡수로 미반영이 **의도**이고, 홈 사본이 동일 finding을 보존하며 홈은 참조된 두 id(t204 원본·t538)를 모두 유지하므로 참조는 유효 — 판정 오류.

**RED (`go test ./internal/kanban/ -run TestVerifyMergedRecordBranches -count=1`):** 신설 서브테스트 `duplicate-absorbed archived finding survives via the home copy` → `todo_merge_procedure_test.go:431: duplicate-absorbed host's finding flagged stale despite the identical home copy: [project archived finding t204->t538 not rewritten]` — `/tmp/l4m_execute.json` 오류 원문과 동일 메시지 재현.

**GREEN (동일 명령):** 9/9 서브테스트 PASS. 패키지 전체 `go test ./internal/kanban/ -count=1` → `ok ... 177.696s` (exit 0). `go vet ./internal/kanban/ ./cmd/t657-merge/` → exit 0.

## 변경 파일 (lane-4 t658 영역 겹침 대비 명시)

- `internal/kanban/todo_merge_procedure.go` — `duplicated` 집합(report.Duplicates) 구축 + archived finding 판정식 1개: host가 duplicate 흡수이고 홈 사본이 동일 finding을 보유하면 satisfied. 홈 사본에 없으면 여전히 FAIL(loss-arm — 유실을 계속 잡음).
- `internal/kanban/todo_merge_procedure_test.go` — RED/loss-arm 서브테스트 2건 추가.
- 실데이터(홈 스토어·백업) 쓰기 0건 — 전 과정 읽기전용.

## Baseline-attribution

- 측정 주체: 본 워크트리(`.claude/worktrees/t835`, branch `WT-queue-verifier`)에서 이번 세션이 실행한 명령 + 관측 출력.
- 코드 기점: develop `d4c45bdcc`. 위 표의 sqlite 질의와 테스트 출력은 전부 본 세션의 실측값.

## Gaps

- **전수 대조 미실행**: 수리된 검증기를 과거 병합 결과 전체에 재적용하지 않았다(t657 판정서 Gaps의 "다른 참조 영향 가능성" — 별도 후속 카드 소관으로 유지).
- 246건 duplicates 목록 자체는 파일로 미보존돼 있어(도구가 count만 JSON 출력) t538 ∈ duplicates는 "매핑 부재 + 양 백업 내용 동일 + 병합 후 홈에 단일 t538"의 정의적 귀속으로 판정(간접 3중 확인).
- M5(프로젝트 스토어 정리) 미실행 — 별도 승인 대기(변동 없음).

## Residual-risk

- loss-arm이 잡는 보장은 **미래 병합**에 적용된다. 과거 병합에서 duplicate 흡수로 홈 사본에 없던 finding이 유실됐을 가능성은 본 카드가 배제하지 못함(전수 대조 후속 소관).
- `mergeCardContentEqual`은 카드 본문 동등성만 비교하므로, 본문 동일·finding 상이한 쌍은 앞으로도 duplicate로 흡수된다 — 다만 이제 검증기가 유실을 FAIL로 잡으므로 조용히 넘어가지는 않는다.
