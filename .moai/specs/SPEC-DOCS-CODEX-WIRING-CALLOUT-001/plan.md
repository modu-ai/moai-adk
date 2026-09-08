---
id: SPEC-DOCS-CODEX-WIRING-CALLOUT-001
title: "Plan — docs-site moai doctor 페이지 Codex Wiring 콜아웃 4로케일 반영"
version: "0.1.0"
created: 2026-09-08
updated: 2026-09-08
author: manager-spec
priority: P2
phase: "v3.1.4"
module: "docs-site/content"
tags: "docs, docs-site, i18n, codex, doctor"
tier: M
---

# Plan: SPEC-DOCS-CODEX-WIRING-CALLOUT-001

## §A Context

- Worktree: `.claude/worktrees/t535` · branch `WT-docs-codex-callout` · baseline `a849d99d2` (origin/develop tip)
- SPEC artifacts: `.moai/specs/SPEC-DOCS-CODEX-WIRING-CALLOUT-001/{spec,plan,acceptance,progress}.md`
- Write surface (전부): `docs-site/content/ko/cli-reference/doctor.md`, `docs-site/content/en/cli-reference/doctor.md`, `docs-site/content/ja/cli-reference/doctor.md`, `docs-site/content/zh/cli-reference/doctor.md` — 로케일당 ~20–40행 추가, H2 7→8
- 읽기 대상 (PRESERVE — 수정 금지): `internal/cli/doctor_codex.go` (콘텐츠 근거), `internal/cli/clean.go` (유령 수거 동사 근거), `docs-site/content/{ko,en,ja,zh}/advanced/codex-dual-harness.md` (크로스링크 대상, 편집 없음), `.moai/docs/docs-site-i18n-rules.md` (§17 규칙)

### A.1 결정 우선순위 (변경 가능성 높은 순 — 검토 집중 지점)

1. **절 배치** — Hook Delivery 뒤·종료 코드 앞 (REQ-DWC-003). ko 66행 경계, 4로케일 동일 상대 위치.
2. **절 제목·배지 형태** — ko `## Codex Wiring 진단 {{< new-badge v3.1.4 >}}` / en `## Codex Wiring check {{< new-badge v3.1.4 >}}` / ja `## Codex Wiring 診断 {{< new-badge v3.1.4 >}}` / zh `## Codex Wiring 诊断 {{< new-badge v3.1.4 >}}`. **배지를 v3.1.3으로 "수정" 금지** (spec §1.5). zh 헤딩이 `诊断`(診斷의 간체) 인 것은 ko 정본 용어 '진단'의 직접 한자 대응을 따른 것이다 — zh 페이지 본문에서 '诊断' 은 9회 관측된 용어이고, 기존 H2 일부의 `检查` 는 '점검' 계열이라 다른 어휘다 (plan-audit iter1 D2 — 근거 명시).
3. **fatal 케이스 서술 수위** — 유일 fatal인 `enabled` 형태 불량을 절의 핵심 경고로, 나머지는 조언형 목록으로.
4. **크로스링크 형태** — locale 접두 절대경로 `/ko|en|ja|zh/advanced/codex-dual-harness` (기존 링크 관례: ko doctor.md 53행 `/ko/advanced/home-hygiene`).
5. (기계적) 표 구성·문장 다듬기 — 마지막에.

## §B Known Issues (도메인 필터링 — B5·B6·B8·B10·B11만 해당)

- **ugrep 래퍼 함정**: 이 리포 셸의 `grep` 은 ugrep 래퍼다 — 부재 판정엔 `/usr/bin/grep` (REQ-DWC-014). AC-DWC-001의 부재·착지 판정이 이 규약을 지키지 않으면 0히트가 관측 왜곡일 수 있다.
- **B4 계열(프론트매터)**: doctor.md 기존 frontmatter(title/weight/draft)는 건드리지 않는다 — 본문에 절만 추가.
- **B8 계열(작업 트리 위생)**: `.moai/state/`, `.moai/harness/` 등 런타임 파일 불건드림. 커밋은 pathspec 명시.
- **B11**: 서브에이전트는 사용자에게 묻지 않는다 — blocker 보고로 반환.

## §C Pre-flight (run 시작 시 실행)

```bash
git branch --show-current && git rev-parse --short HEAD   # WT-docs-codex-callout 계열 확인
/usr/bin/grep -rn 'Codex Wiring' docs-site/               # RED-now 재확인 (REQ-DWC-012) — 0히트 기대
/usr/bin/grep -c '^## ' docs-site/content/ko/cli-reference/doctor.md   # 7 기대
/usr/bin/grep -c '^## ' docs-site/content/en/cli-reference/doctor.md   # 7 기대
/usr/bin/grep -c '^## ' docs-site/content/ja/cli-reference/doctor.md   # 7 기대
/usr/bin/grep -c '^## ' docs-site/content/zh/cli-reference/doctor.md   # 7 기대
# (워크트리 가드가 루프·복합형을 거부하므로 평문 단일 명령으로 로케일당 하나씩 실행한다)
sed -n '55,70p' docs-site/content/ko/cli-reference/doctor.md  # 삽입 경계 육안 확인
```

