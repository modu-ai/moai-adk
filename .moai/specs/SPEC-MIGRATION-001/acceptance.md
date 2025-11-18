---
id: SPEC-MIGRATION-001
type: acceptance-criteria
version: 1.0.0
created: 2025-11-18
updated: 2025-11-18
author: GoosLab
---

# SPEC-MIGRATION-001 수용 기준 (Acceptance Criteria)

## 개요

SPEC-MIGRATION-001 구현의 수용 기준을 5개 주요 시나리오 + 3개 추가 엣지 케이스로 정의합니다.

모든 시나리오는 **Given/When/Then** (BDD) 형식으로 작성됩니다.

---

## 시나리오 1: 정상적인 마이그레이션 흐름 ✅

**시나리오명**: 기존 Alfred 폴더가 있는 프로젝트에서 update 명령 실행

**시나리오 ID**: AC-1

**우선순위**: P0 (Critical)

### Given (주어진 조건)

```
- MoAI-ADK v0.26.x 설치된 프로젝트 존재
- .claude/commands/alfred/ 폴더 존재 (5개 파일)
- .claude/agents/alfred/ 폴더 존재 (31개 파일)
- .claude/hooks/alfred/ 폴더 존재 (39개 파일)
- .claude/settings.json에 Alfred Hook 경로 존재:
  예) "uv run {{PROJECT_DIR}}/.claude/hooks/alfred/session_start__*.py"
- .moai/config/config.json에 migration.alfred_to_moai.migrated = false (또는 필드 없음)
- .moai/backups/ 디렉토리 쓰기 가능
```

### When (실행 단계)

```
1. 터미널에서 "moai-adk update" 명령 실행
2. 버전 확인 통과
3. 마이그레이션 필요 조건 충족
4. Alfred → Moai 마이그레이션 프로세스 시작
```

### Then (예상 결과)

```
✅ 콘솔 메시지 출력 (단계별):
   "[1/5] 프로젝트 백업 중..."
   "[2/5] Alfred 폴더 감지됨: commands, agents, hooks"
   "[3/5] Moai 템플릿 설치 중..."
   "[4/5] 경로 업데이트 중..."
   "[5/5] 정리 작업 중..."
   "✅ Alfred → Moai 마이그레이션 완료!"

✅ 디렉토리 구조 확인:
   - .claude/commands/moai/ 생성됨 (5개 파일)
   - .claude/agents/moai/ 생성됨 (31개 파일)
   - .claude/hooks/moai/ 생성됨 (39개 파일)
   - .claude/commands/alfred/ 삭제됨
   - .claude/agents/alfred/ 삭제됨
   - .claude/hooks/alfred/ 삭제됨

✅ settings.json 검증:
   - 모든 "hooks/alfred/" → "hooks/moai/" 변경됨
   - 모든 "commands/alfred/" → "commands/moai/" 변경됨
   - 구조 유효성 검증 (JSON 파싱 성공)

✅ config.json 상태 기록:
   {
     "migration": {
       "alfred_to_moai": {
         "migrated": true,
         "timestamp": "2025-11-18 14:30:00",
         "folders_installed": 3,
         "folders_removed": 3,
         "package_version": "0.27.0"
       }
     }
   }

✅ 백업 보존:
   - .moai/backups/alfred_to_moai_migration_{timestamp}/ 유지됨
```

### 검증 스크립트

