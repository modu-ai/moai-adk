---
id: SPEC-VERIFICATION-COMPLETENESS-001
title: "t241 하네스 규칙 6건의 규칙 파일 착지 — 검증 완결성(verification completeness) 규칙 + always-loaded 예산 영향 측정"
version: "0.2.1"
status: completed
created: 2026-08-25
updated: 2026-08-25
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: .claude/rules/moai/development
lifecycle: spec-anchored
tags: "rules, verification, check-discipline, template-mirror, always-loaded-budget"
era: V3R6
tier: M
---

# SPEC-VERIFICATION-COMPLETENESS-001 — t241 하네스 규칙 6건 착지 (검증 완결성)

## HISTORY

- 2026-08-25 (plan-phase, iter-2 수정, v0.2.1) — plan-audit review-2 반영 (PASS-WITH-DEBT 0.91, D1–D10 전 CLOSED). **N1**: AC-VC-007 계측 명령의 표-셀 내 이스케이프 표기(파이프 앞 백슬래시)가 셸 BRE에서 교대로 읽혀 verbatim 실행 시 191(=전 행) 관측 — 교정형(백슬래시 제거, 실측 6)을 계측기 펜스 CMD-7로 이전하고 표 셀에서 파이프 포함 명령을 배제했다(표 셀 내 파이프는 GFM 렌더를 깨뜨리므로 파이프 포함 명령은 셀 밖 펜스에 두는 규칙).
- 2026-08-25 (plan-phase, iter-1 수정, v0.2.0) — plan-audit review-1 반영 (PASS-WITH-DEBT 0.81, Tier M 문턱 0.80; BLOCKING 0 / SHOULD-FIX 6 / MINOR 4 — 전량 이번 패스 적용). **D3**: §A A-6 측정 정정 — 재측정 템플릿 VCI 8,310B vs 로컬 8,224B(템플릿이 86B 더 큼; 종전 '더 작다' 오기록), 분기 성격을 '근거행 생략'에서 '통째 중립화 재작성'(diff 36행 실측)으로 재규정 — D4 바이트 동일 채택 강화. **D1**: card rule 4의 '옳은 이유로 RED'(why-red) 요소 — 세 번째 실패 방향(무관 선행 파일의 wrong-reason RED)·RED 셀 why-red 서술·녹색 경로 실격 규칙을 plan §A.3 §2와 REQ-VC-001 축 기술에 추가. **D2**: 돌연변이 탐침 귀속 정정(t197 제안 (b) 원천, t228은 룰 쌍 원리 관측). **D5**: CMD-N 정규식 확대(주입 프로브 4종·초안 어휘 실측 검증) + CI 보정 가드 2종 명시. **D7**: 7번째-규칙 제외의 실제 근거 재인용(t261 '이 여섯' + t241 '한 규칙의 사례'). **D10**: REQ-VC-004 스코프 `**/` 접두 정합화. 연동 수정: acceptance.md AC-VC-009 신설(D4), AC-VC-006/007 측정 정렬(D8), plan.md §A.5 예측 장부(D6)·VC-2 행 재귀속(D2).
- 2026-08-25 (plan-phase, v0.1.0) 최초 작성. 카드 t261 (Class C, Tier M, "t241 하네스 규칙 6건을 .claude/rules/ 에 착지"). 규칙 6건은 카드 t241 본문에 2026-08-24~25 라인 8개가 누적한 관측에서 왔다. 산출 축 2개 — (a) 6규칙의 규칙 파일 착지, (b) 착지가 always-loaded 예산에 미치는 영향의 실측. 카드 경고 반영: "규칙을 옮기면서 근거(어느 카드 어느 관측에서 나왔는지)를 떨어뜨리지 말 것 — 근거 없는 규칙은 다음 사람이 무시한다" → 근거는 (i) 규칙 파일 안에 중립적 결함 형태로 인라인, (ii) 카드/반복 단위 원천 귀속은 plan.md §A.4 근거 행렬로 이중 보존. 측정 기반선은 워크트리 `WT-harness-rules` HEAD `32d2221fa` (= 당시 origin/main) 에서 plan-phase 실측했다.

## §A. 검증된 기반선 (Measured Baseline)

이 절의 모든 주장은 2026-08-25 본 워크트리(`.claude/worktrees/t261`, HEAD `32d2221fa`, branch `WT-harness-rules`) 실측이다. 사전 지식이 아니다. 명령 본문은 파이프(`|`)가 표 셀 렌더를 깨뜨리지 않도록 아래 목록에 코드 펜스로 두고 표에서는 CMD 번호로 인용한다.

