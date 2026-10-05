# SPEC-EVIDENCE-CITATION-CANON-001 — 인수 기준

각 항목은 판정을 결정하는 명령을 함께 적는다. 명령이 없는 항목은 그 사실과 대체 판독법을 적는다.

**수리 전 트리에서 RED임을 확인한 grep만 판정 명령으로 쓴다.** 오늘 이미 통과하는 명령은 편집 전후를 구별하지 못하므로 판정력이 없다(iter1 감사 D6). 아래 grep은 전부 이 트리(HEAD `b64043481`)에서 rc를 확인했다.

## D. AC 매트릭스

| AC | 요구 | 판정 명령 |
|---|---|---|
| AC-ECC-001 | REQ-ECC-002 | grep — 거짓 전제 문장 소멸 |
| AC-ECC-002 | REQ-ECC-001, 003 | grep — 신설 고정 문자열 3종 |
| AC-ECC-003 | REQ-ECC-004 | grep — 인용 넓이 상한 + 통째 반출 금지 |
| AC-ECC-004 | REQ-ECC-005 | grep — 선택 기준 |
| AC-ECC-005 | REQ-ECC-006 | grep — carve-out 2건 모두 |
| AC-ECC-006 | REQ-ECC-006 | `go test ./internal/verify/... ./internal/web/...` |
| AC-ECC-007 | REQ-ECC-002 | grep — manager-lead.md 단정 삭제 |
| AC-ECC-008 | REQ-ECC-002 | grep — output-style 배너 3지점 |
| AC-ECC-009 | REQ-ECC-011 | progress.md §E.2 판별식 기록 3건 |
| AC-ECC-010 | REQ-ECC-007 | `go test` 가드 — 통과 방향, 트리별 하한 |
| AC-ECC-011 | REQ-ECC-010 | `go test` 가드 — 합성 뮤테이션 |
| AC-ECC-012 | REQ-ECC-010 | `go test` 가드 — 실물 뮤테이션 |
| AC-ECC-013 | REQ-ECC-009 | `go test` 가드 — 허용목록 단위 + 개수 |
| AC-ECC-014 | REQ-ECC-012, 013 | grep — `.gitignore` 양쪽 |
| AC-ECC-015 | REQ-ECC-008 | `go test` 가드 — 트리·하위트리 방문 + 미러 정합 |

---

## AC-ECC-001 — 거짓 전제 문장이 사라진다

**Given** 두 규칙 파일이 오늘 `.moai/state/verify/<session>/`를 감사 시점 인용 대상으로 지목하고 있고,
**When** M1 편집을 마친 뒤,
**Then** 그 문장이 두 파일 어디에도 남아 있지 않다.

```
grep -n 'SHALL be persisted under' .claude/rules/moai/core/agent-common-protocol-reference.md   # 기대 0행 (오늘 1행)
grep -n 'evidence is persisted under' .claude/rules/moai/core/agent-common-protocol.md          # 기대 0행 (오늘 1행)
```

## AC-ECC-002 — 스크래치 명명·추적 경로 인용·인용 전 반출이 각각 들어간다

**Given** 세 요소가 모두 필요하고, 넓은 단어 검색(`scratch\|export`)은 오늘 이미 무관한 문장에 걸려 판정력이 없으며(감사 D6: `agent-common-protocol.md:337`의 `session-private scratch dir`),
**When** M1 편집 후,
**Then** M1이 도입하는 **고정 문구 3종**이 각각 존재한다. 세 grep이 모두 매치해야 통과다.

```
grep -c 'machine-local scratch'   .claude/rules/moai/core/agent-common-protocol.md   # 기대 >=1 (오늘 0)
grep -c 'export before citing'    .claude/rules/moai/core/agent-common-protocol.md   # 기대 >=1 (오늘 0)
grep -c '\.moai/reports/<card-id>' .claude/rules/moai/core/agent-common-protocol.md  # 기대 >=1 (오늘 0)
```

세 문구는 M1이 실제로 쓰는 문장과 일치해야 한다. 문구를 바꾸려면 이 AC도 함께 바꾸고, 바꾼 뒤 **수리 전 트리에서 0인지** 다시 확인한다 — 그 확인이 이 AC의 판정력을 지탱한다.

## AC-ECC-003 — 인용 넓이 상한과 통째 반출 금지가 하나의 상한을 이룬다

