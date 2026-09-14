---
spec: SPEC-CLI-TUX-RENDER-I18N-001
card: t756
phase: sync
plan_status: audit-ready (draft 0.1.1)
updated: 2026-09-14
---

# progress — SPEC-CLI-TUX-RENDER-I18N-001

## §A Plan-phase 산출물

- `spec.md` 0.1.0 (draft) — GEARS REQ-TRI-001~008, Out of Scope 3건
- `plan.md` — 재고정표(§B), Pre-flight(§C), M1-M4
- `acceptance.md` — AC-TRI-001~009 (전 AC 캡처 프레임 합격 기준)

## §B Plan-phase 결정 기록

- D1 (카드 위임 결정 "v1 흡수 vs 최소화"): **흡수 유지 권고** — t586 이 흡수를 완료했고 v1 코드·의존이 트리에 없음(§D 항목 3·4). Kickoff 게이트에서 최종 확인 대상.
- D2: 카드 6판정 중 5건이 선행 SPEC-INIT-TUX-I18N-001(completed, 베이스 병합됨)로 전달됨 → 본 SPEC 을 잔여 수리 + 가드 + 재검증으로 범위 확정. REQ-TRI-008(검증만 종결)로 과잉 수리 방지.
- D3: 확인 필드 내부 고정 빈 줄의 기본 수리 경로는 로컬 확인 필드 래퍼(huh.Field 구현, 자체 View 조합)로 판독. 라이브러리 인라인 모드(huh v2.0.3 `field_confirm.go:136`)는 구조적 부적합(`:254` 제목↔설명 개행 제거, `:293` 버튼 앞 개행 부재)이라 M1 측정 참고로만 관찰 — plan-audit 1차 D1 반영.

## §C Plan-phase 조사 기록 (file:line)

- 버튼 정렬: huh v2.0.3 `field_confirm.go:53`(기본 Center), `:361`(WithButtonAlignment). 우리 생성 지점 2곳 모두 좌측 지정: `internal/cli/wizard/wizard.go:555`, `internal/cli/wizard/downgrade_confirm.go:31`.
- 확인 내부 빈 줄: huh v2.0.3 `field_confirm.go:261-263` — `!c.inline`일 때 제목 뒤 `"\n\n"` 고정 출력(테마로 제거 불가 → REQ-TRI-003).
- 필드 사이 간격: `internal/cli/wizard/wizard.go:645` — `FieldSeparator` 개행 축소(t586 REQ-ITI-015).
- 선택 폭: `internal/cli/wizard/wizard.go:299`(`optionColumnWidthCap`), `:317`(`alignOptionLabels`, 표시 폭 패딩).
- init 프로필 confirm 소멸: `internal/cli/init.go:617` 주석(REQ-ITI-001 — 프로필 질문 미포함).
- v1 흡수: `internal/cli/profile_setup.go:330` 부근(v2 위저드 실행, 로케일 해석).
- huh v1 부재: `go.mod`에 v1 무, `internal/`·`cmd/` 에서 v1 import grep 0건.
- 신규 발견(영어 고정 고지문): `internal/cli/profile_setup.go:31` `acceptEditsConfirmationLine` — grep 앵커 토큰 계약(REQ-CCI-006) → REQ-TRI-006/AC-TRI-007.
- 캡처 자산: `internal/cli/wizard/ptycap_test.go`(tmux 케이스), `layout_alignment_test.go`(View 렌더 검사), `internal/cli/ptycaptest/`(하니스).

## §D Plan-phase Evidence (측정 기록)

