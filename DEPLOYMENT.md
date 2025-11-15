# 배포 정책 및 Git 추적 가이드

## 개요

MoAI-ADK는 **패키지 템플릿** (source of truth)과 **로컬 개발 파일**의 이중 구조를 가지고 있습니다. 이 문서는 파일 분류, 배포 정책, 안전 메커니즘을 설명합니다.

## 파일 분류

### Category A: 배포 + 추적 (패키지 템플릿 - Source of Truth)

**위치**: `src/moai_adk/templates/`

**파일 목록**:
- `src/moai_adk/templates/.claude/` (Yoda 제외)
  - `agents/` - Alfred 에이전트 정의
  - `commands/` - 커스텀 명령어 (yoda 제외)
  - `skills/` - Skill 시스템 (moai-yoda-system 제외)
  - `hooks/` - Hook 정의
  - `settings.json` - Claude Code 설정
  - `.mcp.json` - MCP 서버 설정

- `src/moai_adk/templates/.moai/config/`
  - `config.json` - 프로젝트 메타데이터

- `src/moai_adk/templates/.moai/scripts/`
  - `statusline.py` - 상태 표시줄 스크립트

- `src/moai_adk/templates/CLAUDE.md` - 프로젝트 지침 템플릿

**배포 방식**:
- 패키지에 포함됨: `uv build` 및 `pip install moai-adk`
- 변수 형식 유지: `{{PROJECT_DIR}}`, `{{PROJECT_OWNER}}` 등

### Category B: 추적 + 미배포 (로컬 개발 파일)

**위치**: 프로젝트 루트

**파일 목록**:
- `.claude/` - 로컬 개발용 사본 (패키지 템플릿에서 동기화)
  - 변수 치환됨: `{{PROJECT_DIR}}` → 상대 경로 `./`
  - 로컬 개발용으로만 사용 (패키지에 포함 안 됨)

- `.moai/config/` - 로컬 프로젝트 설정
  - 치환된 값: `project.owner`, `moai.version` 등

- `.moai/scripts/statusline.py` - 패키지 템플릿에서 동기화

- `CLAUDE.md` - 로컬 프로젝트 지침
  - 로컬 고유 정보 포함 (프로젝트 이름, 소유자 등)

- `src/moai_adk/core/merge/analyzer.py` - MergeAnalyzer 구현
- `tests/test_merge_analyzer.py` - 단위 테스트 (18개)
- `tests/test_merge_analyzer_integration.py` - 통합 테스트 (6개)

**배포 방식**:
- 패키지에 포함 안 됨: `.gitignore`로 제외
- Git에는 추적됨: 패키지 템플릿 동기화용

### Category C: 무시 + 미배포 (런타임 파일)

**위치**: `.moai/` 디렉토리

**파일 목록**:
- `.moai/cache/` - 런타임 캐시
  - `git-info.json` - Git 정보 캐시
  - `version-check.json` - 버전 확인 결과

- `.moai/logs/` - 세션 로그
  - `session-*.log` - 명령어 실행 로그

- `.moai/memory/` - 런타임 상태
  - `command-execution.json` - 명령어 실행 상태
  - `subagent-*.log` - 서브에이전트 로그

- `.moai/docs/` - 로컬 문서 (배포 안 됨)

- `.moai/temp/` - 임시 파일

- `.moai/backups/` - 백업 파일

**배포 방식**:
- Git에 추적 안 함: `.gitignore`로 제외
- 삭제되어도 괜찮음: 재생성 가능한 파일

### Category D: 절대 배포 금지 (로컬 전용 도구)

**위치**: 로컬 전용

**파일 목록**:
- `.moai/yoda/` - 강의 자료 생성 도구
  - 로컬 개발자만 사용
  - 생성된 강의 자료는 `.moai/yoda/output/`에 저장

- `.moai/release/` - `/moai:release` 명령어 구현
  - 패키지 유지보수자만 사용
  - 버전 관리 및 릴리스 자동화

- `.claude/commands/moai/release.md` - 로컬 명령어 정의

**배포 방식**:
- 절대 포함 안 됨: `pyproject.toml` 및 `.gitignore`에서 제외
- 로컬에서만 사용: 개발자 환경에만 필요

## .gitignore 정책

### 추적할 파일 (✅ DO TRACK)

```gitignore
# 추적 항목 (배포됨):
src/moai_adk/templates/

# 추적 항목 (로컬 개발용):
.claude/                    # 로컬 에이전트, 명령어, Skill
.moai/config/              # 프로젝트 설정
.moai/scripts/             # 상태 스크립트, 안전 스크립트
CLAUDE.md                  # 프로젝트 지침
.mcp.json                  # (자격증명 검토 필수!)
```

### 무시할 파일 (❌ DO NOT TRACK)

```gitignore
# 런타임 파일:
.moai/cache/               # 버전 정보, Git 캐시
.moai/logs/                # 세션 로그
.moai/memory/              # 명령어 실행 상태
.moai/docs/                # 로컬 문서
.moai/temp/                # 임시 파일
.moai/backups/             # 백업 파일

# 로컬 전용 도구:
.moai/release/             # /moai:release 명령어
.moai/yoda/                # 강의 자료 생성 도구
.claude/commands/moai/release.md  # 로컬 명령어 정의
```

## 안전 메커니즘

### 1. Pre-Commit 검증

**목적**: 보호된 파일의 실수적 삭제 방지

**위치**: `.moai/scripts/verify-safe-files.py`

**실행**:
```bash
uv run .moai/scripts/verify-safe-files.py
```

**동작**:
- Git 스테이지에서 삭제된 파일 검사
- 보호된 경로 존재 여부 확인
- 위반 발견 시 에러 반환 (exit code 1)

