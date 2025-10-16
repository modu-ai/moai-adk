---
name: alfred:9-checkpoint
description: "Checkpoint 통합 관리 (생성/조회/복구/정리)"
argument-hint: "[create|list|restore|clean|config] [options]"
allowed-tools:
  - Task
  - Read
  - Write
  - Bash(git:*)
  - Bash(cat:*)
  - Bash(ls:*)
  - TodoWrite
---

# 🛡️ MoAI-ADK Checkpoint 통합 관리 시스템

## 🎯 커맨드 목적

자동/수동 체크포인트 생성, 조회, 복구, 정리를 단일 인터페이스로 제공합니다.

---

## 📋 서브커맨드 개요

| 서브커맨드 | 용도 | Phase | 에이전트 위임 |
|-----------|------|-------|-------------|
| **create** | 수동 체크포인트 생성 | 2-Phase | 없음 (직접 처리) |
| **list** | 체크포인트 조회 | 1-Phase | 없음 (직접 처리) |
| **restore** | 체크포인트 복구 | 2-Phase | git-manager |
| **clean** | 오래된 체크포인트 정리 | 2-Phase | git-manager |
| **config** | 자동 체크포인트 설정 | 1-Phase | 없음 (직접 처리) |

---

## 📊 Checkpoint 시스템 개요

MoAI-ADK는 위험한 작업 전 자동으로 checkpoint를 생성합니다.

### 자동 Checkpoint 생성 트리거

| 작업 유형 | 감지 조건 | Checkpoint 이름 |
|---------|--------|---------------|
| **대규모 삭제** | `rm -rf`, `git rm` | `before-delete-{timestamp}` |
| **Git 병합** | `git merge`, `git reset --hard` | `before-merge-{timestamp}` |
| **스크립트 실행** | `python`, `node`, `bash` | `before-script-{timestamp}` |
| **중요 파일 수정** | `CLAUDE.md`, `config.json` | `before-critical-file-{timestamp}` |
| **대규모 리팩토링** | ≥10개 파일 동시 수정 | `before-refactor-{timestamp}` |

### Checkpoint 특징

- **Local branch**: 원격 저장소 오염 방지 (로컬 전용)
- **최대 10개 유지**: FIFO + 7일 제한 (자동 정리)
- **투명한 동작**: 백그라운드에서 자동 생성, 사용자에게 알림
- **완전한 상태 저장**: Uncommitted changes 포함 (자동 커밋)

---

## 1️⃣ create - 수동 체크포인트 생성

### 용도
중요한 작업 직전 안전망을 수동으로 생성합니다.

### 사용법
```bash
/alfred:9-checkpoint create [--name "description"]
```

### STEP 1: 분석 및 계획 수립

다음 항목을 분석하여 보고합니다:

1. **Git 상태 확인**
   ```bash
   git status
   git branch --show-current
   ```

2. **최근 Checkpoint 이력**
   ```bash
   cat .moai/checkpoints.log | tail -5
   ```

3. **디스크 사용량 확인**
   ```bash
   git branch | grep "^  before-" | wc -l
   # 10개 제한 근접 시 경고
   ```

4. **생성 계획 보고**
   ```
   📊 Checkpoint 생성 계획
   - 브랜치 이름: before-manual-20251016-100000
   - 현재 브랜치: develop
   - Uncommitted changes: 5 files (자동 커밋됨)
   - 예상 디스크 사용: ~5MB
   - 현재 Checkpoint 개수: 7/10

   💡 모든 변경사항이 checkpoint에 포함됩니다.

   승인하시겠습니까? (진행/수정/중단)
   ```

### STEP 2: 실행 (사용자 승인 후)

1. **Uncommitted Changes 자동 커밋**
   ```bash
   # Uncommitted changes 확인
   if [ -n "$(git status --porcelain)" ]; then
       # 모든 변경사항 스테이징
       git add -A

       # 자동 커밋 (checkpoint 메시지)
       git commit -m "🛡️ CHECKPOINT: Uncommitted changes before ${CUSTOM_NAME:-manual}

       Auto-saved by /alfred:9-checkpoint create
       Timestamp: $(date -Iseconds)"
   fi
   ```

2. **Checkpoint 브랜치 생성**
   ```bash
   # 커스텀 이름이 있으면 사용, 없으면 'manual' 사용
   TIMESTAMP=$(date +%Y%m%d-%H%M%S)
   BRANCH_NAME="before-${CUSTOM_NAME:-manual}-${TIMESTAMP}"

   # 현재 상태(커밋 포함)를 checkpoint 브랜치로 생성
   git branch ${BRANCH_NAME}
   ```