**Given** 인용문을 넓게 쓰면(`evidence: .moai/state/verify/<session>/`) "이름 붙였다"는 조건이 문자 그대로 만족돼 전부 반출할 수 있고(감사 D11),
**When** M1 편집 후,
**Then** reference 파일이 (a) 인용이 **파일 하나**를 이름 붙인다는 상한과 (b) 디렉터리 통째 반출 금지를 **함께** 담는다.

```
grep -c 'names one file'   .claude/rules/moai/core/agent-common-protocol-reference.md   # 기대 >=1 (오늘 0)
grep -c -i 'never the directory\|wholesale' .claude/rules/moai/core/agent-common-protocol-reference.md   # 기대 >=1 (오늘 0)
```

## AC-ECC-004 — 선택 기준이 서술된다

**Given** 상한만으로는 그 안에서 무엇을 남길지 정할 수 없고,
**When** M1 편집 후,
**Then** reference 파일이 "판정을 결정한 명령과 그 판정 결정선만"이라는 기준과, 남기지 않은 원문의 손실 위험을 Residual-risk에 적는다는 지시를 함께 담는다.

```
grep -c -i 'residual-risk' .claude/rules/moai/core/agent-common-protocol-reference.md   # 기대 >=1 (오늘 0)
```

## AC-ECC-005 — carve-out이 §1.4의 **둘 다** 열거한다

**Given** 기계 소비자가 둘이고(`internal/verify/store.go:15` 스냅샷 저장소, `internal/web/events.go:29` SSE 감시), 초판은 첫째만 적고 "하나뿐"이라 단정했으며(감사 D5),
**When** M1 편집 후,
**Then** reference 파일이 둘 다 이름 붙인다.

```
grep -c 'state/verify/snapshots' .claude/rules/moai/core/agent-common-protocol-reference.md   # 기대 >=1 (오늘 0)
grep -c 'internal/web/events.go' .claude/rules/moai/core/agent-common-protocol-reference.md   # 기대 >=1 (오늘 0)
```

## AC-ECC-006 — carve-out이 실제 기제를 깨지 않는다

**Given** 문구가 carve-out을 잘못 좁히면 attributable diff-check와 SSE 감시가 규칙 위반이 되고,
**When** M1 편집 후,
**Then** 두 기제의 테스트가 그대로 통과한다.

```
go test ./internal/verify/... ./internal/web/...   # rc=0
```

> **이 AC의 한계.** 이것은 문서 변경이 코드를 깨지 않았음만 보인다 — 문서만 바꾸므로 결과가 변하지 않는 것이 당연하다. 문구가 carve-out을 **올바르게 표현했는지**는 기계가 판정하지 못하고, AC-ECC-005의 grep은 존재만 확인한다. 표현의 정확성은 감사자가 읽어 판정한다. iter1 감사가 이 한계 서술 자체는 정직하다고 판정했고, 실제로 가려진 것은 파손이 아니라 열거 누락이었다(AC-ECC-005가 그것을 맡는다).

## AC-ECC-007 — manager-lead.md의 단정이 사라진다

**Given** `.claude/agents/moai/manager-lead.md:150`이 `.moai/state/verify/`를 "canonical persistence location"으로 단정하고 있고 (REQ-ECC-002의 주어가 doctrine 표면 문서로 넓혀졌으므로 이 파일이 요구 층에 대응한다),
**When** M3 편집 후,
**Then** 그 단정이 사라지고, 증거 표 열(85행)과 fold-row(157행)가 인용하는 경로가 추적 경로로 바뀐다. `mkdir -p` / `tee` 레시피(146-147행)는 스크래치 용도로 남아도 되지만 스크래치임을 밝히는 문장이 붙어야 한다.

```
grep -c 'canonical persistence location' .claude/agents/moai/manager-lead.md   # 기대 0 (오늘 1)
grep -c 'machine-local scratch'          .claude/agents/moai/manager-lead.md   # 기대 >=1 (오늘 0)
```

## AC-ECC-008 — 출력 스타일 배너 3지점이 정정된다

**Given** `.claude/output-styles/moai/moai.md`의 384 / 401 / 587행이 `.moai/state/verify/<session>/`를 증거 경로 예시로 박아 두고 있고,
**When** M3 편집 후,
**Then** 세 지점 모두 `.moai/reports/<card-id>/`를 예시로 쓰고, `(persistent; … survive /tmp clearance)` 괄호 주석이 스크래치와 인용 대상을 혼동하지 않는 문구로 바뀐다.

```
grep -c 'state/verify' .claude/output-styles/moai/moai.md   # 기대 0 (오늘 3)
```

