# `/alfred:3-sync` 커맨드 분석 및 개선 리포트

**날짜**: 2025-11-12
**분석자**: R2-D2 (Claude Code)
**심각도**: 🔴 **높음 (High)**
**상태**: 분석 완료, 개선 권장

---

## 📋 Executive Summary

`/alfred:3-sync` 커맨드는 **Living Document 동기화만 수행**하고 있으며, **Infrastructure File Synchronization을 완전히 누락**하고 있습니다.

결과적으로 다음 파일들이 패키지 템플릿과 동기화되지 않음:
- 11개 에이전트 파일 (`.claude/agents/alfred/`)
- 2개 커맨드 파일 (`.claude/commands/alfred/`)
- 1개 설정 파일 (`.moai/config/config.json`)

---

## 🔍 상세 분석

### 1. 현재 `/alfred:3-sync` 구조 (v3.1.0)

```
PHASE 1: Analysis & Planning
├─ Step 1.1: Prerequisites & Skills 검증
├─ Step 1.2: Project Status 분석
├─ Step 1.3: Tag-Agent 호출 (TAG 검증)
├─ Step 1.4: Doc-Syncer 호출 (동기화 계획)
└─ Step 1.5: 사용자 승인

PHASE 2: Execute Synchronization
├─ Step 2.1: Safety Backup 생성
├─ Step 2.2: Doc-Syncer 호출 ← Living Documents만 처리
├─ Step 2.3: Quality-Gate 호출
└─ Step 2.4: SPEC Status 업데이트

PHASE 3: Git Operations & PR
├─ Step 3.1: Git-Manager 호출 (Commit)
├─ Step 3.2: PR Ready 전환
└─ Step 3.3: PR Auto-Merge (선택사항)

PHASE 4: Completion
├─ Step 4.1: 완료 리포트
└─ Step 4.2: 다음 단계 제시
```

### 2. Doc-Syncer 에이전트의 책임 범위

**✅ 수행 중인 작업**:
1. Living Document 동기화 (README.md, CHANGELOG.md 등)
2. API 문서 자동 생성/업데이트
3. SPEC 문서 동기화
4. @TAG 시스템 업데이트
5. 도메인별 문서 생성

**❌ 수행하지 않는 작업** (인프라 파일):
1. `.claude/agents/` 파일 동기화
2. `.claude/commands/` 파일 동기화
3. `.moai/config/` 파일 동기화
4. `.claude/hooks/` 파일 동기화
5. CLAUDE.md 템플릿 동기화

### 3. 누락된 파일 상세 목록

#### A. `.claude/agents/alfred/` — 11개 파일 불일치

| 파일 | 로컬 상태 | 템플릿 상태 | 심각도 |
|------|---------|-----------|--------|
| backend-expert.md | ❌ 구형 | ✅ 최신 | 높음 |
| cc-manager.md | ❌ 구형 | ✅ 최신 | 높음 |
| database-expert.md | ❌ 구형 | ✅ 최신 | 높음 |
| devops-expert.md | ❌ 구형 | ✅ 최신 | 높음 |
| doc-syncer.md | ❌ 구형 | ✅ 최신 | 높음 |
| format-expert.md | ❌ 구형 | ✅ 최신 | 높음 |
| frontend-expert.md | ❌ 구형 | ✅ 최신 | 높음 |
| project-manager.md | ❌ 구형 | ✅ 최신 | 높음 |
| spec-builder.md | ❌ 구형 | ✅ 최신 | 높음 |
| tdd-implementer.md | ❌ 구형 | ✅ 최신 | 높음 |
| ui-ux-expert.md | ❌ 구형 | ✅ 최신 | 높음 |

**영향**: Alfred가 호출하는 모든 에이전트들의 정의가 오래됨

#### B. `.claude/commands/alfred/` — 2개 파일 불일치

| 파일 | 로컬 상태 | 템플릿 상태 | 심각도 |
|------|---------|-----------|--------|
| 2-run.md | ❌ 구형 | ✅ 최신 | 중간 |
| 3-sync.md | ❌ 구형 | ✅ 최신 | 중간 |

**영향**: 워크플로우 정의가 오래됨, 사용자 지침이 부정확할 수 있음

#### C. `.moai/config/config.json` — 1개 파일 불일치

**영향**: 기본 설정이 패키지 템플릿과 다를 수 있음

---

## 🎯 근본 원인 분석

### 문제 1: 책임 범위의 명확하지 않은 정의

**CLAUDE.md의 명시적 규칙**:
```
항상 @src/moai_adk/templates/.claude/ @src/moai_adk/templates/.moai/
@src/moai_adk/templates/CLAUDE.md 에 변경이 생기면
로컬 프로젝트 폴더에도 동기화를 항상 하도록 하자.
패키지 템플릿이 가장 우선이다.
```

