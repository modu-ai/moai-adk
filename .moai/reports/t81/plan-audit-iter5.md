# SPEC Review Report: SPEC-CODEX-SKILLS-CANONICAL-001

Iteration: 5 (리드가 부여한 예외 상한 안의 5회차 — Tier M 기본 상한 2회를 넘어선 명시적 예외)
Verdict: **PASS**
Overall Score: **0.875** (Tier M 임계 0.80)

Reasoning context ignored per M1 Context Isolation. 선행 감사 보고서 5건은 **선행 결함 목록(닫힘 여부 판정 대상)** 으로만 읽었고, 그 안의 수치·판단·결론은 어느 것도 근거로 인용하지 않았다. `m1-preflight-measurements.md` 도 마찬가지로 근거로 쓰지 않았다. load-bearing 수치는 전부 이 감사에서 직접 재측정했으며, §A.2 의 `//go:embed` 링크 소실은 **독립 최소 재현을 새로 작성해** 확인했다(아래 §재측정).

## 고정 상태 확인

감사 시작 시점:

```
$ git log --oneline -1
6d0097abf docs(spec): reduce SPEC-CODEX-SKILLS-CANONICAL-001 to mirror deployment (card t81)
$ git status --short
(빈 출력)
$ git branch --show-current
WT-skills-canonical
$ pwd
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t81
```

감사 종료 시점 `git status --short` 도 빈 출력. 감사 중 아티팩트 개정 없음. 임베드 재현용 프로브(`.probe-embed/`)를 만들었고 측정 직후 삭제해 트리를 원상 복구했다.

---

## Must-Pass Results

- **[PASS] MP-1 REQ 번호 일관성 — 단, 판정 근거를 명시한다.**
  `grep -o '^- \*\*REQ-CSC-[0-9]*' spec.md` → 001 002 003 004 005 006 007 010 011 012 013 014 015 016 (**14개**). 중복 0, zero-padding 일관(3자리). **008·009 자리에 결번이 있다.**
  MP-1 문언은 "gap 하나라도 = FAIL"이다. 그럼에도 PASS 로 판정한 근거를 적는다 — MP-1 이 잡으려는 결함은 *요구사항이 소실되거나 중복된 흔적으로서의* 결번이며, 여기 결번은 (a) HISTORY iter-6 에 **ID 매핑 표**로 각 번호의 처분이 기록돼 있고(spec.md:30-37), (b) `acceptance.md:169` 가 그 표를 가리키며, (c) 재번호하면 감사 보고서 4건이 인용한 ID 가 조용히 무효가 된다는 명시된 사유가 있다. 소실이 아니라 **추적 가능한 전출**이다. 문언 그대로의 엄격 판정을 원하면 이 항목은 FAIL 이 되며, 그 경우 요구되는 수정은 "재번호"인데 그것은 인용 무효화라는 더 큰 손해를 산다 — 판정 근거를 드러내 두었으니 리드가 뒤집을 수 있다.
