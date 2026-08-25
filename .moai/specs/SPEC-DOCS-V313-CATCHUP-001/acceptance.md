# acceptance — SPEC-DOCS-V313-CATCHUP-001

> 검증 계약. 모든 AC는 Given-When-Then 이중 표현이며, 각 AC는 **RED-now** (baseline `e07a6d0f4`에서의 관측)와 **green-path** (어느 마일스톤이 뒤집는지 + 통과 출력)를 짝으로 갖는다 (verification-completeness §2 two-cell).

## §D AC Matrix

### AC-DVC-001 — 격차 표 재검증 (REQ-DVC-008)

- **Given** CHANGELOG `[3.1.3]` 26항목과 spec.md §1 격차 표
- **When** run-phase가 시작될 때
- **Then** 26행의 판정 셀 전부가 run-phase 트리에서 명령 재실행으로 재관측되고, plan-phase 관측과 다른 셀은 재판정·표 갱신 후 작업된다
- **RED-now**: 관측 완료 상태 (plan-phase에서 26행 전부 관측됨 — §1의 명령 인용) — 이 AC는 run-phase 착수 조건이므로 착수 시점에 green
- **green-path**: M2 착수 전 §C pre-flight 1–2 실행, 재관측 결과를 progress.md §E.2에 기록

### AC-DVC-002 — U 항목 로케일별 착지 (REQ-DVC-001, REQ-DVC-002)

- **Given** U 판정 11항목 (A5·A6·A7-README·A11·A12·C1·C2·C3·F3·F4·F5)
- **When** 문서화가 완료되었을 때
- **Then** 각 항목의 식별 키워드가 docs-site 4로케일 대상 페이지에서 **로케일별로** 1+ 매칭되고, README 대상 항목(A7·A11)은 4파일 모두에 반영된다 [D5: 한 로케일에 4파일이 몰리는 뮤턴트 폐쇄 — 로케일별 판정]
- **RED-now**: 대표 관측 — `grep -rln 'todo\.enabled' docs-site/content README.md README.ko.md README.ja.md README.zh.md` → 0파일; `grep -n 'scrub\|queue' docs-site/content/ko/utility-commands/moai-feedback.md` → 0행 (동사 미문서화); `grep -n 'symlink' docs-site/content/ko/cli-reference/update.md docs-site/content/ko/getting-started/init-wizard.md` → 0행; `sed -n '352p' README.ko.md` → 동사 나열에 `analyze` 부재
- **green-path**: M2+M3 완료 후 항목별 grep을 **로케일별로** 실행 — `grep -rln 'todo\.enabled' docs-site/content/ko` · `/en` · `/ja` · `/zh` 가 각각 비어 있지 않은 목록 반환 (예: `grep -rln 'todo\.enabled' docs-site/content/ko` → `docs-site/content/ko/advanced/config-sections.md` 1+ 행). 총계 `grep -rln 'todo\.enabled' docs-site/content | wc -l` → `4`는 보조 확인일 뿐 판정은 아니다. README: `grep -c 'analyze' README.ko.md README.md README.ja.md README.zh.md` → 각 `1+` (352행 동사 나열)

### AC-DVC-003 — 프로필 매트릭스 정책값 (REQ-DVC-002; C1·C2)

- **Given** v3.1.3 judgment-weighted 정책 (심사·조율 행 high; `manager-spec`·`manager-develop` 전 열 medium; `manager-docs` sonnet/low 전 열; 어떤 행도 max 아님; `manager-lead` 매핑 행 존재)
- **When** ko 정본 갱신이 완료되었을 때
- **Then** `profile-matrix.md`와 `model-policy.md` 양쪽 매트릭스 표가 모두 새 정책값을 담고, 두 페이지가 서로 모순되지 않는다
- **RED-now**: `grep -n 'opus / max' docs-site/content/ko/advanced/profile-matrix.md` → 30·31행 (구 정책 잔존); `grep -c 'manager-lead' docs-site/content/ko/advanced/profile-matrix.md` → `0`
- **green-path**: M2 후 `grep -c 'opus / max' …profile-matrix.md` → `0`, `manager-lead` 행 1+ 매칭 (model-policy.md 동일 기준)

### AC-DVC-004 — version SSOT 일관 (REQ-DVC-004; V1–V8) [D1: 전 표면 열거, D4: 목표값 고정·뮤턴트 폐쇄]

