---
name: alfred:9-update
description: MoAI-ADK 패키지 및 템플릿 업데이트 (백업 자동 생성, 설정 파일 보존)
argument-hint: [--check|--force|--check-quality]
tools: Read, Write, Bash, Grep, Glob
---

<!-- @DOC:UPDATE-REFACTOR-001 | SPEC: SPEC-UPDATE-REFACTOR-001.md -->

# 🔄 MoAI-ADK 프로젝트 업데이트

## HISTORY

### v0.2.6 (TBD) - 지능형 병합 시스템 🔄
- **ADDED**: Step 10 (CLAUDE.md) 지능형 병합 로직
  - 기존: 사용자 수정 → 보존만 (업데이트 없음)
  - 개선: **템플릿 최신 구조 + 프로젝트 정보 유지** 병합
  - 추출: 프로젝트 이름, 설명, 버전, 모드
  - 주입: 템플릿 {{변수}}에 사용자 값 삽입
  - 결과: Alfred 최신 기능 + 사용자 프로젝트 정보 유지
- **ADDED**: Step 11 (.moai/config.json) 스마트 딥 병합
  - JSON 필드별 병합 전략 정의 (표 형식 문서화)
  - `project.*` → 사용자 값 100% 유지
  - `constitution.*` → 사용자 정책 유지, 새 필드는 템플릿 추가
  - `git_strategy.*` → 사용자 워크플로우 완전 보존
  - `tags.categories` → 템플릿 + 사용자 카테고리 병합
  - `pipeline.*` → 최신 명령어 + 사용자 진행 상태 유지
- **IMPROVED**: 데이터 보호 전략 3단계 체계화
  1. 완전 보존 🔒: SPEC, 리포트, 프로젝트 문서
  2. 지능형 병합 🔄: CLAUDE.md, config.json
  3. 템플릿 교체 ✅: 시스템 파일
- **IMPROVED**: 출력 메시지 업데이트
  - "🔄 CLAUDE.md (템플릿 최신화 + 프로젝트 정보 유지)"
  - "🔄 config.json (스마트 병합: 템플릿 구조 + 사용자 설정)"
- **ADDED**: JSON 파싱 실패 시 안전 모드
  - 백업 생성 → 템플릿으로 교체 → 수동 복구 안내
- **PRINCIPLE**: 사용자 데이터 손실 제로, 자동 업데이트 + 수동 개입 최소화
- **AUTHOR**: @alfred
- **RATIONALE**: 템플릿 최신화와 사용자 데이터 보존 양립

### v0.2.5 (2025-10-06) - 사용자 작업물 완전 보존 🔒
- **IMPROVED**: Step 7-9 (프로젝트 문서) 보호 정책 강화
  - 기존: 사용자 수정 → 백업 후 덮어쓰기 (내용 손실)
  - 개선: 사용자 수정 → **완전 보존** (덮어쓰지 않음)
- **IMPROVED**: Step 10 (CLAUDE.md) 보호 정책 강화
  - `{{PROJECT_NAME}}` 패턴 없으면 사용자 작업물로 간주
  - 보존 메시지: "⏭️ 사용자 작업물 보존"
- **IMPROVED**: 출력 예시 업데이트
  - 템플릿 → 최신: "✅ (템플릿 → 최신 버전)"
  - 사용자 수정 → 보존: "⏭️ (사용자 작업물 보존)"
  - 새로 생성: "✨ (새로 생성)"
- **ADDED**: 최신 템플릿 참조 경로 안내
  - 사용자가 수동으로 새 필드 추가 가능하도록 가이드
- **PRINCIPLE**: SPEC 파일처럼 프로젝트 문서도 사용자 작업물 완전 보호
- **AUTHOR**: @alfred
- **RATIONALE**: 프로젝트 컨텍스트 손실 방지, 사용자 신뢰 강화

### v0.2.4 (2025-10-02) - Option C 하이브리드 완성
- **REFACTORED**: Phase 4를 Alfred가 Claude Code 도구로 직접 실행 (TypeScript 코드 제거)
- **REFACTORED**: Phase 5 검증을 Claude Code 도구로 전환 ([Glob], [Read], [Grep])
- **ADDED**: Phase 5.5 품질 검증 독립 섹션 (trust-checker 연동)
- **ADDED**: 카테고리별 복사 절차 상세화 (A-I: 10단계)
- **ADDED**: [Grep] "{{PROJECT_NAME}}" 기반 프로젝트 문서 보호
- **ADDED**: [Bash] chmod +x 훅 파일 권한 자동 부여
- **ADDED**: Output Styles 복사 (.claude/output-styles/alfred/)
- **ADDED**: 오류 복구 시나리오 4가지 (파일 복사 실패, 검증 실패, 버전 불일치, Write 실패)
- **REMOVED**: AlfredUpdateBridge TypeScript 클래스 언급 (문서-구현 일치)
- **PRINCIPLE**: 스크립트 최소화, 커맨드 지침 중심, Claude Code 도구 우선
- **AUTHOR**: @alfred, @cc-manager, @code-builder
- **SPEC**: SPEC-UPDATE-REFACTOR-001
- **REVIEW**: cc-manager 품질 점검 통과 (P0 6개, P1 3개 완료)

