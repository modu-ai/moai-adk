# t297 verdict — launch-ledger 쓰기 정규화 (REQ-009)

카드 t297 (Class B — 원인 판명된 결함, plan 생략 직행) · lane-15 · Factory Mode
브랜치 `WT-launch-ledger-write` · base `304bc8158` (로컬 develop 팁 흡수 확인됨)

## Claim (주장)

1. launch 원장(`launch.yaml`의 `projects:`)은 워크트리마다 행이 쌓이는 단조 증가 결함이 실재했고
   (실행 재현으로 관측), 쓰기 시점 정규화로 새 중복 행 생성이 원천 차단됐다 (REQ-009/AC-009).
2. 워크트리 N개 생성→폐기 사이클에서 원장 행 수가 유계로 남는다 — 실행 기반 회귀로 확인.
3. 의도적으로 만든 중복(레거시 행)이 수거 경로로 접혀 최종 행 수 == 서로 다른 프로젝트 수가 된다 — 실행 확인.
4. t293의 조상-walk 읽기 경로와 정합: 쓰기가 접는 키 == walk가 해석하는 키 (구조적 쌍둔Ѱ walk), 정규화 후에도
   fresh sibling 워크트리가 기록값을 해석한다 (read/write coherence 테스트).
5. 폐기 경로(`moai worktree remove/done/clean`)가 죽은 행의 회수 주체다 — 배선 6지점 + 테스트.

## Evidence (증거 — 실행한 명령 + 관측 출력)

모든 로그는 이 디렉터리(`.moai/reports/t297/`)에 커밋됨:

| 파일 | 내용 |
|---|---|
| `red-prefix-run.log` | 수정 전 RED 전문 — 5회 워크트리 기록 → 6행, fold/정합성 2건 FAIL |
| `teeth-mutant1-fold-disabled.log` | 변이 1(조상 fold 제거) → fold·정합성 2건 RED |
| `teeth-mutant2-prune-nop.log` | 변이 2(prune 삭제 생략) → prune·회귀 4건 RED |
| `teeth-mutant3-precedence-swap.log` | 변이 3(exact↔조상 우선순위 교환) → 중첩 프로젝트·레거시 2건 RED |
| `teeth-mutant4-wiring-nop.log` | 변이 4(CLI 배선 무효화) → 배선 3건 RED |
| `green-profile-final.log` | 수정 후 profile 신규 14건 전부 PASS |
| `green-worktree-final.log` | 수정 후 worktree 배선 13건 전부 PASS |

명령·출력 요약:

- `go test ./internal/profile -count=1 -cover` → `ok ... coverage: 86.0% of statements` (기준 85% 충족)
- `go test ./internal/cli/worktree -count=1` → `ok ... 3.964s`; `go test ./internal/web -count=1` → `ok ... 4.218s`
- `go vet ./internal/profile/ ./internal/cli/worktree/ ./internal/web/` → 통과
- `GOOS=windows GOARCH=amd64 go build ./internal/profile/ ./internal/cli/worktree/ && GOOS=windows ... go vet` → 통과
- `golangci-lint run internal/profile/... internal/cli/worktree/...` → `0 issues.` (초차 2건 errcheck는 `_, _ =` 관례로 수리)

## Baseline-attribution (baseline 귀속)

