# MoAI-ADK WSL 공식 지원 전략

## 🎯 핵심 전략

**Windows 사용자에게 WSL을 공식 권장**
- Native Windows는 "지원하지만 권장하지 않음"
- WSL 2는 "공식 권장 환경"
- 이유: Linux 호환성, Docker 지원, 개발 도구 호환성

---

## 📋 WSL 지원의 이점

### 1. 사용자 관점
- ✅ Linux와 동일한 개발 경험
- ✅ Docker Desktop WSL 2 백엔드 사용
- ✅ VS Code WSL 원격 개발
- ✅ 경로 문제 완전 해결 (Unix 스타일)
- ✅ 성능 향상 (네이티브 Linux)

### 2. 유지보수 관점
- ✅ 크로스 플랫폼 코드 단순화
- ✅ Windows 전용 버그 감소
- ✅ 테스트 복잡도 감소
- ✅ 문서 통일 (Linux 기준)

### 3. 업계 표준
- Docker Desktop: WSL 2 권장
- VS Code: WSL 원격 개발 지원
- Node.js: WSL에서 더 나은 성능
- Git: WSL에서 더 빠른 성능

---

## 🚀 구현 계획

### Phase 1: 환경 감지 및 경고

**목표**: WSL 사용 권장 메시지 표시

**구현**:
```typescript
// src/utils/platform-detector.ts
export function detectEnvironment() {
  const isWindows = process.platform === 'win32';
  const isWSL = process.env.WSL_DISTRO_NAME !== undefined;
  
  return {
    platform: process.platform,
    isWindows,
    isWSL,
    isNativeWindows: isWindows && !isWSL,
    recommendation: isWindows && !isWSL 
      ? 'WSL_RECOMMENDED' 
      : 'OK'
  };
}

// src/cli/commands/init.ts
const env = detectEnvironment();
if (env.isNativeWindows) {
  console.warn(`
⚠️  Windows 환경 감지됨

권장사항: WSL 2 사용을 강력히 권장합니다.
- 더 나은 성능
- Linux 호환성
- Docker 지원

설치 방법: https://moai-adk.dev/docs/install-wsl
  `);
}
```

### Phase 2: WSL 설치 가이드

**파일**: `docs/install/windows-wsl.md`

**내용**:
1. WSL 2 설치 (PowerShell 한 줄)
2. Ubuntu 설치
3. Node.js 설치
4. MoAI-ADK 설치
5. VS Code 연동

### Phase 3: README 업데이트

**변경 사항**:
```markdown
## Installation

### macOS / Linux
```bash
npm install -g moai-adk
```

### Windows
**권장: WSL 2 사용**
```powershell
# PowerShell (관리자 권한)
wsl --install

# WSL 터미널에서
npm install -g moai-adk
```

**Native Windows** (권장하지 않음)
- 경로 문제 발생 가능
- Docker 지원 제한적
```

### Phase 4: CI/CD 업데이트

**GitHub Actions**:
```yaml
jobs:
  test:
    strategy:
      matrix:
        os: 
          - ubuntu-latest
          - macos-latest
          # Native Windows 제외
          # - windows-latest
        
  test-wsl:
    name: Test on Windows (WSL)
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup WSL
        uses: Vampire/setup-wsl@v2
        with:
          distribution: Ubuntu-22.04
      
      - name: Install Node.js in WSL
        run: |
          wsl sudo apt-get update
          wsl sudo apt-get install -y nodejs npm
      
      - name: Run tests in WSL
        run: wsl node test-session-notice.js
```

---

## 📊 지원 매트릭스

| 플랫폼 | 지원 수준 | 권장도 | 비고 |
|--------|-----------|--------|------|
| macOS | ✅ Full | ⭐⭐⭐⭐⭐ | 공식 지원 |
| Linux | ✅ Full | ⭐⭐⭐⭐⭐ | 공식 지원 |
| **WSL 2** | ✅ Full | ⭐⭐⭐⭐⭐ | **Windows 권장** |
| Native Windows | ⚠️ Limited | ⭐⭐ | 지원하지만 권장 안 함 |

---

## 🔧 WSL 감지 코드

