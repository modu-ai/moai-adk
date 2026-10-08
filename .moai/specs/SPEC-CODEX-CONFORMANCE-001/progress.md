# SPEC-CODEX-CONFORMANCE-001 — progress.md

status: in-progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-09
plan_audit: iter 1 FAIL 0.875 (MP-8 fail, blocking 6, required backend codex fail — receipt rcpt-f829a5596d422b6433546552, verdict `.moai/reports/t1607/plan-audit.md`); iter 2 repair applied — fixes 1-7(D1-D8) + lane inputs (D6 CI-trigger correction, 0.161.0 하위명령 실측 해소); iter 2 재감사는 감사자 429 사망으로 미완료; plan amendment round (2026-10-08T17:12:42Z) — 레인 실측(2026-10-08T16:56Z+) 접기: 제너레이터 출력 형상(39파일·v1/v2 분할·소비 8종 전부 존재, 이전 감사의 "이름 누락"은 ls 별칭 인공물로 반증) + vendoring 구성 결정(소비 부분집합, decision-index Q7 기본 적용); iter 2 FAIL 0.94 (MP-8 잔여 AC-CONF-005 + N2/N3/N4, receipt rcpt-2f21e144af2e5bef92bb82f6) — round-3 repairs applied (N4 005 채용 지연+기준선 원리 정정+Q4 앵커 갱신, N2 양세계 family-green 완결 경로, N3 죽은 출처 참조 제거, O2 스탬프 UTC 정정, O3 스테이징 디렉터)
artifacts: spec.md, plan.md, acceptance.md, decision-index.md, progress.md (Tier M set + decision gate on)
red_now_ledger: acceptance.md §B (LEDGER-1/2/3 — tree 81786284e; AC-CONF-006의 RED 채용은 M3 E8 시점)
open_items: decision-index Q4 (EVIDENCE-NEEDED — 0.161.0 login-status 출력; 미인증 형태 포획은 AC-CONF-005의 최소 필수 관측으로 강화, keyring 변형만 기록 갭 허용)

## §E.2 Run-phase Evidence

### M1 — fixture 재생성 + 소비자 이전 (2026-10-09, tree dce9596be 이후 본 카드 브랜치)

**생성 환경 관측 (npx -y @openai/codex@0.161.0, lane 지정 핀)**:

- `npx -y @openai/codex@0.161.0 --version` → stdout `codex-cli 0.161.0`, exit 0 (본 실행 재관측; 레인 선행 실측과 일치).
- `npx -y @openai/codex@0.161.0 app-server generate-json-schema --out /tmp/t1607-m1/schema-staging` → exit 0. 출력: 최상위 39파일 + `v1/`(2파일) + `v2/`(274파일) = 총 315파일, 4.3MB. 소비 8종 Response 스키마는 최상위에 전부 존재(레인 선행 관측과 일치; plan §A "실파일 39개"는 최상위 기준).
- **소비 8종 바이트 동일성**: `cmp`로 0.160.0 vendored 세트와 전수 대조 → 8종 전부 IDENTICAL (ApplyPatchApprovalResponse/CommandExecutionRequestApprovalResponse/DynamicToolCallResponse/ExecCommandApprovalResponse/FileChangeRequestApprovalResponse/JSONRPCError/McpServerElicitationRequestResponse/PermissionsRequestApprovalResponse). 0.160.0 → 0.161.0 응답 스키마 불변 — M1 전환이 동작 보존임을 기계가 증명.
- **REQ-CONF-004 판정 재료 — resume --help 옵션 집합 관측**: `npx -y @openai/codex@0.161.0 resume --help` → exit 0, 115행. `--remote <ADDR>`(37행)와 `--remote-auth-token-env <ENV_VAR>`(42행) **둘 다 존재** → **옵션 존재 세계** 확정. `real_help_supported` 기대값 갱신 불요(acceptance.md §C AC-CONF-003의 첫 Then이 완결 경로). 신규 캡처 본문은 커밋된 0.160.0 resume-help.txt와 `diff` exit 0 = 바이트 동일.
- vendoring: 소비 8종을 스테이징에서 이름 지정 기계 복사(내용 편집 0) + resume-help.txt 캡처본 + README 재생성(0.161.0 명기, 생성 명령, 부분집합 구성 근거) → `internal/cli/testdata/codex-0.161.0/` 8+2 구성. `codex-0.160.0/` 디렉터 삭제. decision-index Q7(소비 부분집합)·Q1(교체) 기본 적용 이행.

**RED 원문 (소비자 이전 직후·vendoring 전, tree dce9596be)**:

```text
$ go test ./internal/cli -run '^(TestManagedServerRequestPolicyMatchesCodexSchema|TestManagedCodexRemoteSupportProbe|TestManagedCodexServerRequestPolicy)$' -count=1
--- FAIL: TestManagedCodexRemoteSupportProbe (0.00s)
    --- FAIL: TestManagedCodexRemoteSupportProbe/real_help_supported (0.00s)
        managed_codex_tui_test.go:1057: open testdata/codex-0.161.0/resume-help.txt: no such file or directory
--- FAIL: TestManagedServerRequestPolicyMatchesCodexSchema (0.00s)
    managed_hardening_test.go:626: read vendored schema JSONRPCError.json: open testdata/codex-0.161.0/JSONRPCError.json: no such file or directory
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.835s
```

(acceptance.md §B가 예고한 두 실패 입력 그대로 — resume-help `os.ReadFile` 실패 + `read vendored schema` 실패. `TestManagedCodexServerRequestPolicy`는 fixture 디렉터를 직접 읽지 않아 적색 대상이 아니며, 그 0.161.0 정합은 AC-CONF-002/003의 가드가 대리한다.)

**GREEN 원문 (vendoring 후)**:

```text
$ go test ./internal/cli -run '^(TestManagedServerRequestPolicyMatchesCodexSchema|TestManagedCodexRemoteSupportProbe|TestManagedCodexServerRequestPolicy)$' -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	2.100s
```

**M1 폐쇄 관측 (두 세계 공통 요구 — family green)**:

```text
$ go test ./internal/cli -run '^TestManagedCodexTUIPreconditionsAndFallback$' -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	13.329s
$ grep -rn "codex-0.160.0" internal/cli | wc -l
       0
```

(프로브는 실제 새 help 텍스트를 입력으로 실행됐다 — fake의 `tuiFakeHelpFileEnv`가 `tuiResumeHelpFixture` 절대경로를 읽고, 그 fixture가 이제 0.161.0 캡처이다. 합성 텍스트만의 녹색 아님.)

**빌드**: `go build ./...` exit 0 / `GOOS=windows GOARCH=amd64 go build ./...` exit 0.

**AC-CONF-005 예비 관측 (M2 소관이지만 동일 생성 환경에서 포획 — REQ-CONF-007 최소 필수 관측)**:

- 명령: `CODEX_HOME=<빈 디렉터> npx -y @openai/codex@0.161.0 login status` (미인증 상태 보장 — 격리 CODEX_HOME)
- verbatim 출력: **stderr에 `Not logged in` 한 줄**, stdout 빈 송출, **exit 1**
- 바이너리 버전: codex-cli 0.161.0
- 판정 재료: 이 출력은 stage 2 전체라인 문법 `logged in using (chatgpt|api key)`와 불일치 → `parseCodexAuthLine`은 `codexAuthUnknown`(갭)이어야 하며, "미인증" 판정을 내선 안 된다 — REQ-CONF-006 계약과 정합. M2가 이 포획본을 vendored 샘플로 삼아 분류 시험으로 봉인한다.


## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

- Input parameters: tier M; scope ~6 files (fixture dir + 2 test consumers + 1 ops doc + SPEC artifacts); domains 1 (Go CLI adapter + its tests); language mix Go + machine-generated fixtures; concurrency benefit LOW (coding-heavy, M1 fixture regeneration gates M2/M3); Agent Teams prereqs n/a (explicit-request-only).
- Mode evaluation: direct — not selected (semantic multi-milestone work); serial — selected; fanout — not selected (not research-heavy, single domain); sweep — not selected (semantic work, not mechanical-uniform bulk); agent-team — not selected (no operator request).
- Decision: serial
- Justification: single-domain coding-heavy conformance work with a strict milestone dependency chain — sequential manager-develop delegation with per-milestone commits is the safe default per the coding-task parallelism caveat; no other mode's selection criteria are unambiguously met.
- Kickoff record (autonomous form): decision record: decided_by=lane-21 evidence_refs=.moai/reports/t1607/plan-audit.md (iter-4 PASS 1.0, must_pass 0, blocking 0; codex required backend pass, governing receipt rcpt-f16b809f87c02e929297a2e9 fresh on the unchanged final state; plan_artifact_hash 93f2e64f8bbb38ba93a0ec664ddfa00796321566a9300da3f7223b115f34b795 via the canonical runtime.ComputeHash recipe — name:len:NUL+bytes+NUL per planArtifactNames order — unchanged since verdict; the earlier 21aba6a7 figure was the concatenation recipe and is superseded per gate review) ladder_path=plan-to-run Kickoff autonomous form — verdict PASS + 1.0 ≥ 0.80 Tier M threshold + artifact hash unchanged + no open blocker — run-phase entry approved.
