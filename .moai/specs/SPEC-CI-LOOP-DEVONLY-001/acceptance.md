# acceptance — SPEC-CI-LOOP-DEVONLY-001

모든 판정 명령은 워크트리 루트에서 실행한다.

**baseline 정직성 고지 (감사 D3 반영).** 초판은 "모든 baseline을 실행했다"고 서술했으나
AC-CLD-010의 선택자는 실행되지 않은 채 존재하지 않는 테스트명(`TestInternalContentLeak`)을
지목하고 있었다. 이 판은 각 AC마다 baseline의 **출처를 개별 표기**한다:

- `[실행됨]` — 이 판을 쓰면서 실제로 명령을 돌려 얻은 값
- `[변이검증됨]` — 실행에 더해, 조건을 인위로 깨뜨려 **실패하는 것까지** 확인한 값
- `[미존재]` — 아직 만들지 않은 가드. 통과 기준이 "존재 + PASS"이므로 공허 통과 불가

---

## REQ ↔ AC 커버리지 (감사 D9 반영)

| REQ | AC | REQ | AC |
|---|---|---|---|
| REQ-CLD-001 | AC-CLD-001 | REQ-CLD-011 | AC-CLD-009, AC-CLD-011 |
| REQ-CLD-002 | AC-CLD-002 | REQ-CLD-012 | AC-CLD-005, AC-CLD-009 |
| REQ-CLD-003 | AC-CLD-003 | REQ-CLD-013 | AC-CLD-009, AC-CLD-010 |
| REQ-CLD-004 | AC-CLD-002 | REQ-CLD-014 | AC-CLD-012 |
| REQ-CLD-005 | AC-CLD-005 | REQ-CLD-015 | AC-CLD-013 |
| REQ-CLD-006 | AC-CLD-004 | REQ-CLD-016 | AC-CLD-001, 002, 005, 017 |
| REQ-CLD-007 | AC-CLD-007 | REQ-CLD-017 | AC-CLD-015 |
| REQ-CLD-008 | AC-CLD-006 | REQ-CLD-018 | AC-CLD-019 |
| REQ-CLD-009 | AC-CLD-008 | REQ-CLD-019 | AC-CLD-014 |
| REQ-CLD-010 | AC-CLD-008 | (결정 B 검증) | AC-CLD-016 |
| (§4 범위 제외 가드) | AC-CLD-018 | | |

**정정 2건 (감사 N9 반영)** — 초판 표는 두 곳을 과대 주장했다.

- **REQ-CLD-016**(dev-only 자산 흔적이 템플릿에 없을 것)을 AC-CLD-014(중립성)에 걸었으나,
  그 AC는 **내부 콘텐츠 누출**(SPEC ID·날짜·커밋 해시)을 판정할 뿐 dev-only 자산 흔적은
  판정하지 않는다. 실제 판정자는 AC-CLD-001·002·005(템플릿 참조 0)와 AC-CLD-017(미러)이다.
- **REQ-CLD-018**(사용자 저작물 미삭제)을 AC-CLD-015에 걸었으나, 그 AC의 명령은
  **목록 등재 여부**와 **분리 테스트**만 관측한다 — 삭제가 아니라 아카이브로 동작하는지는
  어느 명령도 관측하지 않는다. 초판 주석은 "아카이브 경로를 택했으므로 안전하다"는
  **채택한 메커니즘으로부터의 논증**이었지 시험이 아니다.
  → **해소됨 (v0.3.0, AC-CLD-019).** 경위를 기록해 둔다. 초판이 적은 이연 사유
  ("통합 시험이 필요하다")는 부정확했다 — `archiveSkill(projectRoot, skillID string)`
  (`internal/cli/update_archive.go`, 내용 앵커 `func archiveSkill`)은 평범한 루트 경로를
  받으므로 `t.TempDir()`로 **단위 시험이 가능하다**. 실제 이연 사유는 "시험이 불가능해서"가
  아니라 **이 SPEC의 범위를 배포 격리로 한정했기 때문**이었다. run-phase 진입 시점에
  GOOS가 승격을 결정해 AC-CLD-019로 채택했다. 부채가 아니라 판정 대상이다.

**AC-CLD-018의 REQ 귀속 (해당 없음 — 명시)**: AC-CLD-018은 어떤 REQ도 직접 판정하지
않는다. 이것은 요구사항 검증이 아니라 **범위 가드**다 — spec.md §4 "Out of Scope —
CLI 동작 변경"이 선언한 불변 조건(플래그 집합·종료 코드·`--abort`/`--report` 동작)이
M6의 문자열 편집으로 훼손되지 않았음을 관측한다. REQ-CLD-008(부재 스크립트 안내 제거)의
**편집 봉투를 지키는** 짝이지만 REQ-CLD-008 자체는 AC-CLD-006이 판정하므로,
표에 REQ↔AC 행을 만들지 않고 별도 행으로 표기했다. 없는 매핑을 지어내지 않는다.

미커버 REQ: **0건**.

---

## AC-CLD-001 — 템플릿 트리에 스크립트 경로 참조가 없다

**Given** 템플릿 트리가 미배포 셸 스크립트를 참조하고
**When** 배포 격리가 완료되면
**Then** 참조가 0건이다.

```bash
grep -rn 'scripts/ci-watch\|scripts/ci-autofix' internal/template/templates/ | wc -l
```

