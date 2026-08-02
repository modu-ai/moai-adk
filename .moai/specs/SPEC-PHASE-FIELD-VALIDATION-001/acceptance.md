---
id: SPEC-PHASE-FIELD-VALIDATION-001
title: "phase 프론트매터 필드의 값-형태 검증과 오염 코퍼스 교정 — 수용 기준"
version: "0.2.0"
status: in-progress
created: 2026-08-02
updated: 2026-08-02
author: Goos Kim
priority: P2
phase: "v3.0.2"
module: "internal/spec, .claude/agents/moai"
lifecycle: spec-anchored
tier: M
tags: "spec-lint, frontmatter, phase, drift, authoring-guard"
---

# 수용 기준

---

## A. 판정 규약

### A.1 공유 추출기

여러 AC가 프론트매터의 `phase` 값을 읽는다. 다음 함수를 판정 셸에 먼저 정의한다.
YAML 디코딩과 동일하게 인용부호를 벗기고, 첫 번째 프론트매터 블록만 본다.

```bash
phase_of() {
  awk '/^---$/{n++; next} n==1 && /^phase:/{sub(/^phase:[ \t]*/,""); gsub(/^["'"'"']|["'"'"']$/,""); print; exit}' "$1"
}
```

### A.2 공허 통과 방지

- 모든 판정 명령은 저작 시점에 **실제로 실행**했고, 아래 baseline은 그때 관측한
  값이다. 실행하지 않은 명령은 이 문서에 없다.
- 위치는 **내용 토큰**으로 고정한다. 줄 번호와 커밋 SHA는 판정 시점에 이미 밀려
  있으므로 앵커로 쓰지 않는다.
- 0건이 통과 조건인 판정은, 명령이 **0건 아닌 값을 낼 수 있음**을 baseline이
  증명해야 한다. baseline이 이미 0인 명령은 판정력이 없으므로 쓰지 않는다.
- 입력 파일을 소비하는 판정은 **그 파일을 만드는 단계를 같은 AC 안에 포함**하고,
  파일 부재 시 조용히 통과하지 않도록 존재 가드를 둔다.
- `go test -run <선택자>`는 0개 매칭 시 exit 0을 내므로, 선택자 매칭 수를 별도
  AC(AC-PFV-014)로 **먼저** 확인한다.

### A.3 증거 경로와 실행 비용

- 증거 산출물은 `.moai/state/verify/pfv/`에 남긴다. `/tmp`는 OS가 비우므로 감사
  시점에 인용 경로가 해소되지 않는다.
- 저장소 전역 `moai spec lint --json`은 git 조회 때문에 약 3분 걸린다(저작 시 실측
  3:11). 타임아웃을 넉넉히 잡는다.
- 모든 명령은 저장소 루트에서 실행한다. 파일을 변조하는 판정은 **사본**에 대해
  수행하며 추적 파일을 건드리지 않는다. 인플레이스 편집은 BSD/GNU 양쪽에서 동작하는
  `perl -0pi -e`를 쓴다(`sed -i ''`는 BSD 전용).

---

## B. 탐지 — 도구가 위반을 본다

### AC-PFV-001 (maps REQ-PFV-001)

**Given** 값-형태 검증이 도입된 린터,
**When** SPEC 디렉터리 사본의 `phase`에 `plan`을 심고 린트하면,
**Then** `FrontmatterPhaseInvalid` finding이 **error** severity로 1건 이상 나온다.

```bash
mkdir -p .moai/state/verify/pfv
D=.moai/state/verify/pfv/ac001
rm -rf "$D" && cp -R .moai/specs/SPEC-PHASE-FIELD-VALIDATION-001 "$D"
perl -0pi -e 's/^phase: "v3\.0\.2"$/phase: plan/m' "$D/spec.md"
moai spec lint "$D/spec.md" --json \
  | tee .moai/state/verify/pfv/ac001.json \
  | jq '[.[]|select(.code=="FrontmatterPhaseInvalid" and .severity=="error")]|length'
```