### v0.1.0 (2025-09-01) - Initial
- **INITIAL**: /alfred:9-update 명령어 최초 작성
- **AUTHOR**: @alfred

---

**버전 관리 정책**:
- MoAI-ADK 패키지 버전을 따름 (Semantic Versioning)
- `v0.x.y` → 개발 버전 (현재)
- `v1.0.0+` → 정식 릴리스 (프로덕션 준비)
- HISTORY 버전 = 해당 기능이 포함된 MoAI-ADK 릴리스 버전
- 날짜: 실제 릴리스 날짜 또는 TBD (예정)

## 커맨드 개요

MoAI-ADK npm 패키지를 최신 버전으로 업데이트하고, 템플릿 파일(`.claude`, `.moai`, `CLAUDE.md`)을 안전하게 갱신합니다. 자동 백업, 설정 파일 보존, 무결성 검증을 지원합니다.

## 실행 흐름

1. **버전 확인** - 현재/최신 버전 비교
2. **백업 생성** - 타임스탬프 기반 자동 백업
3. **패키지 업데이트** - npm install moai-adk@latest
4. **템플릿 복사** - Claude Code 도구 기반 안전한 파일 복사
5. **검증** - 파일 존재 및 내용 무결성 확인

## Alfred 오케스트레이션

**실행 방식**: Alfred가 직접 실행 (전문 에이전트 위임 없음)
**예외 처리**: 오류 발생 시 `debug-helper` 자동 호출
**품질 검증**: 선택적으로 `trust-checker` 연동 가능 (--check-quality 옵션)

## 사용법

```bash
/alfred:9-update                    # 업데이트 확인 및 실행
/alfred:9-update --check            # 업데이트 가능 여부만 확인
/alfred:9-update --force            # 강제 업데이트 (백업 없이)
/alfred:9-update --check-quality    # 업데이트 후 TRUST 검증 수행
```

## 실행 절차

### Phase 1: 버전 확인 및 검증

```bash
npm list moai-adk --depth=0   # 현재 버전
npm view moai-adk version     # 최신 버전
```

**조건부 실행**: `--check` 옵션이면 여기서 중단하고 결과만 보고

### Phase 2: 백업 생성

```bash
BACKUP_DIR=".moai-backup/$(date +%Y-%m-%d-%H-%M-%S)"
mkdir -p "$BACKUP_DIR"
cp -r .claude .moai CLAUDE.md "$BACKUP_DIR/" 2>/dev/null || true
```

**백업 구조**: `.moai-backup/YYYY-MM-DD-HH-mm-ss/{.claude, .moai, CLAUDE.md}`

**예외**: `--force` 옵션이면 건너뛰기

### Phase 3: npm 패키지 업데이트

```bash
if [ -f "package.json" ]; then
    npm install moai-adk@latest
else
    npm install -g moai-adk@latest
fi
```

### Phase 4: Alfred가 Claude Code 도구로 템플릿 복사/병합

**담당**: Alfred (직접 실행, 에이전트 위임 없음)
**도구**: [Bash], [Glob], [Read], [Grep], [Write]

**처리 방식**:
- Step 2-6: 시스템 파일 (명령어, 에이전트, 훅, 스타일, 가이드) → **전체 교체** ✅
- Step 7-9: 프로젝트 문서 (product, structure, tech) → **사용자 수정 시 보존** 🔒
- Step 10: CLAUDE.md → **지능형 병합** 🔄
- Step 11: config.json → **스마트 딥 병합** 🔄

**실행 절차**:

#### Step 1: npm root 확인

```bash
[Bash] npm root
→ Output: /Users/user/project/node_modules
```

템플릿 경로 설정:
```
TEMPLATE_ROOT="{npm_root}/moai-adk/templates"
```

**보존/병합 대상**:
- `.moai/specs/` - 모든 SPEC 파일 (완전 보존) 🔒
- `.moai/reports/` - 동기화 리포트 (완전 보존) 🔒
- `.moai/config.json` - 프로젝트 설정 (스마트 병합) 🔄
- `.moai/project/*.md` - 프로젝트 문서 (사용자 수정 시 보존) 🔒
- `CLAUDE.md` - 프로젝트 지침 (지능형 병합) 🔄

---

#### Step 2: 명령어 파일 복사 (카테고리 A)

**대상**: `.claude/commands/alfred/*.md` (~10개 파일)