- **baseline `[실행됨]`**: `27`
- **통과 기준**: `0`

---

## AC-CLD-002 — 템플릿 트리와 카탈로그에 스킬 식별자가 없다

**Given** 9개 항목이 `moai-workflow-ci-loop`를 참조하고
**When** 스킬 배포가 중단되면
**Then** 템플릿 트리와 `catalog.yaml` 어디에도 식별자가 없다.

```bash
grep -rln 'moai-workflow-ci-loop' internal/template/templates/ internal/template/catalog.yaml | wc -l
grep -c 'name: moai-workflow-ci-loop' internal/template/catalog.yaml || true
```

- **baseline `[실행됨]`**: `9` / `1`
- **통과 기준**: `0` / `0`

---

## AC-CLD-003 — 컴파일된 임베드 FS에 스킬이 없다 (배포 산출물 판정)

**Given** 스킬이 `//go:embed all:templates`로 바이너리에 컴파일되어 있고
**When** 템플릿에서 제거되면
**Then** `EmbeddedTemplates()` FS에 해당 스킬 하위 파일이 0개다.

```bash
go test ./internal/template/ -run TestEmbeddedTemplatesExcludeCILoopSkill -count=1 -v 2>&1 \
  | grep -q '^--- PASS: TestEmbeddedTemplatesExcludeCILoopSkill'; echo "exit=$?"
```

- **baseline `[변이검증됨]`** — 존재/부재 양쪽 실행:
  ```
  현재 (테스트 부재)                     exit=1   ← 올바르게 거부
  대조군 TestSplitHarnessNamespaceNoLeak  exit=0   ← 존재+PASS를 올바르게 수용
  ```
  참고로 초판 명령(`>/dev/null; echo $?`)은 **테스트가 없어도 `exit=0`** 이었다.
- **통과 기준**: `exit=0`
- **주석 (감사 P1 반영 — 공허 선택자 3회차)**:
  초판은 통과 조건을 산문으로 "테스트가 존재하고 PASS"라고 적었지만 명령은 둘 중
  **어느 것도 단언하지 않았다.** `go test -run`은 0매칭이어도 성공으로 끝나기 때문이다.
  같은 계열의 결함이 iteration-1(존재하지 않는 테스트명), iteration-2(판정이 산문에만 존재)에
  이어 세 번째였다. `grep -q '^--- PASS: <name>'`은 테스트가 실제로 **실행되어 통과했음**을
  출력에서 확인하므로 존재와 통과를 한 번에 단언한다.
  acceptance.md 판정 규율의 "명령에 없는 판정은 없다"를 이 AC에도 적용한 결과다.
- **이 AC가 판정하는 것 / 판정하지 못하는 것 (감사 D3 반영)**:
  소스 트리 grep(AC-CLD-001/002)과 달리 **배포 산출물인 임베드 FS**를 읽으므로
  디렉터리가 남아 있으면 실패한다. **그러나 `make build` 누락은 탐지하지 못한다** —
  `go test`가 컴파일을 수행하므로 `//go:embed`는 항상 현재 소스를 반영하기 때문이다.
  초판이 이 AC에 붙였던 "재빌드 누락 탐지" 근거는 **거짓이며 변이 테스트로 반증되었다**.
  재빌드 누락은 AC-CLD-004가 담당한다.

---

## AC-CLD-004 — catalog 해시가 최신이다 (재빌드 누락 탐지)

**Given** `make build`가 `gen-catalog-hashes.go --all`로 해시를 재생성하고
**When** 템플릿을 편집한 뒤 재빌드를 누락하면
**Then** 생성기를 돌렸을 때 `catalog.yaml`에 diff가 발생한다.

```bash
go run ./internal/template/scripts/gen-catalog-hashes.go --all >/dev/null 2>&1; echo "gen-exit=$?"
git diff --exit-code -- internal/template/catalog.yaml >/dev/null 2>&1; echo "diff-exit=$?"
```

- **baseline `[변이검증됨]`** — 이 SPEC이 실제로 수행하는 **삭제** 상태로 재검증:
  ```
  STATE A (clean)                       gen-exit=0  diff-exit=0
  STATE B (스킬 디렉터리 삭제, M3가 하는 일)  gen-exit=1  diff-exit=0
                                        catalog.yaml 미변경, 낡은 해시 3761e843… 잔존
                                        생성기 로그: "1 entries failed hash computation"
  ```
  변이 후 `git checkout`으로 복원해 `git status --porcelain`이 SPEC 디렉터리
  하나만 남는 것을 확인했다.
- **통과 기준**: `gen-exit=0` **그리고** `diff-exit=0` (두 조건 모두)
- **판정 시점 전제 (감사 P7)**: `git diff --exit-code`는 커밋 상태에 의존한다.
  M3·M4 편집을 **커밋한 뒤** 판정해야 한다 — 커밋 전 완성 상태에서는 생성기가 멱등해도
  HEAD 대비 미커밋 변경 때문에 `diff-exit=1`이 된다. 파탄 상태에서 통과할 수는 없으므로
  구멍은 아니지만, 판정 순서를 지키지 않으면 거짓 실패가 난다.
