---
id: SPEC-REF-SEO-ABSORB-001
title: "moai-ref-seo 레퍼런스 스킬 클린룸 흡수 + 재사용 가능한 클린룸 프로토콜 확립"
version: "0.3.1"
status: completed
created: 2026-08-01
updated: 2026-08-02
author: manager-spec
priority: Medium
phase: "v3.0.2"
module: template-skills
lifecycle: spec-anchored
tags: "skill, seo, reference, clean-room, template, epic-absorb"
tier: M
---

## HISTORY

- v0.3.1 (2026-08-01) — plan-audit iteration-3 **PASS(0.87)** 이후 잔여 SHOULD-FIX 처리. N10 정정: `research.md` §B.4의 `moai-ref-ui-polish` H2 수치 `11개` → `명명 10 + 불변 3 = 13개`(실측 재확인). N4가 이 수치를 물려받은 것이었으므로 발원지를 닫았다 — research.md는 형제 SPEC 2건의 재사용 대상이라 오기가 미착수 작업으로 전파된다. N8(REQ-SEO-020 경로 형태 무판정) / N9(REQ-SEO-004 frontmatter 필드 무판정)는 산출물이 존재해야 수정의 동작을 관측할 수 있으므로 사용자 결정에 따라 run-phase M6로 이연.
- v0.3.0 (2026-08-01) — plan-audit iteration-2 FAIL(0.79, 통과선 0.80에 0.01 미달) 대응 개정. iteration-1 MUST-FIX 6건은 CLOSED 확정되어 재검증 대상이 아니며, 이번 개정은 **v0.2.0에서 새로 쓴 콘텐츠의 1차 감사 결함**을 닫는다. N1 AC-SEO-015 표 헤더 명령 실행 불가(`grep -- '---' -B1` 옵션 파싱), N2 §D.1 추적성 표의 REQ↔AC 매핑 오기재 및 그에 기댄 §B 번호 관례 근거 문장, N3 AC-SEO-013 이중 임계값. SHOULD-FIX N4-N7 + SF-10 동반 정정.
- v0.2.0 (2026-08-01) — plan-audit iteration-1 FAIL(0.72) 대응 개정. MUST-FIX 6건 정정(AC-SEO-025 가드 오지정 / AC-SEO-022 표면 집합 역전 / 접근성 결정-요구사항 충돌 / 요구사항 개수 오기재 / 미해소 마커 5건 / REQ 번호 결번), plan.md §B의 열린 결정 5건을 전부 결정 기록으로 치환, 클린룸 프로토콜에 구조 발산 판정과 원문 해시 고정을 추가.
- v0.1.0 (2026-08-01) — plan-phase 최초 저작. 3-렌즈 조사(research.md) 기반. 미해소 결정 5건은 plan.md §B에 clarification 마커로 보류(v0.2.0에서 전부 결정 기록으로 치환).

---

## §A. 배경과 목적

MoAI-ADK 템플릿에는 SEO 도메인 레퍼런스가 없다. 웹 산출물을 만드는 사용자 프로젝트에서 canonical URL·구조화 데이터·robots/sitemap·엔티티 일관성 같은 기초 규율이 에이전트 지식 밖에 있다.

제3자 레퍼런스 문서 1건이 이 공백을 잘 덮는 개념 집합을 담고 있으나, 해당 배포에는 라이선스·NOTICE·COPYING이 존재하지 않는다(research.md §D.2 실측). 산출물은 템플릿 트리를 통해 전체 사용자에게 배포되므로 원문 산문을 복사하는 것은 허용되지 않는다.

따라서 본 SPEC은 두 가지를 동시에 산출한다.

1. **`moai-ref-seo` 신규 레퍼런스 스킬** — 원문에서 개념과 구조만 취하고 모든 문장을 MoAI 자체 어조로 새로 작성한 결과물
2. **재사용 가능한 클린룸 재작성 프로토콜** — 기계적으로 검증 가능한 형태. 본 SPEC은 Epic 1/3이며, 형제 SPEC 2건이 동일 프로토콜을 따른다

---

## §B. 요구사항 (GEARS)

**번호 부여 관례 (의도적 블록 결번)**: 요구사항 번호는 세 블록으로 나뉘며 **블록 경계에서 번호를 건너뛴다**. B.1 콘텐츠 = 001-009, B.2 프로토콜 = 010-015, B.3 배포 = 020-027. 따라서 016-019는 결번이며 누락된 요구사항이 아니다. 총 23건이다.

연속 재부여를 하지 않는 근거는 다음 둘이며, **"AC 번호와 REQ 번호가 1:1 대응한다"는 아니다** — 그 주장은 §B.2 프로토콜 블록에서 실제로 성립하지 않는다(아래).

