# SPEC-ARTIFACT-STATELESS-001 — 인수 기준

> 모든 명령은 워크트리 루트에서, 프로젝트 루트 상대 경로로 실행한다.
>
> **baseline 귀속 규칙**: 판정 명령은 어떤 계수도 리터럴로 고정하지 않는다. `c6aa61346`의 362와 `3b1830b96`의 389는 **참고값**이며, 판정은 실행 시점 HEAD에서 재측정한 값으로 한다(REQ-AST-001-009).
>
> **모집단**: D1 정리의 모집단은 **전 코퍼스 696**(`.moai/specs/SPEC-*/` 전부, 라이프사이클 status 무관)이다. 362는 그 종결-한정 부분집합이며 판정값이 아니다(§1.6).
>
> **고정 토큰 2종** — 이 문서가 판정하는 이름:
> - 신규 lint finding 코드: `ArtifactStatusFieldForbidden`
> - M1 신설 소절 앵커 제목: `### Artifact Statelessness`
>
> **기준값 슬롯**: `$BASE_M3`(M3 착수 직전 SHA) · `$BASE_SPEC`(SPEC 착수 직전 SHA) · baseline `N`(M3 착수 시 D1 대상 수)은 `progress.md §E.1`의 「기준값」 표에서 읽는다. [HARD] 추출은 **반드시 공용 스니펫 (C)의 3중 가드를 거친다** — 빈 슬롯이 조용히 빈 문자열로 통과하면 `""..HEAD`가 `HEAD..HEAD`로 해석되어 AC가 아무것도 검사하지 않고 PASS를 출력한다.

---

## 공용 스니펫

각 코드 블록은 별개 셸에서 실행되므로, 아래 네 정의는 **필요한 블록마다 인라인한다**(붙여넣기만으로 실행되게 하기 위함).

```bash
# (A) 소절 추출 — 앵커가 없으면 빈 문자열이 되어 이하 모든 grep이 실패한다.
#     종료 조건은 **레벨 3 이하 제목**(`# `/`## `/`### `)이다. `/^##/`로 두면
#     소절 안의 `#### 하위제목`에서도 끊겨, 올바른 소절이 오탐 FAIL한다.
F=.claude/rules/moai/development/spec-frontmatter-schema.md
SEC=$(awk '/^### Artifact Statelessness/{f=1;next} f&&/^#{1,3} /{exit} f' "$F")

# (B) D1 잔여 계수 — 전 코퍼스 696 모집단, frontmatter 블록 안의 status: 만 센다
count_d1() {
  local n=0 f
  for d in .moai/specs/SPEC-*/; do
    for a in plan acceptance design research; do
      f="${d}${a}.md"; [ -f "$f" ] || continue
      awk 'NR==1&&/^---/{p=1;next} p&&/^---/{exit} p' "$f" \
        | grep -qE '^status:[[:space:]]' && n=$((n+1))
    done
  done
  echo "$n"
}

# (C) 기준값 추출 + 3중 가드 — 빈 슬롯 / 오형식 / 이 트리에 없는 SHA 를 모두
#     **큰 소리로** 죽인다. 가드가 없으면 빈 슬롯이 ""..HEAD → HEAD..HEAD 로
#     조용히 해석되어 빈 diff·rc=0 을 내고, AC가 아무것도 검사하지 않은 채
#     PASS를 출력한다.
P=.moai/specs/SPEC-ARTIFACT-STATELESS-001/progress.md
read_sha() {   # $1 = progress.md 「기준값」 표의 행 레이블
  local v
  # 가드1: {7,} — 0회 반복을 막아 빈 슬롯 행에 아예 매치하지 않게 한다
  v=$(sed -n "s/^| $1 | \`\([0-9a-f]\{7,\}\)\` |.*/\1/p" "$P")
  # 가드2: 비었으면(미매치 포함) 즉시 실패
  [ -n "$v" ] || { echo "FAIL — 「$1」 슬롯이 비어 있거나 7자리 미만이다" >&2; return 1; }
  # 가드3: 형식은 맞아도 이 트리에 없는 SHA면 실패
  git rev-parse --verify "$v^{commit}" >/dev/null 2>&1 \
    || { echo "FAIL — 「$1」 = $v 가 이 트리에서 커밋으로 해석되지 않는다" >&2; return 1; }
  echo "$v"
}
read_num() {   # $1 = 행 레이블 — D1 baseline N 용 (숫자, 최소 1자리)
  local v
  v=$(sed -n "s/^| $1 | \`\([0-9]\{1,\}\)\` |.*/\1/p" "$P")
  [ -n "$v" ] || { echo "FAIL — 「$1」 슬롯이 비어 있거나 숫자가 아니다" >&2; return 1; }
  echo "$v"
}
```

> **왜 세 겹인가**: `[0-9a-f]*`는 0회 반복을 허용해 **빈 슬롯 행에 매치하고 빈 문자열을 캡처한다** — 실패가 아니라 조용한 성공이다. `\{7,\}`가 그 매치 자체를 막고, `-n` 검사가 미매치를 잡고, `rev-parse --verify`가 형식은 맞으나 존재하지 않는 SHA(다른 레인의 커밋, 오타)를 잡는다. 셋 중 하나라도 빠지면 빈/잘못된 기준값이 `git diff`에 들어가 AC를 공허하게 만든다.

---

## AC-AST-001-01 — 신설 소절이 세 문장을 각각 담는다

**Given** 규약에 `### Artifact Statelessness` 앵커가 없어 소절 추출이 빈 문자열을 내는 상태에서
**When** M1 소절이 착지하면
**Then** 그 소절 본문 안에서 (1) spec.md 한정, (2) 4종 무상태 + status 축, (3) Tier 무관 **세 문장이 각각 독립으로** 판정된다.