```bash
#!/bin/bash
# 시나리오 1 검증

echo "=== 시나리오 1: 정상 마이그레이션 검증 ==="

# 1. Moai 폴더 존재 확인
[ -d ".claude/commands/moai" ] && echo "✅ commands/moai 폴더 존재" || echo "❌ commands/moai 폴더 없음"
[ -d ".claude/agents/moai" ] && echo "✅ agents/moai 폴더 존재" || echo "❌ agents/moai 폴더 없음"
[ -d ".claude/hooks/moai" ] && echo "✅ hooks/moai 폴더 존재" || echo "❌ hooks/moai 폴더 없음"

# 2. Alfred 폴더 삭제 확인
! [ -d ".claude/commands/alfred" ] && echo "✅ commands/alfred 삭제됨" || echo "❌ commands/alfred 아직 존재"
! [ -d ".claude/agents/alfred" ] && echo "✅ agents/alfred 삭제됨" || echo "❌ agents/alfred 아직 존재"
! [ -d ".claude/hooks/alfred" ] && echo "✅ hooks/alfred 삭제됨" || echo "❌ hooks/alfred 아직 존재"

# 3. settings.json 경로 검증
! grep -q "hooks/alfred" .claude/settings.json && echo "✅ settings.json alfred 참조 제거됨" || echo "❌ settings.json에 alfred 참조 남음"
grep -q "hooks/moai" .claude/settings.json && echo "✅ settings.json moai 경로 추가됨" || echo "❌ settings.json에 moai 경로 없음"

# 4. config.json 상태 확인
python3 -c "
import json
with open('.moai/config/config.json') as f:
    config = json.load(f)
    if config.get('migration', {}).get('alfred_to_moai', {}).get('migrated'):
        print('✅ config.json 마이그레이션 상태 기록됨')
    else:
        print('❌ config.json 마이그레이션 상태 없음')
"
```

---

## 시나리오 2: 백업 실패 시 중단 ❌

**시나리오명**: 백업 디렉토리에 쓰기 권한이 없는 경우

**시나리오 ID**: AC-2

**우선순위**: P0 (Critical)

### Given (주어진 조건)

```
- MoAI-ADK v0.26.x 설치된 프로젝트
- .claude/commands/alfred/ 폴더 존재
- .moai/backups/ 디렉토리 READ-ONLY 상태 (쓰기 권한 없음)
  예) chmod 444 .moai/backups/
```

### When (실행 단계)

```
1. "moai-adk update" 명령 실행
2. 마이그레이션 필요 조건 확인
3. 백업 생성 시도
```

### Then (예상 결과)

```
✅ 콘솔 에러 메시지 출력:
   "❌ 에러: 백업 생성 실패"
   "원인: .moai/backups/ 디렉토리 쓰기 권한 없음"
   "조치: 디렉토리 권한 확인 후 다시 시도하세요"

✅ 마이그레이션 프로세스 즉시 중단:
   - 어떤 파일도 수정되지 않음
   - Alfred 폴더 여전히 존재
   - Moai 폴더 생성되지 않음

✅ config.json 변경 없음:
   - migration.alfred_to_moai.migrated 필드 생성 안 됨
   - 이후 update 재실행 시 다시 마이그레이션 시도

✅ 프로젝트 상태:
   - 프로젝트 파일 손상 없음 (완전히 안전)
   - 백업도 생성 안 됨
```

### 검증 스크립트

```bash
#!/bin/bash
# 시나리오 2 검증

echo "=== 시나리오 2: 백업 실패 시 중단 검증 ==="

# 1. 백업 디렉토리 권한 제한
chmod 444 .moai/backups/

# 2. 마이그레이션 실행
moai-adk update 2>&1 | tee /tmp/migration.log

# 3. 에러 메시지 확인
grep -q "백업 생성 실패" /tmp/migration.log && echo "✅ 에러 메시지 출력됨" || echo "❌ 에러 메시지 없음"

# 4. 파일 변경 없음 확인
[ -d ".claude/commands/alfred" ] && echo "✅ commands/alfred 여전히 존재 (미변경)" || echo "❌ 파일 변경됨"
! [ -d ".claude/commands/moai" ] && echo "✅ commands/moai 생성 안 됨" || echo "❌ moai 폴더 생성됨 (실패함)"

# 5. config.json 변경 없음 확인
! grep -q "migrated.*true" .moai/config/config.json && echo "✅ config.json 변경 없음" || echo "❌ config.json 변경됨"

# 6. 권한 복구
chmod 755 .moai/backups/
```

---

## 시나리오 3: 마이그레이션 실패 및 자동 롤백 🔄

**시나리오명**: Moai 템플릿 설치 중 I/O 에러 발생

**시나리오 ID**: AC-3

**우선순위**: P0 (Critical)

### Given (주어진 조건)