1. **블록 경계가 의미를 나른다.** 세 자리 번호의 십의 자리가 콘텐츠 / 프로토콜 / 배포를 구분하므로, 번호만 보고 요구사항의 성격을 안다. 연속 재부여는 이 정보를 버린다.
2. **번호가 이미 광범위하게 참조되고 있다.** `acceptance.md`의 판정 항목·§D.1 추적성 표·`plan.md` §B 결정 기록·§E DoD 체크박스가 REQ 번호를 직접 인용한다. 재부여는 이 참조를 전부 갱신해야 하며, 그 과정에서 새 불일치가 생길 위험이 관례 선언 1문단의 비용을 넘는다.

**AC 번호와 REQ 번호의 실제 관계**: 수용 기준 번호는 요구사항 번호와 **같은 수 공간에 두되, 1:1 대응은 아니다.** §B.1 콘텐츠와 §B.3 배포 블록에서는 대체로 동일 번호가 대응하지만, §B.2 프로토콜 블록에서는 한 칸 어긋난다(`AC-SEO-010`→`REQ-SEO-011`, `AC-SEO-011`→`REQ-SEO-012`, `AC-SEO-012`→`REQ-SEO-013`). 이는 오류가 아니라 **대응 관계 자체가 다대다이기 때문이다** — `REQ-SEO-011`(라이선스 확인 + 해시 고정)은 두 개의 AC가 절반씩 판정하고, `REQ-SEO-012`(중첩 검사)는 `AC-SEO-011`과 `AC-SEO-013`이 함께 판정한다. 어떤 번호 부여로도 1:1을 만들 수 없다.

따라서 대응 관계는 번호 규칙이 아니라 **`acceptance.md` §D.1 추적성 표가 명시적으로 관리한다**. 형제 SPEC 2건이 이 형식을 복제할 때는 이 선언 문단과 §D.1 표를 **함께** 복사하고, 자신의 실제 매핑을 §D.1에 다시 기록한다 — 번호 관례만 복사하고 §D.1을 비워두면 이 SPEC이 v0.2.0에서 저지른 것과 같은 거짓 1:1 주장이 재생산된다.

### B.1 스킬 콘텐츠

- **REQ-SEO-001** (Ubiquitous) — The `moai-ref-seo` skill shall exist as a single `SKILL.md` file with no subdirectory, at both `.claude/skills/moai-ref-seo/SKILL.md` and `internal/template/templates/.claude/skills/moai-ref-seo/SKILL.md`.
- **REQ-SEO-002** (Ubiquitous) — The skill body shall carry a single H1, 4-10 domain H2 sections whose content is predominantly markdown tables, and shall total between 150 and 220 lines.
- **REQ-SEO-003** (Ubiquitous) — The skill body shall terminate with the invariant triad `## Common Rationalizations`, `## Red Flags`, `## Verification` in that order, each wrapped in evolvable-zone comments carrying the ids `rationalizations`, `red-flags`, `verification`.
- **REQ-SEO-004** (Ubiquitous) — The skill frontmatter shall supply all fields required of a reference skill (`name`, `description`, `when_to_use`, `user-invocable: false`, the quoted `metadata` map, and the `progressive_disclosure` block), and the combined `description` + `when_to_use` length shall stay under 1,536 characters.
- **REQ-SEO-005** (Ubiquitous) — The `description` shall follow the three-movement form: a scope noun-phrase ending in `reference`, the invariant amplification sentence, and a terminal `NOT for:` negative-scope clause naming sibling skills inline. The `NOT for:` clause shall additionally name the three operability accessibility topics that this skill does not own — keyboard operability, visible focus, and form-control labeling — and shall name generative-engine optimization as a deliberate exclusion. The `NOT for:` clause is the sole placement for both exclusions; no trailing body line is added, because the body terminates with the invariant triad per REQ-SEO-003.
- **REQ-SEO-006** (Ubiquitous) — The skill body shall cover the durable concept set: canonical-URL discipline, per-page unique title and meta description, robots.txt and sitemap.xml as host-derived artifacts, JSON-LD structured data (type selection, required fields, absolute URLs, graph cross-referencing), the visible-content-mirroring rule, entity identity hygiene, a single response-header chokepoint covering error and redirect paths, and the four SEO-causal document-semantics items: single-H1 with unskipped heading levels, image `alt` text, descriptive anchor text, and live in-page fragment targets.
- **REQ-SEO-007** (Unwanted) — The skill body shall not reproduce the source document's uncited numeric claims (the engagement-loss figure, the TTFB figure, the preconnect-saving figure), and shall not state any comparable figure without a verifiable citation.
- **REQ-SEO-008** (Unwanted) — The skill body shall not carry volatile stated values as normative thresholds (character budgets for title and description, sitemap `priority` and `changefreq`, content-to-code word thresholds); **where** such a topic is covered, the skill shall express it as a decision rule to be measured rather than a fixed number.
- **REQ-SEO-009** (Unwanted) — The skill body shall not include the platform-bound marketplace-sidecar flow or any paid-asset credit-cost interaction rule.

