# 독립 sync-audit — SPEC-INTEGRATION-LOCK-LIVENESS-001 (카드 t298)

fresh-context `sync-auditor` 판정. 감사자는 구현물도 sync 산출물도 저작하지 않았다.

**판정: PASS (clean)**
**종합 점수: 0.9189** (4차원 조화평균, Tier M PASS 임계 0.80)
측정 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t298`, 브랜치 `WT-integration-lock`.
본문 측정은 HEAD `f8b7264ba`에서 수행했고, 감사 중 착지한 `2b49785de`에서도 그대로 성립한다(§J).

---

## §0 Provenance — 이 경로는 도착했을 때 비어 있지 않았다

**이 절은 여섯 달 뒤의 독자를 위한 것이다.** 판정을 읽기 전에 먼저 읽어야 한다.

이 감사가 시작될 때 `.moai/reports/t298/sync-audit.md`에는 **다른 에이전트(`manager-docs`)가
쓰고 커밋한 별도의 품질 판독**이 이미 들어 있었다. 그 문서는 스스로 자기 §0에서 "이것은
fresh-context `sync-auditor` 판정이 아니다"라고 밝히고 있었고, 헤더에 **PASS-WITH-DEBT /
조화평균 0.924 / Functionality 0.94 · Security 0.88 · Craft 0.95 · Consistency 0.93**을 달고 있었다.

**그 문서는 지워지지 않았다 — 이 경로에서 치워졌을 뿐 커밋에 남아 있다.** 복원 주소는
`2b49785de:.moai/reports/t298/sync-audit.md`이고, `git show`로 그대로 읽을 수 있다
(첫 행이 `# Sync-phase Quality Assessment: …`임을 확인했다). 여기 옮겨 붙이지 않은 것은 의도적이다 —
한 파일에 판정 두 개가 나란히 서면 독자가 어느 것이 유효한지 판단해야 하고, 그건 매달린 포인터보다
나쁘다. 주소를 남기고 본문은 두지 않는다.

**감사자는 그 파일을 읽었다. 그리고 자신의 점수를 정하기 전에 읽었다.** 숨길 일이 아니므로
그대로 적는다 — 당시 판단은 "다른 사슬의 증거를 파괴하지 않는다"였고, 그래서 덮어쓰지 않고
아래에 덧붙이는 형태를 택했다. 그러나 그 선택이 **앵커링 오염**을 만들었다: 첫 도출에서
must-pass 두 차원이 오염원과 **한 자리도 다르지 않은 0.94 / 0.88**로 나왔다.

리드가 그 일치를 지적했고, 이 판이 그에 대한 대응이다. 조치는 셋이다.

1. 오염원 문서는 리드가 이 경로에서 제거했다(그 시점까지 그 에이전트 자신의 내용임이 확인됐다).
   지금 이 파일은 이 감사의 것이다.
2. **must-pass 두 차원을 자기 증거만으로 재도출했다**(§F.1). 재도출은 Functionality를
   0.94 → **0.95**로, Security를 0.88 → **0.87**로 옮겼다. 두 이동의 근거는 각각 §F.1에 적혀 있고,
   둘 다 오염원이 도달할 수 없었던 근거다 — 하나는 §G 잔존을 Functionality에서 빼야 한다는 판단,
   다른 하나는 오염원이 발견하지 못한 F2다.
3. Craft(0.92) · Consistency(0.94)는 첫 도출에서 이미 오염원(0.95 / 0.93)과 달랐고
   근거도 자기 측정이므로 그대로 둔다.

**보존한다고 믿었으나 실제로는 단독 작성이었다(자기 결함, 교정 완료).** 감사자가 덧붙이기를
실행한 시점에 그 에이전트의 제거가 이미 경로를 비운 뒤였다. 그래서 "덧붙인" 내용이 파일 전체가
됐고, 그 결과 서두가 **어디에도 없는 §0을 가리키는 매달린 인용**을 한동안 지니고 있었다.
해소되지 않는 인용은 결함이며, 이 감사가 남의 작업에서 지적하는 것과 같은 종류다.
지금 판은 단독 문서로 다시 썼고 그 인용은 남아 있지 않다.