- 관측 baseline (저작 시점, 변경 전): `moai spec lint <이 SPEC의 spec.md> --json` → `[]`.
  심기 전에는 0건이므로 이 명령은 판정력이 있다.
- 통과: `>= 1`.
- 이 사본의 디렉터리는 `§E.2`를 담고 `sync_commit_sha`가 없어 H-3 → V3R5 → 유산으로
  분류된다. 즉 이 AC는 **유산 분류 상태에서도 error가 나온다**를 동시에 확인한다.
  추적 파일은 건드리지 않는다.

### AC-PFV-002 (maps REQ-PFV-005, REQ-PFV-014)

**Given** 유산 era 픽스처와 terminal 상태 픽스처,
**When** 각각에 부정 토큰을 심고 규칙을 돌리면,
**Then** 두 경우 모두 severity가 `error`이고 `advisory`가 참이 아니다.

- 판정은 M4가 추가하는 Go 테스트의 두 픽스처 케이스로 한다.
- 통과: 두 케이스가 `severity == error && advisory != true`를 단언하고 PASS.
- 이 AC가 없으면 "강등을 벗어난다"는 설계 의도가 관측되지 않은 부작용으로 남는다.
  두 갈래(era / terminal)를 각각 픽스처로 고정한다.

### AC-PFV-003 (maps REQ-PFV-002)

**Given** M1이 도입한 신규 코드,
**When** 강등 대상 코드 집합의 리터럴을 조회하면,
**Then** 신규 코드가 그 집합에 **없다**.

```bash
awk '/eraDemotableCodes = map\[string\]bool\{/,/^\}/' internal/spec/lint.go \
  | tee .moai/state/verify/pfv/ac003.txt
grep -c 'FrontmatterPhaseInvalid' .moai/state/verify/pfv/ac003.txt
grep -c 'FrontmatterInvalid' .moai/state/verify/pfv/ac003.txt
```

- 관측 baseline (변경 전): 집합 블록은 `MissingExclusions`와 `FrontmatterInvalid`
  두 항목을 담는다. 즉 두 번째 grep은 `1`을 낸다 — 명령이 0 아닌 값을 낼 수 있음이
  증명되므로 첫 번째 grep의 `0` 판정은 공허하지 않다.
- 통과: 첫 grep `0`, 둘째 grep `1`(기존 항목 보존 확인).
- 이 등록 여부가 이 SPEC의 핵심 설계 결정이므로 산문이 아니라 기계로 판정한다.

### AC-PFV-004 (maps REQ-PFV-003, REQ-PFV-004)

**Given** `Runtime`처럼 부정 토큰을 부분 문자열로 포함하는 정당한 유산 값,
**When** 정확 일치 판정과 부분 문자열 판정을 같은 코퍼스에 각각 적용하면,
**Then** 정확 일치는 이들을 위반으로 세지 않고, 부분 문자열은 8건을 추가로 센다.

```bash
exact=0; substr_only=0
for f in $(find .moai/specs -name 'spec.md'); do
  v=$(phase_of "$f"); lv=$(printf '%s' "$v" | tr 'A-Z' 'a-z')
  case "$lv" in
    plan|run|sync|mx) exact=$((exact+1));;
    *plan*|*run*|*sync*|*mx*) substr_only=$((substr_only+1)); echo "SUBSTR-ONLY: $v";;
  esac
done
echo "exact=$exact substr_only=$substr_only"
```

- 관측 baseline (저작 시점): `exact=9 substr_only=8`. `SUBSTR-ONLY` 출력은
  `Runtime Hardening` 5건, `Agent Runtime Robustness` 1건, `Runtime Protocol Migration`
  1건, `Runtime Safety Net` 1건.
