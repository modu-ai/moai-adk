# SPEC-WEB-CONSOLE-017 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (Tier M 세트; `design.md`/`research.md` 미산출 — Tier L 전용).
- Status on creation: `draft`. No production code written in plan phase.
- Requirement IDs: REQ-WC-017-001 .. REQ-WC-017-006 (수집-가능 형태). Acceptance IDs: AC-WC17-001 .. AC-WC17-005.
- Evidence base: 본 레인의 재측정 — 트리 `WT-save-observability` @ `616f7451d` (로컬 develop 흡수 직후; 흡수 커밋 21건 중 `internal/web` 접촉 0건으로 배차 전제 좌표 안정 확인). CDP 콘솔 에러 수·본문 크기는 카드 분석의 전제로 표기(재측정 안 함 — spec.md §1.1, §6-1).
- Open clarifications blocking Implementation Kickoff Approval: **0건.** `banner--error` 신설 여부는 운영자 대기 질문이지만 본 SPEC 은 그 결정에 의존하지 않는다(plan.md §F.1, spec.md HARD-6) — Kickoff 를 막지 않는 것이 의도다.
- Run-phase entry is NOT approved. M1(REQ-A 수송 기구)은 사용자 가시 동작을 바꾸므로 Kickoff 게이트의 운영자 판정 대상이다.

## §E.2 Run-phase Evidence

Run-phase executed TDD (RED-GREEN-REFACTOR), 4 milestones, tree `WT-save-observability`. Baseline attribution: all outputs below observed in this run against this tree (HEAD per row in §E.3's `run_commit_sha`; each milestone's own SHA in the commit list). Lane rule respected: `go test ./...` never run locally — package-scoped only, full suite is CI's.

### M1 — REQ-A transport (commit `be94bb36e`)

- **RED-1 (E8 evidence)** — `go test -run TestSaveFailureReasonReachesInlineSlot ./internal/web/`:
  ```
  --- FAIL: TestSaveFailureReasonReachesInlineSlot (0.02s)
      save_observability_test.go:76: failed-save status = 500, want 200 (htmx discards non-2xx boosted bodies, so the phrase never reaches the slot)
  FAIL
  ```
