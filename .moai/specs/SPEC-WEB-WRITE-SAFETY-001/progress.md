# Progress — SPEC-WEB-WRITE-SAFETY-001

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_complete_at: 2026-09-07
plan_status: audit-ready
tier: M
cycle_type: tdd
artifacts:
  - spec.md         # GEARS REQ-WWS-001..008, §1 관측+코드 근거(직접 확인/전달 구분), §3 경계(t509/t510), Out of Scope 4개 H3
  - plan.md         # Class B 조사-선결: M1 재현 → M2/M3 첫 측정 게이트 → M4 수리 → M5 회귀 가드
  - acceptance.md   # AC-WWS-001..008, 부재-가드 RED-first(4요소) + 뮤턴트 필수, AC-WWS-004 양성 통제
  - progress.md     # this file
spec_id: SPEC-WEB-WRITE-SAFETY-001
module: internal/web, internal/config
related_specs: [SPEC-WEB-CONSOLE-011, SPEC-WEB-CONSOLE-010, SPEC-GITSTRATEGY-SAVE-ISOLATION-001, SPEC-FEEDBACK-AUTO-SUBMIT-001]
card: t517
```

plan_status는 plan-audit 통과 시 `audit-ready`로 갱신된다(갱신 소관: plan-audit 반영 오케스트레이터). 갱신 실측: iter1 FAIL 0.88(D1 차단 1건) → 정정 `4e94f9607` → iter2 **PASS 1.00**, D1 RESOLVED, 회귀 없음(`.moai/reports/t517/plan-audit-iter2.md`, Tier M 상한 2/2 도달). 잔여 optional D5(spec.md:141 구(舊) 협의 괄호 해설)는 sync-phase 재량 정정 대상.

## §F Phase 4 Mode Selection

```yaml
inputs:
  tier: M
  scope_files: ~6 (internal/web 3-4 + internal/config 1-2 + tests)
  domains: 2 (Go web handlers, Go config manager) — research-heavy 아님, coding-heavy 조사·수리
  language_mix: 100% Go
  concurrency_benefit: LOW (M1 재현 → M2/M3 측정 → M4 수리가 강한 순서 의존)
mode_eval:
  direct: not_selected — 구현 규모가 trivial 밖 (수리+회귀 가드+격리 재현 절차)
  serial: selected — coding-heavy 단일 도메인, 마일스톤 순서 의존 (Anthropic coding-task caveat)
  fanout: not_selected — 다중 도메인 리서치 아님
  sweep: not_selected — 기계적 대량 변형 아님
  agent_team: not_selected — 운영자 명시 요청 없음
decision: serial
justification: >
  M1-M3은 재현과 측정이 순서 의존(이전 단계의 출력이 다음 단계의 입력)이고 M4-M5는
  측정 결론에 귀속되므로 병렬화 이득이 없다. 단일 manager-develop에 Section A-E
  템플릿으로 위임하고 각 마일스톤을 순차 집행한다.