판정 범위를 소절로 한정하는 것이 핵심이다. 문서 전역 grep은 이미 참인 문장들에 걸려 판정력을 잃는다.

```bash
F=.claude/rules/moai/development/spec-frontmatter-schema.md
SEC=$(awk '/^### Artifact Statelessness/{f=1;next} f&&/^#{1,3} /{exit} f' "$F")
echo "section bytes = ${#SEC}"

# 부정 근접 검사 — 리터럴이 부정문 안에 들어가 있으면 그 문장은 단언이 아니다
neg_hit() {   # $1 = 리터럴
  printf '%s\n' "$SEC" | grep -F "$1" \
    | grep -qiE '\b(not|never|no longer|rejected|incorrect|false)\b|아니|않'
}

# S1 — spec.md 한정
if printf '%s\n' "$SEC" | grep -qF 'binds `spec.md` only'; then
  neg_hit 'binds `spec.md` only' && echo "S1 FAIL — 리터럴이 부정문 안에 있다" || echo "S1 PASS"
else echo "S1 FAIL"; fi

# S2 — 4종 전부 + status 축
s2=1
for a in plan.md acceptance.md design.md research.md; do
  printf '%s\n' "$SEC" | grep -qF "$a" || s2=0
done
printf '%s\n' "$SEC" | grep -qE '`status:`' || s2=0
[ "$s2" -eq 1 ] && echo "S2 PASS" || echo "S2 FAIL"

# S3 — Tier 무관
if printf '%s\n' "$SEC" | grep -qF 'Tier-independent'; then
  neg_hit 'Tier-independent' && echo "S3 FAIL — 리터럴이 부정문 안에 있다" || echo "S3 PASS"
else echo "S3 FAIL"; fi
```

판정: **S1·S2·S3 세 줄이 모두 `PASS`.** 하나라도 FAIL이면 AC 전체 FAIL.

**착지 전 실측 (`3b1830b96` / 재확인 `aacad4f99`, 미수정 트리)** — `section bytes = 0`, `S1 FAIL` / `S2 FAIL` / `S3 FAIL`. 세 판정 모두 착지 전에는 성립할 수 없다.

> **[HARD] 이 검사가 판정하는 것은 리터럴의 존재와 배치이지 문장의 주장(polarity)이 아니다.** 부정 근접 검사는 같은 줄 안의 부정어만 잡는다. 여러 줄에 걸친 부정(`… only.` 다음 줄에 `— 이것은 사실이 아니다`)은 기계로 닫히지 않으므로, DoD의 인간 판독 항목이 유일한 방어다. 잔여는 spec.md §5에 부채로 명시돼 있다.

(요구사항: REQ-AST-001-001 / 002 / 003)

---

## AC-AST-001-02 — 무상태가 status 축으로 한정되고, frontmatter 일체 금지가 아니다

**Given** "무상태"가 frontmatter 전면 금지로 읽힐 여지가 있을 때
**When** 신설 소절을 잘라내 읽으면
**Then** frontmatter 허용을 명시한 문장이 **있고**, frontmatter 자체를 금지하는 문형이 **0건**이다.

