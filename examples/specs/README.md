# MoAI-ADK SPEC Examples

이 디렉토리는 MoAI-ADK에서 사용하는 SPEC 작성 방법을 학습할 수 있는 예제 모음입니다.

## 📚 예제 SPEC 목록

### SPEC-002: Python 코드 품질 개선 시스템
- **파일**: [SPEC-002-quality-system.md](./SPEC-002-quality-system.md)
- **학습 포인트**:
  - EARS (Environment, Assumptions, Requirements, Success criteria) 형식 활용
  - TRUST 5원칙 구현을 위한 구체적 요구사항 정의
  -  TAG 시스템을 통한 추적성 확보
  - TDD Red-Green-Refactor 자동화 설계

### SPEC-010: 온라인 문서 사이트 제작
- **파일**: [SPEC-010-documentation.md](./SPEC-010-documentation.md)
- **학습 포인트**:
  - 복잡한 시스템의 자동화 요구사항 정의
  - Living Document 원칙을 통한 코드-문서 동기화
  - CI/CD 파이프라인과 통합된 문서 시스템 설계
  - 사용자 경험(UX) 요구사항 구체화

## 🎯 SPEC 작성 가이드라인

### 1. EARS 형식 준수
```markdown
## Environment (환경)
- 실행 환경, 기술 스택, 플랫폼 제약

## Assumptions (가정사항)
- 전제 조건, 사용자 지식 수준, 기존 시스템 상태

## Requirements (요구사항)
- 구체적이고 측정 가능한 기능 요구사항
- "WHEN ... THE SYSTEM SHALL ..." 형식 권장

## Success criteria (성공 기준)
- 체크리스트 형태의 검증 가능한 기준
- 실패 시나리오 포함
```

### 2. TAG 시스템 활용 (v5.0)
```markdown
@SPEC:USER-001       # SPEC 문서와 요구사항
@TEST:USER-001       # 테스트 코드 및 검증
@CODE:USER-001       # 실제 구현 코드
@DOC:USER-001        # 문서화 및 주석
```

### 3. 추적성 확보
- 각 요구사항을 TAG로 연결
- 상위 SPEC과의 관계 명시
- 구현 완료 후 완성 태그 추가

## 📝 SPEC 템플릿

새로운 SPEC을 작성할 때는 다음 템플릿을 사용하세요:

```markdown
# SPEC-XXX: [제목]

## @SPEC:[CATEGORY]-XXX 프로젝트 컨텍스트

### 배경
- 현재 상황 설명
- 해결해야 할 문제

### 목표
- 구체적이고 측정 가능한 목표

### Environment (환경)
- 기술 스택
- 플랫폼 요구사항

### Assumptions (가정사항)
- 전제 조건
- 제약사항

## @SPEC:[CATEGORY]-XXX 요구사항 명세

### R1. [요구사항 제목]
**WHEN** [조건]
**THE SYSTEM SHALL** [동작]

**상세 요구사항:**
- 구체적 기능 목록

## @TEST:[CATEGORY]-XXX 검증 기준

### 성공 기준
- [ ] 검증 가능한 체크리스트

### 실패 시나리오
- 예상 실패 케이스와 대응 방안

## 추적성 체인 (v5.0)
```
@SPEC:[CATEGORY]-XXX → @TEST:[CATEGORY]-XXX → @CODE:[CATEGORY]-XXX → @DOC:[CATEGORY]-XXX
```
```

## 🎓 학습 권장사항

1. **SPEC-002부터 시작**: 품질 시스템 예제로 기본 구조 학습
2. **SPEC-010으로 심화**: 복잡한 자동화 시스템 설계 방법 학습
3. **TAG 시스템 연습**: 실제 프로젝트에서 TAG 추적성 구현
4. **점진적 개선**: 작은 SPEC부터 시작해서 단계적으로 확장

## 💡 추가 리소스

- **MoAI-ADK 개발 가이드**: [MOAI-ADK-GUIDE.md](../../MOAI-ADK-GUIDE.md)
- **TRUST 5원칙**: [development-guide.md](../../.moai/memory/development-guide.md)
- ** TAG 시스템**: [tags.json](../../.moai/indexes/tags.json)

이 예제들을 통해 체계적이고 추적 가능한 SPEC 작성 방법을 익히고, MoAI-ADK의 Spec-First TDD 개발 프로세스를 마스터하세요!