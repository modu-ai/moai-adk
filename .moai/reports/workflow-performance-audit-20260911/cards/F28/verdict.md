# F28 카드 판정

## Claim

조건부로만 필요한 대형 workflow/security/skill rule 네 개를 path-scoped로 전환하고, always-loaded 예외는 적용 범위·도달성·실제 `InstructionsLoaded` 관측을 함께 기록하도록 계약화하였다.

## Evidence

- `bash .claude/hooks/tests/test-rule-loading-budget.sh`
- 관찰된 출력: `PASS: large conditional rules declare path-scoped loading and measurement gaps`

## Baseline-attribution

위 출력은 `WT-workflow-audit-f28`의 rule frontmatter과 loading-budget 계약을 기준으로 실행하였다.

## Gaps

현재 호스트에서 실제 InstructionsLoaded 목록·토큰 사용량을 이 카드 시험으로 읽지는 않았다.

## Residual-risk

path-scoped 전환은 runtime loader의 glob 의미와 실제 prompt reachability를 별도 환경에서 재검증해야 하며, always-loaded 핵심 rule은 안전성 때문에 남아 있다.