3. **.moai/checkpoints.log 업데이트**
   ```bash
   echo '{"timestamp": "'$(date -Iseconds)'", "branch": "'${BRANCH_NAME}'", "operation": "manual", "description": "'${DESCRIPTION}'", "has_uncommitted": true}' >> .moai/checkpoints.log
   ```

4. **완료 보고**
   ```
   ✅ Checkpoint 생성 완료
   - 브랜치: before-refactor-start-20251016-100000
   - Uncommitted changes: 자동 커밋됨 (5 files)
   - 커밋 SHA: a1b2c3d4...
   - 복구 방법: /alfred:9-checkpoint restore <ID>

   💡 Checkpoint는 모든 변경사항을 포함합니다 (커밋됨)
   ```

### 예시

```bash
# 기본 이름 (manual)
/alfred:9-checkpoint create
→ before-manual-20251016-093000

# 커스텀 이름
/alfred:9-checkpoint create --name "refactor-start"
→ before-refactor-start-20251016-093000
```

---

## 2️⃣ list - 체크포인트 조회

### 용도
전체 checkpoint 목록 및 상세 정보를 조회합니다.

### 사용법
```bash
/alfred:9-checkpoint list [--filter <type>] [--last <N>] [--details <ID>]
```

### 실행 (단일 Phase)

1. **.moai/checkpoints.log 파싱**
   ```bash
   cat .moai/checkpoints.log | jq -r '.'
   ```

2. **Git 브랜치 목록 크로스체크**
   ```bash
   git branch | grep "^  before-"
   ```

3. **필터 적용 (선택적)**
   - `--filter delete`: 삭제 관련 checkpoint만
   - `--filter merge`: 병합 관련 checkpoint만
   - `--filter manual`: 수동 생성 checkpoint만
   - `--last N`: 최근 N개만 표시

4. **테이블 형식 출력**

```
📋 Checkpoint 목록 (총 5개)

ID  | 브랜치 이름                        | 작업 타입   | 생성 시간           | 안전
----|-----------------------------------|------------|--------------------|---------
1   | before-delete-20251016-090000     | delete     | 2025-10-16 09:00   | No
2   | before-merge-20251016-091500      | merge      | 2025-10-16 09:15   | No
3   | before-restore-20251016-092000    | restore    | 2025-10-16 09:20   | Yes
4   | before-manual-20251016-093000     | manual     | 2025-10-16 09:30   | No
5   | before-refactor-start-20251016... | manual     | 2025-10-16 09:35   | No

💡 상세 정보: /alfred:9-checkpoint list --details <ID>
🔄 복구: /alfred:9-checkpoint restore <ID>
🧹 정리: /alfred:9-checkpoint clean
```

### 상세 조회 모드 (--details <ID>)

```bash
# Git 로그 조회
git log -1 <branch> --format="%H%n%an%n%ai%n%s"

# 현재와의 차이
git diff --stat <branch>..HEAD

# 변경된 파일 목록
git diff --name-status <branch>..HEAD
```

**출력 예시**:
```
📊 Checkpoint 상세 정보 (ID: 3)

브랜치: before-restore-20251016-092000
커밋 SHA: a1b2c3d4...
작성자: @Goos
생성 시간: 2025-10-16 09:20:00
작업 타입: restore (안전 체크포인트)

📁 현재와의 차이 (5 files changed)
  M  src/auth/service.py
  M  src/auth/models.py
  D  src/auth/legacy.py
  A  tests/auth/test_service.py
  M  CLAUDE.md

📈 Diff 통계
  +150 -80 lines
```

### 예시

```bash
# 전체 목록
/alfred:9-checkpoint list

# 최근 3개만
/alfred:9-checkpoint list --last 3

# 수동 생성만
/alfred:9-checkpoint list --filter manual

# 상세 정보
/alfred:9-checkpoint list --details 3
```

---

## 3️⃣ restore - 체크포인트 복구

### 용도
특정 checkpoint로 프로젝트 상태를 복구합니다.

### 사용법
```bash
/alfred:9-checkpoint restore <ID|branch-name> [--strategy <1|2|3>] [--files <paths>]
```

### STEP 1: 분석 및 계획 수립

다음 항목을 분석하여 보고합니다:

1. **대상 Checkpoint 유효성 확인**
   ```bash
   git rev-parse --verify <branch-name>
   ```

2. **현재 Git 상태 확인**
   ```bash
   git status
   # Uncommitted changes 감지
   ```

3. **복구 영향도 분석**
   ```bash
   git diff --name-status <checkpoint-branch>..HEAD
   ```

4. **복구 전략 제안**

```
⚠️ 복구 영향도 분석

대상 Checkpoint: before-delete-20251016-100500 (ID: 3)
현재 브랜치: develop
Uncommitted changes: Yes (3 files)

📁 복구 시 변경될 파일 (23 files)
  A  src/auth/service.py (새로 추가됨)
  M  src/auth/models.py (수정됨)
  D  src/auth/legacy.py (삭제됨)
  ...

🔄 복구 전략 선택:

[전략 1] Branch 전환 (안전, 권장)
  - 현재 작업 보존
  - Checkpoint 브랜치로 전환
  - 언제든 돌아올 수 있음

  실행: git checkout <checkpoint-branch>

[전략 2] Hard Reset (강력, 주의)
  - 현재 브랜치를 Checkpoint로 리셋
  - ⚠️ Uncommitted changes 손실
  - Stash 자동 생성 후 진행

  실행: git stash && git reset --hard <checkpoint-branch>

[전략 3] 선택적 복구 (추천: 부분 복구)
  - 특정 파일/디렉토리만 복구
  - 다른 변경사항 보존

  실행: git checkout <checkpoint-branch> -- <files>

⚠️ 안전 Checkpoint 자동 생성: before-restore-{timestamp}

어떤 전략을 선택하시겠습니까? (1/2/3)
또는 수정 요청: "수정 [내용]"
```

### STEP 2: 실행 (사용자 승인 후)

#### ⚙️ 에이전트 호출 방법

**git-manager에게 복구 작업 위임**:

```markdown
Task tool 호출:
- subagent_type: "git-manager"
- description: "Checkpoint 복구"
- prompt: "다음 checkpoint로 복구를 진행해주세요.

          복구 대상:
          - Checkpoint 브랜치: {checkpoint_branch}
          - Checkpoint ID: {checkpoint_id}
          - 복구 전략: {user_strategy}

          사전 작업:
          1. 안전 checkpoint 생성: before-restore-{timestamp}
          2. Uncommitted changes가 있으면 stash 생성

          전략별 실행:
          - 전략 1: git checkout {checkpoint_branch}
          - 전략 2: git stash && git reset --hard {checkpoint_branch}
          - 전략 3: git checkout {checkpoint_branch} -- {files}

          사후 작업:
          1. 복구 결과 검증 (git status)
          2. .moai/checkpoints.log 업데이트
          3. 완료 보고"
```

#### 복구 프로세스

1. **안전 Checkpoint 생성** (git-manager 실행)
   ```bash
   TIMESTAMP=$(date +%Y%m%d-%H%M%S)
   git branch before-restore-${TIMESTAMP}
   ```

2. **Uncommitted Changes 처리** (git-manager 실행)
   ```bash
   if [ -n "$(git status --porcelain)" ]; then
       git stash push -m "Before checkpoint restore ${TIMESTAMP}"
   fi
   ```

3. **선택된 전략 실행** (git-manager 실행)
   ```bash
   # 전략 1
   git checkout ${CHECKPOINT_BRANCH}

   # 전략 2
   git reset --hard ${CHECKPOINT_BRANCH}

   # 전략 3
   git checkout ${CHECKPOINT_BRANCH} -- ${FILES}
   ```

4. **복구 결과 검증** (git-manager 실행)
   ```bash
   git status
   git log -1 --oneline
   ```

5. **완료 보고**
   ```
   ✅ Checkpoint 복구 완료

   복구 정보:
   - 복구된 Checkpoint: before-delete-20251016-100500
   - 사용된 전략: 전략 3 (선택적 복구)
   - 복구된 파일: 23 files
   - 안전 Checkpoint: before-restore-20251016-101530

   다음 단계:
   1. 변경사항 검토: git diff
   2. 테스트 실행: pytest / npm test
   3. 커밋 및 Push (필요 시)
   ```

### 예시

```bash
# 기본 복구 (대화형)
/alfred:9-checkpoint restore 3

# 브랜치 이름으로 복구
/alfred:9-checkpoint restore before-delete-20251016-100500

# 전략 지정
/alfred:9-checkpoint restore 3 --strategy 1

# 선택적 복구 (특정 파일만)
/alfred:9-checkpoint restore 3 --strategy 3 --files "src/auth/"
```

