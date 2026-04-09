# Acceptance Criteria: SPEC-SKILL-002

## AC-001: 언어 스킬 → Rules 전환

```gherkin
Given 16개 moai-lang-* 스킬 디렉토리가 삭제되었을 때
When .claude/skills/ 디렉토리를 검사하면
Then moai-lang-* 디렉토리가 존재하지 않아야 한다

Given 16개 언어 rule 파일이 병합되었을 때
When .claude/rules/moai/languages/ 파일들을 검사하면
Then 각 rule이 기존 스킬의 핵심 패턴을 포함해야 한다
And paths frontmatter가 유지되어야 한다
And 각 rule의 라인 수가 100줄 이상이어야 한다 (기존 ~47줄에서 증가)
```

## AC-002: 이름 변경

```gherkin
Given moai-foundation-claude가 moai-foundation-cc로 변경되었을 때
When 전체 프로젝트에서 "moai-foundation-claude"를 검색하면
Then 매칭 결과가 0건이어야 한다

When .claude/skills/moai-foundation-cc/SKILL.md를 검사하면
Then name 필드가 "moai-foundation-cc"이어야 한다
```

## AC-003: references/ 링크

```gherkin
Given references/ 디렉토리가 있는 모든 스킬이 수정되었을 때
When SKILL.md에서 "CLAUDE_SKILL_DIR" 또는 "references/"를 검색하면
Then references/를 가진 모든 스킬에서 매칭이 발생해야 한다
```

## AC-004: progressive_disclosure 설정

```gherkin
Given 모든 스킬에 progressive_disclosure가 추가되었을 때
When 전체 스킬의 frontmatter를 검사하면
Then 모든 스킬에 progressive_disclosure: enabled: true가 존재해야 한다
```

## AC-005: 코드 예제 제거

```gherkin
Given moai-tool-svg, moai-platform-deployment의 코드 블록이 이동되었을 때
When SKILL.md의 본문을 검사하면
Then 개념 설명용 코드 블록(```로 시작하는 3줄 이상)이 5개 이하여야 한다
```

## AC-006: Build + Test

```gherkin
Given 모든 변경이 완료되었을 때
When make build를 실행하면 Then 성공해야 한다
When go test ./...를 실행하면 Then 모든 테스트가 통과해야 한다
```

## AC-007: 스킬 수 감소

```gherkin
Given 모든 변경이 완료되었을 때
When .claude/skills/ 디렉토리의 SKILL.md 파일 수를 세면
Then 41개 이하여야 한다 (기존 57개에서 16개 감소)
```
