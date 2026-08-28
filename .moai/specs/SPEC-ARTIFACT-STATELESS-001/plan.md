# SPEC-ARTIFACT-STATELESS-001 — 구현 계획

> 이 문서는 결정의 **번복 가능성 순서**로 배열한다. 바뀔 여지가 큰 결정(lint 술어·심각도·정리 정의)이 앞에, 기계적인 단계(대량 편집)가 뒤에 온다.

## §A 맥락

카드 t357. 카드의 두 전제가 실측으로 뒤집혔고(spec.md §1.1), 운영자가 **안 C(무상태 선언)** 를 선택했다. 이 계획은 그 선택을 세 마일스톤으로 나눈다. 측정 baseline은 트리 `c6aa61346`이며, 이후 `origin/develop`이 `48d8ef4be`로 26 커밋 전진했으므로 **정리 단계 착수 전 재측정**이 전제다.

---

## §B 먼저 검토할 결정 (번복 가능성 높음)

### B1. lint 술어 — "status 필드 거부"이지 "frontmatter 금지"가 아니다

새 규칙이 거부하는 것은 **비-spec.md SPEC 산출물의 frontmatter 안 `status:` 필드** 하나다. frontmatter 자체는 허용한다.

| 대상 파일 | frontmatter 있음 + `status:` 있음 | frontmatter 있음 + `status:` 없음 | frontmatter 없음 |
|---|---|---|---|
| `plan.md` / `acceptance.md` / `design.md` / `research.md` | **거부** | 통과 | 통과 |
| `spec.md` | 통과 (기존 `FrontmatterSchemaRule` 소관) | — | — |
| `progress.md` | 통과 (축이 다름) | 통과 | 통과 |

**왜 이 축인가**: 정리 정의(D1)와 검사가 같은 것을 보게 만들기 위해서다. 정의는 status 라인을 지우는데 검사가 frontmatter 전체를 보면, 이 카드가 고치려던 "정의와 검사의 어긋남"을 새로 만든다.

**번복 지점**: "frontmatter 일체 금지"로 넓히자는 판단이 나올 수 있다. 그 경우 대상은 D1 362 → D2 417로 늘고, 정리 대상 55개가 spec.md frontmatter 복제 블록(12~14 필드)이라 별개 결정이 된다. 넓히려면 이 SPEC이 아니라 후속 카드다.

### B2. era/grandfather 예외를 두지 않는다 — 그리고 그것이 M2·M3의 착지 순서를 묶는다

새 finding 코드는 `eraDemotableCodes`(`internal/spec/lint.go:248`)에 넣지 **않는다**. 성립 조건은 하나다: **정리(M3)와 lint(M2)가 같은 SPEC 안에서 함께 착지한다.** 함께 착지하면 lint가 켜질 때 코퍼스 위반이 0이므로 예외가 필요 없다.

**번복 지점**: M3를 별도 카드로 미루자는 판단이 나오면, 이 결정은 즉시 뒤집혀야 한다 — 그때는 예외가 **필수**가 된다. 즉 "M3를 미룬다"와 "예외를 두지 않는다"는 동시에 성립할 수 없다. AC로 못박은 이유다.

### B3. 심각도 — 어느 수준으로 거부할 것인가

권고: **error**. 근거는 B2와 같다 — 착지 시점에 위반이 0이므로 error가 기존 코퍼스를 깨지 않는다. warning으로 두면 재발이 조용히 쌓여, 선언만 하고 끝내는 것과 결과가 같아진다(spec.md §1.7).

**번복 지점**: 배포 사용자 코퍼스에 기존 위반이 있을 가능성. 이것은 **이 리포에서 측정 불가**하며 Gap이다. 위험을 낮추려면 first-run에서 warning으로 내고 다음 릴리스에 error로 승격하는 2단계도 가능하나, 그 경우 B2의 "예외 불필요" 논거가 약해진다.

### B4. 정리 정의 D1 고정

`status:` 라인만 제거한다. `id`/`title`/`version`/`created`는 건드리지 않는다. 근거와 실측은 spec.md §1.6.

**번복 지점**: 없음에 가깝다 — 운영자가 (H1)로 명시 지정했다. 다만 재측정 결과 D1 대상 수가 `c6aa61346`의 362에서 크게 벗어나면(예: ±20% 초과) 원인을 먼저 규명한 뒤 진행한다.

### B5. 규약 문구를 어디에 넣는가

`spec-frontmatter-schema.md`의 `## Canonical 12 Required Fields` 바로 아래에 **별도 소절**을 신설한다. 현재는 13행 괄호(`(spec.md)`)에 **암시**돼 있을 뿐이라, 읽는 사람이 "나머지는 어떻게 하라는 건가"에 답을 얻지 못한다. 명시 문구는 세 가지를 말해야 한다: (1) 대상은 spec.md 1종, (2) 나머지는 무상태, (3) **Tier와 무관**하다(spec.md §1.5의 106건 구멍을 닫는 절).

---

## §C 마일스톤

### M1 — 규약 명문화

