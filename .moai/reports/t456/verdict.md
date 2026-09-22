# t456 판정 — statusline 이 착지 판정을 읽는다

카드: statusline 렌더러가 `Backlog.Picked/Queued` 를 그대로 찍고 그 착지 판정을 묻지 않는다.
Class B (plan 생략, run → sync). 브랜치 `WT-statusline-landed`.

Implementation Kickoff Approval: 리드 경유 운영자 승인(github.go 패턴 안). 승인 전 커밋은
측정 1건(`8915dcb2a`)뿐이고 구현은 승인 후에 들어갔다.

## 무엇을 바꿨나

렌더 경로는 작은 캐시 파일 1회 읽기만 한다. 착지 질의는 TTL 이 지났을 때 분리된 자식이
`git log <ref> --format=%B` **1회**로 전량 계산한다. 카드마다 묻지 않는다.

표시는 **뺄셈이 아니라 주석**이다:

```
판정 있음:  🔄 TODO: 39/5 ✓23
판정 없음:  🔄 TODO: 39/5
```

`github.go` 의 기전을 그대로 재사용했다 — 캐시 읽기 / 분리 자식 / 캐시 타임스탬프 stampede 가드 /
`isSelfInvocable` fork-bomb 가드. 두 번째 기전을 만들지 않았다(같은 로직이 세 벌이 된 t441 의 반대편).

## Claim / Evidence

측정 스냅샷을 하나로 고정했다. **origin/develop = `d592b0551eeb731e5bbd3ef330bf71b21c0822c9`**,
판정 바이너리는 이 트리에서 빌드한 것(`go build -o bin/moai-t456 ./cmd/moai`), 로컬 darwin, CI 아님.

### C1 — 렌더 경로 git 서브프로세스는 늘지 않았다 (2 → 2)

t305 가 남긴 spawn 카운팅 shim 을 재사용해 기계적으로 셌다. 코드를 읽어 판단하지 않았다.

```
baseline (구현 전, 같은 shim):   rev-parse --git-dir --show-toplevel
                                 status --porcelain --branch          → 2

구현 후 (캐시 웜, 주석 렌더됨):  rev-parse --git-dir --show-toplevel
                                 status --porcelain --branch          → 2
                                 렌더: 🔄 TODO: 39/4 ✓11
```

자식이 뜨지 않도록 non-self-invocable 이름(`bin/moai-t456`)으로 측정했다 — `isSelfInvocable`
가드가 basename `moai` 만 통과시키므로, 이름 자체가 렌더 경로를 격리하는 수단이 된다.

### C2 — 자식은 카드 수와 무관하게 git 을 1회 부른다

콜드/웜 쌍대조. 억제만 보면 판별력이 없으므로 양성 대조를 붙였다.

```
콜드 캐시:  rev-parse … / status … / log origin/develop --format=%B   → 3
웜 캐시:    rev-parse … / status …                                    → 2
```

콜드의 3번째 줄이 자식의 유일한 질의다. 그 시점 picked 는 39장이었고 질의는 1회였다.
웜에서 자식이 뜨지 않는 것이 stampede 가드의 동작이다.

### C3 — 캐시 상태 5종이 판별력을 갖는다

```
A. measured 캐시(landed:11)            → 🔄 TODO: 39/4 ✓11
B. 손상 캐시 ("not json{")             → 🔄 TODO: 39/4
C. placeholder (measured:false)        → 🔄 TODO: 39/4
D. 관측된 0 (measured:true, landed:0)  → 🔄 TODO: 39/4 ✓0
E. 캐시 부재                            → 🔄 TODO: 39/4
```

A·D 는 붙고 B·C·E 는 안 붙는다. **이 매트릭스는 한 번 공허하게 통과한 적이 있다** — §정정 1.

### C4 — 1회 훑기(RE2)가 배포 기준(카드별 git PCRE)과 같은 답을 낸다

핵심 정확성 주장이다. 한 스냅샷 안에서 세 방법이 독립적으로 같은 수를 냈다:

