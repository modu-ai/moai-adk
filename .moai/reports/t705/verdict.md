# Card t705 Verdict — 임시 루트 안 git 저장소의 거짓 안내 수리 (Class B, 재현 우선)

- Date: 2026-09-14 · Lane: lane-4 · Branch: `WT-tempgit-refusal` @ 8dcfa2470 (base 4da5d1c4e, unpushed)
- Class B — plan 생략, 재현→판정→RED→GREEN

## Claim

`t.TempDir()` 아래 git 저장소에서 `StateDirForRoot`는 큐를 프로젝트 로컬에 유지하면서 `TempOriginRefusal`은 `refused=false`(홈 큐 사용 중으로 안내)를 보고하는 **계층 간 모순**이 실재하는 결함임을 나란히 측정으로 확정하고, 안내 계층의 git-first 조기 반환 제거로 수리했다. 독립 검토 PASS(변이 제어 포함), lint 0 issues.

## 결함 판정 (결함 vs 의도)

**결함** — 근거 3가지:
1. 해석 계층(`StateDirForRoot`, state_dir.go:29)은 git을 전혀 보지 않는 경로 기반 판별 — temp-git도 프로젝트 로컬로 강제(temp 가드의 존재 이유: git-init하는 /tmp 픽스처의 홈 오염 방지)
2. t706의 homestate 가드(paths.go:99)도 경로 기반 — git-first를 쓰는 계층은 `TempOriginRefusal` 하나뿐(소외)
3. `TempOriginRefusal` 자신의 doc comment("exactly as the resolvers treat it")는 검증 없이 참이라 여겨진 주장이었고 측정으로 반증됨

## Evidence

| 단계 | 결과 | 증거 |
|---|---|---|
| 재현(RED) | 나란히 측정: `StateDirForRoot` = `<tmp>/project/.moai/state/todo` vs `TempOriginRefusal` = `refused=false` — 패리티 붕괴 관측 | 최초 측정 실행 출력(t.Logf 4행 + PARITY BREAK 에러) |
| 수리 | git-first 조기 반환 제거(todo_root.go), doc comment 정정, `explicitMoaiHome` 면제 유지 | 커밋 `97cd3a562` |
| GREEN | 신규 회귀 테스트 PASS + 기존 가드 스위트 19/19 PASS + 패키지 전체 ok 181.5s + vet·gofmt·build 0 | `go test -v ... 'TempOrigin|TodoQueueRoot'` exit=0, 19 PASS/0 FAIL |
| 독립 검토 | **PASS** — 2(a) 비임시 git 케이스 불변 코드 대조 입증, 2(b) 유일 생산 호출자(todo.go:95)는 refused=false에 의존하는 분기 없음+stderr 안내 1줄뿐, 2(c) 주석 진실성 확인, **revert-check을 go test -overlay로 기계 실행 — 구현으로 되돌리면 정확히 결함 서명으로 FAIL** | 검토 보고(세션 로그) |
| F1 수리 | 회귀 테스트의 폐기 예정 `filepath.HasPrefix` → 동일 패키지 `pathWithin` — lint 1건 → **0 issues** | 커밋 `8dcfa2470`, `golangci-lint run ./internal/kanban/` 0 issues |

## Baseline-attribution

모든 수치는 lane-4 세션에서 이 워크트리(4da5d1c4e 기반)에서 실행한 명령의 관측값. 패키지 전체 판정은 CI 몫(§4 규율).

## Gaps

- 실제 `moai todo`를 temp-git 디렉터리에서 실행한 엔드투엔드 미수행(단위+호출부 코드 판독으로 대체 — PersistentPreRun 배선은 기존 코드)
- 전체 스위트·internal/cli 패키지 미재측정(독립 검토 범위 밖 — CI 몫)
- `--separate-git-dir` 체크아웃 형태 미측정

## Residual-risk

- 런타임 변화 표면: temp-git에서 `moai todo` 실행 시 stderr 안내 1줄이 새로 나옴 — exit 코드·큐 위치·스크립트 계약 불변(검토 확인)
- t706의 앵커 이원화(kanban TempRootsFn vs homestate tempRoots)는 유지 — 단일원천화는 t706 verdict의 후속 소관