---

## 4️⃣ clean - 체크포인트 정리

### 용도
오래되거나 불필요한 checkpoint를 삭제하여 디스크 공간을 확보합니다.

### 사용법
```bash
/alfred:9-checkpoint clean [--older-than <days>] [--keep <N>] [--force]
```

### STEP 1: 분석 및 계획 수립

다음 항목을 분석하여 보고합니다:

1. **현재 Checkpoint 목록 조회**
   ```bash
   git branch | grep "^  before-"
   cat .moai/checkpoints.log
   ```

2. **삭제 대상 필터링**
   - 기본값: 7일 이상 + 최신 5개 유지
   - `--older-than N`: N일 이상된 checkpoint
   - `--keep N`: 최신 N개는 보존

3. **디스크 사용량 계산**
   ```bash
   # 각 브랜치의 커밋 개수 및 크기 추정
   for branch in $(git branch | grep "before-"); do
       git log --oneline $branch | wc -l
   done
   ```

4. **삭제 계획 보고**

```
🧹 Checkpoint 정리 계획

현재 상태:
- 전체 Checkpoint: 12개
- 7일 이상: 7개
- 최신 5개 보존

삭제 대상 (7개):
  1. before-delete-20251009-143000 (7일 전)
  2. before-merge-20251010-091500 (6일 전)
  3. before-manual-20251011-100000 (5일 전)
  4. before-script-20251011-150000 (5일 전)
  5. before-refactor-20251012-090000 (4일 전)
  6. before-critical-file-20251013-110000 (3일 전)
  7. before-delete-20251014-140000 (2일 전)

유지 (5개 - 최신):
  8. before-merge-20251015-091500 (1일 전)
  9. before-delete-20251016-090000 (오늘)
  10. before-manual-20251016-093000 (오늘)
  11. before-restore-20251016-101530 (오늘)
  12. before-refactor-start-20251016-102000 (오늘)

예상 효과:
- 삭제될 브랜치: 7개
- 절감 예상: ~35MB

⚠️ 주의: 삭제된 checkpoint는 복구할 수 없습니다.

승인하시겠습니까? (진행/수정/중단)
```

### STEP 2: 실행 (사용자 승인 후)

#### ⚙️ 에이전트 호출 방법

**git-manager에게 정리 작업 위임**:

```markdown
Task tool 호출:
- subagent_type: "git-manager"
- description: "Checkpoint 브랜치 정리"
- prompt: "다음 checkpoint 브랜치들을 삭제해주세요.

          삭제 대상 브랜치 목록:
          {branch_list}

          실행:
          1. 각 브랜치 삭제: git branch -D <branch>
          2. .moai/checkpoints.log 업데이트 (삭제된 항목 제거)
          3. 정리 결과 보고

          주의사항:
          - 강제 삭제 (-D) 사용
          - 삭제 전 브랜치 존재 확인
          - 에러 발생 시 중단하고 보고"
```

#### 정리 프로세스

1. **브랜치 삭제** (git-manager 실행)
   ```bash
   for branch in ${DELETE_LIST[@]}; do
       git branch -D $branch
   done
   ```

2. **.moai/checkpoints.log 업데이트** (git-manager 실행)
   ```bash
   # 삭제된 브랜치 제외하고 재생성
   cat .moai/checkpoints.log | jq -r 'select(.branch | IN("'${KEEP_LIST}'"))' > .moai/checkpoints.log.tmp
   mv .moai/checkpoints.log.tmp .moai/checkpoints.log
   ```

3. **완료 보고**
   ```
   ✅ Checkpoint 정리 완료

   정리 결과:
   - 삭제된 Checkpoint: 7개
   - 유지된 Checkpoint: 5개
   - 절감된 공간: ~35MB

   남은 Checkpoint:
   - before-merge-20251015-091500
   - before-delete-20251016-090000
   - before-manual-20251016-093000
   - before-restore-20251016-101530
   - before-refactor-start-20251016-102000

   💡 현재 사용량: 5/10
   ```

### 예시

```bash
# 기본 정리 (7일 이상, 최신 5개 유지)
/alfred:9-checkpoint clean

# 30일 이상, 최신 10개 유지
/alfred:9-checkpoint clean --older-than 30 --keep 10

# 강제 삭제 (확인 없이)
/alfred:9-checkpoint clean --force

# 모든 checkpoint 정리 (주의!)
/alfred:9-checkpoint clean --keep 0 --force
```