```text
CMD-1 (naive 열거):  grep -rLE '^paths:' .claude/rules/moai --include='*.md' | sort
CMD-2 (naive 바이트): grep -rLE '^paths:' .claude/rules/moai --include='*.md' | xargs wc -c | tail -1
CMD-3 (frontmatter 한정 열거 — 본 SPEC의 계측기):
  awk 'function flush(){ if(prev!="" && !has) print prev } FNR==1{ flush(); prev=FILENAME; has=0; infm=($0=="---"); next } infm && $0=="---"{ infm=0; next } infm && /^paths:/{ has=1 } END{ flush() }' .claude/rules/moai/**/*.md | sort
CMD-4 (양성 대조):  CMD-3 결과에 askuser-protocol.md 존재 확인
CMD-5 (음성 대조):  CMD-3 결과에 spec-frontmatter-schema.md 부재 확인
```

| # | 주장 | 근거 (명령 → 관측) |
|---|------|-------------------|
| A-1 | `.claude/rules/moai/` 룰 파일 82개, always-loaded 서브셋은 14파일 179,081바이트 | CMD-1 → 14행 목록(하단); CMD-2 → `179081 total` |
| A-2 | CMD-3(frontmatter 한정 계측기)은 CMD-1과 동일한 14행을 반환 — 현재 트리에서 두 명령이 일치. 카드가 지적한 naive 명령의 허위-제외 입력(본문 행이 정확히 `paths:`로 시작)은 현재 트리에 존재하지 않는다. 그럼에도 계측기는 CMD-3을 정본으로 쓴다(더 엄격). | CMD-3 → CMD-1과 동일 목록 |
| A-3 | 계측기 판별력(양성/음성 대조 모두 관측) | CMD-4 → `askuser-protocol.md` 목록 포함(always-loaded 클래스 생성 가능); CMD-5 → `spec-frontmatter-schema.md` 부재(paths 스코프 판별 가능) |
| A-4 | RED-now: 신규 룰 파일의 두 착지 지점 모두 부재 | `test -f .claude/rules/moai/development/verification-completeness.md` → rc=1; `test -f internal/template/templates/.claude/rules/moai/development/verification-completeness.md` → rc=1 |
| A-5 | 파일명 토큰 base-0 | `grep -rn 'verification-completeness' .claude/rules internal/template/templates` → 0행 |
| A-6 | 템플릿 미러의 중립화 선례: 로컬 판이 SPEC-ID를 실어도(로컬 VCI `SPEC-` 2회) 템플릿 판은 싣지 않는다(VCI·`main-checkout-branch-guard.md` 템플릿 판 `SPEC-` 0회). **정정(plan-audit D3, 재측정)**: 템플릿 VCI 8,310B, 로컬 VCI 8,224B — 템플릿 판이 86B **더 크다**(종전 '로컬 판보다 작음'은 오기록). 양 판의 분기(diff 36행)는 근거행 생략이 아니라 **통째 중립화 재작성**이다 — 실측 예: 로컬 "coverage 87%"·"0 0 sync" → 템플릿 "coverage met"·"remote in sync", `go test -cover ./internal/<pkg>/...` 예시 → "the coverage command", `§E`(E1-E7) 구조 → "self-verification deliverables"(로컬 1회/템플릿 0회), SPEC 교차참조 → 일반문 치환. 이 규모의 분기가 관리된다는 전제는 성립하지 않는다 — D4의 바이트 동일 채택은 분기 자체를 원천 봉쇄하므로 이 선례보다 한층 강하게 정당화된다 | `wc -c` 양측 → 8,224 / 8,310 (2026-08-25 재측정); `grep -c 'SPEC-'` → 로컬 VCI 2 / 템플릿 VCI 0 / 템플릿 branch-guard 0; `diff` → 36행 |
| A-7 | `internal/template/catalog.yaml`은 룰 파일을 목록화하지 않는다 → 신규 룰 파일에 catalog 행 불필요 | `grep -c 'rules/moai' internal/template/catalog.yaml` → `0` (총 268행) |
| A-8 | t197 근거 문서 `.moai/reports/t197/procedure-defect.md` 는 본 트리에 없다(t197 미머지 WIP) → 근거는 규칙 파일에 인라인으로 실어야 한다(포인터 금지) | `test -f .moai/reports/t197/procedure-defect.md` → rc=1 (디렉터리 자체 부재) |
| A-9 | zone-registry는 4개 헌법 원천 파일(CLAUDE.md → moai-constitution.md → agent-common-protocol.md → design/constitution.md)의 HARD 조항만 등록 대상으로 열거한다 → development/ 룰 파일의 신규 [ZONE] 조항은 등록 대상 밖 | `.claude/rules/moai/core/zone-registry.md` §ID Allocation Policy (고정 파일 순서 명시) |
| A-10 | rule-authoring.md 의무 (a)신규 always-loaded 파일 = 크기+비용 명시, (b)1,000B 초과 성장, (c)불필요 세션이 지불하는 비용 서술, (d)`paths:` 스코프 우선 — 본 SPEC은 (d)를 채택하므로 (a)는 발화하지 않는다. 다만 카드가 요구하는 예산 영향 실측은 의무화한다(REQ-VC-006) | `.claude/rules/moai/development/rule-authoring.md` §The statement duty |

