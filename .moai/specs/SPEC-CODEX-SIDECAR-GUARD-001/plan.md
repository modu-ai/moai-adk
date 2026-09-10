# SPEC-CODEX-SIDECAR-GUARD-001 — 구현 계획

> 카드 t405 · Tier S · 기준 트리 `64bba61aa`

## §A. 맥락

`internal/cli/init_agent_flag_test.go`의 경로-집합 시험 4종 중 2종이 세 번째 배선 산출물
(`.moai/state/codex-wiring.json`)을 보지 않는다. 결손은 존재·부재 **양방향에 각각 하나씩**이며,
둘 다 적색을 만들지 않으므로 조용하다(spec.md §A.3 · §A.4).

## §B. 핵심 결정 (되돌리기 어려운 순서대로)

여기서 판단이 갈릴 여지가 있는 것은 **뮤턴트 설계** 하나다. 나머지는 기계적이다.

### B.1 뮤턴트는 "격리"여야 한다 — 가장 되돌리기 어려운 결정

새 단언을 추가하고 초록을 보는 것은 아무것도 입증하지 않는다. 빠져 있던 단언은 원래도 적색을
내지 않았기 때문이다. 그래서 각 새 단언이 **적색이 될 수 있음**을 뮤턴트로 보인다.

여기서 흔한 실수는 "아무거나 깨뜨리는" 뮤턴트를 쓰는 것이다. 배선 전체를 끄면 세 단언이 모두
적색이 되지만, 그것은 새 단언이 제 몫을 한다는 증거가 아니라 **기존 두 단언이 이미 하던 일**이다.
따라서 뮤턴트는 기존 두 단언을 **초록으로 남기도록** 좁게 설계한다.

| 방향 | 뮤턴트 | 기존 2단언 | 새 단언 |
|---|---|---|---|
| 부재 (`:109`) | `claude` 분기가 sidecar **만** 기록 — 임시 편집 2건: (1) `internal/codexwiring`에 비exported `writeSidecar`를 호출하는 임시 export 래퍼(`WriteSidecarOnly`) 추가, (2) `init.go` `claude` 분기가 그 래퍼를 호출. 관측 후 즉시 원복 | GREEN (hooks/config 여전히 미기록) | RED |
| 존재 (`:82`) | `Wire`에서 `writeSidecar` 호출 **만** 제거 | GREEN (hooks/config 여전히 기록) | RED |

부재 방향을 임시 2편집으로 심는 이유는 패키지 경계다: `writeSidecar`는 `internal/codexwiring`의
비exported 함수라 `internal/cli`의 `claude` 분기가 직접 호출할 수 없고, 직접 형태는 컴파일되지
않는다.

### B.2 존재 방향 뮤턴트에는 생사 확인이 필요하다

부재 방향 뮤턴트는 코드를 **추가**하므로 심긴 사실이 자명하다. 존재 방향 뮤턴트는 코드를
**제거**하므로, 엉뚱한 곳을 지웠거나 아무것도 안 지웠어도 겉보기가 같다. 그 경우 `:82`의 RED는
다른 이유일 수 있고, 최악의 경우 RED조차 없이 "뮤턴트를 심었는데 초록이니 단언이 약하다"는
반대 결론이 나온다.

`:70`이 판별식이다. `:70`은 기준 트리에서 **이미** sidecar 존재를 단언하므로, `writeSidecar`가
정말 사라졌다면 반드시 함께 깨진다. `:70`이 초록이면 뮤턴트는 죽은 것이고 `:82`의 관측은 무효다
(AC-CSG-006).

### B.3 단언은 파일 경로 단위로만

`.codex/` 디렉터리 부재를 단언하고 싶은 유혹이 있으나 **항상 실패한다** — 템플릿 트리가
`.codex/agents/**`를 배포하므로 배선이 없어도 디렉터리는 정당하게 존재한다. 기존 `:97`의 주석이
이미 이 사실을 기록하고 있다.

### B.4 프로덕션 코드는 읽기 전용

영구 변경은 시험 파일 1개에 국한한다. 뮤턴트는 관측 도구이며 커밋되지 않는다(spec.md §E.2).

## §C. Pre-flight

| # | 확인 | 명령 |
|---|---|---|
| 1 | 기준 트리 확인 | `git rev-parse --short HEAD` |
| 2 | 시험 4종 현재 초록 | `go test ./internal/cli/ -run 'TestRunInit_Agent' -v` |
| 3 | 대상 함수 좌표 재확인(줄 번호는 밀릴 수 있음) | `grep -n 'func TestRunInit_Agent' internal/cli/init_agent_flag_test.go` |

## §D. 마일스톤

### M1 — 단언 2줄 추가

- `TestRunInit_AgentClaudeLeavesNoCodexFiles`의 경로 슬라이스에
  `".moai/state/codex-wiring.json"` 추가 → `:97`과 동일 집합
- `TestRunInit_AgentBothWiresBothSides`의 경로 슬라이스에
  `".moai/state/codex-wiring.json"` 추가 → `:70`과 동일 집합
- 기존 표현 방식(슬라이스 순회 + `os.Stat`) 유지
- 두 시험의 doc comment를 sidecar 포함으로 갱신 (`:82`의 "Codex wiring files exist" 문구가 2종만
  가리키는 것으로 읽히므로)
- 판정: AC-CSG-001 · AC-CSG-002 · AC-CSG-003

### M2 — 격리 뮤턴트 2종 관측

- B.1 표대로 부재 방향 뮤턴트 심기 → 관측 → 원복
- B.1 표대로 존재 방향 뮤턴트 심기 → 관측(`:82` RED + `:70` RED 교차 확인) → 원복
- 각 관측의 **명령과 관측된 출력**을 `progress.md` §E.2에 기록. 요약은 증거가 아니다
- 판정: AC-CSG-004 · AC-CSG-005 · AC-CSG-006 · AC-CSG-007

