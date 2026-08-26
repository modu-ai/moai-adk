# t273 격차 맵 — plan>run>sync + 칸반/팩토리 운영 모델 문서화

작성: 2026-08-25 · 레인: t273 (WT-workflow-docs, base db1362739) · 카드 Class C, Tier M

## 조사 방법

- 정본 3개 헤딩 스캔 + 핵심 섹션 정독 (spec-workflow.md §Phase Overview~Phase Transitions, kanban-dispatch.md 전문 — 세션 리마인더 로드, worktree-integration.md 헤딩)
- docs-site 조사: 4로케일 구조·페이지 수 파리티·관련 키워드 커버리지 (docs-site-survey 서브에이전트 보고 + 직접 grep 교차확인)
- README 조사: 4파일 헤딩 파리티(H2 12개 동일) + ko 정본 관련 섹션 정독
- 갭 키워드 실측: `grep -rn "Class A|클래스 A|카드 클래스|card class" docs-site/content/ko/ README.ko.md` → **0건**

## 카드 요구사항 대비 커버리지 매트릭스

| 카드 요구 | 기존 커버 | 갭 판정 |
|---|---|---|
| (1) plan>run>sync 각 단계 목적·입출력 산출물·관문(킥오프·plan-audit·sync-audit) | workflow-commands/{moai-plan,moai-run,moai-sync}.md (명령별 상세, plan-auditor Phase 11 포함) · core-concepts/spec-based-dev.md (SPEC 문서 구조 중심) · README 3-페이즈 절 (mermaid·Tier 언급) | **부분** — 관문 3종을 라이프사이클 하나의 흐름으로 통합 서술하는 페이지 없음. Tier S/M/L 산출물 세트(2/3/5파일)·REQ/AC 상한(8/16/25) 표도 문서화 안 됨 |
| (2) 칸반 모드: 세션 체인·5컬럼·카드 클래스 A/B/C | advanced/kanban-mode.md (28KB, 세션 체인·진입·증거 기반 완료 :188,:304) · multi-llm/kanban-mode.md (운영 절차) · core-concepts/kanban-board-terms.md (용어 9종) · README v3.1 섹션 | **대부분 커버, 카드 클래스 A/B/C만 전면 부재** (grep 0건). 정본 kanban-dispatch.md §Card classes에 있는 A(직접 종결)·B(결함 원인 불명, plan 스킵)·C(설계 변경, 3컬럼 전부) 미문서화 |
| (3) 팩토리 모드: 구조·통째 라우팅·레인 내 3단계·병렬 상한 | advanced/kanban-mode.md 내부 섹션 (§팩토리 모드 — 레인 N개) · README §팩토리 모드 (진입·최대 10 동시 에이전트·workers.json) | **내용은 있으나 전용 페이지 부재** — 발견가능성 문제. 카드의 "상세한 문서" 요구와 `-f` 진입 토큰의 1급성을 고려하면 분리가 자연스러움 |
| (4) 사용 예시(진입 명령·세션 구성) | cli-reference/launchers.md (-k/-f 스위치 표) · README·kanban-mode.md 예시 블록 | **커버** |

## 확정 갭 (문서화 범위)

- **GAP-1 — 칸반 카드 클래스 A/B/C**: 정본 kanban-dispatch.md §Card classes·§Class A admitted on checked evidence·§Class B skips plan not the sync gate's review. A=한 파일·설계 판단 없음·plan 스킵 / B=결함 원인 불명·plan 스킵(원인 확립 증거는 run에서 progress.md에) / C=설계 결정·3작업 컬럼 전부. 배치: advanced/kanban-mode.md 신규 섹션 + README v3.1 칸반 절 소표.
- **GAP-2 — 팩토리 모드 전용 페이지**: kanban-mode.md의 팩토리 섹션을 advanced/factory-mode.md로 분리·확장. 내용: lead+lane-1..N 구조, 카드 통째 라우팅(칸반=칸 이동 vs 팩토리=레인이 3단계 관통), 레인 내 plan>run>sync(단계마다 서브에이전트), 레인당 동시 스폰 상한 10, 순차 기동, workers.json 점유. kanban-mode.md에는 요약+링크만 남김.
- **GAP-3 — SPEC 라이프사이클 통합 페이지**: core-concepts 또는 guides에 "SPEC 라이프사이클" 통합 페이지. 내용: 3페이즈 표(명령·주체 에이전트·목적), 각 단계 입출력 산출물, 관문 3종(Implementation Kickoff Approval — plan→run 인간 관문·score-independent, plan-audit — 독립 감사 PASS 임계 0.75/0.80/0.85, sync-audit — 4차원 품질), Tier S/M/L 산출물 세트(2/3/5파일)·REQ/AC 상한(8/16/25) 표, Route A/B 요약, /clear 전략. 기존 spec-based-dev.md와 상호링크(중복 최소화 — spec-based-dev는 "SPEC 문서가 무엇인가", 신규 페이지는 "라이프사이클이 어떻게 흐르는가").
- **GAP-4 — README 보강**: 카드 클래스 A/B/C 소표를 README v3.1 칸반 절에 추가(4로케일). 3페이즈·칸반·팩토리는 기존 서술 충실 — 문서 링크 정도만 점검.

## 명시적 비범위 (Out of Scope)

- 기존 페이지 전면 재집필 (t87 드롭 사유 준수 — 기존 4로케일 파리티 524페이지 유지)
- Origin-Trail Chain 내부 구조 (JSONL·WorktreeNode 13필드·two-phase backfill) — kanban-mode.md 기존 심층 섹션 유지, 본 카드에서 재서술 안 함
- 세션 핸드오프·컨텍스트 창 관리 문서화 (독립 주제 — 별도 카드감)
- multi-llm/kanban-mode.md 재구성 (운영 절차 문서로 역할 분담 유지)

## 네비게이션 변경 (리드 보고 대상)

신규 페이지 2개(factory-mode.md, SPEC 라이프사이클 통합)는 _meta.yaml 4로케일 + data/menu/main.yaml 4로케일 항목 추가 수반 → structure-curator 소관. 배치 확정 후 리드에 보고하고 진행.

- advanced/factory-mode.md: advanced 섹션 _meta.yaml에 kanban-mode 인접 weight
- SPEC 라이프사이클 통합: core-concepts 섹션 (spec-based-dev 인접) — 명칭은 plan 단계에서 확정 (후보: spec-lifecycle.md)

## 검증 레시피 (run 종료 게이트)

hns-oss-docs-verify 스킬: warning-free hugo build · sitemap 존재 · URL 블랙리스트 grep · Mermaid TD-only grep · 4로케일 파일 존재+섹션 수 파리티 · README 4파일 헤딩 파리티 · 본문 이모지 스캔. 추가: 카드 클래스 A/B/C가 4로케일 전부에 존재하는 grep(각 로케일 표현), 신규 페이지 2개×4로케일=8파일 존재, _meta.yaml·main.yaml 4로케일 동기.
