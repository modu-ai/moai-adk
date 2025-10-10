# SPEC-PLUGIN-003 구현 계획

## 📋 개요

플러그인 설치 스크립트 작성 및 배포 전략을 수립합니다. bash 스크립트를 우선 구현하고, PowerShell 스크립트는 선택적으로 추가합니다.

---

## 🎯 마일스톤 (우선순위 기반)

### 1차 목표: bash 스크립트 기본 구현

**주요 작업**:
1. `scripts/install.sh` 파일 생성
2. Git 설치 감지 함수 구현
3. Git 클론 방식 구현
4. tar.gz 다운로드 방식 구현 (GitHub Release API)
5. 설치 검증 함수 구현
6. 에러 핸들링 및 사용자 안내 메시지

**산출물**:
- `scripts/install.sh` (실행 가능)
- 설치 검증 테스트 스크립트

### 2차 목표: 에러 처리 고도화

**주요 작업**:
1. 네트워크 오류 처리
2. 권한 오류 처리 (`~/.claude/plugins/` 쓰기 권한)
3. 이미 설치된 경우 덮어쓰기 확인
4. 다운로드 실패 시 재시도 로직
5. 상세한 에러 메시지 템플릿

**산출물**:
- 에러 처리 가이드 문서
- 테스트 케이스 (에러 시나리오)

### 3차 목표: 크로스 플랫폼 지원

**주요 작업**:
1. macOS/Linux 호환성 검증
2. Windows PowerShell 스크립트 작성 (`scripts/install.ps1`)
3. OS 자동 감지 및 스크립트 선택
4. 플랫폼별 진행률 표시 최적화

**산출물**:
- `scripts/install.ps1` (Windows용)
- 플랫폼별 테스트 케이스

### 4차 목표: 배포 및 문서화

**주요 작업**:
1. 설치 스크립트 호스팅 (GitHub Pages 또는 CDN)
2. `docs/installation.md` 문서 작성
3. curl 원라이너 URL 설정
4. Quick Start 가이드에 설치 섹션 추가

**산출물**:
- 공개 설치 URL: `https://moai-adk.dev/install.sh`
- 설치 가이드 문서

---

## 🛠️ 기술적 접근 방법

### 1. Git vs tar.gz 선택 알고리즘

```bash
# install.sh 핵심 로직
if command -v git &> /dev/null; then
    echo "Git detected. Using git clone method..."
    git clone https://github.com/modu-ai/moai-adk ~/.claude/plugins/moai-adk
else
    echo "Git not found. Using tar.gz download method..."
    LATEST_RELEASE=$(curl -s https://api.github.com/repos/modu-ai/moai-adk/releases/latest | grep "tarball_url" | cut -d '"' -f 4)
    curl -L "$LATEST_RELEASE" -o /tmp/moai-adk.tar.gz
    mkdir -p ~/.claude/plugins/moai-adk
    tar -xzf /tmp/moai-adk.tar.gz -C ~/.claude/plugins/moai-adk --strip-components=1
    rm /tmp/moai-adk.tar.gz
fi
```

### 2. GitHub Release API 활용

**API 엔드포인트**:
```bash
GET https://api.github.com/repos/modu-ai/moai-adk/releases/latest
```

**응답 예시** (JSON):
```json
{
  "tag_name": "v0.3.0",
  "tarball_url": "https://api.github.com/repos/modu-ai/moai-adk/tarball/v0.3.0",
  "zipball_url": "https://api.github.com/repos/modu-ai/moai-adk/zipball/v0.3.0"
}
```

**추출 로직**:
```bash
# jq 사용 (선호)
LATEST_RELEASE=$(curl -s https://api.github.com/repos/modu-ai/moai-adk/releases/latest | jq -r '.tarball_url')

# jq 없을 시 grep/cut 대체
LATEST_RELEASE=$(curl -s https://api.github.com/repos/modu-ai/moai-adk/releases/latest | grep "tarball_url" | cut -d '"' -f 4)
```

### 3. 에러 처리 전략

**권한 오류**:
```bash
if [ ! -w ~/.claude/plugins ]; then
    echo "❌ Error: No write permission to ~/.claude/plugins/"
    echo "  → Run: chmod 755 ~/.claude/plugins"
    exit 1
fi
```

**네트워크 오류**:
```bash
if ! curl -s --head https://api.github.com &> /dev/null; then
    echo "❌ Error: Cannot reach GitHub API"
    echo "  → Check your internet connection"
    echo "  → Manual installation: git clone https://github.com/modu-ai/moai-adk ~/.claude/plugins/moai-adk"
    exit 1
fi
```

**이미 설치된 경우**:
```bash
if [ -d ~/.claude/plugins/moai-adk ]; then
    read -p "⚠️ MoAI-ADK already installed. Overwrite? [y/N]: " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Installation cancelled."
        exit 0
    fi
    rm -rf ~/.claude/plugins/moai-adk
fi
```

### 4. 설치 검증 로직

```bash
# plugin.json 존재 확인
if [ ! -f ~/.claude/plugins/moai-adk/plugin.json ]; then
    echo "❌ Error: plugin.json not found. Installation may be corrupted."
    echo "  → Retry installation or check GitHub repository"
    exit 1
fi

# commands/, agents/ 디렉토리 확인
if [ ! -d ~/.claude/plugins/moai-adk/commands ] || [ ! -d ~/.claude/plugins/moai-adk/agents ]; then
    echo "⚠️ Warning: plugin structure incomplete. Plugin may not work correctly."
fi

echo "✅ MoAI-ADK plugin installed successfully!"
echo ""
echo "Next steps:"
echo "  1. Restart Claude Code"
echo "  2. Verify plugin: ls ~/.claude/plugins/moai-adk"
echo "  3. Quick start: /alfred:8-project"
```

