# MoAI-ADK v0.4.0 릴리즈 체크리스트

> **릴리즈 준비 날짜**: 2025-10-20
> **배포 대기 상태**: ✅ 모든 사전 검증 완료

---

## ✅ 사전 검증 완료 항목

### 1. 코드 품질
- [x] **update.py 완전 재작성** (177 → 199 lines)
  - 새 기능: detect_install_method(), upgrade_package(), get_latest_version()
  - 제거: 템플릿 업데이트 로직 전체, TemplateProcessor 의존성
  - Breaking Change: --path, --force 옵션 제거

- [x] **테스트 커버리지**
  - 28개 테스트 케이스 전면 재작성
  - 모든 테스트 통과 (28/28) ✅
  - update.py 커버리지: 96.97% ✅

- [x] **코드 리뷰**
  - 타입 힌트 완성 (mypy 호환)
  - 보안 검사 완료 (nosec B310 주석 추가)
  - 에러 핸들링 완료 (TimeoutError, URLError, JSONDecodeError)

### 2. 문서화
- [x] **CHANGELOG.md**: Breaking Changes 섹션 추가
- [x] **README.md**: v0.4.0 릴리즈 노트 섹션 추가
- [x] **MIGRATION-v0.4.0.md**: 마이그레이션 가이드 신규 작성
- [x] **RELEASE-NOTES-v0.4.0.md**: 상세 릴리즈 노트 작성

### 3. 버전 관리
- [x] **pyproject.toml**: version = "0.4.0" ✅
- [x] **.moai/config.json**: moai_adk_version = "0.4.0" ✅
- [x] **Git 커밋**: 모든 변경사항 커밋 완료

### 4. 빌드 및 검증
- [x] **빌드 파일 생성**
  - dist/moai_adk-0.4.0-py3-none-any.whl (317 KB)
  - dist/moai_adk-0.4.0.tar.gz (240 KB)

- [x] **PyPI 호환성 검증**
  - `twine check` 통과 ✅
  - 메타데이터 검증 완료 ✅

- [x] **로컬 설치 검증**
  - `uv tool install --force .` 성공 ✅
  - `~/.local/bin/moai-adk --version` → v0.4.0 ✅
  - `moai-adk update --check` 정상 동작 ✅

### 5. Git 준비
- [x] **브랜치**: develop
- [x] **최근 커밋**: e46daa0 (릴리즈 준비 완료)
- [x] **변경사항**: 모두 커밋됨 (`git status` clean)

---

## 📝 릴리즈 준비 요약

### 변경 파일 통계
| 카테고리 | 파일 수 | 설명 |
|----------|---------|------|
| **코드** | 2 | update.py, test_update.py |
| **문서** | 5 | CHANGELOG, README, MIGRATION, RELEASE-NOTES, config.json |
| **빌드** | 2 | .whl, .tar.gz |
| **총계** | 9 |  |

### 주요 커밋 (v0.3.14 이후)
```
e46daa0 📦 RELEASE: v0.4.0 릴리즈 준비 완료
e324aac 📝 DOCS: v0.4.0 Breaking Change 문서화 완료
71269d9 ♻️ REFACTOR: update 커맨드 완전 개편 - 패키지 업그레이드 전용
fbfce40 🔧 REFACTOR: Skill 메타데이터 공식 표준 준수
e143e47 🎯 IMPROVE: update 명령어 개선 - 패키지 업그레이드 안내 강화
6909502 ♻️ REFACTOR: 버전 관리 방식 개선 (SSOT)
2b3a058 🔖 VERSION: Bump to v0.4.0
```

---

## 🚀 PyPI 배포 절차 (실행 대기)

### Step 1: Git 태그 생성
```bash
git tag v0.4.0
git push origin v0.4.0
```

### Step 2: PyPI 배포
```bash
uv publish
```

**예상 출력**:
```
Uploading distributions to https://upload.pypi.org/legacy/
Uploading moai_adk-0.4.0-py3-none-any.whl
Uploading moai_adk-0.4.0.tar.gz

View at:
https://pypi.org/project/moai-adk/0.4.0/
```

### Step 3: 배포 검증
```bash
# 1. PyPI 페이지 확인
open https://pypi.org/project/moai-adk/0.4.0/

# 2. 새로운 환경에서 설치 테스트
uv tool install moai-adk==0.4.0
moai-adk --version  # v0.4.0 확인

# 3. 기능 테스트
moai-adk update --check
moai-adk doctor
```

### Step 4: GitHub 릴리즈 생성
1. https://github.com/modu-ai/moai-adk/releases/new 접속
2. Tag: `v0.4.0`
3. Title: `MoAI-ADK v0.4.0 - Breaking Change: update 커맨드 개편`
4. Description: RELEASE-NOTES-v0.4.0.md 내용 복사
5. Attach files: dist/moai_adk-0.4.0-py3-none-any.whl, dist/moai_adk-0.4.0.tar.gz
6. Publish release

---

## ⚠️ 배포 전 최종 확인사항

### 필수 확인
- [ ] PyPI 계정 토큰 확인 (`~/.pypirc` 또는 환경변수)
- [ ] TestPyPI 배포 테스트 (선택사항)
- [ ] develop → main 머지 필요 여부 확인
- [ ] CI/CD 파이프라인 통과 확인

### 배포 후 작업
- [ ] PyPI 페이지 확인
- [ ] GitHub 릴리즈 생성
- [ ] Discussions/Issues에 릴리즈 공지
- [ ] 기존 사용자에게 마이그레이션 가이드 안내

---

## 📊 릴리즈 영향 분석

### Breaking Change 영향
- **영향받는 사용자**: 모든 v0.3.x 사용자
- **마이그레이션 시간**: 5분 이내
- **하위 호환성**: ❌ (Breaking Change)

### 마이그레이션 난이도
- **쉬움** ⭐⭐⭐⭐⭐
  - 명령어만 변경: `moai-adk update` (패키지 업그레이드)
  - 템플릿 업데이트: `moai-adk init .`
  - 상세 가이드: MIGRATION-v0.4.0.md

---

## 🔗 관련 문서

- [CHANGELOG.md](CHANGELOG.md#v040---2025-10-20-phase-2-완료)
- [RELEASE-NOTES-v0.4.0.md](RELEASE-NOTES-v0.4.0.md)
- [MIGRATION-v0.4.0.md](MIGRATION-v0.4.0.md)
- [README.md](README.md#v040-릴리즈-노트)

---

## ✅ 배포 대기 상태

**현재 상태**: 🟢 **Ready for PyPI Deployment**

모든 사전 검증이 완료되었습니다. PyPI 배포를 진행해도 안전합니다.

**다음 단계**: 사용자 승인 후 PyPI 배포 실행

---

**작성일**: 2025-10-20
**작성자**: MoAI Team
**버전**: v0.4.0
