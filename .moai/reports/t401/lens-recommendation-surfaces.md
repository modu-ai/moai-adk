# 렌즈 1 — 권고 방출(recommendation-push) 표면 조사

조사 트리: `.claude/worktrees/t401` @ `ad272be20` (= origin/develop)
범위: `.claude/rules/moai/**`, `.claude/output-styles/moai/*.md`, `.claude/agents/moai/*.md`,
`.claude/skills/moai/**`, `.moai/config/sections/*.yaml`, `internal/config/*.go`
방식: read-only Explore fan-out. 총 77좌표 수집.

분류: **A = 권고 방출이 의무**, **B = 선택/허용**, **C = 이미 존재하는 반-anchoring 가드**

## A — 권고를 반드시 내라고 요구하는 좌표 (핵심)

| 좌표 | 절 | 표시 |
|---|---|---|
| `askuser-protocol.md:64` | 첫 옵션은 `(권장)`/`(Recommended)` 접미사를 **MUST** 달아 권장 선택을 신호한다 | 구조 제약 목록 "all mandatory"; zone-registry 에 Frozen 미러 |
| `askuser-protocol.md:217`, `:245`, `:253` | 절차 단계 · 완료보고 종결 규율 모두 "(권장) first" 를 못박음 | 무표시 |
| `zone-registry.md:869` (CONST-V3R5-035) | Skill body BODP 게이트가 `(권장)` first 를 따를 것 | `zone: Frozen`, `canary_gate: true` |
| `output-styles/moai/moai.md:451`, `:469` | Discovery 배너 `⏭️ Recommended action:` 필드 — "MUST be a single-line actionable directive" | `[HARD]` |
| `moai.md:521`, `:537`, `:555`, `:575` | Epic Stats / Epic Status `⏭️ Next:` — "MUST be concrete … never vague" | `[HARD]` |
| `moai.md:353-356` | Insight 배너 `What/Why/Alternatives/Implications` — 결정과 근거를 함께 방출, 사용자 판단 슬롯 없음 | 레이블 `[HARD]` 축자 |
| `moai.md:592-596` | Error Recovery `A. Retry as-is / B. Alt approach / …` — 순서 자체가 암묵 권고 | 무표시 |
| `plan-auditor.md:433-435` | `## Recommendation — {If FAIL: numbered, actionable fix instructions}` | 출력 계약 |
| `plan-auditor.md:427` | FAIL 시 결함목록이 "machine-consumable fix route" | 무표시 |
| `sync-auditor.md:83` | 각 finding 줄 안에 `- Required fix: {concrete, actionable fix instruction}` — **관측과 처방이 한 줄에 융합** | 리포트 템플릿 |
| `sync-auditor.md:86-87` | `### Recommendations` 절 | 리포트 템플릿 |
| `super-advisor.md:87-92` | 출력 계약 3번 = "Recommendation — the advisor's preferred option + why (non-binding)" | 출력 계약 |
| `manager-spec.md:172-175` | Step 6 에서 도메인 specialist 를 **recommend** | 무표시 |
| `run.md:97`, `:137` | run 완료 시 sync 체인을 "(Recommended)" 첫 옵션으로 표면화; run-phase 진입 라운드도 동일 | `:137` 은 `[HARD]` |
| `context-window-management.md:84` | `/clear` 를 자연어로 **권고** — AskUserQuestion 채널 **밖**이라 그 채널의 가드가 하나도 안 걸림 | 트리거 절차 |
| `ci-autofix-protocol.md:47`, `:95` | "1. (Recommended) Fix manually" / "apply (Recommended)" | 무표시 |
| `goal-directive.md:32-40` | §Proactive Recommendation Triggers — 요청 없이 선제 방출 | 무표시 |
| `session-handoff-examples.md:223`, `:119` | `--branch` 권고 · `ultrathink` 포함 계속 권고 | `:223` 은 `[ZONE:Evolvable][HARD]` |

## C — 이미 있는 반-anchoring 가드 (중복 구축 금지)

