# SPEC-FULL-SUITE-DOCTRINE-001 — 진행 기록

- card: t301
- 워크트리: `.claude/worktrees/t301` (branch `WT-full-suite-doctrine`), 측정 HEAD `d29b8942e`
- Tier: M (`spec.md` frontmatter `tier: M`)
- 현재 버전: 0.5.0

## §E.1 Plan-phase Audit-Ready Signal

### iteration 1 — 산출 (v0.1.0)

- 산출물: `spec.md` / `plan.md` / `acceptance.md` / `progress.md` (Tier M 4종).
- 좌표 재측정에서 정정 1건: `manager-develop.md` 의 전량 지시 지점은 3곳이 아니라 **4곳**(신규 S4).
- SPEC ID 정규식 자가 점검 — 실행 출력:

```
$ ID="SPEC-FULL-SUITE-DOCTRINE-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL
PASS
```

### iteration 1 감사 — FAIL 0.57

Must-Pass 7항 통과, 차단 결함 9건 + optional 2건. 기록: `.moai/reports/t301/plan-audit.md`.

### iteration 2 — 대응 (v0.2.0)

| 결함 | 조치 | 근거 |
|---|---|---|
| D1 브랜치명 유출 | REQ-FSD-006 재작성(`integration branch`), REQ-FSD-007 신설, AC-FSD-009 패턴 확장 | `grep -rl 'origin/develop' internal/template/templates/` → 매치 없음, `rc=1` |
| D2 AC-006 지연 | 구간(`# 1.`~`# 2.`) 한정 판정으로 확정, 위임 문장 삭제 | 구간 4줄, 전량 호출 각 `1` |
| D3 부재형만 존재 | 존재형 AC 2종 신설 (AC-012 / AC-013) | 사전 baseline 각 `0` |
| D4 dangling 무방비 | 판별자 **제거** 확정, AC-014 신설, plan 존치 권고 삭제 | `LARGE_SCALE` 각 `3` |
| D5 시간 추정치 무AC | AC-015 신설 | Group A 행 패턴 각 `1` |
| D6 unstaged 전용 diff | 기준 SHA 고정 + 훑은 줄 수 하한 | 편집 전 훑은 줄 `0` |
| D7 `tier:` 부재 | `tier: M` 추가 | — |
| D8 report-not-verdict | AC-005 자기완결화, AC-009 첫 블록 삭제 | Group A 행 내 전량 호출 각 `1` |
| D9 AC-007 얕음 | STEP 5 구간 한정 + 정본 토큰 판정 | 구간 토큰 사전 baseline 각 `0` |

optional 2건: D11 부분 수용(지시문 층 중립 / 예시 층 기존 유지), D12 기각(`internal/spec/lint.go:447` 의 `reqLinePattern` 이 `\d{3}-\d{3}` 두 그룹 꼬리를 요구하므로 `REQ-FSD-001` 은 표기를 어떻게 바꿔도 매치 불가 — 표기 변경은 GEARS 라벨 가독성만 잃고 안전망은 못 얻는다).

### iteration 2 감사 — FAIL 0.75 (Δ +0.18, 회귀 없음)

iter-1 결함 9건 중 7건 완전 종결, 2건 부분 종결. **D12 기각은 감사가 독립 확인해 정당하다고 판정.** 신규 차단 결함 1건. 기록: `.moai/reports/t301/plan-audit-iter2.md`.

### 운영자 결정 — 반복 상한 연장

Tier M 상한 2회에 도달했으나, 남은 결함 4건이 전부 좌표·기대치가 확정된 국소 편집이고 점수 회귀가 없어 **운영자가 iteration 3 연장을 승인**했다. 감사의 권고 경로(1번)와 일치한다. PASS-with-debt는 채택하지 않았다 — N1은 배포되는 drift를 알면서 남기는 선택이고, 그 상태로 run-phase가 AC를 전부 통과시키면 완료 보고가 관측되지 않은 완료 주장이 된다. iteration 3이 **마지막 라운드**다.

### iteration 3 — 대응 (v0.3.0)

**N1 (critical) — 세 번째 배포 사본을 범위에 편입.**

