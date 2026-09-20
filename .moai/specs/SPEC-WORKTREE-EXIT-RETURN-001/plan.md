# SPEC-WORKTREE-EXIT-RETURN-001 — 구현 계획

> Tier S. 코드 변경 0. 인도물은 독트린 문서 한 절(+ 템플릿 미러)과, 증거의 약한 고리를 닫는
> 재측정 기록이다.

## A. Context

- 근거 파일: `.moai/reports/t965/observations.md` (관측 1~4, Gaps, Residual-risk).
- 대상 파일: `.claude/rules/moai/workflow/worktree-integration.md`
  § `EnterWorktree` / `ExitWorktree` Tools — 현재 복귀 지점에 대한 서술은 "`ExitWorktree`
  returns to the originating checkout" 한 문장뿐이다.
- 템플릿 미러: `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md`
  (§2.0 / Template-First 규율. 같은 커밋에 반영한다).

## B. 가장 먼저 검토돼야 할 결정 (되돌리기 비쌈 순)

### B.1 [결정] 「primary checkout」을 계약 문언으로 채택할 것인가 — 관측 귀속으로 둘 것인가

관측 1~4 는 복귀 지점이 primary checkout 으로 고정된 것처럼 보이게 한다. 그러나 이것은 **우리가
보장할 수 있는 계약이 아니라 상위 런타임의 관측된 동작**이다. 두 갈래:

- (채택) **관측 귀속 형태로 쓴다** — "관측(날짜·버전·기록 경로)에서 복귀 지점은 primary
  checkout 이었다"로 쓰고, 보장 문언(`always`, `guaranteed`)을 쓰지 않는다.
- (기각) 「ExitWorktree 는 항상 primary checkout 으로 복귀한다」 — 우리가 만들지 않은 동작에
  대한 무조건 주장이 되고, 상위가 바꾸면 조용히 거짓이 된다.

이 결정이 REQ-WXR-005 의 형태를 정한다. 되돌리기가 가장 비싸므로 먼저 확정한다.

### B.2 [결정] 「originating checkout」 문장을 고칠 것인가 — 옆에 주석을 붙일 것인가

독트린의 기존 문장은 관측 4(워크트리 출발 세션)에서 오독을 부른다. 두 갈래:

- (채택) **문장을 고친다** — "originating" 이 세션 출발 디렉터리로 읽히는 경로를 없앤다. 모호한
  문장을 남긴 채 아래에 해설을 붙이면, 훑어 읽는 독자는 위 문장만 읽는다.
- (기각) 원문 유지 + 각주 추가.

### B.2b [결정] 두 사본의 귀속 표면 — 되돌리기 가장 비쌈

**이 결정이 B.1 보다 먼저 확정돼야 한다.** B.1 이 정하는 「관측 귀속 형태」가 배포 미러에서
기계적으로 금지되면, B.1 의 채택안 자체가 완료 시점에 CI 를 적색으로 만든다.

- (채택 — **Path A**) **두 사본을 바이트 동일하게 두고, 양쪽 모두 런타임 버전 인라인만으로
  귀속한다.** 어느 사본에도 날짜 리터럴·저장소 내부 증거 경로·SPEC ID 를 싣지 않는다. 증거 경로
  인용은 배포되지 않는 층(`spec.md`·`progress.md`)이 맡는다. 근거·가드 셋·전례·표는
  `spec.md` §1.5.
- (기각 — **Path B**) 비대칭을 유지하고 대상 파일을 `rule_template_mirror_test.go` 의
  `workflowOptMirroredPaths` 에서 빼 `sanitized_pair_parity_test.go` 의 `sanitizedPairPaths` 에
  등재한다 — 두 편집 모두 `internal/` 아래 테스트 파일이라 이 SPEC 의 [HARD] 「코드 변경 없음」과
  충돌하고, 최소 이행(제거만)은 이 파일을 묶는 상시 가드를 0 으로 만든다(`spec.md` §1.5 D2′).
- (기각) 미러에도 날짜를 싣기 위해 `internal_content_leak_test.go` 의 `dateAllowlist` 에 예외를
  추가한다 — 코드 층 변경이고, 배포 문서에 내부 날짜를 들이려고 가드를 여는 거래다.
- (기각) 미러의 귀속을 「측정된 관측이다」로 뭉갠다 — 파리티를 모호함으로 바꾸는 것이지 푸는
  것이 아니다. 낡음을 보이게 만드는 좌표가 사라진다.