- `askuser-protocol.md:83` — 옵션 description 은 중립·사실 언어 **MUST**; 권고 신호는 `(권장)` 레이블 **하나로만** 전달
- `askuser-protocol.md:89-105` §Recommendation Placement Principles 5원칙 — 권장은 "관측된 다수 기본값"이지 시스템이 밀고 싶은 정책 기본값이 아니다; cold-start 는 공개 의무
- `askuser-protocol.md:103-105` **Adaptive strength** — 숙련도 높으면 레이블 **생략**. 현행에서 레이블을 억제할 수 있는 **유일한** 절
- `askuser-protocol.md:118` Report-Before-Ask `[HARD]` — 조사 기반 질문은 **같은 턴 안에서** 실질 보고가 선행
- `askuser-protocol.md:122` **Requested-Deliverable Primacy** `[HARD]` — 사용자가 보고를 요청했으면 그 턴은 보고로 끝내고 결정 질문을 붙이지 않는다 (현행 최강 분리)
- `askuser-protocol.md:136` preview-as-report 치환 금지
- `verification-claim-integrity.md:9`, `:33` — **권고의 전제**도 미관측 주장 금지 대상. 사용자 지시에 반대하는 권고는 결함 주장과 동일한 증거 부담
- `sync-auditor.md` finding-consumption discipline — 발견 단계에선 전부 보고, 브레이크는 **소비 단계**(오케스트레이터)
- `manager-lead.md:207`, `:215` / `kanban-dispatch.md:100` — deputy 는 `RECOMMEND:` 접두사만, `FINAL VERDICT:` 토큰 금지
- `super-advisor.md:34-36` — non-binding 처방과 binding 판정의 역할 분리 필수

## 명시 답변 1 — 사용자의 독립 판단을 AI 권고보다 **먼저** 요구하는 규칙이 있는가

**없음.** 조사 범위 전체에서 0건.

가장 근접한 4개 절 전부 **증거**를 권고보다 앞세울 뿐 **사용자 판단**을 앞세우지 않는다:

1. Report-Before-Ask(`:118`) — 보고가 선행하지만 **같은 턴**이다. 사용자의 첫 행위는 이미 레이블이
   붙은 옵션 중 고르는 것.
2. Requested-Deliverable Primacy(`:122`) — 유일하게 턴을 끊지만, **사용자가 보고를 명시 요청했을 때만**
   발동한다. 그리고 같은 조사 출력에 `[HARD]` 로 요구되는 Discovery 배너의
   `⏭️ Recommended action:` 필드를 억제하지 못한다.
3. Adaptive strength(`:103-105`) — 레이블을 뺄 수 있는 유일한 통로지만 기준이 **추정 숙련도**이지
   사용자 판단이 아니다.
4. Implementation Kickoff Approval — 승인 게이트이되, **이미 권고가 붙은 선택지**에 대한 승인이다.

grep 스윕: `anchor(ing)`, `independent judg(e)ment`, `withhold`, `push, not pull`, `pull, not push`,
`unprompted`, `solicit`, `on request only`, `only when asked`, `권장|recommend|⏭️|prescri|non-binding`
→ judgment-first / recommendation-withholding 구성물 **0히트**.

## 명시 답변 2 — 오늘 권고 방출을 껐다 켤 수 있는 config 키가 있는가

**없음.** `.moai/config/sections/*.yaml` 33개 + `internal/config/*.go` 전수.

근접 키는 전부 다른 것을 통제한다:

| 키 | 실제 통제 대상 |
|---|---|
| `learning.auto_apply` (harness.yaml:180) | Tier-4 제안의 **자동 적용** 여부. false 면 권고가 오히려 **더** 노출됨 |
| `interview.enabled` / `plan.max_rounds` / `skip_conditions` | 라운드 **개수**. 라운드를 없애면 그 레이블도 사라지지만 권고 통제는 아님 |
| `AuditGateOff/Advisory/Required` (internal/config/audit_models.go:36-52) | 감사가 **차단**하는지 여부. `sync-auditor` 의 `### Recommendations` 는 무조건 방출 |
| `gate.yaml` advisory · `sunset.yaml action: advisor` | 품질 게이트 차단 모드 |
| `ralph.yaml:39,51` | 확인 없이 수정 자동적용 |
| `mcp-matrix.yaml` | 정적 추천 매트릭스(데이터, 토글 아님) |

`⏭️ Recommended action:` · `⏭️ Next:` · `(권장)` 첫 옵션 레이블 · super-advisor 의 `Recommendation`
블록 · 감사자의 `Required fix:` / `Recommendations` 절 — **어느 것도 끌 수 없고 대부분 `[HARD]`**.