두 조건의 **논리곱**이 판정이다 — 부정 검사만으로는 소절이 없어도(0건) 통과하므로 공허해진다.

```bash
F=.claude/rules/moai/development/spec-frontmatter-schema.md
SEC=$(awk '/^### Artifact Statelessness/{f=1;next} f&&/^#{1,3} /{exit} f' "$F")

P=0; printf '%s\n' "$SEC" | grep -qF 'Frontmatter itself is permitted' && P=1
# 부정 근접 — 허용 문장이 부정문 안에 있으면 허용이 아니다
printf '%s\n' "$SEC" | grep -F 'Frontmatter itself is permitted' \
  | grep -qiE '\b(not|never|no longer|rejected|incorrect|false)\b|아니|않' && P=0
N=$(printf '%s\n' "$SEC" | grep -cE 'MUST NOT (carry|have|contain) (a |any )?(YAML )?frontmatter')
echo "permission=$P blanket_prohibition=$N"
[ "$P" -eq 1 ] && [ "$N" -eq 0 ] && echo PASS || echo FAIL
```

판정: `permission=1 blanket_prohibition=0` + `PASS`.

**착지 전 실측 (`3b1830b96` / 재확인 `aacad4f99`)** — `permission=0 blanket_prohibition=0` → `FAIL`. 부정 검사는 이미 0이지만 허용 문장이 없어 전체가 FAIL한다. (요구사항: REQ-AST-001-003 / 014)

> **설계 메모 (왜 이 형태인가)**: 종전 판정은 문서 전역에 `grep -inE '(no|never|금지).{0,40}frontmatter'`를 걸고 `||` 분기로 PASS 문자열을 냈다. 그 grep은 미수정 파일에서 이미 4행 이상 매치하므로 `||` 분기가 결코 실행되지 않았고, 선언한 PASS 문자열은 **원리상 출력될 수 없었다.** 범위를 소절로 좁히고 매치 패턴을 "frontmatter 자체 금지" 문형으로 한정해 이 결함을 닫는다.

---

## AC-AST-001-03 — 새 lint 규칙이 등록되고 era 예외에 없다

**Given** `internal/spec/lint.go`가 규칙 배열과 `eraDemotableCodes`를 가질 때
**When** M2가 착지하면
**Then** `ArtifactStatusFieldForbidden`이 규칙으로 등록되고, `eraDemotableCodes` 블록에는 **없다**.

```bash
echo "rule registered:"
grep -c 'ArtifactStatusFieldForbidden' internal/spec/lint.go
echo "--- eraDemotableCodes block ---"
sed -n '/^var eraDemotableCodes/,/^}/p' internal/spec/lint.go
```

판정: 첫 계수 ≥ 1 **그리고** `eraDemotableCodes` 블록 출력에 `ArtifactStatusFieldForbidden`이 나타나지 않는다.

**착지 전 실측** — 첫 계수 `0` → FAIL. (요구사항: REQ-AST-001-004 / 006)

---

## AC-AST-001-04 — [HARD] 심어서 거부를 관측한다

**Given** lint 규칙이 작성되어 있으나 실제로 거부하는 것을 아직 보지 못한 상태에서
**When** 이 SPEC의 `plan.md`에 `status:` 필드를 심고 **이 SPEC만 대상으로** lint를 돌리면
**Then** `ArtifactStatusFieldForbidden`과 심기 전후의 rc 변화를 **관측**하고, 심은 것을 원복한다.

[HARD] lint 호출은 **이 SPEC 하나로 한정한다.** 전 코퍼스 호출은 M3(정리) 미착지 상태에서 코퍼스 잔여(`3b1830b96` 기준 389)에 걸려, **정상 동작하는 lint에 대해서도 before ≠ 0**이 되어 FAIL한다 — 그러면 이 AC가 마일스톤 순서에 종속된다. 단일 SPEC 한정이면 before=0 비공허성 가드가 순서와 무관하게 유지된다.

규칙을 작성한 것은 규칙이 작동한다는 증거가 아니다. 아래를 **실행하고 출력을 인용**한다.