### 5. Windows PowerShell 스크립트 (선택)

```powershell
# install.ps1
$INSTALL_PATH = "$env:USERPROFILE\.claude\plugins\moai-adk"

if (Get-Command git -ErrorAction SilentlyContinue) {
    Write-Host "Git detected. Using git clone method..."
    git clone https://github.com/modu-ai/moai-adk $INSTALL_PATH
} else {
    Write-Host "Git not found. Using zip download method..."
    $LATEST = (Invoke-RestMethod -Uri "https://api.github.com/repos/modu-ai/moai-adk/releases/latest").zipball_url
    Invoke-WebRequest -Uri $LATEST -OutFile "$env:TEMP\moai-adk.zip"
    Expand-Archive -Path "$env:TEMP\moai-adk.zip" -DestinationPath $INSTALL_PATH -Force
    Remove-Item "$env:TEMP\moai-adk.zip"
}

# 검증
if (Test-Path "$INSTALL_PATH\plugin.json") {
    Write-Host "✅ MoAI-ADK plugin installed successfully!"
} else {
    Write-Host "❌ Error: Installation failed. plugin.json not found."
    exit 1
}
```

---

## 🏗️ 아키텍처 설계 방향

### 디렉토리 구조

```
moai-adk/
├── scripts/
│   ├── install.sh          # bash 설치 스크립트 (우선)
│   ├── install.ps1         # PowerShell 스크립트 (선택)
│   └── uninstall.sh        # 제거 스크립트 (향후)
├── tests/
│   └── scripts/
│       ├── install.test.sh # bash 테스트
│       └── install.test.ps1 # PowerShell 테스트
└── docs/
    └── installation.md     # 설치 가이드 문서
```

### 설치 스크립트 설계 원칙

1. **멱등성 (Idempotent)**: 여러 번 실행해도 안전
2. **원자성 (Atomic)**: 설치 실패 시 롤백 또는 정리
3. **투명성 (Transparent)**: 모든 작업을 사용자에게 명확히 표시
4. **안전성 (Safe)**: 예외 상황 완벽 처리

---

## ⚠️ 리스크 및 대응 방안

### 리스크 1: GitHub API Rate Limit

**문제**:
- GitHub API는 비인증 요청 시 60회/시간 제한

**대응**:
- 에러 메시지에 수동 설치 가이드 포함
- GitHub Personal Access Token 사용 옵션 추가 (선택)

### 리스크 2: 플러그인 디렉토리 권한

**문제**:
- `~/.claude/plugins/` 디렉토리 쓰기 권한 없을 수 있음

**대응**:
- 권한 확인 로직 추가
- 권한 없을 시 `chmod 755 ~/.claude/plugins` 안내

### 리스크 3: 네트워크 불안정

**문제**:
- GitHub 다운로드 중 네트워크 끊김

**대응**:
- 재시도 로직 (최대 3회)
- 실패 시 수동 설치 가이드 제공

### 리스크 4: 플러그인 구조 변경

**문제**:
- tar.gz 압축 해제 후 구조가 예상과 다를 수 있음

**대응**:
- `plugin.json` 존재 확인으로 검증
- 구조 불일치 시 에러 메시지 출력

---

## 🧪 테스트 전략

### 수동 테스트 시나리오

**정상 케이스**:
1. Git 설치 환경에서 스크립트 실행
2. Git 미설치 환경에서 스크립트 실행
3. 이미 설치된 상태에서 덮어쓰기 선택/거부

**에러 케이스**:
1. 네트워크 끊김 (GitHub API 접근 불가)
2. 플러그인 디렉토리 권한 없음
3. curl/wget 미설치
4. 다운로드 파일 손상 (plugin.json 없음)

### 자동화 테스트 (선택)

```bash
# tests/scripts/install.test.sh
#!/bin/bash

# Mock 환경 설정
export HOME=/tmp/test-home
mkdir -p $HOME/.claude/plugins

# 테스트 1: Git 설치 환경
which git &> /dev/null && echo "✅ Git test passed"

# 테스트 2: tar.gz 다운로드
# (실제 API 호출 대신 mock 데이터 사용)

# 테스트 3: 설치 검증
[ -f $HOME/.claude/plugins/moai-adk/plugin.json ] && echo "✅ Verification test passed"

# 정리
rm -rf $HOME/.claude/plugins
```

---

## 📚 문서화 요구사항

### installation.md 구성

1. **설치 방법 3가지**:
   - curl 원라이너 (권장)
   - Git 클론
   - 수동 다운로드

2. **문제 해결 가이드**:
   - 권한 오류 해결
   - 네트워크 오류 해결
   - 플러그인 미인식 문제

3. **설치 검증 방법**:
   - `ls ~/.claude/plugins/moai-adk`
   - Claude Code 재시작 후 `/alfred:8-project` 실행

---

## 다음 단계

1. **Phase 1**: `scripts/install.sh` 기본 구현 완료
2. **Phase 2**: 에러 처리 고도화 및 테스트
3. **Phase 3**: Windows PowerShell 스크립트 추가
4. **Phase 4**: 문서화 및 배포 URL 설정

**완료 조건**:
- `scripts/install.sh` 실행 가능
- 모든 에러 시나리오 처리 완료
- `docs/installation.md` 문서 작성 완료
- `/alfred:2-build SPEC-PLUGIN-003` 대기
