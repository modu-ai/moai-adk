# v2.17.0 마이그레이션 가이드 — BC-V3R3-007 정적 스킬 제거

**Release Date**: 2026-04-27  
**Breaking Change**: BC-V3R3-007  
**Grace Window**: v2.17.x (v2.18.0에서 아카이브 지원 종료)

---

## 변경사항 요약

v2.17.0은 MoAI-ADK의 아키텍처를 고정 스킬 카탈로그 (38개)에서 **메타-하니스 기반 동적 생성**으로 전환합니다. 이 변경에 따라 도메인/프레임워크/라이브러리/플랫폼/도구 범주의 16개 정적 스킬이 제거되고, 대신 프로젝트별 요구사항에 맞춘 동적 스킬 생성이 가능해집니다.

**영향받는 사용자**: v2.16.x를 사용 중이며 제거되는 16개 스킬 중 하나 이상을 활용하는 프로젝트.

---

## 제거된 16개 스킬

다음 스킬들이 v2.17.0에서 제거됩니다. 모두 `.moai/archive/skills/v2.16/` 디렉토리에 자동 백업되므로 필요 시 복원 가능합니다.

### 도메인 스킬 (5개)
- `moai-domain-backend` — Backend 개발 패턴 및 아키텍처
- `moai-domain-frontend` — Frontend 컴포넌트 설계 및 상태 관리
- `moai-domain-database` — 데이터베이스 설계 및 최적화
- `moai-domain-db-docs` — 데이터베이스 문서화
- `moai-domain-mobile` — 모바일 애플리케이션 개발 (iOS/Android)

### 프레임워크 스킬 (1개)
- `moai-framework-electron` — Electron 데스크톱 애플리케이션 개발

### 라이브러리 스킬 (3개)
- `moai-library-shadcn` — shadcn/ui 컴포넌트 라이브러리
- `moai-library-mermaid` — Mermaid 다이어그램 생성
- `moai-library-nextra` — Nextra 문서 사이트 생성

### 도구 스킬 (1개)
- `moai-tool-ast-grep` — AST 기반 코드 검색 및 분석

### 플랫폼 스킬 (3개)
- `moai-platform-auth` — 인증 및 인가 구현
- `moai-platform-deployment` — 배포 및 DevOps 자동화
- `moai-platform-chrome-extension` — Chrome 확장 프로그램 개발

### 워크플로우 스킬 (2개)
- `moai-workflow-research` — 연구 및 기술 조사
- `moai-workflow-pencil-integration` — Pencil 디자인 도구 통합

### 포맷/데이터 스킬 (1개)
- `moai-formats-data` — 데이터 포맷 변환 및 처리

**총 16개 스킬** (도메인 5 + 프레임워크 1 + 라이브러리 3 + 도구 1 + 플랫폼 3 + 워크플로우 2 + 포맷 1)

---

## 자동 마이그레이션 (권장)

### 1단계: `moai update` 실행

```bash
cd /path/to/project
moai update
```

`moai update`는 제거되는 16개 스킬을 자동으로 감지하여 다음을 수행합니다:

1. 각 스킬의 전체 콘텐츠를 `.moai/archive/skills/v2.16/<skill-id>/`에 복사
2. `.claude/skills/` 디렉토리에서 원본 스킬 삭제
3. 새 `moai-meta-harness` 스킬 설치

### 2단계: 마이그레이션 결과 확인

```
archive: moai-domain-backend → .moai/archive/skills/v2.16/moai-domain-backend
archive: moai-domain-frontend → .moai/archive/skills/v2.16/moai-domain-frontend
...
install: moai-meta-harness (v0.1.0)

total: 16 skills archived, 1 skill installed, 0 user customizations modified
```

### 3단계: 프로젝트별 메타-하니스 생성 (선택)

프로젝트에 맞춘 동적 스킬 생성은 다음 명령으로 시작합니다:

```bash
/moai project
```

이 명령은 Socratic 인터뷰(16개 질문, 4라운드)를 통해 프로젝트 특성을 파악하고,
필요한 `my-harness-*` 스킬들을 동적으로 생성합니다 (v2.18.0 예정).

---

## 수동 마이그레이션 (fallback)

`moai update`를 실행할 수 없는 경우 수동으로 마이그레이션할 수 있습니다.

### 단계 1: 스킬 백업

```bash
# 기존 스킬 디렉토리 백업
cp -r .claude/skills .claude/skills.backup-v2.16
```

### 단계 2: 제거되는 스킬 삭제

위의 "제거된 16개 스킬" 목록에서 16개 디렉토리를 모두 삭제합니다:

```bash
rm -rf .claude/skills/moai-domain-backend
rm -rf .claude/skills/moai-domain-frontend
rm -rf .claude/skills/moai-domain-database
rm -rf .claude/skills/moai-domain-db-docs
rm -rf .claude/skills/moai-domain-mobile
rm -rf .claude/skills/moai-framework-electron
rm -rf .claude/skills/moai-library-shadcn
rm -rf .claude/skills/moai-library-mermaid
rm -rf .claude/skills/moai-library-nextra
rm -rf .claude/skills/moai-tool-ast-grep
rm -rf .claude/skills/moai-platform-auth
rm -rf .claude/skills/moai-platform-deployment
rm -rf .claude/skills/moai-platform-chrome-extension
rm -rf .claude/skills/moai-workflow-research
rm -rf .claude/skills/moai-workflow-pencil-integration
rm -rf .claude/skills/moai-formats-data
```

