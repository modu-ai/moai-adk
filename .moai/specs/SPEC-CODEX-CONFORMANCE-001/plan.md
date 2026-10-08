# SPEC-CODEX-CONFORMANCE-001 — plan.md

---
id: SPEC-CODEX-CONFORMANCE-001
title: "Codex CLI 0.161.0 adapter conformance re-measure — implementation plan"
version: "0.1.0"
created: 2026-10-09
updated: 2026-10-09
---

## §A. Context

- 카드: t1607 (CARD-4, Codex 0.161.0 어댑터 정합성 재측정). 트리: 카드 워크트리 `.moai/worktrees/t1607`, 브랜치 `WT-codex-0161-conformance`, 계획 기준 HEAD `81786284e`.
- 상위 근거: `.moai/reports/release-update-20261008/upstream-update-20261008.md` 축 2 / `summary.json` CARD-4 / `codex-0.161.0-release-body.md`(14불릿). 상위 리포는 미추적 산출물이므로 본 plan의 인용은 방향 근거이고, 모든 코드 사실은 본 트리 실측으로 다시 잡았다(spec §A.2~§A.4).
- 개발 모드: `quality.yaml` `constitution.development_mode: tdd` → manager-develop `cycle_type=tdd`. 기존 동작 보존형 카드이므로 각 마일스톤은 보존 대상 시험의 green-now 기록(Pre-RED) 위에 새 검증을 얹는다.
- 환경 노트: fixture 재생성은 codex-cli 0.161.0 바이너리를 요구한다. 설치 바이너리는 0.160.1(카드 실측)이므로 M1 생성 단계에서 `npx -y @openai/codex@0.161.0 app-server generate-json-schema --out <staging-dir>`로 **스테이징 디렉터에** 생성한 뒤 소비 8종을 이름 지정 복사한다 — `--out`을 최종 디렉터로 쓰면 39파일이 그대로 놓여 AC-CONF-001의 8+2 검사에 걸린다(plan §F M1 구성 결정과 정합). 하위명령 존재는 **레인 실측으로 해소**(2026-10-08T16:14-16:15Z, 본 트리): `npx -y @openai/codex@0.161.0 --version` → `codex-cli 0.161.0`, exit 0 / `npx -y @openai/codex@0.161.0 app-server --help`에 `generate-json-schema` 확인, exit 0 — blocker 경로는 폐쇄됐고 M1이 생성 **출력**을 관측한다.
- **출력 형상 관측(레인 실측, 이번 실행·본 트리, 2026-10-08T16:56Z+)**: `npx -y @openai/codex@0.161.0 app-server generate-json-schema --out <dir>` → exit 0, **실파일 39개**(소비자가 읽는 형태의 나열이 아닌 ls 별칭 총계 행이 포함된 42가 아니라), **v1/v2 분할 레이아웃**, 소비 8종 이름 전부 존재(ApplyPatchApprovalResponse, CommandExecutionRequestApprovalResponse, DynamicToolCallResponse, ExecCommandApprovalResponse, FileChangeRequestApprovalResponse, JSONRPCError, McpServerElicitationRequestResponse, PermissionsRequestApprovalResponse). old-only = README.md + resume-help.txt(제너레이터 비생성물 — 예정대로 별도 캡처), new-only = 31파일. 전체 출력 = 4.3MB(커밋된 56K 디렉터 대비). **재-공황 방지 한 줄**: 이전 감사 패스가 제너레이터 출력에 8종 이름이 빠졌다고 읽은 것은 ls 별칭 인공물이고 이름은 존재한다 — 검증은 내용 수준(스키마 가드)이며 M1이 이미 그렇게 한다.

### §A.1 버전 드리프트 갱신 판정 (leader 필수 기재)

**삼방 상태**: 리포 fixture 핀 = 0.160.0 / 설치 바이너리 = 0.160.1 (2026-10-08T13:11Z 카드 실측) / 업스트림 안정 = 0.161.0 (2026-10-07T15:58:45Z 승격).