```text
[Step 2.1] 템플릿 파일 검색
  → [Glob] "{npm_root}/moai-adk/templates/.claude/commands/alfred/*.md"
  → 결과: [1-spec.md, 2-build.md, 3-sync.md, 8-project.md, 9-update.md, ...]

[Step 2.2] 각 파일 복사
  FOR EACH file IN glob_results:
    a. [Read] "{npm_root}/moai-adk/templates/.claude/commands/alfred/{file}"
    b. [Write] ".claude/commands/alfred/{file}"
    c. 성공 로그: "✅ {file}"

[Step 2.3] 완료 메시지
  → "✅ .claude/commands/alfred/ (~10개 파일 복사 완료)"
```

**오류 처리**:
- Glob 결과 비어있음 → "⚠️ 템플릿 디렉토리 경로 확인 필요"
- Write 실패 → `[Bash] mkdir -p .claude/commands/alfred` 후 재시도

---

#### Step 3: 에이전트 파일 복사 (카테고리 B)

**대상**: `.claude/agents/alfred/*.md` (~9개 파일)

```text
[Step 3.1-3.3] 카테고리 A와 동일 절차
  → 경로만 변경: .claude/agents/alfred/
  → 예상 파일: spec-builder.md, code-builder.md, doc-syncer.md, ...
```

---

#### Step 4: 훅 파일 복사 + 권한 부여 (카테고리 C)

**대상**: `.claude/hooks/alfred/*.cjs` (~4개 파일)

```text
[Step 4.1] 템플릿 파일 검색
  → [Glob] "{npm_root}/moai-adk/templates/.claude/hooks/alfred/*.cjs"
  → 예상: [policy-block.cjs, pre-write-guard.cjs, session-notice.cjs, tag-enforcer.cjs]

[Step 4.2] 각 파일 복사
  FOR EACH file IN glob_results:
    a. [Read] "{npm_root}/moai-adk/templates/.claude/hooks/alfred/{file}"
    b. [Write] ".claude/hooks/alfred/{file}"

[Step 4.3] 실행 권한 부여
  → [Bash] chmod +x .claude/hooks/alfred/*.cjs
  → IF 성공: "✅ 실행 권한 부여 완료 (755)"
  → IF 실패: "⚠️ chmod 실패 (Windows 환경에서는 정상, 계속 진행)"

[Step 4.4] 완료 메시지
  → "✅ .claude/hooks/alfred/ (4개 파일 복사 + 권한 설정 완료)"
```

**권한 설정**:
- Unix 계열: `755` (rwxr-xr-x)
- Windows: chmod 실패해도 경고만 출력

---

#### Step 5: Output Styles 복사 (카테고리 D) ✨

**대상**: `.claude/output-styles/alfred/*.md` (4개 파일)

```text
[Step 5.1] 템플릿 파일 검색
  → [Glob] "{npm_root}/moai-adk/templates/.claude/output-styles/alfred/*.md"
  → 예상: [beginner-learning.md, pair-collab.md, study-deep.md, moai-pro.md]

[Step 5.2] 각 파일 복사
  FOR EACH file IN glob_results:
    a. [Read] "{npm_root}/moai-adk/templates/.claude/output-styles/alfred/{file}"
    b. [Write] ".claude/output-styles/alfred/{file}"

[Step 5.3] 완료 메시지
  → "✅ .claude/output-styles/alfred/ (4개 파일 복사 완료)"
```

---

#### Step 6: 개발 가이드 복사 (카테고리 E)

**대상**: `.moai/memory/development-guide.md` (무조건 덮어쓰기)

```text
[Step 6.1] 파일 읽기
  → [Read] "{npm_root}/moai-adk/templates/.moai/memory/development-guide.md"

[Step 6.2] 파일 쓰기
  → [Write] ".moai/memory/development-guide.md"
  → IF 실패: [Bash] mkdir -p .moai/memory 후 재시도

[Step 6.3] 완료 메시지
  → "✅ .moai/memory/development-guide.md 업데이트 완료"
```

**참고**: development-guide.md는 항상 최신 템플릿으로 덮어써야 함 (사용자 수정 금지)

---

#### Step 7-9: 프로젝트 문서 복사 (카테고리 F-H) - 사용자 작업물 보존

**대상**:
- `.moai/project/product.md`
- `.moai/project/structure.md`
- `.moai/project/tech.md`

**각 파일마다 다음 절차 반복**:

```text
[Step 7.1] 기존 파일 존재 확인
  → [Read] ".moai/project/product.md"
  → IF 파일 없음: Step 7.5로 이동 (새로 생성)
  → IF 파일 있음: Step 7.2 진행

[Step 7.2] 템플릿 상태 검증
  → [Grep] "{{PROJECT_NAME}}" -n ".moai/project/product.md"
  → IF 검색 결과 있음: 템플릿 상태 (Step 7.5로 이동 - 덮어쓰기)
  → IF 검색 결과 없음: 사용자 수정 상태 (Step 7.3 진행 - 보존)

[Step 7.3] 사용자 작업물 보존 🔒
  → "ℹ️  product.md는 이미 프로젝트 정보가 작성되어 있어서 건너뜁니다"
  → "💡 최신 템플릿 참조: {npm_root}/moai-adk/templates/.moai/project/product.md"
  → "📝 필요시 수동으로 새 필드 추가 가능"
  → 복사하지 않고 다음 파일로 이동 (Step 7.6)

[Step 7.4] (예비 - 현재 미사용)

[Step 7.5] 새 템플릿 복사 (템플릿 상태 또는 파일 없음)
  → [Read] "{npm_root}/moai-adk/templates/.moai/project/product.md"
  → [Write] ".moai/project/product.md"
  → IF 실패: [Bash] mkdir -p .moai/project 후 재시도

[Step 7.6] 완료 메시지
  → 템플릿 상태였던 경우: "✅ .moai/project/product.md (템플릿 → 최신 버전)"
  → 사용자 수정 상태였던 경우: "⏭️  .moai/project/product.md (사용자 작업물 보존)"
  → 새로 생성한 경우: "✨ .moai/project/product.md (새로 생성)"
```

**보호 정책** (개선됨):
- `{{PROJECT_NAME}}` 패턴 존재 → 템플릿 상태, 안전하게 덮어쓰기
- 패턴 없음 → **사용자 수정, 보존 (덮어쓰지 않음)** 🔒
- 파일 없음 → 새로 생성

---

#### Step 10: CLAUDE.md 병합 (카테고리 I) - 지능형 병합

**대상**: `CLAUDE.md` (프로젝트 루트)

**병합 전략**: 템플릿 최신 구조 + 사용자 프로젝트 정보 유지

```text
[Step 10.1] 기존 파일 존재 확인
  → [Read] "./CLAUDE.md"
  → IF 파일 없음: Step 10.6으로 이동 (새로 생성)
  → IF 파일 있음: Step 10.2 진행

[Step 10.2] 템플릿 상태 검증
  → [Grep] "{{PROJECT_NAME}}" -n "./CLAUDE.md"
  → IF 검색 결과 있음: 템플릿 상태 (Step 10.6으로 이동 - 덮어쓰기)
  → IF 검색 결과 없음: 사용자 수정 상태 (Step 10.3 진행 - 병합)

[Step 10.3] 사용자 프로젝트 정보 추출
  → [Grep] "^- \*\*이름\*\*:" "./CLAUDE.md" → project_name
  → [Grep] "^- \*\*설명\*\*:" "./CLAUDE.md" → project_description
  → [Grep] "^- \*\*버전\*\*:" "./CLAUDE.md" → project_version
  → [Grep] "^- \*\*모드\*\*:" "./CLAUDE.md" → project_mode
  → "📋 추출된 정보: {project_name} v{project_version} ({project_mode})"

[Step 10.4] 최신 템플릿 읽기
  → [Read] "{npm_root}/moai-adk/templates/CLAUDE.md"
  → 템플릿 내용을 메모리에 저장

[Step 10.5] 템플릿에 사용자 정보 주입 (병합)
  → {{PROJECT_NAME}} → {project_name}
  → {{PROJECT_DESCRIPTION}} → {project_description}
  → {{PROJECT_VERSION}} → {project_version}
  → {{PROJECT_MODE}} → {project_mode}
  → [Write] "./CLAUDE.md" (병합된 내용)
  → "🔄 CLAUDE.md 병합 완료 (템플릿 최신화 + 프로젝트 정보 유지)"

[Step 10.6] 새 템플릿 복사 (템플릿 상태 또는 파일 없음)
  → [Read] "{npm_root}/moai-adk/templates/CLAUDE.md"
  → [Write] "./CLAUDE.md"

[Step 10.7] 완료 메시지
  → 템플릿 상태였던 경우: "✅ CLAUDE.md (템플릿 → 최신 버전)"
  → 병합한 경우: "🔄 CLAUDE.md (템플릿 최신화 + 프로젝트 정보 유지)"
  → 새로 생성한 경우: "✨ CLAUDE.md (새로 생성)"
```

**병합 정책**:
- 템플릿 상태 → 전체 교체
- 사용자 수정 → **지능형 병합** 🔄
  - 템플릿 최신 구조 사용
  - 프로젝트 정보(이름, 설명, 버전, 모드) 유지
  - Alfred 에이전트 목록, 워크플로우 등은 최신 템플릿 반영
- 파일 없음 → 새로 생성

---

#### Step 11: config.json 병합 (카테고리 J) - 스마트 병합

**대상**: `.moai/config.json` (프로젝트 설정)

**병합 전략**: 템플릿 최신 구조 + 사용자 설정값 유지