- 통과: M3 이후 `exact=0`이고 `substr_only=8`이 유지된다. `substr_only`가 0으로
  떨어지면 정당한 값이 삭제·변조된 것이고, 이들이 `exact`로 넘어오면 술어가 부분
  문자열 판정으로 퇴행한 것이다. 양방향으로 판정력이 있다.
- 추가로 M4 테스트가 이 네 값 각각에 대해 finding 0건을 단언한다.

---

## C. 비회귀 — 아무것도 새로 깨지지 않는다

### AC-PFV-005 (maps REQ-PFV-016)

**Given** 모든 마일스톤이 랜딩된 저장소,
**When** 저장소 전역 린트를 돌리면,
**Then** error severity finding이 0건이고 종료코드가 0이다.

```bash
mkdir -p .moai/state/verify/pfv
moai spec lint --json > .moai/state/verify/pfv/after.json; echo "exit=$?"
jq '[.[]|select(.severity=="error")]|length' .moai/state/verify/pfv/after.json
jq 'length' .moai/state/verify/pfv/after.json
jq -r 'group_by(.code)[]|"\(length) \(.[0].code)"' .moai/state/verify/pfv/after.json
```

- 관측 baseline (저작 시점, 이 SPEC 산출물 포함): 총 62건, error **0**건, exit 0.
  내역은 warning `MissingExclusions` 24 / `StatusGitConsistency` 16 /
  `FrontmatterInvalid` 14 / `LegacyEARSKeyword` 7 / `OwnershipTransitionInvalid` 1.
- 통과: error 0건, exit 0. 총건수는 62에서 줄어들 수 있다(교정으로 기존 warning이
  사라질 수 있음). **늘어나면 조사 대상**이다.

### AC-PFV-006 (maps REQ-PFV-016)

**Given** 변경 전후의 error-severity 파일 집합,
**When** 두 집합을 비교하면,
**Then** 변경 후에 새로 나타난 파일이 없다.

```bash
mkdir -p .moai/state/verify/pfv
B=.moai/state/verify/pfv/before.json
A=.moai/state/verify/pfv/after.json

# (1) BEFORE 생성 — M1/M3 랜딩 **이전**에 실행해야 한다.
#     이미 랜딩했다면 merge-base 트리에서 재생성한다.
[ -s "$B" ] || { echo "BEFORE baseline 없음 — M1/M3 이전에 생성하라"; exit 1; }

# (2) 존재 + 파싱 가드 — 부재/깨짐이면 조용히 통과하지 않고 실패한다.
jq -e 'type=="array"' "$B" >/dev/null || { echo "BEFORE 파싱 실패"; exit 1; }
jq -e 'type=="array"' "$A" >/dev/null || { echo "AFTER 파싱 실패"; exit 1; }

jq -r '[.[]|select(.severity=="error")|.file]|sort|.[]' "$B" > .moai/state/verify/pfv/err-before.txt
jq -r '[.[]|select(.severity=="error")|.file]|sort|.[]' "$A" > .moai/state/verify/pfv/err-after.txt
comm -13 .moai/state/verify/pfv/err-before.txt .moai/state/verify/pfv/err-after.txt
```

- 관측 baseline: 변경 전 error 파일 목록은 **빈 목록**이다(AC-PFV-005의 error 0건과
  정합).
- 통과: `comm -13`(after에만 있는 줄) 무출력. 존재 가드가 없으면 두 빈 파일의
  비교가 무출력으로 통과해 공허 판정이 되므로, (1)과 (2)를 생략할 수 없다.

### AC-PFV-007 (maps REQ-PFV-004)

**Given** 엄격 semver 형태를 벗어나는 정당한 `phase` 값들,
**When** 개수를 세고 error 목록과 대조하면,
**Then** 그 값들 때문에 새 finding이 생기지 않는다 — 허용목록이 도입되지 않았다.

```bash
ns=0; for f in $(find .moai/specs -name 'spec.md'); do
  printf '%s' "$(phase_of "$f")" | grep -qE '^v[0-9]+\.[0-9]+\.[0-9]+$' || ns=$((ns+1)); done
echo "non-strict-semver phase values = $ns"
```