```bash
set -u
S=.moai/specs/SPEC-ARTIFACT-STATELESS-001
T="$S/plan.md"
cp "$T" .moai/cache/t357_ac04_backup.md

# 1) baseline — 심기 전에는 이 코드가 나오지 않아야 한다 (비공허성 가드)
moai spec lint "$S/spec.md" > .moai/reports/t357/t357_ac04_before.txt 2>&1; echo "before rc=$?"
grep -c 'ArtifactStatusFieldForbidden' .moai/reports/t357/t357_ac04_before.txt

# 2) 심는다 — plan.md 선두에 status 필드를 가진 frontmatter 블록을 붙인다
printf -- '---\nstatus: draft\n---\n\n' | cat - "$T" > .moai/cache/t357_planted.md
mv .moai/cache/t357_planted.md "$T"

# 3) 관측한다
moai spec lint "$S/spec.md" > .moai/reports/t357/t357_ac04_after.txt 2>&1; echo "after rc=$?"
grep -n 'ArtifactStatusFieldForbidden' .moai/reports/t357/t357_ac04_after.txt | head

# 4) 원복 — 반드시 실행한다
cp .moai/cache/t357_ac04_backup.md "$T"
git diff --stat -- "$T"
```

판정 (넷 모두 성립해야 PASS):

1. `before` 출력의 `ArtifactStatusFieldForbidden` 매치 수 = **0** — 심기 전에 이미 나오면 그 규칙은 다른 것을 보고 있다
2. `after` 출력에 `ArtifactStatusFieldForbidden`이 **1건 이상**
3. `after rc`가 **비-0** (규칙 심각도가 error일 때). warning 심각도를 택했다면 `after rc`는 0이어도 되나, 그때는 `moai spec lint --strict "$S/spec.md"`를 추가로 돌려 비-0을 관측한다
4. 원복 후 `git diff --stat`이 **빈 출력**

**착지 전 실측 (`3b1830b96`)** — `moai spec lint .moai/specs/SPEC-ARTIFACT-STATELESS-001/spec.md` → `✓ No findings`, rc=0, `ArtifactStatusFieldForbidden` 매치 0. 규칙이 없으므로 2항이 성립할 수 없어 AC 전체 FAIL이며, **동시에 1항(before=0)이 참임이 확인**되어 비공허성 가드가 실제로 작동하는 형태임을 보인다.

(요구사항: REQ-AST-001-007)

---

## AC-AST-001-05 — 통과 방향도 확인한다 (거짓 양성 없음)

**Given** frontmatter가 있으나 `status:`가 없는 산출물, `spec.md`, `progress.md`가 코퍼스에 존재할 때
**When** 전 코퍼스 lint를 돌리면
**Then** 그 셋 중 어느 것에 대해서도 `ArtifactStatusFieldForbidden`이 나오지 않는다.

```bash
moai spec lint > .moai/reports/t357/t357_ac05.txt 2>&1; echo "rc=$?"
grep 'ArtifactStatusFieldForbidden' .moai/reports/t357/t357_ac05.txt \
  | grep -E '/(spec|progress)\.md' \
  && echo "FAIL — spec.md/progress.md 를 지목한 finding 존재" \
  || echo "no spec.md/progress.md hits: PASS"
```

판정: `no spec.md/progress.md hits: PASS`.

> 이 AC는 M3 착지 후에 판정한다 — M2만 착지한 시점에는 코퍼스 잔여로 다른 파일들의 finding이 나오지만, 그것은 이 AC가 보는 대상이 아니다(이 AC는 `spec.md`/`progress.md` 지목 여부만 본다).

(요구사항: REQ-AST-001-005 / 011)

---

## AC-AST-001-06 — D1 정리가 전 코퍼스에서 0으로 떨어진다

**Given** run-phase HEAD에서 재측정한 D1 대상이 N개일 때 (`3b1830b96`의 389는 참고값)
**When** M3가 착지하면
**Then** 같은 측정을 재실행했을 때 **전 코퍼스 696 모집단** 잔여가 **0**이 된다.

```bash
count_d1() {
  local n=0 f
  for d in .moai/specs/SPEC-*/; do
    for a in plan acceptance design research; do
      f="${d}${a}.md"; [ -f "$f" ] || continue
      awk 'NR==1&&/^---/{p=1;next} p&&/^---/{exit} p' "$f" \
        | grep -qE '^status:[[:space:]]' && n=$((n+1))
    done
  done
  echo "$n"
}
echo "HEAD=$(git rev-parse --short HEAD)"
echo "D1 remaining (all 696) = $(count_d1)"
```

판정: `D1 remaining (all 696) = 0`.

M3 **착수 전**에 같은 명령을 한 번 돌려 baseline N과 HEAD SHA를 `progress.md §E.2`의 「기준 SHA」 표에 기록한다.