```
구현 (Go RE2, 1회 훑기):        {"landed":23,"ref":"origin/develop","measured":true}
독립 재현 (BSD grep ERE + comm): 23
picked 모수 (SQLite 직독):        39
```

엔진 대조는 표본이다(전수 아님 — 워크트리 가드가 반복문을 거부해 39회 질의를 못 돌렸다).
fold 가 착지로 본 2건, 미착지로 본 3건을 카드별 PCRE 로 확인:

```
t191 → 20   t216 → 3        (fold: 착지)      일치
t339 → 0    t345 → 0   t348 → 0   (fold: 미착지)  일치
tZZZZ9 → 0  (음성 대조: 존재하지 않는 id)
```

### C5 — 렌더 지연은 회귀하지 않았다 (+1.53ms, +1.0%)

t305 의 교대 실행 방법론을 그대로 썼다. 시각이 다른 두 측정은 부하가 달라 비교가 안 되므로,
두 바이너리를 한 실행 안에서 번갈아 돌린다.

```
13:34  load averages: 6.42 4.88 8.35
base   n=25 median=  161.02ms p95=  187.25ms
after  n=25 median=  162.55ms p95=  186.42ms
delta  median +1.53ms (+1.0% of baseline)
```

파일 1회 읽기 추가분이며 잡음 범위다.

### C6 — 테스트

```
$ go test -timeout 600s ./internal/statusline/...   → ok 11.946s   EXIT=0
$ go test -timeout 900s ./internal/cli/             → EXIT=0
$ go test ./internal/statusline/ -run TestCountNamed → ok 0.382s   EXIT=0
```

`countNamed` 의 단어경계 동작은 처음에 **덮여 있지 않았다**. 뮤턴트(`strings.Contains` 치환)로
비공허성을 확인했고, 두 경우가 정확히 실패했다:

```
--- FAIL: .../a_substring_is_not_a_mention          countNamed(ids=[t45]) = 1, want 0
--- FAIL: .../prefix_collision,_only_the_longer_named  countNamed(ids=[t45 t456]) = 2, want 1
```

양성 대조(경계 문자 5종)는 뮤턴트 아래서도 초록으로 남아, 음성 케이스의 중복이 아니라
대조군 역할을 한다.

## 정정 — 이 판정서가 스스로 뒤집은 것 두 건

기록하는 이유는, 둘 다 **통과처럼 보이는 실패**였기 때문이다.

### 정정 1 — 캐시 매트릭스가 한 번 공허하게 통과했다

처음 C3 을 돌렸을 때 다섯 경우가 **전부** 주석 없이 동일하게 나왔다. B·C·E 만 보면
"unknown 처리 정확"으로 읽힌다. 실제 원인은 구현 **이전**(10:03) 빌드를 썼다는 것이고,
그 빌드에서는 어느 경우에도 주석이 붙지 않는다. 단서는 붙어야 할 A·D 가 함께 실패한 것이었다.

트리에서 다시 빌드해 재측정한 뒤에야 매트릭스가 A·D=붙음 / B·C·E=안 붙음으로 갈린다.
**판정 빌드와 트리가 같아야 한다**는 규율이 여기서 실제로 값을 했다.

### 정정 2 — "엔진 불일치"는 엔진이 아니라 움직인 ref 였다

C4 를 처음 돌렸을 때 fold 는 t191·t305·t327 을 미착지로 봤는데 PCRE 는 20·4·2건을 반환했다.
엔진 차이로 보였다. 실제로는 fold 입력(`--format=%B` 덤프)이 origin/develop 이 `400f37eb9`
였을 때 만들어졌고, 그 사이 ref 가 `d592b0551` 로 전진해 있었다. t191 은 새로 들어온 커밋에서
이름이 불린 것이다.

SHA 를 못박고 한 스냅샷 안에서 다시 재니 세 방법이 23으로 일치한다. **이동하는 ref 를 값으로
붙들면 안 되고, 비교 대상은 같은 스냅샷이어야 한다.**

