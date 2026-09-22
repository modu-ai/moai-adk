# plan.md — SPEC-CODEX-LOCALMD-001

## A. Context

- 카드 t1078 (Tier M — frontmatter `tier: "M"`; 규모 근거: 수정 표면 약 10개(런처·컨트랙트·테스트·헬프·plan-artifacts), 예상 LOC ~450-700 → Tier M 밴드 300-1000 내), 워크트리 `WT-codex-local-md@0314801c2`, 개발 모드: TDD (`codex_local_instructions_test.go`가 characterization base).
- 목표: `internal/cli/codex_launcher.go`의 `codexLocalDeveloperInstructionArgs`(:114-142)를 2-파일 로더로 확장해 `CLAUDE.local.md`도 `-c developer_instructions=` 단일 토큰으로 주입한다. 구조적 균일성은 기존 funnel(`codex_launcher.go:495`) 안에 머무는 것으로 무료로 상속된다(REQ-LMD-008).
- 이 변경은 기록된 계약("CLAUDE-local 상태는 Claude 전용" — 커밋 `6c647bbe2`, `CHANGELOG.md:77`, `AGENTS.md` §8/:264, `AGENTS.md.tmpl` §8/문단 :268-273, docs-site `codex-dual-harness.md:33`, 헬프 카피 "Codex-only")을 되돌린다. Gen1(SPEC-CODEX-INIT-001 REQ-CI-008)의 원 의도를 `-c` 채널로 회복하는 성격이다(research.md §4, C1).

## B. Known Issues (선결 결함·긴장)

- **계약 역전 (C1)**: 다중 표면이 "Claude-only"를 기록 중 — 커밋 `6c647bbe2`, `CHANGELOG.md:77`, 리포 `AGENTS.md` §8(:264), 템플릿 `AGENTS.md.tmpl` §8(문단 :268-273, "…are not forwarded to Codex." 문장 :273 — 본 트리 직접 판독; audit 리포트의 :272 표기와 1행 차이, 본 트리 측정 우선), docs-site `codex-dual-harness.md:33`, 헬프 카피 "Codex-only". coordinated-surface 계획은 §F M4/sync 인계 참조.
- **헬프 카피 핀 (C3 해소)**: `codex_launcher.go:334-335`는 status readout이 아니라 `codexCmd.Long` 헬프 카피가 맞다(직접 판독). `TestCodexLocalInstructions_DocumentedInLauncherHelp`(:141-147)가 `"Codex-only"`를 grep — 문구와 테스트가 같은 변경으로 움직여야 한다(REQ-LMD-009).
- **argv overflow 방어 부재**: `E2BIG|ErrArgListTooLong` grep 전 `*.go` 0건 — 리포 최초 방어(REQ-LMD-007).
- **TOCTOU**: Lstat→ReadFile 사이 심볼릭 링크 치환 시 guard 탈출 가능(기존 AGENTS.local.md에도 존재). 본 SPEC은 창을 넓히지 않는 것을 요구하고, 봉쇄 여부는 decision-index R3.
- **`moai codex status` readout에 local-instruction 행 부재** — 거부 조건이 readout에서 보이지 않음(decision-index R2).

## C. Pre-flight

1. `[ ]` SPEC 디렉터리 유일성 확인 완료 (이번 실행, 부재 확인).
2. `[ ]` baseline 귀속: research.md §Author verification의 file:line 전부 이번 트리에서 직접 판독.
3. `[ ]` `CLAUDE.local.md` 61,360 bytes 측정 완료 — AC-LMD-006 fixture의 기준값.
4. `[ ]` 공식 URL 검증 완료 (VERIFIED — research.md).
5. `[ ]` decision-index R1~R4가 Implementation Kickoff Approval 전에 운영자에게 상신되어야 한다.

## D. Constraints

- 구현 코드 편집 금지 경계: 본 plan phase에서 `internal/`, `internal/template/` 편집 없음 ( artifacts만 ).
- `defaultCodexInstructionRelPaths` 3-path 테이블 불변 (`codex_contract.go:225-227` exactly-3 핀).
- `CLAUDE.local.md` 이름은 `codex_contract.go`의 신규 상수로만 참조 (하드코딩 금지 — CLAUDE.local.md §14).
- argv 상한 상수는 `internal/config/defaults.go` 단일 원천 규율.
- run phase에서 템플릿(`AGENTS.md.tmpl`)·sync 문서(CHANGELOG, docs-site, 리포 AGENTS.md) 편집 금지 — 각각 decision-index R1 상신 후 / manager-docs 인계.

## E. Self-Verification

