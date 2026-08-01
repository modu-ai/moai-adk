# Plan — SPEC-REF-SEO-ABSORB-001

## §A. 맥락

`moai-ref-seo` 신규 레퍼런스 스킬을 클린룸 재작성으로 만들어 템플릿 트리를 통해 전체 사용자에게 배포한다. 동시에 형제 Epic SPEC 2건이 재사용할 클린룸 프로토콜을 확립한다.

조사 결과는 `research.md`에 있다. 아래 §B의 결정 5건은 **전부 확정되었다**(v0.2.0). M2 저작은 이 결정들 위에서 진행한다.

---

## §B. 결정 기록 (5건 전부 확정)

v0.1.0에서 열려 있던 5건은 사용자 결정 라운드로 전부 확정되었다. 각 항목은 선택지 표와 논거를 보존하되, **채택안**과 **그 결과 무엇이 바뀌는가**를 함께 기록한다.

### B.1 [결정 확정] catalog tier = `optional-pack:frontend`

catalog `tier` 값은 장식이 아니라 기능적 선택이다. tier 인지 파일시스템 래퍼가 slim 배포 경로에서 non-core 엔트리를 전부 숨긴다(research.md §C.5).

| 선택지 | 결과 | 근거 |
|---|---|---|
| `core` | slim FS에 포함. SEO를 보편 적용 도메인으로 취급 | canonical URL·robots·sitemap·구조화 데이터는 웹 산출물이 있는 모든 프로젝트에 해당. 프레임워크 무관 |
| **`optional-pack:frontend` (채택)** | slim FS에서 제외. 기존 UI 인접 ref 스킬 2건과 정렬 | `moai-ref-ui-polish`, `moai-ref-react-patterns` 모두 이 값(research.md §D.4 실측) |

**채택 근거** — (a) 웹 산출물이 없는 프로젝트(CLI 도구, 라이브러리, 데이터 파이프라인)에서 SEO 레퍼런스는 순수 토큰 비용이며, 16개 지원 언어 중 상당수의 전형적 프로젝트가 여기 해당한다. (b) 기존 UI 인접 ref 스킬 2건과 동일 팩에 두면 프론트엔드 도메인 활성화 시 함께 켜지는 일관된 묶음이 된다. (c) 되돌리기 방향의 비대칭: 팩 → `core` 승격은 새 사용자에게 자산을 **추가**하는 방향이라 안전하고, `core` → 팩 강등은 이미 받은 사용자에게서 **회수**하는 방향이라 더 시끄럽다. 확신이 없을 때는 좁은 쪽에서 시작한다. (v0.1.0의 근거 (c) 문장은 앞뒤 절이 서로를 부정하는 자기모순이었다 — 위 문장으로 대체한다.)

**반대 논거 정정(실측 반영)**: v0.1.0은 "별도 팩 신설은 본 SPEC 범위를 넘는다"를 근거로 별도 팩 선택지를 각하하면서, **도메인 전용 팩 선례가 이미 존재한다는 사실을 누락했다**. 실측:

```
$ grep -n -A1 'name: moai-ref-secops' internal/template/catalog.yaml
210:                - name: moai-ref-secops
211-                  tier: optional-pack:devops

$ grep -n 'tier: ' internal/template/catalog.yaml | sed 's/.*tier: //' | sort | uniq -c
  29 core
   1 harness-generated
   3 optional-pack:backend
   1 optional-pack:design
   4 optional-pack:devops
   3 optional-pack:frontend
```

즉 본 SPEC이 등록 표면의 표준으로 삼는 `moai-ref-secops` 자신이 `optional-pack:devops`에 있고, 주제별 전용 팩은 확립된 관례다. 따라서 정직한 반대 논거는 "별도 팩 선례가 없다"가 아니라 **"별도 `optional-pack:seo` 신설은 선례가 있으나, 팩 신설 자체가 본 SPEC의 범위를 넘는다"**이다. 이 형태로 각하한다.

**검증기 허용 집합(실측)**: `internal/template/catalog_tier_audit_test.go:300`

```go
tierPattern := regexp.MustCompile(`^(core|optional-pack:[a-z][a-z0-9-]{1,30}|harness-generated)$`)
```

`optional-pack:frontend`는 이 패턴을 만족한다. (`optional-pack:seo` 역시 만족한다 — 각하 사유는 검증기가 아니라 범위다.)