- **GREEN-1** — `renderErrorPage` answers 200; transport mechanism = 2xx re-render (rationale recorded in the M1 commit body per plan.md §D). Existing tests whose status assertion encoded the old 500 contract updated to 200 (status assertion only): `partial_apply_repro_test.go`, `handlers_test.go` (sync failure), `projectconfig_handler_test.go`, `projectnested_error_test.go`. The GET-read-error 500 (`TestIndexReadErrorRendersInlineError`) and profile-CRUD 500s are different paths, out of scope (§F), left untouched.
- **Status transition** — spec.md frontmatter `draft → in-progress` performed on the M1 commit (manager-develop's single owned transition).

### M2 — REQ-B stderr (commit `7a2bf81b9`)

- **RED-2 (E8 evidence)** — `go test -run TestSaveFailureStderrLog ./internal/web/`:
  ```
  save_observability_test.go:155: save-failure stderr lines = 0, want exactly 1 (0 undercounts; 2+ violates REQ-WC-017-003's exactly-one)
  --- FAIL: TestSaveFailureStderrLog/failure_emits_exactly_one_prefixed_line
  ```
  (the success-emits-zero subtest passed already — that direction was true pre-change)
- **GREEN-2** — `logSaveFailure(seam, phrase)` helper via the established `fmt.Fprintf(os.Stderr, ...)` idiom; wired at all 9 §5.1 call sites; `err.Error()` never reaches the line.

### M3 — harness extension + seam coverage (commit `0cb60f263`)

- **RED-3 (E8 evidence)** — compile failure on the missing seams (`go vet ./internal/web/`):
  ```
  vet: internal/web/save_observability_test.go:206:3: undefined: stepApplyPerfTier
  ```
- **GREEN-3** — the three package-level calls (`applyPerfTierEdits`, `glmcred.Save`, `jevcred.Save`) promoted to app fields (calls, not implementations; default wiring byte-identical), `recordingSeams` extended with their recorders — signature preserved (HARD-2). `reproForm` now submits newline-free test keys so the fixture reaches glmcred.Save/jevcred.Save; recordingSeams always wires both, so tests never touch the real credential writers. Positive control asserts the full nine-step order.

### M4 — guards (commit `6640003df`)

- AC-WC17-004 sentinel: sentinel key material asserted absent from stderr AND the inline response at every one of the 9 seams. Classification note (acceptance.md §D.1): could not be RED before M2 existed (no stderr surface to leak into); exercised per seam now that both surfaces exist.
- AC-WC17-005 success-surface guard: banner + saved state + no error slot + zero stderr lines.

### AC Matrix (E1) — final re-measurement

| AC | Status | Verification command | Observed result |
|----|--------|---------------------|-----------------|
| AC-WC17-001 | PASS | `go test -run TestSaveFailureReasonReachesInlineSlot ./internal/web/` | PASS (pre-submit absence + post-failure presence in slot, 2xx) |
| AC-WC17-002 | PASS | `go test -run TestSaveFailureSeamCoverage ./internal/web/` | PASS (9/9 seam subtests, both surfaces) |
| AC-WC17-003 | PASS | `go test -run TestSaveFailureStderrLog ./internal/web/` | PASS (exactly-one prefixed line + success zero lines) |
| AC-WC17-004 | PASS | `go test -run TestSaveFailureNeverLeaksKeyMaterial ./internal/web/` | PASS (9/9 seam subtests, sentinel absent both surfaces) |
| AC-WC17-005 | PASS | `go test -run TestSaveSuccessSurfaceUnchanged ./internal/web/` | PASS (banner/saved-state/no-error-slot/zero-lines) |

Final package run (all of the above together + full package): `go test ./internal/web/` → `ok github.com/modu-ai/moai-adk/internal/web` (exit 0, observed 3× across M1b/M3/post-M4).

### Boundary + lint (E4/E5)

- E4: `grep -rn 'AskUserQuestion|mcp__askuser' internal/web/` excluding `_test.go` and comments → 0 matches (exit 1).
- E4 (HARD-3): sentinel sweep above is the stronger instrument — text-level key-material absence observed on both surfaces, per seam.
- E5: `golangci-lint run --timeout=2m ./internal/web/...` → `0 issues.` (baseline was `0 issues.` pre-change; NEW = 0).
- `go vet ./internal/web/` → exit 0.

### Coverage (E3)

`go test -coverprofile=... ./internal/web/` → package 67.9% overall; changed paths: `handleSave` 97.0%, `renderErrorPage` 100.0%, `logSaveFailure` 100.0%, `successProjectView` 100.0%, `rejectedProjectView` 100.0% — ≥ 85% on every changed path (package-wide figure dominated by pre-existing uncovered render/monitor surfaces, unchanged by this SPEC).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-22
run_commit_sha: "6640003df"
run_status: implemented-run-phase-complete
ac_pass_count: 5
ac_fail_count: 0
preserve_list_post_run_count: 7
l44_pre_commit_fetch: "not-applicable — lane never pushes; lead batch-pushes origin/develop (2026-09-02 operator directive)"
l44_post_push_fetch: "not-applicable — same lane discipline"
new_warnings_or_lints_introduced: 0
cross_platform_build:
  host: "exit 0 (go build ./...)"
  windows: "exit 0 (GOOS=windows GOARCH=amd64 go build ./...)"
total_run_phase_files: 9
m1_to_mN_commit_strategy: "one commit per milestone (M1 be94bb36e transport / M2 7a2bf81b9 stderr / M3 0cb60f263 harness+guard / M4 6640003df sentinel+success guards), each with verbatim RED evidence captured before its GREEN; progress.md §E refresh follows as the final run-phase commit"
```

Preserve-list post-run audit (HARD-2/5/6): `recordingSeams` seam list preserved, signatures unchanged (extended by 3 recorders); `server.go:252` `web: ` site untouched; `root.templ:26-28` single-accent design decision untouched (no `banner--error` introduced); 9-seam order/count/existence unchanged (HARD-1 — order asserted by the positive control); failure phrases remain Go literals (HARD-7).

Spec-lint REQ collection check (DoD 6): `reqLinePattern` (`internal/spec/lint.go:678`) = `-\s+(REQ-[A-Z]{2,5}-\d{3}-\d{3})\s*:\s*(.+)`; `grep -cE '^- REQ-WC-017-[0-9]{3}:' spec.md` = **6** — the lint green is non-vacuous for this SPEC.

## §E.4 Sync-phase Audit-Ready Signal

### B12 CHANGELOG 방출 자가검증 3건

| # | 검사 | 커맨드 | 관측 |
|---|------|--------|------|
| a | 중복 방출 차단 | `grep -c 'SPEC-WEB-CONSOLE-017' CHANGELOG.md` | 방출 **전** `0` (exit 1) — 기존 항목 없음, 방출 진행. 방출 **후** `1` |
| b | AC 개수 일치 | `grep -oE 'AC-([A-Z0-9]+-)*[0-9]+b?' acceptance.md \| sort -u` | **5행** — `AC-WC17-001..005` (release-blocking 3 + regression-guard 2). 예약 토큰(`[RETIRED]`/`[REF]`) 0건. CHANGELOG 문안이 같은 수(5)를 적는다 — 0행이 아니므로 공허 비교가 아니다 |
| c | 파일 경로 실재 | `ls internal/web/handlers.go internal/web/app.go internal/web/save_observability_test.go internal/web/shell.templ .moai/specs/SPEC-WEB-CONSOLE-017/spec.md` | 5개 전부 존재. CHANGELOG 항목이 이름을 대는 경로는 이 다섯뿐이다 |

### 문서 판정 — README · docs-site 변경 없음 (근거 있는 결정, 생략이 아니다)

**결정: 변경하지 않는다.** 근거 셋:

1. **사용자 절차가 변하지 않았다.** 본 카드는 `moai web` 콘솔 저장-실패의 관측 표면(인라인 사유 도달 + stderr 로그)만 만진다 — CLI 커맨드·플래그·출력 형식 변경 없음, 콘솔 사용 절차 동일.
2. **문서 대상 신규 기능이 아니다.** 실패 시 화면에 사유가 보이게 된 것은 기존 실패 슬롯의 수송 수리다. stderr 행은 유지보수자 진단 표면이며 README/docs-site 독자 대상이 아니다.
3. **소급 약속 금지.** 실패 문구의 i18n 화는 spec.md §F 가 명시적 범위 밖(HARD-7) — 문서에 새 문구를 적으면 하지 않는 일을 약속하게 된다.

CHANGELOG `[Unreleased] > Fixed` 항목 하나가 본 카드의 유일한 문서 산출물이다.

### MX 태그 (sync 서브스텝)

- `@MX:ANCHOR` 신설 2건 (각 `@MX:REASON` 하위 행 동반):
  - `logSaveFailure` — `internal/web/handlers.go:675`, fan-in 9 (handleSave 의 9개 persistence seam 이 전부 이 한 줄로 수렴).
  - `renderErrorPage` — `internal/web/handlers.go:661`, fan-in 9 (9개 seam 의 오류 경로 전부). 본 SPEC 이 그 계약(500 → 2xx)과 doc comment 를 다시 썼으므로 sync MX 점검에서 같은 MUST 클래스로 보정.
- `handlers.go` 기존 ANCHOR 0건이므로 per-file 상한 3 이내. 그 외 신규 표면(`app.go` 주입 seam 필드 3개 — struct 필드, `save_observability_test.go` — 테스트 전용)은 프로토콜 의무 클래스 해당 없음, 태그 미부여.

### 상태 전이

단일 sync 커밋이 `spec.md` frontmatter 의 `in-progress → implemented → completed` 를 실어 3-phase close 를 닫는다(별도 Mx 커밋 없음). `updated:` 는 이미 sync 커밋 날짜(2026-09-22)와 동일해 값 변경 없음. **`status:` 외에는 어떤 frontmatter 필드도, 어떤 본문 섹션도 건드리지 않았다** — `spec.md` · `plan.md` · `acceptance.md` 본문은 sync 단계 금지 표면이다.

### Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-22
sync_commit_sha: "b15dbd57e"   # sync 커밋 실측값 — D3 백필 (후속 커밋에서 채움; 커밋은 자기 해시를 인용할 수 없다)
sync_status: audit-ready
b12_self_test_a: pass                 # 중복 grep: 방출 전 0 (exit 1) / 방출 후 1
b12_self_test_b: pass                 # AC 5건 (acceptance.md §D.1 이 SSOT; 예약 토큰 0건)
b12_self_test_c: pass                 # 인용 경로 5개 전부 ls 확인
changelog_entry_position: "[Unreleased] > Fixed, 최상단"
changelog_entry_count: 1
mx_tags_added:
  - "@MX:ANCHOR logSaveFailure (internal/web/handlers.go:675)"
  - "@MX:ANCHOR renderErrorPage (internal/web/handlers.go:661)"
frontmatter_status_transitions:
  spec_md: in-progress -> implemented -> completed   # 단일 sync 커밋
  updated_refreshed: false             # 이미 2026-09-22 — 값 동일
```

## §F Phase 4 Mode Selection

**입력값**: tier M · 범위 약 5-8파일(internal/web 내 Go/templ/CSS + 테스트) · 도메인 1개(웹 콘솔) · 언어 혼합 Go+templ+CSS · 병렬 이득 낮음(coding-heavy) · agent-team 사전요건: 미요청.

| 모드 | 선택 | 근거 |
|------|------|------|
| direct | 미선택 | 의미 있는 다중 파일 변경 — 단일 오타 수정 아님 |
| serial | **선택** | 한 패키지 안의 결합된 변경, coding-heavy — Anthropic caveat 상 serial 기본 |
| fanout | 미선택 | coding-heavy 병렬 위임 금지(caveat) + 동일 워크트리 다중 작성자 금지(한-작성자 규율) |
| sweep | 미선택 | ≥30파일 기계 변환 아님 |

**결정**: serial

**정당화**: REQ-A(전송)와 REQ-B(stderr)가 같은 요청 경로와 같은 패키지를 공유해 단일 구현 흐름이 자연스럽고, 워크트리 단일-작성자 규율상 쓰기 가능 스폰은 한 번에 하나다. 단일 manager-develop 위임(배경 실행)으로 마일스톤을 순차 소진한다.

## §G Implementation Kickoff Approval Record

- **2026-09-22 06:2x** — 1차 상신(`AskUserQuestion`, 진행 방식 축 포함) → 60초 무응답. **2026-09-22 06:4x** — 2차 상신 → 60초 무응답(재시도 없음 원칙 적용). 무응답은 승인이 아니므로 두 차례 모두 run 미진입.
- **2026-09-22 리드 세션 경유 운영자 승인 수령.** 운영자 전역 지시(2026-09-22) 원문: 「각 레인 체크해서 남은 카드 모두 배차해서 완료하고 완료된 카드는 상태 체크 후 로컬 develop 병합 완료」 — 리드가 본 레인의 Kickoff 대기 상태를 보고한 뒤 하달된 지시라고 리드가 명시.
- **승인 성격 고지**: 본 레인 터미널의 직접 답변이 아니라 리드 중계다. 중계 경로(lead → lane agent-20)와 지시 원문을 이 기록으로 보존하며, push 금지(리드 일괄)·워크트리 폐기 금지·banner--error §F.1 비종속(HARD-6)은 그대로 유지한다.
- **진행 방식**: 반자율(경계 보고) — `/moai goal` 미무장, 배경 위임으로 run 진행, 마일스톤 경계마다 보고.
- **Phase 1 plan-audit skip 기록**: verdict PASS(iter-2, be3b0627c) + 1.00 ≥ Tier M 임계 0.80 + plan-artifact 해시 주체 6종(spec/plan/acceptance/design/research/tasks) 불변 — 본 기록·§F 편집은 progress.md 라 해시 주체가 아니다. 3조건 충족으로 Phase 1 재실행 skip.
