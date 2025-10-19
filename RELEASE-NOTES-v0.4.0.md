# MoAI-ADK v0.4.0 Release Notes

> **릴리즈 날짜**: 2025-10-20
> **이전 버전**: v0.3.14 → **현재 버전**: v0.4.0

---

## 🎉 주요 하이라이트

### 🚨 Breaking Change: `moai-adk update` 커맨드 완전 개편

v0.4.0의 가장 큰 변화는 **`moai-adk update` 커맨드의 목적이 완전히 변경**되었다는 점입니다.

**변경 전 (v0.3.x)**:
```bash
moai-adk update  # 템플릿 파일 업데이트
```

**변경 후 (v0.4.0)**:
```bash
moai-adk update        # 패키지 자체 업그레이드 (자동 감지)
moai-adk update --check  # 버전 확인만
moai-adk init .        # 템플릿 파일 업데이트
```

**왜 변경했나요?**
1. **명확한 책임 분리**: 패키지 업그레이드 vs 템플릿 업데이트
2. **사용자 경험 개선**: 업그레이드 실패 원인이 명확해짐
3. **자동화**: 설치 방법(uv-tool, uv-pip, pip)을 자동 감지하여 적절한 명령어 실행
4. **안전성**: 템플릿 업데이트는 `init .`로 명시적으로 실행

---

## ✨ 새로운 기능

### 1. 자동 설치 방법 감지 및 업그레이드

**설치 방법 자동 감지**:
```bash
moai-adk update
```

실행 시 자동으로 다음 중 하나를 감지하고 실행:
- ✅ `uv tool install moai-adk` → `uv tool upgrade moai-adk`
- ✅ `uv pip install moai-adk` → `uv pip install --upgrade moai-adk`
- ✅ `pip install moai-adk` → `pip install --upgrade moai-adk`

**버전 확인**:
```bash
moai-adk update --check

# 출력 예시:
# 🔍 Checking versions...
#    Current version: 0.3.14
#    Latest version:  0.4.0
# ⚠ Update available
```

**Development 버전 감지**:
```bash
# 로컬 버전이 PyPI보다 최신인 경우
moai-adk update --check

# 출력:
# ✓ Development version (newer than PyPI)
```

### 2. Skills 메타데이터 표준화 완료

**Anthropic 공식 표준 100% 준수**:
- ✅ 54개 Skills 메타데이터 표준화
- ✅ 비표준 필드 제거 (tier, auto-load, version, author, license, tags, metadata)
- ✅ 표준 A (최소 구현) 방식 적용
- ✅ 필수 필드: `name`, `description`
- ✅ 선택 필드: `allowed-tools`

**영향받는 파일**: 108개 (54 Skills × 2 파일)

---

## 🔄 변경 사항

### Breaking Changes

#### `moai-adk update` 커맨드

| 항목 | v0.3.x | v0.4.0 |
|------|--------|--------|
| **패키지 업그레이드** | 수동 (`uv tool upgrade moai-adk`) | **자동** (`moai-adk update`) |
| **템플릿 업데이트** | `moai-adk update` | `moai-adk init .` |
| **설치 방법 감지** | ❌ 없음 | ✅ 자동 감지 |
| **버전 확인** | ❌ 없음 | ✅ `--check` 옵션 |

#### 제거된 옵션

- ❌ `--path <dir>` → `cd <dir> && moai-adk init .` 사용
- ❌ `--force` → 백업 자동 생성으로 대체

### 개선 사항

#### 1. update 명령어 완전 재작성
- **새 기능**:
  - `detect_install_method()`: 설치 방법 자동 감지
  - `upgrade_package()`: 자동 업그레이드 실행
  - `get_latest_version()`: PyPI 버전 확인

- **제거된 기능**:
  - 템플릿 파일 업데이트 로직 전체 제거
  - TemplateProcessor 의존성 제거

- **테스트**:
  - 28개 테스트 케이스 전면 재작성
  - 모든 테스트 통과 (28/28)
  - update.py 커버리지 96.97%

#### 2. 버전 관리 개선
- SSOT (Single Source of Truth) 방식 도입
- `pyproject.toml`의 version 필드가 유일한 진실의 원천
- `__init__.py`에서 동적으로 버전 로드

