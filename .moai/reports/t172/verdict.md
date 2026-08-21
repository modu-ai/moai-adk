# t172 — `moai gate` typecheck axis

베이스: `main` @ `4b2f203fe` · 브랜치 `WT-gate-typecheck` · 워크트리 `.claude/worktrees/t172`
출처: mo.ai.kr 사고(10ab1872) 근본 수리 · 설계는 운영자 확정본(재조사 불요)

---

## 1. 카드 실측 좌표 검증 (착수 전 확인, 4/4 일치)

카드가 "재조사 불요"라고 했으나 좌표만은 짚었다. **전부 맞았다.**

| 좌표 | 확인 | 결과 |
|---|---|---|
| `gate.go:63-72` langToolchain 3축, typecheck 부재 | 62-73행 직접 읽음 | `vetSteps`/`lintSteps`/`testStep` 뿐 — **typecheck 없음** 확인 |
| `:699-717` 무음 스킵 3경로 | 695-720행 직접 읽음 | optional 바이너리 부재 / configFiles 부재 / changedExts 불일치 — 전부 `return true, ""` |
| `:289-304` passReason 인프라 | 285-305행 직접 읽음 | 존재. 다만 **덮어쓰기**였다(§3.2) |
| `:622` readPackageJSONScripts | 618-622행 | 존재, 재사용 가능 |
| 언어 엔트리 15개 | `toolchains` 블록 내 `markerFiles:` 정확 카운트 | **15** 일치 |

> 주의 하나: 파일 전체 grep 은 18을 세는데, 그중 3건은 `var toolchains` 블록 밖이다. 블록 내부만
> 세면 카드의 15가 맞다. 파일 전체 grep 으로 확인했다면 카드가 틀렸다고 오판했을 자리다.

## 2. 구현 (9파일)

| 파일 | 내용 |
|---|---|
| `internal/hook/quality/gate_typecheck.go` (신규) | 3티어 해석기 + solution-style 판별 + `appendReason` |
| `internal/hook/quality/gate_typecheck_test.go` (신규) | 14 테스트 |
| `internal/hook/quality/gate.go` | `typecheckStep` 필드 · Node 엔트리 배선 · GateConfig 3필드 · 기본값 · Run 순서 |
| `internal/config/types.go` | `GateTypecheck` 구조체 · `Timeouts.Typecheck` · `TypecheckTimeoutDuration()` |
| `internal/config/defaults.go` | `enabled: true` · `typecheck: 300` |
| `internal/cli/gate.go` | config→quality 매핑 3필드 |
| `.moai/config/sections/gate.yaml` + 템플릿 미러 | 신규 키 + 주석 |
| `internal/config/testdata/shipped_key_inventory.yaml` | 신규 키 3건 triage 등록(§4) |

실행 순서: **vet → typecheck → lint → ast-grep → test**. 타입 오류가 느린 스타일 패스보다 먼저
드러나고, 린터 설정이 없는 프로젝트도 정합성 게이트는 받는다 — 사고의 정확한 형태가 후자였다.

### 2.1 Node 3티어

(a) `gate.typecheck.command` — **어느 언어든** 강제 발화 (b) `package.json` `scripts.typecheck` →
`npm run typecheck` (c) `tsconfig.json` → `npx --no-install tsc --noEmit`.

**티어 (b)가 tsconfig 형태 검사보다 우선**한다. turbo 에 위임하는 모노레포 루트가 solution-style
tsconfig 를 가졌다는 이유로 벌받지 않게 하려는 것이고, 테스트로 고정했다
(`TestSolutionStyleWithScriptStillRuns`).

### 2.2 solution-style 차단

`files: []` + `references` 만인 tsconfig 는 `tsc --noEmit` 이 **아무것도 검사하지 않고 exit 0** 한다.
이걸 커버리지로 치면 이 카드가 닫으려는 사각을 그대로 다시 만든다. 그래서 실행하지 않고 사유와 함께
스킵한다. 파싱 실패는 `false` — 읽을 수 없는 설정 때문에 검사를 거부하는 쪽이 더 나쁜 실패다.

