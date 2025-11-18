# 에이전트 동기화 실행 가이드

**생성**: 2025-11-19
**버전**: v0.26.0
**상태**: 실행 준비 완료

---

## 🎯 목표

로컬 에이전트 파일 31개를 패키지 템플릿과 완벽하게 동기화하여 SSOT(Single Source of Truth) 원칙 유지

---

## 📋 요약

| 항목 | 수치 |
|------|------|
| 총 에이전트 | 31개 |
| 동기화 필요 | 23개 (74%) |
| 최신 상태 | 8개 (26%) |
| 총 변경사항 | 192줄 |
| 예상 소요시간 | 95분 (~1.5시간) |
| 위험 수준 | 낮음 (완벽 롤백 가능) |

---

## 🚀 빠른 시작 (5분)

### 옵션 A: 자동화 스크립트 실행 (권장)

```bash
# 1. 스크립트 실행 권한 부여
chmod +x .moai/scripts/sync-agents.sh

# 2. 동기화 실행 (자동 백업 포함)
.moai/scripts/sync-agents.sh

# 3. 결과 확인
git diff .claude/agents/moai/ | head -50

# 4. 문제 발생 시 복원
cp -r .moai/backup/agents-sync-YYYY-MM-DD-HHMMSS/* .claude/agents/moai/
```

### 옵션 B: 수동 실행 (단계별)

```bash
# 1단계: 백업
cp -r .claude/agents/moai .moai/backup/agents-manual-backup-$(date +%s)

# 2단계: Phase 1 실행
for file in accessibility-expert.md api-designer.md backend-expert.md ...; do
  sed -i 's/moai-alfred-language-detection/moai-core-language-detection/g' ".claude/agents/moai/$file"
done

# 3단계: 검증
grep -r "moai-alfred" .claude/agents/moai/ | wc -l
# 결과: 0 (완벽 동기화)

# 4단계: 커밋
git add .claude/agents/moai/
git commit -m "chore(agents): Sync with v0.26.0 templates (moai-alfred → moai-core)"
```

---

## 📊 상세 동기화 계획

### Phase 1: 단순 변경 (13개 파일)

**시간**: 15분
**변경 유형**: `moai-alfred-language-detection` → `moai-core-language-detection`

**파일**:
```
1. accessibility-expert.md
2. api-designer.md
3. backend-expert.md
4. component-designer.md
5. devops-expert.md
6. figma-expert.md
7. frontend-expert.md
8. migration-expert.md
9. monitoring-expert.md
10. performance-engineer.md
11. ui-ux-expert.md
```

**자동화 명령어**:
```bash
cd .claude/agents/moai
for file in accessibility-expert.md api-designer.md backend-expert.md \
            component-designer.md devops-expert.md figma-expert.md \
            frontend-expert.md migration-expert.md monitoring-expert.md \
            performance-engineer.md ui-ux-expert.md; do
  sed -i 's/moai-alfred-language-detection/moai-core-language-detection/g' "$file"
  echo "✓ Updated: $file"
done
```

**검증**:
```bash
grep "moai-core-language-detection" .claude/agents/moai/accessibility-expert.md
# 결과: - `Skill("moai-core-language-detection")` – Detect project language
```

---

### Phase 2: 복합 변경 (5개 파일)

**시간**: 25분
**변경 유형**: 다중 Skill 참조 + AskUserQuestion 링크 업데이트

#### 2-1. cc-manager.md (10 changes)

**변경 패턴**:
```
1. moai-alfred-workflow → moai-core-workflow
2. moai-alfred-language-detection → moai-core-language-detection
3. moai-alfred-tag-scanning → moai-core-tag-scanning
```

**실행**:
```bash
cd .claude/agents/moai
sed -i 's/moai-alfred-workflow/moai-core-workflow/g' cc-manager.md
sed -i 's/moai-alfred-language-detection/moai-core-language-detection/g' cc-manager.md
sed -i 's/moai-alfred-tag-scanning/moai-core-tag-scanning/g' cc-manager.md
echo "✓ cc-manager.md updated"
```

#### 2-2. debug-helper.md (8 changes)

**변경 패턴**:
```
1. moai-alfred-ask-user-questions → moai-core-ask-user-questions
2. moai-alfred-language-detection → moai-core-language-detection
3. moai-alfred-tag-scanning → moai-core-tag-scanning
```

