# Git Worktree CLI 사용 가이드

## 소개

Git Worktree는 단일 저장소에서 여러 개의 독립적인 작업 디렉토리를 생성할 수 있는 Git 기능입니다. MoAI-ADK의 `moai-wt` 명령어는 이를 활용하여 여러 SPEC을 병렬로 개발할 수 있는 환경을 제공합니다.

### 이점
- 동시 다중 SPEC 개발 (3-5개 병렬 작업)
- 브랜치 전환 없이 즉각적인 컨텍스트 전환
- 독립적인 node_modules/.venv 환경
- 자동 충돌 감지 및 병합 관리

## 요구사항
- Git >= 2.7.0 (worktree 지원)
- Python >= 3.9
- moai-adk >= 0.31.0

## 설치 및 초기 설정

### 1. moai-adk 업데이트
```bash
uv pip install moai-adk --upgrade
```

### 2. Git 설정 확인
```bash
git --version  # 2.7.0 이상 필요
```

### 3. Worktree 디렉토리 생성
```bash
# 자동으로 생성되지만, 수동 생성도 가능
mkdir -p ~/worktrees/{{PROJECT_NAME}}
```

## 기본 명령어

### 1. 새 Worktree 생성
```bash
moai-wt new SPEC-001
# 자동으로 ~/worktrees/PROJECT_NAME/SPEC-001 생성
# feature/SPEC-001 브랜치 자동 생성
```

### 2. 활성 Worktree 목록 확인
```bash
moai-wt list
# 테이블 형식으로 표시:
# SPEC-ID      | Branch           | Path              | Status
# SPEC-001     | feature/SPEC-001 | ~/worktrees/.../  | Active
# SPEC-002     | feature/SPEC-002 | ~/worktrees/.../  | Active
```

### 3. Worktree 전환 (새 셸 실행)
```bash
moai-wt switch SPEC-001
# SPEC-001 디렉토리에서 새로운 셸 실행
# 종료 시 이전 셸로 복귀
```

### 4. Shell eval 패턴으로 전환
```bash
# 별칭 설정 (권장)
alias wt-go='eval $(moai-wt go'

# 사용
wt-go SPEC-001
# 현재 셸에서 SPEC-001 디렉토리로 이동
```

### 5. Worktree 제거
```bash
moai-wt remove SPEC-001 --force
# 워크트리 삭제 및 레지스트리 업데이트
# 미커밋 변경사항 확인 후 진행
```

### 6. 상태 확인
```bash
moai-wt status
# 모든 worktree의 상태를 표시
# Dirty 상태 확인
# 병합 가능 브랜치 탐지
```

### 7. 병합된 Worktree 정리
```bash
moai-wt clean
# 이미 병합된 worktree 자동 감지
# 대화형 확인 후 삭제
```

### 8. 레지스트리 동기화
```bash
moai-wt sync SPEC-001
# Main 브랜치와 최신 변경사항 동기화
# 자동 병합 시도 (충돌 시 알림)
```

## 워크플로우

### 워크플로우 1: 병렬 SPEC 개발

```bash
# Step 1: 첫 번째 SPEC 워크트리 생성
moai-wt new SPEC-002
cd ~/worktrees/PROJECT/SPEC-002

# Step 2: SPEC 계획 수립
/moai:1-plan "사용자 인증 기능" --worktree

# Step 3: 구현 시작
/moai:2-run SPEC-002

# Step 4: 동시에 다른 SPEC 작업 가능
moai-wt new SPEC-003
cd ~/worktrees/PROJECT/SPEC-003
/moai:2-run SPEC-003

# Step 5: SPEC-002 완료 후 병합
wt-go SPEC-002
git add .
git commit -m "feat: implement user authentication"
moai-wt remove SPEC-002

# Step 6: 완료 문서화
/moai:3-sync SPEC-002
```

### 워크플로우 2: 긴급 패치 (Main 브랜치 기반)

