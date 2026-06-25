# design.md — SPEC-V3R6-ASKUSER-DECISION-MEMORY-001 아키텍처 결정

> 아키텍처 결정, 템플릿 중립성 분할 매트릭스, 상충 증거 양면 문서화, advisory 훅 배선 지점, 감쇠 공식, 숙련도 추론 접근.

---

## §A. 아키텍처 결정 (5개 미해결 질문 해법)

### §A.1 결정 A1 — 선호 메모리 저장 위치

**옵션**:
- (a) 기존 `MEMORY.md` 확장 — 인덱스에 선호 라인 추가
- (b) `.moai/config/sections/user_profile.yaml` — 정적 YAML
- (c) `~/.claude/projects/{hash}/memory/user_decisions/` 신규 서브디렉터리 (3계층: core.yaml / recall.jsonl / archival/)

**선택**: (c)

**근거**:
- (a) 기각 — `MEMORY.md`는 기술 교훈 인덱스; 혼합 시 회수 정밀도 저하 (STRONG-3 위반)
- (b) 기각 — 정적 YAML은 동적 upsert에 부적합; 감쇠 스캔이 파일을 매번 재작성해야 함
- (c) 선택 — MemGPT 3계층(core/recall/archival) 원칙 정준 구현; `~/.claude/projects/{hash}/memory/`는 Claude Code memory docs의 정준 경로이므로 하위 서브디렉터리로 자연 통합; 기존 `feedback_*.md`와 네임스페이스 분리 (REQ-ADM-002)

**레이아웃**:
```
~/.claude/projects/{hash}/memory/
├── MEMORY.md                          # 기존 — 기술 교훈 인덱스 (변경 없음)
├── feedback_*.md                      # 기존 — 기술 교훈 상세 (변경 없음)
└── user_decisions/                    # 신규 — 본 SPEC
    ├── core.yaml                      # 항상 로드, ≤4KB, 최근 빈번 stable 사실
    ├── recall.jsonl                   # 최근 N 세션 사실 (N=설정 가능, 초기 100)
    └── archival/                      # 전체 검색 대상 (오래된/드문 사실)
```

**core.yaml 스키마** (YAML, 항상 로드):
```yaml
entries:
  - fact: "prefers Go backend over Python"
    domain: "tech_stack"
    decision_key: "backend_language"
    confidence: observed
    scope: stable
    last_used: "2026-06-24T10:00:00Z"
    valid_time: "2026-06-20T10:00:00Z"
    weight: 0.92
```

### §A.2 결정 A2 — 캡처 범위

**옵션**:
- (a) 모든 AskUserQuestion tool_result 캡처
- (b) 특정 결정 유형만 (Tier / lang / effort / agent-delegation)
- (c) 모든 캡처 + 도메인 분류 후 저장 (민감 도메인은 추천 강도 저하, 캡처는 유지)

**선택**: (c)

**근거**:
- (a)는 너무 관대람 — 노이즈(일회성 선택)가 메모리를 오염
- (b)는 너무 엄격 — 어떤 결정이 "선호 가치가 있는지" 사전에 알기 어려움
- (c) 선택 — 모든 캡처 후 도메인 분류로 일회성 transient / 반복 stable 구분; 민감 도메인(security 등)은 캡처는 하되 추천 강도를 neutral로 저하 (REQ-ADM-014). 이것이 Copilot 모델(모든 선호 추출 + 분류)과 정합

**도메인 분류 휴리스틱**:
- `decision_key`가 AskUserQuestion의 `header` 필드에서 추출 (예: "진행 방향", "Tier 선택", "언어")
- 반복 관측 시 `scope: stable`로 자동 승격 (초기: 동일 도메인 ≥3회 동일 선택)
- 일회성 선택 (header가 작업 특정적)은 `scope: transient`

### §A.3 결정 A3 — 감쇠 파라미터 초기값

**선택**: 멱법칙(power-law) `weight = (age_days + 1)^(-α)`, α=0.5 고정 (Standard tier).

**근거**:
- Ebbinghaus/Murre 망각 곡선은 멱법칍 형태 — 초기 가파른 감소 후 평탄
- α=0.5는 "square root decay" — 28일에 weight ≈ 0.19 (유효 임계); 7일에 weight ≈ 0.38
- 동적 학습(사용자별 α 추정)은 "complete" tier로 이월 — Standard는 안정적 단일 값