`internal/template/templates/.codex/agents/moai/manager-develop.toml` (19,127 바이트)은 같은 에이전트 정의의 Codex 하니스 미러이며 `//go:embed all:templates` 로 배포된다. 네 위반 지점을 문자 그대로 담고 있다(`:80`·`:114`·`:120`·`:123`). **범위에 넣는 쪽(가)** 을 택했다 — 범위 밖 선언(나)은 `spec.md §A.3` 이 배치 정의에 적용한 inert 논리와 비대칭이 되고, 그 비대칭을 정당화할 근거가 없기 때문이다.

조치: `spec.md §A.1` 을 사본 3벌 × 지점 4개 = 12곳 구조로 재작성, §B에 "세 사본 전부" 범위 규약 신설, REQ-FSD-002·009에 사본 3벌·`.codex` 미러 명시, AC-FSD-001·002·003·012·014의 파일 인자에 C3 추가. **새 AC 없음** — AC 총수 15/16 유지.

C3 사전 baseline (전부 이번에 직접 측정, `.md` 사본과 같다고 가정하지 않음):

| 측정 | C3 값 |
|---|---|
| 리터럴 4패턴 | `3` |
| 어휘족 패턴 | `4` |
| `full suite, coverage` | `1` |
| `LARGE_SCALE` | `3` |
| 정본 스코프 문구 | `0` |
| STEP 4 구간 길이 / 정본 문구 | `14` / `0` |
| STEP 5 구간 길이 / `integration branch` / `PENDING at report time` | `10` / `0` / `0` |

C3에는 로컬 쌍이 없음을 확인했다(`ls .codex/agents/moai/manager-develop.toml` → No such file or directory). 따라서 AC-FSD-008(쌍 델타) 대상에서 제외하고, 부재·존재형 AC가 C3를 직접 재도록 했다 — 판정 공백 없음.

**D3' (major) — 존재형 AC를 지점 고정으로.**

AC-FSD-012가 파일 단위 존재 판정이라, 위반 지점은 동의어로 두고 정본 문구를 다른 자리에 한 번 심는 mutant가 통과했다. STEP 4 블록 구간 한정으로 바꿨다(구간 길이를 Then에 포함). 아울러 감사 권고대로 어휘족 패턴을 부재형에 추가했다 — AC-FSD-002를 (기존의 템플릿판 리터럴 반복에서) 어휘족 판정으로 재배정하고, 리터럴 판정은 AC-FSD-001의 인자 목록으로 흡수해 AC 총수를 늘리지 않았다.

**D9' (major) — 토큰별 판정.**

`grep -c -e A -e B` 가 두 대안 중 하나만 있어도 매치 줄을 센다는 지적이 맞다. 두 토큰을 각각 세고 둘 다 `≥1` 을 요구하도록 나눴다. `PENDING at report time` 단독 판정이 REQ-FSD-008("미결 생략 불가")의 계측점이 된다.

**N2 (minor) — 구간 길이 정정 + Then 편입.**

기재 9줄, 실측 **10줄**(C1·C2·C3 동일). awk 범위 연산자가 종료 패턴 줄을 포함하므로 실질 본문 9줄 + 종료 헤딩 1줄이다. 값을 정정하고 Then 안으로 옮겨, 구간 폭주 시 RED가 되도록 했다. 같은 처리를 AC-FSD-006(4줄)·AC-FSD-012(14줄)에도 적용했다.

### 다른 파일의 `.codex` 쌍둥이 여부

`.codex/` 트리 전체를 나열해 확인했다. 담고 있는 것은 `agents/moai/*.toml` **11개뿐**이며 `rules` 디렉터리 자체가 없다. 따라서:

- `agent-common-protocol-reference.md` — codex 쌍둥이 **없음**
- `verification-batch-pattern.md` — codex 쌍둥이 **없음**
- `manager-develop` 정의 — codex 쌍둥이 **있음** (= C3, 편입 완료)

형제 10개(`plan-auditor.toml` 등)는 다른 에이전트 정의이므로 이 SPEC의 수리와 무관하다 — `spec.md §D` 에 범위 밖 소절로 근거와 함께 기록했다.

### 이번 라운드에서 재측정한 baseline (전부 이 트리, HEAD `d29b8942e`)