```bash
# Step 1: 긴급 패치 워크트리 생성
moai-wt new hotfix/critical-bug --base main

# Step 2: 긴급 수정 작업
cd ~/worktrees/PROJECT/hotfix/critical-bug
# ... 코드 수정 ...

# Step 3: 테스트 및 검증
npm test  # 또는 pytest

# Step 4: 병합 및 정리
moai-wt remove hotfix/critical-bug
# Main 브랜치에 수동 병합 또는 PR 생성
```

### 워크플로우 3: 팀 협업

```bash
# 개발자 A: SPEC-001 작업
moai-wt new SPEC-001
cd ~/worktrees/PROJECT/SPEC-001
/moai:2-run SPEC-001

# 개발자 B: SPEC-002 작업 (동시 진행)
moai-wt new SPEC-002
cd ~/worktrees/PROJECT/SPEC-002
/moai:2-run SPEC-002

# 개발자 C: SPEC-003 작업 (동시 진행)
moai-wt new SPEC-003
cd ~/worktrees/PROJECT/SPEC-003
/moai:2-run SPEC-003

# 각자 완료 후 PR 생성
# /moai:3-sync로 문서 동기화
# 팀 리뷰 후 메인 브랜치에 병합
```

## 고급 기능

### Worktree 최적화 설정

```bash
# 설정 조회
moai-wt config worktree_root
moai-wt config auto_cleanup

# 설정 변경
moai-wt config worktree_root ~/my-worktrees
moai-wt config auto_cleanup true
```

### 충돌 해결

```bash
# 병합 중 충돌 발생
moai-wt sync SPEC-001

# 충돌 파일 수정
# 충돌 마커 (<<<<<<, ======, >>>>>>)를 확인하고 수정

# 해결 후 완료
git add .
git commit -m "Merge conflicts resolved"
```

### 성능 최적화

- **Worktree 최대 개수**: 5개 권장 (시스템 리소스에 따라)
- **디스크 공간**: 각 worktree마다 약 100-500MB
- **메모리**: 병렬 작업 시 각각 100-200MB

## 문제 해결

### Q: Worktree 생성 실패

**원인**: Git 버전 불일치 또는 디스크 공간 부족

**해결**:
```bash
git --version  # 2.7.0 이상 확인
df -h          # 디스크 공간 확인
moai-wt clean  # 불필요한 worktree 정리
```

### Q: 충돌이 발생했을 때

**원인**: Main 브랜치와 동시 수정

**해결**:
```bash
moai-wt sync SPEC-001  # 자동 동기화 시도
# 충돌이 있으면 수동으로 해결
# 이후 커밋
```

### Q: Worktree 삭제 후 디렉토리가 남음

**원인**: .git 폴더가 직접 있지 않고 .git 파일에서 참조

**해결**:
```bash
moai-wt remove SPEC-001 --force
rm -rf ~/worktrees/PROJECT/SPEC-001  # 필요시 수동 삭제
```

## 팁과 권장사항

### 1. 셸 별칭 설정 (권장)

```bash
# ~/.zshrc 또는 ~/.bashrc에 추가
alias wt-new='moai-wt new'
alias wt-go='eval $(moai-wt go'
alias wt-list='moai-wt list'
alias wt-sync='moai-wt sync'
alias wt-status='moai-wt status'
alias wt-clean='moai-wt clean'
alias wt-remove='moai-wt remove'
```

### 2. 정기적인 정리

```bash
# 일주일마다 병합된 worktree 정리
moai-wt clean

# 불필요한 브랜치 확인
git branch -a | grep feature/SPEC
```

### 3. 안전한 작업 습관

- 항상 `moai-wt list`로 활성 워크트리 확인
- 삭제 전 `moai-wt status`로 상태 확인
- 중요 작업 전 백업 (git push 권장)

---

**참고**: Git Worktree와 기존 Git 브랜치 전환은 완전히 호환됩니다. 전통적인 `git checkout`과 함께 사용 가능합니다.