**주의해서 읽을 것**: 재도출 뒤 종합 점수는 0.9193 → **0.9189**로 사실상 제자리다. 이것은
앵커링의 잔재가 아니라 **두 보정이 서로 반대 방향으로 움직여 상쇄된 결과**다 — Functionality는
올랐고 Security는 내렸다. 종합의 근접이 아니라 차원별 도출을 읽어야 한다.

판정의 **종류**는 처음부터 오염원과 달랐다(오염원 PASS-WITH-DEBT ↔ 이 감사 PASS clean).
오염원의 debt는 "독립 감사가 돌지 않았다"는 절차 공백이었고, 이 감사가 바로 그 공백을 메운다.

---

## §A Claim — 무엇을 주장하는가

1. AC-INL-001..013 중 **13/13 관측 green, 0 fail, 0 미관측**. `progress.md` §E.3이 "개별
   재단언하지 않았다"고 남긴 AC-INL-005 / AC-INL-006도 이 실행에서 직접 돌려 확인했다.
2. `spec.md` §G의 직렬화 잔존 위험은 **여전히 열려 있고**, M5 산문이 그것을 덮지 않았다.
3. must-pass 두 차원(Functionality **0.95** / Security **0.87**)이 각각 독립적으로 임계를 넘는다.
4. 발견된 결함 4건은 **전부 optional**(차단 아님). blocking 0건.

## §B Evidence — 이 실행에서 돌린 명령과 관측 출력

패키지 스위트 (`go test ./internal/kanban/... ./internal/session/... ./internal/hook/...`, rc=0):

```
ok  github.com/modu-ai/moai-adk/internal/kanban   40.138s
ok  github.com/modu-ai/moai-adk/internal/session  24.650s
ok  github.com/modu-ai/moai-adk/internal/hook     64.006s
(hook 하위 9개 패키지 전부 ok)
```

교차프로세스 + 기존 통합락 CLI 테스트
(`go test ./internal/cli/ -run 'TestIntegrationOwnerLiveness|TestIntegrationAcquire|TestIntegrationRelease|TestIntegrationStatus' -v -count=1`, rc=0):

```
--- PASS: TestIntegrationOwnerLiveness_AncestryPathHoldsAfterAcquireCLIExits (5.27s)
--- PASS: TestIntegrationOwnerLiveness_EnvStampHoldsAfterAcquireCLIExits (5.73s)
--- PASS: TestIntegrationOwnerLiveness_BareAcquireRefusesLiveHolder (4.10s)
--- PASS: TestIntegrationAcquire_RefusesASecondLane (0.09s)
--- PASS: TestIntegrationAcquire_ForceReportsWhatItDisplaced (0.08s)
--- PASS: TestIntegrationAcquire_RefusesWithoutASessionID (0.00s)
--- PASS: TestIntegrationRelease_EmptyIsReported (0.00s)
--- PASS: TestIntegrationStatus_FreeWindow (0.00s)
ok  github.com/modu-ai/moai-adk/internal/cli  17.064s
```

가드 3-leg (`-run TestCheckIntegrationLock_FollowsAnchoredLiveness -v`) — **skip 0건, 3-leg 전부 실행**:

```
--- PASS: .../anchored_live_holder_denies (0.00s)
--- PASS: .../anchored_dead_holder_allows (0.00s)
--- PASS: .../anchored_pid-0_holder_denies_conservatively (0.00s)
```

kanban 락 계열 9건 전부 PASS. AC-INL-005 = `TestReleaseIntegrationLock_HolderAndForeign` PASS,
AC-INL-006 = `TestIntegrationAcquire_ForceReportsWhatItDisplaced` PASS,
AC-INL-007 = `TestReadIntegrationLock_LegacyRecordWithoutPIDSource` PASS(skip 아님).

정적 검사: `go vet ./internal/kanban/ ./internal/session/ ./internal/cli/ ./internal/hook/` → rc=0, 무출력.
`GOOS=windows GOARCH=amd64 go vet ./internal/...` → rc=0, 출력 0바이트
(`.moai/state/verify/t298-audit/vet-windows.txt`).