| 측정 | C1 | C2 | C3 |
|---|---|---|---|
| 리터럴 4패턴 | 3 | 3 | 3 |
| 어휘족 `(full\|complete\|entire\|whole)[ -](test )?suite` | 4 | 4 | 4 |
| `full suite, coverage` | 1 | 1 | 1 |
| `LARGE_SCALE` | 3 | 3 | 3 |
| STEP 4 구간 길이 | 14 | 14 | 14 |
| STEP 4 구간 내 정본 스코프 문구 | 0 | 0 | 0 |
| STEP 5 구간 길이 | 10 | 10 | 10 |
| STEP 5 구간 내 `integration branch` | 0 | 0 | 0 |
| STEP 5 구간 내 `PENDING at report time` | 0 | 0 | 0 |

부수 확인: `.codex/` 전체 트리 나열(11개 `.toml`, `rules` 없음), `.codex` 로컬 사본 부재. rc를 읽는 모든 측정은 파이프 없이 수행했다.

### iteration 3 감사 — FAIL 0.81 (Δ +0.24 누적, 임계선 0.80 첫 통과)

Must-Pass 7항 통과, 기재 baseline 13개 전부 재현, 구간 길이 단언 3개 전부 검증, iter-2 결함 4건 종결. FAIL은 **신규 critical 1건(N3)** 하나에 걸린다. 기록: `.moai/reports/t301/plan-audit-iter3.md`.

### iteration 4 — 대응 (v0.4.0) · N3 단독 델타

**이 라운드는 전면 재작업이 아니라 N3 좁은 델타다.** 0.81에서 건전하다고 검증된 부분을 흔들지 않기 위해, 감사가 지정한 네 곳만 고쳤다. AC 총수 15 불변.

**N3 (critical) — C3는 손으로 관리하는 미러가 아니라 생성 산출물이다.**

이 트리에서 확인한 사실:

- `internal/template/agentemit/` 가 중립 `.md` 층 + 매니페스트(`agents-codex.yaml`)로부터 `.codex/agents/moai/*.toml` 을 결정적으로 만든다.
- `Makefile` 의 `agents-emit` 타깃이 재생성 진입점이다(`AGENTEMIT_UPDATE=1` 로 골든 산출물 갱신).
- **`build:` 의 선행 타깃은 `templ-generate` 뿐이다** — `agents-emit` 을 부르지 않는다. 두 타깃은 독립이다.
- `golden_test.go` 는 손으로 고친 `.toml` 을 `committed artifact differs from emission (sha256 mismatch)` 로 떨어뜨린다.
- 골든 테스트 현재 상태 — 이 세션에서 `AGENTEMIT_UPDATE` 없이 직접 실행:

```
$ go test ./internal/template/agentemit/... -run TestGoldenCommittedArtifactsMatchEmission; echo "exit=$?"
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.402s
exit=0
```

(측정 트리 `pwd` = `.claude/worktrees/t301`, HEAD `d29b8942e`, branch `WT-full-suite-doctrine` — 같은 턴에 확인.)

세 경로 중 하나만 옳다: **C2 수정 후 재생성**(옳음) · 손 편집(골든 가드에 걸림) · 재생성 생략(다음 방출이 조용히 되돌림). v0.3.0의 15개 AC는 **내용**을 재므로 셋을 구별하지 못했다 — 되돌아가기 전까지는 내용이 옳기 때문이다.

네 곳 편집:

| # | 위치 | 변경 |
|---|---|---|
| 1 | `spec.md §A.1` | C3가 `internal/template/agentemit` 로 C2에서 생성된다는 사실, `make build` 가 `agents-emit` 을 부르지 않는다는 사실, 세 경로 중 하나만 옳다는 판정을 추가. 사본 축 구조는 그대로. |
| 2 | `spec.md §C` C7 | TOML 이스케이프 우려(생성 파일에는 해당 없음)를 **손 편집 금지 + `make agents-emit` 재생성**으로 교체. |
| 3 | `plan.md` M2 · §G | 단계 순서를 "C2 편집 → `make agents-emit` → `make build` → C1 확인" 으로 정정. §G에 손 편집 / 재생성 생략 두 경로를 안티패턴으로 추가(후자는 B7 조용한 되돌림의 생성-계보 판본). |
| 4 | `acceptance.md` AC-FSD-010 | 골든 테스트 종료 코드를 Then에 편입. `make agents-emit` 과 `make build` 를 **따로** 실행하도록 명시(빌드가 재생성을 대신한다고 가정하지 않음). 불변 보존형임과 사전 통과 사실을 함께 기재. **새 AC 없음.** |