```text
[Step 11.1] 기존 파일 존재 확인
  → [Read] ".moai/config.json"
  → IF 파일 없음: Step 11.7로 이동 (새로 생성)
  → IF 파일 있음: Step 11.2 진행

[Step 11.2] 템플릿 상태 검증
  → [Grep] "{{PROJECT_NAME}}" -n ".moai/config.json"
  → IF 검색 결과 있음: 템플릿 상태 (Step 11.7로 이동 - 덮어쓰기)
  → IF 검색 결과 없음: 사용자 설정 상태 (Step 11.3 진행 - 병합)

[Step 11.3] 사용자 설정값 추출 (JSON 파싱)
  → project.name → user_project_name
  → project.description → user_project_description
  → project.version → user_project_version
  → project.mode → user_project_mode
  → project.created_at → user_created_at
  → constitution.test_coverage_target → user_test_coverage
  → constitution.simplicity_threshold → user_simplicity_threshold
  → git_strategy.* → user_git_strategy (전체 보존)
  → tags.categories → user_tag_categories (사용자 추가 카테고리)
  → pipeline.current_stage → user_current_stage
  → "📋 추출 완료: {user_project_name} v{user_project_version}"

[Step 11.4] 최신 템플릿 읽기
  → [Read] "{npm_root}/moai-adk/templates/.moai/config.json"
  → 템플릿 JSON을 메모리에 파싱

[Step 11.5] 딥 병합 (Deep Merge)
  → project.* → 사용자 값 우선
  → constitution.* → 사용자 수정값 유지, 새 필드는 템플릿 기본값
  → git_strategy.* → 사용자 설정 완전 유지
  → tags.categories → 템플릿 + 사용자 추가 (중복 제거)
  → pipeline.available_commands → 템플릿 최신 목록
  → pipeline.current_stage → 사용자 값 유지
  → _meta.* → 템플릿 최신 TAG 참조

[Step 11.6] 병합 결과 저장
  → [Write] ".moai/config.json" (병합된 JSON, 들여쓰기 2칸)
  → "🔄 config.json 병합 완료"

[Step 11.7] 새 템플릿 복사 (템플릿 상태 또는 파일 없음)
  → [Read] "{npm_root}/moai-adk/templates/.moai/config.json"
  → [Write] ".moai/config.json"

[Step 11.8] 완료 메시지
  → 템플릿 상태였던 경우: "✅ config.json (템플릿 → 최신 버전)"
  → 병합한 경우: "🔄 config.json (스마트 병합: 템플릿 구조 + 사용자 설정)"
  → 새로 생성한 경우: "✨ config.json (새로 생성)"
```

**병합 정책** (필드별):

| 필드 | 병합 전략 | 이유 |
|------|----------|------|
| `project.*` | 사용자 값 100% 유지 | 프로젝트 식별 정보 |
| `constitution.test_coverage_target` | 사용자 값 유지 | 팀 정책 |
| `constitution.simplicity_threshold` | 사용자 값 유지 | 팀 정책 |
| `git_strategy.*` | 사용자 값 100% 유지 | 워크플로우 설정 |
| `tags.categories` | 병합 (템플릿 + 사용자) | 확장 가능 |
| `pipeline.available_commands` | 템플릿 최신 | 시스템 명령어 |
| `pipeline.current_stage` | 사용자 값 유지 | 진행 상태 |
| `_meta.*` | 템플릿 최신 | TAG 참조 |

---

**전체 오류 처리 원칙**:
- 각 Step별 독립적 오류 처리
- 한 파일 실패가 전체 프로세스를 중단시키지 않음
- 실패한 파일 목록 수집하여 Phase 4 종료 후 보고
- 디렉토리 없음 → `mkdir -p` 자동 실행 후 재시도
- JSON 파싱 실패 → 백업 생성 후 템플릿으로 교체 (안전 모드)

### Phase 5: 업데이트 검증

**담당**: Alfred (직접 실행)
**도구**: [Bash], [Glob], [Read], [Grep]

**검증 항목**:

#### 5.1 파일 개수 검증 (동적)

```text
[Check 1] 명령어 파일
  → [Glob] .claude/commands/alfred/*.md
  → 실제 개수 확인
  → 예상: ~10개
  → IF 실제 < 예상: "⚠️ 명령어 파일 누락 감지"

[Check 2] 에이전트 파일
  → [Glob] .claude/agents/alfred/*.md
  → 예상: ~9개

[Check 3] 훅 파일
  → [Glob] .claude/hooks/alfred/*.cjs
  → 예상: ~4개

[Check 4] Output Styles 파일 ✨
  → [Glob] .claude/output-styles/alfred/*.md
  → 예상: 4개

[Check 5] 프로젝트 문서
  → [Glob] .moai/project/*.md
  → 예상: 3개

[Check 6] 필수 파일 존재
  → [Read] .moai/memory/development-guide.md
  → [Read] CLAUDE.md
  → IF 파일 없음: "❌ 필수 파일 누락"
```

**파일 개수 기준** (동적 확인):
- 템플릿에서 기대하는 파일 개수와 실제 복사된 파일 개수 비교
- 누락 감지 시 Phase 4 재실행 제안

---

#### 5.2 YAML Frontmatter 검증

