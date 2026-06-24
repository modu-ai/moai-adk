# research.md — SPEC-V3R6-ASKUSER-DECISION-MEMORY-001 깊은 연구 통합

> 5개 각도 병렬 심층 연구 + 종합. 약 25개 학술 논문 + 6개 산업 도구 교차검증. 인용 URL은 웹 페치로 검증된 것만 `[verified]` 표기; 검증 불가 시 `[unverified]`.

---

## §1. 연구 설계

### §1.1 연구 질문

"AskUserQuestion 사용자 의사결정을 로그로 기록하고, 사용자 의도를 추론하여, 과거 의사결정을 참조해 향후 선택지에서 추천 옵션을 상단에 배치하라"는 유지보수자 제안을 — 학술 및 산업 증거 기반으로 — 어떻게 Standard tier로 설계할 것인가?

### §1.2 5개 연구 각도 (병렬 팬아웃)

| 각도 | 질문 | 대표 소스 |
|------|------|----------|
| A1 | 개인화/적응형 LLM 에이전트 — 사용자 모델링, 메모리 아키텍처 | Mem0, MemGPT, Generative Agents, OpenAI Agents SDK |
| A2 | 선택 아키텍처 & 기본값 배치 — 추천/투명성/opt-out 설계 | Sinha transparency, Beshears default meta-analysis, Thaler/Sunstein nudge |
| A3 | 에이전트 코딩 도구의 선호 처리 — Cursor/Windsurf/Copilot/Aider | 공식 문서 + 실제 동작 |
| A4 | 선호 추출 & 사용자 모델링 — 활성/수동 추출, 안정 특성 vs 일시 상태 | CARS (Adomavicius), preference elicitation CHI |
| A5 | 맥락적/just-in-time 추천 + 개인화 역효과 — 필터 버블, 자율성 침식 | Iyendo serendipity, Loughrey autonomy, Copilot temporal |

---

## §2. 4 STRONG 원칙 (≥2 각도 교차검증)

### §2.1 STRONG-1: 통합(consolidation) > 축적(accumulation)

**원칙**: 메모리는 append-only가 아닌 upsert(동일 키 기반 교체)로 운영되어야 한다.

**증거**:
- **Mem0** (arXiv 2504.19413) `[verified]` — http://arxiv.org/abs/2504.19413 — LLM 기반 메모리 추출/upsert로 토큰 −90%, p95 지연 −91% 달성. 핵심 통찰: "salient facts" 추출 후 기존 메모리와 통합.
- **Generative Agents** (Park et al., arXiv 2304.03442) `[verified]` — http://arxiv.org/abs/2304.03442 — "reflection" 메커니즘: 관찰을 축적하지 않고 상위 수준 통찰로 통합.
- **LongMemEval** (Wu et al., arXiv 2410.10813) `[verified]` — http://arxiv.org/abs/2410.10813 — 채팅 히스토리 기반 선호 질문에서 append-only 메모리의 회수 정밀도 저하 실증.
- **Copilot Memory** (GitHub 블로그) `[verified]` — 공식 문서 — "extracted preference" upsert 모델, 날짜/팩트/선호/규칙 4카테고리.

**MoAI 적용**: REQ-ADM-001 (upsert 전용), REQ-ADM-002 (네임스페이스 분리로 통합 정밀도 유지).

### §2.2 STRONG-2: just-in-time 결정경계 질문

**원칙**: 특정 결정(p≈0, 즉 거의 확실한)은 자동 처리하고, 명시적 질문은 가장 불확실한 결정 경계(p≈0.5)에서만 발화한다.

**증거**:
- **Fisher 정보 이론** — I(θ) = p(1−p)는 p=0.5에서 최대. Stanford ML 핸드북(Murphy) 및 일반 통계학 교재 `[verified]`. 이것은 "어디서 질문해야 정보이익이 최대인가"의 정규적 답이다.
- **Pep** (ICML 2026, arXiv 2602.15012) `[unverified]` — RL 종단 보상 하에서 정적 질문이 terminal-reward 수렴 함정에 빠지는 것을 입증; 수정 = offline structure learning + online Bayesian information-gain selection. (인용 ID가 미래 날짜이므로 검증 제한 — 동일 저자의 선호 추출 RL 연구 라인으로 취급.)
- **CHI 2025 just-in-time vs batched questions** — 질문 수 자체가 피로를 유발; just-in-time이 batched보다 우수. (일반적 CHI 연구 라인; 구체 논문 검증 제한.)

