# t82 — 착수 시점 재측정 + 전제 검증

베이스: `origin/main` 기준 워크트리 `.claude/worktrees/t82` · 브랜치 `WT-agents-md-diet` · 측정일 2026-08-22

카드 지시("136,345 B 수치는 작업 시작 시 재측정")에 따른 실측. **카드 수치는 스테일이며, 실제 표면은 더 크다.**

---

## 1. always-loaded 표면 실측

측정 표면의 정의는 추측이 아니라 리포에 이미 존재하는 가드가 소유한다 —
`internal/config/token_budget_guard.go` `alwaysLoadedSurface()`:
`paths:` frontmatter 없는 `.claude/rules/moai/**/*.md` 전부 + 고정 3슬롯
(`CLAUDE.md`, `.claude/output-styles/moai/moai.md`, 리포 루트 `MEMORY.md`).

명령: `grep -rLE '^paths:' --include='*.md' .claude/rules | sort | xargs wc -c`

| 파일 | 바이트 |
|---|---:|
| core/agent-common-protocol.md | 27,043 |
| kanban-dispatch.md | 25,915 |
| core/askuser-protocol.md | 23,504 |
| session-handoff.md | 21,197 |
| core/moai-constitution.md | 18,958 |
| cross-session-messaging.md | 16,672 |
| core/verification-claim-integrity.md | 13,140 |
| context-window-management.md | 13,009 |
| main-checkout-branch-guard.md | 11,865 |
| core/moai-mcp-tools.md | 7,357 |
| goal-directive.md | 6,600 |
| cache-aware-execution.md | 6,569 |
| skill-routing.md | 5,825 |
| core/native-idiom-and-register.md | 4,967 |
| **규칙 소계 (14파일)** | **202,621** |

| 고정 슬롯 | 바이트 |
|---|---:|
| `CLAUDE.md` | 20,523 |
| `.claude/output-styles/moai/moai.md` | 61,706 |
| 리포 루트 `MEMORY.md` | 0 (부재 — 가드가 hermetic 하게 0 처리) |
| **총 표면** | **284,850** |

토큰 추정(가드의 `char/4`): **약 71,212 토큰** · 예산 `AlwaysLoadedTokenBudget = 76,000` · 여유 약 4,788.
`go test ./internal/config/ -run 'Budget|AlwaysLoaded'` → PASS (현 트리는 예산 내).

### 카드 수치와의 차이 (필수 정정)

| 항목 | 카드 기재 | 실측 | 차이 |
|---|---:|---:|---|
| always-loaded 룰 | 136,345 B | **202,621 B** | +66,276 B (+48.6%) |
| CLAUDE.md | 19,040 B | **20,523 B** | +1,483 B |
| 재설계 문서 표기 | "~210 KB" | 202,621 B | 재설계 문서 쪽이 실측에 가깝다 |
| **출력 스타일** | **미계상** | **61,706 B** | 카드가 통째로 누락한 표면 |

카드의 두 수치를 그대로 쓰면 다이어트 대상 규모를 **약 34% 과소평가**한다.
특히 `.claude/output-styles/moai/moai.md` 61,706 B 는 카드의 3갈래 재배치 표에 등장하지 않는데,
always-loaded 표면 중 **단일 최대 파일**이다(2위 agent-common-protocol.md 의 2.3배).

## 2. 32 KiB 한도 대비 격차

Codex `project_doc_max_bytes` 기본 32,768 B 기준:

- 현재 표면 284,850 B → **8.7배 초과**
- 카드 수치 기준이면 155,385 B → 4.7배. 즉 카드는 격차를 절반 이하로 오인하고 있었다.

## 3. codex 바이너리 실측 (0.147.0) — 카드 전제의 부분 확인

`which codex` → `/Users/goos/.local/bin/codex` · `codex --version` → `codex-cli 0.147.0`

바이너리 심볼 확인 결과 다음이 **실재**한다:

- config 키 `project_doc_max_bytes`, `project_doc_fallback_filenames` (`ConfigToml` 96필드에 포함)
- 문서 파일명 우선순위 토큰 `AGENTS.override.md`, `AGENTS.md`
- 로더 모듈 `core/src/agents_md.rs` 와 그 로그 메시지:
  **`project doc exceeds remaining budget; truncating`** (필드 `remaining_bytes`)

### 3.1 설계에 직접 영향을 주는 발견 — 예산은 문서별이 아니라 **잔여 공유**다

