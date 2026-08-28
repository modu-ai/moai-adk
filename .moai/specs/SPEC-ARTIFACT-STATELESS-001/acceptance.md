# SPEC-ARTIFACT-STATELESS-001 — 인수 기준

> 모든 명령은 워크트리 루트에서, 프로젝트 루트 상대 경로로 실행한다.
> **baseline 귀속 규칙**: 이 문서의 판정 명령은 어떤 계수도 리터럴로 고정하지 않는다. `c6aa61346`의 362·417은 **참고값**이며, 판정은 실행 시점 HEAD에서 재측정한 값으로 한다(REQ-AST-001-009).
> **finding 코드 이름**: 새 lint 코드의 이름을 `ArtifactStatusFieldForbidden`으로 고정한다(AC-AST-001-03이 이 이름을 판정한다).

---

## AC-AST-001-01 — 규약이 spec.md 한정 + 무상태 + Tier 무관을 명시한다

**Given** `.claude/rules/moai/development/spec-frontmatter-schema.md`가 12필드 의무를 13행 괄호로만 암시하고 있을 때
**When** M1의 신설 소절이 착지하면
**Then** 그 문서는 (1) 의무 대상이 `spec.md` 1종임, (2) 나머지 4종이 무상태임, (3) 그 선언이 Tier와 무관함 — 셋을 모두 명시 문장으로 담는다.

```bash
F=.claude/rules/moai/development/spec-frontmatter-schema.md
grep -qiE 'spec\.md.*(only|한정|1종)' "$F" \
  && grep -qE 'plan\.md' "$F" && grep -qE 'acceptance\.md' "$F" \
  && grep -qE 'design\.md' "$F" && grep -qE 'research\.md' "$F" \
  && grep -qiE 'stateless|무상태' "$F" \
  && grep -qiE 'tier' "$F" \
  && echo PASS || echo FAIL
```

판정: `PASS`. (요구사항: REQ-AST-001-001 / 002 / 003)

---

## AC-AST-001-02 — 무상태의 정의가 status 축으로 한정된다

**Given** 무상태 선언이 "frontmatter 금지"로 읽힐 여지가 있을 때
**When** 규약 문구를 읽으면
**Then** 금지 대상이 `status:` 필드임이 명시되고, frontmatter 자체를 금지하는 문장은 없다.

```bash
F=.claude/rules/moai/development/spec-frontmatter-schema.md
grep -qE 'status:' "$F" && echo "status-axis: present" || echo "status-axis: FAIL"
# 반대 방향 — frontmatter 일체 금지 문구가 없어야 한다
grep -inE '(no|never|금지).{0,40}frontmatter' "$F" || echo "no blanket-frontmatter prohibition: PASS"
```

판정: `status-axis: present` + `no blanket-frontmatter prohibition: PASS`. (요구사항: REQ-AST-001-003)

---

## AC-AST-001-03 — 새 lint 규칙이 존재하고 era 예외에 들어가 있지 않다

**Given** `internal/spec/lint.go`가 규칙 배열에 규칙들을 등록하고 `eraDemotableCodes`로 grandfather 예외를 관리할 때
**When** M2가 착지하면
**Then** 새 코드 `ArtifactStatusFieldForbidden`이 규칙으로 등록되고, `eraDemotableCodes`에는 **없다**.

```bash
grep -n 'ArtifactStatusFieldForbidden' internal/spec/lint.go | head
echo "--- eraDemotableCodes block ---"
sed -n '/^var eraDemotableCodes/,/^}/p' internal/spec/lint.go
```

판정: 첫 grep이 1건 이상 매치하고, `eraDemotableCodes` 블록 출력에 `ArtifactStatusFieldForbidden`이 **나타나지 않는다**. (요구사항: REQ-AST-001-004 / 006)

---

## AC-AST-001-04 — [HARD] 심어서 거부를 관측한다 (H3)

**Given** lint 규칙이 작성되어 있으나 실제로 거부하는 것을 아직 보지 못한 상태에서
**When** 비-spec.md 산출물에 `status:` 필드를 의도적으로 심고 lint를 돌리면
**Then** finding 코드 `ArtifactStatusFieldForbidden`과 비-0 게이팅 결과를 **관측**하고, 심은 것을 원복한다.

규칙을 작성한 것은 규칙이 작동한다는 증거가 아니다. 아래를 **실행하고 출력을 인용**한다.

