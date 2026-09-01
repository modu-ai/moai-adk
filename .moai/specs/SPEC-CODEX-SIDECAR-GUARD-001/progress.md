# SPEC-CODEX-SIDECAR-GUARD-001 — 진행 기록

> 카드 t405 · Tier S · 기준 트리 `64bba61aa`

## §E.1 Plan-phase Audit-Ready Signal

| 항목 | 값 |
|---|---|
| 산출물 | `spec.md` · `plan.md` · `acceptance.md` · `progress.md` |
| 요구 | REQ-CSG-001 … REQ-CSG-005 (5건) |
| 수용 기준 | AC-CSG-001 … AC-CSG-008 (8건) |
| 미해소 clarification | 0건 (`[NEEDS CLARIFICATION]` 스윕 0) |
| SPEC ID 사전 검사 | `PASS` (Bash 정규식 실행) |
| 범위 확대 판정 | `:82` 거울상 편입 — 근거는 `spec.md` §E.1, 상한은 §E.2 |
| 미검증 값 | `phase: "v3.1.5 target"` — 상속값, `plan.md` §I 참조 |
| plan-audit | PASS-WITH-DEBT 0.82 (Tier S threshold 0.75) · must-pass 7/7 · blocking D1+D2 — 리드 지시(Tier S 단일 수리 라운드)로 재감사 없이 현장 수리: D1 = AC-CSG-008에 판정 가능한 명령 부여, D2 = AC-CSG-004 뮤턴트를 컴파일 가능한 임시 2편집 형태로 재작성 |

plan-phase 시점 프로덕션 트리 무변경: `git status --porcelain internal/` 출력 없음.

## §E.2 Run-phase Evidence

> manager-develop 소유. M2의 격리 뮤턴트 2종은 **명령과 관측된 출력**을 여기에 기록한다 —
> 요약은 증거가 아니다(`plan.md` M2 · `acceptance.md` §E).
>
> 모든 측정의 baseline 귀속: 본 워크트리, 브랜치 `WT-codexwiring-sidecar`, HEAD `b98a8779c`
> (2026-09-01, run-phase 세션). 뮤턴트 관측 명령은 AC 명령에 로그 노이즈 제거 grep 필터
> (`grep -E 'FAIL|--- |ok |created by|missing'`)를 덧댄 형태로 실행했으며, 판정 행은 전부
> 보존돼 있다.

### M1 — 단언 2줄 추가 (AC-CSG-001 · 002 · 003)

`TestRunInit_AgentBothWiresBothSides`와 `TestRunInit_AgentClaudeLeavesNoCodexFiles`의 경로
슬라이스에 `".moai/state/codex-wiring.json"`을 추가하고 doc comment를 함께 갱신(기존 표현
방식 — 슬라이스 순회 + `os.Stat` — 유지). M1 직후 판정:

| AC | 판정 | 명령 → 관측 출력 |
|---|---|---|
| AC-CSG-001 | PASS | claude 경로 슬라이스 = absent(`:97`) 경로 슬라이스 — 3원소 동일 (`internal/cli/init_agent_flag_test.go` 판독) |
| AC-CSG-002 | PASS | both 경로 슬라이스 = codex(`:70`) 경로 슬라이스 — 3원소 동일 (같은 판독) |
| AC-CSG-003 | PASS | `grep -cE '"\.codex"\|"\.codex/"' internal/cli/init_agent_flag_test.go` → 출력 `0` (grep exit 1 — 부합 없음) |

M1 직후 `go test ./internal/cli/ -run 'TestRunInit_Agent' -v`: 네 시험 전부 `--- PASS`, exit 0.
새 단언이 초록인 것은 프로덕션이 올바르기 때문이며, 적색 능력은 M2의 격리 뮤턴트로 입증한다.

### M2-부재 — AC-CSG-004 (claude 분기가 sidecar 만 기록)

심은 임시 편집 2건: (1) `internal/codexwiring/mutant_tmp.go` 신설 — 비exported `writeSidecar`를
호출만 하는 임시 export 래퍼 `WriteSidecarOnly`, (2) `internal/cli/init.go`
`wireCodexUnlessClaude`의 claude 조기 반환 분기가 그 래퍼를 호출. 관측 직후 원복(rm + Edit).

명령: `go test ./internal/cli/ -run 'TestRunInit_Agent'`

관측된 출력 (verbatim):

```
--- FAIL: TestRunInit_AgentAbsentLeavesNoCodexFiles (0.21s)
    init_agent_flag_test.go:103: .moai/state/codex-wiring.json created by a flag-absent init (AC-CW-004 violation)
--- FAIL: TestRunInit_AgentClaudeLeavesNoCodexFiles (0.21s)
    init_agent_flag_test.go:117: .moai/state/codex-wiring.json created by --agent claude init (must equal flag-absent)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.917s
FAIL
```