```text
[Sample Check] 명령어 파일 검증
  → [Read] .claude/commands/alfred/1-spec.md
  → 첫 10줄 추출
  → YAML 파싱 시도 (---로 감싸진 블록)
  → IF 파싱 실패: "⚠️ YAML frontmatter 손상 감지"
  → IF 파싱 성공: "✅ YAML 검증 통과"

[필수 필드 확인]
  - name: alfred:1-spec
  - description: (내용 확인)
  - tools: [Read, Write, ...]
```

**검증 방식**:
- 대표 파일 1-2개 샘플링
- YAML 구문 오류 감지
- 손상 시 Phase 4 재실행 제안

---

#### 5.3 버전 정보 확인

```text
[Check 1] development-guide.md 버전
  → [Grep] "version:" -n .moai/memory/development-guide.md
  → 버전 추출: v{X.Y.Z}

[Check 2] package.json 버전
  → [Bash] npm list moai-adk --depth=0
  → 출력: moai-adk@{version}

[Check 3] 버전 일치 확인
  → IF 일치: "✅ 버전 정합성 통과 (v{X.Y.Z})"
  → IF 불일치: "⚠️ 버전 불일치 감지"
```

**버전 불일치 시**:
- Phase 3 재실행 제안 (npm 재설치)
- 또는 Phase 4 재실행 (템플릿 재복사)

---

#### 5.4 훅 파일 권한 검증 (Unix 계열만)

```bash
[Bash] ls -l .claude/hooks/alfred/*.cjs
→ 예상 출력: -rwxr-xr-x (755)
```

**권한 확인**:
- 실행 권한 (`x`) 존재 여부 확인
- Windows 환경은 검증 생략

---

**검증 실패 시 자동 복구 전략**:

| 오류 유형 | 복구 조치 |
|----------|----------|
| 파일 누락 | Phase 4 재실행 제안 |
| 버전 불일치 | Phase 3 재실행 제안 (npm) |
| 내용 손상 | 백업 복원 후 재시작 제안 |
| 권한 오류 | chmod 재실행 ([Bash] chmod +x) |
| 디렉토리 없음 | mkdir -p 후 Phase 4 재실행 |

### Phase 5.5: 품질 검증 (선택적)

**조건**: `--check-quality` 옵션 제공 시에만 실행

**도구**: trust-checker 에이전트

**실행 절차**:

```text
[Step 1] trust-checker 호출
  → @agent-trust-checker "Level 1 빠른 스캔 (3-5초)"
  → 검증 항목: TRUST 5원칙

[Step 2] 검증 항목
  - **T**est: 테스트 커버리지 ≥85% 확인
  - **R**eadable: ESLint/Biome 통과 여부
  - **U**nified: TypeScript 타입 안전성
  - **S**ecured: npm audit 보안 취약점 검사
  - **T**rackable: @TAG 체인 무결성 검증

[Step 3] 결과별 처리
  ✅ **Pass**: 업데이트 성공 완료
    → "✅ 품질 검증 통과"
    → "- 모든 파일 정상"
    → "- 시스템 무결성 유지"

  ⚠️ **Warning**: 경고 표시 후 완료
    → "⚠️ 품질 검증 경고"
    → "- 일부 문서 포맷 이슈 발견"
    → "- 권장사항 미적용 항목 존재"
    → 사용자 확인 권장 (계속 진행 가능)

  ❌ **Critical**: 롤백 제안
    → "❌ 품질 검증 실패 (치명적)"
    → "- 파일 손상 감지"
    → "- 설정 불일치"
    → 조치 선택:
      1. "롤백" → moai restore --from={timestamp}
      2. "무시하고 진행" → 손상된 상태로 완료 (위험)
    → 권장: 롤백 후 재시도
```

**실행 시간**: 추가 3-5초 (Level 1 빠른 스캔)

**검증 생략**:
- `--check-quality` 옵션 없으면 Phase 5.5 건너뛰고 완료

---

## 아키텍처: Alfred 중앙 오케스트레이션

```
┌─────────────────────────────────────────────────────────┐
│                   UpdateOrchestrator                     │
├─────────────────────────────────────────────────────────┤
│ Phase 1: VersionChecker   (자동 - Bash)                  │
│ Phase 2: BackupManager    (자동 - Bash)                  │
│ Phase 3: NpmUpdater       (자동 - Bash)                  │
│ Phase 4: ⏸️  ALFRED 직접 실행 (Claude Code 도구)        │
│          [Glob] [Read] [Grep] [Write] [Bash]            │
│ Phase 5: UpdateVerifier   (자동 - Alfred 직접)           │
│          [Glob] [Read] [Grep] [Bash]                    │
│ Phase 5.5: trust-checker  (선택 - @agent)               │
└─────────────────────────────────────────────────────────┘
```