`PIDSource`를 읽는 판정 경로가 있는지
(`grep -rn "PIDSource" internal/ --include='*.go'` 에서 `_test.go` 제외):

```
internal/cli/integration.go:189       PIDSource: kanban.PIDSourceSessionOwner   ← 쓰기
internal/kanban/integration_lock.go:65,71,80,89                                 ← 상수·주석·필드 선언
```

**읽는 곳 0건.** 이 측정 하나가 F1의 등가성 논거와 F4를 동시에 세운다.

문서 AC 3종 (이 트리에서 재측정):

```
$ grep -hoE '직렬화를[[:space:]]*보장|동시.*acquire.*불가능|acquire.*동시.*불가능' \
    .claude/rules/local/gitflow-lane-protocol.md CLAUDE.local.md | wc -l
0                                                       # AC-INL-013 green
$ grep -n "직렬화" .claude/rules/local/gitflow-lane-protocol.md CLAUDE.local.md
.claude/rules/local/gitflow-lane-protocol.md:38:## 3. 직렬화 — 병합 창은 한 번에 한 레인
                                                        # §3 캐비엇 본문 소멸 → AC-INL-010 green
$ awk '/^### §4\.1 /,/^## 5\./' CLAUDE.local.md | grep -c '세션 프로세스'   → 1
$ awk '/^### §4\.1 /,/^## 5\./' CLAUDE.local.md | grep -c '재획득'          → 1
                                                        # AC-INL-011 green (둘 다 ≥1)
```

RED 성립 근거(뮤턴트 방향): 수정 전 `AcquireIntegrationLock`은
`if want.PID == 0 { want.PID = os.Getpid() }`로 **acquire CLI 자신의 pid**를 채웠다(diff로 확인).
교차프로세스 테스트는 자식이 종료한 뒤 부모에서 조회하므로 그 pid는 반드시 죽어 있다 —
즉 세 테스트는 수정 전 코드에서 기계적으로 RED다. 기록된 RED 전사
(`.moai/reports/t298/red-baseline.txt`)의 실패 메시지가 현재 테스트의 단언 문자열과 일치한다.

## §C Baseline-attribution — 무엇에 대고 쟀는가

모든 수치는 이 실행, 이 트리(`f8b7264ba` @ `WT-integration-lock`)에서 잰 것이다.
`progress.md` §E.3의 숫자는 하나도 이월하지 않았다 — AC-005/006을 포함해 전부 재실행했다.
증거: `.moai/state/verify/t298-audit/{suite-3pkg.txt,cli-integration.txt,vet-windows.txt}`.

**process 관측(판정과 별개로 기록)**: 조상 판정 결과 `f8b7264ba`는 `origin/develop`의 조상이고,
좌우 divergence 계수는 `66 0`이었다. 즉 감사 대상 커밋은 **이미 origin/develop에 착지해 있고**,
이 판정은 게이트가 아니라 확인이다.

## §D AC 귀속 표 (13/13)

| AC | 판정 | 관측 근거 |
|---|---|---|
| AC-INL-001 | PASS | `..._AncestryPathHoldsAfterAcquireCLIExits` — (a) stale=false, pid=부모 (b) `pid_source` 원시 JSON 확인 |
| AC-INL-002 | PASS | `..._EnvStampHoldsAfterAcquireCLIExits` |
| AC-INL-003 | PASS (대체 픽스처) | `TestAcquireIntegrationLock_TakesOverAStaleHolder` — F1 참조 |
| AC-INL-004 | PASS | `TestAcquireIntegrationLock_AnchoredPIDZeroIsLiveNotStale` + 가드 pid-0 leg |
| AC-INL-005 | PASS | `TestReleaseIntegrationLock_HolderAndForeign` — 이 실행에서 직접 확인 |
| AC-INL-006 | PASS | `TestIntegrationAcquire_ForceReportsWhatItDisplaced` + 레거시 인수 시 `replaced` 보고 |
| AC-INL-007 | PASS | `TestReadIntegrationLock_LegacyRecordWithoutPIDSource` (실제 실행, skip 아님) |
| AC-INL-008 | PASS | 가드 3-leg 전부 PASS |
| AC-INL-009 | PASS | `GOOS=windows go vet ./internal/...` rc=0 무출력 + `factory_alive_windows.go` diff 없음 — F2 참조 |
| AC-INL-010 | PASS | 위 grep, 캐비엇 본문 소멸 |
| AC-INL-011 | PASS | awk+grep 둘 다 1 |
| AC-INL-012 | PASS | `..._BareAcquireRefusesLiveHolder` (a)(b) 양 절 — F3 참조 |
| AC-INL-013 | PASS | 금지어 sweep 0 + 판단 조항(권한 경계 부인)은 M5 diff 열람으로 확인 |

