# SPEC-FULL-SUITE-DOCTRINE-001 — 인수 기준

버전 0.3.0. 모든 명령은 리포 루트(작업 워크트리) 기준이며, 각 AC는 **사전 baseline**(구현 전 트리에서 실측한 값)과 **사후 기대치**를 함께 싣는다. baseline이 기재값과 다르면 그 AC는 통과 대상이 아니라 재작성 대상이다.

이 트리의 사전 baseline은 전부 HEAD `d29b8942e` 에서 실행해 얻었다. AC 총수 15개(Tier M 상한 16).

**대상 사본 3벌** — `spec.md §A.1` 참조. 표기 약칭:

- `C1` = `.claude/agents/moai/manager-develop.md` (로컬 md)
- `C2` = `internal/template/templates/.claude/agents/moai/manager-develop.md` (템플릿 md, 배포)
- `C3` = `internal/template/templates/.codex/agents/moai/manager-develop.toml` (템플릿 toml, 배포, **로컬 쌍 없음**)
- `ACPR` / `ACPR_T` = `agent-common-protocol-reference.md` 의 로컬 / 템플릿 사본
- `VBP` / `VBP_T` = `verification-batch-pattern.md` 의 로컬 / 템플릿 사본

AC는 세 갈래다. **부재형**(사라졌는가), **존재형**(새로 생겼는가), **동등형**(두 사본이 같은가). 부재형만으로는 동의어 재작성 mutant를 잡지 못하므로 존재형이 짝을 이루고, 존재형이 파일 단위면 지점 단위 mutant를 놓치므로 **구간 고정**이 짝을 이룬다.

---

## §D. AC 매트릭스

### 부재형 — 전량 지시가 사라졌는가

#### AC-FSD-001 — 리터럴 4패턴 소멸 (세 사본 전부)

**Given** 세 사본이 각각 네 지점에서 전량 실행을 지시하고 있고,
**When** 수리 후 다음을 실행하면,

```bash
grep -c -e 'always runs the full suite' -e 'otherwise the full suite' \
        -e 'COMPLETE test suite' -e 'regardless of LARGE_SCALE' \
        .claude/agents/moai/manager-develop.md \
        internal/template/templates/.claude/agents/moai/manager-develop.md \
        internal/template/templates/.codex/agents/moai/manager-develop.toml
```

**Then** 세 줄 모두 `:0` 이다.

- 사전 baseline: C1 `3` / C2 `3` / C3 `3` (전부 실측). C3가 `.md` 사본들과 같은 값이라는 사실을 가정하지 않고 따로 쟀다.

#### AC-FSD-002 — 전량 어휘족 소멸 (동의어 재작성 차단)

**Given** 리터럴 4패턴만 막으면 형용사를 갈아끼운 재작성(`complete suite`, `entire test suite` 등)이 그대로 통과하고,
**When** 수리 후 어휘족 패턴을 대소문자 무시로 세면,

```bash
grep -c -i -E '(full|complete|entire|whole)[ -](test )?suite' \
        .claude/agents/moai/manager-develop.md \
        internal/template/templates/.claude/agents/moai/manager-develop.md \
        internal/template/templates/.codex/agents/moai/manager-develop.toml
```

**Then** 세 줄 모두 `:0` 이다.

- 사전 baseline: C1 `4` / C2 `4` / C3 `4` (전부 실측). 리터럴 집합의 `3` 보다 1 큰 것은 S4의 `full suite, coverage` 까지 잡기 때문이다.
- 이전 판(v0.2.0)에서 AC-FSD-002는 템플릿판에 같은 리터럴 검사를 반복하는 항목이었다. 그 역할은 AC-FSD-001의 인자 목록으로 흡수했고, 이 번호는 어휘족 판정으로 재배정했다 — AC 총수를 늘리지 않으면서 mutant 공간을 닫기 위해서다.

#### AC-FSD-003 — 배치 호출 문장(S4)의 전량 열거 소멸