baseline이 기대와 다르면(이미 Codex Wiring 절 존재, H2 ≠ 7) 진행하지 말고 blocker 보고.

## §D Constraints

- 쓰기 표면: 위 4파일만. 그 밖의 모든 것 PRESERVE (REQ-DWC-010).
- 금지: Go 소스·템플릿 편집, layouts/shortcodes 편집, 메뉴/사이드바 편집, README/CHANGELOG 편집, 다른 docs 페이지 편집.
- 금지: 배지 `v3.1.3` 사용 (REQ-DWC-008).
- 금지: doctor 의 지시문으로 `moai clean --codex-skills` 기술 (REQ-DWC-004 — 코드에 그 지시문이 없다).
- 본문 장식 이모지 금지; URL은 `adk.mo.ai.kr` 계열만; Mermaid 도입 시 TD-only (이 절에 Mermaid는 불요 — 도입 자체를 권장하지 않음).
- 커밋/푸시 금지 (오케스트레이터 소관).

## §E Self-Verification (완료 보고 필수 항목)

각 항목 VCI §3 5단 형식(주장/증거/귀속/미검증/잔여위험)으로 보고:

- **E1** AC 이진 매트릭스 — acceptance.md §D의 14 AC 전수, 명령+원문 출력.
- **E2** 4로케일 H2 패리티 — `grep -c '^## '` ×4 = 8.
- **E3** 부재 역전 — `/usr/bin/grep -rn 'Codex Wiring' docs-site/` exit 0, 4로케일 doctor.md 최소 4히트.
- **E4** 배지 귀속 — `new-badge v3.1.4` 4히트, 이 절 안의 `v3.1.3` 0히트.
- **E5** hugo 빌드 무경고 (`cd docs-site && hugo --minify`).
- **E6** 범위 판정 — `git diff --stat`이 4파일만.

## §F Milestones

- **M1 — ko 정본 절 작성** (Priority High): `docs-site/content/ko/cli-reference/doctor.md`에 절 추가. 내용 골격: (a) 검사 소개 — 조언형·fail-open·읽기 전용, 읽기만 하고 고치지 않음; (b) 검사 항목 표 (spec §1.4 표의 사용자 친화 재구성); (c) 네 가지 수정 지시문 코드 문구 그대로; (d) fatal 경고 — `enabled` bare TOML 불리언, codex 0.153.4 실측 한정; (e) un-nagging 불변(코덱스 없는 claude-only 프로젝트는 조용한 스킵); (f) 유령 등록 수거는 `moai clean --codex-skills` 별도 동사로 한 줄; (g) `[Codex 듀얼 하니스](/ko/advanced/codex-dual-harness)` 크로스링크. 강조 간격 규칙 준수.
- **M2 — en 파생** (Priority High): M1 ko를 en으로 파생. 진단·명령·명칭 영문 관례는 기존 en doctor.md의 Home Disk Usage/Hook Delivery 절을 따른다 (`## Codex Wiring check`).
- **M3 — ja·zh 병렬 파생** (Priority Medium): en 확인 후 ja/zh 동시 파생. 사실·수치·코드블록·지시문 verbatim 보존.
- **M4 — 검증 스윕** (Priority High): §E 전수 실행 — 패리티·배지·링크·이모지·URL·hugo 빌드·범위 판정. 실패 시 M1로 롤백 재작성.

## §G Anti-Patterns

- 배지를 "정렬"한다며 v3.1.3으로 되돌리는 것 — 허위 귀속 (spec §1.5가 차단 근거).
- doctor 가 `moai clean --codex-skills`을 가리킨다고 쓰는 것 — 코드에 없는 지시문.
- `enabled` fatal을 "일반 경고" 수준으로 표현하는 것 — 이 검사의 유일한 fatal이며 exit 1로 이어지는 유일 경로다.
- ja/zh를 ko 작성 직후 병렬 파생하는 것 — en 확인 후 파생이 체인(ko→en→ja/zh)이다.
- 이 절 작성에 동반해 doctor.md의 다른 절을 "다듬는" 것 — 범위 위반.

## §H Cross-References

- spec.md §1 (측정 근거), §2 (REQ-DWC-001..015), acceptance.md §D (AC-DWC-001..014)
- `.moai/docs/docs-site-i18n-rules.md` §17 (URL·이모지·Mermaid·강조 간격·4로케일 동시)
- `.claude/skills/hns-oss-docs-verify/SKILL.md` (M4 검증 축)
- SPEC-DOCS-V313-CATCHUP-001 (구조 모델 + §1.7 갭 생존 원인)