MUST-PASS 8건(001/002/003/004/007/008/012/013) 전부 green.

## §E Findings — 4건, 전부 optional

- **F1 [MINOR][optional]** `internal/kanban/integration_lock_test.go:214` —
  AC-INL-003의 Given은 "`pid_source: session-owner`가 붙은 앵커 레코드의 owner가 죽은 경우"인데,
  실제 픽스처(`seedDeadHolder`, :238)는 `pid_source` **없는** 레코드에 고정 상수 pid
  `0x7FFFFFF0`을 심는다. 즉 관측된 것은 레거시 형태의 인수다. 등가성은 §B의 grep으로 세웠다 —
  판정 경로에서 `PIDSource`를 읽는 곳이 0건이므로 두 형태의 `Stale()` 결과는 같다. 앵커+사망
  조합 자체는 가드 테스트 `anchored_dead_holder_allows`가 별도로 덮는다.
  **필요한 수정**(선택): `AcquireIntegrationLock` 경로에도
  `PIDSource: PIDSourceSessionOwner` + 죽은 pid 케이스 1건 추가.
- **F2 [MINOR][optional]** `internal/session/anchor_pid_windows.go:14` —
  Windows에서 `isProcessAlive`가 무조건 `true`다. 그래서 `sessionPIDFromEnv`(session_pid.go:99)의
  liveness 검사가 Windows에서 공허해지고, **죽은 pid를 담은 `MOAI_SESSION_PID`가 그대로 기록된다**.
  기록 후 판정은 `kanban.FactoryProcessAlive`(Windows에서는 진짜 `OpenProcess` 프로브)가 하므로
  그 창은 reclaimable로 읽힌다 — `spec.md` §D가 선언한 TREAT-AS-LIVE 방향과 **반대**로 기우는
  유일한 경로이자 `acceptance.md` §D.2 "죽은 스탬프는 건너뛰어야 한다" 엣지 케이스의 미충족이다.
  전제(오래된 env 스탬프 상속)가 좁고 이 SPEC이 도입한 회귀도 아니지만 **§G에 미기재**다.

  **분류 — 이 감사의 판단: (i) §G 문서 누락 + (iii) 후속 카드 감. (ii)는 아니다.**
  근거를 나눠 적는다.
  - **(ii)가 아닌 이유**: 이 카드의 범위는 acquire 경로의 owner 앵커다. 문제의 함수
    `internal/session/anchor_pid_windows.go:14`는 그 경로가 아니라 **세션 레지스트리의 플랫폼
    프로브**이고, 이 SPEC 이전부터 무조건 `true`였다(diff에 이 파일 변경 0건). 이 카드가 고치지
    않아 생긴 결함이 아니라, 이 카드가 만든 앵커가 **기존 플랫폼 프로브 위에 얹히면서 비로소
    보이게 된** 결함이다. 여기서 고치려면 레지스트리의 Windows liveness 의미를 바꿔야 하고,
    그건 이 카드가 명시적으로 건드리지 않기로 한 표면이다(§D.4 간접 검증 경계).
  - **(i)인 이유**: 그러나 `spec.md` §G는 "Windows 런처 스탬프 전파 미검증 → pid-0 보수적 degrade"만
    적고 있다. 그것은 **스탬프가 없을 때**의 경로다. 여기서 빠진 것은 **스탬프가 있는데 죽었을 때**의
    경로이며, 그쪽은 보수적으로 degrade하지 않고 반대로 기운다. §G가 선언한 안전 방향과 어긋나는
    경로가 §G에 없다 — 이것은 이 카드의 산출물이 지금 고칠 수 있고 고쳐야 하는 누락이다.
  - **(iii)인 이유**: 코드 수정 자체(Windows `isProcessAlive` → 같은 파일 :30의
    `probeProcessLiveness`, 이미 undetermined를 반환한다)는 레지스트리·워크트리 앵커 가드까지
    파급되므로 별도 카드에서 그 소비자들과 함께 다뤄야 한다.

  **권고**(카드 발행은 리드의 권한이다): ① `spec.md` §G에 이 비대칭 한 줄 추가 — 이 카드 안에서
  가능하고, 없으면 관리되지 않는 위험이 된다. ② Windows liveness 프로브 교체는 후속 카드.