**`WORKFLOW_UNCOVERED` 별도 선언 필요 여부 — 불필요(실측 확인)**: `catalog_tier_audit_test.go:509`의 `WORKFLOW_UNCOVERED` 단언은 워크플로 frontmatter의 `metadata.required-skills`에 선언된 스킬이 catalog에 있는지만 본다. 두 가지를 실측했다. (a) `knownSkills` 집합은 core 뿐 아니라 **모든 optional-pack 스킬을 bare name으로도 등록**하므로(`catalog_tier_audit_test.go` 내 `knownSkills[e.Name] = true`) optional-pack 소속이 누락 사유가 되지 않는다. (b) 현재 워크플로 파일 중 `required-skills`를 선언한 것은 0건이다:

```
$ grep -rn 'required-skills' .claude/skills/moai/workflows/ internal/template/templates/.claude/skills/moai/workflows/
(출력 없음)
```

따라서 M4에 별도 선언 작업은 추가하지 않는다. 대신 이 판단이 계속 참인지 확인하는 AC를 둔다(`acceptance.md` AC-SEO-020c).

**동반 결정 — `delegation.yaml` 도메인 키 = `frontend`**: `delegation.yaml`의 `domain_skills` 키는 `backend / frontend / database / security / testing / git / authoring`이다(실측). tier를 `optional-pack:frontend`로 정한 이상 도메인 키도 `frontend`가 정합적이다. 두 결정은 같은 축이므로 함께 확정한다.

**이 결정이 바꾸는 것**: `spec.md` REQ-SEO-020(tier 값 명시), REQ-SEO-022(delegation 키 명시), `acceptance.md` AC-SEO-020c 신설, M4 catalog 엔트리의 `tier` 필드.

### B.2 [결정 확정] 감사 게이트 = 중간안 (점검 항목 + 판정 어휘, 수치 임계값·차단 권한 없음)

원문의 가장 특징적인 기여는 10항목 pre-deploy 감사를 BLOCKING 게이트로 패키징한 것이다(research.md §A.4). 개별 점검 항목은 업계 표준이지만 **게이트 패키징과 항목별 수치 임계값은 저작 판단**이므로, 여기가 클린룸 재도출 부담이 가장 큰 지점이다.

| 선택지 | 결과 |
|---|---|
| 순수 레퍼런스 유지 | 스킬은 점검 항목 표만 제공. 게이팅은 `/moai gate` 등 기존 표면에 맡긴다. 재도출 부담 최소, 원문의 최대 가치 미흡수 |
| **중간안 (채택)** | 기존 `## Verification` 불변 섹션이 pre-ship 점검의 그릇이 된다. 점검 항목과 판정 어휘는 담되 수치 임계값·차단 권한은 갖지 않는다. 별도 "BLOCKING gate" H2는 만들지 않는다 |
| 게이트 프레이밍 포함 | 스킬이 pre-ship 점검 목록 + PASS/WARN/FAIL 어휘 + 반복 루프를 서술. 가치 최대, 임계값을 자체 근거로 새로 정해야 함 |

**채택 근거** — (a) ref 스킬은 정의상 에이전트 지식 증강이며 실행 게이트가 아니다. (b) 원문 임계값은 저작 판단이므로 그대로 옮길 수 없고, 새 임계값을 세우려면 근거가 필요한데 본 SPEC에는 그 근거를 만들 조사가 없다. (c) 항목 목록 자체는 표준 실무이므로 안전하게 재작성 가능하다.

**항목 수 — 10항목을 그대로 옮기지 않는다(관례 충돌 해소)**: 중간안을 택하면 점검 목록이 `## Verification` 불변 섹션으로 들어가는데, 그 섹션의 기존 관례는 6-9개 체크박스다(research.md §B.3). 실측으로 재확인했다:

```
$ for f in internal/template/templates/.claude/skills/moai-ref-*/SKILL.md; do
    awk '/^## Verification/{f=1;next} /^## /{f=0} f&&/^- \[ \]/{c++} END{print c+0}' "$f"; done
api-patterns 6 / git-workflow 6 / llm-security 8 / owasp-checklist 6 / react-patterns 6
secops 7 / supply-chain 8 / testing-pyramid 6 / ui-polish 9
```