- **AC-CSG-004 PASS** — `TestRunInit_AgentClaudeLeavesNoCodexFiles` RED, 실패 메시지가 오직
  `.moai/state/codex-wiring.json` 한 경로(:117 신규 단언)뿐. 같은 시험의 기존 두 단언
  (hooks.json · config.toml)은 실패 메시지 0건 → 새 단언 하나가 적색을 만든 것이 격리로 입증됨.
- **동반 관측(기록 의무, AC-CSG-004 인용부)** — 플래그 부재도 `resolveAgentWiring`에서 claude로
  해소되므로 `TestRunInit_AgentAbsentLeavesNoCodexFiles`(`:97`)도 같은 이유로 RED: **예상과
  일치**.

### M2-존재 — AC-CSG-005 · 006 (Wire에서 writeSidecar 호출 제거)

**1차 시도 — 죽은 뮤턴트 (판정 무효, 재설계 근거로 기록).** `wire.go`의
`sidecarNeeded := res.HooksWritten`을 `false`로 고치는 1줄 뮤턴트를 심었으나 4종 전부 초록
(`ok  github.com/modu-ai/moai-adk/internal/cli  1.782s`), `:70` 포함. 원인:
`wireProject`의 부활 경로(wire.go:141-147) — `sidecarNeeded`가 거짓이어도 sidecar 부재 +
온디스크 hooks.json이 렌더와 바이트 일치 시 `sidecarNeeded`를 다시 참으로 되돌려
`writeSidecar`를 수행한다. 신규 init에서 hooks.json은 방금 렌더한 내용이라 항상 일치하므로 이
뮤턴트 형태는 신규 init에서 쓰기를 못 죽이는 **구조적으로 죽은 뮤턴트**였다. AC-CSG-006 교차
확인(`:70`이 초록으로 남음)이 이를 잡아냈고, `plan.md` §B.2 절차대로 재설계했다 — AC-CSG-005는
1차 시도로 통과 처리하지 않는다.

**재설계 뮤턴트** — `writeSidecar` **호출 자체**를 제거(호출 3줄을 blank assignment로 대체해
컴파일 유지; hooks/config 기록 경로 무변경). 관측 직후 원복(Edit).

명령: `go test ./internal/cli/ -run 'TestRunInit_Agent'`

관측된 출력 (verbatim):

```
--- FAIL: TestRunInit_AgentCodexWiresAndSkipsMCPProvisioning (0.45s)
    init_agent_flag_test.go:75: .moai/state/codex-wiring.json missing after --agent codex init: stat /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestRunInit_AgentCodexWiresAndSkipsMCPProvisioning2778433352/002/tier-proj/.moai/state/codex-wiring.json: no such file or directory
--- FAIL: TestRunInit_AgentBothWiresBothSides (0.24s)
    init_agent_flag_test.go:88: .moai/state/codex-wiring.json missing after --agent both init: stat /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestRunInit_AgentBothWiresBothSides2002682935/002/tier-proj/.moai/state/codex-wiring.json: no such file or directory
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	2.101s
FAIL
```

- **AC-CSG-005 PASS** — `TestRunInit_AgentBothWiresBothSides` RED, 실패 메시지가 오직
  `.moai/state/codex-wiring.json` 한 경로(:88 신규 단언)뿐. 기존 두 단언 실패 메시지 0건 → 격리
  입증.
- **AC-CSG-006 PASS** — 같은 실행에서 `TestRunInit_AgentCodexWiresAndSkipsMCPProvisioning`
  (`:70`)도 동일 이유로 RED → 살아 있는 뮤턴트의 생존 교차 확인 충족.

### 원복 청결 — AC-CSG-007

명령: `git status --porcelain internal/cli/init.go internal/codexwiring/`

관측된 출력: **없음** (빈 출력 — 프로덕션 경로 잔여 변경 0; 임시 래퍼 `mutant_tmp.go`는 rm으로
삭제). `git diff --stat`도 해당 경로 0건 — 커밋에 포함되는 코드 변경은
`internal/cli/init_agent_flag_test.go` 한 파일뿐.

### 세 번째 거울상

작업 중 추가 결함 없음 — `:70` · `:82` · `:97` · `:109` 외 시험의 단언 강도 이상 미발견.
`spec.md` §E.2 상한 준수(흡수 없음, 새 카드 발행 불필요).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-01
run_commit_sha: "d528d8766"  # 백필 (D3 면제) — run 커밋 SHA (2026-09-01, sync-phase에서 기재)
run_status: complete
ac_pass_count: 8
ac_fail_count: 0
preserve_list_post_run_count: 0  # 프로덕션 영구 변경 0 — 유일 코드 변경은 시험 파일 1개
l44_pre_commit_fetch: "executed — git fetch origin develop; origin/develop 32 ahead / HEAD 1 ahead (로컬 plan 커밋). 카드 브랜치 로컬 커밋에는 무영향, 통합 창에서 흡수"
l44_post_push_fetch: "N/A — 본 위임은 push 없음(통합은 리드 지정 창 소관)"
new_warnings_or_lints_introduced: 0  # go vet exit 0 진단 0건; golangci-lint 미실행(로컬 금지 아님, CI 판정 면)
cross_platform_build:
  windows_amd64: "exit 0 — GOOS=windows GOARCH=amd64 go build ./internal/cli/... ./internal/codexwiring/... (컴파일 전용; 테스트 실행은 CI 몫)"