| # | Claim | Command | Observed output |
|---|---|---|---|
| 1 | 렌더 정렬·간격 기존 검사가 이 트리에서 통과 | `go test ./internal/cli/wizard/ -run 'TestLayout_NoBlankBetweenFields' -count=1` (plan-audit 1차 -v 관측: 해당 셀렉터 목록과 일치하는 검사는 이 1건) | `ok github.com/modu-ai/moai-adk/internal/cli/wizard 0.607s` |
| 2 | 확인 생성 지점 2곳 + 모두 좌측 정렬 | `grep -rn "huh.NewConfirm" internal/ ; grep -rn "WithButtonAlignment" internal/ \| grep -v _test` | 2 지점(`wizard.go:544`, `downgrade_confirm.go:26`) / 2 정렬 지정(`wizard.go:555`, `downgrade_confirm.go:31`) |
| 3 | huh v1 잔존 없음 | `grep -rn "charmbracelet/huh\"" internal/ cmd/ ; grep -n "charmbracelet/huh" go.mod` | 양쪽 모두 0건(빈 출력) |
| 4 | t586 완료·본 베이스 포함 | `.moai/specs/SPEC-INIT-TUX-I18N-001/spec.md` frontmatter 판독 | `status: completed`, `updated: 2026-09-13` |
| 5 | SPEC ID 규격·고유성 | 정규식 셀프체크 + specs 디렉터리 조회 | `PASS` / 기존 ID 충돌 없음 |
| 6 | 라이브러리 기본 정렬·고정 빈 줄 존재 | huh v2.0.3 `field_confirm.go` 판독 | `:53 buttonAlignment: lipgloss.Center`, `:261-263` `"\n\n"` 출력부, `:136 Inline`, `:361 WithButtonAlignment` |

## §E 감사 신호

## §E.1 Plan-phase Audit-Ready Signal

```yaml
spec: SPEC-CLI-TUX-RENDER-I18N-001
phase: plan
status: draft
tier: M
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
req_count: 8
ac_count: 9
gears_compliance: all REQ use GEARS patterns (Event/State/Ubiquitous), no IF/THEN
out_of_scope: 3 H3 sections with bullets (question-structure overhaul, huh upstream patch, new locale content)
capture_gate: every render/i18n AC judged on PTY frames (card HARD clause honored)
stale_ref_resolution: all card refs re-anchored; table in plan.md §B
blockers: none
needs_clarification_markers: none
```

## §E.2 Run-phase Evidence

### Run-phase 결정 기록

- **D4 (M1 측정 → M2 범위 재확정)**: M1 이 모든 confirm 표면에서 확인 필드 내부 빈 행을 **1**로 실측했다
  (카드 판정문의 "2" 가 아님). huh `field_confirm.go:261-263` 의 고정 `"\n\n"` 은 소스상 그대로지만
  가시 빈 행은 1개다. REQ-TRI-003 의 합격선(≤1)은 이미 충족 → REQ-TRI-008 (적합한 코드 수리 금지)에 따라
  plan.md M2 의 래퍼 수리 경로를 실행하지 않고, 실측값 1을 골드로 고정하는 캡처 가드로 대체했다.
  plan.md M1 의 자체 계약("이 표가 M2/M3 범위를 확정한다")이 이 재확정의 근거다.
- AC-TRI-003 의 전제값은 2 → **1** 로 조정됐다 (acceptance.md 의 observe-first 구조 흡수; 위임 프롬프트가 명시 승인).
- D1 (v1 처분 = v2 흡수 유지): 전 트리 검증 유지 — huh v1 import/require 0건, 가드로 봉쇄. 이견 없음.

### AC 매트릭스 (AC-TRI-001..009)

