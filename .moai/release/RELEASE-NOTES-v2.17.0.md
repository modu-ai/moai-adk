# MoAI-ADK v2.17.0 — Meta-Harness Skill + BC-V3R3-007 (정적 스킬 16개 제거)

**Release Date**: 2026-04-27  
**Type**: Minor (breaking change)  
**Branch**: `feat/SPEC-V3R3-HARNESS-001-impl` → `main`  
**Predecessor**: v2.16.0

---

## ⚠️ BREAKING CHANGE — BC-V3R3-007: 정적 스킬 16개 제거

이번 v2.17.0에서 **도메인/프레임워크/라이브러리/플랫폼/도구 범주의 16개 정적 스킬이 제거됩니다**. 이는 MoAI-ADK 아키텍처를 "고정 스킬 카탈로그" 에서 "메타-하니스 기반 동적 생성" 으로 전환하기 위한 의도적인 변화입니다.

### 제거되는 스킬 (16개)

| 분류 | 스킬 ID |
|------|---------|
| **도메인 (5)** | moai-domain-backend, moai-domain-frontend, moai-domain-database, moai-domain-db-docs, moai-domain-mobile |
| **프레임워크 (1)** | moai-framework-electron |
| **라이브러리 (3)** | moai-library-shadcn, moai-library-mermaid, moai-library-nextra |
| **도구 (1)** | moai-tool-ast-grep |
| **플랫폼 (3)** | moai-platform-auth, moai-platform-deployment, moai-platform-chrome-extension |
| **워크플로우 (2)** | moai-workflow-research, moai-workflow-pencil-integration |
| **포맷/데이터 (1)** | moai-formats-data |

### 자동 마이그레이션

`moai update` 명령 하나로 충분합니다:

```bash
cd /path/to/project
moai update
```

결과:
- 제거되는 16개 스킬 → `.moai/archive/skills/v2.16/<skill-id>/` 자동 백업
- `.claude/skills/` 에서 원본 제거
- 새 `moai-meta-harness` 스킬 자동 설치

### Grace Window & 복원

v2.17.x 기간(약 2-4주) 동안 `moai migrate restore-skill <skill-id>` 명령으로 필요한 스킬 복원 가능:

```bash
moai migrate restore-skill moai-domain-backend
moai migrate restore-skill moai-library-mermaid
```

v2.18.0 이상에서는 아카이브 지원 종료 예정입니다.

**자세한 마이그레이션 가이드**: `.moai/release/MIGRATION-v2.17.0.md` 참조

---

## TL;DR

v2.17.0은 **메타-하니스 기반 아키텍처로의 전환**을 완성합니다:

1. **정적 스킬 (38개) → 동적 스킬 생성 (메타-하니스)**: 프로젝트별 요구사항에 맞춘 `my-harness-*` 스킬 자동 생성
2. **16개 스킬 제거 (자동 백업)**: BC-V3R3-007 with grace window
3. **Namespace 분리**: `moai-*` (정적) vs `my-harness-*` (동적)
4. **revfactory/harness 7-Phase workflow 통합**: Apache 2.0 attribution 준수

---

## Highlights

### 1. 메타-하니스 스킬 (Meta-Harness Skill)

새 `moai-meta-harness` 스킬이 revfactory/harness의 7-Phase 워크플로우를 MoAI-ADK에 적응시켜 프로젝트별 맞춤 스킬 자동 생성을 가능하게 합니다.

**7-Phase 워크플로우**:
1. **Discovery**: Socratic 인터뷰 (16개 질문, 4라운드) → 프로젝트 요구사항 파악
2. **Analysis**: 기존 코드베이스 + 브랜드 컨텍스트 분석
3. **Synthesis**: SPEC 문서 작성 (EARS 형식)
4. **Skeleton**: `.moai/harness/main.md` + 확장 파일 생성
5. **Customization**: `.claude/agents/my-harness/*` + `.claude/skills/my-harness-*/*` 생성
6. **Evaluation**: Sprint Contract 채점 (evaluator-active)
7. **Iteration**: 자학습 루프 (LEARNING-001, v2.18.0)

**메타-하니스 사용**:
```bash
/moai project
```
→ Socratic 인터뷰 → 프로젝트 특성 파악 → `my-harness-*` 스킬 생성 (v2.18.0 예정)

### 2. Namespace 분리 (Static vs Dynamic)

v2.17.0부터 스킬은 두 가지 범주로 명확히 분리됩니다:

**정적 스킬 (Static Core — `moai-*` prefix, 23개)**:
- 4x `moai-foundation-*` (cc, core, quality, thinking)
- 10x `moai-workflow-*` (ddd, design-context, design-import, gan-loop, loop, project, spec, tdd, testing, worktree)
- 5x `moai-ref-*` (api-patterns, git-workflow, owasp-checklist, react-patterns, testing-pyramid)
- 1x `moai-design-system`
- 2x `moai-domain-*` (brand-design, copywriting — FROZEN)
- 1x `moai-meta-harness` (NEW)

`moai update` 시 자동 갱신됨.

**동적 스킬 (User-Managed — `my-harness-*` prefix)**:
- 프로젝트별 메타-하니스로 생성
- `moai update` 미변경 (사용자 정의 보호)
- `.claude/skills/my-harness-*/`, `.claude/agents/my-harness/` 에 위치

### 3. 아카이브 + 복원 시스템

16개 제거 스킬은 완전히 사라지지 않고 `.moai/archive/skills/v2.16/` 에 보존됩니다:

**아카이브 구조**:
```
.moai/archive/skills/v2.16/
├── moai-domain-backend/
│   ├── SKILL.md
│   ├── modules/
│   ├── examples.md
│   └── reference.md
├── moai-domain-frontend/
│   └── ...
└── (14 more)
```

