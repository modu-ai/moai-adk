# MoAI-ADK v0.4.0 마이그레이션 가이드

> v0.3.x → v0.4.0 업그레이드 전 필수 확인 사항

---

## 🚨 Breaking Changes 요약

### `moai-adk update` 커맨드 목적 변경

| 항목 | v0.3.x | v0.4.0 |
|------|--------|--------|
| **패키지 업그레이드** | 수동 (`uv tool upgrade moai-adk`) | **자동** (`moai-adk update`) |
| **템플릿 업데이트** | `moai-adk update` | `moai-adk init .` |
| **설치 방법 감지** | ❌ 없음 | ✅ 자동 감지 (uv-tool, uv-pip, pip) |
| **버전 확인** | ❌ 없음 | ✅ `moai-adk update --check` |

**핵심 변경**:
```bash
# v0.3.x (이전)
moai-adk update              # 템플릿 파일 업데이트
uv tool upgrade moai-adk     # 패키지 업그레이드 (수동)

# v0.4.0 (새로운 방법)
moai-adk update              # 패키지 업그레이드 (자동)
moai-adk init .              # 템플릿 파일 업데이트
```

---

## 📋 마이그레이션 체크리스트

### ✅ 사전 준비 (필수)

1. **현재 버전 확인**
   ```bash
   moai-adk --version
   # v0.3.x 확인
   ```

2. **작업 중인 변경사항 커밋**
   ```bash
   git status
   git add .
   git commit -m "마이그레이션 전 백업"
   ```

3. **백업 디렉토리 확인**
   ```bash
   ls -la .moai-backups/
   # 기존 백업이 있는지 확인
   ```

### ✅ 업그레이드 실행

#### 1단계: 패키지 업그레이드

```bash
# 자동 감지 및 업그레이드
moai-adk update
```

**예상 출력**:
```
🔍 Checking versions...
   Current version: 0.3.14
   Latest version:  0.4.0

🔎 Detected installation method: uv-tool

📦 Upgrading via uv-tool...
   Command: uv tool upgrade moai-adk

✓ Upgraded to version 0.4.0

✓ Update complete!

💡 For template updates, run: moai-adk init .
```

**문제 발생 시**:
```bash
# 수동 업그레이드 (출력된 명령어 사용)
uv tool upgrade moai-adk     # uv-tool 모드
# 또는
uv pip install --upgrade moai-adk  # uv-pip 모드
# 또는
pip install --upgrade moai-adk     # pip 모드
```

#### 2단계: 템플릿 업데이트

```bash
cd your-project
moai-adk init .
```

**예상 동작**:
- `.moai-backups/{timestamp}/` 백업 자동 생성
- 템플릿 파일 업데이트
- 기존 SPEC/Reports 보존

#### 3단계: 검증

```bash
# 버전 확인
moai-adk --version
# Expected: 0.4.0

# 프로젝트 상태 확인
moai-adk status

# 시스템 진단
moai-adk doctor
```

---

## 🔄 명령어 변경 사항

### 패키지 업그레이드

| v0.3.x | v0.4.0 | 설명 |
|--------|--------|------|
| `uv tool upgrade moai-adk` | `moai-adk update` | 자동 감지 및 업그레이드 |
| `uv pip install --upgrade moai-adk` | `moai-adk update` | 자동 감지 및 업그레이드 |
| (없음) | `moai-adk update --check` | 버전 확인만 |

### 템플릿 업데이트

| v0.3.x | v0.4.0 | 설명 |
|--------|--------|------|
| `moai-adk update` | `moai-adk init .` | 템플릿 파일 업데이트 |
| `moai-adk update --path <dir>` | `cd <dir> && moai-adk init .` | 특정 디렉토리 업데이트 |
| `moai-adk update --force` | `moai-adk init .` | 백업 자동 생성 |

### 제거된 옵션

| 옵션 | v0.3.x | v0.4.0 |
|------|--------|--------|
| `--path <dir>` | ✅ 지원 | ❌ 제거 (`cd <dir>` 사용) |
| `--force` | ✅ 지원 | ❌ 제거 (백업 자동) |

---

## 📚 사용 예시

### 시나리오 1: 일반적인 업그레이드

