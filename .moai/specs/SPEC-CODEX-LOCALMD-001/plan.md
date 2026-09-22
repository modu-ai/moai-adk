# plan.md — SPEC-CODEX-LOCALMD-001

## A. Context

- 카드 t1078 (Tier M), 워크트리 `WT-codex-local-md`, 개발 모드: TDD (`codex_local_instructions_test.go`가 characterization base).
- 목표: `internal/cli/codex_launcher.go`의 `codexLocalDeveloperInstructionArgs`(:114-142)를 2-파일 로더로 확장해 `CLAUDE.local.md`도 `-c developer_instructions=` 단일 토큰으로 주입한다. 구조적 균일성은 기존 funnel(`codex_launcher.go:495`) 안에 머무는 것으로 무료로 상속된다(REQ-LMD-008).
- 이 변경은 기록된 계약("CLAUDE-local 상태는 Claude 전용" — 커밋 `6c647bbe2`, `CHANGELOG.md:77`, `AGENTS.md` §8/:264, `AGENTS.md.tmpl` §8/문단 :268-273, docs-site `codex-dual-harness.md:33`, 헬프 카피 "Codex-only")을 되돌린다. Gen1(SPEC-CODEX-INIT-001 REQ-CI-008)의 원 의도를 `-c` 채널로 회복하는 성격이다(research.md §4, C1).

## B. Known Issues (선결 결함·긴장)

- **계약 역전 (C1)**: 다중 표면이 "Claude-only"를 기록 중 — 커밋 `6c647bbe2`, `CHANGELOG.md:77`, 리포 `AGENTS.md` §8(:264), 템플릿 `AGENTS.md.tmpl` §8(문단 :268-273, "…are not forwarded to Codex." 문장 :273 — 본 트리 직접 판독; audit 리포트의 :272 표기와 1행 차이, 본 트리 측정 우선), docs-site `codex-dual-harness.md:33`, 헬프 카피 "Codex-only". coordinated-surface 계획은 §F M4/sync 인계 참조.
- **헬프 카피 핀 (C3 해소)**: `codex_launcher.go:334-335`는 status readout이 아니라 `codexCmd.Long` 헬프 카피가 맞다(직접 판독). `TestCodexLocalInstructions_DocumentedInLauncherHelp`(:141-147)가 `"Codex-only"`를 grep — 문구와 테스트가 같은 변경으로 움직여야 한다(REQ-LMD-009).
- **argv overflow 방어 부재**: `E2BIG|ErrArgListTooLong` grep 전 `*.go` 0건 — 리포 최초 방어(REQ-LMD-007).
- **TOCTOU**: Lstat→ReadFile 사이 심볼릭 링크 치환 시 guard 탈출 가능하다. 두 입력 모두 같은 descriptor를 안전하게 open한 뒤 fstat·read하도록 봉쇄한다.
- **status 표면**: `moai codex status`에는 local-instruction 행을 추가하지 않는다. 거부 조건은 런치 오류로 관측한다.
- **긴 옵션 누락**: 설치된 `codex exec --help`는 `-c, --config <key=value>`를 함께 제공하므로 두 형식을 모두 충돌 검사한다.
- **spawn 인용 팽창**: 작은따옴표 40,000개의 헤더 없는 측정 문자열은 인코딩 토큰 40,025 bytes가 shell quote 후 160,027 bytes로 팽창했다. 실제 합성 payload는 provenance 헤더와 명령 prefix 때문에 더 길어지므로 테스트는 고정 총길이가 아니라 실제 생성값과 팽창 관계를 측정한다.

## C. Pre-flight

1. `[ ]` SPEC 디렉터리 유일성 확인 완료 (이번 실행, 부재 확인).
2. `[ ]` baseline 귀속: research.md §Author verification의 file:line 전부 이번 트리에서 직접 판독.
3. `[ ]` `CLAUDE.local.md` 61,360 bytes 측정 완료 — AC-LMD-006 fixture의 기준값.
4. `[ ]` 공식 URL 검증 완료 (VERIFIED — research.md).
5. `[ ]` 구현 경계 고정: 템플릿 일반 문구 정확화(R1), status 행 제외(R2), 양쪽 파일 안전 open/fstat(R3), 독립 fixture의 Codex 레인 LIVE 수행(R4). 이는 카드 전체 구현·로컬 병합 요청에 따른 구현 판단이며 별도 승인 게이트가 아니다.