```
- MoAI-ADK v0.26.x 설치된 프로젝트
- .claude/commands/alfred/, agents/alfred/, hooks/alfred/ 존재
- 백업 생성 완료됨
- Moai 템플릿 복사 중 I/O 에러 발생 (시뮬레이션)
  예) 디스크 공간 부족, 권한 오류 등
```

### When (실행 단계)

```
1. "moai-adk update" 명령 실행
2. 백업 생성 성공
3. Moai 템플릿 설치 시작
4. 복사 중 I/O 에러 발생
5. 에러 감지되면 자동 롤백 트리거
```

### Then (예상 결과)

```
✅ 콘솔 메시지 출력:
   "[1/5] 프로젝트 백업 중..."
   "✅ 백업 완료: .moai/backups/alfred_to_moai_migration_20251118_143000/"
   "[2/5] Alfred 폴더 감지됨..."
   "[3/5] Moai 템플릿 설치 중..."
   "❌ 에러: 템플릿 설치 실패 (I/O 에러)"
   "🔄 자동 롤백 시작..."
   "[1/3] 프로젝트 복원 중..."
   "[2/3] 설정 복원 중..."
   "[3/3] 마이그레이션 상태 초기화..."
   "✅ 롤백 완료"
   "💡 팁: 디스크 공간을 확인하고 다시 시도하세요"

✅ 프로젝트 상태 (롤백 후):
   - .claude/commands/alfred/ 복원됨 (원본 그대로)
   - .claude/agents/alfred/ 복원됨 (원본 그대로)
   - .claude/hooks/alfred/ 복원됨 (원본 그대로)
   - .moai/backups/alfred_to_moai_migration_*/ 여전히 존재 (증거)

✅ 부분 설치 파일 정리:
   - 불완전한 .claude/commands/moai/ 삭제됨 (있었다면)
   - 불완전한 .claude/agents/moai/ 삭제됨 (있었다면)
   - 불완전한 .claude/hooks/moai/ 삭제됨 (있었다면)

✅ config.json 상태:
   - migration.alfred_to_moai.migrated 필드 없음 (롤백됨)
   - 이후 update 재실행 시 다시 마이그레이션 시도 가능

✅ 복원 검증:
   - 백업된 파일과 복원된 파일 체크섬 일치
```

### 검증 스크립트

```bash
#!/bin/bash
# 시나리오 3 검증

echo "=== 시나리오 3: 실패 및 롤백 검증 ==="

# 1. 초기 상태 저장
find .claude/commands/alfred -type f | wc -l > /tmp/alfred_count_before.txt

# 2. 마이그레이션 실행 (실패 시뮬레이션)
moai-adk update 2>&1 | tee /tmp/migration.log

# 3. 에러 메시지 확인
grep -q "에러\|롤백" /tmp/migration.log && echo "✅ 에러 및 롤백 메시지 출력됨" || echo "❌ 메시지 없음"

# 4. Alfred 폴더 복원 확인
find .claude/commands/alfred -type f | wc -l > /tmp/alfred_count_after.txt
[ "$(cat /tmp/alfred_count_before.txt)" = "$(cat /tmp/alfred_count_after.txt)" ] && \
  echo "✅ Alfred 폴더 완전히 복원됨" || echo "❌ 복원 불완전"

# 5. Moai 폴더 정리 확인
! [ -d ".claude/commands/moai" ] && echo "✅ 부분 설치 모ai 폴더 정리됨" || \
  echo "⚠️  moai 폴더 존재함 (검수 필요)"

# 6. config.json 상태 확인
! grep -q "migrated.*true" .moai/config/config.json && \
  echo "✅ config.json 마이그레이션 상태 없음 (롤백됨)" || \
  echo "❌ config.json 상태 남아있음"

# 7. 백업 존재 확인
[ -d ".moai/backups/alfred_to_moai_migration_"* ] && \
  echo "✅ 백업 파일 보존됨 (증거)" || echo "⚠️  백업 파일 없음"
```

---

## 시나리오 4: 중복 실행 방지 ⏭️

**시나리오명**: 이미 마이그레이션된 프로젝트에서 update 재실행

**시나리오 ID**: AC-4