**Given** S4가 배치 항목을 `full suite, coverage, lint, boundary greps` 로 열거하고 있고,
**When** 다음을 실행하면,

```bash
grep -c 'full suite, coverage' .claude/agents/moai/manager-develop.md \
  internal/template/templates/.claude/agents/moai/manager-develop.md \
  internal/template/templates/.codex/agents/moai/manager-develop.toml
```

**Then** 세 줄 모두 `:0` 이다.

- 사전 baseline: C1 `1` / C2 `1` / C3 `1` (전부 실측).

#### AC-FSD-004 — 배치 1번 항목 표제의 전량 명명 소멸

**Given** 배치 정의 1번 항목 주석이 `Full test suite (Go)` 이고,
**When** 다음을 실행하면,

```bash
grep -c 'Full test suite' .claude/rules/moai/core/agent-common-protocol-reference.md \
  internal/template/templates/.claude/rules/moai/core/agent-common-protocol-reference.md
```

**Then** 두 파일 모두 `0` 이다.

- 사전 baseline: 각 `1` (실측).
- 이 파일에는 codex 미러가 없다 — `.codex/` 트리는 `agents/moai/*.toml` 만 담고 `rules` 디렉터리 자체가 없다(실측).

#### AC-FSD-005 — Group A 행의 무조건 전량 호출 소멸 (자기완결 판정)

**Given** Group A 행이 무조건 전량 호출을 구성원으로 적고 있고,
**When** 해당 행만 추출해 그 안에서 세면,

```bash
grep 'A. Functional' .claude/rules/moai/workflow/verification-batch-pattern.md | grep -c 'go test \./\.\.\.'
grep 'A. Functional' internal/template/templates/.claude/rules/moai/workflow/verification-batch-pattern.md | grep -c 'go test \./\.\.\.'
```

**Then** 두 출력 모두 `0` 이다.

- 사전 baseline: 각 `1` (실측).

#### AC-FSD-006 — 배치 1번 항목 구간의 무조건 전량 호출 소멸 (구간 한정)

**Given** 배치 코드블록 1번 항목이 무조건 전량 호출을 담고 있고,
**When** `# 1.` ~ `# 2.` 구간만 잘라내 그 안에서 세면,

```bash
awk '/^# 1\./,/^# 2\./' .claude/rules/moai/core/agent-common-protocol-reference.md | wc -l
awk '/^# 1\./,/^# 2\./' .claude/rules/moai/core/agent-common-protocol-reference.md | grep -c 'go test \./\.\.\.'
awk '/^# 1\./,/^# 2\./' internal/template/templates/.claude/rules/moai/core/agent-common-protocol-reference.md | grep -c 'go test \./\.\.\.'
```

**Then** 첫 출력이 `4`(구간이 닫혀 있음)이고, 나머지 두 출력이 모두 `0` 이다.

- 사전 baseline: 구간 `4`줄, 전량 호출 각 `1` (실측).
- **구간 한정이 확정 판정이다 — run-phase로 미루지 않는다.** 파일 전체로 세면 `3` 이 나오지만 나머지 2히트(`:65`, `:76`)는 처방이 아니라 예시 산문이다: 하나는 "직렬 검증 안티패턴" 블록에서 하지 말라는 형태를 보여주고, 다른 하나는 "언제 직렬로 실행하는가" 의 의존 관계 예시다. 이 SPEC의 어떤 변경으로도 뒤집히지 않으므로 전체 카운트를 요구하면 도착 시점에도 완료 시점에도 RED인 항목이 된다. `spec.md §A.3`·§D 가 두 지점을 명시적 범위 밖으로 선언한다.
- 커버리지 손실 없음: 실제 처방인 `:22` 의 호출은 이 구간 안에 있다.
- 구간 폭주는 이 AC 자신이 잡는다 — 종료 패턴이 깨져 범위가 EOF까지 달아나면 예시 산문이 딸려 들어와 카운트가 `2` 가 되고, 줄 수도 `4` 를 벗어난다.

