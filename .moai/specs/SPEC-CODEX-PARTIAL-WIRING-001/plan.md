# SPEC-CODEX-PARTIAL-WIRING-001 — 구현 계획

> 카드 t499 · Tier M · 대상 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t499` (branch `WT-codex-partial-wiring`, base `ace1c5440`)

## §A. 맥락

레인이 이미 재현을 끝냈다(`.moai/reports/t499/repro.md`). 이 계획은 재현을 다시 하지 않고,
그 위에 **판별식 한 곳의 확장**과 그것을 잡는 시험을 얹는다.

## §B. 먼저 정할 것 (되돌리기 어려운 순서대로)

아래 3개가 이 카드에서 사람이 실제로 검토해야 할 결정이고, 나머지는 기계적이다.

### B-1. 세 번째 상태를 판별식에 어떻게 들여놓는가 [결정 필요도 高]

현행은 2치(`wired` / not) + `codexInstalled` 조합이다. 여기에 "agent 정의가 있는가"를 넣으면 상태가 세 갈래가 된다.

- 제안: `wired`(배선 파일 ≥1) → 기존 경로 그대로 / `half-wired`(배선 0 + agent 정의 ≥1) → 신규 두 갈래 / `unwired`(둘 다 0) → 기존 두 갈래 그대로
- 되돌리기 난이도: 이 구조가 뒤에 오는 모든 문구·시험의 뼈대다. 여기서 갈래를 잘못 나누면 시험을 전부 다시 쓴다

### B-2. 두 갈래의 상태값(Warn / OK) [결정 필요도 高 — 사용자 체감 직결]

`spec.md` §D-1이 근거다. **codex 존재 → Warn, codex 부재 → OK**. plain init이 모든 프로젝트에 agent TOML을
깔기 때문에, 부재 갈래까지 Warn으로 올리면 Codex를 안 쓰는 모든 사용자의 모든 프로젝트에 경고가 상시로 뜬다.
이 판단이 뒤집히면 AC-CPW-002가 통째로 바뀐다 — run-phase 안에서 임의로 뒤집지 말고 리드에게 올린다.

### B-3. 새 Message 문구 [결정 필요도 中 — 사용자에게 보이는 유일한 표면]

제약 3개: 113 runes 이하(REQ-CPW-008), `claude-only` 금지(REQ-CPW-004), Warn 갈래는 지시문을 Message에
싣는다(REQ-CPW-002 — Detail은 `--verbose`에서만 렌더된다). 부재 경로 인용 같은 증거는 Detail로 내린다.

**문구 결정이 어떻게 고정되는가(0.3.1 추가).** 문구를 고르는 것은 여전히 run-phase의 사람 결정이지만,
고른 즉시 **부재 갈래 Message는 시험 파일의 리터럴과 등가로 못박힌다**(AC-CPW-002 (d)). 이후 문구를
다듬으면 시험이 함께 깨지는 것이 정상이다 — 사용자에게 보이는 문구가 바뀌는데 시험이 안 깨지는 쪽이 나쁘다.
검토자가 읽어야 할 지점이 체크리스트 항목이 아니라 diff에 보이는 코드 한 줄이 된다.

### B-4. agent 정의 경로 상수를 어디에 두는가 [결정 필요도 低]

`.codex/agents`는 현재 `internal/template` 계열의 시험에만 문자열로 등장하고, `internal/codexwiring`에는
상수가 없다(`HooksRelPath` / `ConfigRelPath` / `SidecarPath`만 있음). 형제 상수들 옆에 `AgentsRelPath`를
더하는 쪽이 자연스럽다. 단 `RefreshWiring` 동작에는 손대지 않는다(AC-CPW-008).

## §C. 착수 전 확인

- [ ] `git -C . rev-parse --short HEAD` / `git branch --show-current` 재판독
- [ ] `.moai/reports/t499/repro.md` 재독 (측정 전제 인용용)
- [ ] `internal/cli/doctor_codex.go` 판별식과 `internal/cli/doctor_golden_test.go` 하네스 재독

## §D. 제약

| 제약 | 출처 |
|---|---|
| 존재-게이트 보존 — doctor·update·codexwiring 어느 쪽도 배선 파일을 만들지 않는다 | `spec.md` §A.4 / REQ-CPW-007 |
| doctor는 읽기 전용·비게이트 유지 | `internal/cli/doctor_codex.go` 머리 주석 |
| 진짜 claude-only 프로젝트의 침묵 보존(골든 3본 포함) | REQ-CPW-006 / AC-CPW-004 |
| 판별식에 개수(11) 금지 | REQ-CPW-005 |
| 전체 스위트 로컬 실행 금지 — 건드린 패키지만, 전 패키지 판정은 CI | CLAUDE.local.md §4 / §6 |

## §E. 마일스톤

### M1 — 판별식 확장과 두 갈래 문구

- `checkCodexWiring`의 판별식에 agent 정의 존재 조회를 더해 B-1의 3상태로 만든다
- `half-wired` × codex 존재 → Warn + 지시문(REQ-CPW-002), `half-wired` × codex 부재 → OK + 사실 문구(REQ-CPW-003)
- `unwired` 두 갈래의 기존 문구는 **손대지 않는다**(REQ-CPW-006)
- 필요하면 `internal/codexwiring`에 `AgentsRelPath` 상수만 추가(B-4)

### M2 — 시험

- 신규: `TestCheckCodexWiring_HalfWiredCodexInstalled`, `..._HalfWiredCodexAbsent`, `..._HalfWiredCountIndependent`, `..._HalfWiredReadOnly`, `..._HalfWiredMessageWidthStaysInBand`(다섯 번째 — 감사 D1 수리로 추가. 기존 폭 시험 2종은 `unwired` 경로만 밟아 새 문구의 폭을 관측할 수 없다)
- 픽스처 헬퍼: 반쪽 프로젝트를 만드는 헬퍼(agent TOML n개를 `t.TempDir()` 아래에 쓰고 배선 파일은 만들지 않는다)
- 기존 시험은 수정하지 않는다 — 수정이 필요해 보이면 그것은 B-1/B-2가 틀렸다는 신호다

### M3 — 뮤턴트로 공허 초록 깨기

AC-CPW-009의 뮤턴트 **8종(M1~M7, M6′ 포함)**을 심고 각각 RED를 관측한 뒤 원복. 원복 확인은 `git diff --stat internal/cli/`.
**M6**(패러프레이즈 지시문)과 **M7**(부재 갈래를 사용자 계층 훑기로 흘려보냄)은 감사 iter2 D1/D2 수리의
합격 조건이므로 생략할 수 없다 — 둘 중 하나라도 초록이면 그 수리는 이름만 들어간 것이다.

### M4 — 실물 확인과 증거 기록

`make build` → 임시 프로젝트에 `moai init --non-interactive` → 두 PATH 갈래로 `moai doctor` 실행 →
출력을 `.moai/reports/t499/` 아래에 남기고 `progress.md` §E.2에 인용. `moai doctor`의 `rc=0` 확인.

## §F. 위험

| 위험 | 완화 |
|---|---|
| 골든 3본이 깨진다 | 골든 하네스는 빈 cwd + codex 부재로 고정돼 `.codex/`가 없다(`spec.md` §A.5) → `unwired` 경로 그대로. AC-CPW-004로 못 박는다 |
| 부재 갈래를 Warn으로 올려 전 사용자 잔소리 | B-2 결정 + AC-CPW-002의 `CheckOK` 단언 |
| 개수 기반 판별로 템플릿 편집에 취약 | REQ-CPW-005 + AC-CPW-005(빈 디렉터리/1개/12개 3케이스) |
| 신규 문구가 패널 폭을 넘겨 다른 행 정렬을 깬다 | 기존 폭 시험 2종(AC-CPW-007)이 이미 그 축을 잡고 있다 |

## §G. 안티패턴

- 기존 문구를 재작성해 회귀 반경을 넓히는 것 — 조건을 좁혀라(`spec.md` §D-2)
- 반쪽 상태를 doctor가 "고쳐 주는" 것 — 존재-게이트 위반
- `go test ./...` 로컬 전체 실행
- 실행하지 않은 명령의 결과를 AC 판정으로 적는 것

## §H. 교차 참조

- `spec.md` §A(측정 전제) · §C(요구) · §D(설계 결정) · §E(범위 밖)
- `acceptance.md` AC-CPW-001 ~ 009
- `SPEC-CODEX-WIRING-001`(REQ-CW-009 / AC-CW-012)

## §I. 미검증 전제 (Gap)

- `phase: "v3.1.5 target"` — **격하됨(리드 측정으로 값의 근거는 확보, 다만 갭 자체는 남는다).** 리드 실측: `pkg/version/version.go:8` = `v3.1.3`, 최신 태그 `v3.1.2`(v3.1.3 미태그), 원격 릴리스 브랜치는 `release/v3.1.4`까지 존재 → **다음 미점유 번호가 v3.1.5**이므로 값은 관측과 일치한다. 그러나 이것은 브랜치와 태그를 센 것이지 로드맵을 본 것이 아니므로 **"이 SPEC을 v3.1.5에 넣기로 결정됐다"는 확인이 아니다.** 이 필드는 릴리스 대상 라벨이지 일정 확약이 아니며(스키마 검증·era 분류 어디에도 영향 없음 — 감사 확인), 일정을 아는 쪽이 정정하면 그대로 따른다
- 대화형 위저드 경로가 같은 반쪽 상태를 만들 수 있는지는 레인이 측정하지 않았다(범위 밖으로 기록)
- `moai update`를 반쪽 프로젝트에 돌렸을 때의 거동은 측정되지 않았다 — 존재-게이트상 아무것도 만들지 않아야 하나, 이 카드에서 확인하지는 않는다