- 측정 트리: worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t297`, HEAD `304bc8158` (구현 전)
  → 구현 커밋 시점에 재측정(커밋 직전 상태에서 위 명령 전부 재실행). 병합 트리 재측정은 통합 창 소관.
- 실기계 전제 측정: `~/.moai/claude-profiles/launch.yaml` — 8행 중 워크트리 모양 3행
  (`t267`, `t289`, `release-v313`, 모두 생존 중). 읽기 전용 python3 yaml 파싱, 원장 미수정.
- RED/GREEN/변이는 같은 패키지 같은 커밋 기준선 위의 대조 관측이다 (수정 전 RED → 수정 후 GREEN).

## 설계 결정 (카드 4축)

1. **쓰기 정규화** — `resolveWriteProjectKey`(`internal/profile/profile.go`): 우선순위
   (a) exact/alias 기존 행 → 그 자리 갱신(중첩 독립 프로젝트 보호 — `/mono` vs `/mono/lib`),
   (b) 등록된 가장 깊은 생존 조상 → 그 행으로 접음(REQ-009 핵심),
   (c) 둘 다 없으면 해당 경로에 live 행(cold-start — 폐기 시 수거되는 유계 잔여).
   쓰기 경로는 삭제를 하지 않는다 — 삭제는 수거 주체의 몫(경계 명확화).
2. **회수 주체·시점** — **워크트리 폐기 시점**으로 판정. 세션 종료 기각(세션≠워크트리, 프로젝트 간
   원장에 매 세션 쓰기는 노이즈), 독립 청소 커맨드 신설 기각(재사용 사다리: 기존 표면 존재).
   배선: `runRemove`·`runDoneWorktreeCleanup(--auto, quiet)`·`runDone`·`runClean`(plain)·
   `cleanStaleWorktrees`(apply만)·`cleanMergedWorktrees`. preview/`--json`은 아무것도
   제거하지 않는다(REQ-WR-013 정합). Claude 네이티브 isolation 워크트리 제거 훅은 미배선 —
   해당 트리 클래스는 런처/웹 쓰기가 없어 원장 행을 만들지 않음(카드 전제의 행은 모두
   런처/웹 기원), 잔여는 `moai worktree clean`이 주기 수거.
3. **1회성 정리 경로** — `moai worktree clean`(plain)에 `PruneStaleProjectEntries` 배선.
   술어: 키 디렉터리 부재(모든 읽기 경로가 os.Stat을 요구하므로 죽은 행은 정의상 회수 안전),
   멱등(2회차 0건 + 파일 바이트 동일 — 테스트 고정), 관측 가능(제거 건수 stdout 1행).
   프로필 디렉터리가 죽은 행은 보존(바인딩은 살아있는 프로젝트의 자산 — 테스트 고정).
4. **t293 조상-walk 정합** — `registeredAncestorKey`는 `lookupSubtreeProjectKey`와 동일한
   Dir-step 구조 walk(쓰기가 접는 키 == 읽기가 해석하는 키). 의도적 비대칭 1개: 쓰기 쪽은
   `launchCandidateIsUsable`을 보지 않는다 — 쓰기는 방금 검증한 live 프로필로 값을 덮어쓰므로
   죽은 프로필 조상에 접는 것은 행의 수리다(문서화됨). walk의 stale-skip은 드문 fallback으로 유지.

## Gaps (미검증)

- 실기계 실원장(3개 워크트리 행)에 대한 1회성 prune 실행은 수행하지 않았다 — 트리 생존 중(8/8 alive)이라
  지금은 죽은 행이 0이고, 폐기 후 `moai worktree clean` 1회로 회수된다. 멱등성·술어는 샌드박스 테스트로 검증.
- go1.26.4/darwin/arm64 단일 환경 실행. windows는 vet+build만(테스트 런 아님) — 저장소 관례와 동일.
- `moai cc -p X` 엔드투엔드(claude 바이너리 exec 포함) 실런치는 재현하지 않았다 — 패키지 경계의
  실함수(`RecordLastUsedProfileForProject`) 실행 기반 검증으로 갈음.

## Residual-risk (잔여 위험)

- 레거시 live 중복(워크트리 생존 중인 자기 행)은 폐기 전까지 남는다 — 쓰기는 그 자리를 갱신해
  성장은 없고, 폐기+prune으로 소멸(회귀 (b) 테스트가 이 수렴을 고정). 폐기되지 않는 채로 두면
  행 수는 살아있는 트리 수에 유계(무한 증가 아님).
- `moai worktree` CLI를 거치지 않는 폐기(세션 종료 프롬프트 등)의 죽은 행은 다음
  `moai worktree clean`(또는 아무 폐기)까지 남는다 — 이벤트 수거가 아니라 폐기/청소 시점 수거라는
  선택의 귀결. 유계는 다음 clean까지의 지연일 뿐이다.
- cold-start(루트 미등록 프로젝트) 워크트리는 폐기 전까지 자기 행을 유지한다 — 동일하게 유계.

## 결정 요청

없음 — 카드 범위 안에서 판단 가능한 결정만 내렸다. 리드 판독 대상: 본 판정서 + 커밋 3건.
