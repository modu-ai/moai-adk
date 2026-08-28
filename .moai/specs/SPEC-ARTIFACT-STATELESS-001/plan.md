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

**중간 상태 — M2·M3 순서는 무방하되, 그 근거는 AC 설계에 있다.** M2를 먼저 착지시키면 그 구간에서는 코퍼스 잔여(착수 시점 389 @ `3b1830b96`)에 lint가 걸린다. 이것이 AC를 깨지 않는 이유는 `AC-AST-001-04`의 lint 호출을 **이 SPEC 하나로 한정**했기 때문이다 — 단일 SPEC 범위의 `before` 매치는 코퍼스 잔여와 무관하게 0이므로, 비공허성 가드가 순서에 종속되지 않는다. 반대로 전 코퍼스 호출로 두면 M2 선행 시 **정상 동작하는 lint에도 FAIL**하므로 순서가 강제된다. 즉 "순서 무방"은 자유 서술이 아니라 AC 범위 한정이 사들인 성질이다.

전 코퍼스를 보는 `AC-AST-001-05`·`-06`·`-09`는 M3 착지 후에 판정한다 — 중간 상태에서 돌리지 않는다.

### B3. 심각도 — 어느 수준으로 거부할 것인가

권고: **error**. 근거는 B2와 같다 — 착지 시점에 위반이 0이므로 error가 기존 코퍼스를 깨지 않는다. warning으로 두면 재발이 조용히 쌓여, 선언만 하고 끝내는 것과 결과가 같아진다(spec.md §1.7).

**번복 지점**: 배포 사용자 코퍼스에 기존 위반이 있을 가능성. 이것은 **이 리포에서 측정 불가**하며 Gap이다. 위험을 낮추려면 first-run에서 warning으로 내고 다음 릴리스에 error로 승격하는 2단계도 가능하나, 그 경우 B2의 "예외 불필요" 논거가 약해진다.

### B4. 정리 정의 D1 고정 + 모집단은 전 코퍼스 696

**정의**: `status:` 라인만 제거한다. `id`/`title`/`version`/`created`는 건드리지 않는다.

**모집단**: `.moai/specs/SPEC-*/` **전부** — 라이프사이클 status로 좁히지 않는다. 근거는 spec.md §1.6: 무상태 선언이 Tier에 의존하지 않는 것과 **같은 이유로** status에도 의존하지 않으며, 종결로 좁히면 오늘 `in-progress`인 SPEC이 내일 규약 밖 필드를 달고 닫히는 구멍이 남는다.

두 모집단의 실측(`3b1830b96`, `bash .moai/reports/t357/t357_d1_all.sh .`):

| 모집단 | D1 대상 |
|---|---:|
| **전 코퍼스 696 (채택)** | **389** |
| 종결 633 (참고값, 운영자에게 제시됐던 수치) | 362 |

**번복 지점**: 정의 D1은 운영자가 (H1)로 명시 지정했으므로 없음에 가깝다. 모집단 확대는 이번 개정의 판단이며, 편집 대상이 362 → 389로 늘고 그중 27건이 **미종결 SPEC의 산출물**이라 다른 레인의 미커밋 변경과 겹칠 여지가 커진다(§D 리스크).

**±20% 가드는 같은 모집단으로 잰다.** M3 착수 시 재측정한 전 코퍼스 값이 착수 시점 baseline **389** 대비 ±20%를 넘으면 원인을 먼저 규명한다. 종전 판(362 대비 비교)은 서로 다른 모집단의 수를 견주는 것이라 통과하더라도 우연이었다.

### B5. 규약 문구를 어디에 넣는가 — 앵커 제목이 AC의 판정 범위다

`spec-frontmatter-schema.md`의 `## Canonical 12 Required Fields` 바로 아래에 **`### Artifact Statelessness`** 라는 고정 앵커 제목의 소절을 신설한다. 현재는 13행 괄호(`(spec.md)`)에 **암시**돼 있을 뿐이라, 읽는 사람이 "나머지는 어떻게 하라는 건가"에 답을 얻지 못한다.

**앵커 제목은 장식이 아니라 판정 장치다.** `AC-AST-001-01`·`-02`는 이 앵커로 소절 본문을 잘라낸 뒤 그 안에서만 grep한다. 문서 전역 grep은 이미 참인 문장들에 걸려 판정력을 잃으므로(종전 판에서 7개 conjunct 중 6개가 미수정 파일에서 이미 참이었다), 범위 한정이 AC를 공허하지 않게 만드는 유일한 수단이다. 앵커를 바꾸려면 두 AC를 함께 바꾼다.

소절이 담아야 할 **네 문장**(각각 별도 AC 판정 대상, 리터럴 형태가 고정된 것은 REQ가 명시):

1. 12필드 의무는 `spec.md` 하나만 구속한다 — 리터럴 `` binds `spec.md` only ``
2. `plan.md`·`acceptance.md`·`design.md`·`research.md` 네 종은 무상태다 — frontmatter의 `status:` 필드 금지
3. 이 선언은 **Tier와 무관**하다 — 리터럴 `Tier-independent` (spec.md §1.5의 106건 구멍을 닫는 절)
4. frontmatter 자체는 허용된다 — 리터럴 `Frontmatter itself is permitted` (무상태를 "frontmatter 금지"로 오독하는 것을 막는 절)

### B6. 템플릿 미러는 조건이 아니라 사실이다

미러는 실재하며 착수 시점에 **바이트 동일**하다:

```bash
diff -q .claude/rules/moai/development/spec-frontmatter-schema.md \
        internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md
# (무출력) @ 3b1830b96 — 양쪽 23,317 bytes
```