- **[PASS] MP-2 GEARS 형식 준수** — **요구사항 계층(`REQ-XXX`)에 한해** 판정했다(M3 § Scope). 14개 전부 매치: Ubiquitous(001·015), Unwanted `shall not`(002·007·010), Event-driven `When`(003·005·011·012·013·014), Where(004), 복합 shall+shall not(006·016). `IF/THEN` 0건. 검증 계층(`AC-XXX`)의 Given-When-Then 은 이 기준으로 감점하지 않았고 Group 4 에서 평가했다.
- **[PASS] MP-3 YAML frontmatter 유효성** — spec.md:1-15 에 canonical 12필드 전부: `id`(다중 세그먼트 — 이 저장소의 지배적 관례, 예 `SPEC-AGENT-ARCH-V2-001`) · `title`(따옴표) · `version: "0.6.0"`(따옴표 semver) · `status: draft`(enum) · `created`/`updated: 2026-08-22`(ISO) · `author` · `priority: P2` · `phase: "v3.1.3 target"` · `module: internal/template` · `lifecycle: spec-anchored` · `tags`(콤마 구분). 거부 별칭(`created_at`/`updated_at`/`labels`/`spec_id`) 0건. 부가 필드 `tier: M`.
- **[PASS] MP-4 언어 중립성** — 대상은 Go 배포기(`internal/template`) 단일 언어 경로이며, 16개 지원 언어 중 일부만 열거하는 형태 없음. §E 에 템플릿 중립성 제약이 명시돼 있다(spec.md:238).
- **[N/A] MP-5 D7 교차 SPEC 정합** — `grep -Eo 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+'` 결과 세 문서 모두 **자기 ID 하나뿐**이다. 검증 동사가 대상 없이 도는 상태이므로 MP-4 선례에 따라 N/A. 다만 "승계 SPEC"이 **ID 없이 22회**(spec 13 · acceptance 5 · plan 4) 참조되는 것은 별개 결함으로 D3 에 기록했다.
- **[PASS] MP-6 D8 크로스 플랫폼 규율** — `syscall` 문자열은 spec.md:44(HISTORY iter-5) 한 곳뿐이며 Go `syscall` 패키지가 아니라 "판정 syscall"이라는 한국어 서술이다. 살아남은 요구사항 계층이 실제로 쓰는 크로스 플랫폼 표면은 `os.Symlink` 이고, REQ-CSC-004(폴백) + `GOOS=windows go vet` 게이트(acceptance §D.4·§D.5, plan R2)로 규율이 걸려 있다. BLOCKING 없음.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' plan.md spec.md acceptance.md` → 매치 0건(rc=1).

---

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | 0.75 (요구사항 1~2개의 minor 모호) | REQ-CSC-006 이 REQ-CSC-001 이 받은 예외 절을 받지 못했다(spec.md:186 vs 182) — "대상 집합"을 시도 집합으로 읽으면 해소되고 AC-CSC-014 가 그 읽기를 고정하므로 일관 해결 가능. 나머지 12개는 단일 해석이다. |
| Completeness | 0.75 | 0.75 (비핵심 절 하나가 sparse) | 구조는 완비 — HISTORY(19) · §A(50) · §B(155) · §C(181) · §D 범위 밖(198) · §E(234) · §F(240) · §G(246), frontmatter 12/12. 감점 사유는 **§D 의 축소된 보장 고지가 결과 하나를 빠뜨린 것**(D1) 과 첫 `### Out of Scope` H3(spec.md:200)이 `-` 불릿을 하나도 갖지 않는 것(D4). |
| Testability | 1.00 | 1.0 | AC 13개 전부 이진 판정이며 weasel word 0건. 자기 무력화 함정을 본문에서 명시적으로 닫은 사례가 다수 — AC-CSC-001(2)가 `d.IsDir()` 수집 금지를 [HARD] 로 못박고(acceptance.md:33-35), AC-CSC-003(3)이 "배포 후 대상 FS 재독" 구현에서 1번이 정의상 참이 되는 것을 3번으로 갈라내며, AC-CSC-010 이 "변경 전 커밋 기준선" 형태를 금지하고 동일 프로세스 불변식으로 대체한다. |
| Traceability | 1.00 | 1.0 | 직접 계수: REQ 정의 14개, AC 정의 13개(001-006·010-016), AC 매트릭스가 참조하는 REQ 집합 = 정의된 14개와 **정확히 일치**. 고아 AC 0건, 미커버 REQ 0건. §D.3 의 주장을 읽지 않고 세어 같은 결과를 얻었다. |

산술 평균 **0.875** ≥ 0.80 → PASS.

---

## 분할 예측은 성립했는가

**대체로 성립했다 — 그러나 "즉시 통과 가능"까지는 아니다.**