#### AC-FSD-014 — 판별자 제거로 dangling conditional 원천 차단

**Given** `LARGE_SCALE` 가 세 사본 각각에 정확히 3회 등장하고 그 3회가 전부 위반 지점과 같은 줄이며, 그 유일한 귀결이 "타깃 실행으로 전환" 이고,
**When** 새 독트린에서 타깃 실행이 무조건이 되어 판별자가 아무것도 가르지 않게 된 뒤 다음을 실행하면,

```bash
grep -c 'LARGE_SCALE' .claude/agents/moai/manager-develop.md \
  internal/template/templates/.claude/agents/moai/manager-develop.md \
  internal/template/templates/.codex/agents/moai/manager-develop.toml
```

**Then** 세 줄 모두 `:0` 이다.

- 사전 baseline: C1 `3` / C2 `3` / C3 `3` (전부 실측).
- 이것이 REQ-FSD-003의 기계 판정이다. `otherwise` 만 지우고 판별자를 남기면 문법적 dangling은 사라지되 실질은 남는데, 토큰이 0이면 매달릴 가지 자체가 없다.

#### AC-FSD-015 — Group A 시간 추정치에서 미측정 수치 제거

**Given** Group A 행이 `30-120 s` 라는, 이 SPEC이 측정하지 않은 지속 시간 범위를 주장하고 있고,
**When** 해당 행만 추출해 지속 시간 범위 패턴을 세면,

```bash
grep 'A. Functional' .claude/rules/moai/workflow/verification-batch-pattern.md | grep -c -E '[0-9]+-[0-9]+ s'
grep 'A. Functional' internal/template/templates/.claude/rules/moai/workflow/verification-batch-pattern.md | grep -c -E '[0-9]+-[0-9]+ s'
```

**Then** 두 출력 모두 `0` 이다.

- 사전 baseline: 각 `1` (실측).
- REQ-FSD-005의 두 번째 절을 재는 유일한 항목이다. AC-FSD-005는 호출 형태만 보고 추정치를 보지 않는다.

---

### 존재형 — 대체 문면이 실제로, 올바른 자리에 생겼는가

부재형 AC는 문자열 회피에 무력하고, 파일 단위 존재형 AC는 **지점 단위** 회피에 무력하다. 후자가 실제 mutant다: 위반 지점은 동의어로 남겨두고 정본 문구를 엉뚱한 다른 지점에 한 번 심으면 파일 단위 판정이 초록이 된다. 그래서 아래 존재형 항목은 전부 **구간 고정**이며, 구간 폭주를 막기 위해 추출 줄 수를 **Then 안에** 넣는다.

#### AC-FSD-012 — 변경 범위 스코프 문면이 STEP 4 블록 안에 존재 (REQ-FSD-001)

**Given** 실행을 실제로 지배하는 자리는 STEP 4 변경 루프이고, 파일 어딘가에 문구가 있다는 사실만으로는 그 자리가 고쳐졌음을 증명하지 못하며,
**When** 세 사본 각각에서 STEP 4 블록만 잘라내 정본 문구를 세면,

```bash
awk '/^### STEP 4/,/^### STEP 5/' .claude/agents/moai/manager-develop.md | wc -l
awk '/^### STEP 4/,/^### STEP 5/' .claude/agents/moai/manager-develop.md | grep -c 'the tests the change can affect'
awk '/^### STEP 4/,/^### STEP 5/' internal/template/templates/.claude/agents/moai/manager-develop.md | grep -c 'the tests the change can affect'
awk '/^### STEP 4/,/^### STEP 5/' internal/template/templates/.codex/agents/moai/manager-develop.toml | grep -c 'the tests the change can affect'
```

**Then** 첫 출력이 `14`(구간이 닫혀 있음)이고, 나머지 세 출력이 모두 `1` 이상이다.

