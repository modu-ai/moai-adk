# MoAI-ADK Commands Reference

Alfred가 사용하는 6개의 핵심 MoAI-ADK 커맨드. SPEC-First TDD 실행을 위한 필수 도구다.

## `/moai:0-project` - 프로젝트 초기화

**목적**: 프로젝트 구조 초기화 및 설정 생성

**위임**: `project-manager`

**사용법**:
```
/moai:0-project
/moai:0-project --with-git
```

**출력**: `.moai/` 디렉토리 + config.json

**다음 단계**: SPEC 생성 준비 완료

---

## `/moai:1-plan` - SPEC 생성

**목적**: EARS 포맷 SPEC 문서 생성

**위임**: `spec-builder`

**사용법**:
```
/moai:1-plan "사용자 인증 엔드포인트 구현 (JWT)"
```

**출력**: `.moai/specs/SPEC-001/spec.md` (EARS 포맷 문서)

**필수**: 완료 후 반드시 `/clear` 실행 (45-50K 토큰 절약)

---

## `/moai:2-run` - TDD 구현

**목적**: RED-GREEN-REFACTOR 사이클 실행

**위임**: `tdd-implementer`

**사용법**:
```
/moai:2-run SPEC-001
```

**프로세스**:
1. RED: 실패하는 테스트 작성
2. GREEN: 최소 코드로 통과
3. REFACTOR: 최적화 및 정리

**출력**: 구현된 코드 + 테스트 + 품질 리포트

**조건**: 테스트 커버리지 ≥ 85% (TRUST 5)

---

## `/moai:3-sync` - 문서 동기화

**목적**: API 문서 및 프로젝트 아티팩트 자동 생성

**위임**: `docs-manager`

**사용법**:
```
/moai:3-sync SPEC-001
```

**출력**:
- API 문서 (OpenAPI 포맷)
- 아키텍처 다이어그램
- 프로젝트 리포트

---

## `/moai:9-feedback` - 개선 의견 수집

**목적**: 오류 분석 및 개선사항 제안 수집

**위임**: `quality-gate`

**사용법**:
```
/moai:9-feedback
/moai:9-feedback --analyze SPEC-001
```

**용도**: MoAI-ADK 지속적 개선, 오류 복구

---

## `/moai:99-release` - 프로덕션 릴리스

**목적**: 릴리스 아티팩트 생성 및 배포 준비

**위임**: `release-manager`

**사용법**:
```
/moai:99-release
/moai:99-release --patch
/moai:99-release --minor
```

**검증**: 모든 품질 게이트 통과 필수 (테스트 ≥85%, 문서 완전)

---

## 필수 워크플로우

```
1. /moai:0-project              # 프로젝트 초기화
2. /moai:1-plan "설명"          # SPEC 생성
3. /clear                       # 컨텍스트 초기화 (필수)
4. /moai:2-run SPEC-001         # TDD 구현
5. /moai:3-sync SPEC-001        # 문서 생성
6. /moai:9-feedback             # 피드백 수집
7. /moai:99-release             # 프로덕션 배포
```

---

## Context 초기화 규칙

- `/moai:1-plan` 후 **반드시** `/clear` 실행
- Context > 150K 일 때 `/clear` 실행
- 대화 > 50메시지 후 `/clear` 실행

각 커맨드 사용법의 상세한 정보는 CLAUDE.md를 참고한다.