- **Given** v3.1.3 릴리즈(2026-08-24)와 §1.4의 8개 스테일 표면 (V1 hugo.toml, V2 README:491, V3 README:766, V4 statusline.md:22 ×4, V5 faq.md ×4, V6 claude-cloud.md ×4, V7 README:493, V8 moai-feedback.md:62 ×4)
- **When** run-phase가 완료되었을 때
- **Then** 8개 표면 전부가 §1.4의 목표값을 읽는다 — hugo.toml `v3.1.3`/`2026-08-24`; README 491행 `🗿 v3.1.3`; **766행 `🗿 v3.1.2 -> 🗿 v3.1.3`** (491행만 고치고 766행을 남기는 뮤턴트 폐쇄 — 아래 카운트 `2+`가 둘 다 잡음); 493행 `release/v3.1.3`; docs-site statusline 예시 `🗿 v3.1.3`, faq update 예시 `🗿 v3.1.2 -> 🗿 v3.1.3` + 단독 예시 `🗿 v3.1.3`, claude-cloud `@v3.1.3`, moai-feedback 표 예시값 `v3.1.3`. 역사 인용("v3.1.1부터"류, §1.4 제외 기록)은 변경되지 않는다
- **RED-now**: `grep -n 'version\|releaseDate' docs-site/hugo.toml` → 55행 `v3.1.2`, 56행 `2026-08-21`; `grep -c '🗿 v3\.1\.2' README.ko.md` → `2` (491·766행); `grep -rc 'v3\.1\.2' docs-site/content/ko/advanced/statusline.md docs-site/content/ko/guides/claude-cloud.md` → 각 `1+`; `sed -n '37p;46p' docs-site/content/ko/getting-started/faq.md` → 스테일 예시
- **green-path**: M2+M3 후 (a) `grep -n 'version\|releaseDate' docs-site/hugo.toml` → `v3.1.3`/`2026-08-24`; (b) `grep -c '🗿 v3\.1\.3' README.ko.md README.md README.ja.md README.zh.md` → **각 `2`** (491 statusline + 766 update 예시 — 491행만 고친 뮤턴트는 `1`로 red); (c) `grep -c 'release/v3\.1\.3' README.ko.md README.md README.ja.md README.zh.md` → 각 `1`; (d) `grep -rc 'v3\.1\.2' docs-site/content/ko/advanced/statusline.md docs-site/content/ko/getting-started/faq.md docs-site/content/ko/guides/claude-cloud.md docs-site/content/*/advanced/statusline.md docs-site/content/*/getting-started/faq.md docs-site/content/*/guides/claude-cloud.md` → 전부 `0`; (e) `grep -c '@v3\.1\.3' docs-site/content/ko/guides/claude-cloud.md` → `1`. 역사 인용 보존: `grep -c 'v3\.1\.1부터\|v3\.1\.1에' README.ko.md` → M2 전후 동일 카운트

### AC-DVC-005 — 종료 게이트: hns-oss-docs-verify 8축 (REQ-DVC-007) [D3: version-sync 축 추가·§D.1과 정합]

- **Given** 모든 문서 편집 완료
- **When** 종료 게이트를 실행할 때
- **Then** 8축 전부에서 신규 위반 0 (baseline green 축은 green 유지 — REQ-DVC-007과 동일 계약): (1) warning-free hugo build (2) sitemap.xml 존재 (3) URL 블랙리스트 grep 0히트 (4) Mermaid LR/RL grep 0히트 (5) 4로케일 파일 존재·섹션 수 파리티 — `scripts/docs-i18n-check.sh` exit 0 (6) README 4파일 헤딩 파리티 (7) 본문 emoji 스캔 0히트 (8) version-sync — docs-site content + README의 모든 제품 버전 표시가 hugo.toml `v3.1.3`과 일치 (AC-DVC-004의 (a)–(e) 판정이 이 축의 기계적 내용)
- **RED-now**: 게이트 축은 문서 편집이 낸 격차에 대해 red — 축 5·6은 현재 green (파일 체크섬 파리티 ×4, H2 12 ×4, plan-phase 관측). 축 3·4·7의 baseline 카운트는 §C pre-flight 3에서 측정한다 (기존 부채와 신규 위반 구분)
- **green-path**: M5에서 레시피 실행, 7축 출력 원문을 progress.md §E.2에 기록. 예: `hugo --source docs-site --quiet` rc=0 + 경고 0행; `grep -rc 'docs\.moai-ai\.dev\|adk\.moai\.com\|adk\.moai\.kr' docs-site/content/*/ | grep -v ':0$' | wc -l` → `0`

### AC-DVC-006 — 변경 범위 순結 (REQ-DVC-005)

- **Given** 이 SPEC이 문서 전용임
- **When** run-phase 커밋이 완료되었을 때
- **Then** `internal/`, `pkg/`, `cmd/`, `internal/template/templates/`, 훅 스크립트 경로의 diff가 0파일이다
- **RED-now**: N/A (부정형 관문 — §C.5 음성 점검에서 red 관측: `internal/`·훅 경로에 1행 임시 파일 삽입 시 diff가 비지 않음을 확인 후 철회 [D8])
- **green-path**: `git diff --stat e07a6d0f4 -- internal/ pkg/ cmd/ internal/hook/ .claude/hooks/ internal/template/templates/` → 빈 출력 (훅 경로 포함 [D5]); diff 파일 목록이 spec.md §5의 나열 + progress.md/격차 표로 한정

### AC-DVC-007 — 신규 페이지 승인 관문 (REQ-DVC-003; N: A1–A4)

