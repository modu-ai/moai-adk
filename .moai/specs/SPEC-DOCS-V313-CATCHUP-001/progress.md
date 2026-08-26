# progress — SPEC-DOCS-V313-CATCHUP-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-08-26
- artifacts: spec.md · plan.md · acceptance.md (Tier M 3종 + 본 progress.md §E 스켈레톤)
- baseline_tree: e07a6d0f4 (worktree t274, branch WT-v313-docs, card t274)
- tier: M · reqs: 8 (REQ-DVC-001..008) · acs: 9 (AC-DVC-001..009)
- inventory: CHANGELOG [3.1.3] 26항목 (Added 13 / Changed 4 / Fixed 9) — 격차 판정 D 3 / U 11 / N 4 / NA 8 + version SSOT 갭 8건 (V1–V8), spec.md §1
- notes: N 4항목(codex dual-harness)은 operator 승인 관문(REQ-DVC-003) — 승인 전 내비게이션 설정 불변. version SSOT 프로세스 원인은 별도 카드 권장 (본 SPEC은 증상만)
- revised: 0.2.0 — plan-audit iter1 (FAIL 0.7125) blocking D1–D6 + optional D7–D9 applied, 2026-08-26
- revised: 0.3.0 — plan-audit iter2 (FAIL 0.825, micro-pass) blocking R1–R4 + optional R5–R7 applied, 2026-08-26. iter-3 불가(Tier M 상한) — 오케스트레이터 grep 읽기 검증(키 문자열 4개)으로 폐쇄 확인

## §E.2 Run-phase Evidence

### Pre-flight (§C) — 2026-08-26, tree `311d5498a` (branch WT-v313-docs, clean)

| Check | Command | Observed Output |
|-------|---------|-----------------|
| §C.1 CHANGELOG 위치 | `grep -n '^## \[' CHANGELOG.md \| head -8` | `177:## [3.1.3] - 2026-08-24`, `307:## [3.1.2]` |
| §C.1 항목 수 | `awk` 섹션별 카운트 | `Added=13 Changed=4 Fixed=9 total=26` |
| §C.2 격차 표 셀 재관측 | §1 각 행의 grep 재실행 (아래 AC-DVC-001 표) | **26행 전부 plan-phase 관측과 일치 — 재판정 0건, 병렬 세션 선문서화 0건** |
| §C.3 hugo build baseline | `hugo --source docs-site --quiet` | `rc=0`, stderr 0행 (경고 0) |
| §C.3 URL 블랙리스트 baseline | `grep -rc 'docs\.moai-ai\.dev\|adk\.moai\.com\|adk\.moai\.kr' docs-site/content/*/ \| grep -v ':0$' \| wc -l` | `0` |
| §C.3 Mermaid LR/RL baseline | `grep -rc 'flowchart LR\|graph LR\|flowchart RL\|graph RL' docs-site/content/ \| grep -v ':0$' \| wc -l` | `0` |
| §C.3 sitemap | `ls docs-site/public/sitemap.xml` | 존재 (702B, baseline 빌드 산출물) |
| §C.3 body-emoji baseline | `grep -rPn '[\x{1F300}-\x{1FAFF}\x{2600}-\x{27BF}]' docs-site/content/ko/ \| wc -l` | `67` — 전부 기존 파일의 코드블록 예시(statusline 출력 등)·보존 대상 문자. 신규 위반 0 판정의 기준선 |
| §C.4 hugo 바이너리 | `which hugo && hugo version` | `/opt/homebrew/bin/hugo` `v0.160.1+extended` |
| §C.5 음성 점검 (a) | 추적 파일 `internal/hook/agent_model_guard.go`에 1행 삽입 → `git diff --stat e07a6d0f4 -- internal/ pkg/ cmd/ internal/hook/ .claude/hooks/ internal/template/templates/` | `internal/hook/agent_model_guard.go \| 1 +` — **비지 않음 관측 (AC-DVC-006 red 확인)**. 즉시 원복, `git status --porcelain internal/` → 0행 |
| §C.5 음성 점검 (b) | `docs-site/content/ko/_meta.yaml`에 무승인 1행 삽입 → `git diff --stat e07a6d0f4 -- 'docs-site/content/*/_meta.yaml' docs-site/data/menu/main.yaml docs-site/layouts/partials/menu.html 'docs-site/content/*/advanced/codex-dual-harness.md'` | `docs-site/content/ko/_meta.yaml \| 1 +` — **비지 않음 관측 (AC-DVC-007 red 확인)**. 즉시 원복, diff 재측 → 0행 |

