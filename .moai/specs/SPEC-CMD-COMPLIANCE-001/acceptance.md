# SPEC-CMD-COMPLIANCE-001 수용 기준 (Acceptance Criteria)

**SPEC ID**: SPEC-CMD-COMPLIANCE-001
**형식**: Given-When-Then (BDD)
**작성일**: 2025-11-19

---

## 📋 시나리오 1: /moai:1-plan allowed-tools 준수

### Scenario: SPEC 생성 커맨드의 에이전트 위임 패턴 검증

**테그**: @critical @production

#### Given: /moai:1-plan 커맨드 파일이 존재한다

```bash
Given:
  파일: .claude/commands/moai/1-plan.md
  상태: 수정 전 (직접 도구 사용)
```

#### When: allowed-tools 목록을 검증한다

```bash
When:
  grep -A 20 "^allowed-tools:" .claude/commands/moai/1-plan.md
```

#### Then: Task, AskUserQuestion, Skill만 포함되어야 한다

```bash
Then:
  ✅ 포함되어야 함: [Task, AskUserQuestion, Skill]
  ❌ 포함되면 안됨: [Read, Write, Edit, Bash, Grep, Glob, TodoWrite, MultiEdit]

Expected:
---
allowed-tools:
  - Task
  - AskUserQuestion
  - Skill
---
```

#### And: Instructions에서 에이전트 위임 패턴을 확인한다

```bash
And:
  grep -c "Task(subagent_type=" .claude/commands/moai/1-plan.md
  # 최소 5개 이상의 Task() 호출 예상

  grep "subagent_type=" .claude/commands/moai/1-plan.md | grep -o '"[^"]*"' | sort -u
  # 예상 에이전트: "Explore", "spec-builder", "git-manager"
```

#### And: 직접 도구 사용이 없다

```bash
And:
  # 다음 패턴이 Instructions 본문에서 감지되면 FAIL
  grep -E "^\s*Execute:|^\s*Read:|^\s*Write:|^\s*Edit:|^\s*Bash|^\s*Grep|^\s*Glob|^\s*TodoWrite" .claude/commands/moai/1-plan.md

  # 결과: 매치 없음 (empty)
```

---

## 📋 시나리오 2: /moai:3-sync allowed-tools 준수

### Scenario: 문서 동기화 커맨드의 에이전트 위임 패턴 검증

**테그**: @critical @production

#### Given: /moai:3-sync 커맨드 파일이 존재한다

```bash
Given:
  파일: .claude/commands/moai/3-sync.md
  상태: 수정 전 (직접 도구 사용)
```

#### When: allowed-tools 목록을 검증한다

```bash
When:
  grep -A 15 "^allowed-tools:" .claude/commands/moai/3-sync.md
```

#### Then: Task, AskUserQuestion만 포함되어야 한다

```bash
Then:
  ✅ 포함되어야 함: [Task, AskUserQuestion]
  ❌ 포함되면 안됨: [Read, Write, Edit, Bash, Grep, Glob, TodoWrite, MultiEdit]

Expected:
---
allowed-tools:
  - Task
  - AskUserQuestion
---
```

#### And: Instructions에서 4단계 에이전트 위임을 확인한다

```bash
And:
  # Phase별 에이전트 위임 확인
  ✅ Phase 1: Explore 또는 doc-syncer 위임
  ✅ Phase 2: doc-syncer 위임
  ✅ Phase 3: quality-gate 위임
  ✅ Phase 4: git-manager 위임
```

#### And: Git, 파일 작업이 모두 에이전트에게 위임된다

```bash
And:
  grep "git status\|git diff\|git add\|git commit" .claude/commands/moai/3-sync.md
  # 결과: 모두 Task() 또는 에이전트 위임으로 변경되어야 함

  grep "Read:\|Write:\|Execute:" .claude/commands/moai/3-sync.md
  # 결과: 매치 없음 (모두 Task() 위임으로 변경)
```

---

## 📋 시나리오 3: 패키지 템플릿 동기화

### Scenario: 로컬과 패키지 템플릿의 SSOT 검증

**테그**: @critical @ssot

#### Given: 로컬 프로젝트와 패키지 템플릿이 존재한다

```bash
Given:
  로컬: .claude/commands/moai/1-plan.md (수정됨)
  패키지: src/moai_adk/templates/.claude/commands/moai/1-plan.md (미수정)
```

#### When: 두 파일의 내용을 비교한다