total_run_phase_files: 1  # internal/cli/init_agent_flag_test.go (SPEC 산출물 제외 코드 파일)
m1_to_m1n_commit_strategy: "단일 run 커밋(M1+M2+M3) — 리드 커밋 규율 준수"
```

M3 범위 한정 검증 (AC-CSG-008) — 본 SPEC의 지역 검증 전부이며, 전체 스위트(`go test ./...`)
지역 실행 기록 0건:

| 항목 | 명령 | 관측 |
|---|---|---|
| (i) 패키지 녹색 | `go test ./internal/cli/... ./internal/codexwiring/...` | **exit 0** — 범위 내 전 패키지 `ok` (`internal/codexwiring 6.089s` 포함) |
| (ii) 시험별 판정 | `go test -count=1 ./internal/cli/ -run 'TestRunInit_Agent' -v` | 네 `--- PASS` 행 (`TestRunInit_AgentCodexWiresAndSkipsMCPProvisioning` · `TestRunInit_AgentBothWiresBothSides` · `TestRunInit_AgentAbsentLeavesNoCodexFiles` · `TestRunInit_AgentClaudeLeavesNoCodexFiles`) + `ok`, exit 0 |
| (부가) 정적 분석 | `go vet ./internal/cli/... ./internal/codexwiring/...` | exit 0, 진단 출력 0건 |

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-01
sync_commit_sha: "5483eb4a6"  # 백필 (D3 면제) — 본 항목을 처음 기록한 sync 커밋의 SHA
sync_status: complete
sync_did: "CHANGELOG.md [Unreleased] Fixed 섹션에 close 항목 1건 추가 + spec.md frontmatter in-progress → completed (3-phase close — 본문 무변경, updated: 2026-09-01 유지)"
files_touched:
  - "CHANGELOG.md — [Unreleased] Fixed 섹션 말미에 항목 1건"
  - ".moai/specs/SPEC-CODEX-SIDECAR-GUARD-001/spec.md — frontmatter status 필드만"
  - ".moai/specs/SPEC-CODEX-SIDECAR-GUARD-001/progress.md — §E.4 신설"
b12_self_test_a: "PASS — 사전 중복 검사: grep -c 'SPEC-CODEX-SIDECAR-GUARD-001' CHANGELOG.md → 0 (grep exit 1, 부합 없음)"
b12_self_test_b: "PASS — AC 수 일치: acceptance.md 고유 AC 식별자 8건 (AC-CSG-001..008) ↔ CHANGELOG 항목 기재 8건"
b12_self_test_c: "PASS — 항목이 주장하는 파일 경로 7종 전부 실존 확인 (ls)"
changelog_entry_position: "CHANGELOG.md [Unreleased] §Fixed 섹션 말미 — SPEC-VERSION-STAMP-GUARD-001(t388) 항목 다음, ## [3.1.3] 직전"
frontmatter_status_transitions:
  spec_md: "in-progress → completed (단일 sync 커밋 병합 전환 — manager-docs 소관)"
mx_scan_note: "시험 전용 diff (internal/cli/init_agent_flag_test.go +7/-4) — 프로덕션 심볼 변화 0건, @MX 태그 추가/갱신/제거 0건 (zero-change 기록)"
docs_site_and_readme: "변경 없음 — 이 시험 단언 강화를 다루는 README/docs-site 페이지 부재 (리드 지시: docs-site 불가)"
```

## §F Phase 4 Mode Selection

Decision: serial

| 항목 | 값 |
|---|---|
| tier | S |
| scope (파일 수) | 1 — `internal/cli/init_agent_flag_test.go` |
| domain 수 | 1 (Go 테스트 코드) |
| 동시성 이점 | 낮음 — 코딩 중심 단일 파일 |
| Kickoff | 승인 (운영자, 자율 모드 — 리드 경유, 2026-09-01) |
| 스폰 모델 | GLM 상속 (운영자 묶음 승인 — manager-develop · plan-auditor 포함) |

| 모드 | 선택 | 근거 |
|---|---|---|
| direct | 아니오 | 기계적 변경이 아님 — 격리 뮤턴트 관측 판단 포함 |
| fanout | 아니오 | 다중 도메인 연구가 아님 |
| sweep | 아니오 | 고볼륨 기계 변환이 아님 |
| agent-team | 아니오 | 운영자 명시 요청 없음 |
| **serial** | **예** | 코딩 중심 단일 파일 — 병렬화 주의사항의 기본 경로 |

근거: 단일 파일에 단언 2줄 추가 + 격리 뮤턴트 관측 2종으로, 병렬화로 얻을 것이 없다.
coding-heavy 작업의 기본값인 serial을 따른다. Tier S이므로 위임 프롬프트는 최소 형태를 쓴다.

경계 사례: 없음 — 모든 축이 임계값에서 명확히 떨어져 있다.
