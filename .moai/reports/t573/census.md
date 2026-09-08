# Census — SPEC-AC-LOCALE-TOKEN-001 (card t573)

> 측정 트리: 워크트리 `WT-ascii-token-criterion` 작업 트리, HEAD `0e1f248cd` (base `3ac58b5a1` = origin/develop) · 측정일: 2026-09-08 · M1 재측정
> 모든 grep 은 `/usr/bin/grep` (셸 grep 은 ugrep 래퍼 — t538 REQ-012 계승). 코퍼스에서 `SPEC-AC-LOCALE-TOKEN-001` 디렉터리는 제외했다 (자기 아티팩트 착지 시 계수 표류 방지 — t538 AC-011 동형).
> 아래 모든 수치는 M1 시작 시점 재측정값이다. plan-시 스냅샷(2276/55/98/62/161/66/95)과 동일하게 재현됐다 — 스냅샷을 근거로 쓰지 않고, 같은 명령의 재실행 출력을 근거로 쓴다.

## §1 REQ-002 필드 스키마

각 C-판정 행은 다음 7개 필드를 갖는다:

| 필드 | 의미 |
|---|---|
| 파일 | 기준이 진술된 코퍼스 파일 |
| AC id | 해당 파일 안의 기준 식별자 |
| 토큰 | 계수 대상 토큰 |
| 대상 문서+로케일 | 계수 대상 문서와 그 로케일 |
| 계수 의미 | 행(matching lines, `grep -c`) vs 출현(occurrences, `grep -o…\|wc -l`) |
| 현재 히트 vs 기준 | 본 트리 실측값과 기준이 요구하는 바 |
| 판정 | `distorts` / `benign-code-identifier` / `n-a` |

**판별식(REQ-002·plan §F M1.4)**: (1) 계수 대상 토큰이 ASCII/Latin 인가 (2) 대상 문서가 비(非)Latin 로케일(ko/ja/zh)인가 (3) 기준이 토큰의 **존재/출현을 요구**(≥1, ≥N)하면서 토큰이 산문/라벨 자리에 놓이는가 → `distorts`. 토큰이 코드 식별자(CLI 플래그 값·파일명·SPEC-ID·버전 리터럴)이고 계수식이 코드 위치로 한정되거나, 토큰이 도구가 출력하는 고유 라벨(예: `moai doctor` 진단명 "Home Disk Usage" — `internal/cli/doctor_disk_test.go:68` 실측)의 인용이면 `benign-code-identifier`. **부정 계수(→0 기대)는 ASCII 쓰기를 유도하지 못하므로** 왜곡 압력이 없다. 대상이 비(非)Latin 로케일 문서가 아니면 `n-a`.

## §2 Funnel 재측정 (M1 첫 단계 — plan.md:46 이행)

```bash
L=/tmp/t573-corpus.txt
find .moai/specs -name SPEC-AC-LOCALE-TOKEN-001 -prune -o \
  \( -name acceptance.md -o -name spec.md -o -name plan.md \) -print > $L
wc -l < $L                                            # → 2276  (코퍼스)
cat $L | xargs /usr/bin/grep -l 'docs-site/content/\(ko\|ja\|zh\)' > /tmp/t573-A1.txt
wc -l < /tmp/t573-A1.txt                              # → 55    (A1 슬래시 철자)
cat $L | xargs /usr/bin/grep -lE 'content/\{[ekjz]' > /tmp/t573-A2.txt
wc -l < /tmp/t573-A2.txt                              # → 98    (A2 중괄호 철자)
cat $L | xargs /usr/bin/grep -lE 'README\.(ko|ja|zh)' > /tmp/t573-A3.txt
wc -l < /tmp/t573-A3.txt                              # → 62    (A3 README 로케일)
sort -u /tmp/t573-A1.txt /tmp/t573-A2.txt /tmp/t573-A3.txt > /tmp/t573-A.txt
wc -l < /tmp/t573-A.txt                               # → 161   (A 합집합)
cat /tmp/t573-A.txt | xargs /usr/bin/grep -l 'grep -[co]' > /tmp/t573-AB.txt
wc -l < /tmp/t573-AB.txt                              # → 66    (A∩B)
cat /tmp/t573-A.txt | xargs /usr/bin/grep -L 'grep -[co]' > /tmp/t573-AminusB.txt
wc -l < /tmp/t573-AminusB.txt                         # → 95    (A∖B)
```

- 경계 선언: `progress.md` 는 동기-단계 증거 기록으로서 구속력 있는 검증 기준이 아니므로 코퍼스에서 제외 (plan §F M1.1, REQ-002).
- A∩B 66파일에서 추출된 계수식 행: `/usr/bin/grep -n 'grep -[co]'` → **397행** 전수 판독하여 C-판정했다 (빈 집합 없음 — 각 단계 계수 상기).
- O-1 이행 (plan-audit review-2): AC-009 판독 표(§4)의 행 수와 **유니크 파일 토큰 수**를 교차 확인했다 — §4 표 행 `grep -c '^| \.moai'` 계수와 `sort -u` 파일 토큰 계수가 모두 95로 일치한다 (§4 말미 실측).

## §3 C-판정 — A∩B 66파일 (REQ-002 스키마, O-2 대로 SPEC 디렉터리별 배치)

### 배치 1 — SPEC 범위 밖 대상 (n-a 다수)