같은 이유로 초기 측정의 `picked 76 / landed 48` 도 지금은 `39 / 23` 이다. 어느 쪽도 오측이
아니며 큐와 ref 가 움직인 것이다. 그래서 이 문서는 수치를 재는 **명령**을 함께 적었고,
값에는 SHA 와 시각을 붙였다.

기전 판정은 모수가 바뀌어도 그대로다: 39장이어도 카드별 라이브 계산은 약 6.8초이고,
렌더 경로 예산은 서브프로세스 2회다.

## 표시 형태 — 뺄셈을 고르지 않은 이유

배포 기준은 `--grep=\b<id>\b` 로 **전체 커밋 메시지**를 훑는다. 즉 다른 카드의 보고 커밋이
이 id 를 언급하기만 해도 "착지"로 잡힌다. 이 카드 자신이 그 예다 — t456 의 측정 커밋이
t456 을 부른다.

그러므로 39에서 23을 빼면 실제 in-flight 를 조용히 과소보고하고, `moai todo pr` 자신이
주장하는 것보다 강한 의미를 주장하게 된다. 주석은 "판정이 존재하고 23장이 통합 브랜치에
이름이 있다" 는 재조정 신호이지 재도출한 진실이 아니다.

**뺄셈을 원한다면 한 줄 변경이다.** 이 결정은 되돌리기 쉽고, 운영자 판단 사항으로 남긴다.

## Gaps — 관측하지 않은 것

- **엔진 대조는 표본 5건**이다(전수 39건 아님). 워크트리 가드가 반복문 형태의 git 명령을
  거부해 39회 질의를 한 번에 돌리지 못했다. 표본은 착지 2 / 미착지 3 + 음성 대조 1이다.
- **하이픈 포함 id**(`t45-b`)의 경계 동작. `landedCardToken` 과 kanban 의 `validCardToken`
  둘 다 허용하지만 현재 큐는 전부 `tNNN` 이라 실사례가 없다. `\b` 가 토큰 **내부**의 `-` 에
  걸리므로 두 엔진이 갈릴 수 있는 유일한 지점이다.
- **실제 배포 경로의 자식 spawn 은 워크트리 안에서만 관측했다.** primary 체크아웃에
  `.moai/state/landed/` 를 쓰는 실행은 하지 않았다 — 공유 트리에 쓰지 않기 위해서다.
  큐는 primary 사본을 워크트리에 두고 읽혔다.
- **`golangci-lint` 미실행.** `go vet` 만 돌렸다(무출력, exit 0).
- **CI 판정 없음.** 위는 전부 로컬 darwin 단일 플랫폼이다. windows/linux 매트릭스는 CI 몫이고,
  이 브랜치는 아직 push 되지 않았다.

## 상속된 적색 — 본 카드 소관 아님

`make build` 가 선행 검사에서 실패한다:

```
agent-emit drift: committed .codex/agents/moai/*.toml differ from the .md source layer
```

develop 팁에서도 동일하며 본 브랜치가 develop 대비 더한 것은 `.moai/reports/t456/` 와
statusline 변경뿐이다. 고치지 않았다. 측정 바이너리는 `go build ./cmd/moai` 로 우회 빌드했다.

## Residual-risk

- 캐시가 stale 인 동안 옛 수가 보인다. github 카운트와 같은 성질의 트레이드오프이며,
  TTL 10분이 그 상한이다.
- `--format=%B` 전량 훑기는 이력 길이에 선형이다(현재 5,761 커밋 / 약 0.35s). 자식 쪽이라
  렌더는 물지 않지만, 이력이 크게 자라면 `landedScanBudget`(20s)을 다시 재야 한다.
- 주석이 붙은 수는 "이름이 불렸다"이지 "완료됐다"가 아니다. 운영자가 그것을 완료로 읽으면
  이 카드가 막으려던 것과 반대 방향의 오독이 된다 — 뺄셈을 고르지 않은 이유와 같은 뿌리다.