- 관측 baseline: **310**건.
- 통과: 이 수가 300건대를 유지하면서 AC-PFV-005의 error가 0이다. 급감했다면
  허용목록이 몰래 들어온 것이다.

---

## D. 교정 — 오염이 사라진다

### AC-PFV-008 (maps REQ-PFV-009, REQ-PFV-010)

**Given** M3이 완료된 저장소,
**When** 모든 `spec.md`에 부정 토큰 정확 일치 스윕을 돌리면,
**Then** 위반이 0건이다.

```bash
hits=0
for f in $(find .moai/specs -name 'spec.md'); do
  v=$(phase_of "$f"); lv=$(printf '%s' "$v" | tr 'A-Z' 'a-z')
  case "$lv" in plan|run|sync|mx) echo "VIOLATION $f -> [$v]"; hits=$((hits+1));; esac
done
echo "spec.md violations = $hits"
```

- 관측 baseline (저작 시점): **9**건. 목록은
  `SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001`, `SPEC-WORKTREE-BRANCH-GUARD-001`,
  `SPEC-CI-LOOP-DEVONLY-001`, `SPEC-UPDATE-YAML-PRESERVE-001`,
  `SPEC-UPDATE-REINSTALL-LOOP-002`, `SPEC-UPDATE-GUARD-EFFICACY-001`,
  `SPEC-REF-SEO-ABSORB-001`, `SPEC-PIPELINE-FANOUT-ACTIVATION-001`,
  `SPEC-ENVKEY-ANTHROPIC-SSOT-001`.
- 통과: `0`. 전용 코드에서는 9건 전부가 un-demoted error이므로 부분 교정은 성립하지
  않는다 — 하나라도 남으면 AC-PFV-005가 함께 실패한다.

### AC-PFV-009 (maps REQ-PFV-011)

**Given** in-scope SPEC 2개의 형제 산출물 5건,
**When** 각 파일의 `phase`를 읽으면,
**Then** 모두 `v3.0.2`이며 각자의 `spec.md`와 일치한다.

```bash
for f in .moai/specs/SPEC-ENVKEY-ANTHROPIC-SSOT-001/plan.md \
         .moai/specs/SPEC-ENVKEY-ANTHROPIC-SSOT-001/acceptance.md \
         .moai/specs/SPEC-ENVKEY-ANTHROPIC-SSOT-001/progress.md \
         .moai/specs/SPEC-WORKTREE-BRANCH-GUARD-001/plan.md \
         .moai/specs/SPEC-WORKTREE-BRANCH-GUARD-001/acceptance.md; do
  echo "$f -> [$(phase_of "$f")]"
done
```

- 관측 baseline: 차례로 `plan`, `plan`, `run`, `plan`, `plan`.
- 통과: 다섯 줄 모두 `[v3.0.2]`.
- **판정 성격 주의**: 이 다섯 건은 린터가 읽지 않는 파일이다(발견 함수는
  `SPEC-*/spec.md`만 수집). 따라서 이 AC는 파일 내용 직접 확인으로만 판정하며,
  "린트가 탐지한다"를 주장하지 않는다.

### AC-PFV-010 (maps REQ-PFV-011)

**Given** 범위에서 제외한 레거시 형제 산출물 17건,
**When** 전체 산출물 스윕을 돌리면,
**Then** 잔여 위반이 정확히 17건이며 전부 레거시 SPEC 소속이다.

```bash
all=0
for f in $(find .moai/specs -name '*.md'); do
  v=$(phase_of "$f"); lv=$(printf '%s' "$v" | tr 'A-Z' 'a-z')
  case "$lv" in plan|run|sync|mx) echo "$f"; all=$((all+1));; esac
done
echo "all-artifact violations = $all"
```

