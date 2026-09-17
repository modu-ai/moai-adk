# SPEC-CODEX-AUDIT-GATE-AXES-001 — 진행 기록

카드 **t686** · 이슈 **modu-ai/moai-adk#1632** 잔여분 · 브랜치 `WT-codex-audit-gate`

## §E.1 Plan-phase Audit-Ready Signal

- 산출: `spec.md` · `plan.md` · `acceptance.md` · `progress.md` (Tier M). 요구사항 16, AC 14.
- 기준 트리: develop @ `f67d2193f`.
- 축 (c) `auth_provider: "unknown"` 은 v0.3.0 에서 카드 **t870** 으로 분리(제보자 입력 대기). 이 SPEC 에는 요구사항·AC·마일스톤·파일이 남아 있지 않다.
- 확정 결정(운영자, 2026-09-18): B-1 영수증, B.3 `verdict: fail`+`gate_unmet`+`isError:false`, K1 표면 S1(SubagentStop), N2 시작 표식 + 거부 기록 영속 + PreToolUse `Agent|Task` 소비자(`internal/hook/pre_tool.go:632-640` 경로에 형제 가드, 배선 `.claude/settings.json:69` PreToolUse `"matcher": "Agent|Task"`).
- plan-audit: iter1 FAIL 0.74(D1-D12 반영, v0.2.0), iter2 FAIL 0.82(N1-N4, O1-O4 반영, v0.3.0), iter3 최종 문구 수정 R1·R2·K4·O5-O10 반영(v0.3.1). 운영자가 이 수정 후 Implementation Kickoff 를 승인했고 추가 감사는 없다.
- 남은 열린 결정: 없음. M1 정지 규칙(페이로드에 `agent_type`/`last_assistant_message`/`agent_id` 부재 시 S2+S3 로 되돌리고 리드 보고)이 유일한 조건부 분기.
- 이 세션이 관측하지 않은 것: 테스트 미실행(변경 전 초록 기준선 미확인), SubagentStart/Stop 런타임 페이로드 미관측(필드는 `internal/hook/types.go:212,230,238-241` 선언만 확인).

## §E.2 Run-phase Evidence

모든 명령은 워크트리 `.claude/worktrees/t686`(브랜치 `WT-codex-audit-gate`)에서, 이번 실행에서 직접 돌린 것이다. 증거 파일은 세션 scratchpad `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/e5510854-2e36-4553-ad99-cb9ea452333f/scratchpad/` 에 있으며 세션이 끝나면 사라진다 — 따라서 판독에 필요한 출력은 아래에 직접 옮겨 적었다.

### E.2.0 변경 전 기준선 (plan.md §C)

| Claim | Evidence (command → observed) | Baseline-attribution |
|---|---|---|
| 변경 전 cli 선택 스위트 초록 | `go test ./internal/cli/ -run 'CodexAudit_\|CodexBlankReview_\|Converge_\|RunMultiAudit_\|AuditMulti_\|ReviewGate_\|MultiReviewGate' -count=1 -v` → `=== RUN` 185줄, `ok github.com/modu-ai/moai-adk/internal/cli 9.509s` | HEAD `b28b9f6e1`, 이번 실행, 이 트리 |
| 변경 전 hook 선택 스위트 초록 | `go test ./internal/hook/ -run 'SubagentStart\|SubagentStop\|PreTool\|AgentModel' -count=1 -v` → `=== RUN` 187줄, `ok github.com/modu-ai/moai-adk/internal/hook 1.880s` | 같음 |

### E.2.1 M1 — 훅 페이로드 실측 (정지 규칙 통과)

**측정 방법**: 프로젝트 트리를 건드리지 않는 일회성 headless 세션. `--setting-sources user` + `--settings <scratchpad>/m1probe/.claude/settings.json` 으로 이 저장소의 훅 배선을 로드하지 않고, 훅 커맨드는 stdin 을 scratchpad 파일에 append 만 한다. 프로브용 에이전트는 `--agents` 로 인라인 정의했다(프로젝트 `.claude/agents/` 미수정). 사용 명령:

```
timeout 300 claude -p "Spawn the plan-auditor subagent exactly once with the prompt probe, then reply DONE." \
  --setting-sources user --settings <scratchpad>/m1probe/.claude/settings.json \
  --agents '{"plan-auditor":{...,"tools":["Read"]}}' --allowedTools "Agent,Task,Read" --max-turns 6
```

관측된 페이로드(비밀값 없음, transcript 경로는 그대로 포함):

```
PreToolUse   : {'session_id': '1ce17461-…', 'cwd': '/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t686',
                'hook_event_name': 'PreToolUse', 'tool_name': 'Agent',
                'tool_input': {'description': 'Probe run', 'prompt': 'probe',
                               'subagent_type': 'plan-auditor', 'run_in_background': True},
                'tool_use_id': 'toolu_012LdkrsG4dHuqfmU7NZaGJe'}
SubagentStart: {'session_id': '1ce17461-…', 'cwd': '…/.claude/worktrees/t686',
                'agent_id': 'abf8e739c4b64da11', 'agent_type': 'plan-auditor',
                'hook_event_name': 'SubagentStart'}
SubagentStop : {'session_id': '1ce17461-…', 'cwd': '…/.claude/worktrees/t686',
                'agent_id': 'abf8e739c4b64da11', 'agent_type': 'plan-auditor',
                'hook_event_name': 'SubagentStop', 'stop_hook_active': False,
                'agent_transcript_path': '…/subagents/agent-abf8e739c4b64da11.jsonl',
                'last_assistant_message': 'AUDIT-VERDICT: PASS spec=SPEC-PROBE-001 receipts=none',
                'background_tasks': [{…'agent_type': 'plan-auditor'}]}
```

**정지 규칙 판정: 통과.** `agent_type`(Start·Stop 양쪽), `agent_id`(양쪽 동일 값 `abf8e739c4b64da11` — 두 이벤트를 잇는다), `cwd`(양쪽), `last_assistant_message`(Stop), `stop_hook_active`(Stop) 모두 존재. 따라서 축 (b)를 S1 표면으로 구현했고 S2+S3 로 되돌리지 않았다. 부수 관측: PreToolUse `tool_name` 은 `Agent`, 스폰 대상은 `tool_input.subagent_type` 으로 온다(가드가 읽는 경로와 일치). 프로브 디렉터리·캡처 파일은 scratchpad 안에만 있고 작업 트리에는 아무것도 남기지 않았다.

### E.2.2 M2 — 골든 캡처 (구현 전 커밋)

| Claim | Evidence | Baseline-attribution |
|---|---|---|
| 비-required 골든이 축 (a) 구현 이전 커밋에 있다 | `git log --diff-filter=A --format=%H -- internal/cli/testdata/codex-audit-nonrequired/off.golden` → `6d36579ce56b544c9f0c8ba956d21aba68347f47`; `git merge-base --is-ancestor 6d36579ce… 6831ed3b3` → exit 0 | 커밋 `6d36579ce`(골든) vs `6831ed3b3`(축 a) |
| 골든 7경우 캡처 | `UPDATE_CODEX_AUDIT_GOLDEN=1 go test ./internal/cli/ -run TestCodexAudit_NonRequiredGateGoldenByteIdentical -count=1 -v` → 7 subtest PASS (off / advisory / key-absent / file-absent / corrupt-yaml / `"required "` / `REQUIRED`) | HEAD `6d36579ce` 직전 트리(=`b28b9f6e1`) |

### E.2.3 M3 — 축 (a) RED (§E 항목 E8, GREEN 이전 verbatim)