**착지 전 실측 (`3b1830b96`)** — `bash .moai/reports/t357/t357_d1_all.sh .` → `D1 전체 696 모집단 = 389`. 0이 아니므로 FAIL. (요구사항: REQ-AST-001-008 / 009)

---

## AC-AST-001-07 — 정리가 status 라인만 건드렸다 (D1 준수, D2 아님)

**Given** D1은 `status:` 라인만, D2는 블록 전체를 지우는 정의일 때
**When** M3의 diff를 읽으면
**Then** 제거된 라인이 전부 `status:` 라인이고, `id:`/`title:`/`version:`/`created:` 라인은 하나도 제거되지 않았다.

```bash
P=.moai/specs/SPEC-ARTIFACT-STATELESS-001/progress.md
read_sha() {
  local v
  v=$(sed -n "s/^| $1 | \`\([0-9a-f]\{7,\}\)\` |.*/\1/p" "$P")
  [ -n "$v" ] || { echo "FAIL — 「$1」 슬롯이 비어 있거나 7자리 미만이다" >&2; return 1; }
  git rev-parse --verify "$v^{commit}" >/dev/null 2>&1 \
    || { echo "FAIL — 「$1」 = $v 가 이 트리에서 커밋으로 해석되지 않는다" >&2; return 1; }
  echo "$v"
}
read_num() {
  local v
  v=$(sed -n "s/^| $1 | \`\([0-9]\{1,\}\)\` |.*/\1/p" "$P")
  [ -n "$v" ] || { echo "FAIL — 「$1」 슬롯이 비어 있거나 숫자가 아니다" >&2; return 1; }
  echo "$v"
}

BASE_M3=$(read_sha 'M3 착수 직전') || { echo "AC-07 FAIL — 기준 SHA 없음"; exit 1; }
N=$(read_num 'M3 착수 시 D1 baseline N') || { echo "AC-07 FAIL — baseline N 없음"; exit 1; }
echo "BASE_M3=$BASE_M3  baseline_N=$N"

git diff "$BASE_M3"..HEAD -- '.moai/specs/*/plan.md' '.moai/specs/*/acceptance.md' \
  '.moai/specs/*/design.md' '.moai/specs/*/research.md' \
  | grep '^-' | grep -v '^---' > .moai/reports/t357/t357_removed.txt

TOTAL=$(wc -l < .moai/reports/t357/t357_removed.txt | tr -d ' ')
NONSTATUS=$(grep -cvE '^-status:' .moai/reports/t357/t357_removed.txt | tr -d ' ')
echo "removed lines total:      $TOTAL   (baseline N = $N)"
echo "removed non-status lines: $NONSTATUS"
grep -vE '^-status:' .moai/reports/t357/t357_removed.txt | head -20

# 두 조건을 **명령이** 판정한다 — 실행자의 기억에 맡기지 않는다
[ "$NONSTATUS" -eq 0 ] && [ "$TOTAL" -eq "$N" ] \
  && echo "AC-07 PASS" \
  || echo "AC-07 FAIL — non-status=$NONSTATUS (must be 0), total=$TOTAL vs N=$N (must be equal)"
```

판정: `AC-07 PASS`. 두 조건의 **논리곱**이다 — (1) 제거된 비-status 라인 0 (D1 준수), (2) 제거 총량 == M3 착수 시 baseline N (제거 누락 없음).

> **왜 두 번째 조건이 명령이어야 하는가**: 종전 판은 이 조건을 산문으로만 달아 두었다. 기준 SHA가 비면 `git diff`가 빈 결과를 내 `total=0`이 되고, 첫 조건(`non-status=0`)만 보는 실행자는 그것을 PASS로 읽는다 — 정확히 이 SPEC이 없애려는 "아무것도 보지 않는 검사"다. `total == N` 비교가 그 경로를 닫는다(빈 diff면 `0 != N`으로 즉시 FAIL).

**착지 전 실측 (`aacad4f99`, 빈 슬롯 상태)** — `read_sha`가 가드1·2에서 걸려 `FAIL — 「M3 착수 직전」 슬롯이 비어 있거나 7자리 미만이다` + `AC-07 FAIL — 기준 SHA 없음`, exit 1. **PASS를 출력하지 않는다.** (요구사항: REQ-AST-001-008)

---

## AC-AST-001-08 — 정리가 spec.md / progress.md를 건드리지 않았다

