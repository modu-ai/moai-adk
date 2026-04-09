# Implementation Plan: SPEC-SKILL-002

## 1. Overview

57개 스킬 중 41개에 영향을 미치는 최적화. 핵심: 16개 언어 스킬 → rules 전환, references/ 링크 복원, 이름 규약 준수.

## 2. Implementation Phases

### Phase 1: 16개 언어 스킬 → Rules 전환 (REQ-001 + REQ-007)

각 언어에 대해:
1. 기존 rule 파일 읽기 (`.claude/rules/moai/languages/{lang}.md`, ~47줄)
2. 기존 skill 파일 읽기 (`.claude/skills/moai-lang-{lang}/SKILL.md`, ~173줄)
3. 스킬의 풍부한 콘텐츠를 rule에 병합 (paths frontmatter 유지)
4. 불완전 스킬(swift, csharp, flutter)은 다른 스킬 수준으로 보완 후 병합
5. 스킬 디렉토리 삭제 (SKILL.md + references/)
6. CLAUDE.md output style의 Available agent types 섹션에서 언어 스킬 참조 확인

병합 전략:
- Rule의 paths frontmatter → 유지
- Rule의 간결한 규칙(MUST/MUST NOT) → 유지
- Skill의 상세 패턴, 프레임워크, 테스팅 전략 → 추가
- Skill의 Progressive Disclosure 메타데이터 → 불필요 (rule은 자동 로드)
- Skill의 trigger keywords → 불필요 (paths가 대체)

### Phase 2: moai-foundation-claude 이름 변경 (REQ-002)

1. 디렉토리 rename: `moai-foundation-claude/` → `moai-foundation-cc/`
2. SKILL.md 내부 name 필드: `moai-foundation-claude` → `moai-foundation-cc`
3. Grep으로 모든 참조 찾아 일괄 수정:
   - 에이전트 frontmatter skills 목록
   - 다른 스킬의 참조
   - CLAUDE.md
   - rules 파일

### Phase 3: references/ 링크 복원 (REQ-003)

41개 스킬의 SKILL.md에 references/ 링크 추가:
- 패턴: SKILL.md 하단에 `## References` 섹션 추가
- `For detailed {topic}, see ${CLAUDE_SKILL_DIR}/references/{filename}.md`
- 각 스킬의 references/ 내 실제 파일명 확인 후 매칭

### Phase 4: 대형 스킬 분리 + 코드 예제 이동 (REQ-004 + REQ-006)

4개 대형 스킬에서 코드 예제와 상세 참조를 references/로 이동:
- moai-platform-deployment: JSON/YAML 설정 예제 → references/config-examples.md
- moai-tool-svg: SVG 코드 패턴 → references/svg-patterns.md
- moai-design-tools: 설정 예제 → references/setup-guide.md (신규)
- moai-platform-chrome-extension: API 참조 → references/api-reference.md (신규)
- moai-library-shadcn: TSX 예제 → references/component-examples.md

### Phase 5: progressive_disclosure 추가 (REQ-005)

15개 스킬에 frontmatter 추가:
```yaml
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000
```

### Phase 6: Build + Test + 로컬 동기화

1. `make build` 실행
2. `go test ./...` 전체 통과 확인
3. 로컬 `.claude/` 동기화

## 3. Risk Analysis

| Risk | Impact | Mitigation |
|------|--------|-----------|
| 언어 rule 병합 시 콘텐츠 누락 | High | Before/After 라인 비교 |
| 이름 변경 참조 누락 | High | Grep 전수 검색 |
| references/ 링크 잘못된 파일명 | Medium | 실제 파일 존재 확인 |
| make build 실패 | Low | 즉시 감지 |

## 4. Success Metrics

| Metric | Before | Target |
|--------|--------|--------|
| 스킬 수 | 57 | 41 (-28%) |
| references/ 미링크 스킬 | 41 | 0 |
| "claude" 이름 위반 | 1 | 0 |
| progressive_disclosure 미설정 | 15 | 0 |
| 불완전 파일 | 3 | 0 |
| 코드 예제 위반 | 3 | 0 |
