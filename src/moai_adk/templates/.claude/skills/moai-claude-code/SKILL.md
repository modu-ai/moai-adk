---
name: moai-claude-code
description: Claude Code 5가지 컴포넌트 생성 및 관리 - Agent, Command, Skill, Plugin, Settings
  템플릿 기반 정확한 파일 생성
allowed-tools:
- Read
- Write
- Edit
---

# MoAI Claude Code Manager

Claude Code의 5가지 핵심 컴포넌트를 공식 표준에 맞게 생성하고 관리합니다.

## 지원 컴포넌트

1. **Agent** (.claude/agents/) - 전문 에이전트
2. **Command** (.claude/commands/) - 슬래시 커맨드
3. **Skill** (.claude/skills/) - 재사용 기능 모듈
4. **Plugin** (settings.json의 mcpServers) - MCP 서버 통합
5. **Settings** (.claude/settings.json) - 권한 및 훅 설정

## 템플릿 특징

MoAI-ADK 통합 프로덕션급 템플릿 (5개)

- 완전 상세 (완전하고 실무 사용 가능)
- MoAI-ADK 워크플로우 통합
- 복사-붙여넣기 즉시 사용
- 검증 및 트러블슈팅 가이드 포함

## 사용법

### Agent 생성
"spec-builder Agent를 생성해주세요"

### Settings 최적화
"Python 프로젝트용 settings.json을 생성해주세요"

### 전체 검증
"모든 Claude Code 설정을 검증해주세요"

## 상세 문서

- **reference.md**: 컴포넌트별 작성 가이드
- **examples.md**: 실전 예제 모음
- **templates/**: 5개 프로덕션급 템플릿
- **scripts/**: Python 검증 스크립트 (선택적)

## 작동 방식

1. 사용자 요청 분석 → 컴포넌트 유형 파악
2. 적절한 템플릿 선택 (templates/ 디렉토리)
3. 플레이스홀더 치환 및 파일 생성
4. 자동 검증 (선택적, scripts/ 실행)

## 핵심 원칙

- **공식 표준 준수**: Anthropic 가이드라인 완벽 준수
- **할루시네이션 방지**: 검증된 템플릿만 사용
- **최소 권한**: 필요한 도구만 명시
- **보안 우선**: 민감 정보 환경변수 관리

---

**공식 문서**: https://docs.claude.com/en/docs/claude-code/skills
**버전**: 1.0.0