관측 범위는 정확히 6-9다. 따라서 **원문 10항목은 6-9개로 압축해 옮긴다** — 관례를 넘기지 않는다. 압축은 항목 병합(같은 산출물을 보는 점검을 한 체크박스로)으로 수행하며, 항목을 버려서 달성하지 않는다.

**근거 (a)의 미검증 표시**: v0.1.0은 "`moai-ref-*` 9건 중 차단 권한을 주장하는 사례 0건"을 근거로 제시했으나 9건 본문 전수 판독은 수행되지 않았다. 이 근거는 **미검증 상태로 표시**하며, 결정 자체는 (b)·(c)만으로도 성립한다.

**이 결정이 바꾸는 것**: M3 본문 구성(별도 게이트 H2 없음, `## Verification` 체크박스 6-9개), `spec.md` §C에 임계값·차단 권한 제외 명시.

### B.3 [결정 확정] 접근성 = SEO 인과 항목만 포함

`moai-ref-ui-polish`를 직독한 결과 **실질 중복은 0건**이다(research.md §D.5 — Tier 1 접근성 6종 중 ui-polish가 다루는 항목 없음). 따라서 이것은 "중복 조사" 문제가 아니라 **소유권 경계 결정**이다.

| 선택지 | 결과 |
|---|---|
| `moai-ref-seo`가 전부 포함 | 원문의 의도적 융합(접근성-as-SEO)을 계승. 공백 위험 최저 |
| **SEO 인과 항목만 포함 (채택)** | 크롤러·인덱싱·SERP 표시에 직접 영향하는 항목만 흡수, 순수 조작성 항목은 형제 표면에 위임 |
| 전면 제외하고 상위 계층에 위임 | 형제 SPEC이 접근성 전체를 흡수하도록 남김 |

**채택 범위**:

- **포함(4종)** — 단일 H1과 heading 레벨 구조 / 이미지 `alt` / 서술적 anchor text / 살아있는 in-page fragment 타깃. 이들은 문서 의미 구조로서 SEO가 1차 이해관계자다.
- **제외(3종)** — 키보드 조작 가능성 / 가시적 포커스 / 폼 컨트롤 라벨링. SEO가 아니라 접근성이 1차 이해관계자다. `NOT for:` 절에서 형제 표면에 명시 위임한다.

**항목 수 산술 (6 vs 7 — N7 정정)**: research.md §A.2·§D.5는 Tier 1 접근성 항목을 **6종**으로 센다. 본 결정은 **7종**으로 다룬다. 차이는 누락이 아니라 **분해**다 — research가 "가시적 포커스를 동반한 키보드 조작 가능성"을 1항목으로 묶은 것을, 소유권 경계를 그으려면 나눌 필요가 없으므로 그대로 두어도 되지만, 본 결정은 두 항목이 각각 독립적으로 형제 표면에 위임되어야 함을 분명히 하기 위해 **키보드 조작 가능성**과 **가시적 포커스**로 분리했다. 따라서 6 = 4(포함) + 2(묶인 형태의 제외 2종) 이고, 7 = 4(포함) + 3(분해된 제외 3종)이다. 판정 시 "6종 중 몇 종을 흡수했는가"를 묻는다면 답은 **4종**이며, 이 수는 두 셈법에서 동일하다.

**파급 범위 (v0.1.0이 기록하지 않아 감사에서 지적된 부분)**: 이 결정을 채택하면 요구사항 2곳을 함께 개정해야 한다 — v0.1.0의 REQ-SEO-006 개념 집합에는 접근성 항목이 **0건**이었으므로 결정과 요구사항이 정면 충돌 상태였다. v0.2.0에서 다음을 반영했다.

1. `spec.md` REQ-SEO-006 — 개념 목록에 위 4종 추가
2. `acceptance.md` AC-SEO-006 — 토큰 루프에 4종 각각의 판정 토큰 추가(양성·음성 대조 실측 동반)
3. `spec.md` REQ-SEO-005 — `NOT for:` 절이 제외 3종을 명시적으로 지목하도록 요구
4. `spec.md` §C — `### Out of Scope — 접근성 중 형제 SPEC에 넘기는 항목` 신설

**형제 SPEC 의존 — 현재 기록 수단이 없다(정직한 상태 기록)**: v0.1.0은 "Epic 조율 사항으로 기록한다"고 적었으나 **그 기록을 구현하는 수단이 존재하지 않았다.** 실측:

```
$ find .moai/specs -maxdepth 1 -type d -name 'SPEC-*ABSORB*'
.moai/specs/SPEC-AGENCY-ABSORB-001        # "Agency → MoAI-ADK 흡수 및 Claude Design 통합" — 무관
.moai/specs/SPEC-REF-SEO-ABSORB-001       # 본 SPEC
.moai/specs/SPEC-V3R6-ABSORB-CLEANUP-001  # "Wave 1 Foundation Cleanup" — 무관
```

design-taste-frontend 형제 SPEC도 security/review-rubric 형제 SPEC도 **아직 존재하지 않는다**. 따라서 frontmatter `depends_on:`에 넣을 실제 SPEC ID가 없고, 없는 ID를 지어내지 않는다.

현재 상태를 있는 그대로 기록한다: **이 조율 사항은 현재 어떤 기계적 수단으로도 기록되어 있지 않다.** 기록 지점은 다음 둘이며, 둘 다 형제 SPEC 저작 시점에 발생한다.

1. 형제 SPEC(design-taste-frontend 흡수)의 `spec.md` §A 배경에 "본 SPEC이 조작성 접근성 3종을 흡수한다"를 요구사항으로 고정하고, 그 SPEC의 frontmatter `depends_on: [SPEC-REF-SEO-ABSORB-001]`로 방향을 건다.
2. 그 시점에 본 SPEC의 `spec.md` §C 위임 문단에 형제 SPEC ID를 역참조로 채운다.

**공백 위험(트레이드오프 명시)**: 형제 SPEC이 착수되지 않으면 조작성 3종은 영구 공백이 된다. "전부 포함"은 이 공백 위험이 0이지만 소유권 경계가 흐려진다. 채택안은 경계의 명료함을 택하고 공백 위험을 감수하되, 위 기록 지점을 명시함으로써 위험을 보이게 둔다.

### B.4 [결정 확정] generative-engine-optimization = 전면 제외, `NOT for:` 절에서만 명시

Tier 3 중 가장 덜 정착한 영역이다(research.md §A.2).

| 선택지 | 결과 |
|---|---|
| **전면 제외 (채택)** | 스킬 수명 최장. 최근 관심 주제 미흡수 |
| 포함하되 정착도 표시 | 실무 관심 반영. 표시 수단이 필요하고, 잘못되면 다른 규칙과 동급으로 읽힘 |

**채택 근거** — (a) ref 스킬은 안정된 실무를 담는 표 중심 자산이며, 정착도 표시를 위한 별도 서식 관례가 존재하지 않는다. (b) 배포되는 스킬은 `moai update`로 갱신되므로 정착 이후 추가하는 편이 되돌리기 쉽다. (c) 정착도 표시를 도입하면 그 자체가 새 저작 관례가 되어 9개 기존 ref 스킬과 불일치한다.

**표시 수단 = `NOT for:` 절 단독 (v0.1.0의 대안 제거)**: v0.1.0은 "`NOT for:` 절**이나 본문 말미 1줄**"을 제시했으나, 후자는 REQ-SEO-003과 충돌한다 — 본문은 불변 3종(`## Common Rationalizations` → `## Red Flags` → `## Verification`)으로 **종료**해야 하며, `## Verification` 뒤에 1줄을 붙이면 그 조건이 깨진다. 따라서 **`NOT for:` 절이 유일하게 정합적인 배치**이고, 본문 말미 선택지는 제거한다.

**이 결정이 바꾸는 것**: `spec.md` REQ-SEO-005(`NOT for:` 절이 GEO를 명시하도록 요구), `spec.md` §C에 GEO 제외 명시.

### B.5 [결정 확정 — 이미 선점되어 있었음] docs-site 카운트 드리프트 = 표 행만 추가

skill-guide 산문의 스킬 총계 주장이 **이미 로케일 간 불일치**한다: en 27 / ja 27 / zh 27 / ko 30, 디스크 실측 31(research.md §D.3). 본 SPEC이 만든 부채가 아니다.

| 선택지 | 결과 |
|---|---|
| **표 행만 추가, 산문 카운트 무수정 (채택)** | 변경 최소. 각 로케일의 표와 산문 불일치가 1 더 벌어짐 |
| 각 로케일이 주장하는 숫자를 1씩 증가 | 로케일 내부 상대 일관성 유지. 틀린 숫자를 계속 틀린 채 유지 |
| 4개 로케일 전부 실측(32)으로 재계산 | 드리프트 해소. 카테고리 분해 수치까지 손대야 하므로 범위 확대 |