**핵심 원칙**:
- **Phase 1-3**: UpdateOrchestrator 자동 실행 (Bash 스크립트)
- **Phase 4**: Alfred가 Claude Code 도구로 직접 제어 (문서 기반 지침)
- **Phase 5**: Alfred가 Claude Code 도구로 검증
- **Phase 5.5**: trust-checker 에이전트 선택적 호출
- **스크립트 최소화**: 모든 로직이 지침(텍스트)으로 표현됨
- **TypeScript 코드 없음**: Claude Code 도구만 사용

## 출력 예시

```text
🔍 MoAI-ADK 업데이트 확인 중...
📦 현재 버전: v0.0.1
⚡ 최신 버전: v0.0.2
✅ 업데이트 가능

💾 백업 생성 중...
   → .moai-backup/2025-10-02-15-30-00/

📦 패키지 업데이트 중...
   npm install moai-adk@latest
   ✅ 패키지 업데이트 완료

📄 Phase 4: Alfred가 템플릿 복사/병합 중...
   ✅ .claude/commands/alfred/ (~10개 파일 복사 완료)
   ✅ .claude/agents/alfred/ (~9개 파일 복사 완료)
   ✅ .claude/hooks/alfred/ (4개 파일 복사 + 권한 설정 완료)
   ✅ .claude/output-styles/alfred/ (4개 파일 복사 완료)
   ✅ .moai/memory/development-guide.md (무조건 업데이트)
   ✅ .moai/project/product.md (템플릿 → 최신 버전)
   ⏭️  .moai/project/structure.md (사용자 작업물 보존)
   ✨ .moai/project/tech.md (새로 생성)
   🔄 CLAUDE.md (템플릿 최신화 + 프로젝트 정보 유지)
   🔄 .moai/config.json (스마트 병합: 템플릿 구조 + 사용자 설정)
   ✅ 템플릿 파일 처리 완료

🔍 검증 중...
   [Bash] npm list moai-adk@0.0.2 ✅
   ✅ 검증 완료

✨ 업데이트 완료!

롤백이 필요하면: moai restore --from=2025-10-02-15-30-00
```

## 고급 옵션

### --check-quality (선택)

업데이트 후 TRUST 5원칙 품질 검증을 추가로 수행합니다.

```bash
/alfred:9-update --check-quality
```

**상세 내용**: Phase 5.5 섹션 참조
**실행 시간**: 추가 3-5초 (Level 1 빠른 스캔)
**검증 결과**: Pass / Warning / Critical

### --check (확인만)

업데이트 가능 여부만 확인하고 실제 업데이트는 수행하지 않습니다.

```bash
/alfred:9-update --check
```

### --force (강제 업데이트)

백업 생성 없이 강제 업데이트합니다. **주의**: 롤백 불가능합니다.

```bash
/alfred:9-update --force
```

## 안전 장치

**자동 백업**:
- `--force` 옵션 없으면 항상 백업 생성
- 백업 위치: `.moai-backup/YYYY-MM-DD-HH-mm-ss/`
- 수동 삭제 전까지 영구 보존

**데이터 보호 전략** (✨ v0.2.6):

1. **완전 보존** 🔒:
   - `.moai/specs/` - 사용자 SPEC 파일 절대 건드리지 않음
   - `.moai/reports/` - 동기화 리포트 보존
   - `.moai/project/*.md` - 프로젝트 문서 (사용자 수정 시)

2. **지능형 병합** 🔄:
   - **CLAUDE.md**: 템플릿 최신 구조 + 프로젝트 정보 유지
     - 프로젝트 이름, 설명, 버전, 모드 추출 → 템플릿 주입
     - Alfred 에이전트 목록, 워크플로우는 최신 템플릿 반영

   - **.moai/config.json**: 스마트 딥 병합
     - `project.*` → 사용자 값 100% 유지
     - `constitution.*` → 사용자 정책 유지, 새 필드는 템플릿 기본값
     - `git_strategy.*` → 사용자 워크플로우 완전 보존
     - `tags.categories` → 템플릿 + 사용자 추가 카테고리 병합
     - `pipeline.available_commands` → 템플릿 최신 명령어
     - `pipeline.current_stage` → 사용자 진행 상태 유지

3. **템플릿 판단 기준**:
   - `{{PROJECT_NAME}}` 패턴 존재 → 템플릿 상태 (전체 교체)
   - 패턴 없음 → 사용자 수정 (보존 또는 병합)

**롤백 지원**:
```bash
moai restore --list                       # 백업 목록
moai restore --from=2025-10-02-15-30-00  # 특정 백업 복원
moai restore --latest                     # 최근 백업 복원
```

## 오류 복구 시나리오

### 시나리오 1: 파일 복사 실패

**상황**:
```text
Phase 4 실행 중...
  → [Write] .claude/commands/alfred/1-spec.md ✅
  → [Write] .claude/commands/alfred/2-build.md ❌ (디스크 공간 부족)
```