## D. Constraints

- 구현 코드 편집 금지 경계: 본 plan phase에서 `internal/`, `internal/template/` 편집 없음 ( artifacts만 ).
- `defaultCodexInstructionRelPaths` 3-path 테이블 불변 (`codex_contract.go:225-227` exactly-3 핀).
- `CLAUDE.local.md` 이름은 `codex_contract.go`의 신규 상수로만 참조 (하드코딩 금지 — CLAUDE.local.md §14).
- argv 상한 상수는 `internal/config/defaults.go` 단일 원천 규율.
- run phase에서 템플릿(`AGENTS.md.tmpl`)은 파일명을 재열거하지 않는 일반 문구만 정확화한다. sync 문서(CHANGELOG, docs-site, 리포 AGENTS.md)는 manager-docs에 인계한다.
- `moai codex status` 출력 계약과 행 수는 변경하지 않는다.

## E. Self-Verification

- [ ] E1: `go test ./internal/cli/... -run TestCodexLocal` — 신규·수정 테스트 전부 통과 출력.
- [ ] E2: `GOOS=windows GOARCH=amd64 go build ./...` (안전 open/fstat의 플랫폼별 구현과 fail-closed 대체 경로 포함).
- [ ] E3: scoped test의 coverprofile을 함수별로 판독해 변경 전 패키지 기준선 대비 비회귀를 관측하고, 신규·변경 함수는 coverage 85% 이상이며 충돌·안전 판독·크기 경계 분기가 각 요구 셀에서 실행됨을 증명한다. coverprofile 없는 추정치는 PASS 근거가 아니다.
- [ ] E4: `golangci-lint run ./internal/cli/...` — 0 error.
- [ ] E5: direct 최종 합성 토큰과 spawn 최종 shell-quoted 명령을 각각 측정하고, 헤더 없는 40,000-quote 기준값(40,025 → 160,027 bytes)보다 provenance·prefix를 포함한 실제 launch 표현이 더 길다는 팽창 반례를 기록.
- [ ] E6: provenance-aware parser가 각 body slice를 추출해 원본 fixture와 bytes/hash를 비교하고, launch 후 사용자 로컬 입력의 byte-identity와 rename/link==0을 재단언한다.
- [ ] E7: Codex 레인이 독립 fixture에서 acceptance.md D.12의 exact command를 120초 hard timeout으로 실행하고 양 nonce/source의 response+log predicate와 소스 입력 불변을 기록한다.

## F. Milestones (의사결정 가역성 순 — 변경 가능성 높은 결정이 앞)

### M1 — 합성 의미론 (data-model 축, Priority High)

`codexLocalDeveloperInstructionArgs`를 2-파일 로더로 재작성한다 (REQ-LMD-001/003/004/006):

