# SPEC-VERIFY-RUN-REUSE-001 — 구현 계획

## §A 접근 (한 문단)

`internal/verify` 에 순수 결정 함수를 두고(키·시계·저장소 주입), `internal/cli` 에 얇은 cobra 동사를 얹는다. 비교는 기존 `CheckReceipt` 를 그대로 쓰고 새 저장소·스키마는 없다. 한 번 실행하면 `RecordCheck` 로 기록한다.

## §B 파일 목록 (이 목록 밖은 건드리지 않는다)

| 파일 | 변경 | 용도 |
|---|---|---|
| `internal/verify/run.go` (신규) | 추가 | 명령 정규화(`CanonicalCommand`), 환경 다이제스트(`EnvDigest`), 재사용 결정(`DecideReuse`) |
| `internal/verify/run_test.go` (신규) | 추가 | 결정 함수 단위 테스트 |
| `internal/cli/verify_run.go` (신규) | 추가 | `moai verify run` 동사. `verify_receipts.go` 와 같이 `init()` 에서 `verifyExtraCommands` 에 등록 |
| `internal/cli/verify_run_test.go` (신규) | 추가 | CLI 통합 테스트 (`initVerifyTestRepo` 재사용) |
| `internal/cli/verify.go` | 한 줄 | `Long` 의 Verbs 목록에 `verify run` 추가 |
| `AGENTS.md` | 한 문장 | §4 규율 문장 |
| `internal/template/templates/AGENTS.md.tmpl` | 한 문장 | §4 미러 (동일 문장) |

금지 파일(t1479 소유)은 spec.md §C. 재사용을 거기에 배선해야 하는 변경은 spec.md §F 에 따라 t1479 착지 후로 미룬다.

## §C 단계 (결정이 바뀔 가능성이 큰 것부터)

### M1 — 사용자 노출 표면과 의미 (우선순위 High)

1. 동사 시그니처·플래그·종료 코드 확정: `moai verify run [--check-id <id>] [--ttl <d>] [--env NAME,...] [--tool-version-cmd <elem>]... [--tool-version-timeout <d>] [--timeout <d>] -- <command...>`. `--check-id` 기본값 `run`, 재사용 판정에는 쓰지 않고 `verify check --check` 필터용. `--tool-version-cmd` 는 cobra `StringArray`(반복 플래그, 값 쪼개기 없음)로 구현한다 — 토큰화 규칙이 없으므로 공백 경로(Windows)도 안전하다.
2. 적중/미스 통지는 stderr 한 줄 (`verify run: reuse key=... recorded_at=... duration_ms=...` / `verify run: miss (<reason>)`). stdout 은 실행된 명령의 것만 — 파이프 오염 방지. 적중 통지의 `recorded_at` 은 인용 문구(REQ-VRR-008)가 그대로 옮기는 값이다.
3. 기본값은 spec.md REQ-VRR-003 이 정한다: `--timeout` 60m(카드 실측 30분 스위트를 덮으면서 무한 대기는 아님), `--tool-version-timeout` 30s(`verifyKeyTimeout` 과 동일), `--ttl` 10m(`verify.DefaultTTL`).
4. `--tool-version-cmd` 가 없으면 `unversioned`. 호출자 책임을 규율 문장과 `--help` 에 명시. 실패·타임아웃·빈 출력은 unbound(재사용·기록 없음, 명령은 실행).
5. 타임아웃 종료: Unix 는 자식을 새 프로세스 그룹으로 시작해 그룹 전체에 종료 신호, Windows 는 직접 자식만 종료(손자 잔존은 spec.md §D-5 잔여 위험). 플랫폼별 코드는 빌드 태그 파일로 분리하고(`manager-develop-prompt-template.md` B1) `GOOS=windows GOARCH=amd64 go build ./...` 로 확인한다.

### M2 — 결정 함수 (internal/verify/run.go)

- `CanonicalCommand(argv []string) string` — REQ-VRR-005.
- `EnvDigest(names []string, lookup func(string) (string, bool)) string` — `ConfigDigest` 에 `env:NAME` 키로 위임, unset 은 값에 `\x00unset` 같은 비충돌 표지를 써서 빈 값과 구분 (REQ-VRR-006).
- `DecideReuse(snap *Snapshot, state ReceiptState, now time.Time, ttl time.Duration) (hit bool, entry *CheckEntry, reason string)` — `snap.FindCommand` 로 항목을 얻어 `Receipt` 로 변환 후 `CheckReceipt`, 추가로 exit 0·verdict `pass` 확인 (REQ-VRR-002). `LoadReceipt` 는 `DurationMS` 를 버리므로 쓰지 않고 `Load` + `FindCommand` 로 직접 구성한다.

### M3 — CLI 동사 (internal/cli/verify_run.go)

흐름: 루트 해석(`verifyResolveRoot`) → 저장 루트(`verifyStoreRoot`) → `verify.Key` 전 → 도구 식별 → 환경 다이제스트 → `verify.Load` → `DecideReuse` → 적중이면 통지·exit 0 → 미스면 `exec.CommandContext`(stdio 상속, `Dir=root`, 타임아웃) → `verify.Key` 후 → 같으면 `RecordCheck`. 종료 코드는 `exitCodeError`(`ExitCode()`) 로 전달 — `internal/cli/constitution.go` 의 기존 타입과 같은 패턴 (REQ-VRR-003). 키·저장소 오류는 REQ-VRR-007 에 따라 fail-open, 전후 키 비교는 REQ-VRR-004. 서브에이전트 경계: AskUserQuestion 미사용.