**복구 절차**:
```text
[Step 1] 오류 로그 기록
  → "❌ 2-build.md 복사 실패: 디스크 공간 부족"

[Step 2] 실패 파일 목록에 추가
  → failed_files = [2-build.md]

[Step 3] 나머지 파일 계속 복사
  → Phase 4 중단 없이 진행
  → 모든 파일 처리 완료까지 계속

[Step 4] Phase 4 종료 후 실패 목록 보고
  → "⚠️ {count}개 파일 복사 실패: [2-build.md]"

[Step 5] 사용자 선택
  → "재시도" → Phase 4 재실행 (실패한 파일만)
  → "백업 복원" → moai restore --from={timestamp}
  → "무시" → Phase 5로 진행 (불완전한 상태, 권장하지 않음)
```

**자동 재시도**:
- 각 파일당 최대 2회 재시도
- 재시도 후에도 실패 시 건너뛰고 계속 진행

---

### 시나리오 2: 검증 실패 (파일 누락)

**상황**:
```text
Phase 5 검증 중...
  → [Glob] .claude/commands/alfred/*.md → 8개 (예상: 10개)
  → "❌ 검증 실패: 2개 파일 누락"
```

**복구 절차**:
```text
[Step 1] 누락 파일 파악
  → 템플릿과 실제 파일 목록 비교
  → 누락된 파일: [3-sync.md, 8-project.md]

[Step 2] 사용자에게 선택 제안
  → "Phase 4 재실행" → 전체 복사 다시 시도
  → "백업 복원" → moai restore --from={timestamp}
  → "무시하고 진행" → 불완전한 상태로 완료 (위험)

[Step 3] "Phase 4 재실행" 선택 시
  → Alfred Phase 4 절차 재실행
  → 완료 후 Phase 5 재검증
  → IF 재검증 통과: "✅ 검증 통과 (재시도 성공)"
  → IF 재검증 실패: 시나리오 2 반복 (최대 3회)

[Step 4] "백업 복원" 선택 시
  → [Bash] moai restore --from={timestamp}
  → "✅ 복원 완료, 재시도하시겠습니까?"
  → 재시도 선택 시 처음부터 다시 실행
```

---

### 시나리오 3: 버전 불일치

**상황**:
```text
Phase 5 검증 중...
  → [Grep] "version:" .moai/memory/development-guide.md → v0.0.1
  → [Bash] npm list moai-adk → v0.0.2
  → "❌ 버전 불일치 감지"
```

**복구 절차**:
```text
[Step 1] 사용자에게 보고
  → "⚠️ development-guide.md 버전(v0.0.1)과 패키지 버전(v0.0.2)이 불일치합니다."
  → "이는 템플릿 복사가 제대로 되지 않았음을 의미합니다."

[Step 2] 원인 분석 안내
  → 가능한 원인:
    a. npm 캐시 손상
    b. 템플릿 디렉토리 경로 오류
    c. 파일 복사 중 네트워크 오류

[Step 3] 선택 제안
  → "Phase 3 재실행" → npm 재설치 (npm cache clean + install)
  → "Phase 4 재실행" → 템플릿 재복사
  → "무시" → 버전 불일치 상태로 완료 (권장하지 않음)

[Step 4] 자동 복구 시도 (Phase 3 선택 시)
  → [Bash] npm cache clean --force
  → [Bash] npm install moai-adk@latest
  → Phase 4 재실행
  → Phase 5 재검증
```

---

### 시나리오 4: Write 도구 실패 (디렉토리 없음)

**상황**:
```text
[Write] .claude/commands/alfred/1-spec.md → ❌ (디렉토리 없음)
```

**자동 복구**:
```text
[Step 1] 오류 감지
  → "❌ Write 실패: .claude/commands/alfred/ 디렉토리 없음"

[Step 2] 디렉토리 자동 생성
  → [Bash] mkdir -p .claude/commands/alfred
  → "✅ 디렉토리 생성 완료"

[Step 3] Write 재시도
  → [Write] .claude/commands/alfred/1-spec.md
  → "✅ 파일 복사 성공 (재시도)"
```

**재시도 실패 시**:
- 디스크 공간 확인 안내
- 권한 문제 확인 안내
- 시나리오 1로 진행 (파일 복사 실패)

## 관련 명령어

- `/alfred:8-project` - 프로젝트 초기화/재설정
- `moai restore` - 백업 복원
- `moai doctor` - 시스템 진단
- `moai status` - 현재 상태 확인

## 버전 호환성

**MoAI-ADK Semantic Versioning** (`v0.x.y` 기준):

- **v0.2.x → v0.2.y**: 패치 업데이트 (완전 호환, 버그 수정/문서 개선)
- **v0.2.x → v0.3.x**: 마이너 업데이트 (신규 기능 추가, 설정 확인 권장)
- **v0.x.x → v1.x.x**: 메이저 업데이트 (Breaking Changes, 마이그레이션 가이드 필수)

**현재 버전**: v0.2.5
**다음 릴리스**: v0.2.6 (지능형 병합 시스템)
