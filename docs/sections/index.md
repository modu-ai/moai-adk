# MoAI-ADK Documentation Index

> **AI Navigation Guide**: 빠른 문서 검색을 위한 체계적 인덱스
> **Last Updated**: 2025-09-23

---

## 🚀 Getting Started

### [01-overview.md](01-overview.md) - 시스템 개요
**핵심 내용**: MoAI-ADK 소개, 핵심 가치, Spec-First TDD 철학, 4단계 파이프라인 개요
**키워드**: `overview`, `introduction`, `spec-first`, `tdd`, `pipeline`, `agentic-development`
**난이도**: 🟢 Basic

### [05-installation.md](05-installation.md) - 설치 및 초기화
**핵심 내용**: pip 설치, `moai init`, 환경 설정, 업데이트 시스템, 검증 방법
**키워드**: `installation`, `setup`, `pip`, `moai-init`, `update`, `verification`
**난이도**: 🟢 Basic

### [02-changelog.md](02-changelog.md) - 변경 이력
**핵심 내용**: 버전별 변경사항, v0.1.21 Hook 안정성/버전 동기화, v0.1.17 패키지 구조 개선
**키워드**: `changelog`, `version`, `updates`, `v0.1.21`, `hooks`, `v0.1.17`, `package-restructure`
**난이도**: 🟢 Basic

---

## 🏗️ Core Architecture

### [03-principles.md](03-principles.md) - 핵심 원칙
**핵심 내용**: Constitution 5원칙, Simplicity/Architecture/Testing/Observability/Versioning
**키워드**: `principles`, `constitution`, `simplicity`, `architecture`, `testing`
**난이도**: 🟡 Intermediate

### [04-architecture.md](04-architecture.md) - 프로젝트 아키텍처
**핵심 내용**: .claude/ 디렉토리 구조, .moai/ 시스템, 파일 조직, 템플릿 시스템
**키워드**: `architecture`, `directory-structure`, `claude-code`, `moai-system`
**난이도**: 🟡 Intermediate

### [package-structure.md](package-structure.md) - 패키지 구조 (NEW)
**핵심 내용**: cli/, core/, install/ 서브패키지, 모듈별 책임, import 가이드
**키워드**: `package-structure`, `cli`, `core`, `install`, `modules`, `imports`
**난이도**: 🟡 Intermediate

---

## 🛠️ Development Workflow

### [07-pipeline.md](07-pipeline.md) - 4단계 파이프라인
**핵심 내용**: `/moai:0-project` → `/moai:3-sync` 신규 워크플로우, 개인/팀 모드별 흐름, 체크포인트/PR 연동
**키워드**: `pipeline`, `0-project`, `1-spec`, `2-build`, `3-sync`, `personal`, `team`
**난이도**: 🟡 Intermediate

### [08-commands.md](08-commands.md) - CLI 명령어 시스템
**핵심 내용**: `/moai:0-project~3-sync` + Git 전용 명령어 5종(`/moai:git:*`), 모드별 동작
**키워드**: `commands`, `git`, `checkpoint`, `rollback`, `branch`, `commit`, `sync`
**난이도**: 🟢 Basic

### [06-wizard.md](06-wizard.md) - 대화형 마법사
**핵심 내용**: 대화형 설정, 프로젝트 초기화, 구성 마법사, 설정 변경
**키워드**: `wizard`, `interactive`, `setup`, `configuration`, `initialization`
**난이도**: 🟢 Basic

### [build-system.md](build-system.md) - 빌드 및 버전 관리 ⭐ NEW
**핵심 내용**: 자동 빌드, 버전 동기화, Makefile, CI/CD 통합, VersionSyncManager (BUILD.md + v0.1.17 내용 통합)
**키워드**: `build`, `version-sync`, `makefile`, `ci-cd`, `automation`
**난이도**: 🔴 Advanced

---

## 🤖 Advanced Features

### [10-agents.md](10-agents.md) - Agent 시스템
**핵심 내용**: project-manager, spec-builder, code-builder, doc-syncer, git-manager, cc-manager, codex-bridge, gemini-bridge 등 협업 구조와 브레인스토밍 설정
**키워드**: `agents`, `brainstorming`, `codex`, `gemini`, `parallel`
**난이도**: 🔴 Advanced

