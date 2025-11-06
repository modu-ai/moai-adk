# Tool Hooks 상세 가이드

도구 실행 전/후에 자동으로 실행되는 Hook들입니다.

## 🎯 목적

### PreToolUse Hook

도구 실행 **전**:

- 위험한 명령 차단 (git push --force, rm -rf)
- 권한 검증
- 컨텍스트 전달

### PostToolUse Hook

도구 실행 **후**:

- 결과 분석
- 오류 감지
- 자동 수정 제안

## 🛡️ PreToolUse Hook

### 차단하는 명령들

```bash
# ❌ 절대 차단
git push --force          # 강제 푸시
git reset --hard          # 하드 리셋
rm -rf /                  # 전체 삭제
chmod -R 777 /            # 권한 전체 오픈

# ⚠️ 확인 후 실행
git rebase -i             # 인터랙티브 리베이스
rm *.py                   # 다중 파일 삭제
```

### 권한 검증 로직

```bash
# Permission 확인
if command in dangerous_list:
    # settings.json 확인
    if "deny" in permissions:
        → 실행 차단
    elif "ask" in permissions:
        → 사용자 확인 요청
    else:
        → 실행 허용
```

### 예시: Git Push 검증

```bash
# git push 실행 시
PreToolUse Hook 실행:
1. "push" 감지
2. "push --force" 확인 → NO
3. 대상 브랜치 확인 → develop (OK)
4. 원격 상태 확인 → 업데이트됨
5. ✅ 실행 허용
```

## 📊 PostToolUse Hook

### 결과 분석

```bash
# Tool 실행 후
PostToolUse Hook:
1. 종료 코드 확인
2. stdout/stderr 분석
3. 부작용 감지
4. 자동 수정 제안
```

### 오류 감지 예시

#### Bash 명령 오류

```bash
# 사용자 명령
mkdir /Users/goos/test/nested/dir

# PreToolUse: 부모 디렉토리 확인 → 없음
# PostToolUse 결과:
❌ mkdir: cannot create directory: No such file or directory

🔧 자동 수정 제안:
   mkdir -p /Users/goos/test/nested/dir
```

#### Git 병합 충돌

```bash
# 사용자 명령
git merge feature/auth

# PostToolUse 결과:
⚠️ Merge conflict detected in src/auth.py

🔧 해결 방법:
1. 충돌 부분 수정
2. git add src/auth.py
3. git commit
```

### 자동 수정 프로토콜

```
1️⃣ 오류 분석
   └─→ 원인 파악

2️⃣ 수정 가능성 판단
   ├─ YES → 3단계
   └─ NO → 가이드만 제시

3️⃣ 사용자 확인
   └─→ AskUserQuestion

4️⃣ 자동 수정 실행
   └─→ 재실행

5️⃣ 결과 검증
   └─→ 성공 확인
```

## 🔍 Hook 검증 규칙

| Tool  | PreToolUse     | PostToolUse    |
| ----- | -------------- | -------------- |
| Bash  | 명령 검증      | 종료 코드 확인 |
| Git   | 브랜치 확인    | 병합 상태 확인 |
| Read  | 파일 경로 확인 | 인코딩 검증    |
| Write | 경로 검증      | 사이즈 제한    |
| Edit  | 파일 존재 확인 | 문법 검증      |

## ⚙️ Hook 설정

### .claude/settings.json

```json
{
  "hooks": {
    "pre_tool_use": {
      "enabled": true,
      "timeout": 5000,
      "dangerous_commands": [
        "git push --force",
        "git reset --hard",
        "rm -rf"
      ]
    },
    "post_tool_use": {
      "enabled": true,
      "timeout": 5000,
      "auto_fix": true,
      "error_detection": true
    }
  }
}
```

## 📋 Hook 체인 예시

```
User: git push

↓ PreToolUse Hook
├─→ "push" 감지
├─→ 브랜치 확인: develop
├─→ 강제 푸시 확인: 없음
└─→ ✅ 실행 허용

↓ Git Push 실행
$ git push origin develop

↓ PostToolUse Hook
├─→ 종료 코드: 0 (성공)
├─→ stdout 분석
└─→ ✅ 성공 메시지

완료!
```

## 🆘 Hook 오류 처리

### Hook 자체 오류

```bash
❌ Hook 실행 실패
│
├─ Timeout (5초 초과)
│  └─→ 경고만 출력, 도구 실행
│
├─ Permission 오류
│  └─→ 권한 조정 후 재시도
│
└─ Script 오류
   └─→ 로그 저장, 계속 진행
```

### 디버깅

```bash
# Hook 로그 확인
cat ~/.claude/projects/*/hook-logs/*.log

# Hook 비활성화
# .claude/settings.json:
# "hooks.enabled": false

# 특정 Hook만 비활성화
# "hooks.pre_tool_use.enabled": false
```

______________________________________________________________________

**다음**: [Hooks 개요](index.md) 또는 [SessionStart Hook](session.md)