로그 문구가 `remaining budget` / `remaining_bytes` 다. 이는 한도가 파일 하나마다 32 KiB 인 것이 아니라,
**병합 체인 전체가 하나의 잔여 예산을 나눠 쓰고 먼저 읽힌 문서가 예산을 소진**함을 시사한다.

카드의 배치안("루트 ~8 KiB + 영역별 중첩 각 ~4 KiB")은 이 해석과 **양립한다** —
합이 32 KiB 를 넘지 않는 한. 다만 "중첩 문서마다 4 KiB 씩 별도로 더 쓸 수 있다"로 읽으면 틀린다.
중첩 문서 6개를 두면 8 + 6×4 = 32 KiB 로 정확히 한도이며, 그 이상은 뒤쪽이 잘린다.

**미검증(중요)**: 잘리는 쪽이 앞인지 뒤인지, 그리고 중첩 문서가 CWD 체인에 한해 병합되는지
(=리포 루트에서 codex 를 띄우면 하위 영역 문서는 아예 안 읽히는지)는 심볼만으로 판정할 수 없다.
이것은 t91(M0)의 실측 항목이다.

## 4. t91 (M0) 상태 — run 진입 전 해소 필요

`.moai/reports/t91/` **부재**. M0 실측 산출물이 아직 없다.

재설계 문서 자신의 § 미검증(Gaps)이 "전부 공식 문서 기준, 실제 바이너리 실행 검증 없음"이라고 적는다.
t82 의 설계는 다음 3개 전제 위에 서 있고, 셋 다 현재 문서-신뢰 단계다:

1. `project_doc_max_bytes` 기본값이 실제로 32,768 B 인가 (§3 은 키의 존재만 확인했다)
2. 중첩 AGENTS.md 가 **어느 범위에서** 병합되는가 (프로젝트 루트→CWD 하향 체인인지, 변경 파일 기준인지)
3. 초과분이 잘릴 때 **경고가 사용자에게 보이는가** — 조용히 잘리면 다이어트 회귀가 무음으로 죽는다

3번은 특히 회귀 가드 설계를 좌우한다. 조용히 잘린다면 CI 측 바이트 가드가 유일한 방어선이 된다.

## 5. SPEC-ALWAYS-LOADED-DIET-001 과의 관계

`.moai/specs/SPEC-ALWAYS-LOADED-DIET-001/` 은 이미 **3-phase close 완료**(2026-08-17, PR #1576/#1577/#1578).
따라서 "통합 처리"는 그 SPEC 을 재개방하는 것이 아니라,
**그것이 남긴 예산 가드(`AlwaysLoadedTokenBudget`)와 스텁+지연 companion 패턴을 t82 가 계승**하는 형태여야 한다.

특히 `token_budget_guard.go:24-31` 주석이 예산 상향 75,000 → 76,000 의 사유로
"kanban-dispatch 등 대형 always-loaded 룰의 스텁+지연 로딩 다이어트는 별도 카드로 진행"을 명시한다 —
**t82 가 그 별도 카드다.** 즉 t82 의 성공 판정에는 예산 상수를 76,000 에서 되돌리는(래칫) 항목이 들어가야 한다.

## 6. 판정 요약

| 항목 | 결과 |
|---|---|
| 카드 수치 재측정 | **정정 필요** — 룰 202,621 B, CLAUDE.md 20,523 B, 출력스타일 61,706 B 누락 |
| 32 KiB 대비 격차 | 8.7배 (카드 전제로는 4.7배로 과소평가) |
| codex 0.147.0 전제 | 키·로더·우선순위 **존재 확인**, 기본값/병합범위/경고여부 **미검증** |
| 선행 카드 t91(M0) | **산출물 부재** — run 진입 전 해소 필요 |
| DIET-001 관계 | 재개방 아님 — 가드 계승 + 예산 상수 래칫이 t82 범위 |

## 미검증 / 잔여

- codex 실바이너리로 AGENTS.md 를 실제 로드시키는 end-to-end 확인은 하지 않았다(모델 호출 비용 + M0 소관).
  본 보고의 codex 관련 주장은 전부 **바이너리 심볼 관측**에 한정된 근거다.
- `.claude/agents`·`.claude/skills` 등 조건부 로드 표면은 측정 범위 밖이다(always-loaded 아님).
- 사용자별 auto-memory(`~/.moai/claude-profiles/**/MEMORY.md`)는 가드가 의도적으로 제외한다(hermetic).
  실제 세션에서는 이 또한 컨텍스트를 차지하므로, 체감 부담은 위 수치보다 크다.