**실행**:
```bash
cd .claude/agents/moai
sed -i 's/moai-alfred-ask-user-questions/moai-core-ask-user-questions/g' debug-helper.md
sed -i 's/moai-alfred-language-detection/moai-core-language-detection/g' debug-helper.md
sed -i 's/moai-alfred-tag-scanning/moai-core-tag-scanning/g' debug-helper.md
echo "✓ debug-helper.md updated"
```

#### 2-3. doc-syncer.md (16 changes)

**변경 패턴**:
```
1. moai-alfred-ask-user-questions → moai-core-ask-user-questions
2. moai-alfred-tag-scanning → moai-core-tag-scanning
```

**실행**:
```bash
cd .claude/agents/moai
sed -i 's/moai-alfred-ask-user-questions/moai-core-ask-user-questions/g' doc-syncer.md
sed -i 's/moai-alfred-tag-scanning/moai-core-tag-scanning/g' doc-syncer.md
echo "✓ doc-syncer.md updated"
```

#### 2-4. git-manager.md (12 changes)

**변경 패턴**:
```
1. moai-alfred-ask-user-questions → moai-core-ask-user-questions
2. moai-alfred-git-workflow → moai-core-git-workflow
3. moai-alfred-trust-validation → moai-core-trust-validation
```

**실행**:
```bash
cd .claude/agents/moai
sed -i 's/moai-alfred-ask-user-questions/moai-core-ask-user-questions/g' git-manager.md
sed -i 's/moai-alfred-git-workflow/moai-core-git-workflow/g' git-manager.md
sed -i 's/moai-alfred-trust-validation/moai-core-trust-validation/g' git-manager.md
echo "✓ git-manager.md updated"
```

#### 2-5. implementation-planner.md (20 changes)

**변경 패턴**:
```
1. moai-alfred-ask-user-questions → moai-core-ask-user-questions
2. moai-alfred-language-detection → moai-core-language-detection
```

**실행**:
```bash
cd .claude/agents/moai
sed -i 's/moai-alfred-ask-user-questions/moai-core-ask-user-questions/g' implementation-planner.md
sed -i 's/moai-alfred-language-detection/moai-core-language-detection/g' implementation-planner.md
echo "✓ implementation-planner.md updated"
```

---

### Phase 3: 대규모 네임스페이스 업데이트 (6개 파일)

**시간**: 30분
**변경 유형**: 스킬 팩토리 및 validation 스킬 대규모 재정의

#### 3-1. agent-factory.md (12 changes)

```bash
cd .claude/agents/moai
sed -i 's/moai-alfred-agent-factory/moai-core-agent-factory/g' agent-factory.md
echo "✓ agent-factory.md updated"
```

#### 3-2. quality-gate.md (18 changes)

```bash
cd .claude/agents/moai
sed -i 's/moai-alfred-ask-user-questions/moai-core-ask-user-questions/g' quality-gate.md
sed -i 's/moai-alfred-trust-validation/moai-core-trust-validation/g' quality-gate.md
echo "✓ quality-gate.md updated"
```

#### 3-3. skill-factory.md (30 changes - 최대 규모)

```bash
cd .claude/agents/moai
sed -i 's/moai-alfred-skill-factory/moai-core-skill-factory/g' skill-factory.md
sed -i 's/moai-alfred-ask-user-questions/moai-core-ask-user-questions/g' skill-factory.md
echo "✓ skill-factory.md updated"
```

#### 3-4. spec-builder.md (18 changes)

```bash
cd .claude/agents/moai
sed -i 's/moai-alfred-spec-authoring/moai-core-spec-authoring/g' spec-builder.md
sed -i 's/moai-alfred-ask-user-questions/moai-core-ask-user-questions/g' spec-builder.md
sed -i 's/moai-alfred-ears-authoring/moai-core-ears-authoring/g' spec-builder.md
echo "✓ spec-builder.md updated"
```

#### 3-5. tdd-implementer.md (18 changes)

```bash
cd .claude/agents/moai
sed -i 's/moai-alfred-ask-user-questions/moai-core-ask-user-questions/g' tdd-implementer.md
sed -i 's/moai-alfred-language-detection/moai-core-language-detection/g' tdd-implementer.md
echo "✓ tdd-implementer.md updated"
```

#### 3-6. trust-checker.md (16 changes)

