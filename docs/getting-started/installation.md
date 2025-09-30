# 설치

## 사전 요구사항

- Node.js 18.0.0 이상
- Bun 1.2.19 이상 (권장)
- Git 2.28.0 이상
- TypeScript 5.0.0 이상 (선택)

## Bun으로 설치 (권장)

```bash
bun install -g moai-adk
```

## npm으로 설치

```bash
npm install -g moai-adk
```

## 설치 확인

```bash
moai --version
# 출력: 0.0.1

moai doctor
```

`moai doctor` 명령어는 시스템 요구사항을 확인하고 모든 것이 올바르게 설정되었는지 검증합니다.

## 시스템 진단

MoAI-ADK는 프로젝트 언어를 자동으로 감지하고 필요한 개발 도구를 추천합니다:

```bash
moai doctor
```

**진단 출력 예시:**

```
🔍 Checking system requirements...

  Languages: TypeScript, JavaScript

  ⚙️  Runtime:
    ✅ Node.js (18.19.0)
    ✅ Git (2.42.0)

  🛠️  Development:
    ✅ bun (1.2.19)
    ✅ npm (10.2.5)
    ✅ TypeScript (5.9.2)

  📦 Optional:
    ✅ Vitest (3.2.4)
    ✅ Biome (2.2.4)

─────────────────────────────────────────────────────
  📊 Summary:
     Checks: 7 total
     Status: 7 passed
─────────────────────────────────────────────────────

✅ All requirements satisfied!
```

**지능형 언어 감지:**
- JavaScript/TypeScript: npm, TypeScript, Vitest, Biome 추천
- Python: pytest, mypy, ruff 추천
- Java: Maven/Gradle, JUnit 추천
- Go: go test, golint 추천
- Rust: cargo test, rustfmt 추천

## 문제 해결

설치 중 문제가 발생하면:

1. Node.js/Bun 버전 확인
2. 전역 설치 권한 확인
3. `moai doctor --verbose`로 상세 로그 확인
4. GitHub Issues에 문제 보고

### 권한 오류 (macOS/Linux)

```bash
# npm 전역 경로 변경 (권장)
npm config set prefix ~/.npm-global
export PATH=~/.npm-global/bin:$PATH

# 또는 sudo 사용
sudo npm install -g moai-adk
```

### Bun 설치

```bash
# macOS/Linux
curl -fsSL https://bun.sh/install | bash

# Windows (PowerShell)
powershell -c "irm bun.sh/install.ps1 | iex"
```

## 다음 단계

- [빠른 시작](/getting-started/quick-start) 가이드 확인
- [프로젝트 초기화](/getting-started/project-setup) 학습
- [3단계 워크플로우](/guide/workflow) 이해