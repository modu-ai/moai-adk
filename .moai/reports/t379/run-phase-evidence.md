# SPEC-BOARDLOCK-ERRNO-001 run-phase 증거 — 카드 t379

5절 형식(Claim / Evidence / Baseline-attribution / Gaps / Residual-risk).
전체 서술은 `.moai/specs/SPEC-BOARDLOCK-ERRNO-001/progress.md` §E.2 가 단일 보유처다.

트리: `.claude/worktrees/t379`, 브랜치 `WT-boardlock-errno`.
진입 HEAD `9328a5242`, M1 커밋 `364bc332f`, RED-now 트리 `9c196204c76b8f7ff2cba3873c7d21ca7c128017`.
플랫폼 darwin/arm64. **CI 는 ubuntu이며 그 관측은 여기에 없다.**

---

## 1. Claim

1. Unix board-lock 획득의 `flock(2)` 실패가 errno 를 보존한다 — `EWOULDBLOCK`/`EAGAIN` 만 경합
   sentinel 이고, 나머지는 errno 를 `%w` 로 감싸고 lock 경로를 이름 짓는 하드 오류다.
2. **측정된 도달 가능 입력에서 관측 가능한 행동 변화는 0 이다.** 이 카드는 방어적 좁히기 +
   회귀 잠금이지 살아 있는 결함의 수리가 아니다 — 이 호출 지점의 실측 오분류는 **0건**이다.
3. 착지 차단 AC 셋(001b·002·005) PASS. 회귀-가드 셋(001a·003·004)은 초록 유지가 관측됐고
   `verification-completeness.md` §2.1 에 따라 PASS 로 세지 않는다.
4. 계획이 M1 으로 넘긴 부채 둘을 닫았다 — `/dev/fd` 기제 교체(부채-2), M-leak 재규정(부채-3).
5. `EINTR` 의 재분류는 **받아들이는 미측정 행동 변화**이며 (2)의 범위 밖이다.

## 2. Evidence

명령 + 출력 전문 + 종료 코드 + 선택자 매치 수는 `progress.md` §E.2.5 / §E.2.6 표에 있다. 요약:

| 대상 | 명령 | 결과 | rc |
|---|---|---|---|
| 수리 전 기준선 | `go test ./internal/kanban/... -count=1` | `ok ... 17.413s` (실패 집합 공집합) | 0 |
| RED-now 001b | `go test ./internal/kanban/ -run '^TestBoardFlockErrnoNonContentionIsNotHeld$' -count=1 -v` | 4/4 하위 케이스 FAIL @`9c196204c` | 1 |
| RED-now 002 | `go test ./internal/kanban/ -run '^TestBoardFlockErrnoPreservesErrnoAndPath$' -count=1 -v` | 4/4 하위 케이스 FAIL @`9c196204c` | 1 |
| GREEN 전체 | `go test ./internal/kanban/ -run '^TestBoardFlockErrno' -count=1 -v` | 4 test / 8 subtest 전부 PASS | 0 |
| M-broad | 위 선택자 | 001b·002 RED, 001a·003 GREEN | 1 |
| M-narrow | 위 선택자 | 001a·003 RED, 001b·002 GREEN | 1 |
| M-leak | 위 선택자 | 003 RED — `probe fd 6 before, 206 after 200 failed acquisitions` | 1 |
| 되돌림 | `git status --short` | 추적 파일 변경 0 | 0 |
| 수리 후 | `go test ./internal/kanban/... -count=1` | `ok ... 18.923s` (실패 집합 공집합) | 0 |
| 크로스 빌드 | `GOOS=windows GOARCH=amd64 go build ./...` | 출력 없음 | 0 |
| lint | `golangci-lint run --timeout=5m ./internal/kanban/...` | `0 issues.` | 0 |
| coverage | `go test -cover ./internal/kanban/... -count=1` | `coverage: 86.5% of statements` | 0 |
| vet | `go vet ./internal/kanban/` | 출력 없음 | 0 |

## 3. Baseline-attribution

- **행동 불변(AC-BLE-004)** 은 **편집을 시작하기 전** HEAD `9328a5242` 에서 실측한 기준선
  (`ok ... 17.413s`, rc=0)에 귀속된다. 사후에 만들어 붙인 기준선이 아니다.