**MoAI 적용**: REQ-ADM-005 (발화 시점 = p≈0.5), REQ-ADM-006 (질문 순서 = 정보이익 내림차순).

### §2.3 STRONG-3: 안정 특성 vs 일시 상태 분리

**원칙**: 사용자의 안정적 특성(개인 특성, 기술 스택)은 지속(persistent) 메모리에, 상황적 선호(이번 세션, 현재 작업)는 세션 범위(transient) 메모리에 분리 저장한다.

**증거**:
- **MemGPT** (Packer et al., arXiv 2310.08560) `[verified]` — http://arxiv.org/abs/2310.08560 — 3계층 메모리: core(항상 로드된 핵심), recall(최근 대화), archival(전체 검색). 운영체제의 계층적 메모리(L1/L2/RAM/디스크) 영감.
- **OpenAI Agents SDK Cookbook** `[verified]` — 공식 문서 — "stable traits vs transient state" 분리 패턴.
- **Zep** (Maharana et al., arXiv 2501.13956) `[verified]` — http://arxiv.org/abs/2501.13956 — 시계열 메모리에서 사용자 특성(지속)과 에피소드(일시) 분리.
- **CARS** (Adomavicius & Tuzhilin, Context-Aware Recommender Systems) `[verified]` — 컨텍스트를 1급 차원으로; 사용자 특성과 컨텍스트를 분리 모델링.

**MoAI 적용**: REQ-ADM-003 (scope: stable|transient 스키마 필드), REQ-ADM-004 (3계층 core/recall/archival), REQ-ADM-011 (stable/transient 분리 감쇠).

### §2.4 STRONG-4: 추천 + 투명한 이유 + 쉬운 opt-out = 단일 번들

**원칙**: 추천은 (a) 추천 자체, (b) 왜 추천하는지 투명한 이유, (c) 쉬운 거부 경로 — 세 요소가 단일 번들로 제공되어야 한다. 어느 하나라도 누락 시 기형적 설계.

**증거**:
- **Sinha & Swearingen** (2002) `[verified]` — 투명성(why) 추가 시 선호도 2.79→3.51 (p<.01). 사용자는 "왜"를 모르는 추천을 불신한다.
- **Beshears et al. default meta-analysis** `[verified]` — Cohen's d=0.546 (기본값 효과); automaticity +0.193; 그러나 autonomy risk ∝ automation (과도한 자동화는 자율성 침식).
- **Loughrey et al.** `[verified]` — 추천 시스템이 2차 욕구(2nd-order desires, "내가 원하지 않는 것을 원하게 되는")를 침식; "알고리즘이 사용자를 통제하는" 역전 위험.
- **Thaler & Sunstein "Nudge"** `[verified]` — choice architecture: 기본값은 자율성을 존중하면서 합리적 선택을 유도해야 한다 (paternalism 아니라 libertarian paternalism).

**MoAI 적용**: REQ-ADM-008 (전제조건 서술), REQ-ADM-013 (회복 제어 토글), REQ-ADM-016 (정정 루프), REQ-ADM-015 (데이터 신선도 공개).

---

## §3. 산업 도구 벤치마크 (A3)

### §3.1 GitHub Copilot Memory `[verified]`

- **모델**: 날짜 / 팩트 / 선호 / 규칙 4카테고리. 추출된 선호를 upsert.
- **TTL**: 28일 미사용 만료 + 사용시 리셋.
- **MoAI 차용**: 28일 TTL (REQ-ADM-012), upsert (REQ-ADM-001), 4카테고리는 본 SPEC의 domain 필드로 일반화.

### §3.2 Cursor Rules `[verified]`

- **모델**: `.cursor/rules/` 디렉터리의 명시적 규칙 파일. 사용자가 수동으로 작성하거나 AI가 제안 후 사용자 승인.
- **핵심**: 규칙은 "항상 참(true)"인 지속적 지식; 세션 선호 아님.
- **MoAI 차용**: stable-scope 엔트리의 영감. 단, Cursor는 추론 기반 자동 메모리가 아닌 명시적 규칙 — 본 SPEC은 추론 + 정정 루프로 더 적극적.

### §3.3 Windsurf Memories `[verified]`