**하지만 `/alfred:3-sync` 구현**:
- ✅ Living Documents 동기화만 수행
- ❌ Infrastructure Files 동기화는 별도 단계로 없음

### 문제 2: 에이전트 책임의 모호함

**Doc-Syncer 에이전트 설명**:
```
"Use when: When automatic document synchronization
based on code changes is required."
```

이는 다음을 의미함:
- 코드 변경 기반 "문서" 동기화
- **인프라 파일(agents, commands, config)** 동기화는 아님

**해결**: 새로운 에이전트 또는 단계 필요

### 문제 3: Phase 2의 백업 범위 부족

**현재 Step 2.1 (Create Safety Backup)**:
```bash
Copy: README.md, docs/, .moai/specs/, .moai/indexes/
```

**누락**:
- `.claude/agents/`
- `.claude/commands/`
- `.moai/config/`

---

## 🔧 개선안

### 개선 1: `/alfred:3-sync` 구조 확장

**새로운 구조**:

```
PHASE 1: Analysis & Planning (현재 유지)
PHASE 2: Execute Synchronization
├─ Step 2.1: Safety Backup (확장)
├─ Step 2.2: Living Document Sync (현재 doc-syncer)
├─ Step 2.3: Infrastructure File Sync (신규)
├─ Step 2.4: Quality-Gate 검증 (현재)
└─ Step 2.5: SPEC Status 업데이트 (현재)

PHASE 3: Git Operations & PR (현재 유지)
PHASE 4: Completion (현재 유지)
```

### 개선 2: Infrastructure File Synchronization 추가

**새로운 Step 2.3 구현**:

```bash
# Step 2.3: Infrastructure File Synchronization

1. Agent Files Sync
   ├─ Copy: src/moai_adk/templates/.claude/agents/alfred/*
   │        → .claude/agents/alfred/
   ├─ Verify: 11개 파일 모두 최신 버전 확인
   └─ Report: 동기화된 파일 목록

2. Command Files Sync
   ├─ Copy: src/moai_adk/templates/.claude/commands/alfred/*
   │        → .claude/commands/alfred/
   ├─ Special: release-new.md는 로컬 유지 (로컬 전용)
   └─ Report: 동기화된 파일 목록

3. Config Files Sync
   ├─ Merge: src/moai_adk/templates/.moai/config/config.json
   │         (기본값) + 로컬 변경사항 유지
   ├─ Preserve: 로컬 customization 유지
   │           - alfred-orchestration.yaml (로컬 전용)
   │           - 사용자 정의 설정
   └─ Report: 병합 결과

4. Hooks Files Sync
   ├─ Update: .claude/hooks/alfred/* 최신 버전으로
   ├─ Special: SessionStart.md는 로컬 유지 (로컬 전용)
   └─ Report: 동기화된 파일 목록
```

### 개선 3: 로컬 전용 파일 보호

**로컬 전용 파일 목록** (동기화되지 않음):
```
✅ .claude/commands/alfred/release-new.md
✅ .moai/config/alfred-orchestration.yaml
✅ .claude/hooks/SessionStart.md
```

이들은 **로컬 개발 전용**이므로 패키지 템플릿으로 동기화되지 않아야 함.

### 개선 4: 향상된 백업 범위

**Step 2.1 개선**:
```bash
mkdir -p .moai-backups/sync-$TIMESTAMP/

Backup directories:
├─ README.md, CHANGELOG.md (문서)
├─ docs/, .moai/specs/ (레거시 항목)
├─ .moai/indexes/ (TAG 인덱스)
├─ .claude/agents/ (에이전트 정의) ← NEW
├─ .claude/commands/ (커맨드 정의) ← NEW
└─ .moai/config/ (설정 파일) ← NEW
```

---

## 📊 영향도 분석

### 영향받는 사용자 시나리오

**시나리오 1: 새 프로젝트 구성 (직접적 영향)**
```
1. /alfred:0-project (프로젝트 초기화)
2. /alfred:1-plan (SPEC 작성)
3. /alfred:2-run (구현)
4. /alfred:3-sync (동기화) ← 현재: 에이전트 파일 미동기화
   결과: 이전 버전의 에이전트로 다음 사이클 실행 위험
```

**시나리오 2: 패키지 업데이트 후 (간접적 영향)**
```
1. moai-adk 패키지 업데이트 (새 에이전트 버전)
2. moai-adk sync (로컬 템플릿 동기화)
3. /alfred:3-sync ← 여전히 구형 에이전트 파일 유지
   결과: 패키지 업데이트 효과 미반영
```