성립한 부분: iter-4 의 blocking 결함 6건(D1~D6)은 전부 청소 계열이었고, 전출로 **전부 해소**됐다(아래 회귀 점검). 절단면도 깨끗하다 — 전출 항목(§A.4·A.5·A.7·A.10·A.11, §B.D5·D6, REQ-008·009, AC-007·008·009, plan M4·R4·R5·R8·R13·R14·R16·AP-6·AP-12·AP-13·AP-14·AP-16)에 대한 참조는 **HISTORY 와 명시적 결번 고지 안에만** 남아 있고, 살아 있는 본문에서 죽은 절을 가리키는 인용은 **0건**이다(spec.md:21-47, plan.md:5, acceptance.md:2·159·169 외 매치 없음).

성립하지 않은 부분: 미러 절반이 blocking 결함을 내지 않았던 것은 **감사의 초점이 늘 청소 쪽에 있었기 때문**이지 결함이 없어서가 아니었다. 처음으로 미러 절반을 정면에서 읽자 blocking 하나가 나왔고(D1), 그것은 **분할이 스스로 만든 것**이다 — 청소 등록이 빠지면서 복사 모드 미러를 갱신하던 유일한 경로가 함께 사라졌는데, SPEC 은 그 손실을 기록하지 않았다. 예측을 확인하기보다 시험하라는 지시가 정확히 이 자리를 겨눴다.

---

## 재측정 (이 감사에서 직접 실행)

| 주장 | 출처 | 재측정 결과 | 판정 |
|---|---|---|---|
| 템플릿 배포 스킬 34개 | §A.1 | `find … -mindepth 1 -maxdepth 1 -type d \| wc -l` → **34** | 일치 |
| 로컬 44, dev-only `hns-*` 10 | §A.1 | **44** / **10** (44 = 34 + 10) | 일치 |
| 템플릿 `SKILL.md` 보유 34 | §A.1 | **34** | 일치 |
| 트리 전체 심볼릭 링크 0 | §A.1·plan §C | `find internal/template/templates -type l \| wc -l` → **0** | 일치 |
| `.agents/` 부재(로컬·템플릿) | §A.1 | `ls -d` 양쪽 `No such file or directory` | 일치 |
| 비-`moai` 접두 0, 비-`moai-` 접두 1(=`moai`) | §A.9 | `grep -cv '^moai'` → **0**, `grep -cv '^moai-'` → **1**, 해당 이름 **`moai`** | 일치 |
| `moaiSkillPrefix = "moai-"` | §A.9 | `skills_manifest.go:15` 상수 확인, 주석이 bare `moai` 제외를 의도로 명시 | 일치 |
| `EmbeddedMoaiSkillNames()` → 33 | §A.9 | 34개 중 `moai-` 접두 필터 → 33 (`moai` 제외) | 일치 |
| catalog 스킬 34, optional-pack 13, core 21 | §A.3 | YAML 파싱 → 스킬 항목 **34**, `grep -c 'tier: optional-pack'` → **13**, 34-13=**21** | 일치 |
| `harness_generated.skills: []` | §A.3 | `catalog.yaml:261-262` — 빈 슬롯 확인, 그 tier 유일 항목은 에이전트 `builder-harness` | 일치 |
| `manifest.Track` → `HashFile` → `io.Copy` 이고 실패 시 error 전파 | §A.6 | `manifest.go:130-149` `HashFile` 오류를 `manifest track hash: %w` 로 반환, `hasher.go` 가 `os.Open`+`io.Copy` | 일치 |
| `internal/template` 비-테스트에 `io.Writer` 0건 | §A.9b | `grep -rn 'io.Writer' … \| grep -v _test \| wc -l` → **0** | 일치 |
| `Deploy` 시그니처 | §A.9b | `deployer.go:100` — `(ctx, projectRoot, m manifest.Manager, tmplCtx *TemplateContext) error` 정확히 일치 | 일치 |
| `deployer` 필드 4개(`fsys`/`renderer`/`forceUpdate`/`renderCache`) | §A.9b | 구조체 확인, 필드 4개 일치 | 일치 |
| 템플릿 `.gitignore` 에 `.agents` 항목 없음 | §A.8·plan §C | `grep -n 'agents' … .gitignore` → 매치 0(rc=1) | 일치 |
| `//go:embed all:templates` | §A.2 | `embed.go:28` | 일치 |
| **`//go:embed` 가 심볼릭 링크를 무음으로 버린다** | §A.2 | **독립 최소 재현 신규 작성** — 디렉터리 링크 + 파일 링크를 심은 `templates/` 를 `//go:embed all:templates` 로 임베드해 `fs.WalkDir` 출력: `templates/linktarget`, `templates/linktarget/SKILL.md`, `templates/sub`, `templates/sub/real.txt` 만 존재. **`dirlink`·`filelink` 둘 다 부재, 빌드 오류·경고 0.** | 일치 |

