# t429 판정서 — doctor Harness 5-Layer FAIL 원인 규명

- 카드: t429 (Class B, t425 부수 발견)
- 날짜: 2026-09-02
- 트리: 워크트리 `.claude/worktrees/t429` @ origin/develop **58fbc3b5e** (배차 기반과 동일)
- 브랜치: `WT-doctor-harness-fail`
- 재현 출발점: `.moai/reports/t425/verdict.md` 부수 실측 절 (C3)

## 판정 (결론)

**doctor의 Harness 5-Layer FAIL은 코드 결함이 아니다 — 정확한 진단이다.**
개발 트리는 #913(2e27c14f8)에서 meta-harness를 **의도적으로 인스턴스화**한 프로젝트이며,
활성화가 6개 레이어 모두에서 미완성인 상태다. doctor는 감사된 설계(#1087/F1)대로
"진짜 구성 콘텐츠가 1개라도 있으면 배터리를 실시해 미완성을 노출"하고 있을 뿐이다.

**수리는 단일 수리 범위가 아니다.** exit 1 해소는 (a) 활성화 완성(20+ 파일) 또는
(b) 탈구성(#913 산출물 퇴치) 중 하나인데, 둘 다 운영자 정책 결정 + 위임 강제 스케일이다.
→ **별도 카드 제안과 함께 보고** (dispatch 예상 분기).

## Claim / Evidence

### C1. "트리 루트에서 `moai doctor`가 Harness 5-Layer FAIL로 exit 1" → 확인

- 측정: `moai doctor > .moai/reports/t429/doctor-worktree-58fbc3b5e.log 2>&1` →
  `fail    Harness 5-Layer        L1:FAIL L2:FAIL L3:FAIL L4:FAIL L5:FAIL L6:FAIL`,
  요약 `Pass 24  Warn 3  Fail 1`, **exit=1 직접 관측**(파이프 마스킹 없이 `$?` 직독).
- 로그: `.moai/reports/t429/doctor-worktree-58fbc3b5e.log`

### C2. 6개 레이어 전수 직접 측정 (본 run, 본 워크트리)

| Layer | 측정 명령 | 결과 |
|---|---|---|
| L1 | `grep -L "^triggers:" .claude/skills/hns-*/SKILL.md` | **FAIL** — hns-* 9개 중 8개 부재 (보유: hns-moaiadk-patterns만) |
| L2 | `grep -nE "^\s*harness:\s*$" .moai/config/sections/workflow.yaml` | **FAIL** — 섹션 부재 (rc=1) |
| L3 | `grep -c "moai:harness-start" CLAUDE.md` | **FAIL** — 마커 0쌍 |
| L4 | `grep -l "@.moai/harness/" .claude/skills/moai/workflows/{plan,run,sync,design}.md` | **FAIL** — plan/run/sync 3개 보유, design.md만 부재(파일 존재 확인됨, 8959B) |
| L5 | `ls .moai/harness/` | **FAIL** — baseline 7개 중 2개 보유(main.md, README.md), 5개 부재(plan-extension, run-extension, sync-extension, chaining-rules.yaml, interview-results.md) |
| L6 | `grep -L "^skills:" .claude/agents/harness/*.md` | **FAIL** — 10개 중 6개 부재 (보유: cli-template/hook-ci/quality/workflow-specialist 4개) |

### C3. "harnessConfigured 판정은 옳다" → 확인

- `internal/cli/doctor_harness.go:224-233` MX:NOTE — "진짜 비텔레메트리 파일 1개라도
  구성으로 간주해 배터리를 실시한다(거짓 음성 방지)"는 #1087/plan-audit D2/sync-audit F1의
  **감사된 설계 결정**. 술어를 강화하는 수리는 이 결정을 뒤집는다.
- 개발 트리의 `.moai/harness/main.md`는 크루프가 아니라 #913이 집필한
  "moai-adk-go Domain Harness" 프로젝트 baseline — `harnessConfigured`=true는 정확하다.

### C4. "#913이 의도적 인스턴스화였다" → 확인

- `git show 2e27c14f8 -s`: *"moai-meta-harness 스킬을 dev project에 처음 적용… specialist
  에이전트 + 도메인 지식 스킬을 1쌍씩 생성하고 .moai/harness/main.md에 인덱스화"* —
  main.md는 우연한 잔재가 아니라 구성 행위의 산출물.
- 학습 서브시스템 활동 중 (primary 체크아웃 실측, 2026-09-02): `usage-log.jsonl` 37.9MB
  (당일 04:48 갱신), `proposals/` 81건, `observations.yaml` 11KB, `learning-history/`.
- 추가 축적: hns-* 스킬 9개(002071e55, t259 등), `.claude/agents/harness/` 10개(f55aefef3 이후).

### C5. "FAIL의 폭발 반경은 사람 실행뿐" → 확인

- `Makefile:59` — `go run ./cmd/moai doctor --check "Agent Emit Embed"` 단일 체크로 우회.
- `.github/workflows/*.yaml`, `.claude/hooks/moai/*.sh` — full `moai doctor` 게이트 배선 0건.
- 즉 exit 1을 만나는 것은 트리 루트에서 full doctor를 도는 사람/스크립트뿐이다.

### C6. "설치 바이너리 lag가 판정을 오염하지 않는다" → 확인

- 설치 바이너리: v3.1.2 @ **64bba61aa**, HEAD **58fbc3b5e** — lag 존재.
- `git log --oneline 64bba61aa..58fbc3b5e -- internal/cli/doctor_harness.go internal/harness/`
  → **빈 출력(무변경)** — Harness 체크의 구현 패키지가 구간 내 등가이므로 바이너리 측정이
  현재 HEAD 코드에 그대로 귀속된다.

## 근거 없는 가설의 기각 (t425 가설 검증)

- **"doctor 로직 결함" 가설** → 기각. 게이트 3-상태 모델(미구성 스킵 / 텔레메트리만 스킵 /
  진짜 콘텐츠 → 배터리)은 의도대로 동작. 개발 트리는 3번째 상태가 맞다.
- **"observe-only 모드가 있고 doctor가 이를 무시한다" 가설** → 기각.
  `internal/harness/`·`internal/cli/harness/`에 observe-only 상태 표면 없음
  (`grep -rn "observe"` — applier의 apply-outcome observer뿐).

## 별도 카드 제안 (운영자 결정 필요)

exit 1 해소의 유일한 두 갈래 — 어느 쪽이냐는 **개발 트리의 harness 자세**에 대한
운영자 정책 결정이며, 결정 후 실행은 위임 강제 스케일이다.

| 제안 | 내용 | 규모 | 부담 |
|---|---|---|---|
| **A. 활성화 완성** | L1 triggers ×8, L2 workflow.yaml 섹션, L3 CLAUDE.md 마커 쌍, L4 design.md import, L5 baseline ×5, L6 skills: ×6 | **20+ 파일** — manager-develop/builder-harness 위임 강제 | CLAUDE.md 마커·workflow import는 개발 세션 동작을 바꾼다 |
| **B. 탈구성** | main.md·README.md·seeds를 tracked 트리에서 퇴치(내용은 `.moai/docs/` 등으로 이관 가능) → `harnessConfigured`가 거짓이 되어 OK 스킵 | 2-4 파일 | **#913의 의도된 작업 폐기** + 활동 중인 학습 서브시스템과 충돌(usage-log 당일 갱신) |
| C. 수용 | FAIL을 "정확한 진단"으로 문서화하고 방치 (doctor는 틀리지 않았다) | 0 파일 | exit 1 지속 — full doctor를 도는 스크립트 작성 시 주의 필요 |

- **기각된 중간안**: "L1+L6만 보강(14 파일)"은 L2~L5가 남아 exit 1을 해소하지 못한다 —
  독립 가치 없음, A안 채택 시에만 묶인다.
- **코드 수리(술어 변경)는 기각**: #1087 감사된 설계를 뒤집고 거짓 음성을 재도입한다.

## Baseline-attribution

- 모든 측정은 본 run에서 직접 실행·관측 (lane, 2026-09-02).
- 트리: 워크트리 `.claude/worktrees/t429` @ **58fbc3b5e** (= origin/develop 배차 기반).
- 바이너리: 설치본 v3.1.2 @ 64bba61aa — C6의 무변경 구간 로그로 코드 등가성 입증.
- 참조: `.moai/reports/t425/verdict.md` (부수 절), `internal/cli/doctor_harness.go` (본 트리).

## Gaps (관측하지 못한 것)

- `bin/moai` 재빌드 후의 재측정은 하지 않았다 — C6 무변경 로그가 등가성을 입증하므로 생략.
  (재빌드 시 "Agent Emit Embed" 체크 판정도 함께 바뀌나 본 카드와 무관.)
- L4에서 plan/run/sync의 `@.moai/harness/` import가 **언제·왜** 추가됐는지의 이력 추적은
  하지 않았다 — FAIL/ PASS 경계 판정에 불요.
- 4개 전문 에이전트(cli-template/hook-ci/quality/workflow-specialist)의 `skills:` 키가
  있는 이유(hns-moaiadk-* preload)와 6개가 없는 이유의 의도 차이는 규명하지 않았다 —
  A안 실행 카드의 설계 단계 소관.

## Residual-risk (잔여 위험)

- t425 힌트 재확인: `runDoctor` 테스트(cwd=internal/cli)는 `.moai/harness`를 보지 않아
  이 배터리의 커버 밖이다 — internal/cli 하위에 harness 계열 파일이 생기는 날 테스트가
  환경 의존 적색이 된다. A안 실행 시 테스트 격리 점검 권장.
- 학습 서브시스템이 계속 관찰 중이므로, B안(탈구성) 채택 시 런타임 재생성
  (usage-log 등)이 즉시 재발한다 — B안은 `harnessRuntimeArtifacts` 제외 집합과의
  정합 확인이 선행돼야 한다.
- doctor 체크 자체는 무결하나, 향후 L1 접두어 목록에 새 세대(prefix)가 추가되면
  본 측정치(9 스킬/10 에이전트)는 재측정 대상이다.
