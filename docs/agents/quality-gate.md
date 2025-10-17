# @DOC:AGENT-QUALITY-GATE-001 | Chain: @SPEC:DOCS-003 -> @DOC:AGENT-QUALITY-GATE-001

# quality-gate 🛡️

**모델**: Haiku
**페르소나**: 품질 보증 엔지니어 (QA Engineer)
**전문 영역**: TRUST 원칙 검증, 코드 품질 검증, TAG 체인 무결성

## 역할

TRUST 원칙과 프로젝트 표준을 자동으로 검증하는 품질 게이트입니다.

## 호출 방법

```bash
# TDD 구현 완료 후 자동 호출 (Phase 3)
/alfred:2-build SPEC-ID

# 동기화 전 사전 검증 (Phase 1, 조건부)
/alfred:3-sync
```

## 주요 작업

- **TRUST 원칙 검증**: trust-checker 스크립트 실행 및 결과 파싱
  - Testable: 테스트 커버리지 ≥ 85%
  - Readable: 코드 가독성 (파일≤300 LOC, 함수≤50 LOC)
  - Unified: 아키텍처 통합성
  - Secured: 보안 취약점 없음
  - Traceable: @TAG 체인 무결성
- **코드 스타일**: 린터(ESLint/Pylint) 실행 및 검증
- **테스트 커버리지**: 언어별 커버리지 도구 실행
- **TAG 체인 검증**: 고아 TAG, 누락된 TAG 확인
- **의존성 검증**: 보안 취약점 확인

## 산출물

- 품질 검증 리포트 (Pass/Warning/Critical)
- 파일:라인 정보 포함 상세 보고서
- 수정 제안 (실행 가능한 구체적 방법)

---

**다음**: [doc-syncer →](doc-syncer.md)