**채택 근거** — 카운트 정정은 카테고리 분해(Foundation/Workflow/Domain/Reference/Meta) 재계산을 동반하며 4개 로케일 × 3개 등장 지점이라 본 SPEC의 성격(스킬 1개 추가)과 위험 프로필이 다르다.

**이 항목은 v0.1.0에서 이미 다른 두 곳에 선점되어 있었다**: `spec.md` §C 제외 범위가 "docs-site skill-guide 산문의 선행 카운트 드리프트 전면 재정합"을 고정했고, 본 문서 §E M5도 "산문 카운트는 손대지 않는다"로 실행 계획을 확정했다. 즉 마커만 남아 있었을 뿐 결정은 이미 내려져 있었다. v0.2.0에서 마커를 내리고 상태를 결정 기록으로 정정한다.

**이 결정이 바꾸는 것**: 없음(세 곳이 이미 일치). 만약 이후에 3번째 선택지로 뒤집으면 `spec.md` §C와 본 문서 M5를 **둘 다** 개정해야 한다.

---

## §C. Tier 판정

**Tier M**을 권고한다.

| 판단 축 | 관측 |
|---|---|
| 파일 수 | 17 (스킬 2 + catalog 1 + Go 상수 3 + delegation 2 + 워크플로 본문 4[로컬 2 + 템플릿 미러 2] + manifest 1 + docs-site 4). v0.1.0의 "12-16 + skill-routing 1"은 §B.5 표면 집합 오류를 물려받은 값이었다 — skill-routing은 표면이 아니므로 빠지고, 워크플로 본문 4건이 들어와 순증 +2 |
| Go 로직 변경 | 없음. 테스트 상수 3개만 |
| 되돌리기 | 전 항목 되돌리기 쉬움(신규 파일 추가 + 카운트 증가) |
| 위험 집중 | 배포 가드(leak/neutrality) 통과 여부와 클린룸 검증 두 곳 |

Tier S가 아닌 이유: 등록 표면이 6종이고 4개 로케일 동기 의무가 있어 acceptance criteria를 spec.md 인라인으로 압축하면 판정 명령이 뭉개진다. Tier L이 아닌 이유: 아키텍처 결정이 없고 프로덕션 Go 코드 변경이 0이다. design.md·research.md 중 research.md만 필요하며 이미 저작했다.

파일 수 17은 Tier M 범위(5-15)의 상단을 2건 넘는다. Tier 재판정은 하지 않는다 — 초과분 2건은 로컬 파일의 **템플릿 미러**이며 독립적 판단이 없는 기계적 복제이고, 나머지 축(Go 로직 변경 0, 아키텍처 결정 0, 전 항목 되돌리기 용이)은 Tier M 그대로다.

---

## §D. 제약

- 템플릿 트리 우선 저작 → `catalog.yaml` 수기 엔트리 → `make build` → 재생성 catalog 커밋 → 카운트 상수 → 로컬 미러 동기화 (research.md §C.4)
- 마크다운만 담긴 변경은 CI에서 docs-only로 분류되어 race 테스트 잡이 스킵된다. Go 상수 변경을 **같은 변경에** 넣어야 가드가 실제로 돈다 (research.md §C.6)
- 배포 SKILL.md 본문에 요구사항·수용기준 토큰 형태와 날짜를 넣지 않는다(frontmatter `updated:` 제외)
- 언어 우위 표현 금지 — 정규식 교대 목록에 맨 토큰 `go`와 `r`이 포함된다
- `references/` 하위 디렉터리를 만들지 않는다(ref 스킬 선례 0건)

---

## §E. 마일스톤

되돌리기 어려운 결정을 앞에, 기계적 작업을 뒤에 배치한다.

### M1 — 결정 게이트 (완료)

§B의 5건을 결정 라운드로 해소하고 각 마커를 결정 기록으로 치환한다.

**상태: 완료 (v0.2.0).** 5건 전부 확정되었고 파급 범위를 spec.md·acceptance.md에 반영했다. 이 문서에 열린 clarification 마커는 남아 있지 않다. run-phase는 M2부터 시작한다.

