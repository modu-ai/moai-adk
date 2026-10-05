# Card t670 — Verdict

- card: t670 (moai gpt 시작 시 Bypass Permissions·workspace trust 프롬프트 방지)
- branch: WT-trust-seed-verify (base: develop 5e0f71175)
- evidence: .moai/reports/t670/verdict.md
- date: 2026-09-13, lane-6 · 증거 확정 카드 (코드 변경 없음)

## Claim (주장)

1. **Bypass Permissions 경고** — cd6bad85f(`seedGatewayBypassAcceptance`, settings.json에 `skipDangerousModePermissionPrompt` 시딩)로 수리됐고 운영자 확인을 마쳤다(배차 기재 사실).
2. **workspace trust 경고** — 별도의 미수리 결함이 아니라, 이미 억제돼 있다: t649(ce79ef7ca)부터 `seedGatewayUIState`(internal/cli/gateway_ui_state.go:60-70)가 **매 신규 family 기동 시(exec 전 Prepare 단계, gateway_session.go:69 호출점)** 선택된 설정의 `.claude.json`에서 정확한 cwd 키의 `hasTrustDialogAccepted`를 family 프라이빗 `.claude.json`으로 복사한다.
3. **named profile로 기동할 때** 그 프로필 자체가 해당 워크스페이스를 신뢰한 적 없으면 1회 물음 — 설계된 동작이다(코드 주석: "A named profile owns its trust decisions; do not borrow a global approval").

## Evidence (증거 — 이번 run에서 직접 관측)

1. **코드 경로**: `internal/cli/gateway_ui_state.go:15-92` — 시드 대상 키 `theme`·`lastOnboardingVersion`·`hasCompletedOnboarding`·`projects[<cwd>].hasTrustDialogAccepted`. 소스 우선순위: profile 선택 시 마지막 소스(profile)만 trust 차용, 미선택 시 `~/.claude.json`. 경로 매칭은 `currentLaunchCWD()` = `filepath.Clean(os.Getwd())` (gateway_launcher.go:10-16).
2. **실설정**: `jq -r '.projects["/Users/goos/MoAI/moai-adk-go"].hasTrustDialogAccepted' ~/.claude.json` → `true`. 전체 58 프로젝트에 trust 기록 존재.
3. **기동 직후 사망 family의 결정적 증거**: family b20da60c(2026-09-13 08:31:51Z 기동, **같은 세션이 08:31:57Z 첫 API 호출에서 400으로 사망 — 기동 6초, 모달 수동 수락이 물리적으로 불가한 타이밍**)의 `native/.claude.json`에 `"/Users/goos/MoAI/moai-adk-go": {"hasTrustDialogAccepted": true, ...}` 존재 → 시딩이 상호작용 없이 trust를 심었음.
4. **trust 시딩 착지 시점**: `git log -S hasTrustDialogAccepted -- internal/cli/gateway_ui_state.go` → ce79ef7ca (t649, bypass 수리 cd6bad85f 이전부터 존재).

## Baseline-attribution (baseline 귀속)

- 코드 판독: develop 트리 5e0f71175 기준, 이번 run.
- family 파일: 실가동 게이트웨이 families/e28832af·b20da60c (2026-09-13 17:31-32 기동분) 읽기만 함(수정 없음).

## Gaps (미검증 — 명시적)

1. TUI에서 trust 대화면이 "화면에 안 뜬다"는 것을 사람 눈으로 관측하진 않았다(헤드리스로는 관측 불가). 다음 실기동(`moai gpt` 신규 시작)에서 운영자 확인이 최종 확인이다 — 기동 즉시 대화면 없이 프롬프트가 떠야 정상.
2. 운영자가 레인을 named profile로 띄우는지 여부는 미확인 — named profile 운영이라면 프로필별 1회 trust 수락이 예상된다.

## Residual-risk (잔여 위험)

- named profile이 워크스페이스를 미신뢰 상태로 두면 1회 프롬프트가 뜬다 — 설계 동작이며, 이 정책(프로필이 전역 trust를 차용하지 않음)을 바꿀지는 보안 판단이 포함된 리드 결정 사항.
- trust 시딩은 cwd **문자열 정확 매칭**이다 — symlink 경로로 기동하면 실설정 키와 어긋나 시딩이 빠질 수 있다(t646 교훈: 정규화 비교). 현재 런처는 Clean만 적용하고 EvalSymlinks는 하지 않는다. 재발 시 이 지점부터 볼 것.

## 완료 판정

배차 요구(남은 workspace trust 경고 확인) 충족: 경고는 신규 결함이 아니라 t649 시딩으로 이미 억제된 상태이며, 기동-즉시-사망 family의 파일로 기계적 증거 확보. 코드 변경 불요 — 증거 커밋만으로 카드 종결.
