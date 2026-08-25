# Plan — SPEC-LSEL-DRAIN-STALL-001

> Tier M (3 artifacts + progress.md). Milestone 2개. 카드 t259 (Class B — 원인은 plan-phase에서 확정: 트리거 부재).
> 측정 기반선: worktree `WT-lsel-drain-stall` HEAD `a739d04b4` (2026-08-25). live-state 수치는 primary 체크아웃 runtime state에서 같은 날 측정(spec.md §B.1).

## §A. Context

### §A.1 무엇을 만드나

1. **`session_drain.sh`** (신규, tracked): flock 배타잠금 → clusters.json 후보 보존(`clusters-history/`) → `drain.sh` 실행 → 1행 상태 출력. drain.sh는 무수정(Constraint 7).
2. **`session_drain_test.sh`** (신규, tracked): fixture 기반 wrapper 테스트 — drain 경로 / 잠금 경쟁 / 보존 / no-op / fail-open + **mutant probe**(오프셋만 전진하는 가짜를 넣고 검증 predicate가 이를 걸러내는지).
3. **위생 수정**: `lsel-drain-loop.js` 주장-본문 일치화(REQ-LDS-010), `backlog_check.sh` 죽은 §28 앵커 제거 + wrapper 지시로 교체, `hns-lsel-curator/SKILL.md` 내구-운영 섹션 미러.
4. **로컬 인도물**(PR 미탑재, spec.md §E): settings.local.json SessionStart 배선 2건 + CLAUDE.local.md LSEL 섹션.
5. **M2 백로그 일괄 드레인**: 유지자 머신에서 wrapper로 live 3.5k+ 백로그 소진 + SKILL.md §Verification 레시피로 검증(AC-LDS-010).

### §A.2 Tier M 근거 (카드는 S~M)

- AC 12개 > Tier S 상한 8 — 예산만으로도 M.
- 검증 표면 2종: tracked 코드(wrapper/테스트/위생 — PR로 감사) + 유지자 머신 적용(배선·live 드레인 — §E.2 증거로만 감사). 이 이중 표면이 Tier S의 단일 위임 프롬프트에 안 맞는다.

### §A.3 결정 원장 (가역성 내림차순 — 파토 때낼 확률이 높은 순서)

| # | 결정 | 디폴트 | 기각안 | 근거 |
|---|---|---|---|---|
| D1 | 백로그 방침 | **일괄 드레인** (bulk, offset을 live 끝까지) | 오프셋 리셋 후 최근분만 — 기각(OutOfScope) | mutant-인접 + 측정 부하 사소(§B.4: 3.5k행 클러스터링 dry-run rc=0, 후보 13-15개) |
| D2 | 배선 위치 | settings.local.json (로컬, PR 외) | tracked settings.json — **반드시 기각** | `moai update`가 `.claude/settings.json`을 통째 재배포(§2.3) — 배선이 매 업데이트 유실. 정직한 문서화가 유일한 지속 수단 |
| D3 | loop 레시피 처분 | **수정 유지** (주장=본문 일치화 + wrapper 참조) | 삭제(retire) — 기각 | (a) SessionStart는 세션당 1회 — 긴 세션의 중반 적재는 루프 리마인더가 유일한 세션 내 커버, (b) 최소 diff, (c) 삭제는 문서화된 수동 탈출로를 잃음 |
| D4 | clusters.json 유실 방지 | wrapper에서 **호출-전 무조건 보존** (history 디렉터리) | drain.sh no-op 경로 수정 — 기각 | drain.sh 무수정 원칙(기존 characterization 초록 유지) + no-op 경로(63-76행)조차 덮어쓰므로 조건부 보존은 구멍 |
| D5 | doctor/statusline 신호 | 본 카드 비목표 | 편입 — 기각 | 분산 표면 + dev-local 상태. 후속 카드 후보로 기록 |

### §A.4 Kickoff 확인 사항 (Implementation Kickoff Approval에서 운영자 선택 — 디폴트는 이미 REQ로 굳음)

D1(일괄 vs 최근분), D2(로컬 배선 방침), D3(레시피 수정 vs 삭제), D5(doctor 비목표 유지). 전항 디폴트 권장.

### §A.5 PRESERVE 목록 (무수정 — AC-LDS-011로 증명)

