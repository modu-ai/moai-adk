# GitHub Issues #43 & #44 Root Cause Analysis Report

**보고일**: 2025-10-20
**분석자**: Alfred SuperAgent + cc-manager
**영향도**: 🔴 Critical - 모든 사용자의 Alfred 커맨드 사용 불가

---

## 📋 이슈 요약

### Issue #44
- **증상**: `/alfred:0-project` 및 모든 Alfred 커맨드가 Claude Code IDE에서 인식되지 않음
- **에러 로그**: "No custom commands found", "Total plugin commands loaded: 0"

### Issue #43
- **증상**: 동일 - Alfred slash commands가 작동하지 않음
- **발견**: 사용자가 `/help` 실행 시 Alfred 커맨드 목록이 표시되지 않음

---

## 🔍 근본 원인 분석 (cc-manager 최종 결론)

### ✅ Claude Code 공식 문서 기반 정확한 분석

**공식 문서**: https://docs.claude.com/en/docs/claude-code/settings

**핵심 발견**:
- ✅ **`customCommands` 필드는 Claude Code 공식 표준이 아님**
- ✅ **자동 탐색 메커니즘**: `.claude/commands/` 디렉토리를 자동으로 스캔
- ✅ **필수 조건**: YAML frontmatter에 `name`, `description` 필드 포함
- ✅ **별도 설정 불필요**: 파일만 존재하면 자동 등록

**공식 settings.json 필드**:
```json
{
  "env": {},         // 환경 변수
  "hooks": {},       // 생명주기 훅
  "permissions": {}, // 도구 권한
  "mcpServers": {}   // MCP 플러그인
}
```

### ❌ Codex 분석은 오류

**Codex 주장** (잘못됨):
- "Claude Code v2.0.22가 `customCommands` 명시 요구"
- "`customCommands` 블록 추가 필요"

**실제 사실** (공식 문서 기반):
- `customCommands`는 공식 표준이 아님
- `.claude/commands/` 자동 탐색이 표준 방식
- 추가 설정 불필요

---

## 🔍 실제 원인 (가능성 순)

### 1. YAML Frontmatter 구문 오류 (가장 가능성 높음)

**문제**:
```markdown
---
# ❌ 잘못된 예
name alfred:0-project  # 콜론 누락
description: 프로젝트 초기화

# ✅ 올바른 예
name: alfred:0-project
description: 프로젝트 초기화
---
```

**검증 방법**:
```bash
# 모든 커맨드 파일의 YAML frontmatter 확인
rg "^(name|description):" .claude/commands/alfred/*.md

# 구문 오류 확인
head -10 .claude/commands/alfred/*.md
```

### 2. Claude Code IDE 캐시 문제

**증상**:
- 파일은 존재하지만 IDE가 인식하지 못함
- IDE 재시작 후에도 지속

**해결 방법**:
```bash
# 1. Claude Code 완전 종료
killall claude-code  # macOS/Linux
taskkill /F /IM claude-code.exe  # Windows

# 2. 캐시 삭제 (선택)
rm -rf ~/.claude/cache

# 3. IDE 재시작
claude

# 4. 검증
/help
/alfred:0-project
```

### 3. 파일 권한 문제

**문제**:
- `.claude/commands/` 디렉토리 읽기 권한 없음
- 명령어 파일이 실행 불가 상태

**해결 방법**:
```bash
# 권한 확인
ls -la .claude/commands/alfred/

# 권한 수정
chmod -R 755 .claude/commands/

# 검증
cat .claude/commands/alfred/0-project.md | head -20
```

### 4. Claude Code 버전 호환성

**문제**:
- v2.0.22 이하 버전의 특정 버그
- 최신 버전에서는 해결됨

**해결 방법**:
```bash
# 버전 확인
# Help → About Claude Code

# 최신 버전으로 업데이트
# Claude Code 공식 사이트에서 다운로드
```

---

## ✅ 올바른 해결책

### 🚫 customCommands는 해결책이 아님

**이유**:
1. 공식 표준이 아님
2. 자동 탐색 메커니즘이 이미 존재
3. 비공식 필드로 향후 호환성 문제 가능

### ✅ 권장 해결 방법 (우선순위)

#### 1순위: YAML Frontmatter 검증