| 파일 | AC id | 토큰 | 대상 문서+로케일 | 계수 의미 | 현재 vs 기준 | 판정 |
|---|---|---|---|---|---|---|
| SPEC-AGENT-MEMORY-DRAIN-001/plan.md | — | `t209`, `AGENT-MEMORY` | git/ls 출력 (문서 아님) | 행 | — | n-a |
| SPEC-AGENT-PARALLEL-OPT-001/spec.md | — | (이력 기술) | 문서 아님 | — | — | n-a |
| SPEC-ANALYZE-FIRST-ROUTING-001/plan.md | — | `MUST INVOKE` | .claude/agents/*.md (en 내부) | 행 | — | n-a |
| SPEC-AUDIT-PARTICIPANT-COUNT-001/plan.md | — | `participant` | Go 소스 | 행 | `0 0 0` | n-a |
| SPEC-CC2178-DOCS-ALIGN-001/plan.md | — | `Diet Constraints\|V0 Abort Gate` | session-handoff.md (en 내부 룰) | 행 | ≥3 | n-a |
| SPEC-CODEX-WIRING-001/acceptance.md | 복수 | `--agent`, `'version'`, `my-own-hook` 등 | .codex 설정/CLI 출력 | 행 | 각 절 명시 | n-a |
| SPEC-E2E-REVIVAL-001/plan.md | — | `e2e`, `Subcommands:` | SKILL.md/CLAUDE.md | 행 | — | n-a |
| SPEC-E2E-REVIVAL-001/spec.md | — | `e2e` | SKILL.md (en) | 행 | 0→15행 | n-a |
| SPEC-FEEDBACK-AUTO-SUBMIT-001/plan.md | — | `--title` | gh CLI 호출문 | 행 | — | n-a |
| SPEC-GLM-EFFORT-TUNE-001/plan.md | — | `reasoning_effort`, `thinking-off` | llm.yaml/Go 소스 | 행 | — | n-a |
| SPEC-HANDOFF-MSGMODE-001/plan.md | — | `11 items`, enum 토큰 | output-styles/rules (en 내부) | 행 | ≥1 | n-a |
| SPEC-SUBCOMMAND-RETIRE-001/acceptance.md | AC-SCR-003a | `name: moai-domain-*` | internal/template/catalog.yaml | 행 | 0 | n-a |
| SPEC-TODO-ENABLE-FLAG-001/spec.md | — | (N1 이력 기술 — 한국어 리터럴 함정 기록) | 문서 아님 | — | — | n-a |
| SPEC-V3R4-LINT-SKIP-CLEANUP-001/plan.md | — | `StatusGitConsistency` | lint 출력 | 행 | — | n-a |
| SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001/plan.md | — | `StatusGitConsistency`, `lint.skip` | lint 출력/diff | 행 | 0 | n-a |
| SPEC-V3R5-DOCS-SECURITY-001/acceptance.md | — | `CWE-732\|…` | CHANGELOG.md | 행 | — | n-a |
| SPEC-V3R5-STATUSLINE-V2145-001/acceptance.md | — | `approved\|pending…`, `^--- (PASS\|FAIL)` | Go 테스트/렌더 출력 | 행 | — | n-a |
| SPEC-V3R6-AGENT-FOLDER-SPLIT-001/plan.md | — | `agents/moai/` | catalog.yaml | 행 | — | n-a |
| SPEC-V3R6-DOCS-I18N-PARITY-001/acceptance.md | C1-C3 | `no frontmatter block` 등 | docs-i18n-check.sh 출력 | 행 | 0 | n-a |
| SPEC-V3R6-DOCS-I18N-PARITY-001/plan.md | — | 동일 | 스크립트 출력 | 행 | — | n-a |
| SPEC-V3R6-HOOK-CONTRACT-FIX-001/acceptance.md | 복수 | `WorktreeCreate` 등 훅 이벤트명 | Go/rules (코드) | 행 | — | n-a |
| SPEC-V3R6-LEGACY-CLEANUP-001/acceptance.md | — | `^ok` | 테스트 출력 | 행 | — | n-a |
| SPEC-V3R6-LEGACY-CLEANUP-001/plan.md | — | `agency` | SKILL.md/rules (en 내부) | 행 | ≤1 | n-a |
| SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001/spec.md | 복수 | `GEARS`, `shall not` | .claude/agents/*.md (en 내부) | 행 | ≥5 등 | n-a |
| SPEC-V3R6-SKILL-CONSOLIDATE-001/plan.md | — | `<old-name>` 등 | catalog.yaml/스킬 (en) | 행 | 0/≥3 | n-a |
| SPEC-V3R6-SKILL-GEARS-ALIGN-001/plan.md | — | `GEARS`, `{#gears-notation}` | 스킬/에이전트 + en 페이지 | 행 | ≥5 등 | n-a |
| SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001/acceptance.md | — | `^- ` | 자기 migration-matrix.md | 행 | — | n-a |
| SPEC-VERSION-STAMP-GUARD-001/acceptance.md | :74 | `config.yaml` 경로 | .moai/docs/version-management.md — **영어 문서** (한글 행 0 실측: `/usr/bin/grep -c '[가-힣]' …` → `0`) | 행 | — | n-a |
| SPEC-VERSION-STAMP-GUARD-001/spec.md | — | `release/` 경로 | 버전 태그 목록 | 행 | — | n-a |
| SPEC-VERSION-STAMP-PREDICATE-001/acceptance.md | 복수 | `aged-out token` 등 영어 구절 | .moai/docs/version-management.md — **영어 문서** (동일 실측) | 행 | ≥1 | n-a |

### 배치 2 — 로케일 문서 대상, 코드 식별자·구조·부정 계수 (benign-code-identifier)

| 파일 | AC id | 토큰 | 대상 문서+로케일 | 계수 의미 | 현재 vs 기준 | 판정 |
|---|---|---|---|---|---|---|
| SPEC-CC2178-TEAM-API-ALIGN-001/acceptance.md | :228-264 | `TeamCreate\|TeamDelete`, `Agent(name` | 4-로케일 docs (코드 식별자 문서화 자리) | 행 | 로케일별 명시 | benign-code-identifier (Go API 식별자) |
| SPEC-DOCS-CODEX-WIRING-CALLOUT-001/acceptance.md | AC-1 | `Home Disk Usage` | 4-로케일 doctor.md | 행 | 각 2 (관측) | benign-code-identifier — **도구 출력 고유 라벨의 인용** (`internal/cli` doctor 검사명, 실측 `internal/cli/doctor_disk_test.go:68` `Name = "Home Disk Usage"`) |
| SPEC-DOCS-CODEX-WIRING-CALLOUT-001/acceptance.md | AC-2·AC-4 | `^## `, `advanced/codex-dual-harness` | 4-로케일 doctor.md | 행 | 7·7·7·7 / 경로 | benign-code-identifier (구조/슬러그) |
| SPEC-DOCS-CODEX-WIRING-CALLOUT-001/plan.md | :44-47 | `^## ` | 4-로케일 doctor.md | 행 | 각 7 | benign-code-identifier (구조) |
| SPEC-DOCS-CODEX-WIRING-CALLOUT-001/spec.md | :40 | `^## ` | 4-로케일 doctor.md | 행 | 각 7 | benign-code-identifier (구조) |
| SPEC-DOCS-LOCALE-PARITY-REPAIR-001/acceptance.md | AC-003 | `desktop-native` | en·zh moai-e2e.md — 플래그 행 한정 | 행 | zh 플래그 행 ≥1 (코드 자리) | benign-code-identifier (CLI 플래그 값·코드 위치 한정) |
| SPEC-DOCS-LOCALE-PARITY-REPAIR-001/acceptance.md | AC-001·002 | `not yet provided`(en) / `尚未提供`(zh) | en/zh moai-e2e.md | 행 | 0 (부정) | benign-code-identifier (부정 계수 — 왜곡 압력 없음) |
| SPEC-DOCS-LOCALE-PARITY-REPAIR-001/acceptance.md | AC-005 | `^\|`, `^#`, 호스트-OS 원어 토큰 | ko/en/ja/zh moai-e2e.md | 행 | 패리티 | benign-code-identifier (구조 + 로케일-확정 원어 토큰 — **올바른 형**) |
| SPEC-DOCS-LOCALE-PARITY-REPAIR-001/acceptance.md | AC-006 | `^moai doctor (permission\|sandbox)` | ja·zh doctor.md | 행 | 0→2 | benign-code-identifier (명령·코드) |
| SPEC-DOCS-LOCALE-PARITY-REPAIR-001/acceptance.md | AC-007 | `\*\*[^*]*\(…` 정규식 | 4-로케일 doctor.md | 행 | 0 | benign-code-identifier (구조·스타일 규칙) |
| SPEC-DOCS-LOCALE-PARITY-REPAIR-001/acceptance.md | AC-008 | `SVG060\|SVG070` | 4-로케일 skill-guide.md | **출현** (`-o\|wc -l`) | 각 ≥2 | benign-code-identifier (규칙 ID — 코드 토큰) |
| SPEC-DOCS-LOCALE-PARITY-REPAIR-001/acceptance.md | AC-009 | `^+.*badge` | diff 출력 | 행 | 0 | benign-code-identifier (부정) |
| SPEC-DOCS-SB-REMOVE-001/acceptance.md | :41 | `^"[^"]*":` | ko/en _meta.yaml | 행 | diff | benign-code-identifier (YAML 키 구조) |
| SPEC-DOCS-SITE-001/acceptance.md | :580-604 | `<url>`, `"@type":` | sitemap/JSON-LD (ko 프리뷰) | 행 | ≥200 등 | benign-code-identifier (구조·메타데이터 키) |
| SPEC-DOCS-V313-CATCHUP-001/acceptance.md | 복수 | `🗿 v3\.1\.3` 등 버전 리터럴, `todo\.enabled`, `manager-lead`, `analyze`(행 앵커 명령명) | README.ko + 4-로케일 pages | 행 | 각 명시 | benign-code-identifier (버전 리터럴·슬러그·명령명 — 로케일 불변 코드) |
| SPEC-DOCS-V313-CATCHUP-001/acceptance.md | :28-29 | `inconclusive` | ko multi-model-audit.md | 행 | 0 (부정) | benign-code-identifier (부정 계수) |
| SPEC-DOCS-V313-CATCHUP-001/plan.md | :45 | 버전 리터럴 | README 4파일 + pages | 행 | 각 명시 | benign-code-identifier |
| SPEC-DOCS-V313-CATCHUP-001/spec.md | F3 | `inconclusive` | ko multi-model-audit.md | 행 | 0 (부정) | benign-code-identifier |
| SPEC-DOCSITE-ADVANCED-001/acceptance.md | :167-220 | `^## `, meta 키, `(/goal\|HUMAN-ONLY\|사용자 전용\|ユーザー専用\|用户专用)` | 4-로케일 advanced | 행 | 패리티 | benign-code-identifier (구조 + 다국어 병렬 패턴 — **올바른 형**) |
| SPEC-DOCSITE-ADVANCED-001/plan.md | :93·205 | `^[a-zA-Z_-]+:` | 4-로케일 _meta.yaml | 행 | 4-로케일 동일 | benign-code-identifier |
| SPEC-DOCSITE-E2E-001/acceptance.md | 복수 | `e2e-tester`, `moai-e2e`, `Playwright` 등, `^## ` | 4-로케일 moai-e2e 관련 | 행 | 명시 | benign-code-identifier (에이전트 id·슬러그·제품 고유명·구조) |
| SPEC-GOAL-DOCS-RETIRE-001/acceptance.md | 복수 | `` `/goal` ``, `` `/moai loop` `` 코드 리터럴 | 4-로케일 autonomous-loops 등 | 행/출현 | 명시 | benign-code-identifier (코드 리터럴 — 파일 자체가 근거 대조 기록: `en:1 ja:1 ko:1 zh:1` 대칭) |
| SPEC-GOAL-DOCS-RETIRE-001/acceptance.md | :98 | `auto mode.*\`/goal\`` | 4-로케일 autonomous-loops.md | 행 | — | benign-code-identifier — 외부 제품(Claude Code) 기능 고유명의 인용. 단 §5 2차 관찰 (ko 현재 0 — 잠재 불일치) |
| SPEC-REF-SEO-ABSORB-001/acceptance.md | :637 등 | `moai-ref-seo`, `^## ` | 4-로케일 skill-guide + SKILL.md | 행 | ≥1 | benign-code-identifier (스킬 id·구조) |
| SPEC-SKILL-GALLERY-BENCH-001/acceptance.md | :104 | `^## ` | README 4파일 | 행 | 4-파일 패리티 | benign-code-identifier (구조) |
| SPEC-V3R6-DOCS-CMD-CATALOG-001/acceptance.md | 복수 | `moai-db`, `moai-github`, `moai-harness` 등 | 4-로케일 _meta/menu/pages | 행 | 0/1 명시 | benign-code-identifier (슬러그·메뉴 경로·frontmatter 키) |
| SPEC-V3R6-DOCS-CMD-CATALOG-001/plan.md | :296 | `No such file` | ls 출력 | 행 | 4 | benign-code-identifier |
| SPEC-V3R6-DOCS-DOCSITE-001/acceptance.md | :47·211·232 | `28 agents\|28개…에이전트\|28个\|28個` 병렬 패턴 | 4-로케일 what-is 페이지 | 행 | 0 | benign-code-identifier (**다국어 병렬 패턴 — 올바른 형**) |
| SPEC-V3R6-DOCS-DOCSITE-001/acceptance.md | :96·111 | `M6["manager-spec` 등 | mermaid 블록 (코드) | 행 | — | benign-code-identifier |
| SPEC-V3R6-DOCS-DOCSITE-001/plan.md | :80-81 | 병렬 패턴 | 4-로케일 | 행 | 0 | benign-code-identifier |
| SPEC-V3R6-DOCS-DOCSITE-001/spec.md | :115 | `8 retained\|8 Retained` (en↔ja parity diff, 명령에 `'...'` 생략) | ja agent-guide | 행 | — | n-a — 실행 불가 스케치(`'...'` 리터럴, ja 대상 파일 미해결). §5 2차 관찰 |
| SPEC-V3R6-DOCS-I18N-COMPLETION-001/acceptance.md | 복수 | `Korean (한국어)`(로케일 전환기 고정 라벨), `feedback.repository`, `sync-auditor`(부정), `Anthropic`(고유명), `graph TB`(코드), 한글 글자 수(원어) | 4-로케일 pages | 행 | 명시 | benign-code-identifier |
| SPEC-V3R6-DOCS-I18N-COMPLETION-001/plan.md | :65 | `feedback.repository`, `sync-auditor` | en/ja/zh moai-feedback | 행 | ≥1/0 | benign-code-identifier |
| SPEC-V3R6-DOCS-I18N-COMPLETION-001/spec.md | :49·59 | `[가-힣]`(원어), `feedback.repository` | 4-로케일 pages | 출현/행 | — | benign-code-identifier |
| SPEC-V3R6-DOCS-POSTREBUILD-CLEANUP-001/acceptance.md | 복수 | `goos`(유출 부정), `/path/to/your-project`(플레이스홀더), `중복\|dedupe\|duplicate`(병렬), `Go 툴체인\|go version`(병렬) | 4-로케일 pages + ko moai-feedback | 행 | 명시 | benign-code-identifier (부정·병렬 패턴) |
| SPEC-V3R6-DOCS-POSTREBUILD-CLEANUP-001/plan.md | :98 | `releaseDate = …` | hugo.toml | 행 | 1 | n-a (설정 파일) |
| SPEC-V3R6-DOCS-RC2-DOCSITE-001/acceptance.md | 복수 | 버전 리터럴, `/Users/`(유출 부정), `grandfather`(용어), `era_final`/`sync_commit_sha`(코드), `template-managed`/`user-owned`(네임스페이스 범주 라벨 — 접두사 의미론과 결합된 준-식별자, 표 라벨로 운용) | 4-로케일 pages | 행 | 명시 | benign-code-identifier — 단 §5 2차 관찰 (ja/zh 이중-글로스 스타일) |
| SPEC-V3R6-DOCS-RC2-DOCSITE-001/plan.md | :140·158 | `grandfather\|era_final\|sync_commit_sha\|3-phase` | 4-로케일 pages | 행 | ≥1 | benign-code-identifier (동일) |
| SPEC-V3R6-DOCS-RC2-README-001/acceptance.md | 대부분 | stale-token 행-앵커 검사 (README.md — en) | en README | 행 | 0/≥1 | n-a (대상이 Latin 로케일) — 인용: 동일 원리라 en 대상은 왜곡 클래스 밖 |
| SPEC-V3R6-DOCS-RC2-README-001/acceptance.md | AC-KO-003 등 | `my-harness-`(부정), `moai-\* …template-managed…`(**"(or the KO equivalent phrasing)"** 병행), `16개`(원어) | README.ko.md | 행 | 명시 | benign-code-identifier (부정 + KO-동등 표현 허용 병기 — **올바른 형**) |
| SPEC-V3R6-DOCS-RC2-README-001/plan.md | :40·55·83 | (acceptance 정책 서술) | — | — | — | n-a |
| SPEC-V3R6-DOCS-RC2-README-001/spec.md | :115·161·203 | (동일 정책 서술) | — | — | — | n-a |
| SPEC-V3R6-DOCS-USER-DRIFT-001/acceptance.md | :17-20 | 로케일별 **원어** 헤딩 토큰 (ko `PR 머지 후 CI 모니터링` / ja `PR 作成後の…` / zh `PR 创建后的…`) | 4-로케일 moai-sync.md | 행 | 각 ≥1 | benign-code-identifier (**로케일-확정 원어 토큰 — 본 SPEC 이 지향하는 올바른 형의 기존 사례**) |
| SPEC-V3R6-DOCS-V3-README-001/acceptance.md | :66·89 | `**Agency** \| 6`(stale 부정, 기대 0), `47개 스킬`(원어), `8개.*retained\|8 retained\|retained 에이전트`(병렬) | README.ko.md | 행 | 0/≥1 | benign-code-identifier (부정 + 원어 + 병렬) |
| SPEC-V3R6-GEARS-MIGRATION-001/acceptance.md | :152·332 | `\|`(구조), `6.*month\|6.*개월\|6.*ヶ月\|6.*个月`(병렬) | 로케일 절 | 행 | ≥7 / ≥1 | benign-code-identifier |
| SPEC-V3R6-WORKFLOW-DOCS-001/plan.md | :31 | `^## ` README 패리티 | README 4파일 | 행 | 12×4 | benign-code-identifier (구조) |

### 배치 3 — 왜곡형 확정 (distorts)

| 파일 | AC id | 토큰 | 대상 문서+로케일 | 계수 의미 | 현재 히트 vs 기준 | 판정 |
|---|---|---|---|---|---|---|
| SPEC-DOCS-LOCALE-PARITY-REPAIR-001/acceptance.md | **AC-004** (§D.4 :72) | `desktop-native` (`grep -ci`) | zh moai-e2e.md — 산문·매트릭스 라벨 자리 | **행** (`-ci`) | zh 7행 vs ≥4 — 충족이 **ASCII 보강 표기(:78-80 라벨 + :84·:98·:176 산문, 총 5행)로만** 성립. 전면 보강 제거 뮤턴트에서 2행 → 미달 (본 트리 재실측: `/usr/bin/grep -o … \| wc -l` → `2`) | **distorts** — t538 sync-audit F2 원인. 비(非)Latin 로케일의 쓰기를 ASCII 괄호 보강으로 왜곡한 관측된 실해 |
| SPEC-V3R6-WORKFLOW-DOCS-001/acceptance.md | **AC-WFD-001** (:12) | `Class A`/`Class B`/`Class C` | ja·zh kanban-mode.md (+ko) — 산문 자리 | 행 | ja·zh 각 1 — **글로스 행 한정**: ja :253 「3つのクラス(Class A · Class B · Class C)」· zh :253 「三个类别(Class A · Class B · Class C)」. 나머지 본문은 전부 원어(クラスA :261·263 / 类别A :261 / 클래스 A). 뮤턴트(글로스 제거 사본)에서 ASCII 계수 0 → 미달, 원어 토큰은 1 유지 (§6 실측) | **distorts** — ASCII 글로스가 계수 급여를 위해 산문에 삽입된 관측된 실해 |
| SPEC-V3R6-WORKFLOW-DOCS-001/acceptance.md | **AC-WFD-007** (:54) | `Implementation Kickoff` | ko·ja·zh spec-lifecycle.md — 머메이드 노드 라벨·표제 자리 | 행 | ko·ja·zh 각 2 — 전부 글로스 자리: ja :33 「実装着手承認(Implementation Kickoff Approval…ヒューマンゲート)」· ko :33 「구현 착수 승인(Implementation Kickoff Approval…)」· zh :33 「实现启动审批(Implementation Kickoff Approval…)」. 원어 단독 계수: ko 3 · ja 3 · zh 3 (실측) | **distorts** — 원어 용어가 이미 운용 중이고 ASCII는 괄호 글로스로만 존재 |
| SPEC-V3R6-WORKFLOW-DOCS-001/acceptance.md | **AC-WFD-011** (:82) | `Functionality`/`Security`/`Craft`/`Consistency` | ja·zh spec-lifecycle.md — 산문 자리 | 행 | ja 1 — :79 「4 次元 — **Functionality / Security / Craft / Consistency** — に採点します」 산문 한복판 ASCII 글로스 | **distorts** — 감사-에이전트 어휘를 산문 자리에 ASCII로 강제 |
| SPEC-V3R6-DOCS-RC2-README-001/acceptance.md | **AC-KO-002b** (:207) 세 번째 불릿 | `8 (retained )?agents` | README.ko.md — What's New 섹션 | 행 | **0** vs ≥1 — 동일 SPEC의 AC-KO-003은 "(or the KO equivalent phrasing)"을 병기하는데 이 불릿만 누락. 현재 두 README 모두 8-에이전트 히어로 카운트 자체가 퇴역 (README.md도 0 실측) | **distorts** — en 문구를 ko README에 그대로 요구 (동종 AC의 KO-동등 허용 관례 미적용) |

**C-판정 집계**: 66파일 = **왜곡형 3파일(기준 5건)** · benign-code-identifier 35파일 · n-a 30파일 — 파일별 대표 판정 기준. 배치 구성: 배치1 30파일 · 배치2 35파일 · 배치3 3파일 (LOCALE-PARITY acceptance와 RC2-README acceptance는 배치2 benign 행과 배치3 distorts 행을 함께 가지므로 30+35+3−2중복=66).

## §4 A∖B 전수 사람 판독 (AC-009 — plan-audit D4)

95파일 전수. 판정 범위 명시(VCI — 스캔 토큰): B-필터 부적중 확인 위에 2-스윕 계수-도구 스캔을 수행했다 — 스윕 1: `awk|wc -l|python3?|rg |sort -u|comm -|sed -n`, 스윕 2: `perl|ruby |node |jq |osascript|wc -c|uniq -c|cut -d|difflib|Counter\(`. 35파일이 적중(개별 판독 행 있음), 60파일은 무적중 — 이들은 로케일 경로 참조만 하고 계수기기가 없어 왜곡형이 될 수 없다(기계 근거 + 표본 대독 3건: FACTORY-BOOTSTRAP acceptance `ls` 존재 확인, WORKFLOW-DOCS spec, GEARS-MIGRATION spec).

| 파일 | 판정 | 근거(1줄) |
|---|---|---|
| SPEC-AUDIT-PARTICIPANT-COUNT-001/spec.md | benign | 로케일 경로 참조, 계수기기 없음 |
| SPEC-CC2178-TEAM-API-ALIGN-001/plan.md | 비-grep 계수 관찰 | `awk -F: '{s+=$2}'` 합계 — 로케일 문서 아님 |
| SPEC-CC2178-TEAM-API-ALIGN-001/spec.md | benign | 로케일 경로 참조만 |
| SPEC-CC2219-UPSTREAM-ALIGN-001/spec.md | benign | 로케일 경로 참조만 |
| SPEC-CCSYNC-DYNWF-001/acceptance.md | 비-grep 계수 관찰 | awk 절-경계 추출 — CLAUDE.md/스킬 (en 내부), 로케일 왜곡 아님 |
| SPEC-CCSYNC-DYNWF-001/plan.md | benign | 로케일 경로 참조만 |
| SPEC-CI-MULTI-LLM-001/plan.md | benign | 로케일 경로 참조만 |
| SPEC-COMPLETION-MARKER-RETIRE-001/acceptance.md | 비-grep 계수 관찰 | `grep -rlE … \| wc -l` = 0 부정 계수 (로케일별) — 부정이라 무압력 |
| SPEC-COMPLETION-MARKER-RETIRE-001/plan.md | benign | 로케일 경로 참조만 |
| SPEC-DEAD-CONFIG-001/spec.md | benign | 로케일 경로 참조만 |
| SPEC-DESIGN-DOCS-V31-001/acceptance.md | 비-grep 계수 관찰 | awk 블록 — 디자인 산출물, 로케일 문서 비대상 |
| SPEC-DESIGN-DOCS-V31-001/spec.md | benign | 로케일 경로 참조만 |
| SPEC-DESIGN-DOCSV2-001/spec.md | benign | 로케일 경로 참조만 |
| SPEC-DESIGN-DOCSV2-001/plan.md | 비-grep 계수 관찰 | `grep -rn …\| wc -l` = 0 부정 (색상 토큰 제거) — static/layouts 대상 |
| SPEC-DOCS-LOCALE-PARITY-REPAIR-001/spec.md | benign | 경계 선언 절 — 기준 진술 없음 |
| SPEC-DOCS-SB-REMOVE-001/spec.md | benign | 로케일 경로 참조만 |
| SPEC-DOCS-SITE-001/plan.md | benign | 로케일 경로 참조만 |
| SPEC-DOCS-SITE-001/spec.md | benign | 로케일 경로 참조만 |
| SPEC-DOCSITE-ADVANCED-001/spec.md | 비-grep 계수 관찰 | `grep -rlP`/`node` — 내부 스캔, 로케일 왜곡 아님 |
| SPEC-DOCSITE-E2E-001/plan.md | benign | `node` 스크립트 호출 — 계수 아님 |
| SPEC-DOCSITE-E2E-001/spec.md | benign | 로케일 경로 참조만 |
| SPEC-FACTORY-BOOTSTRAP-001/acceptance.md | benign | `ls` 존재 확인 — 계수기기 없음 |
| SPEC-FACTORY-BOOTSTRAP-001/plan.md | benign | 로케일 경로 참조만 |
| SPEC-FACTORY-BOOTSTRAP-001/spec.md | benign | 로케일 경로 참조만 |
| SPEC-FALSE-ALLCLEAR-GUARD-001/plan.md | benign | R-8 위험 기술 (계수 판정 아님) |
| SPEC-FALSE-ALLCLEAR-GUARD-001/spec.md | benign | 로케일 경로 참조만 |
| SPEC-GLM-EFFORT-REBALANCE-001/spec.md | benign | 로케일 경로 참조만 |
| SPEC-GLM-FLASH-DEFAULT-001/plan.md | benign | 로케일 경로 참조만 |
| SPEC-GOAL-SURFACE-UNIFY-001/spec.md | benign | 로케일 경로 참조만 |
| SPEC-HANDOFF-CTXGUIDE-001/plan.md | benign | 로케일 경로 참조만 |
| SPEC-HOOK-CONFIG-SAFETY-001/spec.md | benign | 로케일 경로 참조만 |
| SPEC-I18N-001-ARCHIVED/acceptance.md | 비-grep 계수 관찰 | `find … \| wc -l` = 165 파일 수 — 구조; 접두사 없는 `content/zh/` 철자는 §5 이름 붙은 잔여 1건 |
| SPEC-MODEL-PROFILE-MATRIX-002/acceptance.md | benign | 로케일 경로 참조만 |
| SPEC-MX-SCANNER-DOCS-001/acceptance.md | benign | 로케일 경로 참조만 |
| SPEC-MX-SCANNER-DOCS-001/plan.md | benign | `node` 아님 — 내부 코드 서술 |
| SPEC-MX-SCANNER-DOCS-001/spec.md | benign | 로케일 경로 참조만 |
| SPEC-REMOVAL-GUARD-EXTRAS-001/plan.md | benign | 로케일 경로 참조만 |
| SPEC-SESSION-TELEMETRY-001/acceptance.md | 비-grep 계수 관찰 | `grep -rln … \| wc -l` = 0/12 — docs-site 대상 부정/파일 수 (토큰 `context-usage` 코드) |
| SPEC-SKILL-GALLERY-BENCH-001/plan.md | 비-grep 계수 관찰 | `node --version` — 도구 점검, 계수 아님 |
| SPEC-SKILL-GALLERY-BENCH-001/spec.md | 비-grep 계수 관찰 | `node check-svg.mjs` 등 — 스크립트 호출, 계수 아님 |
| SPEC-SESSION-TELEMETRY-001/plan.md | benign | 로케일 경로 참조만 |
| SPEC-SUBCOMMAND-RETIRE-001/plan.md | benign | 로케일 경로 참조만 |
| SPEC-TOOLPOLICY-DEPLOY-REVIEW-001/spec.md | benign | 로케일 경로 참조만 |
| SPEC-UTIL-001/spec.md | benign | tree-sitter 언어명 나열 — 계수 아님 |
| SPEC-UPDATE-VERSION-FLAG-001/plan.md | benign | 로케일 경로 참조만 |
| SPEC-V3R3-DESIGN-PIPELINE-001/plan.md | benign | 로케일 경로 참조만 |
| SPEC-V3R3-HARNESS-001/plan.md | 비-grep 계수 관찰 | `ls \| wc -l` = 23 — 스킬 수, 로케일 문서 아님 |
| SPEC-V3R3-BRAIN-001/spec.md | benign | 16 프로그래밍 언어 나열 — 로케일 무관 |
| SPEC-V3R4-LINT-SKIP-CLEANUP-001/spec.md | benign | 로케일 경로 참조만 |
| SPEC-V3R5-DOCS-SECURITY-001/plan.md | 비-grep 계수 관찰 | `find … \| wc -l` = 4 파일 존재 패리티 — 구조 |
| SPEC-V3R5-DOCS-SECURITY-001/spec.md | benign | 로케일 경로 참조만 |
| SPEC-V3R5-STATUSLINE-FULL-MODE-CLEANUP-001/spec.md | 비-grep 계수 관찰 | `git diff … \| grep -v \| wc -l` — Go 트리 대상 |
| SPEC-V3R5-STATUSLINE-V2145-001/plan.md | benign | `python3` 포크 서술 — 계수 아님 |
| SPEC-V3R5-STATUSLINE-V2145-001/spec.md | benign | 동일 서술 |
| SPEC-V3R6-AGENT-MODEL-ROUTING-001/acceptance.md | 비-grep 계수 관찰 | `grep -l … \| wc -l`·`wc -l` — .claude/agents 대상 (로케일 문서 아님) |
| SPEC-V3R6-AGENT-MODEL-ROUTING-001/plan.md | 비-grep 계수 관찰 | 동일 |
| SPEC-V3R6-AGENT-MODEL-ROUTING-001/spec.md | 비-grep 계수 관찰 | 동일 |
| SPEC-V3R6-ASKUSER-DECISION-MEMORY-001/plan.md | benign | 로케일 경로 참조만 |
| SPEC-V3R6-CLI-AUDIT-001/spec.md | benign | 로케일 경로 참조만 |
| SPEC-V3R6-CLI-CONFIG-INTEGRITY-001/acceptance.md | benign | 로케일 경로 참조만 |
| SPEC-V3R6-CLI-CONFIG-INTEGRITY-001/plan.md | benign | 로케일 경로 참조만 |
| SPEC-V3R6-CODE-COMMENTS-EN-001/spec.md | benign | `sed/awk 금지` 서술 — 계수 아님 |
| SPEC-V3R6-DESIGN-SYSTEM-RETIRE-001/plan.md | benign | 로케일 경로 참조만 |
| SPEC-V3R6-DESIGN-SYSTEM-RETIRE-001/spec.md | benign | 로케일 경로 참조만 |
| SPEC-V3R6-DOCS-CMD-CATALOG-001/spec.md | benign | 로케일 경로 참조만 |
| SPEC-V3R6-DOCS-I18N-PARITY-001/spec.md | benign | 로케일 경로 참조만 |
| SPEC-V3R6-DOCS-POSTREBUILD-CLEANUP-001/spec.md | benign | 로케일 경로 참조만 |
| SPEC-V3R6-DOCS-RC2-DOCSITE-001/spec.md | 비-grep 계수 관찰 | `grep -rln … \| wc -l` = 0 부정 + `sed -n` 행 앵커 (버전 리터럴) |
| SPEC-V3R6-DOCS-USER-DRIFT-001/plan.md | 비-grep 계수 관찰 | `wc -l` 4-로케일 행 수 패리티 — **구조 기준 (올바른 형)** |
| SPEC-V3R6-DOCS-USER-DRIFT-001/spec.md | 비-grep 계수 관찰 | 동일 `wc -l` 패리티 |
| SPEC-V3R6-DOCS-V3-REBUILD-001/acceptance.md | 비-grep 계수 관찰 | `ls \| wc -l`·`find \| wc -l` 4-로케일 동일 — 구조 |
| SPEC-V3R6-DOCS-V3-REBUILD-001/plan.md | benign | 로케일 경로 참조만 |
| SPEC-V3R6-DOCS-V3-REBUILD-001/spec.md | benign | 로케일 경로 참조만 |
| SPEC-V3R6-DOCS-V3-README-001/plan.md | 비-grep 계수 관찰 | `ls -1 … \| wc -l` = 17 — 명령 수, 로케일 문서 아님 |
| SPEC-V3R6-DOCS-V3-README-001/spec.md | benign | 동일 서술 |
| SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001/spec.md | benign | 로케일 경로 참조만 |
| SPEC-V3R6-GEARS-MIGRATION-001/plan.md | benign | 로케일 경로 참조만 |
| SPEC-V3R6-GEARS-MIGRATION-001/spec.md | benign | "53 pages reference EARS" 사전-측정 서술 |
| SPEC-V3R6-GRAPH-FRESHNESS-002/acceptance.md | benign | 로케일 경로 참조만 |
| SPEC-V3R6-GRAPH-FRESHNESS-002/plan.md | benign | 로케일 경로 참조만 |
| SPEC-V3R6-HOOK-CONTRACT-FIX-001/plan.md | 비-grep 계수 관찰 | `git ls-files \| wc -l` — 훅 트리 대상 |
| SPEC-V3R6-HOOK-CONTRACT-FIX-001/spec.md | benign | 로케일 경로 참조만 |
| SPEC-V3R6-LINK-FIX-001/plan.md | benign | 로케일 경로 참조만 |
| SPEC-V3R6-LINK-FIX-001/spec.md | benign | 로케일 경로 참조만 |
| SPEC-V3R6-LEGACY-CLEANUP-001/spec.md | 비-grep 계수 관찰 | `sort -u` — 내부 목록 |
| SPEC-V3R6-LEGACY-CLEANUP-002/spec.md | benign | 로케일 경로 참조만 |
| SPEC-V3R6-PROMPT-CACHE-001/plan.md | benign | 로케일 경로 참조만 |
| SPEC-V3R6-SEQ-THINKING-RETIRE-001/spec.md | benign | 로케일 경로 참조만 |
| SPEC-V3R6-SKILL-COMPRESS-001/plan.md | 비-grep 계수 관찰 | `awk '/^triggers:/'` 추출 — 스킬 대상 |
| SPEC-V3R6-SKILL-CONSOLIDATE-001/spec.md | benign | 로케일 경로 참조만 |
| SPEC-V3R6-SKILL-GEARS-ALIGN-001/spec.md | benign | 로케일 경로 참조만 |
| SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001/spec.md | benign | 로케일 경로 참조만 |
| SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001/plan.md | 비-grep 계수 관찰 | `grep -rlP … \| wc -l`·awk — 템플릿 트리 대상 |
| SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001/spec.md | benign | perl PCRE 동등성 서술 |
| SPEC-V3R6-WORKFLOW-DOCS-001/spec.md | benign | §C.4 "locale-verbatim" 정책 근원 — AC-001/007/011 왜곡의 설계 출처 (수리는 acceptance.md 축 — 배치 3) |

**A∖B 집계**: 95 = benign 60 · 비-grep 계수 관찰 35 · 왜곡형 0. (비-grep 계수 관찰 35건 전부 로케일 왜곡형이 아님 — wc/awk/부정-계수가 로케일 문서 산문을 겨냥한 건 없다.)

**O-1 유니크 토큰 교차 확인**: 본 표의 파일 행 수 = `grep -c '^| \.moai'` 기준 95 · `sort -u` 유니크 파일 토큰 = 95 — 중복 행 슬랙 0.

## §5 이름 붙은 잔여 철자 판정 (AC-010 — plan M1.6)

| # | 철자 | 관측 위치 | 판정 | 근거 |
|---|---|---|---|---|
| 1 | 접두사 없는 `content/zh/` | SPEC-I18N-001-ARCHIVED (코퍼스 `find content/{en,zh,ja}` — archaic .mdx 철자, A 필터 밖) | benign | archived SPEC의 당대 경로 관습. `find … -name "*.mdx" \| wc -l` = 165은 파일-수 구조 계수 — 산문 토큰 강제 없음 |
| 2 | `$loc` 변수 보간 경로 | SPEC-CC2178-TEAM-API-ALIGN-001/acceptance.md:228 등 | benign | `grep -c 'TeamCreate\|TeamDelete' "docs-site/content/$loc/…"` — 계수 토큰이 Go API 식별자 (코드 위치). 보간은 경로 기술일 뿐 |
| 3 | 접두사 없는 `ko/advanced/…` | SPEC-DOCS-V313-CATCHUP-001/spec.md:64 | benign | `grep -c 'inconclusive' ko/advanced/multi-model-audit.md` — **부정 계수(→0)** + `inconclusive`는 코드 응답 값. 부정 계수는 ASCII 쓰기를 유도하지 않는다 |
| 4 | `.moai/docs/*.md` 대상 계수 | SPEC-VERSION-STAMP-GUARD-001/acceptance.md:74 (+ SPEC-VERSION-STAMP-PREDICATE-001 전반) | n-a | 대상 `.moai/docs/version-management.md` 는 영어 문서 — 본 트리 실측 `/usr/bin/grep -c '[가-힣]' .moai/docs/version-management.md` → `0`. 비(非)Latin 로케일 문서가 아니므로 왜곡 클래스 밖 |

## §6 뮤턴트 probe 기록 (M2 사전 — REQ-001(d) 근거, 본 트리 재실측)

```bash
sed 's/, desktop-native)/)/g; s/ (desktop-native)//g' \
  docs-site/content/zh/utility-commands/moai-e2e.md > /tmp/t573-zh-reverted.md
/usr/bin/grep -o 'desktop-native' /tmp/t573-zh-reverted.md | wc -l        # → 2  (옛 기준 ≥4 미달 — 옛 기준은 왜곡을 강제)
/usr/bin/grep -c '原生桌面' /tmp/t573-zh-reverted.md                       # → 6  (원어 기준 — 보강 제거 후에도 통과)
/usr/bin/grep -c '^| \*\*原生桌面' docs-site/content/zh/utility-commands/moai-e2e.md   # → 3  (원본)
/usr/bin/grep -c '^| \*\*原生桌面' /tmp/t573-zh-reverted.md               # → 3  (뮤턴트 — 동일)
sed 's/(Class A · Class B · Class C)//g' docs-site/content/ja/advanced/kanban-mode.md > /tmp/t573-ja-mutant.md
/usr/bin/grep -c 'Class A' /tmp/t573-ja-mutant.md                          # → 0  (옛 AC-WFD-001 기준 — 뮤턴트에서 FAIL)
/usr/bin/grep -cE 'クラス ?A' /tmp/t573-ja-mutant.md                       # → 1  (재작성 기준 — 뮤턴트에서 PASS)
```

→ 원어-토큰/병렬-토큰 기준은 ASCII 보강 제거 사본에서도 통과 — 왜곡 압력 없음이 관측으로 확인됐다. 옛 기준은 정확히 왜곡을 **수리하는** 편집(글로스 제거)에서 실패한다 — 기준이 왜곡을 급여하는 구조의 실증.

## §7 2차 관찰 (AC-008 — 기록만, 수정 없음)

1. **SPEC-GOAL-DOCS-RETIRE-001 :98** `auto mode.*\`/goal\`` — ko autonomous-loops.md 현재 히트 0 (ja 1·zh 1 실측). 완료 SPEC의 잠재 불일치 소지 — 본 카드 범위 밖, 기록만.
2. **SPEC-V3R6-DOCS-DOCSITE-001/spec.md:115** — `diff <(grep -c '8 retained…' en/…) <(grep -c '...' ja/…)` — 명령에 `'...'` 리터럴이 남아 실행 불가 스케치이고 ja agent-guide 대상 경로가 미해결. 이 SPEC 자체는 A∩B 소속이므로 C-판정은 배치 2의 n-a 행에 기록했다.
3. **SPEC-V3R6-DOCS-RC2-DOCSITE-001** — ja/zh harness-engineering.md :181 「汎用配布 (template-managed)」 스타일의 이중-글로스 헤딩. 네임스페이스 범주 라벨(접두사 의미론 결합)로 benign 판정했으나, 헤딩 자리의 이중 글로스는 ja/zh 표기 관습상 과잉 — 표기 축(M3 성격)의 스타일 관찰.
4. **계열 근접 미스 (t538 AC-008/AC-011 동형)** — 배치 2의 `grep -c` 행-계수 중 산문 대상(예: AC-WFD-011 dimension 행)은 단위 불일치 위험을 안고 있으나, 본 카드의 재작성으로 해당 기준은 병렬-토큰 형으로 바뀌므로 별도 수리 불요. 기타 단위 불일치 후보는 발견하지 못했다 (397행 전수 판독 근거).

## §8 판정 총계

| 구분 | 대상 수 | distorts | benign-code-identifier | n-a / benign(비-grep 관찰 포함) |
|---|---|---|---|---|
| A∩B C-판정 (파일) | 66 | **3파일 / 기준 5건** | 26 | 37 |
| A∖B 사람 판독 (파일) | 95 | 0 | — (구조·부정 계수 포함) | benign 60 + 비-grep 관찰 35 |
| 이름 붙은 잔여 철자 | 4 | 0 | 3 | 1 (n-a) |

재작성 대상 = C-판정 왜곡형 5건 (§3 배치 3) — M2에서 5건 전량 재작성한다 (census 확정 건수 = 재작성 건수).
