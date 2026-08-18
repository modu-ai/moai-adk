# t47 — en 구골격 → ko 정본 흡수 대조표 (손실-0 검증)

기준: en 구골격 README.md (release/v3.1.1 @ ca8c0b593, 14 H2, 560줄) → ko 신골격 정본 (12 H2, 756줄).
각 en 섹션의 처지와 정보 귀속 위치. 재파생 en/ja/zh는 ko 정본에서 파생되므로 이 표가 4-로케일 공통 근거다.

| en 구골격 섹션/요소 | 처지 | ko 흡수 위치 |
|---|---|---|
| 에피그래프 (stochastic worker 인용) | 흡수 | H1 직후 인용구 블록 (번역) |
| H2 타이틀 "A Three-Axis Agentic Harness" + 하네스 정의 문단 | 흡수 | "왜 moai-adk인가요?" 도입 2문단 + why-harness-infographic-ko.png + 세 가지 핵심 (three axes) 프레임 문단 |
| New in v3.1 — Kanban Mode | 대응 (ko가 상위집합) | "v3.1 새 기능 — 칸반 모드" — ko는 용어집("다섯 세션이 쓰는 말", t59)·moai web 소절 추가 보유 |
| Why Three Axes (상호지지 논증) | 흡수 | H3 "세 핵심이 서로를 지탱한다" + three-axes-infographic-ko.png |
| 🪙 The Cost Axis — Tokenomics (98%/320%, Uber, Meta/Amazon/MS) | 흡수 | H3 "비용은 단가가 아니라 배정이 결정한다" + why-tokenomics-infographic-ko.png |
| Cost — DeepSWE 표 (요약 5행 + 상세 7행) | 흡수 (단일화) | 상세 7행 표 + deepswe-benchmark-2.png. 요약표는 상세표의 부분집합(opus-5 low/med/high/max + sonnet-5 max) → 정보 손실 0 |
| Cost — Routing (No-Haiku, Profile Matrix, CG) | 흡수 | No-Haiku·배정 원칙 → "설정과 프로파일 > 모델 프로파일" + model-routing-infographic-ko.png. CG → "핵심 기능 > CG 모드" + cg-mode-infographic-ko.png |
| Cost — Four stages (measurement→routing→diet→defense) | 흡수 | why-tokenomics 인포그래픽 alt 텍스트에 4단계 명시 + 다이어트/중단은 아래 검증경제 항목으로 |
| Cost — Verification Economy · Budget Defense | 흡수 | H3 "검증 비용을 줄이고, 예산 초과 전에 멈춘다" (디스크 리다이렉트·50줄 꼬리·캐시 0.1×·/clear 임계·토큰 회로 차단기 90%·statusline) — calque 금지 준수로 제목은 자연어 |
| 🧠 The Self-Improvement Axis | 흡수 | 차별점 표 "자가 개선" 행 + "goal 엔진" 절 (+ en의 --max-turns 0 무한 골/--max-duration/정체 가드 파라미터 추가 반영) |
| 🛡️ The Quality-Control Axis — SPEC/TRUST/12-Agent | 흡수 | "어떻게 돌아가나요?" — SPEC 3-페이즈(기존) + 12-에이전트 카탈로그 표(비용 색 포함) + trust-but-verify(기존) |
| Infrastructure Sustains All Three Axes | 중복 폐기 | "핵심 기능 > 크로스 플랫폼"이 동일 정보 (Go 단일 바이너리·무의존·훅 강제·statusline 실시간) 보유 |
| Quick Start | 대응 | "빠르게 시작" (ko 상위집합: z.ai 후원 노트 보유) |
| Reference — /moai 16 서브커맨드 표 | 부분 흡수 | "단일 진입점 /moai"가 16개 이름 전부 나열 + 각 기능의 역할은 "핵심 기능" 각 절이 상세 설명(표의 역할 컬럼 대응). **은퇴 4커맨드(design·brain·coverage·security) 각주 신규 흡수** |
| Reference — CLI 13 표 | 대응+정정 | "CLI 명령표" — worktree 행을 en 최신 8동사(sync·done·remove·clean·recover·snapshot·verify·restore)로 정정 + "(워크트리 진입은 런처의 몫)" 주석 흡수 |
| Reference — MCP Server | 대응 | "핵심 기능 > MCP 서버" (동일 내용, 5그룹 17도구 표 포함) |
| Reference — 12-Agent Catalog 표 | 흡수 | "어떻게 돌아가나요? > 12-에이전트 카탈로그" (비용 색 🔴🟠🔵🩵⚪ + 프로파일 색 해설 포함) |
| Reference — TRUST 5 상세 표 (T/R/U/S/T 검증방법 컬럼) | 요약 수준 흡수 | "자동 품질 게이트"(다섯 알파벳 전개) + "코드 품질 요구사항"(85%·린트0·타입0·Conventional). 표의 항목별 검증 컬럼은 미반영 — Gaps 참조 |
| Reference — Methodology (TDD/DDD mermaid+표) | 흡수 | "SPEC 3-페이즈 라이프사이클"에 mermaid + 2행 표 추가 |
| Reference — Kanban Mode (Origin-Trail) | 대응 | "핵심 기능 > 칸반 모드" (동일: 개념 표·실행 각주 포함) |
| Reading the Statusline (예시+요소 표) | 흡수 | H3 "스테이터스라인 읽기" (렌더링 예시 코드블록 verbatim + 12행 요소 표 + /ko/advanced/statusline 링크) |
| Claude × GLM Multi-LLM ($10·모델 목록·3모드 표·매핑 표) | 흡수 | "Claude + GLM" — ANTHROPIC_DEFAULT_*_MODEL 매핑 표(Opus/Sonnet/Haiku/Fable→glm-5.3 1M)·월 $10·glm-5.3/4.7/4.5-air·무료 모델 추가 |
| FAQ (4 Q) | 흡수 | 신규 H2 "자주 묻는 질문" (골격 11→12의 원인) |
| Community and Documentation | 대응 | "문서와 학습"(12섹션 표·도서·CHANGELOG·Claude Code 공식 링크) + "함께 만들어요"(기여·피드백·Discord) |
| Star History | 대응 | "스타 히스토리" (동일) |
| en 헤더의 book 이중 블록 | 중복 폐기 | ko 헤더의 공식 문서·도서·Discord 블록이 동일 링크 전부 보유 |

