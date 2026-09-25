# t1185 — Claude Code deny 규칙 문법

## Claim

템플릿의 rm -rf 대상 경로 규칙 3개에서 와일드카드 *와 접미사 :*를 함께 쓰던 문법을 제거했다. 루트, 홈, Windows 드라이브 경로의 기존 deny 범위는 유지한다.

## Evidence

- 수정 전 실제 moai cc -w t1185 -- --print --model haiku 'Reply OK' 실행에서 Claude Code 2.1.282가 루트, 홈, Windows 경로 규칙에 대해 "mixes * with the trailing :* prefix syntax" 경고를 출력했다.
- [Claude Code 공식 권한 문서](https://code.claude.com/docs/en/permissions)의 Wildcard patterns 항목은 *의 와일드카드 의미와 :*의 끝 접미사 의미를 설명한다.
- go test ./internal/template -run '^TestSettingsTemplateDenyWildcardSyntax$' -count=1 -v -timeout=60s → macOS, Linux, Windows 렌더링 하위 테스트 모두 PASS.
- go test ./internal/template -count=1 -timeout=90s → ok github.com/modu-ai/moai-adk/internal/template 65.953s.
- git diff --cached --check → 출력 없음, exit 0.

## Baseline-attribution

위 결과는 WT-template-deny-rule-syntax 작업트리에서 직접 관측했다. 초기 기준은 8a0ce2d14이며 구현 커밋은 89edb25b0이다.

## Gaps

수정된 템플릿으로 새 프로젝트를 생성해 Claude Code를 다시 기동하는 실세션 검증은 수행하지 않았다. 테스트는 렌더링된 규칙 문자열과 JSON을 확인한다.

## Residual-risk

Claude Code 권한 규칙은 명령 문자열 패턴이므로, 래퍼 명령이나 다른 철자의 파괴적 명령까지 모두 차단하는 보안 경계로 해석하면 안 된다.