- [ ] E1: `go test ./internal/cli/... -run TestCodexLocal` — 신규·수정 테스트 전부 통과 출력.
- [ ] E2: `GOOS=windows GOARCH=amd64 go build ./...` (internal/cli 크로스 플랫폼; O_NOFOLLOW 채택 시 빌드 태그 분기 검증 포함 — R3 채택 시).
- [ ] E3: `go test -cover ./internal/cli/...` — 패키지 커버리지 85% 이상 관측.
- [ ] E4: `golangci-lint run ./internal/cli/...` — 0 error.
- [ ] E5: argv 크기 실측: 합성 토큰 길이(2×~61KiB escape 포함)와 상한 비교 리포트, spawn 경로 shell-quoted 총길이 측정치 기록.
- [ ] E6: 입력 비가공 재확인 — 테스트가 launch 후 두 파일 byte-identity 재단언 (기존 `codex_contract_link_test.go:173` rename==0 패턴 준용).
- [ ] E7: LIVE AC(AC-LMD-012) 증거 경로가 `.moai/specs/SPEC-CODEX-LOCALMD-001/` 아래에 기록됨.

## F. Milestones (의사결정 가역성 순 — 변경 가능성 높은 결정이 앞)

### M1 — 합성 의미론 (data-model 축, Priority High)

`codexLocalDeveloperInstructionArgs`를 2-파일 로더로 재작성한다 (REQ-LMD-001/003/004/006):

- 신규 상수 `codexClaudeLocalName = "CLAUDE.local.md"`를 `codex_contract.go`에 추가 (3-path 테이블에는 넣지 않는다).
- **RED 과제 (exactly-3 오류 경로 신설 테스트)**: `len(rels) != 3` 분기(`codex_contract.go:225-227`)는 현재 커버리지 0이다 — 기존 변형 테스트(`codex_contract_test.go:350-370`)는 테이블 항목을 in-place 치환해 항상 길이 3을 유지하므로 이 분기를 지나지 않는다(음성 grep 확인, 이번 실행). 2-또는-4-entry 변형 테이블을 `codexInstructionRelPathsFn`에 주입해 `secureCodexInstructionContract`가 `"instruction path table must name exactly three paths"`로 실패하는 것을 단언하는 RED 테스트를 이 마일스톤에서 추가한다 — REQ-LMD-010의 테이블 불변을 실행 가능하게 지키는 유일한 계기이며, acceptance.md D.5.x의 인용 대상이다.
- per-file 수집 헬퍼: Lstat→IsRegular→ReadFile→빈 바디 nil. 파일별 `codexPathGuardError`/`codexModeName` 어휘 재사용.
- 합성: 존재하는 비어있지 않은 바디들을 `[CLAUDE.local.md, AGENTS.local.md]` 순으로 연결 — REQ-LMD-003이 핀한 출처 구분자(각 바디 직전의 자체 줄 헤더 라인 `<!-- source: <filename> -->`)로 provenance 분리 후 연결, 연결 결과를 한 번만 `json.Marshal` → 단일 `developer_instructions` 오버라이드(`-c` + `developer_instructions=<value>` 2요소 쌍 1개).
- RED 테스트 먼저: precedence-order fixture(AC-LMD-003), absent/empty 매트릭스(AC-LMD-004), 61,360-byte UTF-8 round-trip(AC-LMD-006).

### M2 — 가드·충돌·오버플로 (failure 축, Priority High)

- 경로별 거부 매트릭스 (REQ-LMD-002, AC-LMD-005): symlink/FIFO/directory/read-error 각 파일 독립 거부 + `launches == 0`.
- 연산자 `-c developer_instructions=` 충돌 명시적 거부 (REQ-LMD-005, AC-LMD-007): `runCodexLaunch`에서 tail 스캔, 충돌 시 named error로 fail — silent overwrite 금지. **본 SPEC이 채택한 규칙: reject 명시적 거부** (merge 규칙 대비 단순·fail-closed 전통 정합; 운영자 재확인은 decision-index R4와 함께 상신).
- argv-overflow fail-closed (REQ-LMD-007, AC-LMD-008): 합성 오버라이드의 **최종 인코딩 길이**(JSON escape 완료 후 exec에 전달되는 `developer_instructions=<value>` 토큰의 바이트 길이)를 상수 상한과 비교, 초과 시 launch 전 진단과 함께 실패. 상한 상수는 `internal/config/defaults.go` 단일 정의; **마진 정책(핀)**: 기본값 = Linux `MAX_ARG_STRLEN` 131,072 − 마진 4,096 bytes = **126,976 bytes** (마진은 `-c` 접두, spawn 경로의 shell-quote 확장, 커널 계산 변동을 위한 여유). spawn 경로(`buildCodexSpawnCommand` :193-204)는 직전 단계인 이 검사로 보호된다. `@MX:WARN` 대상.

