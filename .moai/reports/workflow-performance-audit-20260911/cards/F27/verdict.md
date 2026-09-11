# F27 카드 판정

## Claim

기술 키워드 수만으로 clarity interview를 건너뛰지 않고 scope·constraints/non-goals·acceptance/stopping condition·authorization/ownership 완결도를 확인한다.

## Evidence

- `bash .claude/hooks/tests/test-interview-completeness-contract.sh`
- 관찰된 출력: `PASS: interview skipping requires intent completeness, not keyword count`

## Baseline-attribution

위 출력은 `WT-workflow-audit-f27`의 context-discovery 지침을 기준으로 실행하였다.

## Gaps

실제 키워드 과다·짧은 완결 요청 fixture에서 재질문 수는 관찰하지 않았다.

## Residual-risk

필드 추출과 명시적 empty 처리의 실제 parser wiring은 별도 runtime 시험이 필요하다.