편집 4가 plan-phase에서 이뤄져야 하는 이유는 iter-1의 D2와 같다 — `acceptance.md` 본문 수정은 run-phase 행위자의 소유권 밖이다.

**비차단 4건 처리**: N4·N5·N6·N7은 손대지 않았다. 위 네 편집 중 어느 것도 그 줄에 닿지 않았기 때문이다. 특히 **N5는 열린 채로 남는다** — `§D.2` 의 REQ-FSD-009 행이 C3 직접 측정 AC를 5종으로 적었으나 AC-FSD-007도 C3를 직접 재므로 실제로는 6종이다. 편집 4는 AC-FSD-010 본문에만 닿았고 `§D.2` 행은 건드리지 않았다. N6(같은 문장 눈 확인)은 감사가 run-phase 진입에 허용 가능하다고 판정한 갭이고, N7(어휘족 전체를 버리는 패러프레이즈)은 grep으로 원리상 닫을 수 없다.

### final 감사 — FAIL 0.83 (Δ +0.02, 네 번째 단조 상승, 임계선 2회 연속 통과)

Must-Pass 7/7, iter-4의 네 편집 전부 착지, 범위 밖 이동 0. 차단 결함 2건(두 파일 약 5줄). 기록: `.moai/reports/t301/plan-audit-final.md`.

### iteration 5 — 대응 (v0.5.0) · N8·N9 좁은 델타

**N8 (major) — AC-FSD-010의 명령 순서가 검사를 동어반복으로 만들었다.**

`internal/template/agentemit/golden_test.go` 의 update 분기를 직접 읽어 확인했다:

```go
if update {
    ... os.MkdirAll ...
    ... os.WriteFile(committedTOMLPath(p), data, 0o644) ...
    t.Logf("updated %s (sha256 %s)", p, sum[:12])
    continue          // 비교를 통째로 건너뛴다
}
```

`make agents-emit` 은 커밋 산출물을 **무조건 덮어쓰고 0으로 끝난다.** v0.4.0의 순서(`agents-emit` → `build` → 골든)에서는 손으로 편집한 C3를 명령 1이 덮어써 없애고, 재생성을 건너뛴 상태도 명령 1이 해소해 버린다. 즉 뒤에 오는 읽기 전용 골든은 **원리상 빨개질 수 없었다** — 잡으려던 두 실패 경로를 검사 직전 명령이 지우는 형태였다.

조치: 읽기 전용 골든을 **맨 앞**으로 옮겨 구현자가 남겨둔 그대로의 트리를 관측하게 했다. 아울러 "골든 테스트가 둘 다를 잡는 유일한 기계 장치" 라는 문장을 **AC-FSD-010 본문과 `plan.md` §G 양쪽에서** 고쳤다 — 실제로 RED를 낼 수 있는 것은 **선행 골든(명령 1)과 fail-closed인 `make agents-emit`(명령 2)** 이고, 마지막 자리의 골든은 검사가 아니라 사후 확인이다.

불변 보존형 분류도 이로써 복구된다. 빨개질 수 없는 명령에 귀속돼 있던 것을 빨개질 수 있는 명령 1에 다시 앵커했고, 그 사실을 AC 본문에 적었다.

**N9 (major) — `plan.md` §D에 v0.3.0판 C7이 남아 있었다.**

`spec.md` C7은 iter-4에서 손 편집 금지로 교체됐는데, `plan.md` §D의 사본은 "문면을 옮길 때 문자열 리터럴 경계를 깨지 않는다" 는 옛 문장 그대로였다. 이 문장은 손 편집을 전제하므로 교체된 C7과 정면으로 충돌한다. **구현자가 제약을 실제로 읽는 곳이 `plan.md` §D** 이므로, 따르게 되는 쪽은 낡은 사본이었다. `spec.md` C7과 같은 문면으로 교체했다.

**교훈 — 검사한 것이 아니라 검사기 쪽의 결함이다.** 이 누락은 iter-4의 자체 델타 점검이 (1) grep 범위를 `spec.md` 하나로 좁히고 (2) 본문이 한국어 `이스케이프` 인데 영어 `escap` 으로 매칭한 탓에 0을 반환했고, **그 0을 확인으로 읽었기** 때문에 생겼다. 사전 baseline 없이 0히트를 근거로 삼는 것 — 이 SPEC의 AC 규율이 존재 이유로 삼는 바로 그 오류를 점검 절차 자신이 저질렀다. 앞으로 델타 점검은 (가) 산출물 4종 전체를 범위로, (나) 대상 문서의 실제 언어로 패턴을 잡고, (다) 0히트일 때 그 패턴이 수정 전 트리에서 비영을 반환했는지부터 확인한다.