```bash
set -u
T=.moai/specs/SPEC-ARTIFACT-STATELESS-001/plan.md
cp "$T" /tmp/t357_ac004_backup.md

# 1) baseline — 심기 전에는 이 코드가 나오지 않아야 한다
moai spec lint > /tmp/t357_ac004_before.txt 2>&1; echo "before rc=$?"
grep -c 'ArtifactStatusFieldForbidden' /tmp/t357_ac004_before.txt

# 2) 심는다 — frontmatter 블록을 가진 파일 선두에 status 필드를 넣는다
printf -- '---\nstatus: draft\n---\n\n' | cat - "$T" > /tmp/t357_planted.md && mv /tmp/t357_planted.md "$T"

# 3) 관측한다
moai spec lint > /tmp/t357_ac004_after.txt 2>&1; echo "after rc=$?"
grep -n 'ArtifactStatusFieldForbidden' /tmp/t357_ac004_after.txt | head

# 4) 원복 — 반드시 실행한다
cp /tmp/t357_ac004_backup.md "$T"
git diff --stat -- "$T"
```

판정 (넷 모두 성립해야 PASS):

1. `before` 출력의 `ArtifactStatusFieldForbidden` 매치 수 = **0** (심기 전 공허하지 않음을 보증)
2. `after` 출력에 `ArtifactStatusFieldForbidden`이 **1건 이상** 나타난다
3. `after rc` 가 `before rc` 와 다르거나, 규칙 심각도가 error라면 `after rc` 가 **비-0**
4. 원복 후 `git diff --stat`이 **빈 출력** (심은 것이 남지 않음)

(요구사항: REQ-AST-001-007)

---

## AC-AST-001-05 — 통과 방향도 확인한다 (거짓 양성 없음)

**Given** frontmatter가 있으나 `status:`가 없는 산출물, `spec.md`, `progress.md`가 코퍼스에 존재할 때
**When** lint를 돌리면
**Then** 그 셋 중 어느 것에 대해서도 `ArtifactStatusFieldForbidden`이 나오지 않는다.

```bash
moai spec lint > /tmp/t357_ac005.txt 2>&1; echo "rc=$?"
# spec.md / progress.md 를 지목한 finding 이 있으면 FAIL
grep 'ArtifactStatusFieldForbidden' /tmp/t357_ac005.txt | grep -E 'spec\.md|progress\.md' && echo FAIL || echo "no spec.md/progress.md hits: PASS"
```

판정: `no spec.md/progress.md hits: PASS`. (요구사항: REQ-AST-001-005 / 011)

---

## AC-AST-001-06 — D1 정리가 완료되고, 재측정된 대상이 0으로 떨어진다

**Given** run-phase HEAD에서 재측정한 D1 대상 집합 N개가 있을 때 (`c6aa61346`의 362는 참고값일 뿐 판정값이 아니다)
**When** M3가 착지하면
**Then** 같은 측정을 재실행했을 때 대상이 **0**이 된다.

```bash
# 재측정 함수 — frontmatter 블록 안의 status: 만 센다
count_d1() {
  local n=0
  for d in .moai/specs/SPEC-*/; do
    for a in plan acceptance design research; do
      f="${d}${a}.md"; [ -f "$f" ] || continue
      awk 'NR==1&&/^---/{f=1;next} f&&/^---/{exit} f' "$f" \
        | grep -qE '^status:[[:space:]]' && n=$((n+1))
    done
  done
  echo "$n"
}
echo "HEAD=$(git rev-parse --short HEAD)"
echo "D1 remaining = $(count_d1)"
```

판정: `D1 remaining = 0`. 정리 **착수 전**에 같은 명령을 한 번 돌려 baseline N을 기록하고, 그 N과 HEAD SHA를 §E.2 증거에 남긴다. (요구사항: REQ-AST-001-008 / 009)

---

## AC-AST-001-07 — 정리가 status 라인만 건드렸다 (D1 준수, D2 아님)

**Given** D1은 `status:` 라인만, D2는 블록 전체를 지우는 정의일 때
**When** M3의 diff를 읽으면
**Then** 제거된 라인은 전부 `status:` 라인이며, `id:`/`title:`/`version:`/`created:` 라인은 **하나도 제거되지 않았다**.

```bash
BASE=<M3 착수 직전 커밋 SHA>
git diff "$BASE"..HEAD -- '.moai/specs/*/plan.md' '.moai/specs/*/acceptance.md' \
  '.moai/specs/*/design.md' '.moai/specs/*/research.md' \
  | grep '^-' | grep -v '^---' > /tmp/t357_removed.txt
echo "removed lines total: $(wc -l < /tmp/t357_removed.txt)"
echo "removed non-status lines:"
grep -vE '^-status:' /tmp/t357_removed.txt | head -20
grep -cvE '^-status:' /tmp/t357_removed.txt
```

판정: `removed non-status lines` 계수 = **0**. (요구사항: REQ-AST-001-008)

---

## AC-AST-001-08 — 정리가 spec.md / progress.md를 건드리지 않았다

**Given** REQ-AST-001-011이 두 파일을 대상에서 제외할 때
**When** M3의 diff를 읽으면
**Then** 변경 파일 목록에 `spec.md`·`progress.md`가 없다 (이 SPEC 자신의 산출물 제외).