**복원 명령** (v2.17.x에서만):
```bash
moai migrate restore-skill moai-domain-backend
moai migrate restore-skill moai-library-nextra -f  # -f to force overwrite
```

### 4. Dry-Run 플래그

마이그레이션 계획을 미리 확인할 수 있습니다:

```bash
moai update --dry-run
```

출력 예:
```
[PLAN] archive: moai-domain-backend → .moai/archive/skills/v2.16/moai-domain-backend
[PLAN] archive: moai-domain-frontend → .moai/archive/skills/v2.16/moai-domain-frontend
...
[PLAN] install: moai-meta-harness (v0.1.0)

total: 16 skills archived, 1 skill installed, 0 user customizations modified
```

실제 파일 변경은 발생하지 않습니다.

---

## Related SPECs (v2.17.0 클러스터)

v2.17.0은 세 개의 관련 SPEC을 통합합니다:

| SPEC ID | 제목 | 역할 |
|---------|------|------|
| **SPEC-V3R3-HARNESS-001** | Meta-Harness Skill — Static Core 22 + Dynamic ∞ | 메타-하니스 신설, 16 스킬 제거, namespace 분리 (이번 릴리스) |
| **SPEC-V3R3-DESIGN-PIPELINE-001** | Design Pipeline Harness | Vibe Design 통합, DTCG token validator (v2.18.0) |
| **SPEC-V3R3-PROJECT-HARNESS-001** | Project Harness Initialization | Socratic 인터뷰 UI, 5-layer integration 메커니즘 (v2.18.0) |

---

## Apache License 2.0 Attribution

메타-하니스 스킬의 7-Phase 워크플로우는 revfactory/harness 프로젝트 (Apache License 2.0)에서 파생되었습니다.

**원본 저장소**: https://github.com/revfactory/harness  
**라이선스**: https://www.apache.org/licenses/LICENSE-2.0

MoAI-ADK에서의 적응:
- revfactory/harness의 7-Phase workflow 구조 유지
- MoAI 에이전트 에코시스템 (manager-spec, manager-tdd/ddd, expert-*, evaluator-active)과 통합
- Frontmatter, trigger declarations, integration points 추가

Apache 2.0 § 4(c) attribution 요구사항 준수:
- `moai-meta-harness/SKILL.md` 상단에 attribution block 포함
- MIGRATION-v2.17.0.md에서 revfactory/harness 명시
- CHANGELOG.md 및 RELEASE-NOTES-v2.17.0.md에서 출처 기록
- `.claude/rules/moai/NOTICE.md` 에서 전체 라이선스 정보 관리

---

## Migration Quick Reference

### 옵션 1: 자동 마이그레이션 (권장)

```bash
moai update
```

✅ 제거 스킬 자동 백업  
✅ meta-harness 설치  
✅ 진행 상황 출력  

### 옵션 2: 계획만 확인

```bash
moai update --dry-run
```

✅ 변경 사항 미리보기  
❌ 실제 변경 없음  

### 옵션 3: 스킬 복원

```bash
moai migrate restore-skill moai-domain-backend
moai migrate restore-skill moai-library-mermaid
```

⚠️ v2.17.x에서만 가능  
⚠️ v2.18.0에서 삭제 예정  

---

## Deprecation Timeline

| 버전 | 상태 | 아카이브 | 복원 |
|------|------|---------|------|
| **v2.17.x** | 현재 (grace window) | ✅ `.moai/archive/skills/v2.16/` | ✅ `moai migrate restore-skill` |
| **v2.18.0+** | 예정 | ❌ 제거 | ❌ 불가능 |

**추천**: v2.17.x 기간에 메타-하니스 기반 마이그레이션을 완료하세요.

---

## What's Next

### v2.18.0 (예정)

- **SPEC-V3R3-PROJECT-HARNESS-001**: Socratic 인터뷰 UI + 5-layer integration
- **SPEC-V3R3-DESIGN-PIPELINE-001**: Vibe Design + DTCG token validator
- **Archive 제거**: `.moai/archive/skills/v2.16/` 삭제
- **`moai migrate restore-skill` 삭제**: 복원 불가능

### v2.18.1+

- **SPEC-V3R3-HARNESS-LEARNING-001**: 메타-하니스 자학습 루프
- **`my-harness-*` 자동 진화**: 프로젝트별 스킬 지속적 개선

---

## Support

- **마이그레이션 가이드**: `.moai/release/MIGRATION-v2.17.0.md`
- **기술 문서**: `SPEC-V3R3-HARNESS-001/spec.md`, `plan.md`
- **문제 해결**: MIGRATION-v2.17.0.md 의 "문제 해결" 섹션
- **GitHub 이슈**: https://github.com/modu-ai/moai-adk/issues

---

## Verification Checklist

- ✅ `moai update` → 16 skills archived, meta-harness installed
- ✅ `moai doctor` → namespace allowlist verified (23 static core)
- ✅ `moai migrate restore-skill <id>` → round-trip validation (SHA-256 match)
- ✅ `moai update --dry-run` → zero filesystem mutations
- ✅ `moai update` (2nd run) → idempotency check
- ✅ TRUST 5 gates: Tested (85%+) / Readable (golangci-lint) / Unified (gofmt) / Secured (archive safety) / Trackable (conventional commits)

---

**v2.17.0 릴리스일**: 2026-04-27  
**Grace Window**: ~2-4주 (v2.18.0 출시 예정 전)  
**커뮤니티**: [GitHub Discussions](https://github.com/modu-ai/moai-adk/discussions)