CMD-1/CMD-3 이 반환한 always-loaded 14파일 (2026-08-25, `32d2221fa`): `core/{agent-common-protocol, askuser-protocol, moai-constitution, moai-mcp-tools, native-idiom-and-register, verification-claim-integrity}.md`, `workflow/{cache-aware-execution, context-window-management, cross-session-messaging, goal-directive, kanban-dispatch, main-checkout-branch-guard, session-handoff, skill-routing}.md`

## §B. 사용자 스토리

운영자(GOOS)로서, 나는 t241의 라인 8개가 피로 얻은 검증 규율 6건이 (1) 다음 라인/다음 사람이 규칙 파일에서 읽고 따를 수 있게 `.claude/rules/` 에 착지되고, (2) 각 규칙이 어느 관측에서 나왔는지 근거와 함께 실려 무시되지 않으며, (3) 그 착지가 always-loaded 예산을 0바이트 소모하는 것을 실측으로 확인받고 싶다. 규칙 파일은 템플릿으로 배포되므로 내부 추적 ID(카드 번호, 날짜, SHA)가 사용자 프로젝트에 새어 나가지 않아야 한다.

## §C. 범위 요약과 범위 외 (Scope Summary & Out of Scope)

**범위 요약**: 신규 path-scoped 룰 파일 1개(`.claude/rules/moai/development/verification-completeness.md`) — 6규칙 + 중립 인라인 근거 + 완결성 단일 축 구조 — 의 작성, 템플릿 소스 미러(바이트 동일) + `make build` 재임베드, always-loaded 예산 영향 실측(전/후, SHA 고정), zone-registry 비적용 결정 기록.

### Out of Scope — 기계적 강제 (mechanical enforcement)

- 어떠한 훅·린트·CI 검사도 본 SPEC에서 구현하지 않는다. 착지물은 doctrine 텍스트(규칙 파일)뿐이다.
- 6규칙의 위반을 자동 탐지하는 도구(예: always-green 체크 탐지기)는 후속 카드/SPEC 소관.

### Out of Scope — 7번째 규칙 (zero-match selector 관측)

- 2026-08-25 레인 9·11의 zero-match selector 관측(-run 정규식 0선택 rc 0, grep 0행=통과, t.Skip, 0룰 sg test)은 규칙으로 채택하지 않는다 — 근거(plan-audit D7 재인용): t261의 스코핑이 '이 여섯을 규칙 파일로 옮기고'로 6건에 한정되고, t241은 이 관측을 별도 규칙이 아닌 '전부 "무엇을 실제로 훑었는지 세지 않으면 통과가 무의미하다"는 한 규칙의 사례'로 규정한다(관측 원천은 plan §A.4 각주 행). 규칙 1(축 §1.1)의 근거 각주로만 인용한다.

### Out of Scope — 기존 SPEC/AC 소급 정비

- 이미 닫힌 SPEC의 acceptance criteria 를 6규칙에 맞게 소급 수정하지 않는다.
- 기존 82개 룰 파일의 재구조화·압축은 건드리지 않는다(SPEC-RULE-DIET-* / SPEC-V3R6-RULES-* 계열 소관).

### Out of Scope — moai update 청소 결함

- `CleanMoaiManagedPaths` 보호목록 부재(CLAUDE.local.md §2.3 결함 ①) 수정은 별도 카드 소관. 본 SPEC은 그 결함을 회피하는 위치 선택(템플릿 미러 = 배포 대상과 동일 파일)으로만 대응한다.

### Out of Scope — 배포 사용자 문서화

- docs-site / README 4-locale 문서화는 sync-phase/별도 카드 판단이다. 본 SPEC의 REQ는 룰 파일+템플릿 미러+측정에 한정한다.

## §D. 요구사항 (GEARS)

