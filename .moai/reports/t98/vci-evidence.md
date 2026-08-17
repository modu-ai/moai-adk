# t98 VCI Evidence — docs-site 오케스트레이션 감사 반영 (4로케일)

Card: t98 (bundled behind t32, second card in the same worktree WT-t32)
Branch: WT-t32 @ fb4bb5676 (fast-forwarded to release tip after t32 merge b18f180e2)
Date: 2026-08-18
Session: db221a6c-e73f-4806-b60e-bc00af9ab6fa (run lane)

## 1. Claim (주장)

1. **(a) 오케스트레이션 모드 4종과 한도** — `utility-commands/moai.md`: 4-모드 카탈로그 표
   (direct/serial/fanout/sweep) + 모드별 한도 추가 (fanout 권고 3-5 / 런타임 상한
   `CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS` 기본 20 / sweep=메인 세션 동적 워크플로우·서브에이전트
   사용자 프롬프트 불가). 낡은 "`--team` 은퇴" 서술(v3.0.0 콜아웃 + 플래그 표 2곳)을 현재 사실
   (실험적 재허용·명시적 선택 전용·`MODE_TEAM_UNAVAILABLE`은 문서화된 역사 센티널)로 정정.
2. **(b) 서브에이전트 도구 필터 2단과 함의** — `advanced/agent-guide.md`: 신규 섹션
   "서브에이전트 도구 필터 — 두 단" (1단 스폰 시점 `tools:` CSV 정적 허용 목록 / 2단 런타임 지연
   로딩 ToolSearch `select:`) + 함의 표 (AskUserQuestion 오케스트레이터 전용·런타임 강제 → 서브에이전트는
   blocker 보고서 반환; sweep 서브에이전트는 메인 세션 하에서 프롬프트 불가).
3. **(c) 칸반 -k 진입점 통일 안내** — 편집 불필요 확인: `advanced/kanban-mode.md` 진입 경로
   섹션(~:123-131, :148-179)과 `multi-llm/kanban-mode.md` "진입 스위치" 섹션이 `-k`/`--kanban`,
   cc·glm 양측 배선, `cg` 거부, 리드/동반 진입 명령을 이미 완전 문서화. 저작자·감사 관점 이중 확인.
4. **(d) 카드 클래스 A/B/C** — `advanced/kanban-mode.md`: 신규 섹션 "카드 클래스 — A/B/C"
   (A 직접 마감 / B 결함·원인 미확립 / C 설계 변경; A 증거 요건 "diff 측정+CI 초록, 속도는 결과이지
   이유가 아님"; B는 plan만 건너뛰고 sync 게이트 리뷰는 유지·원인 증거 기록; WIP 규칙).
   SSOT: 현행 `.claude/rules/moai/workflow/kanban-dispatch.md` (t104 개정판). t115의 5열 보드·
   세 동반 서술과 일치 (재수정 없음, 이중 확인).
5. **4로케일 동시 반영** — ko 정본 저작 → en/ja/zh 표적 미러 (순차). 모든 로케일에서 헤딩 변화량이
   ko와 동일 (moai.md ±0, agent-guide.md +1, kanban-mode.md +1).
6. verify 레시피 전 항목 PASS + 신규 로케일 발산 0 (래칫 유지 55).

## 2. Evidence (증거)

### 헤딩 수 (게이트 메트릭 `grep -c '^#\{2,\}'`) — before → after

| 페이지 | ko | en | ja | zh |
|---|---|---|---|---|
| utility-commands/moai.md | 24→24 | 24→24 | 23→23 | 23→23 |
| advanced/agent-guide.md | 21→22 | 29→30 | 29→30 | 29→30 |
| advanced/kanban-mode.md | 19→20 | 19→20 | 19→20 | 19→20 |

모든 after 값은 오케스트레이터가 12개 파일 전건 재측정 (로케일별 번역자 보고와 무관하게 확정).

### 전체 트리 래칫 (레시피 §4 awk 파이프라인)

발산 집합 크기 = **55** (t32에서 63→55로 조인 baseline과 동일) — t98 변경 후에도 신규 발산 0.
kanban-mode.md는 4로케일 20/20/20/20 수렴 유지 (baseline 미등재 상태 그대로).

### 정적 검증 (모두 exit 1 = 무매치)

- `hugo --source docs-site --minify --gc 2>&1 | grep -E "WARN|ERROR"` → 무매치; `sitemap OK`
- URL 블랙리스트 (content + README 4종) → 무매치
- `flowchart LR|RL` → 무매치 (신규 다이어그램 0 — 표로만 구성)
- 카드 추가 줄 이모지 스캔 (`git diff | rg "^\+…" | rg -v -F '{{<'`) → 무매치

### 변경 센서스

`git status --porcelain -- docs-site/` → 정확히 12 ` M` (3페이지 × ko/en/ja/zh).
외부 파일 0. (en moai.md의 `--solo` 상단 표 1줄은 오케스트레이터가 직접 동기화 — ko:80과
일치, ja:71 미러와 동일 형태; tr-en98이 누락을 자체 플래그했던 항목)

### 하네스 실행 기록

- 저작: hns-oss-docs-content-author-specialist 1회 (ko 정본 3페이지, +51/−11)
- 미러: hns-oss-docs-locale-translator-specialist 3회 순차 (en→ja→zh, 각 3파일 표적 동기화)
- 각 에이전트 자체 출구게이트 통과 (hugo quiet 빌드 exit 0, 로케일별 렌더 index.html 생성·신규
  섹션 grep 확인 보고 포함)

## 3. Baseline-attribution (baseline 귀속)

모든 측정은 본 세션에서 본 트리(WT-t32, HEAD fb4bb5676) 대상으로 직접 실행. before 값은
저작 위임 전 측정, after 값은 zh 미러 완료 후 재측정. hugo = /opt/homebrew/bin/hugo.

## 4. Gaps (미검증)

- 번역 품질의 문장 단위 대조 심사는 번역 전문가 보고 + 오케스트레이터 구조·식별자 검증
  (CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS/MODE_TEAM_UNAVAILABLE/`select:`/A/B/C 등 verbatim
  존재 grep 확인)로 갈음 — 12파일 전문 대역심사 미실시.
- ja/zh moai.md의 기존 발산(23 vs ko 24, 누락 플래그 행 등)은 baseline 등재 상태 그대로
  유지 — 표적 동기화 원칙에 따라 재파생하지 않음 (B5류 후속 카드 소관).
- (c)는 편집 0이므로 4로케일 변경 없음 — 이중 확인(ko 두 페이지 교차 검증)으로 충분하다고 판단.

## 5. Residual-risk (잔여 위험)

- `--team` 정정은 CLAUDE.md §15·agent-guide.md§재허용 섹션·manager-kanban.md와의 4소스 일치를
  확인했으나, Agent Teams의 세션 관찰 증거 자체(재허용 이후 실제 동작)는 이 카드에서 재실측하지 않음.
- moai.md 자동 선택 기준의 숫자(도메인 3개/파일 10개/점수 7)는 기존 문서 값 유지 — 본 카드에서
  재검증하지 않음 (감사 항목 범위 밖).
- 정본 결함 후보 2건(t32에서 플래그: 에이전트 수 12 vs 11, CG 절감 50~70% vs 60-70%)은 이번
  작업 페이지와 무관 — 리드 지시대로 백로그 후보로 남김 (t98 페이지를 건드리는 과정에서 해당
  문장을 만나지 않아 정정하지 않음).