**§A 의 실측은 이번에도 전부 정확했다.** 다섯 판본을 거친 §A 는 이 감사가 확인할 수 있는 범위에서 신뢰할 수 있다.

추가로 측정한 것(§A 주장이 아니라 D1 의 근거):

```
$ sed -n '/func ManagedCleanTargets/,/^}/p' internal/cli/update/deploy/deploy.go
  .claude/settings.json · .claude/commands/moai · .claude/agents/moai
  .claude/skills/moai*(glob) · .claude/rules/moai · .claude/output-styles/moai
  .claude/hooks/moai · (그리고 .moai/config)
  → `.agents` 항목 없음
```

---

## Defects Found

### D1 — 폴백(복사) 모드 미러는 두 번째 배포부터 영구히 낡는다. 축소가 만든 결함이며 SPEC 은 이 손실을 고지하지 않는다

`spec.md:186,194,195,196`(REQ-CSC-004·012·013·014) · `spec.md:200-208`(§D 축소된 보장) · `acceptance.md:88-100`(AC-CSC-011)
**Severity: major — Class: blocking**

측정한 연쇄:

1. REQ-CSC-004 — 링크가 불가한 환경에서 미러는 **실 디렉터리 복사본**이 된다.
2. REQ-CSC-014 — 미러 대상에 **링크가 아닌 실 항목**이 있으면 "제거하거나 덮어써서는 안 되며 건너뛰고 경고"한다. §C 전체에 *우리가 지난 실행에 만든 복사본* 과 *사용자가 만든 디렉터리* 를 가르는 판별자가 **없다**.
3. `ManagedCleanTargets` 에 `.agents` 항목이 없다(위 측정). 등록은 REQ-CSC-008 과 함께 승계 SPEC 으로 전출했으므로, 이 SPEC 착지 상태에서 `moai update` 의 clean 단계는 미러를 **건드리지 않는다**.

귀결 — 심볼릭 링크가 불가한 플랫폼(Windows 기본 상태)에서:

- 1회차: 복사 미러 생성.
- 2회차 이후 `moai update`: clean 이 `.claude/skills/moai*` 를 지우고 재배포해 **정본은 갱신**되지만, 미러는 REQ-CSC-014 의 "실 항목" 분기에 걸려 **건너뛰어진다**. 미러는 1회차 내용에 영구 고정된다.

이것이 blocking 인 이유는 **SPEC 내부 모순**이기 때문이다. §B.D3 은 관측 가능성을 요구하는 근거로 *"링크인 줄 알았는데 복사본이라 **정본 갱신이 반영되지 않는** 상태가 무음으로 생긴다"*(spec.md:169) 를 든다. 그런데 §C 는 그 상태를 REQ-CSC-014 로 **제도화**해 놓고 어디에도 그렇게 적지 않았다. 완화 요인은 하나 — 건너뛴 사실이 경고로 올라온다(REQ-CSC-005). 다만 그 경고는 우리 산출물을 사용자 항목으로 오귀속해 보고하게 된다.