**Given** REQ-AST-001-011이 두 파일을 대상에서 제외할 때
**When** M3의 diff를 읽으면
**Then** 변경 파일 목록에 `spec.md`·`progress.md`가 없다 (이 SPEC 자신의 산출물 제외).

```bash
P=.moai/specs/SPEC-ARTIFACT-STATELESS-001/progress.md
read_sha() {
  local v
  v=$(sed -n "s/^| $1 | \`\([0-9a-f]\{7,\}\)\` |.*/\1/p" "$P")
  [ -n "$v" ] || { echo "FAIL — 「$1」 슬롯이 비어 있거나 7자리 미만이다" >&2; return 1; }
  git rev-parse --verify "$v^{commit}" >/dev/null 2>&1 \
    || { echo "FAIL — 「$1」 = $v 가 이 트리에서 커밋으로 해석되지 않는다" >&2; return 1; }
  echo "$v"
}
BASE_M3=$(read_sha 'M3 착수 직전') || { echo "AC-08 FAIL — 기준 SHA 없음"; exit 1; }
echo "BASE_M3=$BASE_M3"

# 비공허성 가드 — 「정리 대상 집합」이 실제로 이 범위 안에 있는지 확인한다.
# `.moai/specs` 전체를 세면 이 SPEC 자신의 산출물 4개만으로도 충족되므로(그것들은
# 아래에서 제외된다) 가드가 헐거워진다. 세는 대상을 정리 대상과 일치시킨다.
CLEANED=$(git diff --name-only "$BASE_M3"..HEAD -- '.moai/specs/*/plan.md' \
  '.moai/specs/*/acceptance.md' '.moai/specs/*/design.md' '.moai/specs/*/research.md' \
  | grep -v 'SPEC-ARTIFACT-STATELESS-001' | wc -l | tr -d ' ')
[ "$CLEANED" -gt 0 ] || { echo "AC-08 FAIL — $BASE_M3..HEAD 에 정리 대상 편집이 0건 (M3가 아직 착지하지 않았다)"; exit 1; }
echo "cleanup-target files in range = $CLEANED"

git diff --name-only "$BASE_M3"..HEAD -- .moai/specs \
  | grep -E '/(spec|progress)\.md$' \
  | grep -v 'SPEC-ARTIFACT-STATELESS-001' \
  && echo "AC-08 FAIL" || echo "no spec.md/progress.md touched: AC-08 PASS"
```

판정: `changed files under .moai/specs = <N>` (N > 0) **그리고** `no spec.md/progress.md touched: AC-08 PASS`.

> 비공허성 가드가 필요한 이유: 이 AC의 본체는 grep 미매치 시 `||` 분기로 PASS를 낸다. 빈 입력에도 grep은 미매치이므로, 가드가 없으면 **아무것도 검사하지 않고 PASS**한다. 종전 판이 정확히 그 상태였다.

**착지 전 실측 (`aacad4f99`, 빈 슬롯 상태)** — `FAIL — 「M3 착수 직전」 슬롯이 비어 있거나 7자리 미만이다` + `AC-08 FAIL — 기준 SHA 없음`, exit 1. (요구사항: REQ-AST-001-011)

---

## AC-AST-001-09 — lint와 정리가 같은 SPEC에서 함께 착지한다

**Given** era 예외를 두지 않기로 한 결정이 "정리와 lint의 동시 착지"에 의존할 때
**When** 이 SPEC의 종결 시점 트리에서 판정하면
**Then** 규칙이 등록되어 있고 **동시에** 전 코퍼스 D1 잔여가 0이다 — 둘 중 하나만 성립하면 FAIL.

```bash
count_d1() {
  local n=0 f
  for d in .moai/specs/SPEC-*/; do
    for a in plan acceptance design research; do
      f="${d}${a}.md"; [ -f "$f" ] || continue
      awk 'NR==1&&/^---/{p=1;next} p&&/^---/{exit} p' "$f" \
        | grep -qE '^status:[[:space:]]' && n=$((n+1))
    done
  done
  echo "$n"
}
R=$(grep -c 'ArtifactStatusFieldForbidden' internal/spec/lint.go)
D=$(count_d1)
echo "rule_matches=$R d1_remaining=$D"
[ "$R" -ge 1 ] && [ "$D" -eq 0 ] && echo PASS \
  || echo "FAIL — 둘이 갈라졌다면 era 예외가 필수가 된다"
```

판정: `PASS`.

