# MoAI-ADK Skills Quick Reference

**Last Updated**: 2025-10-20 | **Total**: 44 Skills | **Status**: ✅ **100% OPERATIONAL**

---

## 🎯 Skills 빠른 찾기

### 🏛️ Foundation Tier (핵심 기반)

```markdown
# SPEC 작성
→ moai-foundation-specs      (메타데이터 검증)
→ moai-foundation-ears       (EARS 방식 작성)

# 추적성 관리
→ moai-foundation-tags       (@TAG 시스템)
→ moai-foundation-trust      (TRUST 5원칙)

# 워크플로우
→ moai-foundation-git        (Git 자동화)
→ moai-foundation-langs      (언어 감지)
```

### 🛠️ Essentials Tier (필수 도구)

```markdown
# 문제 해결
→ moai-essentials-debug      (디버깅 + 오류 분석)

# 코드 품질
→ moai-essentials-review     (코드 리뷰)
→ moai-essentials-refactor   (리팩토링 가이드)

# 성능
→ moai-essentials-perf       (성능 최적화)
```

### 🌐 Domain Tier (전문 영역)

```markdown
# 아키텍처
→ moai-domain-backend        (백엔드 설계)
→ moai-domain-frontend       (프론트엔드)
→ moai-domain-web-api        (API 설계)

# 데이터
→ moai-domain-database       (DB 최적화)
→ moai-domain-data-science   (데이터 분석)
→ moai-domain-ml             (머신러닝)

# 배포 & 보안
→ moai-domain-devops         (CI/CD, Docker, K8s)
→ moai-domain-security       (보안, OWASP)

# 특화
→ moai-domain-cli-tool       (CLI 도구)
→ moai-domain-mobile-app     (모바일 앱)
→ moai-claude-code           (Claude Code 관리)
```

### 💻 Language Tier (언어별)

#### Mainstream (8)
```markdown
→ moai-lang-python           (pytest, mypy, ruff)
→ moai-lang-typescript       (Vitest, Biome)
→ moai-lang-javascript       (Jest, ESLint)
→ moai-lang-go               (go test, golint)
→ moai-lang-rust             (cargo, clippy)
→ moai-lang-java             (JUnit, Maven)
→ moai-lang-csharp           (xUnit, .NET)
→ moai-lang-php              (PHPUnit, Composer)
```

#### JVM (4)
```markdown
→ moai-lang-kotlin           (JUnit, Gradle)
→ moai-lang-scala            (ScalaTest, sbt)
→ moai-lang-clojure          (clojure.test, Leiningen)
→ moai-lang-julia            (Test, Pkg)
```

#### Systems (4)
```markdown
→ moai-lang-cpp              (Google Test, CMake)
→ moai-lang-c                (Unity, Make)
→ moai-lang-shell            (bats, shellcheck)
→ moai-lang-lua              (busted, luacheck)
```

#### Mobile & Data (4)
```markdown
→ moai-lang-dart             (flutter test)
→ moai-lang-swift            (XCTest, SwiftLint)
→ moai-lang-r                (testthat, lintr)
→ moai-lang-sql              (pgTAP, SQL testing)
```

#### Functional (3)
```markdown
→ moai-lang-elixir           (ExUnit, Mix)
→ moai-lang-haskell          (HUnit, Stack)
→ moai-lang-ruby             (RSpec, Bundler)
```

---

## 📚 Skill Categories by Purpose

### TDD & Testing
- moai-foundation-specs, moai-foundation-ears, moai-foundation-trust
- moai-essentials-review, moai-essentials-debug
- All moai-lang-* skills

### Code Quality & Performance
- moai-essentials-review (SOLID, code smells)
- moai-essentials-refactor (design patterns)
- moai-essentials-perf (profiling, optimization)
- moai-essentials-debug (error analysis)

### Architecture & Design
- moai-domain-backend (server patterns)
- moai-domain-frontend (UI/UX patterns)
- moai-domain-web-api (REST/GraphQL)
- moai-domain-database (schema design)

### Deployment & Operations
- moai-domain-devops (CI/CD, containers)
- moai-domain-security (OWASP, scanning)
- moai-foundation-git (Git workflows)

### Data & AI
- moai-domain-data-science (analysis, visualization)
- moai-domain-ml (model training, deployment)
- moai-lang-python, moai-lang-r, moai-lang-julia

### Domain-Specific
- moai-domain-cli-tool (command-line tools)
- moai-domain-mobile-app (Flutter, React Native)
- moai-claude-code (Claude Code components)