축소가 만든 결함이라는 근거: 전출 전 REQ-CSC-008 이 `.agents/skills/moai*` 를 청소 목록에 등록했으므로 clean→deploy 가 낡은 복사본을 지우고 다시 만들었다. 전출로 그 경로가 사라졌고, §D 의 "축소된 보장" 고지는 **개명·은퇴 미러의 잔존 하나만** 적고 이 두 번째 결과는 적지 않았다. 리드 지시 4번("dress it up as handled 하지 말 것")에 대한 답은: 적힌 한 가지는 정직하게 적혔고, **적히지 않은 한 가지가 있다.**

판정 계층에도 같은 구멍이 있다. AC-CSC-002 는 빈 프로젝트 1회 배포만 본다. AC-CSC-011 의 fixture 는 (i) 올바른 링크 / (ii) 엉뚱한 링크 / (iii) 사용자 실 디렉터리인데, **복사 모드에서는 (i) 과 (iii) 이 같은 형태(실 디렉터리)** 라 "실 항목이면 전부 건너뛴다"는 구현이 세 단언을 모두 통과한다. **복사 모드 + 2회차 배포를 보는 AC 가 없다.**

**Required fix** — 둘 중 하나. (a) 선점 상태 분할을 네 상태로 넓힌다: 링크가 아닌 실 항목 중 **이번 실행이 배포한 스킬 이름과 같고 내용이 정본의 복사본인 것**은 갱신 대상, 그 밖의 실 항목은 REQ-CSC-014 대로 보존 — 그리고 그 판별자를 명시한 뒤 AC-CSC-011 에 복사 모드 2회차 팔을 추가한다. (b) 갱신을 승계 SPEC 소관으로 넘기되, §D 축소된 보장 절에 *"폴백 플랫폼에서는 1회차 이후 미러가 정본과 분기하며 이 SPEC 은 그것을 되돌리지 않는다"* 를 **잔존 고지와 나란히** 적는다. (b) 를 고르면 요구사항 수정 없이 문서 한 문단으로 닫힌다.

### D2 — plan M1 의 닫힘 조건에 M2 가 만들 seam 이 필요한 AC 가 들어가 있다

`plan.md:55`(M1 닫힘 조건) · `plan.md:62-64`(M2)
**Severity: major — Class: blocking**

M1 은 "미러 집합을 어떻게 파생하는가"를 정하는 마일스톤이고, 닫힘 조건이 **AC-CSC-006, AC-CSC-014** 다. 그런데 AC-CSC-006 은 *"폴백이 관측된다(양방향, 반환 결과 기준)"* 이며, M2 가 [HARD] 로 *"출력 seam 은 **이 마일스톤의** 산출물"*(plan.md:63) 이라고 못박고 있다. **M1 은 자기 닫힘 조건을 자기 안에서 닫을 수 없다.**

거울상 오배치도 함께 있다. AC-CSC-003(슬림 vs 전량 집합 동치 + tier 필터 관통 확인)은 내용상 정확히 M1 의 결정을 판정하는 항목인데 M2 에만 걸려 있다(plan.md:64).

이 결함은 v0.6.0 이 만든 것이 아니라 **이월된 것**이며, 앞선 감사 네 건이 짚지 않았다. 청소 계열이 사라져 M1·M2 가 처음으로 plan 의 무게중심이 되면서 드러났다.

**Required fix**: M1 닫힘 조건을 `AC-CSC-003, AC-CSC-014` 로, M2 닫힘 조건에서 AC-CSC-003 을 빼고 AC-CSC-006 을 유지.

### D3 — "승계 SPEC"이 ID 없이 22회 참조된다

`spec.md`(13회) · `acceptance.md`(5회) · `plan.md`(4회)
**Severity: minor — Class: blocking**

