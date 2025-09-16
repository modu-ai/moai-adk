# MoAI-ADK CLI 명령어 가이드

> **명령어 용도별 완전 가이드: 사용자 vs 개발자 구분**

## 🎯 명령어 용도 구분

### 👤 사용자용 명령어 (End User Commands)
**설치된 MoAI-ADK 패키지를 사용하는 일반 사용자**

| 명령어 | 용도 | 설명 |
|--------|------|------|
| `moai init` | 프로젝트 초기화 | Claude Code 프로젝트 설정 |
| `moai update` | 패키지 업데이트 | 최신 버전으로 자동 업그레이드 |
| `moai status` | 상태 확인 | 설치 상태 및 구성 점검 |
| `moai doctor` | 문제 진단 | 설치 문제 해결 및 복구 |
| `moai restore` | 백업 복구 | 백업에서 설정 복원 |

### 🛠️ 개발자용 명령어 (Developer Commands)
**MoAI-ADK 패키지를 개발/수정하는 개발자**

| 명령어 | 용도 | 설명 |
|--------|------|------|
| `make build` | 패키지 빌드 | 자동 버전 동기화 포함 빌드 |
| `./scripts/build.sh` | 직접 빌드 | 스크립트 직접 실행 |
| `python build_hooks.py --sync-only` | 수동 동기화 | 버전 동기화만 실행 |

---

## 📋 상세 명령어 레퍼런스

### 1. 사용자용 명령어

#### `moai init` - 프로젝트 초기화
```bash
# 기본 사용법
moai init                        # 현재 디렉토리에 초기화
moai init my-project             # 새 프로젝트 생성
moai init . --interactive        # 대화형 설정

# 옵션
--template, -t                   # 템플릿 선택 (standard, minimal, advanced)
--interactive, -i                # 대화형 설정 마법사
--backup, -b                     # 설치 전 백업 생성
--force, -f                      # 기존 파일 강제 덮어쓰기
--force-copy                     # 심볼릭 링크 대신 파일 복사
```

#### `moai update` - 자동 업데이트 (★ 메인 기능)
```bash
# 완전 자동 업데이트
moai update                      # 패키지 + 글로벌 리소스 업데이트

# 사전 확인
moai update --check              # 업데이트 가능 여부만 확인

# 부분 업데이트
moai update --package-only       # 패키지만 업데이트
moai update --resources-only     # 글로벌 리소스만 업데이트

# 기타 옵션
moai update --no-backup          # 백업 생성 건너뛰기
moai update --verbose            # 상세 정보 표시
```

**자동 수행 단계:**
1. **PyPI 버전 확인**: 최신 버전 체크
2. **패키지 업그레이드**: `pip install --upgrade moai-adk`
3. **글로벌 리소스 동기화**: 템플릿, 에이전트, 훅 업데이트
4. **백업 및 검증**: 안전한 업데이트 보장

#### `moai status` - 상태 점검
```bash
# 기본 상태 확인
moai status                      # 전체 상태 요약

# 상세 정보
moai status --verbose            # 자세한 설치 정보
moai status --project-path /path # 특정 프로젝트 상태
```

#### `moai doctor` - 문제 해결
```bash
# 문제 진단
moai doctor                      # 설치 상태 진단

# 백업 관리
moai doctor --list-backups       # 사용 가능한 백업 목록
```

#### `moai restore` - 백업 복구
```bash
# 백업 복구
moai restore .moai_backup_20250916_035157

# 드라이 런
moai restore /path/to/backup --dry-run
```

### 2. 개발자용 명령어

#### `make build` - 자동화된 패키지 빌드
```bash
# 기본 빌드 (자동 버전 동기화 포함)
make build

# 강제 빌드
make build-force

# 클린 빌드
make build-clean
```

**자동 수행 작업:**
1. 빌드 전 자동 버전 동기화
2. 24개 파일에서 버전 정보 일괄 업데이트
3. 템플릿 변수 자동 적용
4. 패키지 빌드 (`python -m build`)