- 신규 상수 `codexClaudeLocalName = "CLAUDE.local.md"`를 `codex_contract.go`에 추가 (3-path 테이블에는 넣지 않는다).
- **RED 과제 (exactly-3 오류 경로 신설 테스트)**: `len(rels) != 3` 분기(`codex_contract.go:225-227`)는 현재 커버리지 0이다 — 기존 변형 테스트(`codex_contract_test.go:350-370`)는 테이블 항목을 in-place 치환해 항상 길이 3을 유지하므로 이 분기를 지나지 않는다(음성 grep 확인, 이번 실행). 2-또는-4-entry 변형 테이블을 `codexInstructionRelPathsFn`에 주입해 `secureCodexInstructionContract`가 `"instruction path table must name exactly three paths"`로 실패하는 것을 단언하는 RED 테스트를 이 마일스톤에서 추가한다 — REQ-LMD-010의 테이블 불변을 실행 가능하게 지키는 유일한 계기이며, acceptance.md D.5.x의 인용 대상이다.
- per-file 수집 헬퍼: 각 경로를 symlink-follow 없이 안전하게 open하고 같은 descriptor를 fstat한 뒤 그 descriptor에서 read한다. check-then-reopen은 금지하며 양쪽 파일에 같은 helper를 적용한다. 플랫폼별 차이는 최소 분기로 격리하고 동등한 안전성이 없으면 fail closed한다.
- 합성: 존재하는 비어있지 않은 바디들을 `[CLAUDE.local.md, AGENTS.local.md]` 순으로 연결 — REQ-LMD-003이 핀한 출처 구분자(각 바디 직전의 자체 줄 헤더 라인 `<!-- source: <filename> -->`)로 provenance 분리 후 연결, 연결 결과를 한 번만 `json.Marshal` → 단일 `developer_instructions` 오버라이드(`-c` + `developer_instructions=<value>` 2요소 쌍 1개).
- RED 테스트 먼저: precedence-order fixture(AC-LMD-003), absent/empty 매트릭스(AC-LMD-004), 61,360-byte UTF-8 round-trip(AC-LMD-006).

### M2 — 가드·충돌·오버플로 (failure 축, Priority High)

- 경로별 거부·경합 매트릭스 (REQ-LMD-002, AC-LMD-005): 각 파일의 비정규 leaf를 거부하고, 경로 교체 시 교체된 바이트가 주입되지 않음을 단언한다.
- config 충돌 거부: Codex 0.155.1 parser probe가 수용한 다섯 형태(`--config value`, `--config=value`, `-c value`, `-c=value`, `-cvalue`)를 모두 검사한다. `value`가 `developer_instructions=...`일 때만 충돌이며 unrelated key는 통과시킨다.
- overflow 차단: direct는 JSON escape 후 실제 exec 토큰을, spawn은 `buildCodexSpawnCommand`의 최종 shell-quoted 문자열을 별도로 측정한다. 두 경로의 기본 상한은 각각 126,976 bytes(Linux `MAX_ARG_STRLEN` 131,072 − 4,096)이며 `internal/config/defaults.go`의 명명된 단일 원천을 공유한다. 숫자는 같아도 검사는 독립이다. pre-quote 검사 하나가 quote 팽창을 보호한다는 전제는 폐기한다.

### M3 — 헬프·핀 테스트 정합 (user-facing 문구 축, Priority Medium)

- `codexCmd.Long`(:334-335)은 `CLAUDE.local.md`를 Claude workflow와 공유되는 공통 로컬 입력으로, `AGENTS.local.md`를 Codex-specific 입력으로 구분하고 둘 다 주입됨을 기술한다.
- 핀 테스트는 부정확한 exclusivity만 거부한다. `AGENTS.local.md`를 지칭하는 정확한 "Codex-only" 표현 자체는 금지하지 않는다.
- 동일 변경 내에서 문구·테스트가 함께 움직인다는 점이 AC-LMD-011로 핀된다.
- `AGENTS.md.tmpl` §8은 특정 로컬 파일명을 다시 열거하지 않는 일반 문구로 정확화한다. status 행은 추가하지 않는다.

### M4 — 구조 상수·진단 정리 (mechanical 축, Priority Low — defer)

- 경로 문자열 완전 상수화 잔여 정리, 진단 메시지 문구 다듬기, `spec-compact.md` 재생성 등 기계적 마무리.

### Sync-phase 인계 (manager-docs 소관 — run phase에서 실행하지 않음)

1. `CHANGELOG.md` 신규 항목: 계약 역전 명시("no longer Claude-only for Codex injection"), 커밋 `6c647bbe2` 판 문언 정정.
2. docs-site `content/{en,ko,ja,zh}/advanced/codex-dual-harness.md` 4-locale 동시 갱신 (§17 4-locale 동기화 의무; `:33` "remain Claude-only" 문단).
3. 리포 루트 `AGENTS.md` §8(:264) — 템플릿의 일반 문구에 맞춰 갱신.