`internal/harness/applier.go` · `internal/harness/curator_dispatch.go` · `internal/template/templates/**` (전체) · `.claude/rules/moai/**` · `CLAUDE.md` · `.claude/skills/hns-lsel-curator/drain.sh` · `.claude/skills/hns-lsel-curator/drain_test.sh` (신규 테스트는 별도 파일) · `.moai/lessons-inbox.jsonl` (append-only, 불변) · `memory/` (무쓰기) · `.moai/config/sections/**` (신규 파일 금지)

### §A.6 도달 파일 (PR 탑재)

신규: `.claude/skills/hns-lsel-curator/session_drain.sh`, `.claude/skills/hns-lsel-curator/session_drain_test.sh`. 수정: `.claude/workflows/lsel-drain-loop.js`, `.claude/skills/hns-lsel-curator/backlog_check.sh`, `.claude/skills/hns-lsel-curator/SKILL.md`, 본 SPEC 4건. — PR 총 7파일 + SPEC.

## §B. Known Issues

- **B-analog (워크트리 가드)**: wrapper 테스트는 `mktemp -d` fixture + `trap rm -rf` (기존 drain_test.sh 패턴) — 복합 명령 거부·부하 규율 회피. 세션 훅 시간 예산: 테스트에서 wrapper 런타임 상한 확인(REQ-LDS-005).
- **flock 이식성**: macOS/BSD `flock`은 Linux와 flag 호환성이 다를 수 있음 — `flock`(util-linux) 부재 시 `mkdir` 원자잠금 폴백 또는 `shlock` 검토. run-phase에서 첫 구현 시 실제 동작 기준 선택(AC-LDS-002가 동작을 검증, 구현 수단은 자유).
- **backlog_check.sh의 오프셋 파싱**: `grep -o` 기반 파싱은 drain-offset.json 형식 가정 — wrapper/drain이 쓰는 형식과 불일치 없음 확인(기존 backlog_check_test.sh 초록 유지).
- **M2 live 드레인은 유지자 머신에서만**: 워크트리엔 live 인박스가 없다(untracked runtime state). M2 실행·증거는 primary 체크아웃 대상 — AC-LDS-010 증거는 progress.md §E.2에 명령+출력으로.

## §C. Pre-flight (run-phase M1 착수 전)

```bash
git rev-parse --short HEAD && git branch --show-current   # 트리 재확인
bash .claude/skills/hns-lsel-curator/drain_test.sh          # 기존 characterization 초록 확인
bash .claude/skills/hns-lsel-curator/backlog_check_test.sh  # 동일
grep -rn "session_drain" .claude/skills/hns-lsel-curator/ | wc -l   # 0 (RED-now 확인)
grep -c "The recipe runs drain.sh" .claude/workflows/lsel-drain-loop.js  # 1 (RED-now)
```

live-state 기반선(유지자 머신, M2 직전 재측정 — moving target): `wc -l < .moai/lessons-inbox.jsonl` / `jq .offset .moai/state/lsel/drain-offset.json`.

## §D. Constraints (spec.md §D와 동일 — 원천은 spec.md)

템플릿 미러링 금지(CI guard 3종 초록 유지) · applier/curator_dispatch 무수정 · rules/CLAUDE.md 무수정 · `.moai/config/sections/` 신규 금지 · `.moai/state/lsel/` 외 쓰기 금지 · 인박스 불변 · drain.sh 무수정 · plan-phase 중 live 실행 금지.

## §E. Self-Verification (run-phase 보고 형식)

각 항목 VCI 5-섹션(Claim/Evidence/Baseline-attribution/Gaps/Residual-risk) + attribution triple (a)명령 (b)관측 출력 (c)측정 HEAD. 로컬 인도물은 "적용됨"이 아니라 적용 후 관측(jq/grep 출력)을 증거로.

## §F. Milestones

> 순서는 가역성 원장(§A.3) 순서와 정렬: 정책-민감 M2를 먼저 서술하되 **실행 순서는 M1 → M2** (wrapper 없이는 안전한 백로그 드레인·검증이 없다).

### M2 — 백로그 일괄 드레인 + 검증된 완료 + 로컬 인도물 적용 (정책-민집: D1/D2 실현)

우선순위 High. AC-LDS-010 (MUT — mutant guard), AC-LDS-011 (MUST), AC-LDS-012 (SHOULD, 적용 부).

