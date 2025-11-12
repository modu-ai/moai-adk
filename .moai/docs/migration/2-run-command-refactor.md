# Migration Guide: `/alfred:2-run` Refactoring (v0.23.0)

**Version**: 0.23.0 (Breaking Change)
**Date**: 2025-11-12
**Language**: 한국어

---

## 📋 개요

MoAI-ADK v0.22.5 → v0.23.0으로의 마이그레이션 가이드입니다.

**핵심 변경사항**: `/alfred:2-run` 명령어를 Claude Code 공식 베스트 프랙티스에 따라 **완전 에이전트 위임 구조**로 리팩토링했습니다.

---

## ⚠️ Breaking Changes

### 1. 명령어 구조 변경

**Before (v0.22.5)**:
```yaml
/alfred:2-run SPEC-001
  ├─ Read SPEC file (직접 실행) ❌
  ├─ Bash script 호출 (직접 실행) ❌
  ├─ Display plan (직접 실행) ❌
  ├─ Task(implementation-planner) ✓
  ├─ Task(tdd-implementer) ✓
  ├─ Task(quality-gate) ✓
  ├─ Task(git-manager) ✓
  └─ git log (직접 실행) ❌
```

**After (v0.23.0)**:
```yaml
/alfred:2-run SPEC-001
  └─ Task(run-orchestrator) ✓
      ├─ Phase 1: Analysis & Planning
      ├─ Phase 2: TDD Implementation
      ├─ Phase 3: Git Operations
      └─ Phase 4: Completion
```

### 2. allowed-tools 변경

**Before**:
```yaml
allowed-tools:
  - Read
  - Write
  - Edit
  - MultiEdit
  - Bash(python3:*)
  - Bash(pytest:*)
  - Bash(npm:*)
  - Bash(node:*)
  - Bash(git:*)
  - Task
  - WebFetch
  - Grep
  - Glob
  - TodoWrite
```

**After**:
```yaml
allowed-tools:
  - Task
```

### 3. 에이전트 추가

**새로운 에이전트**:
- `run-orchestrator` - 4-Phase 오케스트레이션 담당

**기존 에이전트** (변경 없음):
- implementation-planner
- tdd-implementer
- quality-gate
- git-manager

---

## ✅ 사용자 관점에서의 변경사항

### 사용 방법 (No Change ✓)
```bash
# 명령어 사용법 동일
/alfred:2-run SPEC-001
/alfred:2-run SPEC-FRONTEND-001
```

### 실행 흐름 (Simplified)

**Before**: 명령어가 직접 여러 작업 수행
- 평균 실행 로직 복잡도: High
- 명령어 코드: ~420줄

**After**: 명령어는 오케스트레이션만, 실행은 에이전트
- 평균 실행 로직 복잡도: Low
- 명령어 코드: ~260줄 (38% 감소)

### 출력 (No Change ✓)
- 실행 계획 출력: 동일
- 사용자 승인 요청: 동일
- 완료 요약: 동일

---

## 🚀 마이그레이션 체크리스트

### 1단계: 업그레이드 전 확인

```bash
# 현재 버전 확인
cat .moai/config/config.json | jq '.version'
# Expected: 0.22.5

# 미커밋 변경사항 확인
git status
# Expected: Clean working directory

# 현재 feature 브랜치가 있으면 기록
git branch | grep feature/
```

### 2단계: 에이전트 설치

새로운 에이전트 파일이 다음 위치에 있는지 확인:

```bash
# 로컬
test -f .claude/agents/run-orchestrator.md && echo "✓ Agent installed" || echo "✗ Missing"

# 패키지 템플릿
test -f src/moai_adk/templates/.claude/agents/run-orchestrator.md && echo "✓ Template updated" || echo "✗ Missing"
```

### 3단계: 스크립트 재배치

스크립트가 새로운 위치로 이동되었는지 확인:

```bash
# 새로운 위치
test -f .claude/skills/moai-alfred-workflow/scripts/spec_status_hooks.py && echo "✓ Relocated" || echo "✗ Missing"

# 이전 위치 (선택적 정리)
test -f .claude/hooks/alfred/spec_status_hooks.py && echo "⚠️  Old location still exists"
```