**`§D.2` 선택 항목 — 닫았다.** REQ-FSD-009 행의 C3 직접 측정 목록에 AC-FSD-007을 더해 5종 → **6종**으로 정정하고, AC-FSD-010에 "(생성 계보)" 를 달았다. 어차피 같은 파일을 편집하는 라운드였고, 열어두면 다음 독자가 목록을 사실로 읽을 위험이 있어 닫는 쪽을 택했다.

AC 총수 15 불변, AC-FSD-011 무변경.

### 잔여 미결

- M1의 문면 3종은 run-phase 산출이다. AC의 계측점이 되는 정본 문자열 4개는 이미 SPEC에 고정돼 있으므로(`spec.md §C` C6), 남은 것은 그 계측점을 품은 문장 짓기뿐이다.
- AC-FSD-007의 "두 토큰이 같은 문장에" 는 여전히 눈 확인이다. 기계 층이 보증하는 것은 "둘 다, STEP 5 블록 안에" 까지이며, 이 갭을 AC 본문과 §D.3에 명시했다.
- AC-FSD-011(회귀 관측)은 관측 창이 열릴 때까지 미결이다.

## §E.2 Run-phase Evidence

측정 트리: 워크트리 `.claude/worktrees/t301`, 브랜치 `WT-full-suite-doctrine`.

### 기준 SHA 갱신 (AC-FSD-009 · `plan.md §C` 2번)

`acceptance.md` 가 기재한 기준 SHA `d29b8942e` 는 이 브랜치의 분기점이 아니다. 이 워크트리는 `origin/develop` 으로 fast-forward 된 상태에서 시작했고, 실측 분기점은 **`7ed6edb3e`** 다. `plan.md §C` 2번이 지시한 대로 실제 분기점으로 갱신해 AC-FSD-009를 실행했다.

```
$ git rev-parse --short HEAD
7ed6edb3e
$ git branch --show-current
WT-full-suite-doctrine
```

### 편집 전 baseline 재채집 (전부 이 트리, HEAD `7ed6edb3e`)

`acceptance.md` 의 15개 사전 baseline을 편집 전에 전부 재실행했고 **기재값과 한 건도 어긋나지 않았다.** 따라서 `plan.md §C` 4번의 "불일치는 SPEC 재작업 사유" 조건은 발동하지 않았다.

| AC | baseline 기재값 | 실측 |
|---|---|---|
| AC-FSD-001 | C1 `3` / C2 `3` / C3 `3` | 동일 |
| AC-FSD-002 | C1 `4` / C2 `4` / C3 `4` | 동일 |
| AC-FSD-003 | 각 `1` | 동일 |
| AC-FSD-004 | 각 `1` | 동일 |
| AC-FSD-005 | 각 `1` | 동일 |
| AC-FSD-006 | 구간 `4`줄 · 호출 각 `1` | 동일 |
| AC-FSD-007 | 구간 `10`줄 · 두 토큰 `0` | 동일 |
| AC-FSD-008 | `1a2 > isolation: worktree` · 무출력 · 무출력 | 동일 |
| AC-FSD-012 | 구간 `14`줄(세 사본) · 정본 문구 `0` | 동일 |
| AC-FSD-013 | 각 `0` | 동일 |
| AC-FSD-014 | C1 `3` / C2 `3` / C3 `3` | 동일 |
| AC-FSD-015 | 각 `1` | 동일 |

### 대체 문면 (M1 산출 — 세 사본에 같은 문면)

1. **범위 규칙(S2·S3)** — `run the tests the change can affect` (정본 문구, `REQ-FSD-001`).
2. **배치 열거(S4)** — `full suite, coverage, …` → `change-scoped tests, coverage, …`.
3. **위임·미결(STEP 5)** — `The report MUST name the CI run on the project's integration branch as the owner of the repository-wide test verdict and state that this verdict is PENDING at report time.`
4. **판별자(S1)** — `LARGE_SCALE` 를 정의하던 STEP 1 항목 자체를 제거. 그 항목의 유일한 내용이 판별자 정의와 그 귀결이었으므로, 토큰만 지우면 뜻이 남지 않는 문장이 된다.