```bash
cd .claude/agents/moai
sed -i 's/moai-alfred-ask-user-questions/moai-core-ask-user-questions/g' trust-checker.md
sed -i 's/moai-alfred-trust-validation/moai-core-trust-validation/g' trust-checker.md
echo "✓ trust-checker.md updated"
```

---

### Phase 4: 최종 검증 (20분)

**단계 1**: 모든 alfred 참조 제거 확인

```bash
# 검증 명령어
grep -r "moai-alfred" .claude/agents/moai/

# 예상 결과
# (아무것도 출력되지 않음 = 성공)
```

**단계 2**: 모든 core 참조 추가 확인

```bash
# 검증 명령어
grep -r "moai-core" .claude/agents/moai/ | wc -l

# 예상 결과
# 200 이상 (이전: 0)
```

**단계 3**: 파일 무결성 검증

```bash
# 파일 수 확인
find .claude/agents/moai -name "*.md" | wc -l
# 결과: 31개

# 파일 크기 확인
ls -lh .claude/agents/moai/*.md | wc -l
# 결과: 31개
```

**단계 4**: Git diff 최종 확인

```bash
# 변경사항 요약
git diff --stat .claude/agents/moai/

# 예상 결과
# .claude/agents/moai/accessibility-expert.md        | 2 +-
# .claude/agents/moai/agent-factory.md              | 12 +++---
# ... (23개 파일)

# 상세 변경사항 확인 (선택사항)
git diff .claude/agents/moai/ | head -100
```

---

## ✅ 최종 체크리스트

### 사전 검사 (5분)
- [ ] 현재 브랜치 확인: `git branch`
  ```bash
  git branch
  # 결과: * release/0.26.0
  ```

- [ ] 작업 디렉토리 클린: `git status`
  ```bash
  git status
  # 결과: On branch release/0.26.0, nothing to commit, working tree clean
  ```

- [ ] 네트워크 연결 확인
  ```bash
  ping github.com -c 1
  ```

### 동기화 실행 (95분)
- [ ] Phase 1 완료 (15분)
  ```bash
  # 13개 파일 확인
  git diff .claude/agents/moai/*.md | grep "moai-core-language-detection" | wc -l
  # 결과: 13 이상
  ```

- [ ] Phase 2 완료 (25분)
  ```bash
  # 5개 파일 확인
  git diff .claude/agents/moai/{cc-manager,debug-helper,doc-syncer,git-manager,implementation-planner}.md | wc -l
  ```

- [ ] Phase 3 완료 (30분)
  ```bash
  # 6개 파일 확인
  git diff .claude/agents/moai/{agent-factory,quality-gate,skill-factory,spec-builder,tdd-implementer,trust-checker}.md | wc -l
  ```

- [ ] Phase 4 검증 완료 (20분)
  ```bash
  # alfred 참조 제거 확인
  grep -r "moai-alfred" .claude/agents/moai/ | wc -l
  # 결과: 0
  ```

### 사후 작업 (10분)
- [ ] 변경사항 리뷰
  ```bash
  git diff .claude/agents/moai/ | less
  ```

- [ ] 커밋 메시지 준비
  ```
  chore(agents): Sync with v0.26.0 templates (moai-alfred → moai-core)

  - Update 13 simple skill references (moai-alfred-language-detection)
  - Update 5 complex agents (cc-manager, debug-helper, doc-syncer, git-manager, implementation-planner)
  - Update 6 large-scale agents (agent-factory, quality-gate, skill-factory, spec-builder, tdd-implementer, trust-checker)
  - Add new Context7 MCP research strategy note
  - Total changes: 192 lines across 23 files
  ```

- [ ] 커밋 실행
  ```bash
  git add .claude/agents/moai/
  git commit -m "chore(agents): Sync with v0.26.0 templates (moai-alfred → moai-core)"
  ```

- [ ] 푸시 (선택사항)
  ```bash
  git push origin release/0.26.0
  ```

---

## 🔧 문제 해결

### 문제 1: Alfred 참조가 남아있음

**증상**:
```bash
grep -r "moai-alfred" .claude/agents/moai/ | wc -l
# 결과: 5 (0이 아님)
```