## AC-ECC-009 — 경계 사례 3건이 판별식으로 결정된다

**Given** `gate.md:122`, `loop.md:115`, `run.md:199`가 기계 소비 문맥일 가능성이 높지만 확정되지 않았고,
**When** M2 완료 후,
**Then** `progress.md` §E.2에 파일별로 판별식 적용 결과(carve-out 또는 교체)와 그 이유가 한 줄씩, 3건 모두 기록된다.

기계 판정 명령 없음 — 판별식은 "최종적으로 사람이 읽는가"를 묻고, 이는 읽어서 정한다. 대체 판독법: 감사자가 `progress.md` §E.2의 3줄과 각 파일의 해당 행을 대조한다. 기록의 **존재**만은 기계로 확인 가능하다:

```
grep -c 'gate.md\|loop.md\|run.md' .moai/specs/SPEC-EVIDENCE-CITATION-CANON-001/progress.md   # 기대 >=3
```

## AC-ECC-010 — 가드가 통과 방향에서 초록이고, 하한이 실제 모집단에서 도출된다

**Given** 가드의 반공허 장치가 모집단보다 두 자릿수 작은 하한(7)을 쓰면 범위를 `.claude/rules/moai/core/` 하나로 좁혀도 걸리지 않고(감사 D1),
**When** M1·M3 편집 후 가드를 실행하면,
**Then** **두 트리 각각**에 대해 스캔 파일 수가 아래 하한을 넘고 위반이 0이다.

하한 도출(이 트리, HEAD `b64043481`). **양쪽 명령이 같은 4개 하위트리를 열거한다** — 초판의 미러 명령은 `templates/.claude` 전체를 훑어 340을 냈는데, 그것은 §D.1이 정한 스캔 범위(4개 하위트리, 338)보다 넓은 모집단이었다(iter2 감사 N3). 차이 2는 범위 밖 파일 `templates/.claude/loop.md`와 `templates/.claude/commands/moai/todo.md`다.

```
find .claude/rules .claude/agents .claude/output-styles .claude/skills -name '*.md' -type f | wc -l   # 363 → 하한 300
T=internal/template/templates/.claude
find $T/rules $T/agents $T/output-styles $T/skills -name '*.md' -type f | wc -l                        # 338 → 하한 300
```

하한 300은 트리별 실측(363 / 338)에서 여유를 두고 내린 값이며, 범위가 한 하위트리로 붕괴하면 넘지 못한다(가장 큰 `.claude/skills` 단독도 251 < 300).

> **하한만으로는 부족하고, 그 부족분을 AC-ECC-015가 맡는다.** 300은 여유를 63(루트) / 38(미러) 남기고, 그 안에 작은 하위트리 둘이 통째로 들어간다: `.claude/agents` 21개를 빼면 342, `.claude/output-styles` 3개를 빼면 360 — 둘 다 통과한다(미러도 같은 모양: 338 − 11 = 327, 338 − 3 = 335). 하필 그 둘이 §1.3이 "가장 강한 사례"라 부른 `manager-lead.md`와 AC-ECC-008이 존재하는 이유인 배너 3지점을 담고 있다. 하한을 올려서 막을 수 있는 문제가 아니다 — 집계 수는 어느 하위트리가 빠졌는지 원리상 구별하지 못한다. 하위트리 **존재**를 따로 단언해야 한다.

```
go test -run 'EvidenceCitation' ./internal/template/...   # rc=0
```

> **하한 근거로 "위반 파일 수"를 쓰지 않는다.** 초판은 "오늘 이 범위의 위반 파일이 7개"라며 하한 7을 정당화했는데, (a) 그 7은 저장소 루트만 센 값이고 두 트리 기준으로는 14였으며, (b) 위반 파일 수와 스캔 파일 수는 서로 다른 모집단이고, (c) M1·M3 수리 후 위반은 0이 되므로 7은 어떤 관측량과도 연결되지 않는다. 하한은 **스캔 모집단**에서만 도출한다.

## AC-ECC-011 — 합성 뮤테이션이 잡힌다

**Given** 스캐너 함수가 있고,
**When** 새 방식 인용 한 줄(`.moai/reports/t999/verify/gotest.log`)과 옛 방식 인용 한 줄(`.moai/state/verify/abc123/gotest.log`)이 든 임시 픽스처를 먹이면,
**Then** 옛 방식 한 줄만 위반으로 보고된다. 둘 다 잡히면 과잉이고, 둘 다 안 잡히면 꺼진 것이다.

