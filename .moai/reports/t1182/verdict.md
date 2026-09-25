# t1182 — launcher auto 모드 안내 문구

## Claim

moai cc 도움말, moai glm의 auto 모드 거부 사유, 네 언어의 런처 문서에서 오래된 플랜·모델 버전 표기를 제거했다. acceptEdits는 프로젝트 기본값이 아닌 moai init 기본값으로 표시한다. GLM의 거부 동작과 moai cc 대안 안내는 유지했다.

## Evidence

- [Claude Code 공식 권한 모드 문서](https://code.claude.com/docs/en/permission-modes)는 auto 모드 지원 조건이 계정, 조직, 모델, 제공 경로에 따라 달라짐을 설명한다.
- go test ./internal/cli/ -run 'TestCharacterize_CC_HelpFlag|TestCharacterize_GLM_AutoMode' -count=1 -timeout=90s → ok github.com/modu-ai/moai-adk/internal/cli 0.843s, exit 0. 테스트는 도움말의 두 문구와 GLM의 두 플래그 철자, 기존 거부 및 대안 안내를 확인한다.
- 인라인 Python 검사 → AC-001..004 PASS; locale link/default line: [47, 47, 47, 47], exit 0.
- go vet ./internal/cli/ → 출력 없음, exit 0. gofmt -l 대상 Go 4파일 → 출력 없음, exit 0. git diff --check → 출력 없음, exit 0.
- git diff a520187f1 -- internal/cli/cc.go internal/cli/glm.go의 추가 줄에서 지정된 플랜·모델 버전 문자열을 검색 → 일치 없음. git diff --name-only a520187f1...HEAD → SPEC 세 파일, 지정한 코드·테스트·문서 8파일만 표시.
- git diff --name-only a520187f1..origin/develop -- 지정한 코드·테스트·문서 8파일 → 출력 없음, exit 0.
- 개정 AC-006 명령: git diff origin/develop...HEAD -- internal/cli/cc.go internal/cli/glm.go | grep '^+' | grep -v '^+++' | grep -cE 'Sonnet|Opus|Fable|[0-9]\.[0-9]|\b(Pro|Max|Team|Enterprise)\b' → 출력 0. grep 파이프라인 exit 1은 일치 줄이 없다는 뜻이다.
- 개정 AC-007의 git diff --name-only origin/develop...HEAD → 아래 12경로, 허용 목록 밖 0건:
    .moai/reports/t1182/verdict.md
    .moai/specs/SPEC-LAUNCHER-AUTOMODE-WORDING-001/plan.md
    .moai/specs/SPEC-LAUNCHER-AUTOMODE-WORDING-001/progress.md
    .moai/specs/SPEC-LAUNCHER-AUTOMODE-WORDING-001/spec.md
    docs-site/content/en/cli-reference/launchers.md
    docs-site/content/ja/cli-reference/launchers.md
    docs-site/content/ko/cli-reference/launchers.md
    docs-site/content/zh/cli-reference/launchers.md
    internal/cli/cc.go
    internal/cli/cc_test.go
    internal/cli/glm.go
    internal/cli/glm_new_test.go
- git rev-list --count --left-right origin/develop...HEAD → 4 5. AC-006/007의 내용은 통과했으나 SPEC의 origin/develop 흡수 선행 조건은 미충족이다.

## Baseline-attribution

기준 커밋은 SPEC 작성 시의 develop a520187f1이다. 현재 구현 커밋은 격리 작업트리 WT-auto-mode-help의 7b69ab3ca이며 위 명령은 이 작업트리에서 직접 실행했다.

## Gaps

origin/develop 통합과 원격 PR 검증은 아직 수행하지 않았다. 격리 작업트리에서 git merge --no-edit origin/develop을 시도했으나 PreToolUse BRANCH_GUARD_VIOLATION이 명령 실행 전에 차단했다. 부모 에이전트가 통합한 뒤 개정 AC-006/007을 다시 실행해야 최종 PASS를 선언할 수 있다.

## Residual-risk

공식 문서의 지원 조건은 계속 바뀔 수 있으므로, 런처는 버전을 고정하지 않고 공식 문서로 안내한다. 이번 수정 범위 밖의 프로필 설정 문구에는 별도 드리프트가 남아 있으며 SPEC §4에 기록했다.