- **F3 [MINOR][optional]** `internal/cli/integration_lock_owner_liveness_test.go:252` —
  AC-INL-012 (a)는 "오류가 `ErrIntegrationLockHeld` 센티널을 지닌다"를 요구하지만, 교차프로세스
  단언은 exit≠0 과 `sess-a`/`lane-a` 문자열 포함(두 조건을 `&&`로 부정 결합)뿐이다. 센티널 자체는
  in-process `TestAcquireIntegrationLock_AnchoredPIDZeroIsLiveNotStale`의 `IsIntegrationLockHeld(err)`가
  관측한다. 두 관측을 합치면 조항은 충족되나 한 테스트가 단독으로 조항 전부를 세우지는 않는다.
- **F4 [MINOR][optional]** `internal/kanban/integration_lock.go:89` — `PIDSource`는 기록만 되고
  어떤 판정 경로도 읽지 않는다(§B grep으로 확인; 주석이 의도임을 명시). 현재로선 옳은 형태지만,
  훗날 이 필드로 분기하기 시작하면 "pid 0 = live" 불변식을 상속이 아니라 **재논증**해야 한다.

세 곳의 `t.Skip` 탈출구(guard:121, guard:206, kanban 레거시)는 이 머신에서 **하나도 발동하지 않았다**
(§B의 -v 출력). 다만 심어둔 고정 pid가 살아 있는 머신에서는 조용히 공허해질 수 있다 — F1과 같은
계열의 잔여 위험이며 결함으로 계수하지 않았다.

## §F 차원 점수

| 차원 | 가중 | 점수 | 근거 |
|---|---|---|---|
| Functionality | 40% | **0.95** | 13/13 AC 직접 관측 green, 기능 결함 0건; 감점은 간접 증거 2건(F1 픽스처 대체, RED 재현 미실행) |
| Security | 25% | **0.87** | 새 입력·권한·서브프로세스 표면 0, 보수적 실패 방향 3곳 확인; 감점은 미선언 F2(무거움) + 선언된 §G 잔존(가벼움) |
| Craft | 20% | 0.92 | 주석이 기계가 아니라 이유를 담음; 스키마 순수 additive; 테스트 격리 완전; 감점은 F4 미사용 필드 + skip 탈출구 3곳 |
| Consistency | 15% | 0.94 | 영어 주석·`snake_case_test.go`·`%w` 래핑 준수; Go diff 6파일 전부 선언된 `module:` 안, drive-by 0건 |

조화평균 = `4 / (1/0.95 + 1/0.87 + 1/0.92 + 1/0.94)` = **0.9189** ≥ 0.80.
must-pass 방화벽: Functionality 0.95 통과, Security 0.87 통과 — 각각 독립 통과.

### §F.1 must-pass 두 차원 재도출 (앵커링 대응)

§0이 기록한 오염 때문에, 아래 두 도출은 오염원 수치를 참조하지 않고 자기 증거만으로 다시 세웠다.