- (초과분) **0.2.0 의 비대칭 결정.** 틀린 결정이라기보다 **전제가 하나 빠진** 결정이었다 —
  누출 가드 둘은 보고 바이트 동일성 가드는 보지 않았고, 그래서 결함을 제거하는 대신 다른 가드로
  옮겼다. 그 비대칭 아래에서는 **SPEC 문언을 만족시키는 트리 상태가 존재하지 않는다.**

이 결정이 REQ-WXR-005 · REQ-WXR-008 · AC-WXR-005 · AC-WXR-007 의 형태를 동시에 정한다.

### B.3 [결정] 재측정을 이 SPEC 의 완료 조건에 넣을 것인가

증거의 가장 약한 고리는 「응답 문자열 의존」이다. 재측정(AC-WXR-004)을 **완료 조건에 포함**한다.
포함하지 않으면 독트린이 미검증 전제 위에 앉고, REQ-WXR-007 의 잔여 위험 문구가 영구화된다.

## C. Milestones (우선순위 순 — 시간 추정 없음)

| # | 내용 | 우선순위 |
|---|---|---|
| M0 | **증거 반출 먼저** — `.moai/reports/t965/observations.md` 는 `.gitignore` 되어 어떤 clone 에도 없다. 조사 기록을 추적되는 경로 `.moai/specs/SPEC-WORKTREE-EXIT-RETURN-001/evidence/` 아래로 반출하고 `git ls-files --error-unmatch` 로 확인한다(`spec.md` §1.5, AC-WXR-005 (3)). **인용하는 층은 독트린이 아니라 `spec.md`·`progress.md` 다** — Path A 아래에서 독트린은 어느 사본도 경로를 인용하지 않는다. 반출 없이 인용하면 그 층에서도 귀속이 해소되지 않는다. | High |
| M1 | **재측정** — 연쇄 진입 후 Exit 직후 작업 디렉터리를 직접 판독하고, 런타임 버전을 함께 기록한다(AC-WXR-004). 결과를 M0 이 만든 evidence 경로에 남기고, **이 기록만을 담은 자기 커밋**으로 착지시킨다(AC-WXR-008). | High |
| M2 | 독트린 § `EnterWorktree` / `ExitWorktree` Tools 개정 — REQ-WXR-001~007 충족 (복귀 지점 명시 + 표본 명시, 연쇄 케이스 명시, 두 트리 지목 문장 경고, originating 문장 정정, 회차별 관측 귀속, 미재현 불일치 기록 + 공존 문장, 잔여 위험). | High |
| M3 | 템플릿 미러에 **바이트 동일**하게 반영(REQ-WXR-008, §B.2b). M2 와 같은 커밋. 커밋 **전에** AC-WXR-007 (3) 의 **세 가드를 모두** 로컬에서 실행해 초록을 확인한다 — 바이트 파리티(`TestRuleTemplateMirrorDrift`), 중립성 default, 중립성 strict. 하나만 돌리면 나머지 두 축이 적색인 트리가 인수조건을 통과한다. | High |
| M4 | 문서 층 검증 — 인수조건의 문자열 판별식을 **두 사본 각각에서** 실행해 같은 결과를 확인하고, AC-WXR-007 (1) 의 바이트 동일성과 (2) 의 **양쪽 공통** 부재 검사, AC-WXR-008 의 조상 판정을 확인한다. | Medium |

**M0 → M1 → M2 순서, 그리고 M1 이 자기 커밋을 갖는 이유.** 재측정 결과가 관측 1~4 와 어긋나면
**M2 에 쓸 문장 자체가 달라진다.** 문서를 먼저 쓰고 나중에 재는 순서는, 측정이 문서를 확인하는
도구로 전락한다.

그런데 그 이유를 여기 산문으로 적어 두는 것만으로는 **아무도 나중에 확인할 수 없다.** git 은
커밋 **내부**의 저작 순서를 원리상 기록하지 않으므로, M1 기록과 M2 개정이 한 커밋에 함께
들어가면 「먼저 쟀다」는 주장은 영구히 검증 불가가 된다 — `verification-claim-integrity.md`
**§2.3** 이 정확히 이 모양을 금지한다. 따라서 순서는 산문이 아니라 **커밋 그래프**가 증언한다:
M1 기록은 자기 커밋에 먼저 착지하고, `git merge-base --is-ancestor <M1> <M2>` 가 exit 0 을 낸다
(AC-WXR-008). 한 커밋에 합쳐야만 하는 사정이 생기면 이 순서 조항을 커밋 그래프가 검증 가능한
문언으로 다시 쓴다 — AC 를 「통과한 것으로 처리」하지 않는다.

## D. Constraints