## AC-ECC-012 — 실물 뮤테이션이 잡힌다

**Given** 수리 이전 `agent-common-protocol.md:268`의 문장이 오늘 트리에 실재하고,
**When** 그 문장을 리터럴로 담은 픽스처를 스캐너에 먹이면,
**Then** 위반으로 보고된다.

이 항목이 이 가드의 진짜 RED다. 합성 픽스처는 스캐너가 자기가 만든 모양을 잡는 것을 보일 뿐이고, 이 항목은 **저장소에 실제로 있던 문장**을 잡는 것을 보인다.

## AC-ECC-013 — 허용목록이 파일 + 리터럴 단위이고, 개수가 못 박힌다

**Given** 허용목록이 **파일 단위**면 §C.3의 carve-out 후보 3파일이 통째로 면제되어, 나중에 그 파일에 진짜 위반 인용이 들어와도 가드가 보지 못한다(감사 D7 — 저장소 루트 위반 대상 7건 중 3건, 43%가 사각지대),
**When** 가드를 실행하면,
**Then** 둘 다 성립한다:

1. **단위** — 허용목록 항목이 **파일 경로 + 정확한 리터럴** 쌍이고, 파일 경로만으로 된 항목은 존재하지 않는다.
2. **개수** — 항목 수가 상수와 일치한다.

두 단언 각각에 판정 명령이 붙는다. 초판은 #2에만 명령을 달고 #1은 산문으로 남겼는데, 이 문서 자신의 서두가 "명령이 없는 항목은 그 사실과 대체 판독법을 적는다"고 정해 두었으므로 그것은 규약 위반이었다(iter2 감사 D7 잔여).

- **#2 판정**: `grep -c 'allowlistSize\|wantAllowlist' internal/template/evidence_citation_guard_test.go` → 기대 ≥1.
- **#1 판정**: 리터럴이 빈 항목을 만들어 검증자에 먹이는 **뮤테이션 서브테스트**가 존재하고, 검증자가 그것을 거부한다. grep으로 "파일 단독 항목이 없음"을 확인하는 것은 오늘의 목록에 대해서만 참이고 내일 추가될 항목을 막지 못하므로, 판정은 **구조에 대한 단언**이어야 한다.

```
go test -run 'EvidenceCitation/Allowlist' ./internal/template/...   # rc=0, 서브테스트 2개(단위·개수)
```

두 단언이 서로를 보완한다: 개수 단언은 목록이 길어지는 것을 막고, 단위 단언은 **항목 하나가 파일 전체를 삼키는 것**을 막는다. 개수만으로는 후자가 통과한다.

## AC-ECC-014 — ignore 항목이 양쪽 트리에 파일 범위로 들어간다

**Given** 저장소 루트와 템플릿 미러 두 `.gitignore`가 있고,
**When** M5 편집 후,
**Then** 네 조건이 모두 성립한다.

1. 양쪽에 `.moai/observability/*.jsonl`이 있다.
2. 양쪽 어디에도 `.moai/observability/.gitkeep` 예외 줄이 **없다** (§4.2 판정).
3. 저장소에도 템플릿에도 `.moai/observability/` 디렉터리 스캐폴드가 **없다** (§4.2 판정 — 있으면 배포 시 훅 opt-in이 뒤집힌다).
4. 양쪽 어디에도 `navigator`를 담은 줄이 **없다** (§4.3 판정).

```
grep -n 'moai/observability' .gitignore internal/template/templates/.gitignore   # 각 1행 (*.jsonl), ! 줄 0
grep -c 'navigator' .gitignore internal/template/templates/.gitignore            # 기대 각 0
find . internal/template/templates -maxdepth 3 -type d -name observability       # 기대 0행
```

## AC-ECC-015 — 두 트리와 그 아래 네 하위트리가 실제로 방문되고, 두 사본이 일치한다

**Given** REQ-ECC-008에 매핑된 유일한 AC가 "스캔 파일 수 + 위반 0"만 보면, **저장소 루트만 스캔한 가드도 그 조건을 만족한다**(감사 D2 — 363 ≥ 하한, 위반 0). 그리고 이 실패 형태는 기존 `internal/template/gitignore_agents_mirror_test.go`(45행 `filepath.Join("templates", ".gitignore")`, 루트를 읽는 코드 없음)에 실재한다. **또한 하한만으로는 작은 하위트리가 조용히 빠지는 것을 막지 못한다**(AC-ECC-010 주석: agents 21 / output-styles 3이 하한 여유 안에 들어간다).
**When** 가드를 실행하면,
**Then** 넷이 모두 성립한다:

1. **트리 방문** — 가드가 방문한 트리 루트 목록을 반환하고, 그 목록이 저장소 루트와 `internal/template/templates` **둘 다** 포함한다.
2. **하위트리 방문** — 트리마다 방문한 하위트리 집합을 반환하고, 그 집합이 아래 4개와 **정확히 일치한다**(부분집합이 아니라 상등):

   ```
   rules, agents, output-styles, skills
   ```

   존재 단언이므로 `output-styles`(3개)처럼 작은 하위트리도 수치 여유에 삼켜지지 않는다. 하한이 원리상 할 수 없는 구별을 이 단언이 맡는다.
3. **트리별 하한** — 트리별 스캔 파일 수가 각각 보고되고, 각각 AC-ECC-010의 하한을 넘는다(집계 합계가 아니라 **트리별**).
4. **미러 정합** — 두 사본이 이 항목에 관해 일치한다: 루트 사본에서 `.moai/state/verify` 위반이 0이면 미러에서도 0이고, 그 역도 성립한다.

**두 방향의 뮤테이션으로 1·2번이 실제로 잡히는 것을 보인다** — 방문 단언 역시 시연 없이는 공허하다.

- 트리 루트 한쪽을 빼면 1번이 실패한다.
- 하위트리 하나(`agents`)를 빼면 2번이 실패한다. **이 뮤테이션은 3번을 통과한 채로 실패해야 한다**(342 ≥ 300) — 그것이 2번이 3번과 독립적으로 값한다는 증거다. 통과·실패가 함께 움직이면 2번은 3번의 재서술일 뿐이다.

---

## 경계 사례

- **`capability-map.md`가 무시되면 실패.** `navigator_enrich.go:75`가 이 경로를 입력으로 읽는다. AC-ECC-014의 4번(navigator 줄 0)이 이것을 막는다.
- **`.moai/observability/.gitkeep`이 템플릿에 들어가면 실패.** `moai init`이 모든 사용자 프로젝트에 배포해 `os.Stat` opt-in을 opt-out으로 뒤집는다(§4.2). AC-ECC-014의 2·3번이 막는다.
- **`.moai/reports/**`와 `.moai/specs/**`가 가드 범위에 들어가면 첫날부터 빨갛다.** 그것이 spec.md §5의 부채이고 이 SPEC의 범위가 아니다.
- **`.codex/agents/moai/manager-lead.toml` 손편집.** 이 파일은 `make agents-emit` 방출물이고 오늘 `canonical persistence location` 문장을 담고 있다. 가드는 `.md` 한정이라 이 파일을 지키지 않는다 — M6의 `make agents-emit` → `make agents-emit-check`가 전이적으로 고칠 뿐이다. 이 비대칭은 알려진 잔여 위험이다.
- **`internal/cli/mcp_glm.go:110`의 코드 주석**이 `.moai/state/verify/` 경로를 측정 근거로 인용한다. `.go`는 가드 범위 밖이므로 이 SPEC에서 닫히지 않는다(spec.md §5 후속 카드 후보).

## 품질 게이트

- `go test ./internal/template/... ./internal/verify/... ./internal/web/...` rc=0
- `make agents-emit-check` rc=0
- `go vet ./internal/template/...` rc=0
- 템플릿 중립성: 미러 편집분에 SPEC ID·카드 id·내부 날짜·커밋 SHA가 없을 것 (`.moai/reports/<card-id>/`는 경로 형태이므로 중립)

## Definition of Done

1. AC-ECC-001 ~ AC-ECC-015가 모두 통과하거나, 통과하지 못한 항목이 GAP으로 표시되고 그 이유가 `progress.md`에 적혀 있다.
2. 가드가 여섯 방향(통과 / 합성 뮤테이션 / 실물 뮤테이션 / 트리 방문 뮤테이션 / 하위트리 방문 뮤테이션 / 허용목록 단위 뮤테이션) 모두 시연된 출력이 `.moai/reports/t375/`에 반출돼 있다 — 이 SPEC이 세우는 규칙을 이 SPEC 자신이 지킨다. 반출은 REQ-ECC-004의 상한을 따른다: 판정을 결정한 줄만 반출하고 전체 로그는 스크래치에 둔다.
3. spec.md §5의 잔여 부채 2건(124개 문서 인용 정정 / `mcp_glm.go:110` + `.go` 범위 확대)이 후속 카드 후보로 리드에게 보고돼 있다.