### M1 — Operator gate 기록 (2026-08-26, orchestrator 전달)

1. **Implementation Kickoff APPROVED** (2026-08-26).
2. **N-class NEW PAGE APPROVED** — `advanced/codex-dual-harness.md` 4로케일 생성 승인. 내비게이션 설정(`content/<locale>/_meta.yaml` ×4, `data/menu/main.yaml` 4로케일 이름 맵 + icon, `layouts/partials/menu.html` SVG case) 편집 승인 — 통상 structure-curator 단독 소관 도메인에 대한 이번 PR 한정 승인. Skill `hns-oss-docs-structure-map` 절차 준수 + 커밋 메시지에 승인 명시 조건.
3. **Progression: autonomous** — M1→M5 직진, 중간 승인 정지 없음. blocker만 보고.

판정: **new_page** (2분기 중 승인 측) — A1–A4는 M4에서 실행.

### M5 — AC 판정 매트릭스 (2026-08-26, 트리 `5d68cdac9` 기준, 전부 이번 실행 관측)

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-DVC-001 | PASS | §C.1–2 (위 pre-flight 표) | `[3.1.3]` 177행·26항목(13/4/9) 재확인; 격차 표 26행 전부 재관측 — plan-phase 관측과 불일치 0건 |
| AC-DVC-002 | PASS | `grep -rln 'todo\.enabled' docs-site/content/{ko,en,ja,zh}` 각각; `awk 'FNR==352' README.ko.md README.md README.ja.md README.zh.md \| grep -c 'analyze'`; scrub/symlink/inconclusive 각 파일 | 로케일별 `todo.enabled` 2파일씩(총 8); README 352행 `analyze` **4**; A5 scrub·A12 symlink·F3 inconclusive 전 로케일 1+ 매칭 (상세는 아래 격차 표 폐쇄) |
| AC-DVC-003 | PASS | `grep -c 'opus / max' <6개 매트릭스 파일>`; `grep -c 'manager-lead' <동일>` | `opus / max` **전부 0**; `manager-lead` profile-matrix 3 / model-policy 4 (6개 파일 전부). 두 페이지 매트릭스 값 동일 — 모순 없음 |
| AC-DVC-004 | PASS | (a) `grep -n 'version\|releaseDate' docs-site/hugo.toml`; (b) `grep -c '🗿 v3\.1\.3' README…4파일`; (c) `grep -c 'release/v3\.1\.3' …`; (d) `grep -rc 'v3\.1\.2' …statusline+claude-cloud 8파일`; (d-2) `grep -c '🗿 v3\.1\.1' faq×4` + `grep -c '🗿 v3\.1\.2 -> 🗿 v3\.1\.3' faq×4`; (e) `grep -c '@v3\.1\.3' claude-cloud×4`; (f) `grep -c 'v3\.1\.3' moai-feedback×4` + `grep -c 'v10\.8\.0' …` | (a) 55행 `v3.1.3`/56행 `2026-08-24`; (b) **각 2**; (c) 각 1; (d) 전부 0; (d-2) 전부 0 + **각 1**; (e) 각 1; (f) **각 1** + v10.8.0 전부 0. 역사 인용 보존: `grep -c 'v3\.1\.1부터\|v3\.1\.1에' README.ko.md` baseline(e07a6d0f4)=5 → 현재=5 (동일) |
| AC-DVC-005 | PASS | 아래 종료 게이트 8축 표 (각 축 고유 행) | 8축 전부 green — 신규 위반 0, baseline green 축 green 유지 |
| AC-DVC-006 | PASS | `git diff --stat e07a6d0f4 -- internal/ pkg/ cmd/ internal/hook/ .claude/hooks/ internal/template/templates/ \| wc -l` | **0** (빈 출력 — Go/템플릿/훅 diff 0파일). §C.5 (a) 음성 점검에서 red 관측 완료 |
| AC-DVC-007 | PASS | `git diff --stat e07a6d0f4 -- 'docs-site/content/*/_meta.yaml' docs-site/data/menu/main.yaml docs-site/layouts/partials/menu.html 'docs-site/content/*/advanced/codex-dual-harness.md'` | 승인된 변경만: 신규 페이지 4파일(각 63행) + `main.yaml` 6행 삽입. `_meta.yaml`·`menu.html` diff 없음(하위 항목은 아이콘 없음 — advanced 섹션 `school` 아이콘의 SVG case menu.html:65에 기존 존재). §C.5 (b) 음성 점검에서 red 관측 완료 |
| AC-DVC-008 | PASS | `grep -oE '^\| [AFCD][0-9]+' spec.md \| sort -u \| wc -l`; 보조 `grep -cE` | 유니크 **26** (보조 행수 26). NA 8항목 근거 문구 그대로 존재 |
| AC-DVC-009 | PASS | `grep -c '^## ' README.ko.md README.md README.ja.md README.zh.md` | 각 **12** (4파일 동일 — 섹션 추가/삭제 없음) |