** stable/transient 분리 적용**:
- `scope: stable` — pure time-decay 면제; `last_used` 갱신 시 `weight` 부양 (spaced repetition)
- `scope: transient` — 멱법칙 적용; 28일(last_used 기준) 경과 시 soft-delete

**사용시 리셋**:
- 재사용(회수 시) → `age_days = 0`, `weight = 1.0`으로 리셋, `confidence` 소폭 부양

### §A.4 결정 A4 — 숙련도 추론 방법

**옵션**:
- (a) 세션 카운트 ≥ 임계값 (예: 20세션)
- (b) 의사결정 일관성 (같은 도메인에서 같은 선택 비율)
- (c) 명시적 자가 평가 ("/moai preference proficiency expert")
- (d) 위 세 가지 앙상블

**선택**: (a) 초기 + 점진적 확장으로 (b), (c) 추가

**근거**:
- (a)는 초기 구현이 단순하고 cold-start에서 작동
- (b), (c)는 더 정밀하지만 데이터가 쌓여야 작동 — 점진적 도입
- (d)는 "complete" tier 목표

**초기 임계값**: 세션 카운트 ≥ 20 → 전문가 (weak recommendation); 미만 → 일반 사용자 (strong recommendation). 임계값은 `.moai/config/sections/preference.yaml`(신규)에서 조정 가능.

**Cold-start 보호**: 세션 카운트 < 5 (초기) → neutral 강도 (REQ-ADM-014 cold-start 게이트와 동일 처리).

### §A.5 결정 A5 — 회복 제어 노출

**옵션**:
- (a) 매 AskUserQuestion에 "개인화 비활성화" 옵션 추가
- (b) 별도 `/moai preference toggle` CLI 명령
- (c) (a) + (b) 혼합

**선택**: (b)

**근거**:
- (a)는 질문 수 증가로 피로 가중 (STRONG-2 위반)
- (b)는 사용자가 필요할 때만 호출; 세션 단위 토글 (NFR-ADM-005)
- 매 추론 공개 시 정정 채널("이 추론이 틀리면 알려주세요")은 (a) 없이도 제공 (REQ-ADM-016)

**토글 상태 저장**: `.moai/state/session-preference-disabled` 센티넬 파일 (존재 = 비활성화); 세션 종료 시 자동 삭제 (또는 신규 세션 시작 시 확인 후 삭제).

---

## §B. 상충 증거 양면 문서화 (research.md §4 확장)

> 각 상충에 대해 양쪽 증거를 명시하고 버퍼(설계적 완화)를 서술. 이것이 "어느 한쪽만 인용하지 않는" 편향 방지 장치.

### §B.1 추천 효과 vs 자율성 침식

**추천 효과 측**:
- Beshears 메타분석 d=0.546 — 기본값(추천)은 합리적 선택을 유도
- Sinha 투명성 — "왜"가 동반된 추천은 선호도 3.51 (불투명 2.79 대비)

**자율성 침식 측**:
- Loughrey — 추천이 2차 욕구(내가 원하지 않는 것을 원하게 되는)를 침식
- 필터 버블 (Pariser/Iyendo) — 과도 추천이 정보 다양성 축소

**버퍼 (본 SPEC 설계)**:
1. REQ-ADM-013 회복 제어 토글 — 세션 단위 비활성화
2. REQ-ADM-014 민감 도메인 게이트 — security/cold-start에서 neutral 강도
3. REQ-ADM-016 정정 루프 — 추론 정정 즉시 반영
4. REQ-ADM-017 적응형 강도 — 전문가에게 약 추천(info-centric)

### §B.2 명시적 선호 정밀 vs 질문 수 피로

**명시적 정밀 측**:
- Fisher 정보 I=p(1−p) p=0.5 최대 — 결정 경계 질문이 최대 정보이익

**질문 수 피로 측**:
- CHI 2025 just-in-time > batched — 질문 수 자체가 피로