- 사전 baseline: 구간 길이 C1 `14` / C2 `14` / C3 `14`(세 사본 동일, 실측), 정본 문구 세 사본 모두 `0` (실측) — RED-now 확보.
- 이 문구는 상위 계약(`AGENTS.md`)이 이미 쓰는 언어 중립 산문이므로 배포 중립성과 충돌하지 않는다.

#### AC-FSD-013 — 배치 1번 항목의 패키지 스코프 호출 존재 (REQ-FSD-004)

**Given** 배치 1번 항목이 리포 전체 호출만 담고 있고,
**When** 수리 후 같은 구간에서 패키지 자리표시자를 세면,

```bash
awk '/^# 1\./,/^# 2\./' .claude/rules/moai/core/agent-common-protocol-reference.md | grep -c 'internal/<pkg>'
awk '/^# 1\./,/^# 2\./' internal/template/templates/.claude/rules/moai/core/agent-common-protocol-reference.md | grep -c 'internal/<pkg>'
```

**Then** 두 출력 모두 `1` 이상이다.

- 사전 baseline: 각 `0` (실측). 같은 자리표시자가 2번 항목에는 이미 있으나 그것은 구간 밖이라 계수되지 않는다 — 구간 한정이 baseline 0을 지켜준다. 구간이 폭주하면 AC-FSD-006이 먼저 RED가 되므로 이 전제는 조용히 무너질 수 없다.

#### AC-FSD-007 — 위임·미결 문장이 STEP 5 블록 안에 존재 (토큰별 판정)

**Given** 전량 판정을 삭제하면 완료 보고가 근거 없이 통과로 읽히고, 미결 상태를 말하지 않는 위임 문장도 같은 방식으로 통과로 읽히며,
**When** 세 사본 각각에서 STEP 5 블록만 잘라내 두 정본 토큰을 **각각** 세면,

```bash
awk '/^### STEP 5/,/^### Checkpoint/' .claude/agents/moai/manager-develop.md | wc -l
awk '/^### STEP 5/,/^### Checkpoint/' .claude/agents/moai/manager-develop.md | grep -c 'integration branch'
awk '/^### STEP 5/,/^### Checkpoint/' .claude/agents/moai/manager-develop.md | grep -c 'PENDING at report time'
awk '/^### STEP 5/,/^### Checkpoint/' internal/template/templates/.claude/agents/moai/manager-develop.md | grep -c 'integration branch'
awk '/^### STEP 5/,/^### Checkpoint/' internal/template/templates/.claude/agents/moai/manager-develop.md | grep -c 'PENDING at report time'
awk '/^### STEP 5/,/^### Checkpoint/' internal/template/templates/.codex/agents/moai/manager-develop.toml | grep -c 'integration branch'
awk '/^### STEP 5/,/^### Checkpoint/' internal/template/templates/.codex/agents/moai/manager-develop.toml | grep -c 'PENDING at report time'
```

**Then** 첫 출력이 `10`(구간이 닫혀 있음)이고, 나머지 여섯 출력이 **모두** `1` 이상이다.

- 사전 baseline: 구간 길이 C1 `10` / C2 `10` / C3 `10`(세 사본 동일, 실측 — awk 범위 연산자가 종료 패턴 줄을 포함하므로 실질 본문 9줄 + 종료 헤딩 1줄), 두 토큰 세 사본 모두 `0` (실측).
- **토큰을 각각 센다.** 이전 판(v0.2.0)은 `grep -c -e A -e B` 였는데, 이 형태는 두 대안 중 **하나만** 있어도 매치 줄을 세므로 `PENDING at report time` 을 빠뜨린 문면이 초록으로 통과했다. 미결을 말하지 않는 위임 문장은 보고 시점에 통과로 읽히므로, 그 구멍은 B5(조용한 삭제)의 재발 경로다.
- **구간 폭주를 Then이 막는다.** 종료 헤딩이 편집으로 바뀌어 범위가 EOF까지 달아나면 줄 수가 `10` 을 벗어나 RED가 된다. 위치 고정이 이 AC의 존재 이유이므로 그 고정 자체가 판정에 들어가야 한다.
- 두 토큰이 **같은 문장**에 있어야 한다는 REQ-FSD-006의 요구는 여전히 눈 확인이다. 기계 층이 보증하는 것은 "둘 다, STEP 5 블록 안에" 까지다 — 남은 갭으로 기록한다.
- **브랜치명을 쓰지 않는다.** `integration branch` 는 일반 명칭이다. 특정 브랜치명은 AC-FSD-009가 차단한다.