### 단계 3: 메타-하니스 스킬 설치

MoAI-ADK 템플릿에서 새 스킬 파일을 복사합니다:

```bash
# 템플릿 설치 위치에서 로컬 프로젝트로 복사
cp -r /path/to/moai-adk-go/internal/template/templates/.claude/skills/moai-meta-harness .claude/skills/
```

### 단계 4: 아카이브 수동 생성 (선택)

복원 기능을 사용하려면 수동으로 아카이브 디렉토리를 생성합니다:

```bash
# 백업에서 아카이브 디렉토리로 복사
mkdir -p .moai/archive/skills/v2.16
cp -r .claude/skills.backup-v2.16/moai-domain-backend .moai/archive/skills/v2.16/
cp -r .claude/skills.backup-v2.16/moai-domain-frontend .moai/archive/skills/v2.16/
# ... (16개 모두 반복)
```

---

## 스킬 복원 명령어

제거된 스킬이 필요한 경우, `moai migrate restore-skill` 명령으로 `.moai/archive/skills/v2.16/`에서 복원합니다.

### 문법

```bash
moai migrate restore-skill <skill-id>
```

### 예제 1: 단일 스킬 복원

```bash
moai migrate restore-skill moai-domain-backend
```

결과:
```
restored: moai-domain-backend from archive
restored to: .claude/skills/moai-domain-backend/
```

### 예제 2: 여러 스킬 복원

```bash
moai migrate restore-skill moai-library-mermaid
moai migrate restore-skill moai-platform-deployment
```

### 예제 3: 이미 존재하는 스킬 복원 (강제)

스킬이 이미 존재하면 `-f` (force) 플래그로 덮어쓸 수 있습니다:

```bash
moai migrate restore-skill moai-domain-database -f
```

### 주의사항

- 복원된 스킬은 v2.16.x 버전이므로, v2.17.0 이후 변경사항이 포함되지 않습니다.
- 복원 후 프로젝트를 다시 시작하거나 Claude Code를 재로드하여 새 스킬을 인식하도록 합니다.

---

## 지원 종료 일정 (Deprecation Timeline)

### v2.17.x 기간 (현재)

- ✅ 제거된 스킬 아카이브 유지: `.moai/archive/skills/v2.16/`
- ✅ `moai migrate restore-skill <skill-id>` 복원 지원
- ✅ 자동 마이그레이션 (`moai update`) 완벽 지원
- ⚠️  Grace window 내: 사용자 선택에 따라 언제든지 복원 가능

### v2.18.0 이상

- ❌ 아카이브 디렉토리 제거 예정
- ❌ `moai migrate restore-skill` 명령 삭제 예정
- ✅ `moai-meta-harness` 기반 동적 스킬 생성이 유일한 옵션

**추천**: v2.17.x 기간(약 2-4주)에 필요한 스킬을 복원하고, v2.18.0 이전에 메타-하니스 기반으로 마이그레이션하세요.

---

## 버전 고정 (선택)

v2.17.0 업그레이드를 보류하려면 `go.mod`에서 버전을 고정할 수 있습니다:

```bash
# v2.16.x 계속 사용
moai update --version v2.16.0
```

단, 보안 및 기능 업데이트가 지연되므로 가능한 한 빨리 마이그레이션하시길 권장합니다.

---

## Apache License 2.0 표시

이번 v2.17.0의 메타-하니스 스킬은 revfactory/harness 프로젝트의 7-Phase 워크플로우를 기반으로 합니다.

**원본 저장소**: https://github.com/revfactory/harness  
**라이선스**: Apache License 2.0 (https://www.apache.org/licenses/LICENSE-2.0)

revfactory/harness의 원저작과 개선사항은 Apache 2.0 라이선스 조건(§4(c) attribution 보존, §4(b) 변경 사항 공시)에 따라 MoAI-ADK에 통합되었습니다.

자세한 저작권 표시는 `.claude/rules/moai/NOTICE.md`를 참조하세요.

---

## 문제 해결

### Q: 마이그레이션 후 스킬을 찾을 수 없습니다

**A**: Claude Code를 재시작하거나 새 세션을 시작하여 업데이트된 스킬 카탈로그를 로드합니다.

### Q: 복원할 수 없다고 나옵니다

**A**: 아카이브가 존재하는지 확인합니다:
```bash
ls -la .moai/archive/skills/v2.16/<skill-id>/
```

### Q: 사용자 정의 스킬이 삭제되었습니다

**A**: 이 마이그레이션은 `moai-*` 정적 스킬만 제거하며, `my-harness-*` 사용자 정의 스킬은 보존됩니다.
수동 마이그레이션 중 `my-harness-*` 스킬은 삭제하지 않도록 주의합니다.

### Q: v2.18.0은 언제 출시되나요?

**A**: 현재 계획은 v2.17.0 출시 후 2-4주 이후입니다. 자세한 일정은 [GitHub 릴리스 페이지](https://github.com/modu-ai/moai-adk)를 참조하세요.

---

## 추가 리소스

- **CHANGELOG.md**: v2.17.0 전체 변경사항
- **RELEASE-NOTES-v2.17.0.md**: 릴리스 하이라이트 및 새로운 기능
- **`.claude/rules/moai/NOTICE.md`**: Apache 2.0 저작권 표시
- **[revfactory/harness](https://github.com/revfactory/harness)**: 원본 프로젝트 (Apache 2.0)