```
$ go test ./internal/cli/ -run 'TestCodexAudit_(RequiredGateBlocks|EngineDefaultRequiredIsNotOptIn|RequiredGateLeavesRealVerdicts)' -count=1 -v
    codex_audit_required_block_test.go:43: verdict = "inconclusive", want fail — an explicitly required gate left without a verdict must block
    codex_audit_required_block_test.go:43: summary = "codex unavailable: codex binary not found in PATH", want it to name the unmet required gate
--- FAIL: TestCodexAudit_RequiredGateBlocksWhenBinaryAbsent (0.00s)
    codex_audit_required_block_test.go:57: verdict = "inconclusive", want fail …
    codex_audit_required_block_test.go:57: summary = "codex unavailable: rpc transport closed", want it to name the unmet required gate
--- FAIL: TestCodexAudit_RequiredGateBlocksOnRPCFailure (0.00s)
    codex_audit_required_block_test.go:66: verdict = "inconclusive", want fail …
    codex_audit_required_block_test.go:66: summary = "codex review output was blank: no verdict text was produced", want it to name the unmet required gate
--- FAIL: TestCodexAudit_RequiredGateBlocksOnBlankOutput (0.00s)
--- PASS: TestCodexAudit_EngineDefaultRequiredIsNotOptIn (0.00s)
--- PASS: TestCodexAudit_RequiredGateLeavesRealVerdicts (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.824s
```

`=== RUN` 7줄(빈 스윕 아님). AC-CAG-004·005 는 회귀 가드라 RED 시점에도 초록이 정상이다. 구현(`applyGateUnmet`) 후 같은 선택자 전부 PASS, 그리고 지명된 두 단언만 뒤집혔음을 관측했다:

```
codex_audit_gate_unmet_test.go:43: verdict = "fail", want "inconclusive" (fail-open verdict is preserved, not rewritten)
codex_blank_review_test.go:419: verdict = "fail", want "inconclusive"
```

축 (b) RED: 저장소 패키지는 빌드 실패(`undefined: RootSourceArgument`, `undefined: stateRel` …)로 시작했고, 훅 가드는 `audit_receipt_guard_test.go:340: blank: output = &{… Decision: …}, want a block naming a missing verdict line` 로 시작했다.

### E.2.4 AC 매트릭스