### B.2 클린룸 프로토콜 (형제 SPEC 재사용 대상)

- **REQ-SEO-010** (Ubiquitous) — The clean-room protocol shall be recorded in this SPEC in a form that a sibling absorption SPEC can adopt without re-deriving it.
- **REQ-SEO-011** (Ubiquitous) — The protocol shall define provenance verification as a per-source check that the source document's own distribution directory contains zero license, NOTICE, or COPYING files, rather than a repository-wide check, and shall pin the source document by recording its `sha256` digest alongside its path, so that a changed third-party distribution invalidates the recorded baselines loudly rather than silently.
- **REQ-SEO-012** (Ubiquitous) — The protocol shall define a mechanical textual-overlap check between the delivered artifact and the source document, stating the exact command, the normalization applied, the numeric threshold, and the expected output.
- **REQ-SEO-013** (Ubiquitous) — The overlap check shall be accompanied by a self-trip demonstration in which the source document is compared against itself and the check is observed to FAIL, proving the check is not vacuous. The self-trip judgment shall assert the pinned source digest before comparing its numeric result against the recorded baseline.
- **REQ-SEO-014** (Event-driven) — **When** the overlap check reports any shared n-gram above the threshold, the implementer shall re-author the offending passage and re-run the check, rather than adjusting the threshold.
- **REQ-SEO-015** (Ubiquitous) — The protocol shall state that structure, section ordering, and concept inventory are facts that may be studied and reused, while sentences, phrasing, and authored numeric thresholds shall be newly written; and **because the n-gram and longest-common-substring checks are both lexical-contiguity measures that a synonym-substituted copy of the source's section ordering and table shape would pass**, the protocol shall additionally require a recorded structural-divergence verdict — a human comparison of the delivered artifact's section sequence and table column composition against the source's, with the verdict and its reasoning written into the SPEC's progress record.

### B.3 배포와 등록

- **REQ-SEO-020** (Ubiquitous) — The skill shall be registered in `internal/template/catalog.yaml` with the five-key entry shape, its `path` carrying the `templates/` prefix and a trailing slash, its `hash` generated by `make build` rather than hand-written, and its `tier` set to `optional-pack:frontend` (per the plan.md §B.1 decision record).
- **REQ-SEO-021** (Ubiquitous) — The three Go count constants (`expectedSkillCount` in `catalog_tier_audit_test.go`, `expectedTotal` in `catalog_loader_test.go`, `wantTotal` in `embed_catalog_test.go`) shall each be incremented by one, and the increment shall land in the same change as the markdown addition.
- **REQ-SEO-022** (Ubiquitous) — The skill shall be registered on the same surface set that `moai-ref-secops` actually occupies, measured rather than assumed: `delegation.yaml` in both the local and template copies (under the `frontend` domain key, per the plan.md §B.1 decision record), the two workflow bodies `.claude/skills/moai/workflows/review.md` and `.claude/skills/moai/workflows/sync/quality-gates-quality.md` together with their two template mirrors, the docs-site skill-guide table, and the skills-manifest spot-check list. The skill-routing rule is **not** part of this set — `moai-ref-secops` does not appear there, and its three skill examples are fixed illustrations rather than a registration surface.
- **REQ-SEO-023** (Ubiquitous) — The docs-site skill-guide reference table shall gain a `moai-ref-seo` row in all four locales.
- **REQ-SEO-024** (Unwanted) — The delivered template-tree files shall not contain any forbidden internal-content token: SPEC identifiers, requirement or acceptance-criteria token shapes, audit citations, internal dates outside the frontmatter `updated:` field, commit hashes, absolute user-home paths, or maintainer-local instruction-file references.
- **REQ-SEO-025** (Unwanted) — The skill body shall not assert primacy, default status, or exclusive support for any one programming language, and shall not add a `GOOS=` token to the template tree.
- **REQ-SEO-026** (Where) — **Where** the skill body names a web framework, it shall name several at equal depth or none, and shall frame guidance at the protocol and output layer rather than as a framework walkthrough.
- **REQ-SEO-027** (Event-driven) — **When** the template-tree guard suite reports a leak or neutrality finding attributable to the new skill, the implementer shall correct the skill body, and shall not add a guard exemption entry.