> **어휘 선택 근거.** `REQ-FSD-006` 산문은 "full-suite verdict" 라고 적지만, 그 표현을 그대로 쓰면 `AC-FSD-002` 의 어휘족 정규식 `(full|complete|entire|whole)[ -](test )?suite` 에 `full-suite` 가 걸려 부재형 AC가 RED가 된다. 정본 토큰은 `integration branch` 와 `PENDING at report time` 둘뿐이므로(`spec.md §C` C6), 나머지 서술은 어휘족을 피해 `repository-wide test verdict` 로 썼다. 두 정본 토큰은 문자 그대로 유지했다.

### 편집 순서 (Template-First + C7)

`C2(템플릿 md) → C1(로컬 md) → 배치 정의 4벌 → make agents-emit(C3 재생성) → make build`.

**C3는 손으로 편집하지 않았다.** 재생성 직전 읽기 전용 골든을 돌려 가드가 살아 있음을 관측했다:

```
$ go test ./internal/template/agentemit/... -run TestGoldenCommittedArtifactsMatchEmission
--- FAIL: TestGoldenCommittedArtifactsMatchEmission (0.00s)
    golden_test.go:109: .codex/agents/moai/manager-develop.toml: committed artifact differs from emission (sha256 mismatch) — regenerate or stop hand-editing
FAIL	github.com/modu-ai/moai-adk/internal/template/agentemit	0.413s
exit=1
```

즉 `plan.md §G` 가 이름 붙인 "C2를 고치고 재생성을 건너뛴 상태" 를 이 가드가 실제로 빨갛게 잡는다는 것을 관측했다 — 추정이 아니다. 이어서 `make agents-emit` 을 돌려 C3를 재생성했고, 네 지점이 실제로 바뀌었음을 diff로 확인했다:

```
$ git diff --stat -- internal/template/templates/.codex/
 .../template/templates/.codex/agents/moai/manager-develop.toml   | 9 ++++-----
 1 file changed, 4 insertions(+), 5 deletions(-)
```

diff 본문에서 S1(삭제) · S2 · S3 · S4 네 지점이 모두 C2와 같은 문면으로 바뀌었다.

### 부수 변경 — `internal/template/catalog.yaml`

`make build` 가 `gen-catalog-hashes --all` 로 `manager-develop.md` 의 SHA256 항목 1줄을 갱신했다(`7872f240… → ad390185…`). 이 SPEC이 편집한 파일의 해시이므로 같은 SPEC 범위 안의 cascade이며, 커밋에 포함한다.

### AC PASS/FAIL 매트릭스

| AC | 판정 | 실행 명령 | 실측 출력 |
|---|---|---|---|
| AC-FSD-001 | PASS | `grep -c -e 'always runs the full suite' -e 'otherwise the full suite' -e 'COMPLETE test suite' -e 'regardless of LARGE_SCALE' <C1> <C2> <C3>` | `C1:0` / `C2:0` / `C3:0` |
| AC-FSD-002 | PASS | `grep -c -i -E '(full\|complete\|entire\|whole)[ -](test )?suite' <C1> <C2> <C3>` | `C1:0` / `C2:0` / `C3:0` |
| AC-FSD-003 | PASS | `grep -c 'full suite, coverage' <C1> <C2> <C3>` | `C1:0` / `C2:0` / `C3:0` |
| AC-FSD-004 | PASS | `grep -c 'Full test suite' <ACPR> <ACPR_T>` | `0` / `0` |
| AC-FSD-005 | PASS | `grep 'A. Functional' <VBP> \| grep -c 'go test \./\.\.\.'` (로컬·템플릿 각각) | `0` / `0` |
| AC-FSD-006 | PASS | `awk '/^# 1\./,/^# 2\./' <ACPR> \| wc -l` 및 `\| grep -c 'go test \./\.\.\.'` (로컬·템플릿) | 구간 `4` · `0` / `0` |
| AC-FSD-007 | PASS | `awk '/^### STEP 5/,/^### Checkpoint/' <사본> \| wc -l` 및 토큰별 `grep -c` ×6 | 구간 `10` · 나머지 여섯 출력 모두 `1` |
| AC-FSD-008 | PASS | `diff` ×3 (로컬↔템플릿) | `1a2 > isolation: worktree` (rc=1) · 무출력(rc=0) · 무출력(rc=0) |
| AC-FSD-009 | PASS | `git diff 7ed6edb3e -- internal/template/templates/` → `grep -c '^+'` → 금칙 패턴 스캔 | 훑은 줄 `15` · 금칙 스캔 무출력 `rc=1` |
| AC-FSD-010 | PASS | 골든(읽기 전용) → `make agents-emit` → `make build` (이 순서) | `exit=0` / `exit=0` / `exit=0`; 골든 출력 `ok github.com/modu-ai/moai-adk/internal/template/agentemit 0.419s`; 재실행한 AC-FSD-008 여전히 통과 |
| AC-FSD-011 | **미결(지연)** | — | 관측 창 필요. 통과로 적지 않는다 |
| AC-FSD-012 | PASS | `awk '/^### STEP 4/,/^### STEP 5/' <사본> \| wc -l` 및 `\| grep -c 'the tests the change can affect'` | 구간 `14` · 세 사본 각 `1` |
| AC-FSD-013 | PASS | `awk '/^# 1\./,/^# 2\./' <ACPR> \| grep -c 'internal/<pkg>'` (로컬·템플릿) | `1` / `1` |
| AC-FSD-014 | PASS | `grep -c 'LARGE_SCALE' <C1> <C2> <C3>` | `C1:0` / `C2:0` / `C3:0` |
| AC-FSD-015 | PASS | `grep 'A. Functional' <VBP> \| grep -c -E '[0-9]+-[0-9]+ s'` (로컬·템플릿) | `0` / `0` |

