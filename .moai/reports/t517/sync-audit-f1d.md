# Sync-Audit (F1-D limited) — SPEC-WEB-WRITE-SAFETY-001 (card t517)

- 감사자: sync-auditor (독립 판정)
- 감사 시점: 2026-09-07, 트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t517` (HEAD `6639e7ecb`)
- 범위: F1-D 한정 — 렌더 극성 1식 수리 + 전체 경로 루프 가드 + RED 진위 (리드 지정 1순위). F3-D/F4-D/F5-D 동반 확인.
- 선수: `.moai/reports/t517/sync-audit-delta.md` (FAIL — F1-D) → 본 재감사는 그 결함의 수리만 판정한다.

---

## Evaluation Report (F1-D limited)

SPEC: SPEC-WEB-WRITE-SAFETY-001
Overall Verdict: **PASS** (F1-D 수리 성립 — 기능 결함 0. 기록·가드 위생 소정리 2건은 병행 수리 요건, 아래 F1-f/F2-f)

### Dimension Scores

| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%) | 1.00 | PASS | 렌더 1식 확인(생성물 `fieldsets_templ.go:925`에 수리식 존재) + 실험 A(커밋 가드 GREEN) + 실험 B(렌더-only 뮤턴트 RED — 가드 이 ±) + 전 타깃 스위트 `ok` ×7 (`go test -count=1 ./internal/settings/... ./internal/web/... ./internal/config/...`, 본 런 본 트리) |
| Security (25%) | 1.00 | PASS | 델타는 보안 표면 불변 — 렌더 식만 변경 |
| Craft (20%) | 0.75 | PASS (소정리 2건) | `golangci-lint run` → `0 issues.`, `gofmt -l` 청결(F4-D 수리 확인). 단 가드 단언 약화 1건(F1-f)과 RED 로그 세대 불일치(F2-f) — 비차단, 기록/테스트 위생 |
| Consistency (15%) | 0.75 | PASS | F5-D 트리 귀속 노트 반영 확인. RED 로그 세대 표기 정정 필요(F2-f) |

가중 평균 91.3.

### 1순위 판정 — 가드 RED의 진위 (리드 지정)

**결론: 가드의 이는 실재이고 수리는 유효하다. 다만 커밋된 RED 로그는 최종 가드 세대의 채득이 아니라 초안(draft) 세대의 것이다 — 기록 정정이 필요하다.**

관측 근거 (모두 본 런, 본 트리):

- **실험 A — 연속 관측 (GREEN)**: 커밋된 가드 `TestHandleSaveUntouchedRenderedBodyLeavesTrackedConfigByteIdentical`을 HEAD에서 실행 → `--- PASS` (0.02s). GET /settings의 실제 렌더에서 absent 키의 checked 라디오 값("1")을 읽어 그대로 재제출 → workflow.yaml byte-identical. 무접촉 저장 불변이 수리 트리에서 성립.
- **실험 B — 렌더-only 뮤턴트 (RED)**: `fieldsets_templ.go:925`의 수리식을 이전 식(`== "true"`)만으로 되돌리는 오버레이 뮤턴트 → 동일 가드 즉시 RED:
  ```
  write_safety_test.go:158: renderer shows no checked radio for the absent default-ON key — expected ON (value=1) after the polarity repair
  --- FAIL: TestHandleSaveUntouchedRenderedBodyLeavesTrackedConfigByteIdentical (0.01s)
  ```
  렌더만 어긋나도 가드가 잡는다 — 리드가 요청한 "한쪽만 되돌리는 뮤턴트" 재확인 충족.
- **실험 C — 결함 본체 독립 재관측 (전 세션, `a8042c630`)**: 렌더가 산출하는 무접촉 본문(`__present=1` + `enabled=`) POST → `todo: enabled: false` 생성 — 커밋된 RED 로그의 before/after 덤프와 **바이트 단위로 동일한 형상**.

**RED 로그(`RED-f1d-render-polarity.log`)의 세대 판별**: 로그의 실패는 158행 `untouched rendered submission rewrote the tracked section file`(byte-diff Errorf)인데, 커밋된 가드의 158행은 `t.Fatal("renderer shows no checked radio...")`이고 Errorf는 그보다 뒤에 있다. 또한 커밋된 가드 코드로 수리 전 트리(a8042c630 — 렌더가 absent를 OFF로 그리던 상태)를 돌리면 step-1 fatal이 Errorf보다 먼저 발화하므로(실험 B가 그 형태를 관측), **커밋된 가드로는 이 로그를 재현할 수 없다**. 로그는 step-1 fatal이 없던 초안 가드의 채득이다. 로그 자체에는 커맨드·트리 SHA가 없어 채득 상태를 내용만으로 인증할 수 없다(VCI §2.1 undecidable).

**처분**: 결함과 가드의 이는 독립 이중 관측(실험 C = 초회 델타 감사 실험, 실험 B = 커밋 가드의 RED 형태)으로 완성됐으므로 기능 판정에는 영향 없음. 단 채택 기록의 RED 셀은 커밋 가드 세대로 정정해야 한다 — **F2-f [minor, 병행 수리]**: (a) `RED-f1d-render-polarity.log`를 실험 B 절차(렌더-only 뮤턴트 오버레이 → 커밋 가드)로 재채득해 교체하거나, (b) progress.md의 해당 RED 셀에 "초안 가드 세대 채득 — 커밋 가드의 수리 전 RED는 158행 fatal(렌더-only 뮤턴트로 재현 가능)"을 명기. 실험 B의 출력이 그대로 재현 절차가 된다.

### 동반 확인 (F3-D / F4-D / F5-D)

- **F3-D — 절반 성립 + 신규 소정리 1건**. fixture에서 `pre_commit` 블록 제거 확인. 본 감사가 MUTANT-D(absent 분기 `==`→`!=` 반전)를 새 fixture에 재적용: `TestApplySchemaEditsSeamAbsentKeyFalseIsNoOp` → **RED** — default-off absent 분기가 이제 실검증된다(이전 라운드에는 fixture에 키가 있어 뮤턴트가 무시됐다). **그러나 F1-f [minor]**: 같은 뮤턴트에서 `TestApplySchemaEditsSeamAbsentKeyTrueStillWrites`가 **PASS로 남는다** — fixture 상위의 `gate.enabled: true`(testdata/sections/gate.yaml:6)가 `strings.Contains(after, "enabled: true")` 단언을 충족해 기록이 스킵돼도 통과하는 공허 통과. 단언을 pre_commit 블록 범위로 좁혀야 한다(예: `pre_commit:` 섹션 파싱 또는 `enabled: true` 앞뒤 문맥 확인).
- **F4-D ✓**: `gofmt -l internal/settings/ internal/web/` → 무출력(청결).
- **F5-D ✓**: progress.md에 F1 RED-first의 트리 귀속 노트 반영("커밋 `43f9ab202` + 미커밋 극성 선언 편집, 동작 게이트 미수리, working-tree 측정으로 트리 소멸 — 상태 기술로 SHA 대체") — 중간 상태 특성상 정직한 기술이며 채택.

### Findings (delta)

- **F1-f [minor] [병행 수리]** `internal/settings/write_safety_test.go` `TestApplySchemaEditsSeamAbsentKeyTrueStillWrites` — fixture 변경 후 단언이 `gate.enabled: true`(상위 키)로 공허 충족될 수 있음(MUTANT-D에서 PASS 실측). 수리: 단언을 pre_commit 블록 스코프로 한정.
- **F2-f [minor] [병행 수리]** `RED-f1d-render-polarity.log` + progress.md 해당 RED 셀 — 초안 가드 세대 채득(커밋 가드로는 재현 불가; 실험 B 형태가 커밋 가드의 수리 전 RED). 수리: 재채득 또는 세대 명기.
- (기능 결함 0. 초회 감사의 F5/F6/F7 advisory·optional은 변화 없음 유지.)

### Gaps (관측하지 못한 것)

- 실험 B의 뮤턴트 RED를 로그 파일로 재채득하지 않았다(본문에 verbatim 기재 — 파일 교체는 레인 몫).
- `origin/develop` CI 전체 판정 미관측 — 리드 push 소관.
- mcpToolRow 경로의 무접촉 저장("1" → skip)은 코드 판독 + 게이트 skip 테스트(c)로 확정(실행 실험 아님 — 이전 감사와 동일하게 유지).

### Residual Risk

- 없음(기능축). 기록 축: F2-f 정정 전까지 RED 셀의 세대 표기가 불일치한다.

---

## 판정 요약

F1-D의 렌더 1식은 생성물까지 반영됐고, 무접촉 저장 불변이 실렌더→파서→게이트 전체 루프 가드로 GREEN 전환됐으며, 렌더-only 뮤턴트에 가드가 즉시 RED — 이가 실임을 본 감사가 독립 관측했다. RED 로그는 초안 세대 채득으로 커밋 가드와 세대가 어긋나지만(F2-f), 결함의 관측 자체는 독립 이중 확인돼 있다. **PASS** — F1-f·F2-f 소정리는 병행 수리 요건으로 창 요청을 막지 않는다.
