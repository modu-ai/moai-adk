# t272 — SVG 인포그래픽 스킬 9종 형태 실측 벤치마크 + 생성 가능한 다이어그램 종류 문서화

SPEC-SKILL-GALLERY-BENCH-001 v0.1.2 · 카드 t272 · 브랜치 `WT-skillstead-gallery`

## 무엇을 측정했나

`moai-domain-svg-infographic` 스킬이 실제로 어느 종류의 다이어그램을 만들 수 있는지, 주장 대신 실측으로 답했습니다. 제3자 카탈로그인 SkillStead TypePack(Apache-2.0, `svg-infographic` v0.11.0)의 9가지 형태를 동일한 생성 과제로 실행하고, 스킬 자체의 품질 게이트 — 결정론적 소스 린트(`check-svg.mjs`)와 치수 검증 2배 PNG 렌더(`render.mjs`) — 를 형태별로 통과했는지 측정했습니다. 판정 기준은 연산자 승인된 정보구조 동등성(형태의 정보 구조가 보존되면 동등 형태로 인정, 편차는 명시)입니다.

## 결과 — 9/9 PRODUCIBLE

9가지 형태 모두 재현 가능했습니다: approval-gate, before-after, cards-kpi-grid, decision-matrix, layer-stack, nested-scope, process-flow, roadmap-timeline, topology-component. 게이트 18건(린트 9 + 렌더 9) 전부 exit=0으로 관측했고, 렌더 로그마다 브라우저 실행 파일과 버전(Chrome 151.0.7922.174)을 공개했습니다. 형태별 편차 문장과 다이얼 편차(approval-gate·process-flow size fit, decision-matrix faithful-banded)는 판정 표에 기록했습니다.

## 문서 변경 (8파일, 순수 추가)

측정 근거 위에 "생성 가능한 다이어그램 종류" 안내를 추가했습니다. docs-site 4개 로케일의 `advanced/skill-guide.md` Domain 섹션에 H3 섹션(실측 설명 + 9행 표, 각 행이 실측 산출물 경로를 인용)을, README 4개 로케일 핵심 기능 섹션에 H3 항목(9가지 형태 명명 + 판정표 포인터)을 각각 추가했습니다. ko 정본 저작 후 en/ja/zh 최소 분파, 신규 H2 없음, 이모지·버전 표시·블랙리스트 URL·새 Mermaid 없음.

## 증거

- 판정 표(9행, 전체 스키마): `.moai/reports/t272/verdict.md`
- 레이아웃 수치 패스: `.moai/reports/t272/layout-notes.md`
- 산출물: `.moai/reports/t272/artifacts/` (SVG 9 + PNG 9)
- 게이트 로그: `.moai/reports/t272/logs/` (18건, 명령줄 + 출력 + exit=N 계약 형식)
- verify 레시피 결과: `.moai/reports/t272/verify.md`

## 검증

`hns-oss-docs-verify` 레시피 전 항목 실행: 경고 없는 hugo 빌드 + sitemap, URL 블랙리스트 0건, Mermaid LR/RL 0건, 4-로케일 파일 존재(150/150/150/150) 및 섹션 수 래칫(신규 분기 0), README H2 패리티 12/12/12/12, 본문 이모지 스캔(추가 라인 0건). 버전 동기화는 원본 main에 이미 존재하던 배지(v3.1.3)/SSOT(v3.1.2) 분기를 귀속 조사와 함께 기록했습니다 — 이 브랜치는 버전 표시를 하나도 추가하지 않았습니다. 스킬 디렉터리와 템플릿 미러는 diff 없음(REQ-SGB-008).

SPEC: SPEC-SKILL-GALLERY-BENCH-001 v0.1.2 · 카드: t272 · 브랜치: WT-skillstead-gallery