**Functionality: 0.94 → 0.95 (상향).**
1.0에서 깎을 근거를 자기 증거에서만 찾으면 두 가지다 — (i) F1: AC-INL-003의 *Then*은
직접 관측했으나 *Given*의 형태가 다르고, 등가성은 실행이 아니라 §B의 grep(판정 경로에서
`PIDSource` 읽기 0건)으로 세웠다. (ii) RED→GREEN 쌍의 RED 절반을 수정 전 코드에 대고 직접
재실행하지 않았다(소스 뮤테이션이 쓰기 범위 밖이라 diff 판독 + 기록 전사 대조로 간접 확인).
**둘 다 기능 결함이 아니라 증거의 간접성**이고, 13개 기준 중 2건이다. 기능 결함은 0건이다.
결정적으로, 첫 도출에서 무의식적으로 반영됐던 "§G 직렬화 잔존 때문에 만점이 아니다"라는
감점을 **여기서 제거했다** — Functionality는 선언된 AC에 대고 채점하는 차원이고, 명시적으로
범위 밖인 잔존은 Security와 잔여 위험이 받아야 한다. 그 감점을 옮기면 **0.95**다.

**Security: 0.88 → 0.87 (하향).**
가점 근거: 새 외부 입력·서브프로세스·네트워크 표면 0, 파일 모드 불변(0o755/0o644, diff로 확인),
`PID <= 0` → live의 보수적 방향을 서로 독립인 3곳(kanban pid-0 테스트 / 가드 pid-0 leg /
pid-0 위 bare acquire 거부)에서 관측, 파싱 실패는 free window가 아니라 error.
감점 근거 둘: **F2(미선언)**와 **§G(선언·문서화·범위 밖)**. §G는 두 배포 문서가 모두 명시하고
카드가 범위 밖으로 못박았으므로 처리 수준이 모범에 가깝고 비용이 작다. F2는 반대다 —
적히지 않은 비대칭은 관리될 수 없고, 하필 §D가 절대 일어나선 안 된다고 선언한 방향이다.
첫 도출은 감점을 사실상 §G 하나로 셈했다. **F2를 발견한 감사는 그만큼 더 깎아야 일관된다.**
→ **0.87**.

**종합이 거의 그대로인 이유**: 두 보정이 반대로 움직여 0.9193 → 0.9189로 상쇄됐다.
근접은 앵커링의 잔재가 아니라 상쇄의 결과다. 판정(PASS clean)은 두 도출 어느 쪽에서도 바뀌지 않는다.

## §G §G 잔존 판정 — 열림, 정직하게 서술됨

재측정(`grep -n "Flock\|flock\|LockFile" internal/kanban/integration_lock.go internal/cli/integration.go`):
`integration_lock.go` 5행(:15, :19 = 패키지 헤더 산문의 `flock`; :38, :39, :128 = 상수 주석·선언·경로 결합),
`integration.go` **0행**. **호출부 없음** — 오케스트레이터의 판독을 **확증**한다.
`AcquireIntegrationLock`(:177-208)은 `ReadIntegrationLock` → 판정 → `writeIntegrationLock`을
사이에 아무 잠금 없이 수행하고, 스테이징 경로는 모든 동시 작성자가 공유하는 고정 `path + ".tmp"`(:257)다.

M5 산문이 이것을 덮지 않았음을 diff로 확인했다. 두 문서 모두 명시적이다 —
gitflow-lane-protocol §3: "`acquire` 자체가 읽고-고치고-쓰는 과정을 갈라 세우지 않으므로 …
**리드 공지가 여전히 첫 번째 층**"; CLAUDE.local.md §4.1: "조율 신호이지 권한 경계가 아니다".
AC-INL-013의 판단 조항(권한 경계로 읽히는 표현 금지)까지 충족한다.
**잔존은 정직하게 유지됐다 — 종이로 덮은 흔적 없음.**

## §H Gaps — 관측하지 **않은** 것

- **전체 스위트 미실행.** 영향 패키지(kanban/session/hook + cli 통합락 계열)만 돌렸다.
  전 패키지 판정은 CI 몫이며 이 감사가 대체하지 않는다.