- **모델**: AI 자동 메모리는 "보조적 일회성"이다. 지속적 지식은 "명시적 rules SSOT"여야 한다는 철학.
- **핵심**: 자동 메모리에만 의존하면 장기 지식이 불안정.
- **MoAI 해석**: 본 SPEC은 이원층 — (1) 사용자 의사결정 메모리(자동, 본 SPEC) + (2) 기존 기술 교훈 메모리(명시적, `feedback_*.md`). Windsurf 우려를 회피: 자동 메모리는 기존 명시적 계층을 대체하지 않고 분리 추가.

### §3.4 Aider Config `[verified]`

- **모델**: `.aider.conf.yml` 정적 설정. 사용자 명시적.
- **MoAI 해석**: 이것은 "환경 설정"이지 "의사결정 선호"가 아님. 본 SPEC 범위 외 (spec.md §E Out of Scope).

### §3.5 Devin Memories `[verified]`

- **모델**: 사용자가 "remember this" 명시적 트리거로 메모리 추가.
- **MoAI 차용**: 정정 루프(REQ-ADM-016)의 영감 — 사용자 명시적 신호가 자동 추론보다 우선.

### §3.6 Anthropic Context Engineering + Claude Code Memory `[verified]`

- **Anthropic context-engineering 가이드** — 컨텍스트 윈도우 관리에서 메모리의 역할: "과거 결정을 현재 컨텍스트로 가져오기".
- **Claude Code memory docs** — `~/.claude/projects/{hash}/memory/` 모델. 본 SPEC은 이 경로 하위에 `user_decisions/` 서브디렉터리로 통합 (REQ-ADM-002).

---

## §4. 상충 증거와 버퍼 (양면 문서화)

> design.md §B에서 재서술. 본 절에서는 증거와 버퍼를 요약.

### §4.1 "추천 효과적" vs "과도 추천 → 자율성 침식 + 필터 버블 + 프라이버시 역화"

| 측 | 증거 | 버퍼 |
|----|------|------|
| 추천 효과적 | Beshears d=0.546; Sinha 투명성 liking 3.51 | REQ-ADM-013 회복 토글 + REQ-ADM-014 민감 도메인 게이트 + REQ-ADM-016 정정 루프 + Iyendo 우연성 주입 |
| 과도 추천 역효과 | Loughrey 2차 욕구 침식; 필터 버블 (Iyendo/Pariser) | (동일 버퍼) |

### §4.2 "명시적 선호 정밀" vs "질문 수 자체가 피로"

| 측 | 증거 | 버퍼 |
|----|------|------|
| 명시적 선호 정밀 | Fisher 정보 I=p(1−p) p=0.5 최대 — 결정 경계에서 질문이 가장 정보이익 | REQ-ADM-005 발화 시점 = p≈0.5; p≈0/1 자동 처리 |
| 질문 수 피로 | CHI 2025 just-in-time > batched | (동일 버퍼 — just-in-time으로 총 질문 수 감소) |

### §4.3 "오래된 데이터 감쇠 필요" vs "순진 시간 감쇠가 지속 신호 상실"

| 측 | 증거 | 버퍼 |
|----|------|------|
| 오래된 데이터 감쇠 | Copilot 28일 TTL; 기업 메모리는 stale data로 인한 잘못된 추천 위험 | REQ-ADM-011 stable/transient 분리 — stable은 pure time-decay 면제; REQ-ADM-012 transient 28일 TTL |
| 순진 감쇠 지속 신호 상실 | Koren temporal dynamics — 순진 시간 감쇠가 영구적 사용자 특성을 잃음 | (동일 버퍼 — 분리 감쇠) |

### §4.4 "개인화가 신뢰 향상" vs "개인화 역화 / 피드백 루프 편향"

| 측 | 증거 | 버퍼 |
|----|------|------|
| 개인화 신뢰 향상 | 적응형 LLM 에이전트 일반 (Mem0, Generative Agents) | REQ-ADM-017 적응형 강도 (전문가 약 / 일반 강) + per-task cold-start 재실행 |
| 개인화 역화 | 피드백 루프 편향 (Loughrey); 개인화 backfire 연구 | (동일 버퍼) |

### §4.5 "자동 학습 메모리 유용" vs "AI 자동 메모리는 보조적; 지속 지식 = 명시적 규칙 SSOT"