### Traceability & Documentation
- moai-foundation-tags (CODE-FIRST tracking)
- moai-foundation-git (versioning)
- All Tier 3+ skills (integration)

---

## 🎓 Skill Usage Patterns

### SPEC 작성 단계
```
사용자: /alfred:1-plan "새 기능"
    ↓
활성 skills:
  • moai-foundation-specs    ← 메타데이터 검증
  • moai-foundation-ears     ← 요구사항 구조화
  • moai-foundation-langs    ← 언어 감지
  • moai-domain-*            ← 도메인별 아키텍처
  • moai-foundation-git      ← 브랜치 생성
```

### TDD 구현 단계
```
사용자: /alfred:2-run SPEC-AUTH-001
    ↓
활성 skills:
  • moai-lang-*              ← 언어별 TDD 도구
  • moai-essentials-*        ← 품질 관리
  • moai-foundation-trust    ← TRUST 검증
  • moai-domain-*            ← 도메인 패턴
  • moai-essentials-debug    ← 오류 해결
```

### 문서 동기화 단계
```
사용자: /alfred:3-sync
    ↓
활성 skills:
  • moai-foundation-tags     ← TAG 체인 검증
  • moai-foundation-specs    ← 메타데이터 확인
  • moai-essentials-review   ← 코드 리뷰
  • moai-foundation-git      ← PR 상태 전환
  • moai-foundation-trust    ← 최종 검증
```

---

## 🔍 Skill Discovery

### 언어별 찾기
```bash
# Python 작업?
→ moai-lang-python

# TypeScript/React?
→ moai-lang-typescript
→ moai-domain-frontend

# 모바일 앱?
→ moai-lang-dart (Flutter)
→ moai-lang-swift (iOS)
→ moai-domain-mobile-app
```

### 도메인별 찾기
```bash
# API 설계?
→ moai-domain-web-api
→ moai-domain-backend

# 데이터 분석?
→ moai-domain-data-science
→ moai-lang-python
→ moai-lang-r

# 배포 자동화?
→ moai-domain-devops
→ moai-foundation-git

# 보안?
→ moai-domain-security
→ moai-essentials-review
```

### 작업별 찾기
```bash
# 오류 해결
→ moai-essentials-debug

# 코드 정리
→ moai-essentials-refactor

# 성능 개선
→ moai-essentials-perf

# 리뷰 요청
→ moai-essentials-review

# SPEC 검증
→ moai-foundation-specs

# TAG 확인
→ moai-foundation-tags
```

---

## 📊 Statistics

| Category | Count | Coverage |
|----------|-------|----------|
| **Languages** | 23 | Complete mainstream + specialized |
| **Domains** | 11 | Full stack development |
| **Testing Frameworks** | 20+ | Covered per language |
| **Build Systems** | 15+ | Maven, Gradle, npm, cargo, etc. |
| **Linters/Formatters** | 25+ | Language standards included |

---

## ✅ Verification Checklist

Before starting a project:

- [ ] Run `/alfred:0-project` for setup
- [ ] Detect language with moai-foundation-langs
- [ ] Check relevant language skill loaded
- [ ] Check relevant domain skill loaded
- [ ] Verify SPEC with moai-foundation-specs
- [ ] Write EARS requirements with moai-foundation-ears
- [ ] Start TDD with `/alfred:2-run`
- [ ] Validate with moai-foundation-trust
- [ ] Sync documentation with `/alfred:3-sync`

---

## 🚀 Quick Start Commands

```bash
# List all available skills
ls .claude/skills/moai-*

# Invoke a specific skill
/skill moai-lang-python

# Check skill documentation
cat .claude/skills/moai-lang-python/SKILL.md

# Verify all skills loaded
ls -la .claude/skills/moai-*/*.md | wc -l  # Should be 44
```

---

## 📞 Support

**All skills are production-ready and fully integrated with:**
- ✅ MoAI-ADK 3-stage pipeline
- ✅ Claude Code ecosystem
- ✅ SPEC-First TDD methodology
- ✅ Multi-language support

**Next Steps:**
1. Choose your domain (frontend, backend, data science, etc.)
2. Identify your language (Python, TypeScript, Go, etc.)
3. Use `/alfred:1-plan` to create SPEC
4. Use `/alfred:2-run` to implement TDD
5. Use `/alfred:3-sync` to finalize documentation

---

**Happy coding! 🎉**