### M3 — 범위 한정 검증

- `go test ./internal/cli/... ./internal/codexwiring/...`
- 전체 스위트 지역 실행 금지 — 전 패키지 판정은 CI 소관
- 판정: AC-CSG-008

## §E. 위험

| 위험 | 완화 |
|---|---|
| 존재 방향 뮤턴트가 죽은 채로 통과 판정 | AC-CSG-006의 `:70` 교차 확인 (B.2) |
| 뮤턴트가 커밋에 섞임 | AC-CSG-007이 `git status --porcelain`으로 프로덕션 경로 청결 확인 |
| 줄 번호로 대상을 지목해 엉뚱한 시험 수정 | 함수명으로 지목. M1 이후 줄 번호는 밀린다 |
| 작업 중 세 번째 거울상 발견 시 범위 확대 | spec.md §E.2 상한 — 흡수하지 않고 새 카드로 리드에 올림 |

## §F. 후속 (본 SPEC에서 행동하지 않음)

카드 **t393**이 아직 착지하지 않았다. 착지하면 그 `AC-IHP-006a`가 wizard 경로를 덮으므로, 그
시점에 **남는 결손 범위를 다시 측정**해야 한다. 본 SPEC은 t393에 게이팅되지 않으며, 지금 그
잔여 범위를 측정하지 않는다 — 착지하지 않은 카드에 대고 잰 수치는 착지 시점에 이미 낡는다.

## §G. 안티패턴

- 새 단언을 추가하고 초록을 보고 끝내기 — 빠진 단언은 원래도 적색을 내지 않았다
- 배선 전체를 끄는 광범위 뮤턴트로 적색 능력을 주장하기 — 기존 단언이 낸 적색이다
- `.codex/` 디렉터리 부재 단언 — 템플릿이 `.codex/agents/**`를 배포하므로 항상 실패
- 뮤턴트를 커밋에 남기기
- 전체 스위트 지역 실행을 근거로 삼기

## §H. 교차 참조

- `spec.md` §A (측정 전제) · §E.1 (범위 확대 판정 기록) · §E.2 (상한)
- `acceptance.md` AC-CSG-001 … AC-CSG-008
- `SPEC-CODEX-WIRING-001` — REQ-CW-001 · AC-CW-004 · AC-CW-005 (본 시험 4종의 발주 근거)

## §I. 미검증 값 기록

`spec.md`의 `phase: "v3.1.5 target"`은 **측정된 값이 아니다.** 관련 SPEC
`SPEC-CODEX-WIRING-001`의 같은 필드에서 상속했다. 최근 착지한 형제 카드 둘
(`SPEC-BOARDLOCK-ERRNO-001`, `SPEC-VACUOUS-FLOOR-GUARD-001`)은 `"v3.1.4 target"`을 달고 있어
분포가 갈린다. `release/v3.1.4`는 준비됐으나 미태그이므로 "develop에 착지하는 작업이 어느
릴리스에 실리는가"는 현재 열려 있는 질문이고, 이 SPEC은 그것을 판정할 근거를 갖고 있지 않다.

**따라서 이 값은 결정이 아니라 상속된 미검증 값으로 읽어야 한다.** 릴리스 상태가 확정되면
그때 정정한다.

## §J. 산출물 집합 편차 기록 (의도적)

**이 SPEC은 산출물 4종을 갖는다 — `spec.md` · `plan.md` · `acceptance.md` · `progress.md`.**
`spec-workflow.md` § SPEC Complexity Tier의 Tier S 규정은 **2종**(`spec.md` + `plan.md`, AC는
`spec.md` §3에 인라인)이다(`progress.md`는 전 Tier 공통이라 Tier 산출물 수에 세지 않는다).
따라서 Tier 규정 대비 초과분은 `acceptance.md` 1종이다.

**이 편차는 드리프트가 아니라 오케스트레이터 결정이다.** 접기지 않고 그대로 둔다.

**근거**: 이 SPEC에서 **수용 기준 자체가 산출물**이다 — 격리 뮤턴트 프로토콜 2종과 뮤턴트 생존
교차 확인을 담은 8개 기준이 본체다. 이 92줄을 `spec.md` §3으로 인라인해도 토큰이 줄지 않고,
옮기는 과정에서 뮤턴트 문구가 훼손될 위험만 생긴다. Tier S의 2종 기본값은 **의례 부담을 막기
위한** 규정인데, 여기서는 그 의례가 곧 payload다.

**예산은 초과하지 않았다**: Tier S 상한은 요구 8 · 수용 기준 8이며, 이 SPEC은 요구 5 · 수용 기준
8로 둘 다 이내다. 편차는 파일 분할 방식 하나뿐이고 분량 초과가 아니다.

> 참고(정보): `.claude/skills/moai-workflow-spec/SKILL.md:219`는 "[HARD] Every SPEC directory MUST
> contain all 3 files"라고 적어 3종을 요구한다 — Tier 분류 체계 도입 이전 문구로 보이며 Tier S
> 규정과 어긋난다. 본 SPEC은 Tier 표(`spec-workflow.md`)를 Tier 산출물 집합의 SSOT로 본다. 이
> 문서 간 불일치는 본 카드 범위 밖이며 손대지 않는다(spec.md §E.2).

**plan-phase 커밋 주체(subject)는 산출물 3종으로 적는다** — Tier S 기본값 2종이 아니다. 감사자가
설명 없는 불일치를 발견해 "결정인가 사고인가"를 추측하게 두지 않기 위한 기록이다.
