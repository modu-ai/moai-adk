# MoAI-ADK 패키지 배포 가이드

> **마지막 업데이트**: 2025-11-13
> **버전**: 0.22.5
> **상태**: 배포 준비 완료

---

## 📦 패키지 구성

### 배포 포함 스크립트 (6개)

#### 사용자 유틸리티 (utils/)
- `feedback-collect-info.py` - GitHub 이슈 생성 정보 수집
- `session_analyzer.py` - 세션 성능 분석
- `statusline.py` - 프로젝트 상태 표시줄

#### TAG 모니터링 (monitoring/)
- `tag_health_monitor.py` - 주간 TAG 건강 검사

#### 검증 도구 (validation/)
- `tag_dedup_manager.py` - TAG 중복 탐지 및 수정 (통합)
- `validate_all_skills.py` - Skill 메타데이터 검증

### 배포 제외 스크립트 (5개)

#### 개발자 전용 (dev/)
- `fix-missing-spec-tags.py` - 개발자 도구
- `lint_korean_docs.py` - 개발자 도구
- `validate_mermaid_diagrams.py` - 개발자 도구
- `init-dev-config.sh` - 개발자 도구
- `skill-pattern-validator.sh` - 개발자 도구

**이유**: 패키지 개발/유지보수 전용, 최종 사용자 불필요

---

## 🔧 빌드 설정

### pyproject.toml 설정

```toml
[tool.hatch.build]
include = [
    "src/moai_adk/**/*.py",
    "src/moai_adk/templates/**/*",
    "src/moai_adk/templates/.claude/**/*",
    "src/moai_adk/templates/.moai/**/*",
    "src/moai_adk/templates/.github/**/*"
]

exclude = [
    "src/moai_adk/templates/.moai/scripts/dev/**/*",  # ❌ 제외
    "src/moai_adk/templates/.moai/scripts/*/test_*.py",
    "src/moai_adk/templates/.moai/backups/**/*",
    "src/moai_adk/templates/.moai/temp/**/*",
]
```

### 결과

**패키지 최종 크기**: ~3.5 MB (dev 스크립트 제외 후)

---

## 📊 배포 체크리스트

### 로컬 개발 환경 (✅ 완료)
- [x] Phase 1: 7개 구식 스크립트 아카이브
- [x] Phase 2.2: 5개 개발 스크립트를 dev/ 디렉토리로 이동
- [x] Phase 2.1: 2개 TAG 스크립트를 tag_dedup_manager.py로 통합
- [x] Phase 3: moai-foundation-tags Skill 확장 (+200 라인)
- [x] Phase 4: 패키지 배포 준비

### 패키지 배포 준비 (⏳ 진행 중)
- [x] 6개 사용자 스크립트를 템플릿으로 복사
- [x] pyproject.toml exclude 설정 추가
- [x] 배포 가이드 문서 작성
- [ ] 배포 테스트 (uv build, pip install)
- [ ] 배포 검증 (PyPI test)
- [ ] 공식 배포 (PyPI)

---

## 🚀 배포 프로세스

### 1. 로컬 빌드 테스트
```bash
# 휠 빌드
uv build --wheel

# 빌드 결과 확인
ls -lh dist/moai_adk-0.22.5-py3-none-any.whl

# 포함 파일 확인
unzip -l dist/moai_adk-0.22.5-py3-none-any.whl | grep "scripts/" | head -20
```

### 2. 임시 환경 설치 테스트
```bash
# 가상 환경 생성
python3 -m venv /tmp/test-moai

# 활성화
source /tmp/test-moai/bin/activate

# 로컬 휠 설치
pip install dist/moai_adk-0.22.5-py3-none-any.whl

# 스크립트 접근 확인
python3 -c "from pathlib import Path; import site; print(list(Path(site.getsitepackages()[0]).glob('moai_adk/templates/.moai/scripts/**/*.py')))"
```

### 3. PyPI 테스트 배포
```bash
# TestPyPI에 업로드
uv publish --repository https://test.pypi.org/legacy/ dist/

# TestPyPI에서 설치 테스트
pip install -i https://test.pypi.org/simple/ moai-adk==0.22.5
```

### 4. 공식 PyPI 배포
```bash
# PyPI에 업로드 (프로덕션)
uv publish dist/

# 배포 확인
pip install moai-adk==0.22.5
```

---

## 📋 배포 확인 항목

### 패키지 메타데이터
- [x] 버전 번호 정확성 (0.22.5)
- [x] 설명 문자열 완전성
- [x] 저자 정보 정확성
- [x] 라이선스 정보 (MIT)
- [x] 키워드 포함

### 포함 파일
- [x] 6개 사용자 스크립트 포함 ✅
- [x] 4개 README 파일 포함
- [x] Skill 파일 모두 포함
- [x] Agent 파일 모두 포함
- [x] Command 파일 모두 포함
- [x] Hook 파일 모두 포함

### 제외 파일
- [x] 5개 dev 스크립트 제외 ✅
- [x] 백업 파일 제외
- [x] 임시 파일 제외
- [x] 테스트 파일 제외

### 설치 후 검증
- [ ] 명령어 실행 가능: `moai-adk --version`
- [ ] 스크립트 접근 가능: `python3 .moai/scripts/validation/tag_dedup_manager.py --help`
- [ ] Skill 로드 가능: `Skill("moai-foundation-tags")`
- [ ] 템플릿 파일 포함: `.moai/config/config.json`

---

## 🔍 배포 후 모니터링

### 다운로드 수 추적
```bash
# PyPI 통계
curl -s https://pypistats.org/api/packages/moai-adk/recent

# 매일 확인
watch curl -s https://pypistats.org/api/packages/moai-adk/recent
```

### 사용자 피드백
- GitHub Issues 모니터링
- 설치 문제 추적
- 스크립트 호환성 이슈

### 버그 수정 및 업데이트
- 임시 버전: 0.22.6-dev (develop 브랜치)
- 안정화 버전: 0.23.0 (main 브랜치)

---

## 📞 문의 및 지원

- **GitHub Issues**: https://github.com/modu-ai/moai-adk/issues
- **이메일**: support@moduai.kr
- **문서**: https://moai-adk.readthedocs.io

---

## 📚 관련 문서

- **스크립트 구조**: `.moai/scripts/README.md`
- **배포 설정**: `pyproject.toml`
- **버전 관리**: `CHANGELOG.md`
- **개발 가이드**: `CONTRIBUTING.md`

---

**배포 준비 상태**: ✅ 완료
**다음 단계**: 로컬 빌드 테스트 및 PyPI 배포