1. **목표 fixture 버전 = 0.161.0.** 0.160.1로의 부분 이동은 안정이 아닌 설치본을 따르는 실수이고, 0.161.0 이후 안정 재승격 시 다음 카드가 따라간다(다음 스윕 기준: 안정 > 0.161.0 또는 0.162 승격 시 델타).
2. **0.160.0 → 0.161.0 전이 검증 = 교체(replace), 보존(retain) 아님.** 판단: 0.160.0 디렉터를 지우고 0.161.0 디렉터로 대체하며 소비자 리터럴 2곳(`hardenCompileSchema` 경로+URL, `tuiResumeHelpFixture`)을 이전한다. 근거: (a) 커밋된 fixture README와 `.moai/docs/factory-managed-session.md:120`의 재생성 규율이 "최소 지원 버전이 움직이면 사본을 다시 만든다"고 규정 — 누적이 아니라 대체가 계약이다; (b) 두 소비자는 디렉터명을 리터럴로 쥐므로 이중 디렉터는 두 번째 진실의 원천이 되고 스키마 드리프트를 숨긴다; (c) 보존해도 그것을 읽는 시험은 하나도 남지 않는다. decision-index Q1에 기록(구현 수준, 기본 적용).
3. **라이브 바이너리 측정 = 자동 AC 범위 밖.** fixture 기반 기계적 정합이 1차 판정면이다(현행 스위트 전체가 그렇다 — 현행 행동 보존). fixture가 판정할 수 없는 세 항목 — ① `codex login status` 0.161.0 출력 형태, ② 0.161.0 `resume --help` 옵션 집합(재생성 자체가 관측), ③ 재생 프레임 확장 필드의 실제 형태 — 는 fixture 생성 환경에서 1회 관측해 vendoring하거나, 불가하면 progress.md §E.2에 수동 검증으로 기록한다. 운영자 환경 의존 검사를 자동 AC로 쓰지 않는다. decision-index Q2에 기록(구현 수준, 기본 적용).

### §A.2 카드 앵커 정정 (구현자가 헷갈리지 않도록)

- 모델 핀 파일: `internal/config/closed_sets.go` (카드문의 `internal/cli/` 아님). `defaults.go` 핀 주석 블록은 :1545 근처.
- `codex_event_adaptation_test.go`는 훅-이벤트 매핑 시험이다. resume 재생 페이로드의 소비면은 관리 소유자 프레임 처리(`managed_codex_tui.go` fake-server 계열)이고, rollout 인출 표면은 `codex_role_fingerprint.go`다.

## §B. Known Issues (본 카드와 관련된 범주만)

- **B1 크로스플랫폼**: 시험 파일 손대면 `GOOS=windows go build ./...` 유지.
- **B4 frontmatter**: 스키마 SSOT 준수(canonical 12필드, snake_case 금지) — spec.md에 적용 완료.
- **B5 CI 3단**: spec-lint / golangci-lint / Test 별도 판정. 카드 워크트리에서 `internal/cli` 전체 스위트는 구조적 적색(경합·자식 env hang 전례) — 레인-로컬 판정은 명명 시험 계열로 한정하고 CI가 등판.
- **B8 작업트리 위생**: `git add`는 명시 경로만.
- **B11 서브에이전트 경계**: 구현 위임 시 blocker 보고 반환, 사용자 질문 금지.
- **범주 외 카드 고유**: 스키마 파일은 기계 생성 산출물이므로 손편집 금지 — 불일치가 나면 생성 단계부터 재검토한다(손으로 스키마를 맞추는 것은 가드를 무력화한다).

## §C. Pre-flight

