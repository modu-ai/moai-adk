# SPEC-UPDATE-002 Acceptance Criteria

## AC-1: Figma MCP 공식 전환

**Given** moai-design-tools/reference/figma.md가 업데이트된 후
**When** 에이전트가 Figma MCP 워크플로우를 참조하면
**Then** 공식 Remote MCP 서버(mcp.figma.com)와 검증된 도구 목록만 안내해야 한다

검증 항목:
- [ ] figma.get_file_metadata, figma.get_components 등 가상 API 잔재 없음
- [ ] generate_figma_design, get_design_context 등 공식 도구 문서화
- [ ] `claude plugin install figma@claude-plugins-official` 설치 안내 포함
- [ ] Code-to-Canvas v1 제한사항 명시

## AC-2: Pencil MCP 팩트 수정

**Given** moai-design-tools/reference/pencil-renderer.md가 업데이트된 후
**When** 에이전트가 .pen 파일 작업을 안내하면
**Then** .pen 파일이 순수 JSON 형식임을 정확히 안내해야 한다

검증 항목:
- [ ] "encrypted" 문구 완전 제거
- [ ] ".pen files are pure JSON" 명시
- [ ] 수동 mcpServers 설정 안내 제거
- [ ] UI Kit 4종 선택 가이드 포함

## AC-3: Pencil-to-Code 가상 API 제거

**Given** moai-design-tools/reference/pencil-code.md가 업데이트된 후
**When** 에이전트가 Pencil→Code 워크플로우를 실행하면
**Then** 실제 prompt 기반 워크플로우만 사용해야 한다

검증 항목:
- [ ] pencil.export_to_react 잔재 없음
- [ ] pencil.config.js 잔재 없음
- [ ] batch_get → 구조 분석 → React/Tailwind 생성 워크플로우 문서화

## AC-4: Anti-AI Slop 디자인 방향성

**Given** moai-domain-uiux/modules/web-interface-guidelines.md가 업데이트된 후
**When** 에이전트가 UI 디자인을 생성하면
**Then** Purpose→Tone→Constraints→Differentiation 프로세스를 따라야 한다

검증 항목:
- [ ] 디자인 방향성 프레임워크 섹션 존재
- [ ] 금지 패턴 목록 (Inter/Roboto/Arial, 보라색 그라데이션) 포함
- [ ] 스타일 극단 가이드 포함
- [ ] 모션/마이크로인터랙션 섹션 존재
- [ ] 모바일 우선 UX 패턴 확장

## AC-5: 기술 스택 버전

**Given** moai-domain-uiux/SKILL.md가 업데이트된 후
**When** 에이전트가 기술 스택을 참조하면
**Then** TypeScript 5.9+, Tailwind CSS 4.x를 반영해야 한다

검증 항목:
- [ ] TypeScript 5.5 → 5.9+ 업데이트
- [ ] Tailwind CSS 3.4 → 4.x 업데이트
- [ ] Hugeicons 아이콘 라이브러리 추가
- [ ] Nova 프리셋 상호참조

## AC-6: 중복 제거

**Given** moai-domain-uiux/modules/design-system-tokens.md가 업데이트된 후
**When** Pencil MCP 관련 내용이 참조되면
**Then** moai-design-tools로의 상호참조만 존재해야 한다

검증 항목:
- [ ] Pencil MCP Integration Workflow 섹션(~115줄) 제거
- [ ] moai-design-tools 상호참조 존재

## AC-7: 빌드 및 테스트

**Given** 모든 스킬 파일 수정 완료 후
**When** `make build && go test ./internal/template/...` 실행 시
**Then** 빌드 성공 및 모든 테스트 통과해야 한다

검증 항목:
- [ ] `make build` 성공
- [ ] `go test ./internal/template/...` 전체 통과
- [ ] 가상 API 문자열 grep 결과 0건

## Edge Cases

### EC-1: Context7 도구 검증 실패

**Given** Context7에서 Pencil/Figma 도구 이름 검증이 실패하면
**Then** 미검증 도구는 스킬에 포함하지 않고, 검증된 도구만 문서화한다

### EC-2: Figma generate_figma_design 가용성

**Given** generate_figma_design이 일부 세션에서 사용 불가능하면
**Then** "Remote MCP server only" 제한사항을 명시하고 fallback 안내한다