### [11-hooks.md](11-hooks.md) - Hook 시스템
**핵심 내용**: Session/PreToolUse/PostToolUse 훅 구성, steering_guard 등 보안 훅, Python 기반 자동 검증 흐름
**키워드**: `hooks`, `validation`, `security`, `pre-post`
**난이도**: 🔴 Advanced

### [12-tag-system.md](12-tag-system.md) - TAG 추적성 시스템
**핵심 내용**: 16-Core @TAG, 추적성 체인, 자동 인덱싱, 무결성 검사
**키워드**: `tags`, `traceability`, `16-Core`, `indexing`, `integrity`
**난이도**: 🔴 Advanced

### [09-output-styles.md](09-output-styles.md) - 출력 스타일
**핵심 내용**: 5가지 출력 모드, expert/beginner/study/mentor/audit 스타일
**키워드**: `output-styles`, `expert`, `beginner`, `mentor`, `audit`
**난이도**: 🟡 Intermediate

---

## ⚙️ Configuration & Templates

### [13-config.md](13-config.md) - 설정 파일 관리
**핵심 내용**: settings.json, config.json, 성능 설정, 글로벌 설정
**키워드**: `configuration`, `settings`, `config-files`, `performance`
**난이도**: 🟡 Intermediate

### [14-templates.md](14-templates.md) - 템플릿 시스템
**핵심 내용**: 동적 템플릿, 변수 주입, SPEC/Steering 템플릿, TemplateEngine
**키워드**: `templates`, `dynamic`, `template-engine`, `spec`, `steering`
**난이도**: 🔴 Advanced

### [15-development-guide.md](15-development-guide.md) - Constitution 거버넌스
**핵심 내용**: 프로젝트 거버넌스, Constitution 업데이트, 체크리스트, 품질 게이트
**키워드**: `constitution`, `governance`, `quality-gate`, `checklist`
**난이도**: 🔴 Advanced

---

## 📚 Quick Reference

### 자주 사용하는 명령어
```bash
# 설치 및 초기화
pip install moai-adk
moai init --personal   # 또는 --team

# 4단계 워크플로우
/moai:0-project        # 프로젝트 문서 갱신 + 메모리 반영
/moai:1-spec           # auto 제안(개인: 로컬 생성 / 팀: GitHub Issue)
/moai:2-build          # TDD (개인: 체크포인트 / 팀: 7단계 커밋)
/moai:3-sync           # 문서/PR 동기화 + 상태 보고

# Git 전용 명령어(5종)
/moai:git:checkpoint   # 자동/수동 체크포인트
/moai:git:rollback     # 체크포인트 롤백
/moai:git:branch       # 모드별 브랜치 전략
/moai:git:commit       # Constitution 기반 커밋
/moai:git:sync         # 원격 동기화

# 상태/검증
moai status
python .moai/scripts/check-traceability.py --update
```

### 중요 파일 경로
- **프로젝트 메모리**: `CLAUDE.md`
- **Constitution**: `.moai/memory/development-guide.md`
- **설정**: `.claude/settings.json`, `.moai/config.json`
- **Git 스크립트**: `.moai/scripts/branch_manager.py`, `commit_helper.py`, `rollback.py`
- **Hook 스크립트**: `.claude/hooks/moai/` (auto_checkpoint, file_watcher 등)

### 문제 해결
- **Hook 실행 실패** → 권한 확인: `chmod +x .claude/hooks/moai/*.py`
- **TAG 불일치** → 자동 복구: `python scripts/repair_tags.py --execute`
- **빌드 실패** → 클린 빌드: `make build-clean`

---

## 🔍 Search Tags for AI
`moai-adk` `claude-code` `spec-first` `tdd` `agentic-development` `pipeline` `automation` `agents` `hooks` `tags` `constitution` `build-system` `templates` `configuration` `installation` `architecture` `package-structure` `cli` `core` `install` `version-management`

---

*📝 이 인덱스는 AI 검색 최적화를 위해 설계되었습니다. 각 문서의 핵심 내용과 키워드를 포함하여 빠른 탐색이 가능합니다.*