## 3. 스킵 의미론 — "스킵은 실패가 아니나 결코 무음이 아님"

### 3.1 사유 부착

축이 없는 언어는 `typecheck: skipped (no default for this language; set gate.typecheck.command to
enable one)` 를 출력한다. 도구 가용성은 **fail-open 유지**하되 가시성은 항상 준다.

### 3.2 발견 — passReason 은 누적이 아니라 덮어쓰기였다

카드 범위를 넘지 않는 선에서 한 가지를 고쳐야 했다. 기존 `passReason = out` 은 **마지막 통지가 앞의
것을 지웠다.** typecheck 와 ast-grep 이 둘 다 스킵하면 하나만 보고된다 — 이 카드가 만드는 통지가
바로 그 자리에서 사라진다. `appendReason` 으로 누적하도록 바꿨고, vet/lint 스텝의 통지도 같이
줍도록 했다(이전엔 버려졌다).

실증(§5 데모 3): 한 실행에서 **두 통지가 함께** 출력된다.

## 4. shipped-key 앤티로트 가드 (예상 밖 1건)

`TestShippedConfigKeysHaveReaders` 가 신규 키 3건을 **미triage** 로 잡았다. 처음 출력이 잘려
기존 실패(cacheStrategy·context_search·hook.*)만 보였고 **내 키가 안 보였다** — 필터를 다시 걸어
확인하니 내 것이 맨 위에 있었다. 잘린 출력만 보고 "기존 baseline" 으로 넘겼으면 놓쳤을 자리다.

`shipped_key_inventory.yaml` 에 3건을 등록했다. class **W**(wire) / evidence `reader` — 실제로
`mapConfigGateToQuality` 가 읽는다. 헤더 카운트도 955→958 로 맞췄다.

(기존 `gate.timeouts.*` 3건이 class R 로 등록돼 있는데 실제로는 읽힌다 — 스테일 triage 지만 내
변경이 아니라 손대지 않았다.)

## 5. 검증

### 5.1 TDD RED 선행 (원문)

```
$ go test ./internal/hook/quality/          # gate_typecheck.go 존재 전
internal/hook/quality/gate_typecheck_test.go:48:22: undefined: resolveTypecheckStep
internal/hook/quality/gate_typecheck_test.go:48:43: undefined: nodeTypecheckStep
FAIL	github.com/modu-ai/moai-adk/internal/hook/quality [build failed]
```

### 5.2 대조절 포함 테스트 14건

차단(RED)만 있으면 "항상 실패하는 게이트"로도 만족된다. 그래서 **통과 대조절**을 같이 뒀다:
`TestGateBlocksOnTypecheckFailure`(command `false` → 차단) ↔ `TestGatePassesWhenTypecheckSucceeds`
(command `true` → 통과). 그 밖: 3티어 각각 · solution-style 차단 · 스크립트 우선 · 오버라이드
타언어 발화 · 사유 부착 · `disabled_steps` · `enabled:false` · 기본값 · Node 엔트리 배선.

### 5.3 실 바이너리 데모 (`make build` 산출물)

`CLAUDE_PROJECT_DIR` 로 픽스처에 스코프. **cwd 도 픽스처 안이어야 한다** — §6 참조.

| 픽스처 | 결과 |
|---|---|
| `broken`(typecheck exit 2) | `quality gate failed: typecheck` + **TS 오류 원문이 그대로 출력**. 게이트 차단 |
| `fixed`(typecheck exit 0) | typecheck 통과 후 다음 스텝으로 진행 — 축이 통과함을 증명 |
| `noconfig`(스크립트·tsconfig 둘 다 없음) | **게이트 통과(EXIT=0)** + `typecheck: skipped (…set gate.typecheck.command…)` 출력. ast-grep 통지와 **함께** 나옴(§3.2 실증) |

### 5.4 명령과 관측