### M2 — 클린룸 프로토콜 고정 (저작보다 먼저)

프로토콜을 먼저 확정한다. 저작을 마친 뒤 검증 방법을 정하면 임계값을 결과에 맞추게 되므로 순서가 뒤집히면 안 된다.

- 출처 확인: 원문 배포 디렉터리에 라이선스·NOTICE·COPYING 0건임을 재확인 (acceptance.md AC-SEO-010)
- 원문 `sha256` 고정값이 여전히 일치하는지 확인 (spec.md §E 원문 고정 표)
- 중첩 검사 명령·정규화·임계값 확정 (acceptance.md AC-SEO-011)
- self-trip 시연 실행 및 FAIL 관측 (acceptance.md AC-SEO-012)
- 구조 발산 판정 절차 확정 (acceptance.md AC-SEO-015) — 8-gram·LCS가 잡지 못하는 실패 양식이므로 사람 판독 항목으로 둔다
- 프로토콜을 형제 SPEC이 복사 가능한 형태로 spec.md §B.2에 고정

### M3 — 스킬 본문 저작 (템플릿 트리)

`internal/template/templates/.claude/skills/moai-ref-seo/SKILL.md` 신규 작성.

- frontmatter 3악장 + `metadata` 따옴표 처리 + `progressive_disclosure`
- 도메인 H2 4-10개, 표 중심, 150-220줄
- 불변 3종(`rationalizations` / `red-flags` / `verification`) evolvable-zone 주석 id 정확 일치
- M1 결정 반영: `## Verification` 체크박스 6-9개로 압축된 pre-ship 점검(§B.2), 접근성 4종 포함·3종 `NOT for:` 위임(§B.3), GEO 제외를 `NOT for:` 절에서만 명시(§B.4). 별도 "BLOCKING gate" H2는 만들지 않는다
- 저작 직후 M2 중첩 검사 + 구조 발산 판정 실행. 임계값 초과 시 해당 문단 재작성(임계값 조정 금지)

### M4 — 등록 표면 채우기 + 카운트 상수 (단일 변경)

`moai-ref-secops`가 **실제로 점유한** 표면을 따른다(§B.5 실측 정정 반영 — skill-routing은 표면이 아니다).

- `internal/template/catalog.yaml` 5키 엔트리 수기 추가 (`tier: optional-pack:frontend` — §B.1 결정값)
- `make build` 실행 → 해시 재생성 → 재생성된 catalog 커밋
- Go 상수 3개 증가: `expectedSkillCount` 31→32, `expectedTotal` 41→42, `wantTotal` 41→42
- `delegation.yaml` 로컬·템플릿 양쪽, `domain_skills.frontend` 키 (§B.1 동반 결정)
- 워크플로 본문 4건: `.claude/skills/moai/workflows/review.md`, `.claude/skills/moai/workflows/sync/quality-gates-quality.md`, 그리고 각각의 `internal/template/templates/` 미러 2건
- `internal/template/skills_manifest_test.go` 스팟체크 목록 — secops 표면 등가를 위해 추가하되, 이 목록은 파생 집합(`EmbeddedMoaiSkillNames()`가 디렉터리를 walk)에 대한 회귀 핀일 뿐 새 가드를 만들지 않는다는 점을 인지한다
- 로컬 미러 `.claude/skills/moai-ref-seo/SKILL.md` 동기화 — **템플릿 트리와 바이트 동일해야 하며** AC-SEO-011b가 이를 판정한다

**`.claude/rules/moai/workflow/skill-routing.md`는 건드리지 않는다.** `moai-ref-secops`가 그 파일에 없음을 실측했고(`grep -c 'moai-ref-secops' … → 0`), 거기 있는 3개 스킬 예시는 고정 예시이지 등록 표면이 아니다. 예시 추가를 원한다면 "secops와 동일 표면"이 아닌 별도 근거로 별개 요구사항을 세워야 한다.

**[HARD] 마크다운과 Go 상수를 분리 커밋하지 않는다** — 분리하면 마크다운 커밋 시 가드가 스킵된다.

### M5 — docs-site 4-로케일

en·ja·zh·ko `advanced/skill-guide.md` 레퍼런스 표에 `moai-ref-seo` 행 추가. 산문 카운트는 손대지 않는다(§B.5 결정에 따름).

### M6 — 검증

acceptance.md 전 항목 실행 및 출력 인용.