```bash
BASE=<M3 착수 직전 커밋 SHA>
git diff --name-only "$BASE"..HEAD -- .moai/specs \
  | grep -E '/(spec|progress)\.md$' \
  | grep -v 'SPEC-ARTIFACT-STATELESS-001' \
  && echo FAIL || echo "no spec.md/progress.md touched: PASS"
```

판정: `no spec.md/progress.md touched: PASS`. (요구사항: REQ-AST-001-011)

---

## AC-AST-001-09 — lint와 정리가 같은 SPEC에서 함께 착지한다 (era 예외 미설정의 성립 조건)

**Given** era 예외를 두지 않기로 한 결정이 "정리와 lint의 동시 착지"에 의존할 때
**When** 이 SPEC의 종결 시점 트리에서 판정하면
**Then** 규칙이 등록되어 있고 **동시에** D1 잔여가 0이다 — 둘 중 하나만 성립하면 FAIL.

```bash
R=$(grep -c 'ArtifactStatusFieldForbidden' internal/spec/lint.go)
# AC-AST-001-06 의 count_d1 을 재사용
[ "$R" -ge 1 ] && [ "$(count_d1)" -eq 0 ] && echo PASS || echo "FAIL — 둘이 갈라졌다면 era 예외가 필수가 된다"
```

판정: `PASS`. (요구사항: REQ-AST-001-006 / 010)

---

## AC-AST-001-10 — 스코프 밖 항목을 실제로 건드리지 않았다

**Given** `tier: 2`(7) / `tier: 3`(5) 12건과 백필이 스코프 밖일 때
**When** 이 SPEC의 전체 diff를 읽으면
**Then** tier 값을 바꾼 편집이 0건이고, 비-spec.md 산출물에 `status:`를 **추가**한 편집도 0건이다.

```bash
BASE=<SPEC 착수 직전 커밋 SHA>
echo "tier edits:"
git diff "$BASE"..HEAD -- .moai/specs | grep -E '^[+-]tier:' | grep -v 'SPEC-ARTIFACT-STATELESS-001' | wc -l
echo "status additions in non-spec artifacts:"
git diff "$BASE"..HEAD -- '.moai/specs/*/plan.md' '.moai/specs/*/acceptance.md' \
  '.moai/specs/*/design.md' '.moai/specs/*/research.md' | grep -cE '^\+status:'
```

판정: 두 계수 모두 **0**. (요구사항: REQ-AST-001-012 / 013)

---

## 엣지 케이스

| 케이스 | 기대 동작 |
|---|---|
| frontmatter가 2행부터 시작하는 변칙 파일 | `NR==1&&/^---/` 판정에서 "frontmatter 없음"으로 계상 — 오분류 가능. Gap으로 기록(spec.md §5), 발견 시 개별 처리 |
| 본문에 `status:` 문자열이 있는 파일 | `---` 블록 밖이므로 대상 아님. AC-AST-001-07의 diff 판독으로 오삭제가 잡힌다 |
| `plan.md`는 있고 frontmatter가 없는 SPEC (다수) | lint 통과, 정리 대상 아님 |
| SPEC 디렉터리에 `spec.md`가 없는 2건 | 규칙이 spec.md 문서에서 형제를 유도하므로 스캔되지 않음. 기존 `discoverSPECs` 동작과 동일 — 새 결함 아님 |
| 배포 사용자 코퍼스에 기존 위반 존재 | **이 리포에서 측정 불가 — Gap.** plan.md §D 리스크 표 및 §B3 2단계 대안 참조 |

---

## 품질 게이트

- `go vet ./internal/spec/...` 통과
- `go test ./internal/spec/...` 통과 (새 규칙의 양성·음성 양방향 테이블 테스트 포함)
- `golangci-lint run` — 새 코드에 신규 지적 0건
- `moai spec lint` rc=0 (정리 착지 후)
- 템플릿 미러가 존재하면 `make build` 후 embed 정합 확인

## Definition of Done

- [ ] AC-AST-001-01 ~ AC-AST-001-10 전부 PASS, 각 판정의 **실제 명령 출력을 인용**
- [ ] AC-AST-001-04의 심기→관측→원복 3단계가 실행되었고, `before` 매치 0 / `after` 매치 ≥1 이 함께 인용됨 (한쪽만 인용하면 미충족)
- [ ] 모든 계수에 측정 명령 + 트리 SHA가 병기됨 (`c6aa61346` 리터럴 재사용 금지)
- [ ] M2와 M3가 같은 SPEC 안에서 착지 (AC-AST-001-09)
- [ ] 품질 게이트 전 항목 통과
- [ ] `progress.md` §E.2 / §E.3에 증거 기록