- **주석 (감사 N1 반영 — 초판 게이트의 치명적 공백)**:
  초판은 `git diff --exit-code`만 조건으로 삼았고, **내용 변경**으로만 변이 테스트했다.
  그러나 M3가 수행하는 것은 **디렉터리 삭제**이며, 그 상태에서 생성기는 해시를 계산할 수
  없는 항목의 재작성을 **거부**한다 — 결과적으로 `catalog.yaml`이 손대지 않은 채 남아
  `diff-exit=0`이 되어 **가장 위험한 상태에서 트리가 깨끗해 보인다.**
  판별자는 생성기 자신의 종료 코드(`gen-exit`)다: clean `0`, 삭제 상태 `1`.
- **탐지 한계 (명시)**: 해시는 공백을 정규화하므로 **공백만 추가한 변경은 `gen-exit=0`,
  `diff-exit=0`으로 통과한다.** 이 게이트는 의미 있는 내용 변경과 삭제를 탐지하며,
  공백 전용 변경은 대상이 아니다.

---

## AC-CLD-005 — ci-watch-protocol.md가 템플릿에서 제거되었다

**Given** 배포되는 어떤 산출물도 watch 루프를 수행하지 않고 (research.md §B)
**When** 결정 A에 따라 파일이 배포에서 제거되면
**Then** 템플릿 트리에 해당 파일이 없다.

```bash
test ! -e internal/template/templates/.claude/rules/moai/workflow/ci-watch-protocol.md; echo $?
```

- **baseline `[실행됨]`**: 파일 존재 (277행) → 명령은 `1`
- **통과 기준**: `0`

---

## AC-CLD-006 — 배포 바이너리가 부재 스크립트 실행을 지시하지 않는다

**Given** `pr_watch_cmd.go`가 사용자에게 `scripts/ci-watch/run.sh` 실행을 지시하는
문자열 3개를 보유하고
**When** 문자열이 교정되면
**Then** 주석을 제외한 사용자 노출 문자열에 해당 경로가 남지 않는다.

```bash
grep -n 'scripts/ci-watch/run.sh' internal/cli/pr_watch_cmd.go | grep -vc ':[[:space:]]*//'
```

- **baseline `[실행됨]`**: `3` — 매치 원문:
  ```
  41:The CI watch loop is invoked via scripts/ci-watch/run.sh.
  72:  fmt.Fprintf(os.Stderr, "[ci-watch] Use scripts/ci-watch/run.sh to start the watch loop.\n")
  73:  fmt.Fprintf(os.Stderr, "[ci-watch] Example: MOAI_CIWATCH_GH=gh sh scripts/ci-watch/run.sh %s %s\n",
  ```
- **통과 기준**: `0`
- **주석**: 107행은 주석이라 선택자가 제외한다(원시 매치는 4건). 이 AC가 없으면
  다른 모든 AC가 통과해도 `moai pr watch` 사용자는 여전히 부재 스크립트를 안내받는다.

---

## AC-CLD-007 — 배포 규칙 트리가 `moai pr watch`를 전혀 언급하지 않는다

**Given** CLI는 watch 루프를 수행하지 않으므로 배포 규칙이 이 명령을 언급할 정당한 이유가 없고
**When** M2·M3가 완료되면
**Then** 템플릿 트리에 `moai pr watch` 문자열이 남지 않는다.

```bash
grep -rn 'moai pr watch' internal/template/templates/ | wc -l
```

- **baseline `[실행됨]`**: `5` — 분포:
  ```
  1  .claude/rules/moai/core/zone-registry.md              (M2가 제거)
  3  .claude/rules/moai/workflow/ci-watch-protocol.md      (M3가 파일 삭제)
  1  .claude/skills/moai-workflow-ci-loop/SKILL.md         (M3가 디렉터리 삭제)
  ```
- **통과 기준**: `0`
- **주석 (감사 N10 반영 — 선택자 전면 교체)**:
  초판 선택자는 `moai pr watch`와 키워드가 **같은 줄**에 있을 것을 요구했다. 변이 테스트로
  두 가지가 드러났다.
  ```
  baseline                        count=1   (ci-watch-protocol.md 단독 — zone-registry.md는
                                             문구는 있으나 키워드가 없어 미매치)
  금지 문구 한 줄로 재삽입          count=2   (탐지됨)
  같은 문구를 두 줄로 줄바꿈          count=1   (탐지 실패)
  ```
  이 SPEC 문서 자체가 95자 부근에서 줄바꿈하므로 **줄바꿈된 형태가 오히려 흔한 형태**다.
  또한 초판 baseline은 매치 파일을 2개로 적었으나 실제로는 1개다(과대 주장).
  덧붙여 유일한 매치가 AC-CLD-005가 삭제하는 파일 안에 있어, M3 이후 이 선택자는
  **자동으로 통과하며 독립적 신호를 전혀 내지 않았다.**
  교체 선택자는 인접성을 요구하지 않으므로 줄바꿈에 영향받지 않고, 세 파일에 걸친
  5개 매치를 baseline으로 가지므로 M2·M3 각각의 누락을 개별적으로 탐지한다.

---

## AC-CLD-008 — autofix 안전 규칙이 보존되고 스크립트 트리거가 제거되었다

**Given** `cycle_type=autofix`가 실재하는 `manager-develop` 사이클이고
**When** `ci-autofix-protocol.md`가 재작성되면
**Then** 파일은 남고, Frozen 절 10개는 보존되며, 스크립트 경로는 사라진다.