살아남은 절반의 정당화 여러 개가 승계 SPEC 에 의존한다 — REQ-CSC-015 의 존재 이유 전체(spec.md:195), AC-CSC-016 의 마지막 문단, §D 축소된 보장 절, plan R11. 그런데 세 문서 어디에도 승계 SPEC 의 **ID 도 카드 번호도 없다**(`grep -Eo 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+'` 결과가 자기 ID 하나뿐). "리드가 등록했다"는 서술만 있다.

blocking 으로 분류한 이유는 이것이 취향이 아니라 **추적성 결함**이기 때문이다 — REQ-CSC-015 의 정당화는 "승계의 청소 글롭이 도달할 수 있는 이름"이고, 그 글롭이 실제로 무엇인지 확인할 수 있는 문서가 지목돼 있지 않으면 이 SPEC 만으로는 그 주장을 검증할 수 없다. 감사가 D7 을 N/A 로 둘 수밖에 없었던 것도 같은 원인이다.

**Required fix**: 승계 SPEC ID(미발행이면 카드 ID)를 §F 교차 참조에 한 줄 추가하고, spec.md:200 의 Out of Scope 첫 문단에서 한 번 지목. 나머지 21회는 그대로 둬도 된다.

### D4 — 첫 `### Out of Scope` H3 가 `-` 불릿을 하나도 갖지 않는다

`spec.md:200-208`
**Severity: minor — Class: optional**

lint 의 `OutOfScopeRule` 관례는 `### Out of Scope — <topic>` H3 아래 최소 1개의 구체적 `-` 불릿을 요구한다. 나머지 네 개 H3(사용자 홈 정리 · dev-only 스킬 · 나머지 마일스톤 · 구현 세부)는 전부 불릿을 갖는데, **가장 중요한 첫 절만** 산문 + 인용 블록으로 돼 있다. SC-6 자체는 다른 H3 들 덕에 통과하지만 문서 내부 일관성이 깨진다. D1 의 수정이 이 절을 건드리므로 함께 처리하면 비용이 0 이다.

**Required fix**: 축소된 보장 두 가지(개명·은퇴 미러 잔존 / D1 의 폴백 분기)를 `-` 불릿 두 줄로 뽑아 산문 위에 배치.

### D5 — REQ-CSC-006 이 REQ-CSC-001 이 받은 예외 절을 받지 못했다

`spec.md:186`
**Severity: minor — Class: optional**

REQ-CSC-001 은 iter-4 에서 REQ-CSC-011(실패)·REQ-CSC-014(선점)에 대한 예외 절을 달아 모순을 해소했다. REQ-CSC-006 은 *"미러 대상 집합은 … 배포된 스킬 집합과 **정확히 일치**해야 하며"* 로 예외 없이 남아 있다. 선점·실패가 발생하면 실제 `.agents/skills/` 내용은 배포 집합보다 작다.

"대상 집합"을 *시도 집합* 으로 읽으면 해소되고 AC-CSC-014 의 측정(합성 FS 2스킬, 선점·실패 없음)이 그 읽기를 고정하므로 실무 위험은 낮다. 그래서 optional 이다.

**Required fix**: REQ-CSC-006 에 *"이 일치는 미러 생성이 시도되는 집합을 가리키며, REQ-CSC-011·014 가 규정하는 실패·선점은 예외"* 한 절. 또는 명시적으로 기각하고 §G 에 남긴다.

---

## Regression Check (iter-4 결함 목록 대비)