**착지 전 실측 (`3b1830b96`)** — `rule_matches=0 d1_remaining=389` → FAIL. (요구사항: REQ-AST-001-006 / 010)

---

## AC-AST-001-10 — 스코프 밖 항목을 실제로 건드리지 않았다

**Given** `tier: 2`(7) / `tier: 3`(5) 12건과 백필이 스코프 밖일 때
**When** 이 SPEC의 전체 diff를 읽으면
**Then** tier 값을 바꾼 편집이 0건이고, 비-spec.md 산출물에 `status:`를 **추가**한 편집도 0건이다.

```bash
P=.moai/specs/SPEC-ARTIFACT-STATELESS-001/progress.md
read_sha() {
  local v
  v=$(sed -n "s/^| $1 | \`\([0-9a-f]\{7,\}\)\` |.*/\1/p" "$P")
  [ -n "$v" ] || { echo "FAIL — 「$1」 슬롯이 비어 있거나 7자리 미만이다" >&2; return 1; }
  git rev-parse --verify "$v^{commit}" >/dev/null 2>&1 \
    || { echo "FAIL — 「$1」 = $v 가 이 트리에서 커밋으로 해석되지 않는다" >&2; return 1; }
  echo "$v"
}
BASE_SPEC=$(read_sha 'SPEC 착수 직전') || { echo "AC-10 FAIL — 기준 SHA 없음"; exit 1; }
echo "BASE_SPEC=$BASE_SPEC"

# 비공허성 가드 — 이 AC의 판정은 "두 계수 모두 0"이라, 빈 diff 에서도 그대로 성립한다.
# 세는 대상을 정리 대상 집합과 일치시킨다(이 SPEC 자신의 산출물은 제외).
CLEANED=$(git diff --name-only "$BASE_SPEC"..HEAD -- '.moai/specs/*/plan.md' \
  '.moai/specs/*/acceptance.md' '.moai/specs/*/design.md' '.moai/specs/*/research.md' \
  | grep -v 'SPEC-ARTIFACT-STATELESS-001' | wc -l | tr -d ' ')
[ "$CLEANED" -gt 0 ] || { echo "AC-10 FAIL — $BASE_SPEC..HEAD 에 정리 대상 편집이 0건 (M3가 아직 착지하지 않았다)"; exit 1; }
echo "cleanup-target files in range = $CLEANED"

TIER=$(git diff "$BASE_SPEC"..HEAD -- .moai/specs | grep -E '^[+-]tier:' \
  | grep -v 'SPEC-ARTIFACT-STATELESS-001' | wc -l | tr -d ' ')
ADDS=$(git diff "$BASE_SPEC"..HEAD -- '.moai/specs/*/plan.md' '.moai/specs/*/acceptance.md' \
  '.moai/specs/*/design.md' '.moai/specs/*/research.md' | grep -cE '^\+status:' | tr -d ' ')
echo "tier edits: $TIER"
echo "status additions in non-spec artifacts: $ADDS"
[ "$TIER" -eq 0 ] && [ "$ADDS" -eq 0 ] && echo "AC-10 PASS" || echo "AC-10 FAIL"
```

판정: `changed files under .moai/specs = <N>` (N > 0) **그리고** 두 계수 모두 0 → `AC-10 PASS`.

**착지 전 실측 (`aacad4f99`, 빈 슬롯 상태)** — `FAIL — 「SPEC 착수 직전」 슬롯이 비어 있거나 7자리 미만이다` + `AC-10 FAIL — 기준 SHA 없음`, exit 1. 종전 판은 이 상태에서 두 계수 모두 0을 내며 **공허하게 PASS**했다. (요구사항: REQ-AST-001-012 / 013)

---

## AC-AST-001-11 — 템플릿 미러가 함께 갱신됐다

**Given** 미러가 실재하며 착수 시점에 바이트 동일할 때 (`3b1830b96`, 양쪽 23,317 bytes)
**When** M1이 착지하면
**Then** 두 파일이 여전히 바이트 동일하고, 앵커 소절이 **양쪽 모두**에 있다.

```bash
F=.claude/rules/moai/development/spec-frontmatter-schema.md
M="internal/template/templates/$F"
diff -q "$F" "$M" && echo "mirror identical: PASS" || echo "mirror identical: FAIL"
grep -qF '### Artifact Statelessness' "$F" && echo "anchor local:  PASS" || echo "anchor local:  FAIL"
grep -qF '### Artifact Statelessness' "$M" && echo "anchor mirror: PASS" || echo "anchor mirror: FAIL"
```