#### 3. 버그 수정
- ✅ 버전 동일 시 update 명령어가 불필요하게 진행되는 문제 수정
- ✅ TemplateBackup AttributeError 수정
- ✅ Domain Tier 템플릿 동기화 문제 해결

---

## 📚 문서화

### 신규 문서

1. **MIGRATION-v0.4.0.md**: 마이그레이션 가이드
   - v0.3.x → v0.4.0 전환 체크리스트
   - 명령어 변경 사항 매핑 테이블
   - 트러블슈팅 및 FAQ (4가지 문제, 5가지 질문)

2. **CHANGELOG.md**: Breaking Changes 섹션 추가
   - 변경 전/후 비교
   - 마이그레이션 가이드 요약
   - 자동 감지 기능 설명

3. **README.md**: 업그레이드 가이드 전면 개편
   - Breaking Change 경고 추가
   - 자동 감지 기능 상세 설명
   - 검증 체크리스트 추가

---

## 📊 통계

| 항목 | 수치 |
|------|------|
| **총 커밋** | 12개 (v0.3.14 이후) |
| **변경된 파일** | 120개+ |
| **새 테스트** | 28개 (update.py) |
| **테스트 통과율** | 100% |
| **update.py 커버리지** | 96.97% |
| **Skills 표준화** | 54개 (108개 파일) |

---

## 🚀 업그레이드 방법

### 1단계: 패키지 업그레이드

```bash
# 자동 감지 및 업그레이드 (권장)
moai-adk update

# 버전 확인만
moai-adk update --check
```

### 2단계: 템플릿 업데이트

```bash
cd your-project
moai-adk init .
```

### 3단계: 검증

```bash
# 패키지 버전 확인
moai-adk --version  # v0.4.0

# 프로젝트 상태 확인
moai-adk status

# 시스템 진단
moai-adk doctor
```

---

## 📝 마이그레이션 가이드

상세한 마이그레이션 가이드는 [MIGRATION-v0.4.0.md](MIGRATION-v0.4.0.md)를 참조하세요.

### 주요 명령어 변경

| v0.3.x | v0.4.0 |
|--------|--------|
| `moai-adk update` | `moai-adk update` (패키지 업그레이드)<br>`moai-adk init .` (템플릿 업데이트) |
| `uv tool upgrade moai-adk` | `moai-adk update` (자동 감지) |
| `moai-adk update --path <dir>` | `cd <dir> && moai-adk init .` |
| `moai-adk update --force` | `moai-adk init .` (백업 자동) |

---

## 🐛 알려진 문제

### pyenv shim 우선순위

**증상**:
```bash
which moai-adk  # /Users/xxx/.pyenv/shims/moai-adk (v0.3.x)
```

**해결 방법**:
```bash
# 직접 경로 사용
~/.local/bin/moai-adk --version  # v0.4.0

# 또는 Alias 설정
echo "alias moai-adk='~/.local/bin/moai-adk'" >> ~/.bashrc
source ~/.bashrc
```

---

## 🙏 감사의 말

이번 릴리즈는 사용자 피드백을 반영하여 만들어졌습니다.

**특별히 감사드립니다**:
- GitHub Discussion #30, #39에서 버그를 보고해주신 분들
- update 명령어 개선을 제안해주신 분들

**기여자**:
- @Goos - 전체 개발 및 릴리즈 관리
- Claude Code - 개발 지원 및 테스트

---

## 📞 지원

문제가 발생하면:
1. [MIGRATION-v0.4.0.md](MIGRATION-v0.4.0.md) 확인
2. [GitHub Issues](https://github.com/modu-ai/moai-adk/issues) 등록
3. [GitHub Discussions](https://github.com/modu-ai/moai-adk/discussions) 질문

---

## 🔗 관련 링크

- [CHANGELOG.md](CHANGELOG.md)
- [MIGRATION-v0.4.0.md](MIGRATION-v0.4.0.md)
- [README.md](README.md)
- [PyPI](https://pypi.org/project/moai-adk/)
- [GitHub Repository](https://github.com/modu-ai/moai-adk)

---

**릴리즈 날짜**: 2025-10-20
**작성자**: MoAI Team
**버전**: v0.4.0