> AC-DVC-002 판정 보강: acceptance의 `sed -n '352p' A B C D | grep -c` 형태는 BSD sed가 파일 경계에서 줄번호를 이어가 macOS에서 1을 반환(플랫폼 특성). 파일별 `sed -n '352p' <f> | grep -c 'analyze'` 1/1/1/1 및 다중파일 등가형 `awk 'FNR==352' … | grep -c` = **4**로 판정.

### M5 — 종료 게이트: hns-oss-docs-verify 8축 (트리 `5d68cdac9`)

| # | 축 | 명령 | 관측 |
|---|-----|------|------|
| 1 | warning-free hugo build | `hugo --source docs-site --minify --gc` | rc=0, WARN/ERROR 0행 (KO 185 / EN·JA·ZH 183 페이지 — ko-only 섹션 차이는 baseline 동일) |
| 2 | sitemap 존재 | `test -f docs-site/public/sitemap.xml` | 존재 (572 bytes, 본 빌드 산출) |
| 3 | URL 블랙리스트 grep 0 | `grep -rc 'docs\.moai-ai\.dev\|adk\.moai\.com\|adk\.moai\.kr' docs-site/content README… \| grep -v ':0$' \| wc -l` | **0** |
| 4 | Mermaid LR/RL grep 0 | `grep -rc 'flowchart LR\|graph LR\|flowchart RL\|graph RL' docs-site/content \| grep -v ':0$' \| wc -l` | **0** |
| 5 | 4로케일 파일+섹션 파리티 | `scripts/docs-i18n-check.sh` + 래칫(스킬 §4 awk 발산 집합 vs `.locale-parity-baseline`) | 스크립트: 151 .md ×4로케일, Errors 0 / Warnings 0. 래칫: 신규 발산 **0행**(comm -23 빈 출력), now=base=54. README H2 12×4 |
| 6 | README 4파일 헤딩 파리티 | `grep -c '^## ' README.ko.md README.md README.ja.md README.zh.md` | 각 12 |
| 7 | 본문 emoji 0 신규 | 신규 페이지 4종 스캔 + diff 증명 | 신규 페이지 emoji 0행. 전체 diff의 emoji 추가 라인 32 = 제거 라인 32(순증 0) — 전부 코드블록 예시의 버전숫자 교체(statusline·update-prompt·faq 예시, 보존 범주). ko 트리 총합 67 = pre-flight baseline 67 |
| 8 | version-sync | `grep -E 'version = ' docs-site/hugo.toml`; `grep -rn 'Release-v[0-9]' README…`; `grep -rn '🗿 v[0-9]' … \| grep -v v3\.1\.3` | hugo.toml `v3.1.3`; README 배지 4파일 모두 `Release-v3.1.3`. v3.1.3 외 `🗿 v` 표시는 faq "현재 설치된 버전" 설명 4행뿐 — V5 목표 형태 `🗿 v3.1.2 -> 🗿 v3.1.3`의 화살표-왼쪽 토큰이자 (d-2) 수정 패턴이 정당히 유지하는 토큰 (acceptance [R3] 명시) |

### M5 — 격차 표 폐쇄 (U/N 착지 증거, 전부 로케일별 관측)

