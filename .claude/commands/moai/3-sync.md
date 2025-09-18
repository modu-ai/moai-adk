---
description: 문서 동기화 및 TAG 시스템 업데이트 - Living Document와 16-Core TAG 완전 동기화
argument-hint: [auto|force|status] [target-path]
allowed-tools: Read, Write, Edit, MultiEdit, Bash, Task
---

# MoAI-ADK 0.2.0 Living Document 동기화

!@ doc-syncer 에이전트가 코드-문서 양방향 동기화와 16-Core TAG 시스템 관리를 완전 자동화합니다.

코드와 문서의 완벽한 일치성을 유지하고 16-Core TAG 시스템으로 추적성을 보장하는 핵심 명령어입니다.

## 🔄 빠른 시작

```bash
# 자동 증분 동기화 (기본값)
/moai:3-sync

# 완전 재동기화 (전체 프로젝트)
/moai:3-sync force

# 동기화 상태 확인
/moai:3-sync status
```

## 🎯 실행 모드

### Auto 모드 (기본값) - 증분 동기화
변경된 파일과 TAG만 선택적으로 동기화하여 빠른 처리를 보장합니다.

```
🔄 증분 동기화 실행 중...

📊 변경사항 스캔:
├── 수정된 파일: 3개 감지
├── 새로운 @TAG: 2개 발견
└── 삭제된 파일: 0개

⚡ 최적화된 처리: 델타만 동기화
```

### Force 모드 - 완전 재동기화
전체 프로젝트를 다시 스캔하고 모든 문서를 재생성합니다.

```
🔄 완전 재동기화 실행 중...

📂 전체 프로젝트 스캔:
├── 소스 파일: 45개 분석
├── 테스트 파일: 70개 검토
├── 문서 파일: 15개 갱신
└── 총 130개 파일 처리

🏗️ 전체 TAG 인덱스 재구축
```

## 🤖 doc-syncer 에이전트 자동화

**doc-syncer 에이전트**가 양방향 동기화를 완전 자동화:

### 코드 → 문서 동기화
- **API 문서**: 코드 변경 시 OpenAPI 스펙 자동 갱신
- **README**: 기능 추가/수정 시 사용법 자동 업데이트
- **아키텍처 문서**: 구조 변경 시 다이어그램 자동 갱신

### 문서 → 코드 동기화
- **SPEC 변경**: 요구사항 수정 시 관련 코드 마킹
- **TODO 추가**: 문서의 할일이 코드 주석으로 반영
- **TAG 업데이트**: 추적성 링크 자동 갱신

## 🏷️ 16-Core TAG 시스템 동기화

### TAG 카테고리별 처리
```markdown
📋 Primary Chain (주요 추적):
├── REQ → DESIGN → TASK → TEST
├── FEATURE → BUG → DEBT → TODO
└── API → UI → DATA → CONFIG

🔍 Quality Chain (품질 추적):
├── PERF → SEC → DOCS → TAG
└── 추적성 매트릭스 100% 유지
```

### 자동 검증 및 복구
- **끊어진 링크**: 자동 감지 및 수정 제안
- **중복 TAG**: 병합 또는 분리 옵션 제공
- **고아 TAG**: 참조 없는 태그 정리

## 📁 동기화 대상 파일

### 소스 코드 연동
```
src/ → docs/api/
├── models/*.py → API 스키마 문서
├── services/*.py → 비즈니스 로직 문서
├── controllers/*.py → 엔드포인트 문서
└── utils/*.py → 유틸리티 문서
```

### 테스트 연동
```
tests/ → docs/testing/
├── unit/ → 단위 테스트 보고서
├── integration/ → 통합 테스트 가이드
└── e2e/ → E2E 시나리오 문서
```

### SPEC 연동
```
.moai/specs/ → 프로젝트 문서
├── SPEC-*/spec.md → 요구사항 문서
├── SPEC-*/plan.md → 설계 문서
└── SPEC-*/tasks.md → 구현 가이드
```

## 📊 동기화 결과

### 성공적인 동기화
```bash
✅ Living Document 동기화 완료!

📊 처리 결과:
├── 업데이트된 파일: 8개
├── 생성된 문서: 3개
├── 수정된 TAG: 12개
└── 추적성 검증: 100% 통과

🏷️ TAG 시스템 상태:
├── Primary Chain: 완전 연결
├── Quality Chain: 완전 연결
├── 고아 TAG: 0개
└── 끊어진 링크: 0개
```

### 부분 동기화 (문제 감지)
```bash
⚠️ 부분 동기화 완료 (문제 발견)

🔴 해결 필요한 문제:
├── 끊어진 링크: 2개
│   ├── REQ:USER-LOGIN-001 → 참조 파일 없음
│   └── TASK:AUTH-IMPL-005 → 구현 파일 삭제됨
├── 중복 TAG: 1개
│   └── API:USERS-001 (2개 파일에 중복)
└── 고아 TAG: 3개

🛠️ 자동 수정 옵션:
1. 끊어진 링크 복구
2. 중복 TAG 병합
3. 고아 TAG 정리

계속 진행하시겠습니까? (y/N)
```

## 🔧 고급 옵션

### 특정 경로 동기화
```bash
# 특정 SPEC만 동기화
/moai:3-sync auto SPEC-001

# 특정 디렉토리만
/moai:3-sync force src/auth/

# 문서 파일만
/moai:3-sync auto docs/
```

### 동기화 상태 확인
```bash
/moai:3-sync status

📊 현재 동기화 상태:
├── 마지막 동기화: 2시간 전
├── 변경된 파일: 5개 (동기화 필요)
├── TAG 상태: 양호
└── 권장 조치: /moai:3-sync auto
```

## 🔄 완료 후 다음 단계

### Git 커밋 권장
```bash
✅ 동기화 완료!

🎯 권장 다음 단계:
> git add .
> git commit -m "docs: sync living documents and TAG system"

📝 변경된 파일:
├── README.md (기능 설명 업데이트)
├── docs/api/ (3개 API 문서 갱신)
├── .moai/indexes/tags.json (TAG 인덱스 업데이트)
└── CLAUDE.md (프로젝트 상태 반영)
```

### 다음 개발 사이클
```bash
🔄 개발 사이클 완료!

전체 MoAI-ADK 워크플로우:
✅ /moai:1-spec → SPEC 작성
✅ /moai:2-build → TDD 구현
✅ /moai:3-sync → 문서 동기화

🎉 다음 기능 개발 준비 완료
> /moai:1-spec "다음 기능 설명"
```

## ⚠️ 에러 처리

### 파일 충돌 감지
```bash
🔴 파일 충돌 감지:
- README.md: 수동 편집과 자동 생성 충돌

해결 옵션:
1. 수동 편집 유지 [기본값]
2. 자동 생성으로 덮어쓰기
3. 대화형 병합 모드
```

### TAG 시스템 오류
```bash
❌ TAG 시스템 오류:
- 인덱스 파일 손상: .moai/indexes/tags.json

자동 복구 중...
✅ 백업에서 복구 완료
```

## 🔁 응답 구조

출력은 반드시 3단계 구조를 따릅니다:
1. **Phase 1 Results**: 동기화 결과 및 변경사항
2. **Phase 2 Plan**: TAG 시스템 업데이트 계획
3. **Phase 3 Implementation**: Git 커밋 및 다음 단계

이 명령어는 MoAI-ADK 0.2.0의 마지막 단계로, 완벽한 문서-코드 일치성을 보장합니다.