---

## §C. 제외 범위

이 섹션은 본 SPEC이 **만들지 않는 것**을 고정한다.

### Out of Scope — 형제 Epic SPEC

- design-taste-frontend 문서 흡수 및 `moai-ref-ui-polish` 상위 계층 구성
- security 문서 흡수
- review-rubric 문서 흡수
- 위 2건은 본 SPEC이 확립한 클린룸 프로토콜을 재사용하되, 각자의 SPEC에서 자체 조사와 요구사항을 갖는다

### Out of Scope — 원문 중 이관하지 않는 내용

- 빌드타임 메타데이터 사이드카 JSON을 마켓플레이스 리스팅 카드에 동기화하는 흐름
- 유료 커버 비디오 생성 및 크레딧 비용 상호작용 규칙
- 출처 없는 성능·참여도 수치 3건
- generative-engine-optimization 원칙 전체 — plan.md §B.4 결정에 따라 전면 제외하며, 제외 사실은 `NOT for:` 절에서만 밝힌다
- 원문의 10항목 감사에 붙은 항목별 수치 임계값과 차단 권한 — plan.md §B.2 결정에 따라 점검 항목과 판정 어휘만 흡수한다

### Out of Scope — 접근성 중 형제 SPEC에 넘기는 항목

- 키보드 조작 가능성
- 가시적 포커스
- 폼 컨트롤 라벨링
- 위 3종은 SEO가 아니라 접근성이 1차 이해관계자이므로 `NOT for:` 절로 형제 표면에 명시 위임한다. plan.md §B.3 결정 기록 참조 — **현재 이 위임을 기록할 형제 SPEC은 존재하지 않으며, 그 사실 자체가 §B.3에 명시되어 있다**

### Out of Scope — 인접하지만 본 SPEC이 손대지 않는 표면

- `/moai gate` 서브커맨드의 동작 변경 또는 신규 게이트 단계 추가
- 실제 SEO 점검을 수행하는 Go 코드·CLI 서브커맨드·훅 신설
- `moai-ref-ui-polish`의 본문 수정
- `moai-ref-ui-polish`가 남긴 불완전 등록 표면(`delegation.yaml`, 워크플로 본문, skills-manifest, en·ja·zh skill-guide 누락)의 소급 보정
- docs-site skill-guide 산문의 선행 카운트 드리프트(로케일 간 27/27/27/30 불일치) 전면 재정합

### Out of Scope — 원문 자체

- 원문 문서의 수정, 재배포, 저장소 내 보관
- 원문 산문의 인용, 발췌, 부분 복사

---

## §D. 성공 기준

1. `moai-ref-seo` 스킬이 템플릿과 로컬 양쪽에 존재하고 ref 스킬 저작 표준(frontmatter 3악장, 불변 3종, 표 중심 본문, 150-220줄)을 만족한다.
2. 클린룸 중첩 검사가 임계값 이내로 통과하고, 동일 명령의 self-trip 실행이 고정된 원문 해시 위에서 FAIL을 관측하며, 구조 발산 판정이 기록된다.
3. 템플릿 가드 스위트(leak / neutrality / catalog / manifest)가 신규 스킬 추가 후에도 통과한다 — 언어 중립성 판정은 `TestLanguageNeutrality`가 실제로 실행되었음을 출력에서 확인한다.
4. Go 카운트 상수 3개가 증가하고 동일 변경에 마크다운이 포함되어 CI 전체 스위트가 실제로 실행된다.
5. 등록 표면이 `moai-ref-secops`의 **실측** 표면 집합과 동일 수준으로 채워진다(워크플로 본문 2종 + 템플릿 미러 2종 포함, skill-routing 규칙 제외).

---

## §E. 참조

- `research.md` — 3-렌즈 조사, 교차 렌즈 모순 해소, 실측 명령
- `plan.md` — 마일스톤, §B 결정 기록 5건(전부 확정)
- `acceptance.md` — 판정 명령과 기대 출력

### 원문 고정 (클린룸 프로토콜 입력)

| 항목 | 값 |
|---|---|
| 경로 | `~/.agents/skills/higgsfield-websites/references/seo.md` |
| 분량 | 861 lines |
| `sha256` | `c088f089f365fae621dac90db77b56abdc47463becb2dbd182bd0cd04de98ee7` |

이 digest는 `acceptance.md` §B의 self-trip 기대값 `4600`이 유효한 조건이다. digest가 달라지면 원문이 갱신된 것이므로 baseline을 재측정해야 하며, `4600` 불일치를 산출물의 결함으로 읽어서는 안 된다. 형제 SPEC 2건은 각자의 원문에 대해 같은 형태로 digest를 고정한다.