```bash
test -e internal/template/templates/.claude/rules/moai/workflow/ci-autofix-protocol.md && echo FILE_OK
grep -c 'ZONE:Frozen' internal/template/templates/.claude/rules/moai/workflow/ci-autofix-protocol.md
grep -c 'scripts/ci-watch\|scripts/ci-autofix' internal/template/templates/.claude/rules/moai/workflow/ci-autofix-protocol.md || true
```

- **baseline `[실행됨]`**: `FILE_OK` / `10` / (스크립트 참조 존재)
- **통과 기준**: `FILE_OK` / `10` (감소 불가) / `0`
- **담당 마일스톤**: M3 7단계 (`ci-autofix-protocol.md` 재작성) — 감사 P5 반영
- **주석**: Frozen 절 수가 10 미만이면 안전 규칙이 삭제된 것이다 — 이 SPEC은
  `004..013`을 재작성하되 **제거하지 않는다**.

---

## AC-CLD-009 — 헌법 검증에서 ci 귀속 findings가 0이다

**Given** ci 귀속 findings 18건이 존재하고
**When** 소스와 레지스트리가 정합화되면
**Then** 두 프로토콜 파일에 귀속된 findings가 0건이다.

```bash
MOAI_CONSTITUTION_REGISTRY=.claude/rules/moai/core/zone-registry.md \
  moai constitution validate > /tmp/cv.txt 2>&1; echo "validate-exit=$?"
grep -E '^\s+\[(DRIFT|SOURCE_FILE_MISSING)\]' /tmp/cv.txt \
  | grep -c 'ci-autofix-protocol.md\|ci-watch-protocol.md'
```

- **baseline `[변이검증됨]`**:
  ```
  STATE A (clean)                              validate-exit=1  ci-count=18  total=77
  STATE B (소스 삭제 + 레지스트리 항목 잔존)       validate-exit=2  ci-count=11  total=49
  ```
- **통과 기준**: `validate-exit=1` **그리고** ci-count `0` (두 조건 모두)
- **주석 (감사 N8 반영 — 치명적 오독)**:
  `SOURCE_FILE_MISSING`은 **fatal(exit 2)** 이라 검증이 중단되고 출력이 잘린다.
  초판은 `grep -c`만 조건으로 삼았으므로, **가장 위험한 부분 완료 상태에서 오히려
  더 작은 숫자를 보고**했다(18→11). 중단 지점에 따라서는 `0`이 나올 수도 있는데,
  그것이 바로 초판의 통과값이었다. `validate-exit`이 판별자다 — clean `1`, 파탄 `2`.
- 감사가 워크트리 소스 빌드 바이너리로 재측정해 릴리스 바이너리와 동일 출력을
  확인했다 — 엔진 버전 차이 없음(research.md §G.2).

---

## AC-CLD-010 — 전체 findings가 정확히 18건 감소한다

**Given** 전체 77건 중 18건이 ci 귀속이고
**When** ci 귀속분만 해소되면
**Then** 전체는 정확히 59건이다.

```bash
MOAI_CONSTITUTION_REGISTRY=.claude/rules/moai/core/zone-registry.md \
  moai constitution validate > /tmp/cv.txt 2>&1; echo "validate-exit=$?"
grep -c 'DRIFT\]\|SOURCE_FILE_MISSING\]' /tmp/cv.txt
```

- **baseline `[변이검증됨]`**:
  ```
  STATE A (clean)                              validate-exit=1  total=77
  STATE B (소스 삭제 + 레지스트리 항목 잔존)       validate-exit=2  total=49
  ```
- **통과 기준**: `validate-exit=1` **그리고** total `59` (등호)
- **주석 (감사 N8 반영 — 초판 주석의 원인 진단이 틀렸다)**:
  초판은 "59 미만 = 봉투 밖 수정"이라고 적었다. 실제 지배적 원인은 **fatal abort로 인한
  출력 절단**이다 — STATE B에서 49가 나온 것은 봉투 밖을 건드려서가 아니라
  검증이 중간에 죽었기 때문이다. `validate-exit`을 함께 보지 않으면 파탄을 개선으로 읽는다.
  59 초과는 여전히 새 드리프트 유입을 뜻한다.

---

## AC-CLD-011 — canary_gate 감소가 정확히 8건이다

**Given** 레지스트리에 `canary_gate: true` 73건이 있고 그중 `014..021` 8건이 삭제 대상이며
**When** M2가 완료되면
**Then** 총계는 정확히 65다.

```bash
grep -c 'canary_gate: true' .claude/rules/moai/core/zone-registry.md
```

- **baseline `[실행됨]`**: `73` (템플릿 레지스트리도 동일하게 `73`).
  `014..021` 8개 전부가 `canary_gate: true`임을 별도 확인 (`004..013`은 10개, 합 18).
- **통과 기준**: `65` (등호)
- **주석**: 65 미만은 주제가 소멸하지 않은 절까지 지운 것이다 — REQ-CLD-011 위반.
- **이 AC 단독으로는 불충분 (변이 검증)**: 레지스트리 8개 항목만 지우고 **소스 파일은
  남긴** 반쪽 상태에서 이 명령은 정확히 `65`를 반환한다 — 즉 통과값이다.
  ```
  STATE B (항목 8개 삭제, 소스 유지)   canary=65  validate-exit=1  ZONE_UNREGISTERED=0
  ```
  그 상태는 AC-CLD-007(템플릿에 `moai pr watch` 잔존)과 AC-CLD-005(파일 잔존)가 잡는다.
  이 AC는 **수량**만 판정하며 소스와의 정합은 판정하지 않는다.