- 관측 baseline: **31**건 (spec.md 9 + in-scope 형제 5 + 레거시 형제 17).
- 통과: `17`, 그리고 출력된 경로가 모두 `SPEC-V3R2-*` / `SPEC-V3R3-*` /
  `SPEC-V3R6-LINK-FIX-001` / `SPEC-GLM-MCP-001` / `SPEC-CI-MULTI-LLM-001` /
  `SPEC-TOKEN-001` 중 하나에 속한다.
- 이 AC는 "제외가 의도된 것임"을 고정한다. 17이 아니라 0이면 범위를 넘어선 것이고,
  17보다 크면 교정이 덜 된 것이다. 양방향으로 판정력이 있다.

### AC-PFV-011 (maps REQ-PFV-012)

**Given** M3이 배정한 타깃 값,
**When** 각 in-scope SPEC의 최초 커밋 시각을 최신 태그와 비교하면,
**Then** 아홉 건 모두 태그 이후이며, 따라서 `v3.0.2` 배정이 이력으로 뒷받침된다.

```bash
git log -1 --format='%ci' v3.0.1
for d in SPEC-ENVKEY-ANTHROPIC-SSOT-001 SPEC-WORKTREE-BRANCH-GUARD-001 \
         SPEC-CI-LOOP-DEVONLY-001 SPEC-PIPELINE-FANOUT-ACTIVATION-001 \
         SPEC-UPDATE-GUARD-EFFICACY-001 SPEC-UPDATE-REINSTALL-LOOP-002 \
         SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001 SPEC-REF-SEO-ABSORB-001 \
         SPEC-UPDATE-YAML-PRESERVE-001; do
  echo "$d first=$(git log --reverse --format='%ci' -- ".moai/specs/$d/" | head -1)"
done
```

- 관측 baseline: 태그 `2026-07-24 01:37:05 +0900`. 최초 커밋은 `2026-07-27 19:50`부터
  `2026-08-02 04:30` 사이에 분포하며 아홉 건 모두 태그 이후.
- 통과: 아홉 줄 모두 태그 시각보다 늦다.

---

## E. 저작 지시

### AC-PFV-012 (maps REQ-PFV-006, REQ-PFV-008)

**Given** plan-phase 저작 에이전트 정의,
**When** `phase` 프론트매터 필드 설명을 찾으면,
**Then** 릴리스 타깃이라는 의미, 부정 토큰, 생명주기 필드가 아니라는 서술이 모두 있다.

```bash
F=.claude/agents/moai/manager-spec.md
grep -c 'release target' "$F"
grep -cE '`plan`.*`run`.*`sync`|워크플로 단계|workflow-stage' "$F"
grep -cE 'not a lifecycle|생명주기 단계 필드가 아' "$F"
```

- 관측 baseline: `grep -n "phase:" "$F"` → **1건**이며, 그 유일한 줄은 `progress.md`의
  `§E.2`~`§E.4` 절에 관한 서술로 프론트매터 필드와 무관하다. 세 grep 모두 현재 `0`.
- 통과: 세 grep 모두 `>= 1`.

### AC-PFV-013 (maps REQ-PFV-007)

**Given** 배포 템플릿 미러,
**When** 같은 지시를 찾고 중립성 위반 집합의 **변화량**을 확인하면,
**Then** 지시가 존재하고 위반 줄 집합이 baseline과 동일하다.

```bash
T=internal/template/templates/.claude/agents/moai/manager-spec.md
P='SPEC-[A-Z0-9-]+-[0-9]{3}|internal/spec/|REQ-[A-Z]{2,}|20[0-9]{2}-[0-9]{2}-[0-9]{2}|\b[0-9a-f]{7,40}\b'
grep -c 'release target' "$T"
grep -nE "$P" "$T" | cut -d: -f1 | tr '\n' ' '
go test ./internal/template/... -count=1
```

