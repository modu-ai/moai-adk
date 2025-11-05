# Windows 환경 테스트 가이드

## 📋 Windows 테스트 방법 비교

| 방법 | 난이도 | 설치 필요 | 장점 | 단점 |
|------|--------|-----------|------|------|
| **WSL 2** | ⭐⭐ | Windows 10/11 기본 | Linux 완벽 호환, Docker 불필요 | 초기 설정 필요 |
| **Native Windows** | ⭐ | Node.js만 | 가장 간단 | 경로 구분자 차이 가능 |
| **Docker Desktop** | ⭐⭐⭐ | Docker Desktop | 완전한 격리 환경 | 리소스 많이 사용 |
| **GitHub Actions** | ⭐ | 없음 | 자동화, 무료 | 로컬 테스트 불가 |

---

## 🚀 방법 1: WSL 2 (권장)

### 장점
- ✅ Linux 환경 완벽 호환
- ✅ Docker 없이 테스트 가능
- ✅ Windows와 파일 시스템 공유
- ✅ 빠른 실행 속도

### 설치 방법

**1단계: WSL 2 설치**
```powershell
# PowerShell (관리자 권한)
wsl --install
```

**2단계: Ubuntu 설치 (기본)**
```powershell
wsl --install -d Ubuntu
```

**3단계: Node.js 설치 (WSL 내부)**
```bash
# WSL 터미널에서
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs
```

### 테스트 실행

**방법 A: Windows에서 WSL 명령 실행**
```powershell
# PowerShell에서
wsl node test-session-notice.js
```

**방법 B: WSL 터미널에서 직접 실행**
```bash
# WSL 터미널에서
cd /mnt/c/Users/[USERNAME]/path/to/MoAI-ADK
node test-session-notice.js
```

**방법 C: Windows 경로 직접 접근**
```powershell
# PowerShell에서
wsl bash -c "cd $(wslpath -u '$(pwd)') && node test-session-notice.js"
```

---

## 🖥️ 방법 2: Native Windows

### PowerShell 테스트 스크립트

**test-session-notice.ps1**:
```powershell
# Windows PowerShell Test Script
$projectRoot = Get-Location

Write-Host "Project Root: $projectRoot" -ForegroundColor Cyan
Write-Host ""

# Test 1: Check .moai directory
$moaiDir = Join-Path $projectRoot ".moai"
$moaiExists = Test-Path $moaiDir
if ($moaiExists) {
    Write-Host "✓ Check .moai directory: ✅ EXISTS" -ForegroundColor Green
} else {
    Write-Host "✓ Check .moai directory: ❌ NOT FOUND" -ForegroundColor Red
}
Write-Host "  Path: $moaiDir"
Write-Host ""

# Test 2: Check .claude/commands/alfred directory
$alfredCommands = Join-Path $projectRoot ".claude\commands\alfred"
$alfredExists = Test-Path $alfredCommands
if ($alfredExists) {
    Write-Host "✓ Check .claude/commands/alfred: ✅ EXISTS" -ForegroundColor Green
} else {
    Write-Host "✓ Check .claude/commands/alfred: ❌ NOT FOUND" -ForegroundColor Red
}
Write-Host "  Path: $alfredCommands"
Write-Host ""

# Test 3: Check old moai path
$moaiCommands = Join-Path $projectRoot ".claude\commands\moai"
$moaiCommandsExists = Test-Path $moaiCommands
if ($moaiCommandsExists) {
    Write-Host "✓ Check .claude/commands/moai (old): ⚠️ EXISTS (not used)" -ForegroundColor Yellow
} else {
    Write-Host "✓ Check .claude/commands/moai (old): ✅ NOT FOUND" -ForegroundColor Green
}
Write-Host "  Path: $moaiCommands"
Write-Host ""

# Test 4: isMoAIProject logic
$isMoAIProject = $moaiExists -and $alfredExists
Write-Host ("=" * 50)
if ($isMoAIProject) {
    Write-Host "isMoAIProject() result: ✅ TRUE (initialized)" -ForegroundColor Green
} else {
    Write-Host "isMoAIProject() result: ❌ FALSE (not initialized)" -ForegroundColor Red
}
Write-Host ("=" * 50)
Write-Host ""

if ($isMoAIProject) {
    Write-Host "✅ PASS: Project should NOT show initialization message" -ForegroundColor Green
} else {
    Write-Host "❌ FAIL: Project WILL show initialization message" -ForegroundColor Red
    if (-not $moaiExists) { Write-Host "  Reason: .moai directory missing" }
    if (-not $alfredExists) { Write-Host "  Reason: .claude/commands/alfred directory missing" }
}
```