```bash
# 1. 패키지 업그레이드
moai-adk update

# 2. 템플릿 업데이트
cd ~/my-project
moai-adk init .

# 3. 검증
moai-adk status
```

### 시나리오 2: 여러 프로젝트 업데이트

```bash
# 1. 패키지 한 번만 업그레이드
moai-adk update

# 2. 각 프로젝트마다 템플릿 업데이트
cd ~/project-1 && moai-adk init .
cd ~/project-2 && moai-adk init .
cd ~/project-3 && moai-adk init .
```

### 시나리오 3: 버전 확인만

```bash
# 업그레이드 가능 여부만 확인
moai-adk update --check

# 예상 출력:
# 🔍 Checking versions...
#    Current version: 0.3.14
#    Latest version:  0.4.0
# ⚠ Update available
```

### 시나리오 4: Development 버전 사용

```bash
# Local 버전이 PyPI보다 최신인 경우
moai-adk update --check

# 예상 출력:
# 🔍 Checking versions...
#    Current version: 0.5.0
#    Latest version:  0.4.0
# ✓ Development version (newer than PyPI)
```

---

## 🐛 트러블슈팅

### 문제 1: "Already up to date" 반복

**증상**:
```
moai-adk update
✓ Already up to date
```

**원인**: 이미 최신 버전 설치됨

**해결**:
```bash
# 버전 확인
moai-adk --version

# 강제 재설치 (필요 시)
uv tool install moai-adk --force
```

### 문제 2: 템플릿 업데이트 안됨

**증상**: `moai-adk update`를 실행했는데 템플릿 파일이 변경되지 않음

**원인**: v0.4.0부터 `update`는 패키지만 업그레이드

**해결**:
```bash
# 템플릿 업데이트는 별도 명령
moai-adk init .
```

### 문제 3: PyPI 접속 실패

**증상**:
```
⚠ Unable to fetch from PyPI
⚠ Cannot check for updates
```

**원인**: 네트워크 문제 또는 PyPI 장애

**해결**:
```bash
# 수동 업그레이드
uv tool upgrade moai-adk

# 또는 특정 버전 지정
uv tool install moai-adk==0.4.0
```

### 문제 4: 업그레이드 실패

**증상**:
```
✗ Upgrade failed
⚠ Upgrade failed. Please try manually:
   uv tool upgrade moai-adk
```

**원인**: 권한 문제 또는 패키지 충돌

**해결**:
```bash
# 출력된 명령어 직접 실행
uv tool upgrade moai-adk

# 또는 강제 재설치
uv tool install moai-adk --force
```

---

## ❓ FAQ

### Q1: v0.3.x에서 사용하던 `moai-adk update`가 동작이 다른가요?

**A**: 네, 완전히 변경되었습니다.
- **v0.3.x**: 템플릿 파일 업데이트
- **v0.4.0**: 패키지 자체 업그레이드

### Q2: 템플릿 업데이트는 어떻게 하나요?

**A**: `moai-adk init .` 명령을 사용하세요.
```bash
cd your-project
moai-adk init .
```

### Q3: 자동 감지가 정확하지 않으면?

**A**: 수동 명령어를 사용하세요.
```bash
# 현재 설치 방법 확인
which moai-adk

# uv tool 확인
uv tool list | grep moai-adk

# 수동 업그레이드
uv tool upgrade moai-adk
```

### Q4: 백업은 어떻게 복구하나요?

**A**: `.moai-backups/` 디렉토리에서 수동 복구
```bash
# 백업 목록 확인
ls -la .moai-backups/

# 최신 백업 복구 (예시)
cp -r .moai-backups/2025-10-20-123456/* .
```

### Q5: v0.3.x로 롤백하려면?

**A**: 특정 버전 설치
```bash
# uv tool
uv tool install moai-adk==0.3.14

# uv pip
uv pip install moai-adk==0.3.14

# pip
pip install moai-adk==0.3.14
```

---

## 📞 지원

문제가 계속되면:
1. [GitHub Issues](https://github.com/modu-ai/moai-adk/issues) 등록
2. [GitHub Discussions](https://github.com/modu-ai/moai-adk/discussions) 질문
3. [CHANGELOG.md](CHANGELOG.md) 확인

---

**작성일**: 2025-10-20
**대상 버전**: v0.3.x → v0.4.0
**작성자**: MoAI Team