## 사실 정정 (흡수 과정에서 발견·측정)

1. **프로파일 매트릭스 셀 수**: en 구골격 "12 agents × 3 profiles = 36 cells" → 실측 `moai model profile --json` = **11 에이전트 = 33셀** (관측: grep -c '"agent":' → 11). 12는 Explore 포함 카탈로그 수이며 매트릭스 행이 아님. ko 정본·재파생 전 로케일 33셀로 통일.
2. **ko CLI 표 worktree 행**: 구 5동사 → en 최신 8동사로 정정 (snapshot·verify·restore 추가).

## ja Cost 절 분기 해소

구 ja는 Cost H2 아래 H3 4개(単価/ルーティング/検証経済/予算防御)로 en/zh(3개)와 어긋났다. ko 골격은 Cost H2 자체가 없고 토크노믹스가 "왜 moai-adk인가요?"의 H3 2개(세 핵심 상호지지·비용은 배정)로 흡수 — 재파생 4파일이 동일 구조를 가지므로 분기는 골격 소멸로 해소. (재파생 완료 후 H3 실측으로 확인 — 검증 리포트 참조)

## Gaps (손실-0 주장의 경계)

- TRUST 5 표의 항목별 검증방법 컬럼(예: Secured=OWASP·입력검증·보안경고 0)은 ko가 요약 수준으로만 보유. 핵심 수치(85%·린트0·타입0)와 다섯 축 전개는 보존됐으나 항목별 상세 검증 기술은 미흡. 필요 시 후속 카드로 표 확장 권장.
- en 구골격의 "Why Three Axes" 절 제목 표현 자체는 ko에서 "세 가지 핵심" 프레임으로 재조직됨 (calque 금지 준수). 정보는 보존, 표현 상이.
