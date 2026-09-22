# t601 — sync-phase 품질 게이트: 레코드 키에 작업‑트리 내용 식별자 추가

카드: t601 · `[hooks 감사 2026-09-11 · H06 · P1]` 동일 커밋의 두 번째 품질 검사에서 앞선 차단 결과가 사라짐
대상: `.claude/hooks/moai/sync-phase-quality-gate.sh` (+ `internal/template/templates/` 미러)
워크트리: `.claude/worktrees/t601` · 브랜치 `WT-quality-gate-sentinel` · 기준 `9935e4e3e`

적용 규칙: `verification-claim-integrity.md` §1.1(1)(3), §2, §3 · `verification-completeness.md` §1.1, §2, §4

---

## Claim

1. 카드가 기록한 증상(같은 HEAD 두 번째 호출의 stdout이 빈값)은 **develop 9935e4e3e에서 이미 재현되지 않는다.** 감사 기준이던 `main 2213871af` 사본과 develop 사본은 서로 다른 파일이다.
2. 카드 완료 조건 세 전이 중 **`수정 → 재호출`만 미달**이었다. 작업 트리를 고쳐도 검사가 재실행되지 않고 낡은 `decision:block`이 그대로 재전달됐다.
3. 원인은 레코드 키가 HEAD SHA 하나뿐인데 검사 대상은 작업 트리라는 불일치다. 파일 헤더가 그 키 설계를 명시하고 있었다.
4. 수리 후 그 전이가 성립하고, **대칭 구멍(저장된 `pass` 아래에서 트리가 깨진 경우)도 함께 닫힌다.**
5. 기존 회귀 가드는 동작 축에서 하나도 깨지지 않는다. 바뀐 것은 레코드 형식 단정 2곳뿐이며, 그 형식 변경이 이 카드의 내용이다.
6. `internal/hook` 루트 패키지에 레드 1건이 남아 있으나 **이 변경에 귀속되지 않는다**(선재 레드).

## Evidence

### E1 — 기준점

```
$ git rev-parse --show-toplevel
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t601
$ git rev-parse --short HEAD
9935e4e3e
$ git show-ref | grep WT-quality-gate-sentinel
9935e4e3e1067cc5d5d4576f8ae4fd5410d838d6 refs/heads/WT-quality-gate-sentinel
```

`origin/HEAD` → `origin/develop`, 그리고 `refs/heads/develop` = `refs/remotes/origin/develop` = `9935e4e3e` (동일 SHA)이므로 워크트리 기본 기준점이 배차된 기준과 일치한다.

### E2 — 감사 기준 사본과 develop 사본은 다른 파일

```
$ wc -l .claude/hooks/moai/sync-phase-quality-gate.sh      # develop
     618
$ wc -l (main 2213871af 작업 사본, 앞선 판독)
     399
```

develop 사본은 이미 `outcome record`(pass/fail/running) + 실패 payload 재전달 + stale 재실행을 갖고 있다. 따라서 카드의 원 관측은 develop에서 **철회**되어야 하고, 남은 것은 권고문 마지막 절(내용 식별자)뿐이다.

### E3 — 5칸 프로브, 수리 전후 대조

계측기: `.moai/reports/t601/probe_t601.py` (셸 훅을 임시 git 픽스처에서 직접 구동)
수리 전 대상: `git show HEAD:.claude/hooks/moai/sync-phase-quality-gate.sh`을 파일로 받아 겨눴다. 그 사본은 커밋하지 않았다 — 훅 파일의 네 번째 사본이 저장소에 남으면 "게이트 사본이 몇 벌인가"를 세는 다음 감사를 부풀린다. 재유도는 위 명령 한 줄이다.