- **세 번째 명령은 패키지 전체를 돌린다 — 선택자로 좁히지 않는다.** `-run 'Leak|Neutral'`로
  좁히면 `TestManifestHashFormat`이 선택자에 걸리지 않는다(실측: 해당 선택자는 15개를
  매칭하고 그 안에 없다). 그런데 `manager-spec.md`는 카탈로그 등재 파일이라 편집하면
  `catalog.yaml`의 SHA256이 무효화돼 **정확히 그 테스트가 깨진다**. 즉 좁힌 형태는
  M2가 유발하는 유일한 실패를 놓치고 PASS를 보고한다. 전체 패키지 실행이 판정 조건이다.

- 관측 baseline (저작 시점): 이 패턴은 **9건**을 맞히며, 줄 번호 집합은
  `68 69 104 134 135 149 150 151 202`이다. 내용은 전부 교육용 예시 식별자
  (`SPEC-AUTH-001`, `SPEC-A-001`, `SPEC-Z-001`, `SPEC-NEW-001`,
  `SPEC-RETIRED-DDD-001`, `SPEC-V3R6-SPEC-ID-VALIDATION-001`) 와 `REQ-ARR-002/003`으로,
  M2가 도입하는 것이 아니다. `grep -c 'release target'` 은 현재 `0`.
- 통과: 첫 grep `>= 1`; 위반 **건수가 9에서 늘지 않고** 새 줄 번호가 baseline 집합
  밖에서 나타나지 않는다(줄 번호는 M2 삽입으로 이동할 수 있으므로 **집합 크기와
  내용 토큰**으로 판정하고, 늘어난 항목이 있으면 그 항목의 텍스트를 확인한다);
  템플릿 가드 테스트 PASS.
- baseline을 기록하지 않으면 "새 위반 vs 기존 위반"을 판별할 수 없으므로 이 AC는
  절대 건수가 아니라 **증분**으로 판정한다.

---

## F. 가드 반증 가능성

### AC-PFV-014 (maps REQ-PFV-015)

**Given** M4가 추가한 회귀 테스트,
**When** 고정된 선택자를 나열 모드로 돌리면,
**Then** 1개 이상 매칭한다.

```bash
go test ./internal/spec/ -list '.*' -count=1 \
  | tee .moai/state/verify/pfv/ac014.txt
grep -c '^TestPhaseValueShape' .moai/state/verify/pfv/ac014.txt
```

- **`-run`을 붙이지 않는다.** `-list`는 `-run`을 **무시하고** 패키지의 전체 테스트를
  나열한다(실측: `-run '^TestPhaseValueShape'` 유무와 무관하게 둘 다 278줄). 따라서
  `-run`을 함께 쓰면 "선택자가 동작하는 것처럼 보이지만 실제로는 `grep -c`가 거른다"는
  오해를 만든다. 실제 필터는 **`grep -c` 한 곳**이며, 명령 형태가 그 사실을 그대로
  드러내야 한다.
- 선택자 접두사는 `^TestPhaseValueShape`로 **고정**한다(plan.md §F M4가 이 접두사를
  명명 규약으로 지정). AC-PFV-015의 `-run`은 실제로 실행 범위를 좁히므로 그쪽에서는
  선택자가 유효하다.
- 통과: `>= 1`.
- 이 AC를 **먼저** 통과시키지 않고 AC-PFV-015를 판정하면 안 된다.

### AC-PFV-015 (maps REQ-PFV-013)

**Given** PASS 상태의 회귀 테스트,
**When** (a) 규칙에서 값-형태 검사의 **호출부**를 제거한 뒤, 그리고 (b) 호출부는
그대로 두고 술어 **본문**을 무력화(항상 거짓 반환)한 뒤 각각 테스트를 돌리면,
**Then** **두 경우 모두** 테스트가 FAIL한다.

```bash
# 왕복 (a) 호출부 제거 → 실행 → 복원
go test ./internal/spec/ -run '^TestPhaseValueShape' -count=1   # 기대: FAIL
# 왕복 (b) 술어 본문 무력화 → 실행 → 복원
go test ./internal/spec/ -run '^TestPhaseValueShape' -count=1   # 기대: FAIL
```