> <subject> 일반화 형을 사용한다(루트 규칙 파일, SPEC 디렉터리, 템플릿 원천, 본 SPEC의 AC). 각 REQ 직후의 한국어 줄은 근거 요약이지 요구사항의 일부가 아니다.

- **REQ-VC-001** (Ubiquitous) — The rule file `.claude/rules/moai/development/verification-completeness.md` shall codify the six harness rules of card t241 as `[ZONE:Evolvable] [HARD]` clauses organized under a single completion axis — *a verification artifact (check, gate, acceptance criterion, rule, or assertion) is incomplete until its failure has been observed on a known input* — where the axis core carries the observed-failure completion rule, the three-part check specification (red timing / red input / red reachability), and the two-cell adoption discipline (a RED-now observation that is red **for the right stated reason**, paired with a green path that no unrelated fix is needed to reach), and the remaining rules (cross-layer revision sweep, evidence pinning with the moving-ref discriminator) form subordinate sections with the corollaries attached beneath them.
  - 근거: 리드의 구조 제안(규칙 1·3·4는 "알려진 입력에서 실패를 관측했을 때만 검증물은 완결"이라는 단일 축)을 채택. 6평행 항목 구조는 카드가 통합하려는 규율을 재파편화한다.
- **REQ-VC-002** (Event-driven) — **When** the rule file states a rule, the rule entry shall carry an inline evidence summary describing the observed failure form that produced it, phrased as distribution-neutral generic prose carrying no SPEC identifier, card identifier, internal date, or commit SHA.
  - 근거: 카드 경고(근거 없는 규칙은 무시된다) + §25.1 중립성(금지 클래스 = 치환 사전의 generic prose 로 표현). A-8: 존재하지 않는 t197 문서를 가리키는 포인터 금지 → 요약 자체가 인라인.
- **REQ-VC-003** (Ubiquitous) — The SPEC directory shall carry a provenance matrix (plan.md §A.4) mapping each of the six rules to its originating card, iteration, observation date, and incident summary, preserving the full internal traceability that the distribution-neutral rule file cannot carry.
  - 근거: 중립 규칙 파일이 실을 수 없는 카드 단위 근거(어느 카드·어느 반복·무엇을 관측)의 원천 보존처. `.moai/specs/` 는 템플릿 동기화 보호 대상이라 안정적.