### 4단계: 명령어 업데이트 확인

```bash
# 로컬 명령어 확인
grep "allowed-tools:" .claude/commands/alfred/2-run.md
# Expected:
#   allowed-tools:
#     - Task

# 명령어가 Task() 호출만 하는지 확인
grep -E "Read|Write|Edit|Bash" .claude/commands/alfred/2-run.md | wc -l
# Expected: 0 (no matches)
```

### 5단계: 테스트

#### 5.1 기본 실행
```bash
# 테스트 SPEC 생성
mkdir -p .moai/specs/SPEC-MIGRATION-TEST/
cat > .moai/specs/SPEC-MIGRATION-TEST/spec.md << 'EOF'
# SPEC-MIGRATION-TEST: Migration Validation

## Requirements
- Test refactored /alfred:2-run command
- Verify all 4 phases work

## Acceptance Criteria
1. Plan created
2. Implementation passes quality gate
3. Commits created
4. Completion shown
EOF

# 실행
/alfred:2-run SPEC-MIGRATION-TEST
```

#### 5.2 기존 SPEC 재실행 (선택적)
```bash
# 기존 완료된 SPEC 다시 실행해보기
/alfred:2-run SPEC-EXISTING-001

# 결과 비교
git log --oneline -10
```

### 6단계: 버전 업그레이드

```bash
# 버전 업그레이드
# .moai/config/config.json에서:
# "version": "0.22.5" → "version": "0.23.0"

# 또는 명령어로:
# /alfred:0-project setting (version 선택)
```

---

## 🎯 에이전트/개발자 관점

### 새로운 에이전트: run-orchestrator

**역할**: 4-Phase 오케스트레이션 완전 담당

**책임**:
1. **Phase 1**: SPEC 분석 및 계획 생성
   - implementation-planner 호출
   - 사용자 승인 처리

2. **Phase 2**: TDD 실행
   - tdd-implementer 호출
   - quality-gate 호출
   - 결과 처리

3. **Phase 3**: Git 작업
   - git-manager 호출
   - 커밋 검증

4. **Phase 4**: 완료 처리
   - 요약 표시
   - 다음 단계 안내

**도구**:
- Task: 전문 에이전트 호출
- AskUserQuestion: 사용자 상호작용
- TodoWrite: 작업 추적
- Read: 설정 파일 읽기

**스킬**:
- moai-alfred-workflow
- moai-alfred-todowrite-pattern
- moai-alfred-ask-user-questions
- moai-alfred-reporting

### 기존 에이전트 변경 사항

**implementation-planner** (변경 없음)
- 권한: Read, Grep, Glob, WebFetch 등
- 역할: SPEC 분석 및 전략 수립
- 위임자: run-orchestrator

**tdd-implementer** (변경 없음)
- 역할: TDD 사이클 실행
- 위임자: run-orchestrator

**quality-gate** (변경 없음)
- 역할: TRUST 5 검증
- 위임자: run-orchestrator

**git-manager** (변경 없음)
- 역할: Git 커밋 관리
- 위임자: run-orchestrator

---

## 📊 영향도 분석

### 긍정적 영향

✅ **코드 복잡도 감소**
- 명령어 라인: 420줄 → 260줄 (38% 감소)
- 직접 도구 사용: 제거됨
- 에이전트 호출: 1개로 단순화

✅ **유지보수성 향상**
- 명령어 수정 불필요 (에이전트만 관리)
- 책임 분리 명확
- 테스트 용이성 증가

✅ **Claude Code 베스트 프랙티스 준수**
- Commands → Task() → Agents → Skills
- 3-계층 아키텍처 명확화
- 도구 권한 최소화

✅ **확장성 개선**
- Phase별 에이전트 추가 용이
- 새 기능 통합 간단
- 다른 워크플로우에 패턴 재사용 가능

### 잠재적 영향

⚠️ **에이전트 의존성**
- run-orchestrator가 작동해야 함
- 에이전트 파일 필수