```
$ git show HEAD:.claude/hooks/moai/sync-phase-quality-gate.sh > <tmp>/base-gate.sh
$ python3 .moai/reports/t601/probe_t601.py <tmp>/base-gate.sh                        > probe-baseline.json   # exit 0
$ python3 .moai/reports/t601/probe_t601.py .claude/hooks/moai/sync-phase-quality-gate.sh      > probe-after.json      # exit 0
$ python3 .moai/reports/t601/compare.py

cell                                           pre-fix   post-fix  want(post)
P0 control: 2nd call re-ran checks              False     False     False  OK
P1 fail->recall: 2nd call re-ran checks         False     False     False  OK
P1 fail->recall: stdout byte-identical          True      True      True   OK
P2a interrupted fresh: re-ran checks            False     False     False  OK
P2b interrupted stale: re-ran checks            True      True      True   OK
P3 repaired tree: re-ran checks                 False     True      True   OK
P3 repaired tree: stale block re-delivered      True      False     False  OK
P3 repaired tree: HEAD unchanged                True      True      True   OK

P3 post-fix stdout: ''
P3 post-fix record: b1fbb1459059ead3639aa7cea7bb22fa2040f9d0 pass 858cd981…d6a8
P1 post-fix record: 747980169f7bf0cba056d3674e5dfa80a404079c fail e3b0c442…b855
P1 pre-fix  record: 3b3ccd7243e2ed08048e707915d7e0eb5b1a2578 fail
```

여덟 칸 중 **바뀐 것은 P3 두 줄뿐**이고 나머지 여섯 칸은 수리 전후 동일하다 — 대조는 값이 아니라 불변을 단언한다.

> 계측기 정정 기록: 최초 프로브는 가짜 `go`의 호출 기록 파일을 **저장소 안**에 두었다. 수리된 훅은 트리 내용으로 키를 잡으므로, 계측기가 측정 대상을 바꿔 P0·P1이 매 호출 재실행으로 관측됐다. 기록 파일을 저장소 밖으로 옮긴 뒤 양쪽을 같은 계측기로 다시 쟀고, 위 표가 그 재측정이다. 정정 전 수치는 근거로 쓰지 않는다.

### E4 — Go 수용 테스트: RED → GREEN

신규 행 4개 (`internal/hook/sync_gate_failstate_test.go`, `TestSyncGateFailState_T601_WorkTreeContentKeysTheRecord`):

| 행 | 등급 | 요구 |
|---|---|---|
| R1 | release-blocking | 트리 수리 후 재호출 → 검사 재실행, decision 없음 |
| R2 | regression-guard | 트리 불변 재호출 → 재실행 0회, stdout 바이트 동일 |
| R3 | release-blocking | 저장된 pass 아래 트리 파손 → 재실행 + block |
| R4 | release-blocking | 다줄 레코드 → 재게이트 |

수리 전 (RED, 관측):

```
--- FAIL: …/R1-repaired-tree-regates
    stub invoked 0 time(s) after the work tree was repaired; want >= 1
    stdout still carries a decision after the repair
--- PASS: …/R2-untouched-tree-redelivers
--- FAIL: …/R3-broken-after-pass-regates
    stub invoked 0 time(s) after the work tree broke under a stored pass; want >= 1
    stdout is not a block after the work tree broke; stdout=""
```

R2가 수리 전에 통과한다는 것이 중요하다 — "항상 재게이트"라는 무성의한 수리가 이 스위트를 통과하지 못한다.

수리 후 (GREEN, 관측):

```
$ go test ./internal/hook/ -run 'TestSyncGateFailState' -count=1
ok  	github.com/modu-ai/moai-adk/internal/hook	121.912s
```

### E5 — 뮤턴트 2건 (새 절이 실제로 테스트에 붙들려 있는가)

| 뮤턴트 | 무엇을 되돌렸나 | 관측 |
|---|---|---|
| M1 | 다줄 레코드 가드 제거 | `R4` FAIL — `stub invoked 0 time(s) on this call; want >= 1`, `stdout is not a block` |
| M2 | 하위호환 절(3번째 필드 부재 = 일치) 제거 → 비어있지 않은 일치 요구 | 기존 release-blocking 행 `AC-005/TA1` FAIL — `stub invoked 2 time(s); want 0`, `stdout carries "decision"; want none` |

두 뮤턴트 모두 실행 후 백업(`gate.orig.sh`)에서 복원했고, 복원 결과가 템플릿 사본과 바이트 동일함을 확인했다(E6).

### E6 — 사본 2벌 동일성

