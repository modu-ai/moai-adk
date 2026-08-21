# diagram-design 흡수 분석 — 요약 (markdown twin)

- 대상: github.com/cathrynlavery/diagram-design v2.6.1 (MIT, LICENSE 정본 존재 — 라이선스 차단 없음). 412 파일: SKILL.md 39,453B / references 52파일 660KB(스킨 SSOT 1·프리미티브 4·시맨틱 패턴 7·워크플로 8·유형 사양 39) / 검증기 15 + 린터 2 + 검증기시험 16(263KB) / ADR 8 / 예제 136+템플릿 5 / 4호스트 매니페스트 6.
- 대조: moai-domain-svg-infographic v1.0.0 (Apache-2.0). SKILL.md 14KB + references 3(24KB) + 스크립트 2(check-svg.mjs, render.mjs). 아키타입 4 + mermaid 라우팅, CJK-first 60% 용량 규칙, 열람시점 무외부자산 계약.
- 데모 3종 공통 특성(지관 배경·코랄 포인트·세리프 제목·모노 서브라벨·직교 커넥터) = diagram-design 시맨틱 롤 스킨 토큰의 산출물 — 재현 열쇠는 카탈로그가 아니라 스킨 SSOT+타이포 역할 분리+강조 1–2 원칙.

## 판정
- **즉시 흡수 6 (A)**: A-1 커넥터 6강제규칙(직교 엘보 r=8·마스크 6–10px 갭·비중복+bridge/hop·부착점 팬 L·k/(N+1) ≥12px·비엔드포인트 뒤 통과 금지·마스크-후속노드 비중첩) → authoring.md. A-2 유형별 복잡도 예산표 → SKILL.md/archetypes.md. A-3 접근성 SVG 계약(role/aria-labelledby·prefixed ID·title 첫 자식·desc 내용 서술) — moai 완전 부재. A-4 "AI slop" 안티패턴 14종 → Red Flags 보강. A-5 출력 다이얼 4개(format/size/detail/audience) → Frame 절. A-6 시맨틱 롤 스킨 토큰+라이트/다크 반전 규칙 → authoring.md 팔레트 재구조.
- **참조 흡수 4 (B)**: B-7 check-svg.mjs 기하 검사 3종 확장(페인트순서 라벨-클립·부착점 간격·마스크-스트로크 갭)+양극 self-test — 최대 갭. B-8 marker-first 스킨 프로파일 → design-dna 보강. B-9 drawio/mermaid IR 임포터+충실도 원장 — 옵트인 참조. B-10 아이콘 라이선스 정규화 기법(24×24 currentColor) — 필요 시, THIRD_PARTY 고지 동반.
- **거부 4 (C)**: 39유형 카탈로그 전체(mermaid 라우팅·"한 다이어그램 한 집" 충돌), Google Fonts CDN(무외부자산 계약 위반), HTML 3변형/모션/터미널 스킨(SVG+PNG 정적 계약 유지), 플러그인 매니페스트 배포(Template-First가 정본).
- **메타 흡수 (D)**: ADR 관례, docs-sync 드리프트 게이트(description 어휘 훅 보존 검사), 검증기-시험-검증기 양극성, 스크린샷 신선도 매니페스트 → skill-authoring craft 반영 후보.

## 실행 제안 (카드 3장)
1. 품질 계층 흡수(A-1~A-6) — Tier S. 2. 검증기 기하 확장(B-7) — Tier S~M, ① 다음(규칙 선행). 3. 프로파일·임포터(B-8/9) — Tier M 후순.
- 라이선스: MIT→Apache-2.0 호환, 출처 고지 의무. 배포: Template-First 미러+make build.

## 경로
- HTML 보고서: .moai/reports/diagram-design-absorption/diagram-design-absorption-20260822.html
- 전수 서베이: .moai/reports/diagram-design-absorption/survey.md
- 원본 클론: /tmp/diagram-design