**버퍼**:
- REQ-ADM-005 발화 시점 = p≈0.5만 질문; p≈0/1 자동 처리 → 총 질문 수 감소
- REQ-ADM-006 정보이익 내림차순 정렬 → 낮은 정보이익 질문 전에 핵심 완료

### §B.3 오래된 데이터 감쇠 vs 지속 신호 상실

**감쇠 필요 측**:
- Copilot 28일 TTL — stale data로 인한 잘못된 추천 위험

**지속 신호 상실 측**:
- Koren temporal dynamics — 순진 time-decay가 영구적 사용자 특성을 잃음

**버퍼**:
- REQ-ADM-011 stable/transient 분리 — stable은 pure time-decay 면제
- REQ-ADM-012 transient 28일 TTL + 사용시 리셋

### §B.4 개인화 신뢰 향상 vs 역화 / 피드백 루프 편향

**신뢰 향상 측**:
- Mem0 / Generative Agents — 적응형 메모리가 사용자 경험 향상

**역화 측**:
- 개인화 backfire 연구 — 과도한 개인화가 거부감 유발
- 피드백 루프 편향 — 추천이 관측을 왜곡, 관측이 추천을 강화

**버퍼**:
- REQ-ADM-017 적응형 강도 — 숙련도별 차등 적용
- per-task cold-start 재실행 — 매 새 작업에서 추정 초기화
- Iyendo 우연성 주입 — 의도적 탐색 허용 (REQ-ADM-014 맥락 게이트와 연동)

### §B.5 자동 메모리 유용 vs 명시적 규칙 SSOT

**자동 유용 측**:
- Copilot, Cursor, Devin 모두 자동 메모리 채택

**명시적 SSOT 측**:
- Windsurf 철학 — 자동 메모리는 보조적; 지속 지식은 명시적 규칙이 SSOT

**버퍼**:
- 이원층 — (1) 본 SPEC의 자동 사용자 의사결정 메모리 + (2) 기존 명시적 기술 교훈 메모리(`feedback_*.md`) 공존
- 상호 대체 아닌 보완 — Windsurf 우려(자동 메모리 불안정)를 기존 명시적 계층 유지로 회피

---

## §C. Advisory 훅 배선 지점 (PostToolUse 파이프라인)

### §C.1 기존 배선 구조

```
Claude Code PostToolUse event
  ↓ (stdin JSON)
.claude/hooks/moai/handle-post-tool.sh   (wrapper, 변경 없음)
  ↓ exec
moai hook post-tool                       (CLI 진입)
  ↓
internal/cli/hook.go runHarnessObserve    (기존 — usage-log.jsonl 기록)
  ↓ (추가)
[user_decision_capture 서브파이프라인]    (신규 — 본 SPEC)
  ↓
internal/cli/preference/ Store.Upsert
```

### §C.2 신규 서브파이프라인 — user_decision_capture

**트리거 조건**: stdin JSON의 `tool_name == "AskUserQuestion"` (또는 MCP 동등). tool_result 페이로드에 `selected_option_label` / `question_header` 존재.

**처리 순서**:
1. Recovery-Signal Carve-Out 탐지 — stdin JSON의 `stopReason` 또는 주변 맥락이 회복 신호면 exit 0 + 미실행 (REQ-ADM-010)
2. tool_name 필터 — AskUserQuestion이 아니면 skip (기존 runHarnessObserve 경로로 전달)
3. tool_result 파싱 — selected_option_label, question_header, options 배열 추출
4. 도메인 분류 — header 기반 decision_key 추출 (§A.2 휴리스틱)
5. upsert — `internal/cli/preference/` Store.Upsert 호출 (domain, decision_key, fact, source_citation=session_id, scope=transient 초기, confidence=observed)
6. 오류 시 — advisory/fail-open (REQ-ADM-009) — 모든 오류 exit 0 + `.moai/logs/hook-stderr.log` warn

### §C.3 cohabitation 규칙 (기존 exit-2 훅과 충돌 회피)

- `status-transition-ownership.sh` (PostToolUse on SPEC body content) — 본 SPEC 캡처 훅과 동일 PostToolUse 이벤트 사용. 그러나:
  - `status-transition-ownership.sh`는 stdin JSON의 `tool_name == Write|Edit` + 파일 경로 `.moai/specs/SPEC-*/{spec,plan,acceptance}.md` 매칭 시에만 동작
  - 본 SPEC 캡처 훅은 `tool_name == AskUserQuestion` 매칭 시에만 동작
  - 두 훅은 상호 배타적 트리거이므로 충돌 없음
