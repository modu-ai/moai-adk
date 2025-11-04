---
name: moai-alfred-autofixes-ko
version: 1.0.0
created: 2025-11-05
updated: 2025-11-05
status: active
description: 자동 코드 수정, 병합 충돌 및 사용자 승인 워크플로우의 안전 프로토콜
keywords: ['auto-fix', 'merge-conflicts', 'safety', 'approval', 'protocol']
allowed-tools:
  - Read
  - Write
  - Bash
  - AskUserQuestion
---

# 자동 수정 및 병합 충돌 프로토콜

Alfred가 자동으로 코드를 수정할 수 있는 문제(병합 충돌, 덮쓰기 변경, 사용되지 않는 코드 등)를 감지하면, 변경하기 전에 다음 프로토콜을 따르세요:

## 1단계: 분석 및 보고

- git 기록, 파일 내용, 로직을 사용하여 문제를 철저히 분석
- 명확한 보고서(일반 텍스트, NO 마크다운) 작성:
  - 문제의 근본 원인
  - 영향받는 파일
  - 제안된 변경사항
  - 영향 분석

**보고서 예시 형식**:
```
병합 충돌 감지:

근본 원인:
- 커밋 c054777b가 develop에서 언어 감지를 제거함
- 병합 커밋 e18c7f98 (main → develop)가 해당 라인을 다시 도입함

영향:
- .claude/hooks/alfred/shared/handlers/session.py
- src/moai_adk/templates/.claude/hooks/alfred/shared/handlers/session.py

제안된 수정:
- detect_language() import와 호출 제거
- "🐍 언어: {language}" 표시 라인 삭제
- 로컬 + 패키지 템플릿 동기화
```

## 2단계: 사용자 확인 (AskUserQuestion)

- 분석 결과를 사용자에게 제시
- AskUserQuestion을 사용하여 명시적 승인 얻기
- 옵션은 명확해야 함: "이 수정을 계속할까요?" YES/NO 선택지
- 진행하기 전에 사용자 응답 대기

## 3단계: 승인 후만 실행

- 사용자가 확인한 후에만 파일 수정
- 로컬 프로젝트와 패키지 템플릿 모두에 변경 적용
- `/`와 `src/moai_adk/templates/` 간 일관성 유지

## 4단계: 전체 컨텍스트로 커밋

- 상세한 메시지로 커밋 생성:
  - 어떤 문제가 수정되었는지
  - 왜 발생했는지
  - 어떻게 해결되었는지
- 관련 커밋이 있다면 참조

## 중요한 규칙

- ❌ 사용자 승인 없이 자동 수정 금지
- ❌ 보고서 단계 건너뛰기 금지
- ✅ 항상 먼저 결과 보고
- ✅ 항상 사용자 확인 요청 (AskUserQuestion)
- ✅ 항상 로컬 + 패키지 템플릿 함께 업데이트

## 템플릿 동기화 규칙

**패키지 템플릿이 진실의 소스입니다**:
- `src/moai_adk/templates/`의 변경사항은 로컬 프로젝트 경로로 동기화되어야 함
- 로컬 수정사항은 패키지 템플릿 변경사항을 덮어써서는 안 됨
- 로컬과 패키지 템플릿 간 충돌이 발생하면 패키지 템플릿 변경사항을 우선

**동기화 체크리스트**:
- [ ] `src/moai_adk/templates/` 경로에 변경 적용
- [ ] 해당 로컬 프로젝트 경로에 변경 적용
- [ ] 파일 내용 확인 동일
- [ ] Git 커밋이 두 경로 업데이트 확인