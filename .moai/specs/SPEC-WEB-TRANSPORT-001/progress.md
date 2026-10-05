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
run_commit_sha: f3a2988e1              # 백필 2026-09-22 (manager-docs, D3 자기참조 예외 경유)
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

```yaml
sync_complete_at: 2026-09-22
sync_commit_sha: 475f9dd23               # 백필 2026-09-22 (D3 자기참조 예외 — 3-phase close 커밋)
sync_status: complete
frontmatter_status_transitions:
  spec_md: in-progress → completed       # 단일 sync 커밋 병합 체결 (implemented 스킵 아님 — 3-phase close 병합)
  updated: 2026-09-22
changelog_entry_position: none           # 측정 전용 SPEC — 운용 코드 0, 사용자 표면 변화 0 (근거: 본 절 산문)
canary_compliance_check:
  b12_pre_emission_grep: PASS            # grep -c 'SPEC-WEB-TRANSPORT-001' CHANGELOG.md → 0 (중복 없음)
  b12_ac_count_match: n/a                # emission 스킵 결정 — acceptance.md AC 5건 대조 불요
  b12_file_path_verification: PASS       # transport400_characterization_test.go / transport400_htmx_contract_test.go 존재 확인
  b12_ac_live_identifiers: 5             # AC-TR400-001..005 — sort -u 기준
```

**CHANGELOG / docs-site 결정 — 미기입 (no entry).** 근거: 본 SPEC 은 측정 전용(measurement-only)으로
운용 코드 변경 0(preserve_list_post_run_count: 0, 신규 `*_test.go` 2개 + 판정서뿐)이고, 사용자 대면
동작 변화가 없다. CHANGELOG 는 사용자 표면 변화를 기술하는 표면이므로 기입 대상이 아니다. docs-site 도
동일 — 내부 측정 하네스는 사용자 문서화 대상이 아니다. 판정 내용(defect-present, contract-level)은
`.moai/reports/t1080/verdict.md` 와 progress.md §E.2 에 이미 기록돼 있으며, 후속 수리는 t1081 축
(브라우저 실측, M4 분할 경계)의 소관이다.

**전이 근거**: spec.md frontmatter `status:` 는 본 단일 sync 커밋에서 `in-progress → completed` 로
체결된다(spec-frontmatter-schema.md § Status Transition Ownership Matrix — `implemented` 중간 전이는
sync 커밋에 병합되는 3-phase close 규약). plan.md / acceptance.md 는 status 축 무상태(stateless)라
전이 대상이 아니다.