| AC | 상태 | 검증 명령 (이번 실행, 이 트리) | 관측 출력 / 근거 |
|---|---|---|---|
| TRI-001 | **PASS** | `MOAI_PTY_CAPTURE=1 MOAI_PTY_CAPTURE_OUT=<spec>/evidence/baseline go test ./internal/cli/wizard/ -run 'TestPtyCapture_BaselineSurfaces\|TestPtyCapture_NormalRun' -count=1 -v` | `ok ... 3.331s` — 8프레임 수출(`evidence/baseline/*.txt`) + 앵커 도달성 전부 통과 + `evidence/residual-defects.md` 잔여표 |
| TRI-002 | **PASS** | `go test ./internal/cli/wizard/ -run 'TestConfirmAlignmentSweep' -count=1 -v` | `swept 2 huh.NewConfirm site(s), all left-aligned` + 뮤턴트 거부: `downgrade_confirm.go: huh.NewConfirm site lacks WithButtonAlignment(lipgloss.Left)` → FAIL 관측 |
| TRI-003 | **PASS** (검증만 종결 + 가드) | M1 실측: `CONFIRM-INTERNAL-BLANK-ROWS confirm-fixture-en = 1` (en/ko fixture, en/ko downgrade 전부 1) / 가드: `MOAI_PTY_CAPTURE=1 go test ./internal/cli/wizard/ -run 'TestPtyCapture_ConfirmGapBudget' -count=1 -v` | 빈 행 1 ≤ 1 충족; `TestPtyCapture_ConfirmGapBudget` 4 서브테스트 전부 PASS (`ok ... 7.455s`) |
| TRI-004 | **PASS** | `go test ./internal/cli/wizard/ -run 'TestLayout_NoBlankBetweenFields' -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli/wizard` (M2 재검증 녹색, 감쇠 없음) |
| TRI-005 | **PASS** | `go test ./internal/cli/wizard/ -run 'TestOptionDescriptionColumn_DisplayWidthAligned' -count=1` | 녹색 (표시 열 동일 유지) |
| TRI-006 | **PASS** | `MOAI_PTY_CAPTURE=1 go test ./internal/cli/wizard/ -run 'TestPtyCapture_I18nKoSweep' -count=1 -v` | 4 서브테스트 PASS — ko 프레임에 번역 존재 + 영어 원문 잔존 0; acceptEdits 고지의 en 잔존 0은 앵커 검사 enResidues가 잠금 |
| TRI-007 | **PASS** | `go test ./internal/cli/ -run 'TestEmitAcceptEditsConfirmationAnchor' -count=1 -v` | PASS — en/ko/ja/zh 전 로케일 앵커 토큰(`acceptEdits`, `settings.local.json`) 보존 + 현지 문장 + en 잔존 0; 증거: `evidence/acceptEdits-localization.md` |
| TRI-008 | **PASS** | `go test ./internal/cli/ -run 'TestHuhV1NonRegressionGuard' -count=1 -v` | `swept 255 production source files + go.mod: no huh v1` + 뮤턴트 2건 거부 관측 (go.mod require 추가 / import 추가) |
| TRI-009 | **PASS** | `git diff --stat a404132e7..HEAD -- internal/cli/wizard/wizard.go internal/cli/wizard/downgrade_confirm.go` | **빈 stat (수리 diff 0)** — 검증만 종결 5행의 표면 코드 무변경; 프레임 증거 = `evidence/baseline/*.txt` |

### 뮤턴트 관측 기록 (acceptance.md §F 3항)

- **AC-TRI-002 뮤턴트**: 생산 소스에서 `WithButtonAlignment(lipgloss.Left)` 제거 → 동일 스윕 함수가
  `downgrade_confirm.go` 파일을 지목하며 FAIL. (초기 뮤턴트 구성이 접미 점(`."+marker`) 패턴이라 무효였던
  사건을 뮤턴트 테스트가 스스로 잡아 수정 — 뮤턴트 무효 자체도 관측됨.)
- **AC-TRI-008 뮤턴트**: (a) go.mod에 `github.com/charmbracelet/huh v1.14.0` 추가 → FAIL;
  (b) 소스에 v1 import 리터럴 추가 → FAIL (파일명 지목).
- **공축 방지**: 두 가드 모두 스윕 대상 수 하한(사이트 0 에러 / 파일 20 미만 에러)을 가져 대상 0 공허 통과가 불가능하다.

### 파일 변경 (run-phase, 6 코드 파일)

- `internal/cli/wizard/ptycap_test.go` — M1/M2/M3 캡처 케이스 + 측정 + i18n 스윕
- `internal/cli/wizard/confirm_alignment_sweep_test.go` — 신규 (REQ-TRI-002 가드)
- `internal/cli/huh_v1_guard_test.go` — 신규 (REQ-TRI-007 가드)
- `internal/cli/profile_setup.go` — acceptEdits 고지 현지화 (유일한 생산 코드 수리)
- `internal/cli/profile_setup_acceptEdits_test.go` — 앵커 검사 4-로케일 갱신
- `internal/cli/profile_setup_absorb_test.go` — once-only 카운트를 로케일 불변 토큰 쌍으로 전환

### 커밋 목록

| SHA | 제목 |
|---|---|
| 3547cb233 | feat(SPEC-...-001): M1 baseline PTY capture + residual defect table |
| 7504c895b | test(SPEC-...-001): M2 alignment sweep guard + confirm gap budget freeze |
| 481ca8d89 | feat(SPEC-...-001): M3 localize acceptEdits notice + ko sweep + huh v1 guard |
| (M4) | docs(SPEC-...-001): M4 closure records (이 커밋) |

## §E.3 Run-phase Audit-Ready Signal