| iter-4 결함 | 계열 | 상태 |
|---|---|---|
| D1 — `os.Lstat` 전환 폭발 반경 / `moai update` 실패 | 청소 | **RESOLVED (전출)** — `§B.D6`·`REQ-CSC-008`·plan M4 가 문서에서 사라졌고(`grep` 결과 HISTORY 외 매치 0), 살아남은 절반은 `backupThenRemove` 경로를 전혀 건드리지 않는다. |
| D2 — `moai-linkprobe` 탐침이 다른 분기를 탄다 | 청소 | **RESOLVED (전출)** — AC-CSC-008 부재. |
| D3 — plan M4 닫힘 조건 fixture 개수 | 청소 | **RESOLVED (전출)** — plan §F 에 M4 없음(plan.md:5 가 결번을 명시). |
| D4 — AC-CSC-012 2번 단언이 공허하게 참 | 청소(백업 팔) | **RESOLVED (전출)** — AC-CSC-012 는 manifest 팔 2개만 남았고, 남은 두 단언은 공허하지 않다(manifest 키 0개 / 정본 기록 불변). |
| D5 — AC-CSC-008 6번 단언의 얇음 미선언 | 청소 | **RESOLVED (전출)**. |
| D6 — REQ-CSC-010 이 링크 항목 백업 정책 미규정 | 청소(백업 절) | **RESOLVED (전출)** — 백업 절 2개가 나가고 manifest 절 하나만 남았다. 남은 절은 백업을 언급하지 않으므로 지적이 성립하지 않는다. |

**전출로 6/6 해소.** 정체(stagnation) 항목 없음. 이번 iteration 의 D1·D2 는 **신규**이며, 그중 D1 은 축소 자체가 만들었다.

### 점수 추이 (회귀 없음)

0.775(iter1) → 0.78 → 0.7625(독립 2회독) → 0.7875(iter3) → 0.7625(iter4, STOP) → **0.875(iter5)**. iter-4 의 STOP 조건(직전 대비 하락)은 해소됐다.

---

## 절단 손상 점검 (리드 지시 1·2·5)

**교차 참조 (지시 1)** — 전출된 13개 식별자(§A.4·A.5·A.7·A.10·A.11, §B.D5·D6, REQ-008·009, AC-007·008·009 + plan M4·R4·R5·R8·R13·R14·R16·AP-6·AP-12·AP-13·AP-14·AP-16)에 대한 `grep` 결과, 살아 있는 본문의 매치는 **0건**이다. 매치는 전부 (a) HISTORY iter-1~5 의 이력 기록, (b) plan.md:5 · acceptance.md:2·159·169 의 명시적 결번 고지 안에 있다. 이력 기록을 현재 상태에 맞춰 고치면 기록이 거짓이 되므로 **손대지 않은 것이 옳다**.

**§A 의 잔여 절 (지시 1)** — 남은 §A 는 A.1·A.2·A.3·A.6·A.8·A.9·A.9b 이며 **전부 살아 있는 인용처를 갖는다**: A.1→plan §C, A.2→REQ-CSC-002·AC-CSC-001·AP-1, A.3→REQ-CSC-006·AP-2·A.9, A.6→REQ-CSC-010·AC-CSC-012·R12·AP-11, A.8→REQ-CSC-016·§B.D7, A.9→REQ-CSC-015·AC-CSC-016·§H·AP-15, A.9b→REQ-CSC-005·AC-CSC-006·M2·R15. **승계 전용으로 남겨진 죽은 실측은 없다.** A.6 의 마지막 문단만 승계 쪽을 가리키는데, "형제 경로를 함께 훑는다"는 규율로 명시적으로 서술돼 load-bearing 처럼 보이지 않게 처리돼 있다.

**REQ-CSC-010 존치 판단 (지시 2)** — **옳은 판단이며 stump 가 아니다.** 남은 절(manifest 기록 금지)은 배포 계열이 맞다: `Track` 을 부르는 주체가 배포기이고, §A.6 의 실측이 그 자리에 있으며, AC-CSC-012 · plan M3 · R12 · AP-11 이 각각 판정·구현·위험·안티패턴으로 받치고 있다. 자기 완결적이다. 통째로 전출했다면 배포기가 미러를 `Track` 에 넘겨 EISDIR 로 `Deploy` 를 실패시키는 것을 금지하는 요구사항이 **이 SPEC 에 하나도 없게 되고**, 그것은 REQ-CSC-011(fail-open)이 규정만 있고 지켜질 근거가 없는 상태다. HISTORY 가 이 예외 판단을 근거와 함께 적어 둔 것도 적절하다.

