# macOS에서 Windows 환경 테스트하는 방법

## 🚨 중요한 사실

**Docker는 macOS에서 Windows 컨테이너를 실행할 수 없습니다!**
- Docker Desktop for Mac은 **Linux 컨테이너만** 지원
- Windows 컨테이너는 **Windows 호스트에서만** 실행 가능

---

## ✅ macOS에서 Windows 테스트하는 실제 방법

### 방법 1: GitHub Actions (권장 ⭐⭐⭐⭐⭐)

**장점**:
- ✅ 완전 무료
- ✅ 설치 불필요
- ✅ 실제 Windows 환경
- ✅ 모든 플랫폼 동시 테스트

**단점**:
- ❌ 로컬 실행 불가 (Git push 필요)

**설정 방법**:

1. `.github/workflows/test-cross-platform.yml` 생성:
```yaml
name: Cross-Platform Test

on:
  push:
    branches: [ '**' ]  # 모든 브랜치
  pull_request:

jobs:
  test:
    name: Test on ${{ matrix.os }}
    runs-on: ${{ matrix.os }}
    
    strategy:
      matrix:
        os: [ubuntu-latest, windows-latest, macos-latest]
    
    steps:
      - uses: actions/checkout@v3
      
      - uses: actions/setup-node@v3
        with:
          node-version: 18
      
      - name: Run test
        run: node test-session-notice.js
```

2. **테스트 실행**:
```bash
git add .github/workflows/test-cross-platform.yml
git commit -m "ci: Add cross-platform tests"
git push
```

3. **결과 확인**: GitHub → Actions 탭

---

### 방법 2: VirtualBox + Windows (로컬 테스트)

**장점**:
- ✅ 완전 무료
- ✅ 로컬 실행 가능
- ✅ 실제 Windows 환경

**단점**:
- ❌ 초기 설정 복잡
- ❌ 디스크 공간 필요 (20GB+)

**설정 방법**:

**1단계: VirtualBox 설치**
```bash
brew install --cask virtualbox
```

**2단계: Windows 10/11 ISO 다운로드**
- [Windows 11 다운로드](https://www.microsoft.com/software-download/windows11)
- 무료 평가판 (90일)

**3단계: VM 생성**
1. VirtualBox 실행
2. "New" → Windows 10/11 선택
3. 메모리: 4GB 이상
4. 디스크: 50GB 이상

**4단계: Windows 설정**
1. Node.js 설치
2. Git 설치 (옵션)
3. 프로젝트 폴더 공유 설정

**5단계: 테스트 실행**
```powershell
# Windows VM에서
cd C:\path\to\MoAI-ADK
node test-session-notice.js
.\test-session-notice.ps1
```

---

### 방법 3: Parallels Desktop (유료, 가장 편함)

**장점**:
- ✅ macOS 통합 최고
- ✅ 파일 공유 쉬움
- ✅ 빠른 성능

**단점**:
- ❌ 유료 ($99/년)

**설정**:
1. Parallels Desktop 설치
2. Windows 11 ARM 자동 설치
3. Coherence 모드로 macOS와 통합

**테스트**:
```bash
# macOS 터미널에서 바로 실행
prlctl exec "Windows 11" node test-session-notice.js
```

---

### 방법 4: 클라우드 VM (임시 테스트)

**Azure/AWS 무료 티어**:

**AWS EC2 (12개월 무료)**:
```bash
# 1. AWS EC2 Windows 인스턴스 생성
# 2. RDP로 연결
# 3. 테스트 실행
```

**장점**:
- ✅ 실제 Windows 서버
- ✅ 무료 티어 사용 가능

**단점**:
- ❌ 네트워크 필요
- ❌ 설정 복잡

---

## 📊 방법 비교표

| 방법 | 비용 | 설치 시간 | 로컬 실행 | 추천도 |
|------|------|-----------|-----------|--------|
| **GitHub Actions** | 무료 | 5분 | ❌ | ⭐⭐⭐⭐⭐ |
| **VirtualBox** | 무료 | 1시간 | ✅ | ⭐⭐⭐⭐ |
| **Parallels** | $99/년 | 30분 | ✅ | ⭐⭐⭐⭐ |
| **Cloud VM** | 무료~유료 | 30분 | ❌ | ⭐⭐⭐ |

---

## 🎯 권장 워크플로우

### 개발 중 (빠른 피드백)
1. **macOS 로컬**: `node test-session-notice.js`
2. **Linux Docker**: `docker run --rm moai-session-test`

### PR 전 (최종 검증)
1. **GitHub Actions**: 모든 플랫폼 자동 테스트
2. **결과 확인**: Actions 탭에서 확인

### 디버깅 필요 시
1. **VirtualBox**: Windows VM에서 직접 디버깅

---

## 🚀 즉시 실행 가능한 방법

### GitHub Actions로 Windows 테스트 (지금 바로!)

```bash
# 1. Workflow 파일 생성
mkdir -p .github/workflows
cat > .github/workflows/test-cross-platform.yml << 'YAML'
name: Cross-Platform Test
on:
  push:
    branches: [ '**' ]
jobs:
  test:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, windows-latest, macos-latest]
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: 18
      - run: node test-session-notice.js
YAML

# 2. 커밋 및 푸시
git add .github/workflows/test-cross-platform.yml
git commit -m "ci: Add cross-platform tests"
git push

# 3. GitHub에서 결과 확인
# https://github.com/[USER]/[REPO]/actions
```

---

## 💡 핵심 요약

**macOS에서 로컬로 Windows 테스트는 불가능합니다 (Docker 제한)**

**현실적인 해결책**:
1. **GitHub Actions** (무료, 권장) ← 지금 바로 사용 가능!
2. **VirtualBox** (무료, 로컬)
3. **Parallels** (유료, 편함)

**가장 빠른 방법**: 
→ GitHub Actions 설정 (5분) → Push → Actions 탭 확인