---

### 동등형 · 중립성

#### AC-FSD-008 — 로컬·템플릿 델타 보존

**Given** 수리 전 세 쌍의 델타가 각각 1줄 / 0줄 / 0줄이고,
**When** 수리 후 다음을 실행하면,

```bash
diff .claude/agents/moai/manager-develop.md \
     internal/template/templates/.claude/agents/moai/manager-develop.md
diff .claude/rules/moai/core/agent-common-protocol-reference.md \
     internal/template/templates/.claude/rules/moai/core/agent-common-protocol-reference.md
diff .claude/rules/moai/workflow/verification-batch-pattern.md \
     internal/template/templates/.claude/rules/moai/workflow/verification-batch-pattern.md
```

**Then** 첫 diff는 `isolation: worktree` 1줄만 출력하고, 나머지 두 diff는 무출력(`rc=0`)이다.

- 사전 baseline: 실측 — `1a2 > isolation: worktree`, 무출력, 무출력.
- **C3는 이 항목의 대상이 아니다.** 로컬 쌍이 존재하지 않으므로(실측: `.codex/agents/moai/manager-develop.toml` 없음) 쌍 동등성을 잴 대상이 없다. C3는 AC-FSD-001·002·003·012·014가 **직접** 측정하므로 판정 공백이 생기지 않는다.

#### AC-FSD-009 — 템플릿 중립성 (커밋 상태와 무관한 판정)

**Given** 세 배포 파일이 하류 프로젝트로 나가고, 이 SPEC이 도입할 위험이 있는 지역 내용에는 **브랜치명도 포함**되며(`origin/develop` 은 현재 배포 템플릿 전체에 0회 등장한다 — 파이프 없이 실측, `rc=1`),
**When** 구현 시작 시점 SHA를 기준으로 추가된 줄에 패턴을 걸면,

```bash
git diff d29b8942e -- internal/template/templates/ > /tmp/t301-neutrality.diff; echo "exit=$?"
grep -c '^+' /tmp/t301-neutrality.diff
grep -i -e 'SPEC-' -e 'CLAUDE.local' -e '2026-' -e 'load 413' -e '/Users/' \
        -e 'origin/develop' -e 'origin/main' /tmp/t301-neutrality.diff | grep '^+'
echo "rc=$?"
```

**Then** 두 번째 명령의 출력이 `1` 이상이고(훑은 줄이 실제로 존재함), 세 번째 명령이 무출력이며 `rc=1` 이다.

- 사전 baseline: 아직 편집이 없으므로 diff가 비어 있고 훑은 줄 수는 `0` — **즉 사전에는 이 AC가 성립하지 않는다.** 이것이 의도한 형태다: 훑은 줄 수 하한이 "아무것도 안 훑은 스윕"을 통과로 읽지 못하게 막는다.
- **기준을 `d29b8942e` 로 고정**하므로 staged / committed 여부와 무관하게 판정된다.
- 경로가 `internal/template/templates/` 전체이므로 `.codex/` 하위 변경도 자동으로 훑는다 — C3 편입으로 이 항목은 수정이 필요 없다.
- 브랜치를 리베이스했다면 기준 SHA를 실제 분기점으로 갱신하고 그 값을 `progress.md` 에 기록한다.

#### AC-FSD-010 — 임베드 재생성 + C3 생성 계보 무결성