⚠️ **학습 곡선**
- 새로운 에이전트 구조 이해 필요
- 기존 패턴과 다름

### 없는 영향 (호환성 유지)

✓ 사용자 인터페이스 동일
✓ 명령어 실행 방법 동일
✓ 출력 형식 일관성 유지
✓ 기능 동일

---

## 🔄 롤백 절차 (필요시)

### 문제 발생 시

```bash
# 1. 현재 브랜치 저장
git stash

# 2. 이전 버전으로 체크아웃
git checkout v0.22.5

# 3. 명령어 다시 실행
/alfred:2-run SPEC-XXX

# 4. Issue 보고
# GitHub issue 생성: 에러 내용, 재현 방법, 로그
```

### 수동 롤백

```bash
# 기존 2-run.md 복구
git show v0.22.5:.claude/commands/alfred/2-run.md > .claude/commands/alfred/2-run.md

# run-orchestrator 에이전트 제거
rm .claude/agents/run-orchestrator.md

# 버전 다운그레이드
# .moai/config/config.json: version → "0.22.5"
```

---

## 📝 로깅 및 디버깅

### 로그 위치

```bash
# 실행 로그
.moai/logs/session_*.log

# 에러 로그
.moai/logs/errors/

# 리포트
.moai/reports/
```

### 디버깅 팁

```bash
# 1. 에이전트 로드 확인
test -f .claude/agents/run-orchestrator.md && echo "✓ Agent found"

# 2. 명령어 구문 확인
cat .claude/commands/alfred/2-run.md | head -20

# 3. SPEC 파일 확인
cat .moai/specs/SPEC-001/spec.md

# 4. 최근 커밋 확인
git log --oneline -5
```

---

## 🆘 문제 해결

### Q1: "run-orchestrator agent not found" 에러

**원인**: 에이전트 파일 누락

**해결**:
```bash
# 에이전트 설치 확인
ls -la .claude/agents/run-orchestrator.md

# 없으면 패키지 재설치
moai-adk install --update
```

### Q2: `/alfred:2-run` 명령어가 작동하지 않음

**원인**: 명령어 파일 손상 또는 권한 문제

**해결**:
```bash
# 명령어 파일 확인
cat .claude/commands/alfred/2-run.md | wc -l

# 권한 확인
ls -la .claude/commands/alfred/2-run.md
```

### Q3: 커밋이 생성되지 않음

**원인**: git-manager 문제 또는 git 설정 누락

**해결**:
```bash
# git 설정 확인
git config user.name
git config user.email

# 설정되지 않았으면 설정
git config user.name "Your Name"
git config user.email "your@email.com"
```

### Q4: 품질 게이트 실패

**원인**: 테스트 커버리지 부족 또는 코드 스타일 문제

**해결**:
```bash
# 테스트 추가
pytest tests/ --cov

# 코드 포맷 수정
black src/
ruff check src/
```

---

## 📞 지원 및 피드백

### 문제 보고

```bash
# GitHub Issue 생성 (모범 사례)
제목: "[alfred:2-run v0.23.0] <문제 설명>"

본문:
- 재현 단계
- 예상 동작
- 실제 동작
- 환경 (MoAI-ADK 버전, Python 버전, OS 등)
- 로그 파일 (.moai/logs/)
```

### 피드백

- 메일: [project contact]
- GitHub Discussions: 아이디어 및 토론
- Issues: 버그 및 기능 요청

---

## 📚 참고 자료

- [Claude Code 공식 문서](https://docs.claude.com/claude-code)
- [MoAI-ADK CLAUDE.md](CLAUDE.md)
- [run-orchestrator 에이전트](.claude/agents/run-orchestrator.md)
- [테스트 계획](.moai/reports/PHASE-4-TEST-PLAN-2025-11-12.md)

---

**마이그레이션 완료 후**: `/clear` 명령어로 새 세션을 시작하는 것을 권장합니다.

**버전**: 0.23.0
**작성일**: 2025-11-12
**상태**: 검수 완료