1. 유지자 머신에서 로컬 인도물 적용: settings.local.json SessionStart 2항(wrapper + backlog_check), CLAUDE.local.md LSEL 섹션 복원. — 적용 후 `jq '.hooks.SessionStart' .claude/settings.local.json` 관측을 §E.2에.
2. live 백로그 재측정 후 wrapper로 일괄 드레인 실행 (kickoff 승인 이후만).
3. SKILL.md §Verification 레시피 + 자기일관성 검증: `offset == live wc -l` && `candidates >= 1` && `offset_after == live` && `total_read == live − offset_before`. mutant probe: 세션 테스트의 오프셋-전진-가짜가 이 predicate에서 FAIL함을 fixture로 입증(AC-LDS-005에서 구축, 여기서 live 적용).
4. memory/ 무쓰기 + 인박스 불변 + PRESERVE diff 증명 (AC-LDS-011).

### M1 — 내구 트리거 wrapper + 정지 신호 + 위생 (기계적 기반)

우선순위 High. AC-LDS-001..009 (전 MUST) + AC-LDS-012 (문서화 부, SHOULD).

1. `session_drain.sh` 작성 (TDD: 테스트先行 — session_drain_test.sh RED부터): flock → 무조건 보존 → drain.sh 호출 → 상태 1행 → fail-open.
2. 테스트 5경로 + mutant probe (AC-LDS-005): drain / 경쟁 / 보존 / no-op / fail-open / offset-only-advance mutant가 검증 predicate를 통과 못함.
3. 위생: lsel-drain-loop.js 헤더·본문 교정(실행 주장 제거 + session_drain.sh 참조), backlog_check.sh §28 앵커 제거·리마인더 텍스트 교체, SKILL.md 내구-운영 섹션(트리거·검증·로컬 배선 안내 — 로컬 항목은 "유지자 머신 적용"으로 명시) 미러.
4. 기존 테스트 회귀: drain_test.sh + backlog_check_test.sh 초록 유지.
5. 로컬 인도물 문서화(spec.md §E 확정 — plan-phase에서 이미 완료 상태로 반영).

실행 순서: **M1 → M2** (M2가 M1의 wrapper·predicate에 의존).

## §G. Anti-Patterns

- **AP-LDS-001 (카드 명명 mutant)**: 오프셋만 최신으로 밀고 클러스터링은 안 하는 구현 — AC-LDS-005 mutant probe + AC-LDS-010 3중 조건(offset==live && candidates≥1 && 자기일관)으로 봉쇄.
- **AP-LDS-002**: tracked settings.json에 SessionStart 배선 — 매 `moai update`마다 유실(§2.3). 배선은 settings.local.json + 문서(REQ-LDS-009).
- **AP-LDS-003**: drain.sh 수정로 clusters.json 유실 방지 시도 — no-op 경로 조건 분기를 놓치기 쉽다; wrapper의 호출-전 무조건 보존이 닫는다.
- **AP-LDS-004**: wrapper가 세션 시작을 막는 형태 (unbounded 동기 드레인, 에러 시 non-zero) — advisory-check discipline 위반(REQ-LDS-005).
- **AP-LDS-005**: 백로그 드레인을 "나중에 조금씩" 분할 — 3주 정지를 재현하는 지름길. 일괄(D1)이 디폴트.
- **AP-LDS-006**: 템플릿 미러링/내부 토큰 유출 — CI guard 3종이 잡지만 PR 전 자가점검(`git diff --stat <merge-base> -- internal/template/templates` 빈 출력).

## §H. Cross-References

- spec.md (요구·제약·로컬 인도물 원천) · acceptance.md (AC·RED-now/green-path 셀)
- `.moai/specs/SPEC-LSEL-LOCAL-EVOLUTION-001/plan.md` (6개 가용 표면·AP-LSEL 계보 — AP-LDS 번호는 그 전통 계승)
- `.claude/rules/moai/development/coding-standards.md` § Advisory-Check Discipline · `.claude/rules/moai/core/verification-claim-integrity.md` §3 (5-섹션 보고)
- `.claude/rules/moai/development/verification-completeness.md` §1.1/§2 (mutant probe·two-cell 채택 규율 — AC 셀 구성의 근거)