- **Given** N 판정 4항목이 구조(curator) 소관의 신규 페이지를 요구함
- **When** operator 승인 기록이 없을 때
- **Then** `content/<locale>/_meta.yaml`, `data/menu/main.yaml`, `layouts/partials/menu.html`, 그리고 신규 페이지 파일의 diff는 0이다. 승인·거부 어느 쪽이든 그 결정은 격차 표에 기록된다 (deferred 포함)
- **RED-now**: N/A (부정형 관문 — §C.5 음성 점검에서 red 관측: `_meta.yaml`에 무승인 1행 삽입 시 아래 관측 명령이 비지 않음을 확인 후 철회 [D8])
- **green-path**: M1의 결정 기록 + (승인 시) M4 완료 후 내비게이션 파일·신규 페이지 diff가 승인된 변경만 포함 [D5: 신규 페이지 경로 관측 추가]. 관측: `git diff --stat e07a6d0f4 -- 'docs-site/content/*/_meta.yaml' docs-site/data/menu/main.yaml docs-site/layouts/partials/menu.html 'docs-site/content/*/advanced/codex-dual-harness.md'` — 승인 전 상태면 이 네 경로 모두 빈 출력, 승인 후면 승인된 변경만 표시

### AC-DVC-008 — 전수성: 26항목 무선drop 금지 (REQ-DVC-006)

- **Given** 격차 표의 NA 8항목이 문서화에서 제외됨
- **When** run-phase가 완료되었을 때
- **Then** 26항목 전부가 격차 표에 판정과 근거와 함께 남아 있고 (D/U/N/NA 어느 것이든), NA 행은 근거 문구를 유지한다
- **RED-now**: 충족 (§1에 26행 전부 존재)
- **green-path**: M5에서 유니크 ID 집합 재관측 [D5: drop-and-duplicate 뮤턴트 폐쇄] — `grep -oE '^\| [AFCD][0-9]+' .moai/specs/SPEC-DOCS-V313-CATCHUP-001/spec.md | sort -u | wc -l` → `26` (유니크 기준; 단순 행 수 `grep -cE '^\| [AFCD][0-9]+'` → `26`은 보조 확인) — 유니크 집합이 `{A1..A13, C1..C4, F1..F9}` 전부 존재함을 의미

### AC-DVC-009 — README 구조 파리티 유지 (REQ-DVC-001 귀속 [D8])

- **Given** README 4파일이 H2 12개 동일 구조
- **When** 편집이 완료되었을 때
- **Then** H2 개수·순서가 4파일 동일하게 유지된다 (기존 섹션 내 갱신만 — 섹션 추가/삭제 없음)
- **RED-now**: green (`grep -c '^## ' README.ko.md README.md README.ja.md README.zh.md` → 각 `12`, plan-phase 관측). 이 AC는 회귀 방지용
- **green-path**: M3 후 동일 grep → 각 `12`

## §D.1 Edge Cases

- **M1 승인 지연/거부**: U 11항목 + V1–V8만으로 SPEC이 완결된다 — A1–A4는 "deferred — separate card"로 기록되고 AC-DVC-002·004·005는 영향 없음. AC-DVC-007은 승인 기록 자체로 green. (흡수 대안은 제거됨 [D6] — 운영자가 기존 페이지 흡수를 원하면 이 SPEC을 amend하는 것이 순서.)
- **run 중 병렬 세션이 항목을 선문서화**: §C 재관측에서 발견 시 해당 행을 D로 재판정 — 중복 문서화하지 않고, 이미 반영된 내용이 정확한지만 확인한다.
- **verify 게이트의 기존 부채**: pre-flight baseline(§C.3)에서 이미 존재하는 경고/위반은 이 SPEC 게이트의 실패로 간주하지 않는다 — 다만 baseline 목록에 기록하고 **새 위반 0 + baseline green 축 green 유지**를 판정 기준으로 한다 (REQ-DVC-007의 "zero NEW violations" 계약과 동일 표현 — 기존 부채 면제가 이 SPEC diff가 만드는 위반까지 면제하지는 않는다) [D3].
- **`[Unreleased]` 항목 유입**: 재관측 시점에 `[3.1.3]` 섹션 자체가 재작성돼 있으면(병합 이상) 착수를 중단하고 blocker 보고 — 26항목 전제가 깨진다.
- **ko–derived 콘텐츠 편차 발견** (§B의 settings-json 사례처럼): 이 SPEC 범위 밖 — 격차 기록만 하고 수정하지 않는다.

## §D.2 Quality Gate / Definition of Done

- [ ] AC-DVC-001 … AC-DVC-009 전부 관측된 출력과 함께 progress.md §E.2에 기록됨 (VCI §2 — 명령 + 출력 원문)
- [ ] hns-oss-docs-verify 7축 통과 (AC-DVC-005)
- [ ] 4로케일 + README 4파일이 같은 PR에 존재 (REQ-DVC-001)
- [ ] N 항목의 operator 결정 기록 존재 (AC-DVC-007)
- [ ] Go/템플릿/훅 diff 0파일 (AC-DVC-006)
- [ ] NA 8항목 근거 보존 + 26행 전수성 (AC-DVC-008)
- [ ] 게시(push·PR)는 human gate 통과 후에만 — run-phase 산출은 편집+커밋 대기 상태로 종료