- 통과: 두 왕복 각각에서 FAIL을 관측하고, 각 복원 후 PASS를 관측한다. run-phase
  보고는 각 왕복의 **FAIL 출력 원문을 인용**한다 — "FAIL 관측함"이라는 서술만으로는
  이 AC가 충족되지 않는다.
- 두 왕복이 모두 필요한 이유: (a)만 잡히면 테스트는 "호출되었다"만 증명하며, 이는
  도달 가능성이지 하중이 아니다. (b)가 FAIL해야 술어의 **내용**이 하중을 받는다.
- **왕복 (b) 전 호출부 존재 grep이 필수다.** 두 왕복의 FAIL 출력은 바이트 단위로
  동일하므로(plan.md §F M4 설계 제약 (a)), 출력만으로는 (a)를 두 번 돌린 것과
  구별되지 않는다. (b) 실행 직전에 호출부가 아직 남아 있음을 grep으로 확인하고 그
  출력을 함께 인용해야 두 왕복이 실제로 서로 다른 지점을 건드렸음이 증명된다.

---

## G. 빌드

### AC-PFV-016 (maps REQ-PFV-001, REQ-PFV-013)

**Given** 모든 마일스톤,
**When** 빌드와 정적 분석과 패키지 테스트를 돌리면,
**Then** 성공한다.

```bash
go build ./...
go vet ./internal/spec/
go test ./internal/spec/ -count=1
```

- 통과: 세 명령 모두 exit 0.

---

## H. Definition of Done

- [ ] AC-PFV-001 ~ AC-PFV-016 전부 PASS, 각 판정 명령의 실제 출력이 인용됨
- [ ] AC-PFV-014를 AC-PFV-015보다 **먼저** 판정
- [ ] AC-PFV-015의 두 왕복에서 FAIL 출력 **원문**을 인용(복원 후 PASS도 관측)
- [ ] AC-PFV-015 왕복 (b) 직전 호출부 존재 grep 출력을 인용 — 두 왕복의 FAIL 출력이
      동일하므로 이 grep이 (a)/(b) 구별의 유일한 증거
- [ ] AC-PFV-013의 세 번째 명령을 `-run` 선택자로 좁히지 않음(전체 패키지 실행)
- [ ] AC-PFV-006의 BEFORE baseline을 M1/M3 랜딩 **이전**에 생성
- [ ] AC-PFV-009/010의 판정을 "린트가 탐지"로 서술하지 않음(린트 불가시 영역)
- [ ] 증거 산출물이 `.moai/state/verify/pfv/`에 남아 감사 시점에 해소됨
- [ ] 미커버 REQ 0건 — REQ-PFV-001~016 전부 아래 대조표로 확인
- [ ] `moai spec lint --json` error 0건, exit 0

### REQ 커버리지 대조

| REQ | 커버하는 AC |
|---|---|
| REQ-PFV-001 | AC-PFV-001, AC-PFV-016 |
| REQ-PFV-002 | AC-PFV-003 |
| REQ-PFV-003 | AC-PFV-004 |
| REQ-PFV-004 | AC-PFV-004, AC-PFV-007 |
| REQ-PFV-005 | AC-PFV-002 |
| REQ-PFV-006 | AC-PFV-012 |
| REQ-PFV-007 | AC-PFV-013 |
| REQ-PFV-008 | AC-PFV-012 |
| REQ-PFV-009 | AC-PFV-008 |
| REQ-PFV-010 | AC-PFV-008 |
| REQ-PFV-011 | AC-PFV-009, AC-PFV-010 |
| REQ-PFV-012 | AC-PFV-011 |
| REQ-PFV-013 | AC-PFV-015, AC-PFV-016 |
| REQ-PFV-014 | AC-PFV-002 |
| REQ-PFV-015 | AC-PFV-014 |
| REQ-PFV-016 | AC-PFV-005, AC-PFV-006 |