## G. 위험 테이블

| # | 위험 | 영향 | 완화 |
|---|---|---|---|
| R-a | argv ceiling 초과 (Linux `MAX_ARG_STRLEN` 131,072/토큰; JSON escape로 2×61KiB가 실질 근접) | exec E2BIG — 런치 실패 | M2 size-check fail-closed가 launch 전 차단; E5 실측 |
| R-b | spawn shell-quote 팽창 | pre-quote 토큰은 허용돼도 spawn만 실행 한도를 넘을 수 있음 | spawn 최종 문자열 독립 측정·상한, 40,000 quote 회귀 테스트 |
| R-c | `-w` provenance — `projectRoot`(:468-474)의 로컬 파일이 worktree 세션(`Dir` :478-486)에 주입됨; 워크트리 자체의 untracked 로컬 파일은 미독 | 워크트리 세션이 primary 루트의 지침을 받음 — 기존 AGENTS.local.md 동작의 2배 확대 | 본 SPEC은 기존 의미 유지; help 문서에 명기하고 후속 카드 후보로 기록 |
| R-d | 연산자 tail의 중복 `-c developer_instructions` | 키 중복 precedence 미검증 — 무음 덮어쓰기 위험 | M2 명시적 거부로 차단 (REQ-LMD-005) |
| R-e | check-then-reopen TOCTOU | guard 후 대체 파일 또는 symlink 주입 | 두 파일 모두 safe open + same-descriptor fstat/read |
| R-f | Codex CLI 외부 의미론 (`developer_instructions` vs `AGENTS.md` discovery 조합) 미검증 | 주입 내용이 예상과 다르게 해석될 가능성 | AC-LMD-012 exact command와 양 nonce/source response+log predicate로 실측 |
| R-g | 템플릿 문구와 namespace 중립성 긴장 | 파일명 재유입 또는 부정확한 단정 | 파일명을 재열거하지 않는 정확한 일반 문구 |

## H. MX Tag Plan

- `@MX:ANCHOR` (신규, `[AUTO]`): 2-파일 합성 함수 — 모든 런치 경로의 유일 argv 생산자(fan-in = 6 런치 형태)이자 `developer_instructions`의 sole argv producer. `@MX:REASON`: 단일 토큰·출처 순서(CLAUDE first) 불변식이 여기서만 강제됨. `@MX:SPEC: SPEC-CODEX-LOCALMD-001`.
- `@MX:WARN` (신규, `[AUTO]`): argv-overflow fail-closed 가드 — 리포 최초의 E2BIG 계열 방어이자 net-new 상한 상수. `@MX:REASON`: 상한값·마진 근거가 코드 밖 규약(Linux `MAX_ARG_STRLEN`)에 의존.
- 주석 언어: `code_comments: en` (language.yaml).

## I. Anti-Patterns (금지)

- 2개의 `-c developer_instructions=` 토큰 출력 — 금지 (duplicate-key precedence 미검증).
- `CLAUDE.local.md`를 `codexInstructionRelPathsFn` 테이블에 추가 — 금지 (exactly-3 핀 위반).
- 입력 파일에 대한 어떤 쓰기/개명/링크 — 금지.
- 경로별 분기로 6 런치 경로 각각에 적재 로직 복제 — 금지 (funnel 밖 구현 금지).
- 연산자 `-c` 충돌의 무음 처리(어느 쪽이든) — 금지.

## J. Cross-References

- spec.md (REQ-LMD-001..010) / acceptance.md (AC-LMD-001..012) / preimplementation-recheck.md (실측 반례) / spec-compact.md
- research.md §4 (계약 3세대), §5 (hard couplings), §6 (risks)
- 선후관: SPEC-CODEX-LAUNCHER-001 REQ-CL-013 (read-only), SPEC-CODEX-INIT-001 REQ-CI-008 (Gen1 원 의도), SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001 M4 (템플릿 중립성 — R1과 긴장)
- 소관 외: t1074/t1075 (factory broker, idle wake)