---

## AC-CLD-012 — 두 레지스트리의 ci 절이 일치한다

**Given** 저장소 루트와 템플릿 레지스트리가 동일한 ci 절을 보유하고
**When** 한쪽만 수정되면
**Then** diff가 발생해 실패한다.

```bash
ci_clauses() { grep -A 6 -E '^- id: CONST-V3R5-(00[4-9]|01[0-9]|02[01])$' "$1" | grep '^  clause:'; }
R=$(ci_clauses .claude/rules/moai/core/zone-registry.md)
T=$(ci_clauses internal/template/templates/.claude/rules/moai/core/zone-registry.md)
diff <(echo "$R") <(echo "$T") >/dev/null 2>&1; echo "diff-exit=$?"
echo "root-lines=$(echo "$R" | grep -c .) tmpl-lines=$(echo "$T" | grep -c .)"
```

- **baseline `[변이검증됨]`**:
  ```
  STATE A (현재)              diff-exit=0  root-lines=18  tmpl-lines=18
  STATE B (템플릿 절 1개 변형)  diff-exit=1
  STATE C (복원)              diff-exit=0
  ```
- **통과 기준**: `diff-exit=0` **그리고** `root-lines=10` **그리고** `tmpl-lines=10`
  (`004..013`만 잔존)
- **주석 (감사 N3 반영 — 초판은 게이트가 아니었다)**:
  초판의 통과 조건은 `diff-exit=0` 하나뿐이었는데, **그 값이 바로 현재 baseline**이다.
  즉 아무 작업도 하지 않아도 통과했고, 양쪽 모두에서 M2를 건너뛰어도 통과했다.
  "각 10행" 조건은 산문에만 있었고 명령은 행 수를 세지 않았다 —
  **명령에 없는 판정은 존재하지 않는 판정이다.**
  행 수를 명령에 넣어 18→10 전이 자체를 판정 대상으로 만들었다.
- `moai constitution validate`는 `file:` 항목을 프로젝트 루트 기준으로 해석하므로
  템플릿 레지스트리를 가리켜도 저장소 소스를 검증한다. 템플릿 측 정합성은 이 AC로만
  판정 가능하다 (감사 D5).

---

## AC-CLD-013 — 개발 전용 자산 9개 스크립트 전부가 보존된다

**Given** ci-loop 스크립트 9개와 스킬이 개발 저장소에만 존재하고
**When** 배포 격리가 완료되면
**Then** 9개 전부와 스킬 디렉터리가 그대로 남는다.

```bash
find scripts/ci-watch scripts/ci-autofix -type f | wc -l
test -d .claude/skills/moai-workflow-ci-loop && echo SKILL_OK
```

- **baseline `[실행됨]`**: `9` / `SKILL_OK`
- **통과 기준**: `9` / `SKILL_OK` (변화 없음)
- **주석 (감사 D6 반영)**: 초판은 3개 파일만 열거해 나머지 6개
  (`lib/_common.sh`, `lib/classify.sh`, `lib/timeout.sh`, `ci-watch/test/run_test.sh`,
  `ci-autofix/test/classify_test.sh`, `ci-autofix/test/log_fetch_test.sh`)의 삭제를
  탐지하지 못했다. `find -type f`가 전수를 판정한다.

---

## AC-CLD-014 — 템플릿 중립성이 유지된다

**Given** 템플릿이 16개 언어에 중립적이고 내부 흔적을 담지 않아야 하며
**When** 변경이 적용되면
**Then** 중립성 가드가 통과한다.

```bash
go test ./internal/template/ -run 'TestTemplateNoInternalContentLeak|TestSplitHarnessNamespaceNoLeak' -count=1 -v
grep -rn 'SPEC-CI-LOOP-DEVONLY-001' internal/template/templates/ internal/template/catalog.yaml | wc -l
```

- **baseline `[실행됨]`**: 두 테스트 모두 실행 확인 —
  ```
  === RUN   TestTemplateNoInternalContentLeak
  --- PASS: TestTemplateNoInternalContentLeak (1.08s)
  === RUN   TestSplitHarnessNamespaceNoLeak
  --- PASS: TestSplitHarnessNamespaceNoLeak (0.00s)
  ```
  두 번째 명령은 `0`.
- **통과 기준**: 두 테스트 PASS, 두 번째 명령 `0`
- **주석 (감사 D3 반영)**: 초판 선택자는 `TestInternalContentLeak`(존재하지 않는 이름)을
  지목했다. 실행 확인 결과 `testing: warning: no tests to run` + `PASS` + exit 0 —
  **공허 통과**였다. 실제 함수명은 `TestTemplateNoInternalContentLeak`
  (`internal_content_leak_test.go:1170`)이며 위 선택자로 교정했다.
  `-v`를 붙여 실행 사실이 출력에 남게 한다.

---

## AC-CLD-015 — 이미 배포된 스킬이 아카이브된다

**Given** `legacySkillIDs`가 은퇴 스킬의 아카이브 목록이고
**When** 이 스킬이 등재되면
**Then** `moai update` 시 사용자 디렉터리가 삭제가 아니라 아카이브로 이동한다.

