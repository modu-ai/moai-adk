# SPEC-008 Acceptance Criteria

## @TEST:ACCEPTANCE-001 수락 기준 (Given-When-Then)

### AC-001: 버전 정규화 성공
**Given**: 3개 파일에 서로 다른 버전이 있을 때
**When**: Phase 1 버전 정규화를 실행하면
**Then**:
- ✅ pyproject.toml의 version이 "0.1.0"이다
- ✅ src/moai_adk/_version.py의 __version__이 "0.1.0"이다
- ✅ src/moai_adk/resources/VERSION이 "0.1.0"이다
- ✅ Development Status가 "5 - Production/Stable"이다

### AC-002: TAG 추적성 완성
**Given**: 메인 패키지 파일들에 @TAG가 없을 때
**When**: Phase 2 TAG 추적성 완성을 실행하면
**Then**:
- ✅ CLI 모듈 4개 파일에 @FEATURE + @TASK 태그가 있다
- ✅ Install 모듈 4개 파일에 @FEATURE + @TASK 태그가 있다
- ✅ Core 모듈 25개 파일에 적절한 @TAG가 있다
- ✅ Commands 모듈 3개 파일에 @FEATURE + @REQ + @TASK 태그가 있다
- ✅ Utils 모듈 3개 파일에 @FEATURE + @TASK 태그가 있다
- ✅ 메인 모듈 3개 파일에 @FEATURE + @TASK 태그가 있다
- ✅ 총 43개 메인 패키지 파일의 TAG 커버리지가 100%이다

### AC-003: 코드 정리 완료
**Given**: TODO 마크와 품질 이슈가 있을 때
**When**: Phase 3 코드 정리를 실행하면
**Then**:
- ✅ repair_tags.py의 TODO가 @TODO:CREATE-TEST-001로 전환된다
- ✅ guideline_checker.py에 대한 Waiver가 적용된다
- ✅ cli/commands.py에 대한 Waiver가 적용된다
- ✅ 모든 TODO 마크가 해결되거나 @TODO TAG로 전환된다

### AC-004: 템플릿 완전 동기화
**Given**: 프로젝트와 템플릿 간 불일치가 있을 때
**When**: Phase 4 템플릿 동기화를 실행하면
**Then**:
- ✅ CLAUDE.md가 루트에서 templates/로 정확히 복사된다 (8191 bytes)
- ✅ .claude 디렉토리가 완전히 동기화된다
- ✅ .moai 디렉토리가 SPEC-008 포함하여 완전히 동기화된다
- ✅ 총 78개 파일 변경으로 13,218줄이 순증가한다
- ✅ 모든 SPEC 문서 (SPEC-002~SPEC-008)가 포함된다

### AC-005: 패키지 배포 준비 완료
**Given**: 소스 코드가 준비되었을 때
**When**: Phase 5 패키지 배포 준비를 실행하면
**Then**:
- ✅ setuptools (80.9.0+)와 wheel (0.45.1+)이 준비된다
- ✅ python -m build가 오류 없이 실행된다
- ✅ moai_adk-0.1.0-py3-none-any.whl이 생성된다 (450KB+)
- ✅ moai_adk-0.1.0.tar.gz가 생성된다 (290KB+)
- ✅ 패키지에 186개 파일이 포함된다
- ✅ 모든 템플릿과 SPEC 문서가 패키지에 포함된다

### AC-006: SPEC-008 문서 완성
**Given**: SPEC-008/spec.md만 있을 때
**When**: Phase 6 문서 완성을 실행하면
**Then**:
- ✅ plan.md가 실제 실행 결과를 반영하여 생성된다
- ✅ acceptance.md가 검증 기준을 포함하여 생성된다
- ✅ SPEC-008이 3-파일 구성 (spec.md, plan.md, acceptance.md)을 갖는다

## @TEST:INTEGRATION-001 통합 테스트

### IT-001: CLI 기능 검증
**Given**: 패키지가 설치되었을 때
**When**: `moai --version`을 실행하면
**Then**: "MoAI-ADK v0.1.17"이 출력된다

### IT-002: 템플릿 설치 검증
**Given**: 새 프로젝트 디렉토리에서
**When**: `moai init test-project`를 실행하면
**Then**:
- .claude/ 디렉토리가 생성된다
- .moai/ 디렉토리가 생성된다
- CLAUDE.md가 생성된다
- 모든 SPEC 문서가 포함된다

### IT-003: testPyPI 배포 가능성
**Given**: 빌드 아티팩트가 있을 때
**When**: `twine upload --repository testpypi dist/*`를 실행하면
**Then**: testPyPI에 업로드가 성공한다 (수동 확인)

## @PERF:PERFORMANCE-001 성능 기준

### 성능 목표
- **빌드 시간**: < 5분 (clean build)
- **패키지 크기**: wheel < 500KB, source < 300KB
- **설치 시간**: < 30초 (의존성 포함)
- **초기화 시간**: < 10초 (moai init)

### 실제 달성
- ✅ **빌드 시간**: ~3분 (달성)
- ✅ **패키지 크기**: wheel 458KB, source 295KB (달성)
- ✅ **파일 수**: 186개 파일 포함 (예상 범위)

## @SEC:SECURITY-001 보안 검증

### 보안 기준
**Given**: 패키지 빌드가 완료되었을 때
**When**: 보안 검증을 실행하면
**Then**:
- ✅ check_secrets.py로 하드코딩된 키가 없음을 확인한다
- ✅ .claude/settings.json이 적절한 권한을 갖는다
- ✅ 모든 민감한 설정이 환경 변수 또는 안전한 저장소를 사용한다

## @QUALITY:GATES-001 품질 게이트

### 필수 통과 항목
- ✅ **버전 일관성**: 모든 버전 파일 0.1.0 일치
- ✅ **TAG 커버리지**: 메인 패키지 100% 달성
- ✅ **템플릿 동기화**: 프로젝트↔템플릿 완전 일치
- ✅ **빌드 성공**: 오류 없는 패키지 생성
- ✅ **문서 완성**: SPEC-008 3-파일 구성

### 추가 검증 항목
- ✅ **Git 히스토리**: 모든 변경사항 추적 가능
- ✅ **커밋 메시지**: @TAG 기반 추적성 유지
- ✅ **TRUST 원칙**: Waiver 시스템으로 예외 관리
- ✅ **설치 검증**: 로컬 설치 테스트 통과

## @ROLLBACK:STRATEGY-001 롤백 전략

### 롤백 조건
- 패키지 빌드 실패 시
- 중요한 기능 회귀 발견 시
- 보안 이슈 발견 시

### 롤백 절차
```bash
# 이전 안정 버전으로 롤백
git reset --hard [previous-stable-commit]

# 버전 되돌리기
git revert [version-commits]

# 패키지 재빌드
python -m build --clean
```

---

**최종 수락 기준**: 모든 AC-001~AC-006 및 통합 테스트 통과
**품질 임계값**: 모든 품질 게이트 항목 100% 달성
**보안 승인**: 모든 보안 검증 통과