### M4 — 규율 문장 (우선순위 Medium)

`AGENTS.md` §4 와 템플릿 미러에 같은 문장 한 개 추가 (REQ-VRR-008). 제안문(한 단락, 핵심 구절 `moai verify run`·`--env`·`recorded_at`·`output not re-observed`·`runs the command directly` 가 각각 한 줄 안에 들어가도록 줄바꿈): "**Run repeated verification through `moai verify run`**: it executes a command once per working-tree state and reuses a passing result within its TTL, so list the environment variables (`--env NAME,...`) and the toolchain identity (`--tool-version-cmd`) the result depends on. A reuse is a prior observation, not one made in this run: a verdict citing it names the key and `recorded_at` from the reuse notice and lists \"output not re-observed\" under Gaps, and a claim that needs verbatim output runs the command directly." 템플릿 미러에 같은 단락을 넣는다. 이 SPEC 은 생성 단계를 주장하지 않으며, 바이트 한도는 `go test ./internal/config -run '^TestCodexContractByteCeiling$' -count=1 -v` 로 확인한다 (iter1 관측 여유 2996/3284 B). 로컬 `.claude/rules/**` 변경은 없다.

### M5 — 기계적 정리 (우선순위 Low)

`verify.go` Long 한 줄, `@MX:NOTE` 한 줄(`run.go` 의 `DecideReuse` 에 `@MX:ANCHOR` + `@MX:REASON`: 재사용 판정의 단일 지점), `gofmt`/`golangci-lint`(CI 판 v2.1.6).

## §D 위험과 완화

| 위험 | 완화 |
|---|---|
| 키가 환경·툴체인을 덮지 않아 오적중 | REQ-VRR-006/007 + spec.md §D 잔여 위험 명시, 규율 문장에 경계 명시 |
| 실행 중 트리가 움직여 결과가 다른 트리에 귀속 | REQ-VRR-004 전후 키 비교, 불일치 시 기록 안 함 |
| 동시 레인이 같은 키에 기록 | `RecordCheck` 키별 잠금(기존). 최악은 같은 트리의 동등 결과 덮어쓰기 |
| 키 계산이 untracked 대용량 파일에서 느림 | 기존 `Key()` 특성. 30초 타임아웃(`verifyKeyTimeout`) 초과 시 fail-open 으로 직접 실행 (REQ-VRR-007) |
| 템플릿 미러가 어긋남 | AC-VRR-007 이 두 파일에서 같은 문장 존재를 단언 |
| 인수 해석(셸 문법) 혼동 | REQ-VRR-005 직접 exec, `sh -c` 사용 예를 `--help` 에 표기 (AC 는 셸을 쓰지 않는 자체 실행 헬퍼 사용) |
| 타임아웃 후 손자 프로세스 잔존 | Unix 프로세스 그룹 종료, Windows 는 잔여 위험(spec.md §D-5) |
| AC 선택자가 이름 변경·삭제에도 초록 | acceptance.md 의 `-v` + N 줄 `--- PASS` + `[no tests to run]` 부재 단언 |
| 재사용 결과를 이번 실행의 관측으로 오인 | REQ-VRR-008 문장(이전 관측, Gaps 에 "output not re-observed") + AC-VRR-007 |

## §E 자체 검증 (run 단계에서 실행)

- acceptance.md 의 AC 별 `-v` 명령(N 줄 `--- PASS`, `[no tests to run]` 부재). CLI 13개 이름 존재 확인: `go test ./internal/cli -list '^(TestVerifyRunHitExecutesZeroTimes|TestVerifyRunMissOnTreeChange|TestVerifyRunMissOnCommandBytes|TestVerifyRunNeverReusesFailure|TestVerifyRunMissOnTTL|TestVerifyRunMissOnEnvChange|TestVerifyRunToolVersionBinding|TestVerifyRunToolVersionTimeout|TestVerifyRunExitCodePassthrough|TestVerifyRunTimeoutAndNotFound|TestVerifyRunTreeMovedNotRecorded|TestVerifyRunStoreFailOpen|TestVerifyRunIgnoresHandRecordedEntry)$'` — 이름 13줄 출력이어야 한다. 패키지 전체는 CI
- `go test ./internal/config -run '^TestCodexContractByteCeiling$' -count=1 -v`, `GOOS=windows GOARCH=amd64 go build ./...`
- `go vet ./internal/verify ./internal/cli`
- `go run ./cmd/moai spec lint .moai/specs/SPEC-VERIFY-RUN-REUSE-001/spec.md`
- 변경 파일이 §B 표 안인지 `git diff --name-only 2b9e4a4d0 HEAD` 로 확인

## §H 교차 참조

spec.md · acceptance.md · `.moai/reports/t1452/verdict.md` · `.claude/rules/moai/workflow/snapshot-consumer-contract.md` (소비자 계약은 변경하지 않음)