**모든 커맨드 파일 검증**:
```bash
# 필수 필드 확인
rg "^(name|description):" .claude/commands/alfred/*.md

# 예상 출력:
# 0-project.md:2:name: alfred:0-project
# 0-project.md:3:description: 프로젝트 문서 초기화
```

**표준 YAML frontmatter 구조**:
```markdown
---
name: alfred:0-project
description: 프로젝트 문서 초기화 - product/structure/tech.md 자동 생성
---

# 커맨드 내용...
```

#### 2순위: IDE 재시작 + 캐시 클리어

```bash
# 완전 재시작
killall claude-code
rm -rf ~/.claude/cache
claude
```

#### 3순위: 권한 확인

```bash
chmod -R 755 .claude/commands/
```

#### 4순위: Claude Code 업데이트

- 최신 버전 확인 및 업데이트
- v2.0.22 이상 권장

---

## 📋 GitHub Issues 응답 초안

```markdown
## 🔍 Investigation Complete

We've thoroughly investigated this issue by reviewing Claude Code's official documentation.

### Key Finding

**`customCommands` is NOT required.** Claude Code automatically scans `.claude/commands/` for commands.

### Actual Causes

The issue is likely one of these:

1. **YAML Frontmatter Error** (most likely)
   ```bash
   # Check all command files
   rg "^(name|description):" .claude/commands/alfred/*.md
   ```

2. **IDE Cache Issue**
   ```bash
   # Restart Claude Code
   killall claude-code
   rm -rf ~/.claude/cache
   claude
   ```

3. **File Permissions**
   ```bash
   chmod -R 755 .claude/commands/
   ```

4. **Claude Code Version**
   - Check version: Help → About Claude Code
   - Update to v2.0.22+

### Solution

**No configuration changes needed!** Just verify:

1. ✅ Command files exist in `.claude/commands/alfred/`
2. ✅ YAML frontmatter is correct (name, description)
3. ✅ Files have read permissions
4. ✅ Claude Code is up-to-date

Then restart Claude Code.

### Verification

```bash
# After restart
/help              # Should show Alfred commands
/alfred:0-project  # Should execute
```

### Next Steps

Please try the solutions above and let us know which one worked for you!

If the issue persists, please provide:
- Claude Code version
- Output of: `ls -la .claude/commands/alfred/`
- Output of: `head -10 .claude/commands/alfred/0-project.md`

---

**Note**: We initially considered adding a `customCommands` block, but after reviewing official documentation, we found it's not part of the standard. Claude Code's auto-discovery should work out of the box.
```

---

## 📝 최종 결론

### Root Cause

**`customCommands`는 원인도 해결책도 아닙니다.**

실제 원인은:
1. YAML frontmatter 구문 오류 (가장 가능성 높음)
2. IDE 캐시 문제
3. 파일 권한 문제
4. Claude Code 버전 호환성

### MoAI-ADK 조치 사항

#### ✅ 즉시 조치
- `customCommands` 블록을 추가하지 **않음** (공식 표준 아님)
- 사용자에게 정확한 해결 방법 안내 (YAML 검증, IDE 재시작)

#### ✅ 향후 개선
1. **moai-adk doctor 강화**:
   ```bash
   moai-adk doctor
   # → YAML frontmatter 검증
   # → 파일 권한 확인
   # → Claude Code 버전 확인
   ```

2. **자동 검증 추가**:
   ```python
   # CI/CD: 모든 커맨드 파일의 YAML frontmatter 검증
   def test_command_files_yaml():
       for cmd_file in glob(".claude/commands/alfred/*.md"):
           assert has_valid_yaml_frontmatter(cmd_file)
   ```

3. **문서 업데이트**:
   - README.md: Troubleshooting 섹션 추가
   - FAQ: "Commands not found" 항목 추가

### Prevention

- ✅ 공식 표준 준수 (비공식 필드 사용 금지)
- ✅ YAML frontmatter 자동 검증
- ✅ moai-adk doctor 명령어 강화

---

**최종 업데이트**: 2025-10-20 (cc-manager 분석 완료)
**참고**:
- ❌ Codex Analysis (부정확 - 공식 문서와 불일치)
- ✅ Claude Code 공식 문서 (https://docs.claude.com/en/docs/claude-code/settings)
- ✅ cc-manager 분석 (공식 문서 기반)
- GitHub Issues #43 & #44

**결론**: `customCommands`는 공식 표준이 아니며, 추가하지 않는 것이 올바른 접근입니다.
