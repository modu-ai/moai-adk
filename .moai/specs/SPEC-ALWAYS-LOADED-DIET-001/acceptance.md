# Acceptance Criteria — SPEC-ALWAYS-LOADED-DIET-001

## §A 판정 규율 (AC 실행 전 필독)

이 SPEC의 AC는 전부 **명령의 출력이 판정한다**. 아래 5개 함정은 과거 SPEC에서 공허한 GREEN 또는 부당한 FAIL을 실제로 만들어냈다. AC를 돌리기 전에 읽는다.

1. **명령을 표 셀에 넣지 않는다.** 마크다운 표 셀 안의 `\|` 는 `grep -E` 에서 리터럴 파이프로 해석돼 패턴이 조용히 아무 것도 매치하지 않는다. 이 문서의 모든 명령은 펜스 코드블록에 있다.
2. **패턴 언어와 파일 언어를 맞춘다.** `.claude/rules/`, `.claude/agents/`, `.claude/skills/` 본문은 **영어 전용**이다. 이 SPEC 본문은 한국어지만, 규칙 파일을 겨냥한 grep 패턴은 반드시 영어여야 한다.
3. **부재(absence) 판정은 양성 대조를 먼저 세운다.** "0건이면 통과" 형태의 AC는 패턴 자체가 고장나도 0건을 낸다. 부재 AC마다 같은 패턴이 분리 이전 파일에서는 매치되는지를 먼저 확인한다.
4. **가드 정규식을 면제 목록 없이 재구현하지 않는다.** `go test ./internal/config/ -run TestAlwaysLoaded` 의 실행 결과가 권위다. §A의 `headroom` 스니펫은 `fm()` 을 통과시켜 `paths:` 판정을 frontmatter 블록에 한정한다(가드의 `alwaysLoadedSurface` 계약과 동일한 스코프). 종전의 `grep -rL '^paths:'` 형태는 본문 어디든 `paths:` 로 시작하는 줄이 있으면 scoped 로 잘못 빼서 가드와 어긋날 수 있었다(CodeRabbit #1576 Major 1). 그래도 이 스니펫은 보조 계측이며, 둘이 어긋나면 **가드가 맞다**.
5. **`--strict` 는 severity 문자열이 아니라 exit code 를 바꾼다.** spec-lint 판정은 exit code 로 한다.
6. **부재 판정 앞에 존재 단언을 세운다.** `grep` 계열은 대상 파일이 없으면 에러를 stderr 로만 내고 stdout 에 아무 것도 출력하지 않는다. 그래서 "출력이 없으면 통과" 형태의 AC 는 **파일을 만들지 않은 것만으로** GREEN 이 되고, `grep -c` 는 `0` 이 아니라 **빈 문자열**을 내어 기계 판정을 미정의로 만든다. 파일을 다루는 AC 는 전부 `test -f` 를 먼저 통과시킨다.
7. **줄 위치로 필드를 뽑지 않는다.** `sed -n '2p'` 같은 위치 기반 추출은 frontmatter 키 순서가 조금만 달라도 무너진다. 실측: 기존 컴패니언 4개는 **전부** 2행이 `description:` 이다. 키는 이름으로 뽑는다(아래 `fm()`).

공통 전제: 모든 명령은 리포 루트에서 실행한다.

```bash
# frontmatter(첫 --- 블록)만 뽑는다. 본문에 나오는 --- 에 영향받지 않는다.
fm() { awk '/^---$/{c++; if(c==2) exit; next} c==1' "$1"; }
# 양성 대조(2026-08-16 실측): 기존 컴패니언 4개 전부 `fm "$f" | grep -c '^paths:'` = 1,
#   always-loaded 파일(kanban-dispatch.md)은 0.

# 이 문서 전반에서 쓰는 계측 함수 — scoped 판정은 fm() 결과에서만 한다(가드 계약 준수)
headroom() {
  tot=0; n=0
  for f in $(find .claude/rules/moai -type f -name '*.md' | sort); do
    fm "$f" | grep -q '^paths:' && continue
    tot=$((tot+$(wc -c < "$f"))); n=$((n+1))
  done
  for f in CLAUDE.md .claude/output-styles/moai/moai.md; do tot=$((tot+$(wc -c < "$f"))); done
  [ -f MEMORY.md ] && tot=$((tot+$(head -200 MEMORY.md | head -c 25600 | wc -c)))
  echo "files=$n bytes=$tot tokens=$((tot/4)) headroom=$((75000 - tot/4))"
}
```

Baseline (2026-08-16 실측; 2026-08-17 에 frontmatter-aware 형태로 재실행해 동일 출력 확인): `files=14 bytes=295044 tokens=73761 headroom=1239`

---

## §B AC 매트릭스

### AC-ALD-001 — 순감소 (구속 조건)

**Given** baseline 여유가 1,239 토큰이고
**When** M1~M4 완료 후 `headroom` 을 재실행하면
**Then** 출력의 `headroom` 값이 **1,239보다 엄격히 커야** 한다.

```bash
headroom
# PASS 조건: headroom > 1239
```

이 SPEC의 단일 종료 조건이다. M1(감소)과 M2(증가)를 개별 판정하지 않는다.

### AC-ALD-002 — 가드 통과 + `internal/config` 불변

**Given** always-loaded 표면이 변경된 상태에서
**When** 권위 있는 가드를 실행하고 `internal/config` 의 상태를 함께 재면
**Then** 가드는 exit 0 이고, 예산 상수는 75,000 그대로이며, Go 변경은 0건이어야 한다.

```bash
go test ./internal/config/ -run TestAlwaysLoaded; echo "exit=$?"
# D4-1: 예산 상수 불변 — 이 SPEC 안에서 AlwaysLoadedTokenBudget 을 움직이지 않는다
echo "budget_const=$(grep -c 'AlwaysLoadedTokenBudget = 75000' internal/config/token_budget_guard.go)"
# D4-2: M3 는 문서 전용 — Go 무변경
echo "go_changed=$(git diff --name-only origin/main...HEAD -- '*.go' | wc -l)"
# PASS 조건: exit=0, budget_const=1, go_changed=0
# baseline 관측(2026-08-16): ok github.com/modu-ai/moai-adk/internal/config, budget_const=1, go_changed=0
```

두 줄을 여기 접어 넣은 이유는 둘 다 `internal/config` 를 겨냥하고 있어 이 AC 의 판정 대상과 같기 때문이다. 별도 AC 로 세우면 17개가 되어 Tier M 상한(16)을 넘는다.

`git diff` 의 `origin/main...HEAD` 앵커는 **생략 불가**다. 인자 없는 `git diff` 는 워킹 트리와 비교하므로 작업을 커밋하는 순간 항상 비어, 아무 것도 판정하지 못하는 공허한 검사로 바뀐다.

`budget_const` 가 잡는 것: `AC-ALD-001` 의 `headroom` 함수는 셸 안에 `75000` 리터럴을 박고 있어 Go 상수가 바뀌어도 비교가 흔들리지 않는다. 뒤집으면 **상수 변경을 탐지하지도 못한다**는 뜻이고, 이 줄이 그 구멍을 막는다.

### AC-ALD-003 — 스텁이 BINDING-EVERYWHERE 구간 7개를 보유

**Given** `kanban-dispatch.md` 가 분리된 뒤
**When** 스텁에서 BINDING 구간 제목을 세면
**Then** 7개 제목이 모두 있어야 한다.

```bash
f=.claude/rules/moai/workflow/kanban-dispatch.md
for h in '^# Kanban Dispatch Protocol' \
         '^## Scope — when this rule is live' \
         '^## Completion is read, never trusted' \
         '^### CodeRabbit is not read from' \
         '^## Isolation is entered, never provisioned' \
         '^## Verification load is lane-local' \
         '^### The env-isolated verification form' \
         '^## Boundaries — what this protocol does not do' \
         '^## Cross-references'; do
  printf '%s => %s\n' "$h" "$(grep -c "$h" "$f")"
done
# PASS 조건: 9개 제목 전부 count=1 (7구간 + 하위 2개 제목)

# (b) LEAD-ONLY 6개 제목의 **부재** — 이동이 아니라 복사로 구현한 경우를 여기서 잡는다
for h in '^## The board' \
         '^## Entry into the board is an operator act' \
         '^## Card classes' \
         '^## The dispatch cycle' \
         '^## Review lens selection' \
         '^## The `/clear` handoff between phases'; do
  printf 'absent:%s => %s\n' "$h" "$(grep -c "$h" "$f")"
done
# PASS 조건: 6개 전부 count=0
# 양성 대조(함정 3): 분리 이전 원본에 같은 패턴을 돌리면 6개 전부 1 이어야 한다 —
#   2026-08-16 실측으로 확인함. 패턴이 살아있음을 본 뒤에야 0 을 근거로 삼는다.

# (c) REQ-ALD-001 이 지정한 Class B 문단이 스텁에 있고 컴패니언에 없는지
anchor='names that path in its completion report'
comp=.claude/rules/moai/workflow/kanban-dispatch-detail.md
echo "relocated_in_stub=$(grep -c "$anchor" "$f")"
if [ -f "$comp" ]; then
  echo "relocated_in_companion=$(grep -c "$anchor" "$comp")"
else
  echo "companion_MISSING=1"
fi
# PASS 조건: relocated_in_stub=1, relocated_in_companion=0, companion_MISSING 출력 없음
# 가드가 필요한 이유는 §A 함정 6 과 같다 — 컴패니언이 없으면 grep -c 는 0 이 아니라 빈 문자열을 낸다.
# 앵커 유일성(2026-08-16 실측): 이 문구는 .claude/rules/moai 전체에서 kanban-dispatch.md 에만, 1회 나온다.
```

**이 AC 는 보존 AC 다.** (a) 는 손대지 않은 트리에서도 통과하는 것이 정상이며 — 분리 이전 원본에 9개 제목이 다 있으므로 — 그 자체로는 진척 신호가 아니다. 진척은 `AC-ALD-005`(컴패니언 존재) · `AC-ALD-009`(바이트 합계) · `AC-ALD-001`(여유)이 판정한다. (b) 와 (c) 는 반대로 **분리 이전에는 반드시 실패**하므로, 이 AC 를 falsifiable 하게 만드는 쪽은 그 둘이다.

### AC-ALD-004 — 스텁의 HARD 절 보존

**Given** 분리 이전 원본이 `[HARD]` 9건을 갖고, 그중 6건이 STAY 구간·3건이 MOVE 구간에 있었으며
**When** 분리 후 스텁과 컴패니언에서 각각 세면
**Then** 스텁 6건 / 컴패니언 3건이어야 한다.

```bash
s=.claude/rules/moai/workflow/kanban-dispatch.md
c=.claude/rules/moai/workflow/kanban-dispatch-detail.md
for p in "$s" "$c"; do test -f "$p" || { echo "MISSING $p"; exit 1; }; done   # §A 함정 6
echo "stub=$(grep -c '\[HARD\]' "$s")"
echo "companion=$(grep -c '\[HARD\]' "$c")"
# PASS 조건: stub=6 companion=3 (합 9 = 분리 이전 총계)
```

이 AC가 R2(안전 관련 HARD 절 유실)를 잡는다. 합이 9가 아니면 절이 사라졌거나 생겼다.

### AC-ALD-005 — 컴패니언이 LEAD-ONLY 구간 6개를 보유

**Given** LEAD-ONLY로 분류된 6구간이 이동한 뒤
**When** 컴패니언에서 제목을 세면
**Then** 6개 제목(+ 하위 1개)이 모두 있어야 한다.

```bash
f=.claude/rules/moai/workflow/kanban-dispatch-detail.md
test -f "$f" || { echo "MISSING $f"; exit 1; }   # §A 함정 6
for h in '^## The board' \
         '^## Entry into the board is an operator act' \
         '^## Card classes' \
         '^## The dispatch cycle' \
         '^### Dispatch language' \
         '^## Review lens selection' \
         '^## The `/clear` handoff between phases'; do
  printf '%s => %s\n' "$h" "$(grep -c "$h" "$f")"
done
# PASS 조건: 7개 전부 count=1
```

### AC-ALD-006 — 내용 보존 (이동한 바이트가 사라지지 않음)

**Given** 분리 이전 원본의 비어있지 않은 줄이 140개이고
**When** 스텁과 컴패니언의 합집합과 대조하면
**Then** 원본의 모든 비어있지 않은 줄이 둘 중 한쪽에 있어야 한다(포인터 줄·frontmatter 등 **추가된** 줄은 허용).

```bash
c=.claude/rules/moai/workflow/kanban-dispatch-detail.md
test -f "$c" || { echo "MISSING $c"; exit 1; }   # §A 함정 6 — 가드가 없으면 분리 이전에도
                                                  # missing_lines=0 으로 통과한다(감사 iter2 D2)
BASE_REF=be1958a4d   # 분리 이전(plan-phase) 커밋 — HEAD 는 구현 착지 후 post-split 상태다
git show "${BASE_REF}:.claude/rules/moai/workflow/kanban-dispatch.md" \
  | grep -v '^[[:space:]]*$' | sort > /tmp/ald-orig.txt
cat .claude/rules/moai/workflow/kanban-dispatch.md \
    .claude/rules/moai/workflow/kanban-dispatch-detail.md \
  | grep -v '^[[:space:]]*$' | sort > /tmp/ald-after.txt
#   sort 에 -u 를 붙이지 않는다 — 중복 줄도 보존해 대조한다(comm 은 정렬만 요구)
comm -23 /tmp/ald-orig.txt /tmp/ald-after.txt
echo "missing_lines=$(comm -23 /tmp/ald-orig.txt /tmp/ald-after.txt | wc -l)"
# PASS 조건: missing_lines=0
# 양성 대조(함정 3): 위 comm 을 /dev/null 대상으로 돌리면 140에 가까운 값이 나와야 한다 —
#   패턴이 살아있음을 확인한 뒤 0을 근거로 삼는다.
```

### AC-ALD-007 — 컴패니언이 domain-keyed

**Given** self-keyed 컴패니언 3개와 domain-keyed 1개가 선례로 존재하고
**When** 새 컴패니언의 `paths:` 를 읽으면
**Then** 부모 규칙 파일 외에 실제 kanban 도메인 경로 2개를 함께 키로 가져야 한다.

```bash
c=.claude/rules/moai/workflow/kanban-dispatch-detail.md
test -f "$c" || { echo "MISSING $c"; }
p=$(fm "$c" | grep '^paths:')
echo "$p"
echo "manager-kanban_key=$(printf '%s' "$p" | grep -c 'manager-kanban.md')"
echo "todo_key=$(printf '%s' "$p" | grep -c 'workflows/todo.md')"
# PASS 조건: MISSING 출력 없음, 두 값 모두 1 (self-keyed 단독이면 둘 다 0 → FAIL)
```

**줄 위치로 뽑지 않는 이유**(실측된 거짓 실패): 종전 판정은 `sed -n '2p'` 로 2행이 `paths:` 라고 가정했는데, 리포의 컴패니언 **4개 전부** 2행이 `description:` 이다(`goal-directive-detail.md` / `session-handoff-examples.md` / `agent-common-protocol-reference.md` / `askuser-protocol-reference.md`). 그중 `goal-directive-detail.md` 는 plan.md D2 가 "유일한 domain-keyed 선례"로 지목하고 REQ-ALD-004 가 준수를 요구하는 바로 그 파일이다. 즉 종전 형태에서는 **선례를 정확히 따르면 이 AC 가 실패하고, 통과시키려면 선례를 어겨야 했다**. `fm()` 은 키를 이름으로 뽑으므로 순서에 무관하다.

### AC-ALD-008 — 선례 3요소 (포인터 열거 / 소유 선언 / 버전 기록)

**Given** `goal-directive.md` 선례가 세 요소를 확립했고
**When** 스텁과 컴패니언을 읽으면
**Then** 셋이 모두 있어야 한다.

```bash
stub=.claude/rules/moai/workflow/kanban-dispatch.md
comp=.claude/rules/moai/workflow/kanban-dispatch-detail.md
for p in "$stub" "$comp"; do test -f "$p" || { echo "MISSING $p"; exit 1; }; done   # §A 함정 6
# (a) 스텁 포인터가 옮긴 구간을 이름으로 열거
grep -n 'kanban-dispatch-detail.md' "$stub"
for n in 'board' 'Card classes' 'dispatch cycle' 'Review lens' 'clear' 'operator act'; do
  printf 'enumerated:%s => %s\n' "$n" "$(grep -c "$n" <(grep -A2 -B2 'kanban-dispatch-detail.md' "$stub"))"
done
# (b) 컴패니언이 자기 쪽에서 소유 경계 선언
echo "ownership_decl=$(grep -ciE 'detail companion|companion (to|for) `?kanban-dispatch' "$comp")"
# (c) 스텁 푸터 버전 줄에 분리 기록
tail -5 "$stub"
# PASS 조건: (a) 6개 구간명이 포인터 근방에 열거, (b) ownership_decl >= 1, (c) 푸터에 분리 문구 존재
```

**`owns` 를 대안에서 뺀 이유**(실측된 공허 GREEN): `owns` 는 분리 이전 원본 60행에 이미 들어 있다 — "The `run` session **owns** both the investigation and the fix."(`kanban-dispatch.md:60`). 종전 패턴 `'detail companion\|Detail Companion\|owns'` 를 원본에 돌리면 **1** 이 나오므로, 순수 이동만 해도 컴패니언이 자동으로 `owns` 를 얻어 소유 선언을 **한 글자도 쓰지 않고** 통과했다. REQ-ALD-004(b)가 요구하는 것을 전혀 검증하지 못한 셈이다.

이 결함은 D5(그 60행을 STAY 로 재배치)와 맞물려 있어 함께 푼다. 재배치 후에는 `owns` 가 컴패니언이 아니라 스텁에 남으므로 종전 패턴은 이번엔 **거짓 실패** 쪽으로 뒤집힌다 — 어느 쪽이든 판별력이 없다는 뜻이며, 그래서 앵커 자체를 선언 문구로 바꿨다.

양성 대조(함정 3, 2026-08-16 실측): 새 패턴을 선례 컴패니언 `goal-directive-detail.md` 에 `goal-directive` 앵커로 돌리면 **3**, 분리 이전 `kanban-dispatch.md` 에 `kanban-dispatch` 앵커로 돌리면 **0** 이다. 패턴이 살아 있고 이동만으로는 충족되지 않음을 둘이 함께 보인다.

### AC-ALD-009 — 컴패니언이 적치장이 되지 않음

**Given** 원본이 21,003 bytes 이고 오버헤드 추정이 600~900 bytes 이며
**When** 분리 후 두 파일의 합계를 재면
**Then** 21,003 이상 21,903 이하여야 한다.

```bash
s=.claude/rules/moai/workflow/kanban-dispatch.md
c=.claude/rules/moai/workflow/kanban-dispatch-detail.md
test -f "$s" || { echo "MISSING $s"; exit 1; }
test -f "$c" || { echo "MISSING $c"; exit 1; }
a=$(wc -c < "$s"); b=$(wc -c < "$c")
echo "stub=$a companion=$b sum=$((a+b))"
# PASS 조건: b >= 1 AND 21003 <= sum <= 21903
# 반례 참고: session-handoff-examples.md 는 40,891 로 부모 23,251 보다 크다.
#
# 존재 가드가 없으면 이 AC 는 무작업 상태에서 통과한다(감사 iter2 D1, 실행으로 확인).
# `wc -c` 가 없는 파일에 실패하면 b 가 빈 문자열이 되고 bash 산술이 이를 0 으로 읽어
# sum 이 원본 크기 21,003 이 되는데, 그 값이 PASS 범위의 하한과 정확히 일치한다.
# 경계의 우연이 아니라 구조적이므로 `b >= 1` 을 조건에 함께 건다. (§A 함정 6)
```

### AC-ALD-010 — G3~G7 다섯 규율이 인라인에 존재

**Given** 다섯 갭 모두 현재 매치 0건으로 실측되었고
**When** 편입 후 `cache-aware-execution.md` 를 읽으면
**Then** 다섯 주제어가 각각 최소 1회 등장해야 한다.

```bash
f=.claude/rules/moai/workflow/cache-aware-execution.md
printf 'G3_at_mention => %s\n' "$(grep -ci 'mention' "$f")"
printf 'G3_context_audit => %s\n' "$(grep -c '`/context`' "$f")"
printf 'G4_output_len => %s\n' "$(grep -c 'BASH_MAX_OUTPUT_LENGTH' "$f")"
printf 'G5_quiet => %s\n' "$(grep -ci 'quiet' "$f")"
printf 'G6_session_length => %s\n' "$(grep -ci 'session length\|one long session' "$f")"
printf 'G7_thinking => %s\n' "$(grep -c 'MAX_THINKING_TOKENS' "$f")"
# PASS 조건: 6개 값 전부 >= 1
# 양성 대조(함정 3): 편입 전 같은 명령은 6개 전부 0 이어야 한다 — 2026-08-16 실측으로 확인함.
#
# G3 패턴이 백틱을 포함하는 이유(실측된 오탐): 맨몸 `grep -c '/context'` 는
# 이 파일 27행의 cross-reference `workflow/context-window-management.md` 를 매치해
# 편입 전에도 1을 낸다. 즉 규율이 없는데도 GREEN 이 되는 공허한 AC 가 된다.
# 백틱 형태(`` `/context` ``)는 편입 전 이 파일 0건, `.claude/rules/moai` 전체 0건이다.
```

### AC-ALD-011 — 인라인 증가가 예산 안

**Given** `cache-aware-execution.md` 가 4,497 bytes 였고
**When** G3~G7 편입 후 재면
**Then** 증가분이 1,000 이상 2,000 이하여야 한다.

```bash
n=$(wc -c < .claude/rules/moai/workflow/cache-aware-execution.md)
echo "before=4497 after=$n delta=$((n-4497))"
# PASS 조건: 1000 <= delta <= 2000
```

상한 2,000 은 순감소 보호에 필요하다 — M2 의 증가가 M1 의 절감을 갉아먹는 양을 묶는다.

하한은 1,500 에서 **1,000 으로 낮췄다**. 종전 값의 목적은 "다섯 규율이 실제로 진술됐는지" 였는데, 그 판정은 이미 `AC-ALD-010` 의 6개 패턴이 하고 있어 중복이었고, 중복된 쪽이 하필 **순감소 방향(더 짧게 쓰기)을 벌하는** 형태였다. 진술 존재 판정은 `AC-ALD-010` 단독에 맡기고, 여기 하한은 "빈 껍데기 방지" 수준으로만 남긴다.

**구현자에게 — 이 창은 현행 지시 1~5 를 복사하면 넘친다.** 실측(2026-08-16, `cache-aware-execution.md`):

```
directive 1: 508 B   directive 2: 614 B   directive 3: 490 B
directive 4: 501 B   directive 5: 355 B   합계 2,468 B (평균 494 B)
```

같은 크기로 다섯 개를 더 쓰면 약 +2,470 B 로 **상한을 초과해 FAIL** 한다. 이는 창이 잘못됐다는 뜻이 아니라 `REQ-ALD-011` 이 요구하는 바 그대로다 — 새 지시는 **HARD 지시 본문만** 인라인이고 근거·수치·예시는 `cache-aware-execution-reference.md` 로 간다. 현행 1~5 는 근거를 문단 안에 품고 있어 494 B 인 것이고, 근거를 덜어낸 형태는 그보다 작아야 정상이다.

역산한 문단 예산: 창 [1,000, 2,000] 에 지시 5개 + 구분 빈 줄이면 **문단당 약 200~400 B**. plan.md M2 step 1 의 "현행 지시 1~5와 같은 형태"는 **형태(번호 매김 + `[ZONE:Evolvable] [HARD]` 접두 + 한 문단)를 뜻하지 크기를 뜻하지 않는다.**

### AC-ALD-012 — 신설 컴패니언 2개가 always-loaded 로 계상되지 않음

**Given** 신설 파일이 `kanban-dispatch-detail.md` 와 `cache-aware-execution-reference.md` 2개이고
**When** 가드가 세는 no-`paths:` 파일 수를 보면
**Then** 여전히 14개여야 한다(신설 2개는 `paths:` 보유).

```bash
headroom
for f in .claude/rules/moai/workflow/kanban-dispatch-detail.md \
         .claude/rules/moai/workflow/cache-aware-execution-reference.md; do
  if [ -f "$f" ]; then
    printf '%s exists=1 paths_key=%s\n' "$f" "$(fm "$f" | grep -c '^paths:')"
  else
    printf '%s exists=0 paths_key=0\n' "$f"
  fi
done
# PASS 조건: 2행 전부 exists=1 paths_key=1, 그리고 headroom 출력의 files=14
```

**`grep -L` 을 버린 이유**(실측된 공허 GREEN): `grep -L` 은 패턴이 없는 파일명을 내지만, **파일이 존재하지 않으면** 에러를 stderr 로만 내고 stdout 에는 아무 것도 내지 않는다. 종전 PASS 조건 "위 grep 이 아무 것도 출력하지 않음"은 그래서 **신설 파일 2개를 아예 만들지 않은 상태에서 충족**됐다. 2026-08-16 실측:

```
$ out=$(grep -L '^paths:' <존재하지 않는 파일 2개> 2>/dev/null); echo "len=${#out}"
len=0
```

`files=14` 쪽도 손대지 않은 트리에서 이미 14라, 두 조건이 함께 무작업 GREEN 을 냈다. 새 형태는 존재 단언을 앞세우고 `paths:` 키를 frontmatter **안에서** 직접 센다(§A 함정 6·7). 무작업 상태에서는 `exists=0` 으로 정확히 FAIL 한다 — 실행해 확인했다.

### AC-ALD-013 — 인용 수치가 인용으로 표기됨

**Given** 캐시 배수와 TTL이 원문 인용이고 이번 세션에 재측정되지 않았으며
**When** 신설 reference 컴패니언을 읽으면
**Then** 해당 수치 근방에 인용 출처 표기가 있어야 한다.

```bash
f=.claude/rules/moai/workflow/cache-aware-execution-reference.md
test -f "$f" || { echo "MISSING $f"; exit 1; }   # §A 함정 6
grep -n -i 'quoted\|not re-measured\|source article' "$f"   # 진단 출력
echo "numeric_paragraphs=$(awk -v RS= 'tolower($0) ~ /0\.1x|5x|ttl/ { n++ } END { print n+0 }' "$f")"
awk -v RS= '
  tolower($0) ~ /0\.1x|5x|ttl/ && tolower($0) !~ /quoted|not re-measured|source article/ {
    print "UNLABELED numeric paragraph:"; print; bad=1
  }
  END { exit bad ? 1 : 0 }
' "$f"; echo "citation_exit=$?"
# PASS 조건: numeric_paragraphs >= 1 (양성 대조, 함정 3) 그리고 citation_exit=0 —
#   0.1x / 5x / TTL 을 담은 문단 각각이 같은 문단(빈 줄로 구분된 블록) 안에 인용 표기를 동반
```

### AC-ALD-014 — 재발 통제가 존재하고 비-호출 세션 비용을 다룸

**Given** D1이 작성 규칙(확장안)을 선택했고 D4-3이 임계값을 1,000 bytes(단일 편집 기준)로 확정했으며
**When** 신설 통제 파일을 읽으면
**Then** `paths:` 스코프를 갖고, 신규 생성과 기존 파일 증가 **양쪽**을 다루며, 확정된 임계값을 적고, 비-호출 세션 비용을 명시해야 한다.

```bash
f=.claude/rules/moai/development/rule-authoring.md
test -f "$f" || { echo "MISSING $f"; exit 1; }   # §A 함정 6·7 — 존재 가드 + 줄 위치 추출 금지
p=$(fm "$f" | grep '^paths:')
echo "$p"
for frag in 'rules' 'CLAUDE.md' 'output-styles' 'MEMORY.md'; do
  printf 'slot_%s => %s\n' "$frag" "$(printf '%s' "$p" | grep -c "$frag")"
done
printf 'new_rule_case => %s\n' "$(grep -ci 'new always-loaded' "$f")"
printf 'growth_case => %s\n' "$(grep -ci 'grow\|increase' "$f")"
printf 'threshold => %s\n' "$(grep -cE '1,000 bytes|1000 bytes' "$f")"
printf 'non_invoking_cost => %s\n' "$(grep -ci 'never invoke\|does not invoke\|sessions that never' "$f")"
# PASS 조건: MISSING 없음, slot_ 4개(rules / CLAUDE.md / output-styles / MEMORY.md) 전부 >= 1
#   (REQ-ALD-013·plan.md D1 글롭의 가드 슬롯 4개 — rules-only 글롭이면 FAIL), 나머지 4개 전부 >= 1
```

`threshold` 가 판정하는 것은 값이 문서에 **적혔는지**뿐이다. 임계값 자체는 측정에서 유도한 값이 아니라 선택한 값이며(spec.md §5), 실제 발화 빈도는 이 AC의 판정 대상이 아니다.

### AC-ALD-015 — 템플릿 미러 패리티 + 빌드

**Given** always-loaded 14개는 이번 세션에 미러 존재를 확인했고
**When** 생성·수정된 파일 전부에 대해 미러를 대조하면
**Then** 모든 대응 파일이 존재하고 `make build` 가 성공해야 한다.

```bash
for rel in \
  .claude/rules/moai/workflow/kanban-dispatch.md \
  .claude/rules/moai/workflow/kanban-dispatch-detail.md \
  .claude/rules/moai/workflow/cache-aware-execution.md \
  .claude/rules/moai/workflow/cache-aware-execution-reference.md \
  .claude/rules/moai/development/rule-authoring.md ; do
  m="internal/template/templates/$rel"
  printf '%s => %s\n' "$rel" "$(test -f "$m" && echo MIRRORED || echo MISSING)"
done
make build; echo "build_exit=$?"
# PASS 조건: 5개 전부 MIRRORED, build_exit=0
```

미러는 **byte-parity 가 아닐 수 있다**(sanitized-pair). 통째 `cp` 금지 — 존재와 중립성만 판정한다.

### AC-ALD-016 — 미러 중립성

**Given** 템플릿은 16개 언어 사용자에게 중립이어야 하고
**When** 미러본에서 내부 전용 토큰을 찾으면
**Then** 매치가 0건이어야 한다.

```bash
for rel in \
  .claude/rules/moai/workflow/kanban-dispatch.md \
  .claude/rules/moai/workflow/kanban-dispatch-detail.md \
  .claude/rules/moai/workflow/cache-aware-execution.md \
  .claude/rules/moai/workflow/cache-aware-execution-reference.md \
  .claude/rules/moai/development/rule-authoring.md ; do
  m="internal/template/templates/$rel"
  test -f "$m" || { printf '%s MISSING\n' "$rel"; continue; }
  printf '%s spec_id=%s req=%s sha=%s date=%s\n' "$rel" \
    "$(grep -cE 'SPEC-[A-Z][A-Z0-9-]*-[0-9]{3}' "$m")" \
    "$(grep -cE 'REQ-[A-Z]+-[0-9]{3}' "$m")" \
    "$(grep -cE '\b[0-9a-f]{9,40}\b' "$m")" \
    "$(grep -cE '20[0-9]{2}-[0-9]{2}-[0-9]{2}' "$m")"
done
MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ -run TestTemplateNoInternalContentLeak; echo "exit=$?"
# PASS 조건: MISSING 0건, 5행 전부 네 값이 0, 그리고 leak 테스트 exit=0
# 양성 대조(함정 3): 같은 패턴을 로컬 .moai/specs/ 파일에 돌리면 0이 아니어야 한다.
# 오탐 없음 확인(2026-08-16): 편집 예정 미러 2개(kanban-dispatch.md, cache-aware-execution.md)에
#   위 네 패턴을 돌려 전부 0 을 관측했다 — 이 AC 는 거짓 실패 기계가 아니다.
```

`test -f` 가드가 필요한 이유(실측): 5개 대상 중 **3개는 아직 미러가 없다**(`kanban-dispatch-detail.md` / `cache-aware-execution-reference.md` / `rule-authoring.md`). 존재하지 않는 파일에 `grep -cE` 를 걸면 stdout 이 비어 `printf` 가 `spec_id= req= sha= date=` 를 출력한다. 빈 값은 `0` 이 아니므로 눈으로는 잡히지만 기계 판정으로는 미정의다. `AC-ALD-015` 가 존재를 따로 확인하긴 해도 두 AC 사이에 기계적 의존이 없어, 016 만 단독 실행하면 조용히 넘어간다(§A 함정 6).

---

## §C 추적 매트릭스

| REQ | AC |
|---|---|
| REQ-ALD-001 | AC-ALD-003, AC-ALD-004 |
| REQ-ALD-002 | AC-ALD-005, AC-ALD-006 |
| REQ-ALD-003 | AC-ALD-007 |
| REQ-ALD-004 | AC-ALD-008 |
| REQ-ALD-005 | AC-ALD-009 |
| REQ-ALD-006 ~ REQ-ALD-010 | AC-ALD-010 |
| REQ-ALD-011 | AC-ALD-011, AC-ALD-012 |
| REQ-ALD-012 | AC-ALD-013 |
| REQ-ALD-013 | AC-ALD-014, AC-ALD-002 (Go 무변경 하위절) |
| REQ-ALD-014 | AC-ALD-014 |
| REQ-ALD-015 | AC-ALD-001, AC-ALD-002 |
| REQ-ALD-016 | AC-ALD-015, AC-ALD-016 |

## §D Definition of Done

- [ ] AC-ALD-001 ~ AC-ALD-016 전부 PASS, 각각 실행 출력을 근거로 인용
- [ ] `moai spec lint .moai/specs/SPEC-ALWAYS-LOADED-DIET-001/spec.md` exit 0 (디렉터리 인자 불가 — 파일 경로로 호출)
- [ ] 세 산출물에 미해결 표식 잔존 0건 — `grep -rc 'NEEDS CLARIFICATION' .moai/specs/SPEC-ALWAYS-LOADED-DIET-001/*.md` 를 돌려 `spec.md` 와 `plan.md` 가 0 이고, `acceptance.md` 는 이 항목이 문자열 자체를 언급하므로 1 이어야 한다(1을 넘으면 다른 곳에 표식이 남은 것)
- [ ] 커밋 제목이 `feat(SPEC-ALWAYS-LOADED-DIET-001):` 접두를 사용
- [ ] feature 브랜치 + PR 로 착지(리포 로컬 정책상 전 Tier PR 필수)
- [ ] `.moai/specs/SPEC-ALWAYS-LOADED-DIET-001/progress.md` §E.2/§E.3 에 run-phase 근거 기록

## §E 잔여 위험

- `paths:` 부착은 Claude Code 런타임 소관이라 관측할 수 없다. AC-ALD-007/012는 **글롭 문자열의 형태**를 판정할 뿐 실제 부착을 판정하지 못한다. 최악의 실패 모드는 "컴패니언이 안 붙음"이고, 스텁이 모든 구속 절을 갖고 있으므로 안전 실패다.
- 토큰 수치는 char/4 추정(±약 15%)이다. AC-ALD-001의 `headroom > 1239` 비교는 **같은 공식의 전후 비교**이므로 추정 오차가 상쇄되지만, 절대값을 실제 토큰 수로 읽어서는 안 된다.
- AC-ALD-006의 원본 판독은 `BASE_REF=be1958a4d`(분리 이전 plan-phase 커밋)로 고정한다. 구현 착지 이후 어느 시점에 재실행해도 HEAD 가 아니라 이 SHA 를 읽으므로 pre-split 원본이 나온다.