따라서 M1은 "미러가 있으면"이 아니라 **"미러를 같은 커밋에서 함께 고친다"** 이다(Template-First, `CLAUDE.local.md` §2 [HARD]). `AC-AST-001-11`이 드리프트를 판정한다 — 한쪽만 고치면 FAIL.

**번복 지점**: 없다. 확인 가능한 사실을 미지로 남길 이유가 없다.

---

## §C 마일스톤

### M1 — 규약 명문화

| | |
|---|---|
| 대상 | `.claude/rules/moai/development/spec-frontmatter-schema.md` **+ 그 템플릿 미러** |
| 요구사항 | REQ-AST-001-001 / 002 / 003 / 014 / 015 |
| 산출 | `### Artifact Statelessness` 소절 1개 × 2벌(로컬 + 미러) |

- 앵커 제목은 **`### Artifact Statelessness`** 로 고정한다 — AC의 판정 범위다(§B5).
- 소절이 담을 네 문장과 그중 리터럴이 고정된 세 개는 §B5 목록 참조: `` binds `spec.md` only `` / `Tier-independent` / `Frontmatter itself is permitted`.
- 무상태의 정의를 **status 축**(= `status:` 필드 부재)으로 한정하고, frontmatter 자체를 금지하는 문장은 쓰지 않는다.
- **템플릿 미러를 같은 커밋에서 함께 수정한다**(§B6 — 미러는 실재하며 현재 바이트 동일). 수정 후 `make build`.

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

관측 의무 (H3, REQ-AST-001-007): 규칙을 작성하는 것으로 끝내지 않는다. 실제로 `status:`를 심어 **거부가 나오는 것을 관측**하고, 원복한다. 명령은 `acceptance.md` `AC-AST-001-04`에 실행 가능한 형태로 있으며, lint 호출은 **이 SPEC 하나로 한정**한다(§B2 — 전 코퍼스 호출이면 M2 선행 시 정상 lint에도 FAIL한다). 이것을 빠뜨린 것이 카드 t355의 실패 형태다 — 존재하지만 아무것도 보지 않는 검사.

### M3 — D1 코퍼스 정리 (기계적)

| | |
|---|---|
| 대상 | `.moai/specs/SPEC-*/{plan,acceptance,design,research}.md` — **전 코퍼스 696, status 무관** |
| 요구사항 | REQ-AST-001-008 / 009 / 010 / 011 |
| 산출 | `status:` 라인 제거 (대상 수는 run-phase 재측정값; 착수 시점 baseline 389 @ `3b1830b96`) |

- **착수 전 재측정 필수**(REQ-AST-001-009). `3b1830b96`의 389를 판정값으로 재사용하지 않는다 — `origin/develop`이 `48d8ef4be`로 26 커밋 전진했다. 재측정 명령은 `bash .moai/reports/t357/t357_d1_all.sh .`, 결과는 `progress.md §E.2`의 「기준 SHA」 표에 HEAD와 함께 기록한다.
- 재측정값이 baseline 389 대비 ±20%를 넘으면 원인 규명이 먼저다(§B4).
- frontmatter 블록 **안**의 `status:` 라인만 제거한다. 본문에 나오는 `status:` 문자열은 건드리지 않는다(범위를 `---` 블록으로 좁힌다).
- `spec.md`·`progress.md`는 대상에서 제외한다(REQ-AST-001-011).
- **미종결 SPEC의 산출물 27건도 대상이다.** 다른 레인이 그 SPEC을 작업 중일 수 있으므로, 명시 pathspec 스테이징 + 스테이징과 **같은 호출에서** `git status --short` 재판독을 지킨다. `git add -A` / `git add .` / `git commit -a` 금지.
- M2와 **같은 SPEC에서 착지**한다(REQ-AST-001-010) — 순서는 무방하나(§B2가 그 근거를 설명한다), 둘이 갈라지면 B2가 깨진다.

---

## §D 리스크

| 리스크 | 성격 | 완화 |
|---|---|---|
| 배포 사용자 코퍼스에 기존 위반이 있어 error가 남의 빌드를 깬다 | **Gap — 이 리포에서 측정 불가** | B3의 2단계(warning → error) 대안을 준비. 단 B2 논거가 약해짐을 감수 |
| 재측정 대상 수가 **389**(전 코퍼스 baseline)에서 크게 벗어남 | 원인 미상 상태의 대량 편집 | ±20% 초과 시 원인 규명 후 진행. **같은 모집단끼리** 견준다(B4) |
| 대량 편집이 다른 레인의 미커밋 변경을 삼킴 — **모집단 확대로 커진 리스크** | 공유 트리 사고. 늘어난 27건이 미종결 SPEC이라 다른 레인이 작업 중일 확률이 종결 SPEC보다 높다 | 워크트리 안에서만 편집, 명시 pathspec 스테이징, **스테이징과 같은 호출에서** `git status --short` 재판독, sweep 스테이징 금지 |
| 본문의 `status:` 문자열 오삭제 | 정규식 과다 매칭 | `---` frontmatter 블록으로 범위 제한, `AC-AST-001-07`의 diff 전수 판독 |
| M3를 별도 카드로 분리하고 싶어짐 | B2 결정과 충돌 | 분리한다면 era 예외가 **필수**가 됨 — 두 결정을 한 쌍으로 다룬다 |
| M1이 로컬만 고치고 미러를 빠뜨림 | Template-First [HARD] 위반, `moai update`가 다음 실행에서 되돌림 | `AC-AST-001-11`이 드리프트를 판정(§B6) |

---

## §E 교차 참조

- `spec.md` §1.6 — D1/D2 실측 및 채택 근거
- `spec.md` §1.7 — lint가 본체인 이유, era 예외 미설정의 설계 결합
- `acceptance.md` — 실행 가능한 판정 명령
- `.moai/reports/t357/plan-measurement.md` — 원 실측 보고