- `sync-phase-quality-gate.sh` (Stop hook) — 본 SPEC 캡처 훅(PostToolUse)과 다른 이벤트. 충돌 없음.
- `team-ac-verify.sh` (TaskCompleted) — 본 SPEC과 무관. 충돌 없음.
- `internal/hook/CLAUDE.md` cohabitation contract 준수.

### §C.4 Recovery-Signal Carve-Out — advisory 훅의 stopReason 파싱 gap 상속 (SHOULD, documentation-only)

> 본 절은 REQ-ADM-010의 SHOULD 성격과 AP-RR-006 준거를 아키텍처 수준에서 문서화한다.

본 SPEC의 advisory 캡처 훅(`user_decision_capture` 서브파이프라인)은 기존 `sync-phase-quality-gate.sh`(Stop) 및 `status-transition-ownership.sh`(PostToolUse) 훅과 **동일한 stopReason-파싱 gap을 상속**한다. `runtime-recovery-doctrine.md §4` + AP-RR-006이 명시하듯:

> "the current hooks do NOT parse stopReason; no mechanical enforcement is possible without a runtime-layer hook that parses stopReason, deferred to a future SPEC (`SPEC-V3R6-HOOK-RECOVERY-SIGNAL-001`)."

따라서:

1. **REQ-ADM-010은 SHOULD이다** (HARD `shall`이 아님). 본 SPEC은 회복 턴이 *탐지되었을 때의 행동*(exit 0 + 캡처 미실행)만 정의한다.
2. **탐지 메커니즘은 out-of-scope이다.** 회복 턴을 식별하는 stopReason 파싱은 현재 advisory 훅 layer에서 불가능하며, future `SPEC-V3R6-HOOK-RECOVERY-SIGNAL-001`에서 구현된다.
3. **Carve-out은 documentation-only policy guidance이다.** 현재 layer에서 기계적 강제가 불가능하므로 (AP-RR-006), 이 carve-out은 훅 소스 코드에 위임 사실을 문서화하는 방식으로만 준수된다.
4. **대체 proxy 탐지 도입 금지.** stopReason을 파싱할 수 없다는 이유로 가짜 proxy 신호(예: stderr 패턴 매칭, 시간 기반 추정)를 도입해 "회복 턴을 탐지했다"고 주장하는 것은 over-claim을 재배치하는 것일 뿐이며 AP-RR-006 위반이다. 정직한 해법은 SHOULD를 유지하고 탐지를 future SPEC에 위임하는 것이다.

**구현 지침 (M3)**: advisory 훅 소스(`internal/hook/post_tool.go` user_decision_capture 분기)에 다음을 주석으로 명시:

```go
// REQ-ADM-010 (SHOULD, doctrine-honest): If this turn is a recovery turn
// (stopReason indicates PTL/max_output_tokens/media_size/compact-failure),
// the hook SHOULD exit 0 without capture. However, the current advisory
// hook CANNOT parse stopReason (per runtime-recovery-doctrine.md §4 +
// AP-RR-006). Detection mechanism is deferred to future
// SPEC-V3R6-HOOK-RECOVERY-SIGNAL-001. This hook does NOT claim to
// mechanically detect recovery turns today.
```

이 주석은 AC-ADM-010의 관측 증거가 된다 (탐지 불가를 정직하게 문서화했는가).

---

## §D. 템플릿 중립성 분할 매트릭스 (NFR-ADM-006 준거)

> `.moai/docs/template-internal-isolation-doctrine.md §25`. INTERNAL-ONLY 산출물은 moai-adk 내부 개발 전용; TEMPLATE-SHIPPED 산출물은 16-언어 범용 배포 (내부 SPEC ID / REQ 토큰 / 날짜 / SHA 금지).

