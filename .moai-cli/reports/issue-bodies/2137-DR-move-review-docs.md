# [DR][01] 설계 검토 산출물을 tracked 위치로 이동 (`docs/design-review/`)

## 배경
2026-11-09 사용자 컨펌된 종합 보고서와 검토 파일 7개가 `references/chatgpt-space/`(untracked)에만 있어 레인·Codex 에이전트가 읽을 수 없다. 설계 문서 개선 카드(DR-02~07)의 공통 입력이므로 먼저 tracked로 옮긴다. (종합 보고서 §10.2 F-D 선행 카드)

## 범위
- 복사(untracked → tracked 신규 추가; `references/chatgpt-space/` 원본은 그대로 둔다 — r1 본문의 "원본 유지 금지"는 오기, 리더 정정 2026-11-09 00:1x): `synthesis-improvement-report-2026-11-09.md`, `review-page_5490c4ee-2026-11-08.md`, `gap-analysis-rc1-vs-design-2026-11-08.md`, `review-group1-rc1-plan.md`, `review-group2-core-brain-sync-owner.md`, `review-group3-install-host-plan-run-release.md`, `review-group4-past-designs.md`, `pages.tsv` → `docs/design-review/`
- `docs/design-review/README.md`: 파일 목록·각 파일의 대조 기준 SHA·Space 페이지 id↔제목 표(pages.tsv 기반)·"page_*.md 원문 사본은 저장소에 두지 않음(ChatGPT Space 원본 참조)" 명시
- **제외**: `page_*.md` 19+1편(ChatGPT 원문 사본, 31k행) · `design-review-synthesis.html`
- 파일 내용은 수정하지 않는다(경로 참조 `references/chatgpt-space/` → `docs/design-review/` 치환만 허용, 치환 건수를 PR 본문에 기록)

## 수용 기준
- AC-1: `git ls-tree origin/v2026.11.0 docs/design-review/` 에 README + 8개 파일 존재, 각 sha256이 원본과 동일(치환한 파일은 치환 전후 diff가 경로 문자열뿐)
- AC-2: 개인 절대 경로(`/Users/...`) 0건 — `grep -rn "/Users/" docs/design-review/` 0행 (있으면 상대 경로로 바꾸고 건수 기록)
- AC-3: 비밀·토큰 문자열 0건(`src/redact.rs` SECRET_SHAPES 기준 grep)
- AC-4: check set 통과(문서만이라 cargo test 영향 0 — 그래도 실측 exit 기록)

## 사용자 결정
- 2026-11-09 00:1x: 보고서 컨펌, Issue 발행·배차 시작(리더 중계)

## 증거
PR 본문에 파일별 sha256 전후, grep 결과, check set exit.