| AC | Status | Verification command | Actual output |
|----|--------|----------------------|---------------|
| AC-CAG-001 | PASS | `go test ./internal/cli/ -run TestCodexAudit_RequiredGateBlocksWhenBinaryAbsent -count=1` | `ok … internal/cli` (verdict fail + gate_unmet + summary 에 "codex binary not found" 병기) |
| AC-CAG-002 | PASS | `go test ./internal/cli/ -run 'TestCodexAudit_RequiredGateBlocksOnRPCFailure\|TestCodexAudit_RequiredGateBlocksOnBlankOutput' -count=1` + `TestCodexBlankReview_AC007…` | `ok … internal/cli`; AC007 은 `gate_unmet` 단언 유지, verdict 단언만 `fail` |
| AC-CAG-003 | PASS | `go test ./internal/cli/ -run TestCodexAudit_NonRequiredGateGoldenByteIdentical -count=1 -v` (축 (b) 착지 후 재실행) + 조상 검사 | 7/7 PASS, 골든과 바이트 동일(영수증 필드 없음); `git merge-base --is-ancestor` exit 0 |
| AC-CAG-004 | PASS | `go test ./internal/cli/ -run TestCodexAudit_EngineDefaultRequiredIsNotOptIn -count=1 -v` | PASS — 전제 단언(엔진 기본값 = required) 통과 후 verdict inconclusive, `gate_unmet` 키 부재 |
| AC-CAG-005 | PASS | `go test ./internal/cli/ -run TestCodexAudit_RequiredGateLeavesRealVerdicts -count=1 -v` | `pass`/`fail` 두 subtest PASS, `gate_unmet` 공백 |
| AC-CAG-006 | PASS | (i) `go test -list "$R" ./internal/cli/` (ii) `go test -run "$R" -count=1 ./internal/cli/` (iii) `git diff --stat f67d2193f..HEAD -- <14파일>` (iv) `grep -lE "^func Test[A-Za-z0-9_]*($R)" internal/cli/*_test.go` | (i) 84줄, 뒤집힌 두 테스트 이름 0건 (ii) `ok … internal/cli 5.300s` (iii) 출력 없음 (iv) 15파일 = 기존 14 + 신규 `mcp_audit_receipt_test.go` |
| AC-CAG-007 | PASS | 7곳 판독 + `make agents-emit-check` | 도구 설명·감사자 본문 4곳(로컬/템플릿)·스킬 2곳(로컬/템플릿) 모두 단일 도구 required 차단 + 판정 줄 의무 기재, "convergence engine reuses" 문구 제거; `make agents-emit-check` → `ok … internal/template/agentemit 0.380s` |
| AC-CAG-008 | PASS | `go test ./internal/cli/ -run 'TestCodexAudit_RecordsReceipt\|TestAuditMulti_RecordsReceipt\|TestCodexAudit_FallbackRootedReceipt' -count=1 -v` | 전 subtest PASS — required 트리만 `audit_receipt` 노출, advisory 트리는 저장만, codex off 는 영수증 0건, 폴백 루트는 `root_source: fallback` |
| AC-CAG-009 | PASS | `go test ./internal/hook/ -run TestSubagentStart_WritesAuditorStartMarker -count=1 -v` | PASS — plan/sync-auditor 만 `starts/<agent_id>.json`, `tree_root` = git toplevel 정규화, manager-develop 은 파일 없음 |
| AC-CAG-010 | PASS | `go test ./internal/hook/ -run TestSubagentStop_BlocksUnprovenPassAndPersistsRejection -count=1 -v` | 두 역할 × 5원인(인용 없음/저장소 없음/다른 트리/시작 이전/표식 없음) 모두 `decision: block` + 원인 명시 + `rejections/<role>--SPEC-X-001.json` 존재 |
| AC-CAG-011 | PASS | `go test ./internal/hook/ -run TestSubagentStop_ReentryWarnsAcceptanceClearsRoleFailIsInert -count=1 -v` | PASS — 재진입은 decision 없음 + systemMessage + `reentry_warned: true`; 유효 PASS 가 plan-auditor 3건 전부 제거하고 sync-auditor 기록은 잔존; FAIL 은 무변화 |
| AC-CAG-012 | PASS | `go test ./internal/hook/ -run TestPreToolUse_DeniesPhaseEntrySpawnsWhileRejectionOutstanding -count=1 -v` | PASS — manager-develop/docs/git × Agent·Task 6조합 deny(`AUDIT_RECEIPT_VIOLATION` 접두 + 역할/SPEC/원인), Explore·깨끗한 트리·advisory 트리는 통과 |
| AC-CAG-013 | PASS | `go test ./internal/hook/ -run TestAuditReceiptGuard_NonRequiredTreesAreInert -count=1 -v` + 누출 검사(아래 E.2.5) | PASS — off/advisory/부재 세 트리 모두 출력 없음, `.moai/state/audit-receipts/` 미생성 |
| AC-CAG-014 | PASS | `go test ./internal/hook/ -run TestAuditReceiptGuard_UnreadableEvidence -count=1 -v` | PASS — (i) 손상 영수증 → block + "receipt unreadable" (ii)(iii) 공백·형식 불일치 → block + "verdict line missing" + `unknown-spec` 기록 + 직후 manager-develop deny 에 `unknown-spec` 포함, 재진입은 차단 없이 기록 유지 (iv) 손상 거부 기록 → deny + 파일 경로 명시 (v) 디렉터리 부재 → 통과 (vi) 끝 공백·`\r` PASS → decision 없음 |

### E.2.5 누출 검사 (AC-CAG-013 (a)(b) + R2 유효성 증거)

| 단계 | 명령 | 관측 |
|---|---|---|
| 전 | `find . -path ./.git -prune -o -type d -name .moai -print \| sort`, `git status --porcelain --ignored` | 기준선 저장 |
| 실행 | 이 SPEC 이 추가한 테스트만 `-run` 선택자로 실행(cli `=== RUN` 23줄 `ok … 0.733s`, hook `=== RUN` 18줄 `ok … 4.160s`) | 빈 스윕 아님 |
| 후 | 같은 두 명령 | 두 diff 모두 **빈 출력**(차이 없음) |
| R2 | 일부러 `.moai/state/audit-receipts/` 를 패키지 디렉터리에 쓰는 임시 테스트 `zz_leakprobe_test.go` 를 한 번 실행 | 검사가 차이를 보고함 — dirs diff `> ./internal/cli/.moai`, status diff `> !! internal/cli/.moai/`, 그리고 내장 residue 가드가 `RESIDUE GUARD FAIL: this test run created …/internal/cli/.moai` 로 exit 1. 임시 테스트와 잔재는 즉시 제거했고, 제거 후 `find` 결과는 기준선과 바이트 동일 |