**Given** 템플릿이 `//go:embed all:templates` 로 바이너리에 들어가고 그 범위에 `.codex/` 도 포함되며, C3는 손으로 관리하는 사본이 아니라 `internal/template/agentemit` 이 C2로부터 만들어내는 **생성 산출물**이고,
**When** **먼저 읽기 전용 골든 가드로 트리를 있는 그대로 관측한 뒤** 재생성과 빌드를 각각 실행하면,

```bash
go test ./internal/template/agentemit/... -run TestGoldenCommittedArtifactsMatchEmission; echo "exit=$?"
make agents-emit; echo "exit=$?"
make build; echo "exit=$?"
```

**Then** 세 `exit=` 가 모두 `0` 이고, 재실행한 AC-FSD-008이 여전히 통과한다.

- **순서가 이 AC의 전부다.** 골든 테스트의 update 분기는 비교를 건너뛰고(`continue`) 커밋 산출물을 무조건 덮어쓴 뒤 0으로 끝난다. 따라서 `make agents-emit` 을 **먼저** 돌리면 손으로 편집한 C3는 그 명령이 덮어써 없애고, 재생성을 건너뛴 상태는 그 명령이 해소해 버린다 — 뒤이어 도는 읽기 전용 골든은 **원리상 빨개질 수 없다.** 잡으려던 두 실패 경로를 검사 직전 명령이 지워버리는 형태다.
- 그래서 **명령 1이 맨 앞**이다. `AGENTEMIT_UPDATE` **없이** 도는 이 판정만이 구현자가 남겨둔 그대로의 트리를 관측한다. 커밋된 C3가 현재 방출 결과와 다르면 `committed artifact differs from emission (sha256 mismatch)` 로 떨어지며, 이것이 **손으로 편집한 C3**와 **재생성을 건너뛴 C3** 둘 다를 잡는다 — 나머지 14개 AC는 내용을 재므로, 다음 방출이 되돌리기 전까지 두 경로 모두 초록으로 보인다.
- 명령 2도 빨개질 수 있는 판정이다. 방출기는 fail-closed라, 방출 자체가 실패하면 `make agents-emit` 이 비영으로 끝난다. 즉 이 AC에서 실제로 RED를 낼 수 있는 것은 **명령 1(선행 골든)과 명령 2(fail-closed 방출)** 이고, 마지막 자리의 골든은 검사가 아니라 사후 확인이다.
- **`make build` 는 `agents-emit` 을 부르지 않는다** — 두 타깃은 독립이며 `build:` 의 선행 타깃은 `templ-generate` 뿐이다(실측). 따라서 이 AC는 빌드가 재생성을 대신해 준다고 가정하지 않고 세 명령을 따로 실행한다.
- 사전 baseline: 이 트리에서 명령 1이 통과한다(커밋된 C3가 현재 방출 결과와 일치). 즉 이 AC는 수리 전후 모두 통과해야 하는 **불변 보존형**이며, 그 분류는 빨개질 수 있는 명령 1에 귀속된다. 잡는 것은 "수리가 계보를 깨뜨렸는가" 다.

---

### 지연 항목

#### AC-FSD-011 — 회귀 관측 (이 SPEC 안에서 닫히지 않음)

**Given** 수정 전 최종 배치 소요로 **49분 40초**가 물려받은 값으로 존재하고,
**When** 수리 착지 후 실제 run-phase 1회의 STEP 5 배치가 실행되면,
**Then** 그 배치의 실측 소요와 실행 명령을 기록하고, 물려받은 값과 나란히 제시한다.

- **귀속 명시**: 49분 40초는 리드에게서 물려받은 주장이며 이 SPEC의 어떤 세션에서도 재측정하지 않았다. 재측정하려면 로컬 전량 스위트를 돌려야 하는데 그것이 `spec.md §C` C4가 금지한 행위다.
- **갭**: 따라서 이 항목은 사전/사후 **대조 실험이 아니다.** 사후 값만 실측이며 사전 값은 인용이다. 개선폭을 정량 주장으로 내세우지 않는다.