**우선순위**: P1 (High)

### Given (주어진 조건)

```
- MoAI-ADK v0.27.0+ 설치된 프로젝트 (이미 마이그레이션됨)
- .claude/commands/moai/ 폴더 존재
- .claude/agents/moai/ 폴더 존재
- .claude/hooks/moai/ 폴더 존재
- .claude/commands/alfred/, agents/alfred/, hooks/alfred/ 없음 (이미 삭제됨)
- .moai/config/config.json에 다음 내용 존재:
  {
    "migration": {
      "alfred_to_moai": {
        "migrated": true,
        "timestamp": "2025-10-01 10:00:00"
      }
    }
  }
```

### When (실행 단계)

```
1. "moai-adk update" 명령 실행
2. 마이그레이션 필요 여부 확인
3. config.json 상태 확인
```

### Then (예상 결과)

```
✅ 콘솔 메시지 출력:
   "ℹ️  정보: Alfred → Moai 마이그레이션이 이미 완료되었습니다"
   "타임스탬프: 2025-10-01 10:00:00"

✅ 마이그레이션 로직 스킵됨:
   - 백업 생성 안 됨
   - 폴더 복사 안 됨
   - 경로 업데이트 안 됨

✅ 일반 update 프로세스 계속 진행:
   - 템플릿 동기화 (moai 폴더만)
   - 버전 확인
   - 기타 필요한 작업

✅ 프로젝트 상태 (변경 없음):
   - Moai 폴더 유지
   - 이전 백업 유지 (.moai/backups/alfred_to_moai_migration_*)
```

### 검증 스크립트

```bash
#!/bin/bash
# 시나리오 4 검증

echo "=== 시나리오 4: 중복 실행 방지 검증 ==="

# 1. 이미 마이그레이션된 상태 설정
python3 -c "
import json
with open('.moai/config/config.json') as f:
    config = json.load(f)
config['migration'] = {
    'alfred_to_moai': {
        'migrated': True,
        'timestamp': '2025-10-01 10:00:00'
    }
}
with open('.moai/config/config.json', 'w') as f:
    json.dump(config, f, indent=2)
"

# 2. update 재실행
moai-adk update 2>&1 | tee /tmp/migration2.log

# 3. 스킵 메시지 확인
grep -q "이미 완료되었습니다\|마이그레이션 스킵" /tmp/migration2.log && \
  echo "✅ 마이그레이션 스킵 메시지 출력됨" || echo "❌ 메시지 없음"

# 4. 상태 미변경 확인
[ -d ".claude/commands/moai" ] && echo "✅ moai 폴더 유지됨" || echo "❌ moai 폴더 손상됨"
! [ -d ".claude/commands/alfred" ] && echo "✅ alfred 폴더 없음 (유지됨)" || echo "❌ alfred 폴더 생성됨"

# 5. 백업 유지 확인
backup_count=$(ls -d .moai/backups/alfred_to_moai_migration_* 2>/dev/null | wc -l)
[ "$backup_count" -ge 1 ] && echo "✅ 기존 백업 유지됨" || echo "❌ 백업 제거됨"
```

---

## 시나리오 5: settings.json Hook 경로 자동 업데이트 🔗

**시나리오명**: settings.json의 모든 Alfred Hook 경로가 Moai로 변경됨

**시나리오 ID**: AC-5

**우선순위**: P1 (High)

### Given (주어진 조건)

```
- MoAI-ADK v0.26.x 프로젝트
- Alfred 폴더 존재
- .claude/settings.json에 다음 내용 존재:
  {
    "hooks": {
      "SessionStart": [
        {
          "command": "uv run {{PROJECT_DIR}}/.claude/hooks/alfred/session_start__show_project_info.py"
        }
      ],
      "PreToolUse": [
        {
          "command": "uv run {{PROJECT_DIR}}/.claude/hooks/alfred/pre_tool__validate_command.py"
        }
      ],
      "PostToolUse": [
        {
          "command": "uv run {{PROJECT_DIR}}/.claude/hooks/alfred/post_tool__log_execution.py"
        }
      ]
    }
  }
```

### When (실행 단계)