---

## 5️⃣ config - 자동 체크포인트 설정

### 용도
PreToolUse hook의 자동 checkpoint 트리거 조건을 관리합니다.

### 사용법
```bash
/alfred:9-checkpoint config [--set <key>=<value>] [--show]
```

### 실행 (단일 Phase)

1. **.moai/config.json 읽기**
   ```bash
   cat .moai/config.json | jq '.git_strategy.checkpoint_*'
   ```

2. **설정 표시 (--show 또는 기본)**

```
⚙️ 자동 Checkpoint 설정

[전역 설정]
✅ checkpoint_enabled: true (활성화)
📊 checkpoint_max_count: 10 (최대 개수)
📅 checkpoint_retention_days: 7 (보관 기간)

[트리거 설정]
✅ large_deletion: true (대규모 삭제 감지)
✅ risky_refactoring: true (대규모 리팩토링 감지)
✅ git_merge: true (Git 병합 감지)
✅ script_execution: true (스크립트 실행 감지)
✅ critical_file_modification: true (중요 파일 수정 감지)

💡 설정 변경: /alfred:9-checkpoint config --set <key>=<value>

예시:
  /alfred:9-checkpoint config --set checkpoint_max_count=15
  /alfred:9-checkpoint config --set checkpoint_enabled=false
```

3. **설정 변경 (--set 옵션)**

```bash
# .moai/config.json 업데이트
jq '.git_strategy.checkpoint_max_count = 15' .moai/config.json > tmp.json
mv tmp.json .moai/config.json
```

4. **변경 완료 보고**
   ```
   ✅ 설정 변경 완료

   변경 내용:
   - checkpoint_max_count: 10 → 15

   💡 변경사항은 다음 세션부터 적용됩니다.
   ```

### 설정 가능 항목

**전역 설정**:
- `checkpoint_enabled`: 자동 checkpoint 활성화 (true/false)
- `checkpoint_max_count`: 최대 보관 개수 (1-50)
- `checkpoint_retention_days`: 보관 기간 (1-90일)

**트리거 설정** (각 트리거 개별 활성화):
- `large_deletion`: 대규모 삭제 감지
- `risky_refactoring`: 대규모 리팩토링 감지
- `git_merge`: Git 병합 감지
- `script_execution`: 스크립트 실행 감지
- `critical_file_modification`: 중요 파일 수정 감지

### 예시

```bash
# 현재 설정 조회
/alfred:9-checkpoint config

# 최대 개수 변경
/alfred:9-checkpoint config --set checkpoint_max_count=15

# 보관 기간 변경
/alfred:9-checkpoint config --set checkpoint_retention_days=14

# 자동 checkpoint 비활성화
/alfred:9-checkpoint config --set checkpoint_enabled=false

# 특정 트리거 비활성화
/alfred:9-checkpoint config --set script_execution=false
```

---

## ⚠️ 에러 처리

### 예상 실패 케이스

| 케이스 | 증상 | 심각도 | 대응 |
|--------|------|--------|------|
| **Dirty working directory** | 커밋되지 않은 변경사항 | ⚠️ Warning | Stash 자동 생성 후 계속 |
| **Checkpoint 브랜치 없음** | 복구 대상 브랜치 미존재 | ❌ Critical | debug-helper 자동 호출 |
| **Git 충돌** | 복구 시 충돌 발생 | ❌ Critical | 안전 checkpoint로 롤백 |
| **로그 파일 손상** | JSON 파싱 실패 | ⚠️ Warning | Git 브랜치로 재생성 |
| **디스크 부족** | Checkpoint 생성 실패 | ❌ Critical | 자동 정리 후 재시도 |

### 에러 메시지 예시

```bash
# 복구 실패 시
❌ Checkpoint 복구 실패: before-delete-20251016-100500
  → 원인: Git 충돌 (src/auth/service.py)
  → 안전 checkpoint로 자동 롤백: before-restore-20251016-101530
  → git-manager가 debug-helper 호출 중...

⚠️ Uncommitted changes 감지 (3 files)
  → 자동 Stash 생성: stash@{0} "Before checkpoint restore"
  → 계속 진행하시겠습니까? (y/N)

ℹ️ 최대 checkpoint 개수 초과 (12/10)
  → 자동 정리 권장: /alfred:9-checkpoint clean
  → 또는 설정 변경: /alfred:9-checkpoint config --set checkpoint_max_count=15

❌ Checkpoint 브랜치를 찾을 수 없음: before-delete-20251016-999999
  → 가능한 Checkpoint 목록: /alfred:9-checkpoint list
  → 로그 파일 불일치 감지 - 자동 재생성 중...
```