---

## §D.1 심각도

| AC | 유형 | 심각도 | 근거 |
|---|---|---|---|
| AC-FSD-001 | 부재(리터럴) | MUST | 부분 수리 탐지 — 세 사본 × 네 지점 |
| AC-FSD-002 | 부재(어휘족) | MUST | 동의어 재작성 차단 |
| AC-FSD-003·004 | 부재 | MUST | 단일 패턴이 못 잡는 지점의 전용 판정 |
| AC-FSD-005·006 | 부재(구간·행) | MUST | 무효 수리 방어 — 배치 정의가 전량을 되살리지 못하게 |
| AC-FSD-012·013 | 존재(구간 고정) | MUST | 지점 단위 mutant 방어 |
| AC-FSD-014 | 부재 | MUST | dangling conditional 원천 차단 (REQ-FSD-003) |
| AC-FSD-015 | 부재(행 한정) | MUST | 미측정 수치 주장 제거 (REQ-FSD-005 후반절) |
| AC-FSD-007 | 존재(구간 + 토큰별) | MUST | 관측되지 않은 완료 주장 방어 — VCI §1.1 surface 2 |
| AC-FSD-008·010 | 동등 | MUST | 다음 `moai update` 회귀 방어 |
| AC-FSD-009 | 중립성 | MUST | 배포 중립성 + 브랜치명 유출 차단 |
| AC-FSD-011 | 관측 | SHOULD (지연) | 관측 창 필요 |

## §D.2 추적성

| REQ | 이를 재는 AC | 유형 |
|---|---|---|
| REQ-FSD-001 | AC-FSD-012 (STEP 4 구간 고정) | 존재 |
| REQ-FSD-002 | AC-FSD-001(리터럴) + AC-FSD-002(어휘족) + AC-FSD-003·004 | 부재 |
| REQ-FSD-003 | AC-FSD-014 | 부재 |
| REQ-FSD-004 | AC-FSD-006(부재) + AC-FSD-013(존재) | 양방향 |
| REQ-FSD-005 | AC-FSD-005(호출) + AC-FSD-015(추정치) | 절별 분리 |
| REQ-FSD-006 | AC-FSD-007 — 두 토큰을 각각 판정 | 존재(위치 고정) |
| REQ-FSD-007 | AC-FSD-009 | 중립성 |
| REQ-FSD-008 | AC-FSD-007 — `PENDING at report time` 토큰 단독 판정이 "미결 생략" 을 구별한다 | 존재 |
| REQ-FSD-009 | AC-FSD-008(쌍 있는 사본) + AC-FSD-001·002·003·007·012·014(쌍 없는 C3 직접 측정 6종) + AC-FSD-010(생성 계보) | 동등·직접 |
| REQ-FSD-010 | AC-FSD-009 | 중립성 |
| REQ-FSD-011 | AC-FSD-011 | 관측(지연) |

공허한 매핑이 없도록 각 REQ의 **모든 절**에 그 절을 실제로 재는 항목을 붙였다. REQ-FSD-004·005는 절이 둘이므로 AC도 둘이고, REQ-FSD-006·008은 AC-FSD-007의 서로 다른 토큰 판정이 담당한다.

## §D.3 완료 정의

- MUST 항목 전부 통과, verbatim 출력이 `progress.md` §E.2에 기록됨.
- AC-FSD-009의 기준 SHA가 실제 분기점과 일치함이 기록됨.
- AC-FSD-007의 "두 토큰이 같은 문장" 눈 확인 결과가 기록됨(기계 층 밖의 유일한 잔여 판정).
- AC-FSD-011은 미결로 명시된 채 남는다 — 미결을 통과로 적지 않는다.
- 커밋은 명시 pathspec만. 스윕 스테이징 금지.