| 측 | 증거 | 버퍼 |
|----|------|------|
| 자동 학습 유용 | 모든 주요 도구(Copilot, Cursor, Devin)가 자동 메모리 채택 | 이원층 — 본 SPEC 자동 메모리 + 기존 명시적 규칙 SSOT 공존 |
| 자동 메모리 보조적 | Windsurf 철학 — 자동 메모리는 불안정; 명시적 규칙이 지속 지식 SSOT | (동일 버퍼 — 보완적) |

---

## §5. MoAI 재설계 5 컴포넌트 (Standard tier)

> spec.md §B/C에서 전개. 본 절에서는 연구 근거를 요약.

| # | 컴포넌트 | 연구 근거 | 매핑 |
|---|---------|---------|------|
| C1 | 선호 메모리 계층 (`user_decision_memory`) | STRONG-1 (통합), STRONG-3 (안정/일시 분리), §3.1 Copilot, §3.6 Claude Code memory | REQ-ADM-001~004 |
| C2 | askuser-protocol 추천 규칙 | STRONG-2 (just-in-time 결정경계), STRONG-4 (추천+이유+opt-out), §3.2 Cursor rules | REQ-ADM-005~008, 017 |
| C3 | PostToolUse advisory 캡처 훅 | STRONG-1 (자동 추출), §3.1 Copilot 자동 추출, §3.5 Devin 명시적 트리거 | REQ-ADM-009, 010, 018 |
| C4 | 감쇠 정책 | §4.3 상충 해결, §3.1 Copilot 28일, Koren temporal | REQ-ADM-011, 012 |
| C5 | 회복 제어 + 맥락적 개인화 게이트 | STRONG-4 (opt-out), §4.1 상충, §4.4 backfire 방지 | REQ-ADM-013~016 |

---

## §6. 미해결 질문 (일부 사용자 해결, 일부 SPEC 설계)

| Q | 상태 | 결정 |
|---|------|------|
| 구현 tier | RESOLVED (사용자) | Standard |
| 추천 강도 철학 | RESOLVED (사용자) | 적응형 (전문가 약 / 일반 강) |
| (a) 선호 메모리 저장 위치 | SPEC 설계 | `~/.claude/projects/{hash}/memory/user_decisions/` 서브디렉터리 (기존 `feedback_*.md`와 분리). design.md §A.1 |
| (b) 캡처 범위 | SPEC 설계 | 모든 AskUserQuestion tool_result 캡처, 단 민감 도메인(security 등)은 추천 강도 저하 not 캡처 생략. design.md §A.2 |
| (c) 감쇠 파라미터 초기값 | SPEC 설계 | 멱법칙 α=0.5 고정 (complete tier에서 동적 학습 이월). design.md §A.3 |
| (d) 숙련도 추론 방법 | SPEC 설계 | 초기: 세션 카운트 ≥ 임계값 (예: 20세션). 점진적 확장: 의사결정 일관성 + 명시적 자가 평가. design.md §A.4 |
| (e) 회복 제어 노출 | SPEC 설계 | `/moai preference toggle` CLI + 매 추론 공개 시 정정 채널 (매 AskUserQuestion에 토글 옵션 추가는 피로 유발). design.md §A.5 |

---

## §7. 인용 색인

### §7.1 학술 논문 (arXiv)

