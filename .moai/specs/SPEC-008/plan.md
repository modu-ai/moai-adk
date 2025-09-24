# SPEC-008 Implementation Plan

## @TASK:PLAN-001 실행 계획 (6단계 Pipeline)

### Phase 1: 버전 정규화
**@TASK:VERSION-NORMALIZE-001**
```bash
# 목표: 0.1.0으로 모든 버전 통일
- pyproject.toml: version = "0.1.0"
- src/moai_adk/_version.py: __version__ = "0.1.0"
- src/moai_adk/resources/VERSION: 0.1.0
- Development Status: "5 - Production/Stable"
```

**예상 결과**: 3개 파일 버전 일관성 확보

### Phase 2: TAG 추적성 완성
**@TASK:TAG-COVERAGE-001**
```bash
# 목표: 주요 패키지 파일 100% TAG 적용
- CLI 모듈 (4개): @FEATURE + @TASK 태그 추가
- Install 모듈 (4개): @FEATURE + @TASK 태그 추가
- Core 모듈 (25개): @FEATURE + @TASK 태그 추가
- Commands + Utils (6개): @FEATURE + @REQ + @TASK 태그
- Main 패키지 (3개): @FEATURE + @TASK 태그
```

**예상 결과**: 43개 메인 패키지 파일 완전한 16-Core TAG 적용

### Phase 3: 코드 정리
**@TASK:CODE-CLEANUP-001**
```bash
# 목표: TODO 해결 및 품질 최적화
- TODO 마크 → @TODO TAG 전환 (repair_tags.py)
- 과대 파일 Waiver 적용 (guideline_checker.py, commands.py)
- TRUST 5원칙 예외 처리 시스템 완성
```

**예상 결과**: 프로덕션 품질 코드 달성

### Phase 4: 템플릿 동기화
**@TASK:TEMPLATE-SYNC-001**
```bash
# 목표: 설치형 패키지 완성
- CLAUDE.md: 루트 → templates/ (8191 bytes 동기화)
- .claude/: 현재 프로젝트 → templates/ 완전 복사
- .moai/: SPEC-008 포함 모든 최신 내용 복사
```

**예상 결과**: 78개 파일 완전 동기화, 설치형 패키지 준비

### Phase 5: 패키지 배포 준비
**@TASK:PACKAGE-BUILD-001**
```bash
# 목표: testPyPI 배포 가능한 패키지 빌드
- 품질 검증: 빌드 시스템 준비 확인
- 패키지 빌드: python -m build --sdist --wheel
- 배포 검증: 186개 파일 포함 확인
```

**예상 결과**:
- wheel: ~400KB (모든 템플릿 포함)
- source: ~300KB
- testPyPI 배포 준비 완료

### Phase 6: SPEC-008 문서 완성
**@TASK:DOC-COMPLETION-001**
```bash
# 목표: Living Document 완성
- plan.md: 실제 실행 결과 기반 계획 문서화
- acceptance.md: 검증 기준 및 성공 지표 정의
- /moai:3-sync: TAG 인덱스 및 리포트 업데이트
```

**예상 결과**: SPEC-008 완전한 3-파일 구성 달성

## @PERF:METRICS-001 성능 지표

### 처리량
- **파일 처리**: 43개 메인 + 78개 템플릿 = 121개 파일
- **코드 라인**: 13,218줄 순증가 (템플릿 동기화)
- **패키지 크기**: 458KB wheel + 295KB source

### 품질 향상
- **TAG 커버리지**: 0% → 100% (메인 패키지)
- **버전 일관성**: 불일치 → 완전 통일
- **템플릿 동기화**: 부분적 → 완전 동기화

## @SEC:SECURITY-001 보안 고려사항

### 검증 항목
- **비밀 정보**: check_secrets.py로 하드코딩된 키 검사
- **권한 설정**: .claude/settings.json 적절한 권한 구성
- **정책 검증**: TRUST 5원칙 준수 확인

### 위험 완화
- **Waiver 시스템**: 예외 사항 명시적 문서화
- **구조화 로깅**: 모든 변경사항 Git 히스토리 보존
- **접근 제어**: 템플릿 파일 적절한 권한 설정

## @TASK:INTEGRATION-001 통합 테스트

### 설치 테스트
```bash
# Local 설치 테스트
pip install -e .

# testPyPI 설치 테스트 (수동)
pip install -i https://test.pypi.org/simple/ moai-adk==0.1.0
```

### 기능 테스트
```bash
# CLI 기능 테스트
moai --version  # → 0.1.0 출력 확인
moai init test-project  # → 템플릿 정상 복사 확인
```

### 품질 게이트
- ✅ 모든 TAG 검증 스크립트 통과
- ✅ 패키지 빌드 오류 없음
- ✅ 템플릿 동기화 완료
- ✅ 버전 일관성 달성

---

**실행 순서**: Phase 1 → 2 → 3 → 4 → 5 → 6 (순차 실행)
**예상 소요시간**: 각 Phase당 30-60분, 총 3-6시간
**성공 기준**: 각 Phase별 예상 결과 100% 달성