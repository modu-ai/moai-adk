# Harness Learning Subsystem

## Overview

Harness Learning Subsystem은 `/moai` 서브커맨드, 에이전트 호출, SPEC 참조 패턴을 관찰하여 SKILL.md frontmatter를 자동으로 개선하는 학습 파이프라인이다. PostToolUse hook을 통해 이벤트를 수집하고, 관찰 횟수가 임계값에 도달하면 5-Layer Safety Architecture를 거쳐 사용자 승인 후 적용된다.

이 subsystem은 `.moai/config/sections/harness.yaml`의 `learning:` 섹션으로 제어되며, `learning.enabled: false`로 완전히 비활성화할 수 있다.

관련 SPEC: SPEC-V3R3-HARNESS-LEARNING-001

---

## CLI Reference (`/moai harness`)

모든 `/moai harness` 서브커맨드는 `moai harness <verb>` 형식으로 실행된다.

### `status`

현재 학습 subsystem 상태와 tier 분포를 출력한다.

```bash
moai harness status
```

예시 출력:
```
Harness Learning Status
  Enabled:        true
  Log entries:    247
  Patterns:       38

Tier Distribution:
  observation:    28 patterns
  heuristic:       5 patterns
  rule:            3 patterns
  auto_update:     2 patterns (Tier 4 — 사용자 승인 대기 중)

Pending Proposals: 2
  [PENDING] prop-2026-04-27-001: .claude/skills/my-harness-x/SKILL.md (description)
  [PENDING] prop-2026-04-27-002: .claude/skills/my-harness-y/SKILL.md (triggers)
```

### `apply`

Tier 4에 도달한 패턴에 대한 업데이트 제안을 5-Layer Safety Pipeline을 통해 적용한다.

```bash
moai harness apply
```

플로우:
1. `cmdApply()` — 대기 중인 Tier 4 제안 목록 조회
2. `safety.Pipeline.Evaluate()` — L1→L2→L3→L4→L5 순차 평가
3. `DecisionPendingApproval` 반환 시 → orchestrator가 `AskUserQuestion`으로 사용자에게 승인 요청
4. 사용자 승인 후 → `Applier.Apply()` 실행 (스냅샷 생성 → frontmatter 수정)

`learning.auto_apply: true`로 설정하면 L5 사용자 승인 단계가 생략된다.

### `rollback <YYYY-MM-DD>`

지정된 날짜의 스냅샷으로 파일을 복원한다.

```bash
moai harness rollback 2026-04-27
```

날짜에 해당하는 스냅샷 디렉토리를 `.moai/harness/learning-history/snapshots/` 에서 탐색하여 `RestoreSnapshot(snapshotDir)`을 호출한다. 복원 시 원본 파일이 byte-identical하게 복구된다.

날짜가 중복될 경우 (같은 날 여러 스냅샷) 가장 최근 스냅샷을 사용한다.

### `disable`

학습 subsystem을 비활성화한다. `.moai/config/sections/harness.yaml`의 `learning.enabled`를 `false`로 변경한다.

```bash
moai harness disable
```

비활성화 후:
- `Observer.RecordEvent()` — no-op (로그 파일 미생성)
- `Applier.Apply()` — no-op (evaluator 미호출, 파일 미수정)
- 기존 학습 데이터는 보존됨

재활성화: `learning.enabled: true`로 수동 변경.

---

## Configuration (`.moai/config/sections/harness.yaml` `learning:`)

| Key | Default | Description |
|-----|---------|-------------|
| `enabled` | `true` | observer + learner 활성화 여부. `false`이면 완전 no-op |
| `auto_apply` | `false` | Tier 4 도달 시 자동 적용 여부. `false`이면 AskUserQuestion으로 승인 요청 |
| `tier_thresholds` | `[1, 3, 5, 10]` | observation/heuristic/rule/auto_update 임계값 |
| `rate_limit.max_per_week` | `3` | 7일 슬라이딩 윈도우 내 최대 자동 업데이트 횟수 |
| `rate_limit.cooldown_hours` | `24` | 업데이트 간 최소 대기 시간(h) |
| `log_retention_days` | `90` | `usage-log.jsonl` 보존 일수. 초과분은 월별 gzip 아카이브로 이동 |

**설정 예시:**

```yaml
learning:
  enabled: true
  auto_apply: false
  tier_thresholds: [1, 3, 5, 10]
  rate_limit:
    max_per_week: 3
    cooldown_hours: 24
  log_retention_days: 90
```

---

## Tier Ladder

패턴 관찰 횟수와 신뢰도(confidence)에 따라 Tier가 결정된다.
신뢰도 < 0.70이면 횟수에 관계없이 `observation`으로 강제된다.

| Tier | 임계값 | 이름 | 시스템 동작 |
|------|--------|------|-------------|
| 1 | count >= 1 | `observation` | 로그에만 기록, 아무 동작 없음 |
| 2 | count >= 3 | `heuristic` | `EnrichDescription()`: description 필드에 heuristic note 추가 |
| 3 | count >= 5 | `rule` | `InjectTrigger()`: triggers 목록에 키워드 추가 |
| 4 | count >= 10 | `auto_update` | 5-Layer Safety Pipeline 통과 후 사용자 승인 → 자동 적용 |