---

## §F. 위험

| 위험 | 완화 |
|---|---|
| 언어 중립성 정규식이 우발적 문장을 hard-fail | 프레임워크 워크스루 금지, 프로토콜·출력 레이어 서술 고정. M3 직후 **`internal/template/lang_boundary_audit_test.go`의 `TestLanguageNeutrality`**(sentinel `LANG_NEUTRALITY_VIOLATION`, 언어 우위 정규식은 같은 파일 199행)를 단독 실행하고, 동일 파일의 `TestSkillBodyNoLangReference`도 함께 돌린다. (v0.1.0은 이 가드를 `internal_content_leak_test.go`로 오지정했고 `-run` 선택자도 매칭하지 않아 최상위 위험이 무검증 상태였다 — AC-SEO-025에서 정정) |
| 중첩 검사가 공허 통과(0매치를 성공으로 오독) | self-trip 시연을 AC로 강제. 원문 대 원문 비교가 반드시 FAIL이어야 함 |
| 어휘만 바꾸고 원문 구조를 그대로 복제 | 8-gram·LCS 둘 다 어휘 연속성 지표라 이 실패 양식을 통과시킨다. AC-SEO-015 구조 발산 판정(사람 판독 + 판정 기록)으로 별도 방어 |
| 원문(제3자 배포물)이 갱신되어 baseline이 조용히 어긋남 | spec.md §E에 원문 `sha256` 고정. AC-SEO-012가 비교 전에 digest를 단언하므로 갱신이 시끄럽게 드러난다 |
| catalog 해시 수기 작성 | `make build` 산출물만 커밋. 수기 해시 금지 |
| 마크다운 단독 커밋으로 가드 스킵 | M4에서 Go 상수와 동일 변경 강제 |
| 로컬 미러와 템플릿 트리 불일치 — 로컬만 클린룸 통과 | 클린룸 판정(AC-SEO-011/013)의 좌변을 **템플릿 트리**로 통일하고, AC-SEO-011b가 두 파일의 동일성을 판정 |
| 형제 SPEC이 접근성 조작성 항목을 흡수하지 않아 공백 발생 | **현재 기록 수단 없음** — 형제 SPEC이 아직 존재하지 않아 `depends_on:`을 걸 대상이 없다. §B.3에 기록 지점 2곳(형제 SPEC 저작 시 `depends_on` + 본 SPEC §C 역참조 채우기)과 공백 위험을 명시했다. 이 위험은 완화되지 않은 상태로 열려 있다 |

---

## §G. 안티패턴

- 중첩 검사가 실패했을 때 임계값을 올려 통과시키기 (REQ-SEO-014 위반)
- 원문 표를 구조 그대로 두고 단어만 치환하기 — 구조 재사용은 허용되나 문장은 새로 써야 한다. 8-gram·LCS는 이 패턴을 잡지 못하므로 AC-SEO-015 사람 판독이 유일한 방어다
- 가드가 신규 스킬을 지적했을 때 면제 목록에 추가하기 (REQ-SEO-027 위반)
- 등록 표면을 절반만 채우고 완료 선언하기 (`moai-ref-ui-polish` 선례가 그 결과다)
- **판정 명령이 의도한 가드를 실제로 실행하는지 확인하지 않고 통과 선언하기** — `go test -run <pattern>`은 0개 매칭 시 exit 0이므로, 기대 테스트 함수명이 `--- PASS:` 행으로 출력에 등장하는지 확인하지 않으면 검사 자체가 돌지 않은 채 통과한다. v0.1.0의 AC-SEO-025가 정확히 이 상태였다
- 표면 집합·개념 집합 같은 목록을 조사 문서에서 다른 문서로 옮길 때 실측으로 재확인하지 않기 — v0.1.0의 secops 표면 집합은 research.md에서는 정확했으나 이관 중 역전되었다
- Tier 3 수치를 "참고용"이라 적고 그대로 표에 싣기

---

## §H. 교차 참조

- `research.md` §D.1 — Go 상수 3개 실측과 검증 명령
- `research.md` §D.2 — 라이선스 부재 근거 정정
- `research.md` §D.3 — docs-site 등록 표면 및 카운트 드리프트 실측
- `research.md` §D.5 — `moai-ref-ui-polish` 접근성 중복 직독 결과
- `acceptance.md` — 판정 명령
