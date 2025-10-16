# MoAI-ADK (Agentic Development Kit)

[![PyPI version](https://img.shields.io/pypi/v/moai-adk)](https://pypi.org/project/moai-adk/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Python](https://img.shields.io/badge/Python-3.13+-blue)](https://www.python.org/)
[![Tests](https://github.com/modu-ai/moai-adk/actions/workflows/moai-gitflow.yml/badge.svg)](https://github.com/modu-ai/moai-adk/actions/workflows/moai-gitflow.yml)
[![codecov](https://codecov.io/gh/modu-ai/moai-adk/branch/develop/graph/badge.svg)](https://codecov.io/gh/modu-ai/moai-adk)
[![Coverage](https://img.shields.io/badge/coverage-87.66%25-brightgreen)](https://github.com/modu-ai/moai-adk)

## MoAI-ADK: 모두의AI 에이전틱 코딩 개발 프레임워크

**안내**: MoAI-ADK는 모두의AI 연구실에서 집필 중인 "(가칭) 에이전틱 코딩" 서적의 별책 부록 오픈 소스 프로젝트입니다.

![MoAI-ADK CLI Interface](https://github.com/modu-ai/moai-adk/raw/main/docs/public/moai-tui_screen-light.png)

> **"SPEC이 없으면 CODE도 없다."**

---

## 목차

- [v0.3.0 주요 개선사항](#-v030-주요-개선사항)
- [Meet Alfred](#-meet-alfred---10개-ai-에이전트-팀)
- [Quick Start](#-quick-start-3분-실전)
- [3단계 워크플로우](#-3단계-워크플로우)
- [CLI Reference](#-cli-reference)
- [출력 스타일](#-alfreds-output-styles)
- [언어 지원](#-universal-language-support)
- [TRUST 5원칙](#-trust-5원칙)
- [FAQ](#-faq)
- [문제 해결](#-문제-해결)

---

## 🆕 v0.3.0 주요 개선사항

### 🚀 핵심 기능 강화

#### 1. Template Processor 개선 - 안전한 업데이트 시스템
- **Alfred 폴더 자동 백업**: 업데이트 전 `.moai-backups/alfred-{timestamp}/` 폴더에 자동 백업
- **선택적 복사 전략**: Alfred 시스템 폴더만 덮어쓰고, 사용자 커스터마이징 파일 보존
- **지능형 병합**: `product/structure/tech.md` 등 프로젝트 문서를 BackupMerger가 자동으로 병합
- **롤백 지원**: 문제 발생 시 백업에서 복구 가능

#### 2. Event-Driven Checkpoint 시스템
- **자동 백업**: 위험한 작업(`rm -rf`, 병합, 스크립트 실행) 전 자동 checkpoint 생성
- **Hooks 통합**: `SessionStart`, `PreToolUse`, `PostToolUse` 훅이 실시간 감지
- **최대 10개 유지**: FIFO + 7일 보존 정책으로 디스크 효율 관리
- **투명한 동작**: 백그라운드 자동 생성, 사용자에게 알림

#### 3. Hooks vs Agents vs Commands 역할 분리
- **Hooks** (가드레일): 위험 차단, 자동 백업, JIT Context (<100ms)
- **Agents** (분석): SPEC 검증, TRUST 원칙 확인, TAG 관리 (수 초)
- **Commands** (워크플로우): 여러 단계 오케스트레이션 (수 분)

#### 4. Context Engineering 전략 완성
- **JIT Retrieval**: 필요한 순간에만 문서 로드 (초기 컨텍스트 최소화)
- **Compaction**: 토큰 사용량 >70% 시 요약 후 새 세션 시작 권장
- **Explore 에이전트**: 대규모 코드베이스 효율적 탐색 가이드 추가

### 🛠️ 도구 & 명령어 개선

#### CLI 명령어 표준화
```bash
# 새 프로젝트 생성
moai-adk init project-name

# 기존 프로젝트에 설치
moai-adk init .

# 상태 확인
moai-adk status

# 업데이트
moai-adk update
```

#### Alfred 커맨드 단계별 커밋 지침 추가
- **0-project**: 문서 생성 완료 시 커밋
- **1-spec**: SPEC 작성 + Git 브랜치/PR 생성 시 커밋
- **2-build**: TDD 전체 사이클(RED→GREEN→REFACTOR) 완료 시 1회 커밋
- **3-sync**: 문서 동기화 완료 시 커밋

#### PyPI 배포 자동화
- GitHub Actions 워크플로우 추가 (`.github/workflows/publish-pypi.yml`)
- 템플릿 프로젝트에도 배포 워크플로우 제공
- 버전 관리 및 자동 배포 지원

### 📚 문서 강화

#### SPEC 메타데이터 표준 (SSOT)
- **필수 필드 7개**: id, version, status, created, updated, author, priority
- **선택 필드 9개**: category, labels, depends_on, blocks, related_specs, related_issue, scope
- **HISTORY 섹션**: 모든 버전 변경 이력 기록 (필수)
- `.moai/memory/spec-metadata.md`에 전체 가이드 문서화

#### Explore 에이전트 활용 가이드
- 코드 분석 권장 상황 명확화
- thoroughness 레벨별 사용법 (quick/medium/very thorough)
- JIT Retrieval 최적화 전략

### 🔒 보안 & 안정성

#### 크로스 플랫폼 지원 강화
- Windows/macOS/Linux 동일 동작 보장
- 플랫폼별 에러 메시지 제공
- PowerShell + Python 보안 스캔 스크립트

#### .gitignore 및 프로젝트 정리
- 로컬 설정 파일 자동 제외 (`.claude/settings.local.json`)
- 임시 테스트 파일 제외 (`*-test-report.md`)
- 불필요한 파일 자동 정리

### 🎨 출력 스타일 개선

#### 3가지 표준 스타일
- **MoAI Beginner Learning**: 개발 입문자를 위한 친절한 가이드
- **MoAI Professional**: 전문 개발자를 위한 효율적인 출력
- **MoAI Alfred (기본)**: 균형잡힌 AI 협업 스타일

---

## ▶◀ Meet Alfred - 10개 AI 에이전트 팀

안녕하세요, 모두의AI SuperAgent **▶◀ Alfred**입니다!

![Alfred Logo](https://github.com/modu-ai/moai-adk/raw/main/docs/public/alfred_logo.png)

저는 MoAI-ADK의 SuperAgent이자 중앙 오케스트레이터 AI입니다. **10개의 AI 에이전트 팀**(Alfred + 9개 전문 에이전트)을 조율하여 Claude Code 환경에서 완벽한 개발 지원을 제공합니다.

### 🌟 흥미로운 사실: AI가 만든 AI 개발 도구

이 프로젝트의 모든 코드는 **100% AI에 의해 작성**되었습니다.

- **AI 협업 설계**: GPT-5 Pro와 Claude 4.1 Opus가 함께 아키텍처를 설계
- **Agentic Coding 적용**: 10개 AI 에이전트 팀이 자율적으로 SPEC 작성, TDD 구현, 문서 동기화 수행
- **투명성**: 완벽하지 않은 부분을 숨기지 않고, 커뮤니티와 함께 개선해나가는 오픈소스

### ▶◀ Alfred가 제공하는 4가지 핵심 가치

#### 1️⃣ 일관성 (Consistency)
**SPEC → TDD → Sync** 3단계 파이프라인으로 플랑켄슈타인 코드 방지

#### 2️⃣ 품질 (Quality)
**TRUST 5원칙** 자동 적용 및 검증 (Test First, Readable, Unified, Secured, Trackable)

#### 3️⃣ 추적성 (Traceability)
**@TAG 시스템**으로 `@SPEC → @TEST → @CODE → @DOC` 완벽 연결

#### 4️⃣ 범용성 (Universality)
**모든 주요 언어 지원** (Python, TypeScript, Java, Go, Rust, Dart, Swift, Kotlin 등)

---

## 🚀 Quick Start (3분 실전)

### 📋 준비물

- ✅ Python 3.13+ 또는 uv 설치
- ✅ Claude Code 실행 중
- ✅ Git 설치 (선택사항)

### ⚡ 3단계로 시작하기

#### 1️⃣ 설치 (30초)

```bash
# uv 권장 (빠른 성능)
pip install uv
uv pip install moai-adk

# 또는 pip 사용
pip install moai-adk

# 설치 확인
moai-adk --version
```

#### 2️⃣ 초기화 (1분)

**새 프로젝트 생성:**
```bash
moai-adk init my-project
cd my-project

# Claude Code 실행
claude
```

**기존 프로젝트에 설치:**
```bash
cd existing-project
moai-adk init .

# Claude Code 실행
claude
```

**Claude Code에서 프로젝트 초기화 (필수):**
```text
/alfred:0-project
```

Alfred가 자동으로:
- `.moai/project/` 문서 3종 생성 (product/structure/tech.md)
- 언어별 최적 도구 체인 설정
- 프로젝트 컨텍스트 완벽 이해

#### 3️⃣ 첫 기능 개발 (1분 30초)

**Claude Code에서 3단계 워크플로우:**
```text
# SPEC 작성
/alfred:1-spec "JWT 기반 사용자 로그인 API"

# TDD 구현
/alfred:2-build AUTH-001

# 문서 동기화
/alfred:3-sync
```

### 🎉 완료!

**생성된 것들:**
- ✅ `.moai/specs/SPEC-AUTH-001/spec.md` (명세)
- ✅ `tests/test_auth_login.py` (테스트)
- ✅ `src/auth/service.py` (구현)
- ✅ `docs/api/auth.md` (문서)
- ✅ `@SPEC → @TEST → @CODE → @DOC` TAG 체인

---

## ⬆️ 업그레이드 가이드 (v0.2.x → v0.3.0)

### 1단계: 패키지 업데이트

```bash
# pip
pip install --upgrade moai-adk

# uv 권장
uv pip install --upgrade moai-adk
```

### 2단계: 프로젝트 업데이트

```bash
cd your-project
moai-adk update
```

**자동 백업**: 업데이트 전 `.moai-backups/{timestamp}/`에 자동 백업 생성

### 3단계: Claude Code 최적화

```text
claude
/alfred:0-project
```

병합 프롬프트에서 **Merge** 선택 → 기존 문서 유지 + 새 템플릿 추가

### 검증 체크리스트

```bash
# 상태 확인
moai-adk status

# 확인 항목
# ✅ .moai/config.json → project.moai_adk_version: "0.3.x"
# ✅ .moai/config.json → project.optimized: true
# ✅ 모든 커맨드 정상 작동
```

---

## 🔄 3단계 워크플로우

Alfred의 핵심은 **체계적인 3단계 워크플로우**입니다.

### 1️⃣ SPEC - 명세 작성

**명령어**: `/alfred:1-spec "JWT 기반 사용자 로그인 API"`

**Alfred가 자동 수행:**
- EARS 형식 명세 자동 생성
- `@SPEC:ID` TAG 부여
- Git 브랜치 자동 생성 (Team 모드)
- Draft PR 생성 (Team 모드)
- HISTORY 섹션 자동 추가

**산출물:**
- `.moai/specs/SPEC-AUTH-001/spec.md`
- `.moai/specs/SPEC-AUTH-001/plan.md`
- `.moai/specs/SPEC-AUTH-001/acceptance.md`

### 2️⃣ BUILD - TDD 구현

**명령어**: `/alfred:2-build AUTH-001`

**Alfred가 자동 수행:**
- **RED**: 실패하는 테스트 작성
- **GREEN**: 최소 구현으로 테스트 통과
- **REFACTOR**: 코드 품질 개선
- TRUST 5원칙 자동 검증
- 단계별 Git 커밋 (TDD 완료 시 1회)

**산출물:**
- `tests/test_auth_login.py` (테스트 코드)
- `src/auth/service.py` (구현 코드)
- `@TEST:AUTH-001` → `@CODE:AUTH-001` TAG 체인

### 3️⃣ SYNC - 문서 동기화

**명령어**: `/alfred:3-sync`

**Alfred가 자동 수행:**
- Living Document 업데이트
- TAG 시스템 무결성 검증
- sync-report.md 생성
- PR Ready 전환 (Team 모드)
- 선택적 자동 머지 (`--auto-merge`)

**산출물:**
- `docs/api/auth.md` (API 문서)
- `.moai/reports/sync-report.md`
- `@DOC:AUTH-001` TAG 추가

---

## 🛠️ CLI Reference

### 프로젝트 관리

```bash
# 새 프로젝트 생성
moai-adk init project-name

# 기존 프로젝트에 설치
moai-adk init .

# 프로젝트 상태 확인
moai-adk status

# 프로젝트 업데이트
moai-adk update

# 시스템 진단
moai-adk doctor

# 버전 확인
moai-adk --version

# 도움말
moai-adk --help
```

### Alfred 커맨드 (Claude Code 내)

```text
# 프로젝트 초기화
/alfred:0-project

# SPEC 작성
/alfred:1-spec "기능 설명"
/alfred:1-spec SPEC-001 "수정 내용"

# TDD 구현
/alfred:2-build SPEC-001
/alfred:2-build all

# 문서 동기화
/alfred:3-sync
/alfred:3-sync --auto-merge
/alfred:3-sync force
```

---

## 🎨 Alfred's Output Styles

### 3가지 표준 스타일

#### 1. MoAI Beginner Learning (학습 전용)
- **대상**: 개발 입문자, 프로그래밍 초보자
- **특징**: 친절한 설명, 단계별 안내, 격려와 응원
- **사용**: `/output-style beginner-learning`

#### 2. MoAI Professional (실무 전용)
- **대상**: 시니어 개발자, 프로덕션 환경
- **특징**: 간결한 출력, 빠른 의사결정, 효율 중심
- **사용**: `/output-style alfred-pro`

#### 3. MoAI Alfred (기본)
- **대상**: 일반 개발자, 균형잡힌 협업
- **특징**: 체계적인 보고, 명확한 구조, 검증 중심
- **사용**: `/output-style agentic-coding` (기본값)

---

## 🌍 Universal Language Support

Alfred는 **모든 주요 프로그래밍 언어**를 지원하며, 각 언어에 최적화된 도구 체인을 자동으로 선택합니다.

### 지원 언어 & 도구

| 언어 | 테스트 프레임워크 | 린터/포매터 | 빌드 도구 |
|------|------------------|-------------|----------|
| **Python** | pytest, mypy | ruff, black | uv, pip |
| **TypeScript** | Vitest, Jest | Biome, ESLint | npm, pnpm |
| **Java** | JUnit | Checkstyle | Maven, Gradle |
| **Go** | go test | gofmt, golint | go build |
| **Rust** | cargo test | rustfmt, clippy | cargo |
| **Dart** | flutter test | dart analyze | flutter |
| **Swift** | XCTest | SwiftLint | xcodebuild |
| **Kotlin** | JUnit | ktlint | Gradle |

### 자동 언어 감지

Alfred는 다음 파일을 자동으로 감지합니다:
- `pyproject.toml`, `requirements.txt` → Python
- `package.json`, `tsconfig.json` → TypeScript
- `pom.xml`, `build.gradle` → Java
- `go.mod` → Go
- `Cargo.toml` → Rust
- `pubspec.yaml` → Dart

---

## 🛡️ TRUST 5원칙

Alfred가 모든 코드에 자동으로 적용하는 품질 기준입니다.

### T - Test First (테스트 우선)
- SPEC 기반 테스트 케이스 작성
- TDD RED → GREEN → REFACTOR 사이클
- 테스트 커버리지 ≥ 85%

### R - Readable (가독성)
- 파일 ≤ 300 LOC
- 함수 ≤ 50 LOC
- 매개변수 ≤ 5개
- 복잡도 ≤ 10

### U - Unified (통일성)
- 타입 안전성 또는 런타임 검증
- 아키텍처 일관성
- 코딩 스타일 통일

### S - Secured (보안)
- 입력 검증
- 로깅 및 감사
- 비밀 관리
- 정적 분석

### T - Trackable (추적성)
- `@SPEC → @TEST → @CODE → @DOC` TAG 체인
- CODE-FIRST 원칙 (코드 직접 스캔)
- HISTORY 섹션 기록

### 자동 검증

```text
# TDD 구현 완료 후 자동 실행
/alfred:2-build AUTH-001

# 또는 수동 실행
/alfred:3-sync

# trust-checker 에이전트가 자동으로 검증:
# ✅ Test Coverage: 87% (목표: 85%)
# ✅ Code Constraints: 모든 파일 300 LOC 이하
# ✅ TAG Chain: 무결성 확인 완료
```

---

## ❓ FAQ

### Q1: MoAI-ADK는 어떤 프로젝트에 적합한가요?

**A**: 다음과 같은 프로젝트에 적합합니다:
- ✅ 새로운 프로젝트 (그린필드)
- ✅ 기존 프로젝트 (레거시 도입)
- ✅ 개인 프로젝트 (Personal 모드)
- ✅ 팀 프로젝트 (Team 모드, GitFlow 지원)
- ✅ 모든 주요 프로그래밍 언어

### Q2: Claude Code가 필수인가요?

**A**: 네, MoAI-ADK는 Claude Code 환경에서 동작하도록 설계되었습니다. Claude Code는 Anthropic의 공식 CLI 도구로, AI 에이전트 시스템을 완벽하게 지원합니다.

### Q3: 기존 프로젝트에 도입할 수 있나요?

**A**: 네, `moai-adk init .` 명령으로 기존 프로젝트에 안전하게 설치할 수 있습니다. Alfred는 기존 코드 구조를 분석하여 `.moai/` 폴더에 문서와 설정만 추가합니다.

### Q4: Personal 모드와 Team 모드의 차이는?

**A**:
- **Personal 모드**: 로컬 작업 중심, 체크포인트만 생성
- **Team 모드**: GitFlow 지원, Draft PR 자동 생성, develop 브랜치 기반

### Q5: SPEC 메타데이터는 어떻게 관리하나요?

**A**: `.moai/memory/spec-metadata.md`에 전체 가이드가 있습니다.
- **필수 7개**: id, version, status, created, updated, author, priority
- **선택 9개**: category, labels, depends_on, blocks, related_specs, related_issue, scope
- **HISTORY 섹션**: 모든 변경 이력 기록 (필수)

### Q6: TDD 단계별로 커밋하나요?

**A**: 아니요, v0.3.0부터 **TDD 전체 사이클(RED→GREEN→REFACTOR) 완료 후 1회만 커밋**합니다. 이전처럼 각 단계별로 3번 커밋하지 않습니다.

### Q7: Context Engineering이란?

**A**:
- **JIT Retrieval**: 필요한 순간에만 문서 로드 (초기 컨텍스트 최소화)
- **Compaction**: 토큰 사용량 >70% 시 요약 후 새 세션 권장
- **Explore 에이전트**: 대규모 코드베이스 효율적 탐색

### Q8: 자동 백업은 어떻게 작동하나요?

**A**:
- **Template Processor**: 업데이트 전 `.moai-backups/alfred-{timestamp}/` 자동 백업
- **Event-Driven Checkpoint**: 위험한 작업 전 자동 checkpoint 생성
- **보존 정책**: 최대 10개 유지, 7일 후 자동 정리

---

## 🔧 문제 해결

### 설치 문제

```bash
# Python 버전 확인 (3.13+ 필요)
python --version

# uv 설치 (권장)
pip install uv

# 캐시 정리 후 재설치
pip cache purge
uv pip install moai-adk --force-reinstall
```

### 초기화 문제

```bash
# 프로젝트 상태 확인
moai-adk status

# 시스템 진단
moai-adk doctor

# 강제 재초기화
moai-adk init . --force
```

### Claude Code 문제

```text
# 설정 확인
ls -la .claude/

# Alfred 커맨드 확인
ls -la .claude/commands/alfred/

# 출력 스타일 확인
/output-style agentic-coding
```

### 일반적인 에러

#### 에러: "moai-adk: command not found"
```bash
# PATH 확인 및 전체 경로 사용
~/.local/bin/moai-adk --version

# 또는 pip로 재설치
pip install --force-reinstall moai-adk
```

#### 에러: ".moai/ 디렉토리를 찾을 수 없습니다"
```bash
# 초기화 실행
moai-adk init .

# 또는 Claude Code에서
/alfred:0-project
```

#### 에러: "SPEC ID 중복"
```bash
# 기존 SPEC 확인
rg "@SPEC:" -n .moai/specs/

# 새로운 ID 사용
/alfred:1-spec "새 기능 설명"
```

---

## 📚 문서 및 지원

### 공식 문서
- **GitHub Repository**: https://github.com/modu-ai/moai-adk
- **PyPI Package**: https://pypi.org/project/moai-adk/
- **Issue Tracker**: https://github.com/modu-ai/moai-adk/issues
- **Discussions**: https://github.com/modu-ai/moai-adk/discussions

### 커뮤니티
- **GitHub Discussions**: 질문, 아이디어, 피드백 공유
- **Issue Tracker**: 버그 리포트, 기능 요청
- **Email**: support@moduai.kr

### 기여하기

MoAI-ADK는 오픈소스 프로젝트입니다. 여러분의 기여를 환영합니다!

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### 라이선스

MIT License - 자유롭게 사용하실 수 있습니다.

---

## 🙏 감사의 말

MoAI-ADK는 다음 프로젝트와 커뮤니티의 도움으로 만들어졌습니다:

- **Anthropic Claude Code**: AI 에이전트 시스템의 기반
- **OpenAI GPT Models**: 초기 설계 협업
- **Python & TypeScript Communities**: 언어 지원 및 도구 체인
- **모두의AI Community**: 지속적인 피드백과 개선 아이디어

---

**Made with ❤️ by MoAI Team**

**▶◀ Alfred**: "여러분의 개발 여정을 함께하겠습니다!"