- [HARD] 코드 변경 없음. `internal/`, `pkg/`, `cmd/` 를 건드리지 않는다. **가드 파일 둘을 모두
  포함한다** — 중립성 가드(`internal/template/internal_content_leak_test.go`)의 `dateAllowlist`
  와 바이트 동일성 가드(`internal/template/rule_template_mirror_test.go`)의
  `workflowOptMirroredPaths`. 둘 중 어느 쪽을 여는 길도 §B.2b 에서 기각했다. **읽기만 한다.**
- [HARD] **두 사본을 바이트 동일하게 유지한다**(§B.2b, AC-WXR-007 (1)). 로컬 사본에만 무언가를
  더하는 편집은 이 제약 위반이다 — 대상 파일이 바이트 동일성 허용목록에 등재돼 있다.
- [HARD] **어느 사본에도** 날짜 리터럴·저장소 내부 경로(`.moai/reports/`, `.moai/specs/`)·
  SPEC ID 를 싣지 않는다(§B.2b, AC-WXR-007 (2)). 「로컬 사본에는 괜찮다」가 성립하지 않는 이유는
  바이트 파리티다 — 로컬에 실은 것은 그대로 배포된다.
- [HARD] Gaps 를 findings 로 격상하지 않는다 — 특히 원 제보의 불일치에 원인을 부여하지 않는다.
- [HARD] 상위 도구 설명에 대고 요구사항을 쓰지 않는다(우리 표면이 아니다).
- 재측정은 이 카드가 소유한 버리는 워크트리 안에서만 한다. 살아 있는 남의 카드 트리를 쓰지
  않는다.

## E. Risks

| 위험 | 완화 |
|---|---|
| 재측정이 관측 1~4 와 다른 답을 낸다 | 그 경우 M2 의 문언은 **차이 자체**를 기록한다. 두 회차를 합쳐 「고정」이라고 쓰지 않는다. |
| 런타임 버전이 바뀌어 서술이 낡는다 | REQ-WXR-005 의 버전 귀속이 낡음을 **보이게** 만든다. 버전 없는 주장이 조용히 낡는 것이 더 나쁘다. |
| 독트린 개정이 다른 절과 충돌한다 | 개정 범위를 § `EnterWorktree` / `ExitWorktree` Tools 한 절로 한정한다. |
| 미러 반영이 CI 를 적색으로 만든다 | 이 파일을 지배하는 가드는 **셋**이고 서로 배타다(바이트 파리티 / 중립성 default / 중립성 strict). §B.2b 의 Path A + AC-WXR-007 (2) 양쪽 공통 부재 검사 + (3) **세 가드 전부** 로컬 선행 실행. 커밋 **전에** 관측한다 — CI 가 처음 알려주는 자리가 되면 이미 늦다. |
| 결함을 제거하는 대신 다른 가드로 옮긴다 | 0.2.0 에서 실제로 일어났다(누출 가드는 피하고 파리티 가드를 적색으로 만듦). AC-WXR-007 (3) 이 **세 축을 한 검사에 묶어** 한 축만 보고 초록이라 부르는 이행을 막는다. |
| 재측정 회차의 버전이 `2.1.278` 과 달라 서술이 갈린다 | 갈리는 것이 정상이고 **합치는 것이 결함**이다. AC-WXR-005 (2) 가 회차별 분리를 검증하고, 다를 경우 다르다는 사실 자체를 싣게 한다. |

## F. Anti-Patterns

- 재측정 없이 「primary checkout 고정」을 계약 문언으로 굳히기.
- 원 제보 불일치를 「구버전 동작」이나 「측정 오류」로 **추정해** 닫기 — 조사가 명시적으로
  추정을 금지했다.
- 상위 도구 설명 인용문을 우리 독트린에 복사해 두고 그것이 우리 계약인 것처럼 읽히게 하기.
- **표본 2점·1점 위에 `independent of` 같은 전칭 술어를 올리기.** 근거 파일은 「…로 **보인다**」로
  유보했다. 규범 문장이 그 유보를 성질 주장으로 굳히면 SPEC 에서 가장 강한 일반화가 가장 약한
  증거 위에 앉는다(AC-WXR-001 (3) 이 기계적으로 막는다).
- **해소되지 않는 경로를 귀속 앵커로 인용하기.** `.gitignore` 된 경로를 `spec.md`·`progress.md`
  에서 인용하는 것(반출 전 인용), 그리고 **저장소 내부 경로를 독트린 본문에 싣는 것** — 후자는
  바이트 파리티 때문에 사본을 가리지 않고 모든 사용자 프로젝트로 배포된다.