```bash
grep -c 'moai-workflow-ci-loop' internal/cli/update_archive.go
go test ./internal/cli/ -run 'TestLegacySkillIDsNotEmbedded' -count=1
```

- **baseline `[실행됨]`**: `0` (미등재). 두 번째 명령은 현재 PASS.
- **통과 기준**: `1` 이상, 그리고 `TestLegacySkillIDsNotEmbedded` 계속 PASS
- **주석**: 두 번째 명령이 AC-CLD-003과 상호 검증한다 — 임베드에서 제거되지 않은 채
  목록에 넣으면 이 테스트가 실패하므로 M4→M5 순서가 기계적으로 강제된다.
- **판정하지 않는 것 (감사 N9)**: 이 AC는 **목록 등재**와 **분리**만 관측한다.
  "삭제가 아니라 아카이브로 동작하는가"는 어느 명령도 관측하지 않으므로
  REQ-CLD-018은 이 AC로 충족되지 않는다 — 커버리지 표의 부채 항목 참조.

---

## AC-CLD-016 — required-checks.yml 전제가 소멸한다 (결정 B 검증)

**Given** 배포 트리 4개 파일이 `.github/required-checks.yml`을 전제로 요구하고
**When** M2·M3가 완료되면
**Then** 배포 트리에 이 전제를 요구하는 참조가 남지 않는다.

```bash
grep -rln 'required-checks.yml' internal/template/templates/ | wc -l
```

- **baseline `[실행됨]`**: `4` — `zone-registry.md`(M2), `ci-watch-protocol.md`(M3),
  ci-loop `SKILL.md`(M3), `sync/delivery.md:460`(M3, ci-loop 스킬 서술 내부)
- **통과 기준**: `0`
- **주석**: 결정 B는 이 결함을 이연하되 "소멸 여부를 확인하라"고 요구했다.
  네 참조가 전부 제거 범위 안에 있으므로 `0`이 기대값이며, `0`이 아니면
  이연이 아니라 **별도 SPEC 승격**이 필요하다는 신호다.

---

## AC-CLD-017 — 개발 저장소 미러가 동기화되었다

**Given** 미러 측 파일들이 스크립트 또는 스킬을 참조하고
**When** M7의 Template-First 동기화가 완료되면
**Then** 선언된 dev-only 자산을 제외한 미러 파일에 참조가 남지 않는다.

```bash
{ grep -rl 'scripts/ci-watch\|scripts/ci-autofix' .claude/ .moai/config/sections/delegation.yaml 2>/dev/null
  grep -rl 'moai-workflow-ci-loop'                .claude/ .moai/config/sections/delegation.yaml 2>/dev/null; } \
  | grep -v '^\.claude/agent-memory/' \
  | grep -v '^\.claude/skills/moai-workflow-ci-loop/' \
  | grep -vE '^\.claude/rules/moai/workflow/ci-(watch|autofix)-protocol\.md$' \
  | sort -u | wc -l
```

- **baseline `[변이검증됨]`**: `10`. 제외 단계별 관측: 원시 `14` → agent-memory 제외 `13`
  → 보존 스킬 디렉터리 제외 `12` → 보존 `ci-watch-protocol.md` 제외 `11`
  → 보존 `ci-autofix-protocol.md` 제외 **`10`**.
  **도달 가능성 증명** (산술이 아니라 변이로 확인): 10개 파일에서 토큰 보유 행을 제거하자
  ```
  BEFORE=10   AFTER=0   RESTORED=10   git status --porcelain → SPEC 디렉터리만
  ```
  즉 통과값 `0`은 실제로 도달 가능하다 — 구조적으로 제거 불가능한 참조는 없다.
- **통과 기준**: `0`
- **선언된 dev-only 보존 집합** (선택자가 제외하는 대상, plan.md §B와 일치):
  `scripts/ci-watch/`, `scripts/ci-autofix/`, `.claude/skills/moai-workflow-ci-loop/`,
  `.claude/rules/moai/workflow/ci-watch-protocol.md`,
  `.claude/rules/moai/workflow/ci-autofix-protocol.md`.