```
$ shasum -a 256 .claude/hooks/moai/sync-phase-quality-gate.sh \
                internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh
a95a733873ba5918bc160d631b9610fe6b1f0a75d86f6ffb75f39d24134d96a1  .claude/hooks/…
a95a733873ba5918bc160d631b9610fe6b1f0a75d86f6ffb75f39d24134d96a1  internal/template/templates/.claude/hooks/…
$ bash -n (양쪽)   # 둘 다 exit 0
$ grep -nE 't[0-9]{3}|SPEC-|H0[0-9]|2026-' internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh
(무출력 — 템플릿 중립성: 카드 id·SPEC ID·내부 날짜 유출 0건)
```

### E7 — 형식·정적검사·영향 패키지

```
$ gofmt -l internal/hook/                 # 무출력
$ go vet ./internal/hook/ ./internal/template/    # exit 0
$ go test ./internal/template/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/template	67.508s
```

### E8 — 선재 레드 귀속

`go test ./internal/hook/... -count=1`에서 루트 패키지 FAIL. 두 건 모두 `session_start_parallel_test.go`:

- `TestSessionStart_DeferredScanJoinsWithinBound/slow_scan_drops_advisory` — 단독 실행 시 **PASS**. 부하 민감.
- `TestSessionStart_DeferredScanDoesNotBlockReturn` — 단독 실행에서도 FAIL (`Handle blocked 514ms; expected deferred`).

후자의 귀속 측정: 이 카드의 테스트 파일만 `git show HEAD:` 버전으로 되돌린 뒤 같은 테스트 재실행 →

```
--- FAIL: TestSessionStart_DeferredScanDoesNotBlockReturn (0.52s)
    session_start_parallel_test.go:97: Handle blocked 514.263208ms waiting for advisory scan; expected deferred (non-blocking) return
```

변경 전 트리에서 동일하게 실패하므로 **develop 9935e4e3e의 선재 레드**이며 이 카드에 귀속되지 않는다. 측정 후 파일을 복원했다.

## Baseline-attribution

모든 수치는 이 실행에서, 이 트리에 대해 측정했다.

| 항목 | 귀속 |
|---|---|
| 트리 | `.claude/worktrees/t601`, `WT-quality-gate-sentinel` @ `9935e4e3e` |
| 수리 전 대상 | `git show HEAD:.claude/hooks/moai/sync-phase-quality-gate.sh` → `.moai/reports/t601/base/…` |
| 수리 후 대상 | 작업 트리 사본, sha256 `a95a7338…d96a1` |
| 프로브 증거 | `probe-baseline.json` / `probe-after.json` / `compare.txt` (모두 이 실행에서 생성) |
| Go 스위트 | `go test ./internal/hook/ -run TestSyncGateFailState -count=1` → `ok … 121.912s` |
| 템플릿 스위트 | `go test ./internal/template/ -count=1` → `ok … 67.508s` |

카드가 인용한 감사 증거(`probes-shell.json` sync_sentinel, `main 2213871af` 사본)는 **다른 트리의 측정**이므로 이 실행의 baseline으로 쓰지 않았다. E2가 그 사실을 명시한다.

## Gaps — 관측하지 않은 것

- **전체 스위트를 돌리지 않았다.** 범위는 `internal/hook`(루트 + 하위)과 `internal/template`뿐이다. 전 패키지 판정은 CI 몫이다(§4 검증 부하 규율).
- **darwin 외 플랫폼에서 재지 않았다.** `:(top,exclude)` pathspec, `shasum`/`sha256sum`/`cksum` 폴백 사슬, `read -r … <<<` 히어스트링을 Linux·Windows에서 실행하지 않았다. CI 매트릭스가 판정한다.
- **`hash_stdin`의 2·3번째 분기를 실행하지 않았다.** 이 머신에는 `shasum`이 있어 첫 분기만 탔다. `sha256sum`/`cksum` 경로는 미실행이다.
- **`moai` CLI가 PATH에 없는 픽스처에서만 쟀다.** `consume_snapshot`이 실제 스냅숏을 반환하는 경로와 이 변경의 상호작용은 관측하지 않았다.
- **실제 Claude 세션에서 실행하지 않았다.** 게이트가 내는 JSON을 런타임이 수용하는지는 카드의 검증 한계 그대로 미검증이며, 이 카드의 범위도 아니다.
- **성능을 수치로 재지 않았다.** `git diff HEAD`가 이 저장소에서 warm 1.6~2.4s로 관측됐지만(Residual-risk 참조), 훅 전체 지연을 before/after로 측정하지는 않았다.
- **H07(243행)·H08(267행)·H09(62행)은 건드리지 않았다.** 같은 파일이지만 별개 카드이며 리드가 직렬로 배차한다.