```
1. "moai-adk update" 명령 실행
2. 마이그레이션 실행
3. settings.json 훅 경로 업데이트 단계
```

### Then (예상 결과)

```
✅ settings.json 자동 변경:
   원본:
   "uv run {{PROJECT_DIR}}/.claude/hooks/alfred/session_start__show_project_info.py"

   변경 후:
   "uv run {{PROJECT_DIR}}/.claude/hooks/moai/session_start__show_project_info.py"

✅ 모든 경로 변경 완료:
   - hooks/alfred → hooks/moai (모두)
   - commands/alfred → commands/moai (해당하는 경우)
   - agents/alfred → agents/moai (해당하는 경우)

✅ 변경 검증:
   ! grep -q "hooks/alfred\|commands/alfred\|agents/alfred" .claude/settings.json
   → 모든 alfred 참조 제거됨 ✅

   grep -q "hooks/moai" .claude/settings.json
   → 모든 moai 참조 추가됨 ✅

✅ 파일 구조 유지:
   - JSON 형식 유효 (파싱 가능)
   - 들여쓰기 보존
   - 주석 보존 (있었다면)
   - 기타 필드 변경 없음
```

### 검증 스크립트

```bash
#!/bin/bash
# 시나리오 5 검증

echo "=== 시나리오 5: Hook 경로 업데이트 검증 ==="

# 1. 마이그레이션 실행
moai-adk update

# 2. Alfred 참조 제거 확인
if ! grep -r "alfred" .claude/settings.json > /dev/null 2>&1; then
  echo "✅ settings.json에서 alfred 참조 제거됨"
else
  echo "❌ settings.json에 아직 alfred 참조 존재:"
  grep "alfred" .claude/settings.json
fi

# 3. Moai 경로 추가 확인
if grep -q "hooks/moai\|commands/moai" .claude/settings.json; then
  echo "✅ settings.json에 moai 경로 추가됨"
else
  echo "❌ settings.json에 moai 경로 없음"
fi

# 4. JSON 유효성 검사
python3 -m json.tool .claude/settings.json > /dev/null 2>&1 && \
  echo "✅ JSON 형식 유효" || echo "❌ JSON 형식 오류"

# 5. 특정 경로 검증
python3 -c "
import json
with open('.claude/settings.json') as f:
    settings = json.load(f)

hooks = settings.get('hooks', {})
for hook_type, commands in hooks.items():
    for cmd_obj in commands:
        cmd = cmd_obj.get('command', '')
        if 'alfred' in cmd:
            print(f'❌ {hook_type}에 alfred 경로 남음: {cmd}')
        elif 'moai' in cmd:
            print(f'✅ {hook_type}에 moai 경로 있음')
"
```

---

## 추가 엣지 케이스 시나리오

### 시나리오 6: Alfred 폴더 부분만 존재하는 경우

**시나리오 ID**: AC-6

**상황**: commands/alfred는 있지만 agents/alfred는 없는 경우

**기대 동작**:
```
- 존재하는 폴더만 백업/마이그레이션
- 존재하지 않는 폴더는 오류 아님
- 부분 마이그레이션 허용
- 경로 업데이트도 부분 적용 가능
```

**검증**:
```bash
# Alfred 폴더 부분 제거
rm -rf .claude/agents/alfred/

# 마이그레이션 실행
moai-adk update

# 기대 결과
[ -d ".claude/commands/moai" ] && echo "✅ commands 마이그레이션됨"
! [ -d ".claude/agents/alfred" ] && echo "✅ agents도 마이그레이션 필요 없음"
```

---

### 시나리오 7: 수동으로 settings.json 커스터마이징한 경우

**시나리오 ID**: AC-7

**상황**: 사용자가 settings.json을 커스텀 설정한 경우 (주석, 추가 필드 등)

**기대 동작**:
```
- 정규식 기반 경로 치환 (안전)
- 커스텀 설정 보존
- 주석 보존
- 추가 필드 보존
```