```bash
git branch --show-current && git rev-parse --short HEAD   # WT-codex-0161-conformance 기대
grep -rn "codex-0.160.0" internal/cli | wc -l             # 이전 대상 리터럴 기준선 (3행: hardening 2 + tui 1)
grep -rn "codex-0.161.0" internal/cli | wc -l             # 0 기대 (RED-now LEDGER-1)
go test ./internal/cli -run '^(TestManagedServerRequestPolicyMatchesCodexSchema|TestManagedCodexRemoteSupportProbe|TestManagedCodexServerRequestPolicy)$' -count=1
go test ./internal/cli -run '^(TestClassifyCodexAuth_LadderIntegration|TestClassifyCodexAuth_RejectedAuthFileFallsBackToProbe|TestClassifyCodexAuth_UnreadableProbeIsAGap)$' -count=1   # AC-CONF-004 green-now 기준선 (auth 사다리 — 부재·기각·갭 하강 케이스 포함, codex_auth_ladder_test.go:523/568/595)
go test ./internal/config -count=1                         # 핀 회귀 가드 green-now 기준선
```

기준선 원리: 보존 AC(004/007/008)는 지금 green이어야 하고, 신설 AC(001~003)는 LEDGER-1/2/3가 RED-now다. 005의 차단 채용은 M2 최소 관측 시점, 006은 M3 E8 시점이다(acceptance.md 증거 장부 — LEDGER는 디렉터·참조 부재를 관측하지 auth 분류 실패를 관측하지 않는다). 기준선이 어긋나면 시작하지 않고 blocker 보고.

## §D. Constraints

- **PRESERVE**: `internal/cli/mcp_codex.go` auth 사다리 본문(부재=하강 계약은 변경 대상이 아니다 — 문법 파괴가 실측되지 않는 한 코드 손대지 않는다), 관리 기각 정책표, `internal/config/closed_sets.go`·`defaults.go`·`audit_models.go` 핀, `internal/template/templates/**`(무관), 타 SPEC 산출물.
- **금지**: fixture 손편집, 네트워크·codex 바이너리를 요구하는 자동 시험 신설, `--no-verify`, 라이브 바이너리 의존 AC, 카드 6개 범위 항목 밖 기능 추가(발견분은 비목표 또는 후속 기록).
- **필수**: Conventional Commits + 카드 id(t1607) + `🗿 MoAI` 트레일러, 마일스톤별 커밋.

## §E. Self-Verification

E1 AC PASS/FAIL 행렬(acceptance.md §D), E2 `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...`, E3 영향 패키지 커버리지 — 전체 패키지 스위트는 구조적 적색이므로 영향 계열만 스코핑한다: `TestManagedServerRequestPolicyMatchesCodexSchema`, `TestManagedCodexServerRequestPolicy`, `TestManagedCodexRemoteSupportProbe`, 관리 TUI 계열(`TestManagedCodexTUI*`), auth 사다리 계열(`codex_auth_ladder_test.go`, `mcp_codex_test.go`), `./internal/config` 전체. 예시 명령은 전 브랜치 앵커 형태로: `go test -cover ./internal/cli -run '^TestManagedCodexRemoteSupportProbe$' -count=1`, E4 하위에이전트 경계 grep, E5 lint(신규/기존 구분), E6 커밋 SHA·push 상태, E8 RED 원문(LEDGER-1/2 + 신규 시험의 사전 적색 출력). 각 항목 VCI 5섹션 형식 + 귀속 삼인조(command/verbatim output/tree SHA).

## §F. Milestones (결정 가역성 순 — 바뀔 가능성이 큰 결정을 앞에)