| 산출물 | 분류 | 중립성 제약 | CI guard |
|--------|------|------------|----------|
| `.moai/specs/SPEC-V3R6-ASKUSER-DECISION-MEMORY-001/` (SPEC 디렉터리 전체) | INTERNAL-ONLY | 없음 (내부 개발 산출물) | — |
| `internal/cli/preference/` (Go 패키지) | INTERNAL-ONLY | 없음 (Go 코드) | — |
| `internal/hook/post_tool.go` (Go 확장) | INTERNAL-ONLY | 없음 (Go 코드) | — |
| `internal/cli/hook.go runHarnessObserve` 확장 | INTERNAL-ONLY | 없음 (Go 코드) | — |
| `~/.claude/projects/{hash}/memory/user_decisions/` (데이터 경로) | INTERNAL-ONLY | 없음 (사용자 데이터) | — |
| `.moai/state/session-preference-disabled` (센티넬) | INTERNAL-ONLY | 없음 (상태 파일) | — |
| `.moai/logs/hook-stderr.log` (로그) | INTERNAL-ONLY | 없음 (로그) | — |
| `.moai/config/sections/preference.yaml` (신규 설정) | INTERNAL-ONLY | 없음 (내부 설정) | — |
| **`internal/template/templates/.claude/rules/moai/core/askuser-protocol.md`** (template-shipped 카피) | **TEMPLATE-SHIPPED** | **중립성 제약 적용** — 범용 추천 배치 원칙만; "SPEC-V3R6-ASKUSER-DECISION-MEMORY-001" / "REQ-ADM-001" 등 내부 ID 금지 | `internal_content_leak_test.go` |
| **`internal/template/templates/.claude/output-styles/moai/moai.md`** (template-shipped 카피) | **TEMPLATE-SHIPPED** | **중립성 제약 적용** — 범용 렌더 규칙; 내부 ID 금지 | `internal_content_leak_test.go` |
| `internal/template/templates/.claude/hooks/moai/handle-post-tool.sh` (template-shipped wrapper) | TEMPLATE-SHIPPED | 이미 범용 (변경 없음) | 기존 guard |
| docs-site `content/en/workflow-commands/moai-plan.md` + ko/ja/zh (M6) | TEMPLATE-SHIPPED (사용자 대상) | 중립성 — 범용 사용자 대상; 내부 SPEC ID는 "Epic" 맥락에서만 | docs-site CI |

### §D.1 template-shipped 카피 작성 원칙

`askuser-protocol.md`의 template-shipped 카피에는:
- **포함**: "추천 옵션은 사용자의 관측된 통계적 다수 선호 기반이어야 한다" (범용 원칙)
- **포함**: "추천 서술은 전제조건을 명시해야 한다" (범용 원칙)
- **포함**: "cold-start 시 정적 기본값 + 공개" (범용 원칙)
- **제외**: "REQ-ADM-007" (내부 토큰)
- **제외**: "SPEC-V3R6-ASKUSER-DECISION-MEMORY-001" (내부 SPEC ID)
- **제외**: "2026-06-24" / commit SHA (내부 날짜/식별자)

live 카피(`.claude/rules/moai/core/askuser-protocol.md`)는 template-shipped와 동일 내용이지만, 내부 컨텍스트에서 SPEC ID를 교차 참조할 수 있다 (예: "본 규칙은 SPEC-V3R6-ASKUSER-DECISION-MEMORY-001에서 정의됨"). 단, 이 내부 참조는 live 카피에만 존재하고 template-shipped에는 누출되지 않는다.

### §D.2 CI guard 검증

```bash
# M2/M6 완료 시 실행
grep -rE 'SPEC-V3R6|REQ-ADM|AC-ADM|2026-06-24|[0-9a-f]{40}|\.moai/(specs|reports)' \
  internal/template/templates/.claude/rules/moai/core/askuser-protocol.md \
  internal/template/templates/.claude/output-styles/moai/moai.md
# 기대: 0건

go test ./internal/template/ -run TestInternalContentLeak
# 기대: PASS
```

---

## §E. 감쇠 공식 상세

### §E.1 멱법칙 weight 함수

```
weight(age_days) = (age_days + 1)^(-α)     # α = 0.5 (Standard tier 고정)
```

| age_days | weight |
|----------|--------|
| 0 | 1.000 |
| 1 | 0.707 |
| 7 | 0.354 |
| 14 | 0.258 |
| 28 | 0.186 |
| 56 | 0.133 |