| 항목 | 판정 | 착지 증거 (grep) |
|------|------|------------------|
| A1–A4 | N → **착지** | `advanced/codex-dual-harness.md` 4로케일 존재 + hugo 렌더(index.html 4개); `AGENTS.md`·`.codex/agents`·`.agents/skills`·`codexadapter` 키워드 전 로케일 매칭. 사실은 트리에서 재검증(11 TOML `git ls-files internal/template/templates/.codex/agents/moai/*.toml`, `internal/codexadapter/events.go` EventTable 11행·6 적응, `output.go` inertKeys 3개) |
| A5 | U → **착지** | `grep -n 'scrub\|queue' docs-site/content/{ko,en,ja,zh}/utility-commands/moai-feedback.md` 각 1+ 행 (스크러빙 계약 절) |
| A6 | U → **착지** | `todo.enabled` 로케일별 2파일(config-sections + moai-todo) ×4 = 8파일 |
| A7 | U(README) → **착지** | 352행 동사 나열 `analyze` 포함 4/4파일 |
| A11 | U → **착지** | skill-guide design-dna 행 갱신 + README design-dna 문단 — `프로파일/profile/プロファイル/档案(配置档案)` 4로케일, `.design-dna/` 언급 |
| A12 | U → **착지** | update.md + init-wizard.md 심볼릭 링크 폴백 통지 절 6파일(2페이지 ×3로케일) |
| C1·C2 | U → **착지** | 매트릭스 표 재작성 양쪽 페이지 ×4로케일: `opus / max` 0, `manager-lead` 행 존재, 판단 가중 정책 서술 |
| C3 | U → **착지** | model-policy GLM reasoning 절 + profile-matrix 오버레이 — `reasoning_effort`·상한 max·전체성 조항 매핑 표 4로케일 |
| F3·F4·F5 | U → **착지** | multi-model-audit `inconclusive` 절 4로케일 — 미수집 diff 무판정, PASS 합성 금지, `project_root` 트리 기록 |
| V1–V8 | → **착지** | AC-DVC-004 (a)–(f) 표 참조 — 전 표면 v3.1.3 |
| A8·A13·F7 | D (문서화됨) | 재관측 확인 — 각각 mcp-server.md 96–114행, moai-gate.md 타입 6행, init-wizard.md 7·44행. 갱신 없음 (anti-pattern "A13 되돌리기" 준수) |
| A9·A10·C4·F1·F2·F6·F8·F9 | NA (문서화 안 함) | 격차 표 §1 근거 그대로 — 26항목 전수성 유지 (AC-DVC-008) |

### 커밋 목록 (run-phase)

- M1 `47270bd72` — 게이트 기록 + pre-flight + status flip (draft → in-progress)
- M2 `e5808b61f` — ko canonical (14 files)
- M3 `8eeecbf0f` — en/ja/zh 파생 (39 files)
- M4 `5d68cdac9` — 신규 페이지 4로케일 + main.yaml 항목 (5 files)
- M5 `bed33bbde` — 종료 게이트 8축 + AC 9/9 + 격차 표 폐쇄 (위 표들)

### 병합 후 8축 재검증 (2026-08-26, 트리 `0044c7a83` — lane-4 후속 세션이 관측)