#### 직접 빌드 스크립트 실행
```bash
# 빌드 스크립트 직접 사용
./scripts/build.sh

# 수동 버전 동기화만
python build_hooks.py --sync-only
python build_hooks.py --pre-build
```

---

## 🚀 실제 사용 시나리오

### 시나리오 1: 일반 사용자 - 첫 설치
```bash
# 1. 패키지 설치
pip install moai-adk

# 2. 프로젝트 초기화
moai init my-project

# 3. Claude Code에서 사용
cd my-project
claude
```

### 시나리오 2: 일반 사용자 - 업데이트
```bash
# 1. 업데이트 확인
moai update --check

# 2. 자동 업데이트
moai update

# 3. 상태 확인
moai status
```

### 시나리오 3: 개발자 - 패키지 빌드 및 배포
```bash
# 1. 코드 수정 후 빌드
make build

# 2. 결과 확인
ls -la dist/

# 3. 배포
git add -A
git commit -m "feat: new feature implementation"
git tag v0.1.17
git push origin main --tags
python -m twine upload dist/*
```

### 시나리오 4: 문제 해결
```bash
# 1. 문제 발생 시 진단
moai doctor

# 2. 백업 확인
moai doctor --list-backups

# 3. 필요시 복구
moai restore .moai_backup_20250916_035157
```

---

## 🔧 고급 사용법

### 개발 환경 설정
```bash
# 개발용 프로젝트 초기화
moai init dev-project --template advanced --interactive

# 개발 중 빈번한 빌드
make build

# 프로덕션 릴리스
make build-clean
./scripts/build.sh
```

### CI/CD 파이프라인 통합
```yaml
# .github/workflows/release.yml
- name: Build with auto version sync
  run: make build

- name: Deploy to PyPI
  run: |
    python -m twine upload dist/*

- name: Create Git tag
  run: |
    VERSION=$(python -c "from src.moai_adk._version import __version__; print(__version__)")
    git tag v$VERSION
    git push origin v$VERSION
```

### 자동화 스크립트
```bash
#!/bin/bash
# release.sh

echo "🚀 Building and releasing MoAI-ADK"

# 빌드 (자동 버전 동기화 포함)
make build-clean

# 배포
python -m twine upload dist/*

# Git 태그 및 푸시
VERSION=$(python -c "from src.moai_adk._version import __version__; print(__version__)")
git add -A
git commit -m "build: release v$VERSION"
git tag v$VERSION
git push origin main --tags
```

---

## 💡 Pro Tips

### 사용자용 팁
1. **정기 업데이트**: 월 1회 `moai update --check` 실행
2. **백업 관리**: 중요한 프로젝트는 `--backup` 옵션 사용
3. **문제 발생 시**: `moai doctor` 먼저 실행
4. **Claude Code 재시작**: 업데이트 후 Claude Code 재시작 권장

### 개발자용 팁
1. **빌드 자동화**: `make build`로 버전 동기화와 빌드를 한 번에
2. **수동 동기화**: 필요시 `python build_hooks.py --sync-only` 실행
3. **Git 통합**: 빌드 후 수동으로 Git 커밋 및 태그 관리
4. **템플릿 시스템**: 새 파일은 자동으로 올바른 버전 적용

---

## 🚨 주의사항

### 명령어 구분 주의
- **`moai update`**: 사용자가 MoAI-ADK를 최신 버전으로 업그레이드
- **`make build`**: 개발자가 패키지 빌드 시 자동 버전 동기화 수행

### 권한 문제
```bash
# Windows에서 관리자 권한 필요시
moai init --force-copy

# macOS/Linux에서 심볼릭 링크 권한 문제
sudo moai init  # 권장하지 않음
moai init --force-copy  # 대안 사용
```

### 백업 관리
- 자동 백업은 `~/.moai-adk-backup/`에 저장
- 정기적으로 오래된 백업 정리 권장
- 중요한 설정은 수동 백업 병행

---

**마지막 업데이트**: 2025-09-16
**MoAI-ADK 버전**: v0.1.17
**CLI 버전**: v2.0.0

**🎯 "올바른 명령어로 효율적인 워크플로우를 경험하세요!"**