- **REQ-VC-004** (Capability gate) — **Where** a session touches a verification-artifact authoring context, the rule file shall be loadable through a top-level `paths:` frontmatter scope covering SPEC artifacts (`**/.moai/specs/**`), rule files (`**/.claude/rules/**` — which subsumes the template tree's nested `.claude/rules/`, listed explicitly as `internal/template/templates/.claude/rules/**` for readability), hook and gate code (`internal/hook/**`), check scripts (`scripts/**`), and rule/ruleset roots (`**/.moai/astgrep-rules/**`, `**/.moai/hooks/**`) — and the rule file shall NOT join the always-loaded surface.
  - 근거: rule-authoring.md 의무 (d) SCOPE FIRST. 스코프 문법과 값은 `spec-frontmatter-schema.md` 의 `paths: "**/.moai/specs/**,internal/spec/**"` 선례 형식을 따른다.
- **REQ-VC-005** (Ubiquitous) — The template source counterpart `internal/template/templates/.claude/rules/moai/development/verification-completeness.md` shall exist byte-identical to the local rule file (`cmp` exit 0), shall be re-embedded via `make build`, and the template copy shall contain zero occurrences of every forbidden class of `.moai/docs/template-internal-isolation-doctrine.md` §25.1 (internal SPEC/card/REQ/AC identifiers, internal dates, commit SHAs, audit citations, internal archive/memory paths).
  - 근거: Template-First [HARD] + A-6 선례. 바이트 동일 채택 이유는 plan.md D4.
- **REQ-VC-006** (Event-driven) — **When** the rule file lands, the always-loaded rule surface shall remain identical to the baseline measured at tree SHA `32d2221fa` (14 files, 179,081 bytes, CMD-3 enumeration) — the new file shall not appear in the enumeration, and any count or byte delta shall be attributed to named foreign files rather than ignored or explained away without re-measurement.
  - 근거: 카드의 산출 축 (b). 규칙 6의 자기적용 — 기반선을 SHA에 고정하고, run-phase HEAD에서 재측정(재인용 아님)한다.
- **REQ-VC-007** (Event-driven) — **When** an acceptance criterion of this SPEC asserts a delta, an invariant, or a new check, the criterion shall carry a RED-now observation pinned to the tree SHA where it was observed (stating why it is red — the right reason), and a green path naming the milestone that flips it and the passing output it becomes — the SPEC's own acceptance criteria shall comply with the rules this SPEC lands, wherever mechanically applicable.
  - 근거: 카드 제약 6(자기적용) — 제약이자 시위. 규칙 1·4·6을 본 SPEC의 AC에 기계적으로 적용한다.
- **REQ-VC-008** (Capability gate) — **Where** the rule file introduces `[ZONE:Evolvable] [HARD]` clauses outside the four constitutional registry source files, this SPEC shall record the zone-registry non-applicability decision and its policy basis in plan.md, keeping any future registry promotion explicitly deferred rather than silent.
  - 근거: A-9 — registry의 고정 4-파일 할당 범위 밖. 비적용을 근거와 함께 기록하여 다음 사람이 "빠뜨렸다"로 읽지 않게 한다.

## §E. 제약 (Hard Constraints)

1. **Template-First** — 신규 `.claude/` 파일은 템플릿 원천(`internal/template/templates/.claude/rules/moai/development/`)에 추가 후 `make build` 로 재임베드한다(CLAUDE.local.md §2 [HARD]).
2. **템플릿 중립성** — 템플릿 판(= 로컬 판과 동일 파일)에는 §25.1 금지 클래스 0건. 특히 카드 ID(t197/t217/t228/t230/t241), 내부 날짜(2026-08-2x), 커밋 SHA(f7eec06c7), SPEC ID가 템플릿 판에 등장해서는 안 된다. 이 토큰들은 plan.md §A.4(로컬 전용)에만 실린다.
3. **언어** — 규칙 파일 본문은 영어(지시문 정책, 코딩 스탠다드 Language Policy). 한국어는 SPEC 산출물에만.
4. **워크트리 규율** — run-phase 는 본 워크트리(`WT-harness-rules`)에서 상대 경로로 작업. 브랜치 생성·push 금지는 카드 레인 운영 정책을 따른다.
5. **always-loaded 예산** — rule-authoring.md 의무 (d) 위배 없이 신규 always-loaded 파일을 만들지 않는다. 의무 (a)는 발화하지 않으나(스코프 채택) 카드 요구 실측(REQ-VC-006)으로 대체 증명한다.
6. **무변경(PRESERVE)** — 기존 82개 룰 파일, zone-registry.md, catalog.yaml, 기존 SPEC 디렉터리 일체를 수정하지 않는다(신규 추가만).
7. **계측기 형식** — run-phase 검증 명령은 본 워크트리의 worktree guard 가 수용하는 형태여야 한다(단일 awk 파이프라인은 실측 확인됨; 복합 루프는 거부 관측됨 — plan.md §B).

## §F. 인접 경계

- `core/verification-claim-integrity.md` — 쌍둥이. 본 SPEC은 "검증/완료 **주장**은 관측되어야 한다"(claim)를 "검증 **도구**는 알려진 입력에서 실패가 관측되어야 완결이다"(check)로 확장한다. 주장 무결성이 거짓 PASS를 막는다면, 검증 완결성은 애초에 판정을 내릴 수 없는 도구를 막는다.
- `development/spec-frontmatter-schema.md` — `paths:` 스코프 문법 선례 및 SPEC 산출물 스키마 SSOT.
- `development/rule-authoring.md` — always-loaded 비용 규율 (a)~(d). 본 SPEC의 예산 축을 이 의무로 운영화한다.
- `workflow/spec-workflow.md` §SPEC Complexity Tier — Tier M 산출 세트(3 files + progress.md).
- SPEC-EVIDENCE-CLAIM-INVARIANT-001 — claim-무결성 규칙의 원천 SPEC. 본 SPEC은 그 축의 check-측 확장이지 대체가 아니다.

## §G. 교차 참조

- 카드 t261 (본 카드), 카드 t241 (6규칙·근거 원천, 2026-08-24~25 라인 8개 기여)
- `.moai/docs/template-internal-isolation-doctrine.md` §25.1~§25.3 (중립성 클래스·치환·자기점검)
- `.claude/rules/moai/core/zone-registry.md` §ID Allocation Policy
- CLAUDE.local.md §2 (Template-First), §2.3 (managed-roots 와이프 — 위치 선택 동기)
- plan.md §A.4 (6규칙 × 카드/반복/관측 근거 행렬 — 내부 추적 SSOT), acceptance.md §D (RED-now 관측 + 녹색-경로 쌍을 갖춘 AC 매트릭스)
