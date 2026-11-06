# t2128 — lane record (lane-28, run to2bbk)

카드: PR #1969 Codex P2 수리(config) — AF 설정 읽기의 secret 검사가 자기 참조 YAML alias(`webhook.routes` 안 `nested: *cycle`)에서 끝없이 재귀해 stack overflow로 죽음. Class B, SPEC 없음. 트리 `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2128`, 브랜치 `WT-pr-codex-repair-6`, 기준 HEAD `04d8c27a7e`. 착지는 t2121+rc.3 뒤 카드 PR 경로(리더 배차문).

## 1. 출발점 (2026-11-06)

- 배차: 리더 `uds:/tmp/cc-socks/82930.sock`, card t2128, evidence `.moai/reports/t2128/verdict.md`(audited_sha=HEAD) + `card-review.md`, Codex 단계당 최대 2회(규칙 `d-20261106T080039Z-305c`).
- 지적 원문 위치: `internal/config/agentic_factory_settings.go:434` @04d8c27a7e (PR #1969 리뷰 코멘트).
- 레인 읽기(이 트리, HEAD 04d8c27a7e): `afsHasSecretKey`(파일 429행 근처)는 `yaml.AliasNode`를 만나면 `n.Alias`로 재귀하고, mapping·sequence 자식으로도 재귀하며 방문 기록이 없음. alias가 자기를 포함하는 앵커를 가리키면 종료 조건이 없음 — 지적과 일치(가설, 재현 전).
- 재현은 manager-develop이 서브프로세스 시험으로 먼저 확인(카드 본문 요구).

## 2. 진행 기록

(이하 단계별로 추가)

### 2.1 재현 (RED, HEAD 04d8c27a7e, 수정 전 코드)

- 진입점: `LoadAgenticFactorySettings(projectRoot)` → `readAgenticFactorySettings` → `r.guardSecrets()` → `afsHasSecretKey`(agentic_factory_settings.go:429-447).
- 시험: `internal/config/agentic_factory_settings_cycle_test.go` `TestAgenticFactorySettings_RefCycle`(부모)가 자기 바이너리를 `-test.run=^TestAgenticFactorySettings_RefCycleChild$`로 재실행(환경변수 `MOAI_AFS_REF_CHILD_ROOT`)하고 60초 제한 컨텍스트로 묶음. 최소 YAML: `webhook.routes` 첫 항목이 `&cycle` 앵커 매핑이고 그 안에 `nested: *cycle`.
- 명령: `GOCACHE=<scratch>/gocache-lane28 go test -count=1 -run TestAgenticFactorySettings_RefCycle ./internal/config/` → exit=1. 자식 stderr 머리말(원문):

```
--- FAIL: TestAgenticFactorySettings_RefCycle (3.86s)
    agentic_factory_settings_cycle_test.go:78: loading a self-referencing alias crashed the process (exit status 2):
        === RUN   TestAgenticFactorySettings_RefCycleChild
        runtime: goroutine stack exceeds 1000000000-byte limit
        fatal error: stack overflow
```

- 자식 단독 실행(같은 YAML, 테스트 바이너리 직접 실행) 스택 프레임(원문 발췌): `afsHasSecretKey ... agentic_factory_settings.go:429` → `:434`(AliasNode→n.Alias) → `:439`(매핑 값 재귀) → `:434` → `:439` … 반복, exit=2.

### 2.2 다른 보행 함수 점검 (이 입력 기준)

- `afsLookup`: dotted 경로 세그먼트 수만큼만 순회(유한), alias는 세그먼트당 1회만 따라감 — 순환 입력에서 무한 루프 없음.
- `afsChild`: 비재귀.
- `routes()`: 항목이 `MappingNode`가 아니면 거부, 매핑이면 `repository` 한 키만 조회 — 순환과 무관. 단 `guardSecrets`가 먼저 실행되므로 크래시는 그 단계에서 발생.
- `yaml.Node.Decode`: 스칼라(`enabled`, 정수 값)에만 호출 — 순환 노드를 디코드하지 않음.
- 결론: 이 입력으로 닿는 재귀는 `afsHasSecretKey` 하나뿐. 수정 대상은 그것만.

## 3. 원인

`afsHasSecretKey`(internal/config/agentic_factory_settings.go, 수정 전 429-447행)가 `yaml.AliasNode`를 `n.Alias`로 따라가며(434행) 매핑 값(439행)·시퀀스 자식(444행)으로 재귀하는데 방문 기록·종료 조건이 없었음. `&cycle` 앵커 매핑이 `nested: *cycle`로 자신을 가리키면 434↔439가 끝없이 반복되어 Go 고루틴 스택 한도(1GB) 초과 → `fatal error: stack overflow`(recover 불가). yaml.v3 파서는 앵커를 매핑 파싱 시작 시점에 등록하므로 자기 참조 입력을 그대로 허용함.

## 4. 구현 기록

커밋(RED가 수정보다 앞):
- `d5e205208c` [t2128] test(config): RED — 서브프로세스 시험이 수정 전 코드에서 stack overflow로 실패.
- `15d6a8648a` [t2128] fix(config): `afsHasSecretKey`를 `(found, cycle bool)` 반환 + `onPath`(현재 하강 경로, 순환 판정)·`clean`(다 훑은 노드, 공유 앵커 1회 보행) 집합으로 교체. `guardSecrets`는 cycle이면 `webhook.routes[i]`에 "a YAML alias refers back to itself (a cycle), which this file does not allow (the value was not echoed)"를 `r.fail`로 보고(값 미노출). 사용 안 하게 된 `slices` import 제거.
- 시험 추가: `TestAgenticFactorySettings_RefCycle`(서브프로세스, 오류 문구·키·값 미노출), `..._SharedAnchorNotRejected`(순환 아닌 공유 앵커 허용), `..._SecretThroughAnchorStillRefused`, `TestAfsHasSecretKey_CycleAndSharing`(7 케이스 직접 단위 시험: 맵/리스트/간접 순환, 공유, 공유 안 secret 등).

검증(수정 후, 한 패키지씩):
- `go test -count=1 ./internal/config/` → `ok  github.com/modu-ai/moai-adk/internal/config	31.781s` (exit 0)
- `go test -count=1 -v -run TestAgenticFactorySettings_RefCycle ./internal/config/` → `--- SKIP: ..._RefCycleChild`, `--- PASS: TestAgenticFactorySettings_RefCycle (0.01s)`, `ok ... 0.171s`
- `GOOS=windows GOARCH=amd64 go vet ./internal/config/` → exit 0, 출력 없음
- `gofmt -l` 변경 파일 → 출력 없음

Gaps: 전체 스위트·다른 패키지·CI 미실행(리더가 CI에 맡김); lint(golangci-lint)·`-race` 미실행; 자식 프로세스 방식은 Windows에서 실행 검증 못 함(vet만); 서로 다른 앵커 간 매우 깊은 중첩(깊이 폭주)은 순환이 아니므로 별도 깊이 제한 없음; 독립 감사(verdict.md/card-review.md)는 이 단계 범위 밖.

## 5. iter-1 repair

sync-audit iter-1 F1 (BLOCKING): `internal/config/agentic_factory_settings_cycle_test.go` lines 51 and 54 called `os.Stdout.WriteString(...)` without checking the result (errcheck). Repair commit `fa2a30ca6e`: both calls now `_, _ = os.Stdout.WriteString(...)`; child-process protocol (`RESULT:...` lines) unchanged. F2-F4 (optional) left as is.

Verification (GOCACHE=.../scratchpad/gocache-lane28, one command at a time):
- `golangci-lint run ./internal/config/` -> exit 1; tail: `7 issues:` / `* govet: 7`. All 7 are `inline: Constant reflect.Ptr should be inlined (govet)` in `harness_ceiling_test.go:61` and `shipped_key_reader_test.go:369,588,625,706,777` (and one more cut from the tail). `grep -c errcheck` = 0, `grep -c cycle_test` = 0 on the output. Those files are not in this card's diff (`git diff --name-only 04d8c27a7e HEAD` lists only progress.md, sync-audit-iter1.md, agentic_factory_settings.go, agentic_factory_settings_cycle_test.go), so the govet findings are pre-existing and out of scope.
- `go test -count=1 ./internal/config/` -> `ok  	github.com/modu-ai/moai-adk/internal/config	17.980s`

Gaps: lint not clean package-wide (7 pre-existing govet findings, not touched); no full suite; no CI.