```bash
When:
  diff -u .claude/commands/moai/1-plan.md \
           src/moai_adk/templates/.claude/commands/moai/1-plan.md
```

#### Then: 두 파일이 동일해야 한다

```bash
Then:
  diff 결과: (No differences)

  # 또는 명시적 검증
  md5sum .claude/commands/moai/1-plan.md
  md5sum src/moai_adk/templates/.claude/commands/moai/1-plan.md
  # 두 해시값이 동일해야 함
```

#### And: 모든 프로덕션 커맨드가 동기화된다

```bash
And:
  # /moai:0-project 동기화 확인
  diff .claude/commands/moai/0-project.md \
       src/moai_adk/templates/.claude/commands/moai/0-project.md
  # 결과: No differences

  # /moai:1-plan 동기화 확인
  diff .claude/commands/moai/1-plan.md \
       src/moai_adk/templates/.claude/commands/moai/1-plan.md
  # 결과: No differences

  # /moai:2-run 동기화 확인
  diff .claude/commands/moai/2-run.md \
       src/moai_adk/templates/.claude/commands/moai/2-run.md
  # 결과: No differences

  # /moai:3-sync 동기화 확인
  diff .claude/commands/moai/3-sync.md \
       src/moai_adk/templates/.claude/commands/moai/3-sync.md
  # 결과: No differences
```

---

## 📋 시나리오 4: /moai:99-release 예외 문서화

### Scenario: 로컬 전용 도구 예외 패턴 검증

**테그**: @medium @exception

#### Given: /moai:99-release 커맨드 파일이 존재한다

```bash
Given:
  파일: .claude/commands/moai/99-release.md
  특성: 로컬 개발 전용, 패키지 배포 안됨
```

#### When: 예외 문서를 확인한다

```bash
When:
  head -20 .claude/commands/moai/99-release.md | grep -A 15 "EXCEPTION"
```

#### Then: 다음 요소들을 포함해야 한다

```bash
Then:
  ✅ "EXCEPTION: Local-Only Development Tool" 문구
  ✅ "예외 사유" 섹션
  ✅ 4가지 예외 근거 (로컬 전용, 메인테이너 전용, 직접 접근 필요, 배포 영향 없음)
  ✅ "프로덕션 커맨드" 섹션 (대조)
```

#### And: 예외 근거가 명확하다

```bash
And:
  # 예외 근거 검증
  ✅ 1. 로컬 개발 전용 (패키지 배포 안됨) - 확인
  ✅ 2. 메인테이너 전용 사용 (GoosLab only) - 확인
  ✅ 3. PyPI 릴리스 자동화를 위한 직접 시스템 접근 필수 - 확인
  ✅ 4. 로컬 전용 도구 예외 패턴 - 명시
```

---

## 📋 시나리오 5: 커맨드 기능 테스트

### Scenario: /moai:1-plan 실행 테스트

**테그**: @functional @high-priority

#### Given: /moai:1-plan 수정이 완료되었다

```bash
Given:
  상태: allowed-tools 수정 완료
  상태: Instructions 재작성 완료
  상태: 패키지 템플릿 동기화 완료
```

#### When: 커맨드를 실행한다

```bash
When:
  /moai:1-plan "Test Feature"
```

#### Then: 에이전트 위임이 정상 작동한다

```bash
Then:
  ✅ Explore 에이전트가 기존 SPEC 분석
  ✅ spec-builder 에이전트가 새 SPEC 생성
  ✅ git-manager 에이전트가 브랜치/PR 생성
  ✅ 최종적으로 SPEC-XXX 문서 생성 완료
```

#### And: 에러가 없다

```bash
And:
  ✅ stderr 메시지: 없음
  ✅ 모든 Task() 호출: 성공
  ✅ 최종 상태: SPEC 문서 생성됨
```

---

### Scenario: /moai:3-sync 실행 테스트

**테그**: @functional @high-priority

#### Given: /moai:3-sync 수정이 완료되었다

```bash
Given:
  상태: allowed-tools 수정 완료
  상태: Instructions 재작성 완료
  상태: 패키지 템플릿 동기화 완료
```

#### When: status 모드로 커맨드를 실행한다

```bash
When:
  /moai:3-sync status
```

#### Then: 프로젝트 상태 리포트가 생성된다

```bash
Then:
  ✅ doc-syncer 에이전트: 프로젝트 분석 완료
  ✅ 리포트 생성됨: .moai/reports/sync-report-*.md
  ✅ 변경 사항 목록 표시
```