MUST 14개 전부 PASS. AC-FSD-011은 SHOULD(지연)이며 미결로 남는다.

### AC-FSD-007 눈 확인 (기계 층 밖의 유일한 잔여 판정)

`§D.3` 이 요구한 "두 정본 토큰이 **같은 문장**에" 를 육안으로 확인했다. 세 사본 모두 STEP 5의 완료 보고 항목에 한 문장으로 들어 있다:

> The report MUST name the CI run on the project's **integration branch** as the owner of the repository-wide test verdict and state that this verdict is **PENDING at report time**.

주절 하나에 `and` 로 이어진 단일 문장이며, 마침표는 문장 끝에 한 번만 나온다. 위임과 미결이 갈라진 두 문장이 아니다.

### 영향 패키지 테스트

전량 스위트는 돌리지 않았다 — 이 SPEC이 금지 대상으로 다루는 바로 그 행위이며 `spec.md §C` C4가 막는다. 변경이 영향 줄 수 있는 패키지만 돌렸다:

```
$ go test ./internal/template/...
ok  	github.com/modu-ai/moai-adk/internal/template	23.418s
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.654s
?   	github.com/modu-ai/moai-adk/internal/template/scripts	[no test files]
exit=0
```

여기에 미러 패리티 가드(`rule_template_mirror_test.go`)와 방출 골든이 포함된다. 리포 전량 판정은 push 이후 통합 브랜치 CI 몫이며 **보고 시점에는 미결**이다.

### 증거 경로

`.moai/state/verify/t301/` — `golden-pre-emit.log`(재생성 전 RED), `ac010-1-golden.log`, `ac010-2-emit.log`, `ac010-3-build.log`, `template-pkg-tests.log`, `t301-neutrality.diff`.

### 갭

- **AC-FSD-011** — 사후 관측이 없다. 관측 창이 열릴 때까지 미결.
- **전량 스위트 판정** — 로컬에서 관측하지 않았다(C4 금지). 통합 브랜치 CI가 판정 주체이고 보고 시점에는 미결이다.
- **기준 SHA** — `acceptance.md` 기재값(`d29b8942e`)이 아니라 실측 분기점(`7ed6edb3e`)으로 판정했다. `acceptance.md` 는 plan-phase 산출이라 run-phase에서 고치지 않는다.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-27
run_commit_sha: pending-backfill-t301
run_status: audit-ready
ac_pass_count: 14
ac_fail_count: 0
ac_deferred_count: 1
preserve_list_post_run_count: 0
new_warnings_or_lints_introduced: 0
cross_platform_build:
  status: not-applicable
  reason: "markdown·toml 문면 편집뿐이며 Go 코드 변경이 없다(spec.md §C C1)."
total_run_phase_files: 8
m1_to_mN_commit_strategy: single-commit
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