---

## Snapshot Locations

학습 subsystem이 생성하는 파일 경로:

| 파일/디렉토리 | 경로 | 설명 |
|---------------|------|------|
| 이벤트 로그 | `.moai/harness/usage-log.jsonl` | PostToolUse hook이 기록하는 JSONL 로그 |
| Tier 승격 로그 | `.moai/harness/learning-history/tier-promotions.jsonl` | tier 승격 이벤트 기록 |
| 스냅샷 | `.moai/harness/learning-history/snapshots/<ISO-DATE>/` | Apply() 직전 원본 파일 백업 |
| 스냅샷 manifest | `.moai/harness/learning-history/snapshots/<ISO-DATE>/manifest.json` | 스냅샷 메타데이터 |
| 위반 로그 | `.moai/harness/learning-history/frozen-guard-violations.jsonl` | L1 Frozen Guard 차단 기록 |
| Rate limit 상태 | `.moai/harness/learning-history/rate-limit-state.json` | L4 슬라이딩 윈도우 상태 |
| 월별 아카이브 | `.moai/harness/learning-history/archive/<YYYY-MM>.jsonl.gz` | log_retention_days 초과 이벤트 |

---

## Rollback Procedure

학습 subsystem이 파일을 잘못 수정했을 때 복원하는 절차:

**Step 1: 스냅샷 목록 확인**

```bash
ls .moai/harness/learning-history/snapshots/
# 출력 예시:
# 2026-04-27T10-30-05.123456789Z/
# 2026-04-26T15-20-10.987654321Z/
```

**Step 2: 해당 날짜 스냅샷 내용 확인**

```bash
cat .moai/harness/learning-history/snapshots/2026-04-27T10-30-05.123456789Z/manifest.json
# 출력 예시:
# { "proposal_id": "prop-2026-04-27-001", "created_at": "...", "files": [...] }
```

**Step 3: CLI로 롤백 실행**

```bash
moai harness rollback 2026-04-27
```

또는 Go 코드로 직접 실행:

```go
snapshotDir := ".moai/harness/learning-history/snapshots/2026-04-27T10-30-05.123456789Z"
if err := harness.RestoreSnapshot(snapshotDir); err != nil {
    log.Fatalf("롤백 실패: %v", err)
}
```

**Step 4: 파일 복원 확인**

```bash
git diff .claude/skills/my-harness-x/SKILL.md
# 변경 없음 = 롤백 성공
```

**Step 5: 학습 비활성화 (선택)**

롤백 원인이 해결되기 전에 동일 패턴이 다시 적용되지 않도록:

```bash
moai harness disable
```

---

## Frozen Zones (5-Layer Safety L1)

다음 경로 접두사는 **절대 자동 업데이트 불가**한 FROZEN 영역이다. configuration이나 환경변수로 변경할 수 없다.

| 경로 접두사 | 이유 |
|-------------|------|
| `.claude/agents/moai/` | MoAI 핵심 에이전트 정의 — 임의 변경 시 오케스트레이션 파괴 |
| `.claude/skills/moai-` | MoAI upstream 스킬 — `moai update`로만 동기화되어야 함 |
| `.claude/rules/moai/` | MoAI 규칙 파일 — 헌법 수준의 불변 규칙 |
| `.moai/project/brand/` | 브랜드 컨텍스트 — 헌법 제약으로 인한 수동 변경만 허용 |

FROZEN 경로 접근 시도는 `.moai/harness/learning-history/frozen-guard-violations.jsonl`에 기록되며 stderr에 경고가 출력된다.

**사용자 생성 스킬** (`.claude/skills/my-harness-*/`, `.claude/skills/project-*/` 등)은 FROZEN 영역이 아니므로 자동 업데이트가 허용된다.

---

## Disabling Learning

학습 subsystem을 완전히 비활성화하려면:

**방법 1: CLI 사용**

```bash
moai harness disable
```

**방법 2: 설정 파일 직접 편집**

`.moai/config/sections/harness.yaml`:

```yaml
learning:
  enabled: false
```

비활성화 후 동작:
- `Observer.RecordEvent()` — 즉시 반환, 로그 파일 미생성/미기록
- `Applier.Apply()` — 즉시 반환, evaluator 미호출, 파일 미수정, 스냅샷 미생성
- 기존 학습 데이터(`.moai/harness/`)는 보존됨 (삭제되지 않음)
- 재활성화 시 기존 데이터 기반으로 학습 재개

---

## T-P5-10 Post-Merge Trial (Out of Scope)

T-P5-10 (1주 실사용자 trial)은 머지 후 사용자 검증 단계이다.

**검증 절차:**
1. 브랜치 머지 후 7일간 일반 개발 세션 진행
2. 7일 후 `moai harness status` 출력 기록
3. `frozen-guard-violations.jsonl` 위반 횟수 확인
4. `tier-promotions.jsonl` 승격 이벤트 검토
5. Tier 4 제안이 있으면 `moai harness apply`로 승인 플로우 검증
6. `acceptance.md` Definition of Done 체크리스트 완료

---

Version: 1.0.0
SPEC: SPEC-V3R3-HARNESS-LEARNING-001 (Phase 5, T-P5-08)
Last Updated: 2026-04-27