- **autofix 프로토콜 편입 (감사 P2 — 오케스트레이터 결정)**:
  초판은 watch 프로토콜만 보존 대상으로 삼았으나, 그 근거("개발 저장소가 스크립트를
  보유하므로 규율 규칙도 남는다")는 autofix 프로토콜에 **똑같이** 적용된다 —
  이 파일은 `scripts/ci-autofix/log-fetch.sh`·`classify.sh`를 실행 가능 전제 조건으로
  명시하며 스크립트 참조 6건을 보유한다. 한쪽만 보존하면
  "스크립트는 남기고 그 문서는 지운다"는 모순이 남는다. 동일 근거에 동일 처우를 적용한다.
- **템플릿/미러 비대칭 (의도적, 오류 아님)**: 같은 파일이 양측에서 다르게 처리된다.
  | 측 | `ci-autofix-protocol.md` | 이유 |
  |---|---|---|
  | 템플릿 | 스크립트 의존성을 제거하도록 **재작성** (AC-CLD-008) | 배포 사용자에게 스크립트가 없다 |
  | 미러 | **원형 보존** (이 AC의 제외 대상) | 개발 저장소는 스크립트를 보유한다 |
  두 처우 모두 "규칙은 그것이 규율하는 대상이 존재하는 곳에만 있어야 한다"는
  하나의 원칙에서 나온다.
- **[HARD] `.claude/agent-memory/` 제외 필수**: 이 디렉터리의 gitignored 메모리 파일이
  문자열을 포함해 순진한 집계를 `13`에서 `14`로 부풀린다. 제외하지 않으면 이 AC는
  **영원히 통과하지 못한다.**
- **주석 (감사 N4 반영 — 커버리지 구멍)**:
  이 기준은 초판(iteration 1)에 있었으나 개정 중 **대체 없이 사라졌다.**
  그 결과 M7 전체가 무판정 상태였고, 미러가 모든 참조를 그대로 보유해도
  나머지 16개 AC가 전부 통과했다. 어떤 판정 명령도 `.claude/` 경로를
  (레지스트리와 보존 스킬 디렉터리를 빼면) 건드리지 않았기 때문이다.

---

## AC-CLD-018 — `moai pr watch`의 동작·플래그·종료 코드가 M6 편집으로 변하지 않는다

**Given** M6가 `pr_watch_cmd.go`의 사용자 노출 문자열만 편집하고
**When** 편집 후 바이너리를 재빌드하면
**Then** 네 가지 관측값이 편집 전과 동일하다.

```bash
make build >/dev/null 2>&1; echo "build-exit=$?"
test ! -e .moai/state/ci-watch-active.flag; echo "flag-absent=$?"
moai pr watch 999 --branch main >/dev/null 2>&1; echo "default-exit=$?"
echo "flags=$(moai pr watch --help 2>&1 | grep -oE '\-\-(abort|branch|report)' | sort -u | wc -l | tr -d ' ')"
moai pr watch --abort >/dev/null 2>&1; echo "abort-exit=$?"
moai pr watch 999 --report >/dev/null 2>&1; echo "report-exit=$?"
```

- **baseline `[실행됨]`** — 이 판을 쓰면서 네 단언을 재실행해 관측:
  ```
  build-exit=0
  flag-absent=0
  default-exit=0
  flags=3
  abort-exit=1
  report-exit=0
  ```
- **통과 기준**: 위 여섯 값과 **동일** (baseline == 통과값, 설계상 의도)
- **[HARD] 재빌드 선행**: 판정은 반드시 M6 편집을 반영해 `make build`한 바이너리로
  수행한다. 낡은 바이너리로 실행하면 이 AC는 편집을 전혀 관측하지 못하는
  **공허한 통과**가 된다 — 그래서 `make build`를 명령의 첫 줄에 넣었다.
  같은 이유로 `build-exit=0`이 통과 조건에 포함된다.
- **`flag-absent=0`이 전제인 이유**: `abort-exit=1`은 `.moai/state/ci-watch-active.flag`가
  **부재할 때**의 값이다. 상태 파일이 존재하면 `SetAbortFlag`가 성공해 종료 코드가 달라진다.
  전제를 명령에 넣지 않으면 이 줄은 환경에 따라 흔들린다.
- **baseline == 통과값인 이유 (판정 규율의 의도적 예외)**:
  이 문서의 § 판정 규율은 "baseline이 통과값과 같으면 게이트가 아니다"라고 적었다
  (AC-CLD-012의 교훈). AC-CLD-018은 그 발견적 규칙을 **의도적으로 위반한다** —
  이것은 **전이(transition) 단언이 아니라 불변(invariance) 단언**이기 때문이다.
  전이 AC는 "무엇이 바뀌었는가"를 묻고 불변 AC는 "무엇이 바뀌지 않았는가"를 묻는다.
  후자에서 baseline과 통과값이 같은 것은 결함이 아니라 정의다.
- **그럼에도 공허하지 않은 이유**: 통과값이 baseline과 같아도 이 AC는 **실패할 수 있다.**
  M6의 문자열 편집이 `RunE` 분기(`flags.abort`/`flags.report` 조기 반환), 플래그 정의
  (`cmd.Flags().BoolVar` 3줄), 또는 반환값(`return nil` vs `return fmt.Errorf`)을
  건드리면 `default-exit`·`flags`·`abort-exit`·`report-exit` 중 하나가 즉시 어긋난다.
  실제로 문자열 3곳 중 2곳(72·73행)은 `RunE` **본문 안**에 있어 편집 사고 반경 안에 있다.
  감시 대상이 없는 것이 아니라, 감시 대상이 "변하지 않음"인 것이다.
- **담당 마일스톤**: M6.
- **이 AC가 판정하지 않는 것**: 문자열이 실제로 제거되었는지는 AC-CLD-006이 판정한다.
  둘은 짝이다 — AC-CLD-006은 "바뀌어야 할 것이 바뀌었나", AC-CLD-018은
  "바뀌지 말아야 할 것이 그대로인가".

---

## AC-CLD-019 — `archiveSkill`이 사용자 저작물을 삭제가 아니라 복사로 보존한다

**Given** 은퇴 스킬 디렉터리에 사용자가 작성한 내용이 들어 있고
**When** `archiveSkill`이 호출되면
**Then** 아카이브 대상에 같은 내용이 존재하고, 원본은 이 호출로 사라지지 않는다.

```bash
go test ./internal/cli/ -run TestArchiveSkill_PreservesUserContent -count=1 -v 2>&1 \
  | grep -q '^--- PASS: TestArchiveSkill_PreservesUserContent'; echo "exit=$?"
```

- **baseline `[미존재]`**: 이 가드는 아직 없다. 통과 기준이 "존재 + PASS"이므로
  공허 통과가 불가능하다 (AC-CLD-003과 같은 계열).
- **통과 기준**: `exit=0`
- **테스트가 관측해야 하는 것 (run-phase 구현 계약)**:
  1. `t.TempDir()` 아래 `.claude/skills/<id>/`에 **식별 가능한 사용자 저작 내용**을 둔다
  2. `archiveSkill(root, id)` 호출
  3. `.moai/archive/skills/<archiveVersion>/<id>/` 아래에 **바이트 동일한** 내용이 존재
  4. 원본 `.claude/skills/<id>/`가 **호출 후에도 그대로 존재**
- **3번이 아카이브와 삭제를 가르는 관측이다**: 삭제 구현이었다면 아카이브 경로에
  아무것도 없다. 존재만 확인하고 내용을 대조하지 않으면 빈 디렉터리를 만들어 놓는
  구현도 통과하므로, **바이트 대조**까지가 판정이다.
- **4번은 복사→이동 회귀를 막는다**: `archiveSkill`은 복사만 하고 제거하지 않는다
  (내용 앵커 `func archiveSkill` — `copyDirAll` 후 `return nil`, `os.Remove` 없음).
  누군가 이를 `os.Rename`(이동)으로 바꾸면 4번이 실패한다.
- **이 AC가 판정하지 않는 것 (중요)**: "원본이 결국 사라지는가"는 여기서 관측하지
  **않는다.** `archiveSkill`도 `archiveLegacySkills`도 원본을 제거하지 않기 때문이다 —
  원본의 소멸은 update의 clean-reinstall이 `.claude/skills/`를 재배포할 때
  은퇴 스킬이 템플릿에 없어서 다시 깔리지 않는 결과일 뿐이다. 따라서 이 단위 시험에
  "호출 후 원본 부재"를 단언하면 **코드에 없는 모델을 기입**하게 된다.
  REQ-CLD-018이 요구하는 것은 "삭제하지 않을 것"이며, 그것은 3+4로 관측된다.
- **담당 마일스톤**: M5.

---

## 판정 규율

- **도달성**: AC-CLD-003(임베드 FS), AC-CLD-004(해시 신선도), AC-CLD-006(바이너리 문자열)이
  배포 산출물을 판정한다. grep 기반 AC는 소스 정합성만 판정하므로 단독으로는
  "배포되지 않음"을 증명하지 못한다.
- **이 SPEC이 실제로 만드는 상태로 변이할 것**: AC-CLD-004는 초판에서 **내용 변경**으로
  변이 테스트했으나 M3가 수행하는 것은 **삭제**였고, 삭제 상태에서는 게이트가 울리지
  않았다. 변이는 "이 마일스톤이 이 파일에 무엇을 하는가"에 맞춰야 한다.
- **도구가 중단될 수 있으면 종료 코드를 볼 것**: AC-CLD-009·010은 `grep -c`만 보다가
  fatal abort로 잘린 출력을 개선으로 읽었다(18→11, 77→49). 이제 `validate-exit`을 함께 본다.
- **baseline이 통과값과 같으면 게이트가 아니다**: AC-CLD-012는 초판에서 `diff-exit=0`
  하나만 요구했고 그것이 현재 값이었다 — 작업 없이 통과했다.
  **의도적 예외 1건**: AC-CLD-018은 불변(invariance) 단언이라 baseline == 통과값이
  정의상 성립한다. 대신 재빌드를 명령에 넣어 낡은 바이너리로 인한 공허 통과를 막았다
  (해당 AC 주석 참조).
- **명령에 없는 판정은 없다**: 같은 AC의 "각 10행" 조건은 산문에만 있었다. 명령에 넣었다.
- **변이 검증 완료**: AC-CLD-004, 007, 009, 010, 012 (양쪽 상태 출력을 각 AC에 기록).
  나머지 전이 AC는 baseline ≠ 통과값이므로(27≠0, 9≠0, 5≠0, 73≠65, 3≠0, 4≠0, 11≠0)
  수정을 되돌리면 자동 실패한다.
  AC-CLD-018은 전이가 아니라 불변 단언이라 이 논거가 적용되지 않으며,
  비공허성은 재빌드 선행 + `RunE` 본문 내 편집 사고 반경으로 확보한다.
- **비공허성**: 전 선택자가 현재 0이 아닌 매치를 반환함을 확인했다.
  예외는 미존재 가드 2건 — AC-CLD-003, AC-CLD-019 — 이며 둘 다 통과 기준이
  "존재 + PASS"라 공허 통과가 불가능하다.
- **내용 앵커**: 행 번호·커밋 SHA를 앵커로 쓰지 않는다. AC-CLD-006의 41/72/73은
  baseline 원문 인용일 뿐이고, 선택자 자체는 문자열 토큰에 고정한다.

## Definition of Done

- AC-CLD-001 ~ AC-CLD-019 전부 통과
- 미해결 명확화 항목 0건 (2건 모두 종결 — plan.md §D)
- 미커버 REQ 0건 (REQ-CLD-018은 AC-CLD-019가 판정)
- `go test ./...` 회귀 없음
- 커밋이 EXTEND 봉투(plan.md §B) 밖 파일을 건드리지 않음