**보호된 경로**:
```
.moai/yoda/
.moai/release/
.moai/docs/
.claude/commands/moai/
```

### 2. 안전 삭제 스크립트

**목적**: 의도적 삭제 시 자동 백업

**위치**: `.moai/scripts/backup-before-delete.sh`

**사용법**:
```bash
# 보호된 경로 확인
.moai/scripts/backup-before-delete.sh

# 파일 삭제 전 백업 (드라이런)
.moai/scripts/backup-before-delete.sh --dry-run .moai/docs/old-file.txt

# 실제 백업 수행
.moai/scripts/backup-before-delete.sh .moai/docs/old-file.txt
```

**백업 위치**:
```
.moai-backups/delete-backup_YYYYMMDD_HHMMSS/
```

### 3. MergeAnalyzer

**목적**: moai-adk 업데이트 시 백업 vs 템플릿 비교 분석

**위치**: `src/moai_adk/core/merge/analyzer.py`

**기능**:
- Claude Code 헤드리스 모드로 지능형 분석
- 충돌 위험도 평가
- 사용자 확인 후 병합
- 자동 폴백 (Claude 불가 시)

**사용**:
```python
from moai_adk.core.merge import MergeAnalyzer

analyzer = MergeAnalyzer(Path("."))
analysis = analyzer.analyze_merge(backup_path, template_path)

if analyzer.ask_user_confirmation(analysis):
    # 병합 진행
    pass
```

## 배포 프로세스

### 1. 패키지 빌드

```bash
uv build
```

**포함 사항**:
- ✅ `src/moai_adk/` (패키지 코드)
- ✅ `src/moai_adk/templates/` (템플릿 파일)
- ✅ `MANIFEST.in` (명시적 포함)

**제외 사항**:
- ❌ `.claude/` (로컬 개발용)
- ❌ `.moai/` (로컬 런타임)
- ❌ `CLAUDE.md` (로컬 프로젝트 지침)
- ❌ `.moai/yoda/` (로컬 전용 도구)
- ❌ `.moai/release/` (로컬 전용 도구)

### 2. 패키지 배포

```bash
pip install moai-adk
```

**사용자가 받는 파일**:
- `src/moai_adk/templates/.claude/` (변수 형식)
- `src/moai_adk/templates/.moai/` (변수 형식)
- `src/moai_adk/templates/CLAUDE.md` (변수 형식)

**사용자가 받지 않는 파일**:
- 로컬 개발 파일 (`.claude/`, `.moai/config/` 등)
- 로컬 전용 도구 (Yoda, `/moai:release`)
- 런타임 파일

### 3. 프로젝트 초기화

```bash
moai-adk init [project-name]
```

**처리 순서**:
1. 패키지 템플릿 복사: `src/moai_adk/templates/` → 프로젝트
2. 변수 치환: `{{PROJECT_DIR}}` → 상대 경로
3. 메타데이터 설정: `project.owner`, `moai.version` 등
4. `.gitignore` 적용

## 검증 체크리스트

### 패키지 빌드

- [ ] `uv build` 성공
- [ ] `dist/moai_adk-*.tar.gz` 생성됨
- [ ] 패키지에 `src/moai_adk/templates/` 포함
- [ ] 패키지에 `.claude/` 미포함
- [ ] 패키지에 `.moai/yoda/` 미포함

**검증 명령어**:
```bash
tar -tzf dist/moai_adk-*.tar.gz | grep -E "templates|\.claude|yoda" | head -20
```

### Git 추적

- [ ] 608개 파일 스테이징됨 (`.claude/`, `.moai/config/` 등)
- [ ] `.moai/cache/` 등 런타임 파일 제외됨
- [ ] `.moai/yoda/` 제외됨
- [ ] `.moai/release/` 제외됨

**검증 명령어**:
```bash
git status
git diff --cached --name-only | wc -l
```

### 안전 메커니즘

- [ ] `verify-safe-files.py` 스크립트 실행 가능
- [ ] `backup-before-delete.sh` 스크립트 실행 가능
- [ ] Pre-commit 검증 통과
- [ ] 보호된 파일 존재 확인

**검증 명령어**:
```bash
uv run .moai/scripts/verify-safe-files.py
.moai/scripts/backup-before-delete.sh
```

## 트러블슈팅

### 문제: 보호된 파일이 실수로 삭제됨

**해결책**:
```bash
# 백업에서 복구
cp -r .moai-backups/delete-backup_*/path/to/file .

# 또는 git 히스토리에서 복구
git checkout HEAD~1 -- path/to/file
git add path/to/file
git commit -m "Restore accidentally deleted protected file"
```

### 문제: 커밋이 보호된 파일 삭제로 차단됨

**해결책**:
1. 삭제가 의도적인지 확인
2. 필요하면 `.gitignore` 업데이트
3. 또는 Git 스테이지에서 제거:
   ```bash
   git reset HEAD -- path/to/file
   git checkout -- path/to/file
   ```

### 문제: 패키지에 로컬 파일이 포함됨

**해결책**:
1. `MANIFEST.in` 확인
2. `.gitignore` 확인
3. `pyproject.toml`에서 `packages` 설정 확인:
   ```toml
   [tool.poetry]
   packages = [
       { include = "moai_adk", from = "src" }
   ]
   exclude = [".moai/", ".claude/", ...]
   ```

## 참고

- **패키지 템플릿**: `src/moai_adk/templates/`가 source of truth
- **로컬 동기화**: 패키지 템플릿 → 로컬 프로젝트 (일방향)
- **보호 정책**: 로컬 전용 파일은 절대 삭제 금지
- **변수 관리**: 패키지는 변수 유지, 로컬은 치환

---

**마지막 업데이트**: 2025-11-14
**관련 파일**: `.gitignore`, `.moai/scripts/`, `CLAUDE.md`
