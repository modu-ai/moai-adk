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

### B.3 [결정] 재측정을 이 SPEC 의 완료 조건에 넣을 것인가

증거의 가장 약한 고리는 「응답 문자열 의존」이다. 재측정(AC-WXR-004)을 **완료 조건에 포함**한다.
포함하지 않으면 독트린이 미검증 전제 위에 앉고, REQ-WXR-007 의 잔여 위험 문구가 영구화된다.

## C. Milestones (우선순위 순 — 시간 추정 없음)

| # | 내용 | 우선순위 |
|---|---|---|
| M1 | **재측정 먼저** — 연쇄 진입 후 Exit 직후 작업 디렉터리를 직접 판독하고, 런타임 버전을 함께 기록한다(AC-WXR-004). 결과를 `.moai/reports/t965/` 에 추가 기록으로 남긴다. | High |
| M2 | 독트린 § `EnterWorktree` / `ExitWorktree` Tools 개정 — REQ-WXR-001~007 충족 (복귀 지점 명시, 연쇄 케이스 명시, 두 트리 지목 문장 경고, originating 문장 정정, 관측 귀속, 미재현 불일치 기록, 잔여 위험). | High |
| M3 | 템플릿 미러에 동일 반영(REQ-WXR-008). 같은 커밋. | High |
| M4 | 문서 층 검증 — 인수조건의 문자열 판별식 실행, 두 사본 간 해당 절의 동일성 확인. | Medium |

M1 을 M2 앞에 두는 이유: 재측정 결과가 관측 1~4 와 어긋나면 **M2 에 쓸 문장 자체가 달라진다.**
문서를 먼저 쓰고 나중에 재는 순서는, 측정이 문서를 확인하는 도구로 전락한다.

## D. Constraints

- [HARD] 코드 변경 없음. `internal/`, `pkg/`, `cmd/` 를 건드리지 않는다.
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

## F. Anti-Patterns

- 재측정 없이 「primary checkout 고정」을 계약 문언으로 굳히기.
- 원 제보 불일치를 「구버전 동작」이나 「측정 오류」로 **추정해** 닫기 — 조사가 명시적으로
  추정을 금지했다.
- 상위 도구 설명 인용문을 우리 독트린에 복사해 두고 그것이 우리 계약인 것처럼 읽히게 하기.

## G. Cross-References

- `.moai/reports/t965/observations.md` — 근거 관측 전량.
- `.claude/rules/moai/workflow/worktree-integration.md` § Terminology Glossary (L1/L2 구분),
  § `EnterWorktree` / `ExitWorktree` Tools (개정 대상).
- `.claude/rules/moai/core/verification-claim-integrity.md` §2 — 관측 귀속 요구(명령 + 관측
  출력 + 이 트리·이 회차).