---

## 💡 사용 예시 (시나리오별)

### 시나리오 1: 대규모 리팩토링 전후

```bash
# 1. 리팩토링 시작 전 checkpoint 생성
/alfred:9-checkpoint create --name "refactor-auth-module"

# Phase 1 보고
📊 Checkpoint 생성 계획
- 브랜치: before-refactor-auth-module-20251016-100000
- 현재 브랜치: develop (clean)

사용자: "진행"

# Phase 2 완료
✅ Checkpoint 생성: before-refactor-auth-module-20251016-100000

# 2. 리팩토링 작업 수행...
# (여러 파일 수정)

# 3. 문제 발생 → 복구
/alfred:9-checkpoint list
# ID 5: before-refactor-auth-module-20251016-100000

/alfred:9-checkpoint restore 5

# Phase 1: 영향도 분석
⚠️ 23 files 변경됨
전략 선택: 1 (Branch 전환)

사용자: "진행"

# Phase 2: git-manager 실행
✅ 복구 완료
```

---

### 시나리오 2: 자동 Checkpoint 활용

```bash
# 1. 세션 시작 메시지
🚀 MoAI-ADK Session Started
   Checkpoints: 3 available

# 2. 실수로 파일 삭제
rm -rf src/auth/

# Alfred가 자동 감지 및 checkpoint 생성
🛡️ Checkpoint created: before-delete-20251016-100500

# 3. 즉시 복구
/alfred:9-checkpoint restore before-delete-20251016-100500

# 전략 3 선택: 선택적 복구
사용자: "3"
파일 경로: "src/auth/"

✅ 복구 완료: src/auth/ (23 files)
```

---

### 시나리오 3: 정기적인 정리

```bash
# 1. Checkpoint 개수 확인
/alfred:9-checkpoint list
# 총 12개 ⚠️ 최대 10개 초과

# 2. 정리 실행
/alfred:9-checkpoint clean

# Phase 1: 정리 계획
🧹 삭제 대상: 7개 (7일 이상)
유지: 5개 (최신)
절감: ~35MB

사용자: "진행"

# Phase 2: git-manager 실행
✅ 정리 완료: 7개 삭제, 35MB 절감
```

---

## 🔗 관련 시스템

### PreToolUse Hook 통합

Checkpoint는 Claude Code의 PreToolUse hook에서 자동으로 생성됩니다:

```python
# .claude/hooks/alfred/moai_hooks.py

def handle_pre_tool_use(payload):
    """위험한 작업 전 자동 checkpoint 생성"""
    tool = payload.get("tool")
    args = payload.get("arguments")

    is_risky, operation_type = detect_risky_operation(tool, args)

    if is_risky:
        checkpoint_id = create_checkpoint(operation_type)
        print(f"🛡️ Checkpoint created: {checkpoint_id}")
```

### Checkpoint 로그

모든 checkpoint는 `.moai/checkpoints.log`에 기록됩니다:

```bash
# 로그 파일 위치
.moai/checkpoints.log

# 로그 형식 (JSON Lines)
{"timestamp": "2025-10-16T10:05:00+09:00", "branch": "before-delete-20251016-100500", "operation": "delete"}

# 최근 10개 조회
tail -10 .moai/checkpoints.log | jq
```

---

## 🎯 다음 단계

Checkpoint 작업 후:

1. **변경사항 검토**: `git diff` 또는 `git log`로 확인
2. **테스트 실행**: 복구 후 정상 동작 확인
3. **커밋 및 Push**: 필요 시 원격에 반영
4. **다음 작업**: `/alfred:1-spec`, `/alfred:2-build` 등 계속 진행

---

## 🔗 관련 문서

- **SPEC**: `.moai/specs/SPEC-CHECKPOINT-EVENT-001/spec.md`
- **구현**: `src/moai_adk/core/git/checkpoint.py`
- **Hook**: `.claude/hooks/alfred/moai_hooks.py`
- **설정**: `.moai/config.json` (git_strategy.checkpoint_*)
- **에이전트**: `git-manager` (Git 작업 위임)
