# 메모리 시스템 인덱스(프로젝트용)

> 이 문서는 프로젝트 메모리 파일들을 빠르게 탐색하기 위한 지도입니다. Claude Code 메모리 계층과 임포트 규칙, 역할/환경별 구성, 문제 해결 팁을 요약합니다.

## 1) 메모리 계층(요약)
- 조직 정책(Enterprise): IT/DevOps가 배포하는 조직 공통 정책(시스템 경로)
- 프로젝트 메모리(Project): 저장소 루트의 `CLAUDE.md`와 이 폴더의 문서들(팀 공유)
- 사용자 메모리(User): `~/.claude/CLAUDE.md` (개인 선호, 저장소 외부)
- CLAUDE.local.md: 더 이상 권장하지 않음(Deprecated)

우선순위는 조직 → 프로젝트 → 사용자 순서로 적용됩니다.

## 2) 임포트(@path) 규칙(요약)
- 상대/절대 경로, 홈 디렉토리(`~/`) 임포트 지원
- 최대 5-hop 제한(순환/과도한 중첩 방지)
- 코드 스팬/코드 블록 내부는 임포트 처리하지 않음

예시
```
@.claude/memory/project_guidelines.md
@docs/api-guidelines.md
```

## 3) 역할/환경별 구성 패턴
- 역할 기반(Frontend/Backend/QA/PO/DevOps) 지침은 별도 파일로 분리 후 필요 시 임포트
- 환경별(Staging/Prod) 운영 가이드는 배포 문서에서 분리 임포트

샘플
```
# 역할 기반
@.claude/memory/coding_standards/typescript.md
@.claude/memory/coding_standards/python.md

# 환경 기반
@docs/deployment/staging.md
@docs/deployment/production.md
```

## 4) 문서 지도
- 프로세스 표준: @.claude/memory/three_phase_process.md
- 테스트 표준: @.claude/memory/tdd_guidelines.md
- 커밋 표준: @.claude/memory/git_commit_rules.md
- 보안 표준: @.claude/memory/security_rules.md
- 코딩 표준(공통): @.claude/memory/coding_standards.md
- 팀 운영: @.claude/memory/team_conventions.md
- Git 워크플로우: @.claude/memory/git_workflow.md
- 운영 원칙(전체 가이드): @.claude/memory/project_guidelines.md
- 마스터 원칙 요약: @.claude/memory/software_principles.md
- Bash 도구 모음: @.claude/memory/bash_commands.md
- 프로젝트 메모리 템플릿: `src/moai_adk/resources/templates/.moai/_templates/memory/*.template.md`

> `moai init` 실행 시 공통 템플릿(`common.md`)과 선택한 기술 스택에 해당하는 문서(예: `backend-python.md`, `frontend-react.md`)가 `.moai/memory/`에 자동 생성됩니다.

## 5) 문제 해결(트러블슈팅)
- 임포트가 동작하지 않을 때: 경로 확인, 5-hop 초과 여부 점검
- 메모리 충돌: 계층/우선순위 확인(조직→프로젝트→사용자)
- 성능 문제: 깊은 임포트 체인/순환 의심 지점 축약

디버깅 명령(Claude Code 내부)
```
/memory            # 로드된 메모리 확인/편집
claude config list # 메모리/설정 계층 확인
```

## 6) 운영 원칙
- 저장소에는 ‘팀 공유’ 지침만 포함(개인 선호는 사용자 메모리로 분리)
- 큰 문서는 기능별로 분리하고 CLAUDE.md에서 임포트
- 문서는 주기적으로 점검하고 불용/중복 내용을 정리
- 템플릿 추가 시 `_templates/memory/<layer>-<tech>.template.md` 형식을 따르고 이 README를 갱신