**실행 방법**:
```powershell
# PowerShell에서
.\test-session-notice.ps1
```

---

## 🐳 방법 3: Docker Desktop (Windows)

### 전제 조건
- Docker Desktop for Windows 설치
- WSL 2 백엔드 활성화 (권장)

### 실행 방법
```powershell
# PowerShell에서
docker build -f Dockerfile.test -t moai-session-test .
docker run --rm moai-session-test
```

---

## 🤖 방법 4: GitHub Actions (CI/CD)

### 장점
- ✅ 별도 설치 불필요
- ✅ 자동화된 테스트
- ✅ 모든 플랫폼 동시 테스트

### GitHub Actions 워크플로우

**.github/workflows/test-cross-platform.yml**:
```yaml
name: Cross-Platform Test

on:
  push:
    branches: [ main, develop, feature/** ]
  pull_request:
    branches: [ main, develop ]

jobs:
  test:
    name: Test on ${{ matrix.os }}
    runs-on: ${{ matrix.os }}
    
    strategy:
      matrix:
        os: [ubuntu-latest, windows-latest, macos-latest]
        node-version: [18.x, 20.x]
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Node.js ${{ matrix.node-version }}
        uses: actions/setup-node@v3
        with:
          node-version: ${{ matrix.node-version }}
      
      - name: Run session-notice test
        run: node test-session-notice.js
        
      - name: Run session-notice hook
        run: node .claude/hooks/alfred/session-notice.cjs
```

---

## 📊 경로 처리 차이점

### Unix (macOS/Linux)
```javascript
path.join(projectRoot, '.claude', 'commands', 'alfred')
// 결과: /path/to/project/.claude/commands/alfred
```

### Windows
```javascript
path.join(projectRoot, '.claude', 'commands', 'alfred')
// 결과: C:\path\to\project\.claude\commands\alfred
```

**Node.js `path.join()`이 자동으로 플랫폼별 구분자 처리** ✅

---

## ✅ 권장 순서

### 1순위: WSL 2
- 가장 빠르고 정확한 Linux 호환 테스트
- Docker 불필요
- Windows와 파일 공유 용이

### 2순위: Native Windows (PowerShell)
- 가장 간단한 방법
- Node.js만 설치하면 즉시 가능
- Windows 고유 환경 테스트

### 3순위: GitHub Actions
- 자동화 및 모든 플랫폼 동시 테스트
- 로컬 환경 설정 불필요

### 4순위: Docker Desktop
- 완전한 격리 환경
- 리소스 많이 사용
- WSL 2 백엔드 필요 (결국 WSL 설치 필요)

---

## 🔧 트러블슈팅

### WSL 2 설치 오류
```powershell
# Windows 버전 확인 (Windows 10 버전 2004 이상 필요)
winver

# WSL 기능 활성화
dism.exe /online /enable-feature /featurename:Microsoft-Windows-Subsystem-Linux /all /norestart
dism.exe /online /enable-feature /featurename:VirtualMachinePlatform /all /norestart

# WSL 2를 기본값으로 설정
wsl --set-default-version 2
```

### PowerShell 실행 정책 오류
```powershell
# 실행 정책 변경 (관리자 권한)
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### 경로 인식 오류
```powershell
# 절대 경로 사용
$scriptPath = Split-Path -Parent $MyInvocation.MyCommand.Path
cd $scriptPath
```

---

## 📌 요약

**최고의 선택**: **WSL 2** 
- Linux 완벽 호환
- Docker 불필요
- 빠른 실행 속도
- Windows 파일 접근 용이

**가장 간단한 방법**: **Native PowerShell**
- Node.js만 설치
- 즉시 테스트 가능

**자동화 최고**: **GitHub Actions**
- 별도 설정 불필요
- 모든 플랫폼 동시 테스트