## Residual-risk

1. **스킵 경로에 워크트리 순회 비용이 새로 생긴다.** `git diff HEAD`가 이 저장소에서 warm 1.6~2.4s로 관측됐다. 이 지점은 이미 후보 언어마다 `git diff --name-only`를 돌던 곳이고 게이트의 Stop 예산은 60s, 실행 경로는 `go vet ./...`+`go build ./...`이므로 비율상 작다. 그러나 더 큰 저장소에서 이 순회가 예산을 압박할 수 있다.
2. **좁은 경우에 오늘보다 느슨해질 수 있다.** 트리를 고친 뒤 툴체인이 PATH에서 사라지면, 오늘은 낡은 block을 재전달하지만 수리 후에는 재게이트 → 도구 부재 → skipped → pass가 된다. SPEC-SYNC-GATE-FAILSTATE-001 D2의 "어떤 경로도 오늘보다 느슨해선 안 된다"에 대한 유일한 예외 후보이며, 인위적인 조합이다.
3. **git-ignore된 파일은 식별자에 들어가지 않는다**(`--exclude-standard`). 무시되는 `.go` 파일을 빌드가 실제로 읽는 구성에서는 그 변경이 재게이트를 일으키지 않는다. 대안(무시 파일 포함)은 `.moai/logs` 류의 매턴 변동으로 메모를 무력화하므로 택하지 않았다.
4. **하위호환 절이 낡은 레코드에 오늘의 동작을 남긴다.** 3번째 필드 없는 레코드는 일치로 취급되므로, 그 레코드가 살아 있는 동안은 트리를 고쳐도 낡은 block이 재전달된다. 첫 실행이 3필드 레코드를 쓰면 해소된다. 이 절을 빼면 기존 release-blocking 행이 깨진다(E5 M2).
5. **`cksum` 폴백은 충돌 저항이 약하다.** 서로 다른 두 트리가 같은 `cksum` 값을 갖는 경우 재게이트가 일어나지 않는다. 충돌 저항이 아니라 안정성만 필요한 자리라 택했고, 그 판단을 코드 주석에 남겼다.
6. **이름에 개행이 든 파일.** 식별자 계산은 줄 기반이라 그런 이름 앞에서 의미가 흐트러질 수 있다. 결정성은 유지되므로 재게이트 판정이 틀리지는 않으나, 관측하지 않았다.

## 변경 파일

| 파일 | 성격 |
|---|---|
| `.claude/hooks/moai/sync-phase-quality-gate.sh` | 수리 + 헤더 계약 문서 갱신 |
| `internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` | 미러 (바이트 동일) |
| `internal/hook/sync_gate_failstate_test.go` | 신규 행 4개 + 레코드 형식 단정 2곳을 필드 기준으로 |
| `.moai/reports/t601/**` | 증거 (프로브·대조·base 사본·본 문서) |

## 후속 (리드 판단 사항)

- **F1** — `.git_hooks/pre-commit:51`이 `go vet` 출력을 `>/dev/null 2>&1`로 버린다(GH #1679 잔여). 프로젝트 큐가 t601로 담고 있던 항목이며, 리드가 신규 카드 후보로 등록했다.
- **F2** — H07·H08·H09는 같은 파일이므로 이 병합 뒤 직렬로.
- **F3** — 선재 레드 `TestSessionStart_DeferredScanDoesNotBlockReturn`은 별개 결함이다(E8). 카드가 없다면 발행 대상.
- **F4** — `hash_stdin`의 `sha256sum`/`cksum` 분기는 이 머신에서 도달 불가라 미실행이다. Linux CI가 두 번째 분기를 태우는지 확인할 가치가 있다.