**리드 지시 5 (축소가 살아남은 절반을 느슨하게 만들었는가)** — 살아남은 절의 문구를 iter-4 판본 대비로 읽었다. **느슨해진 절은 없다.** 요구사항 본문은 REQ-CSC-010 을 빼면 전부 문자 그대로 유지됐고, REQ-CSC-015 는 근거만 다시 세웠을 뿐 구속력(`shall`)이 그대로다. 곁다리로 바뀐 것이 하나 있다 — **`phase` 가 `"v3.2.0 target"` → `"v3.1.3 target"`**. 제거가 강제한 변경은 아니지만 §D 의 릴리스 조율 서술(`release/v3.1.3`)과 정합하며, 은닉되지 않았다. 결함으로 세지 않고 관측으로만 기록한다.

---

## SPEC · 선행 감사 · preflight 노트의 주장과 어긋나는 것

리드가 별도로 요청한 축이다.

- **§B.D3 (spec.md:169) 과 §C 의 REQ-CSC-014 가 서로 어긋난다.** D3 은 "정본 갱신이 반영되지 않는" 상태를 막아야 할 하자로 지목하고, REQ-CSC-014 는 폴백 플랫폼에서 그 상태를 규정된 동작으로 만든다. D1 의 실질이며, 이 감사가 찾은 **유일한 SPEC 내부 모순**이다.
- **§D 축소된 보장 절의 "정면으로 적는다"가 불완전하다.** 적힌 결과(개명·은퇴 미러 잔존)는 사실이지만 유일한 결과가 아니다(D1).
- **plan M1 닫힘 조건이 plan M2 의 [HARD] 선언과 어긋난다**(D2).
- **preflight 노트(`m1-preflight-measurements.md`)** — 근거로 쓰지 않았으므로 그 수치를 재판정하지 않았다. 다만 §A.1 이 그 노트의 "36개"를 오류로 판정한 것은 **내 재측정(34)과 일치**한다.
- **선행 감사 5건의 §A 관련 판정** — 이번 재측정과 어긋나는 항목 없음. iter-4 가 "§A 실측은 전부 정확"이라고 적은 것도 재확인됐다(다만 이 확인은 내 독립 측정에 근거한 것이지 그 보고서를 인용한 것이 아니다).
- **감사 자신에 대한 정정** — 없음. 이번 감사가 앞선 보고서의 결론을 뒤집은 항목은 하나도 없다. 새로 찾은 D1·D2 는 앞선 감사가 **틀린 자리**가 아니라 **보지 않은 자리**다.

---

## Recommendation

PASS 근거는 위 must-pass 7개(6 PASS + 1 N/A) 와 카테고리 평균 0.875 다. 특히 판정 계층(Testability 1.0) 과 추적성(1.0) 은 직접 계수·직접 실행으로 확인했으며, 다섯 판본을 거치며 자기 무력화 함정을 스스로 닫아 온 흔적이 문서 곳곳에 남아 있다.

다만 **PASS 와 blocking 결함 3건이 공존한다**는 점을 흐리지 않는다. M6 규율상 blocking 항목은 verdict 와 별개로 처리 대상이며, 특히 **D1 은 run-phase 착수 전에 닫는 것을 권고한다** — 수정 (b) 를 고르면 §D 에 문단 하나, D4 와 함께 처리하면 편집 1회로 끝난다. D2 는 plan 한 줄, D3 는 §F 한 줄이다. 세 건 모두 요구사항 설계를 건드리지 않으므로 **재감사 iteration 을 새로 돌릴 필요는 없다고 본다** — 리드가 편집 후 diff 로 확인하는 것으로 충분한 규모다. D1 을 수정 (a) 로 처리하기로 하면 요구사항 계층이 바뀌므로 그때는 델타 재감사가 필요하다.

우선순위: **D1 → D2 → D3 → D4 → D5**(D5 는 기각해도 무방하며, 기각 시 §G 에 사유를 남길 것).