### M3 — 헬프·핀 테스트 정합 (user-facing 문구 축, Priority Medium)

- `codexCmd.Long`(:334-335) 문구 갱신: 두 파일 모두 기술, "Codex-only" 문구를 정확한 서술로 교체 (REQ-LMD-009).
- `TestCodexLocalInstructions_DocumentedInLauncherHelp` 기대 토큰 갱신 (`"Codex-only"` → 새 문구 토큰).
- 동일 변경 내에서 문구·테스트가 함께 움직인다는 점이 AC-LMD-011로 핀된다.

### M4 — 구조 상수·진단 정리 (mechanical 축, Priority Low — defer)

- 경로 문자열 완전 상수화 잔여 정리, 진단 메시지 문구 다듬기, `spec-compact.md` 재생성 등 기계적 마무리. 템플릿 문구는 여기서도 만지지 않는다(R1 상신 대기).

### Sync-phase 인계 (manager-docs 소관 — run phase에서 실행하지 않음)

1. `CHANGELOG.md` 신규 항목: 계약 역전 명시("no longer Claude-only for Codex injection"), 커밋 `6c647bbe2` 판 문언 정정.
2. docs-site `content/{en,ko,ja,zh}/advanced/codex-dual-harness.md` 4-locale 동시 갱신 (§17 4-locale 동기화 의무; `:33` "remain Claude-only" 문단).
3. 리포 루트 `AGENTS.md` §8(:264) — 템플릿 렌더 재생성 경로이므로 R1 결정 이후 템플릿 문구에 맞춰 갱신.
4. [선택, R2 채택 시] `moai codex status` readout에 local-instruction 상태 행.

## G. 위험 테이블

| # | 위험 | 영향 | 완화 |
|---|---|---|---|
| R-a | argv ceiling 초과 (Linux `MAX_ARG_STRLEN` 131,072/토큰; JSON escape로 2×61KiB가 실질 근접) | exec E2BIG — 런치 실패 | M2 size-check fail-closed가 launch 전 차단; E5 실측 |
| R-b | spawn shell-quote ceiling — argv 전체가 tmux 명령 문자열 1개로 합쳐짐 (`buildCodexSpawnCommand` :193-204) | spawn 경로만의 제2 상한, tmux/`/bin/sh -c` 측 미측정 | 직전 단계의 M2 검사가 동일 상한으로 보호; 본 호스트 실측치를 E5에 기록 (미측정 상태면 Gaps 명시) |
| R-c | `-w` provenance — `projectRoot`(:468-474)의 로컬 파일이 worktree 세션(`Dir` :478-486)에 주입됨; 워크트리 자체의 untracked 로컬 파일은 미독 | 워크트리 세션이 primary 루트의 지침을 받음 — 기존 AGENTS.local.md 동작의 2배 확대 | 본 SPEC은 기존 의미 유지; help 문서에 명기하고 후속 카드 후보로 기록 |
| R-d | 연산자 tail의 중복 `-c developer_instructions` | 키 중복 precedence 미검증 — 무음 덮어쓰기 위험 | M2 명시적 거부로 차단 (REQ-LMD-005) |
| R-e | Lstat→ReadFile TOCTOU — `ReadFile`이 심볼릭 링크를 따라감 | guard 탈출 후 대체 파일 주입 가능성 (기존 결함, 파일 수만 2배) | 창 확대 금지는 요구; 봉쇄(open-with-`O_NOFOLLOW`+fstat)는 R3 결정 |
| R-f | Codex CLI 외부 의미론 (`developer_instructions` vs `AGENTS.md` discovery 조합) 미검증 | 주입 내용이 예상과 다르게 해석될 가능성 | AC-LMD-012 LIVE 수용이 실측 수단; `AGENTS.md` 런타임 비가공이므로 rollback 용이 |
| R-g | 템플릿 §8 문구 vs SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001 M4 긴장 | 템플릿에 CLAUDE.local.md 참조 재유입 | R1 decision row로 운영자 결정 전까지 템플릿 무변경 |

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

- spec.md (REQ-LMD-001..010) / acceptance.md (AC-LMD-001..012) / decision-index.md (R1~R4) / spec-compact.md
- research.md §4 (계약 3세대), §5 (hard couplings), §6 (risks)
- 선후관: SPEC-CODEX-LAUNCHER-001 REQ-CL-013 (read-only), SPEC-CODEX-INIT-001 REQ-CI-008 (Gen1 원 의도), SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001 M4 (템플릿 중립성 — R1과 긴장)
- 소관 외: t1074/t1075 (factory broker, idle wake)