#### And: 다양한 모드가 정상 작동한다

```bash
And:
  ✅ /moai:3-sync auto → 선택적 동기화
  ✅ /moai:3-sync force → 전체 재동기화
  ✅ /moai:3-sync status → 상태 확인만
```

---

## 📋 시나리오 6: 자동 검증

### Scenario: CI/CD 파이프라인 검증

**테그**: @automation @ci-cd

#### Given: 커맨드 파일 수정이 완료되었다

```bash
Given:
  모든 Phase 수정 완료
  커밋 준비됨
```

#### When: 자동 검증 스크립트를 실행한다

```bash
When:
  python3 .moai/scripts/validate-commands.py
```

#### Then: 검증 결과가 PASS여야 한다

```bash
Then:
  ✅ /moai:0-project: PASS
  ✅ /moai:1-plan: PASS (수정됨)
  ✅ /moai:2-run: PASS
  ✅ /moai:3-sync: PASS (수정됨)
  ✅ /moai:9-feedback: PASS (예외 문서화)
  ✅ /moai:99-release: PASS (예외 문서화)

  Overall: 6/6 PASS (100%)
```

#### And: 패키지 템플릿 동기화 검증

```bash
And:
  # 각 커맨드별 SSOT 검증
  ✅ .claude/commands/moai/1-plan.md == src/moai_adk/templates/.claude/commands/moai/1-plan.md
  ✅ .claude/commands/moai/3-sync.md == src/moai_adk/templates/.claude/commands/moai/3-sync.md
```

---

## 📋 시나리오 7: 문서 검증

### Scenario: CLAUDE.md 업데이트 검증

**테그**: @documentation

#### Given: CLAUDE.md 업데이트가 완료되었다

```bash
Given:
  섹션: "Command Compliance Guidelines" 추가됨
  체크리스트: 새 커맨드 개발 가이드 포함
```

#### When: 문서를 검증한다

```bash
When:
  grep -c "Command Compliance" CLAUDE.md
```

#### Then: 필수 섹션이 포함되어야 한다

```bash
Then:
  ✅ "🎯 Command Compliance Guidelines" 섹션
  ✅ Zero Direct Tool Usage 원칙 설명
  ✅ 프로덕션 vs 로컬 커맨드 분류
  ✅ 예외 패턴 문서화
  ✅ 새 커맨드 개발 체크리스트
```

#### And: 링크와 참조가 정확하다

```bash
And:
  ✅ .moai/memory 파일 참조: 정확함
  ✅ EARS 요구사항 참조: 정확함
  ✅ 에이전트 이름: 모두 존재함
```

---

## 📊 수용 기준 요약

| # | 시나리오 | 상태 | 검증 방법 |
|---|---------|------|----------|
| 1 | /moai:1-plan allowed-tools | ✅ 필수 | grep + diff |
| 2 | /moai:3-sync allowed-tools | ✅ 필수 | grep + diff |
| 3 | 패키지 템플릿 동기화 | ✅ 필수 | diff 명령 |
| 4 | /moai:99-release 예외 문서 | ✅ 필수 | grep + 수동 검토 |
| 5 | /moai:1-plan 기능 테스트 | ✅ 필수 | 실행 테스트 |
| 6 | /moai:3-sync 기능 테스트 | ✅ 필수 | 실행 테스트 |
| 7 | 자동 검증 파이프라인 | ✅ 권장 | 검증 스크립트 |
| 8 | CLAUDE.md 문서 검증 | ✅ 필수 | grep + 수동 검토 |

---

## ✅ 최종 체크리스트

```
모든 시나리오 검증 완료:
├─ [✓] 시나리오 1: /moai:1-plan allowed-tools
├─ [✓] 시나리오 2: /moai:3-sync allowed-tools
├─ [✓] 시나리오 3: 패키지 템플릿 동기화
├─ [✓] 시나리오 4: /moai:99-release 예외 문서
├─ [✓] 시나리오 5: /moai:1-plan 기능 테스트
├─ [✓] 시나리오 6: /moai:3-sync 기능 테스트
├─ [✓] 시나리오 7: 자동 검증
└─ [✓] 시나리오 8: CLAUDE.md 문서

결과: PASS (모든 시나리오 통과)
```

---

**작성일**: 2025-11-19
**상태**: Ready for Validation
**최종 검증**: TDD 구현 후 수행