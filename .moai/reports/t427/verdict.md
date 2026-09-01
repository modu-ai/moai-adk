# t427 verdict — nearestProjectRoot 홈 위장 수리

- 카드: t427 · Tier S · 워크트리 `.claude/worktrees/t427` · 브랜치 `WT-root-ghosthome`
- 기원: t426 판정서 §Residual #1 (재현 경로: `.moai/reports/t426/verdict.md`)
- 범위: `internal/cli` 1패키지 — `doctor_agentemit_embed.go` 수리 + `doctor_agentemit_embed_test.go` 테스트 6건 추가

## Claim

`nearestProjectRoot`의 조상 탐색에서 사용자 홈 디렉터리를 `os.SameFile` OS 동일성 판정으로 후보에서 제외하면, 전역 `~/.moai`를 보유한 홈(모든 moai 사용자) 아래 비프로젝트 경로에서의 doctor가 더 이상 "in project" 가닥으로 읽히지 않는다. 제외는 철자 불문이다 — windows 8.3 단축명(TMP=`RUNNER~1` vs USERPROFILE=`runneradmin`)·대소문자 분기·심링크 경로 모두 동일 디렉터리로 흡수된다. 건드린 패키지(`internal/cli`) 전체 테스트·vet·lint·3-OS 빌드가 이 트리에서 통과한다.

## 원인 (root cause)

`internal/cli/doctor_agentemit_embed.go`의 `nearestProjectRoot`가 조상 탐색에서 `.moai/` 디렉터리 존재만 보고 홈을 후보에서 제외하지 않았다. 모든 moai 사용자의 홈은 전역 상태 디렉터리 `~/.moai/`를 갖고, 홈 아래 비프로젝트 경로에서 doctor를 돌리면 탐색이 홈에서 멈춰 not-applicable 메시지가 "no committed emission set at internal/..."(프로젝트 안에 있다는 어조) 가닥으로 나갔다. 실관측 2점: (1) t426 CI 러너에서 러너 홈의 `~/.moai`가 Agent Emit Embed golden 행의 분기를 바꿈(축 3), (2) 본 카드의 로컬 RED 재현 — 아래 Evidence.

t426이 이 수리를 넣지 않은 이유는 windows 8.3 단축명 때문이었다: 홈을 **문자열 비교**로 제외하면 같은 디렉터리가 `C:\Users\RUNNER~1\...`(단축명 cwd)와 `C:\Users\runneradmin`(USERPROFILE)으로 다른 철자로 도달해 비교가 놓친다. 본 수리는 문자열 비교를 버리고 `os.SameFile`(unix dev+inode / windows 볼륨+파일 ID)로 판정해 철자 축 자체를 제거했다.

## Evidence (명령 + 출력, 이 런 — darwin 로컬)

RED (수리 전, 시임 변수만 놓고 실행):

```
go test ./internal/cli/ -run 'TestNearestProjectRoot|TestAgentEmitEmbed_HomeOnly' -count=1 -v
--- FAIL: TestNearestProjectRoot_HomeWithGlobalMoaiIsNotARoot
    nearestProjectRoot(".../001/work/scratch") = ".../001", want not found
--- FAIL: TestNearestProjectRoot_HomeItselfIsNotARoot
--- PASS: TestNearestProjectRoot_RealProjectUnderHomeStillFound   (기존 올바른 동작 고정)
--- FAIL: TestNearestProjectRoot_SymlinkedSpellingStillExcluded
--- FAIL: TestNearestProjectRoot_CaseFlippedSpellingStillExcluded (실매치 실패 — 공허 통과 아님)
--- FAIL: TestAgentEmitEmbed_HomeOnlyReadsAsNotAProject
    message = "not applicable: no committed emission set at internal/template/templates/.codex/agents/moai/"
FAIL    github.com/modu-ai/moai-adk/internal/cli
```

GREEN (수리 후, 같은 판정식):

```
go test ./internal/cli/ -run 'TestNearestProjectRoot|TestAgentEmitEmbed|TestFindEmbedCheckRoot|TestBoundedTail|TestExtractEmissionViaInit' -count=1 -v
--- PASS (신규 6건 전부 RED→GREEN 전환)
--- PASS (기존 embed-check 11건 회귀 없음)
ok    github.com/modu-ai/moai-adk/internal/cli
```

풀 패키지(golden 포함):

```
go test ./internal/cli/ -count=1   → ok github.com/modu-ai/moai-adk/internal/cli 345.913s, exit=0
```

정적·크로스 플랫폼:

```
go vet ./internal/cli/                        → rc=0 (darwin)
GOOS=windows go vet ./internal/cli/           → rc=0
GOOS=linux   go vet ./internal/cli/           → rc=0
golangci-lint run internal/cli/...            → "0 issues."
GOOS=windows|linux go build ./internal/cli/   → rc=0, 로컬 darwin build rc=0
```

수리 본체 (`internal/cli/doctor_agentemit_embed.go`): `homeDirFn` 시임 변수(테스트 주입) + `nearestProjectRoot`가 `os.Stat(home)` FileInfo를 한 번 잡아 조상마다 `os.SameFile` 비교 — 동일이면 후보에서 제외, 그 외 기존 `.moai/` 판정 그대로. 홈 미해석·stat 실패 시 제외 없이 기존 동작(fail-open toward old behavior). `findEmbedCheckRoot`와 `walkUp`은 건드리지 않았다(committed-set 조상은 홈에 존재할 수 없음 — 범위 절제).

## Gaps (로컬에서 관측되지 않은 것)

- windows·linux **런타임** 실측은 CI 몫이다. 로컬 GOOS vet/build는 컴파일 계층만 증명한다(규칙: cross-build는 테스트 미컴파일). case-flip 서브테스트는 windows 러너에서 8.3 형제 축을 실제로 발화시키는지 CI windows 레그에서 확인된다.
- `go test ./...` 전체 스위트는 로컬에서 실행하지 않았다(규율) — develop push의 CI가 통합 판정.

## Residual-risk / 후속

1. t426 §Residual #2(windows 조상 탐색 부재 — Toolhelp32 스냅샷)는 본 카드와 무관하게 잔존. 별도 카드 규모.
2. t426 상호작용: t426의 `normalizeAgentEmitRow`는 두 메시지 가닥을 한 캐노니컬 행으로 붕괴시키므로 본 수리로 메시지 가닥이 바뀌어도 golden 판정은 무영향. t426(창 순번 선순위) 착지 후 본 브랜치 병합 시 develop 흡수로 정합 확인.
3. 케이스 민감 파일시스템(예: Linux case-sensitive, 케이스 민감 APFS 볼륨)에서는 대소문자 분기 시나리오 자체가 존재하지 않아 서브테스트가 skip된다 — 결함이 아니라 시나리오 부재.

## 다음

리드가 develop 병합 창에서 `git merge --no-ff WT-root-ghosthome` 후 push — CI windows 레그 판독은 리드 몫. 워크트리 폐기는 원격 머지 확인 후. 단, t324 창 순번(3번: t412 → t426 → t324)이 도래하면 본 카드는 중단되고 t324 병합이 먼저 실행된다(리드 지시 조건).