**검증**:
```bash
# 커스텀 내용 추가
cat >> .claude/settings.json << 'EOF'
{
  "custom_field": "user_value",  // 사용자 주석
  "alfred_custom_hook": "uv run {{PROJECT_DIR}}/.claude/hooks/alfred/custom.py"
}
EOF

# 마이그레이션 실행
moai-adk update

# 검증
grep -q "custom_field" .claude/settings.json && echo "✅ 커스텀 필드 보존됨"
grep -q "alfred_custom_hook.*moai" .claude/settings.json && echo "✅ 커스텀 경로도 업데이트됨"
```

---

### 시나리오 8: Alfred와 Moai 폴더가 동시에 존재하는 경우

**시나리오 ID**: AC-8

**상황**: 이전 마이그레이션이 부분 실패해서 Alfred와 Moai 폴더가 동시 존재

**기대 동작**:
```
- Alfred 폴더 감지 (우선권)
- 기존 Moai 폴더 백업
- 새로운 Moai 설치 (덮어쓰기 아님, 충돌 감지)
- 사용자에게 선택지 제공 (자동 처리 vs 수동 조치)
```

**검증**:
```bash
# Moai 폴더 미리 생성
mkdir -p .claude/commands/moai

# 마이그레이션 실행
moai-adk update 2>&1 | tee /tmp/migration8.log

# 충돌 감지 메시지 확인
grep -q "충돌\|경고\|선택" /tmp/migration8.log && \
  echo "✅ 충돌 감지 및 메시지 출력됨"
```

---

## 수용 기준 요약

| 시나리오 | ID | 상태 | 기준 |
|---------|----|----|------|
| 정상 마이그레이션 | AC-1 | ✅ | 모든 폴더 마이그레이션, 경로 변경, 상태 기록 |
| 백업 실패 중단 | AC-2 | ❌ | 에러 후 즉시 중단, 파일 미변경 |
| 실패 후 롤백 | AC-3 | 🔄 | 에러 감지 시 자동 롤백 |
| 중복 실행 방지 | AC-4 | ⏭️ | 상태 확인 후 스킵 |
| Hook 경로 업데이트 | AC-5 | 🔗 | 모든 경로 자동 변경 |
| 부분 폴더 존재 | AC-6 | ⚙️ | 부분 마이그레이션 |
| 커스텀 설정 보존 | AC-7 | 🔐 | 사용자 커스터마이징 유지 |
| 동시 폴더 존재 | AC-8 | ⚠️ | 충돌 감지 및 처리 |

---

## 테스트 실행 순서

**권장 순서** (의존성 고려):

1. **AC-1** (정상): 기본 경로 확보
2. **AC-4** (중복 방지): AC-1 이후, 이미 마이그레이션된 상태에서
3. **AC-2** (백업 실패): 초기 상태 재설정 후
4. **AC-3** (롤백): AC-1 상황에서 에러 시뮬레이션
5. **AC-5** (경로 업데이트): AC-1 결과 검증
6. **AC-6, AC-7, AC-8** (엣지 케이스): 차례대로

---

## 자동화 테스트 스크립트

모든 시나리오를 자동화할 통합 테스트 스크립트:

```bash
#!/bin/bash
# tests/integration/test_migration_acceptance.sh

set -e

echo "================================"
echo "SPEC-MIGRATION-001 수용 기준 테스트"
echo "================================"

# AC-1: 정상 마이그레이션
echo "[AC-1] 정상 마이그레이션..."
./tests/acceptance/ac-1-normal-migration.sh

# AC-2: 백업 실패
echo "[AC-2] 백업 실패 시 중단..."
./tests/acceptance/ac-2-backup-failure.sh

# AC-3: 롤백
echo "[AC-3] 실패 및 롤백..."
./tests/acceptance/ac-3-failure-rollback.sh

# AC-4: 중복 실행 방지
echo "[AC-4] 중복 실행 방지..."
./tests/acceptance/ac-4-duplicate-prevention.sh

# AC-5: 경로 업데이트
echo "[AC-5] Hook 경로 업데이트..."
./tests/acceptance/ac-5-path-update.sh

echo ""
echo "================================"
echo "✅ 모든 수용 기준 테스트 완료"
echo "================================"
```

---

## End of Acceptance Criteria
