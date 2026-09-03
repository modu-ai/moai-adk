# t440 판정서 — docs-site 4로케일 delivery 동작 고지

카드: t440 · 브랜치 `WT-delivery-notice-docs` · 측정 트리 `.claude/worktrees/t440` · 측정일 2026-09-02.

## 판정

lane-4가 t315에서 Gap(G2)으로 남긴 축에 대한 판정: **docs-site에 실는다 — 실었다.**
릴리스노트 별도 문안(t315 §5)과 별개로, 사용자가 delivery 동작을 배우는 페이지에 동작
고지가 없어야 docs가 거짓말을 하는 상태였다.

## Claim (산출물)

`docs-site/content/{ko,en,ja,zh}/workflow-commands/moai-sync.md` — 신규 절 1개 × 4로케일
(ko: `### 브랜치별 전달 동작 (git strategy)`), tier 절 포인터 문장 ×3 (zh는 tier 절이 없어
제외), FAQ 보강 문장 ×4. 내용:

- 전략 키 `git_strategy.{mode}.workflow` (값 도메인 `{github-flow, git-flow}`, 3모드 기본
  `github-flow`, 미매칭 값 = 보고 후 중단)
- github-flow 브랜치 라우팅 3축: main+Tier S/M+--pr 없음 → 변화 없음 / main 아닌 브랜치 →
  PR 생성 (v3.1.2까지는 브랜치 무관 main push·PR 없음) / 어느 경로에도 안 맞는 상태 →
  중단·보고 (종전엔 무조건 push로 완주)
- `main_branch` 키와 표의 "main" 관계, git-flow 브랜치 라우팅 한 절, 구 키
  `github.spec_git_workflow` 폐기(v3.3.0) 안내

## Evidence

| 검증 | 명령 | 관측 |
|---|---|---|
| 섹션 수 패리티 | `grep -rc '^\#\{2,\} ' ko en ja zh --include='*.md'` (moai-sync) | ko 46 · en 46 · ja 46 · zh 46 — 45→46 동시 증가 |
| 랫chet 신규 이격 | verify recipe 전체 트리 grep+awk → `comm -23 now base` | **0건** (54페이지 기존 베이스라인 불변) |
| 랫chet 수렴 | `comm -13 now base` | 0건 (베이스라인 정리 대상 없음) |
| hugo build | `hugo --minify --gc --source <wt>/docs-site` | **rc=0, WARN/ERROR 0건**, Total 2926 ms (hugo-build.log) |
| sitemap | `test -f docs-site/public/sitemap.xml` | OK |
| URL 블랙리스트 | `grep -rn 'docs\.moai-ai\.dev\|adk\.moai\.com\|adk\.moai\.kr' content` | rc=1 (0매치) |
| Mermaid 방향 | `grep -rn 'flowchart LR\|graph LR\|flowchart RL\|graph RL'` | rc=1 (0매치 — 신규 다이어그램 없음) |
| 본문 이모지 | `grep -rnP '[\x{1F300}-…]'` (4편집 파일) | rc=1 (0매치) |

## Baseline-attribution

- 페이지 실측: 편집 전 4로케일 moai-sync.md 각 45섹션, 베이스라인(`.locale-parity-baseline`,
  54페이지)에 moai-sync.md 미등재 = 편집 전 패리티 상태. 측정 트리: 본 워크트리 `e45054c56`
  (로컬 develop 흡수 후).
- 동작 근거: lane-4 실측 리포 `.claude/worktrees/t315/.moai/reports/t315/d7a-behavior-delta.md`
  §2 표 + 본 트리 `internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md`
  Step 3.0/3.2 원문 직독 (github-flow 라우팅·미매칭 중단·폴백 조건).
- 문서 톤 통제: 릴리스노트 문안 복사 없음 — 배차 지시의 대표 mutant 축. "무엇이 바뀌었나"
  가 아니라 "지금 어떻게 동작하나 + 어디서 확인하나" 서술로 재작성.

## Gaps

- **G1** 실사용 프로젝트에서 `/moai sync`를 돌려 문서 서술과 실제 동작의 end-to-end 대조는
  안 했다. 서술은 배포되는 delivery.md 원문 + lane-4 표 이중 대조에 근거한다.
- **G2** zh 페이지의 선존재 구조 이격(tier 절 부재·`--merge` 절 존재·Phase 번호 체계 상이,
  ko/en/ja와 다름)은 본 카드에서 수리하지 않았다 — 섹션 수는 46으로 동일해 랫chet에는 무해.
  content 정합성 부채로 별도 카드 소관.
- **G3** en/zh FAQ 답변이 ko/ja와 다른 답(auto_pr 설정 안내 vs Hybrid Trunk 서술)을 담고 있는
  것도 선존재 이격 — 본 카드는 각 로케일 기존 답변 뒤에 동일한 사실 1문장만 보강했다.
- **G4** t440 카드 자체의 develop 병합·창 절차는 이 판정서 작성 시점에 미수행 (창 지명 후 별도).

## Residual-risk

- R1: 페이지의 tier 절이 "(Hybrid Trunk 1-person OSS 기본 동작)" 프레임을 유지하는데,
  `CLAUDE.local.md` §18 기준 Hybrid Trunk는 RETIRED다. 본 카드 범위 밖이라 손대지 않았고,
  별도 문서 정리 카드의 후보로 리드에 보고한다.
- R2: docs-server 렌더(HTML) 단에서의 표 렌더링은 hugo 빌드 성공으로만 검증됐다 — 브라우저
  육안 대조는 안 했다 (빌드 warning-free가 배차 검증 기준).
- R3: v3.1.4 릴리스(PR #1685)가 머지되면 본문의 "v3.1.2까지는" 서술은 시점상 과거형이지만
  정확성은 유지된다 (역사 인용 — version-sync 게이트의 비대상).
