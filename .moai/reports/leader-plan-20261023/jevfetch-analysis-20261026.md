# jevfetch(a2sys-ai) 분석과 moai-adk 적용안 (2026-10-26 21:1x KST, 리더)

출처: https://github.com/a2sys-ai/jevfetch (MIT, Python 3.12, 1,241줄 코어; 스크래치 clone depth 1, 읽은 파일 jev.py·scoring.py·pipeline.py·blocks.py·README). 벤치 수치(47% 시간 단축 등)는 재현하지 않음(bench/RESULTS.md 미실행).

## jevfetch가 하는 일
URL + 질문 → 페이지를 블록(문단·리스트 항목·표 행·코드)으로 쪼개고 Jev(noul 예/아니오 확률)로 2단계 채점해 **관련 구절만 원문 그대로** 돌려준다. 통과 기준: 1단계 창(≈3,000토큰) 안에서 블록별 ≥0.6 → 2단계 단독+이웃 400자 맥락으로 ≥0.8 유지(같은 제목 이웃은 ≥0.6). 첫 블록 항상 유지, 300토큰 미만이면 점수순 보충, ≤2,000토큰 문서는 통째 통과, 정적 HTML 200토큰 미만이면 헤드리스 렌더 재시도, 표 행은 헤더 행 동반, `<aside>` 노트는 별도 채점. 출력은 `#` 상태 줄 + `##` 제목 경로 + 원문. 상태 filtered/passthrough/no_match/empty/fetch_error, 종료 코드 0/1/2.

## moai-adk 현황 대조
- `internal/jev` Go 클라이언트: 모델 별칭 `jev-latest`·해석 모델 검증·429/529 종류 구분·표시 전용 라벨은 있음. **없는 것**: 요청 본문 해시(모델 포함) 디스크 캐시, 신규/캐시 사용량 분리 집계, Retry-After 헤더 파싱 백오프, 동시성 세마포어.
- 웹 읽기: 내장 WebFetch(요약 모델이 바꿔 씀) 또는 GLM 시 z.ai web_reader. 원문 발췌 도구 없음. MCP 카탈로그에 `jev_ask`만 있음.
- `internal/jevmeasure`: 측정 틀 있음 → 벤치 케이스(jevfetch bench/cases.json 형식: 질문·기대 인용문·페이지 스냅샷) 수용 가능.

## 가져올 것(우선순위)
1. **`moai fetch web <url> "<question>"` + MCP `fetch_web`** — jevfetch 파이프라인의 Go 이식(blocks/scoring/pipeline). 기존 `internal/jev` 클라이언트 사용, `workflow.jev.enabled=false`(기본)이면 WebFetch 폴백(비용 0). 출력 형식·상태·종료 코드 호환. GLM 라우팅 규칙(glm-web-tooling.md)에 "jev 활성 시 fetch_web 우선" 행 추가.
2. **Jev 클라이언트 보강**: 본문 SHA 캐시(`~/.moai/cache/jev/`, 모델 포함 키, 원자 교체), 사용량 신규/캐시 분리, Retry-After 파싱, 세마포어. 별칭 캐시가 모델 갱신 뒤 구답을 돌려줄 수 있다는 경고(README 지적) → 캐시 키에 해석 모델 id 를 쓰거나 pinned model 권고.
3. **벤치 틀**: `bench/cases.json`+`matching.py` 방식(기대 인용문 포함 여부·토큰 절감률)을 jevmeasure 에 케이스 형식으로 추가, baseline.json 비교 명령.
4. **배포 패키징**: moai-adk 템플릿 스킬 `moai-fetch-web`(SKILL.md: 언제 WebFetch 대신 쓰는지, `--json`·no_match 처리), settings 허용 규칙 `Bash(moai fetch web:*)`, 선택적 WebFetch deny 안내, Codex `network_access` 메모.

## 가져오지 않을 것
Python 의존(httpx·bs4·tiktoken)·헤드리스 Chrome 자동 탐지(실행 환경 의존·설치 금지 원칙) — 렌더 폴백은 2단계로 미룸(aside/플랫폼 브라우저 t1965 와 연동 후).

## Gaps
벤치 수치 미재현; Jev 비용(README 추정 $0.0015/호출)은 운영자 키·크레딧 사용이라 기본 off 유지; 토큰 계수(tiktoken cl100k) Go 대체는 근사치.

> 2026-10-26 21:2x KST 개명: 운영자 지시로 `moai web read` → `moai fetch web` (MCP `web_read` → `fetch_web`, 스킬 `moai-web-read` → `moai-fetch-web`, 허용 규칙 `Bash(moai fetch web:*)`). `moai web`은 기존 웹 UI 명령이라 하위 동사 충돌을 피한다.