```yaml
spec: SPEC-CLI-TUX-RENDER-I18N-001
phase: run
run_complete_at: 2026-09-14
run_commit_sha: "66d790960"  # M4 커밋 — pending-backfill-run 플레이스홀더 백필 (D3 자기참조 면제)
run_status: complete
ac_pass_count: 9
ac_fail_count: 0
verification_only_closed_rows: 5   # 잔여표 6행 중 수리 1행(acceptEdits) + 종결 5행
repair_rows: 1
preserve_list_post_run_count: 0    # plan §B PRESERVE 대상(적합 표면) 전부 무변경 — TRI-009 빈 diff로 입증
l44_pre_commit_fetch: "not-run (worktree-isolated card branch; develop integration is the lead's window)"
l44_post_push_fetch: "n/a (lane does not push; lead batch-pushes develop per gitflow-lane-protocol §4)"
new_warnings_or_lints_introduced: 0  # golangci-lint run internal/cli/... → 0 issues
cross_platform_build:
  darwin: pass   # go vet + 전체 테스트가 이 플랫폼에서 실행됨
  windows: pass  # GOOS=windows GOARCH=amd64 go build ./... → exit 0
  linux: not-run # 미측정 — CI 매트릭스 몫
coverage_wizard_package: "93.6% (go test -cover ./internal/cli/wizard/ -count=1)"
coverage_internal_cli: "not-measured (전체 스위트는 CI 몫 — 공유 머신 규율)"
capture_gate_contract: "skip observable without gate (TestPtyCapture_SkipWithoutGate: 6 SKIP 관측), fail without tmux (TestPtyCapture_FailWithoutTmux: 5 FAIL 관측)"
total_run_phase_files: 6
m1_to_mN_commit_strategy: "per-milestone commits (M1 capture -> M2 guards -> M3 i18n -> M4 closure), each with Authored-By-Agent trailer"
deviations:
  - "plan.md M2 래퍼 수리 경로 미실행 — M1 실측 1로 REQ-TRI-003 이미 충족 (REQ-TRI-008 경계); 캡처 가드로 대체"
  - "AC-TRI-003 전제값 2 -> 1 조정 (observe-first 흡수)"
  - "골든 갱신 없음 — 렌더 변경 0 (TestConfirmButton_LeftAligned 골든 비교 녹색 유지로 입증)"
blockers: none
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
spec: SPEC-CLI-TUX-RENDER-I18N-001
card: t756
phase: sync
sync_complete_at: 2026-09-14
sync_commit_sha: "4c4534419"  # backfilled (D3 exemption): the sync commit itself wrote the pending-backfill-sync placeholder
sync_status: complete
b12_self_test_a: "pre-emission grep 'SPEC-CLI-TUX-RENDER-I18N-001' CHANGELOG.md = 0 (no duplicate entry)"
b12_self_test_b: "AC count — 9 live AC-TRI-001..009 (acceptance.md §C matrix); AC-ITI-019 grep hit is the predecessor SPEC's inherited-contract reference, excluded"
b12_self_test_c: "all file paths cited in the CHANGELOG entry verified via ls (profile_setup.go, wizard ptycap_test.go, confirm_alignment_sweep_test.go, huh_v1_guard_test.go, 2 profile_setup test files)"
changelog_entry_position: "CHANGELOG.md [Unreleased] ### Added, first bullet"
frontmatter_status_transitions:
  spec.md: "in-progress → completed (status only; updated already 2026-09-14)"
  plan.md: "none — stateless on the status axis (no frontmatter, per spec-frontmatter-schema § Artifact Statelessness)"
  acceptance.md: "none — stateless on the status axis (no frontmatter)"
  progress.md: "phase: run → sync; no status field by design"
canary_compliance_check:
  readme_docs_site: "no change — wizard flow/options/prompts behaviorally unchanged; the acceptEdits notice locale is an undocumented stderr detail, so the README 4-locale same-change obligation is not triggered"
  codemaps: "skipped — no architecture change in this SPEC (test guards + one function-body localization)"
  mx_tags: "validated during sync sub-step; no new exported production symbols beyond emitAcceptEditsConfirmation change (existing anchor contract unchanged)"
verification_basis: "run-phase §E.2 AC matrix (9/9 PASS, this tree, orchestrator-verified zero repair diff); lint/vet/GOOS=windows results carried from §E.3; sync phase added docs only — no source re-measurement needed beyond B12 self-tests"
blockers: none
```