kickoff: pending — Implementation Kickoff Approval 게이트 대기 중 (이 로그는 승인이 아니다)
```

## §E.2 Run-phase Evidence

> 측정 환경: 워크트리 `.claude/worktrees/t517`, 브랜치 `WT-web-write-safety`, pre-fix 트리 SHA `2031ccf7f` (바이너리: 본 트리에서 `go build -o /tmp/t517/moai ./cmd/moai`, `moai version` → `v3.1.3 none built unknown` — LDFLAGS 미주입 빌드, 재현용). 격리 fixture: `/tmp/t517/fixture` (워크트리 `.moai/config/` 복사, 스냅샷 SHA-256은 `.moai/reports/t517/snapshots/SHA256SUMS.txt`). 동시 세션: `moai session list --json` → `[]` (이 SPEC 스코프 동시 세션 없음). primary checkout 무접촉.

### M1 — 격리 실물 재현 (RED-now, 3단계 판별 + 값-불변 Save)

| 단계 | 절차 | 판정 커맨드 (단일 호출) | verbatim 출력 | exit | 트리 SHA |
|---|---|---|---|---|---|
| (a) 기동만(브라우저 미접속, 8s) | `timeout 8 /tmp/t517/moai web --no-open --port 3941` | `diff -rq <worktree>/.moai/config/sections /tmp/t517/fixture/.moai/config/sections` | (출력 없음 — diff 0행) | 0 | `2031ccf7f` |
| (b) GET 라우트 전부 순회 | `/`·`/kanban`·`/monitor`·`/todo`·`/settings`·`/specs`·`/static/*` 전부 200 + `/events` 200 | 동일 diff | (출력 없음 — diff 0행) | 0 | `2031ccf7f` |
| (c) 탐색·폴링 유지 | 12라운드 `/events`+`/settings` 반복 요청 | 동일 diff | (출력 없음 — diff 0행) | 0 | `2031ccf7f` |
| (d) 값-불변 POST /save | 렌더된 settings 폼 71필드를 **값 변경 없이** 재제출 (mcp.* 29필드 제외 — fixture에 mcp.yaml 부재로 PatchFile stat-fail 500이 먼저 관측됨, 별도 결함으로 기록) | 동일 diff | `feedback.yaml`, `git-strategy.yaml`, `gate.yaml`, `workflow.yaml` differ + `llm.yaml` 신규 생성 | 1 | `2031ccf7f` |

**(d)의 O1 서명 재현 (M-b의 1차 증거):**

- `feedback.yaml`: `7d6 <` — **빈 줄 1행 삭제** (O1과 정확히 일치)
- `git-strategy.yaml`: `worktree_base_branch: develop` 키 5행→7행 **이동** + manual/team 프로필에 `develop_branch: ""`·`release_branch_prefix: ""`·`rc_version_format: ""` **신규 추가** (O1의 "키 순서 변경 + develop_branch 추가"와 일치 — 실제로는 3키)
- `llm.yaml`: fixture에 **부재**했던 파일이 신규 생성됨 — O2(primary llm.yaml 오늘 기록)의 기제와 일치하는 시그니처 (`Save()`의 llm 무조건 재기록이 부재 파일을 생성)
- `workflow.yaml` (O1에 없던 추가 관측): `effort: high`→`effort:` 값 소실 2건, `failure_pattern_detection`/`auto_merge`/`tmux_preferred`/`enabled` true→false 전환, `codex.*`/`todo.enabled` 신규 추가 — **중복 폼값 첫-값 채택 + false 폴백의 실재 손상**
- `gate.yaml`: `pre_commit.enabled: false` 신규 추가
- `user/language/quality/git-convention.yaml`: diff 없음 — `Save()`가 무조건 재기록하되 **내용 동일 round-trip**(mtime만 변화, §D.2 내용 기준 판정상 무차이)

### M2 — 첫 측정 A (M-a): 무저장 쓰기 경로 귀속

| 후보 기제 (spec §1.4) | 판정 | 근거 |
|---|---|---|
| 1. htmx 자동 POST (`hx-trigger` 자동 제출) | **소거** | 정적: `hx-trigger` 매치 0 (htmx.min.js 제외 전 트리), `hx-boost="true"`는 실제 form 제출에만 반응(`root_templ.go:96`). app.js의 change 리스너는 전부 클라이언트 상태 조작만 하고 submit 없음. 동적: (a)(b)(c)단계 diff 0 |
| 2. lazy GET 쓰기 | **소거** | 정적: internal/web의 쓰기 호출 전부 `handleSave` 내부 9 seam(handlers.go:465-546). 동적: (b) GET 9라우트+events 순회 diff 0 |
| 3. CLI 래퍼 기동 접촉 | **소거** | `internal/cli/web.go`는 thin 진입점 — 플래그 파싱 + 포트 확보 + `web.Run` 위임, config 쓰기 호출 0. 동적: (a) 기동만 diff 0 |
| 4. 외부 동시 작성자 | **불필요** | O1의 두 파일 변경이 값-불변 POST /save 1회로 **완전히 재현**됨 (서명 일치). 외부 작성자 가설로 설명할 필요가 없다 |

**M-a 결론**: 무저장 쓰기 경로는 본 코드베이스에 존재하지 않는다(3단계 diff 0 + 정적 소거). O1의 변경은 **POST /save가 실제로 제출된 것**으로 귀속된다. Save 버튼 클릭 없이도 폼 안 Enter 키 제출 등 브라우저의 모든 폼-제출 이벤트가 POST /save로 수렴하고(`hx-boost` 폼, 상단바 Save 버튼은 `form="settings-form"` 원격 제출 — shell_templ.go:909), 서버는 제출 여부만 보고 **값이 바뀌지 않은 제출**도 전부 기록한다. 즉 결함의 실체는 "저장 없는 쓰기 경로"가 아니라 **"값-불변 제출에 대한 무차별 기록"**(REQ-WWS-003 위반)이다.

**6종 판별 (D2)**: 내용 변경 = feedback(seam)/git-strategy(typed-gated·dirty)/llm(생성)/workflow·gate(seam·부수 손상). 내용 동일 round-trip = user/language/quality/git-convention. → 무조건 6종 전체가 "기록"은 되지만, O1의 **내용 변경**은 값-변경 판정이 없는 seam/typed-gated 경로의 편집이 통과한 결과다 — "전체 Save() 통과"와 "부분 경로"가 섞인 하이브리드이며, POST /save 단일 제출로 모두 설명된다.

### M3 — 첫 측정 B (M-b): gate 우회 + seam 충실도 원인 + 환원 RED 테스트

**원인 1 — git-strategy dirty-gate 우회**: `ApplySchemaEdits`(`internal/settings/sectionapply.go:88-150`)의 `applyTypedEdits`는 제출된 git_strategy 필드를 **현재 값과 비교하지 않고 무조건** `SetSection("git_strategy")`을 호출한다(`:142`). `gitStrategyDirty`가 세팅되고(`manager.go` dirty-flag 계약), `Save()`가 git-strategy.yaml을 전체 재마샬한다 — struct 순서로 키 재배열(`worktree_base_branch` 이동) + struct에는 있는데 디스크에 없던 키를 zero value로 추가(`develop_branch: ""` 등 3키).

**원인 2 — feedback seam 빈 줄 삭제**: `yamlpatch.PatchFile`(`internal/settings/yamlpatch/yamlpatch.go:42-82`)은 노드 수술 후 **문서 전체를 yaml.v3 Encoder로 재직렬화**한다(`encode`, `:173`). yaml.v3 인코더는 빈 줄을 표현 요소로 유지하지 못한다 — 패키지 헤더 `:11-12`에 문서화된 한계가 AC-WWS-005 충실도 요구를 위반하는 것으로 실증됐다.

**환원 RED 테스트 (커맨드·출력·exit·트리 SHA)**:

| 테스트 | 커맨드 (단일 호출) | 판정 | exit | 트리 SHA |
|---|---|---|---|---|
| `TestPatchFileValueInvariantPreservesBytes` (AC-WWS-005) | `go test -count=1 -run TestPatchFileValueInvariantPreservesBytes ./internal/settings/` | FAIL — 값-불변 패치가 파일을 재작성 (빈 줄 삭제) | 1 | `2031ccf7f` |
| `TestPatchFileScalarChangePreservesPresentation` (AC-WWS-005) | `go test -count=1 -run TestPatchFileScalarChangePreservesPresentation ./internal/settings/` | FAIL — 값 변경 시 빈 줄 미보존 | 1 | `2031ccf7f` |
| `TestApplySchemaEditsValueInvariantTouchesNothing` (AC-WWS-003/004) | `go test -count=1 -run TestApplySchemaEditsValueInvariantTouchesNothing ./internal/settings/` | FAIL — 값-불변 제출이 두 파일 재기록 | 1 | `2031ccf7f` |
| `TestHandleSaveValueInvariantLeavesSectionsByteIdentical` (AC-WWS-003/004 통합) | `go test -count=1 -run TestHandleSaveValueInvariantLeavesSectionsByteIdentical ./internal/web/` | FAIL — 전체 쓰기 경로에서 값-불변 저장이 두 파일 재기록 | 1 | `2031ccf7f` |
| `TestParseSchemaFormDuplicateFormValuesNotSilentlyFirst` (AC-WWS-006) | `go test -count=1 -run TestParseSchemaFormDuplicateFormValuesNotSilentlyFirst ./internal/web/` | FAIL — 중복 첫 값 조용한 채택 | 1 | `2031ccf7f` |

verbatim RED 출력: `.moai/reports/t517/evidence/RED-settings-write-safety.log`, `.moai/reports/t517/evidence/RED-web-write-safety.log`. RED 테스트 파일은 untracked 상태(`2031ccf7f` 트리)에서 실행·관측 후 본 커밋에 포함된다.

**AC-WWS-004 양성 통제 (필수, 사전 관측)**: `TestApplySchemaEditsGitStrategyRealChangeStillRewrites` — **PASS** (실제 값 변경 `team→personal` 시 git-strategy.yaml 재기록 관측). `TestHandleSaveGitStrategyRealChangeStillRewrites` — **PASS** (전체 handleSave 경로 동일). → gate가 실제 변경에는 반응하는 살아 있는 gate다. 값-불변 테스트의 GREEN은 "gate이 지킨 것"이 아니라 "수리 후 편집이 아예 SetSection에 도달하지 않는 것"으로 해석된다.

**M4 수리 설계가 인용하는 측정 결론 (REQ-WWS-008)**: (i) 값-불변 edit은 ApplySchemaEdits 진입 시 현재 값과 비교해 제거한다 — M-a 결론("값-불변 제출의 무차별 기록")의 직접 수리. (ii) git_strategy/llm/quality typed는 apply 전후 구조체 비교로 실변경 시에만 SetSection한다 — M-b 원인 1의 수리. (iii) yamlpatch의 기존-스칼라 교체는 원문 라인 스플라이싱으로 전환해 표현 요소를 byte 보존한다 — M-b 원인 2의 수리. (iv) parseSchemaForm은 동일 name 다중 제출을 감지해 atomic reject에 합류시킨다 — workflow.yaml 부수 손상 관측의 수리.

### 부수 관측 (scope 밖 — 리드 전달 대상)

1. **mcp.yaml 부재 시 Save 500**: fixture에 `mcp.yaml`이 없으면 `applySchemaEdits→PatchFile→atomicWrite`가 `stat mcp.yaml: no such file or directory`로 500. PatchFile은 read 단계에서 greenfield(`{}\n`)를 허용하지만 atomicWrite의 `os.Stat`이 부재 파일에서 실패한다(`yamlpatch.go:190-193`). 실제 프로젝트 트리(`2031ccf7f`)에도 mcp.yaml이 없다 — Save 전체 제출 시 재현 가능성. (쓰기 안전 축과 별개의 **쓰기 위치/완결성** 축 — t510 경계 인접, 본 카드에서 수리하지 않음)

### M4 — 수리 (커밋 `03cea2f75`, M-a/M-b 결론 인용 — REQ-WWS-008)

| 수리 | 위치 | 인용 결론 |
|---|---|---|
| (i) 값-불변 seam edit 제거 (동일 값 + 부재키·bool-false 형태) | `internal/settings/sectionapply.go` `ApplySchemaEdits` | M-a: 값-불변 제출의 무차별 기록 |
| (ii) typed 실변경 게이트 (apply 전후 DeepEqual, 변경 0이면 Save 생략) | `internal/settings/sectionapply.go` `applyTypedEdits` | M-b 원인 1: 무조건 SetSection → dirty → 재마샬 |
| (iii) yamlpatch 라인 스플라이싱 (기존 스칼라 교체는 대상 라인만 재작성 + 재파싱 검증, upsert만 재직렬화 폴백) | `internal/settings/yamlpatch/yamlpatch.go` `PatchFile`/`lineSplice` | M-b 원인 2: 재직렬화의 빈 줄 정규화 |
| (iv) parseSchemaForm 중복 폼값 감지 (동의 중복은 통과, 불일치 중복은 atomic reject 합류) | `internal/web/schemaform.go` `parseSchemaForm` | M1(d): workflow 스칼라 부수 손상 (첫 값 채택) |
| (v) WriteProjectNestedConfig 실변경 게이트 | `internal/settings/nested.go` | M1(d): llm.yaml 무의미 생성 (동일 값 *Set 플래그가 Save를 유발) |

**수리 후 같은 재현 절차 GREEN (M1(d) 절차, 수리 바이너리로 재실행)**:

- 판정 커맨드: `diff -rq <worktree>/.moai/config/sections /tmp/t517/fixture-post5/.moai/config/sections`
- verbatim 출력: `Only in /tmp/t517/fixture-post5/.moai/config/sections: llm.yaml` (유일 잔여) — **git 추적 파일 전부 diff 0행**
- exit: 1 (untracked 파일 존재에 의한 것 — 아래 판정 참조) / 트리 SHA: `03cea2f75`
- **llm.yaml 생성 잔여 판정**: git 미추적 파일(greenfield)에 대한 **실변경 저장**(agentfm 프로필 매트릭스 해상값과 디스크 부재의 차이)으로 REQ-WWS-003의 "git 추적 기준 diff 없음" 판정에 영향 없음. feedback 빈 줄 삭제·git-strategy 재마샬·workflow/gate 부수 손상은 전부 소멸 — O1 서명 완전 해결.

**git-strategy 양성 통제 (수리 후)**: `TestApplySchemaEditsGitStrategyRealChangeStillRewrites` PASS, `TestHandleSaveGitStrategyRealChangeStillRewrites` PASS — 실변경 제출 시 재기록이 관측됨(gate 생존의 반대 방향 증거).

### M5 — 회귀 가드 채택 + 뮤턴트 검증

**2-cell 채택 (부재-가드 5건)**: RED-now 셀 = 상기 M3 표 (트리 `2031ccf7f`, verbatim 로그 링크 포함) / green path 셀 = M4 수리가 동일 커맨드를 PASS로 뒤집음 (커밋 `03cea2f75` 트리, 아래 최종 실행).

**뮤턴트 검증 (AC-WWS-007) — 재도입 변형 3건 모두 포착:**

| 뮤턴트 | 재도입 변형 | 포착 가드 | 판정 |
|---|---|---|---|
| MUTANT-A | applyTypedEdits의 DeepEqual 게이트 제거 (무조건 SetSection 복원) | `TestApplySchemaEditsValueInvariantTouchesNothing` FAIL + `TestHandleSaveValueInvariantLeavesSectionsByteIdentical` FAIL | 포착 |
| MUTANT-B | lineSplice 조기 `return nil, false, nil` (재직렬화 폴백 강제) | `TestPatchFileValueInvariantPreservesBytes` FAIL | 포착 |
| MUTANT-C | parseSchemaForm 중복 감지 제거 | `TestParseSchemaFormDuplicateFormValuesNotSilentlyFirst` FAIL | 포착 |

**못 잡은 뮤턴트 기록**: AC-WWS-001/002(무저장 기동·GET 순회 무쓰기)의 가드는 유닛 테스트가 아니라 **실물 재현 절차**(M1 a~c단계 + 스냅샷 diff)다. "startup에 Save 주입" 뮤턴트를 이 절차로 잡는 것은 가능하나 본 카드에서 미수행 — 사유: startup 주입은 lazy GET 쓰기(M-a 후보 2)와 동일 코드 영역이고, 정적 소거(쓰기 호출이 handleSave 외 0건) + 동적 관측(3단계 diff 0)이 그 결함 클래스를 이미 커버한다. 가드 경계: **새 GET/기동 경로에 쓰기 호출이 추가되면 본 가드 체계는 잡지 못한다** — 후속 카드 후보.

**최종 가드 상태 (뮤턴트 복원 후, 트리 = M5 커밋)**:

| 테스트 | 판정 |
|---|---|
| TestPatchFileValueInvariantPreservesBytes | PASS |
| TestPatchFileScalarChangePreservesPresentation | PASS |
| TestApplySchemaEditsValueInvariantTouchesNothing | PASS |
| TestApplySchemaEditsGitStrategyRealChangeStillRewrites (양성 통제) | PASS |
| TestHandleSaveValueInvariantLeavesSectionsByteIdentical | PASS |
| TestHandleSaveGitStrategyRealChangeStillRewrites (양성 통제) | PASS |
| TestParseSchemaFormDuplicateFormValuesNotSilentlyFirst | PASS |

### E1 — AC-WWS-001..008 최종 매트릭스

| AC | Status | 검증 | 실측 |
|---|---|---|---|
| AC-WWS-001 | **PASS** | M1 실물 재현 3단계(기동·렌더·폴링) + 스냅샷 diff | 3단계 모두 diff exit 0 (트리 `2031ccf7f`) — 결함 미재현 = 코드에 무저장 경로 부재 확인 |
| AC-WWS-002 | **PASS** | M1 (b) GET 라우트 9종 + /events 순회 | diff exit 0 |
| AC-WWS-003 | **PASS** | 값-불변 Save 유닛/전체경로 가드 + 수리 후 실물 diff | RED(`2031ccf7f`) → GREEN(`03cea2f75`), git 추적 파일 diff 0 |
| AC-WWS-004 | **PASS** | dirty-gate 가드 + 양성 통제 2건 | RED → GREEN, 양성 통제 PASS×2 (gate 생존 확인) |
| AC-WWS-005 | **PASS** | golden round-trip (값-불변 + 값-변경) | RED → GREEN (빈 줄·주석·키 순서·unknown key byte 보존) |
| AC-WWS-006 | **PASS** | 중복 폼값 가드 | RED → GREEN (불일치 중복 reject, 동의 중복 통과 — 문서화된 규칙) |
| AC-WWS-007 | **PASS** | 뮤턴트 3건 재도입 → 포착 → 복원 | 포착 3/3, 못 잡은 뮤턴트 1건 기록 (AC-001/002 경계) |
| AC-WWS-008 | **PASS** | 본 문서 M4 표의 측정 결론 인용 + M1~M3 귀속 완료 | 커밋 `911d9bbcc`(측정)가 `03cea2f75`(수리)에 선행 — 측정-선결 위반 코드 0 |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-07
run_commit_sha: "3d31ffff2"  # M5 커밋 — run-phase 최종 커밋 (백필 완료)
run_status: complete
ac_pass_count: 8
ac_fail_count: 0
preserve_list_post_run_count: 0  # PRESERVE 대상 침범 0 — t509(codex 패널)/t510(쓰기 위치) 영역 미접촉
l44_pre_commit_fetch: not-performed  # 레인은 push·fetch를 하지 않는다 (git-flow lane protocol §4 — 리드 일괄)
l44_post_push_fetch: not-performed  # 상동 — 원격 착지 검증은 리드 소관
new_warnings_or_lints_introduced: 0  # golangci-lint ./internal/web/... ./internal/settings/... → 0 issues
cross_platform_build:
  darwin_amd64: pass   # go build ./... → exit 0
  windows_amd64: pass  # GOOS=windows GOARCH=amd64 go build ./... → exit 0
total_run_phase_files: 23  # M1-M3 커밋 14 + M4 커밋 5 + M5 커밋 4 (progress/spec/증거 포함, 커밋별 산출)
m1_to_m5_commit_strategy: "3 commits — M1-M3 measured reproduction + RED tests (911d9bbcc) / M4 repairs (03cea2f75) / M5 guard adoption + evidence (this commit)"
known_baseline_failures: "internal/template TestManifestHashFormat — CATALOG_HASH_UNSTABLE (agents 소스 vs catalog.yaml stored hash 불일치). 본 카드 수정 파일과 무관한 pre-existing baseline 결함(별도 카드 소관)."
side_observations:
  - "mcp.yaml 부재 시 PatchFile→atomicWrite stat 실패로 Save 500 (yamlpatch greenfield read와 atomicWrite stat의 불일치) — t510 경계 인접, 별도 소관"
  - "llm.yaml 부재 greenfield에서 agentfm 프로필 매트릭스 해상값이 실변경 저장으로 파일 생성 — git 미추적 파일, REQ-WWS-003 판정 축 밖"
primary_checkout_untouched: true  # 모든 재현·검증은 /tmp/t517/fixture* 격리 트리에서 수행
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-07
sync_commit_sha: "e94d2f2b4"  # backfill 완료 (2026-09-07) — sync 커밋 본체 e94d2f2b4, 본 백필 커밋이 실측 SHA로 교체 (D3 placeholder-backfill exemption)
sync_status: complete
changelog_entry_position: "CHANGELOG.md [Unreleased] > Fixed 섹션 선두"
b12_self_test_a_pre_emission_grep: "grep -c 'SPEC-WEB-WRITE-SAFETY-001' CHANGELOG.md → 0 (중복 0건 — 발행 허용)"
b12_self_test_b_ac_count_match: "acceptance.md 고유 AC 식별자 8건 (AC-WWS-001..008, grep -oE | sort -u 실측) == CHANGELOG 엔트리 인용 8건"
b12_self_test_c_file_path_verification: "엔트리가 이름 대는 파일 7건 + spec.md — ls 실측 전부 존재 (OK 8/8)"
frontmatter_status_transitions:
  spec_md_status: "in-progress → completed (sync 커밋 탑승 — 3-phase close)"
  spec_md_updated: "2026-09-07"
  body_edits: none  # spec.md/plan.md/acceptance.md 본문 무변경 — D5(spec.md:141 구 협의 괄호)는 auditor 유지 허가로 존속
mx_tag_changes:
  added:
    - "@MX:ANCHOR + @MX:REASON → internal/settings/yamlpatch/yamlpatch.go PatchFile (호출 파일 3개 5호출점 — fan_in ≥ 3 계약 + REQ-WWS-005 lineSplice 분기 구조 계약)"
  verified_no_change:
    - "ApplySchemaEdits — sectionapply.go 파일헤더 @MX:WARN 존재, 신규 게이트는 REQ-WWS-003 인라인 문서화로 충분"
    - "WriteProjectNestedConfig — nested.go 파일헤더 @MX:WARN 존재, 실변경 게이트는 REQ-WWS-003 인라인 문서화로 충분"
    - "handleSave — handlers.go @MX:WARN 존재 (변경 없음, 배선만 확인)"
    - "parseSchemaForm — 비-export 함수, REQ-WWS-006 문서 주석 보유 (태그 불필요)"
readme_docs_site_judgment: "변경 없음 — README 4-locale의 moai web 문구는 화면·탭·프로필 관리 구조만 기술하고 저장 의미론에 대한 주장이 없어 정정 대상 0건(README.ko/en/ja/zh :133/:408-414/:610/:750 실측); docs-site 콘텐츠는 본 트리에 부재(docs/ 하위 design/ 뿐)"
go_code_touched_by_sync: "MX 태그 주석 1건만 추가 (yamlpatch.go) — 주석 전용 변경, go test -count=1 ./internal/settings/... ./internal/web/... 재실행으로 확인"
```