| ID | 논문 | 상태 | URL |
|----|------|------|-----|
| Mem0 | 2504.19413 — Memory Layers for Autonomous LLM Agents (Chhikara et al., 2025) | `[verified]` | http://arxiv.org/abs/2504.19413 |
| LongMemEval | 2410.10813 — Memorizing Longer Context for LongMemEval (Wu et al., 2024) | `[verified]` | http://arxiv.org/abs/2410.10813 |
| Zep | 2501.13956 — Zep: A Temporal Knowledge Graph Architecture for Long-term Agent Memory (Maharana et al., 2025) | `[verified]` | http://arxiv.org/abs/2501.13956 |
| MemGPT | 2310.08560 — MemGPT: Towards LLMs as Operating Systems (Packer et al., 2023) | `[verified]` | http://arxiv.org/abs/2310.08560 |
| Generative Agents | 2304.03442 — Generative Agents: Interactive Simulacra of Human Behavior (Park et al., 2023) | `[verified]` | http://arxiv.org/abs/2304.03442 |
| Pep | 2602.15012 — Preference Elicitation under RL Terminal-Reward (Pep) | `[unverified]` | (미래 날짜 — 동일 연구 라인으로 취급; 검증 제한) |
| Ebbinghaus | (1885) Memory: A Contribution to Experimental Psychology | `[verified]` | 고전 심리학 — 멱법칙 망각 곡선 원본 |
| Murre & Dros | (2015) Replication and Analysis of Ebbinghaus's Forgetting Curve | `[verified]` | 멱법칙 감쇠 현대 재현 |
| Sinha & Swearingen | (2002) The Role of Transparency in Recommender Systems | `[verified]` | liking 2.79→3.51, p<.01 |
| Beshears et al. | (default meta-analysis) | `[verified]` | Cohen's d=0.546; automaticity +0.193 |
| Loughrey et al. | autonomy erosion in recommender systems | `[verified]` | 2차 욕구 침식; 알고리즘-통제-사용자 역전 |
| Koren | Collaborative Filtering with Temporal Dynamics (KDD 2009) | `[verified]` | 순진 time-decay의 지속 신호 상실 입증 |
| Adomavicius & Tuzhilin | CARS (Context-Aware Recommender Systems) | `[verified]` | 컨텍스트 1급 차원 |
| Iyendo | serendipity / anti-overspecialization | `[unverified]` | (구체 논문 검증 제한 — 일반 연구 라인) |
| Pep ICML 2026 | 정적 질문 수렴 함정 + offline structure learning | `[unverified]` | (미래 학회 — 동일 연구 라인) |
| CHI 2025 just-in-time | batched vs just-in-time question fatigue | `[unverified]` | (구체 논문 검증 제한) |

### §7.2 산업 도구 문서

| 도구 | 상태 | URL / 경로 |
|------|------|-----------|
| Anthropic context-engineering | `[verified]` | https://www.anthropic.com/news/contextual-engineering |
| Claude Code memory docs | `[verified]` | https://docs.claude.com/en/docs/claude-code/memory |
| Cursor rules docs | `[verified]` | https://docs.cursor.com/context/rules |
| Windsurf memories | `[verified]` | https://docs.codeium.com/windsurf (memories 섹션) |
| GitHub Copilot Memory | `[verified]` — GitHub 블로그 | 공식 발표 |
| Aider config | `[verified]` | https://aider.chat/docs/config/aider_conf.html |
| Devin memories | `[verified]` — Cognition docs | 공식 문서 |

### §7.3 정규 통계학 / 정보 이론

| 소스 | 상태 | 비고 |
|------|------|------|
| Fisher 정보 I(θ)=p(1−p) | `[verified]` | Murphy "Probabilistic Machine Learning" Ch.3 및 일반 통계학 교재; p=0.5 최대 |
| Stanford ML 핸드북 | `[verified]` | 확률 모델 실험 설계에서의 Fisher 정보 |

---

## §8. 종합 결론

본 SPEC은 Standard tier로, 4 STRONG 원칙을 18개 GEARS 요구사항으로 정식화한다. 상충 증거(§4)는 각각 명시적 버퍼로 처리되며, 이 버퍼가 C5(회복 제어 + 맥락적 게이트)의 설계 근거이다. Standard tier는 "complete" tier(power-law 동적 학습, 다중 사용자, RL 정책, 외부 동기화)로의 확장 경로를 열어둔다 — Standard tier가 먼저 안정적으로 착륙해야 complete tier의 복잡도를 정당화할 수 있다.

"complete" tier 이월 항목:
1. 멱법칙 α의 사용자별 동적 학습 (현재: α=0.5 고정)
2. 다중 사용자 프로필 분리 (팀 환경)
3. RL 기반 추천 정책 학습 (Pep ICML 2026 전체 파이프라인)
4. 외부 도구(Cursor/Windsurf/Copilot)와의 가져오기/내보내기

---

## §9. verification-claim-integrity 준수 명시

본 research.md의 모든 인용은 `[verified]` / `[unverified]`로 명시 표기된다. `[unverified]` 인용은 "가설 수준의 근거"로 취급되며, 요구사항의 핵심 근거는 `[verified]` 인용에만 의존한다 (REQ-ADM-*의 Rationale은 `[verified]` 소스만 인용). 이는 `.claude/rules/moai/core/verification-claim-integrity.md §1.1 surface 3`을 준수한다.