- **M1 — fixture 재생성 + 소비자 이전** (최고 변경 가능성: 새 vendored 데이터 + 시험 리터럴 2곳). 0.161.0 바이너리로 생성 → **vendoring 구성은 소비 부분집합**(decision-index Q7, 기본 적용): 소비 8종 Response 스키마를 제너레이터 출력에서 **이름 지정 기계 복사**(내용 편집 없음 — 손편집 금지 불변) + `resume-help.txt` 캡처 + README 재생성 — 커밋된 디렉터 구성(8+2)을 그대로 미러한다. 31개 신규 파일(request-side Params 동반, JSONRPC 봉투 계열, ToolRequestUserInput*, FuzzyFileSearch*, Attestation*/ChatgptAuthTokensRefresh*, RequestId, v1/v2 분할, 통합 번들 2종)은 소비자가 없어 벤더링하지 않는다 — 요청측 Params 동반이 나중에 필요해지면 README 규율대로 재벤더링한다. → `codex-0.160.0` 디렉터 삭제 → `hardenCompileSchema`(경로·URL)·`tuiResumeHelpFixture` 이전 → 명명 계열 3종 green. 부산물 관측: 0.161.0 `resume --help` 옵션 집합 기록(REQ-CONF-004 판정 재료), 생성 **출력** 관측(하위명령 존재·출력 형상은 레인 실측으로 해소 — plan §A). **옵션 제거 세계의 M1 폐쇄 경로**: resume-help가 두 옵션을 모두 제공하지 않으면 AC-CONF-003의 명시적 완결 경로대로 — `real_help_supported` 기대값을 관측 현실로 갱신(실제 새 help 원문 관측을 근거로 기록)하고 프로브+폴백(`TestManagedCodexTUIPreconditionsAndFallback`)을 실제 새 help 입력으로 실행해 **명명 계열 전부 green**을 만들며, 어댑터 후속은 blocker 보고로 남긴다.
- **M2 — auth/keyring 정합**: 부재-하강 계약의 명시 시험 확인(없으면 추가 — RED/GREEN 한 쌍) + M1 환경에서 포획한 0.161.0 `codex login status` 출력의 정화 샘플 vendoring·`parseCodexAuthLine` 분류 시험. 샘플 불가 시 REQ-CONF-007의 수동 검증 경로로 기록. keyring 저장 상태의 라이브 확인(운영자 환경 의존)은 수동 검증 항목.
- **M3 — resume 표면 재검증**: 재생 프레임 내성 시험(확장 필드를 실은 fake-server 재생 프레임 fixture 추가, REQ-CONF-005) + `codex_role_fingerprint.go` rollout 소비가 0.161.0 형상 변화를 겪는지 재확인(변화 없음이 기본 기대 — 발견 시 비목표/후속 기록, 본 카드에서 고치지 않는다).
- **M4 — 불변 표면 검증 + 문서**: `TestManagedCodexServerRequestPolicy` green, `internal/config` 핀 시험 green, `.moai/docs/factory-managed-session.md` 버전 표기 갱신(0.160.0 → 0.161.0), CHANGELOG는 sync 소관이므로 손대지 않음, 증거 취합.

## §G. Anti-Patterns

- 리포트의 수치·라인을 본 트리 측정 없이 인용(이 카드가 정정한 앵커 3건이 실례 — §A.2).
- auth 부재를 "미인증"으로 읽는 신규 분기 추가(계약 위반).
- 스키마 불일치를 손으로 고치기(기계 생성 산출물의 손편집).
- 0.162-alpha 테마(51xxx번대: thread 예측 프로토콜, compaction 전체컨텍스트, MCP elicitation 등)를 0.161.0에 끌여들이기 — 미탑재 확인된 관찰목록이다.
- 설치 바이너리(0.160.1)로 생성한 fixture를 0.161.0이라고 기록하기.

## §H. Cross-References

- 선행 SPEC: SPEC-FACTORY-MANAGED-HARDEN-001 (스키마 가드·기각 정책), SPEC-CODEX-LAUNCHER-001 (auth 2단 사다리, REQ-CL-008/009/010), SPEC-CODEX-RESUME-SCOPE-001 (resume 선택 계약), SPEC-MODEL-MATRIX-UPDATE-001 (`{gpt-6.1-sol, high}` 핀).
- 운영 문서: `.moai/docs/factory-managed-session.md` (재생성 규율 본문).
- 상위 관측: `.moai/reports/release-update-20261008/` (미추적 — 근거 방향 참조 전용).