**해결책**:
```bash
# 문제 파일 찾기
grep -r "moai-alfred" .claude/agents/moai/

# 해당 파일의 모든 alfred 참조 변경
grep -r "moai-alfred" .claude/agents/moai/ | cut -d: -f1 | sort -u | while read file; do
  sed -i 's/moai-alfred-[a-z-]*/moai-core-&/g' "$file"
  echo "✓ Cleaned: $file"
done

# 재검증
grep -r "moai-alfred" .claude/agents/moai/ | wc -l
# 결과: 0
```

### 문제 2: 파일 손상

**증상**: 파일이 비어있거나 YAML 프론트매터가 손상됨

**확인**:
```bash
# YAML 프론트매터 확인
head -20 .claude/agents/moai/spec-builder.md | grep "^---\|^name:\|^description:"

# 예상:
# ---
# name: spec-builder
# description: "Use when: ..."
```

**해결책**:
```bash
# 백업에서 복원
cp .moai/backup/agents-sync-YYYY-MM-DD-HHMMSS/spec-builder.md .claude/agents/moai/

# 또는 템플릿에서 복사
cp src/moai_adk/templates/.claude/agents/moai/spec-builder.md .claude/agents/moai/
sed -i 's/moai-alfred-spec-authoring/moai-core-spec-authoring/g' .claude/agents/moai/spec-builder.md
# (모든 변경 재적용)
```

### 문제 3: 동기화 실패

**증상**: 스크립트 오류 또는 일부 파일만 업데이트됨

**디버깅**:
```bash
# 스크립트 verbose 모드로 실행
set -x
.moai/scripts/sync-agents.sh
set +x

# 또는 파일별 diff 확인
for file in .claude/agents/moai/*.md; do
  echo "=== $(basename $file) ==="
  diff "$file" "src/moai_adk/templates/.claude/agents/moai/$(basename $file)" | head -20
done
```

**해결책**:
```bash
# 특정 파일만 재동기화
sed -i 's/moai-alfred-ask-user-questions/moai-core-ask-user-questions/g' .claude/agents/moai/spec-builder.md

# 모든 파일 초기화 후 재시작
cp -r .moai/backup/agents-sync-YYYY-MM-DD-HHMMSS/* .claude/agents/moai/
.moai/scripts/sync-agents.sh
```

---

## 📞 롤백 절차

**만약 문제가 발생하면**:

```bash
# Step 1: 현재 변경사항 저장 (선택사항)
git stash

# Step 2: 백업에서 복원
BACKUP_DIR=".moai/backup/agents-sync-2025-11-19-HHMMSS"
cp -r "$BACKUP_DIR"/* .claude/agents/moai/

# Step 3: 복원 확인
git diff .claude/agents/moai/ | wc -l
# 결과: 0 (변경사항 없음)

# Step 4: Git 상태 확인
git status
```

---

## 📊 성공 기준

| 항목 | 기준 | 검증 명령어 |
|------|------|-----------|
| Alfred 참조 제거 | 0개 | `grep -r "moai-alfred" .claude/agents/moai/ \| wc -l` |
| Core 참조 추가 | 150+ | `grep -r "moai-core" .claude/agents/moai/ \| wc -l` |
| 파일 수 유지 | 31개 | `find .claude/agents/moai -name "*.md" \| wc -l` |
| YAML 프론트매터 유효 | 31개 | `head -5 .claude/agents/moai/*.md \| grep "^---"` |
| 노선 끝 | 0개 | `grep -r "---$" .claude/agents/moai/` |

---

## 🎯 다음 단계

### 즉시 (동기화 후)
1. 변경사항 검증: `git diff .claude/agents/moai/`
2. 커밋 생성: `git commit -m "chore(agents): Sync with v0.26.0"`
3. 테스트 실행: `/moai:0-project`

### 단기 (24시간 내)
1. 에이전트 기능성 테스트
2. SPEC 문서 생성 테스트 (`/moai:1-plan`)
3. TDD 구현 테스트 (`/moai:2-run`)

### 장기 (1주일 내)
1. 로컬 프로젝트에 변경사항 배포
2. 팀과 동기화
3. 문서 업데이트

---

## 📎 참고 자료

- **상세 분석**: `.moai/reports/agent-sync-analysis-2025-11-19.md`
- **자동화 스크립트**: `.moai/scripts/sync-agents.sh`
- **현재 브랜치**: release/0.26.0
- **메인 브랜치**: main

---

**문서 버전**: 1.0.0
**작성일**: 2025-11-19
**마지막 업데이트**: 2025-11-19
**상태**: 실행 준비 완료 ✅