- **한 축의 가드만 돌려 보고 「초록」이라 부르기.** 이 파일에는 가드가 셋 있고 서로 배타다.
  하나만 돌린 관측은 나머지 두 축에 대해 아무것도 말하지 않는다(AC-WXR-007 (3)).
- **결함을 다른 가드로 옮기고 해소로 보고하기.** 0.2.0 의 비대칭 결정이 이 모양이었다 — 누출
  가드는 피했으나 파리티 가드를 적색으로 만들었고, 조건이 사라진 것이 아니라 이동했다. 수정이
  닫는 축과 **열어 두는 축**을 함께 적는다.
- **두 측정 회차의 귀속을 한 묶음으로 합치기.** 「2026-09-19 에 관측, 런타임 <재측정 버전>」은
  어느 회차에도 참이 아닌 거짓 귀속이다.
- **M1 기록과 M2 개정을 한 커밋에 넣고 순서를 커밋 메시지로 주장하기.** 커밋 메시지는 순서를
  **주장**하고 커밋 그래프만이 순서를 **증언**한다.

## G. Cross-References

- `.moai/reports/t965/observations.md` — 근거 관측 전량. **`.gitignore` 되어 있으므로 인용
  대상이 아니다** — M0 이 `.moai/specs/SPEC-WORKTREE-EXIT-RETURN-001/evidence/` 로 반출한 사본이
  독트린이 인용하는 경로다(`spec.md` §1.5).
- `internal/template/rule_template_mirror_test.go` — **바이트 동일성 가드.** 대상 파일이
  `workflowOptMirroredPaths` 허용목록에 등재돼 있고, 단언부는
  `bytes.Equal(srcContent, mirrorContent)`, 실패 센티널은 `RULE_TEMPLATE_MIRROR_DRIFT` 다. 이
  SPEC 의 산출물을 지배하는 가드이며, 0.2.0 이 이것을 제약 목록에서 빠뜨린 것이 D1′ 의 발생
  경로였다. **읽기만 한다 — 고치지 않는다.**
- `internal/template/sanitized_pair_parity_test.go` — 구조 드리프트 가드(`sanitizedPairPaths`
  레지스트리, 센티널 `SANITIZED_PAIR_PARITY_DRIFT`). 바이트 동일성을 지킬 수 없는 쌍이 옮겨
  가는 자리다. Path B 가 요구했을 등재처이며, **이 카드는 등재하지 않는다**(§B.2b 기각).
- `internal/template/internal_content_leak_test.go` — 템플릿 중립성 가드. default 계층 클래스
  `C1-spec-id-prefix` 가 SPEC ID 를, strict 계층 클래스 `S1-internal-date` 가 날짜 리터럴을
  차단한다(§B.2b). 저장소 내부 경로는 **어느 클래스도 잡지 않는다**(AC-WXR-007 (2) 의 잔여
  위험). **읽기만 한다 — 고치지 않는다.**
- `.claude/rules/moai/workflow/kanban-dispatch.md` 와 그 템플릿 미러 — **전례.** 허용목록에
  등재되지 않았는데도 두 사본이 바이트 동일하고, 미러가 "Measured on Claude Code 2.1.27x" 로
  관측을 귀속시키며, 날짜 리터럴 0건으로 두 티어 모두 초록이다(`spec.md` §1.5).
- `.github/workflows/template-neutrality-check.yaml` — 위 가드를 `MOAI_TEMPLATE_LEAK_STRICT: '1'`
  로 실행하는 CI. 미러 경로 변경이 트리거다.
- `.claude/rules/moai/workflow/worktree-integration.md` § Terminology Glossary (L1/L2 구분),
  § `EnterWorktree` / `ExitWorktree` Tools (개정 대상).
- `.claude/rules/moai/core/verification-claim-integrity.md` **§2** — 관측 귀속 요구(명령 + 관측
  출력 + 이 트리·이 회차). 회차를 섞은 귀속을 금지하는 조항이 여기다(AC-WXR-005 (2)).
- `.claude/rules/moai/core/verification-claim-integrity.md` **§2.3** — 순서 귀속. 「baseline 이
  먼저 착지했다」는 주장은 커밋 그래프만이 증언한다(AC-WXR-008). §2 만 인용하고 §2.3 을
  빠뜨리면 순서 조항이 산문 단언으로 남는다.
- `.claude/rules/moai/core/agent-common-protocol.md` § Parallel Execution — 「인용 전 반출」
  의무. 인용 대상은 추적되는 경로여야 한다(M0, AC-WXR-005 (3)).
