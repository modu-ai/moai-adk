# Progress — SPEC-WEB-TRANSPORT-001

> 카드 t1080 · lane 워크트리 `WT-verify-400-path` · base `0314801c2`

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-09-22
tier: M
artifacts: 4  # spec.md, plan.md, acceptance.md, progress.md
spec_id_check: PASS  # ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ — verbatim Bash 실행
base_sha: 0314801c2
```

plan-phase 측정 근거(자체 측정, 본 레인 트리): htmx 로컬 임베드 확인(`assets.go:20-24`), 핀 버전 2.0.4(자산 내 `version:"2.0.4"`), `responseHandling` 식별자 1회 존재 — 브라우저 불요 계약 수준 측정 경로 성립. 400 본문 피드백 주장은 미증명 상태로 SPEC 에 명시(REQ-TR400-001 이 승격 담당).

## §E.2 Run-phase Evidence

Run-phase 측정 (본 레인, 트리 `WT-verify-400-path` @ `a8d31593a` — 2026-09-22):

| 측정 | 값 | 근거 |
|---|---|---|
| D-400a (서버 축) | **true** — 400 본문이 배너 + 렌더되는 필드 오류를 운반 | `go test ./internal/web/ -run TestCharacterizeValidation400BodyFeedback -v` → `D-400a MEASURED: 400 body carries banner "Validation failed — no changes were saved." AND per-field error "unrecognized permission mode: bogus" (113765 bytes)` — PASS |
| D-400b (클라이언트 계약) | **swap: false (error: true)** — 400 본문 기본 폐기 | `go test ./internal/web/ -run TestHtmxResponseHandlingTableExtraction -v` → 테이블 조각 전문 `responseHandling:[{code:"204",swap:false},{code:"[23]..",swap:true},{code:"[45]..",swap:false,error:true}]`, 400 → 엔트리 #2 `{code:"[45]..", swap:false, error:true}` — PASS |

**게이트된 판정 (REQ-TR400-003 사상표)**: (true, swap:false) → **defect-present (contract-level)** ·
신뢰도 등급 **contract-level** (REQ-TR400-004 상한 — browser-observed 아님). 브라우저 실측은
t1081 축(분할 경계 유지 확인 — M4). 운용 코드 변경 0, t1051 판정 비인용(REQ-TR400-005).

프로브 결함 배제 기록: 1차 프로브는 development_mode 필드 오류 마커로 본문 미포함을 관측 —
발산 아님(배너는 존재), 원인은 해당 필드 오류가 렌더 표면 은퇴로 본문에 렌더되지 않는 것.
렌드되는 permission_mode 필드 오류로 재조준해 재측정 통과.

전문 증거: `.moai/reports/t1080/verdict.md` (§1 측정 출력, §3 AC 매트릭스).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-22
run_commit_sha: pending-backfill-run   # 레인 커밋 후 백필 (D3 자기참조 예외)
run_status: complete
ac_pass_count: 5
ac_fail_count: 0
preserve_list_post_run_count: 0        # 운용 소스 변경 0 — 신규 *_test.go 2개만
l44_pre_commit_fetch: n/a              # 카드 워크트리 레인 — push 는 리드 일괄 (2026-09-02)
l44_post_push_fetch: n/a
new_warnings_or_lints_introduced: 0    # go vet ./internal/web/ 정상 (exit 0)
total_run_phase_files: 4               # test 2 + verdict.md + progress.md
m1_to_mN_commit_strategy: lane-single-commit  # 레인이 커밋 — 본 레인은 커밋하지 않음
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs 소관>_