**시나리오 3: 다중 팀 협업 (높은 위험)**
```
1. 팀 멤버 A: 새 버전의 에이전트로 작업
2. 팀 멤버 B: /alfred:3-sync로 동기화 (구형 에이전트 사용)
3. 결과: 팀원 간 일관성 부족, 버그 증가
```

---

## 🚀 구현 계획

### Phase 1: 커맨드 파일 업데이트

**파일**: `.claude/commands/alfred/3-sync.md`

**추가 내용**:
```markdown
## 🔧 PHASE 2.3: Infrastructure File Synchronization (NEW)

**Goal**: 패키지 템플릿의 인프라 파일들을 로컬로 동기화

### 동기화 범위

1. **에이전트 파일** (.claude/agents/alfred/)
   - 11개 파일: backend-expert, cc-manager, database-expert, ...
   - 패키지 템플릿에서 복사
   - 로컬 커스터마이징 없음 (읽기 전용 에이전트 정의)

2. **커맨드 파일** (.claude/commands/alfred/)
   - 2-run.md, 3-sync.md 최신화
   - release-new.md는 로컬 전용 (보존)

3. **설정 파일** (.moai/config/)
   - config.json 기본값 병합 (사용자 정의 유지)
   - alfred-orchestration.yaml는 로컬 전용 (보존)

4. **Hook 파일** (.claude/hooks/)
   - 최신 버전으로 동기화
   - SessionStart.md는 로컬 전용 (보존)
```

### Phase 2: 동기화 로직 구현

**bash 스크립트**:
```bash
#!/bin/bash

# Infrastructure File Synchronization

TEMPLATE_DIR="src/moai_adk/templates"
LOCAL_DIR="."
TIMESTAMP=$(date +%Y-%m-%d-%H%M%S)

echo "🔄 Infrastructure File Synchronization..."

# 1. Agent Files
echo "📌 Syncing agent files..."
cp -r $TEMPLATE_DIR/.claude/agents/alfred/* .claude/agents/alfred/
echo "✅ Agent files synced (11 files)"

# 2. Command Files
echo "📌 Syncing command files..."
cp $TEMPLATE_DIR/.claude/commands/alfred/2-run.md .claude/commands/alfred/
cp $TEMPLATE_DIR/.claude/commands/alfred/3-sync.md .claude/commands/alfred/
echo "✅ Command files synced (2 files)"

# 3. Config Files (with merge)
echo "📌 Merging config files..."
# Template의 기본값과 로컬 커스터마이징 병합
python3 << 'PYTHON'
import json

# 템플릿 로드
with open("src/moai_adk/templates/.moai/config/config.json") as f:
    template = json.load(f)

# 로컬 로드
with open(".moai/config/config.json") as f:
    local = json.load(f)

# 병합 (로컬 값 우선)
merged = {**template, **local}

# 저장
with open(".moai/config/config.json", "w") as f:
    json.dump(merged, f, indent=2)

print("✅ Config files merged")
PYTHON

# 4. Report
echo ""
echo "📊 Infrastructure Synchronization Complete"
```

---

## ✅ 검증 체크리스트

- [ ] Step 2.3 추가 (Infrastructure File Synchronization)
- [ ] 11개 에이전트 파일 동기화 로직 구현
- [ ] 2개 커맨드 파일 동기화 로직 구현
- [ ] config.json 병합 로직 구현
- [ ] 로컬 전용 파일 보호 로직 검증
- [ ] 백업 범위 확장
- [ ] doc-syncer 에이전트 호출 유지
- [ ] git-manager 호출 유지
- [ ] 동기화 리포트에 인프라 파일 포함

---

## 🔗 관련 CLAUDE.md 규칙

```
항상 @src/moai_adk/templates/.claude/ @src/moai_adk/templates/.moai/
@src/moai_adk/templates/CLAUDE.md 에 변경이 생기면
로컬 프로젝트 폴더에도 동기화를 항상 하도록 하자.
패키지 템플릿이 가장 우선이다.
```

이 규칙이 `/alfred:3-sync` 커맨드에 명시적으로 구현되지 않음.

---

## 📝 결론

**현재 상태**: `/alfred:3-sync` 커맨드는 Living Document 동기화만 수행

**개선 필요사항**: Infrastructure File Synchronization 단계 추가

**권장 우선순위**: 🔴 **높음 (High)**

**예상 작업량**: 2-3시간 (구현 + 테스트)

**예상 효과**: 패키지 템플릿과 로컬 프로젝트 간 100% 동기화 달성

---

**작성자**: R2-D2 (Claude Code)
**날짜**: 2025-11-12
**상태**: 분석 완료, 구현 준비 중