- **lint 신규 0** 은 같은 HEAD 의 `0 issues.` baseline 에 귀속된다.
- **RED-now 3셀**은 트리 `9c196204c76b8f7ff2cba3873c7d21ca7c128017` 에 귀속된다. 이 트리는
  `git write-tree` 가 만든 실재 객체이며 `.moai/reports/t379/red-now-tree.diff` 로 재구성된다.
- **GREEN·게이트 관측 전부**는 트리 `364bc332f` 에 귀속된다.
- **오분류 0건**은 계획 단계 프로브(`errno-probe-output.txt`, darwin/arm64)에 귀속되며, **이 실행의
  측정이 아니라 계획 단계 측정의 인용이다.**

## 4. Gaps — 관측하지 않은 것

- **리눅스 관측 없음.** 모든 측정이 darwin/arm64 다. CI 러너는 ubuntu이며 이 보고서는 그 결과를
  담고 있지 않다.
- **Windows 테스트 미실행.** 크로스 빌드 rc=0 만 관측했다. windows substrate 의 런타임 동작은
  이 실행에서 관측되지 않았다.
- **`EINTR`/`ENOLCK`/`EOPNOTSUPP` 커널 유도 안 함.** 합성 errno 로 분류 술어만 쟀다.
  실동작은 미측정이며 **0 이 아니다.**
- **`-race` 미실행.** 이 카드는 동시성을 도입하지 않으므로 돌리지 않았다. 돌리지 않았다는 사실이
  race 가 없다는 증거는 아니다.
- **패키지 밖 미측정.** `./internal/kanban/...` 한정. 전 패키지 판정은 CI 몫이다.
- **소비자 3곳의 테스트 피복을 직접 세지 않았다.** 호출 지점 3개는 읽었지만, 그것을 지나는 기존
  테스트가 몇 개인지는 세지 않았다 — AC-BLE-004 의 대조가 실제로 무엇을 배제했는지는 그 수에
  달려 있고, 그 수를 모른다.
- **`l44` pre-commit fetch 미수행.** 이 워크트리는 카드 전용이고 푸시하지 않았다. `origin` 대조는
  통합 시점에 리드가 낸다.
- **커버리지 86.5% 는 패키지 전체 수치**이며, 이번에 추가한 분류 함수의 개별 피복률은 따로 재지
  않았다.
- **감사 재실행 없음.** plan-audit iter3(PASS-WITH-DEBT 0.83) 이후 재감사는 하지 않았다.

## 5. Residual-risk

- **`EINTR` 이 실제로 도착하는 배포에서 행동이 바뀐다.** 재시도 예산 흡수 → 즉시 하드 오류.
  도달 가능성 논증은 syscall 계약을 읽은 **추론이지 측정이 아니다**. 이 변화는 어떤 검사로도
  잡히지 않는다 — 받아들이기로 한 것이기 때문이다.
- **AC-BLE-003 의 fd 단언을 재는 뮤턴트는 M-leak 하나다.** 그 하나가 실제로 RED 를 냈으므로
  가드는 작동하지만, 단일 뮤턴트에 걸려 있다.
- **AC-BLE-003 은 분류에 결합돼 있다** — M-narrow 에도 반응한다(경합 sentinel 전제가 깨지므로).
  의도한 전제 가드이지만, 순수한 fd-위생 가드는 아니다.
- **AC-BLE-004 는 뮤턴트 가드가 없다.** 전후가 모두 `ok` 인 대조가 무엇을 배제했는지는 스윕 범위에
  달려 있고, 그 범위를 정량화하지 않았다(위 Gaps).
- **RED-now 트리 객체 `9c196204c` 는 어떤 ref 에서도 도달 불가라 gc 대상이다.** 사라지면
  `git ls-tree` 검증이 불가능해진다. 복구 수단은 `red-now-tree.diff` 이며, 그것이 진짜 증거다.
- **`slack = 16` 은 이 머신에서 관측한 여유(before 6 → after 6)에 견주면 넉넉하지만, Go 런타임이
  더 많은 디스크립터를 여는 환경에서는 위양성이 날 수 있다.** ubuntu CI 에서의 여유는 미측정이다.
- **darwin 측정을 리눅스 판정으로 쓰지 않았다는 것이 리눅스에서 통과한다는 뜻은 아니다.**