### E.2.6 품질 게이트

| 항목 | 명령 | 관측 |
|---|---|---|
| vet | `go vet ./internal/cli/ ./internal/hook/ ./internal/auditreceipt/` | 출력 없음(0건) |
| lint | `golangci-lint run ./internal/cli/... ./internal/hook/... ./internal/auditreceipt/...` | 신규 패키지 `0 issues.`; 저장소 전체로는 기존 2건(`internal/cli/gtd_answer.go:84` S1038, `internal/cli/launcher.go:811` S1021)만 남으며 이번 변경이 만든 것이 아니다(두 파일 미수정) |
| 커버리지 | `go test ./internal/auditreceipt/ -cover -count=1` | `coverage: 87.3% of statements` (목표 85% 상회) |
| gofmt | `gofmt -l <이번 변경 파일들>` | 출력 없음 |
| 빌드·미러 | `go build ./...`, `make agents-emit`, `make agents-emit-check`, `make build` | 모두 성공; 방출본 2개(.codex toml)와 `internal/template/catalog.yaml` 해시가 같은 커밋에 동반 |
| 최종 회귀 | `go test ./internal/cli/ -run '<기준선 선택자>' -count=1 -v` / `go test ./internal/hook/ -run 'SubagentStart\|SubagentStop\|PreTool\|AgentModel\|AuditReceipt' -count=1 -v` | cli `=== RUN` 208줄 `ok … 9.439s`, hook `=== RUN` 205줄 `ok … 5.357s` (기준선 185/187 → 신규 테스트만큼 증가) |

### E.2.7 Gaps (관측하지 않은 것)

- **이 저장소의 `.moai/config/sections/workflow.yaml` 은 `workflow.audit.gates` 를 설정하지 않는다.** 따라서 축 (b)의 required 경로는 이 저장소의 실제 세션에서 발동하지 않으며, 검증은 전부 `t.TempDir()` 픽스처 트리로만 이뤄졌다(라이브 세션에서 SubagentStop 차단이 실제로 걸리는 것은 관측하지 않았다).
- 전체 스위트(`go test ./...`, `internal/cli`·`internal/hook` 전 패키지)는 로컬에서 돌리지 않았다 — 머신 부하 정책에 따라 전체 판정은 리드 push 후 CI 몫이다.
- 크로스플랫폼 빌드(`GOOS=windows`)는 이번 실행에서 측정하지 않았다(테스트 바이너리 미컴파일 한계 포함) — CI 매트릭스 소관.
- M1 실측은 headless 세션 1회, plan-auditor 1개 인스턴스에 대한 관측이다. sync-auditor 스폰, 팀 모드, `stop_hook_active: true` 재진입의 실제 런타임 페이로드는 관측하지 않았다(테스트는 픽스처로 덮었다).
- codex 바이너리 실물로 `codex_audit` 를 돌린 적은 없다(모든 경로가 seam 이중화).

### E.2.8 Residual-risk

