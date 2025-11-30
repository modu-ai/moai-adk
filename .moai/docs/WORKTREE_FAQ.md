# Git Worktree FAQ

## 기본 개념

### Q1: Git Worktree란?
Git Worktree는 하나의 Git 저장소에서 여러 개의 작업 디렉토리(working tree)를 동시에 유지할 수 있게 해주는 Git 기능입니다. 각 worktree는 독립적인 브랜치를 가지므로 여러 작업을 병렬로 진행할 수 있습니다.

### Q2: Worktree와 브랜치의 차이?
- **브랜치**: 하나의 작업 디렉토리 내에서 코드 히스토리의 다른 라인
- **Worktree**: 여러 개의 작업 디렉토리를 물리적으로 분리
- **결합 사용 가능**: 각 worktree는 다른 브랜치를 가짐

### Q3: Worktree가 필요한 이유?
- 빠른 컨텍스트 전환 (브랜치 전환 불필요)
- 병렬 개발 (여러 팀원이 동시 작업)
- 독립적인 환경 (node_modules, .venv 분리)
- 병합 충돌 감소

---

## 사용 관련

### Q4: Worktree는 몇 개까지 만들 수 있나?
기술적으로는 제한이 없지만, 실무 권장 사항:
- **최적**: 3-4개
- **최대**: 5개 (시스템 리소스 고려)
- **초과**: 디스크 공간과 메모리 부족 주의

### Q5: Worktree 생성 후 어떻게 진행하나?
```bash
moai-wt new SPEC-001    # 생성
wt-go SPEC-001          # 이동 (현재 셸)
# 또는
moai-wt switch SPEC-001 # 새 셸에서 열기
```

### Q6: 여러 SPEC을 병렬로 개발할 수 있나?
네! 각 SPEC마다 워크트리를 생성하면 병렬 개발 가능:
```bash
moai-wt new SPEC-002 & moai-wt new SPEC-003 & moai-wt new SPEC-004
```

### Q7: Worktree에서 git push는?
각 worktree의 브랜치에서 독립적으로 push 가능:
```bash
wt-go SPEC-001
git push origin feature/SPEC-001
```

---

## 문제 해결

### Q8: "Worktree already exists" 에러
**원인**: 같은 이름의 worktree가 이미 있음
**해결**:
```bash
moai-wt list          # 확인
moai-wt remove OLD-NAME --force
moai-wt new SPEC-001  # 다시 생성
```

### Q9: 충돌이 발생했을 때?
**원인**: Main과의 동시 수정
**해결**:
```bash
wt-go SPEC-001
moai-wt sync SPEC-001  # 자동 시도
# 충돌 파일 수정
git add .
git commit -m "Resolve conflicts"
```

### Q10: Worktree 삭제 후 폴더가 안 지워짐
**원인**: .git 파일이 남아있음 (worktree 참조)
**해결**:
```bash
moai-wt remove SPEC-001 --force
# 필요시 수동 삭제
rm -rf ~/worktrees/PROJECT/SPEC-001
```

### Q11: 실수로 worktree 삭제했을 때?
**복구 불가능**: Git worktree는 복구되지 않음
**예방**: 항상 push 후 삭제
```bash
git push origin feature/SPEC-001  # 먼저 푸시
moai-wt remove SPEC-001           # 그 후 삭제
```

### Q12: Worktree 내에서 브랜치 전환 불가?
**정상**: Worktree는 특정 브랜치 고정
**해결**: 다른 worktree로 이동하거나 새로 생성

---

## 성능 & 리소스

### Q13: Worktree의 디스크 공간 요구량?
각 worktree마다 대략:
- 소스 코드: 10-50 MB
- node_modules: 100-500 MB
- .venv: 50-200 MB
- **총계**: 150-750 MB per worktree

### Q14: 여러 worktree가 느려지나?
아니오. 각 worktree는 독립적:
- 컴파일 속도: 영향 없음
- 메모리: 각각 약 100MB (활성 상태)
- 총 용량: 누적

### Q15: node_modules 공유 가능?
권장하지 않음. 각 worktree별 독립 설치 권장.

---

## 팀 협업

### Q16: 팀원과 worktree 공유 가능?
아니오. Worktree는 로컬 개발용.
대신 브랜치를 push하고 협업:
```bash
git push origin feature/SPEC-001
# 다른 팀원이 체크아웃 가능
```

### Q17: CI/CD와 연동되나?
네! Push된 브랜치는 자동으로 CI/CD 실행.

### Q18: PR 생성은?
```bash
wt-go SPEC-001
git push origin feature/SPEC-001
gh pr create --base main  # PR 생성
```

---

## 마이그레이션

### Q19: 기존 브랜치 기반 워크플로우에서 변경?
점진적 마이그레이션 권장:
```bash
# 기존: git checkout feature/X
# 신규: moai-wt new feature/X && wt-go feature/X
```

### Q20: Worktree 대신 도커 컨테이너?
Worktree가 더 가볍고 빠름. 도커는 보조적으로 사용.

---

## 베스트 프랙티스

### Q21: Worktree 사용 시 주의사항?
1. 항상 push 후 삭제
2. 정기적인 `moai-wt clean`
3. 중요 작업 전 백업
4. 팀과 컨벤션 공유

### Q22: 이상적인 워크플로우?
```
1. moai-wt new SPEC-XXX
2. wt-go SPEC-XXX
3. /moai:2-run SPEC-XXX (개발)
4. git push
5. gh pr create (리뷰)
6. git merge (병합)
7. moai-wt remove SPEC-XXX
```

### Q23: Shell 별칭 설정 권장?
네! 매우 권장:
```bash
alias wt-go='eval $(moai-wt go'
alias wt-new='moai-wt new'
# ~/.zshrc 또는 ~/.bashrc에 추가
```

---

## 지원 및 문제

### Q24: 문제 발생 시 어떻게?
```bash
# 1단계: 상태 확인
moai-wt status

# 2단계: 로그 확인
git log --oneline

# 3단계: Git 상태 확인
git status
```

### Q25: 버그 리포팅
GitHub Issues에 다음 정보와 함께:
- `moai-wt --version`
- `git --version`
- 재현 단계
- 오류 메시지