판정: 세 줄 모두 `PASS`.

**착지 전 실측 (`3b1830b96`)** — `mirror identical: PASS` / `anchor local: FAIL` / `anchor mirror: FAIL`. 미러 동일성은 이미 참이나 앵커 두 줄이 FAIL하므로 AC 전체 FAIL이며, 이 AC가 잡는 것은 **미러 드리프트**(M1이 한쪽만 고치는 경우)다. (요구사항: REQ-AST-001-015)

---

## 엣지 케이스

| 케이스 | 기대 동작 |
|---|---|
| frontmatter가 2행부터 시작하는 변칙 파일 | `NR==1&&/^---/` 판정에서 "frontmatter 없음"으로 계상 — 오분류 가능. Gap으로 기록(spec.md §5), 발견 시 개별 처리 |
| 본문에 `status:` 문자열이 있는 파일 | `---` 블록 밖이므로 대상 아님. AC-AST-001-07의 diff 판독으로 오삭제가 잡힌다 |
| `plan.md`는 있고 frontmatter가 없는 SPEC (다수) | lint 통과, 정리 대상 아님 |
| SPEC 디렉터리에 `spec.md`가 없는 2건 | 규칙이 spec.md 문서에서 형제를 유도하므로 스캔되지 않음. 기존 `discoverSPECs` 동작과 동일 — 새 결함 아님. 단 `count_d1`은 디렉터리를 순회하므로 이 2건의 산출물도 정리 대상에 포함된다 |
| 미종결 SPEC(draft·in-progress)의 산출물 27건 | **정리 대상이다**(전 코퍼스 모집단). 다른 레인이 작업 중일 수 있으므로 M3는 명시 pathspec 스테이징 + `git status --short` 재판독 필수 |
| 배포 사용자 코퍼스에 기존 위반 존재 | **이 리포에서 측정 불가 — Gap.** `plan.md` §D 리스크 표 및 §B3 2단계 대안 참조 |

---

## 품질 게이트

- `go vet ./internal/spec/...` 통과
- `go test ./internal/spec/...` 통과 (새 규칙의 양성·음성 양방향 테이블 테스트 포함)
- `golangci-lint run` — 새 코드에 신규 지적 0건
- `moai spec lint` rc=0 (정리 착지 후)
- **템플릿 미러 정합** — `diff -q` 무출력 (AC-AST-001-11), 그리고 `make build`

## Definition of Done

- [ ] AC-AST-001-01 ~ AC-AST-001-11 전부 PASS, 각 판정의 **실제 명령 출력을 인용**
- [ ] AC-AST-001-04의 심기→관측→원복 4단계가 실행되었고, `before` 매치 0 / `after` 매치 ≥1 / `after rc` 비-0 / 원복 후 빈 diff 가 **함께** 인용됨 (일부만 인용하면 미충족)
- [ ] AC-AST-001-01의 S1·S2·S3가 각각 인용됨 (합산 PASS 한 줄로 갈음 불가)
- [ ] **[인간 판독]** `### Artifact Statelessness` 소절 원문을 인용하고, S1·S2·S3의 각 리터럴이 **평서 단언**으로 쓰였음을(반례·인용·"이렇게 읽지 말 것" 문형 안이 아님을) 사람이 확인함 — 기계 검사는 같은 줄 부정어만 잡으므로 여러 줄에 걸친 부정은 이 항목이 유일한 방어다(spec.md §5 부채)
- [ ] `progress.md §E.1` 「기준값」 표의 세 슬롯(`SPEC 착수 직전` / `M3 착수 직전` / `M3 착수 시 D1 baseline N`)이 채워졌고, AC-07·08·10이 가드를 통과해 실행됨 — 빈 슬롯 상태의 출력을 PASS 근거로 인용하는 것은 미충족
- [ ] `progress.md §E.2`의 「기준 SHA」 표에 `SPEC 착수 직전` / `M3 착수 직전` 두 SHA가 기록됨
- [ ] 모든 계수에 측정 명령 + 트리 SHA가 병기됨 (`c6aa61346`의 362, `3b1830b96`의 389 리터럴을 판정값으로 재사용 금지)
- [ ] M2와 M3가 같은 SPEC 안에서 착지 (AC-AST-001-09)
- [ ] 품질 게이트 전 항목 통과
- [ ] `progress.md` §E.2 / §E.3에 증거 기록