- **`internal/cli` 전체 미실행.** `-run` 선택자로 8개 케이스만 돌렸다(실행 케이스 수 확인).
  나머지 케이스의 회귀 여부는 미관측 — 단, 이 트리는 이미 origin/develop 조상이므로 CI가 판정했다.
- **Windows 동작 미실행.** `go vet`은 컴파일만 증명한다. F2도 실제 Windows 호스트 실행이 아니라
  코드 판독으로 도출했다.
- **PID 재사용 미시험.** §G 선언 사항, 이 감사에서도 재현하지 않았다.
- **수정 전 코드에 대한 실제 RED 재현 미실행.** 소스 뮤테이션은 이 감사의 쓰기 범위 밖이라,
  RED 성립은 (i) 수정 전 diff 판독과 (ii) 기록된 `red-baseline.txt` 전사의 단언 문자열 일치로
  간접 확인했다. 직접 재현은 아니다.
- **경쟁 acquire 미재현.** §G가 서술한 두 레인 동시 acquire 경합은 재현하지 않았다.
- **부수 관측(감사자 귀속)**: `go test ./internal/hook/...` 실행이
  `.moai/specs/SPEC-HOOK-PRETOOL-PERF-001/{baseline,postchange}.md`를 재작성했다(알려진 픽스처
  재작성 동작). 감사자가 즉시 원복했고 작업 트리 상태로 확인했다. 이 SPEC의 결함 아님.

## §I Residual risk — 관측한 것에도 불구하고 틀릴 수 있는 것

- 이 수정은 창을 **의미 있게** 만들었고, 그래서 §G 경합 창의 비용을 낮춘 게 아니라 **높였다**.
  수정 전에는 어차피 아무것도 직렬화되지 않아 경합이 가려져 있었다. 리드 공지 층을 지금
  걷어내면 F2/§G 조합이 곧바로 사고로 나타난다.
- 업그레이드 이전에 잡힌 레거시 창은 여전히 인수 가능하게 읽힌다. 두 문서가 재획득을 지시하지만
  이는 문서 층 보장이지 기계 층 보장이 아니다.
- 조상 걷기가 앵커를 실제 세션이 아닌 더 오래 사는 조상(터미널·tmux)에 걸면 창이 세션 종료 뒤에도
  살아 있게 읽힌다. `acceptance.md` §D.2가 선언한 방향이고 보수적 쪽이지만 `--force` 없이는 풀리지 않는다.
- F4: `PIDSource`의 무해함은 "읽는 곳이 없다"에 전적으로 의존한다(§B에서 재확인).
- **이 판정 자체의 잔여 위험**: §0의 오염은 재도출로 처리했으나, 앵커링은 정의상 자기관측이
  어렵다. 재도출된 두 수치가 여전히 오염의 영향 아래 있을 가능성을 완전히 배제하지는 못한다.
  그래서 §F.1은 결론이 아니라 **도출 과정**을 적었다 — 독자가 수치가 아니라 논거를 검증할 수 있도록.

## §J HEAD 드리프트

본문의 모든 측정은 `f8b7264ba`에서 수행했다. 감사 도중 `manager-docs`의 sync 커밋 `2b49785de`가
착지했고, 그 diff는 `.moai/reports/t298/{lane-verify-cli.txt,sync-audit.md}`, `progress.md`,
`spec.md`(1행), `CHANGELOG.md`(1행)뿐이다 — **Go 파일 0건, 금지어 sweep 대상 두 문서 0건**이므로
위 판정과 모든 수치는 `2b49785de`에서도 그대로 성립한다. 재도출 시점에 현재 HEAD를 다시 읽어
`2b49785de`임을 확인했다.

---

독립 `sync-auditor` 감사. 감사자는 구현물과 sync 산출물 어느 쪽도 저작하지 않았고,
쓰기 범위는 이 파일 하나였다. 증거: `.moai/state/verify/t298-audit/`.
오염 이력과 그 처리: §0. must-pass 두 차원의 재도출 과정: §F.1.