| | |
|---|---|
| 대상 | `.claude/rules/moai/development/spec-frontmatter-schema.md` |
| 요구사항 | REQ-AST-001-001 / 002 / 003 |
| 산출 | 신설 소절 1개 (§B5 배치) |

- 12필드 의무가 `spec.md` 1종만 구속함을 **명시 문장으로** 적는다(현재는 괄호 암시).
- `plan.md`·`acceptance.md`·`design.md`·`research.md`를 **무상태**로 선언하고, 무상태의 정의를 **status 축**(= `status:` 필드 부재)으로 한정한다.
- **Tier와 무관**함을 명시한다 — `design.md`를 가지면서 `tier: L`이 아닌 종결 SPEC 106건이 규칙 밖에 남지 않게 하는 절이다.
- 템플릿 미러 확인: 이 규칙 파일이 `internal/template/templates/.claude/rules/moai/development/` 에 미러돼 있는지 확인하고, 있으면 **같은 커밋에서 함께 수정**한 뒤 `make build`.

### M2 — 재발 방지 lint (이 SPEC의 본체)

| | |
|---|---|
| 대상 | `internal/spec/lint.go` (+ 새 테스트 파일) |
| 요구사항 | REQ-AST-001-004 / 005 / 006 / 007 |
| 산출 | 새 `Rule` 1개 + 등록 + 테스트 |

구현 좌표 (설계 메모, spec.md §1.8):

- `Rule.Check(doc *SPECDoc, ...)`가 spec.md를 받으므로, 형제 산출물은 `filepath.Dir(doc.Path)`에서 유도한다. **`discoverSPECs` 변경은 불필요**하다.
- 규칙 등록은 `lint.go:133` 인근 `l.rules = []Rule{...}` 배열. 트리 스캔형 선례는 `HaikuResidualRule`.
- finding 코드는 신설한다(기존 `FrontmatterInvalid` 재사용 금지 — 그것은 spec.md 12필드 스키마 소관이고, 재사용하면 `eraDemotableCodes`의 기존 항목에 딸려 들어가 B2 결정이 무력화된다).
- `eraDemotableCodes`(`lint.go:248`)에 **넣지 않는다**.

관측 의무 (H3, REQ-AST-001-007): 규칙을 작성하는 것으로 끝내지 않는다. 실제로 `status:`를 심어 **거부가 나오는 것을 관측**하고, 원복한다. 명령은 acceptance.md AC-AST-004에 실행 가능한 형태로 있다. 이것을 빠뜨린 것이 카드 t355의 실패 형태다 — 존재하지만 아무것도 보지 않는 검사.

### M3 — D1 코퍼스 정리 (기계적)

| | |
|---|---|
| 대상 | `.moai/specs/SPEC-*/{plan,acceptance,design,research}.md` |
| 요구사항 | REQ-AST-001-008 / 009 / 010 / 011 |
| 산출 | `status:` 라인 제거 (대상 수는 run-phase 재측정값) |

- **착수 전 재측정 필수**(REQ-AST-001-009). `c6aa61346`의 362를 재사용하지 않는다 — `origin/develop`이 `48d8ef4be`로 26 커밋 전진했다.
- frontmatter 블록 **안**의 `status:` 라인만 제거한다. 본문에 나오는 `status:` 문자열은 건드리지 않는다(범위를 `---` 블록으로 좁힌다).
- `spec.md`·`progress.md`는 대상에서 제외한다(REQ-AST-001-011).
- M2와 **같은 SPEC에서 착지**한다(REQ-AST-001-010) — 순서는 M2 → M3든 M3 → M2든 무방하나, 둘이 갈라지면 B2가 깨진다.

---

## §D 리스크

| 리스크 | 성격 | 완화 |
|---|---|---|
| 배포 사용자 코퍼스에 기존 위반이 있어 error가 남의 빌드를 깬다 | **Gap — 이 리포에서 측정 불가** | B3의 2단계(warning → error) 대안을 준비. 단 B2 논거가 약해짐을 감수 |
| 재측정 대상 수가 362에서 크게 벗어남 | 원인 미상 상태의 대량 편집 | ±20% 초과 시 원인 규명 후 진행(B4) |
| 대량 편집이 다른 레인의 미커밋 변경을 삼킴 | 공유 트리 사고 | 워크트리 안에서만 편집, 명시 pathspec 스테이징, `git status --short` 재판독 |
| 본문의 `status:` 문자열 오삭제 | 정규식 과다 매칭 | `---` frontmatter 블록으로 범위 제한, 편집 후 diff 전수 확인 |
| M3를 별도 카드로 분리하고 싶어짐 | B2 결정과 충돌 | 분리한다면 era 예외가 **필수**가 됨 — 두 결정을 한 쌍으로 다룬다 |

---

## §E 교차 참조

- `spec.md` §1.6 — D1/D2 실측 및 채택 근거
- `spec.md` §1.7 — lint가 본체인 이유, era 예외 미설정의 설계 결합
- `acceptance.md` — 실행 가능한 판정 명령
- `.moai/reports/t357/plan-measurement.md` — 원 실측 보고