### 환경 변수 확인
```typescript
function isWSL(): boolean {
  // WSL 1
  if (process.env.WSL_DISTRO_NAME) return true;
  
  // WSL 2
  if (process.platform === 'linux') {
    const release = fs.readFileSync('/proc/version', 'utf8').toLowerCase();
    if (release.includes('microsoft') || release.includes('wsl')) {
      return true;
    }
  }
  
  return false;
}
```

### 플랫폼 정보 출력
```typescript
function displayPlatformInfo() {
  const env = detectEnvironment();
  
  console.log(`Platform: ${env.platform}`);
  
  if (env.isWSL) {
    console.log(`Environment: WSL ${process.env.WSL_DISTRO_NAME || '2'}`);
    console.log(`✅ Recommended environment`);
  } else if (env.isNativeWindows) {
    console.log(`Environment: Native Windows`);
    console.log(`⚠️  Consider using WSL 2 for better experience`);
  } else {
    console.log(`✅ Supported environment`);
  }
}
```

---

## 📚 문서 구조

```
docs/
├── install/
│   ├── macos.md
│   ├── linux.md
│   ├── windows-wsl.md      # ⭐ 메인 가이드
│   └── windows-native.md   # 비권장
├── guides/
│   ├── wsl-setup.md        # WSL 상세 설정
│   └── vscode-wsl.md       # VS Code 연동
└── troubleshooting/
    └── wsl-issues.md       # WSL 문제 해결
```

---

## 🎯 마이그레이션 전략

### 기존 Windows 사용자

**1. 감지 및 안내**:
```bash
$ moai init
⚠️  Native Windows 환경 감지됨

더 나은 경험을 위해 WSL 2로 마이그레이션하세요:
1. wsl --install
2. WSL 터미널에서 moai-adk 재설치

자세한 가이드: https://moai-adk.dev/docs/migrate-to-wsl
```

**2. 마이그레이션 도구**:
```bash
$ moai migrate-to-wsl
# 자동으로:
# - 현재 프로젝트 설정 백업
# - WSL 설치 확인
# - WSL로 프로젝트 복사
# - 의존성 재설치
```

### 새 사용자

**README 첫 화면**:
```markdown
## Quick Start

### Windows Users
**Use WSL 2** (Recommended)
```powershell
wsl --install
# Restart, then in WSL terminal:
npm install -g moai-adk
```

### macOS / Linux
```bash
npm install -g moai-adk
```
```

---

## ✅ 체크리스트

### 코드 변경
- [ ] 환경 감지 함수 추가
- [ ] Native Windows 경고 메시지
- [ ] WSL 최적화 코드
- [ ] 경로 처리 로직 개선

### 문서 작성
- [ ] WSL 설치 가이드
- [ ] VS Code 연동 가이드
- [ ] 트러블슈팅 문서
- [ ] README 업데이트

### 테스트
- [ ] WSL 환경 테스트
- [ ] Native Windows 테스트 (제한적)
- [ ] GitHub Actions WSL 테스트

### 커뮤니케이션
- [ ] 블로그 포스트 (WSL 권장)
- [ ] Discord/커뮤니티 공지
- [ ] 마이그레이션 가이드

---

## 🚨 주의사항

### Native Windows 지원 유지 이유
1. **기업 환경**: WSL 설치 불가능한 경우
2. **레거시 시스템**: 오래된 Windows 버전
3. **제한된 권한**: 관리자 권한 없는 경우

### 하지만 명확하게
- "지원은 하지만 권장하지 않음"
- "문제 발생 시 WSL 사용 권장"
- "새 기능은 WSL 우선 테스트"

---

## 📈 성공 지표

### 단기 (3개월)
- WSL 사용자 비율 > 70%
- Native Windows 버그 리포트 < 10%

### 장기 (1년)
- WSL 사용자 비율 > 90%
- Native Windows 지원 최소화 검토

---

## 💡 최종 권장사항

1. **즉시 실행**: README에 WSL 권장 추가
2. **다음 릴리스**: 환경 감지 및 경고 추가
3. **문서화**: WSL 설치 가이드 완성
4. **커뮤니티**: WSL 사용 장려 캠페인

**목표**: "Windows에서 MoAI-ADK 쓴다 = WSL 쓴다"