```
$ go test ./internal/hook/quality/ -count=1        -> ok 28.031s
$ go test ./internal/config/... -count=1           -> ok (config / atomicfile / toolpolicy)
$ go test ./internal/config/ -run TestStructYAMLSymmetry_Gate -> ok   (CONFIG_STRUCT_YAML_MISMATCH 통과)
$ go test ./internal/config/ -run TestShippedConfigKeysHaveReaders -> ok
$ golangci-lint run --timeout=3m <3 packages>      -> 0 issues.
$ go build ./...                                   -> exit 0
$ make build                                       -> 성공, catalog.yaml 변동 없음
$ gofmt -l <변경 파일>                              -> 출력 없음
```

### 5.5 `internal/cli` 실패 3건 — 환경 오염, 내 변경 아님

첫 실행에서 `TestCC_FactoryEntryThroughRunCC` / `TestGLM_FactoryWorkerEntry` /
`TestHookCommandFlushesLastHandlerEntry` 3건이 붉었다. **이 세션이 팩토리 세션이라
`MOAI_FACTORY_WORKERS` 등이 `go test` 까지 물린다**(기록된 실패 양식). env 스크럽 후 재실행:

```
$ unset MOAI_KANBAN … MOAI_FACTORY_WORKERS MOAI_FACTORY_WORKER && go test ./internal/cli/ -run '…'
ok  	github.com/modu-ai/moai-adk/internal/cli	1.156s
ok  	github.com/modu-ai/moai-adk/internal/cli	9.721s
```

3건 전부 초록. **내 변경과 무관한 로컬 환경 거짓 실패**로 판정한다.

## 6. 알려진 한계 — cmd.Dir (카드 범위 밖, 그러나 실측으로 드러남)

데모 1차에서 `fixed` 픽스처가 실패했다. 원인: 스텝 **해석**은 픽스처를 보는데 **실행**은 프로세스
cwd 에서 일어나 `npm` 이 워크트리 루트의 없는 `package.json` 을 읽었다. 카드가 범위 밖으로 명시한
`runStep` `cmd.Dir` 미지정과 같은 뿌리다.

**소비자 영향은 없다** — 실제 사용은 저장소 루트에서 `moai gate` 를 부르므로 cwd == 프로젝트 디렉터리다
(데모 2·3이 그 구성이고 전부 정상). 다만 **이 축이 그 결함을 상속한다**는 사실은 기록해 둔다:
프로젝트 디렉터리와 cwd 가 갈리는 호출에서는 typecheck 도 엉뚱한 디렉터리에서 실행된다.

## 7. 미검증 / 잔여 위험

- **전체 스위트 로컬 미실행** — 부하 규율. 전 패키지 판정은 CI(PR) 몫이다.
- **실제 `tsc` 를 돌려 본 적 없다.** 티어 (c)는 `npx --no-install tsc --noEmit` 커맨드를 **구성**하는
  것까지만 테스트했고, tsc 가 실제로 그 인자로 기대대로 동작하는지는 관측하지 않았다. 데모는 전부
  스크립트 티어(b)로 돌렸다.
- **solution-style 판별은 `files: []` 명시만 결론으로 삼는다.** `files` 키 자체가 없고 `references`
  만 있는 변형은 실무상 같은 공허 통과를 낳을 수 있으나 보수적으로 통과시킨다.
- 소비자(mo.ai.kr) 자동 발화는 **추론**이다 — `scripts.typecheck` 보유 → 티어 (b) 발화로 예상되나,
  ADK 릴리스 + `moai update` 이후 그쪽에서 실측해야 확정된다.
- `pkill -x moai` 로 데모 잔여 프로세스를 정리하다 **`moai mcp-server` 도 함께 죽였다**(MCP 도구가
  세션에서 끊김). 복구는 세션 재시작으로 되지만, 광범위 매칭 kill 을 쓴 내 실수다. PID 지정이 옳았다.

## 8. 전달

브랜치 `WT-gate-typecheck` → push → **PR 직접**(release 배치와 무관, 사고 근본 수리이므로).
브랜치명은 규약상 `WT-` 접두(카드 스케치의 `feat/*` 대신); 의도는 PR 제목이 전달한다.

금지된 untracked 4건(`ci-autofix-protocol.md`·`astgrep-rules/`·`t127/`·`diagram-design-absorption/`)은
스테이징에 **포함되지 않았다** — `git status --short` 로 확인.