**유효 임계**: weight < 0.15 → recall에서 archival로 강등 가능 (회수 빈도 기반)

### §E.2 stable/transient 분기 로직

```
if entry.scope == "stable":
    # pure time-decay 면제
    entry.weight = max(entry.weight, 0.5)  # 최소 0.5 보장
    entry.last_used 갱신 시 weight 부양: weight = min(1.0, weight + 0.1)
elif entry.scope == "transient":
    # 멱법칙 적용
    entry.weight = power_law(age_days)
    if age_days > 28:
        soft_delete(entry)  # archival/로 이동
```

### §E.3 사용시 리셋

```
on_reuse(entry):
    entry.last_used = now()
    entry.age_days = 0
    entry.weight = 1.0
    entry.confidence = boost(entry.confidence)  # observed는 1.0 유지; inferred는 소폭 부양
```

---

## §F. 위험 분석 (plan.md §F 확장)

| 위험 | 확률 | 영향 | 완화 |
|------|------|------|------|
| 캡처 훅 race (다중 세션 동시 캡처) | 중 | 중 | 원자적 upsert (파일 락 또는 single-writer 패턴); Pre-Spawn Sync Check |
| core 크기 4KB 초과 | 중 | 저 | 자동 강등 (오래된/낮은 weight 엔트리 recall로) |
| 숙련도 추정 부정확 (초기) | 높 | 중 | cold-start neutral 강도; 점진적 정제 |
| 템플릿 중립성 위반 (실수로 내부 ID 누출) | 중 | 고 | CI guard + 사전 커밋 체크리스트 |
| 회복 턴에서 캡처가 death-spiral 유발 | 저 | 치명 | REQ-ADM-010 Recovery-Signal Carve-Out (SHOULD, doctrine-honest — 탐지 메커니즘 future SPEC-V3R6-HOOK-RECOVERY-SIGNAL-001 이연; AP-RR-006 준거 — 현재 layer에서 기계적 강제 불가, documentation-only) |
| 추천이 filter bubble 유발 | 중 | 중 | REQ-ADM-014 맥락 게이트 + 우연성 주입 |
| 멱법칙 α=0.5가 특정 도메인에 부적합 | 중 | 저 | 도메인별 α 오버라이드 (complete tier); Standard는 단일 값 |

---

## §G. 대안 (고려되었으나 기각)

| 대안 | 기각 이유 |
|------|----------|
| RL 기반 추천 정책 학습 (Pep ICML 2026) | 복잡도 과다; Standard tier 범위 초과 — complete tier 이월 |
| 외부 도구(Cursor/Copilot)와의 메모리 동기화 | 프라이버시 민감; 별도 SPEC 필요 |
| 매 AskUserQuestion에 토글 옵션 추가 | 질문 수 증가 → 피로 (STRONG-2 위반) |
| append-only 메모리 | 토큰 비용 선형 증가; 통합 원칙 위반 (STRONG-1) |
| 정적 YAML만 (감쇠 없음) | stale data 위험; Copilot/Koren 증거 무시 |
| 순진 time-decay (stable/transient 분리 없음) | 지속 신호 상실 (Koren); 핵심 선호 반복 질문 |

---

## §H. 교차 참조

- spec.md §B/C — 5 컴포넌트 + 18 REQ
- plan.md §F — M1~M6 마일스톤
- acceptance.md §D — 18+ AC
- research.md — 5 연구 각도 + 25 인용
- `.claude/rules/moai/core/askuser-protocol.md` — M2 수정 대상
- `.claude/rules/moai/workflow/runtime-recovery-doctrine.md §4` — REQ-ADM-010 준거
- `.claude/rules/moai/core/verification-claim-integrity.md §1.1 surface 3` — REQ-ADM-018 준거
- `.moai/docs/template-internal-isolation-doctrine.md §25` — NFR-ADM-006 준거
- `internal/cli/hook.go runHarnessObserve` — M3 배선 지점
- `internal/hook/CLAUDE.md` cohabitation contract — §C.3 준거
- `internal/template/internal_content_leak_test.go` — §D.2 CI guard