- 시작 표식을 **첫 종료 차단 시에는 남긴다**(plan.md §B.5 "Stop 처리 후 표식 제거"의 좁은 해석 변경). 차단 뒤 감사자가 이어서 codex 를 부르고 재진입에서 유효 PASS 를 내려면 그 인스턴스의 표식이 살아 있어야 하기 때문이다. 표식은 수용·FAIL·재진입 시에 제거된다. 부작용: 차단된 인스턴스의 표식은 감사자가 끝내 재진입하지 않으면 남는다(다음 유효 PASS 나 사람의 삭제로 정리).
- 거부 기록은 역할 단위 해제라, 한 SPEC 의 유효 PASS 가 같은 역할의 다른 SPEC 거부까지 지운다(spec.md §C 에 이미 기록된 잔여 위험).
- `agent_id` 로 파일명을 만들 때 허용 문자 밖은 `_` 로 치환한다. 서로 다른 두 `agent_id` 가 같은 파일명으로 정규화되면 표식이 덮인다(런타임 id 는 16진 문자열이라 실제로는 발생하지 않는다).
- 영수증 기록 실패는 로그만 남기고 감사 결과를 바꾸지 않는다(디스크 문제를 판정으로 번역하지 않기 위함). 그 상태에서 required 트리 감사자는 인용할 영수증이 없어 차단된다 — 의도한 방향이지만, 디스크 장애가 감사 중단으로 나타난다.
- `internal/cli` 테스트의 폴백 루트는 `TestMain` seam 으로 임시 디렉터리에 고정했다. 이 seam 이 제거되면 `project_root` 없이 도구를 부르는 기존 테스트가 저장소 트리에 영수증을 쓴다(그 경우 residue 가드가 잡는다 — R2 에서 실측).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-18
run_commit_sha: pending-backfill-run   # 이 절을 담는 커밋 자신의 SHA는 커밋 전에 알 수 없다
run_status: audit-ready
milestone_commits:
  M2_golden: 6d36579ce
  M3_axis_a: 6831ed3b3
  M4_axis_b: ffe36d8e3
  M5_docs_mirrors: 5d160d95a
ac_pass_count: 14
ac_fail_count: 0
m1_payload_measurement: pass            # agent_type / agent_id / cwd / last_assistant_message / stop_hook_active 전부 관측
preserve_list_post_run_count: 14        # AC-CAG-006 의 14개 파일, git diff --stat 출력 없음
new_warnings_or_lints_introduced: 0     # 신규 패키지 0 issues, 기존 2건은 미수정 파일
coverage_new_package: 87.3
cross_platform_build:
  local_build: pass                     # go build ./...
  windows_cross: unmeasured             # CI 매트릭스 소관
total_run_phase_files: 36               # git diff --stat 6d36579ce^..HEAD
mN_commit_strategy: milestone-per-commit (M2 -> M3 -> M4 -> M5), 미푸시
push_state: NOT PUSHED                  # 리드 일괄 push 대기
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-18
sync_commit_sha: 391113875   # completes the 3-phase close (spec.md status -> completed)
sync_status: audit-ready
b12_self_test_a: pass   # grep -c 'CODEX-AUDIT-GATE-AXES-001' CHANGELOG.md == 0 before emission (checked pre-edit)
b12_self_test_b: pass   # acceptance.md 고유 AC id 14개, CHANGELOG 본문 "14/14 acceptance criteria PASS"로 동일 수 인용
b12_self_test_c: pass   # 인용 경로 spec.md·internal/auditreceipt 존재 확인(ls)
changelog_entry_position: "### Fixed, [Unreleased] 절 최상단 항목"
frontmatter_status_transitions:
  spec_md: "in-progress -> completed (status 필드만, updated 불변 2026-09-18) — 단일 sync 커밋에 3-phase close 적용"
  plan_md: "frontmatter 없음 — 대상 아님"
  acceptance_md: "frontmatter 없음 — 대상 아님"
  completed_transition: "단일 sync 커밋(19d90298b 및 이 백필 커밋)에서 spec.md status를 completed로 전이 — SPEC 본문(§A-§H)은 미수정, frontmatter status/updated만 대상"
canary_compliance_check: n/a   # 이 SPEC은 미래를 향한 정책을 스스로 시험하지 않음
docs_site_readme_check:
  searched: "grep -rl 'codex_audit|audit.gates.codex|workflow.audit.gates' docs-site/content/en; README*.md"
  found: "docs-site/content/en/advanced/config-sections.md (tool/model-effort pinning listing only), docs-site/content/en/guides/mcp-server.md (tool catalog + fail-open description), docs-site/content/en/advanced/multi-model-audit.md (audit_multi convergence semantics only — REQ-CAG-006 unchanged), README*.md (tool-name table row only)"
  disposition: "no docs-site or README change made — none of the found pages describe single-codex_audit-path required-gate enforcement or the audit-receipt mechanism; all describe either the unchanged audit_multi convergence algorithm or a bare tool-name listing"
```
