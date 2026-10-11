# decision-index — SPEC-SESSION-START-GUIDE-I18N-001

decision_gate: on (.moai/config/sections/interview.yaml:6)
authored: 2026-10-08 (card t1603 plan-phase)

### Q1: 부트스트랩 안내의 agent-facing additionalContext 채널은 어느 로케일 키를 따르는가

Label: FOUNDER
Class: product-level
Authority anchor: 해당 없음 — `.moai/config/sections/language.yaml`의 `agent_prompt_language`(주석 "Always en")와 `internal/hook/session_start.go:503`의 두-청중 서술은 반대 방향(영어 고정)을 가리키고, 카드 지시(conversation_language별 렌더, spawn authority 문구 포함)는 커밋된 권위 장부에 없다. 서술과 코드도 이미 어긋난다(:503 서술 vs :518 langEnglish 하드코딩).
Why unresolved: (해소됨 — 아래 판정) 카드 지시는 conversation_language별 다국어 렌더를 요구하고, language.yaml의 정책 키(agent_prompt_language: en)와 훅의 기존 two-audience 규칙(lane_spawn_authority.go:38-40)은 agent-facing 채널의 영어 고정을 규정해 두 출처가 같은 질문에 다른 답을 줬다.
Operator verdict: RESOLVED (레인 카드-우위 판정, 2026-10-08) — agent-facing additionalContext 채널(session_start.go:518)은 conversation_language를 따른다. 근거: 카드 본문(운영자 지시 2026-10-08)이 리더·레인 SessionStart 안내 전체를 conversation_language별 렌더로 명시하고 standing spawn authority 문구를 다국어 대상에 이름으로 포함한다 — 운영자 지시가 우위다. 기존 two-audience 규칙과 `agent_prompt_language: en` 정책은 본 카드가 SessionStart 부트스트랩 안내 면에 한해 개정(amend)한다; `agent_prompt_language`는 그 외 모든 agent-facing 면(서브에이전트 프롬프트 등)에 계속 지배한다. 가역적 텍스트 렌더 결정, 카드 내 판정 — keep-set 게이트 아님. 레인 결정 기록: .moai/reports/t1603/ (리더가 착지).

### Q2: 2모드 설명 문구의 착지 순서 — t1600 이전 병합을 허용하는가

Label: FOUNDER
Class: implementation-level
Authority anchor: 해당 없음 — 카드 지시문(dispatch)이 "플래그를 발명하지 말고 t1600 착지 의존을 명시적 가정/후속으로 표기"하라고 정했으나, dispatch는 커밋된 권위 장부 밖이다.
Why unresolved: 안내 문구가 플래그 없는 빌드에 먼저 착지하면 REQ-005의 설계-노면 마커가 진실성을 유지하는지, 아니면 문구를 t1600과 같은 릴리스 열차로 맞춰 마커 없이 출하할지 순서 판정이 남는다.
Default: 설계-노면 마커(카드 t1600 귀속)를 붙인 채 본 SPEC 선착지 — 마커는 t1600 착지 후 M3 재확인에서 제거 판정 (rule: the option that preserves current behavior — 플래그 없는 빌드에 살아 있는 명령처럼 읽히는 텍스트를 출하하지 않는다)
Alternate: t1600 병합 후 동시 착지(마커 불필요)
Operator verdict: DEFAULT-APPLIED 2026-10-08 manager-spec (card t1603, lane)

### Q3: 안내 메시지 테이블의 로케일 집합 — 4-로케일 유지인가 확장인가

Label: FOUNDER
Class: implementation-level
Authority anchor: 해당 없음 — `.moai/config/sections/language.yaml`은 conversation_language로 7개(ko/en/ja/zh/es/fr/de)를 광고하지만 `internal/hook/session_start_factory_i18n.go:74-77`은 4-로케일+영어 폴백을 "the complete set of conversation languages"로 진술한다. 어느 쪽도 본 질문(테이블 확장 여부)을 그대로 덮지 않는다.
Why unresolved: 카드가 "four-locale set"을 전제해 4-로케일 유지가 자연 답이지만, 광고된 es/fr/de 사용자는 폴백 영어를 읽게 된다 — 확장 여부는 사용자면 크기 판정이다.
Default: 기존 4-로케일(en/ko/ja/zh) + 영어 폴백 유지, 테이블 미확장 (rule: the option that preserves current behavior — 더 작은 사용자 노출면)
Alternate: es/fr/de 항목 추가(테이블 7-로케일 확장)
Operator verdict: DEFAULT-APPLIED 2026-10-08 manager-spec (card t1603, lane)