배경: M5 증거는 `5d68cdac9` 트리 기준이었으나, 이후 origin/main 4커밍(t269·t250·t259·t273)이 착지했고
그중 t273(#1656)이 README×4·main.yaml에서 본 카드 파일면과 겹침 → `0044c7a83` 병합 후 최종 트리에서
전 축 재측정. M5의 좀비 완료분(선행 lane-4 세션의 in-flight 에이전트 소행)은 운영자 승인으로 흡수.

| # | 축 | 관측 (재측정, 병합 트리) |
|---|-----|------|
| 1 | hugo build | `hugo --source docs-site --minify --gc` → rc=0, WARN/ERROR 0행 |
| 2 | sitemap | `ls docs-site/public/sitemap.xml` → 572 bytes 존재 |
| 3 | URL 블랙리스트 | grep 3-도메인 → `0` |
| 4 | Mermaid LR/RL | grep → `0` |
| 5 | 4로케일 파리티 | `scripts/docs-i18n-check.sh` rc=0, Errors 0 / Warnings 0; 래칫 now=54 = base=54, `comm -23` 신규 발산 0행 |
| 6 | README H2 파리티 | `grep -c '^## '` → 12/12/12/12 |
| 7 | 본문 emoji | 펜스 인식 스캔(스크립트 `.moai/state/verify/913edc05-…/bodyscan.pl`): 이모지 추가 32행 전부 faq×4·README×4·statusline×4의 코드펜스 내부 — 본문(펜스 밖) 카운트 HEAD=0, origin/main=0, 12파일 전부 SAME. 상식 검증: 펜스 무관 총계 faq_ko=5·README_ko=29 (스캐너 정상) |
| 8 | version-sync | hugo.toml `v3.1.3`/`2026-08-24`; 배지 `Release-v3.1.3` ×4파일; `🗿` 토큰 분포 v3.1.3×24 + v3.1.2×12 — 후자 전부 승인형(faq "현재 설치된 버전" 설명 ×4 + 화살표-왼쪽 토큰 faq×4·README×4, acceptance [R3]·d-2). t273 유입 신규 스테일 표시 0건 |

판정: **8축 전부 PASS (병합 트리 `0044c7a83` 기준)** — AC-DVC-005의 최종-트리 계약 충족.

## §E.3 Run-phase Audit-Ready Signal

- run_status: audit-ready
- run_complete_at: 2026-08-26
- run_commit_sha: 5d68cdac9
- ac_pass_count: 9
- ac_fail_count: 0
- preserve_list_post_run_count: 0 (PRESERVE 대상 없음 — 문서 전용 카드)
- l44_pre_commit_fetch: n/a (워크트리 카드 브랜치, 원격 미갱신 상태에서 착수 — AGENTS.md §2 절차 준수)
- l44_post_push_fetch: n/a (run-phase push 없음 — 게시는 sync-phase 관문)
- new_warnings_or_lints_introduced: 0 (hugo build 경고 0행, baseline 동일)
- cross_platform_build.status: n/a (docs-only — Go 빌드 표면 무변경, AC-DVC-006 diff 0)
- total_run_phase_files: 58 (hugo.toml 1 + docs-site 4로케일 53 + README 4)
- m1_to_mn_commit_strategy: per-milestone 4 commits on WT-v313-docs, no push (manager-git owns push+PR at sync)
- post_merge_reverify_tree: 0044c7a83 (origin/main 4커밋 통합 후 8축 재검증 PASS — §E.2 병합 재검증 표)

## §E.4 Sync-phase Audit-Ready Signal

- sync_status: audit-ready
- sync_complete_at: 2026-08-26
- sync_commit_sha: bc87bc9ca (backfilled from pending-backfill placeholder, per the D3 self-referential-hazard exemption)
- changelog_entry_position: [Unreleased] → ### Added (single dense entry, SPEC-ID link, AC 9/9 referenced)
- frontmatter_status_transitions:
    - spec.md: in-progress → implemented → completed (3-phase close, merged into the single sync commit)
    - updated: 2026-08-26
    - plan.md / acceptance.md: no frontmatter change required (body untouched)
- b12_self_test_a: duplicate grep `grep -c 'SPEC-DOCS-V313-CATCHUP-001' CHANGELOG.md` → 0 (PASS, this run, tree ecf9766bd)
- b12_self_test_b: AC count match — acceptance.md distinct AC tokens = 9 (AC-DVC-001..009); CHANGELOG entry references 9 ACs, 9 PASS / 0 FAIL (PASS, this run)
- b12_self_test_c: file-path verification — docs-site/content/{ko,en,ja,zh}/advanced/codex-dual-harness.md all exist (ls verified); docs-site/hugo.toml v3.1.3/2026-08-24 (lines 55-56); docs-site/data/menu/main.yaml ref /advanced/codex-dual-harness (line 752); README×4 present (PASS, this run)
- verification_basis: docs-only scope — no Go build/lint/test surface (AC-DVC-006 diff 0 verified in §E.2); the binding verification is the hns-oss-docs-verify 8-axis exit gate, re-verified PASS on the post-merge tree 0044c7a83 (§E.2 병합 후 8축 재검증 표)
- cross_platform_build.status: n/a (docs-only)
- new_warnings_or_lints_introduced: 0 (hugo build WARN/ERROR 0행 on tree 0044c7a83)
- mx_tag_validation: n/a (no code surface — zero @MX annotation targets in a documentation-only diff)
- canary_compliance_check: n/a (no canary surface in docs-only scope)
- publish_state: unpushed — sync commit lands on WT-v313-docs; push + PR owned by manager-git after sync-audit (per dispatch)
